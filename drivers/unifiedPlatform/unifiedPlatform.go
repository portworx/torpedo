package unifiedPlatform

import (
	pdsV2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platform"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api/api_v1"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api/api_v2"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/platform/api/grpc"
	platformv2 "github.com/pure-px/platform-api-go-client/v1alpha1"
	"google.golang.org/grpc"
	"os"
)

type UnifiedPlatformComponents struct {
	Platform platform.Platform
}

func NewUnifiedPlatformComponents(platformApiClient *platformv2.APIClient, pdsApiClient *pdsV2.APIClient, grpcClient *grpc.ClientConn, AccountId string) *UnifiedPlatformComponents {
	VARIABLE_FROM_JENKINS := os.Getenv("TYPEOFINTERFACE")

	switch VARIABLE_FROM_JENKINS {
	case "v1":
		return &UnifiedPlatformComponents{
			Platform: &API_V1{
				ApiClientV2: platformApiClient,
			},
		}
	case "v2":
		return &UnifiedPlatformComponents{
			Platform: &API_V2{
				ApiClientV2: platformApiClient,
			},
		}
	case "grpc":
		//get the grpc client as an argument and initialize the GRPC struct
		return &UnifiedPlatformComponents{
			Platform: &GRPC{
				ApiClientV2: grpcClient,
			},
		}
	default:
		return &UnifiedPlatformComponents{
			Platform: &API_V1{
				ApiClientV2: platformApiClient,
			},
		}
	}

}
