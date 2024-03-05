package apiStructs

import (
	"time"
)

type ConfigBackupType string

type ConfigBackupLevel string

type ConfigReclaimPolicyType string

type PDSBackupConfig struct {
	V1   PDSBackupConfigV1   `copier:"must,nopanic"`
	GRPC PDSBackupConfigGRPC `copier:"must,nopanic"`
}

type PDSBackupConfigV1 struct {
	Create CreatePDSBackupConfig `copier:"must,nopanic"`
	Update UpdatePDSBackupConfig `copier:"must,nopanic"`
	Get    GetPDSBackupConfig    `copier:"must,nopanic"`
	Delete DeletePDSBackupConfig `copier:"must,nopanic"`
	List   ListPDSBackupConfig   `copier:"must,nopanic"`
}

type PDSBackupConfigGRPC struct {
	Create CreatePDSBackupConfig `copier:"must,nopanic"`
	Update UpdatePDSBackupConfig `copier:"must,nopanic"`
	Get    GetPDSBackupConfig    `copier:"must,nopanic"`
	Delete DeletePDSBackupConfig `copier:"must,nopanic"`
	List   ListPDSBackupConfig   `copier:"must,nopanic"`
}

type CreatePDSBackupConfig struct {
	ProjectId      string          `copier:"must,nopanic"`
	DeploymentId   string          `copier:"must,nopanic"`
	V1BackupConfig *V1BackupConfig `copier:"must,nopanic"`
}

type UpdatePDSBackupConfig struct {
	BackupConfigMetaUid        string                      `copier:"must,nopanic"`
	DesiredBackupConfiguration *DesiredBackupConfiguration `copier:"must,nopanic"`
}

type GetPDSBackupConfig struct {
	Id string `copier:"must,nopanic"`
}

type DeletePDSBackupConfig struct {
	Id string `copier:"must,nopanic"`
}

type ListPDSBackupConfig struct {
	AccountId            *string `copier:"must,nopanic"`
	TenantId             *string `copier:"must,nopanic"`
	ProjectId            *string `copier:"must,nopanic"`
	TargetClusterId      *string `copier:"must,nopanic"`
	NamespaceId          *string `copier:"must,nopanic"`
	DeploymentId         *string `copier:"must,nopanic"`
	PaginationPageNumber *string `copier:"must,nopanic"`
	PaginationPageSize   *string `copier:"must,nopanic"`
	SortSortBy           *string `copier:"must,nopanic"`
	SortSortOrder        *string `copier:"must,nopanic"`
}

type V1BackupConfig struct {
	Meta   *Meta                 `copier:"must,nopanic"`
	Config *Config               `copier:"must,nopanic"`
	Status *Backupconfigv1Status `copier:"must,nopanic"`
}

type Backupconfigv1Status struct {
	Phase                  *StatusPhase          `copier:"must,nopanic"`
	CustomResourceName     *string               `copier:"must,nopanic"`
	IsScheduleSynchronized *bool                 `copier:"must,nopanic"`
	DeploymentMetaData     *V1DeploymentMetaData `copier:"must,nopanic"`
}

type MetadataOfTheBackupConfiguration struct {
	Name            *string            `copier:"must,nopanic"`
	Description     *string            `copier:"must,nopanic"`
	ResourceVersion *string            `copier:"must,nopanic"`
	CreateTime      *time.Time         `copier:"must,nopanic"`
	UpdateTime      *time.Time         `copier:"must,nopanic"`
	Labels          *map[string]string `copier:"must,nopanic"`
	Annotations     *map[string]string `copier:"must,nopanic"`
}

type DesiredBackupConfiguration struct {
	Meta   *MetadataOfTheBackupConfiguration `copier:"must,nopanic"`
	Config *BackupV1Config                   `copier:"must,nopanic"`
	Status *Backupconfigv1Status             `copier:"must,nopanic"`
}

type BackupV1Config struct {
	References      *BackupV1References1     `copier:"must,nopanic"`
	JobHistoryLimit *int32                   `copier:"must,nopanic"`
	Schedule        *BackupV1Schedule        `copier:"must,nopanic"`
	Suspend         *bool                    `copier:"must,nopanic"`
	BackupType      *ConfigBackupType        `copier:"must,nopanic"`
	BackupLevel     *ConfigBackupLevel       `copier:"must,nopanic"`
	ReclaimPolicy   *ConfigReclaimPolicyType `copier:"must,nopanic"`
}

type BackupV1References1 struct {
	DeploymentId     *string `copier:"must,nopanic"`
	BackupLocationId *string `copier:"must,nopanic"`
	DataServiceId    *string `copier:"must,nopanic"`
}

type BackupV1Schedule struct {
	Id              *string `copier:"must,nopanic"`
	ResourceVersion *string `copier:"must,nopanic"`
}
