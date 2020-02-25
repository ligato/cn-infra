// Code generated by protoc-gen-go. DO NOT EDIT.
// source: process.proto

// Package process provides a data model for process manager plugin template

package process

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
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Template struct {
	Name                 string            `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Cmd                  string            `protobuf:"bytes,2,opt,name=cmd,proto3" json:"cmd,omitempty"`
	POptions             *TemplatePOptions `protobuf:"bytes,3,opt,name=p_options,json=pOptions,proto3" json:"p_options,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *Template) Reset()         { *m = Template{} }
func (m *Template) String() string { return proto.CompactTextString(m) }
func (*Template) ProtoMessage()    {}
func (*Template) Descriptor() ([]byte, []int) {
	return fileDescriptor_54c4d0e8c0aaf5c3, []int{0}
}

func (m *Template) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Template.Unmarshal(m, b)
}
func (m *Template) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Template.Marshal(b, m, deterministic)
}
func (m *Template) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Template.Merge(m, src)
}
func (m *Template) XXX_Size() int {
	return xxx_messageInfo_Template.Size(m)
}
func (m *Template) XXX_DiscardUnknown() {
	xxx_messageInfo_Template.DiscardUnknown(m)
}

var xxx_messageInfo_Template proto.InternalMessageInfo

func (m *Template) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Template) GetCmd() string {
	if m != nil {
		return m.Cmd
	}
	return ""
}

func (m *Template) GetPOptions() *TemplatePOptions {
	if m != nil {
		return m.POptions
	}
	return nil
}

