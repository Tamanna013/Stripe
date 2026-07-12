import os
import logging
from concurrent import futures
import grpc

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("root-cause-engine")

def serve():
    port = os.environ.get("PORT", "50054")
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    # TODO: Register Root Cause Engine servicer here once proto is generated
    
    server.add_insecure_port(f'[::]:{port}')
    server.start()
    logger.info(f"Root Cause Engine (LangGraph Orchestrator) started on port {port}")
    server.wait_for_termination()

if __name__ == '__main__':
    serve()
