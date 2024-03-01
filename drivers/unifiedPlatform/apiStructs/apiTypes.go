package apiStructs

type WorkFlowResponse struct {
	Meta          Meta          `copier:"must,nopanic"`
	Config        Config        `copier:"must,nopanic"`
	Id            string        `copier:"must,nopanic"`
	Status        Status        `copier:"must,nopanic"`
	V1AccessToken V1AccessToken `copier:"must,nopanic"`
}

type WorkFlowRequest struct {
	Deployment                 PDSDeployment              `copier:"must,nopanic"`
	Meta                       Meta                       `copier:"must,nopanic"`
	Config                     Config                     `copier:"must,nopanic"`
	Id                         string                     `copier:"must,nopanic"`
	ClusterId                  string                     `copier:"must,nopanic"`
	TenantId                   string                     `copier:"must,nopanic"`
	PdsAppId                   string                     `copier:"must,nopanic"`
	ServiceAccountRequest      ServiceAccountRequest      `copier:"must,nopanic"`
	ServiceAccountTokenRequest ServiceAccountTokenRequest `copier:"must,nopanic"`
	CreateIAM                  CreateIAM                  `copier:"must,nopanic"`
	V1RoleBinding              V1RoleBinding              `copier:"must,nopanic"`
	V1AccessPolicy             V1AccessPolicy             `copier:"must,nopanic"`
	V1IAM                      V1IAM                      `copier:"must,nopanic"`
	Pagination                 PaginationRequest
	TargetClusterManifest      TargetClusterManifest
}
