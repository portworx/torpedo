package platform

import (
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

//type Account interface {
//	GetAccountList() ([]Account, *status.Response, error)
//	GetAccount(accountID string) (Account, *status.Response, error)
//	CreateAccount(accountName, displayName, userMail string) (Account, *status.Response, error)
//	DeleteBackupLocation(accountId string) (*status.Response, error)
//}
//
//type Tenant interface {
//}

type AccountInterface interface {
	GetAccountList() ([]Account, error)
}

type TenantInterface interface {
	GetTenantList()
}

type Platform interface {
	AccountInterface
	TenantInterface
	//GetAccount(accountID string) (Account, *status.Response, error)
	//CreateAccount(accountName, displayName, userMail string) (Account, *status.Response, error)
	//DeleteBackupLocation(accountId string) (*status.Response, error)
}
