package automationModels

type CloudCredentialsRequest struct {
	Create CloudCredentials
	Get    GetCloudCredentials
	Update CloudCredentials
}

type CloudCredentialsResponse struct {
	Create CloudCredentials
	Get    GetCloudCredentials
	List   ListCloudCredentials
	Update CloudCredentials
}

type ListCloudCredentials struct {
	CloudCredentials []CloudCredentials
}

type GetCloudCredentials struct {
	CloudCredentialsId string
	IsConfigRequired   bool
}

type CloudCredentials struct {
	TenantID string
	Meta     Meta
	Config   CloudConfig
}

type Provider struct {
	CloudProvider int32
	Name          string
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
