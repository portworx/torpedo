package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// Project struct
type Project struct {
	context   context.Context
	apiClient *pds.APIClient
}

// GetprojectsList func
func (project *Project) GetprojectsList(tenantID string) ([]pds.ModelsProject, error) {
	projectClient := project.apiClient.ProjectsApi
	log.Info("Get list of Projects.")
	projectsModel, res, err := projectClient.ApiTenantsIdProjectsGet(project.context, tenantID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdProjectsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return projectsModel.GetData(), nil
}

// Getproject func
func (project *Project) Getproject(projectID string) (*pds.ModelsProject, error) {
	projectClient := project.apiClient.ProjectsApi
	log.Info("Get the project details.")
	projectModel, res, err := projectClient.ApiProjectsIdGet(project.context, projectID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiProjectsIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return projectModel, nil
}
