// Code generated by protoc-gen-go. DO NOT EDIT.
// source: item.proto

/*
Package internal is a generated protocol buffer package.

It is generated from these files:
	item.proto
	bucket.proto

It has these top-level messages:
	Item
	Bucket
*/
package internal

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Item struct {
	// @inject_tag: storm:"id,increment"
	Id int64 `protobuf:"varint,1,opt,name=id" json:"id,omitempty" storm:"id,increment"`
	// @inject_tag: storm:"unique"
	Key  string `protobuf:"bytes,2,opt,name=key" json:"key,omitempty" storm:"unique"`
	Data []byte `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *Item) Reset()                    { *m = Item{} }
func (m *Item) String() string            { return proto.CompactTextString(m) }
func (*Item) ProtoMessage()               {}
func (*Item) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Item) GetId() int64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *Item) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *Item) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
	proto.RegisterType((*Item)(nil), "internal.Item")
}

func init() { proto.RegisterFile("item.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 108 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0xca, 0x2c, 0x49, 0xcd,
	0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0xc8, 0xcc, 0x2b, 0x49, 0x2d, 0xca, 0x4b, 0xcc,
	0x51, 0xb2, 0xe1, 0x62, 0xf1, 0x2c, 0x49, 0xcd, 0x15, 0xe2, 0xe3, 0x62, 0xca, 0x4c, 0x91, 0x60,
	0x54, 0x60, 0xd4, 0x60, 0x0e, 0x62, 0xca, 0x4c, 0x11, 0x12, 0xe0, 0x62, 0xce, 0x4e, 0xad, 0x94,
	0x60, 0x52, 0x60, 0xd4, 0xe0, 0x0c, 0x02, 0x31, 0x85, 0x84, 0xb8, 0x58, 0x52, 0x12, 0x4b, 0x12,
	0x25, 0x98, 0x15, 0x18, 0x35, 0x78, 0x82, 0xc0, 0xec, 0x24, 0x36, 0xb0, 0x71, 0xc6, 0x80, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x30, 0x7f, 0x44, 0xba, 0x5c, 0x00, 0x00, 0x00,
}