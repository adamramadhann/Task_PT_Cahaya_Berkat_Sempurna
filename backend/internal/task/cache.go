package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/redis/go-redis/v9"
)

const listKeyPrefix = "tasks:list:"

// TaskCache caches the task list response (Task 2). Every method is best
// effort: failures are logged and treated as a miss, so a Redis outage
// degrades to plain DB reads instead of failing requests.
type TaskCache interface {
	GetList(ctx context.Context, key string) (*ListResponse, bool)
	SetList(ctx context.Context, key string, resp *ListResponse)
	InvalidateList(ctx context.Context)
}

// ListCacheKey builds a cache key that includes every list query parameter.
// Fields are serialized in one fixed order, so two requests that only differ
// in parameter order share a single key.
func ListCacheKey(q ListTasksQuery) string {
	return fmt.Sprintf("%sassignee=%s&keyword=%s&limit=%d&page=%d&sort=%s&status=%s",
		listKeyPrefix,
		url.QueryEscape(q.Assignee),
		url.QueryEscape(q.Keyword),
		q.Limit,
		q.Page,
		url.QueryEscape(q.Sort),
		url.QueryEscape(q.Status),
	)
}

type redisCache struct {
	rdb *redis.Client
	ttl time.Duration
}

// NewRedisCache returns a Redis-backed TaskCache with the given TTL.
func NewRedisCache(rdb *redis.Client, ttl time.Duration) TaskCache {
	return &redisCache{rdb: rdb, ttl: ttl}
}

func (c *redisCache) GetList(ctx context.Context, key string) (*ListResponse, bool) {
	raw, err := c.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, false
	}
	if err != nil {
		log.Printf("[cache] get %s: %v", key, err)
		return nil, false
	}
	var resp ListResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		log.Printf("[cache] unmarshal %s: %v", key, err)
		return nil, false
	}
	return &resp, true
}

func (c *redisCache) SetList(ctx context.Context, key string, resp *ListResponse) {
	raw, err := json.Marshal(resp)
	if err != nil {
		log.Printf("[cache] marshal %s: %v", key, err)
		return
	}
	if err := c.rdb.Set(ctx, key, raw, c.ttl).Err(); err != nil {
		log.Printf("[cache] set %s: %v", key, err)
	}
}

// InvalidateList drops every cached list key. SCAN is used instead of KEYS so
// Redis is never blocked; the prefix flush keeps invalidation correct after
// any mutation regardless of which query combos were cached.
func (c *redisCache) InvalidateList(ctx context.Context) {
	var keys []string
	iter := c.rdb.Scan(ctx, 0, listKeyPrefix+"*", 100).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		log.Printf("[cache] scan %s*: %v", listKeyPrefix, err)
		return
	}
	if len(keys) == 0 {
		return
	}
	if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
		log.Printf("[cache] del %d keys: %v", len(keys), err)
	}
}
