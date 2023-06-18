package cluster

import (
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/tests"
)

type NamespaceConfig struct {
	ClusterMetaData   *ClusterMetaData
	NamespaceMetaData *NamespaceMetaData
	ClusterController *ClusterController
}

func (c *NamespaceConfig) App(appKey string, identifier ...string) *AppConfig {
	return &AppConfig{
		ClusterMetaData:   c.ClusterMetaData,
		NamespaceMetaData: c.NamespaceMetaData,
		AppMetaData:       NewAppMetaData(appKey, identifier...),
		ScheduleAppConfig: &ScheduleAppConfig{
			ScheduleOptions: &scheduler.ScheduleOptions{
				AppKeys:            []string{appKey},
				StorageProvisioner: tests.Inst().Provisioner,
				Namespace:          c.NamespaceMetaData.GetName(),
				// ToDo: handle non hyper-converged cluster
				Nodes:  nil,
				Labels: nil,
			},
			InstanceID: tests.Inst().InstanceID,
		},
		ValidateAppConfig: &ValidateAppConfig{
			WaitForRunningTimeout:       DefaultWaitForRunningTimeout,
			WaitForRunningRetryInterval: DefaultWaitForRunningRetryInterval,
			ValidateVolumeTimeout:       DefaultValidateVolumeTimeout,
			ValidateVolumeRetryInterval: DefaultValidateVolumeRetryInterval,
		},
		TearDownAppConfig: &TearDownAppConfig{
			WaitForDestroy:             DefaultWaitForDestroy,
			WaitForResourceLeakCleanup: DefaultWaitForResourceLeakCleanup,
			SkipClusterScopedObjects:   DefaultSkipClusterScopedObjects,
		},
		ClusterController: c.ClusterController,
	}
}
