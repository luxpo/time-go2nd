package message

import (
	"encoding/binary"
)

type Response struct {
	HeaderLength uint32 // 协议头长度
	BodyLength   uint32 // 协议体长度
	RequestID    uint32 // 消息ID
	Version      uint8  // 版本
	Compressor   uint8  // 压缩算法
	Serializer   uint8  // 序列化协议

	Error []byte // 错误信息

	Data []byte // 响应数据
}

func EncodeResp(resp *Response) []byte {
	bs := make([]byte, resp.BodyLength+resp.HeaderLength)

	binary.BigEndian.PutUint32(bs[:4], resp.HeaderLength)
	binary.BigEndian.PutUint32(bs[4:8], resp.BodyLength)
	binary.BigEndian.PutUint32(bs[8:12], resp.RequestID)
	bs[12] = resp.Version
	bs[13] = resp.Compressor
	bs[14] = resp.Serializer
	cur := bs[15:]

	copy(cur, resp.Error)
	cur = cur[len(resp.Error):]
	copy(cur, resp.Data)

	return bs
}

func DecodeResp(data []byte) *Response {
	resp := &Response{}

	resp.HeaderLength = binary.BigEndian.Uint32(data[:4])
	resp.BodyLength = binary.BigEndian.Uint32(data[4:8])
	resp.RequestID = binary.BigEndian.Uint32(data[8:12])
	resp.Version = data[12]
	resp.Compressor = data[13]

	resp.Serializer = data[14]

	if resp.HeaderLength > 15 {
		resp.Error = data[15:resp.HeaderLength]
	}

	if resp.BodyLength != 0 {
		resp.Data = data[resp.HeaderLength:]
	}

	return resp
}

func (resp *Response) CalculateHeaderLength() {
	resp.HeaderLength = 15 + uint32(len(resp.Error))
}

func (resp *Response) CalculateBodyLength() {
	resp.BodyLength = uint32(len(resp.Data))
}
