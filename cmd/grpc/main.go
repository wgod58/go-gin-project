package main

import (
	"log"

	"go-gin-project/config"
	"go-gin-project/grpc/server"
	"go-gin-project/internal/pkg/cache"
)

func main() {
	if err := config.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	cacheService, err := cache.New()
	if err != nil {
		log.Printf("Warning: Failed to initialize Redis: %v", err)
		log.Println("Application will continue without caching")
	}

	if err := server.StartGrpcServer(cacheService); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
