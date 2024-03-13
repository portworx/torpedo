package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	publicnamespaceapis "github.com/pure-px/apis/public/portworx/platform/namespace/apiv1"
	"google.golang.org/grpc"
)

// GetClient updates the header with bearer token and returns the new client
func (NamespaceGrpcV1 *PlatformGrpc) getNamespaceClient() (context.Context, publicnamespaceapis.NamespaceServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var accountClient publicnamespaceapis.NamespaceServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	accountClient = publicnamespaceapis.NewNamespaceServiceClient(NamespaceGrpcV1.ApiClientV1)

	return ctx, accountClient, token, nil
}

func (NamespaceGrpcV1 *PlatformGrpc) ListNamespaces(request *PlatformNamespace) (*PlatformNamespaceResponse, error) {
	ctx, nsClient, _, err := NamespaceGrpcV1.getNamespaceClient()

	namespaceResponse := PlatformNamespaceResponse{
		List: V1ListNamespacesResponse{
			Namespaces: []V1Namespace{},
		},
	}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var firstPageRequest *publicnamespaceapis.ListNamespacesRequest
	err = copier.Copy(&firstPageRequest, request)
	if err != nil {
		return nil, err
	}
	nsResponse, err := nsClient.ListNamespaces(ctx, firstPageRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error when calling `AccountServiceListTenants`: %v\n.", err)
	}

	for _, ns := range nsResponse.Namespaces {
		log.Infof("namespace -  [%v]", ns.Meta.Name)
	}

	err = copier.Copy(&namespaceResponse.List, nsResponse)
	if err != nil {
		return nil, err
	}

	log.Infof("Value of namespace after copy - [%v]", nsResponse)
	for _, ten := range namespaceResponse.List.Namespaces {
		log.Infof("namespace -  [%v]", ten.Meta.Name)
	}

	return &namespaceResponse, nil
}
