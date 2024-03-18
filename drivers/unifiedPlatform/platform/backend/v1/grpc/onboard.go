package grpc

import (
	"context"
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	commonapiv1 "github.com/pure-px/apis/public/portworx/common/apiv1"
	publicaccountapis "github.com/pure-px/apis/public/portworx/platform/account/apiv1"
	publiconboardapis "github.com/pure-px/apis/public/portworx/platform/onboard/apiv1"
)

// GetClient updates the header with bearer token and returns the new client
func (onboard *PlatformGrpc) getOnboardClient() (context.Context, publiconboardapis.OnboardServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var onboardClient publiconboardapis.OnboardServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	onboardClient = publiconboardapis.NewOnboardServiceClient(onboard.ApiClientV1)

	return ctx, onboardClient, token, nil
}

func (onboard *PlatformGrpc) OnboardNewAccount(onboardAccountRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {

	response := &automationModels.WorkFlowResponse{}

	onboardRequest := publiconboardapis.CreateAccountRegistrationRequest{
		AccountRegistration: &publiconboardapis.AccountRegistration{
			Meta: &commonapiv1.Meta{
				Name: *onboardAccountRequest.OnboardAccount.Register.AccountRegistration.Meta.Name,
			},
			Config: &publiconboardapis.AccountConfig{
				AccountConfig: &publicaccountapis.Config{
					DisplayName: onboardAccountRequest.OnboardAccount.Register.AccountRegistration.Config.AccountConfig.DisplayName,
					UserEmail:   onboardAccountRequest.OnboardAccount.Register.AccountRegistration.Config.AccountConfig.UserEmail,
				},
			},
		},
	}

	ctx, client, _, err := onboard.getOnboardClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting grpc client: %v\n", err)
	}

	// TODO: Add support for opts if required
	apiResponse, err := client.CreateAccountRegistration(ctx, &onboardRequest, nil)
	if err != nil {
		return nil, fmt.Errorf("Some error occurred while calling CreateAccountRegistration: %v\n", err)
	}

	err = utilities.CopyStruct(apiResponse, response.OnboardAccount)

	return response, nil
}
