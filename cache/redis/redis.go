package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	errFailedToSetCache = errors.New("cache: 写入 redis 失败")
)

type RedisCache struct {
	client redis.Cmdable
}

func NewRedisCache(client redis.Cmdable) *RedisCache {
	return &RedisCache{
		client: client,
	}
}

func (c *RedisCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	res, err := c.client.Set(ctx, key, val, expiration).Result()
	if err != nil {
		return err
	}
	if res != "OK" {
		return fmt.Errorf("%w, 返回信息 %s", errFailedToSetCache, res)
	}
	return nil
}

func (c *RedisCache) Get(ctx context.Context, key string) (any, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	_, err := c.client.Del(ctx, key).Result()
	return err
}
