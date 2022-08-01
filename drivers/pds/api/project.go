package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type Project struct {
	context   context.Context
	apiClient *pds.APIClient
}

func (project *Project) GetprojectsList(tenantId string) ([]pds.ModelsProject, error) {
	projectClient := project.apiClient.ProjectsApi
	log.Info("Get list of Projects.")
	projectsModel, res, err := projectClient.ApiTenantsIdProjectsGet(project.context, tenantId).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdProjectsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return projectsModel.GetData(), nil
}

func (project *Project) Getproject(projectId string) (*pds.ModelsProject, error) {
	projectClient := project.apiClient.ProjectsApi
	log.Info("Get the project details.")
	projectModel, res, err := projectClient.ApiProjectsIdGet(project.context, projectId).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiProjectsIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return projectModel, nil
}
