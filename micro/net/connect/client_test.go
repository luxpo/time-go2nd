package net

import (
	"context"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	go func() {
		err := Serve("tcp", ":8082")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	err := Connect(ctx, "tcp", "localhost:8082")
	t.Log(err)
}
