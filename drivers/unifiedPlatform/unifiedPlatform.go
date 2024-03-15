package unifiedPlatform

import (
	"crypto/tls"
	"fmt"
	pdsapi "github.com/portworx/torpedo/drivers/unifiedPlatform/pds/backend/v1/api"
	pdsGrpc "github.com/portworx/torpedo/drivers/unifiedPlatform/pds/backend/v1/grpc"
	platformapi "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/backend/v1/api"
	"github.com/portworx/torpedo/pkg/log"
	accountv1 "github.com/pure-px/platform-api-go-client/platform/v1/account"
	backuplocationv1 "github.com/pure-px/platform-api-go-client/platform/v1/backuplocation"
	cloudCredentialv1 "github.com/pure-px/platform-api-go-client/platform/v1/cloudcredential"
	iamv1 "github.com/pure-px/platform-api-go-client/platform/v1/iam"
	namespacev1 "github.com/pure-px/platform-api-go-client/platform/v1/namespace"
	onboardv1 "github.com/pure-px/platform-api-go-client/platform/v1/onboard"
	projectv1 "github.com/pure-px/platform-api-go-client/platform/v1/project"
	serviceaccountv1 "github.com/pure-px/platform-api-go-client/platform/v1/serviceaccount"
	targetClusterv1 "github.com/pure-px/platform-api-go-client/platform/v1/targetcluster"
	targetClusterManifestv1 "github.com/pure-px/platform-api-go-client/platform/v1/targetclusterregistrationmanifest"
	tenantv1 "github.com/pure-px/platform-api-go-client/platform/v1/tenant"
	whoamiv1 "github.com/pure-px/platform-api-go-client/platform/v1/whoami"
	"os"
	"strconv"
	"strings"

	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/pds"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platform"
	platformGrpc "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/backend/v1/grpc"
	. "github.com/portworx/torpedo/drivers/utilities"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"net/url"
)

const (
	UNIFIED_PLATFORM_INTERFACE = "BACKEND_TYPE"
	REST_API                   = "REST_API"
	GRPC                       = "GRPC"
	GRPC_PORT                  = "443"
)

type UnifiedPlatformComponents struct {
	Platform platform.Platform
	PDS      pds.Pds
}

func NewUnifiedPlatformComponents(controlPlaneURL string, AccountId string) (*UnifiedPlatformComponents, error) {
	VARIABLE_FROM_JENKINS := GetEnv(UNIFIED_PLATFORM_INTERFACE, REST_API)

	switch VARIABLE_FROM_JENKINS {
	case REST_API:
		//generate platform api_v1 client
		platformApiClient, err := GetPlatformRESTClientForAutomation(controlPlaneURL, AccountId)
		if err != nil {
			return nil, err
		}
		//generate pds api_v2 client
		pdsClient, err := GetPDSRESTClientForAutomation(controlPlaneURL, AccountId)
		if err != nil {
			return nil, err
		}

		return &UnifiedPlatformComponents{
			Platform: &platformApiClient,
			PDS:      &pdsClient,
		}, nil
	case GRPC:
		//Trim the controlplane url and add port number to it
		_, grpcUrl, isFound := strings.Cut(controlPlaneURL, "//")
		if !isFound {
			return nil, fmt.Errorf("Unable to parse control plane url\n")
		}
		grpcUrl = grpcUrl + ":" + GRPC_PORT
		log.Infof("Generating grpc client for controlplane [%s]", grpcUrl)

		//generate grpc client
		insecureDialOptStr := os.Getenv("INSECURE_FLAG")
		insecureDialOpt, err := strconv.ParseBool(insecureDialOptStr)
		if err != nil {
			return nil, err
		}

		dialOpts := []grpc.DialOption{}
		if insecureDialOpt {
			dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		} else {
			tlsConfig := &tls.Config{}
			dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
		}
		grpcClient, err := grpc.Dial(grpcUrl, dialOpts...)
		if err != nil {
			return nil, err
		}

		return &UnifiedPlatformComponents{
			Platform: &platformGrpc.PlatformGrpc{
				ApiClientV1: grpcClient,
				AccountId:   AccountId,
			},
			PDS: &pdsGrpc.PdsGrpc{
				ApiClientV2: grpcClient,
				AccountId:   AccountId,
			},
		}, nil
	default:
		//generate platform api_v1 client
		platformApiClient, err := GetPlatformRESTClientForAutomation(controlPlaneURL, AccountId)
		if err != nil {
			return nil, err
		}
		//generate pds api_v2 client
		pdsClient, err := GetPDSRESTClientForAutomation(controlPlaneURL, AccountId)
		if err != nil {
			return nil, err
		}

		return &UnifiedPlatformComponents{
			Platform: &platformApiClient,
			PDS:      &pdsClient,
		}, nil
	}
}

