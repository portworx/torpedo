package api

import (
	"context"
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	accountv1 "github.com/pure-px/platform-api-go-client/v1/account"
	backuplocationv1 "github.com/pure-px/platform-api-go-client/v1/backuplocation"
	cloudCredentialv1 "github.com/pure-px/platform-api-go-client/v1/cloudcredential"
	iamv1 "github.com/pure-px/platform-api-go-client/v1/iam"
	namespacev1 "github.com/pure-px/platform-api-go-client/v1/namespace"
	onboardv1 "github.com/pure-px/platform-api-go-client/v1/onboard"
	projectv1 "github.com/pure-px/platform-api-go-client/v1/project"
	serviceaccountv1 "github.com/pure-px/platform-api-go-client/v1/serviceaccount"
	targetClusterv1 "github.com/pure-px/platform-api-go-client/v1/targetcluster"
	targetClusterManifestv1 "github.com/pure-px/platform-api-go-client/v1/targetclusterregistrationmanifest"
	tenantv1 "github.com/pure-px/platform-api-go-client/v1/tenant"
	whoamiv1 "github.com/pure-px/platform-api-go-client/v1/whoami"
)

type PLATFORM_API_V1 struct {
	AccountV1APIClient               *accountv1.APIClient
	TenantV1APIClient                *tenantv1.APIClient
	TargetClusterV1APIClient         *targetClusterv1.APIClient
	BackupLocationV1APIClient        *backuplocationv1.APIClient
	CloudCredentialV1APIClient       *cloudCredentialv1.APIClient
	IamV1APIClient                   *iamv1.APIClient
	NamespaceV1APIClient             *namespacev1.APIClient
	OnboardV1APIClient               *onboardv1.APIClient
	ProjectV1APIClient               *projectv1.APIClient
	TargetClusterManifestV1APIClient *targetClusterManifestv1.APIClient
	WhoamiV1APIClient                *whoamiv1.APIClient
	ServiceAccountV1Client           *serviceaccountv1.APIClient
	AccountID                        string
}

// GetClient updates the header with bearer token and returns the new client
func (account *PLATFORM_API_V1) getAccountClient() (context.Context, *accountv1.AccountServiceAPIService, error) {
	log.Infof("Creating client from PLATFORM_API_V1 package")
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	account.AccountV1APIClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	account.AccountV1APIClient.GetConfig().DefaultHeader["px-account-id"] = account.AccountID

	client := account.AccountV1APIClient.AccountServiceAPI
	return ctx, client, nil
}

// GetAppClient updates the header with bearer token and returns the new client
func (applications *PLATFORM_API_V1) getTenantAppClient() (context.Context, *tenantv1.ApplicationServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	applications.TenantV1APIClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	applications.TenantV1APIClient.GetConfig().DefaultHeader["px-account-id"] = applications.AccountID
	client := applications.TenantV1APIClient.ApplicationServiceAPI

	return ctx, client, nil
}

// GetAppClient updates the header with bearer token and returns the new client
func (applications *PLATFORM_API_V1) getClusterAppClient() (context.Context, *targetClusterv1.ApplicationServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	applications.TargetClusterV1APIClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	applications.TargetClusterV1APIClient.GetConfig().DefaultHeader["px-account-id"] = applications.AccountID
	client := applications.TargetClusterV1APIClient.ApplicationServiceAPI

	return ctx, client, nil
}

// GetBackupLocClient updates the header with bearer token and returbackuploc the new client
func (backuploc *PLATFORM_API_V1) getBackupLocClient() (context.Context, *backuplocationv1.BackupLocationServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	backuploc.BackupLocationV1APIClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	backuploc.BackupLocationV1APIClient.GetConfig().DefaultHeader["px-account-id"] = backuploc.AccountID
	client := backuploc.BackupLocationV1APIClient.BackupLocationServiceAPI

	return ctx, client, nil
}

// GetCloudCredentialClient updates the header with bearer token and return cloudCreds the new client
func (cloudCred *PLATFORM_API_V1) getCloudCredentialClient() (context.Context, *cloudCredentialv1.CloudCredentialServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	cloudCred.CloudCredentialV1APIClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	cloudCred.CloudCredentialV1APIClient.GetConfig().DefaultHeader["px-account-id"] = cloudCred.AccountID
	client := cloudCred.CloudCredentialV1APIClient.CloudCredentialServiceAPI

	return ctx, client, nil
}

