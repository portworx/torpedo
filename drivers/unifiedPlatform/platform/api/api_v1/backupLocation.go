package apiv1

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/portworx/torpedo/pkg/log"

	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	platformv1 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// GetBackupLocClient updates the header with bearer token and returns the new client
func (ns *PLATFORM_API_V1) GetBackupLocClient() (context.Context, *platformv1.BackupLocationServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	ns.ApiClientV1.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	ns.ApiClientV1.GetConfig().DefaultHeader["px-account-id"] = ns.AccountID
	client := ns.ApiClientV1.BackupLocationServiceAPI

	return ctx, client, nil
}

// ListBackupLocations return lis of backup locations
func (ns *PLATFORM_API_V1) ListBackupLocations() ([]ApiResponse, error) {
	ctx, backupLocationClient, err := ns.GetBackupLocClient()
	bckpLocResponse := []ApiResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	backupLocationModels, res, err := backupLocationClient.BackupLocationServiceListBackupLocations(ctx).Execute()
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupLocationServiceListBackupLocations`: %v\n.Full HTTP response: %v", err, res)
	}
	log.Infof("Value of backupLocation - [%v]", backupLocationModels)
	copier.Copy(&bckpLocResponse, backupLocationModels.BackupLocations)
	log.Infof("Value of backupLocation after copy - [%v]", bckpLocResponse)
	return bckpLocResponse, nil
}

// GetBackupLocation get backup location model by its ID.
func (ns *PLATFORM_API_V1) GetBackupLocation(backupLocID string) (*ApiResponse, error) {
	ctx, backupLocationClient, err := ns.GetBackupLocClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	bckpLocResp := ApiResponse{}
	var getRequest platformv1.ApiBackupLocationServiceGetBackupLocationRequest
	getRequest = getRequest.ApiService.BackupLocationServiceGetBackupLocation(ctx, backupLocID)
	backupLocationModel, res, err := backupLocationClient.BackupLocationServiceGetBackupLocationExecute(getRequest)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when called `BackupLocationServiceGetBackupLocation`, Full HTTP response: %v\n", res)
	}
	log.Infof("Value of backupLocation - [%v]", backupLocationModel)
	copier.Copy(&bckpLocResp, backupLocationModel)
	log.Infof("Value of backupLocation after copy - [%v]", bckpLocResp)
	return &bckpLocResp, nil
}

// CreateBackupLocation return newly created backup location model.
func (ns *PLATFORM_API_V1) CreateBackupLocation(tenantID string, createReq platformv1.ApiBackupLocationServiceCreateBackupLocationRequest) (*ApiResponse, error) {
	ctx, backupLocationClient, err := ns.GetBackupLocClient()
	bckpLocResp := ApiResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	createReq = createReq.ApiService.BackupLocationServiceCreateBackupLocation(ctx, tenantID)
	backupLocationModel, _, err := backupLocationClient.BackupLocationServiceCreateBackupLocationExecute(createReq)
	if err != nil {
		return nil, fmt.Errorf("error when called `BackupLocationServiceCreateBackupLocation` to create backup target - %v", err)
	}
	copier.Copy(&bckpLocResp, backupLocationModel)
	log.Infof("Value of backupLocation after copy - [%v]", bckpLocResp)
	return &bckpLocResp, nil
}

// UpdateBackupLocation return updated backup location model.
func (ns *PLATFORM_API_V1) UpdateBackupLocation(backupLocationID string, updateReq platformv1.ApiBackupLocationServiceUpdateBackupLocationRequest, updateValue string) (*ApiResponse, error) {
	ctx, backupLocationClient, err := ns.GetBackupLocClient()
	bckpLocResp := ApiResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	updateReq = updateReq.UpdateMask(updateValue)
	updateReq = updateReq.ApiService.BackupLocationServiceUpdateBackupLocation(ctx, backupLocationID)
	backupLocationModel, res, err := backupLocationClient.BackupLocationServiceUpdateBackupLocationExecute(updateReq)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupLocationServiceUpdateBackupLocation`: %v\n.Full HTTP response: %v", err, res)
	}
	copier.Copy(&bckpLocResp, backupLocationModel)
	log.Infof("Value of backupLocation after copy - [%v]", bckpLocResp)
	return &bckpLocResp, nil

}

// SyncToBackupLocation returned synced backup location model.

// DeleteBackupLocation delete backup location and return status.
func (ns *PLATFORM_API_V1) DeleteBackupLocation(backupLocationID string) (*status.Response, error) {
	ctx, backupLocationClient, err := ns.GetBackupLocClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, err := backupLocationClient.BackupLocationServiceDeleteBackupLocation(ctx, backupLocationID).Execute()
	if err != nil {
		return nil, fmt.Errorf("Error when calling `BackupLocationServiceDeleteBackupLocation`: %v\n.Full HTTP response: %v", err, res)
	}
	return res, nil
}
