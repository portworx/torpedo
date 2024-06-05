package api

import (
	"context"
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	backupV1 "github.com/pure-px/platform-api-go-client/pds/v1/backup"
	backupConfigV1 "github.com/pure-px/platform-api-go-client/pds/v1/backupconfig"
	catalogV1 "github.com/pure-px/platform-api-go-client/pds/v1/catalog"
	deploymentV1 "github.com/pure-px/platform-api-go-client/pds/v1/dataservicedeployment"
	deploymentsConfigUpdateV1 "github.com/pure-px/platform-api-go-client/pds/v1/dataservicedeploymentconfigupdate"
	restoreV1 "github.com/pure-px/platform-api-go-client/pds/v1/restore"
)

type PDS_API_V1 struct {
	BackupV1APIClient                  *backupV1.APIClient
	BackupConfigV1APIClient            *backupConfigV1.APIClient
	DeploymentV1APIClient              *deploymentV1.APIClient
	DeploymentsConfigUpdateV1APIClient *deploymentsConfigUpdateV1.APIClient
	RestoreV1APIClient                 *restoreV1.APIClient
	CatalogV1APIClient                 *catalogV1.APIClient
	AccountID                          string
}

// getBackupConfigClient updates the header with bearer token and returns the new client
func (backupConf *PDS_API_V1) getBackupConfigClient() (context.Context, *backupConfigV1.BackupConfigServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()

	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	backupConf.BackupConfigV1APIClient.GetConfig().DefaultHeader = utils.GetDefaultHeader(token, backupConf.AccountID)
	client := backupConf.BackupConfigV1APIClient.BackupConfigServiceAPI

	return ctx, client, nil
}

// getBackupClient updates the header with bearer token and returns the new client
func (backup *PDS_API_V1) getBackupClient() (context.Context, *backupV1.BackupServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()

	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	backup.BackupV1APIClient.GetConfig().DefaultHeader = utils.GetDefaultHeader(token, backup.AccountID)
	client := backup.BackupV1APIClient.BackupServiceAPI

	return ctx, client, nil
}

// getRestoreClient updates the header with bearer token and returns the new client
func (restore *PDS_API_V1) getRestoreClient() (context.Context, *restoreV1.RestoreServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()

	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	restore.RestoreV1APIClient.GetConfig().DefaultHeader = utils.GetDefaultHeader(token, restore.AccountID)
	client := restore.RestoreV1APIClient.RestoreServiceAPI

	return ctx, client, nil
}

// GetClient updates the header with bearer token and returns the new client
func (ds *PDS_API_V1) getDeploymentClient() (context.Context, *deploymentV1.DataServiceDeploymentServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	ds.DeploymentV1APIClient.GetConfig().DefaultHeader = utils.GetDefaultHeader(token, ds.AccountID)
	client := ds.DeploymentV1APIClient.DataServiceDeploymentServiceAPI

	return ctx, client, nil
}

// GetDeploymentConfigClient updates the header with bearer token and returns the new client
func (ds *PDS_API_V1) getDeploymentConfigClient() (context.Context, *deploymentsConfigUpdateV1.DataServiceDeploymentConfigUpdateServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	ds.DeploymentsConfigUpdateV1APIClient.GetConfig().DefaultHeader = utils.GetDefaultHeader(token, ds.AccountID)
	client := ds.DeploymentsConfigUpdateV1APIClient.DataServiceDeploymentConfigUpdateServiceAPI

	return ctx, client, nil
}

func (ds *PDS_API_V1) getTemplateDefinitionClient() (context.Context, *catalogV1.TemplateDefinitionServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	ds.CatalogV1APIClient.GetConfig().DefaultHeader = utils.GetDefaultHeader(token, ds.AccountID)
	client := ds.CatalogV1APIClient.TemplateDefinitionServiceAPI

	return ctx, client, nil
}

func (ds *PDS_API_V1) getCatalogClient() (context.Context, *catalogV1.DataServicesServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	ds.CatalogV1APIClient.GetConfig().DefaultHeader = utils.GetDefaultHeader(token, ds.AccountID)

	client := ds.CatalogV1APIClient.DataServicesServiceAPI
	return ctx, client, nil
}

func (ds *PDS_API_V1) getDSVersionsClient() (context.Context, *catalogV1.DataServiceVersionServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	ds.CatalogV1APIClient.GetConfig().DefaultHeader = utils.GetDefaultHeader(token, ds.AccountID)

	client := ds.CatalogV1APIClient.DataServiceVersionServiceAPI
	return ctx, client, nil
}

func (ds *PDS_API_V1) getDSImagesClient() (context.Context, *catalogV1.ImageServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	ds.CatalogV1APIClient.GetConfig().DefaultHeader = utils.GetDefaultHeader(token, ds.AccountID)

	client := ds.CatalogV1APIClient.ImageServiceAPI
	return ctx, client, nil
}
