package apiStructs

import pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"

type PDSBackup struct {
	Get    GetPDSBackup
	Delete DeletePDSBackup
	List   ListPDSBackup
}

type DeletePDSBackup struct {
	V1   pdsv2.ApiBackupServiceDeleteBackupRequest
	GRPC pdsv2.ApiBackupServiceDeleteBackupRequest
}

type GetPDSBackup struct {
	V1   pdsv2.ApiBackupServiceGetBackupRequest
	GRPC pdsv2.ApiBackupServiceGetBackupRequest
}

type ListPDSBackup struct {
	V1   pdsv2.ApiBackupServiceListBackupsRequest
	GRPC pdsv2.ApiBackupServiceListBackupsRequest
}
