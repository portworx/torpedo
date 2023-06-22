package cluster

import (
	"sync"
	"time"
)

const (
	// DefaultWaitForRunningTimeout indicates the duration to wait for an app to reach the running state
	DefaultWaitForRunningTimeout = 10 * time.Minute
	// DefaultWaitForRunningRetryInterval indicates the interval between retries when waiting for an app to reach the running state
	DefaultWaitForRunningRetryInterval = 10 * time.Second
	// DefaultValidateVolumeTimeout indicates the duration to wait for volume validation of an app
	DefaultValidateVolumeTimeout = 10 * time.Minute
	// DefaultValidateVolumeRetryInterval indicates the interval between retries when performing volume validation of an app
	DefaultValidateVolumeRetryInterval = 10 * time.Second
)

const (
	// DefaultWaitForDestroy indicates whether to wait for resources to be destroyed during the teardown process
	DefaultWaitForDestroy = true // indicates the value of the `scheduler.OptionsWaitForDestroy` key
	// DefaultWaitForResourceLeakCleanup indicates whether to wait for resource leak cleanup during the teardown process
	DefaultWaitForResourceLeakCleanup = true // indicates the value of the `scheduler.OptionsWaitForResourceLeakCleanup` key
	// DefaultSkipClusterScopedObjects indicates whether to skip cluster-scoped objects during the teardown process
	DefaultSkipClusterScopedObjects = false // indicates the value of the `SkipClusterScopedObject` field in the `scheduler.Context`
)

// NamespaceMetaData represents the metadata for a Namespace
type NamespaceMetaData struct {
	Namespace string
}

// GetNamespace returns the Namespace associated with the NamespaceMetaData
func (m *NamespaceMetaData) GetNamespace() string {
	return m.Namespace
}

// SetNamespace sets the Namespace string for the NamespaceMetaData
func (m *NamespaceMetaData) SetNamespace(namespace string) {
	m.Namespace = namespace
}

// GetNamespaceUid returns the Namespace uid
func (m *NamespaceMetaData) GetNamespaceUid() string {
	return m.GetNamespace()
}

// NewNamespaceMetaData creates a new instance of the NamespaceMetaData
func NewNamespaceMetaData() *NamespaceMetaData {
	newNamespaceMetaData := &NamespaceMetaData{}
	newNamespaceMetaData.SetNamespace("")
	return newNamespaceMetaData
}

// NamespaceConfig represents the configuration for a Namespace
type NamespaceConfig struct {
	ClusterMetaData   *ClusterMetaData
	NamespaceMetaData *NamespaceMetaData
	ClusterController *ClusterController
}

// GetClusterMetaData returns the ClusterMetaData associated with the NamespaceConfig
func (c *NamespaceConfig) GetClusterMetaData() *ClusterMetaData {
	return c.ClusterMetaData
}

// SetClusterMetaData sets the ClusterMetaData for the NamespaceConfig
func (c *NamespaceConfig) SetClusterMetaData(clusterMetaData *ClusterMetaData) {
	c.ClusterMetaData = clusterMetaData
}

// GetNamespaceMetaData returns the NamespaceMetaData associated with the NamespaceConfig
func (c *NamespaceConfig) GetNamespaceMetaData() *NamespaceMetaData {
	return c.NamespaceMetaData
}

// SetNamespaceMetaData sets the NamespaceMetaData for the NamespaceConfig
func (c *NamespaceConfig) SetNamespaceMetaData(namespaceMetaData *NamespaceMetaData) {
	c.NamespaceMetaData = namespaceMetaData
}

// GetClusterController returns the ClusterController associated with the NamespaceConfig
func (c *NamespaceConfig) GetClusterController() *ClusterController {
	return c.ClusterController
}

// SetClusterController sets the ClusterController for the NamespaceConfig
func (c *NamespaceConfig) SetClusterController(clusterController *ClusterController) {
	c.ClusterController = clusterController
}

