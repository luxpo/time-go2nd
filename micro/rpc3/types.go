package rpc

import (
	"context"

	"github.com/luxpo/time-go2nd/micro/rpc2/message"
)

type Service interface {
	Name() string
}

type Proxy interface {
	Invoke(ctx context.Context, req *message.Request) (resp *message.Response, err error)
}
