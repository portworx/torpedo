package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// DeploymentTarget struct
type DeploymentTarget struct {
	context   context.Context
	apiClient *pds.APIClient
}

// ListDeploymentTargetsBelongsToTenant func
func (dt *DeploymentTarget) ListDeploymentTargetsBelongsToTenant(tenantID string) ([]pds.ModelsDeploymentTarget, error) {
	dtClient := dt.apiClient.DeploymentTargetsApi

	dtModels, res, err := dtClient.ApiTenantsIdDeploymentTargetsGet(dt.context, tenantID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdDeploymentTargetsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return dtModels.GetData(), nil
}

// ListDeploymentTargetsBelongsToProject func
func (dt *DeploymentTarget) ListDeploymentTargetsBelongsToProject(projectID string) ([]pds.ModelsDeploymentTarget, error) {
	dtClient := dt.apiClient.DeploymentTargetsApi

	dtModels, res, err := dtClient.ApiProjectsIdDeploymentTargetsGet(dt.context, projectID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiProjectsIdDeploymentTargetsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return dtModels.GetData(), nil
}

// GetTarget func
func (dt *DeploymentTarget) GetTarget(targetID string) (*pds.ModelsDeploymentTarget, error) {
	dtClient := dt.apiClient.DeploymentTargetsApi
	log.Infof("Get cluster details having uuid - %v", targetID)
	dtModel, res, err := dtClient.ApiDeploymentTargetsIdGet(dt.context, targetID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDeploymentTargetsIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return dtModel, nil
}

// UpdateTarget func
func (dt *DeploymentTarget) UpdateTarget(targetID string, name string) (*pds.ModelsDeploymentTarget, error) {
	dtClient := dt.apiClient.DeploymentTargetsApi
	log.Infof("Get cluster details having uuid - %v", targetID)
	upateRequest := pds.ControllersUpdateDeploymentTargetRequest{Name: &name}
	dtModel, res, err := dtClient.ApiDeploymentTargetsIdPut(dt.context, targetID).Body(upateRequest).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDeploymentTargetsIdPut``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return dtModel, nil
}

// DeleteTarget func
func (dt *DeploymentTarget) DeleteTarget(targetID string) (*status.Response, error) {
	dtClient := dt.apiClient.DeploymentTargetsApi
	log.Infof("Get cluster details having uuid - %v", targetID)
	res, err := dtClient.ApiDeploymentTargetsIdDelete(dt.context, targetID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDeploymentTargetsIdDelete``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return res, nil
}
