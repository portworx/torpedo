package automationModels

import "time"

type PDSBackup struct {
	Delete PDSDeleteBackup `copier:"must,nopanic"`
	List   PDSListBackup   `copier:"must,nopanic"`
	Get    PDSGetRestore   `copier:"must,nopanic"`
}

type PDSBackupResponse struct {
	Meta   Meta
	Config NewV1BackupConfig
	Status Backupv1Status
}

type PDSDeleteBackup struct {
	Id string `copier:"must,nopanic"`
}

// V1Config Desired configuration of the Backup.
type NewV1BackupConfig struct {
	References *V1BackupReferences `json:"references,omitempty"`
	// BackupCapability of the deployment target when the snapshot was created.
	BackupCapability *string `json:"backupCapability,omitempty"`
}

// V1References References to other resources.
type V1BackupReferences struct {
	// UID of the image of the data service which will needs to be backup .
	ImageId *string `json:"imageId,omitempty"`
}

type PDSListBackup struct {
	Meta            Meta
	Config          Config
	Status          Backupv1Status
	Pagination      *PageBasedPaginationRequest `copier:"must,nopanic"`
	Sort            *Sort                       `copier:"must,nopanic"`
	BackupConfigId  string                      `copier:"must,nopanic"`
	TargetClusterId string                      `copier:"must,nopanic"`
	NamespaceId     string                      `copier:"must,nopanic"`
	DeploymentId    string                      `copier:"must,nopanic"`
}

type PDSGetBackupRequest struct {
	Id string `copier:"must,nopanic"`
}

// Backupv1Status Status of the Backup.
type Backupv1Status struct {
	// CloudSnapID snapshot of the backup volume.
	CloudSnapId *string `copier:"must,nopanic"`
	// Start time of the backup.
	StartTime *time.Time `copier:"must,nopanic"`
	// Completion time of the backup.
	CompletionTime *time.Time   `copier:"must,nopanic"`
	Phase          *StatusPhase `copier:"must,nopanic"`
	// ErrorCode if CompletionStatus is \"Failed\".
	ErrorCode *string `copier:"must,nopanic"`
	// ErrorMessage associated with the ErrorCode.
	ErrorMessage *string `copier:"must,nopanic"`
	// FileSize of the CloudSnap image.
	FileSize *string `copier:"must,nopanic"`
}
