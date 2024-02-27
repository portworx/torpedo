package pds

import (
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

type Pds interface {
	Deployment
	DeploymentConfig
	Backup
	BackupConfig
	Restore
}

type Deployment interface {
	CreateDeployment(depRequest *WorkFlowRequest) (*WorkFlowResponse, error)
}

type DeploymentConfig interface {
	UpdateDeploymentConfig(updateRequest *WorkFlowRequest) (*WorkFlowResponse, error)
}

type BackupConfig interface {
	CreateBackupConfig(*WorkFlowRequest) (*WorkFlowResponse, error)
	UpdateBackupConfig(*WorkFlowRequest) (*WorkFlowResponse, error)
	GetBackupConfig(*WorkFlowRequest) (*WorkFlowResponse, error)
	DeleteBackupConfig(*WorkFlowRequest) (*WorkFlowResponse, error)
	ListBackupConfig(*WorkFlowRequest) ([]WorkFlowResponse, error)
}

type Backup interface {
	GetBackup(*WorkFlowRequest) (*WorkFlowResponse, error)
	DeleteBackup(*WorkFlowRequest) (*WorkFlowResponse, error)
	ListBackup(*WorkFlowRequest) ([]WorkFlowResponse, error)
}

type Restore interface {
	CreateRestore(*WorkFlowRequest) (*WorkFlowResponse, error)
	ReCreateRestore(*WorkFlowRequest) (*WorkFlowResponse, error)
	GetRestore(*WorkFlowRequest) (*WorkFlowResponse, error)
	DeleteRestore(*WorkFlowRequest) (*WorkFlowResponse, error)
	ListRestore(*WorkFlowRequest) ([]WorkFlowResponse, error)
}
