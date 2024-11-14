package mocks

import (
	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go/v72"
)

// MockStripe implements the StripeService interface for testing
type MockStripe struct {
	mock.Mock
}

// New mocks the creation of a new payment intent
func (m *MockStripe) New(params *stripe.PaymentIntentParams) (*stripe.PaymentIntent, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.PaymentIntent), args.Error(1)
}

// Get mocks retrieving a payment intent
func (m *MockStripe) Get(id string, params *stripe.PaymentIntentParams) (*stripe.PaymentIntent, error) {
	args := m.Called(id, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripe.PaymentIntent), args.Error(1)
}

// Helper function to create a mock payment intent for testing
func CreateMockPaymentIntent(id string, amount int64, currency string, status stripe.PaymentIntentStatus) *stripe.PaymentIntent {
	return &stripe.PaymentIntent{
		ID:           id,
		Amount:       amount,
		Currency:     string(currency),
		Status:       status,
		ClientSecret: id + "_secret",
		Created:      1234567890,
	}
}
