package platform

import (
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
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
	Project
}

type AccountInterface interface {
	//GetAccountList() ([]WorkFlowResponse, error) // not used as of now
	GetAccount(*PlatformAccount) (*WorkFlowResponse, error)
}

type TenantInterface interface {
	ListTenants() ([]WorkFlowResponse, error)
}

type TargetClusterInterface interface {
	ListTargetClusters(*PlatformTargetCluster) ([]WorkFlowResponse, error)
	GetTargetCluster(*PlatformTargetCluster) (*WorkFlowResponse, error)
	PatchTargetCluster(*PlatformTargetCluster) (*WorkFlowResponse, error)
	DeleteTargetCluster(request *PlatformTargetCluster) error
}

type TargetClusterManifestInterface interface {
	GetTargetClusterRegistrationManifest(cluster *PlatformTargetCluster) (*WorkFlowResponse, error)
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
	CreateBackupLocation(*BackupLocationRequest) (*BackupLocationResponse, error)
	ListBackupLocations(*BackupLocationRequest) ([]*BackupLocationResponse, error)
	DeleteBackupLocation(*BackupLocationRequest) error
	GetBackupLocation(*WorkFlowRequest) (*WorkFlowResponse, error)
	UpdateBackupLocation(*WorkFlowRequest) (*WorkFlowResponse, error)
}

type CloudCredentialsInterface interface {
	ListCloudCredentials(*CloudCredentials) ([]CloudCredentials, error)
	GetCloudCredentials(*CloudCredentials) (*CloudCredentialsResponse, error)
	CreateCloudCredentials(*CloudCredentials) (*CloudCredentialsResponse, error)
	DeleteCloudCredential(*CloudCredentials) error
	UpdateCloudCredentials(*WorkFlowRequest) (*WorkFlowResponse, error)
}
type NamespaceInterface interface {
	ListNamespaces(namespace *PlatformNamespace) (*PlatformNamespaceResponse, error)
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
	ListTemplates(*PlatformTemplatesRequest) ([]PlatformTemplatesResponse, error)
	CreateTemplates(*PlatformTemplatesRequest) (*PlatformTemplatesResponse, error)
	UpdateTemplates(*PlatformTemplatesRequest) (*PlatformTemplatesResponse, error)
	GetTemplates(*PlatformTemplatesRequest) (*PlatformTemplatesResponse, error)
	DeleteTemplate(*PlatformTemplatesRequest) error
}

type Onboard interface {
	OnboardNewAccount(*WorkFlowRequest) (*WorkFlowResponse, error)
}

type Project interface {
	GetProjectList() ([]WorkFlowResponse, error)
	CreateProject(*PlaformProject, string) (WorkFlowResponse, error)
	DeleteProject(*PlaformProject) error
	GetProject(*PlaformProject) (WorkFlowResponse, error)
	AssociateToProject(*PlaformProject) (WorkFlowResponse, error)
	DissociateFromProject(*PlaformProject) (WorkFlowResponse, error)
}
