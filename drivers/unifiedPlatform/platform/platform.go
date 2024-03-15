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
	TemplatesInterface
	Onboard
}

type AccountInterface interface {
	//GetAccountList() ([]WorkFlowResponse, error) // not used as of now
	GetAccount(string) (*WorkFlowResponse, error)
}

type TenantInterface interface {
	ListTenants() ([]WorkFlowResponse, error)
}

type TargetClusterInterface interface {
	ListTargetClusters(*WorkFlowRequest) ([]WorkFlowResponse, error)
	GetTarget(*WorkFlowRequest) (*WorkFlowResponse, error)
	PatchTargetCluster(*WorkFlowRequest) (*WorkFlowResponse, error)
	DeleteTargetCluster(request *WorkFlowRequest) error
}

type TargetClusterManifestInterface interface {
	GetTargetClusterRegistrationManifest(*WorkFlowRequest) (*WorkFlowResponse, error)
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
	ListBackupLocations(*BackupLocation) ([]BackupLocation, error)
	GetBackupLocation(*WorkFlowRequest) (*WorkFlowResponse, error)
	CreateBackupLocation(*BackupLocation) (*BackupLocation, error)
	UpdateBackupLocation(*WorkFlowRequest) (*WorkFlowResponse, error)
	DeleteBackupLocation(*WorkFlowRequest) error
}

type CloudCredentialsInterface interface {
	ListCloudCredentials(*CloudCredentials) ([]CloudCredentials, error)
	GetCloudCredentials(*CloudCredentials) (*CloudCredentials, error)
	CreateCloudCredentials(*CloudCredentials) (*CloudCredentials, error)
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

type TemplatesInterface interface {
	ListTemplates(*WorkFlowRequest) ([]WorkFlowResponse, error)
	CreateTemplates(*WorkFlowRequest) (*WorkFlowResponse, error)
	UpdateTemplates(*WorkFlowRequest) (*WorkFlowResponse, error)
	DeleteTemplate(*WorkFlowRequest) error
}

type Onboard interface {
	OnboardNewAccount(*WorkFlowRequest) (*WorkFlowResponse, error)
}
