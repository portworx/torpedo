package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// Image struct
type Image struct {
	context   context.Context
	apiClient *pds.APIClient
}

// ListImages func
func (img *Image) ListImages(versionID string) ([]pds.ModelsImage, error) {
	imgClient := img.apiClient.ImagesApi

	imgModels, res, err := imgClient.ApiVersionsIdImagesGet(img.context, versionID).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiVersionsIdImagesGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return imgModels.GetData(), err
}

// GetImage func
func (img *Image) GetImage(imageID string) (*pds.ModelsImage, error) {
	imgClient := img.apiClient.ImagesApi

	imgModel, res, err := imgClient.ApiImagesIdGet(img.context, imageID).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiImagesIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return imgModel, err
}
