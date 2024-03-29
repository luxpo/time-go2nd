package rpc

import (
	"context"
	"testing"

	"github.com/bmizerany/assert"
	"github.com/golang/mock/gomock"
	"github.com/luxpo/time-go2nd/micro/rpc2/message"
	"github.com/luxpo/time-go2nd/micro/rpc2/serialize/json"
)

//go:generate mockgen -destination=./mock_proxy.gen.go -package=rpc github.com/luxpo/time-go2nd/micro/rpc2 Proxy

func Test_setFuncField(t *testing.T) {
	type args struct {
		service Service
		proxy   Proxy
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		// {
		// 	name: "nil",
		// 	args: args{
		// 		service: nil,
		// 		proxy: func() Proxy {
		// 			ctrl := gomock.NewController(t)
		// 			proxy := NewMockProxy(ctrl)
		// 			return proxy
		// 		}(),
		// 	},
		// 	wantErr: ErrServiceNil,
		// },
		// {
		// 	name: "no pointer",
		// 	args: args{
		// 		service: UserService{},
		// 		proxy: func() Proxy {
		// 			ctrl := gomock.NewController(t)
		// 			proxy := NewMockProxy(ctrl)
		// 			return proxy
		// 		}(),
		// 	},
		// 	wantErr: ErrServiceWrongType,
		// },
		{
			name: "user service",
			args: args{
				service: &UserService{},
				proxy: func() Proxy {
					ctrl := gomock.NewController(t)
					proxy := NewMockProxy(ctrl)
					req := &message.Request{
						ServiceName: "user-service",
						MethodName:  "GetByID",
						Data:        []byte(`{"ID":123}`),
					}
					req.CalculateHeaderLength()
					req.CalculateBodyLength()

					resp := &message.Response{
						Data: []byte(`{"ID":123}`),
					}
					resp.CalculateHeaderLength()
					resp.CalculateBodyLength()

					proxy.EXPECT().
						Invoke(gomock.Any(), req).
						Return(resp, nil)
					return proxy
				}(),
			},
			wantErr: nil,
		},
	}
	s := &json.Serializer{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setFuncField(tt.args.service, tt.args.proxy, s)
			assert.Equal(t, tt.wantErr, err)
			if err != nil {
				return
			}
			resp, err := tt.args.service.(*UserService).GetByID(context.Background(), &GetByIDReq{ID: 123})
			assert.Equal(t, tt.wantErr, err)
			t.Log(resp)
		})
	}
}
