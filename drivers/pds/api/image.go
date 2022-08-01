package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type Image struct {
	context   context.Context
	apiClient *pds.APIClient
}

func (img *Image) ListImages(versionId string) ([]pds.ModelsImage, error) {
	imgClient := img.apiClient.ImagesApi

	imgModels, res, err := imgClient.ApiVersionsIdImagesGet(img.context, versionId).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiVersionsIdImagesGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return imgModels.GetData(), err
}

func (img *Image) GetImage(imageId string) (*pds.ModelsImage, error) {
	imgClient := img.apiClient.ImagesApi

	imgModel, res, err := imgClient.ApiImagesIdGet(img.context, imageId).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiImagesIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return imgModel, err
}
