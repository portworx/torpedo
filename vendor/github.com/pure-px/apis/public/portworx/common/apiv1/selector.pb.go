// Please use the following editor setup for this file:
// Tab size=2; Tabs as spaces; Clean up trailing whitepsace
//
// In vim add: au FileType proto setl sw=2 ts=2 expandtab list
//
// In vscode install vscode-proto3 extension and add this to your settings.json:
//    "[proto3]": {
//        "editor.tabSize": 2,
//        "editor.insertSpaces": true,
//        "editor.rulers": [80],
//        "editor.detectIndentation": true,
//        "files.trimTrailingWhitespace": true
//    }
//

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v4.25.1
// source: public/portworx/common/apiv1/selector.proto

package common

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

// RespData provides flags which provides info about the fields that should be populated in the response.
type RespData int32

const (
	// RespData Unspecified. complete resource will be populated.
	RespData_RESP_DATA_UNSPECIFIED RespData = 0
	// only uid, name, labels should be populated.
	RespData_INDEX RespData = 1
	// only meta data should be populated.
	RespData_LITE RespData = 2
	// complete resource should be populated.
	RespData_FULL RespData = 3
)

// Enum value maps for RespData.
var (
	RespData_name = map[int32]string{
		0: "RESP_DATA_UNSPECIFIED",
		1: "INDEX",
		2: "LITE",
		3: "FULL",
	}
	RespData_value = map[string]int32{
		"RESP_DATA_UNSPECIFIED": 0,
		"INDEX":                 1,
		"LITE":                  2,
		"FULL":                  3,
	}
)

func (x RespData) Enum() *RespData {
	p := new(RespData)
	*p = x
	return p
}

func (x RespData) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (RespData) Descriptor() protoreflect.EnumDescriptor {
	return file_public_portworx_common_apiv1_selector_proto_enumTypes[0].Descriptor()
}

func (RespData) Type() protoreflect.EnumType {
	return &file_public_portworx_common_apiv1_selector_proto_enumTypes[0]
}

func (x RespData) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use RespData.Descriptor instead.
func (RespData) EnumDescriptor() ([]byte, []int) {
	return file_public_portworx_common_apiv1_selector_proto_rawDescGZIP(), []int{0}
}

// Operator specifies the relationship between the provided (key,value) pairs in the response.
type Selector_Operator int32

const (
	// Unspecified, do not use.
	Selector_OPERATOR_UNSPECIFIED Selector_Operator = 0
	// IN specifies that the key should be associated with atleast 1 of the element in value list.
	Selector_IN Selector_Operator = 1
	// NOT_IN specifies that the key should not be associated with any of the element in value list.
	Selector_NOT_IN Selector_Operator = 2
)

// Enum value maps for Selector_Operator.
var (
	Selector_Operator_name = map[int32]string{
		0: "OPERATOR_UNSPECIFIED",
		1: "IN",
		2: "NOT_IN",
	}
	Selector_Operator_value = map[string]int32{
		"OPERATOR_UNSPECIFIED": 0,
		"IN":                   1,
		"NOT_IN":               2,
	}
)

func (x Selector_Operator) Enum() *Selector_Operator {
	p := new(Selector_Operator)
	*p = x
	return p
}

func (x Selector_Operator) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Selector_Operator) Descriptor() protoreflect.EnumDescriptor {
	return file_public_portworx_common_apiv1_selector_proto_enumTypes[1].Descriptor()
}

func (Selector_Operator) Type() protoreflect.EnumType {
	return &file_public_portworx_common_apiv1_selector_proto_enumTypes[1]
}

func (x Selector_Operator) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Selector_Operator.Descriptor instead.
func (Selector_Operator) EnumDescriptor() ([]byte, []int) {
	return file_public_portworx_common_apiv1_selector_proto_rawDescGZIP(), []int{0, 0}
}

// Selector is used to query resources using the associated labels or field names.
type Selector struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// FilterList is the list of all filters that should be applied.
	Filters []*Selector_Filter `protobuf:"bytes,1,rep,name=filters,proto3" json:"filters,omitempty"`
}

func (x *Selector) Reset() {
	*x = Selector{}
	if protoimpl.UnsafeEnabled {
		mi := &file_public_portworx_common_apiv1_selector_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Selector) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Selector) ProtoMessage() {}

