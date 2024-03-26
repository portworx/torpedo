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
	CreateDeployment(*PDSDeploymentRequest) (*WorkFlowResponse, error)
	ListDeployment() (*WorkFlowResponse, error)
	GetDeployment(string) (*WorkFlowResponse, error)
	DeleteDeployment(string) error
}

type DeploymentConfig interface {
	UpdateDeployment(updateRequest *PDSDeploymentRequest) (*WorkFlowResponse, error)
}

type BackupConfig interface {
	CreateBackupConfig(*WorkFlowRequest) (*WorkFlowResponse, error)
	UpdateBackupConfig(*WorkFlowRequest) (*WorkFlowResponse, error)
	GetBackupConfig(*WorkFlowRequest) (*WorkFlowResponse, error)
	DeleteBackupConfig(*WorkFlowRequest) (*WorkFlowResponse, error)
	ListBackupConfig(*WorkFlowRequest) ([]WorkFlowResponse, error)
}

type Backup interface {
	DeleteBackup(*WorkFlowRequest) (*WorkFlowResponse, error)
	GetBackup(*WorkFlowRequest) (*WorkFlowResponse, error)
	ListBackup(*WorkFlowRequest) ([]PDSBackupResponse, error)
}

type Catalog interface {
	ListDataServices() ([]WorkFlowResponse, error)
	ListDataServiceVersions(*WorkFlowRequest) ([]WorkFlowResponse, error)
	ListDataServiceImages(*WorkFlowRequest) ([]WorkFlowResponse, error)
}

type RestoreInterface interface {
	CreateRestore(*WorkFlowRequest) (*WorkFlowResponse, error)
	ReCreateRestore(*WorkFlowRequest) (*WorkFlowResponse, error)
	GetRestore(*WorkFlowRequest) (*Restore, error)
	DeleteRestore(*WorkFlowRequest) (*WorkFlowResponse, error)
	ListRestore(*WorkFlowRequest) ([]WorkFlowResponse, error)
}

type TemplateDefinitionsInterface interface {
	ListTemplateKinds() (*TemplateDefinitionResponse, error)
	ListTemplateRevisions() (*TemplateDefinitionResponse, error)
	GetTemplateRevisions() (*TemplateDefinitionResponse, error)
}
