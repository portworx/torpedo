package api

import (
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	projectv1 "github.com/pure-px/platform-api-go-client/platform/v1/project"
	status "net/http"
)

// GetProjectList returns the list of projects under account
func (ProjectV1 *PLATFORM_API_V1) GetProjectList(pageNumber int, pageSize int) (*PlaformProjectResponse, error) {
	ctx, client, err := ProjectV1.getProjectClient()
	projectResponse := PlaformProjectResponse{
		List: V1ListProjectsResponse{},
	}

	if err != nil {
		return nil, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	var getRequest projectv1.ApiProjectServiceListProjectsRequest
	getRequest = getRequest.ApiService.ProjectServiceListProjects(ctx)

	projectsList, res, err := client.ProjectServiceListProjectsExecute(getRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ProjectServiceListProjectsExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(projectsList, &projectResponse.List)
	if err != nil {
		return nil, err
	}
	return &projectResponse, nil
}

// GetProject returns the project details of the project id
func (ProjectV1 *PLATFORM_API_V1) GetProject(getProject *PlaformProjectRequest) (*PlaformProjectResponse, error) {
	ctx, client, err := ProjectV1.getProjectClient()
	projectResponse := PlaformProjectResponse{
		Get: V1Project{},
	}

	if err != nil {
		return &projectResponse, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	var getRequest projectv1.ApiProjectServiceGetProjectRequest
	getRequest = getRequest.ApiService.ProjectServiceGetProject(ctx, getProject.Get.ProjectId)

	project, res, err := client.ProjectServiceGetProjectExecute(getRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return &projectResponse, fmt.Errorf("Error when calling `ProjectServiceGetProjectExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(project, &projectResponse.Get)
	return &projectResponse, nil
}

// CreateProject creates a new project under the given tenant
func (ProjectV1 *PLATFORM_API_V1) CreateProject(createProject *PlaformProjectRequest, tenantId string) (*PlaformProjectResponse, error) {
	ctx, client, err := ProjectV1.getProjectClient()
	projectResponse := PlaformProjectResponse{
		Create: V1Project{},
	}

	if err != nil {
		return &projectResponse, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	var projectCreateBody projectv1.ProjectServiceCreateProjectBody
	projectCreateBody = projectv1.ProjectServiceCreateProjectBody{
		Project: &projectv1.V1Project{
			Meta: &projectv1.V1Meta{
				Name: createProject.Create.Project.Meta.Name,
				ParentReference: &projectv1.V1Reference{
					Uid: createProject.Create.Project.Meta.ParentReference.Uid,
				},
			},
		},
	}

	var createRequest projectv1.ApiProjectServiceCreateProjectRequest
	createRequest = createRequest.ApiService.ProjectServiceCreateProject(ctx, tenantId)
	createRequest = createRequest.ProjectServiceCreateProjectBody(projectCreateBody)

	project, res, err := client.ProjectServiceCreateProjectExecute(createRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return &projectResponse, fmt.Errorf("Error when calling `ProjectServiceCreateProjectExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(project, &projectResponse.Create)
	return &projectResponse, nil
}

// DeleteProject deletes the project
func (ProjectV1 *PLATFORM_API_V1) DeleteProject(deleteProject *PlaformProjectRequest) error {
	ctx, client, err := ProjectV1.getProjectClient()
	if err != nil {
		return fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	var deleteRequest projectv1.ApiProjectServiceDeleteProjectRequest
	deleteRequest = deleteRequest.ApiService.ProjectServiceDeleteProject(ctx, deleteProject.Delete.ProjectId)

	_, res, err := client.ProjectServiceDeleteProjectExecute(deleteRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return fmt.Errorf("Error when calling `ProjectServiceDeleteProjectExecute`: %v\n.Full HTTP response: %v", err, res)
	}
	return nil
}

// AssociateToProject associates the given resurces to the project
func (ProjectV1 *PLATFORM_API_V1) AssociateToProject(associateProject *PlaformProjectRequest) (*PlaformProjectResponse, error) {
	ctx, client, err := ProjectV1.getProjectClient()
	response := PlaformProjectResponse{
		Associate: V1Project{},
	}
	if err != nil {
		return &response, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	associateReq := client.ProjectServiceAssociateResources(ctx, associateProject.Associate.ProjectId)
	associateReq = associateReq.ProjectServiceAssociateResourcesBody(projectv1.ProjectServiceAssociateResourcesBody{
		InfraResource: &projectv1.V1Resources{
			Clusters:        associateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.Clusters,
			Namespaces:      associateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.Namespaces,
			Credentials:     associateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.Credentials,
			BackupLocations: associateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.BackupLocations,
			Templates:       associateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.Templates,
			BackupPolicies:  associateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.BackupPolicies,
		},
	})

	log.Infof("Associate request [%v]", associateReq)
	projectDetails, res, err := client.ProjectServiceAssociateResourcesExecute(associateReq)
	log.Infof("Project Details [%v]", projectDetails)
	log.Infof("Project response [%v]", res)
	log.Infof("Project error [%v]", err)
	if err != nil && res.StatusCode != status.StatusOK {
		return &response, fmt.Errorf("Error when calling `ProjectServiceAssociateResourcesExecute`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(projectDetails, &response.Associate)

	return &response, nil
}

// DissociateFromProject dissociates the given resurces from the project
func (ProjectV1 *PLATFORM_API_V1) DissociateFromProject(dissociateProject *PlaformProjectRequest) (*PlaformProjectResponse, error) {
	ctx, client, err := ProjectV1.getProjectClient()
	response := PlaformProjectResponse{
		Dissociate: V1Project{},
	}
	if err != nil {
		return &response, fmt.Errorf("Error while getting updated client with auth header: %v\n", err)
	}
	dissociateReq := client.ProjectServiceDisassociateResources(ctx, dissociateProject.Associate.ProjectId)

	dissociateReq = dissociateReq.ProjectServiceDisassociateResourcesBody(projectv1.ProjectServiceDisassociateResourcesBody{
		InfraResource: &projectv1.V1Resources{
			Clusters:        dissociateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.Clusters,
			Namespaces:      dissociateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.Namespaces,
			Credentials:     dissociateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.Credentials,
			BackupLocations: dissociateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.BackupLocations,
			Templates:       dissociateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.Templates,
			BackupPolicies:  dissociateProject.Associate.ProjectServiceAssociateResourcesBody.InfraResource.BackupPolicies,
		},
	})

	projectDetails, res, err := client.ProjectServiceDisassociateResourcesExecute(dissociateReq)

	if err != nil && res.StatusCode != status.StatusOK {
		return &response, fmt.Errorf("Error when calling `ProjectServiceAssociateResourcesExecute`: %v\n.Full HTTP response: %v", err, res)
	}

	err = utilities.CopyStruct(projectDetails, &response.Associate)

	return &response, nil

}
