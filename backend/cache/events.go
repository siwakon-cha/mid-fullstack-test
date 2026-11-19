package cache

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const listVersionKey = "events:list:version"

type Events struct {
	client *redis.Client
}

func New(redisURL string) (*Events, error) {
	if strings.TrimSpace(redisURL) == "" {
		return &Events{}, nil
	}
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse REDIS_URL: %w", err)
	}
	return &Events{client: redis.NewClient(options)}, nil
}

func (c *Events) Ping(ctx context.Context) error {
	if c.client == nil {
		return fmt.Errorf("redis is disabled")
	}
	return c.client.Ping(ctx).Err()
}

func (c *Events) Key(ctx context.Context, page, pageSize int, search string) string {
	version := "0"
	if c.client != nil {
		if value, err := c.client.Get(ctx, listVersionKey).Result(); err == nil {
			version = value
		}
	}
	return fmt.Sprintf("events:list:v%s:p%d:s%d:q%s", version, page, pageSize, url.QueryEscape(search))
}

func (c *Events) Get(ctx context.Context, key string) ([]byte, bool) {
	if c.client == nil {
		return nil, false
	}
	value, err := c.client.Get(ctx, key).Bytes()
	return value, err == nil
}

func (c *Events) Set(ctx context.Context, key string, value []byte) {
	if c.client != nil {
		_ = c.client.Set(ctx, key, value, 30*time.Second).Err()
	}
}

func (c *Events) Invalidate(ctx context.Context) {
	if c.client != nil {
		_ = c.client.Incr(ctx, listVersionKey).Err()
	}
}
