package api

import (
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
)

// DeploymentTopologyV2 struct
type DeploymentTopologyV2 struct {
	ApiClientV2 *pdsv2.APIClient
	AccountID   string
}
