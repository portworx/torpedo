package api

import (
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	platformV2 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// TargetClusterRegistrationV2 struct
type TargetClusterRegistrationV2 struct {
	ApiClientV2 *platformV2.APIClient
}

// GenerateTargetClusterRegistrationManifest return deployment targets models.
func (tc *TargetClusterRegistrationV2) GenerateTargetClusterRegistrationManifest(tenantId string) (*platformV2.V1TargetClusterRegistrationManifest, error) {
	tcClient := tc.ApiClientV2.TargetClusterRegistrationManifestServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	tcModels, res, err := tcClient.TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifest(ctx, tenantId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TargetClusterServiceListTargetClusters`: %v\n.Full HTTP response: %v", err, res)
	}
	return tcModels, nil
}
