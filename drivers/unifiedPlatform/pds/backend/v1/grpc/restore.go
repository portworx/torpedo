package grpc

import (
	"context"
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	publicRestoreapis "github.com/pure-px/apis/public/portworx/pds/restore/apiv1"
	"google.golang.org/grpc"
)

func (restore *PdsGrpc) getRestoreClient() (context.Context, publicRestoreapis.RestoreServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var depClient publicRestoreapis.RestoreServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	credentials = &Credentials{
		Token: token,
	}
	depClient = publicRestoreapis.NewRestoreServiceClient(restore.ApiClientV2)
	return ctx, depClient, token, nil
}

// CreateRestore will create restore for a given backup
func (restore *PdsGrpc) CreateRestore(createRestoreRequest *automationModels.PDSRestoreRequest) (*automationModels.PDSRestoreResponse, error) {
	// log.Infof("Backup Create - [%+v]", createBackupConfigRequest.BackupConfig.Create)

	response := &automationModels.PDSRestoreResponse{
		Create: automationModels.PDSRestore{},
	}

	createRequest := &publicRestoreapis.CreateRestoreRequest{}
	// log.Infof("Restore Create Request - [%v], Restore Config - [%v]", createRequest, createRequest.Restore)
	err := utilities.CopyStruct(createRestoreRequest.Create, createRequest)
	if err != nil {
		return response, err
	}
	// log.Infof("Restore Create Request - [%v], Restore Config - [%v]", createRequest, createRequest.Restore)

	ctx, client, _, err := restore.getRestoreClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting grpc client: %v\n", err)
	}

	ctx = WithAccountIDMetaCtx(ctx, restore.AccountId)

	apiResponse, err := client.CreateRestore(ctx, createRequest, grpc.PerRPCCredentials(credentials))
	log.Infof("api response [+%v]", apiResponse)
	if err != nil {
		return nil, fmt.Errorf("Error while creating the restore: %v\n", err)
	}

	err = utilities.CopyStruct(apiResponse, response.Create)
	if err != nil {
		return response, err
	}

	return response, nil
}

// ReCreateRestore will recreate restore for a given deployment
func (restore *PdsGrpc) ReCreateRestore(recretaeRestoreRequest *automationModels.PDSRestoreRequest) (*automationModels.PDSRestoreResponse, error) {

	// log.Infof("Backup Update - [%+v]", updateBackupConfigRequest.BackupConfig.Update)

	response := &automationModels.PDSRestoreResponse{
		ReCreate: automationModels.PDSRestore{},
	}

	recreateRequest := &publicRestoreapis.RecreateRestoreRequest{}
	// log.Infof("Restore Recretae - [%v]", recreateRequest)
	err := utilities.CopyStruct(recretaeRestoreRequest.ReCreate, recreateRequest)
	if err != nil {
		return response, err
	}
	// log.Infof("Restore Recretae - [%v]", recreateRequest)

	ctx, client, _, err := restore.getRestoreClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting grpc client: %v\n", err)
	}

	ctx = WithAccountIDMetaCtx(ctx, restore.AccountId)

	apiResponse, err := client.RecreateRestore(ctx, recreateRequest, grpc.PerRPCCredentials(credentials))
	log.Infof("api response [+%v]", apiResponse)
	if err != nil {
		return nil, fmt.Errorf("Error while recreating the restore: %v\n", err)
	}

	err = utilities.CopyStruct(apiResponse, response)
	if err != nil {
		return response, err
	}

	return response, nil

}

// GetRestore will fetch restore for a given deployment
func (restore *PdsGrpc) GetRestore(getRestoreRequest *automationModels.PDSRestoreRequest) (*automationModels.PDSRestoreResponse, error) {
	// log.Infof("Backup Get - [%+v]", getBackupConfigRequest.BackupConfig.Get)

	response := &automationModels.PDSRestoreResponse{
		Get: automationModels.PDSRestore{},
	}

	getRequest := &publicRestoreapis.GetRestoreRequest{}
	// log.Infof("Restore Get - [%v]", getRequest)
	err := utilities.CopyStruct(getRestoreRequest.Get, getRequest)
	if err != nil {
		return response, err
	}
	// log.Infof("Restore Get - [%v]", getRequest)

	ctx, client, _, err := restore.getRestoreClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting grpc client: %v\n", err)
	}

	ctx = WithAccountIDMetaCtx(ctx, restore.AccountId)

	apiResponse, err := client.GetRestore(ctx, getRequest, grpc.PerRPCCredentials(credentials))
	log.Infof("api response [+%v]", apiResponse)
	if err != nil {
		return nil, fmt.Errorf("Error while fetching the restore: %v\n", err)
	}

	err = utilities.CopyStruct(apiResponse, response)
	if err != nil {
		return response, err
	}

	return response, nil

}

// DeleteRestore will delete restore for a given deployment
func (restore *PdsGrpc) DeleteRestore(deleteRestoreRequest *automationModels.PDSRestoreRequest) error {
	// log.Infof("Backup Delete - [%+v]", deleteBackupConfigRequest.BackupConfig.Delete)

	//deleteRequest := &publicRestoreapis.DeleteRestoreRequest{}
	//// log.Infof("Restore Delete - [%v]", deleteRequest)
	//err := utilities.CopyStruct(deleteRestoreRequest.Delete, deleteRequest)
	//if err != nil {
	//	return err
	//}
	// log.Infof("Restore Delete - [%v]", deleteRequest)

	//ctx, client, _, err := restore.getRestoreClient()
	//if err != nil {
	//	return fmt.Errorf("Error while getting grpc client: %v\n", err)
	//}
	//
	//ctx = WithAccountIDMetaCtx(ctx, restore.AccountId)
	//
	//_, err = client.DeleteRestore(ctx, deleteRequest, grpc.PerRPCCredentials(credentials))
	//if err != nil {
	//	return fmt.Errorf("Error while deleting the restore: %v\n", err)
	//}

	return nil
}

// ListRestore will list restores for a given deployment
func (restore *PdsGrpc) ListRestore(listRestoresRequest *automationModels.PDSRestoreRequest) (*automationModels.PDSRestoreResponse, error) {
	// log.Infof("Backup List - [%+v]", listBackupConfigRequest.BackupConfig.List)

	response := &automationModels.PDSRestoreResponse{
		List: automationModels.PDSListRestoreResponse{},
	}

	listRequest := &publicRestoreapis.ListRestoresRequest{}
	// log.Infof("Restore List - [%v]", listRequest)
	err := utilities.CopyStruct(listRestoresRequest.List, listRequest)
	if err != nil {
		return response, err
	}
	// log.Infof("Restore List - [%v]", listRequest)

	ctx, client, _, err := restore.getRestoreClient()
	if err != nil {
		return nil, fmt.Errorf("Error while getting grpc client: %v\n", err)
	}

	ctx = WithAccountIDMetaCtx(ctx, restore.AccountId)

	apiResponse, err := client.ListRestores(ctx, listRequest, grpc.PerRPCCredentials(credentials))
	log.Infof("api response [+%v]", apiResponse)
	if err != nil {
		return nil, fmt.Errorf("Error while listing the backupConfig: %v\n", err)
	}

	err = utilities.CopyStruct(apiResponse, response.List)
	if err != nil {
		return response, err
	}

	return response, nil
}
