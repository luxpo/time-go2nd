package rpc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrServiceNil       = errors.New("rpc: nil service is not supported")
	ErrServiceWrongType = errors.New("rpc: only first-level pointer to struct is supported")
)

func InitClientProxy(service Service) error {
	return setFuncField(service, nil)
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

				req := &Request{
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,
					Arg:         args[1].Interface(),
				}
				fmt.Println(req)

				retVal := reflect.New(fieldTyp.Type.Out(0)).Elem()
				resp, err := proxy.Invoke(ctx, req)
				if err != nil {
					return []reflect.Value{
						retVal,
						reflect.ValueOf(err),
					}
				}
				fmt.Println(resp)
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
