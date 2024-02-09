package unifiedPlatform

import (
	"crypto/tls"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platform"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api/api_v1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api/grpc"
	platformv2 "github.com/pure-px/platform-api-go-client/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"os"

	"net/url"
)

type UnifiedPlatformComponents struct {
	Platform platform.Platform
}

func NewUnifiedPlatformComponents(controlPlaneURL string, AccountId string) (*UnifiedPlatformComponents, error) {
	VARIABLE_FROM_JENKINS := os.Getenv("TYPE_OF_INTERFACE")

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
	case "grpc":
		//generate platform grpc client
		insecureDialOpt := true
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
			Platform: &GRPC{
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
			Platform: &API_V1{
				ApiClientV2: platformV2apiClient,
			},
		}, nil
	}
}
