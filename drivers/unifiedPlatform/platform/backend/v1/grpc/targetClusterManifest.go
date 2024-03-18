package grpc

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	publictcapis "github.com/pure-px/apis/public/portworx/platform/targetclusterregistrationmanifest/apiv1"
	"google.golang.org/grpc"
)

// getTargetClusterClient updates the header with bearer token and returns the new client
func (tcGrpc *PlatformGrpc) getTargetClusterManifestClient() (context.Context, publictcapis.TargetClusterRegistrationManifestServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var tcClient publictcapis.TargetClusterRegistrationManifestServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	tcClient = publictcapis.NewTargetClusterRegistrationManifestServiceClient(tcGrpc.ApiClientV1)

	return ctx, tcClient, token, nil
}

func (tcGrpc *PlatformGrpc) GetTargetClusterRegistrationManifest(getManifestRequest *PlatformTargetCluster) (*WorkFlowResponse, error) {

	response := &WorkFlowResponse{
		TargetCluster: PlatformTargetClusterOutput{
			Manifest: PlatformManifestOutput{},
		},
	}

	ctx, client, _, err := tcGrpc.getTargetClusterManifestClient()
	if err != nil {
		return response, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	getTcManifestRequest := &publictcapis.GenerateTargetClusterRegistrationManifestRequest{
		ClusterName: getManifestRequest.GetManifest.ClusterName,
		TenantId:    getManifestRequest.GetManifest.TenantId,
	}

	ctx = WithAccountIDMetaCtx(ctx, tcGrpc.AccountId)

	apiResponse, err := client.GenerateTargetClusterRegistrationManifest(ctx, getTcManifestRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return response, fmt.Errorf("Error when calling `GenerateTargetClusterRegistrationManifest`: %v\n.", err)
	}

	response.TargetCluster.Manifest.Manifest = apiResponse.GetManifest()
	log.Infof("Manifest - [%s]", response.TargetCluster.Manifest.Manifest)

	return response, nil

}
