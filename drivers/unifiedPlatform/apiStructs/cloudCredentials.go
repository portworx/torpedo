package apiStructs

type CloudCredentials struct {
	Create CreateCloudCredentials
	Get    GetCloudCredentials
}

type GetCloudCredentials struct {
	CloudCredentialsId string
	IsConfigRequired   bool
}

type CreateCloudCredentials struct {
	TenantID string
	Meta     Meta
	Config   CloudConfig
}

type Provider struct {
	CloudProvider int32
}

type CloudConfig struct {
	Provider    Provider
	Credentials CCredentials
}

type CCredentials struct {
	S3Credentials
	AzureCredentials
	GcpCredentials
}

type S3Credentials struct {
	AccessKey string
	SecretKey string
}

type AzureCredentials struct {
	AccountName string
	AccountKey  string
}

type GcpCredentials struct {
	ProjectId string
	Key       string
}
