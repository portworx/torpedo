package apiv1

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	platformv1 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// GetClient updates the header with bearer token and returns the new client
func (project *PLATFORM_API_V1) getProjectClient() (context.Context, *platformv1.ProjectServiceAPIService, error) {
	log.Infof("Creating client from PLATFORM_API_V1 package")
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	project.ApiClientV1.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	project.ApiClientV1.GetConfig().DefaultHeader["px-account-id"] = project.AccountID

	client := project.ApiClientV1.ProjectServiceAPI
	return ctx, client, nil
}

// GetProjectList returns the list of projects under account
func (ProjectV1 *PLATFORM_API_V1) GetProjectList() ([]WorkFlowResponse, error) {
	ctx, client, err := ProjectV1.getProjectClient()
	projectResponse := []WorkFlowResponse{}

	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	var getRequest platformv1.ApiProjectServiceListProjectsRequest
	getRequest = getRequest.ApiService.ProjectServiceListProjects(ctx)

	projectsList, res, err := client.ProjectServiceListProjectsExecute(getRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ProjectServiceListProjectsExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	err = copier.Copy(&projectResponse, projectsList.Projects)
	if err != nil {
		return nil, err
	}
	return projectResponse, nil
}

// GetProject returns the project details of the project id
func (ProjectV1 *PLATFORM_API_V1) GetProject(projectId string) (WorkFlowResponse, error) {
	ctx, client, err := ProjectV1.getProjectClient()
	projectResponse := WorkFlowResponse{}

	if err != nil {
		return projectResponse, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	var getRequest platformv1.ApiProjectServiceGetProjectRequest
	getRequest = getRequest.ApiService.ProjectServiceGetProject(ctx, projectId)

	project, res, err := client.ProjectServiceGetProjectExecute(getRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return projectResponse, fmt.Errorf("Error when calling `ProjectServiceGetProjectExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	err = copier.Copy(&projectResponse, project)
	return projectResponse, nil
}

// CreateProject creates a new project under the given tenant
func (ProjectV1 *PLATFORM_API_V1) CreateProject(projectName string, tenantId string) (WorkFlowResponse, error) {
	ctx, client, err := ProjectV1.getProjectClient()
	projectResponse := WorkFlowResponse{}

	if err != nil {
		return projectResponse, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	var projectCreateBody platformv1.ProjectServiceCreateProjectBody
	projectCreateBody.Project.Meta.Name = &projectName
	projectCreateBody.Project.Meta.ParentReference.Uid = &tenantId

	var createRequest platformv1.ApiProjectServiceCreateProjectRequest
	createRequest = createRequest.ApiService.ProjectServiceCreateProject(ctx, tenantId)
	createRequest = createRequest.ProjectServiceCreateProjectBody(projectCreateBody)

	project, res, err := client.ProjectServiceCreateProjectExecute(createRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return projectResponse, fmt.Errorf("Error when calling `ProjectServiceCreateProjectExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	err = copier.Copy(&projectResponse, project)
	return projectResponse, nil
}

// DeleteProject deletes the project
func (ProjectV1 *PLATFORM_API_V1) DeleteProject(projectId string) error {
	ctx, client, err := ProjectV1.getProjectClient()
	if err != nil {
		return fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	var deleteRequest platformv1.ApiProjectServiceDeleteProjectRequest
	deleteRequest = deleteRequest.ApiService.ProjectServiceDeleteProject(ctx, projectId)

	_, res, err := client.ProjectServiceDeleteProjectExecute(deleteRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return fmt.Errorf("Error when calling `ProjectServiceDeleteProjectExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	return nil
}
