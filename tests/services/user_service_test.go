package services_test

import (
	"go-gin-project/interfaces"
	"go-gin-project/models"
	"go-gin-project/services"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockDB for testing
type MockDB struct {
	mock.Mock
}

var _ interfaces.DBInterface = (*MockDB)(nil)

func (m *MockDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(dest, conds)
	if user, ok := dest.(*models.User); ok && args.Get(0) != nil {
		mockUser := args.Get(0).(*models.User)
		*user = *mockUser
	}
	return &gorm.DB{Error: args.Error(1)}
}

func (m *MockDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	if user, ok := value.(*models.User); ok && args.Error(0) == nil {
		user.ID = 1
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
	}
	return &gorm.DB{Error: args.Error(0)}
}

func (m *MockDB) Save(value interface{}) *gorm.DB {
	args := m.Called(value)
	return &gorm.DB{Error: args.Error(0)}
}

func (m *MockDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	m.Called(query, args)
	return &gorm.DB{Error: gorm.ErrRecordNotFound}
}

func (m *MockDB) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(value, conds)
	return &gorm.DB{Error: args.Error(0)}
}

// MockRedisCache for testing
type MockRedisCache struct {
	mock.Mock
}

var _ interfaces.CacheInterface = (*MockRedisCache)(nil)

func (m *MockRedisCache) GetCache(key string, dest interface{}) error {
	args := m.Called(key, dest)
	return args.Error(0)
}

func (m *MockRedisCache) SetCache(key string, value interface{}, expiration time.Duration) error {
	args := m.Called(key, value, expiration)
	return args.Error(0)
}

func (m *MockRedisCache) DeleteCache(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func TestUserService_Create(t *testing.T) {
	mockDB := &MockDB{}
	mockCache := &MockRedisCache{}
	userService := services.NewUserService(mockDB, mockCache)

	t.Run("successful user creation", func(t *testing.T) {
		user := &models.User{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "password123",
		}

		// Setup mock expectations
		mockDB.On("Where", mock.AnythingOfType("string"), mock.AnythingOfType("[]interface {}")).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
		mockDB.On("First", mock.AnythingOfType("*models.User"), mock.Anything).Return(nil, gorm.ErrRecordNotFound)
		mockDB.On("Create", mock.AnythingOfType("*models.User")).Return(&gorm.DB{Error: nil})

		// Execute test
		createdUser, err := userService.Create(user)

		// Assert results
		assert.NoError(t, err)
		assert.NotNil(t, createdUser)
		assert.Equal(t, user.Name, createdUser.Name)
		assert.Equal(t, user.Email, createdUser.Email)
		assert.Empty(t, createdUser.Password)

		mockDB.AssertExpectations(t)
	})

	t.Run("duplicate email", func(t *testing.T) {
		user := &models.User{
			Email: "existing@example.com",
		}

		// Setup mock expectations
		mockDB.On("Where", mock.AnythingOfType("string"), mock.AnythingOfType("[]interface {}")).Return(&gorm.DB{Error: nil})
		mockDB.On("First", mock.AnythingOfType("*models.User"), mock.Anything).Return(&models.User{}, nil)

		// Execute test
		createdUser, err := userService.Create(user)

		// Assert results
		assert.Error(t, err)
		assert.Nil(t, createdUser)
		assert.Contains(t, err.Error(), "user already exists")
		mockDB.AssertExpectations(t)
	})
}

func TestUserService_Get(t *testing.T) {
	mockDB := &MockDB{}
	mockCache := &MockRedisCache{}
	userService := services.NewUserService(mockDB, mockCache)

	t.Run("get user successfully", func(t *testing.T) {
		userID := "1"
		mockUser := &models.User{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}

		// Setup mock expectations
		mockDB.On("First", mock.AnythingOfType("*models.User"), userID).Return(mockUser, nil)
		mockCache.On("GetCache", "user:"+userID, mock.AnythingOfType("*models.User")).Return(gorm.ErrRecordNotFound)
		mockCache.On("SetCache", "user:"+userID, mock.AnythingOfType("*models.User"), 5*time.Minute).Return(nil)

		// Execute test
		user, err := userService.Get(userID)

		// Assert results
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, mockUser.ID, user.ID)
		assert.Equal(t, mockUser.Name, user.Name)
		mockDB.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})
}

func TestUserService_Update(t *testing.T) {
	mockDB := &MockDB{}
	mockCache := &MockRedisCache{}
	userService := services.NewUserService(mockDB, mockCache)

	t.Run("successful update", func(t *testing.T) {
		userID := "1"
		existingUser := &models.User{
			ID:    1,
			Name:  "Original Name",
			Email: "original@example.com",
		}
		updateData := &models.User{
			Name: "Updated Name",
		}

		// Setup mock expectations
		mockDB.On("First", mock.AnythingOfType("*models.User"), userID).Return(existingUser, nil)
		mockDB.On("Save", mock.AnythingOfType("*models.User")).Return(&gorm.DB{Error: nil})
		mockCache.On("DeleteCache", "user:"+userID).Return(nil)

		// Execute test
		updatedUser, err := userService.Update(userID, updateData)

		// Assert results
		assert.NoError(t, err)
		assert.NotNil(t, updatedUser)
		assert.Equal(t, updateData.Name, updatedUser.Name)
		mockDB.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})
}

func TestUserService_Delete(t *testing.T) {
	mockDB := &MockDB{}
	mockCache := &MockRedisCache{}
	userService := services.NewUserService(mockDB, mockCache)

	t.Run("successful deletion", func(t *testing.T) {
		userID := "1"
		existingUser := &models.User{
			ID:    1,
			Name:  "To Delete",
			Email: "delete@example.com",
		}

		// Setup mock expectations
		mockDB.On("First", mock.AnythingOfType("*models.User"), userID).Return(existingUser, nil)
		mockDB.On("Delete", mock.AnythingOfType("*models.User")).Return(&gorm.DB{Error: nil})
		mockCache.On("DeleteCache", "user:"+userID).Return(nil)

		// Execute test
		err := userService.Delete(userID)

		// Assert results
		assert.NoError(t, err)
		mockDB.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})
}
