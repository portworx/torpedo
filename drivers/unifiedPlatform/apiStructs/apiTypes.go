package apiStructs

type WorkFlowResponse struct {
	Meta   Meta   `copier:"must,nopanic"`
	Config Config `copier:"must,nopanic"`
	Id     string `copier:"must,nopanic"`
}

type WorkFlowRequest struct {
	Deployment   PDSDeployment
	BackupConfig PDSBackupConfig
	Backup       PDSBackup
	Restore      PDSRestore
	Meta         Meta
	Config       Config
	Id           string
	ClusterId    string
	TenantId     string
	PdsAppId     string
	Pagination   PaginationRequest
}
