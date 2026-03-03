package domain

import "time"

// CacheService defines cache operations.
// Implemented by infrastructure/redis, consumed by application.
type CacheService interface {
	Get(key string, dest interface{}) error
	Set(key string, value interface{}, expiration time.Duration) error
	Delete(key string) error
}
