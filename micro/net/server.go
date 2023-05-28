package net

import (
	"errors"
	"net"
)

func Serve(network, address string) error {
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
			if herr := handleConn(conn); herr != nil {
				conn.Close()
			}
		}()
	}
}

func handleConn(conn net.Conn) error {
	for {
		bs := make([]byte, 8)
		n, err := conn.Read(bs)
		if err != nil {
			return err
		}
		if n != 8 {
			return errors.New("micro: 没读够数据")
		}
		res := handleMsg(bs)
		n, err = conn.Write(res)
		if err != nil {
			return err
		}
		if n != len(res) {
			return errors.New("micro: 没写完数据")
		}
	}
}

func handleMsg(req []byte) []byte {
	res := make([]byte, 2*len(req))
	copy(res[:len(req)], req)
	copy(res[len(req):], req)
	return res
}
