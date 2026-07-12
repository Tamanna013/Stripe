package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// IngestionServer implements the flowguard.ingestion.v1.IngestionGateway service.
type IngestionServer struct {
	// UnimplementedIngestionGatewayServer
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	limiter := ratelimiter.NewTokenBucketLimiter(redisClient)

	// Initialize Kafka Producer
	producer := kafkaproducer.NewProducer([]string{"localhost:9092"})
	defer producer.Close()

	// Initialize gRPC Server
	grpcServer := grpc.NewServer()
	ingestionServer := &server.Server{
		Limiter:  limiter,
		Producer: producer,
	}
	
	// pb.RegisterIngestionGatewayServer(grpcServer, ingestionServer) // Uncomment when proto is built
	_ = ingestionServer

	reflection.Register(grpcServer)

	// Graceful shutdown setup
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Ingestion Gateway started on port %s", port)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down Ingestion Gateway gracefully...")
	grpcServer.GracefulStop()
	log.Println("Ingestion Gateway shutdown complete.")
}
