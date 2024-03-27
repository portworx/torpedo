package stworkflows

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/pds/parameters"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/pkg/log"
	"strconv"
)

type CustomTemplates struct {
	Platform                WorkflowPlatform
	ResourceTemplateId      string
	StorageTemplatetId      string
	ServiceConfigTemplateId string
}

func (cusTemp *CustomTemplates) CreatePdsCustomTemplatesAndFetchIds(templates *parameters.NewPDSParams, updateTemplate bool) (string, string, string, error) {

	//Todo: Mechanism to populate dynamic/Unknown key-value pairs for App config
	//ToDo: Take configurationValue incrementation count from user/testcase

	//Initializing the parameters required for template generation
	appConfigParams := pdslibs.ServiceConfiguration{
		HeapSize: templates.ServiceConfiguration.HeapSize,
		Username: templates.ServiceConfiguration.Username,
		Password: templates.ServiceConfiguration.Password,
	}
	stConfigParams := pdslibs.StorageConfiguration{
		FSType:         templates.StorageConfiguration.FSType,
		ReplFactor:     templates.StorageConfiguration.ReplFactor,
		StorageRequest: templates.StorageConfiguration.StorageRequest,
		NewStorageSize: templates.StorageConfiguration.NewStorageSize,
	}
	resConfigParams := pdslibs.ResourceConfiguration{
		CpuLimit:      templates.ResourceConfiguration.CpuLimit,
		CpuRequest:    templates.ResourceConfiguration.CpuRequest,
		MemoryLimit:   templates.ResourceConfiguration.MemoryLimit,
		MemoryRequest: templates.ResourceConfiguration.MemoryRequest,
	}
	if updateTemplate {
		//create new templates with changed values of CPU Values -
		newCpuLimits, _ := strconv.Atoi(templates.ResourceConfiguration.CpuLimit)
		templates.ResourceConfiguration.CpuLimit = fmt.Sprint(string(rune(newCpuLimits + 1)))
		newCpuReq, _ := strconv.Atoi(templates.ResourceConfiguration.CpuLimit)
		templates.ResourceConfiguration.CpuRequest = fmt.Sprint(string(rune(newCpuReq + 1)))

		//create new templates with changed values of MEM Values -
		newMemLimits, _ := strconv.Atoi(templates.ResourceConfiguration.MemoryLimit)
		templates.ResourceConfiguration.MemoryLimit = fmt.Sprint(string(rune(newMemLimits + 1)))
		newMemReq, _ := strconv.Atoi(templates.ResourceConfiguration.MemoryLimit)
		templates.ResourceConfiguration.MemoryRequest = fmt.Sprint(string(rune(newMemReq + 1)))

		//create new templates with new storage Req-
		newStorageReq, _ := strconv.Atoi(templates.StorageConfiguration.NewStorageSize)
		templates.StorageConfiguration.StorageRequest = fmt.Sprint(string(rune(newStorageReq+1))) + "G"
	}
	appConfig, _ := pdslibs.CreateServiceConfigTemplate(cusTemp.Platform.TenantId, appConfigParams)
	log.InfoD("appConfig ID-  %v", *appConfig.Create.Meta.Uid)
	appConfigId := appConfig.Create.Meta.Uid
	cusTemp.ServiceConfigTemplateId = *appConfigId

	stConfig, _ := pdslibs.CreateStorageConfigTemplate(cusTemp.Platform.TenantId, stConfigParams)
	log.InfoD("stConfig ID-  %v", *stConfig.Create.Meta.Uid)
	stConfigId := stConfig.Create.Meta.Uid
	cusTemp.StorageTemplatetId = *stConfigId

	resConfig, _ := pdslibs.CreateResourceConfigTemplate(cusTemp.Platform.TenantId, resConfigParams)
	log.InfoD("resConfig ID-  %v", *resConfig.Create.Meta.Uid)
	resourceConfigId := resConfig.Create.Meta.Uid
	cusTemp.ResourceTemplateId = *stConfigId
	return *appConfigId, *stConfigId, *resourceConfigId, nil
}

func (cusTemp *CustomTemplates) DeleteCreatedCustomPdsTemplates(tempList []string) error {
	for _, id := range tempList {
		err := pdslibs.DeleteTemplate(id)
		if err != nil {
			return err
		}
	}
	return nil
}
