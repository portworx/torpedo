package grpc

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	commonapiv1 "github.com/pure-px/apis/public/portworx/common/apiv1"
	publictempdefapis "github.com/pure-px/apis/public/portworx/pds/catalog/templatedefinition/apiv1"
	"google.golang.org/grpc"
)

func (tempDefGrpcGrpc *PdsGrpc) getTemplateDefClient() (context.Context, publictempdefapis.TemplateDefinitionServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var depClient publictempdefapis.TemplateDefinitionServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	credentials = &Credentials{
		Token: token,
	}
	depClient = publictempdefapis.NewTemplateDefinitionServiceClient(tempDefGrpcGrpc.ApiClientV2)
	return ctx, depClient, token, nil
}

// ListTemplateKinds will list all template kinds available for PDS
func (tempDefGrpc *PdsGrpc) ListTemplateKinds() (*TemplateDefinitionResponse, error) {
	ctx, tempDefGrpcClient, _, err := tempDefGrpc.getTemplateDefClient()
	templateResponse := TemplateDefinitionResponse{
		ListKinds: ListTemplateKindsResponse{},
	}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var firstPageRequest *publictempdefapis.ListTemplateKindsRequest
	apiResponse, err := tempDefGrpcClient.ListTemplateKinds(ctx, firstPageRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `ListTemplates` call: %v\n", err)
	}
	log.Infof("Value of Template - [%v]", apiResponse)
	err = utilities.CopyStruct(&templateResponse, apiResponse.Kinds)
	log.Infof("Value of Template after copy - [%v]", templateResponse)
	return &templateResponse, nil
}

func (tempDefGrpc *PdsGrpc) ListTemplateRevisions() (*TemplateDefinitionResponse, error) {
	ctx, tempDefGrpcClient, _, err := tempDefGrpc.getTemplateDefClient()
	templateResponse := TemplateDefinitionResponse{
		ListRevision: ListRevisionResponse{},
	}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var firstPageRequest *commonapiv1.ListRevisionsRequest
	apiResponse, err := tempDefGrpcClient.ListRevisions(ctx, firstPageRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `ListTemplates` call: %v\n", err)
	}
	log.Infof("Value of Template - [%v]", apiResponse)
	err = utilities.CopyStruct(&templateResponse, apiResponse.Revisions)
	log.Infof("Value of Template after copy - [%v]", templateResponse)
	return &templateResponse, nil
}

func (tempDefGrpcGrpc *PdsGrpc) GetTemplateRevisions() (*TemplateDefinitionResponse, error) {
	ctx, tempDefGrpcClient, _, err := tempDefGrpcGrpc.getTemplateDefClient()
	templateResponse := TemplateDefinitionResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var getReq *commonapiv1.GetRevisionRequest
	apiResponse, err := tempDefGrpcClient.GetRevision(ctx, getReq, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `ListTemplates` call: %v\n", err)
	}
	log.Infof("Value of Template - [%v]", apiResponse)
	err = utilities.CopyStruct(&templateResponse, apiResponse)
	log.Infof("Value of Template after copy - [%v]", templateResponse)
	return &templateResponse, nil
}
