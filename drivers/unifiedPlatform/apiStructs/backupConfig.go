package apiStructs

import pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"

type PDSBackupConfig struct {
	Create CreatePDSBackupConfig
	Get    GetPDSBackupConfig
	Update UpdatePDSBackupConfig
	Delete DeletePDSBackupConfig
	List   ListPDSBackupConfig
}

type CreatePDSBackupConfig struct {
	V1   pdsv2.ApiBackupConfigServiceCreateBackupConfigRequest
	GRPC pdsv2.ApiBackupConfigServiceCreateBackupConfigRequest
}

type UpdatePDSBackupConfig struct {
	V1   pdsv2.ApiBackupConfigServiceUpdateBackupConfigRequest
	GRPC pdsv2.ApiBackupConfigServiceUpdateBackupConfigRequest
}

type DeletePDSBackupConfig struct {
	V1   pdsv2.ApiBackupConfigServiceDeleteBackupConfigRequest
	GRPC pdsv2.ApiBackupConfigServiceDeleteBackupConfigRequest
}

type GetPDSBackupConfig struct {
	V1   pdsv2.ApiBackupConfigServiceGetBackupConfigRequest
	GRPC pdsv2.ApiBackupConfigServiceGetBackupConfigRequest
}

type ListPDSBackupConfig struct {
	V1   pdsv2.ApiBackupConfigServiceListBackupConfigsRequest
	GRPC pdsv2.ApiBackupConfigServiceListBackupConfigsRequest
}
