package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd

	GetTimeout() time.Duration
}

type RedisConn struct {
	*redis.Client

	timeout time.Duration
}

func NewRedis(ctx context.Context, config Config) (*RedisConn, error) {
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.Password,
		DB:       config.DB,
	})

	result := rdb.Ping(ctx)
	if err := result.Err(); err != nil {
		return nil, err
	}

	return &RedisConn{
		Client:  rdb,
		timeout: config.Timeout,
	}, nil
}

func (r *RedisConn) GetTimeout() time.Duration {
	return r.timeout
}
