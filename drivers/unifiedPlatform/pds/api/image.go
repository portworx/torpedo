package api

import (
	"context"
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	status "net/http"
)

// ImageV2 struct
type ImageV2 struct {
	ApiClientV2 *pdsv2.APIClient
	AccountID   string
}

// GetClient updates the header with bearer token and returns the new client
func (img *ImageV2) GetClient() (context.Context, *pdsv2.ImageServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	img.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	img.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = img.AccountID
	client := img.ApiClientV2.ImageServiceAPI
	return ctx, client, nil
}

// ListImages return images models for given version.
func (img *ImageV2) ListImages() ([]pdsv2.V1Image, error) {
	ctx, imgClient, err := img.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	imgModels, res, err := imgClient.ImageServiceListImages(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiDeploymentsIdBackupsGet`: %v\n.Full HTTP response: %v", err, res)
	}
	return imgModels.Images, err
}

// GetImage return image model.
func (img *ImageV2) GetImage(imageID string) (*pdsv2.V1Image, error) {
	ctx, imgClient, err := img.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	imgModel, res, err := imgClient.ImageServiceGetImage(ctx, imageID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiDeploymentsIdBackupsGet`: %v\n.Full HTTP response: %v", err, res)
	}
	return imgModel, err
}
