package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/drivers/utilities"
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

	ctx = WithAccountIDMetaCtx(ctx, tcGrpc.AccountId)

	tcClient = publictcapis.NewTargetClusterServiceClient(tcGrpc.ApiClientV1)

	return ctx, tcClient, token, nil
}

func (tcGrpc *PlatformGrpc) ListTargetClusters(tcRequest *PlatformTargetClusterRequest) (*PlatformTargetClusterResponse, error) {
	//tcResponse := PlatformTargetClusterResponse{
	//	ListTargetClusters: V1ListTargetClustersResponse{},
	//}
	ctx, client, _, err := tcGrpc.getTargetClusterClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	firstPageRequest := &publictcapis.ListTargetClustersRequest{
		TenantId: tcRequest.ListTargetClusters.TenantId,
	}

	log.Infof("Request - [%+v]", firstPageRequest)
	log.Infof("Ctx - [%+v]", ctx)

	apiResponse, err := client.ListTargetClusters(ctx, firstPageRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error when calling `ListTargetClusters`: %v\n.", err)
	}
	tcResponse := ConvertListToPlatformResponse(apiResponse)
	if err != nil {
		return nil, err
	}

	return tcResponse, nil

}

func (tcGrpc *PlatformGrpc) GetTargetCluster(getTCRequest *PlatformTargetClusterRequest) (*PlatformTargetClusterResponse, error) {
	tcResponse := PlatformTargetClusterResponse{
		GetTargetCluster: V1TargetCluster{},
	}
	ctx, dtClient, _, err := tcGrpc.getTargetClusterClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	getRequest := &publictcapis.GetTargetClusterRequest{
		Id: getTCRequest.GetTargetCluster.Id,
	}
	if err != nil {
		return nil, err
	}
	apiResponse, err := dtClient.GetTargetCluster(ctx, getRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error while getting the target cluster: %v\n", err)
	}
	err = utilities.CopyStruct(apiResponse, &tcResponse.GetTargetCluster)
	if err != nil {
		return nil, err
	}
	log.Infof("Response - [%+v]", tcResponse.GetTargetCluster)
	return &tcResponse, nil
}

//func (tcGrpc *PlatformGrpc) PatchTargetCluster(tcRequest *PlatformTargetClusterRequest) (*PlatformTargetClusterResponse, error) {
//	var patchRequest *publictcapis.UpdateTargetClusterRequest
//	tcResponse := WorkFlowResponse{}
//	ctx, dtClient, _, err := tcGrpc.getTargetClusterClient()
//	if err != nil {
//		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
//	}
//	err = copier.Copy(&patchRequest, tcRequest)
//	if err != nil {
//		return nil, err
//	}
//	dtModel, err := dtClient.UpdateTargetCluster(ctx, patchRequest)
//	if err != nil {
//		return nil, fmt.Errorf("Error when calling `UpdateTargetCluster`: %v\n.", err)
//	}
//	err = copier.Copy(&tcResponse, dtModel)
//	if err != nil {
//		return nil, err
//	}
//	return &tcResponse, nil
//}

func (tcGrpc *PlatformGrpc) DeleteTargetCluster(tcRequest *PlatformTargetClusterRequest) error {
	var deleteRequest *publictcapis.DeleteTargetClusterRequest
	ctx, dtClient, _, err := tcGrpc.getTargetClusterClient()
	if err != nil {
		return fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	err = copier.Copy(&deleteRequest, tcRequest)
	if err != nil {
		return err
	}
	_, err = dtClient.DeleteTargetCluster(ctx, deleteRequest)
	if err != nil {
		return fmt.Errorf("Error when calling `DeleteTargetCluster`: %v\n.", err)
	}
	return nil
}

func ConvertListToPlatformResponse(tcResponse *publictcapis.ListTargetClustersResponse) *PlatformTargetClusterResponse {
	response := &PlatformTargetClusterResponse{
		ListTargetClusters: V1ListTargetClustersResponse{
			Clusters: []V1TargetCluster{},
		},
	}

	for _, eachCluster := range tcResponse.Clusters {
		cluster := V1TargetCluster{
			Meta: &V1Meta{},
			Status: &PlatformTargetClusterv1Status{
				Phase: V1TargetClusterPhasePhase(eachCluster.Status.Phase),
			},
		}
		// Copying all cluster related values manually
		// Any value which is not copied will be nil
		cluster.Meta.Uid = &eachCluster.Meta.Uid
		cluster.Meta.Name = &eachCluster.Meta.Name
		cluster.Meta.Description = &eachCluster.Meta.Description
		cluster.Meta.Labels = &eachCluster.Meta.Labels
		cluster.Meta.Annotations = &eachCluster.Meta.Annotations

		response.ListTargetClusters.Clusters = append(response.ListTargetClusters.Clusters, cluster)
	}

	return response
}
