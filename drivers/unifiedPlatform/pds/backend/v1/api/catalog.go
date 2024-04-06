package api

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	catalogv1 "github.com/pure-px/platform-api-go-client/pds/v1/catalog"
	status "net/http"
)

// ListDataServices return list of data services in the catalog
func (ds *PDS_API_V1) ListDataServices() (*automationModels.CatalogResponse, error) {
	dsListResponse := &automationModels.CatalogResponse{
		DataServiceList: []automationModels.V1DataService{},
	}
	ctx, dsClient, err := ds.getCatalogClient()
	if err != nil {
		return dsListResponse, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	dsListRequest := dsClient.DataServicesServiceListDataServices(ctx)

	//dsListRequest := catalogv1.ApiDataServicesServiceListDataServicesRequest{}

	dsModel, res, err := dsClient.DataServicesServiceListDataServicesExecute(dsListRequest)

	if err != nil || res.StatusCode != status.StatusOK {
		return dsListResponse, fmt.Errorf("Error when calling `DataServicesServiceListDataServices`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(dsModel.DataServices, &dsListResponse.DataServiceList)
	if err != nil {
		return nil, err
	}
	return dsListResponse, nil
}

func (ds *PDS_API_V1) ListDataServiceVersions(input *automationModels.WorkFlowRequest) (*automationModels.CatalogResponse, error) {
	dsListResponse := &automationModels.CatalogResponse{
		DataServiceVersionList: []automationModels.V1DataService{},
	}
	ctx, dsClient, err := ds.getDSVersionsClient()

	if err != nil {
		return dsListResponse, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	dsRequest := catalogv1.ApiDataServiceVersionServiceListDataServiceVersionsRequest{}
	dsRequest = dsRequest.ApiService.DataServiceVersionServiceListDataServiceVersions(ctx, input.DataServiceId)

	dsVersionsModel, res, err := dsClient.DataServiceVersionServiceListDataServiceVersionsExecute(dsRequest)

	if err != nil || res.StatusCode != status.StatusOK {
		return dsListResponse, fmt.Errorf("Error when calling `DataServicesServiceGetDataServiceVersions`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(dsVersionsModel.DataServiceVersions, &dsListResponse.DataServiceVersionList)
	if err != nil {
		return nil, err
	}
	return dsListResponse, nil
}

func (ds *PDS_API_V1) ListDataServiceImages(input *automationModels.WorkFlowRequest) (*automationModels.CatalogResponse, error) {
	dsListResponse := &automationModels.CatalogResponse{
		DataServiceImageList: []automationModels.V1Image{},
	}
	_, dsClient, err := ds.getDSImagesClient()

	if err != nil {
		return dsListResponse, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	dsRequest := catalogv1.ApiImageServiceListImagesRequest{}
	dsRequest = dsRequest.DataServiceId(input.DataServiceId)
	dsRequest.DataServiceVersionId(input.DataServiceVersionId)

	dsImagesModel, res, err := dsClient.ImageServiceListImagesExecute(dsRequest)

	if err != nil && res.StatusCode != status.StatusOK {
		return dsListResponse, fmt.Errorf("Error when calling `DataServicesServiceGetDataServiceImages`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(dsImagesModel.Images, &dsListResponse.DataServiceImageList)
	if err != nil {
		return nil, err
	}
	return dsListResponse, nil
}
