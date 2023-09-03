package rpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"reflect"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/luxpo/time-go2nd/micro/rpc2/message"
	"github.com/silenceper/pool"
)

var (
	ErrServiceNil       = errors.New("rpc: nil service is not supported")
	ErrServiceWrongType = errors.New("rpc: only first-level pointer to struct is supported")
)

func InitClientProxy(network, address string, service Service) error {
	client, err := NewClient(network, address)
	if err != nil {
		return err
	}
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
				req := &message.Request{
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,
					Data:        reqArg,
				}
				fmt.Println(req)

				req.CalculateHeaderLength()
				req.CalculateBodyLength()

				resp, err := proxy.Invoke(ctx, req)
				if err != nil {
					return []reflect.Value{
						retVal,
						reflect.ValueOf(err),
					}
				}
				fmt.Println(string(resp.Data))

				var retErr error
				if len(resp.Error) > 0 {
					// 服务端传来的 error
					retErr = errors.New(string(resp.Error))
				}

				if len(resp.Data) > 0 {
					err = jsoniter.Unmarshal(resp.Data, retVal.Interface())
					if err != nil {
						// 反序列化的 error
						return []reflect.Value{
							retVal,
							reflect.ValueOf(err),
						}
					}
				}

				var retErrVal reflect.Value
				if retErr != nil {
					retErrVal = reflect.ValueOf(retErr)
				} else {
					retErrVal = reflect.Zero(reflect.TypeOf(new(error)).Elem())
				}

				return []reflect.Value{
					retVal,
					retErrVal,
				}
			})
			fieldVal.Set(fnVal)
		}
	}

	return nil
}

type Client struct {
	pool pool.Pool
}

func NewClient(network, addr string) (*Client, error) {
	p, err := pool.NewChannelPool(
		&pool.Config{
			InitialCap:  1,
			MaxCap:      30,
			MaxIdle:     10,
			IdleTimeout: time.Minute,
			Factory: func() (interface{}, error) {
				return net.DialTimeout(network, addr, time.Second*3)
			},
			Close: func(i interface{}) error {
				return i.(net.Conn).Close()
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		pool: p,
	}, nil
}

func (c *Client) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	data := message.EncodeReq(req)
	resp, err := c.Send(ctx, data)
	if err != nil {
		return nil, err
	}
	return message.DecodeResp(resp), nil
}

func (c *Client) Send(ctx context.Context, data []byte) ([]byte, error) {
	val, err := c.pool.Get()
	if err != nil {
		return nil, err
	}
	conn := val.(net.Conn)
	defer func() {
		_ = conn.Close()
	}()

	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}

	respBs, err := ReadMsg(conn)
	return respBs, err
}
