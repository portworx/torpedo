package cluster

import (
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/tests"
)

// App creates a new AppConfig and configures it
func (c *NamespaceConfig) App(appKey string) *AppConfig {
	appMetaData := NewAppMetaData()
	appMetaData.SetAppKey(appKey)
	scheduleAppConfig := NewScheduleAppConfig()
	scheduleOptions := &scheduler.ScheduleOptions{
		AppKeys:            []string{appKey},
		StorageProvisioner: tests.Inst().Provisioner,
		Namespace:          c.GetNamespaceMetaData().GetNamespace(),
		// TODO: Handle non hyper-converged cluster
		Nodes:  nil,
		Labels: nil,
	}
	scheduleAppConfig.SetScheduleOptions(scheduleOptions)
	validateAppConfig := NewValidateAppConfig()
	tearDownAppConfig := NewTearDownAppConfig()
	return &AppConfig{
		ClusterMetaData:   c.ClusterMetaData,
		NamespaceMetaData: c.NamespaceMetaData,
		AppMetaData:       appMetaData,
		ScheduleAppConfig: scheduleAppConfig,
		ValidateAppConfig: validateAppConfig,
		TearDownAppConfig: tearDownAppConfig,
		ClusterController: c.ClusterController,
	}
}
