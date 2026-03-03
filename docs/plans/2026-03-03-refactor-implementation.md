# Refactor: Go Standard Layout + DDD Clean Architecture — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Restructure the project into DDD flat-internal layers and add a clean-code rule.

**Architecture:** `transport → application → domain ← infrastructure`. Domain defines entities and interfaces. Infrastructure implements them. Application contains business logic. Transport handles HTTP/gRPC. `main.go` wires everything.

**Tech Stack:** Go 1.26, Gin, GORM (MySQL), Redis, Stripe, JWT, gRPC, go-sqlmock, testify/mock

---

## Overview of changes

```
OLD                          →  NEW
─────────────────────────────────────────
models/                      →  internal/domain/
services/user_service.go     →  internal/application/user_service.go
services/auth_service.go     →  internal/application/auth_service.go
services/payment_service.go  →  internal/application/payment_service.go
services/redis_service.go    →  internal/infrastructure/redis/cache.go
controllers/                 →  internal/transport/http/handler/
middleware/auth.go            →  internal/transport/http/middleware/auth.go
routes/routes.go             →  internal/transport/http/routes.go
grpc/                        →  internal/transport/grpc/
proto/                       →  api/proto/
tests/                       →  test/
interfaces/                  →  deleted (interfaces move to domain/)
                             →  NEW: internal/infrastructure/mysql/ (repos)
                             →  NEW: internal/infrastructure/stripe/client.go
```

Import path prefix stays `go-gin-project/` throughout.

---

## Task 1: Add clean-code rule

**Files:**
- Create: `.claude/rules/clean-code.md`

**Step 1: Create the rule file**

```markdown
---
description: Clean code standards combining Go idioms and general principles
globs: ["**/*.go"]
---

# Clean Code Standards

## Naming

- Package names: lowercase single word, no underscores (`userservice` → use `application` or `user` sub-package instead)
- Avoid redundant names: `user.UserService` → `user.Service`, `payment.PaymentRepo` → `payment.Repository`
- Exported types: PascalCase. Unexported: camelCase
- Acronyms: consistent casing — `userID` not `userId`, `httpServer` not `HTTPServer` in local vars
- Boolean vars/funcs: use `is`/`has`/`can` prefix: `isValid`, `hasPermission`

## Functions

- Single responsibility: one function does one thing
- Prefer ≤20 lines; if longer, extract helpers
- No boolean flag parameters: `process(user, true)` → two named functions
- Early returns over nested if-else
- Named return values only when they add clarity to complex returns

## Error Handling

- Always handle errors — never discard with `_` unless intentionally ignoring
- Wrap with context: `fmt.Errorf("create user: %w", err)` not `fmt.Errorf("error: %v", err)`
- No sentinel errors exported from packages — callers use `errors.Is` / `errors.As`
- Return errors, don't panic (except in `main` or truly unrecoverable situations)

## Packages

- Organize by responsibility, not by type (no `utils/`, `helpers/`, `common/`)
- Interfaces belong in the consuming package, not the implementing package
- Avoid circular dependencies — use the layer rule: `transport → application → domain ← infrastructure`
- Keep package surface small: prefer unexported types, export only what callers need

## Structs & Interfaces

- Small interfaces (1–3 methods): prefer `io.Reader` style over large interface blobs
- No god structs with 10+ fields doing multiple things
- Constructor functions (`NewXxx`) for all exported types that need initialization
- Don't export struct fields of internal types — use getters/setters or keep unexported

## Testing

- Table-driven tests for multiple cases: `tests := []struct{ name, input, want }{ ... }`
- Mock via interfaces — never mock concrete types
- No global state in tests — each test sets up its own dependencies
- Test file: `xxx_test.go` in same package (white-box) or `xxx_test` package (black-box)
- Each test must be independent and runnable in isolation

## Dependency Direction

```
transport (http/grpc handlers, middleware, routes)
    ↓ imports
application (services, use cases)
    ↓ imports
domain (entities, repository interfaces, service interfaces)
    ↑ implements
infrastructure (mysql repos, redis cache, stripe client)
```

- `domain` imports nothing from other internal layers
- `infrastructure` imports `domain` only (to implement interfaces)
- `application` imports `domain` only (uses interfaces, not concrete types)
- `transport` imports `application` only (calls service methods)
- `main.go` is the only file that imports all layers (for wiring)

## Tooling

Run before every commit:
- `gofmt -w .` — format all files
- `go vet ./...` — catch common mistakes
- `staticcheck ./...` — advanced static analysis
- `go test ./... -v` — all tests must pass
```

