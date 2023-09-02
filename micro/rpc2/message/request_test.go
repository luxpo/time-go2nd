package message

import (
	"testing"

	"github.com/bmizerany/assert"
)

func TestEncodeDecode(t *testing.T) {
	testCases := []struct {
		name string
		req  *Request
	}{
		{
			name: "normal",
			req: &Request{
				RequestID:   1,
				Version:     12,
				Compressor:  13,
				Serializer:  14,
				ServiceName: "user-service",
				MethodName:  "GetByID",
				Meta: map[string]string{
					"trace-id": "123456",
					"a/b":      "a",
				},
				Data: []byte("Hello, world"),
			},
		},
		{
			// 禁止用户在 meta 里使用 \n 和 \r
			name: "data with '\n'",
			req: &Request{
				RequestID:   1,
				Version:     12,
				Compressor:  13,
				Serializer:  14,
				ServiceName: "user-service",
				MethodName:  "GetByID",
				Meta: map[string]string{
					"trace-id": "123456",
					"a/b":      "a",
				},
				Data: []byte("Hello,\n world"),
			},
		},
		{
			name: "no meta",
			req: &Request{
				RequestID:   1,
				Version:     12,
				Compressor:  13,
				Serializer:  14,
				ServiceName: "user-service",
				MethodName:  "GetByID",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.req.calculateHeaderLength()
			tc.req.calculateBodyLength()
			data := EncodeReq(tc.req)
			req := DecodeReq(data)
			assert.Equal(t, tc.req, req)
		})

	}
}

func (req *Request) calculateHeaderLength() {
	// 不要忘了分隔符的长度
	headerLength := 15 + len(req.ServiceName) +
		1 + len(req.MethodName) +
		1
	for key, value := range req.Meta {
		headerLength += len(key)
		headerLength++
		headerLength += len(value)
		headerLength++
	}
	req.HeaderLength = uint32(headerLength)
}

func (req *Request) calculateBodyLength() {
	req.BodyLength = uint32(len(req.Data))
}
