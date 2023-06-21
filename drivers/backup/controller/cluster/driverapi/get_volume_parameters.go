package driverapi

import (
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/tests"
)

type VolumeParameters map[string]map[string]string

type GetVolumeParametersRequest struct {
	Context *scheduler.Context
}

func (r *GetVolumeParametersRequest) GetContext() *scheduler.Context {
	return r.Context
}

func (r *GetVolumeParametersRequest) SetContext(context *scheduler.Context) {
	r.Context = context
}

type GetVolumeParametersResponse struct {
	VolumeParameters VolumeParameters
}

func (r *GetVolumeParametersResponse) GetVolumeParameters() VolumeParameters {
	return r.VolumeParameters
}

func (r *GetVolumeParametersResponse) SetVolumeParameters(parameters VolumeParameters) {
	r.VolumeParameters = parameters
}

func GetVolumeParameters(request *GetVolumeParametersRequest) (*GetVolumeParametersResponse, error) {
	response := &GetVolumeParametersResponse{}
	volumeParameters, err := tests.Inst().S.GetVolumeParameters(request.GetContext())
	if err != nil {
		return nil, err
	}
	response.SetVolumeParameters(volumeParameters)
	return response, nil
}
