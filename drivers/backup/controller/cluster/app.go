package cluster

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/backup/controller/cluster/driver/schedulerapi"
	"github.com/portworx/torpedo/drivers/backup/utils"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/scheduler/spec"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/tests"
	"strings"
)

// CanSchedule checks if App CanSchedule
func (c *AppConfig) CanSchedule() error {
	if !c.GetClusterController().GetClusterManager().IsClusterPresent(c.GetClusterMetaData().GetClusterUid()) {
		err := fmt.Errorf("cluster [%s] is not registered", c.GetClusterMetaData().GetConfigPath())
		return utils.ProcessError(err)
	}
	return nil
}

// Schedule schedules App
func (c *AppConfig) Schedule() error {
	err := c.CanSchedule()
	if err != nil {
		return utils.ProcessError(err)
	}
	appSpec, err := utils.GetAppSpec(c.GetAppMetaData().GetAppKey())
	if err != nil {
		debugStruct := struct {
			AppKey string
		}{
			AppKey: c.GetAppMetaData().GetAppKey(),
		}
		return utils.ProcessError(err, utils.StructToString(debugStruct))
	}
	cluster := c.GetClusterController().GetClusterManager().GetCluster(c.GetClusterMetaData().GetClusterUid())
	scheduleRequest := schedulerapi.NewScheduleRequest()
	scheduleRequest.SetApps([]*spec.AppSpec{appSpec})
	scheduleRequest.SetInstanceID(c.GetScheduleAppConfig().GetInstanceID())
	scheduleRequest.SetScheduleOptions(*c.GetScheduleAppConfig().GetScheduleOptions())
	log.Infof("Scheduling app [%s] on namespace [%s]", c.GetAppMetaData().GetApp(), c.GetNamespaceMetaData().GetNamespace())
	resp, err := cluster.ProcessClusterRequest(scheduleRequest)
	if err != nil {
		debugStruct := struct {
			ScheduleRequest *schedulerapi.ScheduleRequest
		}{
			ScheduleRequest: scheduleRequest,
		}
		return utils.ProcessError(err, utils.StructToString(debugStruct))
	}
	if !cluster.GetNamespaceManager().IsNamespacePresent(c.GetNamespaceMetaData().GetNamespaceUid()) {
		cluster.GetNamespaceManager().SetNamespace(c.GetNamespaceMetaData().GetNamespaceUid(), NewNamespace())
	}
	app := NewApp()
	scheduleResponse := resp.(*schedulerapi.ScheduleResponse)
	app.SetContexts(scheduleResponse.GetContexts())
	cluster.GetNamespaceManager().GetNamespace(c.GetNamespaceMetaData().GetNamespaceUid()).GetAppManager().SetApp(c.GetAppMetaData().GetAppUid(), app)
	return nil
}

// Validate validates App
func (c *AppConfig) Validate() error {
	var errors []error
	err := c.CanSchedule()
	if err != nil {
		return utils.ProcessError(err)
	}
	cluster := c.GetClusterController().GetClusterManager().GetCluster(c.GetClusterMetaData().GetClusterUid())
	app := cluster.GetNamespaceManager().GetNamespace(c.GetNamespaceMetaData().GetNamespaceUid()).GetAppManager().GetApp(c.GetAppMetaData().GetAppUid())
	if app == nil {
		return utils.ProcessError(fmt.Errorf("app [%s] in namespace [%s] of cluster [%s] is not recorded", c.AppMetaData.AppKey, c.NamespaceMetaData.Namespace, c.ClusterMetaData.ConfigPath))
	}
	errChan := make(chan error, len(app.Contexts))
	for _, ctx := range app.Contexts {
		tests.ValidateContext(ctx, &errChan)
	}
	for err = range errChan {
		errors = append(errors, err)
	}
	errStrings := make([]string, 0)
	for _, err = range errors {
		if err != nil {
			errStrings = append(errStrings, err.Error())
		}
	}
	if len(errStrings) != 0 {
		return fmt.Errorf(strings.Join(errStrings, " "))
	}
	return nil
}

// TearDown tears down App
func (c *AppConfig) TearDown() error {
	err := c.CanSchedule()
	if err != nil {
		return utils.ProcessError(err)
	}
	cluster := c.GetClusterController().GetClusterManager().GetCluster(c.GetClusterMetaData().GetClusterUid())
	app := cluster.GetNamespaceManager().GetNamespace(c.GetNamespaceMetaData().GetNamespaceUid()).GetAppManager().GetApp(c.GetAppMetaData().GetAppUid())
	if app == nil {
		return utils.ProcessError(fmt.Errorf("app [%s] in namespace [%s] of cluster [%s] is not recorded", c.AppMetaData.AppKey, c.NamespaceMetaData.Namespace, c.ClusterMetaData.ConfigPath))
	}
	for _, ctx := range app.Contexts {
		tests.TearDownContext(ctx, map[string]bool{
			tests.SkipClusterScopedObjects:              c.GetTearDownAppConfig().GetSkipClusterScopedObjects(),
			scheduler.OptionsWaitForResourceLeakCleanup: c.GetTearDownAppConfig().GetWaitForResourceLeakCleanup(),
			scheduler.OptionsWaitForDestroy:             c.GetTearDownAppConfig().GetWaitForDestroy(),
		})
	}
	return nil
}
