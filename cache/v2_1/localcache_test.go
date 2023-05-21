package v2_1

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLocalCache_Loop(t *testing.T) {
	cnt := 0
	c := NewLocalCache(time.Second, LocalCacheWithEvictedCallback(func(k string, v any) {
		cnt++
	}))
	err := c.Set(context.Background(), "k", 123, time.Second)
	require.NoError(t, err)

	time.Sleep(time.Second * 3)
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.data["k"]
	require.False(t, ok)
	require.Equal(t, 1, cnt)
}
