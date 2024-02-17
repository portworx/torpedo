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
type PlatformGrpc struct {
	ApiClientV1 *grpc.ClientConn
}

// GetClient updates the header with bearer token and returns the new client
func (AccountV1 *PlatformGrpc) getAccountClient() (context.Context, publicaccountapis.AccountServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var accountClient publicaccountapis.AccountServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	accountClient = publicaccountapis.NewAccountServiceClient(AccountV1.ApiClientV1)

	return ctx, accountClient, token, nil
}

func NewPaginationRequest(pageNumber, pageSize int) *commonapis.PageBasedPaginationRequest {
	return &commonapis.PageBasedPaginationRequest{
		PageNumber: int64(pageNumber),
		PageSize:   int64(pageSize),
	}
}

// GetAccountList returns the list of accounts
func (AccountV1 *PlatformGrpc) GetAccountList() ([]WorkFlowResponse, error) {
	accountsResponse := []WorkFlowResponse{}

	ctx, client, _, err := AccountV1.getAccountClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
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

func (AccountV1 *PlatformGrpc) GetAccount(accountID string) (*WorkFlowResponse, error) {
	accountsResponse := WorkFlowResponse{}
	ctx, client, _, err := AccountV1.getAccountClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	getAccRequest := &publicaccountapis.GetAccountRequest{
		AccountId: accountID,
	}

	ctx = WithAccountIDMetaCtx(ctx, accountID)

	apiResponse, err := client.GetAccount(ctx, getAccRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error while getting the account: %v\n", err)
	}

	log.Infof("Value of accounts before copy - [%v]", apiResponse.Meta.Name)
	copier.Copy(&accountsResponse, apiResponse)
	log.Infof("Value of accounts after copy - [%v]", *accountsResponse.Meta.Name)

	return &accountsResponse, nil
}

// CreateAccount return pds account model.
func (AccountV1 *PlatformGrpc) CreateAccount(accountName, displayName, userMail string) (WorkFlowResponse, error) {
	var accResponse WorkFlowResponse
	return accResponse, fmt.Errorf("Create Account is not implemented in grpc client\n")
}

func (AccountV1 *PlatformGrpc) DeleteBackupLocation(accountId string) error {
	return fmt.Errorf("DeleteBackuplocation is not implemented in grpc client\n")
}
