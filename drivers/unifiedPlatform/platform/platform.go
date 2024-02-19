package platform

import (
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

type AccountInterface interface {
	GetAccountList() ([]WorkFlowResponse, error)
	GetAccount(string) (*WorkFlowResponse, error)
	CreateAccount(string, string, string) (WorkFlowResponse, error)
	DeleteAccount(string) error
}

type TenantInterface interface {
	ListTenants(string) ([]WorkFlowResponse, error)
}

type ApplicationInterface interface {
	ListAllApplicationsInCluster(string) ([]WorkFlowResponse, error)
	ListAvailableApplicationsForTenant(string) ([]WorkFlowResponse, error)
	GetApplicationAtClusterLevel(string, string) (*WorkFlowResponse, error)
	GetApplicationByAppId(string) (*WorkFlowResponse, error)
	InstallApplication(*WorkFlowRequest, string) (*WorkFlowResponse, error)
	UninstallApplicationByAppId(string, *WorkFlowRequest) (*WorkFlowResponse, error)
	UninstallAppByAppIdClusterId(string, string, *WorkFlowRequest) (*WorkFlowResponse, error)
}

type BackupLocationInterface interface {
	ListBackupLocations() ([]WorkFlowResponse, error)
	GetBackupLocation(string) (*WorkFlowResponse, error)
	CreateBackupLocation(string, *WorkFlowRequest) (*WorkFlowResponse, error)
	UpdateBackupLocation(*WorkFlowRequest, string) (*WorkFlowResponse, error)
	DeleteBackupLocation(string) error
}

type CloudCredentialsInterface interface {
	ListCloudCredentials() ([]WorkFlowResponse, error)
	GetCloudCredentials(string) (*WorkFlowResponse, error)
	CreateCloudCredentials(*WorkFlowRequest, string) (*WorkFlowResponse, error)
	UpdateCloudCredentials(*WorkFlowRequest, string) (*WorkFlowResponse, error)
	DeleteCloudCredential(string) error
}
type NamespaceInterface interface {
	ListNamespaces(string) ([]WorkFlowResponse, error)
}

type Platform interface {
	AccountInterface
	TenantInterface
	ApplicationInterface
	BackupLocationInterface
	CloudCredentialsInterface
	NamespaceInterface
}
