package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// StorageSettingsTemplate struct
type StorageSettingsTemplate struct {
	context   context.Context
	apiClient *pds.APIClient
}

// ListTemplates func
func (st *StorageSettingsTemplate) ListTemplates(tenantID string) ([]pds.ModelsStorageOptionsTemplate, error) {
	stClient := st.apiClient.StorageOptionsTemplatesApi
	log.Infof("Get list of storage templates for tenant ID - %v", tenantID)
	pdsStorageTemplates, res, err := stClient.ApiTenantsIdStorageOptionsTemplatesGet(st.context, tenantID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdStorageOptionsTemplatesGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return pdsStorageTemplates.GetData(), nil
}

// GetTemplate func
func (st *StorageSettingsTemplate) GetTemplate(templateID string) (*pds.ModelsStorageOptionsTemplate, error) {
	stClient := st.apiClient.StorageOptionsTemplatesApi
	log.Infof("Get storage template details for UUID - %v", templateID)
	stModel, res, err := stClient.ApiStorageOptionsTemplatesIdGet(st.context, templateID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiStorageOptionsTemplatesIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return stModel, nil
}

// CreateTemplate func
func (st *StorageSettingsTemplate) CreateTemplate(tenantID string, fg bool, fs string, name string, repl int32, secure bool) (*pds.ModelsStorageOptionsTemplate, error) {
	stClient := st.apiClient.StorageOptionsTemplatesApi
	log.Info("Create new storage template.")
	createRequest := pds.ControllersCreateStorageOptionsTemplatesRequest{Fg: &fg, Fs: &fs, Name: &name, Repl: &repl, Secure: &secure}
	stModel, res, err := stClient.ApiTenantsIdStorageOptionsTemplatesPost(st.context, tenantID).Body(createRequest).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdStorageOptionsTemplatesPost``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return stModel, nil
}

// UpdateTemplate func
func (st *StorageSettingsTemplate) UpdateTemplate(templateID string, fg bool, fs string, name string, repl int32, secure bool) (*pds.ModelsStorageOptionsTemplate, error) {
	stClient := st.apiClient.StorageOptionsTemplatesApi
	log.Info("Create new storage template.")
	updateRequest := pds.ControllersUpdateStorageOptionsTemplateRequest{Fg: &fg, Fs: &fs, Name: &name, Repl: &repl, Secure: &secure}
	stModel, res, err := stClient.ApiStorageOptionsTemplatesIdPut(st.context, templateID).Body(updateRequest).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiStorageOptionsTemplatesIdPut``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return stModel, nil
}

// DeleteTemplate func
func (st *StorageSettingsTemplate) DeleteTemplate(templateID string) (*status.Response, error) {
	stClient := st.apiClient.StorageOptionsTemplatesApi
	log.Infof("Delete strogae template: %v", templateID)
	res, err := stClient.ApiStorageOptionsTemplatesIdDelete(st.context, templateID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiStorageOptionsTemplatesIdDelete``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return res, nil
}
