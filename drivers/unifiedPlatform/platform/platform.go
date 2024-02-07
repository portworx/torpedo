package platform

import (
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	status "net/http"
)

type Platform interface {
	GetAccountList() ([]Account, *status.Response, error)
	GetAccount(accountID string) (Account, *status.Response, error)
	CreateAccount(accountName, displayName, userMail string) (Account, *status.Response, error)
	DeleteBackupLocation(accountId string) (*status.Response, error)
}