// GetPlatformRESTClientForAutomation returns the platform client for automation
func GetPlatformRESTClientForAutomation(controlPlaneURL string, AccountId string) (platformapi.PLATFORM_API_V1, error) {

	endpointURL, err := url.Parse(controlPlaneURL)
	if err != nil {
		return platformapi.PLATFORM_API_V1{}, err
	}
	log.Infof("Generating REST(V1) client for Platform [%s]", endpointURL)

	// Creating account api client
	accountAPIv1Config := accountv1.NewConfiguration()
	accountAPIv1Config.Host = endpointURL.Host
	accountAPIv1Config.Scheme = endpointURL.Scheme
	accountAPIv1Client := accountv1.NewAPIClient(accountAPIv1Config)

	// Creating tenant api client
	tenantAPIv1Config := tenantv1.NewConfiguration()
	tenantAPIv1Config.Host = endpointURL.Host
	tenantAPIv1Config.Scheme = endpointURL.Scheme
	tenantAPIv1Client := tenantv1.NewAPIClient(tenantAPIv1Config)

	// Creating target cluster api client
	targetClusterAPIv1Config := targetClusterv1.NewConfiguration()
	targetClusterAPIv1Config.Host = endpointURL.Host
	targetClusterAPIv1Config.Scheme = endpointURL.Scheme
	targetClusterAPIv1Client := targetClusterv1.NewAPIClient(targetClusterAPIv1Config)

	// Creating backup location api client
	backupLocationAPIv1Config := backuplocationv1.NewConfiguration()
	backupLocationAPIv1Config.Host = endpointURL.Host
	backupLocationAPIv1Config.Scheme = endpointURL.Scheme
	backupLocationAPIv1Client := backuplocationv1.NewAPIClient(backupLocationAPIv1Config)

	// Creating cloud credential api client
	cloudCredentialAPIv1Config := cloudCredentialv1.NewConfiguration()
	cloudCredentialAPIv1Config.Host = endpointURL.Host
	cloudCredentialAPIv1Config.Scheme = endpointURL.Scheme
	cloudCredentialAPIv1Client := cloudCredentialv1.NewAPIClient(cloudCredentialAPIv1Config)

	// Creating iam api client
	iamAPIv1Config := iamv1.NewConfiguration()
	iamAPIv1Config.Host = endpointURL.Host
	iamAPIv1Config.Scheme = endpointURL.Scheme
	iamAPIv1Client := iamv1.NewAPIClient(iamAPIv1Config)

	// Creating namespace api client
	namespaceAPIv1Config := namespacev1.NewConfiguration()
	namespaceAPIv1Config.Host = endpointURL.Host
	namespaceAPIv1Config.Scheme = endpointURL.Scheme
	namespaceAPIv1Client := namespacev1.NewAPIClient(namespaceAPIv1Config)

	// Creating onboard api client
	onboardAPIv1Config := onboardv1.NewConfiguration()
	onboardAPIv1Config.Host = endpointURL.Host
	onboardAPIv1Config.Scheme = endpointURL.Scheme
	onboardAPIv1Client := onboardv1.NewAPIClient(onboardAPIv1Config)

	// Creating project api client
	projectAPIv1Config := projectv1.NewConfiguration()
	projectAPIv1Config.Host = endpointURL.Host
	projectAPIv1Config.Scheme = endpointURL.Scheme
	projectAPIv1Client := projectv1.NewAPIClient(projectAPIv1Config)

	// Creating target cluster manifest api client
	targetClusterManifestAPIv1Config := targetClusterManifestv1.NewConfiguration()
	targetClusterManifestAPIv1Config.Host = endpointURL.Host
	targetClusterManifestAPIv1Config.Scheme = endpointURL.Scheme
	targetClusterManifestAPIv1Client := targetClusterManifestv1.NewAPIClient(targetClusterManifestAPIv1Config)

	// Creating whoami api client
	whoAmIAPIv1Config := whoamiv1.NewConfiguration()
	whoAmIAPIv1Config.Host = endpointURL.Host
	whoAmIAPIv1Config.Scheme = endpointURL.Scheme
	whoAmIAPIv1Client := whoamiv1.NewAPIClient(whoAmIAPIv1Config)

	// Creating service account api client
	serviceAccountv1Config := serviceaccountv1.NewConfiguration()
	serviceAccountv1Config.Host = endpointURL.Host
	serviceAccountv1Config.Scheme = endpointURL.Scheme
	serviceAccountv1Client := serviceaccountv1.NewAPIClient(serviceAccountv1Config)

	return platformapi.PLATFORM_API_V1{
		AccountV1APIClient:               accountAPIv1Client,
		TenantV1APIClient:                tenantAPIv1Client,
		TargetClusterV1APIClient:         targetClusterAPIv1Client,
		BackupLocationV1APIClient:        backupLocationAPIv1Client,
		CloudCredentialV1APIClient:       cloudCredentialAPIv1Client,
		IamV1APIClient:                   iamAPIv1Client,
		NamespaceV1APIClient:             namespaceAPIv1Client,
		OnboardV1APIClient:               onboardAPIv1Client,
		ProjectV1APIClient:               projectAPIv1Client,
		TargetClusterManifestV1APIClient: targetClusterManifestAPIv1Client,
		WhoamiV1APIClient:                whoAmIAPIv1Client,
		ServiceAccountV1Client:           serviceAccountv1Client,
		AccountID:                        AccountId,
	}, nil
}

// GetPlatformRESTClientForAutomation returns the platform client for automation
func GetPDSRESTClientForAutomation(controlPlaneURL string, AccountId string) (pdsapi.PDSV2_API, error) {

	endpointURL, err := url.Parse(controlPlaneURL)
	if err != nil {
		return pdsapi.PDSV2_API{}, err
	}
	log.Infof("Generating REST(V1) client for PDS [%s]", endpointURL)

	// Creating PDS REST api client
	pdsApiConf := pdsv2.NewConfiguration()
	pdsApiConf.Host = endpointURL.Host
	pdsApiConf.Scheme = endpointURL.Scheme
	pdsV2apiClient := pdsv2.NewAPIClient(pdsApiConf)

	return pdsapi.PDSV2_API{
		ApiClientV2: pdsV2apiClient,
		AccountID:   AccountId,
	}, nil
}
