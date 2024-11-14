package services

import (
	"fmt"
	"go-gin-project/interfaces"
	"go-gin-project/models"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	DB    *gorm.DB
	Cache interfaces.CacheInterface
}

func NewUserService(db *gorm.DB, cache interfaces.CacheInterface) *UserService {
	return &UserService{
		DB:    db,
		Cache: cache,
	}
}

func (s *UserService) Create(user *models.User) (*models.User, error) {
	// Start transaction
	tx := s.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", tx.Error)
	}

	// Defer rollback in case of error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if user exists within transaction
	var existingUser models.User
	if err := tx.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		tx.Rollback()
		return nil, fmt.Errorf("user already exists")
	} else if err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return nil, fmt.Errorf("database error: %v", err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}
	user.Password = string(hashedPassword)

	// Create new user within transaction
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Remove password from response
	user.Password = ""
	return user, nil
}

func (s *UserService) Get(id string) (*models.User, error) {
	cacheKey := fmt.Sprintf("user:%s", id)

	// Try to get from cache first
	var user models.User
	err := s.Cache.GetCache(cacheKey, &user)
	if err == nil {
		return &user, nil
	}

	// If not in cache, get from database
	if err := s.DB.First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	fmt.Println("**************** get user ****************")
	fmt.Println(user)
	fmt.Println(err)
	// Store in cache for future requests
	s.Cache.SetCache(cacheKey, &user, time.Minute*5) // Changed from user to &user

	return &user, nil
}

func (s *UserService) Update(id string, userData *models.User) (*models.User, error) {
	// Start transaction
	tx := s.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", tx.Error)
	}

	// Defer rollback in case of error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var user models.User
	if err := tx.First(&user, id).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("user not found: %v", err)
	}

	// Update allowed fields
	user.Name = userData.Name
	user.Email = userData.Email

	// Handle password update if provided
	if userData.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to hash password: %v", err)
		}
		user.Password = string(hashedPassword)
	}

	// Save updates within transaction
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update user: %v", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Invalidate cache after successful update
	cacheKey := fmt.Sprintf("user:%s", id)
	s.Cache.DeleteCache(cacheKey)

	user.Password = "" // Remove password from response
	return &user, nil
}

func (s *UserService) Delete(id string) error {
	// Start transaction
	tx := s.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %v", tx.Error)
	}

	// Defer rollback in case of error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var user models.User
	if err := tx.First(&user, id).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("user not found: %v", err)
	}

	// Delete user within transaction
	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete user: %v", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Invalidate cache after successful deletion
	cacheKey := fmt.Sprintf("user:%s", id)
	s.Cache.DeleteCache(cacheKey)

	return nil
}
