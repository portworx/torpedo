package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// BackupCredential struct
type BackupCredential struct {
	context   context.Context
	apiClient *pds.APIClient
}

// ListBackupCredentials func
func (backupCredential *BackupCredential) ListBackupCredentials(tenantID string) ([]pds.ModelsBackupCredentials, error) {
	backupClient := backupCredential.apiClient.BackupCredentialsApi
	backupModels, res, err := backupClient.ApiTenantsIdBackupCredentialsGet(backupCredential.context, tenantID).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdBackupCredentialsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupModels.GetData(), err
}

// GetBackupCredential func
func (backupCredential *BackupCredential) GetBackupCredential(backupCredID string) (*pds.ModelsBackupCredentials, error) {
	backupClient := backupCredential.apiClient.BackupCredentialsApi
	backupModel, res, err := backupClient.ApiBackupCredentialsIdGet(backupCredential.context, backupCredID).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupCredentialsIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupModel, err
}

// CreateAzureBackupCredential func
func (backupCredential *BackupCredential) CreateAzureBackupCredential(tenantID string, name string, accountKey string, accountName string) (*pds.ModelsBackupCredentials, error) {
	backupClient := backupCredential.apiClient.BackupCredentialsApi
	azureCredsModel := pds.ModelsAzureCredentials{
		AccountKey:  &accountKey,
		AccountName: &accountName,
	}
	controllerCreds := pds.ControllersCredentials{
		Azure: &azureCredsModel,
	}
	createRequest := pds.ControllersCreateBackupCredentialsRequest{
		Credentials: &controllerCreds,
		Name:        &name,
	}
	backupModel, res, err := backupClient.ApiTenantsIdBackupCredentialsPost(backupCredential.context, tenantID).Body(createRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdBackupCredentialsPost``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupModel, err

}

// CreateS3BackupCredential func
func (backupCredential *BackupCredential) CreateS3BackupCredential(tenantID string, name string, accessKey string, endpoint string, secretKey string) (*pds.ModelsBackupCredentials, error) {
	backupClient := backupCredential.apiClient.BackupCredentialsApi
	s3CredsModel := pds.ModelsS3Credentials{
		AccessKey: &accessKey,
		Endpoint:  &endpoint,
		SecretKey: &secretKey,
	}
	controllerCreds := pds.ControllersCredentials{
		S3: &s3CredsModel,
	}
	createRequest := pds.ControllersCreateBackupCredentialsRequest{
		Credentials: &controllerCreds,
		Name:        &name,
	}
	backupModel, res, err := backupClient.ApiTenantsIdBackupCredentialsPost(backupCredential.context, tenantID).Body(createRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdBackupCredentialsPost``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupModel, err

}

// CreateS3CompatibleBackupCredential func
func (backupCredential *BackupCredential) CreateS3CompatibleBackupCredential(tenantID string, name string, accessKey string, endpoint string, secretKey string) (*pds.ModelsBackupCredentials, error) {
	backupClient := backupCredential.apiClient.BackupCredentialsApi
	s3CompatibleCredsModel := pds.ModelsS3CompatibleCredentials{
		AccessKey: &accessKey,
		Endpoint:  &endpoint,
		SecretKey: &secretKey,
	}
	controllerCreds := pds.ControllersCredentials{
		S3Compatible: &s3CompatibleCredsModel,
	}
	createRequest := pds.ControllersCreateBackupCredentialsRequest{
		Credentials: &controllerCreds,
		Name:        &name,
	}
	backupModel, res, err := backupClient.ApiTenantsIdBackupCredentialsPost(backupCredential.context, tenantID).Body(createRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdBackupCredentialsPost``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupModel, err

}

// UpdateAzureBackupCredential func
func (backupCredential *BackupCredential) UpdateAzureBackupCredential(backupCredsID string, name string, accountKey string, accountName string) (*pds.ModelsBackupCredentials, error) {
	backupClient := backupCredential.apiClient.BackupCredentialsApi
	azureCredsModel := pds.ModelsAzureCredentials{
		AccountKey:  &accountKey,
		AccountName: &accountName,
	}
	controllerCreds := pds.ControllersCredentials{
		Azure: &azureCredsModel,
	}
	updateRequest := pds.ControllersUpdateBackupCredentialsRequest{
		Credentials: &controllerCreds,
		Name:        &name,
	}
	backupModel, res, err := backupClient.ApiBackupCredentialsIdPut(backupCredential.context, backupCredsID).Body(updateRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupCredentialsIdPut``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupModel, err

}

// UpdateS3BackupCredential func
func (backupCredential *BackupCredential) UpdateS3BackupCredential(backupCredsID string, name string, accessKey string, endpoint string, secretKey string) (*pds.ModelsBackupCredentials, error) {
	backupClient := backupCredential.apiClient.BackupCredentialsApi
	s3CredsModel := pds.ModelsS3Credentials{
		AccessKey: &accessKey,
		Endpoint:  &endpoint,
		SecretKey: &secretKey,
	}
	controllerCreds := pds.ControllersCredentials{
		S3: &s3CredsModel,
	}
	updateRequest := pds.ControllersUpdateBackupCredentialsRequest{
		Credentials: &controllerCreds,
		Name:        &name,
	}
	backupModel, res, err := backupClient.ApiBackupCredentialsIdPut(backupCredential.context, backupCredsID).Body(updateRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupCredentialsIdPut``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupModel, err

}

// UpdateS3CompatibleBackupCredential func
func (backupCredential *BackupCredential) UpdateS3CompatibleBackupCredential(backupCredsID string, name string, accessKey string, endpoint string, secretKey string) (*pds.ModelsBackupCredentials, error) {
	backupClient := backupCredential.apiClient.BackupCredentialsApi
	s3CompatibleCredsModel := pds.ModelsS3CompatibleCredentials{
		AccessKey: &accessKey,
		Endpoint:  &endpoint,
		SecretKey: &secretKey,
	}
	controllerCreds := pds.ControllersCredentials{
		S3Compatible: &s3CompatibleCredsModel,
	}
	updateRequest := pds.ControllersUpdateBackupCredentialsRequest{
		Credentials: &controllerCreds,
		Name:        &name,
	}
	backupModel, res, err := backupClient.ApiBackupCredentialsIdPut(backupCredential.context, backupCredsID).Body(updateRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupCredentialsIdPut``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupModel, err

}

// DeleteBackupCredential func
func (backupCredential *BackupCredential) DeleteBackupCredential(backupCredsID string) (*status.Response, error) {
	backupClient := backupCredential.apiClient.BackupCredentialsApi
	res, err := backupClient.ApiBackupCredentialsIdDelete(backupCredential.context, backupCredsID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupCredentialsIdDelete``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return res, nil
}

// GetCloudCredentials func
func (backupCredential *BackupCredential) GetCloudCredentials(backupCredsID string) (*pds.ControllersPartialCredentials, error) {
	backupClient := backupCredential.apiClient.BackupCredentialsApi
	cloudCredsModel, res, err := backupClient.ApiBackupCredentialsIdCredentialsGet(backupCredential.context, backupCredsID).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupCredentialsIdCredentialsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return cloudCredsModel, err
}
