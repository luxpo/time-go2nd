package v2_1

import (
	"context"
	"testing"
	"time"
)

func TestSet(t *testing.T) {
	c := NewLocalCache(time.Millisecond)
	c.Set(context.Background(), "a", "aa", time.Second)
	t.Log(c.data)
	time.Sleep(time.Second + time.Millisecond)
	t.Log(c.data)
}

func TestClose(t *testing.T) {
	c := NewLocalCache(time.Second)
	time.Sleep(time.Second * 2)
	t.Log(1, c.Close())
	t.Log(2, c.Close())
}
