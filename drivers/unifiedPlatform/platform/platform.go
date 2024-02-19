package platform

import (
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

type Platform interface {
	AccountInterface
	TenantInterface
	TargetClusterInterface
}

type AccountInterface interface {
	GetAccountList() ([]WorkFlowResponse, error)
	GetAccount(string) (*WorkFlowResponse, error)
	CreateAccount(string, string, string) (WorkFlowResponse, error)
	DeleteBackupLocation(string) error
}

type TenantInterface interface {
	ListTenants(string) ([]WorkFlowResponse, error)
}

type TargetClusterInterface interface {
	ListTargetClusters() ([]WorkFlowResponse, error)
	GetTarget(*WorkFlowRequest) (*WorkFlowResponse, error)
	PatchTargetCluster(*WorkFlowRequest) (*WorkFlowResponse, error)
	DeleteTarget(request *WorkFlowRequest) error
}

type TargetClusterManifestInterface interface {
	GetTargetClusterRegistrationManifest(getRequest *WorkFlowRequest) (string, error)
}