// GetIamClient updates the header with bearer token and returns the  client
func (iam *PLATFORM_API_V1) getIAMClient() (context.Context, *iamv1.IAMServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	iam.IamV1APIClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	iam.IamV1APIClient.GetConfig().DefaultHeader["px-account-id"] = iam.AccountID
	client := iam.IamV1APIClient.IAMServiceAPI

	return ctx, client, nil
}

// GetNamespaceClient updates the header with bearer token and returns the new client
func (ns *PLATFORM_API_V1) getNamespaceClient() (context.Context, *namespacev1.NamespaceServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	ns.NamespaceV1APIClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	ns.NamespaceV1APIClient.GetConfig().DefaultHeader["px-account-id"] = ns.AccountID
	client := ns.NamespaceV1APIClient.NamespaceServiceAPI
	return ctx, client, nil
}

// GetClient updates the header with bearer token and returns the new client
func (onboard *PLATFORM_API_V1) getOnboardClient() (context.Context, *onboardv1.OnboardServiceAPIService, error) {
	log.Infof("Creating client from PLATFORM_API_V1 package")
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	onboard.OnboardV1APIClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	onboard.OnboardV1APIClient.GetConfig().DefaultHeader["px-account-id"] = onboard.AccountID

	client := onboard.OnboardV1APIClient.OnboardServiceAPI
	return ctx, client, nil
}

// GetClient updates the header with bearer token and returns the new client
func (project *PLATFORM_API_V1) getProjectClient() (context.Context, *projectv1.ProjectServiceAPIService, error) {
	log.Infof("Creating client from PLATFORM_API_V1 package")
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	project.ProjectV1APIClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	project.ProjectV1APIClient.GetConfig().DefaultHeader["px-account-id"] = project.AccountID

	client := project.ProjectV1APIClient.ProjectServiceAPI
	return ctx, client, nil
}

// GetClient updates the header with bearer token and returns the new client
func (tc *PLATFORM_API_V1) getTargetClusterClient() (context.Context, *targetClusterv1.TargetClusterServiceAPIService, error) {
	log.Infof("Creating client from PLATFORM_API_V1 package")
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	tc.TargetClusterV1APIClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	tc.TargetClusterV1APIClient.GetConfig().DefaultHeader["px-account-id"] = tc.AccountID

	client := tc.TargetClusterV1APIClient.TargetClusterServiceAPI
	return ctx, client, nil
}

// GetClient updates the header with bearer token and returns the new client
func (tcManifest *PLATFORM_API_V1) getTargetClusterManifestClient() (context.Context, *targetClusterManifestv1.TargetClusterRegistrationManifestServiceAPIService, error) {
	log.Infof("Creating client from PLATFORM_API_V1 package")
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	tcManifest.TargetClusterManifestV1APIClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	tcManifest.TargetClusterManifestV1APIClient.GetConfig().DefaultHeader["px-account-id"] = tcManifest.AccountID

	client := tcManifest.TargetClusterManifestV1APIClient.TargetClusterRegistrationManifestServiceAPI
	return ctx, client, nil
}

// GetClient updates the header with bearer token and returns the new client
func (tenant *PLATFORM_API_V1) getTenantClient() (context.Context, *tenantv1.TenantServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	tenant.TenantV1APIClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	tenant.TenantV1APIClient.GetConfig().DefaultHeader["px-account-id"] = tenant.AccountID
	client := tenant.TenantV1APIClient.TenantServiceAPI

	return ctx, client, nil
}

// GetClient updates the header with bearer token and returns the new client
func (whoAmI *PLATFORM_API_V1) getWhoAmIClient() (context.Context, *whoamiv1.WhoAmIServiceAPIService, error) {
	log.Infof("Creating client from PLATFORM_API_V1 package")
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}

	whoAmI.WhoamiV1APIClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	whoAmI.WhoamiV1APIClient.GetConfig().DefaultHeader["px-account-id"] = whoAmI.AccountID

	client := whoAmI.WhoamiV1APIClient.WhoAmIServiceAPI
	return ctx, client, nil
}

// getSAClient updates the header with bearer token and returns the  client
func (sa *PLATFORM_API_V1) getSAClient() (context.Context, *serviceaccountv1.ServiceAccountServiceAPIService, error) {
	ctx, token, err := GetBearerToken()
	if err != nil {
		return nil, nil, fmt.Errorf("Error in getting bearer token: %v\n", err)
	}
	sa.ServiceAccountV1Client.GetConfig().DefaultHeader["Authorization"] = "Bearer " + token
	sa.ServiceAccountV1Client.GetConfig().DefaultHeader["px-account-id"] = sa.AccountID
	client := sa.ServiceAccountV1Client.ServiceAccountServiceAPI

	return ctx, client, nil
}