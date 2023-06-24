package schedulerapi

import (
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/scheduler/spec"
	"github.com/portworx/torpedo/tests"
)

// ScheduleRequest represents a ScheduleRequest
type ScheduleRequest struct {
	Apps            []*spec.AppSpec
	InstanceID      string
	ScheduleOptions scheduler.ScheduleOptions
}

// GetApps returns the Apps associated with the ScheduleRequest
func (r *ScheduleRequest) GetApps() []*spec.AppSpec {
	return r.Apps
}

// SetApps sets the Apps for the ScheduleRequest
func (r *ScheduleRequest) SetApps(apps []*spec.AppSpec) {
	r.Apps = apps
}

// GetInstanceID returns the InstanceID associated with the ScheduleRequest
func (r *ScheduleRequest) GetInstanceID() string {
	return r.InstanceID
}

// SetInstanceID sets the InstanceID for the ScheduleRequest
func (r *ScheduleRequest) SetInstanceID(instanceID string) {
	r.InstanceID = instanceID
}

// GetScheduleOptions returns the ScheduleOptions associated with the ScheduleRequest
func (r *ScheduleRequest) GetScheduleOptions() scheduler.ScheduleOptions {
	return r.ScheduleOptions
}

// SetScheduleOptions sets the ScheduleOptions for the ScheduleRequest
func (r *ScheduleRequest) SetScheduleOptions(options scheduler.ScheduleOptions) {
	r.ScheduleOptions = options
}

// NewScheduleRequest creates a new instance of the ScheduleRequest
func NewScheduleRequest() *ScheduleRequest {
	newScheduleRequest := &ScheduleRequest{}
	newScheduleRequest.SetApps(make([]*spec.AppSpec, 0))
	newScheduleRequest.SetInstanceID("")
	newScheduleRequest.SetScheduleOptions(scheduler.ScheduleOptions{})
	return newScheduleRequest
}

// ScheduleResponse represents a ScheduleResponse
type ScheduleResponse struct {
	Contexts []*scheduler.Context
}

// GetContexts returns the Contexts associated with the ScheduleResponse
func (r *ScheduleResponse) GetContexts() []*scheduler.Context {
	return r.Contexts
}

// SetContexts sets the Contexts for the ScheduleResponse
func (r *ScheduleResponse) SetContexts(contexts []*scheduler.Context) {
	r.Contexts = contexts
}

// NewScheduleResponse creates a new instance of the ScheduleResponse
func NewScheduleResponse() *ScheduleResponse {
	newScheduleResponse := &ScheduleResponse{}
	newScheduleResponse.SetContexts(make([]*scheduler.Context, 0))
	return newScheduleResponse
}

// Schedule schedules cluster.App
func Schedule(request *ScheduleRequest) (*ScheduleResponse, error) {
	response := NewScheduleResponse()
	contexts, err := tests.Inst().S.ScheduleWithCustomAppSpecs(request.GetApps(), request.GetInstanceID(), request.GetScheduleOptions())
	if err != nil {
		return nil, err
	}
	response.SetContexts(contexts)
	return response, nil
}
