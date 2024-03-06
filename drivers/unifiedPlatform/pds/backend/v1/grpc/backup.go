package grpc

import (
	"context"
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	publicBackupapis "github.com/pure-px/apis/public/portworx/pds/backup/apiv1"
	"google.golang.org/grpc"
)

func (backup *PdsGrpc) getBackupClient() (context.Context, publicBackupapis.BackupServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var depClient publicBackupapis.BackupServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	credentials = &Credentials{
		Token: token,
	}
	depClient = publicBackupapis.NewBackupServiceClient(backup.ApiClientV2)
	return ctx, depClient, token, nil
}

// DeleteBackup will delete backup for a given deployment
func (backup *PdsGrpc) DeleteBackup(deleteBackupRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	// log.Infof("Backup Delete - [%+v]", deleteBackupConfigRequest.BackupConfig.Delete)

	response := &apiStructs.WorkFlowResponse{}

	deleteRequest := &publicBackupapis.DeleteBackupRequest{}
	log.Infof("Backup Delete - [%v]", deleteRequest)
	err := utilities.CopyStruct(deleteBackupRequest.Backup.Delete, deleteRequest)
	if err != nil {
		return response, err
	}
	log.Infof("Backup Delete - [%v]", deleteRequest)

	ctx, client, _, err := backup.getBackupClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting grpc client: %v\n", err)
	}

	ctx = WithAccountIDMetaCtx(ctx, backup.AccountId)

	apiResponse, err := client.DeleteBackup(ctx, deleteRequest, grpc.PerRPCCredentials(credentials))
	log.Infof("api response [+%v]", apiResponse)
	if err != nil {
		return nil, fmt.Errorf("Error while deleting the backup: %v\n", err)
	}

	err = utilities.CopyStruct(apiResponse, response)
	if err != nil {
		return response, err
	}

	return response, nil
}

// ListBackup will list backup for a given deployment
func (backup *PdsGrpc) ListBackup(listBackupConfigRequest *apiStructs.WorkFlowRequest) ([]apiStructs.WorkFlowResponse, error) {
	// log.Infof("Backup List - [%+v]", listBackupConfigRequest.BackupConfig.List)

	response := []apiStructs.WorkFlowResponse{}

	listRequest := &publicBackupapis.ListBackupsRequest{}
	log.Infof("Backup List - [%v]", listRequest)
	err := utilities.CopyStruct(listBackupConfigRequest.Backup.List, listRequest)
	if err != nil {
		return response, err
	}
	log.Infof("Backup List - [%v]", listRequest)

	ctx, client, _, err := backup.getBackupClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting grpc client: %v\n", err)
	}

	ctx = WithAccountIDMetaCtx(ctx, backup.AccountId)

	apiResponse, err := client.ListBackups(ctx, listRequest, grpc.PerRPCCredentials(credentials))
	log.Infof("api response [+%v]", apiResponse)
	if err != nil {
		return nil, fmt.Errorf("Error while listing the backups: %v\n", err)
	}

	err = utilities.CopyStruct(apiResponse, response)
	if err != nil {
		return response, err
	}

	return response, nil
}

// GetBackup will fetch backup for a given deployment
func (backup *PdsGrpc) GetBackup(getBackupConfigRequest *apiStructs.WorkFlowRequest) (*apiStructs.WorkFlowResponse, error) {
	// log.Infof("Backup Get - [%+v]", getBackupConfigRequest.BackupConfig.Get)

	response := &apiStructs.WorkFlowResponse{}

	getRequest := &publicBackupapis.GetBackupRequest{}
	// log.Infof("Backup Get - [%v]", getRequest)
	err := utilities.CopyStruct(getBackupConfigRequest.Backup.Get, getRequest)
	if err != nil {
		return response, err
	}
	// log.Infof("Backup Get - [%v]", getRequest)

	ctx, client, _, err := backup.getBackupClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting grpc client: %v\n", err)
	}

	ctx = WithAccountIDMetaCtx(ctx, backup.AccountId)

	apiResponse, err := client.GetBackup(ctx, getRequest, grpc.PerRPCCredentials(credentials))
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
