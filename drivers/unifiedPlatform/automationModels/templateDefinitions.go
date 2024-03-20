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
	Uid         *string `copier:"must,nopanic"`
	Name        *string `copier:"must,nopanic"`
	Description *string
}

type ListRevisionResponse struct {
	Meta *Meta         `copier:"must,nopanic"`
	Info *RevisionInfo `copier:"must,nopanic"`
}

type RevisionInfo struct {
	SemanticVersion string           `copier:"must,nopanic"`
	Deprecated      bool             `copier:"must,nopanic"`
	Schema          *structpb.Struct `copier:"must,nopanic"`
}

type ListTemplateKindsResponse struct {
	Kinds      []string `copier:"must,nopanic"`
	Pagination *PageBasedPaginationResponse
}

type Revision struct {
	Meta *Meta         `copier:"must,nopanic"`
	Info *RevisionInfo `copier:"must,nopanic"`
}

type PageBasedPaginationResponse struct {
	TotalRecords int64 `copier:"must,nopanic"`
	CurrentPage  int64 `copier:"must,nopanic"`
	PageSize     int64 `copier:"must,nopanic"`
	TotalPages   int64 `copier:"must,nopanic"`
	NextPage     int64 `copier:"must,nopanic"`
	PrevPage     int64 `copier:"must,nopanic"`
}
