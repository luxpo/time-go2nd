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
			tc.req.CalculateHeaderLength()
			tc.req.CalculateBodyLength()
			data := EncodeReq(tc.req)
			req := DecodeReq(data)
			assert.Equal(t, tc.req, req)
		})

	}
}
