// Code generated by protoc-gen-go.
// source: endpoint.proto
// DO NOT EDIT!

/*
Package boltpb is a generated protocol buffer package.

It is generated from these files:
	endpoint.proto

It has these top-level messages:
	Endpoint
*/
package boltpb

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

type Endpoint struct {
	// @inject_tag: storm:"id"
	Path    string `protobuf:"bytes,1,opt,name=Path" json:"Path,omitempty" storm:"id"`
	Backend string `protobuf:"bytes,2,opt,name=Backend" json:"Backend,omitempty"`
}

func (m *Endpoint) Reset()                    { *m = Endpoint{} }
func (m *Endpoint) String() string            { return proto.CompactTextString(m) }
func (*Endpoint) ProtoMessage()               {}
func (*Endpoint) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func init() {
	proto.RegisterType((*Endpoint)(nil), "boltpb.Endpoint")
}

func init() { proto.RegisterFile("endpoint.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 93 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4b, 0xcd, 0x4b, 0x29,
	0xc8, 0xcf, 0xcc, 0x2b, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x4b, 0xca, 0xcf, 0x29,
	0x29, 0x48, 0x52, 0xb2, 0xe0, 0xe2, 0x70, 0x85, 0xca, 0x08, 0x09, 0x71, 0xb1, 0x04, 0x24, 0x96,
	0x64, 0x48, 0x30, 0x2a, 0x30, 0x6a, 0x70, 0x06, 0x81, 0xd9, 0x42, 0x12, 0x5c, 0xec, 0x4e, 0x89,
	0xc9, 0xd9, 0xa9, 0x79, 0x29, 0x12, 0x4c, 0x60, 0x61, 0x18, 0x37, 0x89, 0x0d, 0x6c, 0x90, 0x31,
	0x20, 0x00, 0x00, 0xff, 0xff, 0x39, 0x06, 0x87, 0x46, 0x5a, 0x00, 0x00, 0x00,
}