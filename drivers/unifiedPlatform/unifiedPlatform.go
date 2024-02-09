package unifiedPlatform

import (
	"net/url"

	"github.com/portworx/torpedo/drivers/unifiedPlatform/platform"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api/api_v1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api/api_v2"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api/grpc"
	. "github.com/portworx/torpedo/drivers/utilities"
	platformv2 "github.com/pure-px/platform-api-go-client/v1alpha1"
)

const (
	UNIFIED_PLATFORM_INTERFACE = "UNIFIED_PLATFORM_INTERFACE"
	API_V1                     = "v1"
	API_V2                     = "v2"
	GRPC                       = "grpc"
)

type UnifiedPlatformComponents struct {
	Platform platform.Platform
}

func NewUnifiedPlatformComponents(controlPlaneURL string, AccountId string) (*UnifiedPlatformComponents, error) {
	// Check the API version to be used during test, fallback to API_v1 if not specified
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
		//generate platform api_v2 client
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
		platformApiConf := platformv2.NewConfiguration()
		endpointURL, err := url.Parse(controlPlaneURL)
		if err != nil {
			return nil, err
		}
		platformApiConf.Host = endpointURL.Host
		platformApiConf.Scheme = endpointURL.Scheme
		platformV2apiClient := platformv2.NewAPIClient(platformApiConf)
		return &UnifiedPlatformComponents{
			Platform: &PLATFORM_GRPC{
				ApiClientV2: platformV2apiClient,
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
