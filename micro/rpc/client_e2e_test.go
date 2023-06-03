package rpc

import (
	"context"
	"testing"
	"time"

	"github.com/bmizerany/assert"
	"github.com/stretchr/testify/require"
)

func TestInitClientProxy(t *testing.T) {
	server := NewServer()
	server.RegisterService(&UserServiceServer{})
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second)
	client := &UserService{}
	err := InitClientProxy("tcp", ":8081", client)
	require.NoError(t, err)
	resp, err := client.GetByID(context.Background(), &GetByIDReq{ID: 123})
	require.NoError(t, err)
	assert.Equal(t, &GetByIDResp{"hi"}, resp)
}
