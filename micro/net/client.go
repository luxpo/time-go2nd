package net

import (
	"context"
	"fmt"
	"net"
)

func Connect(ctx context.Context, network string, address string) error {
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, network, address)
	if err != nil {
		return err
	}
	defer conn.Close()
	for {
		_, err = conn.Write([]byte("hello"))
		if err != nil {
			return err
		}
		res := make([]byte, 128)
		_, err = conn.Read(res)
		if err != nil {
			return err
		}
		fmt.Println(string(res))
	}
}
