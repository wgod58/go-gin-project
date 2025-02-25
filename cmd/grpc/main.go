package main

import (
	"log"

	"go-gin-project/config"
	"go-gin-project/grpc/server"
	"go-gin-project/services"
)

func main() {
	// Initialize DB first
	if err := config.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize cache
	cache, err := services.NewRedisCache()
	if err != nil {
		log.Printf("Warning: Failed to initialize Redis: %v", err)
		log.Println("Application will continue without caching")
	}

	// Start gRPC server
	if err := server.StartGrpcServer(cache); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
