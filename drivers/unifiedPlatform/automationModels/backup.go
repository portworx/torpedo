package automationModels

import "time"

type PDSBackupRequest struct {
	Delete PDSDeleteBackup `copier:"must,nopanic"`
	List   PDSListBackup   `copier:"must,nopanic"`
	Get    PDSGetRestore   `copier:"must,nopanic"`
}

type PDSBackupResponse struct {
	List PDSBackupListResponse `copier:"must,nopanic"`
	Get  V1Backup              `copier:"must,nopanic"`
}

type PDSDeleteBackup struct {
	Id string `copier:"must,nopanic"`
}

type PDSGetBackupRequest struct {
	Id string `copier:"must,nopanic"`
}

type PDSBackupListResponse struct {
	Backups []V1Backup
}

type V1Backup struct {
	Meta   V1Meta         `json:"meta,omitempty"`
	Config V1Config       `json:"config,omitempty"`
	Status Backupv1Status `json:"status,omitempty"`
}

type Backupv1Status struct {
	CloudSnapId    string      `json:"cloudSnapId,omitempty"`
	StartTime      time.Time   `json:"startTime,omitempty"`
	CompletionTime time.Time   `json:"completionTime,omitempty"`
	Phase          StatusPhase `json:"phase,omitempty"`
	ErrorCode      string      `json:"errorCode,omitempty"`
	ErrorMessage   string      `json:"errorMessage,omitempty"`
	FileSize       string      `json:"fileSize,omitempty"`
}

type PDSListBackup struct {
	DeploymentId         string
	TargetClusterId      string
	NamespaceId          string
	BackupConfigId       string
	PaginationPageNumber string
	PaginationPageSize   string
	SortSortBy           string
	SortSortOrder        string
}
