package services

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"go-gin-project/interfaces"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements Redis caching functionality
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisCache creates a new Redis cache service
func NewRedisCache() (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{
		client: client,
		ctx:    ctx,
	}, nil
}

// GetCache retrieves data from Redis cache
func (r *RedisCache) GetCache(key string, dest interface{}) error {
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// SetCache stores data in Redis cache with expiration
func (r *RedisCache) SetCache(key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(r.ctx, key, jsonData, expiration).Err()
}

// DeleteCache removes data from Redis cache
func (r *RedisCache) DeleteCache(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

var _ interfaces.CacheInterface = (*RedisCache)(nil) // Verify RedisCache implements CacheInterface
