package cluster

import (
	"github.com/portworx/torpedo/drivers/backup/utils"
	"github.com/portworx/torpedo/tests"
	"sync"
)

type ContextManager struct {
	sync.RWMutex
	DstConfigPath string
	SrcConfigPath string
}

func (m *ContextManager) GetDstConfigPath() string {
	m.RLock()
	defer m.RUnlock()
	return m.DstConfigPath
}

func (m *ContextManager) SetDstConfigPath(dstConfigPath string) {
	m.Lock()
	defer m.Unlock()
	m.DstConfigPath = dstConfigPath
}

func (m *ContextManager) GetSrcConfigPath() string {
	m.RLock()
	defer m.RUnlock()
	return m.SrcConfigPath
}

func (m *ContextManager) SetSrcConfigPath(srcConfigPath string) {
	m.Lock()
	defer m.Unlock()
	m.SrcConfigPath = srcConfigPath
}

func (m *ContextManager) SwitchContext() error {
	m.Lock()
	defer m.Unlock()
	currentConfigPath := tests.CurrentClusterConfigPath
	err := utils.SwitchClusterContext(m.DstConfigPath)
	if err != nil {
		debugStruct := struct {
			DstConfigPath string
		}{
			DstConfigPath: m.DstConfigPath,
		}
		return utils.ProcessError(err, utils.StructToString(debugStruct))
	}
	m.SetSrcConfigPath(currentConfigPath)
	return nil
}

func NewContextManager() *ContextManager {
	newContextManager := &ContextManager{}
	newContextManager.SetDstConfigPath(GlobalInClusterConfigPath)
	newContextManager.SetSrcConfigPath(GlobalInClusterConfigPath)
	return newContextManager
}
