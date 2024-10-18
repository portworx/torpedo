package platformLibs

import (
	"github.com/pure-px/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/pure-px/torpedo/pkg/log"
)

func OnboardAccount(name string, displayName string, userEmail string) (*automationModels.PlatformOnboardAccountResponse, error) {

	onboardAccountRequest := automationModels.PlatformOnboardAccountRequest{
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

	log.Infof("Onboard Request - [%+v]", onboardAccountRequest)

	response, err := v2Components.Platform.OnboardNewAccount(&onboardAccountRequest)
	log.Infof("Onboard Response - [%+v]", response)
	if err != nil {
		return response, err
	}
	return response, nil
}
