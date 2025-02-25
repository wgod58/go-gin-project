package server

import (
	"fmt"
	"log"
	"net"
	"os"

	"go-gin-project/config"
	"go-gin-project/interfaces"
	"go-gin-project/proto"
	"go-gin-project/services"

	"google.golang.org/grpc"
)

func StartGrpcServer(cache interfaces.CacheInterface) error {
	// Create services using the already initialized DB and cache
	if config.DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	// Create services
	userService := services.NewUserService(config.DB, cache)
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
