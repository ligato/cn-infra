// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: phonebook.proto

package phonebook

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Contact struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name        string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Company     string `protobuf:"bytes,2,opt,name=company,proto3" json:"company,omitempty"`
	Phonenumber string `protobuf:"bytes,3,opt,name=phonenumber,proto3" json:"phonenumber,omitempty"`
}

func (x *Contact) Reset() {
	*x = Contact{}
	if protoimpl.UnsafeEnabled {
		mi := &file_phonebook_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Contact) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Contact) ProtoMessage() {}

func (x *Contact) ProtoReflect() protoreflect.Message {
	mi := &file_phonebook_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Contact.ProtoReflect.Descriptor instead.
func (*Contact) Descriptor() ([]byte, []int) {
	return file_phonebook_proto_rawDescGZIP(), []int{0}
}

func (x *Contact) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Contact) GetCompany() string {
	if x != nil {
		return x.Company
	}
	return ""
}

func (x *Contact) GetPhonenumber() string {
	if x != nil {
		return x.Phonenumber
	}
	return ""
}

var File_phonebook_proto protoreflect.FileDescriptor

var file_phonebook_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x70, 0x68, 0x6f, 0x6e, 0x65, 0x62, 0x6f, 0x6f, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x09, 0x70, 0x68, 0x6f, 0x6e, 0x65, 0x62, 0x6f, 0x6f, 0x6b, 0x22, 0x59, 0x0a, 0x07,
	0x43, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63,
	0x6f, 0x6d, 0x70, 0x61, 0x6e, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f,
	0x6d, 0x70, 0x61, 0x6e, 0x79, 0x12, 0x20, 0x0a, 0x0b, 0x70, 0x68, 0x6f, 0x6e, 0x65, 0x6e, 0x75,
	0x6d, 0x62, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x70, 0x68, 0x6f, 0x6e,
	0x65, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x42, 0x3c, 0x5a, 0x3a, 0x67, 0x6f, 0x2e, 0x6c, 0x69,
	0x67, 0x61, 0x74, 0x6f, 0x2e, 0x69, 0x6f, 0x2f, 0x63, 0x6e, 0x2d, 0x69, 0x6e, 0x66, 0x72, 0x61,
	0x2f, 0x76, 0x32, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x73, 0x2f, 0x65, 0x74, 0x63,
	0x64, 0x2d, 0x6c, 0x69, 0x62, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2f, 0x70, 0x68, 0x6f, 0x6e,
	0x65, 0x62, 0x6f, 0x6f, 0x6b, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_phonebook_proto_rawDescOnce sync.Once
	file_phonebook_proto_rawDescData = file_phonebook_proto_rawDesc
)

func file_phonebook_proto_rawDescGZIP() []byte {
	file_phonebook_proto_rawDescOnce.Do(func() {
		file_phonebook_proto_rawDescData = protoimpl.X.CompressGZIP(file_phonebook_proto_rawDescData)
	})
	return file_phonebook_proto_rawDescData
}

var file_phonebook_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_phonebook_proto_goTypes = []interface{}{
	(*Contact)(nil), // 0: phonebook.Contact
}
var file_phonebook_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_phonebook_proto_init() }
func file_phonebook_proto_init() {
	if File_phonebook_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_phonebook_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Contact); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_phonebook_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_phonebook_proto_goTypes,
		DependencyIndexes: file_phonebook_proto_depIdxs,
		MessageInfos:      file_phonebook_proto_msgTypes,
	}.Build()
	File_phonebook_proto = out.File
	file_phonebook_proto_rawDesc = nil
	file_phonebook_proto_goTypes = nil
	file_phonebook_proto_depIdxs = nil
}
