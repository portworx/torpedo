package cluster

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/backup/utils"
	"reflect"
	"sync"
)

type (
	Request          interface{}
	Response         interface{}
	RequestProcessor func(Request) (Response, error)
)

type RequestManager struct {
	sync.RWMutex
	RequestProcessorMap map[reflect.Type]RequestProcessor
}

func (m *RequestManager) GetRequestProcessorMap() map[reflect.Type]RequestProcessor {
	m.RLock()
	defer m.RUnlock()
	return m.RequestProcessorMap
}

func (m *RequestManager) SetRequestProcessorMap(requestProcessorMap map[reflect.Type]RequestProcessor) {
	m.Lock()
	defer m.Unlock()
	m.RequestProcessorMap = requestProcessorMap
}

func (m *RequestManager) SetRequestProcessor() {

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

func NewRequestManager() {
	//reflect.TypeOf(&driverapi.ScheduleRequest{}): func(req interface{}) (interface{}, error) {
	//	return driverapi.Schedule(req.(*driverapi.ScheduleRequest))
	//},
	//	reflect.TypeOf(&driverapi.WaitForRunningRequest{}): func(req interface{}) (interface{}, error) {
	//	return driverapi.WaitForRunning(req.(*driverapi.WaitForRunningRequest))
	//}
}
