// Package api comprises of all the components and associated CRUD functionality
package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// ResourceSettingsTemplate struct
type ResourceSettingsTemplate struct {
	context   context.Context
	apiClient *pds.APIClient
}

// ListTemplates func
func (rt *ResourceSettingsTemplate) ListTemplates(tenantID string) ([]pds.ModelsResourceSettingsTemplate, error) {
	rtClient := rt.apiClient.ResourceSettingsTemplatesApi
	rtModel, res, err := rtClient.ApiTenantsIdResourceSettingsTemplatesGet(rt.context, tenantID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdResourceSettingsTemplatesGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return rtModel.GetData(), nil
}

// GetTemplate func
func (rt *ResourceSettingsTemplate) GetTemplate(templateID string) (*pds.ModelsResourceSettingsTemplate, error) {
	rtClient := rt.apiClient.ResourceSettingsTemplatesApi
	rtModel, res, err := rtClient.ApiResourceSettingsTemplatesIdGet(rt.context, templateID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiResourceSettingsTemplatesIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return rtModel, nil
}

// CreateTemplate func
func (rt *ResourceSettingsTemplate) CreateTemplate(tenantID string, cpuLimit string, cpuRequest string, dataServiceID string, memoryLimit string, memoryRequest string, name string, storageRequest string) (*pds.ModelsResourceSettingsTemplate, error) {
	rtClient := rt.apiClient.ResourceSettingsTemplatesApi
	createRequest := pds.ControllersCreateResourceSettingsTemplatesRequest{CpuLimit: &cpuLimit, CpuRequest: &cpuRequest, DataServiceId: &dataServiceID, MemoryLimit: &memoryLimit, MemoryRequest: &memoryRequest, Name: &name, StorageRequest: &storageRequest}
	rtModel, res, err := rtClient.ApiTenantsIdResourceSettingsTemplatesPost(rt.context, tenantID).Body(createRequest).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdResourceSettingsTemplatesPost``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return rtModel, nil
}

// UpdateTemplate func
func (rt *ResourceSettingsTemplate) UpdateTemplate(templateID string, cpuLimit string, cpuRequest string, memoryLimit string, memoryRequest string, name string, storageRequest string) (*pds.ModelsResourceSettingsTemplate, error) {
	rtClient := rt.apiClient.ResourceSettingsTemplatesApi
	updateRequest := pds.ControllersUpdateResourceSettingsTemplateRequest{CpuLimit: &cpuLimit, CpuRequest: &cpuRequest, MemoryLimit: &memoryLimit, MemoryRequest: &memoryRequest, Name: &name, StorageRequest: &storageRequest}
	rtModel, res, err := rtClient.ApiResourceSettingsTemplatesIdPut(rt.context, templateID).Body(updateRequest).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiResourceSettingsTemplatesIdPut``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return rtModel, nil
}

// DeleteTemplate func
func (rt *ResourceSettingsTemplate) DeleteTemplate(templateID string) (*status.Response, error) {
	rtClient := rt.apiClient.ResourceSettingsTemplatesApi
	res, err := rtClient.ApiResourceSettingsTemplatesIdDelete(rt.context, templateID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiResourceSettingsTemplatesIdDelete``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return res, nil
}
