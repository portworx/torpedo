package stworkflows

import (
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"google.golang.org/protobuf/types/known/structpb"
)

type CustomTemplates struct {
	TemplateName map[string]string
}

type InputsTemplates struct {
	TenantId        string
	TemplateName    string
	Kind            string
	SemanticVersion string
	//RevisionUid     string
	TemplateValues structpb.Struct
}

func (cusTemp *CustomTemplates) CreateApplicationTemplateAndGetID(inputs InputsTemplates) (string, error) {
	revisionUuid, err := pdslibs.GetRevisionUuidForApplication()
	if err != nil {
		return "", err
	}
	appConfig := pdslibs.TemplateInputs{
		TemplateName:   inputs.TemplateName,
		Kind:           inputs.Kind,
		RevisionUid:    revisionUuid,
		TemplateValues: inputs.TemplateValues,
	}
	templateResponse, err := pdslibs.CreateServiceConfigTemplate(appConfig)
	appConfigId := templateResponse.Create.Meta.Uid
	appConfigName := templateResponse.Create.Meta.Name
	cusTemp.TemplateName[*appConfigName] = *appConfigId
	return *appConfigId, nil
}

func (cusTemp *CustomTemplates) CreateStorageTemplateAndGetID(inputs InputsTemplates) (string, error) {
	semanticVersion, err := pdslibs.GetSemanticVersion()
	if err != nil {
		return "", err
	}
	stConfig := pdslibs.TemplateInputs{
		TemplateName:    inputs.TemplateName,
		Kind:            inputs.Kind,
		SemanticVersion: semanticVersion,
		TemplateValues:  inputs.TemplateValues,
	}
	templateResponse, err := pdslibs.CreateStorageConfigTemplate(stConfig)
	stConfigId := templateResponse.Create.Meta.Uid
	stConfigName := templateResponse.Create.Meta.Name
	cusTemp.TemplateName[*stConfigName] = *stConfigId
	return *stConfigId, nil
}

func (cusTemp *CustomTemplates) CreateResourceTemplateAndGetID(inputs InputsTemplates) (string, error) {
	semanticVersion, err := pdslibs.GetSemanticVersion()
	if err != nil {
		return "", err
	}
	stConfig := pdslibs.TemplateInputs{
		TemplateName:    inputs.TemplateName,
		Kind:            inputs.Kind,
		SemanticVersion: semanticVersion,
		TemplateValues:  inputs.TemplateValues,
	}
	templateResponse, err := pdslibs.CreateStorageConfigTemplate(stConfig)
	stConfigId := templateResponse.Create.Meta.Uid
	stConfigName := templateResponse.Create.Meta.Name
	cusTemp.TemplateName[*stConfigName] = *stConfigId
	return *stConfigId, nil
}
