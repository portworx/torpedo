// Package api comprises of all the components and associated CRUD functionality
package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// Namespace struct
type Namespace struct {
	context   context.Context
	apiClient *pds.APIClient
}

// ListNamespaces func
func (ns *Namespace) ListNamespaces(targetID string) ([]pds.ModelsNamespace, error) {
	nsClient := ns.apiClient.NamespacesApi
	nsModels, res, err := nsClient.ApiDeploymentTargetsIdNamespacesGet(ns.context, targetID).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDeploymentTargetsIdNamespacesGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return nsModels.GetData(), err
}

// CreateNamespace func
func (ns *Namespace) CreateNamespace(targetID string, name string) (*pds.ModelsNamespace, error) {
	nsClient := ns.apiClient.NamespacesApi

	createRequest := pds.ControllersCreateNamespace{Name: &name}
	nsModel, res, err := nsClient.ApiDeploymentTargetsIdNamespacesPost(ns.context, targetID).Body(createRequest).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDeploymentTargetsIdNamespacesPost``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return nsModel, nil
}

// GetNamespace func
func (ns *Namespace) GetNamespace(namespaceID string) (*pds.ModelsNamespace, error) {
	nsClient := ns.apiClient.NamespacesApi

	nsModel, res, err := nsClient.ApiNamespacesIdGet(ns.context, namespaceID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiNamespacesIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return nsModel, nil
}

// DeleteNamespace func
func (ns *Namespace) DeleteNamespace(namespaceID string) (*status.Response, error) {
	nsClient := ns.apiClient.NamespacesApi

	res, err := nsClient.ApiNamespacesIdDelete(ns.context, namespaceID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiNamespacesIdDelete``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return res, nil
}
