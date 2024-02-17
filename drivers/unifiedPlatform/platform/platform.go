package platform

import (
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

type AccountInterface interface {
	GetAccountList() ([]WorkFlowResponse, error)
	GetAccount(string) (*WorkFlowResponse, error)
	CreateAccount(string, string, string) (WorkFlowResponse, error)
	DeleteBackupLocation(string) error
}

type TenantInterface interface {
	ListTenants(string) ([]WorkFlowResponse, error)
}

type Platform interface {
	AccountInterface
	TenantInterface
}
