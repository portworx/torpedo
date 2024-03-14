package api

import (
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/pure-px/platform-api-go-client/platform/v1/backuplocation"

	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	status "net/http"
)


// ListBackupLocations return lis of backup locatiobackuploc
func (backuploc *PLATFORM_API_V1) ListBackupLocations(request *BackupLocation) ([]BackupLocation, error) {
	ctx, backupLocationClient, err := backuploc.getBackupLocClient()
	backupLocResp := []BackupLocation{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	backupLocationModels, res, err := backupLocationClient.BackupLocationServiceListBackupLocations(ctx).Execute()
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupLocatiobackuplocerviceListBackupLocatiobackuploc`: %v\n.Full HTTP respobackuploce: %v", err, res)
	}
	log.Infof("Value of backupLocation - [%v]", backupLocationModels)
	err = copier.Copy(&backupLocResp, backupLocationModels.BackupLocations)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of backupLocation after copy - [%v]", backupLocResp)
	return backupLocResp, nil
}

// GetBackupLocation get backup location model by its ID.
func (backuploc *PLATFORM_API_V1) GetBackupLocation(getReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	_, backupLocationClient, err := backuploc.getBackupLocClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	bckpLocResp := WorkFlowResponse{}
	var getRequest backuplocation.ApiBackupLocationServiceGetBackupLocationRequest
	err = copier.Copy(&getRequest, getReq)
	if err != nil {
		return nil, err
	}
	backupLocationModel, res, err := backupLocationClient.BackupLocationServiceGetBackupLocationExecute(getRequest)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when called `BackupLocatiobackuplocerviceGetBackupLocation`, Full HTTP respobackuploce: %v\n", res)
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
func (backuploc *PLATFORM_API_V1) CreateBackupLocation(createReq *BackupLocation) (*BackupLocation, error) {
	_, backupLocationClient, err := backuploc.getBackupLocClient()
	bckpLocResp := BackupLocation{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var createBackupLoc backuplocation.ApiBackupLocationServiceCreateBackupLocationRequest
	err = copier.Copy(&createBackupLoc, createReq)
	if err != nil {
		return nil, err
	}
	backupLocationModel, _, err := backupLocationClient.BackupLocationServiceCreateBackupLocationExecute(createBackupLoc)
	if err != nil {
		return nil, fmt.Errorf("error when called `BackupLocatiobackuplocerviceCreateBackupLocation` to create backup target - %v", err)
	}
	err = copier.Copy(&bckpLocResp, backupLocationModel)
	if err != nil {
		return nil, err
	}
	log.Infof("Value of backupLocation after copy - [%v]", bckpLocResp)
	return &bckpLocResp, nil
}

// UpdateBackupLocation return updated backup location model.
func (backuploc *PLATFORM_API_V1) UpdateBackupLocation(updateReq *WorkFlowRequest) (*WorkFlowResponse, error) {
	_, backupLocationClient, err := backuploc.getBackupLocClient()
	bckpLocResp := WorkFlowResponse{}
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	var updateBackupLoc backuplocation.ApiBackupLocationServiceUpdateBackupLocationRequest
	err = copier.Copy(&updateBackupLoc, updateReq)
	if err != nil {
		return nil, err
	}
	backupLocationModel, res, err := backupLocationClient.BackupLocationServiceUpdateBackupLocationExecute(updateBackupLoc)
	if res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `BackupLocatiobackuplocerviceUpdateBackupLocation`: %v\n.Full HTTP respobackuploce: %v", err, res)
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
func (backuploc *PLATFORM_API_V1) DeleteBackupLocation(backupLocationID *WorkFlowRequest) error {
	ctx, backupLocationClient, err := backuploc.getBackupLocClient()
	if err != nil {
		return fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, err := backupLocationClient.BackupLocationServiceDeleteBackupLocation(ctx, backupLocationID.Id).Execute()
	if err != nil {
		return fmt.Errorf("Error when calling `BackupLocatiobackuplocerviceDeleteBackupLocation`: %v\n.Full HTTP respobackuploce: %v", err, res)
	}
	return nil
}
