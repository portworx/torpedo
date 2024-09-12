package platformLibs

import (
	automationModels "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
)

type TemplateInputs struct {
	TenantId        string
	TemplateName    string
	Kind            string
	SemanticVersion string
	RevisionUid     string
	TemplateValues  map[string]interface{}
}

func ListAvailableTemplates(tenantId string) (*automationModels.PlatformTemplatesResponse, error) {
	request := automationModels.PlatformTemplatesRequest{List: automationModels.ListTemplates{V1ListTemplatesRequest: automationModels.V1ListTemplatesRequest{TenantId: tenantId}}}
	listResponse, err := v2Components.Platform.ListTemplates(&request)
	if err != nil {
		return listResponse, err
	}

	return listResponse, nil
}

func CreateTemplates(templateInputs TemplateInputs) (*automationModels.PlatformTemplatesResponse, error) {
	createReq := automationModels.PlatformTemplatesRequest{Create: automationModels.CreatePlatformTemplates{
		TenantId: templateInputs.TenantId,
		Template: &automationModels.V1Template{
			Meta: &automationModels.V1Meta{Name: &templateInputs.TemplateName},
			Config: &automationModels.V1Config{
				Kind:            &templateInputs.Kind,
				SemanticVersion: &templateInputs.SemanticVersion,
				RevisionUid:     &templateInputs.RevisionUid,
				TemplateValues:  templateInputs.TemplateValues,
			},
			Status: nil,
		},
	}}
	templateResponse, err := v2Components.Platform.CreateTemplates(&createReq)
	if err != nil {
		return templateResponse, err
	}
	return templateResponse, nil
}

func GetTemplate(templateId string) (*automationModels.PlatformTemplatesResponse, error) {
	getTemplateReq := automationModels.PlatformTemplatesRequest{
		Get: automationModels.GetPlatformTemplates{Id: templateId},
	}
	templateResp, err := v2Components.Platform.GetTemplates(&getTemplateReq)
	if err != nil {
		return templateResp, err
	}
	return templateResp, nil
}

func ListTemplates(tenantId string) (*automationModels.PlatformTemplatesResponse, error) {
	listTempReq := automationModels.PlatformTemplatesRequest{List: automationModels.ListTemplates{V1ListTemplatesRequest: automationModels.V1ListTemplatesRequest{TenantId: tenantId}}}
	templates, err := v2Components.Platform.ListTemplates(&listTempReq)
	if err != nil {
		return templates, err
	}
	return templates, nil

}

func UpdateTemplate(templateId string, templateInputs *TemplateInputs) (*automationModels.PlatformTemplatesResponse, error) {
	updateReq := automationModels.PlatformTemplatesRequest{Update: automationModels.UpdatePlatformTemplates{
		Id: templateId,
		Template: &automationModels.V1Template{
			Meta: &automationModels.V1Meta{Name: &templateInputs.TemplateName},
			Config: &automationModels.V1Config{
				Kind:            &templateInputs.Kind,
				SemanticVersion: &templateInputs.SemanticVersion,
				RevisionUid:     &templateInputs.RevisionUid,
				TemplateValues:  templateInputs.TemplateValues,
			},
			Status: nil,
		},
	}}
	templateResponse, err := v2Components.Platform.UpdateTemplates(&updateReq)
	if err != nil {
		return templateResponse, err
	}
	return templateResponse, nil
}
