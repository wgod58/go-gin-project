package model

import (
	"time"

	"github.com/stripe/stripe-go/v72"
)

// Payment represents the payment domain entity.
type Payment struct {
	ID            uint
	UserID        uint
	Amount        float64
	Currency      string
	StripeID      string
	PaymentStatus string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

// PaymentRepository defines persistence operations for payments.
// Implemented by infrastructure/repository, consumed by application.
type PaymentRepository interface {
	Create(payment *Payment) (*Payment, error)
	FindByStripeID(stripeID string) (*Payment, error)
	UpdateStatus(payment *Payment) (*Payment, error)
}

// StripeService defines external Stripe payment operations.
// Implemented by infrastructure/stripe, consumed by application.
type StripeService interface {
	New(params *stripe.PaymentIntentParams) (*stripe.PaymentIntent, error)
	Get(id string, params *stripe.PaymentIntentParams) (*stripe.PaymentIntent, error)
}
