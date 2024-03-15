package stworkflows

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
)

type WorkflowProject struct {
	Platform            WorkflowPlatform
	ProjectName         string
	ProjectId           string
	AssociatedResources AssociatedResources
}

type AssociatedResources struct {
	Clusters        []string `copier:"must,nopanic"`
	Namespaces      []string `copier:"must,nopanic"`
	Credentials     []string `copier:"must,nopanic"`
	BackupLocations []string `copier:"must,nopanic"`
	Templates       []string `copier:"must,nopanic"`
	BackupPolicies  []string `copier:"must,nopanic"`
}

// CreateProject will create a project with given project name
func (workflowProject *WorkflowProject) CreateProject() (*WorkflowProject, error) {
	projectDetails, err := platformLibs.CreateProject(workflowProject.ProjectName, workflowProject.Platform.TenantId)
	if err != nil {
		return workflowProject, err
	}

	log.Infof("Created [%s] successfully with ID [%s]", workflowProject.ProjectName, *projectDetails.Meta.Uid)
	workflowProject.ProjectId = *projectDetails.Meta.Uid

	return workflowProject, nil
}

// GetProject will get the project details of given project id
func (workflowProject *WorkflowProject) GetProject() (*automationModels.WorkFlowResponse, error) {
	projectDetails, err := platformLibs.GetProject(workflowProject.ProjectId)
	if err != nil {
		return &automationModels.WorkFlowResponse{}, err
	}

	log.Infof("Project details [%s]", projectDetails)
	return &automationModels.WorkFlowResponse{}, nil
}

// DeleteProject will delete the project of given project id
func (workflowProject *WorkflowProject) DeleteProject() error {
	err := platformLibs.DeleteProject(workflowProject.ProjectId)
	if err != nil {
		return err
	}

	// TODO: This needs to be enabled once DS-8599 is resolved
	//log.Infof("Project Deleted [%s]", workflowProject.ProjectId)
	//err = ValidateProjectDeletion(workflowProject.ProjectId)

	return err
}

// GetProjectList will get the list of projects in given tenant
func (workflowProject *WorkflowProject) GetProjectList() ([]automationModels.WorkFlowResponse, error) {
	projects, err := platformLibs.GetProjectList()
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// Associate will associate a request to project
func (workflowProject *WorkflowProject) Associate(clusters []string, namespaces []string, credentials []string, backupLocations []string, templates []string, backupPolicies []string) error {
	// Associate the resources to the project
	_, err := platformLibs.Associate(clusters, namespaces, credentials, backupLocations, templates, backupPolicies, workflowProject.ProjectId)
	if err != nil {
		return err
	}

	workflowProject.AssociatedResources.Clusters = append(workflowProject.AssociatedResources.Clusters, clusters...)
	workflowProject.AssociatedResources.Namespaces = append(workflowProject.AssociatedResources.Namespaces, namespaces...)
	workflowProject.AssociatedResources.Credentials = append(workflowProject.AssociatedResources.Credentials, credentials...)
	workflowProject.AssociatedResources.BackupLocations = append(workflowProject.AssociatedResources.BackupLocations, backupLocations...)
	workflowProject.AssociatedResources.Templates = append(workflowProject.AssociatedResources.Templates, templates...)
	workflowProject.AssociatedResources.BackupPolicies = append(workflowProject.AssociatedResources.BackupPolicies, backupPolicies...)

	return nil
}

// Dissociate will dissociate a request from project
func (workflowProject *WorkflowProject) Dissociate(clusters []string, namespaces []string, credentials []string, backupLocations []string, templates []string, backupPolicies []string) error {
	// Associate the resources to the project
	_, err := platformLibs.Dissociate(clusters, namespaces, credentials, backupLocations, templates, backupPolicies, workflowProject.ProjectId)
	if err != nil {
		return err
	}

	for _, cluster := range clusters {
		workflowProject.AssociatedResources.Clusters, err = utilities.DeleteElementFromSlice(workflowProject.AssociatedResources.Clusters, cluster)
		if err != nil {
			return err
		}
	}

	for _, namespace := range namespaces {
		workflowProject.AssociatedResources.Namespaces, err = utilities.DeleteElementFromSlice(workflowProject.AssociatedResources.Namespaces, namespace)
		if err != nil {
			return err
		}
	}

	for _, credential := range credentials {
		workflowProject.AssociatedResources.Credentials, err = utilities.DeleteElementFromSlice(workflowProject.AssociatedResources.Credentials, credential)
		if err != nil {
			return err
		}
	}

	for _, backuplocation := range backupLocations {
		workflowProject.AssociatedResources.BackupLocations, err = utilities.DeleteElementFromSlice(workflowProject.AssociatedResources.BackupLocations, backuplocation)
		if err != nil {
			return err
		}
	}

	for _, template := range templates {
		workflowProject.AssociatedResources.Templates, err = utilities.DeleteElementFromSlice(workflowProject.AssociatedResources.Clusters, template)
		if err != nil {
			return err
		}
	}

	for _, backuppolicy := range backupPolicies {
		workflowProject.AssociatedResources.BackupPolicies, err = utilities.DeleteElementFromSlice(workflowProject.AssociatedResources.BackupPolicies, backuppolicy)
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidateProjectDeletion will validate the project deletion
func ValidateProjectDeletion(projectId string) error {
	// Validate the project deletion
	_, err := platformLibs.GetProject(projectId)
	if err == nil {
		return fmt.Errorf("Project [%s] is not deleted", projectId)
	}
	log.Infof("Project [%s] is deleted successfully", projectId)
	return nil
}
