package pdslibs

import (
	automationModels "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	"reflect"
)

type StorageConfiguration struct {
	FSType         []string
	ReplFactor     []int32
	StorageRequest string
	NewStorageSize string
}
type ResourceConfiguration struct {
	CpuLimit      string
	CpuRequest    string
	MemoryLimit   string
	MemoryRequest string
}
type ServiceConfiguration struct {
	HeapSize int
	Username string
	Password string
}

func CreateServiceConfigTemplate(tenantId string, templateConfigs ServiceConfiguration) (*automationModels.PlatformTemplatesResponse, error) {
	revisionUid, err := GetRevisionUidForApplication()
	templateKind, err := GetTemplateKind()
	templateName := "pdsAutoSVCTemp" + utilities.RandomString(5)
	templateValue := structToMap(templateConfigs)
	if err != nil {
		return nil, err
	}
	createReq := automationModels.PlatformTemplatesRequest{Create: automationModels.CreatePlatformTemplates{
		TenantId: tenantId,
		Template: &automationModels.Template{
			Meta: &automationModels.V1Meta{Name: &templateName},
			Config: &automationModels.V1Config{
				Kind:           &templateKind,
				RevisionUid:    &revisionUid,
				TemplateValues: templateValue,
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

func CreateStorageConfigTemplate(tenantId string, templateConfigs StorageConfiguration) (*automationModels.PlatformTemplatesResponse, error) {
	semanticVersion, err := GetSemanticVersion()
	templateKind, err := GetTemplateKind()
	templateName := "pdsAutoStTemp" + utilities.RandomString(5)
	templateValue := structToMap(templateConfigs)
	if err != nil {
		return nil, err
	}
	createReq := automationModels.PlatformTemplatesRequest{Create: automationModels.CreatePlatformTemplates{
		TenantId: tenantId,
		Template: &automationModels.Template{
			Meta: &automationModels.V1Meta{Name: &templateName},
			Config: &automationModels.V1Config{
				Kind:            &templateKind,
				SemanticVersion: &semanticVersion,
				TemplateValues:  templateValue,
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

func CreateResourceConfigTemplate(tenantId string, templateConfigs ResourceConfiguration) (*automationModels.PlatformTemplatesResponse, error) {
	semanticVersion, err := GetSemanticVersion()
	templateKind, err := GetTemplateKind()
	templateName := "pdsAutoResTemp" + utilities.RandomString(5)
	templateValue := structToMap(templateConfigs)
	if err != nil {
		return nil, err
	}
	createReq := automationModels.PlatformTemplatesRequest{Create: automationModels.CreatePlatformTemplates{
		TenantId: tenantId,
		Template: &automationModels.Template{
			Meta: &automationModels.V1Meta{Name: &templateName},
			Config: &automationModels.V1Config{
				Kind:            &templateKind,
				SemanticVersion: &semanticVersion,
				TemplateValues:  templateValue,
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

func GetRevisionUidForApplication() (string, error) {
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

func GetTemplateKind() (string, error) {
	semantic, err := v2Components.PDS.GetTemplateRevisions()
	if err != nil {
		return "", err
	}
	kind := semantic.GetRevision.Meta.Name
	return *kind, nil
}

func structToMap(structType interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	val := reflect.ValueOf(structType)
	typ := reflect.TypeOf(structType)

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		key := typ.Field(i).Name
		value := field.Interface()
		result[key] = value
	}
	println(result)
	return result
}
