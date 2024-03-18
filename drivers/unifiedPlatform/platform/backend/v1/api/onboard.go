package api

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
)

func (onboard *PLATFORM_API_V1) OnboardNewAccount(onboardAccountRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	log.Warnf("OnboardNewAccount is not implemented for API")
	return &automationModels.WorkFlowResponse{}, nil
}
