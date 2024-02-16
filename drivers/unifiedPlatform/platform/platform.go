package platform

import (
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

type AccountInterface interface {
	GetAccountList() ([]ApiResponse, error)
	GetAccount(string) (*ApiResponse, error)
	CreateAccount(string, string, string) (ApiResponse, error)
	DeleteAccount(string) error
}

type TenantInterface interface {
	ListTenants(string) ([]ApiResponse, error)
}

type ApplicationInterface interface {
	ListAllApplicationsInCluster(string) ([]ApiResponse, error)
	ListAvailableApplicationsForTenant(string) ([]ApiResponse, error)
	GetApplicationAtClusterLevel(string, string) (*ApiResponse, error)
	GetApplicationByAppId(string) (*ApiResponse, error)
	InstallApplication(string, string, string, string) (*ApiResponse, error)
	UninstallApplicationByAppId(string) (*ApiResponse, error)
	UninstallAppByAppIdClusterId(string, string) error
}

type NamespaceInterface interface {
	ListNamespaces(string) ([]ApiResponse, error)
}

type BackupLocationInterface interface {
	GetBackupLocClient() ([]ApiResponse, error)
	ListBackupLocations() ([]ApiResponse, error)
	GetBackupLocation(string) (*ApiResponse, error)
	CreateBackupLocation(string, *BackupLocationParams)
}

type Platform interface {
	AccountInterface
	TenantInterface
	ApplicationInterface
	NamespaceInterface
	BackupLocationInterface
}
