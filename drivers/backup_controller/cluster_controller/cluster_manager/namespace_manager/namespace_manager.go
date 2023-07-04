package namespace_manager

import (
	. "github.com/portworx/torpedo/drivers/backup_controller/cluster_controller/cluster_manager/namespace_manager/pod_by_name_manager"
	. "github.com/portworx/torpedo/drivers/backup_controller/cluster_controller/cluster_metadata"
)

// Namespace represents Namespace
type Namespace struct {
	PodByNameManager *PodByNameManager
}

// GetPodByNameManager returns the PodByNameManager associated with the Namespace
func (n *Namespace) GetPodByNameManager() *PodByNameManager {
	return n.PodByNameManager
}

// SetPodByNameManager sets the PodByNameManager for the Namespace
func (n *Namespace) SetPodByNameManager(manager *PodByNameManager) *Namespace {
	n.PodByNameManager = manager
	return n
}

// NewNamespace creates a new instance of the Namespace
func NewNamespace(podByNameManager *PodByNameManager) *Namespace {
	newNamespace := &Namespace{}
	newNamespace.SetPodByNameManager(podByNameManager)
	return newNamespace
}

// NewDefaultNamespace creates a new instance of the Namespace with default values
func NewDefaultNamespace() *Namespace {
	return NewNamespace(NewDefaultPodByNameManager())
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
func (m *NamespaceManager) SetNamespaceMap(namespaceMap map[string]*Namespace) *NamespaceManager {
	m.NamespaceMap = namespaceMap
	return m
}

// GetRemovedNamespacesMap returns the RemovedNamespacesMap of the NamespaceManager
func (m *NamespaceManager) GetRemovedNamespacesMap() map[string][]*Namespace {
	return m.RemovedNamespacesMap
}

// SetRemovedNamespacesMap sets the RemovedNamespacesMap of the NamespaceManager
func (m *NamespaceManager) SetRemovedNamespacesMap(removedNamespacesMap map[string][]*Namespace) *NamespaceManager {
	m.RemovedNamespacesMap = removedNamespacesMap
	return m
}

// GetNamespace returns the Namespace with the given Namespace UID
func (m *NamespaceManager) GetNamespace(namespaceUID string) *Namespace {
	return m.NamespaceMap[namespaceUID]
}

// SetNamespace sets the Namespace with the given Namespace UID
func (m *NamespaceManager) SetNamespace(namespaceUID string, namespace *Namespace) *NamespaceManager {
	m.NamespaceMap[namespaceUID] = namespace
	return m
}

// DeleteNamespace deletes the Namespace with the given Namespace UID
func (m *NamespaceManager) DeleteNamespace(namespaceUID string) *NamespaceManager {
	delete(m.NamespaceMap, namespaceUID)
	return m
}

// RemoveNamespace removes the Namespace with the given Namespace UID
func (m *NamespaceManager) RemoveNamespace(namespaceUID string) *NamespaceManager {
	if namespace, isPresent := m.NamespaceMap[namespaceUID]; isPresent {
		m.RemovedNamespacesMap[namespaceUID] = append(m.RemovedNamespacesMap[namespaceUID], namespace)
		delete(m.NamespaceMap, namespaceUID)
	}
	return m
}

// IsNamespacePresent checks if the Namespace with the given Namespace UID is present
func (m *NamespaceManager) IsNamespacePresent(namespaceUID string) bool {
	_, isPresent := m.NamespaceMap[namespaceUID]
	return isPresent
}

// IsNamespaceRemoved checks if the Namespace with the given Namespace UID is removed
func (m *NamespaceManager) IsNamespaceRemoved(namespaceUID string) bool {
	_, isPresent := m.RemovedNamespacesMap[namespaceUID]
	return isPresent
}

// IsNamespaceRecorded checks if the Namespace with the given Namespace UID is recorded
func (m *NamespaceManager) IsNamespaceRecorded(namespaceUID string) bool {
	return m.IsNamespacePresent(namespaceUID) || m.IsNamespaceRemoved(namespaceUID)
}

// NewNamespaceManager creates a new instance of the NamespaceManager
func NewNamespaceManager(namespaceMap map[string]*Namespace, removedNamespacesMap map[string][]*Namespace) *NamespaceManager {
	newNamespaceManager := &NamespaceManager{}
	newNamespaceManager.SetNamespaceMap(namespaceMap)
	newNamespaceManager.SetRemovedNamespacesMap(removedNamespacesMap)
	return newNamespaceManager
}

// NewDefaultNamespaceManager creates a new instance of the NamespaceManager with default values
func NewDefaultNamespaceManager() *NamespaceManager {
	return NewNamespaceManager(make(map[string]*Namespace, 0), make(map[string][]*Namespace, 0))
}

// NamespaceConfig represents the configuration for a Namespace
type NamespaceConfig struct {
	NamespaceManager  *NamespaceManager
	NamespaceMetaData *NamespaceMetaData
}

// GetNamespaceManager returns the NamespaceManager associated with the NamespaceConfig
func (c *NamespaceConfig) GetNamespaceManager() *NamespaceManager {
	return c.NamespaceManager
}

// SetNamespaceManager sets the NamespaceManager for the NamespaceConfig
func (c *NamespaceConfig) SetNamespaceManager(manager *NamespaceManager) *NamespaceConfig {
	c.NamespaceManager = manager
	return c
}

// GetNamespaceMetaData returns the NamespaceMetaData associated with the NamespaceConfig
func (c *NamespaceConfig) GetNamespaceMetaData() *NamespaceMetaData {
	return c.NamespaceMetaData
}

// SetNamespaceMetaData sets the NamespaceMetaData for the NamespaceConfig
func (c *NamespaceConfig) SetNamespaceMetaData(metaData *NamespaceMetaData) *NamespaceConfig {
	c.NamespaceMetaData = metaData
	return c
}

// NewNamespaceConfig creates a new instance of the NamespaceConfig
func NewNamespaceConfig(manager *NamespaceManager, metaData *NamespaceMetaData) *NamespaceConfig {
	newNamespaceConfig := &NamespaceConfig{}
	newNamespaceConfig.SetNamespaceManager(manager)
	newNamespaceConfig.SetNamespaceMetaData(metaData)
	return newNamespaceConfig
}

// NewDefaultNamespaceConfig creates a new instance of the NamespaceConfig with default values
func NewDefaultNamespaceConfig() *NamespaceConfig {
	newNamespaceConfig := &NamespaceConfig{}
	newNamespaceConfig.SetNamespaceManager(NewDefaultNamespaceManager())
	newNamespaceConfig.SetNamespaceMetaData(NewDefaultNamespaceMetaData())
	return newNamespaceConfig
}
