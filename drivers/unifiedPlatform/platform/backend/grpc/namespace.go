package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	publicnamespaceapis "github.com/pure-px/apis/public/portworx/platform/namespace/apiv1"
	"google.golang.org/grpc"
)

type NamespaceGrpc struct {
	ApiClientV1 *grpc.ClientConn
}

// GetClient updates the header with bearer token and returns the new client
func (NamespaceGrpcV1 *NamespaceGrpc) getNamespaceClient() (context.Context, publicnamespaceapis.NamespaceServiceClient, string, error) {
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

func (NamespaceGrpcV1 *NamespaceGrpc) ListNamespaces() ([]WorkFlowResponse, error) {
	ctx, nsClient, _, err := NamespaceGrpcV1.getNamespaceClient()
	namespaceResponse := []WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	firstPageRequest := &publicnamespaceapis.ListNamespacesRequest{
		Pagination: NewPaginationRequest(1, 50),
	}

	nsResponse, err := nsClient.ListNamespaces(ctx, firstPageRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error when calling `AccountServiceListTenants`: %v\n.", err)
	}

	for _, ns := range nsResponse.Namespaces {
		log.Infof("namespace -  [%v]", ns.Meta.Name)
	}

	copier.Copy(&namespaceResponse, nsResponse.Namespaces)

	log.Infof("Value of namespace after copy - [%v]", nsResponse)
	for _, ten := range namespaceResponse {
		log.Infof("namespace -  [%v]", ten.Meta.Name)
	}

	return namespaceResponse, nil
}
