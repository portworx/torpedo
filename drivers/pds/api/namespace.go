// Package api comprises of all the components and associated CRUD functionality
package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type Namespace struct {
	context   context.Context
	apiClient *pds.APIClient
}

func (ns *Namespace) ListNamespaces(targetId string) ([]pds.ModelsNamespace, error) {
	nsClient := ns.apiClient.NamespacesApi
	nsModels, res, err := nsClient.ApiDeploymentTargetsIdNamespacesGet(ns.context, targetId).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDeploymentTargetsIdNamespacesGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return nsModels.GetData(), err
}

func (ns *Namespace) CreateNamespace(targetId string, name string) (*pds.ModelsNamespace, error) {
	nsClient := ns.apiClient.NamespacesApi

	createRequest := pds.ControllersCreateNamespace{Name: &name}
	nsModel, res, err := nsClient.ApiDeploymentTargetsIdNamespacesPost(ns.context, targetId).Body(createRequest).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDeploymentTargetsIdNamespacesPost``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return nsModel, nil
}

func (ns *Namespace) GetNamespace(namespaceId string) (*pds.ModelsNamespace, error) {
	nsClient := ns.apiClient.NamespacesApi

	nsModel, res, err := nsClient.ApiNamespacesIdGet(ns.context, namespaceId).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiNamespacesIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return nsModel, nil
}

func (ns *Namespace) DeleteNamespace(namespaceId string) (*status.Response, error) {
	nsClient := ns.apiClient.NamespacesApi

	res, err := nsClient.ApiNamespacesIdDelete(ns.context, namespaceId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiNamespacesIdDelete``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return res, nil
}
