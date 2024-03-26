package api

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	onboardv1 "github.com/pure-px/platform-api-go-client/platform/v1/onboard"
	status "net/http"
)

func (onboard *PLATFORM_API_V1) OnboardNewAccount(onboardAccountRequest *automationModels.PlatformOnboardAccountRequest) (*automationModels.PlatformOnboardAccountResponse, error) {
	registerationResponse := automationModels.PlatformOnboardAccountResponse{
		Register: &automationModels.AccountRegistration{},
	}
	ctx, client, err := onboard.getOnboardClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	v1Account := onboardv1.V1AccountRegistration{}
	log.Infof("V1 Account before copy - [%+v]", v1Account)
	utilities.CopyStruct(onboardAccountRequest.Register.AccountRegistration, &v1Account)
	log.Infof("V1 Account After copy - [%+v]", v1Account)

	request := client.OnboardServiceCreateAccountRegistration(ctx)
	request = request.V1AccountRegistration(v1Account)

	registerationModel, res, err := request.Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `AccountServiceGetAccount`: %v\n.Full HTTP response: %v", err, res)
	}

	log.Infof("Registeration Model - [%+v]", registerationModel)

	err = utilities.CopyStruct(registerationModel, &registerationResponse.Register)
	return &registerationResponse, nil
}
