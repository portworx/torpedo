package api

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
)

func (onboard *PLATFORM_API_V1) OnboardNewAccount(onboardAccountRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("OnboardNewAccount is not implemented for API")
	return &apiStructs.WorkFlowResponse{}, nil
}
