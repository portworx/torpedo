package apiStructs

type WorkFlowResponse struct {
	Meta   Meta   `copier:"must,nopanic"`
	Config Config `copier:"must,nopanic"`
	Id     string `copier:"must,nopanic"`
	Status Status
}

type WorkFlowRequest struct {
	Deployment            PDSDeployment         `copier:"must,nopanic"`
	BackupConfig          PDSBackupConfig       `copier:"must,nopanic"`
	Backup                PDSBackupConfig       `copier:"must,nopanic"`
	Restore               PDSRestore            `copier:"must,nopanic"`
	Meta                  Meta                  `copier:"must,nopanic"`
	Config                Config                `copier:"must,nopanic"`
	Id                    string                `copier:"must,nopanic"`
	ClusterId             string                `copier:"must,nopanic"`
	TenantId              string                `copier:"must,nopanic"`
	PdsAppId              string                `copier:"must,nopanic"`
	Pagination            PaginationRequest     `copier:"must,nopanic"`
	TargetClusterManifest TargetClusterManifest `copier:"must,nopanic"`
}
