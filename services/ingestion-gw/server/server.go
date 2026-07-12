package server

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/flowguard/ingestion-gw/kafkaproducer"
	"github.com/flowguard/ingestion-gw/ratelimiter"
	
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// In an actual compiled proto setup, you'd import the generated pb package.
// For scaffolding purposes, we define the interfaces we expect from the proto.
type PushSpansRequest struct {
	SourceId    string
	OtlpPayload []byte
}
type PushSpansResponse struct {
	Success       bool
	AcceptedSpans int32
	DroppedSpans  int32
}

// Server implements the gRPC interface for Ingestion Gateway
type Server struct {
	Limiter  *ratelimiter.TokenBucketLimiter
	Producer *kafkaproducer.Producer
}

func (s *Server) PushSpans(ctx context.Context, req *PushSpansRequest) (*PushSpansResponse, error) {
	// 1. Rate Limiting Check
	allowed, err := s.Limiter.Allow(ctx, req.SourceId, 5000, 10000) // Default config
	if err != nil {
		log.Printf("Rate limiter error (failing open): %v", err)
	} else if !allowed {
		return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded for source")
	}

	// 2. Inject FlowGuard Request ID & Validate (mocked injection logic here)
	flowguardReqID := uuid.New().String()
	_ = flowguardReqID // In real logic, unmarshal OTLP, inject attribute, re-marshal

	// 3. Publish to Kafka (partitioned by a trace_id hash ideally; here we use SourceId)
	err = s.Producer.PublishSpan(ctx, []byte(req.SourceId), req.OtlpPayload)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish to kafka: %v", err)
	}

	return &PushSpansResponse{
		Success:       true,
		AcceptedSpans: 1, // Mocked parsed span count
		DroppedSpans:  0,
	}, nil
}
