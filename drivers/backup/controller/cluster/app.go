package cluster

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/backup/controller/cluster/driver/schedulerapi"
	"github.com/portworx/torpedo/drivers/backup/utils"
	"github.com/portworx/torpedo/drivers/scheduler/spec"
	"github.com/portworx/torpedo/pkg/log"
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
