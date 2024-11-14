package mocks

import (
	"time"

	"github.com/stretchr/testify/mock"
)

// MockCache implements the CacheService interface for testing
type MockCache struct {
	mock.Mock
}

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
