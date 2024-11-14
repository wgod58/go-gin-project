package services

import (
	"fmt"
	"go-gin-project/models"
	"time"

	"go-gin-project/interfaces"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	DB    interfaces.DBInterface
	Cache interfaces.CacheInterface
}

func NewUserService(db interfaces.DBInterface, cache interfaces.CacheInterface) *UserService {
	return &UserService{
		DB:    db,
		Cache: cache,
	}
}

func (s *UserService) Create(user *models.User) (*models.User, error) {
	// Check if user exists
	var existingUser models.User
	if err := s.DB.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("user already exists")
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("database error")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password")
	}
	user.Password = string(hashedPassword)

	// Create new user
	if err := s.DB.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user")
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
		return nil, fmt.Errorf("user not found")
	}

	// Store in cache for future requests
	s.Cache.SetCache(cacheKey, user, time.Minute*5) // Cache for 5 minutes

	return &user, nil
}

func (s *UserService) Update(id string, userData *models.User) (*models.User, error) {
	var user models.User
	if err := s.DB.First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update allowed fields
	user.Name = userData.Name
	user.Email = userData.Email

	// Handle password update if provided
	if userData.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password")
		}
		user.Password = string(hashedPassword)
	}

	if err := s.DB.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user")
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("user:%s", id)
	s.Cache.DeleteCache(cacheKey)

	user.Password = "" // Remove password from response
	return &user, nil
}

func (s *UserService) Delete(id string) error {
	var user models.User
	if err := s.DB.First(&user, id).Error; err != nil {
		return fmt.Errorf("user not found")
	}

	if err := s.DB.Delete(&user).Error; err != nil {
		return fmt.Errorf("failed to delete user")
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("user:%s", id)
	s.Cache.DeleteCache(cacheKey)

	return nil
}
