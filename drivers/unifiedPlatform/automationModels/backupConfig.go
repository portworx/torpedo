package automationModels

import (
	"time"
)

type ConfigBackupType string

type ConfigBackupLevel string

type ConfigReclaimPolicyType string

type PDSBackupConfigRequest struct {
	Create CreatePDSBackupConfig `copier:"must,nopanic"`
	Update UpdatePDSBackupConfig `copier:"must,nopanic"`
	Get    GetPDSBackupConfig    `copier:"must,nopanic"`
	Delete DeletePDSBackupConfig `copier:"must,nopanic"`
	List   ListPDSBackupConfig   `copier:"must,nopanic"`
}

type PDSBackupConfigResponse struct {
	Create V1BackupConfig        `copier:"must,nopanic"`
	Update V1BackupConfig        `copier:"must,nopanic"`
	Get    V1BackupConfig        `copier:"must,nopanic"`
	List   ListPDSBackupResponse `copier:"must,nopanic"`
}

type ListPDSBackupResponse struct {
	BackupConfigs []V1BackupConfig
	Pagination    *V1PageBasedPaginationResponse `copier:"must,nopanic"`
}

type CreatePDSBackupConfig struct {
	ProjectId    string          `copier:"must,nopanic"`
	DeploymentId string          `copier:"must,nopanic"`
	BackupConfig *V1BackupConfig `copier:"must,nopanic"`
}

type UpdatePDSBackupConfig struct {
	ID          string
	Labels      *map[string]string `copier:"must,nopanic"`
	Annotations *map[string]string `copier:"must,nopanic"`
}

type GetPDSBackupConfig struct {
	Id string `copier:"must,nopanic"`
}

type DeletePDSBackupConfig struct {
	Id string `copier:"must,nopanic"`
}

type ListPDSBackupConfig struct {
	TenantId             string
	PaginationPageNumber string
	PaginationPageSize   string
	SortSortBy           string
	SortSortOrder        string
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

type BackupV1References1 struct {
	DeploymentId     *string `copier:"must,nopanic"`
	BackupLocationId *string `copier:"must,nopanic"`
	DataServiceId    *string `copier:"must,nopanic"`
}

type BackupV1Schedule struct {
	Id              *string `copier:"must,nopanic"`
	ResourceVersion *string `copier:"must,nopanic"`
}

type References struct {
	BackupLocationId *string `copier:"must,nopanic"`
}
