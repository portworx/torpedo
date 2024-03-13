package api

import (
	"context"
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	platformv1 "github.com/pure-px/platform-api-go-client/v1alpha1"
)

// GetClient updates the header with bearer token and returns the new client
func (onboard *PLATFORM_API_V1) getOnboardClient() (context.Context, *platformv1.TenantServiceAPIService, error) {
	log.Infof("Creating client from PLATFORM_API_V1 package")
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	onboard.ApiClientV1.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	onboard.ApiClientV1.GetConfig().DefaultHeader["px-account-id"] = onboard.AccountID

	client := onboard.ApiClientV1.TenantServiceAPI
	return ctx, client, nil
}

func (onboard *PLATFORM_API_V1) OnboardNewAccount(onboardAccountRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("OnboardNewAccount is not implemented for API")
	return &apiStructs.WorkFlowResponse{}, nil
}
