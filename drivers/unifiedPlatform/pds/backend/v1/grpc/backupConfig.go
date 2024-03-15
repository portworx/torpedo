package grpc

import (
	"context"
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	publicBackupConfigapis "github.com/pure-px/apis/public/portworx/pds/backupconfig/apiv1"
	"google.golang.org/grpc"
)

func (backupConf *PdsGrpc) getBackupConfigClient() (context.Context, publicBackupConfigapis.BackupConfigServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var depClient publicBackupConfigapis.BackupConfigServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	credentials = &Credentials{
		Token: token,
	}
	depClient = publicBackupConfigapis.NewBackupConfigServiceClient(backupConf.ApiClientV2)
	return ctx, depClient, token, nil
}

// CreateBackupConfig will create backup config for a given deployment
func (backupConf *PdsGrpc) CreateBackupConfig(createBackupConfigRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	// log.Infof("Backup Create - [%+v]", createBackupConfigRequest.BackupConfig.Create)

	response := &automationModels.WorkFlowResponse{}

	createRequest := &publicBackupConfigapis.CreateBackupConfigRequest{}
	// log.Infof("Backup Create Request - [%v], Backup Config - [%v]", createRequest, createRequest.BackupConfig)
	err := utilities.CopyStruct(createBackupConfigRequest.BackupConfig.Create, createRequest)
	if err != nil {
		return response, err
	}
	// log.Infof("Backup Create Request - [%v], Backup Config - [%v]", createRequest, createRequest.BackupConfig)

	ctx, client, _, err := backupConf.getBackupConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting grpc client: %v\n", err)
	}

	ctx = WithAccountIDMetaCtx(ctx, backupConf.AccountId)

	apiResponse, err := client.CreateBackupConfig(ctx, createRequest, grpc.PerRPCCredentials(credentials))
	log.Infof("api response [+%v]", apiResponse)
	if err != nil {
		return nil, fmt.Errorf("Error while creating the backupConfig: %v\n", err)
	}

	err = utilities.CopyStruct(apiResponse, response)
	if err != nil {
		return response, err
	}

	return response, nil
}

// UpdateBackupConfig will update backup config for a given deployment
func (backupConf *PdsGrpc) UpdateBackupConfig(updateBackupConfigRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {

	// log.Infof("Backup Update - [%+v]", updateBackupConfigRequest.BackupConfig.Update)

	response := &automationModels.WorkFlowResponse{}

	updateRequest := &publicBackupConfigapis.UpdateBackupConfigRequest{}
	// log.Infof("Backup Update - [%v], Backup Config - [%v]", updateRequest, updateRequest.BackupConfig)
	err := utilities.CopyStruct(updateBackupConfigRequest.BackupConfig.Update, updateRequest)
	if err != nil {
		return response, err
	}
	// log.Infof("Backup Update - [%v], Backup Config - [%v]", updateRequest, updateRequest.BackupConfig)

	ctx, client, _, err := backupConf.getBackupConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting grpc client: %v\n", err)
	}

	ctx = WithAccountIDMetaCtx(ctx, backupConf.AccountId)

	apiResponse, err := client.UpdateBackupConfig(ctx, updateRequest, grpc.PerRPCCredentials(credentials))
	log.Infof("api response [+%v]", apiResponse)
	if err != nil {
		return nil, fmt.Errorf("Error while updating the backupConfig: %v\n", err)
	}

	err = utilities.CopyStruct(apiResponse, response)
	if err != nil {
		return response, err
	}

	return response, nil

}

// GetBackupConfig will fetch backup config for a given deployment
func (backupConf *PdsGrpc) GetBackupConfig(getBackupConfigRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	// log.Infof("Backup Get - [%+v]", getBackupConfigRequest.BackupConfig.Get)

	response := &automationModels.WorkFlowResponse{}

	getRequest := &publicBackupConfigapis.GetBackupConfigRequest{}
	// log.Infof("Backup Get - [%v]", getRequest)
	err := utilities.CopyStruct(getBackupConfigRequest.BackupConfig.Get, getRequest)
	if err != nil {
		return response, err
	}
	// log.Infof("Backup Get - [%v]", getRequest)

	ctx, client, _, err := backupConf.getBackupConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting grpc client: %v\n", err)
	}

	ctx = WithAccountIDMetaCtx(ctx, backupConf.AccountId)

	apiResponse, err := client.GetBackupConfig(ctx, getRequest, grpc.PerRPCCredentials(credentials))
	log.Infof("api response [+%v]", apiResponse)
	if err != nil {
		return nil, fmt.Errorf("Error while fetching the backupConfig: %v\n", err)
	}

	err = utilities.CopyStruct(apiResponse, response)
	if err != nil {
		return response, err
	}

	return response, nil

}

// DeleteBackupConfig will delete backup config for a given deployment
func (backupConf *PdsGrpc) DeleteBackupConfig(deleteBackupConfigRequest *automationModels.WorkFlowRequest) (*automationModels.WorkFlowResponse, error) {
	// log.Infof("Backup Delete - [%+v]", deleteBackupConfigRequest.BackupConfig.Delete)

	response := &automationModels.WorkFlowResponse{}

	deleteRequest := &publicBackupConfigapis.DeleteBackupConfigRequest{}
	// log.Infof("Backup Delete - [%v]", deleteRequest)
	err := utilities.CopyStruct(deleteBackupConfigRequest.BackupConfig.Delete, deleteRequest)
	if err != nil {
		return response, err
	}
	// log.Infof("Backup Delete - [%v]", deleteRequest)

	ctx, client, _, err := backupConf.getBackupConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting grpc client: %v\n", err)
	}

	ctx = WithAccountIDMetaCtx(ctx, backupConf.AccountId)

	apiResponse, err := client.DeleteBackupConfig(ctx, deleteRequest, grpc.PerRPCCredentials(credentials))
	log.Infof("api response [+%v]", apiResponse)
	if err != nil {
		return nil, fmt.Errorf("Error while deleting the backupConfig: %v\n", err)
	}

	err = utilities.CopyStruct(apiResponse, response)
	if err != nil {
		return response, err
	}

	return response, nil
}

// ListBackupConfig will list backup config for a given deployment
func (backupConf *PdsGrpc) ListBackupConfig(listBackupConfigRequest *automationModels.WorkFlowRequest) ([]automationModels.WorkFlowResponse, error) {
	// log.Infof("Backup List - [%+v]", listBackupConfigRequest.BackupConfig.List)

	response := []automationModels.WorkFlowResponse{}

	listRequest := &publicBackupConfigapis.ListBackupConfigsRequest{}
	// log.Infof("Backup List - [%v]", listRequest)
	err := utilities.CopyStruct(listBackupConfigRequest.BackupConfig.List, listRequest)
	if err != nil {
		return response, err
	}
	// log.Infof("Backup List - [%v]", listRequest)

	ctx, client, _, err := backupConf.getBackupConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting grpc client: %v\n", err)
	}

	ctx = WithAccountIDMetaCtx(ctx, backupConf.AccountId)

	apiResponse, err := client.ListBackupConfigs(ctx, listRequest, grpc.PerRPCCredentials(credentials))
	log.Infof("api response [+%v]", apiResponse)
	if err != nil {
		return nil, fmt.Errorf("Error while listing the backupConfig: %v\n", err)
	}

	err = utilities.CopyStruct(apiResponse, response)
	if err != nil {
		return response, err
	}

	return response, nil
}
