package net

import (
	"errors"
	"net"
	"testing"

	"github.com/bmizerany/assert"
	"github.com/golang/mock/gomock"
	"github.com/luxpo/time-go2nd/micro/net/connect/mocks"
)

//go:generate mockgen -destination=./mocks/mock_net_conn.gen.go -package=mocks net Conn

func Test_handleConn(t *testing.T) {
	type args struct {
		conn net.Conn
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "read error",
			args: args{
				conn: func() net.Conn {
					ctrl := gomock.NewController(t)
					conn := mocks.NewMockConn(ctrl)
					conn.EXPECT().Read(gomock.Any()).Return(0, errors.New("read error"))
					return conn
				}(),
			},
			wantErr: errors.New("read error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handleConn(tt.args.conn)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
