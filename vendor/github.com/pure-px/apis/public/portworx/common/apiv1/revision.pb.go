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

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v4.25.1
// source: public/portworx/common/apiv1/revision.proto

package common

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	structpb "google.golang.org/protobuf/types/known/structpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Revision holds the template schema along with version details.
type Revision struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Metadata of the revision.
	Meta *Meta `protobuf:"bytes,1,opt,name=meta,proto3" json:"meta,omitempty"`
	// Info of the revision.
	Info *RevisionInfo `protobuf:"bytes,2,opt,name=info,proto3" json:"info,omitempty"`
}

func (x *Revision) Reset() {
	*x = Revision{}
	if protoimpl.UnsafeEnabled {
		mi := &file_public_portworx_common_apiv1_revision_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Revision) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Revision) ProtoMessage() {}

func (x *Revision) ProtoReflect() protoreflect.Message {
	mi := &file_public_portworx_common_apiv1_revision_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Revision.ProtoReflect.Descriptor instead.
func (*Revision) Descriptor() ([]byte, []int) {
	return file_public_portworx_common_apiv1_revision_proto_rawDescGZIP(), []int{0}
}

func (x *Revision) GetMeta() *Meta {
	if x != nil {
		return x.Meta
	}
	return nil
}

func (x *Revision) GetInfo() *RevisionInfo {
	if x != nil {
		return x.Info
	}
	return nil
}

// RevisionInfo contains info.
type RevisionInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Semantic version of the revision: 1.2 (major.minor - patch version not required).
	SemanticVersion string `protobuf:"bytes,1,opt,name=semantic_version,json=semanticVersion,proto3" json:"semantic_version,omitempty"`
	// Whether this revision has been deprecated.
	Deprecated bool `protobuf:"varint,2,opt,name=deprecated,proto3" json:"deprecated,omitempty"`
	// Schema of the revision, if schema is backward compatible, update the revision, else upgrade.
	Schema *structpb.Struct `protobuf:"bytes,3,opt,name=schema,proto3" json:"schema,omitempty"`
}

func (x *RevisionInfo) Reset() {
	*x = RevisionInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_public_portworx_common_apiv1_revision_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RevisionInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RevisionInfo) ProtoMessage() {}

func (x *RevisionInfo) ProtoReflect() protoreflect.Message {
	mi := &file_public_portworx_common_apiv1_revision_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RevisionInfo.ProtoReflect.Descriptor instead.
func (*RevisionInfo) Descriptor() ([]byte, []int) {
	return file_public_portworx_common_apiv1_revision_proto_rawDescGZIP(), []int{1}
}

func (x *RevisionInfo) GetSemanticVersion() string {
	if x != nil {
		return x.SemanticVersion
	}
	return ""
}

func (x *RevisionInfo) GetDeprecated() bool {
	if x != nil {
		return x.Deprecated
	}
	return false
}

func (x *RevisionInfo) GetSchema() *structpb.Struct {
	if x != nil {
		return x.Schema
	}
	return nil
}

// GetRevisionRequest is the request body to get a revision.
type GetRevisionRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Get the revision by uid or name_semantic_version.
	//
	// Types that are assignable to GetBy:
	//
	//	*GetRevisionRequest_Uid
	//	*GetRevisionRequest_NameSemanticVersion_
	GetBy isGetRevisionRequest_GetBy `protobuf_oneof:"get_by"`
}

func (x *GetRevisionRequest) Reset() {
	*x = GetRevisionRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_public_portworx_common_apiv1_revision_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetRevisionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRevisionRequest) ProtoMessage() {}

