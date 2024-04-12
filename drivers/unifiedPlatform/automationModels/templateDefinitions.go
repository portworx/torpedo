package automationModels

import structpb "google.golang.org/protobuf/types/known/structpb"

type TemplateDefinitionRequest struct {
	GetType     GetTemplateDefinitionType `copier:"must,nopanic"`
	GetRevision GetTemplateRevisions      `copier:"must,nopanic"`
}

type TemplateDefinitionResponse struct {
	GetType      V1TemplateType              `copier:"must,nopanic"`
	ListRevision ListRevisionResponse        `copier:"must,nopanic"`
	ListKinds    ListTemplateKindsResponse   `copier:"must,nopanic"`
	ListTypes    ListTemplateTypesResponse   `copier:"must,nopanic"`
	ListSamples  ListTemplateSamplesResponse `copier:"must,nopanic"`
	GetRevision  Revision                    `copier:"must,nopanic"`
}

type GetTemplateDefinitionType struct {
	Id string `copier:"must,nopanic"`
}

type GetTemplateRevisions struct {
	Uid string `copier:"must,nopanic"`
}

type V1TemplateType struct {
	Uid         *string `copier:"must,nopanic"`
	Name        *string `copier:"must,nopanic"`
	Description *string `copier:"must,nopanic"`
}

type ListRevisionResponse struct {
	Revisions []Revision
}

type RevisionInfo struct {
	SemanticVersion string           `copier:"must,nopanic"`
	Deprecated      bool             `copier:"must,nopanic"`
	Schema          *structpb.Struct `copier:"must,nopanic"`
}

type ListTemplateKindsResponse struct {
	Kinds      []string                     `copier:"must,nopanic"`
	Pagination *PageBasedPaginationResponse `copier:"must,nopanic"`
}

type ListTemplateTypesResponse struct {
	TemplateTypes []string `copier:"must,nopanic"`
}

type ListTemplateSamplesResponse struct {
	TemplateSamples []string `copier:"must,nopanic"`
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
