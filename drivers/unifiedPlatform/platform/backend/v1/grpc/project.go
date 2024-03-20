package grpc

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/drivers/utilities"
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

	ctx = WithAccountIDMetaCtx(ctx, ProjectV1.AccountId)

	projectClient = publicprojectapis.NewProjectServiceClient(ProjectV1.ApiClientV1)

	return ctx, projectClient, token, nil
}

// GetProjectList returns the list of projects under account
func (ProjectV1 *PlatformGrpc) GetProjectList(project *PlaformProjectRequest) (*PlaformProjectResponse, error) {
	projectsResponse := PlaformProjectResponse{
		List: V1ListProjectsResponse{},
	}
	ctx, client, _, err := ProjectV1.getProjectClient() //AccountV1.getAccountClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	log.Infof("tenant id in grpc %s", project.List.TenantId)

	listProjRequest := publicprojectapis.ListProjectsRequest{
		TenantId:   project.List.TenantId,
		Pagination: NewPaginationRequest(1, 50),
	}

	apiResponse, err := client.ListProjects(ctx, &listProjRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error while listing the projects: %v\n", err)
	}
	err = utilities.CopyStruct(apiResponse.Projects, &projectsResponse.List.Projects)
	if err != nil {
		return nil, err
	}

	return &projectsResponse, nil
}

// GetProject returns the project details of the project id
func (ProjectV1 *PlatformGrpc) GetProject(projectReq *PlaformProjectRequest) (
	*PlaformProjectResponse, error) {

	projectResponse := PlaformProjectResponse{
		Get: V1Project{},
	}
	ctx, client, _, err := ProjectV1.getProjectClient()
	if err != nil {
		return &projectResponse, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	getProjRequest := publicprojectapis.GetProjectRequest{
		ProjectId: projectReq.Get.ProjectId,
	}

	apiResponse, err := client.GetProject(ctx, &getProjRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return &projectResponse, fmt.Errorf("Error while getting the project: %v\n", err)
	}

	err = utilities.CopyStruct(apiResponse, &projectResponse.Get)
	if err != nil {
		return &projectResponse, err
	}

	return &projectResponse, nil
}

// CreateProject creates a new project under the given tenant
func (ProjectV1 *PlatformGrpc) CreateProject(projectReq *PlaformProjectRequest, tenantId string) (*PlaformProjectResponse, error) {
	projectResponse := PlaformProjectResponse{
		Create: V1Project{},
	}
	ctx, client, _, err := ProjectV1.getProjectClient()
	if err != nil {
		return &projectResponse, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	createProjRequest := publicprojectapis.CreateProjectRequest{
		TenantId: tenantId,
		Project: &publicprojectapis.Project{
			Meta: &commonapiv1.Meta{
				Name: *projectReq.Create.Project.Meta.Name,
			},
		},
	}

	apiResponse, err := client.CreateProject(ctx, &createProjRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return &projectResponse, fmt.Errorf("Error while creating the project: %v\n", err)
	}

	err = utilities.CopyStruct(apiResponse, &projectResponse.Create)
	if err != nil {
		return &projectResponse, err
	}

	return &projectResponse, nil
}

// DeleteProject deletes the project
func (ProjectV1 *PlatformGrpc) DeleteProject(projectReq *PlaformProjectRequest) error {

	ctx, client, _, err := ProjectV1.getProjectClient()
	if err != nil {
		return fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	deleteProjRequest := publicprojectapis.DeleteProjectRequest{
		ProjectId: projectReq.Delete.ProjectId,
	}
	_, err = client.DeleteProject(ctx, &deleteProjRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return fmt.Errorf("Error while deleting the project: %v\n", err)
	}

	return nil
}

// AssociateToProject associates the given resurces to the project
func (ProjectV1 *PlatformGrpc) AssociateToProject(associateProject *PlaformProjectRequest) (*PlaformProjectResponse, error) {

	projectsResponse := PlaformProjectResponse{
		Associate: V1Project{},
	}

	ctx, client, _, err := ProjectV1.getProjectClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	request := publicprojectapis.AssociateResourcesRequest{
		ProjectId: associateProject.Associate.ProjectId,
		InfraResource: &publicprojectapis.Resources{
			Clusters:        associateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.Clusters,
			Namespaces:      associateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.Namespaces,
			Credentials:     associateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.Credentials,
			BackupLocations: associateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.BackupLocations,
			Templates:       associateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.Templates,
			BackupPolicies:  associateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.BackupPolicies,
		},
	}

	projectDetails, err := client.AssociateResources(ctx, &request, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error while associating the resources: %v\n", err)
	}

	err = utilities.CopyStruct(projectDetails, &projectsResponse.Associate)

	return &projectsResponse, nil
}

// DissociateFromProject dissociates the given resurces from the project
func (ProjectV1 *PlatformGrpc) DissociateFromProject(dissociateProject *PlaformProjectRequest) (*PlaformProjectResponse, error) {
	projectsResponse := PlaformProjectResponse{
		Dissociate: V1Project{},
	}

	ctx, client, _, err := ProjectV1.getProjectClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}

	request := publicprojectapis.DisassociateResourcesRequest{
		ProjectId: dissociateProject.Associate.ProjectId,
		InfraResource: &publicprojectapis.Resources{
			Clusters:        dissociateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.Clusters,
			Namespaces:      dissociateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.Namespaces,
			Credentials:     dissociateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.Credentials,
			BackupLocations: dissociateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.BackupLocations,
			Templates:       dissociateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.Templates,
			BackupPolicies:  dissociateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.BackupPolicies,
		},
	}

	projectDetails, err := client.DisassociateResources(ctx, &request, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error while dissociating the resources: %v\n", err)
	}

	err = utilities.CopyStruct(projectDetails, &projectsResponse.Dissociate)

	return &projectsResponse, nil

}