**Step 2: Verify file created**

```bash
cat .claude/rules/clean-code.md
```
Expected: file contents printed

**Step 3: Commit**

```bash
git add .claude/rules/clean-code.md
git commit -m "feat: add clean-code rule for all Go files"
```

---

## Task 2: Create domain layer

**Files:**
- Create: `internal/domain/user.go`
- Create: `internal/domain/payment.go`
- Create: `internal/domain/cache.go`

**Step 1: Create `internal/domain/user.go`**

```go
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
```

**Step 2: Create `internal/domain/payment.go`**

```go
package domain

import (
	"time"

	"github.com/stripe/stripe-go/v72"
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

// PaymentRepository defines persistence operations for payments.
type PaymentRepository interface {
	Create(payment *Payment) (*Payment, error)
	FindByStripeID(stripeID string) (*Payment, error)
	UpdateStatus(payment *Payment) (*Payment, error)
}

// StripeService defines the external Stripe payment operations.
// Implemented by infrastructure/stripe, consumed by application.
type StripeService interface {
	New(params *stripe.PaymentIntentParams) (*stripe.PaymentIntent, error)
	Get(id string, params *stripe.PaymentIntentParams) (*stripe.PaymentIntent, error)
}
```

**Step 3: Create `internal/domain/cache.go`**

```go
package domain

import "time"

// CacheService defines cache operations.
// Implemented by infrastructure/redis, consumed by application.
type CacheService interface {
	Get(key string, dest interface{}) error
	Set(key string, value interface{}, expiration time.Duration) error
	Delete(key string) error
}
```

**Step 4: Verify it compiles**

```bash
go build ./internal/domain/...
```
Expected: no output (success)

**Step 5: Commit**

```bash
git add internal/domain/
git commit -m "feat: add domain layer with entities and repository interfaces"
```

---

## Task 3: Create infrastructure/mysql repositories

**Files:**
- Create: `internal/infrastructure/mysql/user_repository.go`
- Create: `internal/infrastructure/mysql/payment_repository.go`

**Step 1: Create `internal/infrastructure/mysql/user_repository.go`**

```go
package mysql

import (
	"fmt"

	"go-gin-project/internal/domain"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a MySQL-backed UserRepository.
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

	var existing domain.User
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
	user.Password = string(hashed)

	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("create user: insert: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("create user: commit: %w", err)
	}

	user.Password = ""
	return user, nil
}

func (r *userRepository) FindByID(id string) (*domain.User, error) {
	var user domain.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &user, nil
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

	var user domain.User
	if err := tx.First(&user, id).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("update user: not found: %w", err)
	}

	user.Name = data.Name
	user.Email = data.Email

	if data.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("update user: hash password: %w", err)
		}
		user.Password = string(hashed)
	}

	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("update user: save: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("update user: commit: %w", err)
	}

	user.Password = ""
	return &user, nil
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

	var user domain.User
	if err := tx.First(&user, id).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("delete user: not found: %w", err)
	}
	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("delete user: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("delete user: commit: %w", err)
	}
	return nil
}
```

**Step 2: Create `internal/infrastructure/mysql/payment_repository.go`**

```go
package mysql

import (
	"fmt"

	"go-gin-project/internal/domain"

	"gorm.io/gorm"
)

type paymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository creates a MySQL-backed PaymentRepository.
func NewPaymentRepository(db *gorm.DB) domain.PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(payment *domain.Payment) (*domain.Payment, error) {
	if err := r.db.Create(payment).Error; err != nil {
		return nil, fmt.Errorf("create payment: %w", err)
	}
	return payment, nil
}

func (r *paymentRepository) FindByStripeID(stripeID string) (*domain.Payment, error) {
	var payment domain.Payment
	if err := r.db.Where("stripe_id = ?", stripeID).First(&payment).Error; err != nil {
		return nil, fmt.Errorf("find payment: %w", err)
	}
	return &payment, nil
}

func (r *paymentRepository) UpdateStatus(payment *domain.Payment) (*domain.Payment, error) {
	if err := r.db.Save(payment).Error; err != nil {
		return nil, fmt.Errorf("update payment status: %w", err)
	}
	return payment, nil
}
```

