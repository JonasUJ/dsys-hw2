// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.6
// source: tcp/tcp.proto

package tcp

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

type Flag int32

const (
	Flag_NONE   Flag = 0
	Flag_SYN    Flag = 1
	Flag_ACK    Flag = 2
	Flag_SYNACK Flag = 3
	Flag_FIN    Flag = 4
)

// Enum value maps for Flag.
var (
	Flag_name = map[int32]string{
		0: "NONE",
		1: "SYN",
		2: "ACK",
		3: "SYNACK",
		4: "FIN",
	}
	Flag_value = map[string]int32{
		"NONE":   0,
		"SYN":    1,
		"ACK":    2,
		"SYNACK": 3,
		"FIN":    4,
	}
)

func (x Flag) Enum() *Flag {
	p := new(Flag)
	*p = x
	return p
}

func (x Flag) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Flag) Descriptor() protoreflect.EnumDescriptor {
	return file_tcp_tcp_proto_enumTypes[0].Descriptor()
}

func (Flag) Type() protoreflect.EnumType {
	return &file_tcp_tcp_proto_enumTypes[0]
}

func (x Flag) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Flag.Descriptor instead.
func (Flag) EnumDescriptor() ([]byte, []int) {
	return file_tcp_tcp_proto_rawDescGZIP(), []int{0}
}

type Packet struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Flag Flag   `protobuf:"varint,1,opt,name=flag,proto3,enum=tcp.Flag" json:"flag,omitempty"`
	Seq  uint32 `protobuf:"varint,2,opt,name=seq,proto3" json:"seq,omitempty"`
	Ack  uint32 `protobuf:"varint,3,opt,name=ack,proto3" json:"ack,omitempty"`
	Data string `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *Packet) Reset() {
	*x = Packet{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tcp_tcp_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Packet) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Packet) ProtoMessage() {}

func (x *Packet) ProtoReflect() protoreflect.Message {
	mi := &file_tcp_tcp_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Packet.ProtoReflect.Descriptor instead.
func (*Packet) Descriptor() ([]byte, []int) {
	return file_tcp_tcp_proto_rawDescGZIP(), []int{0}
}

func (x *Packet) GetFlag() Flag {
	if x != nil {
		return x.Flag
	}
	return Flag_NONE
}

func (x *Packet) GetSeq() uint32 {
	if x != nil {
		return x.Seq
	}
	return 0
}

func (x *Packet) GetAck() uint32 {
	if x != nil {
		return x.Ack
	}
	return 0
}

func (x *Packet) GetData() string {
	if x != nil {
		return x.Data
	}
	return ""
}

var File_tcp_tcp_proto protoreflect.FileDescriptor

var file_tcp_tcp_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x74, 0x63, 0x70, 0x2f, 0x74, 0x63, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x03, 0x74, 0x63, 0x70, 0x22, 0x5f, 0x0a, 0x06, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x1d,
	0x0a, 0x04, 0x66, 0x6c, 0x61, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x09, 0x2e, 0x74,
	0x63, 0x70, 0x2e, 0x46, 0x6c, 0x61, 0x67, 0x52, 0x04, 0x66, 0x6c, 0x61, 0x67, 0x12, 0x10, 0x0a,
	0x03, 0x73, 0x65, 0x71, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x03, 0x73, 0x65, 0x71, 0x12,
	0x10, 0x0a, 0x03, 0x61, 0x63, 0x6b, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x03, 0x61, 0x63,
	0x6b, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x64, 0x61, 0x74, 0x61, 0x2a, 0x37, 0x0a, 0x04, 0x46, 0x6c, 0x61, 0x67, 0x12, 0x08, 0x0a,
	0x04, 0x4e, 0x4f, 0x4e, 0x45, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x03, 0x53, 0x59, 0x4e, 0x10, 0x01,
	0x12, 0x07, 0x0a, 0x03, 0x41, 0x43, 0x4b, 0x10, 0x02, 0x12, 0x0a, 0x0a, 0x06, 0x53, 0x59, 0x4e,
	0x41, 0x43, 0x4b, 0x10, 0x03, 0x12, 0x07, 0x0a, 0x03, 0x46, 0x49, 0x4e, 0x10, 0x04, 0x32, 0x2e,
	0x0a, 0x03, 0x54, 0x63, 0x70, 0x12, 0x27, 0x0a, 0x07, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74,
	0x12, 0x0b, 0x2e, 0x74, 0x63, 0x70, 0x2e, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x1a, 0x0b, 0x2e,
	0x74, 0x63, 0x70, 0x2e, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x28, 0x01, 0x30, 0x01, 0x42, 0x29,
	0x5a, 0x27, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x4a, 0x6f, 0x6e, 0x61, 0x73, 0x55, 0x4a, 0x2f, 0x64, 0x73, 0x79,
	0x73, 0x2d, 0x68, 0x77, 0x32, 0x3b, 0x74, 0x63, 0x70, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_tcp_tcp_proto_rawDescOnce sync.Once
	file_tcp_tcp_proto_rawDescData = file_tcp_tcp_proto_rawDesc
)

func file_tcp_tcp_proto_rawDescGZIP() []byte {
	file_tcp_tcp_proto_rawDescOnce.Do(func() {
		file_tcp_tcp_proto_rawDescData = protoimpl.X.CompressGZIP(file_tcp_tcp_proto_rawDescData)
	})
	return file_tcp_tcp_proto_rawDescData
}

var file_tcp_tcp_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_tcp_tcp_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_tcp_tcp_proto_goTypes = []interface{}{
	(Flag)(0),      // 0: tcp.Flag
	(*Packet)(nil), // 1: tcp.Packet
}
var file_tcp_tcp_proto_depIdxs = []int32{
	0, // 0: tcp.Packet.flag:type_name -> tcp.Flag
	1, // 1: tcp.Tcp.Connect:input_type -> tcp.Packet
	1, // 2: tcp.Tcp.Connect:output_type -> tcp.Packet
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_tcp_tcp_proto_init() }
func file_tcp_tcp_proto_init() {
	if File_tcp_tcp_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_tcp_tcp_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Packet); i {
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
			RawDescriptor: file_tcp_tcp_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_tcp_tcp_proto_goTypes,
		DependencyIndexes: file_tcp_tcp_proto_depIdxs,
		EnumInfos:         file_tcp_tcp_proto_enumTypes,
		MessageInfos:      file_tcp_tcp_proto_msgTypes,
	}.Build()
	File_tcp_tcp_proto = out.File
	file_tcp_tcp_proto_rawDesc = nil
	file_tcp_tcp_proto_goTypes = nil
	file_tcp_tcp_proto_depIdxs = nil
}