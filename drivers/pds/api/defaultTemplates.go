package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// DefaultTemplates struct
type DefaultTemplates struct {
	context   context.Context
	apiClient *pds.APIClient
}

// ListApplicationConfigurationTemplates func
func (ds *DefaultTemplates) ListApplicationConfigurationTemplates() ([]pds.ModelsApplicationConfigurationTemplate, error) {
	dsClient := ds.apiClient.DefaultTemplatesApi
	dsModels, res, err := dsClient.ApiDefaultTemplatesApplicationConfigurationGet(ds.context).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDataServicesGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return dsModels.GetData(), err
}

// ListResourceSettingTemplates func
func (ds *DefaultTemplates) ListResourceSettingTemplates() ([]pds.ModelsResourceSettingsTemplate, error) {
	dsClient := ds.apiClient.DefaultTemplatesApi
	dsModels, res, err := dsClient.ApiDefaultTemplatesResourceSettingsGet(ds.context).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDataServicesGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return dsModels.GetData(), err
}

// ListStorageOptionsTemplates func
func (ds *DefaultTemplates) ListStorageOptionsTemplates() ([]pds.ModelsStorageOptionsTemplate, error) {
	dsClient := ds.apiClient.DefaultTemplatesApi
	dsModels, res, err := dsClient.ApiDefaultTemplatesStorageOptionsGet(ds.context).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDataServicesGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return dsModels.GetData(), err
}
