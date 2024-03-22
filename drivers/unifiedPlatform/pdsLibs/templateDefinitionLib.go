package pdslibs

import (
	automationModels "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

type TemplateInputs struct {
	TenantId        string
	TemplateName    string
	Kind            string
	SemanticVersion string
	RevisionUid     string
	TemplateValues  structpb.Struct
}

func CreateServiceConfigTemplate(templateInputs TemplateInputs) (*automationModels.PlatformTemplatesResponse, error) {
	createReq := automationModels.PlatformTemplatesRequest{Create: automationModels.CreatePlatformTemplates{
		TenantId: templateInputs.TenantId,
		Template: &automationModels.Template{
			Meta: &automationModels.V1Meta{Name: &templateInputs.TemplateName},
			Config: &automationModels.V1Config{
				Kind:           &templateInputs.Kind,
				RevisionUid:    &templateInputs.RevisionUid,
				TemplateValues: &templateInputs.TemplateValues,
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

func CreateStorageConfigTemplate(templateInputs TemplateInputs) (*automationModels.PlatformTemplatesResponse, error) {
	createReq := automationModels.PlatformTemplatesRequest{Create: automationModels.CreatePlatformTemplates{
		TenantId: templateInputs.TenantId,
		Template: &automationModels.Template{
			Meta: &automationModels.V1Meta{Name: &templateInputs.TemplateName},
			Config: &automationModels.V1Config{
				Kind:            &templateInputs.Kind,
				SemanticVersion: &templateInputs.SemanticVersion,
				TemplateValues:  &templateInputs.TemplateValues,
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

func CreateResourceConfigTemplate(templateInputs TemplateInputs) (*automationModels.PlatformTemplatesResponse, error) {
	createReq := automationModels.PlatformTemplatesRequest{Create: automationModels.CreatePlatformTemplates{
		TenantId: templateInputs.TenantId,
		Template: &automationModels.Template{
			Meta: &automationModels.V1Meta{Name: &templateInputs.TemplateName},
			Config: &automationModels.V1Config{
				Kind:            &templateInputs.Kind,
				SemanticVersion: &templateInputs.SemanticVersion,
				TemplateValues:  &templateInputs.TemplateValues,
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

func GetRevisionUuidForApplication() (string, error) {
	revision, err := v2Components.PDS.GetTemplateRevisions()
	if err != nil {
		return "", err
	}
	revisionUuid := revision.GetRevision.Meta.Uid
	return *revisionUuid, nil

}

func GetSemanticVersion() (string, error) {
	semantic, err := v2Components.PDS.GetTemplateRevisions()
	if err != nil {
		return "", err
	}
	semanticVersion := semantic.GetRevision.Info.SemanticVersion
	return semanticVersion, nil
}
