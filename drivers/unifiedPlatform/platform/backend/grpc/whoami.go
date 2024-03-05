package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	publicwhoamiapis "github.com/pure-px/apis/public/portworx/platform/whoami/apiv1"
	"google.golang.org/grpc"
)

// GetClient updates the header with bearer token and returns the new client
func (WhoAmIV1 *PlatformGrpc) getWhoAmIClient() (context.Context, publicwhoamiapis.WhoAmIServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var whoAmIClient publicwhoamiapis.WhoAmIServiceClient
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	whoAmIClient = publicwhoamiapis.NewWhoAmIServiceClient(WhoAmIV1.ApiClientV1)

	return ctx, whoAmIClient, token, nil
}

func (WhoAmIV1 *PlatformGrpc) WhoAmI() (WorkFlowResponse, error) {
	whoAmIResponse := WorkFlowResponse{}
	ctx, client, _, err := WhoAmIV1.getWhoAmIClient()
	if err != nil {
		return whoAmIResponse, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	apiResponse, err := client.WhoAmI(ctx, nil, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return whoAmIResponse, fmt.Errorf("Error calling whoAMI: %v\n", err)
	}

	err = copier.Copy(&whoAmIResponse, apiResponse)
	if err != nil {
		return whoAmIResponse, err
	}

	return whoAmIResponse, nil
}
