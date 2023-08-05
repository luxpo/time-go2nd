package rpc

import (
	"encoding/binary"
	"net"
)

func ReadMsg(conn net.Conn) ([]byte, error) {
	lenBs := make([]byte, numOfLengthBytes)
	_, err := conn.Read(lenBs)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint64(lenBs)
	data := make([]byte, length)
	_, err = conn.Read(data)
	return data, err
}

func EncodeMsg(data []byte) []byte {
	dataLen := len(data)
	res := make([]byte, dataLen+numOfLengthBytes)
	binary.BigEndian.PutUint64(res[:numOfLengthBytes], uint64(dataLen))
	copy(res[numOfLengthBytes:], data)
	return res
}
