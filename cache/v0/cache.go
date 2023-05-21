package v0

import (
	"context"
	"time"
)

// Cache 屏蔽不同的缓存中间件的差异
type Cache interface {
	// val, err := Get(ctx, key)
	// str = val.(string)
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, val any, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}
