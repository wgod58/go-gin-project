package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go-gin-project/internal/pkg/model"

	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client *redis.Client
	ctx    context.Context
}

// New creates a Redis-backed model.CacheService.
func New() (model.CacheService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connect: %w", err)
	}

	return &redisCache{client: client, ctx: ctx}, nil
}

func (c *redisCache) Get(key string, dest interface{}) error {
	val, err := c.client.Get(c.ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

func (c *redisCache) Set(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache set marshal: %w", err)
	}
	return c.client.Set(c.ctx, key, data, expiration).Err()
}

func (c *redisCache) Delete(key string) error {
	return c.client.Del(c.ctx, key).Err()
}

var _ model.CacheService = (*redisCache)(nil) // compile-time interface check
