package main

import (
	"go-gin-project/config"
	"go-gin-project/controllers"
	"go-gin-project/routes"
	"go-gin-project/services"
	"log"

	_ "go-gin-project/docs" // This is required for swagger

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/paymentintent"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// DefaultStripeService implements the StripeService interface using the default Stripe client
type DefaultStripeService struct{}

func (s *DefaultStripeService) New(params *stripe.PaymentIntentParams) (*stripe.PaymentIntent, error) {
	return paymentintent.New(params)
}

func (s *DefaultStripeService) Get(id string, params *stripe.PaymentIntentParams) (*stripe.PaymentIntent, error) {
	return paymentintent.Get(id, params)
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Initialize Stripe
	if err := controllers.InitStripe(); err != nil {
		log.Printf("Warning: Failed to initialize Stripe: %v", err)
		log.Println("Payment features may not work properly")
	}

	// Initialize database
	if err := config.InitDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer func() {
		if err := config.CloseDB(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// Initialize Redis cache service
	cacheService, err := services.NewRedisCache()
	if err != nil {
		log.Printf("Warning: Failed to initialize Redis: %v", err)
		log.Println("Application will continue without caching")
	}

	// Create Gin router
	r := gin.Default()

	// Swagger documentation endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Initialize services
	stripeService := &DefaultStripeService{}
	userService := services.NewUserService(config.DB, cacheService)
	paymentService := services.NewPaymentService(config.DB, cacheService, stripeService)
	authService := services.NewAuthService(config.DB)

	// Initialize controllers
	userController := controllers.NewUserController(userService)
	paymentController := controllers.NewPaymentController(paymentService)
	authController := controllers.NewAuthController(authService)

	// Setup routes with the initialized controllers
	routes.SetupRoutesWithControllers(r, paymentController, userController, authController)

	// Run the server
	r.Run(":8080")
}
