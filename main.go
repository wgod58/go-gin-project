package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-gin-project/config"
	apppkg "go-gin-project/internal/app"
	"go-gin-project/internal/app/handler"
	"go-gin-project/internal/app/service"
	"go-gin-project/internal/pkg/cache"
	"go-gin-project/internal/pkg/repository"
	stripepkg "go-gin-project/internal/pkg/stripe"
	grpcserver "go-gin-project/grpc/server"

	_ "go-gin-project/docs"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	if err := config.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer config.CloseDB() //nolint:errcheck

	// Infrastructure layer
	cacheService, err := cache.New()
	if err != nil {
		log.Printf("Warning: Redis unavailable, continuing without cache: %v", err)
	}

	stripeClient, err := stripepkg.New()
	if err != nil {
		log.Printf("Warning: Stripe unavailable: %v", err)
	}

	userRepo := repository.NewUserRepository(config.DB)
	paymentRepo := repository.NewPaymentRepository(config.DB)

	// Application layer
	userService := service.NewUserService(userRepo, cacheService)
	authService := service.NewAuthService(userRepo)
	paymentService := service.NewPaymentService(paymentRepo, userRepo, cacheService, stripeClient)

	// Transport layer
	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(authService)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	apppkg.SetupRoutes(r, userHandler, authHandler, paymentHandler)

	srv := &http.Server{Addr: ":8080", Handler: r}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	go func() {
		if err := grpcserver.StartGrpcServer(cacheService); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Servers exited properly")
}
