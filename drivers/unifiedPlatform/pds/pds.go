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
}

type Deployment interface {
	CreateDeployment(depRequest *WorkFlowRequest) (*WorkFlowResponse, error)
	ListDeployment() (*WorkFlowResponse, error)
	GetDeployment(string) (*WorkFlowResponse, error)
	DeleteDeployment(string) (*WorkFlowResponse, error)
}

type DeploymentConfig interface {
	UpdateDeployment(updateRequest *WorkFlowRequest) (*WorkFlowResponse, error)
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
	ListBackup(*WorkFlowRequest) ([]WorkFlowResponse, error)
}

type RestoreInterface interface {
	CreateRestore(*WorkFlowRequest) (*WorkFlowResponse, error)
	ReCreateRestore(*WorkFlowRequest) (*WorkFlowResponse, error)
	GetRestore(*WorkFlowRequest) (*WorkFlowResponse, error)
	DeleteRestore(*WorkFlowRequest) (*WorkFlowResponse, error)
	ListRestore(*WorkFlowRequest) ([]WorkFlowResponse, error)
}

type TemplateDefinitionsInterface interface {
	ListTemplateKinds() ([]WorkFlowResponse, error)
	ListTemplateRevisions() ([]WorkFlowResponse, error)
	GetTemplateRevisions() (*WorkFlowResponse, error)
}
