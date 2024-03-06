package api

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	platformv1 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// GetClient updates the header with bearer token and returns the new client
func (whoAmI *PLATFORM_API_V1) getWhoAmIClient() (context.Context, *platformv1.WhoAmIServiceAPIService, error) {
	log.Infof("Creating client from PLATFORM_API_V1 package")
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	whoAmI.ApiClientV1.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	whoAmI.ApiClientV1.GetConfig().DefaultHeader["px-account-id"] = whoAmI.AccountID

	client := whoAmI.ApiClientV1.WhoAmIServiceAPI
	return ctx, client, nil
}

func (WhoAmI *PLATFORM_API_V1) WhoAmI() (WorkFlowResponse, error) {
	ctx, client, err := WhoAmI.getWhoAmIClient()
	whoAmIResponse := WorkFlowResponse{}
	if err != nil {
		return whoAmIResponse, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	var getRequest platformv1.ApiWhoAmIServiceWhoAmIRequest
	getRequest = getRequest.ApiService.WhoAmIServiceWhoAmI(ctx)

	whoAmI, res, err := client.WhoAmIServiceWhoAmIExecute(getRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return whoAmIResponse, fmt.Errorf("Error when calling `\tWhoAmIServiceWhoAmIExecute\n`: %v\n.Full HTTP response: %v", err, res)
	}
	err = copier.Copy(&whoAmIResponse, whoAmI)
	return whoAmIResponse, nil
}
