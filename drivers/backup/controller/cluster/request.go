package cluster

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/backup/controller/cluster/driverapi/schedulerapi"
	"github.com/portworx/torpedo/drivers/backup/utils"
	"reflect"
	"sync"
)

type (
	RequestType      reflect.Type
	Request          interface{}
	Response         interface{}
	RequestProcessor func(Request) (Response, error)
)

type RequestManager struct {
	sync.RWMutex
	RequestProcessorMap map[RequestType]RequestProcessor
}

func (m *RequestManager) GetRequestProcessorMap() map[RequestType]RequestProcessor {
	m.RLock()
	defer m.RUnlock()
	return m.RequestProcessorMap
}

func (m *RequestManager) SetRequestProcessorMap(requestProcessorMap map[RequestType]RequestProcessor) {
	m.Lock()
	defer m.Unlock()
	m.RequestProcessorMap = requestProcessorMap
}

func (m *RequestManager) SetRequestProcessor(request Request, processor RequestProcessor) {
	m.Lock()
	defer m.Unlock()
	m.RequestProcessorMap[reflect.TypeOf(request)] = processor
}

func (m *RequestManager) ProcessRequest(request Request) (response Response, err error) {
	handler, ok := m.GetRequestProcessorMap()[reflect.TypeOf(request)]
	if !ok {
		err = fmt.Errorf("unknown request type")
		return nil, utils.ProcessError(err)
	}
	response, err = handler(request)
	if err != nil {
		return nil, utils.ProcessError(err, utils.StructToString(request))
	}
	return response, nil
}

func NewRequestManager() *RequestManager {
	newRequestManager := NewRequestManager()
	newRequestManager.SetRequestProcessor(
		&schedulerapi.ScheduleRequest{},
		func(request Request) (Response, error) {
			return schedulerapi.Schedule(request.(*schedulerapi.ScheduleRequest))
		},
	)
	newRequestManager.SetRequestProcessor(
		&schedulerapi.WaitForRunningRequest{},
		func(request Request) (Response, error) {
			return schedulerapi.WaitForRunning(request.(*schedulerapi.WaitForRunningRequest))
		},
	)
	return newRequestManager
}
