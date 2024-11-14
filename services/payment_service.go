package services

import (
	"fmt"
	"go-gin-project/models"
	"time"

	"github.com/stripe/stripe-go/v72"
	"gorm.io/gorm"
)

type PaymentService struct {
	DB     *gorm.DB
	Cache  *RedisCache
	Stripe StripeService
}

type StripeService interface {
	New(*stripe.PaymentIntentParams) (*stripe.PaymentIntent, error)
	Get(string, *stripe.PaymentIntentParams) (*stripe.PaymentIntent, error)
}

func NewPaymentService(db *gorm.DB, cache *RedisCache, stripe StripeService) *PaymentService {
	return &PaymentService{
		DB:     db,
		Cache:  cache,
		Stripe: stripe,
	}
}

func (s *PaymentService) CreatePaymentIntent(amount float64, currency string, userID uint) (*models.Payment, string, error) {
	// Verify user exists
	var user models.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return nil, "", fmt.Errorf("invalid user ID")
	}

	// Create Stripe PaymentIntent
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(amount * 100)), // Convert to cents
		Currency: stripe.String(currency),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	pi, err := s.Stripe.New(params)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create payment intent: %v", err)
	}

	// Save payment record to database
	payment := &models.Payment{
		UserID:        userID,
		Amount:        amount,
		Currency:      currency,
		StripeID:      pi.ID,
		PaymentStatus: string(pi.Status),
	}

	if err := s.DB.Create(payment).Error; err != nil {
		return nil, "", fmt.Errorf("failed to save payment record")
	}

	return payment, pi.ClientSecret, nil
}

func (s *PaymentService) RetrievePaymentIntent(paymentIntentID string) (*models.Payment, *stripe.PaymentIntent, error) {
	cacheKey := fmt.Sprintf("payment:%s", paymentIntentID)

	// Try to get from cache first
	var cachedPayment models.Payment
	if err := s.Cache.GetCache(cacheKey, &cachedPayment); err == nil {
		return &cachedPayment, nil, nil
	}

	// If not in cache, get from Stripe
	pi, err := s.Stripe.Get(paymentIntentID, nil)
	if err != nil {
		return nil, nil, err
	}

	// Update payment status in database
	var payment models.Payment
	result := s.DB.Where("stripe_id = ?", pi.ID).First(&payment)
	if result.Error != nil {
		return nil, nil, fmt.Errorf("payment record not found")
	}

	// Update payment status
	payment.PaymentStatus = string(pi.Status)
	if err := s.DB.Save(&payment).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to update payment status")
	}

	// Store in cache for future requests
	s.Cache.SetCache(cacheKey, payment, time.Minute*5)

	return &payment, pi, nil
}
