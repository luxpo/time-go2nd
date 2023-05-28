package net

import (
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
		_, err := conn.Read(bs)
		if err != nil {
			return err
		}
		res := handleMsg(bs)
		_, err = conn.Write(res)
		if err != nil {
			return err
		}
	}
}

func handleMsg(req []byte) []byte {
	res := make([]byte, 2*len(req))
	copy(res[:len(req)], req)
	copy(res[len(req):], req)
	return res
}
