package api

import (
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
)

// DeploymentManifestV2 struct
type DeploymentManifestV2 struct {
	ApiClientv2 *pdsv2.APIClient
}
