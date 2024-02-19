package platform

import (
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

type Platform interface {
	AccountInterface
	TenantInterface
	TargetClusterInterface
	ApplicationInterface
	BackupLocationInterface
	CloudCredentialsInterface
	NamespaceInterface
}

type AccountInterface interface {
	GetAccountList() ([]WorkFlowResponse, error)
	GetAccount(string) (*WorkFlowResponse, error)
	CreateAccount(string, string, string) (WorkFlowResponse, error)
	DeleteAccount(string) error
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

type ApplicationInterface interface {
	ListAllApplicationsInCluster(*WorkFlowResponse) ([]WorkFlowResponse, error)
	ListAvailableApplicationsForTenant(*WorkFlowResponse) ([]WorkFlowResponse, error)
	GetApplicationAtClusterLevel(*WorkFlowResponse) (*WorkFlowResponse, error)
	GetApplicationByAppId(*WorkFlowResponse) (*WorkFlowResponse, error)
	InstallApplication(*WorkFlowRequest) (*WorkFlowResponse, error)
	UninstallApplicationByAppId(*WorkFlowRequest) (*WorkFlowResponse, error)
	UninstallAppByAppIdClusterId(*WorkFlowRequest) (*WorkFlowResponse, error)
}

type BackupLocationInterface interface {
	ListBackupLocations() ([]WorkFlowResponse, error)
	GetBackupLocation(*WorkFlowResponse) (*WorkFlowResponse, error)
	CreateBackupLocation(*WorkFlowRequest) (*WorkFlowResponse, error)
	UpdateBackupLocation(*WorkFlowRequest) (*WorkFlowResponse, error)
	DeleteBackupLocation(*WorkFlowRequest) error
}

type CloudCredentialsInterface interface {
	ListCloudCredentials() ([]WorkFlowResponse, error)
	GetCloudCredentials(*WorkFlowRequest) (*WorkFlowResponse, error)
	CreateCloudCredentials(*WorkFlowRequest) (*WorkFlowResponse, error)
	UpdateCloudCredentials(*WorkFlowRequest) (*WorkFlowResponse, error)
	DeleteCloudCredential(*WorkFlowRequest) error
}
type NamespaceInterface interface {
	ListNamespaces(*WorkFlowRequest) ([]WorkFlowResponse, error)
}
