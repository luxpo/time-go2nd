package v2_1

import (
	"context"
	"errors"
	"sync"
	"time"
)

type LocalCache struct {
	mu   sync.RWMutex
	data map[string]*item

	closeOnce sync.Once
	close     chan struct{}
}

type item struct {
	val      any
	deadline time.Time
}

func NewLocalCache(interval time.Duration) *LocalCache {
	c := &LocalCache{
		data:  make(map[string]*item),
		close: make(chan struct{}),
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
						delete(c.data, k)
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
}

// Get 的时候，粗暴的做法是直接加写锁，但是也可以考虑用 double-check 写法。
// 我们使用的就是这个方案。
// 前面我们提到不能使用 sync.Map，从这里也可以看出来。
func (c *LocalCache) Get(ctx context.Context, k string) (any, error) {
	c.mu.RLock()
	i, ok := c.data[k]
	c.mu.RUnlock()
	if !ok {
		return nil, errors.New("key not found")
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
			return nil, errors.New("key not found")
		}
		if i.deadlineBeforeNow(now) {
			delete(c.data, k)
			return nil, errors.New("key expired")
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
