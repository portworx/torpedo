// Package api comprises of all the components and associated CRUD functionality
package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// AppConfigTemplate struct
type AppConfigTemplate struct {
	context   context.Context
	apiClient *pds.APIClient
}

// ListTemplates func
func (at *AppConfigTemplate) ListTemplates(tenantID string) ([]pds.ModelsApplicationConfigurationTemplate, error) {
	atClient := at.apiClient.ApplicationConfigurationTemplatesApi
	atModel, res, err := atClient.ApiTenantsIdApplicationConfigurationTemplatesGet(at.context, tenantID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdApplicationConfigurationTemplatesGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return atModel.GetData(), nil
}

// GetTemplate func
func (at *AppConfigTemplate) GetTemplate(templateID string) (*pds.ModelsApplicationConfigurationTemplate, error) {
	atClient := at.apiClient.ApplicationConfigurationTemplatesApi
	log.Infof("Get list of storage templates for tenant ID - %v", templateID)
	atModel, res, err := atClient.ApiApplicationConfigurationTemplatesIdGet(at.context, templateID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiApplicationConfigurationTemplatesIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return atModel, nil
}

// CreateTemplate func
func (at *AppConfigTemplate) CreateTemplate(tenantID string, dataServiceID string, name string, data []pds.ModelsConfigItem) (*pds.ModelsApplicationConfigurationTemplate, error) {
	atClient := at.apiClient.ApplicationConfigurationTemplatesApi
	log.Info("Create new resource template.")
	createRequest := pds.ControllersCreateApplicationConfigurationTemplatesRequest{ConfigItems: data, DataServiceId: &dataServiceID, Name: &name}
	atModel, res, err := atClient.ApiTenantsIdApplicationConfigurationTemplatesPost(at.context, tenantID).Body(createRequest).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdApplicationConfigurationTemplatesPost``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return atModel, nil
}

// UpdateTemplate func
func (at *AppConfigTemplate) UpdateTemplate(templateID string, deployTime bool, key string, value string, name string) (*pds.ModelsApplicationConfigurationTemplate, error) {
	atClient := at.apiClient.ApplicationConfigurationTemplatesApi
	log.Info("Create new resource template.")
	data := []pds.ModelsConfigItem{{DeployTime: &deployTime, Key: &key, Value: &value}}
	updateRequest := pds.ControllersUpdateApplicationConfigurationTemplateRequest{ConfigItems: data, Name: &name}
	atModel, res, err := atClient.ApiApplicationConfigurationTemplatesIdPut(at.context, templateID).Body(updateRequest).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiApplicationConfigurationTemplatesIdPut``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return atModel, nil
}

// DeleteTemplate func
func (at *AppConfigTemplate) DeleteTemplate(templateID string) (*status.Response, error) {
	atClient := at.apiClient.ApplicationConfigurationTemplatesApi
	log.Infof("Get list of storage templates for tenant ID - %v", templateID)
	res, err := atClient.ApiApplicationConfigurationTemplatesIdDelete(at.context, templateID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiApplicationConfigurationTemplatesIdDelete``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return res, nil
}
