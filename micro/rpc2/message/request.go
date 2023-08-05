package message

type Request struct {
	HeaderLength uint32 // 协议头长度
	BodyLength   uint32 // 协议体长度
	MessageID    uint32 // 消息ID
	Version      uint8  // 版本
	Compressor   uint8  // 压缩算法
	Serializer   uint8  // 序列化协议

	ServiceName string // 服务名
	MethodName  string // 方法名

	Meta map[string]string // 扩展字段，用于传递自定义元数据

	Data []byte // 协议体
}
