package grpc

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	commonapiv1 "github.com/pure-px/apis/public/portworx/common/apiv1"
	publictemplateapis "github.com/pure-px/apis/public/portworx/platform/template/apiv1"
	"google.golang.org/grpc"
)

// getTemplateClient updates the header with bearer token and returns the new client
func (templateGrpc *PlatformGrpc) getTemplateClient() (context.Context, publictemplateapis.TemplateServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var tenantClient publictemplateapis.TemplateServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	tenantClient = publictemplateapis.NewTemplateServiceClient(templateGrpc.ApiClientV1)

	return ctx, tenantClient, token, nil
}

// ListTemplates return service identities models for a project.
func (templateGrpc *PlatformGrpc) ListTemplates(templateReqReq *PlatformTemplates) ([]WorkFlowResponse, error) {
	ctx, templateClient, _, err := templateGrpc.getTemplateClient()
	templateResponse := []WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var firstPageRequest *publictemplateapis.ListTemplatesRequest
	firstPageRequest.TenantId = templateReqReq.List.V1ListTemplatesRequest.TenantId
	apiResponse, err := templateClient.ListTemplates(ctx, firstPageRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `ListTemplates` call: %v\n", err)
	}
	log.Infof("Value of Template - [%v]", apiResponse)
	err = utilities.CopyStruct(&templateResponse, apiResponse.Templates)
	log.Infof("Value of Template after copy - [%v]", templateResponse)
	return templateResponse, nil
}

// CreateTemplates returns newly create template  object
func (templateGrpc *PlatformGrpc) CreateTemplates(templateReq *PlatformTemplates) (*WorkFlowResponse, error) {
	ctx, templateClient, _, err := templateGrpc.getTemplateClient()
	templateResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	createTemplateReq := publictemplateapis.CreateTemplateRequest{
		TenantId: templateReq.Create.TenantId,
		Template: &publictemplateapis.Template{
			Meta: &commonapiv1.Meta{Name: *templateReq.Create.Template.Meta.Name},
			Config: &publictemplateapis.Config{
				Kind:            *templateReq.Create.Template.Config.Kind,
				SemanticVersion: *templateReq.Create.Template.Config.SemanticVersion,
				RevisionUid:     *templateReq.Create.Template.Config.RevisionUid,
				TemplateValues:  templateReq.Create.Template.Config.TemplateValues},
		},
	}
	apiResponse, err := templateClient.CreateTemplate(ctx, &createTemplateReq, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error while creating the project: %v\n", err)
	}
	err = utilities.CopyStruct(&templateResponse, apiResponse)
	return &templateResponse, nil
}

func (templateGrpc *PlatformGrpc) UpdateTemplates(templateReq *PlatformTemplates) (*WorkFlowResponse, error) {
	ctx, templateClient, _, err := templateGrpc.getTemplateClient()
	templateResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	updateTemplateReq := publictemplateapis.UpdateTemplateRequest{
		Id: templateReq.Update.Id,
		Template: &publictemplateapis.Template{
			Meta: &commonapiv1.Meta{Name: *templateReq.Create.Template.Meta.Name},
			Config: &publictemplateapis.Config{
				Kind:            *templateReq.Create.Template.Config.Kind,
				SemanticVersion: *templateReq.Create.Template.Config.SemanticVersion,
				RevisionUid:     *templateReq.Create.Template.Config.RevisionUid,
				TemplateValues:  templateReq.Create.Template.Config.TemplateValues},
		},
	}
	apiResponse, err := templateClient.UpdateTemplate(ctx, &updateTemplateReq, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error while creating the project: %v\n", err)
	}
	err = utilities.CopyStruct(&templateResponse, apiResponse)
	return &templateResponse, nil
}

// GetTemplates return template model.
func (templateGrpc *PlatformGrpc) GetTemplates(templateReq *PlatformTemplates) (*WorkFlowResponse, error) {
	ctx, templateClient, _, err := templateGrpc.getTemplateClient()
	templateResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	templateGetRequest := publictemplateapis.GetTemplateRequest{Id: templateReq.Get.Id}

	apiResponse, err := templateClient.GetTemplate(ctx, &templateGetRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `GetTemplate` call: %v\n", err)
	}
	log.Infof("Value of Template - [%v]", apiResponse)
	err = utilities.CopyStruct(&templateResponse, apiResponse)
	log.Infof("Value of Template after copy - [%v]", templateResponse)
	return &templateResponse, nil
}

// DeleteTemplate delete template and return status.
func (templateGrpc *PlatformGrpc) DeleteTemplate(templateReq *PlatformTemplates) error {
	ctx, templateClient, _, err := templateGrpc.getTemplateClient()
	templateResponse := WorkFlowResponse{}
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	templateDelRequest := publictemplateapis.DeleteTemplateRequest{Id: templateReq.Get.Id}

	apiResponse, err := templateClient.DeleteTemplate(ctx, &templateDelRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return fmt.Errorf("Error in calling `GetTemplate` call: %v\n", err)
	}
	log.Infof("Value of Template - [%v]", apiResponse)
	err = utilities.CopyStruct(&templateResponse, apiResponse)
	log.Infof("Value of Template after copy - [%v]", templateResponse)
	return nil
}
