package grpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	publicbackuplocapi "github.com/pure-px/apis/public/portworx/platform/backuplocation/apiv1"
	"google.golang.org/grpc"
)

type BackupLocGrpc struct {
	ApiClientV1 *grpc.ClientConn
}

// getBackupLocClient updates the header with bearer token and returns the new client
func (BackupLocGrpcV1 *BackupLocGrpc) getBackupLocClient() (context.Context, publicbackuplocapi.BackupLocationServiceClient, string, error) {
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
func (BackupLocGrpcV1 *BackupLocGrpc) ListBackupLocations() ([]ApiResponse, error) {
	ctx, backupLocationClient, _, err := BackupLocGrpcV1.getBackupLocClient()
	bckpLocResponse := []ApiResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	firstPageRequest := &publicbackuplocapi.ListBackupLocationsRequest{
		Pagination: NewPaginationRequest(1, 50),
	}
	backupLocationModels, err := backupLocationClient.ListBackupLocations(ctx, firstPageRequest, grpc.PerRPCCredentials(credentials))
	log.Infof("Value of tenants - [%v]", backupLocationModels)
	copier.Copy(&bckpLocResponse, backupLocationModels.BackupLocations)
	log.Infof("Value of accounts after copy - [%v]", bckpLocResponse)
	return bckpLocResponse, nil
}

// GetBackupLocation get backup location model by its ID.
func (BackupLocGrpcV1 *BackupLocGrpc) GetBackupLocation(backupLocID string) (*ApiResponse, error) {
	ctx, backupLocationClient, _, err := BackupLocGrpcV1.getBackupLocClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	bckpLocResp := ApiResponse{}
	getRequest := &publicbackuplocapi.GetBackupLocationRequest{Id: backupLocID}
	backupLocationModel, err := backupLocationClient.GetBackupLocation(ctx, getRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	log.Infof("Value of backupLocation - [%v]", backupLocationModel)
	copier.Copy(&bckpLocResp, backupLocationModel)
	log.Infof("Value of accounts after copy - [%v]", bckpLocResp)
	return &bckpLocResp, nil
}

// CreateBackupLocation return newly created backup location model.
func (BackupLocGrpcV1 *BackupLocGrpc) CreateBackupLocation(tenantID string) (*ApiResponse, error) {
	ctx, backupLocationClient, _, err := BackupLocGrpcV1.getBackupLocClient()
	bckpLocResp := ApiResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	createRequest := &publicbackuplocapi.CreateBackupLocationRequest{
		TenantId: tenantID,
		BackupLocation: &publicbackuplocapi.BackupLocation{
			Meta:   nil,
			Config: nil,
			Status: nil,
		},
	}
	backupLocationModel, err := backupLocationClient.CreateBackupLocation(ctx, createRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("error when called `BackupLocationServiceCreateBackupLocation` to create backup target - %v", err)
	}
	copier.Copy(&bckpLocResp, backupLocationModel)
	log.Infof("Value of accounts after copy - [%v]", bckpLocResp)
	return &bckpLocResp, nil
}

// UpdateBackupLocation return updated backup location model.
func (BackupLocGrpcV1 *BackupLocGrpc) UpdateBackupLocation(backupLocationID string) (*ApiResponse, error) {
	ctx, backupLocationClient, _, err := BackupLocGrpcV1.getBackupLocClient()
	bckpLocResp := ApiResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	updateRequest := &publicbackuplocapi.UpdateBackupLocationRequest{
		UpdateMask: nil,
		Id:         backupLocationID,
		BackupLocation: &publicbackuplocapi.BackupLocation{
			Meta:   nil,
			Config: nil,
			Status: nil,
		},
	}
	backupLocationModel, err := backupLocationClient.UpdateBackupLocation(ctx, updateRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	copier.Copy(&bckpLocResp, backupLocationModel)
	log.Infof("Value of accounts after copy - [%v]", bckpLocResp)
	return &bckpLocResp, nil

}

// SyncToBackupLocation returned synced backup location model.

// DeleteBackupLocation delete backup location and return status.
func (BackupLocGrpcV1 *BackupLocGrpc) DeleteBackupLocation(backupLocationID string) error {
	ctx, backupLocationClient, _, err := BackupLocGrpcV1.getBackupLocClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	deleteRequest := &publicbackuplocapi.DeleteBackupLocationRequest{Id: backupLocationID}
	_, err = backupLocationClient.DeleteBackupLocation(ctx, deleteRequest, grpc.PerRPCCredentials(credentials))
	if err != nil {
		return fmt.Errorf("Error when calling `BackupLocationServiceDeleteBackupLocation`: %v\n", err)
	}
	return nil
}
