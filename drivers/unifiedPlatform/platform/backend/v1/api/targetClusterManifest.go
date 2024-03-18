package api

import (
	"context"
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
	targetClusterManifestv1 "github.com/pure-px/platform-api-go-client/platform/v1/targetclusterregistrationmanifest"
	status "net/http"
	"time"
)

func (tcManifest *PLATFORM_API_V1) GetTargetClusterRegistrationManifest(getManifestRequest *automationModels.PlatformTargetCluster) (*automationModels.WorkFlowResponse, error) {

	response := &automationModels.WorkFlowResponse{
		TargetCluster: automationModels.PlatformTargetClusterOutput{
			Manifest: automationModels.PlatformManifestOutput{},
		},
	}
	var tcManifestRequest targetClusterManifestv1.ApiTargetClusterRegistrationManifestServiceGenerateTargetClusterRegistrationManifestRequest

	clusterName := getManifestRequest.GetManifest.ClusterName
	tenantId := getManifestRequest.GetManifest.TenantId

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
