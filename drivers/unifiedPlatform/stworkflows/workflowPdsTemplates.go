package stworkflows

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/pds/parameters"
	pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"strconv"
)

const (
	resourceTempID      = "RESOURCE_CONFIGURATION_ID"
	storageTempID       = "STORAGE_CONFIGURATION_ID"
	serviceConfigTempID = "SERVICE_CONFIGURATION_ID"
)

type CustomTemplates struct {
	ResourceTemplate      map[string]string
	StorageTemplate       map[string]string
	ServiceConfigTemplate map[string]string
}

func (cusTemp *CustomTemplates) CreatePdsCustomTemplatesAndFetchIds(tenantId string, templates *parameters.NewPDSParams, updateTemplate bool) (string, string, string, error) {

	//Todo: Mechanism to populate dynamic/Unknown key-value pairs for App config
	//ToDo: Take configurationValue incrementation count from user/testcase

	//Initializing the parameters required for workload generation
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
	}
	appConfig, _ := pdslibs.CreateServiceConfigTemplate(tenantId, appConfigParams)
	appConfigId := appConfig.Create.Meta.Uid
	cusTemp.ServiceConfigTemplate[serviceConfigTempID] = *appConfigId

	stConfig, _ := pdslibs.CreateStorageConfigTemplate(tenantId, stConfigParams)
	stConfigId := stConfig.Create.Meta.Uid
	cusTemp.StorageTemplate[storageTempID] = *stConfigId

	resConfig, _ := pdslibs.CreateResourceConfigTemplate(tenantId, resConfigParams)
	resourceConfigId := resConfig.Create.Meta.Uid
	cusTemp.ResourceTemplate[resourceTempID] = *resourceConfigId
	return *appConfigId, *stConfigId, *resourceConfigId, nil
}
