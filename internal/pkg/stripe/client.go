package stripe

import (
	"fmt"
	"os"

	"go-gin-project/internal/pkg/model"

	stripelib "github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/paymentintent"
)

type client struct{}

// New creates a Stripe client and initializes the Stripe SDK key from STRIPE_SECRET_KEY env var.
func New() (model.StripeService, error) {
	key := os.Getenv("STRIPE_SECRET_KEY")
	if key == "" {
		return nil, fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}
	stripelib.Key = key
	return &client{}, nil
}

func (c *client) New(params *stripelib.PaymentIntentParams) (*stripelib.PaymentIntent, error) {
	return paymentintent.New(params)
}

func (c *client) Get(id string, params *stripelib.PaymentIntentParams) (*stripelib.PaymentIntent, error) {
	return paymentintent.Get(id, params)
}

var _ model.StripeService = (*client)(nil) // compile-time interface check
