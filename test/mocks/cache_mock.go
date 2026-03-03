package mocks

import (
	"time"

	"github.com/stretchr/testify/mock"
)

// MockCache implements model.CacheService for testing.
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
