package api

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
)

// ListTemplateKinds will list all template kinds available for PDS
func (tempDef *PDS_API_V1) ListTemplateKinds(listTempKindReq *automationModels.WorkFlowRequest) ([]automationModels.WorkFlowResponse, error) {
	log.Infof("Value of Template - [%v]", listTempKindReq)
	return nil, nil
}

func (tempDef *PDS_API_V1) ListTemplateRevisions(listTempRevReq *automationModels.WorkFlowRequest) ([]automationModels.WorkFlowResponse, error) {
	log.Infof("Value of Template - [%v]", listTempRevReq)
	return nil, nil
}

func (tempDef *PDS_API_V1) GetTemplateRevisions(getTempRevReq *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	log.Infof("Value of Template - [%v]", getTempRevReq)
	return nil, nil
}
