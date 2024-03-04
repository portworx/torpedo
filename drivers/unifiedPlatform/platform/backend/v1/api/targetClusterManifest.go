package api

import (
	"context"
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	platformv1 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
	"time"
)

// GetClient updates the header with bearer token and returns the new client
func (tcManifest *PLATFORM_API_V1) getTargetClusterManifestClient() (context.Context, *platformv1.TargetClusterRegistrationManifestServiceAPIService, error) {
	log.Infof("Creating client from PLATFORM_API_V1 package")
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	tcManifest.ApiClientV1.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	tcManifest.ApiClientV1.GetConfig().DefaultHeader["px-account-id"] = tcManifest.AccountID

	client := tcManifest.ApiClientV1.TargetClusterRegistrationManifestServiceAPI
	return ctx, client, nil
}

func (tcManifest *PLATFORM_API_V1) GetTargetClusterRegistrationManifest(getManifestRequest *apiStructs.WorkFlowRequest) (string, error) {

	var tcManifestRequest platformv1.ApiTargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifestRequest

	clusterName := getManifestRequest.TargetClusterManifest.ClusterName
	tenantId := getManifestRequest.TargetClusterManifest.TenantId

	tcManifestRequest = tcManifestRequest.ApiService.TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifest(context.Background(), tenantId)

	if clusterName == "" {
		clusterName = fmt.Sprintf("Cluster_%v", time.Now().Unix())
	}

	tcManifestRequest = tcManifestRequest.TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifestBody(
		platformv1.TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifestBody{
			ClusterName: &clusterName,
		})

	_, dtClient, err := tcManifest.getTargetClusterManifestClient()
	dtModels, res, err := dtClient.TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifestExecute(tcManifestRequest)

	if err != nil || res.StatusCode != status.StatusOK {
		return "", fmt.Errorf("Error when calling `TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifest`: %v\n.Full HTTP response: %v", err, res)
	}
	return *dtModels.Manifest, nil

}
