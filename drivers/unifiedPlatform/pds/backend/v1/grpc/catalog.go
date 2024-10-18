package grpc

import (
	"github.com/pure-px/torpedo/drivers/unifiedPlatform/automationModels"
)

func (deployment *PdsGrpc) ListDataServices() (*automationModels.CatalogResponse, error) {
	return nil, nil
}

func (deployment *PdsGrpc) ListDataServiceVersions(input *automationModels.WorkFlowRequest) (*automationModels.CatalogResponse, error) {
	return nil, nil
}

func (deployment *PdsGrpc) ListDataServiceImages(input *automationModels.WorkFlowRequest) (*automationModels.CatalogResponse, error) {
	return nil, nil
}
