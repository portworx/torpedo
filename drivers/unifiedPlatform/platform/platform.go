package platform

import (
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

type Platform interface {
	AccountInterface
	TenantInterface
	TargetClusterInterface
	TargetClusterManifestInterface
	ApplicationInterface
	BackupLocationInterface
	CloudCredentialsInterface
	NamespaceInterface
	IamRoleBindingsInterface
	ServiceAccountsInterface
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
	ListTargetClusters(*WorkFlowRequest) ([]WorkFlowResponse, error)
	GetTarget(*WorkFlowRequest) (*WorkFlowResponse, error)
	PatchTargetCluster(*WorkFlowRequest) (*WorkFlowResponse, error)
	DeleteTarget(request *WorkFlowRequest) error
}

type TargetClusterManifestInterface interface {
	GetTargetClusterRegistrationManifest(getRequest *WorkFlowRequest) (string, error)
}

type ApplicationInterface interface {
	ListAllApplicationsInCluster(*WorkFlowRequest) ([]WorkFlowResponse, error)
	ListAvailableApplicationsForTenant(*WorkFlowRequest) ([]WorkFlowResponse, error)
	GetApplicationAtClusterLevel(*WorkFlowRequest) (*WorkFlowResponse, error)
	GetApplicationByAppId(*WorkFlowRequest) (*WorkFlowResponse, error)
	InstallApplication(*WorkFlowRequest) (*WorkFlowResponse, error)
	UninstallApplicationByAppId(*WorkFlowRequest) (*WorkFlowResponse, error)
	UninstallAppByAppIdClusterId(*WorkFlowRequest) (*WorkFlowResponse, error)
}

type BackupLocationInterface interface {
	ListBackupLocations() ([]WorkFlowResponse, error)
	GetBackupLocation(*WorkFlowRequest) (*WorkFlowResponse, error)
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

type IamRoleBindingsInterface interface {
	ListIamRoleBindings(*WorkFlowRequest) ([]WorkFlowResponse, error)
	CreateIamRoleBinding(*WorkFlowRequest) (*WorkFlowResponse, error)
	UpdateIamRoleBindings(*WorkFlowRequest) (*WorkFlowResponse, error)
	GetIamRoleBindingByID(*WorkFlowRequest) (*WorkFlowResponse, error)
	GrantIAMRoles(*WorkFlowRequest) (*WorkFlowResponse, error)
	RevokeAccessForIAM(*WorkFlowRequest) (*WorkFlowResponse, error)
	DeleteIamRoleBinding(*WorkFlowRequest) error
}

type ServiceAccountsInterface interface {
	ListAllServiceAccounts(*WorkFlowRequest) ([]WorkFlowResponse, error)
	GetServiceAccount(*WorkFlowRequest) (*WorkFlowResponse, error)
	CreateServiceAccount(*WorkFlowRequest) (*WorkFlowResponse, error)
	RegenerateServiceAccountSecret(*WorkFlowRequest) (*WorkFlowResponse, error)
	UpdateServiceAccount(*WorkFlowRequest) (*WorkFlowResponse, error)
	GenerateServiceAccountAccessToken(*WorkFlowRequest) (*WorkFlowResponse, error)
	DeleteServiceAccount(*WorkFlowRequest) error
}
