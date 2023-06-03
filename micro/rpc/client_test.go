package rpc

import (
	"context"
	"testing"

	"github.com/bmizerany/assert"
)

func TestInitClientProxy(t *testing.T) {
	type args struct {
		service Service
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
		// 	},
		// 	wantErr: ErrServiceNil,
		// },
		// {
		// 	name: "no pointer",
		// 	args: args{
		// 		service: UserService{},
		// 	},
		// 	wantErr: ErrServiceWrongType,
		// },
		{
			name: "user service",
			args: args{
				service: &UserService{},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InitClientProxy(tt.args.service)
			assert.Equal(t, tt.wantErr, err)
			_, err = tt.args.service.(*UserService).GetByID(context.Background(), &GetByIDReq{ID: 123})
		})
	}
}
