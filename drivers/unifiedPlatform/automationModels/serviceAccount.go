package automationModels

// PDSServiceAccountRequest struct
type PDSServiceAccountRequest struct {
	Create      CreateServiceAccounts
	Get         GetServiceAccount
	CreateToken CreatePdsServiceAccountToken
	GetToken    GetServiceAccountToken
}

type PDSServiceAccountResponse struct {
	Create V1ServiceAccount
}

type CreateServiceAccounts struct {
	V1ServiceAccount V1ServiceAccount
	TenantId         string
}

type GetServiceAccount struct {
	Meta   Meta
	Config V1Config2
}

type CreatePdsServiceAccountToken struct {
	TenantId                                string
	ServiceAccountServiceGetAccessTokenBody ServiceAccountServiceGetAccessTokenBody
}

type GetServiceAccountToken struct {
	Token string
}

type V1ServiceAccount struct {
	Meta   Meta
	Config Config
}

type ServiceAccountServiceGetAccessTokenBody struct {
	ClientId     *string `json:"clientId,omitempty"`
	ClientSecret *string
}
