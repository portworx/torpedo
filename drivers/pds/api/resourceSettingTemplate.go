// Package api comprises of all the components and associated CRUD functionality
package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type ResourceSettingsTemplate struct {
	context   context.Context
	apiClient *pds.APIClient
}

func (rt *ResourceSettingsTemplate) ListTemplates(tenantId string) ([]pds.ModelsResourceSettingsTemplate, error) {
	rtClient := rt.apiClient.ResourceSettingsTemplatesApi
	rtModel, res, err := rtClient.ApiTenantsIdResourceSettingsTemplatesGet(rt.context, tenantId).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdResourceSettingsTemplatesGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return rtModel.GetData(), nil
}

func (rt *ResourceSettingsTemplate) GetTemplate(templateId string) (*pds.ModelsResourceSettingsTemplate, error) {
	rtClient := rt.apiClient.ResourceSettingsTemplatesApi
	rtModel, res, err := rtClient.ApiResourceSettingsTemplatesIdGet(rt.context, templateId).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiResourceSettingsTemplatesIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return rtModel, nil
}

func (rt *ResourceSettingsTemplate) CreateTemplate(tenantId string, cpuLimit string, cpuRequest string, dataServiceId string, memoryLimit string, memoryRequest string, name string, storageRequest string) (*pds.ModelsResourceSettingsTemplate, error) {
	rtClient := rt.apiClient.ResourceSettingsTemplatesApi
	createRequest := pds.ControllersCreateResourceSettingsTemplatesRequest{CpuLimit: &cpuLimit, CpuRequest: &cpuRequest, DataServiceId: &dataServiceId, MemoryLimit: &memoryLimit, MemoryRequest: &memoryRequest, Name: &name, StorageRequest: &storageRequest}
	rtModel, res, err := rtClient.ApiTenantsIdResourceSettingsTemplatesPost(rt.context, tenantId).Body(createRequest).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdResourceSettingsTemplatesPost``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return rtModel, nil
}

func (rt *ResourceSettingsTemplate) UpdateTemplate(templateId string, cpuLimit string, cpuRequest string, memoryLimit string, memoryRequest string, name string, storageRequest string) (*pds.ModelsResourceSettingsTemplate, error) {
	rtClient := rt.apiClient.ResourceSettingsTemplatesApi
	updateRequest := pds.ControllersUpdateResourceSettingsTemplateRequest{CpuLimit: &cpuLimit, CpuRequest: &cpuRequest, MemoryLimit: &memoryLimit, MemoryRequest: &memoryRequest, Name: &name, StorageRequest: &storageRequest}
	rtModel, res, err := rtClient.ApiResourceSettingsTemplatesIdPut(rt.context, templateId).Body(updateRequest).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiResourceSettingsTemplatesIdPut``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return rtModel, nil
}

func (rt *ResourceSettingsTemplate) DeleteTemplate(templateId string) (*status.Response, error) {
	rtClient := rt.apiClient.ResourceSettingsTemplatesApi
	res, err := rtClient.ApiResourceSettingsTemplatesIdDelete(rt.context, templateId).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiResourceSettingsTemplatesIdDelete``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return res, nil
}
