// Package api contains all the components and associated CRUD functionality
package api

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	platformV2 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// NamespaceV2 struct
type NamespaceV2 struct {
	ApiClientV2 *platformV2.APIClient
	AccountID   string
}

// GetClient updates the header with bearer token and returns the new client
func (ns *NamespaceV2) GetClient() (context.Context, *platformV2.NamespaceServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	ns.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	ns.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = ns.AccountID
	client := ns.ApiClientV2.NamespaceServiceAPI
	return ctx, client, nil
}

// ListNamespaces return namespaces models in a target cluster.
func (ns *NamespaceV2) ListNamespaces(targetID string) ([]platformV2.V1Namespace, error) {
	ctx, nsClient, err := ns.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	nsModels, res, err := nsClient.NamespaceServiceListNamespaces(ctx, targetID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `NamespaceServiceListNamespaces`: %v\n.Full HTTP response: %v", err, res)
	}
	return nsModels.Namespaces, err
}

// CreateNamespace return newly created namespaces model in the target cluster. Not Available

// GetNamespace return namespaces model in the target cluster. Not Available

// DeleteNamespace delete the namespace and return status. Not Available
