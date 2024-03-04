package apiStructs

type ServiceAccountRequest struct {
	V1ServiceAccount V1ServiceAccount
	TenantId         string
}

type V1ServiceAccount struct {
	Meta   Meta
	Config Config
}

type ServiceAccountTokenRequest struct {
	TenantId                                string
	ServiceAccountServiceGetAccessTokenBody ServiceAccountServiceGetAccessTokenBody
}

type ServiceAccountServiceGetAccessTokenBody struct {
	ClientId     *string `json:"clientId,omitempty"`
	ClientSecret *string
}

type V1AccessToken struct {
	Token string
}
