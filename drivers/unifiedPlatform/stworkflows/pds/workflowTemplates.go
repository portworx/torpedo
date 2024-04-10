package pds

import (
	"fmt"
	"strconv"

	"github.com/portworx/torpedo/drivers/pds/parameters"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowPDSTemplates struct {
	Platform                platform.WorkflowPlatform
	ResourceTemplateId      string
	StorageTemplateId       string
	ServiceConfigTemplateId string
}

func (cusTemp *WorkflowPDSTemplates) CreatePdsCustomTemplatesAndFetchIds(templates *parameters.NewPDSParams, dsName string) (string, string, string, error) {

	//Todo: Mechanism to populate dynamic/Unknown key-value pairs for App config

	//Initializing the parameters required for template generation
	appConfigParams := pdslibs.ServiceConfiguration{
		MaxConnection: templates.ServiceConfiguration.MaxConnection,
	}
	stConfigParams := pdslibs.StorageConfiguration{
		FSType:      templates.StorageConfiguration.FSType,
		ReplFactor:  templates.StorageConfiguration.ReplFactor,
		Provisioner: templates.StorageConfiguration.Provisioner,
		FG:          templates.StorageConfiguration.FG,
		Secure:      templates.StorageConfiguration.Secure,
	}
	resConfigParams := pdslibs.ResourceConfiguration{
		CpuLimit:       templates.ResourceConfiguration.CpuLimit,
		CpuRequest:     templates.ResourceConfiguration.CpuRequest,
		MemoryLimit:    templates.ResourceConfiguration.MemoryLimit,
		MemoryRequest:  templates.ResourceConfiguration.MemoryRequest,
		StorageRequest: templates.ResourceConfiguration.StorageRequest,
		NewStorageSize: templates.ResourceConfiguration.NewStorageSize,
	}
	appConfig, _ := pdslibs.CreateServiceConfigTemplate(cusTemp.Platform.TenantId, dsName, appConfigParams)
	log.InfoD("appConfig ID-  %v", *appConfig.Create.Meta.Uid)
	appConfigId := appConfig.Create.Meta.Uid
	cusTemp.ServiceConfigTemplateId = *appConfigId

	stConfig, _ := pdslibs.CreateStorageConfigTemplate(cusTemp.Platform.TenantId, dsName, stConfigParams)
	log.InfoD("stConfig ID-  %v", *stConfig.Create.Meta.Uid)
	stConfigId := stConfig.Create.Meta.Uid
	cusTemp.StorageTemplateId = *stConfigId

	resConfig, _ := pdslibs.CreateResourceConfigTemplate(cusTemp.Platform.TenantId, dsName, resConfigParams)
	log.InfoD("resConfig ID-  %v", *resConfig.Create.Meta.Uid)
	resourceConfigId := resConfig.Create.Meta.Uid
	cusTemp.ResourceTemplateId = *resourceConfigId
	return *appConfigId, *stConfigId, *resourceConfigId, nil
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

func (cusTemp *WorkflowPDSTemplates) CreateResourceTemplateWithCustomValue(templates *parameters.NewPDSParams, dsName string, updateValue int) (string, error) {
	resConfigParams := pdslibs.ResourceConfiguration{
		CpuLimit:       templates.ResourceConfiguration.CpuLimit,
		CpuRequest:     templates.ResourceConfiguration.CpuRequest,
		MemoryLimit:    templates.ResourceConfiguration.MemoryLimit,
		MemoryRequest:  templates.ResourceConfiguration.MemoryRequest,
		StorageRequest: templates.ResourceConfiguration.StorageRequest,
		NewStorageSize: templates.ResourceConfiguration.NewStorageSize,
	}
	newCpuLimits, _ := strconv.Atoi(templates.ResourceConfiguration.CpuLimit)
	templates.ResourceConfiguration.CpuLimit = fmt.Sprint(string(rune(newCpuLimits + updateValue)))
	newCpuReq, _ := strconv.Atoi(templates.ResourceConfiguration.CpuLimit)
	templates.ResourceConfiguration.CpuRequest = fmt.Sprint(string(rune(newCpuReq + updateValue)))

	//create new templates with changed values of MEM Values -
	newMemLimits, _ := strconv.Atoi(templates.ResourceConfiguration.MemoryLimit)
	templates.ResourceConfiguration.MemoryLimit = fmt.Sprint(string(rune(newMemLimits + updateValue)))
	newMemReq, _ := strconv.Atoi(templates.ResourceConfiguration.MemoryLimit)
	templates.ResourceConfiguration.MemoryRequest = fmt.Sprint(string(rune(newMemReq + updateValue)))

	//create new templates with new storage Req-
	newStorageReq, _ := strconv.Atoi(templates.ResourceConfiguration.NewStorageSize)
	templates.ResourceConfiguration.StorageRequest = fmt.Sprint(string(rune(newStorageReq+1))) + "G"
	resConfig, _ := pdslibs.CreateResourceConfigTemplate(cusTemp.Platform.TenantId, dsName, resConfigParams)
	log.InfoD("resConfig ID-  %v", *resConfig.Create.Meta.Uid)
	resourceConfigId := resConfig.Create.Meta.Uid
	cusTemp.ResourceTemplateId = *resourceConfigId
	return *resourceConfigId, nil
}
