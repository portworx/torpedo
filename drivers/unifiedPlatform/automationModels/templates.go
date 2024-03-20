package automationModels

type PlatformTemplatesRequest struct {
	Create        CreatePlatformTemplates
	List          ListTemplates
	ListForTenant ListTemplatesForTenant
	Update        UpdatePlatformTemplates
	Get           GetPlatformTemplates
	Delete        DeletePlatformTemplates
}

type PlatformTemplatesResponse struct {
	Create        V1Template
	List          V1ListTemplateResopnse
	ListForTenant V1ListTemplateResopnse
	Update        V1Template
	Get           V1Template
	Delete        V1Template
}

type CreatePlatformTemplates struct {
	TenantId string
	Template *Template
}

type UpdatePlatformTemplates struct {
	Id       string
	Template *Template
}

type DeletePlatformTemplates struct {
	Id string
}

type GetPlatformTemplates struct {
	Id string
}

type ListTemplatesForTenant struct {
	TenantId string
}

type ListTemplates struct {
	V1ListTemplatesRequest V1ListTemplatesRequest
}

type V1ListTemplatesRequest struct {
	TenantId              string             `copier:"must,nopanic"`
	Pagination            PaginationRequest  `copier:"must,nopanic"`
	LabelSelector         V1Selector         `copier:"must,nopanic"`
	FieldSelector         V1Selector         `copier:"must,nopanic"`
	InfraResourceSelector V1ResourceSelector `copier:"must,nopanic"`
	RespData              V1RespData         `copier:"must,nopanic"`
	Sort                  Sort               `copier:"must,nopanic"`
}

type V1Selector struct {
	Filters []SelectorFilter
}

type SelectorFilter struct {
	Key    string           `copier:"must,nopanic"`
	Op     SelectorOperator `copier:"must,nopanic"`
	Values []string         `copier:"must,nopanic"`
}

type V1ResourceSelector struct {
	InfraResourceFilters []ResourceSelectorResourceFilter `copier:"must,nopanic"`
}

type ResourceSelectorResourceFilter struct {
	ResourceType V1InfraResourceType `copier:"must,nopanic"`
	Op           SelectorOperator    `copier:"must,nopanic"`
	Values       []string            `copier:"must,nopanic"`
}

type SelectorOperator string

type V1InfraResourceType string

type V1RespData string

type V1Template struct {
	Meta   *V1Meta           `copier:"must,nopanic"`
	Config *V1Config         `copier:"must,nopanic"`
	Status *Templatev1Status `copier:"must,nopanic"`
}

type V1ListTemplateResopnse struct {
	Templates []V1Template
}
