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
	ListTargetClusters(*PlatformTargetClusterRequest) (*PlatformTargetClusterResponse, error)
	GetTargetCluster(*PlatformTargetClusterRequest) (*PlatformTargetClusterResponse, error)
	// TODO: Uncomment this method when it is implemented with validation
	//PatchTargetCluster(*PlatformTargetClusterRequest) (*WorkFlowResponse, error)
	DeleteTargetCluster(request *PlatformTargetClusterRequest) error
}

type TargetClusterManifestInterface interface {
	GetTargetClusterRegistrationManifest(cluster *PlatformTargetClusterRequest) (*PlatformTargetClusterResponse, error)
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
	GetProjectList(int, int) (*PlaformProjectResponse, error)
	CreateProject(*PlaformProjectRequest, string) (*PlaformProjectResponse, error)
	DeleteProject(*PlaformProjectRequest) error
	GetProject(*PlaformProjectRequest) (*PlaformProjectResponse, error)
	AssociateToProject(*PlaformProjectRequest) (*PlaformProjectResponse, error)
	DissociateFromProject(*PlaformProjectRequest) (*PlaformProjectResponse, error)
}
