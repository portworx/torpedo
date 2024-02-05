package api

import (
	"context"
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	status "net/http"
)

// DataServiceV2 struct
type DataServiceV2 struct {
	ApiClientV2 *pdsv2.APIClient
	AccountID   string
}

// GetClient updates the header with bearer token and returns the new client
func (ds *DataServiceV2) GetClient() (context.Context, *pdsv2.DataServicesServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	ds.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	ds.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = ds.AccountID
	client := ds.ApiClientV2.DataServicesServiceAPI

	return ctx, client, nil
}

// ListDataServices return data services models.
func (ds *DataServiceV2) ListDataServices() ([]pdsv2.V1DataService, error) {
	ctx, dsClient, err := ds.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dsModels, res, err := dsClient.DataServicesServiceListDataServices(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DataServicesServiceListDataServices`: %v\n.Full HTTP response: %v", err, res)
	}
	return dsModels.DataServices, err
}

// GetDataService return data service model.
func (ds *DataServiceV2) GetDataService(dataServiceID string) (*pdsv2.V1DataService, error) {
	ctx, dsClient, err := ds.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dsModel, res, err := dsClient.DataServicesServiceGetDataService(ctx, dataServiceID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DataServicesServiceGetDataService`: %v\n.Full HTTP response: %v", err, res)
	}
	return dsModel, err
}
