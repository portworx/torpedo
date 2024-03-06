package pds

import (
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

type Pds interface {
	Deployment
	DeploymentConfig
	BackupConfig
	Backup
}

type Deployment interface {
	CreateDeployment(depRequest *WorkFlowRequest) (*WorkFlowResponse, error)
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
