package pdslibs

import (
	automationModels "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	"reflect"
)

type StorageConfiguration struct {
	FSType         string
	ReplFactor     int32
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

func CreateServiceConfigTemplate(tenantId string, serviceConfig ServiceConfiguration) (*automationModels.PlatformTemplatesResponse, error) {
	//revisionUid, err := GetRevisionUidForApplication() PDS APIs not Available
	//templateKind, err := GetTemplateKind()

	//Dummy values for testing
	revisionUid := utilities.RandomString(12)
	templateKind := utilities.RandomString(12)
	templateName := "pdsAutoSVCTemp" + utilities.RandomString(5)

	keyValue := ServiceConfiguration{
		HeapSize: serviceConfig.HeapSize,
		Username: serviceConfig.Username,
		Password: serviceConfig.Password,
	}
	templateValue := structToMap(keyValue)
	if err != nil {
		return nil, err
	}
	createReq := automationModels.PlatformTemplatesRequest{Create: automationModels.CreatePlatformTemplates{
		TenantId: tenantId,
		Template: &automationModels.V1Template{
			Meta: &automationModels.V1Meta{Name: &templateName},
			Config: &automationModels.V1Config{
				Kind:           &templateKind,
				RevisionUid:    &revisionUid,
				TemplateValues: templateValue,
			},
		},
	}}
	log.InfoD("Create ServiceConfiguration Request formed is- {%v}", createReq)
	templateResponse, err := v2Components.Platform.CreateTemplates(&createReq)
	if err != nil {
		return templateResponse, err
	}
	return templateResponse, nil
}

func CreateStorageConfigTemplate(tenantId string, templateConfigs StorageConfiguration) (*automationModels.PlatformTemplatesResponse, error) {
	//PDS APIs not Available
	//templateKind, err := GetTemplateKind()
	//semanticVersion, err := GetSemanticVersion()

	//Dummy values for testing
	semanticVersion := utilities.RandomString(12)
	templateKind := utilities.RandomString(12)

	templateName := "pdsAutoStTemp" + utilities.RandomString(5)
	templateValue := structToMap(templateConfigs)
	log.InfoD("Temp value fromed in lib folder is- [%v]", templateValue)
	if err != nil {
		return nil, err
	}
	createReq := automationModels.PlatformTemplatesRequest{Create: automationModels.CreatePlatformTemplates{
		TenantId: tenantId,
		Template: &automationModels.V1Template{
			Meta: &automationModels.V1Meta{Name: &templateName},
			Config: &automationModels.V1Config{
				Kind:            &templateKind,
				SemanticVersion: &semanticVersion,
				TemplateValues:  templateValue,
			},
		},
	}}
	log.InfoD("Create StorageConfiguration Request formed is- {%v}", createReq)
	templateResponse, err := v2Components.Platform.CreateTemplates(&createReq)
	if err != nil {
		return templateResponse, err
	}
	return templateResponse, nil
}

func CreateResourceConfigTemplate(tenantId string, templateConfigs ResourceConfiguration) (*automationModels.PlatformTemplatesResponse, error) {
	//PDS APIs not Available
	//templateKind, err := GetTemplateKind()
	//semanticVersion, err := GetSemanticVersion()

	//Dummy values for testing
	semanticVersion := utilities.RandomString(12)
	templateKind := utilities.RandomString(12)
	templateName := "pdsAutoResTemp" + utilities.RandomString(5)
	templateValue := structToMap(templateConfigs)
	if err != nil {
		return nil, err
	}
	createReq := automationModels.PlatformTemplatesRequest{Create: automationModels.CreatePlatformTemplates{
		TenantId: tenantId,
		Template: &automationModels.V1Template{
			Meta: &automationModels.V1Meta{Name: &templateName},
			Config: &automationModels.V1Config{
				Kind:            &templateKind,
				SemanticVersion: &semanticVersion,
				TemplateValues:  templateValue,
			},
			Status: nil,
		},
	}}
	log.InfoD("Create ResourceConfiguration Request formed is- {%v}", createReq)
	templateResponse, err := v2Components.Platform.CreateTemplates(&createReq)
	if err != nil {
		return templateResponse, err
	}
	return templateResponse, nil
}

func DeleteTemplate(id string) error {
	delReq := automationModels.PlatformTemplatesRequest{Delete: automationModels.DeletePlatformTemplates{Id: id}}
	templateResponse := v2Components.Platform.DeleteTemplate(&delReq)
	if err != nil {
		return templateResponse
	}
	return nil
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
		log.InfoD("templateValue key is- [KEY- %v]", key)
		log.InfoD("templateValue value is- [KEY- %v]", value)
		result[key] = value
		log.InfoD("templateValue key-value pair formed is- [%v]", result)
	}
	log.InfoD("templateValue formed is- [%v]", result)
	return result
}
