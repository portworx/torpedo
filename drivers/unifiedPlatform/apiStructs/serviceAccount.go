package apiStructs

// PDSServiceAccountRequest struct
type PDSServiceAccount struct {
	Create      CreateServiceAccounts
	Get         GetServiceAccount
	CreateToken CreatePdsServiceAccountToken
	GetToken    GetServiceAccountToken
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
