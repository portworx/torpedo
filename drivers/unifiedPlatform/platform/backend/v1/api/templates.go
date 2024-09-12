package api

import (
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
	listRequest = client.TemplateServiceListTemplates(ctx)
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
	listRequest = client.TemplateServiceListTemplates2(ctx)
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
	ctx, client, err := template.getTemplateClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}
	templateResponse := PlatformTemplatesResponse{}
	tempValueBody := templatesv1.V1Template{
		Meta: &templatesv1.V1Meta{Name: templateReq.Create.Template.Meta.Name},
		Config: &templatesv1.V1Config{
			RevisionUid:    templateReq.Create.Template.Config.RevisionUid,
			TemplateValues: templateReq.Create.Template.Config.TemplateValues,
		},
	}
	templateCreateRequest := client.TemplateServiceCreateTemplate(ctx, templateReq.Create.TenantId)
	templateCreateRequest = templateCreateRequest.V1Template(tempValueBody)
	templateModel, res, err := templateCreateRequest.Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TemplateServiceCreateTemplateExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(templateModel, &templateResponse.Create)
	return &templateResponse, err
}

func (template *PLATFORM_API_V1) UpdateTemplates(templateReq *PlatformTemplatesRequest) (*PlatformTemplatesResponse, error) {
	ctx, client, err := template.getTemplateClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for backend call: %v\n", err)
	}
	templateResponse := PlatformTemplatesResponse{}
	tempValueBody := templatesv1.TemplateServiceUpdateTemplateBody{
		Template: &templatesv1.DesiredTemplateConfiguration{
			Meta: &templatesv1.MetadataOfTheResource{Name: templateReq.Create.Template.Meta.Name},
			Config: &templatesv1.V1Config{
				Kind:            templateReq.Create.Template.Config.Kind,
				SemanticVersion: templateReq.Create.Template.Config.SemanticVersion,
				RevisionUid:     templateReq.Create.Template.Config.RevisionUid,
				TemplateValues:  templateReq.Create.Template.Config.TemplateValues,
			},
		},
	}
	templateUpdateRequest := client.TemplateServiceUpdateTemplate(ctx, templateReq.Update.Id)
	templateUpdateRequest = templateUpdateRequest.TemplateServiceUpdateTemplateBody(tempValueBody)
	templateModel, res, err := templateUpdateRequest.Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TemplateServiceUpdateTemplateExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(templateModel, &templateResponse.Create)
	return &templateResponse, err
}

// GetTemplates return template model.
func (template *PLATFORM_API_V1) GetTemplates(templateReq *PlatformTemplatesRequest) (*PlatformTemplatesResponse, error) {
	ctx, client, err := template.getTemplateClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	templateResponse := PlatformTemplatesResponse{
		Get: V1Template{},
	}
	templateModel, res, err := client.TemplateServiceGetTemplate(ctx, templateReq.Get.Id).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TemplateServiceGetTemplateExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Successfully fetched the template Roles")
	log.Infof("Value of template - [%v]", templateModel)
	err = utilities.CopyStruct(templateModel, &templateResponse.Get)
	log.Infof("Value of template after copy - [%v]", templateResponse.Get)
	return &templateResponse, nil
}

// DeleteTemplate delete template and return status.
func (template *PLATFORM_API_V1) DeleteTemplate(templateReq *PlatformTemplatesRequest) error {
	ctx, templateClient, err := template.getTemplateClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	templateResponse := PlatformTemplatesResponse{Delete: DeletePlatformTemplates{}}
	templateDelRequest := templateClient.TemplateServiceDeleteTemplate(ctx, templateReq.Delete.Id)
	log.InfoD("Template create req formed is- %v", templateDelRequest)
	templateModel, res, err := templateDelRequest.Execute()
	if err != nil || res.StatusCode != status.StatusOK {
		return fmt.Errorf("Error when calling `TemplateServiceCreateTemplateExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(templateModel, &templateResponse.Create)
	return nil
}
