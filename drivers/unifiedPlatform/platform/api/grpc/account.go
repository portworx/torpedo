package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	publicaccountapis "github.com/pure-px/apis/public/portworx/platform/account/apiv1"
	"google.golang.org/grpc"
)

// AccountV2 struct
type GRPC struct {
	ApiClientV2 *grpc.ClientConn
}

var (
	credentials *Credentials
)

// GetClient updates the header with bearer token and returns the new client
func (grpc *GRPC) getClient() (context.Context, publicaccountapis.AccountServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var accountClient publicaccountapis.AccountServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	accountClient = publicaccountapis.NewAccountServiceClient(grpc.ApiClientV2)

	//AccountV2.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	//client := AccountV2.ApiClientV2.AccountServiceAPI

	return ctx, accountClient, token, nil
}

// GetAccountList returns the list of accounts
func (AccountV2 *GRPC) GetAccountList() ([]Account, error) {
	accountsResponse := []Account{}

	ctx, client, token, err := AccountV2.getClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	apiResponse, err := client.ListAccounts(ctx, nil, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error when calling `AccountServiceListAccounts`: %v\n.", err)
	}

	for _, acc := range apiResponse.Accounts {
		log.Infof("accounts - [%v]", acc.Meta.Name)
	}

	copier.Copy(&accountsResponse, apiResponse.Accounts)

	log.Infof("Value of accounts after copy - [%v]", accountsResponse)
	for _, acc := range accountsResponse {
		log.Infof("accounts - [%v]", acc.Meta.Name)
	}

	return accountsResponse, nil
}

// GetAccount return pds account model.
//func (AccountV2 *GRPC) GetAccount(accountID string) (Account, *status.Response, error) {
//	log.Infof("Get the account detail having UUID: %v", accountID)
//
//	accountResponse := Account{}
//
//	ctx, client, err := AccountV2.getClient()
//	if err != nil {
//		return accountResponse, nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
//	}
//	accountModel, res, err := client.AccountServiceGetAccount(ctx, accountID).Execute()
//
//	if err != nil && res.StatusCode != status.StatusOK {
//		return accountResponse, nil, fmt.Errorf("Error when calling `AccountServiceGetAccount`: %v\n.Full HTTP response: %v", err, res)
//	}
//
//	log.Infof("Value of account - [%v]", accountResponse)
//	copier.Copy(&accountResponse, accountModel)
//	log.Infof("Value of account after copy - [%v]", accountResponse)
//
//	return accountResponse, res, nil
//}
//
//// CreateAccount return pds account model.
//func (AccountV2 *GRPC) CreateAccount(accountName, displayName, userMail string) (Account, *status.Response, error) {
//	_, client, err := AccountV2.getClient()
//
//	accountResponse := Account{}
//
//	if err != nil {
//		return accountResponse, nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
//	}
//
//	var createRequest platformV2.ApiAccountServiceCreateAccountRequest
//	createRequest = createRequest.V1Account1(platformV2.V1Account1{
//		Meta: &platformV2.V1Meta{
//			Name: &accountName,
//		},
//		Config: &platformV2.V1Config6{
//			UserEmail:   &userMail,
//			DisplayName: &displayName,
//		},
//	})
//
//	accountModel, res, err := client.AccountServiceCreateAccountExecute(createRequest)
//
//	if err != nil && res.StatusCode != status.StatusOK {
//		return accountResponse, nil, fmt.Errorf("Error when calling `AccountServiceCreateAccount`: %v\n.Full HTTP response: %v", err, res)
//	}
//
//	log.Infof("Value of account - [%v]", accountResponse)
//	copier.Copy(&accountResponse, accountModel)
//	log.Infof("Value of account after copy - [%v]", accountResponse)
//
//	return accountResponse, res, nil
//}
//
//// DeleteBackupLocation delete backup location and return status.
//func (AccountV2 *GRPC) DeleteBackupLocation(accountId string) (*status.Response, error) {
//	ctx, client, err := AccountV2.getClient()
//	if err != nil {
//		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
//	}
//	_, res, err := client.AccountServiceDeleteAccount(ctx, accountId).Execute()
//	if err != nil {
//		return nil, fmt.Errorf("Error when calling `AccountServiceDeleteAccount`: %v\n.Full HTTP response: %v", err, res)
//	}
//	return res, nil
//}
