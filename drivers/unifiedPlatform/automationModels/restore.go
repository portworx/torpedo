package automationModels

import "time"

type PDSRestore struct {
	Create   PDSCreateRestore
	ReCreate PDSReCreateRestore
	Get      PDSGetRestore
	Delete   PDSDeleteRestore
	List     PDSListRestores
}

type PDSCreateRestore struct {
	NamespaceId string   `copier:"must,nopanic"`
	ProjectId   string   `copier:"must,nopanic"`
	Restore     *Restore `copier:"must,nopanic"`
	// SourceReferences for the restore.
	SourceReferences *SourceReferences `copier:"must,nopanic"`
	// Destination references for the restore.
	DestinationReferences *DestinationReferences `copier:"must,nopanic"`
	// K8s resource name for restore, built from ["restore-" + name + short-id].
	CustomResourceName string `copier:"must,nopanic"`
}

type Restore struct {
	Meta   *Meta            `copier:"must,nopanic"`
	Config *RestoreConfig   `copier:"must,nopanic"`
	Status *Restorev1Status `copier:"must,nopanic"`
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
}

type PDSGetRestore struct {
	Id string `copier:"must,nopanic"`
}

type PDSDeleteRestore struct {
	Id string `copier:"must,nopanic"`
}

type PDSListRestores struct {
	Sort       *Sort                       `copier:"must,nopanic"`
	Pagination *PageBasedPaginationRequest `copier:"must,nopanic"`
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
