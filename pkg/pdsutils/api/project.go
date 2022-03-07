// Package api comprises of all the components and associated CRUD functionality
package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

/*
	Project struct consists the context and apiclient.
	It can be used to utilize CRUD functionality with respect to Project API.
*/
type Project struct {
	Context   context.Context
	apiClient *pds.APIClient
}

/*
	Get the List of all the Pojects.
	@param tenantId string - Account UUID.
	@return []pds.ModelsProject, error
*/
func (project *Project) GetprojectsList(tenantId string) ([]pds.ModelsProject, error) {
	projectClient := project.apiClient.ProjectsApi
	log.Info("Get list of Projects.")
	projectsModel, res, err := projectClient.ApiTenantsIdProjectsGet(project.Context, tenantId).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsGet``: %v\n", err)
		log.Error("Full HTTP response: %v\n", res)
		return nil, err
	}
	return projectsModel.GetData(), nil
}

/*
	Get the Project details.
	@param projectId string - Project UUID.
	@return *pds.ModelsProject, error
*/
func (project *Project) Getproject(projectId string) (*pds.ModelsProject, error) {
	projectClient := project.apiClient.ProjectsApi
	log.Info("Get the project details.")
	projectModel, res, err := projectClient.ApiProjectsIdGet(project.Context, projectId).Execute()

	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsGet``: %v\n", err)
		log.Error("Full HTTP response: %v\n", res)
		return nil, err
	}
	return projectModel, nil
}