**Step 3: Verify compile**

```bash
go build ./internal/infrastructure/mysql/...
```
Expected: no output

**Step 4: Commit**

```bash
git add internal/infrastructure/mysql/
git commit -m "feat: add MySQL repository implementations"
```

---

## Task 4: Create infrastructure/redis cache

**Files:**
- Create: `internal/infrastructure/redis/cache.go`

**Step 1: Create `internal/infrastructure/redis/cache.go`**

```go
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go-gin-project/internal/domain"

	"github.com/redis/go-redis/v9"
)

type cache struct {
	client *redis.Client
	ctx    context.Context
}

// NewCache creates a Redis-backed CacheService.
func NewCache() (domain.CacheService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connect: %w", err)
	}

	return &cache{client: client, ctx: ctx}, nil
}

func (c *cache) Get(key string, dest interface{}) error {
	val, err := c.client.Get(c.ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

func (c *cache) Set(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache set marshal: %w", err)
	}
	return c.client.Set(c.ctx, key, data, expiration).Err()
}

func (c *cache) Delete(key string) error {
	return c.client.Del(c.ctx, key).Err()
}

var _ domain.CacheService = (*cache)(nil) // compile-time interface check
```

**Step 2: Verify compile**

```bash
go build ./internal/infrastructure/redis/...
```
Expected: no output

**Step 3: Commit**

```bash
git add internal/infrastructure/redis/
git commit -m "feat: add Redis cache infrastructure implementation"
```

---

## Task 5: Create infrastructure/stripe client

**Files:**
- Create: `internal/infrastructure/stripe/client.go`

**Step 1: Create `internal/infrastructure/stripe/client.go`**

```go
package stripe

import (
	"fmt"
	"os"

	"go-gin-project/internal/domain"

	stripelib "github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/paymentintent"
)

type client struct{}

// NewClient creates a Stripe StripeService and initializes the Stripe SDK key.
func NewClient() (domain.StripeService, error) {
	key := os.Getenv("STRIPE_SECRET_KEY")
	if key == "" {
		return nil, fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}
	stripelib.Key = key
	return &client{}, nil
}

func (c *client) New(params *stripelib.PaymentIntentParams) (*stripelib.PaymentIntent, error) {
	return paymentintent.New(params)
}

func (c *client) Get(id string, params *stripelib.PaymentIntentParams) (*stripelib.PaymentIntent, error) {
	return paymentintent.Get(id, params)
}

var _ domain.StripeService = (*client)(nil) // compile-time interface check
```

**Step 2: Verify compile**

```bash
go build ./internal/infrastructure/stripe/...
```
Expected: no output

**Step 3: Commit**

```bash
git add internal/infrastructure/stripe/
git commit -m "feat: add Stripe infrastructure client"
```

---

## Task 6: Create application layer (services)

Services now depend on domain interfaces — no direct GORM, Redis, or Stripe imports.

**Files:**
- Create: `internal/application/user_service.go`
- Create: `internal/application/auth_service.go`
- Create: `internal/application/payment_service.go`

**Step 1: Create `internal/application/user_service.go`**

```go
package application

import (
	"fmt"
	"time"

	"go-gin-project/internal/domain"
)

type UserService struct {
	repo  domain.UserRepository
	cache domain.CacheService
}

func NewUserService(repo domain.UserRepository, cache domain.CacheService) *UserService {
	return &UserService{repo: repo, cache: cache}
}

func (s *UserService) Create(user *domain.User) (*domain.User, error) {
	return s.repo.Create(user)
}

func (s *UserService) Get(id string) (*domain.User, error) {
	cacheKey := fmt.Sprintf("user:%s", id)

	var user domain.User
	if err := s.cache.Get(cacheKey, &user); err == nil {
		return &user, nil
	}

	found, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	s.cache.Set(cacheKey, found, 5*time.Minute) //nolint:errcheck
	return found, nil
}

func (s *UserService) Update(id string, data *domain.User) (*domain.User, error) {
	updated, err := s.repo.Update(id, data)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	s.cache.Delete(fmt.Sprintf("user:%s", id)) //nolint:errcheck
	return updated, nil
}

func (s *UserService) Delete(id string) error {
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	s.cache.Delete(fmt.Sprintf("user:%s", id)) //nolint:errcheck
	return nil
}
```

