package api

import (
	"context"
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
	targetClusterManifestv1 "github.com/pure-px/platform-api-go-client/platform/v1/targetclusterregistrationmanifest"
	status "net/http"
	"time"
)

func (tcManifest *PLATFORM_API_V1) GetTargetClusterRegistrationManifest(getManifestRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {

	response := &apiStructs.WorkFlowResponse{
		TargetCluster: apiStructs.PlatformTargetClusterOutput{
			Manifest: apiStructs.PlatformManifestOutput{},
		},
	}
	var tcManifestRequest targetClusterManifestv1.ApiTargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifestRequest

	clusterName := getManifestRequest.TargetCluster.GetManifest.ClusterName
	tenantId := getManifestRequest.TargetCluster.GetManifest.TenantId

	tcManifestRequest = tcManifestRequest.ApiService.TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifest(context.Background(), tenantId)

	if clusterName == "" {
		clusterName = fmt.Sprintf("Cluster_%v", time.Now().Unix())
	}

	tcManifestRequest = tcManifestRequest.TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifestBody(
		targetClusterManifestv1.TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifestBody{
			ClusterName: &clusterName,
		})

	_, dtClient, err := tcManifest.getTargetClusterManifestClient()
	dtModels, res, err := dtClient.TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifestExecute(tcManifestRequest)

	if err != nil || res.StatusCode != status.StatusOK {
		return response, fmt.Errorf("Error when calling `TargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifest`: %v\n.Full HTTP response: %v", err, res)
	}

	response.TargetCluster.Manifest.Manifest = dtModels.GetManifest()

	log.Infof("Manifest - [%s]", response.TargetCluster.Manifest.Manifest)

	return response, nil

}
