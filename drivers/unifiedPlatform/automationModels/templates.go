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
	TenantId              string             `json:"tenantId,omitempty"`
	Pagination            PaginationRequest  `json:"pagination,omitempty"`
	LabelSelector         V1Selector         `json:"labelSelector,omitempty"`
	FieldSelector         V1Selector         `json:"fieldSelector,omitempty"`
	InfraResourceSelector V1ResourceSelector `json:"infraResourceSelector,omitempty"`
	RespData              V1RespData         `json:"respData,omitempty"`
	Sort                  Sort               `json:"sort,omitempty"`
}

type V1Selector struct {
	Filters []SelectorFilter
}

type SelectorFilter struct {
	Key    string           `json:"key,omitempty"`
	Op     SelectorOperator `json:"op,omitempty"`
	Values []string         `json:"values,omitempty"`
}

type V1ResourceSelector struct {
	InfraResourceFilters []ResourceSelectorResourceFilter `json:"infraResourceFilters,omitempty"`
}

type ResourceSelectorResourceFilter struct {
	ResourceType V1InfraResourceType `json:"resourceType,omitempty"`
	Op           SelectorOperator    `json:"op,omitempty"`
	Values       []string            `json:"values,omitempty"`
}

type SelectorOperator string

type V1InfraResourceType string

type V1RespData string

type V1Template struct {
	Meta   *V1Meta           `json:"meta,omitempty"`
	Config *V1Config         `json:"config,omitempty"`
	Status *Templatev1Status `json:"status,omitempty"`
}

type V1ListTemplateResopnse struct {
	Templates []V1Template
}
