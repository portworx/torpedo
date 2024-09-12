package pds

import (
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
)

type Pds interface {
	Deployment
	DeploymentConfig
	BackupConfig
	Backup
	RestoreInterface
	TemplateDefinitionsInterface
	Catalog
}

type Deployment interface {
	CreateDeployment(*PDSDeploymentRequest) (*PDSDeploymentResponse, error)
	ListDeployment(string) (*PDSDeploymentResponse, error)
	GetDeployment(string) (*PDSDeploymentResponse, error)
	DeleteDeployment(string) error
	GetDeploymentCredentials(string) (string, error)
}

type DeploymentConfig interface {
	UpdateDeployment(updateRequest *PDSDeploymentRequest) (*PDSDeploymentResponse, error)
	GetDeploymentConfig(getRequest *PDSDeploymentRequest) (*PDSDeploymentResponse, error)
}

type BackupConfig interface {
	CreateBackupConfig(*PDSBackupConfigRequest) (*PDSBackupConfigResponse, error)
	UpdateBackupConfig(*PDSBackupConfigRequest) (*PDSBackupConfigResponse, error)
	GetBackupConfig(*PDSBackupConfigRequest) (*PDSBackupConfigResponse, error)
	DeleteBackupConfig(*PDSBackupConfigRequest) error
	ListBackupConfig(*PDSBackupConfigRequest) (*PDSBackupConfigResponse, error)
}

type Backup interface {
	DeleteBackup(*PDSBackupRequest) error
	GetBackup(*PDSBackupRequest) (*PDSBackupResponse, error)
	ListBackup(*PDSBackupRequest) (*PDSBackupResponse, error)
}

type Catalog interface {
	ListDataServices() (*CatalogResponse, error)
	ListDataServiceVersions(*WorkFlowRequest) (*CatalogResponse, error)
	ListDataServiceImages(*WorkFlowRequest) (*CatalogResponse, error)
}

type RestoreInterface interface {
	CreateRestore(*PDSRestoreRequest) (*PDSRestoreResponse, error)
	ReCreateRestore(*PDSRestoreRequest) (*PDSRestoreResponse, error)
	GetRestore(*PDSRestoreRequest) (*PDSRestoreResponse, error)
	ListRestore(*PDSRestoreRequest) (*PDSRestoreResponse, error)
}

type TemplateDefinitionsInterface interface {
	ListTemplateKinds() (*TemplateDefinitionResponse, error)
	ListTemplateRevisions() (*TemplateDefinitionResponse, error)
	GetTemplateRevisions() (*TemplateDefinitionResponse, error)
}
