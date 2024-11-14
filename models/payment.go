package models

import (
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        uint           `json:"user_id"`
	Amount        float64        `json:"amount"`
	Currency      string         `json:"currency"`
	StripeID      string         `json:"stripe_id"`
	PaymentStatus string         `json:"payment_status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
	User          User           `json:"-" gorm:"foreignKey:UserID"`
}
