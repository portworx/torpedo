package cluster

import (
	"sync"
)

const (
	// GlobalInClusterConfigPath is the config-path of the cluster in which Torpedo and Px-Backup are running
	GlobalInClusterConfigPath = "" // as described in the doc string of the `SetConfig` function in the k8s.go file of the k8s package
)

type Context string

// ContextManager represents a manager for Context
type ContextManager struct {
	sync.RWMutex
	DstConfigPath string
	SrcConfigPath string
}

// GetDstConfigPath returns the DstConfigPath associated with the ContextManager
func (m *ContextManager) GetDstConfigPath() string {
	m.RLock()
	defer m.RUnlock()
	return m.DstConfigPath
}

// SetDstConfigPath sets the DstConfigPath for the ContextManager
func (m *ContextManager) SetDstConfigPath(dstConfigPath string) {
	m.Lock()
	defer m.Unlock()
	m.DstConfigPath = dstConfigPath
}

// GetSrcConfigPath returns the SrcConfigPath associated with the ContextManager
func (m *ContextManager) GetSrcConfigPath() string {
	m.RLock()
	defer m.RUnlock()
	return m.SrcConfigPath
}

// SetSrcConfigPath sets the SrcConfigPath for the ContextManager
func (m *ContextManager) SetSrcConfigPath(srcConfigPath string) {
	m.Lock()
	defer m.Unlock()
	m.SrcConfigPath = srcConfigPath
}

// NewContextManager creates a new instance of the ContextManager
func NewContextManager() *ContextManager {
	newContextManager := &ContextManager{}
	newContextManager.SetDstConfigPath(GlobalInClusterConfigPath)
	newContextManager.SetSrcConfigPath(GlobalInClusterConfigPath)
	return newContextManager
}
