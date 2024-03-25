package api

import (
	"context"
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	backupV1 "github.com/pure-px/platform-api-go-client/pds/v1/backup"
	backupConfigV1 "github.com/pure-px/platform-api-go-client/pds/v1/backupconfig"
	catalogV1 "github.com/pure-px/platform-api-go-client/pds/v1/catalog"
	deploymentV1 "github.com/pure-px/platform-api-go-client/pds/v1/deployment"
	deploymentsConfigUpdateV1 "github.com/pure-px/platform-api-go-client/pds/v1/deploymentconfigupdate"
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
	backupConf.BackupV1APIClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	backupConf.BackupV1APIClient.GetConfig().DefaultHeader["px-account-id"] = backupConf.AccountID
	client := backupConf.BackupConfigV1APIClient.BackupConfigServiceAPI

	return ctx, client, nil
}

// GetClient updates the header with bearer token and returns the new client
func (ds *PDS_API_V1) getDeploymentClient() (context.Context, *deploymentV1.DeploymentServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	ds.DeploymentV1APIClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	ds.DeploymentV1APIClient.GetConfig().DefaultHeader["px-account-id"] = ds.AccountID
	client := ds.DeploymentV1APIClient.DeploymentServiceAPI

	return ctx, client, nil
}

// GetDeploymentConfigClient updates the header with bearer token and returns the new client
func (ds *PDS_API_V1) getDeploymentConfigClient() (context.Context, *deploymentsConfigUpdateV1.DeploymentConfigUpdateServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	ds.DeploymentsConfigUpdateV1APIClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	ds.DeploymentsConfigUpdateV1APIClient.GetConfig().DefaultHeader["px-account-id"] = ds.AccountID
	client := ds.DeploymentsConfigUpdateV1APIClient.DeploymentConfigUpdateServiceAPI

	return ctx, client, nil
}

func (ds *PDS_API_V1) getTemplateDefinitionClient() (context.Context, *catalogV1.TemplateDefinitionServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	ds.CatalogV1APIClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	ds.CatalogV1APIClient.GetConfig().DefaultHeader["px-account-id"] = ds.AccountID
	client := ds.CatalogV1APIClient.TemplateDefinitionServiceAPI

	return ctx, client, nil
}

func (ds *PDS_API_V1) getCatalogClient() (context.Context, *catalogV1.DataServicesServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	ds.CatalogV1APIClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	ds.CatalogV1APIClient.GetConfig().DefaultHeader["px-account-id"] = ds.AccountID

	client := ds.CatalogV1APIClient.DataServicesServiceAPI
	return ctx, client, nil
}

func (ds *PDS_API_V1) getDSVersionsClient() (context.Context, *catalogV1.DataServiceVersionServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	ds.CatalogV1APIClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	ds.CatalogV1APIClient.GetConfig().DefaultHeader["px-account-id"] = ds.AccountID

	client := ds.CatalogV1APIClient.DataServiceVersionServiceAPI
	return ctx, client, nil
}

func (ds *PDS_API_V1) getDSImagesClient() (context.Context, *catalogV1.ImageServiceAPIService, error) {
	ctx, token, err := utils.GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	ds.CatalogV1APIClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	ds.CatalogV1APIClient.GetConfig().DefaultHeader["px-account-id"] = ds.AccountID

	client := ds.CatalogV1APIClient.ImageServiceAPI
	return ctx, client, nil
}
