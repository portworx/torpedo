package apiStructs

import pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"

type PDSRestore struct {
	Create   CreatePDSRestore
	ReCreate ReCreatePDSRestore
	Get      GetPDSRestore
	List     ListPDSRestore
	Delete   DeletePDSRestore
}

type CreatePDSRestore struct {
	V1   pdsv2.ApiRestoreServiceCreateRestoreRequest
	GRPC pdsv2.ApiRestoreServiceCreateRestoreRequest
}

type ReCreatePDSRestore struct {
	V1   pdsv2.ApiRestoreServiceRecreateRestoreRequest
	GRPC pdsv2.ApiRestoreServiceRecreateRestoreRequest
}

type GetPDSRestore struct {
	V1   pdsv2.ApiRestoreServiceGetRestoreRequest
	GRPC pdsv2.ApiRestoreServiceGetRestoreRequest
}

type ListPDSRestore struct {
	V1   pdsv2.ApiRestoreServiceListRestoresRequest
	GRPC pdsv2.ApiRestoreServiceListRestoresRequest
}

type DeletePDSRestore struct {
	V1   pdsv2.ApiRestoreServiceDeleteRestoreRequest
	GRPC pdsv2.ApiRestoreServiceDeleteRestoreRequest
}
