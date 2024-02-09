package grpc

import (
	"context"
	"fmt"

	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	commonapis "github.com/pure-px/apis/public/portworx/common/apiv1"
	publicaccountapis "github.com/pure-px/apis/public/portworx/platform/account/apiv1"
	"google.golang.org/grpc"
)

// AccountV2 struct
type PLATFORM_GRPC struct {
	ApiClientV2 *grpc.ClientConn
}

var (
	credentials *Credentials
)

// GetClient updates the header with bearer token and returns the new client
func (grpc *PLATFORM_GRPC) getClient() (context.Context, publicaccountapis.AccountServiceClient, string, error) {
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

func NewPaginationRequest(pageNumber, pageSize int) *commonapis.PageBasedPaginationRequest {
	return &commonapis.PageBasedPaginationRequest{
		PageNumber: int64(pageNumber),
		PageSize:   int64(pageSize),
	}
}

// GetAccountList returns the list of accounts
func (AccountV2 *PLATFORM_GRPC) GetAccountList() ([]Account, error) {
	accountsResponse := []Account{}

	ctx, client, token, err := AccountV2.getClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	firstPageRequest := &publicaccountapis.ListAccountsRequest{
		Pagination: NewPaginationRequest(1, 50),
	}

	apiResponse, err := client.ListAccounts(ctx, firstPageRequest, grpc.PerRPCCredentials(credentials))
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
//func (AccountV2 *GRPC) GetAccount(accountID string) (*Account, error) {
//	accountsResponse := []Account{}
//
//	ctx, client, token, err := AccountV2.getClient()
//	if err != nil {
//		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
//	}
//
//	credentials = &Credentials{
//		Token: token,
//	}
//
//	client.GetAccount(ctx)
//
//	return nil, err
//
//}

//
// CreateAccount return pds account model.
//func (AccountV2 *GRPC) CreateAccount(accountName, displayName, userMail string) (*Account, error) {
//	_, client, token, err := AccountV2.getClient()
//	if err != nil {
//		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
//	}
//
//	credentials = &Credentials{
//		Token: token,
//	}
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