**Step 2: Create `internal/application/auth_service.go`**

Note: `Claims` moves here from `middleware` to break the cross-layer dependency.

```go
package application

import (
	"errors"
	"os"
	"time"

	"go-gin-project/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Claims is the JWT payload. Defined here so transport/middleware can import it.
type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
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

type AuthService struct {
	userRepo domain.UserRepository
}

func NewAuthService(userRepo domain.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	expiry := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:     tokenStr,
		ExpiresIn: expiry.Unix(),
		UserID:    user.ID,
	}, nil
}
```

**Step 3: Create `internal/application/payment_service.go`**

```go
package application

import (
	"fmt"
	"time"

	"go-gin-project/internal/domain"

	"github.com/stripe/stripe-go/v72"
)

type PaymentService struct {
	paymentRepo domain.PaymentRepository
	userRepo    domain.UserRepository
	cache       domain.CacheService
	stripe      domain.StripeService
}

func NewPaymentService(
	paymentRepo domain.PaymentRepository,
	userRepo domain.UserRepository,
	cache domain.CacheService,
	stripe domain.StripeService,
) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		userRepo:    userRepo,
		cache:       cache,
		stripe:      stripe,
	}
}

func (s *PaymentService) CreatePaymentIntent(amount float64, currency string, userID uint) (*domain.Payment, string, error) {
	if _, err := s.userRepo.FindByID(fmt.Sprintf("%d", userID)); err != nil {
		return nil, "", fmt.Errorf("create payment intent: invalid user: %w", err)
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(amount * 100)),
		Currency: stripe.String(currency),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}
	pi, err := s.stripe.New(params)
	if err != nil {
		return nil, "", fmt.Errorf("create payment intent: stripe: %w", err)
	}

	payment := &domain.Payment{
		UserID:        userID,
		Amount:        amount,
		Currency:      currency,
		StripeID:      pi.ID,
		PaymentStatus: string(pi.Status),
	}
	saved, err := s.paymentRepo.Create(payment)
	if err != nil {
		return nil, "", fmt.Errorf("create payment intent: save: %w", err)
	}

	return saved, pi.ClientSecret, nil
}

func (s *PaymentService) RetrievePaymentIntent(paymentIntentID string) (*domain.Payment, *stripe.PaymentIntent, error) {
	cacheKey := fmt.Sprintf("payment:%s", paymentIntentID)

	var cached domain.Payment
	if err := s.cache.Get(cacheKey, &cached); err == nil {
		return &cached, nil, nil
	}

	pi, err := s.stripe.Get(paymentIntentID, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("retrieve payment: stripe: %w", err)
	}

	payment, err := s.paymentRepo.FindByStripeID(pi.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("retrieve payment: not found: %w", err)
	}

	payment.PaymentStatus = string(pi.Status)
	updated, err := s.paymentRepo.UpdateStatus(payment)
	if err != nil {
		return nil, nil, fmt.Errorf("retrieve payment: update status: %w", err)
	}

	s.cache.Set(cacheKey, updated, 5*time.Minute) //nolint:errcheck
	return updated, pi, nil
}
```

**Step 4: Verify compile**

```bash
go build ./internal/application/...
```
Expected: no output

**Step 5: Commit**

```bash
git add internal/application/
git commit -m "feat: add application layer with services using domain interfaces"
```

---

## Task 7: Create transport/http layer

**Files:**
- Create: `internal/transport/http/middleware/auth.go`
- Create: `internal/transport/http/handler/user_handler.go`
- Create: `internal/transport/http/handler/auth_handler.go`
- Create: `internal/transport/http/handler/payment_handler.go`
- Create: `internal/transport/http/routes.go`

**Step 1: Create `internal/transport/http/middleware/auth.go`**

Note: imports `application.Claims` — transport imports application, not the other way.

```go
package middleware

import (
	"net/http"
	"os"
	"strings"

	"go-gin-project/internal/application"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		claims := &application.Claims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}
```

**Step 2: Create `internal/transport/http/handler/user_handler.go`**

