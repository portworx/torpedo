package pdslibs

import (
	"fmt"
	automationModels "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	"reflect"
	"strings"
)

type StorageConfiguration struct {
	FS          string
	Repl        int32
	Provisioner string
	FG          bool
	Secure      bool
}
type ResourceConfiguration struct {
	Cpu_Limit       string
	Cpu_Request     string
	Memory_Limit    string
	Memory_Request  string
	Storage_Request string
}
type ServiceConfiguration struct {
	MAX_CONNECTIONS string
}

const (
	STORAGE_OPTIONS   = "storage_options"
	RESOURCE_SETTINGS = "resource_settings"
	SERVICE_OPTIONS   = "service_settings"
)

func CreateServiceConfigTemplate(tenantId string, dsName string, serviceConfig ServiceConfiguration) (*automationModels.PlatformTemplatesResponse, error) {
	log.InfoD("DsName fetched is- [%v]", dsName)
	revisionUid, err := GetRevisionUidForApplication(dsName)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch revisionUid for the dataservice - [%v] under test", dsName)
	}
	templateName := "pdsAutoSVCTemp" + utilities.RandomString(5)

	keyValue := ServiceConfiguration{
		MAX_CONNECTIONS: serviceConfig.MAX_CONNECTIONS,
	}
	templateValue := structToMap(keyValue, SERVICE_OPTIONS)
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
	revisionUid, err := GetRevisionUidForStorageOptions()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch revisionUid for the dataservice - [%v] under test", dsName)
	}
	templateName := "pdsAutoSVCTemp" + utilities.RandomString(5)
	templateValue := structToMap(templateConfigs, STORAGE_OPTIONS)
	log.InfoD("Temp value formed is- [%v]", templateValue)
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
	revisionUid, err := GetRevisionUidForResourceConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch revisionUid for the dataservice - [%v] under test", dsName)
	}
	templateName := "pdsAutoResTemp" + utilities.RandomString(5)
	templateValue := structToMap(templateConfigs, RESOURCE_SETTINGS)
	log.InfoD("Temp value formed is- [%v]", templateValue)
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
	err = v2Components.Platform.DeleteTemplate(&delReq)
	if err != nil {
		return fmt.Errorf("unable to delete templates due to error - [%v]", err)
	}
	return nil
}

func GetRevisionUidForApplication(dsName string) (string, error) {
	var revisionUid string
	revisionList, err := v2Components.PDS.ListTemplateRevisions()
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

func GetRevisionUidForStorageOptions() (string, error) {
	var revisionUid string
	revisionList, err := v2Components.PDS.ListTemplateRevisions()
	if err != nil {
		return "", err
	}
	for _, revision := range revisionList.ListRevision.Revisions {
		mainStringLower := strings.ToLower(*revision.Meta.Name)
		subStringLower := strings.ToLower(STORAGE_OPTIONS)
		if strings.Contains(mainStringLower, subStringLower) {
			revisionUid = *revision.Meta.Uid
		}
	}
	log.InfoD("RevisionUid for the ds under test [%v] is- [%v]", STORAGE_OPTIONS, revisionUid)
	return revisionUid, nil
}

func GetRevisionUidForResourceConfig() (string, error) {
	var revisionUid string
	revisionList, err := v2Components.PDS.ListTemplateRevisions()
	if err != nil {
		return "", err
	}
	for _, revision := range revisionList.ListRevision.Revisions {
		mainStringLower := strings.ToLower(*revision.Meta.Name)
		subStringLower := strings.ToLower(RESOURCE_SETTINGS)
		if strings.Contains(mainStringLower, subStringLower) {
			revisionUid = *revision.Meta.Uid
		}
	}
	log.InfoD("RevisionUid for the ds under test [%v] is- [%v]", RESOURCE_SETTINGS, revisionUid)
	return revisionUid, nil
}

func structToMap(structType interface{}, tempType string) map[string]interface{} {
	result := make(map[string]interface{})
	val := reflect.ValueOf(structType)
	typ := reflect.TypeOf(structType)
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		key := typ.Field(i).Name
		value := field.Interface()
		result[key] = value
	}
	if tempType == SERVICE_OPTIONS {
		log.InfoD("templateValue formed is- [%v]", result)
		return result
	}
	// Convert keys to lowercase
	lowercaseMap := make(map[string]interface{})
	for key, value := range result {
		lowercaseKey := strings.ToLower(key)
		lowercaseMap[lowercaseKey] = value
	}
	log.InfoD("templateValue formed is- [%v]", lowercaseMap)
	return lowercaseMap
}
