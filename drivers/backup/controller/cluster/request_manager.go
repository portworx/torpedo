package cluster

import (
	"github.com/portworx/torpedo/drivers/backup/controller/cluster/driver/schedulerapi"
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

func NewRequestManager() *RequestManager {
	newRequestManager := &RequestManager{}
	newRequestManager.SetRequestProcessorMap(make(map[RequestType]RequestProcessor, 0))
	newRequestManager.SetRequestProcessor(schedulerapi.NewScheduleRequest(),
		func(request Request) (Response, error) {
			return schedulerapi.Schedule(request.(*schedulerapi.ScheduleRequest))
		},
	)
	return newRequestManager
}
