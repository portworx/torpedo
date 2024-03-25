package api

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	templatesv1 "github.com/pure-px/platform-api-go-client/platform/v1/template"
	status "net/http"
)

// ListTemplatesForTenants return service identities models for a template.
func (template *PLATFORM_API_V1) ListTemplatesForTenants(templateReq *PlatformTemplatesRequest) (*PlatformTemplatesResponse, error) {
	templateResponse := PlatformTemplatesResponse{
		ListForTenant: V1ListTemplateResopnse{},
	}
	ctx, client, err := template.getTemplateClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var listRequest templatesv1.ApiTemplateServiceListTemplatesRequest
	listRequest = listRequest.ApiService.TemplateServiceListTemplates(ctx)
	listRequest = listRequest.TenantId(templateReq.ListForTenant.TenantId)
	templatesList, res, err := client.TemplateServiceListTemplatesExecute(listRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TemplateServiceListTemplatesExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(&templateResponse, templatesList.Templates)
	if err != nil {
		return nil, err
	}
	return &templateResponse, nil
}

// ListTemplates return service identities models for a template.
func (template *PLATFORM_API_V1) ListTemplates(templateReq *PlatformTemplatesRequest) (*PlatformTemplatesResponse, error) {
	ctx, client, err := template.getTemplateClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	templateResponse := PlatformTemplatesResponse{
		List: V1ListTemplateResopnse{},
	}
	var listRequest templatesv1.ApiTemplateServiceListTemplates2Request
	listRequest = listRequest.ApiService.TemplateServiceListTemplates2(ctx)
	templatesList, res, err := client.TemplateServiceListTemplates2Execute(listRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TemplateServiceListTemplates2Execute`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(&templateResponse, templatesList.Templates)
	if err != nil {
		return nil, err
	}
	return &templateResponse, nil
}

// CreateTemplates returns newly create template  object
func (template *PLATFORM_API_V1) CreateTemplates(templateReq *PlatformTemplatesRequest) (*PlatformTemplatesResponse, error) {
	_, client, err := template.getTemplateClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}
	templateResponse := PlatformTemplatesResponse{}
	templateCreateRequest := templatesv1.ApiTemplateServiceCreateTemplateRequest{}
	templateCreateRequest = templateCreateRequest.ApiService.TemplateServiceCreateTemplate(context.Background(), templateReq.Create.TenantId)
	var tempCreate templatesv1.V1Template
	templateCreateRequest = templateCreateRequest.V1Template(tempCreate)
	templateModel, res, err := client.TemplateServiceCreateTemplateExecute(templateCreateRequest)
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TemplateServiceCreateTemplateExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(&templateResponse, templateModel)
	return &templateResponse, err

}

func (template *PLATFORM_API_V1) UpdateTemplates(templateReq *PlatformTemplatesRequest) (*PlatformTemplatesResponse, error) {
	_, client, err := template.getTemplateClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}
	templateResponse := PlatformTemplatesResponse{}
	templateUpdateRequest := templatesv1.ApiTemplateServiceUpdateTemplateRequest{}
	templateUpdateRequest = templateUpdateRequest.ApiService.TemplateServiceUpdateTemplate(context.Background(), templateReq.Update.Id)
	var updateRequest templatesv1.V1Template
	templateUpdateRequest = templateUpdateRequest.V1Template(updateRequest)
	templateModel, res, err := client.TemplateServiceUpdateTemplateExecute(templateUpdateRequest)
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TemplateServiceCreateTemplateExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(&templateResponse, templateModel)
	return &templateResponse, err
}

// GetTemplates return template model.
func (template *PLATFORM_API_V1) GetTemplates(templateReq *PlatformTemplatesRequest) (*PlatformTemplatesResponse, error) {

	ctx, client, err := template.getTemplateClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	templateResponse := PlatformTemplatesResponse{}
	templateModel, res, err := client.TemplateServiceGetTemplate(ctx, templateReq.Get.Id).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TemplateServiceGetTemplateExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	log.InfoD("Successfully fetched the template Roles")
	log.Infof("Value of template - [%v]", templateModel)
	err = utilities.CopyStruct(&templateResponse, templateModel)
	log.Infof("Value of template after copy - [%v]", templateResponse)
	return &templateResponse, nil
}

// DeleteTemplate delete template and return status.
func (template *PLATFORM_API_V1) DeleteTemplate(templateReq *PlatformTemplatesRequest) error {
	ctx, templateClient, err := template.getTemplateClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	templateResponse := PlatformTemplatesResponse{}
	templateModel, res, err := templateClient.TemplateServiceDeleteTemplate(ctx, templateReq.Delete.Id).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return fmt.Errorf("Error when calling `templateServiceDeletetemplateExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	log.InfoD("Successfully DELETED the template Roles")
	log.Infof("Value of template - [%v]", templateModel)
	err = utilities.CopyStruct(&templateResponse, templateModel)
	log.Infof("Value of template after copy - [%v]", templateResponse)
	return nil
}
