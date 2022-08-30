// Package api comprises of all the components and associated CRUD functionality
package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// DataServiceDeployment struct
type DataServiceDeployment struct {
	context   context.Context
	apiClient *pds.APIClient
}

// ListDeployments func
func (ds *DataServiceDeployment) ListDeployments(projectID string) ([]pds.ModelsDeployment, error) {
	dsClient := ds.apiClient.DeploymentsApi

	dsModels, res, err := dsClient.ApiProjectsIdDeploymentsGet(ds.context, projectID).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiProjectsIdDeploymentsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return dsModels.GetData(), err
}

// CreateDeployment func
func (ds *DataServiceDeployment) CreateDeployment(projectID string, deploymentTargetID string, dnsZone string, name string, namespaceID string, appConfigID string, imageID string, nodeCount int32, serviceType string, resourceTemplateID string, storageTemplateID string) (*pds.ModelsDeployment, error) {
	dsClient := ds.apiClient.DeploymentsApi
	createRequest := pds.ControllersCreateProjectDeployment{
		// ApplicationConfigurationOverrides:  &appConfigOverride,
		ApplicationConfigurationTemplateId: &appConfigID,
		DeploymentTargetId:                 &deploymentTargetID,
		DnsZone:                            &dnsZone,
		ImageId:                            &imageID,
		// LoadBalancerSourceRanges: lbSourceRange,
		Name:        &name,
		NamespaceId: &namespaceID,
		NodeCount:   &nodeCount,
		// ScheduledBackup:                    &scheduledBackup,
		ResourceSettingsTemplateId: &resourceTemplateID,
		ServiceType:                &serviceType,
		StorageOptionsTemplateId:   &storageTemplateID,
	}
	dsModel, res, err := dsClient.ApiProjectsIdDeploymentsPost(ds.context, projectID).Body(createRequest).Execute()

	if res.StatusCode != status.StatusCreated {
		log.Errorf("Error when calling `ApiProjectsIdDeploymentsPost``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return dsModel, err
}

// CreateDeploymentWithScehduleBackup func
func (ds *DataServiceDeployment) CreateDeploymentWithScehduleBackup(projectID string, deploymentTargetID string, dnsZone string, name string, namespaceID string, appConfigID string, imageID string, nodeCount int32, serviceType string, resourceTemplateID string, storageTemplateID string, backupPolicyID string, backupTargetID string) (*pds.ModelsDeployment, error) {
	dsClient := ds.apiClient.DeploymentsApi
	scheduledBackup := pds.ControllersCreateDeploymentScheduledBackup{
		BackupPolicyId: &backupPolicyID,
		BackupTargetId: &backupTargetID,
	}
	createRequest := pds.ControllersCreateProjectDeployment{
		ApplicationConfigurationTemplateId: &appConfigID,
		DeploymentTargetId:                 &deploymentTargetID,
		DnsZone:                            &dnsZone,
		ImageId:                            &imageID,
		Name:                               &name,
		NamespaceId:                        &namespaceID,
		NodeCount:                          &nodeCount,
		ResourceSettingsTemplateId:         &resourceTemplateID,
		ScheduledBackup:                    &scheduledBackup,
		ServiceType:                        &serviceType,
		StorageOptionsTemplateId:           &storageTemplateID,
	}
	dsModel, res, err := dsClient.ApiProjectsIdDeploymentsPost(ds.context, projectID).Body(createRequest).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiProjectsIdDeploymentsPost``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return dsModel, err
}

//UpdateDeployment func
func (ds *DataServiceDeployment) UpdateDeployment(deploymentID string, appConfigID string, imageID string, nodeCount int32, resourceTemplateID string) (*pds.ModelsDeployment, error) {
	dsClient := ds.apiClient.DeploymentsApi
	createRequest := pds.ControllersUpdateDeploymentRequest{
		// ApplicationConfigurationOverrides:  &appConfigOverride,
		ApplicationConfigurationTemplateId: &appConfigID,
		ImageId:                            &imageID,
		// LoadBalancerSourceRanges: lbSourceRange,
		NodeCount: &nodeCount,
		// ScheduledBackup:                    &scheduledBackup,
		ResourceSettingsTemplateId: &resourceTemplateID,
	}
	dsModel, res, err := dsClient.ApiDeploymentsIdPut(ds.context, deploymentID).Body(createRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDeploymentsIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return dsModel, err
}

// GetDeployment func
func (ds *DataServiceDeployment) GetDeployment(deploymentID string) (*pds.ModelsDeployment, error) {
	dsClient := ds.apiClient.DeploymentsApi
	dsModel, res, err := dsClient.ApiDeploymentsIdGet(ds.context, deploymentID).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDeploymentsIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return dsModel, err
}

// GetDeploymentSatus func
func (ds *DataServiceDeployment) GetDeploymentSatus(deploymentID string) (*pds.ControllersStatusResponse, *status.Response, error) {
	dsClient := ds.apiClient.DeploymentsApi
	dsModel, res, err := dsClient.ApiDeploymentsIdStatusGet(ds.context, deploymentID).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDeploymentsIdStatusGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return dsModel, res, err
}

// GetDeploymentEvents func
func (ds *DataServiceDeployment) GetDeploymentEvents(deploymentID string) (*pds.ControllersEventsResponse, error) {
	dsClient := ds.apiClient.DeploymentsApi
	dsModel, res, err := dsClient.ApiDeploymentsIdEventsGet(ds.context, deploymentID).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDeploymentsIdEventsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return dsModel, err
}

// GetDeploymentCredentials func
func (ds *DataServiceDeployment) GetDeploymentCredentials(deploymentID string) (*pds.DeploymentsCredentials, error) {
	dsClient := ds.apiClient.DeploymentsApi
	dsModel, res, err := dsClient.ApiDeploymentsIdCredentialsGet(ds.context, deploymentID).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDeploymentsIdCredentialsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return dsModel, err
}

// GetConnectionDetails func returns information about the host and connection string
func (ds *DataServiceDeployment) GetConnectionDetails(deploymentID string) (pds.DeploymentsConnectionDetails, map[string]interface{}, error) {
	dsClient := ds.apiClient.DeploymentsApi
	dsModel, res, err := dsClient.ApiDeploymentsIdConnectionInfoGet(ds.context, deploymentID).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDeploymentsIdConnectionInfoGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return dsModel.GetConnectionDetails(), dsModel.GetClusterDetails(), err
}

// DeleteDeployment func
func (ds *DataServiceDeployment) DeleteDeployment(deploymentID string) (*status.Response, error) {
	dsClient := ds.apiClient.DeploymentsApi
	res, err := dsClient.ApiDeploymentsIdDelete(ds.context, deploymentID).Execute()
	// if res.StatusCode != status.StatusOK {
	// 	log.Errorf("Error when calling `ApiDeploymentsIdDelete``: %v\n", err)
	// 	log.Errorf("Full HTTP response: %v\n", res)
	// }
	return res, err
}
