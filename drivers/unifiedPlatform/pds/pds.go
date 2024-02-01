package pds

import pdsV2api "github.com/portworx/torpedo/drivers/unifiedPlatform/pds/api"

type Pds struct {
	DataServiceV2                   *pdsV2api.DataServiceV2
	DataServiceVersionsV2           *pdsV2api.DataServiceVersionsV2
	DeploymentV2                    *pdsV2api.DeploymentV2
	DeploymentConfigurationUpdateV2 *pdsV2api.DeploymentConfigurationUpdateV2
	DeploymentEventsV2              *pdsV2api.DeploymentEventsV2
	DeploymentManifestV2            *pdsV2api.DeploymentManifestV2
	ImageV2                         *pdsV2api.ImageV2
}
