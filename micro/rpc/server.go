package rpc

import (
	"context"
	"errors"
	"net"
	"reflect"

	jsoniter "github.com/json-iterator/go"
)

// 长度字段使用的字节数量
const numOfLengthBytes = 8

type Server struct {
	stubs map[string]reflectionStub
}

func NewServer() *Server {
	return &Server{
		stubs: make(map[string]reflectionStub, 16),
	}
}

func (s *Server) RegisterService(service Service) {
	s.stubs[service.Name()] = reflectionStub{
		service: service,
		value:   reflect.ValueOf(service),
	}
}

func (s *Server) Start(network, address string) error {
	listener, err := net.Listen(network, address)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			if herr := s.handleConn(conn); herr != nil {
				_ = conn.Close()
			}
		}()
	}
}

// 我们可以认为，一个请求包含两部分
// 1. 长度字段：用八个字节表示
// 2. 请求数据：
// 响应也是这个规范
func (s *Server) handleConn(conn net.Conn) error {
	for {
		reqBs, err := ReadMsg(conn)
		if err != nil {
			return err
		}

		// 还原调用信息
		req := &Request{}
		err = jsoniter.Unmarshal(reqBs, req)
		if err != nil {
			return err
		}

		resp, err := s.Invoke(context.Background(), req)
		if err != nil {
			// 这个可能你的业务 error
			// 暂时不知道怎么回传 error，所以我们简单记录一下
			return err
		}

		encodedMsg := EncodeMsg(resp.Data)
		_, err = conn.Write(encodedMsg)
		if err != nil {
			return err
		}
	}
}

func (s *Server) Invoke(ctx context.Context, req *Request) (*Response, error) {
	// 发起业务调用
	stub, ok := s.stubs[req.ServiceName]
	if !ok {
		return nil, errors.New("service not available")
	}

	resp, err := stub.invoke(ctx, req.MethodName, req.Arg)
	if err != nil {
		return nil, err
	}

	return &Response{
		Data: resp,
	}, err
}

type reflectionStub struct {
	service Service
	value   reflect.Value
}

func (s *reflectionStub) invoke(ctx context.Context, methodName string, data []byte) ([]byte, error) {
	// 反射找到方法，并且执行调用
	method := s.value.MethodByName(methodName)

	inReq := reflect.New(method.Type().In(1).Elem())
	err := jsoniter.Unmarshal(data, inReq.Interface())
	if err != nil {
		return nil, err
	}

	in := []reflect.Value{
		reflect.ValueOf(context.Background()),
		inReq,
	}
	result := method.Call(in)
	if result[1].Interface() != nil {
		return nil, result[1].Interface().(error)
	}
	return jsoniter.Marshal(result[0].Interface())
}
