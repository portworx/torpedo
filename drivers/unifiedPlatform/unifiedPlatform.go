package unifiedPlatform

import (
	"crypto/tls"
	"fmt"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	"github.com/portworx/torpedo/pkg/log"
	"os"
	"strconv"
	"strings"

	"github.com/portworx/torpedo/drivers/unifiedPlatform/pds"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/pds/api/api_v1"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platform"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api/api_v1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api/grpc"
	. "github.com/portworx/torpedo/drivers/utilities"
	platformv1 "github.com/pure-px/platform-api-go-client/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"net/url"
)

const (
	UNIFIED_PLATFORM_INTERFACE = "BACKEND_TYPE"
	API_V1                     = "v1"
	GRPC                       = "grpc"
	GRPC_PORT                  = "443"
)

type UnifiedPlatformComponents struct {
	Platform platform.Platform
	PDS      pds.PDS
}

func NewUnifiedPlatformComponents(controlPlaneURL string, AccountId string) (*UnifiedPlatformComponents, error) {
	VARIABLE_FROM_JENKINS := GetEnv(UNIFIED_PLATFORM_INTERFACE, API_V1)

	switch VARIABLE_FROM_JENKINS {
	case API_V1:
		//generate platform api_v1 client
		platformApiConf := platformv1.NewConfiguration()
		endpointURL, err := url.Parse(controlPlaneURL)
		if err != nil {
			return nil, err
		}
		platformApiConf.Host = endpointURL.Host
		platformApiConf.Scheme = endpointURL.Scheme
		platformV2apiClient := platformv1.NewAPIClient(platformApiConf)

		//generate pds api_v2 client
		pdsApiConf := pdsv2.NewConfiguration()
		pdsApiConf.Host = endpointURL.Host
		pdsApiConf.Scheme = endpointURL.Scheme
		pdsV2apiClient := pdsv2.NewAPIClient(pdsApiConf)

		return &UnifiedPlatformComponents{
			Platform: &PLATFORM_API_V1{
				ApiClientV1: platformV2apiClient,
				AccountID:   AccountId,
			},
			PDS: &PDSV2{
				ApiClientV2: pdsV2apiClient,
				AccountID:   AccountId,
			},
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
			Platform: &AccountGrpc{
				ApiClientV1: grpcClient,
			},
		}, nil
	default:
		//generate platform api_v1 client
		platformApiConf := platformv1.NewConfiguration()
		endpointURL, err := url.Parse(controlPlaneURL)
		if err != nil {
			return nil, err
		}
		platformApiConf.Host = endpointURL.Host
		platformApiConf.Scheme = endpointURL.Scheme
		platformV2apiClient := platformv1.NewAPIClient(platformApiConf)

		//generate pds api_v2 client
		pdsApiConf := pdsv2.NewConfiguration()
		pdsApiConf.Host = endpointURL.Host
		pdsApiConf.Scheme = endpointURL.Scheme
		pdsV2apiClient := pdsv2.NewAPIClient(pdsApiConf)

		return &UnifiedPlatformComponents{
			Platform: &PLATFORM_API_V1{
				ApiClientV1: platformV2apiClient,
				AccountID:   AccountId,
			},
			PDS: &PDSV2{
				ApiClientV2: pdsV2apiClient,
				AccountID:   AccountId,
			},
		}, nil
	}
}
