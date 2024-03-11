package apiStructs

type WorkFlowResponse struct {
	Meta                      Meta   `copier:"must,nopanic"`
	Config                    Config `copier:"must,nopanic"`
	Id                        string `copier:"must,nopanic"`
	Status                    Status
	CloudCredentials          CreateCloudCredentials
	PdsServiceAccountResponse PdsServiceAccount
	PdsRbacAccessToken        PdsRbacAccessToken
}

type WorkFlowRequest struct {
	Deployment                 PDSDeployment          `copier:"must,nopanic"`
	BackupConfig               PDSBackupConfig        `copier:"must,nopanic"`
	Backup                     PDSBackupConfig        `copier:"must,nopanic"`
	Restore                    PDSRestore             `copier:"must,nopanic"`
	CloudCredentials           CloudCredentials       `copier:"must,nopanic"`
	Meta                       Meta                   `copier:"must,nopanic"`
	Config                     Config                 `copier:"must,nopanic"`
	Id                         string                 `copier:"must,nopanic"`
	ClusterId                  string                 `copier:"must,nopanic"`
	TenantId                   string                 `copier:"must,nopanic"`
	PdsAppId                   string                 `copier:"must,nopanic"`
	Pagination                 PaginationRequest      `copier:"must,nopanic"`
	TargetClusterManifest      TargetClusterManifest  `copier:"must,nopanic"`
	ServiceAccountRequest      PDSServiceAccount      `copier:"must,nopanic"`
	ServiceAccountTokenRequest PDSServiceAccountToken `copier:"must,nopanic"`
	CreateIAM                  PDSCreateIAM           `copier:"must,nopanic"`
	ListNamespacesRequest      PDSListNamespaces      `copier:"must,nopanic"`
}
