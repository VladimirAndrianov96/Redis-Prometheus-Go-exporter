// mockgen -destination=mocks/redis_client.go -package=mocks exporter/exporter/client RedisClient
package client

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

// SliceOfClients is used to pass around slice of clients.
type SliceOfClients struct {
	RedisClients []RedisClient
}

// RedisClient interface to mock the network requests to Redis.
type RedisClient interface {
	Info(ctx context.Context, section ...string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
}