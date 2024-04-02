package automationModels

// PDSServiceAccountRequest struct
type PDSServiceAccountRequest struct {
	Create      CreateServiceAccounts
	Get         GetServiceAccount
	CreateToken CreatePdsServiceAccountToken
	GetToken    GetServiceAccountTokenRequest
}

type PDSServiceAccountResponse struct {
	Create          V1ServiceAccount
	List            []V1ServiceAccount
	Get             V1ServiceAccount
	Update          V1ServiceAccount
	RegenerateToken GetServiceAccountTokenResponse
	GetToken        GetServiceAccountTokenResponse
}

type CreateServiceAccounts struct {
	V1ServiceAccount V1ServiceAccount
	TenantId         string
}

type GetServiceAccount struct {
	Meta     Meta
	Config   V1Config2
	Id       string
	TenantId string
}

type CreatePdsServiceAccountToken struct {
	TenantId                                string
	ServiceAccountServiceGetAccessTokenBody ServiceAccountServiceGetAccessTokenBody
}

type GetServiceAccountTokenResponse struct {
	Token string
}

type GetServiceAccountTokenRequest struct {
	Token    string
	TenantId string
}

type V1ServiceAccount struct {
	Meta   Meta
	Config Config
}

type ServiceAccountServiceGetAccessTokenBody struct {
	ClientId     *string `json:"clientId,omitempty"`
	ClientSecret *string
}
