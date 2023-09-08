package rpc

import (
	"context"
	"errors"
	"net"
	"reflect"
	"strconv"

	"github.com/luxpo/time-go2nd/micro/rpc2/message"
	"github.com/luxpo/time-go2nd/micro/rpc2/serialize"
	"github.com/luxpo/time-go2nd/micro/rpc2/serialize/json"
)

// 长度字段使用的字节数量
const numOfLengthBytes = 8

type Server struct {
	stubs       map[string]reflectionStub
	serializers map[uint8]serialize.Serializer
}

func NewServer() *Server {
	s := &Server{
		stubs:       make(map[string]reflectionStub, 16),
		serializers: make(map[uint8]serialize.Serializer, 4),
	}
	s.RegisterSerializer(&json.Serializer{})
	return s
}

func (s *Server) RegisterSerializer(sl serialize.Serializer) {
	s.serializers[sl.Code()] = sl
}

func (s *Server) RegisterService(service Service) {
	s.stubs[service.Name()] = reflectionStub{
		service:     service,
		value:       reflect.ValueOf(service),
		serializers: s.serializers,
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

func (s *Server) handleConn(conn net.Conn) error {
	for {
		reqBs, err := ReadMsg(conn)
		if err != nil {
			return err
		}

		// 还原调用信息
		req := message.DecodeReq(reqBs)
		resp, err := s.Invoke(context.Background(), req)
		if err != nil {
			// 处理业务 error
			resp.Error = []byte(err.Error())
		}

		resp.CalculateHeaderLength()
		resp.CalculateBodyLength()

		_, err = conn.Write(message.EncodeResp(resp))
		if err != nil {
			return err
		}
	}
}

func (s *Server) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	// 发起业务调用
	stub, ok := s.stubs[req.ServiceName]

	resp := &message.Response{
		RequestID:  req.RequestID,
		Version:    req.Version,
		Compressor: req.Compressor,
		Serializer: req.Serializer,
	}

	if !ok {
		return resp, errors.New("service not available")
	}

	respData, err := stub.invoke(ctx, req)
	resp.Data = respData
	if err != nil {
		return resp, err
	}

	return resp, nil
}

type reflectionStub struct {
	service     Service
	value       reflect.Value
	serializers map[uint8]serialize.Serializer
}

func (s *reflectionStub) invoke(ctx context.Context, req *message.Request) ([]byte, error) {
	// 反射找到方法，并且执行调用
	method := s.value.MethodByName(req.MethodName)

	inReq := reflect.New(method.Type().In(1).Elem())
	serializer, ok := s.serializers[req.Serializer]
	if !ok {
		return nil, errors.New("micro: not supported serializer " + strconv.FormatUint(uint64(req.Serializer), 10))
	}
	err := serializer.Decode(req.Data, inReq.Interface())
	if err != nil {
		return nil, err
	}

	in := []reflect.Value{
		reflect.ValueOf(context.Background()),
		inReq,
	}
	result := method.Call(in)

	if result[1].Interface() != nil {
		err = result[1].Interface().(error)
	}

	var res []byte
	if result[0].IsNil() {
		return nil, err
	} else {
		var er error
		res, er = serializer.Encode(result[0].Interface())
		if er != nil {
			return nil, er
		}
	}

	return res, err
}
