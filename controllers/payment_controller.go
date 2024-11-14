package controllers

import (
	"errors"
	"fmt"
	"go-gin-project/services"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v72"
)

type PaymentController struct {
	paymentService *services.PaymentService
}

type CreatePaymentRequest struct {
	Amount   float64 `json:"amount" binding:"required,gt=0"`
	Currency string  `json:"currency" binding:"required"`
	UserID   uint    `json:"user_id" binding:"required"`
}

type RetrievePaymentRequest struct {
	PaymentIntentID string `json:"payment_intent_id" binding:"required"`
}

func NewPaymentController(paymentService *services.PaymentService) *PaymentController {
	return &PaymentController{
		paymentService: paymentService,
	}
}

func InitStripe() error {
	key := os.Getenv("STRIPE_SECRET_KEY")
	if key == "" {
		return fmt.Errorf("STRIPE_SECRET_KEY is not set in environment")
	}
	stripe.Key = key
	return nil
}

// CreatePaymentIntent godoc
// @Summary Create a payment intent
// @Description Create a new Stripe payment intent
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body CreatePaymentRequest true "Payment request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /payments/payment-intent [post]
func (p *PaymentController) CreatePaymentIntent(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, clientSecret, err := p.paymentService.CreatePaymentIntent(req.Amount, req.Currency, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"clientSecret": clientSecret,
		"payment":      payment,
	})
}

// RetrievePaymentIntent godoc
// @Summary Retrieve payment intent
// @Description Retrieve a Stripe payment intent status
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body RetrievePaymentRequest true "Payment intent request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /payments/retrieve [post]
func (p *PaymentController) RetrievePaymentIntent(c *gin.Context) {
	var req RetrievePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, pi, err := p.paymentService.RetrievePaymentIntent(req.PaymentIntentID)
	if err != nil {
		var stripeErr *stripe.Error
		if ok := errors.As(err, &stripeErr); ok {
			switch stripeErr.Code {
			case stripe.ErrorCodeResourceMissing:
				c.JSON(http.StatusNotFound, gin.H{"error": "Payment intent not found"})
			case stripe.ErrorCode(stripe.ErrorTypeAuthentication):
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Stripe API key"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": stripeErr.Msg})
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	response := gin.H{
		"payment": payment,
	}

	if pi != nil {
		response["payment_intent"] = gin.H{
			"id":            pi.ID,
			"status":        pi.Status,
			"amount":        float64(pi.Amount) / 100,
			"currency":      pi.Currency,
			"client_secret": pi.ClientSecret,
			"created":       pi.Created,
		}
	}

	c.JSON(http.StatusOK, response)
}
