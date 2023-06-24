package request_manager

import (
	"reflect"
)

type (
	RequestType      reflect.Type
	Request          interface{}
	Response         interface{}
	RequestProcessor func(Request) (Response, error)
)

type RequestManager struct {
	RequestProcessorMap map[RequestType]RequestProcessor
}

func (m *RequestManager) GetRequestProcessorMap() map[RequestType]RequestProcessor {
	return m.RequestProcessorMap
}

func (m *RequestManager) SetRequestProcessorMap(requestProcessorMap map[RequestType]RequestProcessor) {
	m.RequestProcessorMap = requestProcessorMap
}

func (m *RequestManager) SetRequestProcessor(request Request, processor RequestProcessor) {
	m.RequestProcessorMap[reflect.TypeOf(request)] = processor
}

func NewRequestManager() *RequestManager {
	newRequestManager := &RequestManager{}
	newRequestManager.SetRequestProcessorMap(make(map[RequestType]RequestProcessor, 0))
	return newRequestManager
}
