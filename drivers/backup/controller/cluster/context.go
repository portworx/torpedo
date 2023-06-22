package cluster

import (
	"github.com/portworx/torpedo/drivers/backup/utils"
	"github.com/portworx/torpedo/tests"
)

// SwitchContext switches Cluster Context
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
