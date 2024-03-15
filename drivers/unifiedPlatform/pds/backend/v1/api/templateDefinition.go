package api

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
)

// ListTemplateKinds will list all template kinds available for PDS
func (tempDef *PDS_API_V2) ListTemplateKinds(listTempKindReq *apiStructs.WorkFlowRequest) ([]apiStructs.WorkFlowResponse, error) {
	log.Infof("Value of Template - [%v]", listTempKindReq)
	return nil, nil
}

func (tempDef *PDS_API_V2) ListTemplateRevisions(listTempRevReq *apiStructs.WorkFlowRequest) ([]apiStructs.WorkFlowResponse, error) {
	log.Infof("Value of Template - [%v]", listTempRevReq)
	return nil, nil
}

func (tempDef *PDS_API_V2) GetTemplateRevisions(getTempRevReq *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Infof("Value of Template - [%v]", getTempRevReq)
	return nil, nil
}
