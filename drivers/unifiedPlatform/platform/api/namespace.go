// Package api contains all the components and associated CRUD functionality
package api

import (
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	platformV2 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// NamespaceV2 struct
type NamespaceV2 struct {
	ApiClientV2 *platformV2.APIClient
}

// ListNamespaces return namespaces models in a target cluster.
func (ns *NamespaceV2) ListNamespaces(targetID string) ([]platformV2.V1Namespace, error) {
	nsClient := ns.ApiClientV2.NamespaceServiceAPI
	ctx, err := GetContext()
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
