package service

import (
	"fmt"
	"time"

	"go-gin-project/internal/pkg/model"

	"github.com/stripe/stripe-go/v72"
)

type PaymentService struct {
	paymentRepo model.PaymentRepository
	userRepo    model.UserRepository
	cache       model.CacheService
	stripe      model.StripeService
}

func NewPaymentService(
	paymentRepo model.PaymentRepository,
	userRepo model.UserRepository,
	cache model.CacheService,
	stripe model.StripeService,
) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		userRepo:    userRepo,
		cache:       cache,
		stripe:      stripe,
	}
}

func (s *PaymentService) CreatePaymentIntent(amount float64, currency string, userID uint) (*model.Payment, string, error) {
	if _, err := s.userRepo.FindByID(fmt.Sprintf("%d", userID)); err != nil {
		return nil, "", fmt.Errorf("create payment intent: invalid user: %w", err)
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(amount * 100)),
		Currency: stripe.String(currency),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}
	pi, err := s.stripe.New(params)
	if err != nil {
		return nil, "", fmt.Errorf("create payment intent: stripe: %w", err)
	}

	payment := &model.Payment{
		UserID:        userID,
		Amount:        amount,
		Currency:      currency,
		StripeID:      pi.ID,
		PaymentStatus: string(pi.Status),
	}
	saved, err := s.paymentRepo.Create(payment)
	if err != nil {
		return nil, "", fmt.Errorf("create payment intent: save: %w", err)
	}

	return saved, pi.ClientSecret, nil
}

func (s *PaymentService) RetrievePaymentIntent(paymentIntentID string) (*model.Payment, *stripe.PaymentIntent, error) {
	cacheKey := fmt.Sprintf("payment:%s", paymentIntentID)

	var cached model.Payment
	if err := s.cache.Get(cacheKey, &cached); err == nil {
		return &cached, nil, nil
	}

	pi, err := s.stripe.Get(paymentIntentID, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("retrieve payment: stripe: %w", err)
	}

	payment, err := s.paymentRepo.FindByStripeID(pi.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("retrieve payment: not found: %w", err)
	}

	payment.PaymentStatus = string(pi.Status)
	updated, err := s.paymentRepo.UpdateStatus(payment)
	if err != nil {
		return nil, nil, fmt.Errorf("retrieve payment: update status: %w", err)
	}

	s.cache.Set(cacheKey, updated, 5*time.Minute) //nolint:errcheck
	return updated, pi, nil
}
