// Package api comprises of all the components and associated CRUD functionality
package api

import (
	"context"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
)

// Components struct comprise of all the component stuct to leverage the usage of the functionality and act as entry point.
type Components struct {
	Account                  *Account
	Tenant                   *Tenant
	Project                  *Project
	AccountRoleBinding       *AccountRoleBinding
	AppConfigTemplate        *AppConfigTemplate
	DataService              *DataService
	DataServiceDeployment    *DataServiceDeployment
	DeploymentTarget         *DeploymentTarget
	Image                    *Image
	Namespace                *Namespace
	ResourceSettingsTemplate *ResourceSettingsTemplate
	StorageSettingsTemplate  *StorageSettingsTemplate
	Version                  *Version
	Backup                   *Backup
	BackupCredential         *BackupCredential
	BackupJob                *BackupJob
	BackupTarget             *BackupTarget
	BackupPolicy             *BackupPolicy
	APIVersion               *PDSVersion
	ServiceAccount           *ServiceAccount
}

// NewComponents create a struct literal that can be leveraged to call all the components functions.
func NewComponents(ctx context.Context, apiClient *pds.APIClient) *Components {
	return &Components{
		Account: &Account{
			Context:   ctx,
			apiClient: apiClient,
		},
		Tenant: &Tenant{
			context:   ctx,
			apiClient: apiClient,
		},
		Project: &Project{
			context:   ctx,
			apiClient: apiClient,
		},
		AccountRoleBinding: &AccountRoleBinding{
			context:   ctx,
			apiClient: apiClient,
		},
		AppConfigTemplate: &AppConfigTemplate{
			context:   ctx,
			apiClient: apiClient,
		},
		DataService: &DataService{
			context:   ctx,
			apiClient: apiClient,
		},
		DataServiceDeployment: &DataServiceDeployment{
			context:   ctx,
			apiClient: apiClient,
		},
		DeploymentTarget: &DeploymentTarget{
			context:   ctx,
			apiClient: apiClient,
		},
		Image: &Image{
			context:   ctx,
			apiClient: apiClient,
		},
		Namespace: &Namespace{
			context:   ctx,
			apiClient: apiClient,
		},
		Version: &Version{
			context:   ctx,
			apiClient: apiClient,
		},
		StorageSettingsTemplate: &StorageSettingsTemplate{
			context:   ctx,
			apiClient: apiClient,
		},
		ResourceSettingsTemplate: &ResourceSettingsTemplate{
			context:   ctx,
			apiClient: apiClient,
		},
		Backup: &Backup{
			context:   ctx,
			apiClient: apiClient,
		},
		BackupCredential: &BackupCredential{
			context:   ctx,
			apiClient: apiClient,
		},
		BackupJob: &BackupJob{
			context:   ctx,
			apiClient: apiClient,
		},
		BackupPolicy: &BackupPolicy{
			context:   ctx,
			apiClient: apiClient,
		},
		BackupTarget: &BackupTarget{
			context:   ctx,
			apiClient: apiClient,
		},
		APIVersion: &PDSVersion{
			context:   ctx,
			apiClient: apiClient,
		},
		ServiceAccount: &ServiceAccount{
			context:   ctx,
			apiClient: apiClient,
		},
	}
}
