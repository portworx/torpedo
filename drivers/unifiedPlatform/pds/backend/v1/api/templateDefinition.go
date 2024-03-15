package api

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
)

// ListTemplateKinds will list all template kinds available for PDS
func (tempDef *PDSV2_API) ListTemplateKinds(listTempKindReq *automationModels.WorkFlowRequest) ([]automationModels.WorkFlowResponse, error) {
	log.Infof("Value of Template - [%v]", listTempKindReq)
	return nil, nil
}

func (tempDef *PDSV2_API) ListTemplateRevisions(listTempRevReq *automationModels.WorkFlowRequest) ([]automationModels.WorkFlowResponse, error) {
	log.Infof("Value of Template - [%v]", listTempRevReq)
	return nil, nil
}

func (tempDef *PDSV2_API) GetTemplateRevisions(getTempRevReq *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	log.Infof("Value of Template - [%v]", getTempRevReq)
	return nil, nil
}
