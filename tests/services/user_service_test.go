package services_test

import (
	"database/sql"
	"fmt"
	"go-gin-project/interfaces"
	"go-gin-project/models"
	"go-gin-project/services"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MockCache for testing
type MockCache struct {
	mock.Mock
}

var _ interfaces.CacheInterface = (*MockCache)(nil)

func (m *MockCache) GetCache(key string, dest interface{}) error {
	args := m.Called(key, dest)
	return args.Error(0)
}

func (m *MockCache) SetCache(key string, value interface{}, expiration time.Duration) error {
	args := m.Called(key, value, expiration)
	return args.Error(0)
}

func (m *MockCache) DeleteCache(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, error) {
	// Create new mock database
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	// Connect to mock database using GORM
	dialector := mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	return db, mock, nil
}

func TestUserService_Create(t *testing.T) {
	db, sqlMock, err := setupTestDB(t)
	assert.NoError(t, err)

	mockCache := new(MockCache)
	userService := services.NewUserService(db, mockCache)

	t.Run("duplicate email", func(t *testing.T) {
		user := &models.User{
			Email: "existing@example.com",
		}

		// Expect begin transaction
		sqlMock.ExpectBegin()

		// Expect check for existing user with soft delete
		sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE email = ? AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT ?")).
			WithArgs(user.Email, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name"}).
				AddRow(1, user.Email, "Existing User"))

		// Expect rollback since user exists
		sqlMock.ExpectRollback()

		// Execute test
		createdUser, err := userService.Create(user)

		// Assert results
		assert.Error(t, err)
		assert.Nil(t, createdUser)
		assert.Contains(t, err.Error(), "user already exists")

		// Verify all expectations were met
		assert.NoError(t, sqlMock.ExpectationsWereMet())
	})

	t.Run("successful user creation", func(t *testing.T) {
		user := &models.User{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "password123",
		}

		// Expect user creation
		sqlMock.ExpectBegin()

		// Expect check for existing user with soft delete
		sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE email = ? AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT ?")).
			WithArgs(user.Email, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		sqlMock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users`")).
			WithArgs(
				user.Name,        // Name
				user.Email,       // Email
				sqlmock.AnyArg(), // Password (hashed)
				sqlmock.AnyArg(), // CreatedAt
				sqlmock.AnyArg(), // UpdatedAt
				nil,              // DeletedAt
			).WillReturnResult(sqlmock.NewResult(1, 1))
		sqlMock.ExpectCommit()

		// Execute test
		createdUser, err := userService.Create(user)

		// Assert results
		assert.NoError(t, err)
		assert.NotNil(t, createdUser)
		assert.Equal(t, user.Name, createdUser.Name)
		assert.Equal(t, user.Email, createdUser.Email)
		assert.Empty(t, createdUser.Password)

		// Verify all expectations were met
		assert.NoError(t, sqlMock.ExpectationsWereMet())
	})
}

func TestUserService_Get(t *testing.T) {
	db, sqlMock, err := setupTestDB(t)
	assert.NoError(t, err)

	mockCache := new(MockCache)
	userService := services.NewUserService(db, mockCache)

	t.Run("get user successfully", func(t *testing.T) {
		userID := "1"
		expectedUser := &models.User{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}

		// Mock cache miss
		mockCache.On("GetCache", "user:"+userID, mock.AnythingOfType("*models.User")).Return(sql.ErrNoRows)

		// Expect database query with exact SQL pattern
		sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE `users`.`id` = ? AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT ?")).
			WithArgs("1", 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
				AddRow(expectedUser.ID, expectedUser.Name, expectedUser.Email))

		// Mock cache set
		mockCache.On("SetCache", "user:"+userID, mock.AnythingOfType("*models.User"), 5*time.Minute).Return(nil)

		// Execute test
		user, err := userService.Get(userID)

		fmt.Println("**************** get user ****************")
		fmt.Println(user)
		fmt.Println(err)

		// Assert results
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Equal(t, expectedUser.Name, user.Name)

		// Verify all expectations were met
		assert.NoError(t, sqlMock.ExpectationsWereMet())
		mockCache.AssertExpectations(t)
	})

	// t.Run("user not found", func(t *testing.T) {
	// 	userID := "999"

	// 	// Mock cache miss
	// 	mockCache.On("GetCache", "user:"+userID, mock.AnythingOfType("*models.User")).Return(sql.ErrNoRows)

	// 	// Expect database query
	// 	sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE `users`.`id` = ? AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT ?")).
	// 		WithArgs(999, 1).
	// 		WillReturnError(gorm.ErrRecordNotFound)

	// 	// Execute test
	// 	user, err := userService.Get(userID)

	// 	// Assert results
	// 	assert.Error(t, err)
	// 	assert.Nil(t, user)
	// 	assert.Contains(t, err.Error(), "user not found")

	// 	// Verify all expectations were met
	// 	assert.NoError(t, sqlMock.ExpectationsWereMet())
	// 	mockCache.AssertExpectations(t)
	// })
}

// func TestUserService_Update(t *testing.T) {
// 	db, mock, err := setupTestDB(t)
// 	assert.NoError(t, err)

// 	mockCache := new(MockCache)
// 	userService := services.NewUserService(db, mockCache)

// 	t.Run("successful update", func(t *testing.T) {
// 		userID := "1"
// 		updateData := &models.User{
// 			Name: "Updated Name",
// 		}

// 		// Expect find user query
// 		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE `users`.`id` = ? ORDER BY `users`.`id` LIMIT 1")).
// 			WithArgs(1).
// 			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
// 				AddRow(1, "Original Name", "test@example.com"))

// 		// Expect update query
// 		mock.ExpectBegin()
// 		mock.ExpectExec(regexp.QuoteMeta("UPDATE `users`")).
// 			WillReturnResult(sqlmock.NewResult(1, 1))
// 		mock.ExpectCommit()

// 		// Mock cache delete
// 		mockCache.On("DeleteCache", "user:"+userID).Return(nil)

// 		// Execute test
// 		updatedUser, err := userService.Update(userID, updateData)

// 		// Assert results
// 		assert.NoError(t, err)
// 		assert.NotNil(t, updatedUser)
// 		assert.Equal(t, updateData.Name, updatedUser.Name)

// 		// Verify all expectations were met
// 		assert.NoError(t, mock.ExpectationsWereMet())
// 		mockCache.AssertExpectations(t)
// 	})
// }

// func TestUserService_Delete(t *testing.T) {
// 	db, mock, err := setupTestDB(t)
// 	assert.NoError(t, err)

// 	mockCache := new(MockCache)
// 	userService := services.NewUserService(db, mockCache)

// 	t.Run("successful deletion", func(t *testing.T) {
// 		userID := "1"

// 		// Expect find user query
// 		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE `users`.`id` = ? ORDER BY `users`.`id` LIMIT 1")).
// 			WithArgs(1).
// 			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Test User"))

// 		// Expect delete query
// 		mock.ExpectBegin()
// 		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `users`")).
// 			WillReturnResult(sqlmock.NewResult(1, 1))
// 		mock.ExpectCommit()

// 		// Mock cache delete
// 		mockCache.On("DeleteCache", "user:"+userID).Return(nil)

// 		// Execute test
// 		err := userService.Delete(userID)

// 		// Assert results
// 		assert.NoError(t, err)

// 		// Verify all expectations were met
// 		assert.NoError(t, mock.ExpectationsWereMet())
// 		mockCache.AssertExpectations(t)
// 	})
// }
