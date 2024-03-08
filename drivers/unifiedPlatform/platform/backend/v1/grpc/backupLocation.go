package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	commonapiv1 "github.com/pure-px/apis/public/portworx/common/apiv1"
	publicbackuplocapi "github.com/pure-px/apis/public/portworx/platform/backuplocation/apiv1"
	"google.golang.org/grpc"
)

// getBackupLocClient updates the header with bearer token and returns the new client
func (BackupLocGrpcV1 *PlatformGrpc) getBackupLocClient() (context.Context, publicbackuplocapi.BackupLocationServiceClient, string, error) {
	log.Infof("Creating client from grpc package")
	var backupLocClient publicbackuplocapi.BackupLocationServiceClient

	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	credentials = &Credentials{
		Token: token,
	}

	backupLocClient = publicbackuplocapi.NewBackupLocationServiceClient(BackupLocGrpcV1.ApiClientV1)

	return ctx, backupLocClient, token, nil
}

// ListBackupLocations return lis of backup locations
func (BackupLocGrpcV1 *PlatformGrpc) ListBackupLocations() ([]WorkFlowResponse, error) {
	ctx, backupLocationClient, _, err := BackupLocGrpcV1.getBackupLocClient()
	bckpLocResponse := []WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	firstPageRequest := &publicbackuplocapi.ListBackupLocationsRequest{
		Pagination: NewPaginationRequest(1, 50),
	}
	backupLocationModels, err := backupLocationClient.ListBackupLocations(ctx, firstPageRequest, grpc.PerRPCCredentials(credentials))
	log.Infof("Value of tenants - [%v]", backupLocationModels)
	err = copier.Copy(&bckpLocResponse, backupLocationModels.BackupLocations)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of backupLocation after copy - [%v]", bckpLocResponse)
	return bckpLocResponse, nil
}

// GetBackupLocation get backup location model by its ID.
func (BackupLocGrpcV1 *PlatformGrpc) GetBackupLocation(getReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, backupLocationClient, _, err := BackupLocGrpcV1.getBackupLocClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	bckpLocResp := WorkFlowResponse{}
	var getRequest *publicbackuplocapi.GetBackupLocationRequest
	err = copier.Copy(&getRequest, getReq)
	if err != nil {
		return nil, err
	}
	backupLocationModel, err := backupLocationClient.GetBackupLocation(ctx, getRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	log.Infof("Value of backupLocation - [%v]", backupLocationModel)
	err = copier.Copy(&bckpLocResp, backupLocationModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of backupLocation after copy - [%v]", bckpLocResp)
	return &bckpLocResp, nil
}

// CreateBackupLocation return newly created backup location model.
func (BackupLocGrpcV1 *PlatformGrpc) CreateBackupLocation(createRequest *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, backupLocationClient, _, err := BackupLocGrpcV1.getBackupLocClient()
	bckpLocResp := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	createAppRequest := &publicbackuplocapi.CreateBackupLocationRequest{
		TenantId: "",
		BackupLocation: &publicbackuplocapi.BackupLocation{
			Meta: &commonapiv1.Meta{
				Uid:             "",
				Name:            "",
				Description:     "",
				ResourceVersion: "",
				CreateTime:      nil,
				UpdateTime:      nil,
				Labels:          nil,
				Annotations:     nil,
				ParentReference: nil,
				ResourceNames:   nil,
			},
			Config: &publicbackuplocapi.Config{
				Provider:          nil,
				CloudCredentialId: "",
				Location:          nil,
			},
			Status: nil,
		},
	}
	//err = copier.Copy(&createAppRequest, createRequest)
	//if err != nil {
	//	return nil, err
	//}
	backupLocationModel, err := backupLocationClient.CreateBackupLocation(ctx, createAppRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("error when called `BackupLocationServiceCreateBackupLocation` to create backup target - %v", err)
	}
	err = copier.Copy(&bckpLocResp, backupLocationModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of backupLocation after copy - [%v]", bckpLocResp)
	return &bckpLocResp, nil
}

// UpdateBackupLocation return updated backup location model.
func (BackupLocGrpcV1 *PlatformGrpc) UpdateBackupLocation(updateRequest *WorkFlowRequest) (*WorkFlowResponse, error) {
	ctx, backupLocationClient, _, err := BackupLocGrpcV1.getBackupLocClient()
	bckpLocResp := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var updateAppRequest *publicbackuplocapi.UpdateBackupLocationRequest
	err = copier.Copy(&updateAppRequest, updateRequest)
	if err != nil {
		return nil, err
	}
	backupLocationModel, err := backupLocationClient.UpdateBackupLocation(ctx, updateAppRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	err = copier.Copy(&bckpLocResp, backupLocationModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of backupLocation after copy - [%v]", bckpLocResp)
	return &bckpLocResp, nil

}

// SyncToBackupLocation returned synced backup location model.

// DeleteBackupLocation delete backup location and return status.
func (BackupLocGrpcV1 *PlatformGrpc) DeleteBackupLocation(backupLocationID *WorkFlowRequest) error {
	ctx, backupLocationClient, _, err := BackupLocGrpcV1.getBackupLocClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	deleteRequest := &publicbackuplocapi.DeleteBackupLocationRequest{Id: backupLocationID.Id}
	_, err = backupLocationClient.DeleteBackupLocation(ctx, deleteRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return fmt.Errorf("Error when calling `BackupLocationServiceDeleteBackupLocation`: %v\n", err)
	}
	return nil
}
