package api

import (
	"fmt"
	status "net/http"

	"github.com/jinzhu/copier"

	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
	accountv1 "github.com/pure-px/platform-api-go-client/platform/v1/account"
)

// GetAccount return pds account model.
func (AccountV1 *PLATFORM_API_V1) GetAccount(accountRequest *PlatformAccount) (*WorkFlowResponse, error) {
	log.Infof("Get the account detail having UUID: %v", accountRequest.Get.AccountId)
	accountResponse := WorkFlowResponse{}
	ctx, client, err := AccountV1.getAccountClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	var getRequest accountv1.ApiAccountServiceGetAccountRequest
	getRequest = getRequest.ApiService.AccountServiceGetAccount(ctx, accountRequest.Get.AccountId)
	accountModel, res, err := client.AccountServiceGetAccountExecute(getRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `AccountServiceGetAccount`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of account - [%v]", *accountModel.Meta.Name)
	err = copier.Copy(&accountResponse, accountModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of account after copy - [%v]", *accountResponse.Meta.Name)
	return &accountResponse, nil
}
