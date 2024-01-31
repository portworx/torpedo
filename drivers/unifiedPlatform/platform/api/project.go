package api

import (
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	status "net/http"
)

// ProjectV2 struct
type ProjectV2 struct {
	ApiClientv2 *pdsv2.APIClient
}

// GetProjectsList return pdsv2 projects models.
func (project *ProjectV2) GetProjectsList() ([]pdsv2.V1Project, error) {
	projectClient := project.ApiClientv2.ProjectServiceApi
	log.Info("Get list of Projects.")
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	projectsModel, res, err := projectClient.ProjectServiceListProjects(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ProjectServiceListProjects`: %v\n.Full HTTP response: %v", err, res)
	}
	return projectsModel.Projects, nil
}

// GetProject return project model.
func (project *ProjectV2) GetProject(projectID string) (*pdsv2.V1Project, error) {
	projectClient := project.ApiClientv2.ProjectServiceApi
	log.Info("Get the project details.")
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	projectModel, res, err := projectClient.ProjectServiceGetProject(ctx, projectID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ProjectServiceGetProject`: %v\n.Full HTTP response: %v", err, res)
	}
	return projectModel, nil
}
