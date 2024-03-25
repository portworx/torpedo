package automationModels

type WorkFlowResponse struct {
	Meta              Meta                        `copier:"must,nopanic"`
	Config            Config                      `copier:"must,nopanic"`
	Id                string                      `copier:"must,nopanic"`
	OnboardAccount    AccountRegistration         `copier:"must,nopanic"`
	Status            Status                      `copier:"must,nopanic"`
	TargetCluster     PlatformTargetClusterOutput `copier:"must,nopanic"`
	PDSDeployment     PDSDeployment               `copier:"must,nopanic"`
	PdsServiceAccount PDSServiceAccount
	Info              V1Info
}

type WorkFlowRequest struct {
	Deployment            PDSDeployment                 `copier:"must,nopanic"`
	BackupConfig          PDSBackupConfig               `copier:"must,nopanic"`
	Backup                PDSBackup                     `copier:"must,nopanic"`
	Restore               PDSRestore                    `copier:"must,nopanic"`
	OnboardAccount        PlatformOnboardAccountRequest `copier:"must,nopanic"`
	PDSApplication        PDSApplicaition               `copier:"must,nopanic"`
	TargetCluster         PlatformTargetClusterRequest  `copier:"must,nopanic"`
	Meta                  Meta                          `copier:"must,nopanic"`
	Config                Config                        `copier:"must,nopanic"`
	Id                    string                        `copier:"must,nopanic"`
	ClusterId             string                        `copier:"must,nopanic"`
	TenantId              string                        `copier:"must,nopanic"`
	PdsAppId              string                        `copier:"must,nopanic"`
	Pagination            PaginationRequest             `copier:"must,nopanic"`
	ServiceAccountRequest PDSServiceAccount             `copier:"must,nopanic"`
	Iam                   PDSIam                        `copier:"must,nopanic"`
	DataServiceId         string                        `copier:"must,nopanic"`
	DataServiceVersionId  string                        `copier:"must,nopanic"`
}