func (x *Selector) ProtoReflect() protoreflect.Message {
	mi := &file_public_portworx_common_apiv1_selector_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Selector.ProtoReflect.Descriptor instead.
func (*Selector) Descriptor() ([]byte, []int) {
	return file_public_portworx_common_apiv1_selector_proto_rawDescGZIP(), []int{0}
}

func (x *Selector) GetFilters() []*Selector_Filter {
	if x != nil {
		return x.Filters
	}
	return nil
}

// ResourceSelector is used to query resources using the associated infra resources.
type ResourceSelector struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Infra_resource_filters is the list of all filters that should be applied to fetch data related to infra resource.
	// Each filter will have AND relationship.
	InfraResourceFilters []*ResourceSelector_ResourceFilter `protobuf:"bytes,1,rep,name=infra_resource_filters,json=infraResourceFilters,proto3" json:"infra_resource_filters,omitempty"`
}

func (x *ResourceSelector) Reset() {
	*x = ResourceSelector{}
	if protoimpl.UnsafeEnabled {
		mi := &file_public_portworx_common_apiv1_selector_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResourceSelector) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceSelector) ProtoMessage() {}

func (x *ResourceSelector) ProtoReflect() protoreflect.Message {
	mi := &file_public_portworx_common_apiv1_selector_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceSelector.ProtoReflect.Descriptor instead.
func (*ResourceSelector) Descriptor() ([]byte, []int) {
	return file_public_portworx_common_apiv1_selector_proto_rawDescGZIP(), []int{1}
}

func (x *ResourceSelector) GetInfraResourceFilters() []*ResourceSelector_ResourceFilter {
	if x != nil {
		return x.InfraResourceFilters
	}
	return nil
}

// Filter for a given key.
type Selector_Filter struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Key of key,value pair against which filtering needs to be performs.
	Key string `protobuf:"bytes,101,opt,name=key,proto3" json:"key,omitempty"`
	// Op provides the relationship between the key,value pair in the resp element(s).
	Op Selector_Operator `protobuf:"varint,102,opt,name=op,proto3,enum=public.portworx.common.v1.Selector_Operator" json:"op,omitempty"`
	// Value of key,value pair against which filtering needs to be performs if operator is EXIST, value should be an empty array.
	Values []string `protobuf:"bytes,103,rep,name=values,proto3" json:"values,omitempty"`
}

func (x *Selector_Filter) Reset() {
	*x = Selector_Filter{}
	if protoimpl.UnsafeEnabled {
		mi := &file_public_portworx_common_apiv1_selector_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Selector_Filter) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Selector_Filter) ProtoMessage() {}

func (x *Selector_Filter) ProtoReflect() protoreflect.Message {
	mi := &file_public_portworx_common_apiv1_selector_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Selector_Filter.ProtoReflect.Descriptor instead.
func (*Selector_Filter) Descriptor() ([]byte, []int) {
	return file_public_portworx_common_apiv1_selector_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Selector_Filter) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Selector_Filter) GetOp() Selector_Operator {
	if x != nil {
		return x.Op
	}
	return Selector_OPERATOR_UNSPECIFIED
}

func (x *Selector_Filter) GetValues() []string {
	if x != nil {
		return x.Values
	}
	return nil
}

// ResourceFilter is filter for a given resource type.
type ResourceSelector_ResourceFilter struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Key of key,value pair against which filtering needs to be performs based on associated infra resource type.
	ResourceType InfraResource_Type `protobuf:"varint,101,opt,name=resource_type,json=resourceType,proto3,enum=public.portworx.common.v1.InfraResource_Type" json:"resource_type,omitempty"`
	// Op provides the relationship between the key,value pair in the resp element(s).
	Op Selector_Operator `protobuf:"varint,102,opt,name=op,proto3,enum=public.portworx.common.v1.Selector_Operator" json:"op,omitempty"`
	// Value of key,value pair against which filtering needs to be performs.
	Values []string `protobuf:"bytes,103,rep,name=values,proto3" json:"values,omitempty"`
}

func (x *ResourceSelector_ResourceFilter) Reset() {
	*x = ResourceSelector_ResourceFilter{}
	if protoimpl.UnsafeEnabled {
		mi := &file_public_portworx_common_apiv1_selector_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResourceSelector_ResourceFilter) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceSelector_ResourceFilter) ProtoMessage() {}

