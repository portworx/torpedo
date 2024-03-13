package platformLibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
)

func OnboardAccount(name string, displayName string, userEmail string) (*automationModels.WorkFlowResponse, error) {

	onboardAccountRequest := automationModels.WorkFlowRequest{}
	onboardAccountRequest.OnboardAccount = automationModels.PlatformOnboardAccount{
		Register: &automationModels.PlatformRegisterAccount{
			AccountRegistration: &automationModels.AccountRegistration{
				Meta: &automationModels.Meta{
					Name: &name,
				},
				Config: &automationModels.AccountConfig{AccountConfig: &automationModels.PlatformAccountConfig{
					DisplayName: displayName,
					UserEmail:   userEmail,
				}},
			},
		},
	}

	log.Infof("Onboard Request - [%+v]", onboardAccountRequest.OnboardAccount)

	response, err := v2Components.Platform.OnboardNewAccount(&onboardAccountRequest)
	if err != nil {
		return response, err
	}
	return response, nil
}
