package driverapi

import (
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/tests"
	"time"
)

type WaitForRunningRequest struct {
	Context       *scheduler.Context
	Timeout       time.Duration
	RetryInterval time.Duration
}

func (r *WaitForRunningRequest) GetContext() *scheduler.Context {
	return r.Context
}

func (r *WaitForRunningRequest) SetContext(context *scheduler.Context) {
	r.Context = context
}

func (r *WaitForRunningRequest) GetTimeout() time.Duration {
	return r.Timeout
}

func (r *WaitForRunningRequest) SetTimeout(timeout time.Duration) {
	r.Timeout = timeout
}

func (r *WaitForRunningRequest) GetRetryInterval() time.Duration {
	return r.RetryInterval
}

func (r *WaitForRunningRequest) SetRetryInterval(retryInterval time.Duration) {
	r.RetryInterval = retryInterval
}

type WaitForRunningResponse struct{}

func WaitForRunning(request *WaitForRunningRequest) (*WaitForRunningResponse, error) {
	response := &WaitForRunningResponse{}
	err := tests.Inst().S.WaitForRunning(request.GetContext(), request.GetTimeout(), request.GetRetryInterval())
	if err != nil {
		return nil, err
	}
	return response, nil
}
