package apiStructs

// PDSServiceAccountRequest struct
type PDSServiceAccount struct {
	V1ServiceAccount V1ServiceAccount
	TenantId         string
}

type V1ServiceAccount struct {
	Meta   Meta
	Config Config
}

type PDSServiceAccountToken struct {
	TenantId                                string
	ServiceAccountServiceGetAccessTokenBody ServiceAccountServiceGetAccessTokenBody
}

type ServiceAccountServiceGetAccessTokenBody struct {
	ClientId     *string `json:"clientId,omitempty"`
	ClientSecret *string
}

type PdsRbacAccessToken struct {
	Token string
}

type PdsServiceAccount struct {
	Meta   Meta
	Config V1Config2
}
