package pdsrestore

import (
	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	pdsapi "github.com/portworx/torpedo/drivers/pds/api"
	tc "github.com/portworx/torpedo/drivers/pds/targetcluster"
)

type RestoreClient struct {
	controlPlaneURL      string
	Components           *pdsapi.Components
	Deployment           pds.ModelsDeployment
	RestoreTargetCluster tc.TargetCluster
}

type DSEntity struct {
	Deployment        pds.ModelsDeployment
	ApplicationConfig pds.ModelsApplicationConfigurationTemplate
	ResourceConfig    pds.ModelsResourceSettingsTemplate
	StorageTemplate   pds.ModelsStorageOptionsTemplate
	DataHash          string
}

func (restoreClient *RestoreClient) TriggerAndValidateRestore(validate bool) error {
	restoreClient
}
