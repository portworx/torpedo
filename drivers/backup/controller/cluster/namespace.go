package cluster

import (
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

// GetNamespace returns the Namespace associated with the metadata
func (m *NamespaceMetaData) GetNamespace() string {
	return m.Namespace
}

// SetNamespace sets the Namespace string for the metadata
func (m *NamespaceMetaData) SetNamespace(namespace string) {
	m.Namespace = namespace
}

// GetNamespaceUid returns the Namespace uid
func (m *NamespaceMetaData) GetNamespaceUid() string {
	return m.GetNamespace()
}

func NewNamespaceMetaData(namespace string) *NamespaceMetaData {
	newNamespaceMetaData := &NamespaceMetaData{}
	newNamespaceMetaData.SetNamespace(namespace)
	return newNamespaceMetaData
}

// NamespaceConfig represents the configuration for a Namespace
type NamespaceConfig struct {
	ClusterMetaData   *ClusterMetaData
	NamespaceMetaData *NamespaceMetaData
	ClusterController *ClusterController
}

// GetClusterMetaData returns the ClusterMetaData associated with the Namespace configuration
func (c *NamespaceConfig) GetClusterMetaData() *ClusterMetaData {
	return c.ClusterMetaData
}

// SetClusterMetaData sets the ClusterMetaData for the Namespace configuration
func (c *NamespaceConfig) SetClusterMetaData(clusterMetaData *ClusterMetaData) {
	c.ClusterMetaData = clusterMetaData
}

// GetNamespaceMetaData returns the NamespaceMetaData associated with the Namespace configuration
func (c *NamespaceConfig) GetNamespaceMetaData() *NamespaceMetaData {
	return c.NamespaceMetaData
}

// SetNamespaceMetaData sets the NamespaceMetaData for the Namespace configuration
func (c *NamespaceConfig) SetNamespaceMetaData(namespaceMetaData *NamespaceMetaData) {
	c.NamespaceMetaData = namespaceMetaData
}

// GetClusterController returns the ClusterController associated with the Namespace configuration
func (c *NamespaceConfig) GetClusterController() *ClusterController {
	return c.ClusterController
}

// SetClusterController sets the ClusterController for the Namespace configuration
func (c *NamespaceConfig) SetClusterController(clusterController *ClusterController) {
	c.ClusterController = clusterController
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
func (n *Namespace) SetAppManager(appManager *AppManager) {
	n.AppManager = appManager
}

// NewNamespace creates a new instance of the Namespace
func NewNamespace() *Namespace {
	newNamespace := &Namespace{}
	newNamespace.SetAppManager(NewAppManager())
	return newNamespace
}

// NamespaceManager represents a manager for Namespace
type NamespaceManager struct {
	NamespaceMap         map[string]*Namespace
	RemovedNamespacesMap map[string][]*Namespace
}

// GetNamespaceMap returns the NamespaceMap of the NamespaceManager
func (m *NamespaceManager) GetNamespaceMap() map[string]*Namespace {
	return m.NamespaceMap
}

// SetNamespaceMap sets the NamespaceMap of the NamespaceManager
func (m *NamespaceManager) SetNamespaceMap(namespaceMap map[string]*Namespace) {
	m.NamespaceMap = namespaceMap
}

// GetRemovedNamespacesMap returns the RemovedNamespacesMap of the NamespaceManager
func (m *NamespaceManager) GetRemovedNamespacesMap() map[string][]*Namespace {
	return m.RemovedNamespacesMap
}

// SetRemovedNamespacesMap sets the RemovedNamespacesMap of the NamespaceManager
func (m *NamespaceManager) SetRemovedNamespacesMap(removedNamespacesMap map[string][]*Namespace) {
	m.RemovedNamespacesMap = removedNamespacesMap
}

// GetNamespace returns the Namespace with the given Namespace uid
func (m *NamespaceManager) GetNamespace(namespaceUid string) *Namespace {
	return m.GetNamespaceMap()[namespaceUid]
}

// IsNamespacePresent checks if the Namespace with the given Namespace uid is present
func (m *NamespaceManager) IsNamespacePresent(namespaceUid string) bool {
	_, isPresent := m.GetNamespaceMap()[namespaceUid]
	return isPresent
}

// SetNamespace sets the Namespace with the given Namespace uid
func (m *NamespaceManager) SetNamespace(namespaceUid string, namespace *Namespace) {
	m.GetNamespaceMap()[namespaceUid] = namespace
}

// DeleteNamespace deletes the Namespace with the given Namespace uid
func (m *NamespaceManager) DeleteNamespace(namespaceUid string) {
	delete(m.GetNamespaceMap(), namespaceUid)
}

// RemoveNamespace removes the Namespace with the given Namespace uid
func (m *NamespaceManager) RemoveNamespace(namespaceUid string) {
	m.GetRemovedNamespacesMap()[namespaceUid] = append(m.GetRemovedNamespacesMap()[namespaceUid], m.GetNamespace(namespaceUid))
	m.DeleteNamespace(namespaceUid)
}

// NewNamespaceManager creates a new instance of the NamespaceManager
func NewNamespaceManager() *NamespaceManager {
	newNamespaceManager := &NamespaceManager{}
	newNamespaceManager.SetNamespaceMap(make(map[string]*Namespace, 0))
	newNamespaceManager.SetRemovedNamespacesMap(make(map[string][]*Namespace, 0))
	return newNamespaceManager
}
