package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
)

// Config Service maintains dynamic configuration and pushes updates via Redis Pub/Sub
func main() {
	log.Println("Starting Config Service...")

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	
	// Test Redis connection
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Printf("Warning: Redis not connected: %v", err)
	} else {
		log.Println("Connected to Redis for Config Pub/Sub")
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down Config Service...")
}
