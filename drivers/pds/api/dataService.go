package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type DataService struct {
	context   context.Context
	apiClient *pds.APIClient
}

func (ds *DataService) ListDataServices() ([]pds.ModelsDataService, error) {
	dsClient := ds.apiClient.DataServicesApi
	dsModels, res, err := dsClient.ApiDataServicesGet(ds.context).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDataServicesGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return dsModels.GetData(), err
}

func (ds *DataService) GetDataService(dataServiceId string) (*pds.ModelsDataService, error) {
	dsClient := ds.apiClient.DataServicesApi
	dsModel, res, err := dsClient.ApiDataServicesIdGet(ds.context, dataServiceId).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDataServicesIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return dsModel, err
}
