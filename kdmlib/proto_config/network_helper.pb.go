// Code generated by protoc-gen-go. DO NOT EDIT.
// source: network_helper.proto

package proto_config

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

type ContactList struct {
	Contacts             []*ContactList_Contact `protobuf:"bytes,1,rep,name=Contacts,proto3" json:"Contacts,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *ContactList) Reset()         { *m = ContactList{} }
func (m *ContactList) String() string { return proto.CompactTextString(m) }
func (*ContactList) ProtoMessage()    {}
func (*ContactList) Descriptor() ([]byte, []int) {
	return fileDescriptor_3a5406aa8b4ed766, []int{0}
}
func (m *ContactList) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ContactList.Unmarshal(m, b)
}
func (m *ContactList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ContactList.Marshal(b, m, deterministic)
}
func (m *ContactList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ContactList.Merge(m, src)
}
func (m *ContactList) XXX_Size() int {
	return xxx_messageInfo_ContactList.Size(m)
}
func (m *ContactList) XXX_DiscardUnknown() {
	xxx_messageInfo_ContactList.DiscardUnknown(m)
}

var xxx_messageInfo_ContactList proto.InternalMessageInfo

func (m *ContactList) GetContacts() []*ContactList_Contact {
	if m != nil {
		return m.Contacts
	}
	return nil
}

type ContactList_Contact struct {
	ID                   string   `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Address              string   `protobuf:"bytes,2,opt,name=Address,proto3" json:"Address,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ContactList_Contact) Reset()         { *m = ContactList_Contact{} }
func (m *ContactList_Contact) String() string { return proto.CompactTextString(m) }
func (*ContactList_Contact) ProtoMessage()    {}
func (*ContactList_Contact) Descriptor() ([]byte, []int) {
	return fileDescriptor_3a5406aa8b4ed766, []int{0, 0}
}
func (m *ContactList_Contact) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ContactList_Contact.Unmarshal(m, b)
}
func (m *ContactList_Contact) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ContactList_Contact.Marshal(b, m, deterministic)
}
func (m *ContactList_Contact) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ContactList_Contact.Merge(m, src)
}
func (m *ContactList_Contact) XXX_Size() int {
	return xxx_messageInfo_ContactList_Contact.Size(m)
}
func (m *ContactList_Contact) XXX_DiscardUnknown() {
	xxx_messageInfo_ContactList_Contact.DiscardUnknown(m)
}

var xxx_messageInfo_ContactList_Contact proto.InternalMessageInfo

func (m *ContactList_Contact) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *ContactList_Contact) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

type Package struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Type                 string   `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	Message              string   `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
	Time                 string   `protobuf:"bytes,4,opt,name=time,proto3" json:"time,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Package) Reset()         { *m = Package{} }
func (m *Package) String() string { return proto.CompactTextString(m) }
func (*Package) ProtoMessage()    {}
func (*Package) Descriptor() ([]byte, []int) {
	return fileDescriptor_3a5406aa8b4ed766, []int{1}
}
func (m *Package) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Package.Unmarshal(m, b)
}
func (m *Package) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Package.Marshal(b, m, deterministic)
}
func (m *Package) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Package.Merge(m, src)
}
func (m *Package) XXX_Size() int {
	return xxx_messageInfo_Package.Size(m)
}
func (m *Package) XXX_DiscardUnknown() {
	xxx_messageInfo_Package.DiscardUnknown(m)
}

var xxx_messageInfo_Package proto.InternalMessageInfo

func (m *Package) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Package) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *Package) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *Package) GetTime() string {
	if m != nil {
		return m.Time
	}
	return ""
}

func init() {
	proto.RegisterType((*ContactList)(nil), "network_helper.ContactList")
	proto.RegisterType((*ContactList_Contact)(nil), "network_helper.ContactList.Contact")
	proto.RegisterType((*Package)(nil), "network_helper.Package")
}

func init() { proto.RegisterFile("network_helper.proto", fileDescriptor_3a5406aa8b4ed766) }

var fileDescriptor_3a5406aa8b4ed766 = []byte{
	// 192 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0xc9, 0x4b, 0x2d, 0x29,
	0xcf, 0x2f, 0xca, 0x8e, 0xcf, 0x48, 0xcd, 0x29, 0x48, 0x2d, 0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9,
	0x17, 0xe2, 0x43, 0x15, 0x55, 0x6a, 0x66, 0xe4, 0xe2, 0x76, 0xce, 0xcf, 0x2b, 0x49, 0x4c, 0x2e,
	0xf1, 0xc9, 0x2c, 0x2e, 0x11, 0xb2, 0xe7, 0xe2, 0x80, 0x72, 0x8b, 0x25, 0x18, 0x15, 0x98, 0x35,
	0xb8, 0x8d, 0x94, 0xf5, 0xd0, 0x0c, 0x42, 0x52, 0x0e, 0x63, 0x07, 0xc1, 0x35, 0x49, 0x19, 0x73,
	0xb1, 0x43, 0xd9, 0x42, 0x7c, 0x5c, 0x4c, 0x9e, 0x2e, 0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x9c, 0x41,
	0x4c, 0x9e, 0x2e, 0x42, 0x12, 0x5c, 0xec, 0x8e, 0x29, 0x29, 0x45, 0xa9, 0xc5, 0xc5, 0x12, 0x4c,
	0x60, 0x41, 0x18, 0x57, 0x29, 0x9a, 0x8b, 0x3d, 0x20, 0x31, 0x39, 0x3b, 0x31, 0x3d, 0x15, 0xa4,
	0x29, 0x33, 0x05, 0xa6, 0x29, 0x33, 0x45, 0x48, 0x88, 0x8b, 0xa5, 0xa4, 0xb2, 0x20, 0x15, 0xaa,
	0x03, 0xcc, 0x06, 0x19, 0x94, 0x9b, 0x5a, 0x5c, 0x9c, 0x98, 0x9e, 0x2a, 0xc1, 0x0c, 0x31, 0x08,
	0xca, 0x05, 0xab, 0xce, 0xcc, 0x4d, 0x95, 0x60, 0x81, 0xaa, 0xce, 0xcc, 0x4d, 0x4d, 0x62, 0x03,
	0xfb, 0xdc, 0x18, 0x10, 0x00, 0x00, 0xff, 0xff, 0x3f, 0xf2, 0xda, 0x4e, 0x11, 0x01, 0x00, 0x00,
}