```go
package handler

import (
	"net/http"

	"go-gin-project/internal/application"
	"go-gin-project/internal/domain"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service *application.UserService
}

func NewUserHandler(service *application.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// @title Go Payment Service API
// @version 1.0
// @description Payment service with Stripe integration
// @host localhost:8080
// @BasePath /api

// Create godoc
// @Summary Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Param user body domain.User true "User object"
// @Success 201 {object} domain.User
// @Failure 400 {object} map[string]string
// @Router /users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.service.Create(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": created})
}

// Get godoc
// @Summary Get user by ID
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} domain.User
// @Failure 404 {object} map[string]string
// @Router /users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
	user, err := h.service.Get(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// Update godoc
// @Summary Update user
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body domain.User true "User object"
// @Success 200 {object} domain.User
// @Router /users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	var data domain.User
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.service.Update(c.Param("id"), &data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": updated})
}

// Delete godoc
// @Summary Delete user
// @Tags users
// @Param id path int true "User ID"
// @Success 204
// @Router /users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
```

**Step 3: Create `internal/transport/http/handler/auth_handler.go`**

```go
package handler

import (
	"net/http"

	"go-gin-project/internal/application"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service *application.AuthService
}

func NewAuthHandler(service *application.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Login godoc
// @Summary User login
// @Tags auth
// @Accept json
// @Produce json
// @Param login body application.LoginRequest true "Login credentials"
// @Success 200 {object} application.LoginResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req application.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.service.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
```

**Step 4: Create `internal/transport/http/handler/payment_handler.go`**

```go
package handler

import (
	"errors"
	"net/http"

	"go-gin-project/internal/application"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v72"
)

type PaymentHandler struct {
	service *application.PaymentService
}

type createPaymentRequest struct {
	Amount   float64 `json:"amount" binding:"required,gt=0"`
	Currency string  `json:"currency" binding:"required"`
	UserID   uint    `json:"user_id" binding:"required"`
}

type retrievePaymentRequest struct {
	PaymentIntentID string `json:"payment_intent_id" binding:"required"`
}

func NewPaymentHandler(service *application.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

// CreatePaymentIntent godoc
// @Summary Create a payment intent
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body createPaymentRequest true "Payment request"
// @Success 200 {object} map[string]interface{}
// @Router /payments/payment-intent [post]
func (h *PaymentHandler) CreatePaymentIntent(c *gin.Context) {
	var req createPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	payment, clientSecret, err := h.service.CreatePaymentIntent(req.Amount, req.Currency, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"clientSecret": clientSecret, "payment": payment})
}

// RetrievePaymentIntent godoc
// @Summary Retrieve payment intent
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body retrievePaymentRequest true "Payment intent request"
// @Success 200 {object} map[string]interface{}
// @Router /payments/retrieve [post]
func (h *PaymentHandler) RetrievePaymentIntent(c *gin.Context) {
	var req retrievePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	payment, pi, err := h.service.RetrievePaymentIntent(req.PaymentIntentID)
	if err != nil {
		var stripeErr *stripe.Error
		if errors.As(err, &stripeErr) {
			switch stripeErr.Code {
			case stripe.ErrorCodeResourceMissing:
				c.JSON(http.StatusNotFound, gin.H{"error": "Payment intent not found"})
			case stripe.ErrorCode(stripe.ErrorTypeAuthentication):
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Stripe API key"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": stripeErr.Msg})
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	resp := gin.H{"payment": payment}
	if pi != nil {
		resp["payment_intent"] = gin.H{
			"id": pi.ID, "status": pi.Status,
			"amount": float64(pi.Amount) / 100, "currency": pi.Currency,
			"client_secret": pi.ClientSecret, "created": pi.Created,
		}
	}
	c.JSON(http.StatusOK, resp)
}
```

**Step 5: Create `internal/transport/http/routes.go`**

```go
package http

import (
	"go-gin-project/internal/transport/http/handler"
	"go-gin-project/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	userHandler *handler.UserHandler,
	authHandler *handler.AuthHandler,
	paymentHandler *handler.PaymentHandler,
) {
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/admin-user", userHandler.Create)
	}

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		users := api.Group("/users")
		{
			users.POST("/", userHandler.Create)
			users.GET("/:id", userHandler.Get)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
		}

		payments := api.Group("/payments")
		{
			payments.POST("/payment-intent", paymentHandler.CreatePaymentIntent)
			payments.POST("/retrieve", paymentHandler.RetrievePaymentIntent)
		}
	}
}
```

**Step 6: Verify compile**

```bash
go build ./internal/transport/...
```
Expected: no output

