package repository

import (
	"fmt"

	"go-gin-project/internal/domain"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a GORM-backed domain.UserRepository.
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *domain.User) (*domain.User, error) {
	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("create user: begin transaction: %w", tx.Error)
	}
	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
		}
	}()

	var existing userModel
	if err := tx.Where("email = ?", user.Email).First(&existing).Error; err == nil {
		tx.Rollback()
		return nil, fmt.Errorf("create user: user already exists")
	} else if err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return nil, fmt.Errorf("create user: %w", err)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("create user: hash password: %w", err)
	}

	m := toUserModel(user)
	m.Password = string(hashed)

	if err := tx.Create(m).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("create user: insert: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("create user: commit: %w", err)
	}

	result := toUserDomain(m)
	result.Password = ""
	return result, nil
}

func (r *userRepository) FindByID(id string) (*domain.User, error) {
	var m userModel
	if err := r.db.First(&m, id).Error; err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	return toUserDomain(&m), nil
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var m userModel
	if err := r.db.Where("email = ?", email).First(&m).Error; err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return toUserDomain(&m), nil
}

func (r *userRepository) Update(id string, data *domain.User) (*domain.User, error) {
	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("update user: begin transaction: %w", tx.Error)
	}
	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
		}
	}()

	var m userModel
	if err := tx.First(&m, id).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("update user: not found: %w", err)
	}

	m.Name = data.Name
	m.Email = data.Email

	if data.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("update user: hash password: %w", err)
		}
		m.Password = string(hashed)
	}

	if err := tx.Save(&m).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("update user: save: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("update user: commit: %w", err)
	}

	result := toUserDomain(&m)
	result.Password = ""
	return result, nil
}

func (r *userRepository) Delete(id string) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("delete user: begin transaction: %w", tx.Error)
	}
	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
		}
	}()

	var m userModel
	if err := tx.First(&m, id).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("delete user: not found: %w", err)
	}
	if err := tx.Delete(&m).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("delete user: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("delete user: commit: %w", err)
	}
	return nil
}
