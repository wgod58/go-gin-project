package domain

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"type:varchar(255);not null"`
	Email     string         `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	Password  string         `json:"password,omitempty" gorm:"type:varchar(255);not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// UserRepository defines persistence operations for users.
// Implemented by infrastructure/mysql, consumed by application.
type UserRepository interface {
	Create(user *User) (*User, error)
	FindByID(id string) (*User, error)
	FindByEmail(email string) (*User, error)
	Update(id string, data *User) (*User, error)
	Delete(id string) error
}
