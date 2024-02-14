package platform

import (
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

type AccountInterface interface {
	GetAccountList() ([]ApiResponse, error)
	GetAccount(string) (*ApiResponse, error)
	CreateAccount(string, string, string) (ApiResponse, error)
	DeleteBackupLocation(string) error
}

type TenantInterface interface {
	ListTenants(string) ([]ApiResponse, error)
}

type Platform interface {
	AccountInterface
	TenantInterface
}
