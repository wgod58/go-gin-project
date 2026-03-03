package server

import (
	"fmt"
	"log"
	"net"
	"os"

	"go-gin-project/api/proto"
	"go-gin-project/config"
	"go-gin-project/internal/app/service"
	"go-gin-project/internal/pkg/model"
	"go-gin-project/internal/pkg/repository"

	"google.golang.org/grpc"
)

func StartGrpcServer(cache model.CacheService) error {
	// Create services using the already initialized DB and cache
	if config.DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	// Create services
	userRepo := repository.NewUserRepository(config.DB)
	userService := service.NewUserService(userRepo, cache)
	grpcServer := grpc.NewServer()
	userGrpcService := NewUserGrpcService(userService)
	proto.RegisterUserServiceServer(grpcServer, userGrpcService)

	// Start listening
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	log.Printf("gRPC server listening on port %s", port)
	return grpcServer.Serve(lis)
}
