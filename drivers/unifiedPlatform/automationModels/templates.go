package automationModels

type PlatformTemplatesRequest struct {
	Create        CreatePlatformTemplates `copier:"must,nopanic"`
	List          ListTemplates           `copier:"must,nopanic"`
	ListForTenant ListTemplatesForTenant  `copier:"must,nopanic"`
	Update        UpdatePlatformTemplates `copier:"must,nopanic"`
	Get           GetPlatformTemplates    `copier:"must,nopanic"`
	Delete        DeletePlatformTemplates `copier:"must,nopanic"`
}

type PlatformTemplatesResponse struct {
	Create        V1Template             `copier:"must,nopanic"`
	List          V1ListTemplateResopnse `copier:"must,nopanic"`
	ListForTenant V1ListTemplateResopnse `copier:"must,nopanic"`
	Update        V1Template             `copier:"must,nopanic"`
	Get           V1Template             `copier:"must,nopanic"`
	Delete        V1Template             `copier:"must,nopanic"`
}

type CreatePlatformTemplates struct {
	TenantId string    `copier:"must,nopanic"`
	Template *Template `copier:"must,nopanic"`
}

type UpdatePlatformTemplates struct {
	Id       string    `copier:"must,nopanic"`
	Template *Template `copier:"must,nopanic"`
}

type DeletePlatformTemplates struct {
	Id string `copier:"must,nopanic"`
}

type GetPlatformTemplates struct {
	Id string `copier:"must,nopanic"`
}

type ListTemplatesForTenant struct {
	TenantId string `copier:"must,nopanic"`
}

type ListTemplates struct {
	V1ListTemplatesRequest V1ListTemplatesRequest `copier:"must,nopanic"`
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
	Filters []SelectorFilter `copier:"must,nopanic"`
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
	Templates []V1Template `copier:"must,nopanic"`
}
