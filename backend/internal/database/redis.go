package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ConnectRedis opens a Redis client and verifies it with a ping (fail fast).
func ConnectRedis(addr, password string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{Addr: addr, Password: password})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return rdb, nil
}
