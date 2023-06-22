package cluster

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/backup/utils"
	"reflect"
)

func (m *RequestManager) ProcessRequest(request Request) (response Response, err error) {
	handler, ok := m.RequestProcessorMap[reflect.TypeOf(request)]
	if !ok {
		err = fmt.Errorf("unknown cluster request type [%T]", reflect.TypeOf(request))
		return nil, utils.ProcessError(err)
	}
	response, err = handler(request)
	if err != nil {
		return nil, utils.ProcessError(err, utils.StructToString(request))
	}
	return response, nil
}
