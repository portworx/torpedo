package api

import (
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	catalogv1 "github.com/pure-px/platform-api-go-client/pds/v1/catalog"
	status "net/http"
)

// ListDataServices return list of data services in the catalog
func (ds *PDS_API_V1) ListDataServices() ([]automationModels.WorkFlowResponse, error) {
	dsListResponse := []automationModels.WorkFlowResponse{}
	_, dsClient, err := ds.getCatalogClient()
	if err != nil {
		return dsListResponse, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	dsListRequest := catalogv1.ApiDataServicesServiceListDataServicesRequest{}

	dsModel, res, err := dsClient.DataServicesServiceListDataServicesExecute(dsListRequest)

	if err != nil && res.StatusCode != status.StatusOK {
		return dsListResponse, fmt.Errorf("Error when calling `DataServicesServiceListDataServices`: %v\n.Full HTTP response: %v", err, res)
	}
	err = copier.Copy(&dsListResponse, dsModel)
	if err != nil {
		return nil, err
	}
	return dsListResponse, nil
}

func (ds *PDS_API_V1) ListDataServiceVersions(input *automationModels.WorkFlowRequest) ([]automationModels.WorkFlowResponse, error) {
	dsVersionsResponse := []automationModels.WorkFlowResponse{}
	ctx, dsClient, err := ds.getDSVersionsClient()

	if err != nil {
		return dsVersionsResponse, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	dsRequest := catalogv1.ApiDataServiceVersionServiceListDataServiceVersionsRequest{}
	dsRequest = dsRequest.ApiService.DataServiceVersionServiceListDataServiceVersions(ctx, input.DataServiceId)

	dsVersionsModel, res, err := dsClient.DataServiceVersionServiceListDataServiceVersionsExecute(dsRequest)

	if err != nil && res.StatusCode != status.StatusOK {
		return dsVersionsResponse, fmt.Errorf("Error when calling `DataServicesServiceGetDataServiceVersions`: %v\n.Full HTTP response: %v", err, res)
	}
	err = copier.Copy(&dsVersionsResponse, dsVersionsModel)
	if err != nil {
		return nil, err
	}
	return dsVersionsResponse, nil
}

func (ds *PDS_API_V1) ListDataServiceImages(input *automationModels.WorkFlowRequest) ([]automationModels.WorkFlowResponse, error) {
	dsImagesResponse := []automationModels.WorkFlowResponse{}
	_, dsClient, err := ds.getDSImagesClient()

	if err != nil {
		return dsImagesResponse, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	dsRequest := catalogv1.ApiImageServiceListImagesRequest{}
	dsRequest = dsRequest.DataServiceId(input.DataServiceId)
	dsRequest.DataServiceVersionId(input.DataServiceVersionId)

	dsImagesModel, res, err := dsClient.ImageServiceListImagesExecute(dsRequest)

	if err != nil && res.StatusCode != status.StatusOK {
		return dsImagesResponse, fmt.Errorf("Error when calling `DataServicesServiceGetDataServiceImages`: %v\n.Full HTTP response: %v", err, res)
	}
	err = copier.Copy(&dsImagesResponse, dsImagesModel)
	if err != nil {
		return nil, err
	}
	return dsImagesResponse, nil
}
