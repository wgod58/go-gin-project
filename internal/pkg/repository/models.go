package repository

import (
	"time"

	"go-gin-project/internal/pkg/model"

	"gorm.io/gorm"
)

// userModel is the GORM persistence model for User.
type userModel struct {
	ID        uint           `gorm:"primaryKey"`
	Name      string         `gorm:"type:varchar(255);not null"`
	Email     string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	Password  string         `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (userModel) TableName() string { return "users" }

func toUserDomain(m *userModel) *model.User {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}
	return &model.User{
		ID:        m.ID,
		Name:      m.Name,
		Email:     m.Email,
		Password:  m.Password,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

func toUserModel(u *model.User) *userModel {
	return &userModel{
		ID:       u.ID,
		Name:     u.Name,
		Email:    u.Email,
		Password: u.Password,
	}
}

// paymentModel is the GORM persistence model for Payment.
type paymentModel struct {
	ID            uint           `gorm:"primaryKey"`
	UserID        uint
	Amount        float64        `gorm:"type:decimal(10,2);not null"`
	Currency      string         `gorm:"type:varchar(3);not null"`
	StripeID      string         `gorm:"type:varchar(255);not null"`
	PaymentStatus string         `gorm:"type:varchar(255);not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (paymentModel) TableName() string { return "payments" }

func toPaymentDomain(m *paymentModel) *model.Payment {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}
	return &model.Payment{
		ID:            m.ID,
		UserID:        m.UserID,
		Amount:        m.Amount,
		Currency:      m.Currency,
		StripeID:      m.StripeID,
		PaymentStatus: m.PaymentStatus,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
		DeletedAt:     deletedAt,
	}
}

func toPaymentModel(p *model.Payment) *paymentModel {
	return &paymentModel{
		ID:            p.ID,
		UserID:        p.UserID,
		Amount:        p.Amount,
		Currency:      p.Currency,
		StripeID:      p.StripeID,
		PaymentStatus: p.PaymentStatus,
	}
}

// Migrate runs GORM AutoMigrate for all repository models.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&userModel{}, &paymentModel{})
}
