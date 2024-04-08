package pdslibs

import (
	automationModels "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	"reflect"
	"strings"
)

type StorageConfiguration struct {
	FSType     string
	ReplFactor int32
}
type ResourceConfiguration struct {
	CpuLimit       string
	CpuRequest     string
	MemoryLimit    string
	MemoryRequest  string
	StorageRequest string
	NewStorageSize string
}
type ServiceConfiguration struct {
	MaxConnection string
}

func CreateServiceConfigTemplate(tenantId string, dsName string, serviceConfig ServiceConfiguration) (*automationModels.PlatformTemplatesResponse, error) {
	log.InfoD("DSNAME fetched is- [%v]", dsName)
	revisionUid, err := GetRevisionUidForApplication(dsName)
	templateName := "pdsAutoSVCTemp" + utilities.RandomString(5)

	keyValue := ServiceConfiguration{
		MaxConnection: serviceConfig.MaxConnection,
	}
	templateValue := structToMap(keyValue)
	log.InfoD("Tenant is- [%v]", tenantId)
	createReq := automationModels.PlatformTemplatesRequest{Create: automationModels.CreatePlatformTemplates{
		TenantId: tenantId,
		Template: &automationModels.V1Template{
			Meta: &automationModels.V1Meta{Name: &templateName},
			Config: &automationModels.V1Config{
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

func CreateStorageConfigTemplate(tenantId string, dsName string, templateConfigs StorageConfiguration) (*automationModels.PlatformTemplatesResponse, error) {
	revisionUid, err := GetRevisionUidForApplication(dsName)
	templateName := "pdsAutoSVCTemp" + utilities.RandomString(5)

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
				TemplateValues: templateValue,
				RevisionUid:    &revisionUid,
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

func CreateResourceConfigTemplate(tenantId string, dsName string, templateConfigs ResourceConfiguration) (*automationModels.PlatformTemplatesResponse, error) {
	revisionUid, err := GetRevisionUidForApplication(dsName)
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
				RevisionUid:    &revisionUid,
				TemplateValues: templateValue,
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
	log.InfoD("Template to be deleted is- [%v]", id)
	templateResponse := v2Components.Platform.DeleteTemplate(&delReq)
	if err != nil {
		return templateResponse
	}
	return nil
}

func GetRevisionUidForApplication(dsName string) (string, error) {
	revisionList, err := v2Components.PDS.ListTemplateRevisions()
	var revisionUid string
	if err != nil {
		return "", err
	}
	for _, revision := range revisionList.ListRevision.Revisions {
		mainStringLower := strings.ToLower(*revision.Meta.Name)
		subStringLower := strings.ToLower(dsName)
		if strings.Contains(mainStringLower, subStringLower) {
			revisionUid = *revision.Meta.Uid
		}
	}
	log.InfoD("RevisionUid for the ds under test [%v] is- [%v]", dsName, revisionUid)
	return revisionUid, nil

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
	log.InfoD("templateValue formed is- [%v]", result)
	return result
}
