package apiStructs

type PDSBackup struct {
	V1   PDSBackupV1
	GRPC PDSBackupGRPC
}

type PDSBackupV1 struct {
	Get    GetPDSBackup
	Delete DeletePDSBackup
	List   ListPDSBackup
}

type PDSBackupGRPC struct {
	Get    GetPDSBackup
	Delete DeletePDSBackup
	List   ListPDSBackup
}

type GetPDSBackup struct {
	Id string
}

type DeletePDSBackup struct {
	Id string
}

type ListPDSBackup struct {
	AccountId            *string
	TenantId             *string
	ClusterId            *string
	NamespaceId          *string
	ProjectId            *string
	BackupConfigId       *string
	PaginationPageNumber *string
	PaginationPageSize   *string
	SortSortBy           *string
	SortSortOrder        *string
}
