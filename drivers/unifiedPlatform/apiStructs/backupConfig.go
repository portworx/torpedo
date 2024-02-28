package apiStructs

import (
	"time"
)

type ConfigBackupType string

type ConfigBackupLevel string

type ConfigReclaimPolicyType string

type PDSBackupConfig struct {
	V1   PDSBackupConfigV1
	GRPC PDSBackupConfigGRPC
}

type PDSBackupConfigV1 struct {
	Create CreatePDSBackupConfig
	Update UpdatePDSBackupConfig
	Get    GetPDSBackupConfig
	Delete DeletePDSBackupConfig
	List   ListPDSBackupConfig
}

type PDSBackupConfigGRPC struct {
	Create CreatePDSBackupConfig
	Update UpdatePDSBackupConfig
	Get    GetPDSBackupConfig
	Delete DeletePDSBackupConfig
	List   ListPDSBackupConfig
}

type CreatePDSBackupConfig struct {
	ProjectId      string
	DeploymentId   *string
	V1BackupConfig *V1BackupConfig
}

type UpdatePDSBackupConfig struct {
	BackupConfigMetaUid        string
	DesiredBackupConfiguration *DesiredBackupConfiguration
}

type GetPDSBackupConfig struct {
	Id string
}

type DeletePDSBackupConfig struct {
	Id string
}

type ListPDSBackupConfig struct {
	AccountId            *string
	TenantId             *string
	ProjectId            *string
	TargetClusterId      *string
	NamespaceId          *string
	DeploymentId         *string
	PaginationPageNumber *string
	PaginationPageSize   *string
	SortSortBy           *string
	SortSortOrder        *string
}

type V1BackupConfig struct {
	Meta   Meta                  `json:"meta,omitempty"`
	Config Config                `json:"config,omitempty"`
	Status *Backupconfigv1Status `json:"status,omitempty"`
}

type Backupconfigv1Status struct {
	Phase                  *StatusPhase          `json:"phase,omitempty"`
	CustomResourceName     *string               `json:"customResourceName,omitempty"`
	IsScheduleSynchronized *bool                 `json:"isScheduleSynchronized,omitempty"`
	DeploymentMetaData     *V1DeploymentMetaData `json:"deploymentMetaData,omitempty"`
}

type MetadataOfTheBackupConfiguration struct {
	Name            *string            `json:"name,omitempty"`
	Description     *string            `json:"description,omitempty"`
	ResourceVersion *string            `json:"resourceVersion,omitempty"`
	CreateTime      *time.Time         `json:"createTime,omitempty"`
	UpdateTime      *time.Time         `json:"updateTime,omitempty"`
	Labels          *map[string]string `json:"labels,omitempty"`
	Annotations     *map[string]string `json:"annotations,omitempty"`
}

type DesiredBackupConfiguration struct {
	Meta   *MetadataOfTheBackupConfiguration `json:"meta,omitempty"`
	Config *BackupV1Config                   `json:"config,omitempty"`
	Status *Backupconfigv1Status             `json:"status,omitempty"`
}

type BackupV1Config struct {
	References      *BackupV1References1     `json:"references,omitempty"`
	JobHistoryLimit *int32                   `json:"jobHistoryLimit,omitempty"`
	Schedule        *BackupV1Schedule        `json:"schedule,omitempty"`
	Suspend         *bool                    `json:"suspend,omitempty"`
	BackupType      *ConfigBackupType        `json:"backupType,omitempty"`
	BackupLevel     *ConfigBackupLevel       `json:"backupLevel,omitempty"`
	ReclaimPolicy   *ConfigReclaimPolicyType `json:"reclaimPolicy,omitempty"`
}

type BackupV1References1 struct {
	DeploymentId     *string `json:"deploymentId,omitempty"`
	BackupLocationId *string `json:"backupLocationId,omitempty"`
	DataServiceId    *string `json:"dataServiceId,omitempty"`
}

type BackupV1Schedule struct {
	Id              *string `json:"id,omitempty"`
	ResourceVersion *string `json:"resourceVersion,omitempty"`
}
