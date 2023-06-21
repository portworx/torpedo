package schedulerapi

import (
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/scheduler/spec"
	"github.com/portworx/torpedo/tests"
)

type ScheduleRequest struct {
	Apps            []*spec.AppSpec
	InstanceID      string
	ScheduleOptions scheduler.ScheduleOptions
}

func (r *ScheduleRequest) GetApps() []*spec.AppSpec {
	return r.Apps
}

func (r *ScheduleRequest) SetApps(apps []*spec.AppSpec) {
	r.Apps = apps
}

func (r *ScheduleRequest) GetInstanceID() string {
	return r.InstanceID
}

func (r *ScheduleRequest) SetInstanceID(instanceID string) {
	r.InstanceID = instanceID
}

func (r *ScheduleRequest) GetScheduleOptions() scheduler.ScheduleOptions {
	return r.ScheduleOptions
}

func (r *ScheduleRequest) SetScheduleOptions(options scheduler.ScheduleOptions) {
	r.ScheduleOptions = options
}

type ScheduleResponse struct {
	Contexts []*scheduler.Context
}

func (r *ScheduleResponse) GetContexts() []*scheduler.Context {
	return r.Contexts
}

func (r *ScheduleResponse) SetContexts(contexts []*scheduler.Context) {
	r.Contexts = contexts
}

func Schedule(request *ScheduleRequest) (*ScheduleResponse, error) {
	response := &ScheduleResponse{}
	contexts, err := tests.Inst().S.ScheduleWithCustomAppSpecs(request.GetApps(), request.GetInstanceID(), request.GetScheduleOptions())
	if err != nil {
		return nil, err
	}
	response.SetContexts(contexts)
	return response, nil
}
