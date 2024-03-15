package automationModels

type PDSBackup struct {
	Delete PDSDeleteBackup `copier:"must,nopanic"`
	List   PDSListBackup   `copier:"must,nopanic"`
	Get    PDSGetRestore   `copier:"must,nopanic"`
}

type PDSDeleteBackup struct {
	Id string `copier:"must,nopanic"`
}

type PDSListBackup struct {
	Pagination *PageBasedPaginationRequest `copier:"must,nopanic"`
	Sort       *Sort                       `copier:"must,nopanic"`
}

type PDSGetBackupRequest struct {
	Id string `copier:"must,nopanic"`
}
