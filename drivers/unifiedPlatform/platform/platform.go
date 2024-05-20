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
	WhoAmI
}

type AccountInterface interface {
	//GetAccountList() ([]WorkFlowResponse, error) // not used as of now
	GetAccount(*PlatformAccount) (*PlatformAccountResponse, error)
}

type TenantInterface interface {
	ListTenants() ([]PlatformTenant, error)
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
	ListBackupLocations(*BackupLocationRequest) (*BackupLocationResponse, error)
	DeleteBackupLocation(*BackupLocationRequest) error
	GetBackupLocation(*WorkFlowRequest) (*BackupLocationResponse, error)
	UpdateBackupLocation(*WorkFlowRequest) (*BackupLocationResponse, error)
}

type CloudCredentialsInterface interface {
	ListCloudCredentials(*CloudCredentialsRequest) (*CloudCredentialsResponse, error)
	GetCloudCredentials(*CloudCredentialsRequest) (*CloudCredentialsResponse, error)
	CreateCloudCredentials(*CloudCredentialsRequest) (*CloudCredentialsResponse, error)
	DeleteCloudCredential(*CloudCredentialsRequest) error
	UpdateCloudCredentials(*CloudCredentialsRequest) (*CloudCredentialsResponse, error)
}
type NamespaceInterface interface {
	ListNamespaces(*PlatformNamespace) (*PlatformNamespaceResponse, error)
	DeleteNamespace(*PlatformNamespace) error
}

type IamRoleBindingsInterface interface {
	ListIamRoleBindings(*IAMRequest) (*IAMResponse, error)
	CreateIamRoleBinding(*IAMRequest) (*IAMResponse, error)
	UpdateIamRoleBindings(*IAMRequest) (*IAMResponse, error)
	GetIamRoleBindingByID(*IAMRequest) (*IAMResponse, error)
	GrantIAMRoles(*IAMRequest) (*IAMResponse, error)
	RevokeAccessForIAM(*IAMRequest) (*IAMResponse, error)
	DeleteIamRoleBinding(*IAMRequest) error
}

type ServiceAccountsInterface interface {
	ListAllServiceAccounts(*PDSServiceAccountRequest) (*PDSServiceAccountResponse, error)
	GetServiceAccount(*PDSServiceAccountRequest) (*PDSServiceAccountResponse, error)
	CreateServiceAccount(*PDSServiceAccountRequest) (*PDSServiceAccountResponse, error)
	RegenerateServiceAccountSecret(*PDSServiceAccountRequest) (*PDSServiceAccountResponse, error)
	UpdateServiceAccount(*PDSServiceAccountRequest) (*PDSServiceAccountResponse, error)
	GenerateServiceAccountAccessToken(*PDSServiceAccountRequest) (*PDSServiceAccountResponse, error)
	DeleteServiceAccount(*PDSServiceAccountRequest) error
}

type TemplatesInterface interface {
	ListTemplates(*PlatformTemplatesRequest) (*PlatformTemplatesResponse, error)
	CreateTemplates(*PlatformTemplatesRequest) (*PlatformTemplatesResponse, error)
	UpdateTemplates(*PlatformTemplatesRequest) (*PlatformTemplatesResponse, error)
	GetTemplates(*PlatformTemplatesRequest) (*PlatformTemplatesResponse, error)
	DeleteTemplate(*PlatformTemplatesRequest) error
}

type Onboard interface {
	OnboardNewAccount(*PlatformOnboardAccountRequest) (*PlatformOnboardAccountResponse, error)
}

type Project interface {
	GetProjectList(*PlaformProjectRequest) (*PlaformProjectResponse, error)
	CreateProject(*PlaformProjectRequest, string) (*PlaformProjectResponse, error)
	DeleteProject(*PlaformProjectRequest) error
	GetProject(*PlaformProjectRequest) (*PlaformProjectResponse, error)
	AssociateToProject(*PlaformProjectRequest) (*PlaformProjectResponse, error)
	DissociateFromProject(*PlaformProjectRequest) (*PlaformProjectResponse, error)
}

type WhoAmI interface {
	WhoAmI() (*WhoamiResponse, error)
}
