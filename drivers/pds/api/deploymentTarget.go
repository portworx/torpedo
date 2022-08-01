package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type DeploymentTarget struct {
	context   context.Context
	apiClient *pds.APIClient
}

func (dt *DeploymentTarget) ListDeploymentTargetsBelongsToTenant(tenantId string) ([]pds.ModelsDeploymentTarget, error) {
	dtClient := dt.apiClient.DeploymentTargetsApi

	dtModels, res, err := dtClient.ApiTenantsIdDeploymentTargetsGet(dt.context, tenantId).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdDeploymentTargetsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return dtModels.GetData(), nil
}

func (dt *DeploymentTarget) ListDeploymentTargetsBelongsToProject(projectId string) ([]pds.ModelsDeploymentTarget, error) {
	dtClient := dt.apiClient.DeploymentTargetsApi

	dtModels, res, err := dtClient.ApiProjectsIdDeploymentTargetsGet(dt.context, projectId).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiProjectsIdDeploymentTargetsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return dtModels.GetData(), nil
}

func (dt *DeploymentTarget) GetTarget(targetId string) (*pds.ModelsDeploymentTarget, error) {
	dtClient := dt.apiClient.DeploymentTargetsApi
	log.Infof("Get cluster details having uuid - %v", targetId)
	dtModel, res, err := dtClient.ApiDeploymentTargetsIdGet(dt.context, targetId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDeploymentTargetsIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return dtModel, nil
}

func (dt *DeploymentTarget) UpdateTarget(targetId string, name string) (*pds.ModelsDeploymentTarget, error) {
	dtClient := dt.apiClient.DeploymentTargetsApi
	log.Infof("Get cluster details having uuid - %v", targetId)
	upateRequest := pds.ControllersUpdateDeploymentTargetRequest{Name: &name}
	dtModel, res, err := dtClient.ApiDeploymentTargetsIdPut(dt.context, targetId).Body(upateRequest).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDeploymentTargetsIdPut``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return dtModel, nil
}

func (dt *DeploymentTarget) DeleteTarget(targetId string) (*status.Response, error) {
	dtClient := dt.apiClient.DeploymentTargetsApi
	log.Infof("Get cluster details having uuid - %v", targetId)
	res, err := dtClient.ApiDeploymentTargetsIdDelete(dt.context, targetId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiDeploymentTargetsIdDelete``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return res, nil
}