func (x *GetRevisionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_public_portworx_common_apiv1_revision_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRevisionRequest.ProtoReflect.Descriptor instead.
func (*GetRevisionRequest) Descriptor() ([]byte, []int) {
	return file_public_portworx_common_apiv1_revision_proto_rawDescGZIP(), []int{2}
}

func (m *GetRevisionRequest) GetGetBy() isGetRevisionRequest_GetBy {
	if m != nil {
		return m.GetBy
	}
	return nil
}

func (x *GetRevisionRequest) GetUid() string {
	if x, ok := x.GetGetBy().(*GetRevisionRequest_Uid); ok {
		return x.Uid
	}
	return ""
}

func (x *GetRevisionRequest) GetNameSemanticVersion() *GetRevisionRequest_NameSemanticVersion {
	if x, ok := x.GetGetBy().(*GetRevisionRequest_NameSemanticVersion_); ok {
		return x.NameSemanticVersion
	}
	return nil
}

type isGetRevisionRequest_GetBy interface {
	isGetRevisionRequest_GetBy()
}

type GetRevisionRequest_Uid struct {
	// UID of the revision.
	// (-- api-linter: core::0148::uid-format=disabled
	//
	//	aip.dev/not-precedent: We need to do this because of prefix. --)
	Uid string `protobuf:"bytes,1,opt,name=uid,proto3,oneof"`
}

type GetRevisionRequest_NameSemanticVersion_ struct {
	// Name and semantic version of the revision.
	NameSemanticVersion *GetRevisionRequest_NameSemanticVersion `protobuf:"bytes,2,opt,name=name_semantic_version,json=nameSemanticVersion,proto3,oneof"`
}

func (*GetRevisionRequest_Uid) isGetRevisionRequest_GetBy() {}

func (*GetRevisionRequest_NameSemanticVersion_) isGetRevisionRequest_GetBy() {}

// Request parameters for listing revisions.
type ListRevisionsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Filtering list of revisions based on the provided column filters.
	FieldSelector *Selector `protobuf:"bytes,1,opt,name=field_selector,json=fieldSelector,proto3" json:"field_selector,omitempty"`
	// Sort parameters for listing revisions.
	Sort *Sort `protobuf:"bytes,2,opt,name=sort,proto3" json:"sort,omitempty"`
	// Pagination parameters for listing revisions.
	Pagination *PageBasedPaginationRequest `protobuf:"bytes,3,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

func (x *ListRevisionsRequest) Reset() {
	*x = ListRevisionsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_public_portworx_common_apiv1_revision_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListRevisionsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListRevisionsRequest) ProtoMessage() {}

func (x *ListRevisionsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_public_portworx_common_apiv1_revision_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListRevisionsRequest.ProtoReflect.Descriptor instead.
func (*ListRevisionsRequest) Descriptor() ([]byte, []int) {
	return file_public_portworx_common_apiv1_revision_proto_rawDescGZIP(), []int{3}
}

func (x *ListRevisionsRequest) GetFieldSelector() *Selector {
	if x != nil {
		return x.FieldSelector
	}
	return nil
}

func (x *ListRevisionsRequest) GetSort() *Sort {
	if x != nil {
		return x.Sort
	}
	return nil
}

func (x *ListRevisionsRequest) GetPagination() *PageBasedPaginationRequest {
	if x != nil {
		return x.Pagination
	}
	return nil
}

// Revisions listing response.
type ListRevisionsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Revisions is the list of revisions.
	Revisions []*Revision `protobuf:"bytes,1,rep,name=revisions,proto3" json:"revisions,omitempty"`
	// Pagination metadata for this response.
	// (-- api-linter: core::0132::response-unknown-fields=disabled
	//
	//	aip.dev/not-precedent: We need this field for pagination. --)
	Pagination *PageBasedPaginationResponse `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

func (x *ListRevisionsResponse) Reset() {
	*x = ListRevisionsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_public_portworx_common_apiv1_revision_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListRevisionsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListRevisionsResponse) ProtoMessage() {}

func (x *ListRevisionsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_public_portworx_common_apiv1_revision_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListRevisionsResponse.ProtoReflect.Descriptor instead.
func (*ListRevisionsResponse) Descriptor() ([]byte, []int) {
	return file_public_portworx_common_apiv1_revision_proto_rawDescGZIP(), []int{4}
}

func (x *ListRevisionsResponse) GetRevisions() []*Revision {
	if x != nil {
		return x.Revisions
	}
	return nil
}

func (x *ListRevisionsResponse) GetPagination() *PageBasedPaginationResponse {
	if x != nil {
		return x.Pagination
	}
	return nil
}

// Name and semantic version of the revision.
type GetRevisionRequest_NameSemanticVersion struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name(kind) of the revision.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Version of the revision.
	SemanticVersion string `protobuf:"bytes,2,opt,name=semantic_version,json=semanticVersion,proto3" json:"semantic_version,omitempty"`
}

func (x *GetRevisionRequest_NameSemanticVersion) Reset() {
	*x = GetRevisionRequest_NameSemanticVersion{}
	if protoimpl.UnsafeEnabled {
		mi := &file_public_portworx_common_apiv1_revision_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetRevisionRequest_NameSemanticVersion) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRevisionRequest_NameSemanticVersion) ProtoMessage() {}

func (x *GetRevisionRequest_NameSemanticVersion) ProtoReflect() protoreflect.Message {
	mi := &file_public_portworx_common_apiv1_revision_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRevisionRequest_NameSemanticVersion.ProtoReflect.Descriptor instead.
func (*GetRevisionRequest_NameSemanticVersion) Descriptor() ([]byte, []int) {
	return file_public_portworx_common_apiv1_revision_proto_rawDescGZIP(), []int{2, 0}
}

func (x *GetRevisionRequest_NameSemanticVersion) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *GetRevisionRequest_NameSemanticVersion) GetSemanticVersion() string {
	if x != nil {
		return x.SemanticVersion
	}
	return ""
}

var File_public_portworx_common_apiv1_revision_proto protoreflect.FileDescriptor

var file_public_portworx_common_apiv1_revision_proto_rawDesc = []byte{
	0x0a, 0x2b, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x2f, 0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f, 0x72,
	0x78, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69, 0x76, 0x31, 0x2f, 0x72,
	0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x19, 0x70,
	0x75, 0x62, 0x6c, 0x69, 0x63, 0x2e, 0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f, 0x72, 0x78, 0x2e, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x62, 0x65, 0x68, 0x61, 0x76, 0x69, 0x6f,
	0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x27, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x2f,
	0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f, 0x72, 0x78, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f,
	0x61, 0x70, 0x69, 0x76, 0x31, 0x2f, 0x6d, 0x65, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x2b, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x2f, 0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f, 0x72,
	0x78, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69, 0x76, 0x31, 0x2f, 0x73,
	0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2d, 0x70,
	0x75, 0x62, 0x6c, 0x69, 0x63, 0x2f, 0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f, 0x72, 0x78, 0x2f, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69, 0x76, 0x31, 0x2f, 0x70, 0x61, 0x67, 0x69,
	0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x27, 0x70, 0x75,
	0x62, 0x6c, 0x69, 0x63, 0x2f, 0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f, 0x72, 0x78, 0x2f, 0x63, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69, 0x76, 0x31, 0x2f, 0x73, 0x6f, 0x72, 0x74, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x7c, 0x0a, 0x08, 0x52, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f,
	0x6e, 0x12, 0x33, 0x0a, 0x04, 0x6d, 0x65, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1f, 0x2e, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x2e, 0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f, 0x72,
	0x78, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x65, 0x74, 0x61,
	0x52, 0x04, 0x6d, 0x65, 0x74, 0x61, 0x12, 0x3b, 0x0a, 0x04, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x2e, 0x70, 0x6f,
	0x72, 0x74, 0x77, 0x6f, 0x72, 0x78, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31,
	0x2e, 0x52, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x04, 0x69,
	0x6e, 0x66, 0x6f, 0x22, 0x94, 0x01, 0x0a, 0x0c, 0x52, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e,
	0x49, 0x6e, 0x66, 0x6f, 0x12, 0x2e, 0x0a, 0x10, 0x73, 0x65, 0x6d, 0x61, 0x6e, 0x74, 0x69, 0x63,
	0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03,
	0xe0, 0x41, 0x02, 0x52, 0x0f, 0x73, 0x65, 0x6d, 0x61, 0x6e, 0x74, 0x69, 0x63, 0x56, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x12, 0x1e, 0x0a, 0x0a, 0x64, 0x65, 0x70, 0x72, 0x65, 0x63, 0x61, 0x74,
	0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0a, 0x64, 0x65, 0x70, 0x72, 0x65, 0x63,
	0x61, 0x74, 0x65, 0x64, 0x12, 0x34, 0x0a, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x42, 0x03, 0xe0,
	0x41, 0x02, 0x52, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x22, 0x8b, 0x02, 0x0a, 0x12, 0x47,
	0x65, 0x74, 0x52, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x12, 0x0a, 0x03, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00,
	0x52, 0x03, 0x75, 0x69, 0x64, 0x12, 0x77, 0x0a, 0x15, 0x6e, 0x61, 0x6d, 0x65, 0x5f, 0x73, 0x65,
	0x6d, 0x61, 0x6e, 0x74, 0x69, 0x63, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x41, 0x2e, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x2e, 0x70, 0x6f,
	0x72, 0x74, 0x77, 0x6f, 0x72, 0x78, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31,
	0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x2e, 0x4e, 0x61, 0x6d, 0x65, 0x53, 0x65, 0x6d, 0x61, 0x6e, 0x74, 0x69, 0x63,
	0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x48, 0x00, 0x52, 0x13, 0x6e, 0x61, 0x6d, 0x65, 0x53,
	0x65, 0x6d, 0x61, 0x6e, 0x74, 0x69, 0x63, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x1a, 0x5e,
	0x0a, 0x13, 0x4e, 0x61, 0x6d, 0x65, 0x53, 0x65, 0x6d, 0x61, 0x6e, 0x74, 0x69, 0x63, 0x56, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x17, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x2e,
	0x0a, 0x10, 0x73, 0x65, 0x6d, 0x61, 0x6e, 0x74, 0x69, 0x63, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x0f, 0x73,
	0x65, 0x6d, 0x61, 0x6e, 0x74, 0x69, 0x63, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x42, 0x08,
	0x0a, 0x06, 0x67, 0x65, 0x74, 0x5f, 0x62, 0x79, 0x22, 0xf3, 0x01, 0x0a, 0x14, 0x4c, 0x69, 0x73,
	0x74, 0x52, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x4f, 0x0a, 0x0e, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x73, 0x65, 0x6c, 0x65, 0x63,
	0x74, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x70, 0x75, 0x62, 0x6c,
	0x69, 0x63, 0x2e, 0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f, 0x72, 0x78, 0x2e, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x42, 0x03,
	0xe0, 0x41, 0x01, 0x52, 0x0d, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74,
	0x6f, 0x72, 0x12, 0x33, 0x0a, 0x04, 0x73, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1f, 0x2e, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x2e, 0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f,
	0x72, 0x78, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x6f, 0x72,
	0x74, 0x52, 0x04, 0x73, 0x6f, 0x72, 0x74, 0x12, 0x55, 0x0a, 0x0a, 0x70, 0x61, 0x67, 0x69, 0x6e,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x35, 0x2e, 0x70, 0x75,
	0x62, 0x6c, 0x69, 0x63, 0x2e, 0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f, 0x72, 0x78, 0x2e, 0x63, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x61, 0x67, 0x65, 0x42, 0x61, 0x73, 0x65,
	0x64, 0x50, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x52, 0x0a, 0x70, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0xb2,
	0x01, 0x0a, 0x15, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x73,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x41, 0x0a, 0x09, 0x72, 0x65, 0x76, 0x69,
	0x73, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x70, 0x75,
	0x62, 0x6c, 0x69, 0x63, 0x2e, 0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f, 0x72, 0x78, 0x2e, 0x63, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e,
	0x52, 0x09, 0x72, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x56, 0x0a, 0x0a, 0x70,
	0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x36, 0x2e, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x2e, 0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f, 0x72,
	0x78, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x61, 0x67, 0x65,
	0x42, 0x61, 0x73, 0x65, 0x64, 0x50, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x52, 0x0a, 0x70, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x42, 0x6d, 0x0a, 0x1d, 0x63, 0x6f, 0x6d, 0x2e, 0x70, 0x75, 0x62, 0x6c, 0x69,
	0x63, 0x2e, 0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f, 0x72, 0x78, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2e, 0x76, 0x31, 0x42, 0x0d, 0x52, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x3b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x70, 0x75, 0x72, 0x65, 0x2d, 0x70, 0x78, 0x2f, 0x61, 0x70, 0x69, 0x73, 0x2f, 0x70,
	0x75, 0x62, 0x6c, 0x69, 0x63, 0x2f, 0x70, 0x6f, 0x72, 0x74, 0x77, 0x6f, 0x72, 0x78, 0x2f, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69, 0x76, 0x31, 0x3b, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_public_portworx_common_apiv1_revision_proto_rawDescOnce sync.Once
	file_public_portworx_common_apiv1_revision_proto_rawDescData = file_public_portworx_common_apiv1_revision_proto_rawDesc
)

func file_public_portworx_common_apiv1_revision_proto_rawDescGZIP() []byte {
	file_public_portworx_common_apiv1_revision_proto_rawDescOnce.Do(func() {
		file_public_portworx_common_apiv1_revision_proto_rawDescData = protoimpl.X.CompressGZIP(file_public_portworx_common_apiv1_revision_proto_rawDescData)
	})
	return file_public_portworx_common_apiv1_revision_proto_rawDescData
}

var file_public_portworx_common_apiv1_revision_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_public_portworx_common_apiv1_revision_proto_goTypes = []interface{}{
	(*Revision)(nil),                               // 0: public.portworx.common.v1.Revision
	(*RevisionInfo)(nil),                           // 1: public.portworx.common.v1.RevisionInfo
	(*GetRevisionRequest)(nil),                     // 2: public.portworx.common.v1.GetRevisionRequest
	(*ListRevisionsRequest)(nil),                   // 3: public.portworx.common.v1.ListRevisionsRequest
	(*ListRevisionsResponse)(nil),                  // 4: public.portworx.common.v1.ListRevisionsResponse
	(*GetRevisionRequest_NameSemanticVersion)(nil), // 5: public.portworx.common.v1.GetRevisionRequest.NameSemanticVersion
	(*Meta)(nil),                                   // 6: public.portworx.common.v1.Meta
	(*structpb.Struct)(nil),                        // 7: google.protobuf.Struct
	(*Selector)(nil),                               // 8: public.portworx.common.v1.Selector
	(*Sort)(nil),                                   // 9: public.portworx.common.v1.Sort
	(*PageBasedPaginationRequest)(nil),             // 10: public.portworx.common.v1.PageBasedPaginationRequest
	(*PageBasedPaginationResponse)(nil),            // 11: public.portworx.common.v1.PageBasedPaginationResponse
}
var file_public_portworx_common_apiv1_revision_proto_depIdxs = []int32{
	6,  // 0: public.portworx.common.v1.Revision.meta:type_name -> public.portworx.common.v1.Meta
	1,  // 1: public.portworx.common.v1.Revision.info:type_name -> public.portworx.common.v1.RevisionInfo
	7,  // 2: public.portworx.common.v1.RevisionInfo.schema:type_name -> google.protobuf.Struct
	5,  // 3: public.portworx.common.v1.GetRevisionRequest.name_semantic_version:type_name -> public.portworx.common.v1.GetRevisionRequest.NameSemanticVersion
	8,  // 4: public.portworx.common.v1.ListRevisionsRequest.field_selector:type_name -> public.portworx.common.v1.Selector
	9,  // 5: public.portworx.common.v1.ListRevisionsRequest.sort:type_name -> public.portworx.common.v1.Sort
	10, // 6: public.portworx.common.v1.ListRevisionsRequest.pagination:type_name -> public.portworx.common.v1.PageBasedPaginationRequest
	0,  // 7: public.portworx.common.v1.ListRevisionsResponse.revisions:type_name -> public.portworx.common.v1.Revision
	11, // 8: public.portworx.common.v1.ListRevisionsResponse.pagination:type_name -> public.portworx.common.v1.PageBasedPaginationResponse
	9,  // [9:9] is the sub-list for method output_type
	9,  // [9:9] is the sub-list for method input_type
	9,  // [9:9] is the sub-list for extension type_name
	9,  // [9:9] is the sub-list for extension extendee
	0,  // [0:9] is the sub-list for field type_name
}

func init() { file_public_portworx_common_apiv1_revision_proto_init() }
func file_public_portworx_common_apiv1_revision_proto_init() {
	if File_public_portworx_common_apiv1_revision_proto != nil {
		return
	}
	file_public_portworx_common_apiv1_meta_proto_init()
	file_public_portworx_common_apiv1_selector_proto_init()
	file_public_portworx_common_apiv1_pagination_proto_init()
	file_public_portworx_common_apiv1_sort_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_public_portworx_common_apiv1_revision_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Revision); i {
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
		file_public_portworx_common_apiv1_revision_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RevisionInfo); i {
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
		file_public_portworx_common_apiv1_revision_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetRevisionRequest); i {
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
		file_public_portworx_common_apiv1_revision_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListRevisionsRequest); i {
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
		file_public_portworx_common_apiv1_revision_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListRevisionsResponse); i {
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
		file_public_portworx_common_apiv1_revision_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetRevisionRequest_NameSemanticVersion); i {
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
	file_public_portworx_common_apiv1_revision_proto_msgTypes[2].OneofWrappers = []interface{}{
		(*GetRevisionRequest_Uid)(nil),
		(*GetRevisionRequest_NameSemanticVersion_)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_public_portworx_common_apiv1_revision_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_public_portworx_common_apiv1_revision_proto_goTypes,
		DependencyIndexes: file_public_portworx_common_apiv1_revision_proto_depIdxs,
		MessageInfos:      file_public_portworx_common_apiv1_revision_proto_msgTypes,
	}.Build()
	File_public_portworx_common_apiv1_revision_proto = out.File
	file_public_portworx_common_apiv1_revision_proto_rawDesc = nil
	file_public_portworx_common_apiv1_revision_proto_goTypes = nil
	file_public_portworx_common_apiv1_revision_proto_depIdxs = nil
}