**Step 7: Commit**

```bash
git add internal/transport/http/
git commit -m "feat: add HTTP transport layer (handlers, middleware, routes)"
```

---

## Task 8: Move gRPC to internal/transport/grpc

**Files:**
- Create: `internal/transport/grpc/server/` (move from `grpc/server/`)
- Create: `internal/transport/grpc/client/` (move from `grpc/client/`)
- Delete: `grpc/` (old location)

**Step 1: Move gRPC server files**

```bash
mkdir -p internal/transport/grpc/server internal/transport/grpc/client
cp grpc/server/main.go internal/transport/grpc/server/main.go
cp grpc/server/user_service.go internal/transport/grpc/server/user_service.go
cp -r grpc/client/. internal/transport/grpc/client/
```

**Step 2: Update package import paths in the copied files**

In `internal/transport/grpc/server/main.go`, update imports:
- `go-gin-project/grpc/server` → stays as `package server`
- Any internal imports: update from old paths to new `go-gin-project/internal/...` paths

In `internal/transport/grpc/server/user_service.go`, update:
- `go-gin-project/services` → `go-gin-project/internal/application`
- `go-gin-project/models` → `go-gin-project/internal/domain`
- `go-gin-project/interfaces` → `go-gin-project/internal/domain`

**Step 3: Verify compile**

```bash
go build ./internal/transport/grpc/...
```
Expected: no output

**Step 4: Commit**

```bash
git add internal/transport/grpc/
git commit -m "feat: move gRPC to internal/transport/grpc"
```

---

## Task 9: Move proto to api/proto

**Files:**
- Create: `api/proto/` (move from `proto/`)

**Step 1: Move proto files**

```bash
mkdir -p api/proto
cp proto/*.proto api/proto/
cp proto/*.pb.go api/proto/
```

**Step 2: Update package declarations**

In the copied `.pb.go` files, update the `option go_package` and any import references from `go-gin-project/proto` to `go-gin-project/api/proto`.

Update `internal/transport/grpc/server/` files to import from `go-gin-project/api/proto` instead of `go-gin-project/proto`.

**Step 3: Update Makefile proto target**

In `Makefile`, update the proto generation command:
```makefile
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/*.proto
```

**Step 4: Verify compile**

```bash
go build ./api/...
go build ./internal/transport/grpc/...
```

**Step 5: Commit**

```bash
git add api/ Makefile
git commit -m "feat: move proto definitions to api/proto"
```

---

## Task 10: Update tests directory

**Files:**
- Create: `test/mocks/cache_mock.go` (replaces `tests/mocks/cache_mock.go`)
- Create: `test/mocks/stripe_mock.go` (replaces `tests/mocks/stripe_mock.go`)
- Create: `test/services/user_service_test.go` (updated from `tests/services/`)
- Delete: `tests/` (old location)

**Step 1: Create `test/mocks/cache_mock.go`**

Update to implement `domain.CacheService` (method names changed: `GetCache`→`Get`, `SetCache`→`Set`, `DeleteCache`→`Delete`):

```go
package mocks

import (
	"time"

	"github.com/stretchr/testify/mock"
)

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(key string, dest interface{}) error {
	args := m.Called(key, dest)
	return args.Error(0)
}

func (m *MockCache) Set(key string, value interface{}, expiration time.Duration) error {
	args := m.Called(key, value, expiration)
	return args.Error(0)
}

func (m *MockCache) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}
```

**Step 2: Create `test/mocks/stripe_mock.go`**

Copy `tests/mocks/stripe_mock.go` and update import from `go-gin-project/services` to `go-gin-project/internal/domain`.

**Step 3: Update `test/services/user_service_test.go`**

Update all imports:
- `go-gin-project/interfaces` → `go-gin-project/internal/domain`
- `go-gin-project/models` → `go-gin-project/internal/domain`
- `go-gin-project/services` → `go-gin-project/internal/application`

Remove the duplicate `MockCache` type in the test file — import it from `go-gin-project/test/mocks` instead.

Update `MockCache` method calls: `GetCache`→`Get`, `SetCache`→`Set`, `DeleteCache`→`Delete`.

**Step 4: Run tests**

```bash
go test ./test/... -v
```
Expected: all tests pass

**Step 5: Commit**

```bash
git add test/
git commit -m "feat: move and update tests to test/ directory with new import paths"
```

