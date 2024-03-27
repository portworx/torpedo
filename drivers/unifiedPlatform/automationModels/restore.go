package automationModels

import "time"

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
	Config *RestoreConfig `copier:"must,nopanic"`
	Status *Status `copier:"must,nopanic"`
}

// V1Config Desired configuration of the restore.
type RestoreConfig struct {
	SourceReferences      *SourceReferences      `copier:"must,nopanic"`
	DestinationReferences *DestinationReferences `copier:"must,nopanic"`
	// K8s resource name for restore, built from [\"restore-\" + name + short-id].
	CustomResourceName *string `copier:"must,nopanic"`
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


// Restorev1Status Status of the restore.
type Restorev1Status struct {
	//  Time when restore was started.
	StartedAt *time.Time `copier:"must,nopanic"`
	//  Time when restore was completed.
	CompletedAt *time.Time   `copier:"must,nopanic"`
	ErrorCode   *V1ErrorCode `copier:"must,nopanic"`
	// Error message is description of the error in restore.
	ErrorMessage *string  `copier:"must,nopanic"`
	Phase        *V1Phase `copier:"must,nopanic"`
}