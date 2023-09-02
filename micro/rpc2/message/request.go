package message

import (
	"bytes"
	"encoding/binary"
)

const splitter = '\n'

type Request struct {
	HeaderLength uint32 // 协议头长度
	BodyLength   uint32 // 协议体长度
	RequestID    uint32 // 消息ID
	Version      uint8  // 版本
	Compressor   uint8  // 压缩算法
	Serializer   uint8  // 序列化协议

	ServiceName string // 服务名
	MethodName  string // 方法名

	Meta map[string]string // 扩展字段，用于传递自定义元数据

	Data []byte // 协议体
}

func EncodeReq(req *Request) []byte {
	bs := make([]byte, req.BodyLength+req.HeaderLength)

	binary.BigEndian.PutUint32(bs[:4], req.HeaderLength)
	binary.BigEndian.PutUint32(bs[4:8], req.BodyLength)
	binary.BigEndian.PutUint32(bs[8:12], req.RequestID)
	bs[12] = req.Version
	bs[13] = req.Compressor
	bs[14] = req.Serializer
	cur := bs[15:]
	copy(cur, req.ServiceName)
	cur = cur[len(req.ServiceName):]
	cur[0] = splitter
	cur = cur[1:]
	copy(cur, req.MethodName)
	cur = cur[len(req.MethodName):]
	cur[0] = splitter
	cur = cur[1:]

	for key, value := range req.Meta {
		copy(cur, key)
		cur = cur[len(key):]
		cur[0] = '\r'
		cur = cur[1:]
		copy(cur, value)
		cur = cur[len(value):]
		cur[0] = splitter
		cur = cur[1:]
	}

	copy(cur, req.Data)

	return bs
}

func DecodeReq(data []byte) *Request {
	req := &Request{}

	req.HeaderLength = binary.BigEndian.Uint32(data[:4])
	req.BodyLength = binary.BigEndian.Uint32(data[4:8])
	req.RequestID = binary.BigEndian.Uint32(data[8:12])
	req.Version = data[12]
	req.Compressor = data[13]

	req.Serializer = data[14]

	header := data[15:req.HeaderLength]

	index := bytes.IndexByte(header, splitter)
	req.ServiceName = string(header[:index])
	header = header[index+1:]

	index = bytes.IndexByte(header, splitter)
	req.MethodName = string(header[:index])
	header = header[index+1:]

	index = bytes.IndexByte(header, splitter)
	if index != -1 {
		meta := make(map[string]string, 16)
		for index != -1 {
			pair := header[:index]
			pairIndex := bytes.IndexByte(pair, '\r')
			key := string(pair[:pairIndex])
			value := string(pair[pairIndex+1:])
			meta[key] = value

			header = header[index+1:]
			index = bytes.IndexByte(header, splitter)
		}
		req.Meta = meta
	}

	if req.BodyLength != 0 {
		req.Data = data[req.HeaderLength:]
	}

	return req
}
