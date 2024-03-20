package platformLibs

import "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"

// CreateProject will create a project in given tenant
func CreateProject(projectName string, tenantID string) (*automationModels.V1Project, error) {
	request := automationModels.PlaformProjectRequest{
		Create: automationModels.PlatformCreateProject{
			Project: &automationModels.V1Project{
				Meta: &automationModels.V1Meta{
					Name: &projectName,
					ParentReference: &automationModels.V1Reference{
						Uid: &tenantID,
					},
				},
			},
		},
	}
	project, err := v2Components.Platform.CreateProject(&request, tenantID)
	if err != nil {
		return nil, err
	}
	return &project.Create, nil
}

// GetProjectList will get the list of projects in given tenant
func GetProjectList(pageNumber int, pageSize int) (*automationModels.V1ListProjectsResponse, error) {
	projects, err := v2Components.Platform.GetProjectList(pageNumber, pageSize)
	if err != nil {
		return nil, err
	}
	return &projects.List, nil
}

// GetProject will get the project details of given project id
func GetProject(projectID string) (*automationModels.V1Project, error) {
	request := automationModels.PlaformProjectRequest{
		Get: automationModels.PlatformGetProject{
			ProjectId: projectID,
		},
	}
	project, err := v2Components.Platform.GetProject(&request)
	if err != nil {
		return nil, err
	}
	return &project.Get, nil
}

// DeleteProject will delete the project of given project id
func DeleteProject(projectID string) error {
	request := automationModels.PlaformProjectRequest{
		Delete: automationModels.PlatformDeleteProject{
			ProjectId: projectID,
		},
	}
	err := v2Components.Platform.DeleteProject(&request)
	if err != nil {
		return err
	}
	return nil
}

// Associate will associate the resources to the project
func Associate(clusters []string, namespaces []string, credentials []string, backupLocations []string, templates []string, backupPolicies []string, projectId string) (*automationModels.V1Project, error) {
	request := automationModels.PlaformProjectRequest{
		Associate: automationModels.PlatformAssociateProject{
			ProjectId: projectId,
			ProjectServiceAssociateResourcesBody: &automationModels.ProjectServiceAssociateResourcesBody{
				InfraResource: &automationModels.V1Resources{
					Clusters:        clusters,
					Namespaces:      namespaces,
					Credentials:     credentials,
					BackupLocations: backupLocations,
					Templates:       templates,
					BackupPolicies:  backupPolicies,
				},
			},
		},
	}
	response, err := v2Components.Platform.AssociateToProject(&request)
	if err != nil {
		return &response.Associate, err
	}
	return &response.Associate, nil
}

// Dissociate will dissociate the resources from the project
func Dissociate(clusters []string, namespaces []string, credentials []string, backupLocations []string, templates []string, backupPolicies []string, projectId string) (*automationModels.V1Project, error) {
	request := automationModels.PlaformProjectRequest{
		Associate: automationModels.PlatformAssociateProject{
			ProjectId: projectId,
			ProjectServiceAssociateResourcesBody: &automationModels.ProjectServiceAssociateResourcesBody{
				InfraResource: &automationModels.V1Resources{
					Clusters:        clusters,
					Namespaces:      namespaces,
					Credentials:     credentials,
					BackupLocations: backupLocations,
					Templates:       templates,
					BackupPolicies:  backupPolicies,
				},
			},
		},
	}
	response, err := v2Components.Platform.AssociateToProject(&request)
	if err != nil {
		return &response.Dissociate, err
	}
	return &response.Dissociate, nil
}
