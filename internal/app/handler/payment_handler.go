package handler

import (
	"errors"
	"net/http"

	"go-gin-project/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v72"
)

type PaymentHandler struct {
	service *service.PaymentService
}

type createPaymentRequest struct {
	Amount   float64 `json:"amount" binding:"required,gt=0"`
	Currency string  `json:"currency" binding:"required"`
	UserID   uint    `json:"user_id" binding:"required"`
}

type retrievePaymentRequest struct {
	PaymentIntentID string `json:"payment_intent_id" binding:"required"`
}

func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: svc}
}

// CreatePaymentIntent godoc
// @Summary Create a payment intent
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body createPaymentRequest true "Payment request"
// @Success 200 {object} map[string]interface{}
// @Router /payments/payment-intent [post]
func (h *PaymentHandler) CreatePaymentIntent(c *gin.Context) {
	var req createPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	payment, clientSecret, err := h.service.CreatePaymentIntent(req.Amount, req.Currency, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"clientSecret": clientSecret, "payment": payment})
}

// RetrievePaymentIntent godoc
// @Summary Retrieve payment intent
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body retrievePaymentRequest true "Payment intent request"
// @Success 200 {object} map[string]interface{}
// @Router /payments/retrieve [post]
func (h *PaymentHandler) RetrievePaymentIntent(c *gin.Context) {
	var req retrievePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	payment, pi, err := h.service.RetrievePaymentIntent(req.PaymentIntentID)
	if err != nil {
		var stripeErr *stripe.Error
		if errors.As(err, &stripeErr) {
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

	resp := gin.H{"payment": payment}
	if pi != nil {
		resp["payment_intent"] = gin.H{
			"id": pi.ID, "status": pi.Status,
			"amount": float64(pi.Amount) / 100, "currency": pi.Currency,
			"client_secret": pi.ClientSecret, "created": pi.Created,
		}
	}
	c.JSON(http.StatusOK, resp)
}
