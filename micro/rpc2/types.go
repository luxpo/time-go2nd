package rpc

import "context"

type Service interface {
	Name() string
}

type Proxy interface {
	Invoke(ctx context.Context, req *Request) (resp *Response, err error)
}

type Request struct {
	ServiceName string
	MethodName  string
	Arg         []byte
}

type Response struct {
	Data []byte
}
