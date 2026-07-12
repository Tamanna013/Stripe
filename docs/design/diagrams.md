# FlowGuard Diagrams

This document consolidates the key architecture diagrams for FlowGuard.

## Full Deployment Topology

```mermaid
flowchart TB
    subgraph GH["GitHub Actions CI/CD"]
        Build[Build/Test/Scan/Sign] --> ArgoCD
    end

    subgraph ArgoCD["Argo CD / Argo Rollouts"]
        StagingSync[Sync: Staging]
        ProdGate{Manual Approval Gate}
        ProdCanary[Canary Rollout: Prod]
    end

    Build --> StagingSync
    StagingSync --> ProdGate
    ProdGate -->|approved| ProdCanary

    subgraph StagingCluster["EKS: flowguard-staging"]
        StagingPods[All 17 services @ N=1-2 replicas]
    end

    subgraph ProdCluster["EKS: flowguard-prod (multi-AZ)"]
        subgraph AZ1["AZ-1"]
            Pods1[Service replicas]
        end
        subgraph AZ2["AZ-2"]
            Pods2[Service replicas]
        end
        subgraph AZ3["AZ-3"]
            Pods3[Service replicas]
        end
    end

    StagingSync --> StagingCluster
    ProdCanary --> ProdCluster

    ProdCluster --> RDS[(RDS PostgreSQL Multi-AZ)]
    ProdCluster --> MSK[(MSK Kafka, 3 AZ)]
    ProdCluster --> Elasticache[(ElastiCache Redis, Multi-AZ w/ Sentinel)]
    ProdCluster --> ClickHouse[(ClickHouse cluster, replicated)]
```

## gRPC Service Communication

```mermaid
flowchart LR
    APIGW[API Gateway] -->|gRPC| IncidentEngine
    APIGW -->|gRPC| TraceCorrelation
    APIGW -->|gRPC| MetricsAgg[Metrics Aggregation]
    APIGW -->|gRPC| DepGraph[Dependency Graph]
    APIGW -->|gRPC| SemanticSearch
    APIGW -->|gRPC| TrafficReplay
    APIGW -->|gRPC| ChaosEngine

    BFF -->|GraphQL resolvers, internally calling| APIGW

    IncidentEngine -->|gRPC| RootCauseEngine
    IncidentEngine -->|gRPC| PostmortemGen[Postmortem Generator]
    IncidentEngine -->|gRPC| Auth[Auth Service - AuthZ check]

    RootCauseEngine -->|gRPC| SemanticSearch
    RootCauseEngine -->|gRPC| TraceCorrelation
    RootCauseEngine -->|gRPC| MetricsAgg
    RootCauseEngine -->|gRPC| DepGraph

    PostmortemGen -->|gRPC| MetricsAgg
    PostmortemGen -->|gRPC| DepGraph

    AlertGrouping[Alert Grouping] -->|gRPC| IncidentEngine
    AlertGrouping -->|gRPC| DepGraph

    TrafficReplay -->|gRPC| TraceCorrelation
    ChaosEngine -->|gRPC| MetricsAgg
    ChaosEngine -->|gRPC| IncidentEngine

    IngestionGateway[Ingestion Gateway] -.->|no gRPC to biz services, Kafka only| MSK[(Kafka)]

    APIGW -.mTLS via mesh.-> IncidentEngine
    APIGW -.mTLS via mesh.-> TraceCorrelation
```

## Kafka Message Flow Map

```mermaid
flowchart TB
    IG[Ingestion Gateway] --> T1[otel.spans.raw]
    IG --> T2[otel.logs.raw]
    IG --> T3[otel.metrics.raw]

    T1 --> TC[Trace Correlation Service]
    TC --> T4[traces.completed]
    T4 --> AG[Alert Grouping]
    T4 --> DG[Dependency Graph Service]
    T4 --> RCE[Root Cause Engine - evidence pull]

    T2 --> LP[Log Processing Service]
    LP --> T5[log.anomaly.detected]
    LP --> T6[logs.embedding.queue]
    T5 --> AG
    T6 --> SS[Semantic Search Service]

    T3 --> MA[Metrics Aggregation Service]
    MA --> T7[metric.slo_burn.alert]
    T7 --> AG

    AG -->|gRPC, not Kafka| IE[Incident Engine]
    IE --> T8[incident.state.changed]
    T8 --> PG[Postmortem Generator]
    T8 --> Notify[Notifier / Paging]

    CE[Chaos Engine] --> T9[chaos.experiment.events]
    T9 --> IE

    TRE[Traffic Replay Engine] --> T10[replay.session.events]
    T10 --> BFF[BFF - live progress]

    T1 -.dlq.-> T1DLQ[otel.spans.raw.dlq]
    T2 -.dlq.-> T2DLQ[otel.logs.raw.dlq]
    T4 -.dlq.-> T4DLQ[traces.completed.dlq]
```

## Master Request Lifecycle — Full Sequence

```mermaid
sequenceDiagram
    autonumber
    participant User as End User
    participant Svc as Monitored Service
    participant OTel as OTel Pipeline
    participant Kafka
    participant TC as Trace Correlation
    participant LP as Log Processing
    participant MA as Metrics Aggregation
    participant AG as Alert Grouping
    participant IE as Incident Engine
    participant RCE as Root Cause Engine
    participant LG as LangGraph
    participant PM as Postmortem Generator
    participant Eng as On-call Engineer

    User->>Svc: Request (fails downstream)
    Svc->>OTel: Emit spans/logs/metrics
    OTel->>Kafka: Publish raw telemetry
    Kafka->>TC: Consume spans
    Kafka->>LP: Consume logs
    Kafka->>MA: Consume metrics
    TC->>Kafka: traces.completed (has_error=true)
    LP->>Kafka: log.anomaly.detected
    MA->>Kafka: metric.slo_burn.alert
    Kafka->>AG: correlate signals
    AG->>AG: fingerprint + cluster (6.1)
    AG->>IE: OpenIncident(cluster)
    IE->>IE: persist, state=DETECTED
    IE->>RCE: AnalyzeIncident(incident_id)
    RCE->>RCE: retrieve_evidence, correlate_with_topology
    RCE->>LG: generate_hypotheses
    LG-->>RCE: ranked hypotheses
    RCE->>IE: RootCauseReport
    IE->>Eng: Page (via Notifier)
    Eng->>IE: transition -> MITIGATING
    Eng->>Eng: apply fix
    Eng->>IE: transition -> RESOLVED
    IE->>PM: GeneratePostmortem(incident_id)
    PM->>PM: draft (pass 1) + grounding check (pass 2)
    PM-->>Eng: Draft postmortem, flags if any
    Eng->>PM: Resolve flags, Publish
```
