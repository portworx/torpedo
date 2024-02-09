package unifiedPlatform

import (
	"crypto/tls"
	"os"
	"strconv"

	"github.com/portworx/torpedo/drivers/unifiedPlatform/platform"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api/api_v1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api/api_v2"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api/grpc"
	. "github.com/portworx/torpedo/drivers/utilities"
	platformv2 "github.com/pure-px/platform-api-go-client/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"net/url"
)

const (
	UNIFIED_PLATFORM_INTERFACE = "BACKEND_TYPE"
	API_V1                     = "v1"
	API_V2                     = "v2"
	GRPC                       = "grpc"
)

type UnifiedPlatformComponents struct {
	Platform platform.Platform
}

func NewUnifiedPlatformComponents(controlPlaneURL string, AccountId string) (*UnifiedPlatformComponents, error) {
	VARIABLE_FROM_JENKINS := GetEnv(UNIFIED_PLATFORM_INTERFACE, API_V1)

	switch VARIABLE_FROM_JENKINS {
	case API_V1:
		//generate platform api_v1 client
		platformApiConf := platformv2.NewConfiguration()
		endpointURL, err := url.Parse(controlPlaneURL)
		if err != nil {
			return nil, err
		}
		platformApiConf.Host = endpointURL.Host
		platformApiConf.Scheme = endpointURL.Scheme
		platformV2apiClient := platformv2.NewAPIClient(platformApiConf)
		return &UnifiedPlatformComponents{
			Platform: &PLATFORM_API_V1{
				ApiClientV2: platformV2apiClient,
			},
		}, nil

	case API_V2:
		//generate platform api_v1 client
		platformApiConf := platformv2.NewConfiguration()
		endpointURL, err := url.Parse(controlPlaneURL)
		if err != nil {
			return nil, err
		}
		platformApiConf.Host = endpointURL.Host
		platformApiConf.Scheme = endpointURL.Scheme
		platformV2apiClient := platformv2.NewAPIClient(platformApiConf)
		return &UnifiedPlatformComponents{
			Platform: &PLATFORM_API_V2{
				ApiClientV2: platformV2apiClient,
			},
		}, nil
	case GRPC:
		//generate platform grpc client
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
		grpcClient, err := grpc.Dial(controlPlaneURL, dialOpts...)
		if err != nil {
			return nil, err
		}

		return &UnifiedPlatformComponents{
			Platform: &PLATFORM_GRPC{
				ApiClientV2: grpcClient,
			},
		}, nil
	default:
		//generate platform api_v1 client
		platformApiConf := platformv2.NewConfiguration()
		endpointURL, err := url.Parse(controlPlaneURL)
		if err != nil {
			return nil, err
		}
		platformApiConf.Host = endpointURL.Host
		platformApiConf.Scheme = endpointURL.Scheme
		platformV2apiClient := platformv2.NewAPIClient(platformApiConf)
		return &UnifiedPlatformComponents{
			Platform: &PLATFORM_API_V1{
				ApiClientV2: platformV2apiClient,
			},
		}, nil
	}
}
