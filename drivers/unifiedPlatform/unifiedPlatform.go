package unifiedPlatform

import (
	"net/url"

	"github.com/portworx/torpedo/drivers/unifiedPlatform/platform"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api/api_v1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api/api_v2"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api/grpc"
	platformv2 "github.com/pure-px/platform-api-go-client/v1alpha1"
)

type UnifiedPlatformComponents struct {
	Platform platform.Platform
}

func NewUnifiedPlatformComponents(controlPlaneURL string, AccountId string) (*UnifiedPlatformComponents, error) {
	// VARIABLE_FROM_JENKINS := os.Getenv("TYPEOFINTERFACE")
	VARIABLE_FROM_JENKINS := "grpc"

	switch VARIABLE_FROM_JENKINS {
	case "v1":
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
			Platform: &API_V1{
				ApiClientV2: platformV2apiClient,
			},
		}, nil
	case "v2":
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
			Platform: &API_V2{
				ApiClientV2: platformV2apiClient,
			},
		}, nil
	case "grpc":
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
			Platform: &API_V1{
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
			Platform: &GRPC{
				ApiClientV2: platformV2apiClient,
			},
		}, nil
	}

}
