package rpc

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"reflect"

	jsoniter "github.com/json-iterator/go"
)

var (
	ErrServiceNil       = errors.New("rpc: nil service is not supported")
	ErrServiceWrongType = errors.New("rpc: only first-level pointer to struct is supported")
)

func InitClientProxy(network, address string, service Service) error {
	client := NewClient(network, address)
	return setFuncField(service, client)
}

func setFuncField(service Service, proxy Proxy) error {
	if service == nil {
		return ErrServiceNil
	}

	val := reflect.ValueOf(service)
	typ := val.Type()
	if typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
		return ErrServiceWrongType
	}

	val = val.Elem()
	typ = typ.Elem()

	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		fieldTyp := typ.Field(i)
		fieldVal := val.Field(i)
		if fieldVal.CanSet() {
			fnVal := reflect.MakeFunc(fieldTyp.Type, func(args []reflect.Value) (results []reflect.Value) {
				ctx := args[0].Interface().(context.Context)
				retVal := reflect.New(fieldTyp.Type.Out(0).Elem())

				reqArg, err := jsoniter.Marshal(args[1].Interface())
				if err != nil {
					return []reflect.Value{
						retVal,
						reflect.ValueOf(err),
					}
				}
				req := &Request{
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,
					Arg:         reqArg,
				}
				fmt.Println(req)

				resp, err := proxy.Invoke(ctx, req)
				if err != nil {
					return []reflect.Value{
						retVal,
						reflect.ValueOf(err),
					}
				}
				fmt.Println(resp)

				err = jsoniter.UnmarshalFromString(resp.Data, retVal.Interface())
				if err != nil {
					return []reflect.Value{
						retVal,
						reflect.ValueOf(err),
					}
				}

				return []reflect.Value{
					retVal,
					reflect.Zero(reflect.TypeOf(new(error)).Elem()),
				}
			})
			fieldVal.Set(fnVal)
		}
	}

	return nil
}

type Client struct {
	address string
	network string
}

func NewClient(network, addr string) *Client {
	return &Client{
		address: addr,
		network: network,
	}
}

func (c *Client) Invoke(ctx context.Context, req *Request) (*Response, error) {
	data, err := jsoniter.MarshalToString(req)
	if err != nil {
		return nil, err
	}
	resp, err := c.Send(ctx, data)
	if err != nil {
		return nil, err
	}
	return &Response{
		Data: resp,
	}, nil
}

func (c *Client) Send(ctx context.Context, data string) (string, error) {
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, c.network, c.address)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = conn.Close()
	}()

	reqLen := len(data)

	// 我要在这，构建请求数据
	// data = reqLen 的 64 位表示 + respData
	req := make([]byte, reqLen+numOfLengthBytes)
	// 第一步：
	// 把长度写进去前八个字节
	binary.BigEndian.PutUint64(req[:numOfLengthBytes], uint64(reqLen))
	// 第二步：
	// 写入数据
	copy(req[numOfLengthBytes:], data)

	_, err = conn.Write(req)
	if err != nil {
		return "", err
	}

	lenBs := make([]byte, numOfLengthBytes)
	_, err = conn.Read(lenBs)
	if err != nil {
		return "", err
	}

	// 我响应有多长？
	length := binary.BigEndian.Uint64(lenBs)

	respBs := make([]byte, length)
	_, err = conn.Read(respBs)
	if err != nil {
		return "", err
	}

	return string(respBs), nil
}
