package v2_0

import (
	"context"
	"sync"
	"time"
)

type LocalCache struct {
	mu   sync.RWMutex
	data map[string]*item
}

type item struct {
	val      any
	deadline time.Time
}

func (c *LocalCache) Set(ctx context.Context, k string, v any, expiration time.Duration) {
	var dl time.Time
	if expiration > 0 {
		dl = time.Now().Add(expiration)
	}

	c.mu.Lock()
	c.data[k] = &item{
		val:      v,
		deadline: dl,
	}
	c.mu.Unlock()

	if expiration > 0 {
		time.AfterFunc(expiration, func() {
			c.mu.RLock()
			defer c.mu.RUnlock()
			v, ok := c.data[k]
			if ok && !v.deadline.IsZero() && v.deadline.Before(time.Now()) {
				delete(c.data, k)
			}
		})
	}
}
