package api

import (
	"context"
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	status "net/http"
)

// DeploymentV2 struct
type DeploymentV2 struct {
	ApiClientV2 *pdsv2.APIClient
	AccountID   string
}

// GetClient updates the header with bearer token and returns the new client
func (ds *DeploymentV2) GetClient() (context.Context, *pdsv2.DeploymentServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	ds.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	ds.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = ds.AccountID
	client := ds.ApiClientV2.DeploymentServiceAPI

	return ctx, client, nil
}

// ListDeployments return deployments models for a given project.
func (ds *DeploymentV2) ListDeployments() ([]pdsv2.V1Deployment, error) {
	ctx, dsClient, err := ds.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dsModels, res, err := dsClient.DeploymentServiceListDeployments(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentServiceListDeployments`: %v\n.Full HTTP response: %v", err, res)
	}
	return dsModels.Deployments, nil
}

// CreateDeployment return newly created deployment model.
func (ds *DeploymentV2) CreateDeployment(namespaceID string) (*pdsv2.V1Deployment, error) {
	dsClient := ds.ApiClientV2.DeploymentServiceAPI
	context, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dsModel, res, err := dsClient.DeploymentServiceCreateDeployment(context, namespaceID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentServiceCreateDeployment`: %v\n.Full HTTP response: %v", err, res)
	}
	return dsModel, err
}

//CreateDeploymentWithRbac

// CreateDeploymentWithScheduleBackup return newly created deployment model with schedule backup enabled.

// GetDeployment return deployment model.

func (ds *DeploymentV2) GetDeployment(deploymentID string) (*pdsv2.V1Deployment, error) {
	ctx, dsClient, err := ds.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dsModel, res, err := dsClient.DeploymentServiceGetDeployment(ctx, deploymentID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentServiceGetDeployment`: %v\n.Full HTTP response: %v", err, res)
	}
	return dsModel, err
}

// GetDeploymentStatus return deployment status.
func (ds *DeploymentV2) GetDeploymentStatus(deploymentID string) (*pdsv2.Deploymentv1Status, *status.Response, error) {
	ctx, dsClient, err := ds.GetClient()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dsModel, res, err := dsClient.DeploymentServiceGetDeployment(ctx, deploymentID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, nil, fmt.Errorf("Error when calling `DeploymentServiceGetDeployment`: %v\n.Full HTTP response: %v", err, res)
	}
	return dsModel.Status, res, err
}

// GetDeploymentCredentials return deployment credentials.
func (ds *DeploymentV2) GetDeploymentCredentials(deploymentID string) (*pdsv2.V1DeploymentCredentials, error) {
	ctx, dsClient, err := ds.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dsModel, res, err := dsClient.DeploymentServiceGetDeploymentCredentials(ctx, deploymentID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentServiceGetDeploymentCredentials`: %v\n.Full HTTP response: %v", err, res)
	}
	return dsModel, err
}

// UpdateDeploymentWithTls updates the deployment with TLS enabled/disabled

// UpdateDeployment func
func (ds *DeploymentV2) UpdateDeployment() (*pdsv2.V1Deployment, error) {
	dsClient := ds.ApiClientV2.DeploymentServiceAPI
	context, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	dsModel, res, err := dsClient.DeploymentServiceUpdateDeployment(context).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentServiceUpdateDeployment`: %v\n.Full HTTP response: %v", err, res)
	}
	return dsModel, err
}

// GetConnectionDetails return connection details for the given deployment.

// DeleteDeployment delete deployment and return status.
func (ds *DeploymentV2) DeleteDeployment(deploymentID string) (*status.Response, error) {
	ctx, dsClient, err := ds.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, err := dsClient.DeploymentServiceDeleteDeployment(ctx, deploymentID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `DeploymentServiceDeleteDeployment`: %v\n.Full HTTP response: %v", err, res)
	}
	return res, err
}
