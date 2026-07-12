package server_test

import (
	"context"
	"testing"
	"github.com/flowguard/ingestion-gw/server"
	"github.com/flowguard/ingestion-gw/ratelimiter"
	"github.com/flowguard/ingestion-gw/kafkaproducer"
	
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestFailOpenRateLimiter validates the failure mode defined in Part 11.3:
// "Ingestion Gateway's test suite must include a test asserting fail-open rate limiting when Redis is unreachable"
func TestFailOpenRateLimiter(t *testing.T) {
	// 1. Setup Redis Client pointing to an invalid/down endpoint
	downRedis := redis.NewClient(&redis.Options{
		Addr: "localhost:9999", // Deliberately wrong port
	})
	limiter := ratelimiter.NewTokenBucketLimiter(downRedis)

	// 2. Setup mocked Kafka producer (nil brokers so it won't connect)
	producer := kafkaproducer.NewProducer([]string{"localhost:9998"})

	// 3. Initialize server
	srv := &server.Server{
		Limiter:  limiter,
		Producer: producer,
	}

	req := &server.PushSpansRequest{
		SourceId: "test-source-1",
		OtlpPayload: []byte("mock-data"),
	}

	// 4. Assert
	_, err := srv.PushSpans(context.Background(), req)
	
	// We expect a Kafka error here (since Kafka is also unreachable in this unit test),
	// but we MUST NOT get a Rate Limiter / RESOURCE_EXHAUSTED error. 
	// If it fails open, the rate limiter check passes through cleanly.
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.ResourceExhausted {
			t.Fatalf("Expected fail-open when Redis is down, but got rate limited: %v", err)
		}
	}
}
