package api

import (
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
	status "net/http"
)

// ListNamespaces return namespaces models in a target cluster.
func (ns *PLATFORM_API_V1) ListNamespaces(request *WorkFlowRequest) ([]WorkFlowResponse, error) {
	ctx, nsClient, err := ns.getNamespaceClient()
	namespaceResponse := []WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	nsModels, res, err := nsClient.NamespaceServiceListNamespaces(ctx).Execute()
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
