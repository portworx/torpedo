package automationModels

type PlatformNamespace struct {
	List PlatformListNamespace
}

type PlatformNamespaceResponse struct {
	List V1ListNamespacesResponse
}

// ListNamespacesRequest struct
type PlatformListNamespace struct {
	TenantId      string `copier:"must,nopanic"`
	Label         string `copier:"must,nopanic"`
	SortSortBy    string `copier:"must,nopanic"`
	SortSortOrder string `copier:"must,nopanic"`
}

type V1ListNamespacesResponse struct {
	Namespaces []V1Namespace `copier:"must,nopanic"`
}

type V1Namespace struct {
	Meta   *V1Meta            `copier:"must,nopanic"`
	Status *V1NamespaceStatus `copier:"must,nopanic"`
}

type V1NamespaceStatus struct {
	Phase *NamespacePhasePhase `copier:"must,nopanic"`
}

type NamespacePhasePhase string
