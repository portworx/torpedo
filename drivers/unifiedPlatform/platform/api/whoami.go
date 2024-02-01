package api

import (
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	platformV2 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// WhoAmiV2 struct
type WhoAmiV2 struct {
	ApiClientV2 *platformV2.APIClient
}

func (whoami *WhoAmiV2) WhoAmi() (*platformV2.V1WhoAmIResponse, error) {
	client := whoami.ApiClientV2.WhoAmIServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	whoamiResp, res, err := client.WhoAmIServiceWhoAmI(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `WhoAmIServiceWhoAmI`: %v\n.Full HTTP response: %v", err, res)
	}
	return whoamiResp, nil
}
