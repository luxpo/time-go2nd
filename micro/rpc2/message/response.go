package message

type Response struct {
	HeaderLength uint32 // 协议头长度
	BodyLength   uint32 // 协议体长度
	MessageID    uint32 // 消息ID
	Version      uint8  // 版本
	Compressor   uint8  // 压缩算法
	Serializer   uint8  // 序列化协议

	Error []byte // 错误信息

	Data []byte // 响应数据
}
