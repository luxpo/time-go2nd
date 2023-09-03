package message

import (
	"testing"

	"github.com/bmizerany/assert"
)

func TestRespEncodeDecode(t *testing.T) {
	testCases := []struct {
		name string
		resp *Response
	}{
		{
			name: "normal",
			resp: &Response{
				RequestID:  1,
				Version:    12,
				Compressor: 13,
				Serializer: 14,
				Error:      []byte("Hello, world"),
				Data:       []byte("Hello, world"),
			},
		},
		{
			name: "no data",
			resp: &Response{
				RequestID:  1,
				Version:    12,
				Compressor: 13,
				Serializer: 14,
				Error:      []byte("Hello,\n world"),
			},
		},
		{
			name: "no error",
			resp: &Response{
				RequestID:  1,
				Version:    12,
				Compressor: 13,
				Serializer: 14,
				Data:       []byte("Hello,\n world"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.resp.calculateHeaderLength()
			tc.resp.calculateBodyLength()
			data := EncodeResp(tc.resp)
			req := DecodeResp(data)
			assert.Equal(t, tc.resp, req)
		})

	}
}

func (resp *Response) calculateHeaderLength() {
	resp.HeaderLength = 15 + uint32(len(resp.Error))
}

func (resp *Response) calculateBodyLength() {
	resp.BodyLength = uint32(len(resp.Data))
}
