// Package api comprises of all the components and associated CRUD functionality
package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// Version struct
type Version struct {
	context   context.Context
	apiClient *pds.APIClient
}

// ListDataServiceVersions func
func (v *Version) ListDataServiceVersions(dataServiceID string) ([]pds.ModelsVersion, error) {
	versionClient := v.apiClient.VersionsApi
	versionModels, res, err := versionClient.ApiDataServicesIdVersionsGet(v.context, dataServiceID).Execute()
	data := versionModels.GetData()
	for i := range data {
		log.Info(data[i])
	}
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDataServicesIdVersionsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return versionModels.GetData(), err
}

// GetVersion func
func (v *Version) GetVersion(versionID string) (*pds.ModelsVersion, error) {
	versionClient := v.apiClient.VersionsApi
	versionModel, res, err := versionClient.ApiVersionsIdGet(v.context, versionID).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiVersionsIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return versionModel, err
}
