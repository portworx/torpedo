package grpc

import (
	"context"
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
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
func (backupConf *PdsGrpc) CreateBackupConfig(createBackupConfigRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Infof("Backup Config - [%+v]", createBackupConfigRequest.BackupConfig.GRPC.Create.V1BackupConfig)

	response := &apiStructs.WorkFlowResponse{}

	backupConfig := &publicBackupConfigapis.BackupConfig{}
	log.Infof("Backup Config - [%v]", backupConfig)
	err := utilities.CopyStruct(createBackupConfigRequest.BackupConfig.GRPC.Create.V1BackupConfig, backupConfig)
	if err != nil {
		return response, err
	}
	log.Infof("Backup Config - [%v]", backupConfig)
	backupRequest := &publicBackupConfigapis.CreateBackupConfigRequest{
		DeploymentId: createBackupConfigRequest.BackupConfig.V1.Create.DeploymentId,
		ProjectId:    createBackupConfigRequest.BackupConfig.V1.Create.ProjectId,
		BackupConfig: backupConfig,
	}

	ctx, client, _, err := backupConf.getBackupConfigClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting grpc client: %v\n", err)
	}

	ctx = WithAccountIDMetaCtx(ctx, backupConf.AccountId)

	apiResponse, err := client.CreateBackupConfig(ctx, backupRequest, grpc.PerRPCCredentials(credentials))
	log.Infof("api response [+%v]", apiResponse)
	if err != nil {
		return nil, fmt.Errorf("Error while creating the deployment: %v\n", err)
	}

	err = utilities.CopyStruct(apiResponse, response)
	if err != nil {
		return response, err
	}

	return response, nil
}

// UpdateBackupConfig will update backup config for a given deployment
func (backupConf *PdsGrpc) UpdateBackupConfig(updateBackupConfigRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("UpdateBackupConfig is not implemented for GRPC")
	return &apiStructs.WorkFlowResponse{}, nil

}

// GetBackupConfig will fetch backup config for a given deployment
func (backupConf *PdsGrpc) GetBackupConfig(getBackupConfigRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("GetBackupConfig is not implemented for GRPC")
	return &apiStructs.WorkFlowResponse{}, nil

}

// DeleteBackupConfig will delete backup config for a given deployment
func (backupConf *PdsGrpc) DeleteBackupConfig(deleteBackupConfigRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	log.Warnf("DeleteBackupConfig is not implemented for GRPC")
	return &apiStructs.WorkFlowResponse{}, nil

}

// ListBackupConfig will list backup config for a given deployment
func (backupConf *PdsGrpc) ListBackupConfig(listBackupConfigRequest *apiStructs.WorkFlowRequest) ([]apiStructs.WorkFlowResponse, error) {
	log.Warnf("ListBackupConfig is not implemented for GRPC")
	return []apiStructs.WorkFlowResponse{}, nil

}
