package automationModels

type CloudCredentialsRequest struct {
	Create CloudCredentials    `copier:"must,nopanic"`
	Get    GetCloudCredentials `copier:"must,nopanic"`
	Update CloudCredentials    `copier:"must,nopanic"`
}

type CloudCredentialsResponse struct {
	Create CloudCredentials     `copier:"must,nopanic"`
	Get    GetCloudCredentials  `copier:"must,nopanic"`
	List   ListCloudCredentials `copier:"must,nopanic"`
	Update CloudCredentials     `copier:"must,nopanic"`
}

type ListCloudCredentials struct {
	CloudCredentials []CloudCredentials `copier:"must,nopanic"`
}

type GetCloudCredentials struct {
	CloudCredentialsId string `copier:"must,nopanic"`
	IsConfigRequired   bool   `copier:"must,nopanic"`
}

type CloudCredentials struct {
	TenantID string      `copier:"must,nopanic"`
	Meta     Meta        `copier:"must,nopanic"`
	Config   CloudConfig `copier:"must,nopanic"`
}

type Provider struct {
	CloudProvider int32  `copier:"must,nopanic"`
	Name          string `copier:"must,nopanic"`
}

type CloudConfig struct {
	Provider          Provider         `copier:"must,nopanic"`
	S3Credentials     S3Credentials    `copier:"must,nopanic"`
	AzureCredentials  AzureCredentials `copier:"must,nopanic"`
	GoogleCredentials GcpCredentials   `copier:"must,nopanic"`
}

type S3Credentials struct {
	AccessKey string `copier:"must,nopanic"`
	SecretKey string `copier:"must,nopanic"`
}

type AzureCredentials struct {
	AccountName string `copier:"must,nopanic"`
	AccountKey  string `copier:"must,nopanic"`
}

type GcpCredentials struct {
	ProjectId string `copier:"must,nopanic"`
	Key       string `copier:"must,nopanic"`
}
