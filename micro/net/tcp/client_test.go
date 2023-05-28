package tcp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Send(t *testing.T) {
	server := &Server{}
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)
	client := &Client{
		network: "tcp",
		address: "localhost:8081",
	}
	resp, err := client.Send(context.Background(), "hello")
	require.NoError(t, err)
	assert.Equal(t, "hellohello", resp)
}
