package namespace_manager

import (
	"github.com/portworx/torpedo/drivers/backup/controller/cluster_controller/cluster_manager/namespace_manager/app_manager"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/tests"
)

// App creates a new AppConfig and configures it
func (c *NamespaceConfig) App(appKey string) *app_manager.AppConfig {
	appMetaData := app_manager.NewAppMetaData()
	appMetaData.SetAppKey(appKey)
	scheduleAppConfig := app_manager.NewScheduleAppConfig()
	scheduleOptions := &scheduler.ScheduleOptions{
		AppKeys:            []string{appKey},
		StorageProvisioner: tests.Inst().Provisioner,
		Namespace:          c.GetNamespaceMetaData().GetNamespace(),
		// TODO: Handle non hyper-converged cluster
		Nodes:  nil,
		Labels: nil,
	}
	scheduleAppConfig.SetScheduleOptions(scheduleOptions)
	validateAppConfig := app_manager.NewValidateAppConfig()
	tearDownAppConfig := app_manager.NewTearDownAppConfig()
	return &app_manager.AppConfig{
		ClusterMetaData:   c.ClusterMetaData,
		NamespaceMetaData: c.NamespaceMetaData,
		AppMetaData:       appMetaData,
		ScheduleAppConfig: scheduleAppConfig,
		ValidateAppConfig: validateAppConfig,
		TearDownAppConfig: tearDownAppConfig,
		ClusterController: c.ClusterController,
	}
}
