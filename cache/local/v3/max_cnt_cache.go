package v3

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

var (
	errOverCapacity = errors.New("cache：capacity limit exceeded")
)

type MaxCntCache struct {
	*LocalCache
	cnt    int32
	maxCnt atomic.Int32
}

func NewMaxCntCache(c *LocalCache, maxCnt int32) *MaxCntCache {
	newCache := &MaxCntCache{
		LocalCache: c,
	}
	newCache.maxCnt.Store(maxCnt)

	evictFunc := c.onEvicted
	newCache.onEvicted = func(k string, v any) {
		newCache.maxCnt.Add(-1)
		if evictFunc != nil {
			evictFunc(k, v)
		}
	}

	return newCache
}

func (c *MaxCntCache) Set(ctx context.Context, k string, v any, expiration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.data[k]
	if !ok {
		if c.cnt+1 > c.maxCnt.Load() {
			// 淘汰策略
			return errOverCapacity
		}
		c.cnt++
	}

	return c.Set(ctx, k, v, expiration)
}
