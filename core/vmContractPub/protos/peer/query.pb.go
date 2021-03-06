// Code generated by protoc-gen-go. DO NOT EDIT.
// source: query.proto

package peer

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// ChaincodeQueryResponse returns information about each chaincode that pertains
// to a query in lscc.go, such as GetChaincodes (returns all chaincodes
// instantiated on a channel), and GetInstalledChaincodes (returns all chaincodes
// installed on a peer)
type PtnChaincodeQueryResponse struct {
	Chaincodes           []*PtnChaincodeInfo `protobuf:"bytes,1,rep,name=chaincodes,proto3" json:"chaincodes,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *PtnChaincodeQueryResponse) Reset()         { *m = PtnChaincodeQueryResponse{} }
func (m *PtnChaincodeQueryResponse) String() string { return proto.CompactTextString(m) }
func (*PtnChaincodeQueryResponse) ProtoMessage()    {}
func (*PtnChaincodeQueryResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_5c6ac9b241082464, []int{0}
}

func (m *PtnChaincodeQueryResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PtnChaincodeQueryResponse.Unmarshal(m, b)
}
func (m *PtnChaincodeQueryResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PtnChaincodeQueryResponse.Marshal(b, m, deterministic)
}
func (m *PtnChaincodeQueryResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PtnChaincodeQueryResponse.Merge(m, src)
}
func (m *PtnChaincodeQueryResponse) XXX_Size() int {
	return xxx_messageInfo_PtnChaincodeQueryResponse.Size(m)
}
func (m *PtnChaincodeQueryResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PtnChaincodeQueryResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PtnChaincodeQueryResponse proto.InternalMessageInfo

func (m *PtnChaincodeQueryResponse) GetChaincodes() []*PtnChaincodeInfo {
	if m != nil {
		return m.Chaincodes
	}
	return nil
}

// ChaincodeInfo contains general information about an installed/instantiated
// chaincode
type PtnChaincodeInfo struct {
	Name    string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Version string `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	// the path as specified by the install/instantiate transaction
	Path string `protobuf:"bytes,3,opt,name=path,proto3" json:"path,omitempty"`
	// the chaincode function upon instantiation and its arguments. This will be
	// blank if the query is returning information about installed chaincodes.
	Input string `protobuf:"bytes,4,opt,name=input,proto3" json:"input,omitempty"`
	// the name of the ESCC for this chaincode. This will be
	// blank if the query is returning information about installed chaincodes.
	Escc string `protobuf:"bytes,5,opt,name=escc,proto3" json:"escc,omitempty"`
	// the name of the VSCC for this chaincode. This will be
	// blank if the query is returning information about installed chaincodes.
	Vscc string `protobuf:"bytes,6,opt,name=vscc,proto3" json:"vscc,omitempty"`
	// the chaincode unique id.
	// computed as: H(
	//                H(name || version) ||
	//                H(CodePackage)
	//              )
	Id                   []byte   `protobuf:"bytes,7,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PtnChaincodeInfo) Reset()         { *m = PtnChaincodeInfo{} }
func (m *PtnChaincodeInfo) String() string { return proto.CompactTextString(m) }
func (*PtnChaincodeInfo) ProtoMessage()    {}
func (*PtnChaincodeInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_5c6ac9b241082464, []int{1}
}

func (m *PtnChaincodeInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PtnChaincodeInfo.Unmarshal(m, b)
}
func (m *PtnChaincodeInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PtnChaincodeInfo.Marshal(b, m, deterministic)
}
func (m *PtnChaincodeInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PtnChaincodeInfo.Merge(m, src)
}
func (m *PtnChaincodeInfo) XXX_Size() int {
	return xxx_messageInfo_PtnChaincodeInfo.Size(m)
}
func (m *PtnChaincodeInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_PtnChaincodeInfo.DiscardUnknown(m)
}

var xxx_messageInfo_PtnChaincodeInfo proto.InternalMessageInfo

func (m *PtnChaincodeInfo) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *PtnChaincodeInfo) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *PtnChaincodeInfo) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *PtnChaincodeInfo) GetInput() string {
	if m != nil {
		return m.Input
	}
	return ""
}

func (m *PtnChaincodeInfo) GetEscc() string {
	if m != nil {
		return m.Escc
	}
	return ""
}

func (m *PtnChaincodeInfo) GetVscc() string {
	if m != nil {
		return m.Vscc
	}
	return ""
}

func (m *PtnChaincodeInfo) GetId() []byte {
	if m != nil {
		return m.Id
	}
	return nil
}

// ChannelQueryResponse returns information about each channel that pertains
// to a query in lscc.go, such as GetChannels (returns all channels for a
// given peer)
type PtnChannelQueryResponse struct {
	Channels             []*PtnChannelInfo `protobuf:"bytes,1,rep,name=channels,proto3" json:"channels,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *PtnChannelQueryResponse) Reset()         { *m = PtnChannelQueryResponse{} }
func (m *PtnChannelQueryResponse) String() string { return proto.CompactTextString(m) }
func (*PtnChannelQueryResponse) ProtoMessage()    {}
func (*PtnChannelQueryResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_5c6ac9b241082464, []int{2}
}

func (m *PtnChannelQueryResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PtnChannelQueryResponse.Unmarshal(m, b)
}
func (m *PtnChannelQueryResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PtnChannelQueryResponse.Marshal(b, m, deterministic)
}
func (m *PtnChannelQueryResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PtnChannelQueryResponse.Merge(m, src)
}
func (m *PtnChannelQueryResponse) XXX_Size() int {
	return xxx_messageInfo_PtnChannelQueryResponse.Size(m)
}
func (m *PtnChannelQueryResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PtnChannelQueryResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PtnChannelQueryResponse proto.InternalMessageInfo

func (m *PtnChannelQueryResponse) GetChannels() []*PtnChannelInfo {
	if m != nil {
		return m.Channels
	}
	return nil
}

// ChannelInfo contains general information about channels
type PtnChannelInfo struct {
	ChannelId            string   `protobuf:"bytes,1,opt,name=channel_id,json=channelId,proto3" json:"channel_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PtnChannelInfo) Reset()         { *m = PtnChannelInfo{} }
func (m *PtnChannelInfo) String() string { return proto.CompactTextString(m) }
func (*PtnChannelInfo) ProtoMessage()    {}
func (*PtnChannelInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_5c6ac9b241082464, []int{3}
}

func (m *PtnChannelInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PtnChannelInfo.Unmarshal(m, b)
}
func (m *PtnChannelInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PtnChannelInfo.Marshal(b, m, deterministic)
}
func (m *PtnChannelInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PtnChannelInfo.Merge(m, src)
}
func (m *PtnChannelInfo) XXX_Size() int {
	return xxx_messageInfo_PtnChannelInfo.Size(m)
}
func (m *PtnChannelInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_PtnChannelInfo.DiscardUnknown(m)
}

var xxx_messageInfo_PtnChannelInfo proto.InternalMessageInfo

func (m *PtnChannelInfo) GetChannelId() string {
	if m != nil {
		return m.ChannelId
	}
	return ""
}

func init() {
	proto.RegisterType((*PtnChaincodeQueryResponse)(nil), "protos.PtnChaincodeQueryResponse")
	proto.RegisterType((*PtnChaincodeInfo)(nil), "protos.PtnChaincodeInfo")
	proto.RegisterType((*PtnChannelQueryResponse)(nil), "protos.PtnChannelQueryResponse")
	proto.RegisterType((*PtnChannelInfo)(nil), "protos.PtnChannelInfo")
}

func init() { proto.RegisterFile("query.proto", fileDescriptor_5c6ac9b241082464) }

var fileDescriptor_5c6ac9b241082464 = []byte{
	// 304 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x91, 0xbf, 0x4e, 0xf3, 0x30,
	0x14, 0xc5, 0x95, 0xfe, 0xfd, 0x7a, 0xfb, 0xa9, 0x42, 0x16, 0x02, 0x33, 0x20, 0x55, 0x99, 0xba,
	0x50, 0x4b, 0x65, 0x61, 0x44, 0x74, 0xea, 0x80, 0x28, 0x91, 0x58, 0x58, 0x50, 0xea, 0x5c, 0x5a,
	0x4b, 0x8d, 0x6d, 0x6c, 0x27, 0x12, 0xaf, 0xc3, 0x93, 0x22, 0xdb, 0x09, 0x34, 0x4c, 0xb9, 0xe7,
	0xdc, 0x5f, 0x8e, 0x7c, 0x6c, 0x98, 0x7e, 0x54, 0x68, 0x3e, 0x97, 0xda, 0x28, 0xa7, 0xc8, 0x28,
	0x7c, 0x6c, 0xfa, 0x02, 0x57, 0x5b, 0x27, 0xd7, 0x87, 0x5c, 0x48, 0xae, 0x0a, 0x7c, 0xf6, 0x48,
	0x86, 0x56, 0x2b, 0x69, 0x91, 0xdc, 0x01, 0xf0, 0x76, 0x63, 0x69, 0x32, 0xef, 0x2f, 0xa6, 0x2b,
	0x1a, 0x03, 0xec, 0xf2, 0xf4, 0xb7, 0x8d, 0x7c, 0x57, 0xd9, 0x09, 0x9b, 0x7e, 0x25, 0x70, 0xf6,
	0x17, 0x20, 0x04, 0x06, 0x32, 0x2f, 0x91, 0x26, 0xf3, 0x64, 0x31, 0xc9, 0xc2, 0x4c, 0x28, 0x8c,
	0x6b, 0x34, 0x56, 0x28, 0x49, 0x7b, 0xc1, 0x6e, 0xa5, 0xa7, 0x75, 0xee, 0x0e, 0xb4, 0x1f, 0x69,
	0x3f, 0x93, 0x73, 0x18, 0x0a, 0xa9, 0x2b, 0x47, 0x07, 0xc1, 0x8c, 0xc2, 0x93, 0x68, 0x39, 0xa7,
	0xc3, 0x48, 0xfa, 0xd9, 0x7b, 0xb5, 0xf7, 0x46, 0xd1, 0xf3, 0x33, 0x99, 0x41, 0x4f, 0x14, 0x74,
	0x3c, 0x4f, 0x16, 0xff, 0xb3, 0x9e, 0x28, 0xd2, 0x47, 0xb8, 0x8c, 0x67, 0x94, 0x12, 0x8f, 0xdd,
	0xe6, 0x2b, 0xf8, 0xc7, 0xa3, 0xdf, 0xf6, 0xbe, 0xe8, 0xf6, 0xf6, 0xab, 0xd0, 0xfa, 0x87, 0x4b,
	0x19, 0xcc, 0xba, 0x3b, 0x72, 0x1d, 0xee, 0xcf, 0xcb, 0x37, 0x51, 0x34, 0xb5, 0x27, 0x8d, 0xb3,
	0x29, 0x1e, 0x9e, 0x60, 0xda, 0x64, 0x6a, 0x44, 0xf3, 0x7a, 0xbf, 0x17, 0xee, 0x50, 0xed, 0x96,
	0x5c, 0x95, 0x4c, 0xe7, 0xc7, 0x23, 0x3a, 0x25, 0x91, 0xed, 0xd5, 0xcd, 0xaf, 0xe0, 0xca, 0x20,
	0xab, 0xcb, 0xb5, 0x92, 0xce, 0xe4, 0xdc, 0x6d, 0xab, 0x1d, 0x8b, 0x09, 0xcc, 0x27, 0xec, 0xe2,
	0xa3, 0xde, 0x7e, 0x07, 0x00, 0x00, 0xff, 0xff, 0xba, 0x62, 0x3b, 0x20, 0xea, 0x01, 0x00, 0x00,
}
