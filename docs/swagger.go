// Package docs Documentation
//
// Documentation for Go Payment Service API
//
//	Schemes: http
//	BasePath: /api
//	Version: 1.0.0
//	Title: Go Payment Service API
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Security:
//	- api_key:
//
// SecurityDefinitions:
//
//	api_key:
//	 type: apiKey
//	 name: Authorization
//	 in: header
//
// swagger:meta
package docs

// User represents the user model
// swagger:model User
type User struct {
	// The user ID
	// example: 1
	ID uint `json:"id"`

	// The user's name
	// example: John Doe
	// required: true
	Name string `json:"name"`

	// The user's email
	// example: john@example.com
	// required: true
	Email string `json:"email"`

	// The user's password (not returned in responses)
	// example: password123
	// required: true
	Password string `json:"password,omitempty"`
}

// Payment represents a payment record
// swagger:model Payment
type Payment struct {
	// The payment ID
	// example: 1
	ID uint `json:"id"`

	// The user ID associated with this payment
	// example: 1
	UserID uint `json:"user_id"`

	// The payment amount
	// example: 1000
	Amount float64 `json:"amount"`

	// The payment currency
	// example: usd
	Currency string `json:"currency"`

	// The Stripe payment intent ID
	// example: pi_1234567890
	StripeID string `json:"stripe_id"`

	// The payment status
	// example: succeeded
	PaymentStatus string `json:"payment_status"`
}

// CreatePaymentRequest represents a request to create a payment intent
// swagger:parameters createPaymentIntent
type CreatePaymentRequest struct {
	// The payment amount
	// required: true
	// example: 1000
	Amount float64 `json:"amount"`

	// The payment currency
	// required: true
	// example: usd
	Currency string `json:"currency"`

	// The user ID
	// required: true
	// example: 1
	UserID uint `json:"user_id"`
}

// RetrievePaymentRequest represents a request to retrieve a payment intent
// swagger:parameters retrievePaymentIntent
type RetrievePaymentRequest struct {
	// The Stripe payment intent ID
	// required: true
	// example: pi_1234567890
	PaymentIntentID string `json:"payment_intent_id"`
}
