package apiStructs

type PDSBackup struct {
	Delete PDSDeleteBackup `copier:"must,nopanic"`
	List   PDSListBackup   `copier:"must,nopanic"`
}

type PDSDeleteBackup struct {
	Id string `copier:"must,nopanic"`
}

type PDSListBackup struct {
	Pagination *PageBasedPaginationRequest `copier:"must,nopanic"`
	Sort       *Sort                       `copier:"must,nopanic"`
}

type GetBackupRequest struct {
	Id string `copier:"must,nopanic"`
}
