package rpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"reflect"
	"time"

	jsoniter "github.com/json-iterator/go"
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
				fmt.Println(string(resp.Data))

				err = jsoniter.Unmarshal(resp.Data, retVal.Interface())
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
		Data: []byte(resp),
	}, nil
}

func (c *Client) Send(ctx context.Context, data string) (string, error) {
	val, err := c.pool.Get()
	if err != nil {
		return "", err
	}
	conn := val.(net.Conn)
	defer func() {
		_ = conn.Close()
	}()

	encodedMsg := EncodeMsg([]byte(data))

	_, err = conn.Write(encodedMsg)
	if err != nil {
		return "", err
	}

	respBs, err := ReadMsg(conn)
	return string(respBs), err
}