func (x *ResourceSelector_ResourceFilter) ProtoReflect() protoreflect.Message {
	mi := &file_public_portworx_common_apiv1_selector_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceSelector_ResourceFilter.ProtoReflect.Descriptor instead.
func (*ResourceSelector_ResourceFilter) Descriptor() ([]byte, []int) {
	return file_public_portworx_common_apiv1_selector_proto_rawDescGZIP(), []int{1, 0}
}

func (x *ResourceSelector_ResourceFilter) GetResourceType() InfraResource_Type {
	if x != nil {
		return x.ResourceType
	}
	return InfraResource_TYPE_UNSPECIFIED
}

func (x *ResourceSelector_ResourceFilter) GetOp() Selector_Operator {
	if x != nil {
		return x.Op
	}
	return Selector_OPERATOR_UNSPECIFIED
}

func (x *ResourceSelector_ResourceFilter) GetValues() []string {
	if x != nil {
		return x.Values
	}
	return nil
}

var File_public_portworx_common_apiv1_selector_proto protoreflect.FileDescriptor

var file_public_portworx_common_apiv1_selector_proto_rawDesc = []byte{
	0x0a, 0x2b, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x2f, 0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f, 0x72,
	0x78, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69, 0x76, 0x31, 0x2f, 0x73,
	0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x19, 0x70,
	0x75, 0x62, 0x6c, 0x69, 0x63, 0x2e, 0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f, 0x72, 0x78, 0x2e, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x1a, 0x33, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63,
	0x2f, 0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f, 0x72, 0x78, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e,
	0x2f, 0x61, 0x70, 0x69, 0x76, 0x31, 0x2f, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x72,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xfc, 0x01,
	0x0a, 0x08, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x12, 0x44, 0x0a, 0x07, 0x66, 0x69,
	0x6c, 0x74, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2a, 0x2e, 0x70, 0x75,
	0x62, 0x6c, 0x69, 0x63, 0x2e, 0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f, 0x72, 0x78, 0x2e, 0x63, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72,
	0x2e, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x52, 0x07, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73,
	0x1a, 0x70, 0x0a, 0x06, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65,
	0x79, 0x18, 0x65, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x3c, 0x0a, 0x02,
	0x6f, 0x70, 0x18, 0x66, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2c, 0x2e, 0x70, 0x75, 0x62, 0x6c, 0x69,
	0x63, 0x2e, 0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f, 0x72, 0x78, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x4f, 0x70,
	0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x52, 0x02, 0x6f, 0x70, 0x12, 0x16, 0x0a, 0x06, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x73, 0x18, 0x67, 0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x73, 0x22, 0x38, 0x0a, 0x08, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x12, 0x18,
	0x0a, 0x14, 0x4f, 0x50, 0x45, 0x52, 0x41, 0x54, 0x4f, 0x52, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45,
	0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x06, 0x0a, 0x02, 0x49, 0x4e, 0x10, 0x01,
	0x12, 0x0a, 0x0a, 0x06, 0x4e, 0x4f, 0x54, 0x5f, 0x49, 0x4e, 0x10, 0x02, 0x22, 0xc1, 0x02, 0x0a,
	0x10, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f,
	0x72, 0x12, 0x70, 0x0a, 0x16, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x5f, 0x72, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x5f, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x3a, 0x2e, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x2e, 0x70, 0x6f, 0x72, 0x74, 0x77,
	0x6f, 0x72, 0x78, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x52,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x52, 0x14, 0x69,
	0x6e, 0x66, 0x72, 0x61, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x46, 0x69, 0x6c, 0x74,
	0x65, 0x72, 0x73, 0x1a, 0xba, 0x01, 0x0a, 0x0e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x12, 0x52, 0x0a, 0x0d, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x65, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2d, 0x2e,
	0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x2e, 0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f, 0x72, 0x78, 0x2e,
	0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x6e, 0x66, 0x72, 0x61, 0x52,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0c, 0x72, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x3c, 0x0a, 0x02, 0x6f, 0x70,
	0x18, 0x66, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2c, 0x2e, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x2e,
	0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f, 0x72, 0x78, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e,
	0x76, 0x31, 0x2e, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x4f, 0x70, 0x65, 0x72,
	0x61, 0x74, 0x6f, 0x72, 0x52, 0x02, 0x6f, 0x70, 0x12, 0x16, 0x0a, 0x06, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x73, 0x18, 0x67, 0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73,
	0x2a, 0x44, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x70, 0x44, 0x61, 0x74, 0x61, 0x12, 0x19, 0x0a, 0x15,
	0x52, 0x45, 0x53, 0x50, 0x5f, 0x44, 0x41, 0x54, 0x41, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43,
	0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x09, 0x0a, 0x05, 0x49, 0x4e, 0x44, 0x45, 0x58,
	0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x4c, 0x49, 0x54, 0x45, 0x10, 0x02, 0x12, 0x08, 0x0a, 0x04,
	0x46, 0x55, 0x4c, 0x4c, 0x10, 0x03, 0x42, 0x6d, 0x0a, 0x1d, 0x63, 0x6f, 0x6d, 0x2e, 0x70, 0x75,
	0x62, 0x6c, 0x69, 0x63, 0x2e, 0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f, 0x72, 0x78, 0x2e, 0x63, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x42, 0x0d, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f,
	0x72, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x3b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x75, 0x72, 0x65, 0x2d, 0x70, 0x78, 0x2f, 0x61, 0x70, 0x69,
	0x73, 0x2f, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x2f, 0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f, 0x72,
	0x78, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69, 0x76, 0x31, 0x3b, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_public_portworx_common_apiv1_selector_proto_rawDescOnce sync.Once
	file_public_portworx_common_apiv1_selector_proto_rawDescData = file_public_portworx_common_apiv1_selector_proto_rawDesc
)

func file_public_portworx_common_apiv1_selector_proto_rawDescGZIP() []byte {
	file_public_portworx_common_apiv1_selector_proto_rawDescOnce.Do(func() {
		file_public_portworx_common_apiv1_selector_proto_rawDescData = protoimpl.X.CompressGZIP(file_public_portworx_common_apiv1_selector_proto_rawDescData)
	})
	return file_public_portworx_common_apiv1_selector_proto_rawDescData
}

var file_public_portworx_common_apiv1_selector_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_public_portworx_common_apiv1_selector_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_public_portworx_common_apiv1_selector_proto_goTypes = []interface{}{
	(RespData)(0),                           // 0: public.portworx.common.v1.RespData
	(Selector_Operator)(0),                  // 1: public.portworx.common.v1.Selector.Operator
	(*Selector)(nil),                        // 2: public.portworx.common.v1.Selector
	(*ResourceSelector)(nil),                // 3: public.portworx.common.v1.ResourceSelector
	(*Selector_Filter)(nil),                 // 4: public.portworx.common.v1.Selector.Filter
	(*ResourceSelector_ResourceFilter)(nil), // 5: public.portworx.common.v1.ResourceSelector.ResourceFilter
	(InfraResource_Type)(0),                 // 6: public.portworx.common.v1.InfraResource.Type
}
var file_public_portworx_common_apiv1_selector_proto_depIdxs = []int32{
	4, // 0: public.portworx.common.v1.Selector.filters:type_name -> public.portworx.common.v1.Selector.Filter
	5, // 1: public.portworx.common.v1.ResourceSelector.infra_resource_filters:type_name -> public.portworx.common.v1.ResourceSelector.ResourceFilter
	1, // 2: public.portworx.common.v1.Selector.Filter.op:type_name -> public.portworx.common.v1.Selector.Operator
	6, // 3: public.portworx.common.v1.ResourceSelector.ResourceFilter.resource_type:type_name -> public.portworx.common.v1.InfraResource.Type
	1, // 4: public.portworx.common.v1.ResourceSelector.ResourceFilter.op:type_name -> public.portworx.common.v1.Selector.Operator
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_public_portworx_common_apiv1_selector_proto_init() }
func file_public_portworx_common_apiv1_selector_proto_init() {
	if File_public_portworx_common_apiv1_selector_proto != nil {
		return
	}
	file_public_portworx_common_apiv1_platformresource_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_public_portworx_common_apiv1_selector_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Selector); i {
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
		file_public_portworx_common_apiv1_selector_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResourceSelector); i {
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
		file_public_portworx_common_apiv1_selector_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Selector_Filter); i {
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
		file_public_portworx_common_apiv1_selector_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResourceSelector_ResourceFilter); i {
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
			RawDescriptor: file_public_portworx_common_apiv1_selector_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_public_portworx_common_apiv1_selector_proto_goTypes,
		DependencyIndexes: file_public_portworx_common_apiv1_selector_proto_depIdxs,
		EnumInfos:         file_public_portworx_common_apiv1_selector_proto_enumTypes,
		MessageInfos:      file_public_portworx_common_apiv1_selector_proto_msgTypes,
	}.Build()
	File_public_portworx_common_apiv1_selector_proto = out.File
	file_public_portworx_common_apiv1_selector_proto_rawDesc = nil
	file_public_portworx_common_apiv1_selector_proto_goTypes = nil
	file_public_portworx_common_apiv1_selector_proto_depIdxs = nil
}