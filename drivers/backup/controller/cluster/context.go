package cluster

import (
	"github.com/portworx/torpedo/drivers/backup/utils"
	"github.com/portworx/torpedo/tests"
)

type ContextManager struct {
	DstConfigPath string
	SrcConfigPath string
}

func (m *ContextManager) GetDstConfigPath() string {
	return m.DstConfigPath
}

func (m *ContextManager) SetDstConfigPath(dstConfigPath string) {
	m.DstConfigPath = dstConfigPath
}

func (m *ContextManager) GetSrcConfigPath() string {
	return m.SrcConfigPath
}

func (m *ContextManager) SetSrcConfigPath(srcConfigPath string) {
	m.SrcConfigPath = srcConfigPath
}

func (m *ContextManager) SwitchContext() error {
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
