package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
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
func (templateGrpc *PlatformGrpc) ListTemplates(listReq *WorkFlowRequest) ([]WorkFlowResponse, error) {
	ctx, templateClient, _, err := templateGrpc.getTemplateClient()
	templateResponse := []WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var firstPageRequest *publictemplateapis.ListTemplatesRequest
	err = copier.Copy(&firstPageRequest, listReq)
	if err != nil {
		return nil, err
	}
	templateModel, err := templateClient.ListTemplates(ctx, firstPageRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `ListTemplates` call: %v\n", err)
	}
	log.Infof("Value of Template - [%v]", templateModel)
	err = copier.Copy(&templateResponse, templateModel.Templates)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of Template after copy - [%v]", templateResponse)
	return templateResponse, nil
}

// CreateTemplates returns newly create template RoleBinding object
func (templateGrpc *PlatformGrpc) CreateTemplates(createReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, templateClient, _, err := templateGrpc.getTemplateClient()
	templateResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var templateCreateRequest *publictemplateapis.CreateTemplateRequest
	err = copier.Copy(&templateCreateRequest, createReq)
	if err != nil {
		return nil, err
	}
	templateModel, err := templateClient.CreateTemplate(ctx, templateCreateRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `CreateTemplate` call: %v\n", err)
	}
	log.Infof("Value of Template - [%v]", templateModel)
	err = copier.Copy(&templateResponse, templateModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of Template after copy - [%v]", templateResponse)
	return &templateResponse, nil
}

func (templateGrpc *PlatformGrpc) UpdateTemplates(updateReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, templateClient, _, err := templateGrpc.getTemplateClient()
	templateResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var templateUpdateRequest *publictemplateapis.UpdateTemplateRequest
	err = copier.Copy(&templateUpdateRequest, updateReq)
	if err != nil {
		return nil, err
	}
	templateModel, err := templateClient.UpdateTemplate(ctx, templateUpdateRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `UpdateTemplate` call: %v\n", err)
	}
	log.Infof("Value of Template - [%v]", templateModel)
	err = copier.Copy(&templateResponse, templateModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of Template after copy - [%v]", templateResponse)
	return &templateResponse, nil
}

// GetTemplateByID return template model.
func (templateGrpc *PlatformGrpc) GetTemplateByID(templateId *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, templateClient, _, err := templateGrpc.getTemplateClient()
	templateResponse := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var templateGetRequest *publictemplateapis.GetTemplateRequest
	err = copier.Copy(&templateGetRequest, templateId)
	if err != nil {
		return nil, err
	}
	templateModel, err := templateClient.GetTemplate(ctx, templateGetRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in calling `GetTemplate` call: %v\n", err)
	}
	log.Infof("Value of Template - [%v]", templateModel)
	err = copier.Copy(&templateResponse, templateModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of Template after copy - [%v]", templateResponse)
	return &templateResponse, nil
}

// DeleteTemplate delete template and return status.
func (templateGrpc *PlatformGrpc) DeleteTemplate(templateId *WorkFlowRequest) error {
	ctx, templateClient, _, err := templateGrpc.getTemplateClient()
	templateResponse := WorkFlowResponse{}
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var templateDelRequest *publictemplateapis.DeleteTemplateRequest
	err = copier.Copy(&templateDelRequest, templateId)
	if err != nil {
		return err
	}
	templateModel, err := templateClient.DeleteTemplate(ctx, templateDelRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return fmt.Errorf("Error in calling `DeleteTemplate` call: %v\n", err)
	}
	log.Infof("Value of Template - [%v]", templateModel)
	err = copier.Copy(&templateResponse, templateModel)
	if err != nil {
		return err
	}
	log.Infof("Value of Template after copy - [%v]", templateResponse)
	return nil
}
