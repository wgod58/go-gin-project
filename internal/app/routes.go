package app

import (
	"go-gin-project/internal/app/handler"
	"go-gin-project/internal/app/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	userHandler *handler.UserHandler,
	authHandler *handler.AuthHandler,
	paymentHandler *handler.PaymentHandler,
) {
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/admin-user", userHandler.Create)
	}

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		users := api.Group("/users")
		{
			users.POST("/", userHandler.Create)
			users.GET("/:id", userHandler.Get)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
		}

		payments := api.Group("/payments")
		{
			payments.POST("/payment-intent", paymentHandler.CreatePaymentIntent)
			payments.POST("/retrieve", paymentHandler.RetrievePaymentIntent)
		}
	}
}
