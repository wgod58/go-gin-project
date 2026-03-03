package domain

import "time"

// User represents the user domain entity.
type User struct {
	ID        uint
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// UserRepository defines persistence operations for users.
// Implemented by infrastructure/repository, consumed by application.
type UserRepository interface {
	Create(user *User) (*User, error)
	FindByID(id string) (*User, error)
	FindByEmail(email string) (*User, error)
	Update(id string, data *User) (*User, error)
	Delete(id string) error
}
