package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	publictenantapis "github.com/pure-px/apis/public/portworx/platform/tenant/apiv1"
	"google.golang.org/grpc"
)

// getTenantClient updates the header with bearer token and returns the new client
func (Tenantgrpc *PlatformGrpc) getTenantClient() (context.Context, publictenantapis.TenantServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var tenantClient publictenantapis.TenantServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	tenantClient = publictenantapis.NewTenantServiceClient(Tenantgrpc.ApiClientV1)

	return ctx, tenantClient, token, nil
}

func (Tenantgrpc *PlatformGrpc) ListTenants(accountID string) ([]WorkFlowResponse, error) {
	tenantsResponse := []WorkFlowResponse{}
	ctx, client, _, err := Tenantgrpc.getTenantClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	firstPageRequest := &publictenantapis.ListTenantsRequest{
		Pagination: NewPaginationRequest(1, 50),
	}

	ctx = WithAccountIDMetaCtx(ctx, Tenantgrpc.AccountId)

	apiResponse, err := client.ListTenants(ctx, firstPageRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error when calling `AccountServiceListTenants`: %v\n.", err)
	}

	for _, ten := range apiResponse.Tenants {
		log.Infof("accounts - [%v]", ten.Meta.Name)
	}

	err = copier.Copy(&tenantsResponse, apiResponse.Tenants)
	if err != nil {
		return nil, err
	}

	log.Infof("Value of tenants after copy - [%v]", tenantsResponse)
	for _, ten := range tenantsResponse {
		log.Infof("tenants - [%v]", ten.Meta.Name)
	}

	return tenantsResponse, nil
}
