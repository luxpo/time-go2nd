package v2_0

import (
	"context"
	"testing"
	"time"
)

func TestSet(t *testing.T) {
	c := &LocalCache{
		data: make(map[string]*item),
	}
	c.Set(context.Background(), "k", "v", time.Hour)
	t.Log("set sucessfully")
}
