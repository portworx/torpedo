package pds

import (
	"github.com/portworx/torpedo/drivers/pds/parameters"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	"github.com/portworx/torpedo/pkg/log"
	"strings"
)

type WorkflowPDSTemplates struct {
	Platform                platform.WorkflowPlatform
	ResourceTemplateId      string
	StorageTemplateId       string
	ServiceConfigTemplateId string
	AppTempIdAndDsName      map[string]string
	UpdateTemplateNameAndId map[string]string
}

func (cusTemp *WorkflowPDSTemplates) CreateAppTemplate(params *parameters.NewPDSParams) map[string]string {
	appTempIdAndDsName := make(map[string]string)
	for _, conf := range params.DataserviceConfigurationsToTest {
		log.Infof(conf.Name)
		for key, value := range conf.Configurations {
			log.Infof("key: %s", key)
			log.Infof("value: %s", value)
		}
		appConfig, _ := pdslibs.NewCreateServiceConfigTemplate(cusTemp.Platform.TenantId, conf.Name, conf.Configurations)
		log.InfoD("appConfig ID-  %v", *appConfig.Create.Meta.Uid)
		appConfigId := appConfig.Create.Meta.Uid
		cusTemp.ServiceConfigTemplateId = *appConfigId
		appTempIdAndDsName[conf.Name] = *appConfigId
	}
	return appTempIdAndDsName
}

func (cusTemp *WorkflowPDSTemplates) CreatePdsCustomTemplatesAndFetchIds(templates *parameters.NewPDSParams) (map[string]string, string, string, error) {
	cusTemp.UpdateTemplateNameAndId = make(map[string]string)
	stConfigParams := pdslibs.StorageConfiguration{
		FS:          templates.StorageConfiguration.FS,
		Repl:        templates.StorageConfiguration.Repl,
		Provisioner: templates.StorageConfiguration.Provisioner,
		FG:          templates.StorageConfiguration.FG,
		Secure:      templates.StorageConfiguration.Secure,
	}
	resConfigParams := pdslibs.ResourceConfiguration{
		Cpu_Limit:       templates.ResourceConfiguration.Cpu_Limit,
		Cpu_Request:     templates.ResourceConfiguration.Cpu_Request,
		Memory_Limit:    templates.ResourceConfiguration.Memory_Limit,
		Memory_Request:  templates.ResourceConfiguration.Memory_Request,
		Storage_Request: templates.ResourceConfiguration.Storage_Request,
	}

	appTemplateNameAndId := cusTemp.CreateAppTemplate(templates)
	cusTemp.AppTempIdAndDsName = appTemplateNameAndId

	stConfig, _ := pdslibs.CreateStorageConfigTemplate(cusTemp.Platform.TenantId, stConfigParams)
	log.InfoD("stConfig ID-  %v", *stConfig.Create.Meta.Uid)
	stConfigId := stConfig.Create.Meta.Uid
	cusTemp.StorageTemplateId = *stConfigId

	resConfig, _ := pdslibs.CreateResourceConfigTemplate(cusTemp.Platform.TenantId, resConfigParams)
	log.InfoD("resConfig ID-  %v", *resConfig.Create.Meta.Uid)
	resourceConfigId := resConfig.Create.Meta.Uid
	cusTemp.ResourceTemplateId = *resourceConfigId
	cusTemp.UpdateTemplateNameAndId[*resConfig.Create.Meta.Name] = *resourceConfigId
	return appTemplateNameAndId, *stConfigId, *resourceConfigId, nil
}

func (cusTemp *WorkflowPDSTemplates) DeleteCreatedCustomPdsTemplates(tempList []string) error {
	for _, id := range tempList {
		err := pdslibs.DeleteTemplate(id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cusTemp *WorkflowPDSTemplates) CreateResourceTemplateWithCustomValue(templates *parameters.NewPDSParams) (string, error) {
	cusTemp.UpdateTemplateNameAndId = make(map[string]string)

	resConfigParams := pdslibs.ResourceConfiguration{
		Cpu_Limit:       templates.ResourceConfiguration.New_Cpu_Limit,
		Cpu_Request:     templates.ResourceConfiguration.New_Cpu_Request,
		Memory_Limit:    templates.ResourceConfiguration.New_Memory_Limit,
		Memory_Request:  templates.ResourceConfiguration.New_Memory_Request,
		Storage_Request: templates.ResourceConfiguration.New_Storage_Request,
	}

	resConfig, _ := pdslibs.CreateResourceConfigTemplate(cusTemp.Platform.TenantId, resConfigParams)
	log.InfoD("resConfig ID-  %v", *resConfig.Create.Meta.Uid)
	resourceConfigId := resConfig.Create.Meta.Uid
	cusTemp.ResourceTemplateId = *resourceConfigId
	cusTemp.UpdateTemplateNameAndId[*resConfig.Create.Meta.Name] = *resourceConfigId
	return *resourceConfigId, nil
}

func (cusTemp *WorkflowPDSTemplates) Purge(ignoreError bool) error {

	if cusTemp.ResourceTemplateId != "" {
		log.Infof("Deleting ResourceTemplate - [%s]", cusTemp.ResourceTemplateId)
		err := cusTemp.DeleteCreatedCustomPdsTemplates([]string{cusTemp.ResourceTemplateId})
		if err != nil {
			return err
		}
	}

	if cusTemp.StorageTemplateId != "" {
		log.Infof("Deleting StorageTemplate - [%s]", cusTemp.StorageTemplateId)
		err := cusTemp.DeleteCreatedCustomPdsTemplates([]string{cusTemp.StorageTemplateId})
		if err != nil {
			return err
		}
	}

	for _, template := range cusTemp.AppTempIdAndDsName {
		log.Infof("Deleting ServiceConfigTemplate - [%s]", template)
		err := cusTemp.DeleteCreatedCustomPdsTemplates([]string{template})
		if err != nil {
			return err
		}
	}

	log.Debugf("length of UpdateTemplateNameAndId [%d]", len(cusTemp.UpdateTemplateNameAndId))
	for _, template := range cusTemp.UpdateTemplateNameAndId {
		log.Infof("Deleting ResourceConfigTemplate - [%s]", template)
		err = cusTemp.DeleteCreatedCustomPdsTemplates([]string{template})
		if err != nil {
			if ignoreError && !strings.Contains(err.Error(), "404 Not Found") {
				return err
			}
		}
	}

	return nil
}
