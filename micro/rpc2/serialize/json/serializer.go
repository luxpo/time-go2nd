package json

import jsoniter "github.com/json-iterator/go"

type Serializer struct {
}

func (s *Serializer) Code() uint8 {
	return 1
}

func (s *Serializer) Encode(val any) ([]byte, error) {
	return jsoniter.Marshal(val)
}

func (s *Serializer) Decode(data []byte, val any) error {
	return jsoniter.Unmarshal(data, val)
}
