package api

import (
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	status "net/http"
)

// ListNamespaces return namespaces models in a target cluster.
func (ns *PLATFORM_API_V1) ListNamespaces(request *PlatformNamespace) (*PlatformNamespaceResponse, error) {
	ctx, nsClient, err := ns.getNamespaceClient()
	namespaceResponse := PlatformNamespaceResponse{
		List: V1ListNamespacesResponse{
			Namespaces: []V1Namespace{},
		},
	}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}

	req := nsClient.NamespaceServiceListNamespaces(ctx)
	req = req.TenantId(request.List.TenantId)

	if request.List.SortSortBy != "" {
		req = req.SortSortBy(request.List.SortSortBy)
	}
	if request.List.SortSortOrder != "" {
		req = req.SortSortOrder(request.List.SortSortOrder)
	}
	if request.List.PaginationPageNumber != "" {
		req = req.PaginationPageNumber(request.List.PaginationPageNumber)
	}
	if request.List.PaginationPageSize != "" {
		req = req.PaginationPageSize(request.List.PaginationPageSize)
	}

	nsModels, res, err := req.Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `NamespaceServiceListNamespaces`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(nsModels, &namespaceResponse.List)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of namespace after copy - [%v]", namespaceResponse)
	return &namespaceResponse, nil
}

// DeleteNamespace will delete the namespace from control plane
func (ns *PLATFORM_API_V1) DeleteNamespace(request *PlatformNamespace) error {
	ctx, nsClient, err := ns.getNamespaceClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}

	req := nsClient.NamespaceServiceDeleteNamespace(ctx, request.Delete.Id)
	_, res, err := req.Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		return fmt.Errorf("Error when calling `NamespaceServiceDeleteNamespace`: %v\n.Full HTTP response: %v", err, res)
	}
	return nil
}
