import os
import logging
from concurrent import futures
import grpc

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("postmortem-generator")

def serve():
    port = os.environ.get("PORT", "50056")
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    # TODO: Register Postmortem Generator servicer here
    
    server.add_insecure_port(f'[::]:{port}')
    server.start()
    logger.info(f"Postmortem Generator Service started on port {port}")
    server.wait_for_termination()

if __name__ == '__main__':
    serve()
