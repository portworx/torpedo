package cluster

import (
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/scheduler/spec"
	"github.com/portworx/torpedo/tests"
)

type AppScheduleRequest struct {
	Apps            []*spec.AppSpec
	InstanceID      string
	ScheduleOptions scheduler.ScheduleOptions
}

type AppScheduleResponse struct {
	Contexts []*scheduler.Context
}

func ScheduleApp(request *AppScheduleRequest) (*AppScheduleResponse, error) {
	contexts, err := tests.Inst().S.ScheduleWithCustomAppSpecs(request.Apps, request.InstanceID, request.ScheduleOptions)
	if err != nil {
		return nil, err
	}
	return &AppScheduleResponse{
		Contexts: contexts,
	}, nil
}

type AppValidateRequest struct {
}

type AppValidateResponse struct {
}

type AppTearDownRequest struct {
}

type AppTearDownResponse struct {
}
