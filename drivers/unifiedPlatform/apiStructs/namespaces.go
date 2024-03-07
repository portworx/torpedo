package apiStructs

// ListNamespacesRequest struct
type ListNamespacesRequest struct {
	ClusterId     string `copier:"must,nopanic"`
	TenantId      string `copier:"must,nopanic"`
	ProjectId     string `copier:"must,nopanic"`
	Label         string `copier:"must,nopanic"`
	SortSortBy    string `copier:"must,nopanic"`
	SortSortOrder string `copier:"must,nopanic"`
}
