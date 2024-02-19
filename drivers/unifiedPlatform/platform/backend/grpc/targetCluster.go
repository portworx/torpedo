package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	publictcapis "github.com/pure-px/apis/public/portworx/platform/targetcluster/apiv1"
	"google.golang.org/grpc"
)

// getTargetClusterClient updates the header with bearer token and returns the new client
func (tcGrpc *PlatformGrpc) getTargetClusterClient() (context.Context, publictcapis.TargetClusterServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var tcClient publictcapis.TargetClusterServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	tcClient = publictcapis.NewTargetClusterServiceClient(tcGrpc.ApiClientV1)

	return ctx, tcClient, token, nil
}

func (tcGrpc *PlatformGrpc) ListTargetClusters() ([]WorkFlowResponse, error) {
	tcResponse := []WorkFlowResponse{}

	ctx, client, _, err := tcGrpc.getTargetClusterClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	firstPageRequest := &publictcapis.ListTargetClustersRequest{
		Pagination: NewPaginationRequest(1, 50),
	}

	apiResponse, err := client.ListTargetClusters(ctx, firstPageRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error when calling `ListTargetClusters`: %v\n.", err)
	}
	copier.Copy(&tcResponse, apiResponse.Clusters)

	return tcResponse, nil

}

func (tcGrpc *PlatformGrpc) GetTarget(getTCRequest *WorkFlowRequest) (*WorkFlowResponse, error) {
	var getRequest *publictcapis.GetTargetClusterRequest
	getTcResponse := WorkFlowResponse{}
	ctx, dtClient, _, err := tcGrpc.getTargetClusterClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	copier.Copy(&getRequest, getTCRequest)

	apiResponse, err := dtClient.GetTargetCluster(ctx, getRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error while getting the target cluster: %v\n", err)
	}

	copier.Copy(&getTcResponse, apiResponse)
	return &getTcResponse, nil
}

func (tcGrpc *PlatformGrpc) PatchTargetCluster(tcRequest *WorkFlowRequest) (*WorkFlowResponse, error) {
	var patchRequest *publictcapis.UpdateTargetClusterRequest
	tcResponse := WorkFlowResponse{}
	ctx, dtClient, _, err := tcGrpc.getTargetClusterClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	copier.Copy(&patchRequest, tcRequest)

	dtModel, err := dtClient.UpdateTargetCluster(ctx, patchRequest)
	if err != nil {
		return nil, fmt.Errorf("Error when calling `UpdateTargetCluster`: %v\n.", err)
	}

	copier.Copy(&tcResponse, dtModel)
	return &tcResponse, nil
}

func (tcGrpc *PlatformGrpc) DeleteTarget(tcRequest *WorkFlowRequest) error {
	var deleteRequest *publictcapis.DeleteTargetClusterRequest
	ctx, dtClient, _, err := tcGrpc.getTargetClusterClient()
	if err != nil {
		return fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	copier.Copy(&deleteRequest, tcRequest)

	_, err = dtClient.DeleteTargetCluster(ctx, deleteRequest)
	if err != nil {
		return fmt.Errorf("Error when calling `DeleteTargetCluster`: %v\n.", err)
	}
	return nil
}
