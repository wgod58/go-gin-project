package models

import (
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        uint           `json:"user_id"`
	Amount        float64        `json:"amount" gorm:"type:decimal(10,2);not null"`
	Currency      string         `json:"currency" gorm:"type:varchar(3);not null"`
	StripeID      string         `json:"stripe_id" gorm:"type:varchar(255);not null"`
	PaymentStatus string         `json:"payment_status" gorm:"type:varchar(255);not null"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
	User          User           `json:"-" gorm:"foreignKey:UserID"`
}
