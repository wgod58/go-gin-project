package routes

import (
	"go-gin-project/controllers"
	"go-gin-project/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutesWithControllers(r *gin.Engine, paymentController *controllers.PaymentController, userController *controllers.UserController, authController *controllers.AuthController) {
	// Public routes
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", authController.Login)
		auth.POST("/admin-user", userController.Create)
	}

	// Protected routes
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		// User routes
		users := api.Group("/users")
		{
			users.POST("/", userController.Create)
			users.GET("/:id", userController.Get)
			users.PUT("/:id", userController.Update)
			users.DELETE("/:id", userController.Delete)
		}

		// Payment routes
		payments := api.Group("/payments")
		{
			payments.POST("/payment-intent", paymentController.CreatePaymentIntent)
			payments.POST("/retrieve", paymentController.RetrievePaymentIntent)
		}
	}
}
