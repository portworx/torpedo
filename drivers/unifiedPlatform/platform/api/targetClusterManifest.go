package api

import (
	"context"
	"fmt"
	utils "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	platformV2 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
	"time"
)

// TargetClusterManifestV2 struct
type TargetClusterManifestV2 struct {
	ApiClientV2 *platformV2.APIClient
}

func (dt *TargetClusterManifestV2) GetTargetClusterRegistrationManifest(tenantId string, clusterName string, pConfig *utils.ProxyConfig, crConfig *utils.CustomRegistryConfig) (string, error) {
	var tcManifestClient platformV2.ApiTargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifestRequest

	tcManifestRequest := tcManifestClient.ApiService.TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifest(context.Background(), tenantId)

	if clusterName == "" {
		clusterName = fmt.Sprintf("Cluster_%v", time.Now())
	}
	tcManifestRequest = tcManifestRequest.TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifestBody(
		platformV2.TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifestBody{
			ClusterName: &clusterName,
			Config: &platformV2.V1Config5{
				CustomImageRegistryConfig: &platformV2.V1CustomImageRegistryConfig{
					RegistryUrl:       &crConfig.RegistryUrl,
					RegistryNamespace: &crConfig.RegistryNamespace,
					Username:          &crConfig.RegistryUserName,
					Password:          &crConfig.RegistryPassword,
					CaCert:            &crConfig.CaCert,
				},
				ProxyConfig: &platformV2.V1ProxyConfig{
					HttpUrl:  &pConfig.HttpUrl,
					HttpsUrl: &pConfig.HttpsUrl,
					Username: &pConfig.Username,
					Password: &pConfig.Password,
					NoProxy:  &pConfig.NoProxy,
					CaCert:   &pConfig.CaCert,
				},
			},
		})

	dtClient := dt.ApiClientV2.TargetClusterRegistrationManifestServiceAPI
	dtModels, res, err := dtClient.TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifestExecute(tcManifestRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return "", fmt.Errorf("Error when calling `TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifest`: %v\n.Full HTTP response: %v", err, res)
	}
	return *dtModels.Manifest, nil
}
