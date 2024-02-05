package api

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	platformV2 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// WhoAmI struct
type WhoAmI struct {
	ApiClientV2 *platformV2.APIClient
	AccountID   string
}

// GetClient updates the header with bearer token and returns the new client
func (WhoAmI *WhoAmI) GetClient() (context.Context, *platformV2.WhoAmIServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	log.Infof("Token %s", token)
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	WhoAmI.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	WhoAmI.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = WhoAmI.AccountID
	client := WhoAmI.ApiClientV2.WhoAmIServiceAPI

	return ctx, client, nil
}

func (WhoamI *WhoAmI) WhoAmI() (*platformV2.V1WhoAmIResponse, error) {
	ctx, client, err := WhoamI.GetClient()
	if err != nil {
		return nil, err
	}
	whoAmiResp, res, err := client.WhoAmIServiceWhoAmI(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `AccountServiceListAccounts`: %v\n.Full HTTP response: %v", err, res)
	}
	return whoAmiResp, nil
}
