package platformLibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
)

func OnboardAccount(name string, displayName string, userEmail string) (*apiStructs.WorkFlowResponse, error) {

	onboardAccountRequest := apiStructs.WorkFlowRequest{}
	onboardAccountRequest.OnboardAccount = apiStructs.PlatformOnboardAccount{
		Register: &apiStructs.PlatformRegisterAccount{
			AccountRegistration: &apiStructs.AccountRegistration{
				Meta: &apiStructs.Meta{
					Name: &name,
				},
				Config: &apiStructs.AccountConfig{AccountConfig: &apiStructs.PlatformAccountConfig{
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
