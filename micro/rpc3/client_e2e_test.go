package rpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bmizerany/assert"
	"github.com/luxpo/time-go2nd/micro/proto/gen"
	"github.com/luxpo/time-go2nd/micro/rpc2/serialize/proto"
	"github.com/stretchr/testify/require"
)

func TestInitClientProxy(t *testing.T) {
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterService(service)
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second)
	usClient := &UserService{}
	client, err := NewClient("tcp", ":8081")
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)

	testCases := []struct {
		name string
		mock func()

		wantErr  error
		wantResp *GetByIDResp
	}{
		{
			name: "no error",
			mock: func() {
				service.Msg = "hi"
				service.Err = nil
			},
			wantErr:  nil,
			wantResp: &GetByIDResp{"hi"},
		},
		{
			name: "error",
			mock: func() {
				service.Msg = ""
				service.Err = errors.New("mock error")
			},
			wantErr:  errors.New("mock error"),
			wantResp: &GetByIDResp{},
		},
		{
			name: "both",
			mock: func() {
				service.Msg = "hi"
				service.Err = errors.New("mock error")
			},
			wantErr:  errors.New("mock error"),
			wantResp: &GetByIDResp{"hi"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			resp, err := usClient.GetByID(context.Background(), &GetByIDReq{ID: 123})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func TestInitServiceProto(t *testing.T) {
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterService(service)
	server.RegisterSerializer(&proto.Serializer{})
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second)
	usClient := &UserService{}
	client, err := NewClient("tcp", ":8081", ClientWithSerializer(&proto.Serializer{}))
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)

	testCases := []struct {
		name string
		mock func()

		wantErr  error
		wantResp *GetByIDResp
	}{
		{
			name: "no error",
			mock: func() {
				service.Msg = "hi"
				service.Err = nil
			},
			wantErr:  nil,
			wantResp: &GetByIDResp{"hi"},
		},
		{
			name: "error",
			mock: func() {
				service.Msg = ""
				service.Err = errors.New("mock error")
			},
			wantErr:  errors.New("mock error"),
			wantResp: &GetByIDResp{},
		},
		{
			name: "both",
			mock: func() {
				service.Msg = "hi"
				service.Err = errors.New("mock error")
			},
			wantErr:  errors.New("mock error"),
			wantResp: &GetByIDResp{"hi"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			resp, err := usClient.GetByIDProto(context.Background(), &gen.GetByIDReq{Id: 123})
			assert.Equal(t, tc.wantErr, err)
			if resp != nil && resp.User != nil {
				assert.Equal(t, tc.wantResp.Msg, resp.User.Name)
			}
		})
	}
}

func TestOneway(t *testing.T) {
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterService(service)
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second)
	usClient := &UserService{}
	client, err := NewClient("tcp", ":8081")
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)

	testCases := []struct {
		name string
		mock func()

		wantErr  error
		wantResp *GetByIDResp
	}{
		{
			name: "oneway",
			mock: func() {
				service.Msg = "hi"
				service.Err = errors.New("mock error")
			},
			wantErr:  errors.New("micro: oneway"),
			wantResp: &GetByIDResp{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			ctx := CtxWithOneway(context.Background())
			resp, err := usClient.GetByID(ctx, &GetByIDReq{ID: 123})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}
