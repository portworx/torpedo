package automationModels

import "time"

// PDSServiceAccountRequest struct
type PDSServiceAccountRequest struct {
	Create      CreateServiceAccounts         `copier:"must,nopanic"`
	Get         GetServiceAccount             `copier:"must,nopanic"`
	CreateToken CreatePdsServiceAccountToken  `copier:"must,nopanic"`
	GetToken    GetServiceAccountTokenRequest `copier:"must,nopanic"`
}

type PDSServiceAccountResponse struct {
	Create          V1ServiceAccount               `copier:"must,nopanic"`
	List            []V1ServiceAccount             `copier:"must,nopanic"`
	Get             V1ServiceAccount               `copier:"must,nopanic"`
	Update          V1ServiceAccount               `copier:"must,nopanic"`
	RegenerateToken GetServiceAccountTokenResponse `copier:"must,nopanic"`
	CreateToken     GetServiceAccountTokenResponse `copier:"must,nopanic"`
}

type CreateServiceAccounts struct {
	V1ServiceAccount V1ServiceAccount `copier:"must,nopanic"`
	TenantId         string           `copier:"must,nopanic"`
}

type GetServiceAccount struct {
	Meta     Meta      `copier:"must,nopanic"`
	Config   V1Config2 `copier:"must,nopanic"`
	Id       string    `copier:"must,nopanic"`
	TenantId string    `copier:"must,nopanic"`
}

type CreatePdsServiceAccountToken struct {
	TenantId                                string                                  `copier:"must,nopanic"`
	ServiceAccountServiceGetAccessTokenBody ServiceAccountServiceGetAccessTokenBody `copier:"must,nopanic"`
}

type GetServiceAccountTokenResponse struct {
	Token string `copier:"must,nopanic"`
}

type GetServiceAccountTokenRequest struct {
	Token    string `copier:"must,nopanic"`
	TenantId string `copier:"must,nopanic"`
}

type V1ServiceAccount struct {
	Meta   Meta                   `copier:"must,nopanic"`
	Config ServiceAccountV1Config `copier:"must,nopanic"`
	Status Serviceaccountv1Status `copier:"must,nopanic"`
}

type ServiceAccountServiceGetAccessTokenBody struct {
	ClientId     *string `copier:"must,nopanic"`
	ClientSecret *string `copier:"must,nopanic"`
}

type ServiceAccountV1Config struct {
	ClientId     *string `copier:"must,nopanic"`
	ClientSecret *string `copier:"must,nopanic"`
	Disabled     *bool   `copier:"must,nopanic"`
}

type Serviceaccountv1Status struct {
	SecretGenerationCount *int32     `copier:"must,nopanic"`
	LastSecretUpdateTime  *time.Time `copier:"must,nopanic"`
}
