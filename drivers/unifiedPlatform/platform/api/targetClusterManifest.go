package api

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	platformV2 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// TargetClusterManifestV2 struct
type TargetClusterManifestV2 struct {
	ApiClientV2 *platformV2.APIClient
	AccountID   string
}

// GetClient updates the header with bearer token and returns the new client
func (dtm *TargetClusterManifestV2) GetClient() (context.Context, *platformV2.TargetClusterRegistrationManifestServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	dtm.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	dtm.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = dtm.AccountID
	client := dtm.ApiClientV2.TargetClusterRegistrationManifestServiceAPI
	return ctx, client, nil
}

func (dtm *TargetClusterManifestV2) GetTargetClusterRegistrationManifest(tenantId string) (string, error) {
	ctx, dtmClient, err := dtm.GetClient()
	if err != nil {
		return "", fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dtmModels, res, err := dtmClient.TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifest(ctx, tenantId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return "", fmt.Errorf("Error when calling `TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifest`: %v\n.Full HTTP response: %v", err, res)
	}
	return *dtmModels.Manifest, nil
}

func (dtm *TargetClusterManifestV2) GetTargetClusterRegistrationManifest1(tenantId string, clusterName string) (string, error) {
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

	dtmClient := dtm.ApiClientV2.TargetClusterRegistrationManifestServiceAPI

	dtmModels, res, err := dtmClient.TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifestExecute(createRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return "", fmt.Errorf("Error when calling `TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifest`: %v\n.Full HTTP response: %v", err, res)
	}
	return *dtmModels.Manifest, nil
}
