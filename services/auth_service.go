package services

import (
	"errors"
	"go-gin-project/middleware"
	"go-gin-project/models"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	DB *gorm.DB
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"`
	UserID    uint   `json:"user_id"`
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{
		DB: db,
	}
}

func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	var user models.User
	if err := s.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &middleware.Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:     tokenString,
		ExpiresIn: expirationTime.Unix(),
		UserID:    user.ID,
	}, nil
}