---

## Task 11: Update main.go and config

**Files:**
- Modify: `main.go`
- Modify: `config/config.go`

**Step 1: Update `config/config.go`**

Update model imports from `go-gin-project/models` to `go-gin-project/internal/domain`:

```go
// In migrateModels(), change:
import "go-gin-project/internal/domain"

// models list:
models := []interface{}{
    &domain.User{},
    &domain.Payment{},
}
```

**Step 2: Rewrite `main.go`**

Replace the old wiring with new layer-aware wiring:

```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-gin-project/config"
	"go-gin-project/internal/application"
	"go-gin-project/internal/infrastructure/mysql"
	redisinfra "go-gin-project/internal/infrastructure/redis"
	stripeinfra "go-gin-project/internal/infrastructure/stripe"
	httphandler "go-gin-project/internal/transport/http/handler"
	httptransport "go-gin-project/internal/transport/http"
	grpcserver "go-gin-project/internal/transport/grpc/server"

	_ "go-gin-project/docs"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	if err := config.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer config.CloseDB() //nolint:errcheck

	// Infrastructure
	cache, err := redisinfra.NewCache()
	if err != nil {
		log.Printf("Warning: Redis unavailable, continuing without cache: %v", err)
	}

	stripeClient, err := stripeinfra.NewClient()
	if err != nil {
		log.Printf("Warning: Stripe unavailable: %v", err)
	}

	userRepo := mysql.NewUserRepository(config.DB)
	paymentRepo := mysql.NewPaymentRepository(config.DB)

	// Application
	userService := application.NewUserService(userRepo, cache)
	authService := application.NewAuthService(userRepo)
	paymentService := application.NewPaymentService(paymentRepo, userRepo, cache, stripeClient)

	// Transport
	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	userHandler := httphandler.NewUserHandler(userService)
	authHandler := httphandler.NewAuthHandler(authService)
	paymentHandler := httphandler.NewPaymentHandler(paymentService)
	httptransport.SetupRoutes(r, userHandler, authHandler, paymentHandler)

	srv := &http.Server{Addr: ":8080", Handler: r}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	go func() {
		if err := grpcserver.StartGrpcServer(cache); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
}
```

**Step 3: Build the whole project**

```bash
go build ./...
```
Expected: no output (success)

**Step 4: Run tests**

```bash
go test ./... -v
```
Expected: all tests pass

**Step 5: Commit**

```bash
git add main.go config/config.go
git commit -m "feat: update main.go and config to use new DDD layer structure"
```

---

## Task 12: Remove old directories + update CLAUDE.md

**Files:**
- Delete: `models/`, `services/`, `controllers/`, `middleware/`, `routes/`, `grpc/`, `proto/`, `tests/`, `interfaces/`
- Modify: `CLAUDE.md`
- Modify: `.claude/rules/go-standard.md`

**Step 1: Verify nothing imports from old paths**

```bash
grep -r "go-gin-project/models\|go-gin-project/services\|go-gin-project/controllers\|go-gin-project/middleware\|go-gin-project/routes\|go-gin-project/interfaces" --include="*.go" .
```
Expected: no output

**Step 2: Delete old directories**

```bash
rm -rf models/ services/ controllers/ middleware/ routes/ grpc/ proto/ tests/ interfaces/
```

**Step 3: Final build and test**

```bash
go build ./...
go test ./... -v
```
Expected: no errors

**Step 4: Update `CLAUDE.md`** — update Architecture section to reflect new structure.

**Step 5: Update `.claude/rules/go-standard.md`** — update "This Project's Layout" section to show the new structure.

**Step 6: Final commit**

```bash
git add -A
git commit -m "refactor: remove old directories after DDD restructure"
```

---

## Verification Checklist

After all tasks complete:

- [ ] `go build ./...` — no errors
- [ ] `go test ./... -v` — all pass
- [ ] `go vet ./...` — no issues
- [ ] No imports from old paths (`models`, `services`, `controllers`, `middleware`, `routes`, `interfaces`)
- [ ] `internal/domain` imports nothing from other internal layers
- [ ] `internal/application` imports only `internal/domain`
- [ ] `internal/infrastructure` imports only `internal/domain`
- [ ] `internal/transport` imports only `internal/application`
- [ ] `main.go` is the only file importing across all layers
