// Package api comprises of all the components and associated CRUD functionality
package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type ApiVersion struct {
	context   context.Context
	apiClient *pds.APIClient
}

func (v *ApiVersion) GetHelmChartVersion() (string, error) {
	versionClient := v.apiClient.APIVersionApi
	versionModel, res, err := versionClient.ApiVersionGet(v.context).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiVersionGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return versionModel.GetHelmChartVersion(), err
}
