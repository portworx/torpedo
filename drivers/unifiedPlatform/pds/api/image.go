package api

import (
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	status "net/http"
)

// ImageV2 struct
type ImageV2 struct {
	ApiClientV2 *pdsv2.APIClient
}

// ListImages return images models for given version.
func (img *ImageV2) ListImages() ([]pdsv2.V1Image, error) {
	imgClient := img.ApiClientV2.ImageServiceAPI
	ctx, err := GetContext()
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
	imgClient := img.ApiClientV2.ImageServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	imgModel, res, err := imgClient.ImageServiceGetImage(ctx, imageID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ApiDeploymentsIdBackupsGet`: %v\n.Full HTTP response: %v", err, res)
	}
	return imgModel, err
}
