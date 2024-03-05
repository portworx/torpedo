package pds

import (
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

type Pds interface {
	Deployment
	DeploymentConfig
}

type Deployment interface {
	CreateDeployment(depRequest *WorkFlowRequest) (*WorkFlowResponse, error)
}

type DeploymentConfig interface {
	UpdateDeployment(updateRequest *WorkFlowRequest) (*WorkFlowResponse, error)
}
