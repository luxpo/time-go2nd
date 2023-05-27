package v3

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	errKeyNotFound = errors.New("cache：key not found")
)

type LocalCache struct {
	mu   sync.RWMutex
	data map[string]*item

	closeOnce sync.Once
	close     chan struct{}

	onEvicted func(k string, v any)
}

type LocalCacheOption func(cache *LocalCache)

type item struct {
	val      any
	deadline time.Time
}

func NewLocalCache(interval time.Duration, opts ...LocalCacheOption) *LocalCache {
	c := &LocalCache{
		data:  make(map[string]*item),
		close: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(c)
	}

	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case t := <-ticker.C:
				c.mu.Lock()
				i := 0
				for k, v := range c.data {
					if i > 10000 {
						break
					}
					if v.deadlineBeforeNow(t) {
						c.delete(k)
					}
					i++
				}
				c.mu.Unlock()
			case <-c.close:
				return
			}
		}
	}()

	return c
}

func LocalCacheWithEvictedCallback(fn func(k string, v any)) LocalCacheOption {
	return func(cache *LocalCache) {
		cache.onEvicted = fn
	}
}

func (c *LocalCache) Set(ctx context.Context, k string, v any, expiration time.Duration) error {
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

	return nil
}

// Get 的时候，粗暴的做法是直接加写锁，但是也可以考虑用 double-check 写法。
// 我们使用的就是这个方案。
// 前面我们提到不能使用 sync.Map，从这里也可以看出来。
func (c *LocalCache) Get(ctx context.Context, k string) (any, error) {
	c.mu.RLock()
	i, ok := c.data[k]
	c.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w, key: %s", errKeyNotFound, k)
	}

	now := time.Now()
	// 定时轮询单独使用肯定是不行的，
	// 因为一个 key 可能已经过期了，但是还没轮到它，
	// 一般都是跟 Get 的时候检查过期时间配合使用。
	if i.deadlineBeforeNow(now) {
		c.mu.Lock()
		defer c.mu.Unlock()
		// double check
		i, ok = c.data[k]
		if !ok {
			return nil, fmt.Errorf("%w, key: %s", errKeyNotFound, k)
		}
		if i.deadlineBeforeNow(now) {
			c.delete(k)
			return nil, fmt.Errorf("%w, key: %s", errKeyNotFound, k)
		}
	}

	return i.val, nil
}

func (c *LocalCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
	return nil
}

func (i *item) deadlineBeforeNow(t time.Time) bool {
	return !i.deadline.IsZero() && i.deadline.Before(t)
}

func (c *LocalCache) delete(k string) {
	item, ok := c.data[k]
	if !ok {
		return
	}
	delete(c.data, k)
	c.onEvicted(k, item.val)
}

func (c *LocalCache) Close() error {
	// 方法一
	// // 要确保 cache 已启动
	// select {
	// case c.close <- struct{}{}:
	// default:
	// 	return errors.New("already closed")
	// }

	// 方法二
	c.closeOnce.Do(func() {
		c.close <- struct{}{}
	})

	return nil
}

func (c *LocalCache) LoadAndDelete(ctx context.Context, key string) (any, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.data[key]
	if !ok {
		return nil, errKeyNotFound
	}
	c.delete(key)
	return v.val, nil
}
