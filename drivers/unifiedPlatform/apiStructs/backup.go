package apiStructs

type PDSBackup struct {
	V1   PDSBackupV1   `copier:"must,nopanic"`
	GRPC PDSBackupGRPC `copier:"must,nopanic"`
}

type PDSBackupV1 struct {
	Get    GetPDSBackup    `copier:"must,nopanic"`
	Delete DeletePDSBackup `copier:"must,nopanic"`
	List   ListPDSBackup   `copier:"must,nopanic"`
}

type PDSBackupGRPC struct {
	Get    GetPDSBackup    `copier:"must,nopanic"`
	Delete DeletePDSBackup `copier:"must,nopanic"`
	List   ListPDSBackup   `copier:"must,nopanic"`
}

type GetPDSBackup struct {
	Id string `copier:"must,nopanic"`
}

type DeletePDSBackup struct {
	Id string `copier:"must,nopanic"`
}

type ListPDSBackup struct {
	AccountId            *string `copier:"must,nopanic"`
	TenantId             *string `copier:"must,nopanic"`
	ClusterId            *string `copier:"must,nopanic"`
	NamespaceId          *string `copier:"must,nopanic"`
	ProjectId            *string `copier:"must,nopanic"`
	BackupConfigId       *string `copier:"must,nopanic"`
	PaginationPageNumber *string `copier:"must,nopanic"`
	PaginationPageSize   *string `copier:"must,nopanic"`
	SortSortBy           *string `copier:"must,nopanic"`
	SortSortOrder        *string `copier:"must,nopanic"`
}
