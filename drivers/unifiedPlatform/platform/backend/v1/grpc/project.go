package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	commonapiv1 "github.com/pure-px/apis/public/portworx/common/apiv1"
	publicprojectapis "github.com/pure-px/apis/public/portworx/platform/project/apiv1"
	"google.golang.org/grpc"
)

// GetClient updates the header with bearer token and returns the new client
func (ProjectV1 *PlatformGrpc) getProjectClient() (context.Context, publicprojectapis.ProjectServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var projectClient publicprojectapis.ProjectServiceClient
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	projectClient = publicprojectapis.NewProjectServiceClient(ProjectV1.ApiClientV1)

	return ctx, projectClient, token, nil
}

// GetProjectList returns the list of projects under account
func (ProjectV1 *PlatformGrpc) GetProjectList() ([]WorkFlowResponse, error) {
	projectsResponse := []WorkFlowResponse{}
	ctx, client, _, err := ProjectV1.getProjectClient() //AccountV1.getAccountClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	listProjRequest := publicprojectapis.ListProjectsRequest{
		Pagination: NewPaginationRequest(1, 50),
	}

	apiResponse, err := client.ListProjects(ctx, &listProjRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error while listing the projects: %v\n", err)
	}

	err = copier.Copy(&projectsResponse, apiResponse)
	if err != nil {
		return nil, err
	}

	return projectsResponse, nil
}

// GetProject returns the project details of the project id
func (ProjectV1 *PlatformGrpc) GetProject(projectId string) (
	WorkFlowResponse, error) {

	projectResponse := WorkFlowResponse{}
	ctx, client, _, err := ProjectV1.getProjectClient()
	if err != nil {
		return projectResponse, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	getProjRequest := publicprojectapis.GetProjectRequest{
		ProjectId: projectId,
	}

	apiResponse, err := client.GetProject(ctx, &getProjRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return projectResponse, fmt.Errorf("Error while getting the project: %v\n", err)
	}

	err = copier.Copy(&projectResponse, apiResponse)
	if err != nil {
		return projectResponse, err
	}

	return projectResponse, nil
}

// CreateProject creates a new project under the given tenant
func (ProjectV1 *PlatformGrpc) CreateProject(projectName string, tenantId string) (WorkFlowResponse, error) {
	projectResponse := WorkFlowResponse{}
	ctx, client, _, err := ProjectV1.getProjectClient()
	if err != nil {
		return projectResponse, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	createProjRequest := publicprojectapis.CreateProjectRequest{
		TenantId: tenantId,
		Project: &publicprojectapis.Project{
			Meta: &commonapiv1.Meta{
				Name: projectName,
			},
		},
	}

	apiResponse, err := client.CreateProject(ctx, &createProjRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return projectResponse, fmt.Errorf("Error while creating the project: %v\n", err)
	}

	err = copier.Copy(&projectResponse, apiResponse)
	if err != nil {
		return projectResponse, err
	}

	return projectResponse, nil
}

// DeleteProject deletes the project
func (ProjectV1 *PlatformGrpc) DeleteProject(projectId string) error {

	ctx, client, _, err := ProjectV1.getProjectClient()
	if err != nil {
		return fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	deleteProjRequest := publicprojectapis.DeleteProjectRequest{
		ProjectId: projectId,
	}
	_, err = client.DeleteProject(ctx, &deleteProjRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return fmt.Errorf("Error while deleting the project: %v\n", err)
	}

	return nil
}