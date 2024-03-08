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
	CloudCredentials      CloudCredentials      `copier:"must,nopanic"`
	Meta                  Meta                  `copier:"must,nopanic"`
	Config                Config                `copier:"must,nopanic"`
	Id                    string                `copier:"must,nopanic"`
	ClusterId             string                `copier:"must,nopanic"`
	TenantId              string                `copier:"must,nopanic"`
	PdsAppId              string                `copier:"must,nopanic"`
	Pagination            PaginationRequest     `copier:"must,nopanic"`
	TargetClusterManifest TargetClusterManifest `copier:"must,nopanic"`
	ServiceAccountRequest      ServiceAccountRequest`copier:"must,nopanic"`
	ServiceAccountTokenRequest ServiceAccountTokenRequest `copier:"must,nopanic"`
	CreateIAM                  CreateIAM                  `copier:"must,nopanic"`
	ListNamespacesRequest      ListNamespacesRequest      `copier:"must,nopanic"`
}
