package peer

import (
	"github.com/golang/protobuf/proto"
)

type Response PtnResponse

func (m *Response) Reset()         { *m = Response{} }
func (m *Response) String() string { return proto.CompactTextString(m) }
func (*Response) ProtoMessage()    {}

func (m *Response) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PtnResponse.Unmarshal(m, b)
}
func (m *Response) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PtnResponse.Marshal(b, m, deterministic)
}