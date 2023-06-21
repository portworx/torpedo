package schedulerapi

import (
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/tests"
	"time"
)

type ValidateVolumesRequest struct {
	Context       *scheduler.Context
	Timeout       time.Duration
	RetryInterval time.Duration
	VolumeOptions *scheduler.VolumeOptions
}

func (r *ValidateVolumesRequest) GetContext() *scheduler.Context {
	return r.Context
}

func (r *ValidateVolumesRequest) SetContext(context *scheduler.Context) {
	r.Context = context
}

func (r *ValidateVolumesRequest) GetTimeout() time.Duration {
	return r.Timeout
}

func (r *ValidateVolumesRequest) SetTimeout(timeout time.Duration) {
	r.Timeout = timeout
}

func (r *ValidateVolumesRequest) GetRetryInterval() time.Duration {
	return r.RetryInterval
}

func (r *ValidateVolumesRequest) SetRetryInterval(retryInterval time.Duration) {
	r.RetryInterval = retryInterval
}

func (r *ValidateVolumesRequest) GetVolumeOptions() *scheduler.VolumeOptions {
	return r.VolumeOptions
}

func (r *ValidateVolumesRequest) SetVolumeOptions(options *scheduler.VolumeOptions) {
	r.VolumeOptions = options
}

type ValidateVolumesResponse struct{}

func ValidateVolumes(request *ValidateVolumesRequest) (*ValidateVolumesResponse, error) {
	response := &ValidateVolumesResponse{}
	err := tests.Inst().S.ValidateVolumes(request.GetContext(), request.GetTimeout(), request.GetRetryInterval(), request.GetVolumeOptions())
	if err != nil {
		return nil, err
	}
	return response, nil
}
