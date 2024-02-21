package apiv1

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	platformv1 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// GetNamespaceClient updates the header with bearer token and returns the new client
func (ns *PLATFORM_API_V1) GetNamespaceClient() (context.Context, *platformv1.NamespaceServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	ns.ApiClientV1.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	ns.ApiClientV1.GetConfig().DefaultHeader["px-account-id"] = ns.AccountID
	client := ns.ApiClientV1.NamespaceServiceAPI
	return ctx, client, nil
}

// ListNamespaces return namespaces models in a target cluster.
func (ns *PLATFORM_API_V1) ListNamespaces(targetID *WorkFlowRequest) ([]WorkFlowResponse, error) {
	ctx, nsClient, err := ns.GetNamespaceClient()
	namespaceResponse := []WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	nsModels, res, err := nsClient.NamespaceServiceListNamespaces(ctx, targetID.Id).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `NamespaceServiceListNamespaces`: %v\n.Full HTTP response: %v", err, res)
	}
	err = copier.Copy(&namespaceResponse, nsModels.Namespaces)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of namespace after copy - [%v]", namespaceResponse)
	return namespaceResponse, nil
}
