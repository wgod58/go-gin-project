package routes

import (
	"go-gin-project/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRoutesWithControllers(r *gin.Engine, paymentController *controllers.PaymentController, userController *controllers.UserController) {
	// User routes
	users := r.Group("/api/users")
	{
		users.POST("/", userController.Create)
		users.GET("/:id", userController.Get)
		users.PUT("/:id", userController.Update)
		users.DELETE("/:id", userController.Delete)
	}

	// Payment routes
	payments := r.Group("/api/payments")
	{
		payments.POST("/payment-intent", paymentController.CreatePaymentIntent)
		payments.POST("/retrieve", paymentController.RetrievePaymentIntent)
	}
}
