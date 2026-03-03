package repository

import (
	"fmt"

	"go-gin-project/internal/pkg/model"

	"gorm.io/gorm"
)

type paymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository creates a GORM-backed model.PaymentRepository.
func NewPaymentRepository(db *gorm.DB) model.PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(payment *model.Payment) (*model.Payment, error) {
	m := toPaymentModel(payment)
	if err := r.db.Create(m).Error; err != nil {
		return nil, fmt.Errorf("create payment: %w", err)
	}
	return toPaymentDomain(m), nil
}

func (r *paymentRepository) FindByStripeID(stripeID string) (*model.Payment, error) {
	var m paymentModel
	if err := r.db.Where("stripe_id = ?", stripeID).First(&m).Error; err != nil {
		return nil, fmt.Errorf("find payment: %w", err)
	}
	return toPaymentDomain(&m), nil
}

func (r *paymentRepository) UpdateStatus(payment *model.Payment) (*model.Payment, error) {
	var m paymentModel
	if err := r.db.Where("stripe_id = ?", payment.StripeID).First(&m).Error; err != nil {
		return nil, fmt.Errorf("update payment status: not found: %w", err)
	}
	m.PaymentStatus = payment.PaymentStatus
	if err := r.db.Save(&m).Error; err != nil {
		return nil, fmt.Errorf("update payment status: save: %w", err)
	}
	return toPaymentDomain(&m), nil
}
