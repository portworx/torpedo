package pds

import (
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
)

type PDS interface {
	CreateDeployment(pdsv2.ApiDeploymentServiceCreateDeploymentRequest) (*ApiResponse, error)
}
