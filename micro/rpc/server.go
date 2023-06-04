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
	services map[string]Service
}

func NewServer() *Server {
	return &Server{
		services: make(map[string]Service, 16),
	}
}

func (s *Server) RegisterService(service Service) {
	s.services[service.Name()] = service
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

		respData, err := s.handleMsg(reqBs)
		if err != nil {
			// 业务 error
			return err
		}

		encodedMsg := EncodeMsg(respData)
		_, err = conn.Write(encodedMsg)
		if err != nil {
			return err
		}
	}
}

func (s *Server) handleMsg(reqData []byte) ([]byte, error) {
	req := Request{}
	err := jsoniter.Unmarshal(reqData, &req)
	if err != nil {
		return nil, err
	}
	// 发起业务调用
	service, ok := s.services[req.ServiceName]
	if !ok {
		return nil, errors.New("service not available")
	}
	val := reflect.ValueOf(service)
	method := val.MethodByName(req.MethodName)

	inReq := reflect.New(method.Type().In(1).Elem())
	err = jsoniter.Unmarshal(req.Arg, inReq.Interface())
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
	resp, err := jsoniter.Marshal(result[0].Interface())
	return resp, err
}
