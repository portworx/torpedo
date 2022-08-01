package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type BackupCredential struct {
	context   context.Context
	apiClient *pds.APIClient
}

func (backupCredential *BackupCredential) ListBackupCredentials(tenantId string) ([]pds.ModelsBackupCredentials, error) {
	backupClient := backupCredential.apiClient.BackupCredentialsApi
	backupModels, res, err := backupClient.ApiTenantsIdBackupCredentialsGet(backupCredential.context, tenantId).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdBackupCredentialsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupModels.GetData(), err
}

func (backupCredential *BackupCredential) GetBackupCredential(backupCredId string) (*pds.ModelsBackupCredentials, error) {
	backupClient := backupCredential.apiClient.BackupCredentialsApi
	backupModel, res, err := backupClient.ApiBackupCredentialsIdGet(backupCredential.context, backupCredId).Execute()

	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupCredentialsIdGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupModel, err
}

func (backupCredential *BackupCredential) CreateAzureBackupCredential(tenantId string, name string, accountKey string, accountName string) (*pds.ModelsBackupCredentials, error) {
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
	backupModel, res, err := backupClient.ApiTenantsIdBackupCredentialsPost(backupCredential.context, tenantId).Body(createRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdBackupCredentialsPost``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupModel, err

}
func (backupCredential *BackupCredential) CreateS3BackupCredential(tenantId string, name string, accessKey string, endpoint string, secretKey string) (*pds.ModelsBackupCredentials, error) {
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
	backupModel, res, err := backupClient.ApiTenantsIdBackupCredentialsPost(backupCredential.context, tenantId).Body(createRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdBackupCredentialsPost``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupModel, err

}
func (backupCredential *BackupCredential) CreateS3CompatibleBackupCredential(tenantId string, name string, accessKey string, endpoint string, secretKey string) (*pds.ModelsBackupCredentials, error) {
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
	backupModel, res, err := backupClient.ApiTenantsIdBackupCredentialsPost(backupCredential.context, tenantId).Body(createRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiTenantsIdBackupCredentialsPost``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupModel, err

}
func (backupCredential *BackupCredential) UpdateAzureBackupCredential(backupCredsId string, name string, accountKey string, accountName string) (*pds.ModelsBackupCredentials, error) {
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
	backupModel, res, err := backupClient.ApiBackupCredentialsIdPut(backupCredential.context, backupCredsId).Body(updateRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupCredentialsIdPut``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupModel, err

}
func (backupCredential *BackupCredential) UpdateS3BackupCredential(backupCredsId string, name string, accessKey string, endpoint string, secretKey string) (*pds.ModelsBackupCredentials, error) {
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
	backupModel, res, err := backupClient.ApiBackupCredentialsIdPut(backupCredential.context, backupCredsId).Body(updateRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupCredentialsIdPut``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupModel, err

}
func (backupCredential *BackupCredential) UpdateS3CompatibleBackupCredential(backupCredsId string, name string, accessKey string, endpoint string, secretKey string) (*pds.ModelsBackupCredentials, error) {
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
	backupModel, res, err := backupClient.ApiBackupCredentialsIdPut(backupCredential.context, backupCredsId).Body(updateRequest).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupCredentialsIdPut``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return backupModel, err

}

func (backupCredential *BackupCredential) DeleteBackupCredential(backupCredsId string) (*status.Response, error) {
	backupClient := backupCredential.apiClient.BackupCredentialsApi
	res, err := backupClient.ApiBackupCredentialsIdDelete(backupCredential.context, backupCredsId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupCredentialsIdDelete``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return res, nil
}

func (backupCredential *BackupCredential) GetCloudCredentials(backupCredsId string) (*pds.ControllersPartialCredentials, error) {
	backupClient := backupCredential.apiClient.BackupCredentialsApi
	cloudCredsModel, res, err := backupClient.ApiBackupCredentialsIdCredentialsGet(backupCredential.context, backupCredsId).Execute()
	if res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiBackupCredentialsIdCredentialsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
	}
	return cloudCredsModel, err
}