// NewNamespaceConfig creates a new instance of NamespaceConfig
func NewNamespaceConfig() *NamespaceConfig {
	newNamespaceConfig := &NamespaceConfig{}
	clusterMetaData := NewClusterMetaData()
	newNamespaceConfig.SetClusterMetaData(clusterMetaData)
	namespaceMetaData := NewNamespaceMetaData()
	newNamespaceConfig.SetNamespaceMetaData(namespaceMetaData)
	newNamespaceConfig.SetClusterController(nil)
	return newNamespaceConfig
}

// Namespace represents a Namespace
type Namespace struct {
	AppManager *AppManager
}

// GetAppManager returns the AppManager associated with the Namespace
func (n *Namespace) GetAppManager() *AppManager {
	return n.AppManager
}

// SetAppManager sets the AppManager for the Namespace
func (n *Namespace) SetAppManager(manager *AppManager) {
	n.AppManager = manager
}

// NewNamespace creates a new instance of the Namespace
func NewNamespace() *Namespace {
	newNamespace := &Namespace{}
	newNamespace.SetAppManager(NewAppManager())
	return newNamespace
}

// NamespaceManager represents a manager for Namespace
type NamespaceManager struct {
	sync.RWMutex
	NamespaceMap         map[string]*Namespace
	RemovedNamespacesMap map[string][]*Namespace
}

// GetNamespaceMap returns the NamespaceMap of the NamespaceManager
func (m *NamespaceManager) GetNamespaceMap() map[string]*Namespace {
	m.RLock()
	defer m.RUnlock()
	return m.NamespaceMap
}

// SetNamespaceMap sets the NamespaceMap of the NamespaceManager
func (m *NamespaceManager) SetNamespaceMap(namespaceMap map[string]*Namespace) {
	m.Lock()
	defer m.Unlock()
	m.NamespaceMap = namespaceMap
}

// GetRemovedNamespacesMap returns the RemovedNamespacesMap of the NamespaceManager
func (m *NamespaceManager) GetRemovedNamespacesMap() map[string][]*Namespace {
	m.RLock()
	defer m.RUnlock()
	return m.RemovedNamespacesMap
}

// SetRemovedNamespacesMap sets the RemovedNamespacesMap of the NamespaceManager
func (m *NamespaceManager) SetRemovedNamespacesMap(removedNamespacesMap map[string][]*Namespace) {
	m.Lock()
	defer m.Unlock()
	m.RemovedNamespacesMap = removedNamespacesMap
}

// GetNamespace returns the Namespace with the given Namespace uid
func (m *NamespaceManager) GetNamespace(namespaceUid string) *Namespace {
	m.RLock()
	defer m.RUnlock()
	return m.NamespaceMap[namespaceUid]
}

// IsNamespacePresent checks if the Namespace with the given Namespace uid is present
func (m *NamespaceManager) IsNamespacePresent(namespaceUid string) bool {
	m.RLock()
	defer m.RUnlock()
	_, isPresent := m.NamespaceMap[namespaceUid]
	return isPresent
}

// SetNamespace sets the Namespace with the given Namespace uid
func (m *NamespaceManager) SetNamespace(namespaceUid string, namespace *Namespace) {
	m.Lock()
	defer m.Unlock()
	m.NamespaceMap[namespaceUid] = namespace
}

// DeleteNamespace deletes the Namespace with the given Namespace uid
func (m *NamespaceManager) DeleteNamespace(namespaceUid string) {
	m.Lock()
	defer m.Unlock()
	delete(m.NamespaceMap, namespaceUid)
}

// RemoveNamespace removes the Namespace with the given Namespace uid
func (m *NamespaceManager) RemoveNamespace(namespaceUid string) {
	m.Lock()
	defer m.Unlock()
	m.RemovedNamespacesMap[namespaceUid] = append(m.RemovedNamespacesMap[namespaceUid], m.NamespaceMap[namespaceUid])
	m.DeleteNamespace(namespaceUid)
}

// NewNamespaceManager creates a new instance of the NamespaceManager
func NewNamespaceManager() *NamespaceManager {
	newNamespaceManager := &NamespaceManager{}
	newNamespaceManager.SetNamespaceMap(make(map[string]*Namespace, 0))
	newNamespaceManager.SetRemovedNamespacesMap(make(map[string][]*Namespace, 0))
	return newNamespaceManager
}
