package platformLibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
)

// CreateProject will create a project in given tenant
func CreateProject(projectName string, tenantID string) (*automationModels.WorkFlowResponse, error) {
	request := automationModels.PlaformProject{
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
	return &project, nil
}

func GetDefaultProjectId(projectName, tenantId string) (string, error) {
	projects, err := GetProjectList(tenantId)
	if err != nil {
		return "", err
	}

	for _, project := range projects {
		if *project.Meta.Name == projectName {
			log.Debugf("Default ProjectId [%s]", *project.Meta.Uid)
			return *project.Meta.Uid, nil
		}
	}

	return "", nil

}

// GetProjectList will get the list of projects in given tenant
func GetProjectList(tenantId string) ([]automationModels.WorkFlowResponse, error) {
	request := automationModels.PlaformProject{
		List: automationModels.PlatformListProject{
			TenantId: tenantId,
		},
	}

	projects, err := v2Components.Platform.GetProjectList(&request)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// GetProject will get the project details of given project id
func GetProject(projectID string) (*automationModels.WorkFlowResponse, error) {
	request := automationModels.PlaformProject{
		Get: automationModels.PlatformGetProject{
			ProjectId: projectID,
		},
	}
	project, err := v2Components.Platform.GetProject(&request)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// DeleteProject will delete the project of given project id
func DeleteProject(projectID string) error {
	request := automationModels.PlaformProject{
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
func Associate(clusters []string, namespaces []string, credentials []string, backupLocations []string, templates []string, backupPolicies []string, projectId string) (*automationModels.WorkFlowResponse, error) {
	request := automationModels.PlaformProject{
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
		return &response, err
	}
	return &response, nil
}

// Dissociate will dissociate the resources from the project
func Dissociate(clusters []string, namespaces []string, credentials []string, backupLocations []string, templates []string, backupPolicies []string, projectId string) (*automationModels.WorkFlowResponse, error) {
	request := automationModels.PlaformProject{
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
		return &response, err
	}
	return &response, nil
}
