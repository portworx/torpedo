// Package api comprises of all the components and associated CRUD functionality
package api

import (
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	status "net/http"
)

// NamespaceV2 struct
type NamespaceV2 struct {
	ApiClientv2 *pdsv2.APIClient
}

// ListNamespaces return namespaces models in a target cluster.
func (ns *NamespaceV2) ListNamespaces(targetID string) ([]pdsv2.V1Namespace, error) {
	nsClient := ns.ApiClientv2.NamespaceServiceApi
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
