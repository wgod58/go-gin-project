package mocks

import (
	"go-gin-project/internal/pkg/model"

	stripelib "github.com/stripe/stripe-go/v72"
	"github.com/stretchr/testify/mock"
)

// MockStripe implements model.StripeService for testing.
type MockStripe struct {
	mock.Mock
}

var _ model.StripeService = (*MockStripe)(nil)

func (m *MockStripe) New(params *stripelib.PaymentIntentParams) (*stripelib.PaymentIntent, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripelib.PaymentIntent), args.Error(1)
}

func (m *MockStripe) Get(id string, params *stripelib.PaymentIntentParams) (*stripelib.PaymentIntent, error) {
	args := m.Called(id, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*stripelib.PaymentIntent), args.Error(1)
}
