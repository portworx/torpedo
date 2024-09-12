package api

import (
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	whoamiv1 "github.com/pure-px/platform-api-go-client/platform/v1/whoami"
	status "net/http"
)

func (WhoAmI *PLATFORM_API_V1) WhoAmI() (*WhoamiResponse, error) {
	ctx, client, err := WhoAmI.getWhoAmIClient()
	whoAmIResponse := WhoamiResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	var getRequest whoamiv1.ApiWhoAmIServiceWhoAmIRequest
	getRequest = getRequest.ApiService.WhoAmIServiceWhoAmI(ctx)

	whoAmI, res, err := client.WhoAmIServiceWhoAmIExecute(getRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `\tWhoAmIServiceWhoAmIExecute\n`: %v\n.Full HTTP response: %v", err, res)
	}
	err = copier.Copy(&whoAmIResponse, whoAmI)
	return &whoAmIResponse, nil
}
