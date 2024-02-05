package api

import (
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	platformV2 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// TargetClusterManifestV2 struct
type TargetClusterManifestV2 struct {
	ApiClientV2 *platformV2.APIClient
}

func (dt *TargetClusterManifestV2) GetTargetClusterRegistrationManifest(tenantId string) (string, error) {
	dtClient := dt.ApiClientV2.TargetClusterRegistrationManifestServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return "", fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dtModels, res, err := dtClient.TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifest(ctx, tenantId).Execute()
	// dtModels, res, err := dtClient.TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifestExecute()
	if err != nil && res.StatusCode != status.StatusOK {
		return "", fmt.Errorf("Error when calling `TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifest`: %v\n.Full HTTP response: %v", err, res)
	}
	return *dtModels.Manifest, nil
}

func (dt *TargetClusterManifestV2) GetTargetClusterRegistrationManifest1(tenantId string, clusterName string) (string, error) {
	var createRequest platformV2.ApiTargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifestRequest

	createRequest.TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifestBody(platformV2.TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifestBody{
		ClusterName: &clusterName,
		Config: &platformV2.V1Config5{
			CustomImageRegistryConfig: &platformV2.V1CustomImageRegistryConfig{
				RegistryUrl:       nil,
				RegistryNamespace: nil,
				Username:          nil,
				Password:          nil,
				CaCert:            nil,
			},
			ProxyConfig: &platformV2.V1ProxyConfig{
				HttpUrl:  nil,
				HttpsUrl: nil,
				Username: nil,
				Password: nil,
				NoProxy:  nil,
				CaCert:   nil,
			},
		},
	})

	dtClient := dt.ApiClientV2.TargetClusterRegistrationManifestServiceAPI

	dtModels, res, err := dtClient.TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifestExecute(createRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return "", fmt.Errorf("Error when calling `TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifest`: %v\n.Full HTTP response: %v", err, res)
	}
	return *dtModels.Manifest, nil
}
