package api

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	platformV2 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// ProjectV2 struct
type ProjectV2 struct {
	ApiClientV2 *platformV2.APIClient
	AccountID   string
}

// GetClient updates the header with bearer token and returns the new client
func (project *ProjectV2) GetClient() (context.Context, *platformV2.ProjectServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	project.ApiClientV2.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	project.ApiClientV2.GetConfig().DefaultHeader["px-account-id"] = project.AccountID
	client := project.ApiClientV2.ProjectServiceAPI
	return ctx, client, nil
}

// ListProjects return platformV2 projects models.
func (project *ProjectV2) ListProjects() ([]platformV2.V1Project, error) {
	ctx, projectClient, err := project.GetClient()
	log.Info("Get list of Projects.")
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
func (project *ProjectV2) GetProject(projectID string) (*platformV2.V1Project, error) {
	ctx, projectClient, err := project.GetClient()
	log.Info("Get the project details.")
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	projectModel, res, err := projectClient.ProjectServiceGetProject(ctx, projectID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ProjectServiceGetProject`: %v\n.Full HTTP response: %v", err, res)
	}
	return projectModel, nil
}

// CreateProject return project model.
func (project *ProjectV2) CreateProject(tenantId string) (*platformV2.V1Project, error) {
	ctx, projectClient, err := project.GetClient()
	log.Info("Get the project details.")
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	projectModel, res, err := projectClient.ProjectServiceCreateProject(ctx, tenantId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ProjectServiceCreateProject`: %v\n.Full HTTP response: %v", err, res)
	}
	return projectModel, nil
}

// AssociateResourcesToProject return project model.
func (project *ProjectV2) AssociateResourcesToProject(projectID string) (*platformV2.V1Project, error) {
	ctx, projectClient, err := project.GetClient()
	log.Info("Get the project details.")
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	projectModel, res, err := projectClient.ProjectServiceAssociateResources(ctx, projectID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ProjectServiceAssociateResources`: %v\n.Full HTTP response: %v", err, res)
	}
	return projectModel, nil
}

// DisAssociateResourcesToProject return project model.
func (project *ProjectV2) DisAssociateResourcesToProject(projectID string) (*platformV2.V1Project, error) {
	ctx, projectClient, err := project.GetClient()
	log.Info("Get the project details.")
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	projectModel, res, err := projectClient.ProjectServiceDisassociateResources(ctx, projectID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `ProjectServiceDisassociateResources`: %v\n.Full HTTP response: %v", err, res)
	}
	return projectModel, nil
}

// DeleteProject delete IAM RoleBinding and return status.
func (project *ProjectV2) DeleteProject(projectId string) (*status.Response, error) {
	ctx, projectClient, err := project.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, _ := projectClient.ProjectServiceDeleteProject(ctx, projectId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `IAMServiceDeleteIAM`: %v\n.Full HTTP response: %v", err, res)
	}
	return res, nil
}
