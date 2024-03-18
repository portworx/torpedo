package grpc

import (
	"context"
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	commonapiv1 "github.com/pure-px/apis/public/portworx/common/apiv1"
	publicTempDefapis "github.com/pure-px/apis/public/portworx/pds/catalog/templatedefinition/apiv1"
	"google.golang.org/grpc"
)

func (tempDef *PdsGrpc) getTemplateDefClient() (context.Context, publicTempDefapis.TemplateDefinitionServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var depClient publicTempDefapis.TemplateDefinitionServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	credentials = &Credentials{
		Token: token,
	}
	depClient = publicTempDefapis.NewTemplateDefinitionServiceClient(tempDef.ApiClientV2)
	return ctx, depClient, token, nil
}

// ListTemplateKinds will list all template kinds available for PDS
func (tempDef *PdsGrpc) ListTemplateKinds(listTempKindReq *automationModels.WorkFlowRequest) ([]automationModels.WorkFlowResponse, error) {
	ctx, tempDefClient, _, err := tempDef.getTemplateDefClient()
	templateResponse := []automationModels.WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var firstPageRequest *publicTempDefapis.ListTemplateKindsRequest
	err = utilities.CopyStruct(&firstPageRequest, listTempKindReq)
	templateModel, err := tempDefClient.ListTemplateKinds(ctx, firstPageRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `ListTemplates` call: %v\n", err)
	}
	log.Infof("Value of Template - [%v]", templateModel)
	err = utilities.CopyStruct(&templateResponse, templateModel.Kinds)
	log.Infof("Value of Template after copy - [%v]", templateResponse)
	return templateResponse, nil
}

func (tempDef *PdsGrpc) ListTemplateRevisions(listTempRevReq *automationModels.WorkFlowRequest) ([]automationModels.WorkFlowResponse, error) {
	ctx, tempDefClient, _, err := tempDef.getTemplateDefClient()
	templateResponse := []automationModels.WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var firstPageRequest *commonapiv1.ListRevisionsRequest
	err = utilities.CopyStruct(&firstPageRequest, listTempRevReq)
	templateModel, err := tempDefClient.ListRevisions(ctx, firstPageRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `ListTemplates` call: %v\n", err)
	}
	log.Infof("Value of Template - [%v]", templateModel)
	err = utilities.CopyStruct(&templateResponse, templateModel.Revisions)
	log.Infof("Value of Template after copy - [%v]", templateResponse)
	return templateResponse, nil
}

func (tempDef *PdsGrpc) GetTemplateRevisions(getTempRevReq *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	ctx, tempDefClient, _, err := tempDef.getTemplateDefClient()
	templateResponse := automationModels.WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var getRequest *commonapiv1.GetRevisionRequest
	err = utilities.CopyStruct(&getRequest, getTempRevReq)
	templateModel, err := tempDefClient.GetRevision(ctx, getRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `ListTemplates` call: %v\n", err)
	}
	log.Infof("Value of Template - [%v]", templateModel)
	err = utilities.CopyStruct(&templateResponse, templateModel)
	log.Infof("Value of Template after copy - [%v]", templateResponse)
	return &templateResponse, nil
}
