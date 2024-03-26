package automationModels

type PDSRestoreRequest struct {
	Create   PDSCreateRestore
	ReCreate PDSReCreateRestore
	Get      PDSGetRestore
	Delete   PDSDeleteRestore
	List     PDSListRestores
}

type PDSRestoreResponse struct {
	Create   PDSRestore
	ReCreate PDSRestore
	Get      PDSRestore
	Delete   PDSRestore
	List     PDSListRestoreResponse
}

type PDSListRestores struct {
	AccountId            string
	TenantId             string
	ProjectId            string
	DeploymentId         string
	BackupId             string
	SortSortBy           string
	SortSortOrder        string
	PaginationPageNumber string
	PaginationPageSize   string
}

type PDSCreateRestore struct {
	NamespaceId string      `copier:"must,nopanic"`
	ProjectId   string      `copier:"must,nopanic"`
	Restore     *PDSRestore `copier:"must,nopanic"`
	// SourceReferences for the restore.
	SourceReferences *SourceReferences `copier:"must,nopanic"`
	// Destination references for the restore.
	DestinationReferences *DestinationReferences `copier:"must,nopanic"`
	// K8s resource name for restore, built from ["restore-" + name + short-id].
	CustomResourceName string `copier:"must,nopanic"`
}

type PDSListRestoreResponse struct {
	Restores []PDSRestore
}

type PDSRestore struct {
	Meta   *Meta   `copier:"must,nopanic"`
	Config *Config `copier:"must,nopanic"`
	Status *Status `copier:"must,nopanic"`
}

type PDSReCreateRestore struct {
	Id              string `copier:"must,nopanic"`
	TargetClusterId string `copier:"must,nopanic"`
	Name            string `copier:"must,nopanic"`
	NamespaceId     string `copier:"must,nopanic"`
	ProjectId       string `copier:"must,nopanic"`
}

type PDSGetRestore struct {
	Id string `copier:"must,nopanic"`
}

type PDSDeleteRestore struct {
	Id string `copier:"must,nopanic"`
}