type TemplatePOptions struct {
	Args                 []string `protobuf:"bytes,1,rep,name=args,proto3" json:"args,omitempty"`
	OutWriter            bool     `protobuf:"varint,2,opt,name=out_writer,json=outWriter,proto3" json:"out_writer,omitempty"`
	ErrWriter            bool     `protobuf:"varint,3,opt,name=err_writer,json=errWriter,proto3" json:"err_writer,omitempty"`
	Restart              int32    `protobuf:"varint,4,opt,name=restart,proto3" json:"restart,omitempty"`
	Detach               bool     `protobuf:"varint,5,opt,name=detach,proto3" json:"detach,omitempty"`
	RunOnStartup         bool     `protobuf:"varint,6,opt,name=run_on_startup,json=runOnStartup,proto3" json:"run_on_startup,omitempty"`
	Notify               bool     `protobuf:"varint,7,opt,name=notify,proto3" json:"notify,omitempty"`
	AutoTerminate        bool     `protobuf:"varint,8,opt,name=auto_terminate,json=autoTerminate,proto3" json:"auto_terminate,omitempty"`
	CpuAffinityMask      string   `protobuf:"bytes,9,opt,name=cpu_affinity_mask,json=cpuAffinityMask,proto3" json:"cpu_affinity_mask,omitempty"`
	CpuAffinityList      string   `protobuf:"bytes,10,opt,name=cpu_affinity_list,json=cpuAffinityList,proto3" json:"cpu_affinity_list,omitempty"`
	CpuAffinityDelay     string   `protobuf:"bytes,11,opt,name=cpu_affinity_delay,json=cpuAffinityDelay,proto3" json:"cpu_affinity_delay,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TemplatePOptions) Reset()         { *m = TemplatePOptions{} }
func (m *TemplatePOptions) String() string { return proto.CompactTextString(m) }
func (*TemplatePOptions) ProtoMessage()    {}
func (*TemplatePOptions) Descriptor() ([]byte, []int) {
	return fileDescriptor_54c4d0e8c0aaf5c3, []int{0, 0}
}

func (m *TemplatePOptions) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TemplatePOptions.Unmarshal(m, b)
}
func (m *TemplatePOptions) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TemplatePOptions.Marshal(b, m, deterministic)
}
func (m *TemplatePOptions) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TemplatePOptions.Merge(m, src)
}
func (m *TemplatePOptions) XXX_Size() int {
	return xxx_messageInfo_TemplatePOptions.Size(m)
}
func (m *TemplatePOptions) XXX_DiscardUnknown() {
	xxx_messageInfo_TemplatePOptions.DiscardUnknown(m)
}

var xxx_messageInfo_TemplatePOptions proto.InternalMessageInfo

func (m *TemplatePOptions) GetArgs() []string {
	if m != nil {
		return m.Args
	}
	return nil
}

func (m *TemplatePOptions) GetOutWriter() bool {
	if m != nil {
		return m.OutWriter
	}
	return false
}

func (m *TemplatePOptions) GetErrWriter() bool {
	if m != nil {
		return m.ErrWriter
	}
	return false
}

func (m *TemplatePOptions) GetRestart() int32 {
	if m != nil {
		return m.Restart
	}
	return 0
}

func (m *TemplatePOptions) GetDetach() bool {
	if m != nil {
		return m.Detach
	}
	return false
}

func (m *TemplatePOptions) GetRunOnStartup() bool {
	if m != nil {
		return m.RunOnStartup
	}
	return false
}

func (m *TemplatePOptions) GetNotify() bool {
	if m != nil {
		return m.Notify
	}
	return false
}

func (m *TemplatePOptions) GetAutoTerminate() bool {
	if m != nil {
		return m.AutoTerminate
	}
	return false
}

func (m *TemplatePOptions) GetCpuAffinityMask() string {
	if m != nil {
		return m.CpuAffinityMask
	}
	return ""
}

func (m *TemplatePOptions) GetCpuAffinityList() string {
	if m != nil {
		return m.CpuAffinityList
	}
	return ""
}

func (m *TemplatePOptions) GetCpuAffinityDelay() string {
	if m != nil {
		return m.CpuAffinityDelay
	}
	return ""
}

func init() {
	proto.RegisterType((*Template)(nil), "process.Template")
	proto.RegisterType((*TemplatePOptions)(nil), "process.Template.pOptions")
}

func init() { proto.RegisterFile("process.proto", fileDescriptor_54c4d0e8c0aaf5c3) }

var fileDescriptor_54c4d0e8c0aaf5c3 = []byte{
	// 327 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x91, 0x4d, 0x4b, 0x33, 0x31,
	0x10, 0xc7, 0xd9, 0x6e, 0x5f, 0x76, 0xa7, 0x4f, 0xfb, 0xd4, 0x1c, 0x24, 0x14, 0x84, 0x45, 0x14,
	0x8a, 0x48, 0x0f, 0x7a, 0xf0, 0x2c, 0x78, 0x54, 0x0a, 0x6b, 0xc1, 0x63, 0x88, 0xdb, 0x54, 0x43,
	0xbb, 0x49, 0x48, 0x26, 0x48, 0x3f, 0xb0, 0x1f, 0xc2, 0x9b, 0x24, 0xbb, 0x2b, 0x4a, 0x6f, 0xff,
	0x97, 0xdf, 0xcc, 0x40, 0x02, 0x13, 0x63, 0x75, 0x25, 0x9c, 0x5b, 0x1a, 0xab, 0x51, 0x93, 0x51,
	0x6b, 0xcf, 0x3f, 0x53, 0xc8, 0xd6, 0xa2, 0x36, 0x7b, 0x8e, 0x82, 0x10, 0xe8, 0x2b, 0x5e, 0x0b,
	0x9a, 0x14, 0xc9, 0x22, 0x2f, 0xa3, 0x26, 0x33, 0x48, 0xab, 0x7a, 0x43, 0x7b, 0x31, 0x0a, 0x92,
	0xdc, 0x41, 0x6e, 0x98, 0x36, 0x28, 0xb5, 0x72, 0x34, 0x2d, 0x92, 0xc5, 0xf8, 0x66, 0xbe, 0xec,
	0xd6, 0x77, 0xbb, 0x96, 0x66, 0xd5, 0x10, 0x65, 0xd6, 0xa9, 0xf9, 0x57, 0x0f, 0x7e, 0x4c, 0xb8,
	0xc5, 0xed, 0x9b, 0xa3, 0x49, 0x91, 0x86, 0x5b, 0x41, 0x93, 0x33, 0x00, 0xed, 0x91, 0x7d, 0x58,
	0x89, 0xc2, 0xc6, 0x93, 0x59, 0x99, 0x6b, 0x8f, 0x2f, 0x31, 0x08, 0xb5, 0xb0, 0xb6, 0xab, 0xd3,
	0xa6, 0x16, 0xd6, 0xb6, 0x35, 0x85, 0x91, 0x15, 0x0e, 0xb9, 0x45, 0xda, 0x2f, 0x92, 0xc5, 0xa0,
	0xec, 0x2c, 0x39, 0x85, 0xe1, 0x46, 0x20, 0xaf, 0xde, 0xe9, 0x20, 0x0e, 0xb5, 0x8e, 0x5c, 0xc0,
	0xd4, 0x7a, 0xc5, 0xb4, 0x62, 0x91, 0xf3, 0x86, 0x0e, 0x63, 0xff, 0xcf, 0x7a, 0xb5, 0x52, 0xcf,
	0x4d, 0x16, 0xa6, 0x95, 0x46, 0xb9, 0x3d, 0xd0, 0x51, 0x33, 0xdd, 0x38, 0x72, 0x09, 0x53, 0xee,
	0x51, 0x33, 0x14, 0xb6, 0x96, 0x8a, 0xa3, 0xa0, 0x59, 0xec, 0x27, 0x21, 0x5d, 0x77, 0x21, 0xb9,
	0x82, 0x93, 0xca, 0x78, 0xc6, 0xb7, 0x5b, 0xa9, 0x24, 0x1e, 0x58, 0xcd, 0xdd, 0x8e, 0xe6, 0xf1,
	0x39, 0xff, 0x57, 0xc6, 0xdf, 0xb7, 0xf9, 0x13, 0x77, 0xbb, 0x23, 0x76, 0x2f, 0x1d, 0x52, 0x38,
	0x62, 0x1f, 0xa5, 0x43, 0x72, 0x0d, 0xe4, 0x0f, 0xbb, 0x11, 0x7b, 0x7e, 0xa0, 0xe3, 0x08, 0xcf,
	0x7e, 0xc1, 0x0f, 0x21, 0x7f, 0x1d, 0xc6, 0x7f, 0xbf, 0xfd, 0x0e, 0x00, 0x00, 0xff, 0xff, 0x82,
	0x49, 0xcc, 0x03, 0x08, 0x02, 0x00, 0x00,
}
