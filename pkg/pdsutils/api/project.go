// Package api comprises of all the components and associated CRUD functionality
package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// Project struct comprise of Context , PDS api client.
type Project struct {
	Context   context.Context
	apiClient *pds.APIClient
}

// GetprojectsList function return Project objects.
func (project *Project) GetprojectsList(tenantID string) ([]pds.ModelsProject, error) {
	projectClient := project.apiClient.ProjectsApi
	log.Info("Get list of Projects.")
	projectsModel, res, err := projectClient.ApiTenantsIdProjectsGet(project.Context, tenantID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsGet``: %v\n", err)
		log.Debugf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return projectsModel.GetData(), nil
}

// Getproject return Project object.
func (project *Project) Getproject(projectID string) (*pds.ModelsProject, error) {
	projectClient := project.apiClient.ProjectsApi
	log.Info("Get the project details.")
	projectModel, res, err := projectClient.ApiProjectsIdGet(project.Context, projectID).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsGet``: %v\n", err)
		log.Debugf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return projectModel, nil
}
