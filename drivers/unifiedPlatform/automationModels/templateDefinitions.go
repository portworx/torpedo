package automationModels

import structpb "google.golang.org/protobuf/types/known/structpb"

type TemplateDefinitionRequest struct {
	GetType GetTemplateDefinitionType
}

type TemplateDefinitionResponse struct {
	GetType      V1TemplateType
	ListRevision ListRevisionResponse
	ListKinds    ListTemplateKindsResponse
	GetRevision  Revision
}

type GetTemplateDefinitionType struct {
	Id string
}

type V1TemplateType struct {
	Uid         *string `json:"uid,omitempty"`
	Name        *string `json:"name,omitempty"`
	Description *string
}

type ListRevisionResponse struct {
	Meta *Meta         `protobuf:"bytes,1,opt,name=meta,proto3" json:"meta,omitempty"`
	Info *RevisionInfo `protobuf:"bytes,2,opt,name=info,proto3" json:"info,omitempty"`
}

type RevisionInfo struct {
	SemanticVersion string           `protobuf:"bytes,1,opt,name=semantic_version,json=semanticVersion,proto3" json:"semantic_version,omitempty"`
	Deprecated      bool             `protobuf:"varint,2,opt,name=deprecated,proto3" json:"deprecated,omitempty"`
	Schema          *structpb.Struct `protobuf:"bytes,3,opt,name=schema,proto3" json:"schema,omitempty"`
}

type ListTemplateKindsResponse struct {
	Kinds      []string `protobuf:"bytes,1,rep,name=kinds,proto3" json:"kinds,omitempty"`
	Pagination *PageBasedPaginationResponse
}

type Revision struct {
	Meta *Meta `protobuf:"bytes,1,opt,name=meta,proto3" json:"meta,omitempty"`
	Info *RevisionInfo
}

type PageBasedPaginationResponse struct {
	TotalRecords int64 `protobuf:"varint,1,opt,name=total_records,json=totalRecords,proto3" json:"total_records,omitempty"`
	CurrentPage  int64 `protobuf:"varint,2,opt,name=current_page,json=currentPage,proto3" json:"current_page,omitempty"`
	PageSize     int64 `protobuf:"varint,3,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
	TotalPages   int64 `protobuf:"varint,4,opt,name=total_pages,json=totalPages,proto3" json:"total_pages,omitempty"`
	NextPage     int64 `protobuf:"varint,5,opt,name=next_page,json=nextPage,proto3" json:"next_page,omitempty"`
	PrevPage     int64 `protobuf:"varint,6,opt,name=prev_page,json=prevPage,proto3" json:"prev_page,omitempty"`
}
