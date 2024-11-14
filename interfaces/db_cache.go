package interfaces

import (
	"time"

	"gorm.io/gorm"
)

// DBInterface defines the database operations we need
type DBInterface interface {
	First(dest interface{}, conds ...interface{}) *gorm.DB
	Create(value interface{}) *gorm.DB
	Save(value interface{}) *gorm.DB
	Where(query interface{}, args ...interface{}) *gorm.DB
	Delete(value interface{}, conds ...interface{}) *gorm.DB
}

// CacheInterface defines the cache operations we need
type CacheInterface interface {
	GetCache(key string, dest interface{}) error
	SetCache(key string, value interface{}, expiration time.Duration) error
	DeleteCache(key string) error
}
