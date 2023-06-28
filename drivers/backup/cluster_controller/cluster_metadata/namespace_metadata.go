package cluster_metadata

import . "github.com/portworx/torpedo/drivers/backup/cluster_controller/cluster_utils"

// NamespaceMetaData represents the metadata for namespace_manager.Namespace
type NamespaceMetaData struct {
	ClusterMetaData *ClusterMetaData
	NamespaceName   string
}

// GetClusterMetaData returns the ClusterMetaData associated with the NamespaceMetaData
func (m *NamespaceMetaData) GetClusterMetaData() *ClusterMetaData {
	return m.ClusterMetaData
}

// SetClusterMetaData sets the ClusterMetaData for the NamespaceMetaData
func (m *NamespaceMetaData) SetClusterMetaData(metaData *ClusterMetaData) *NamespaceMetaData {
	m.ClusterMetaData = metaData
	return m
}

// GetNamespaceName returns the NamespaceName associated with the NamespaceMetaData
func (m *NamespaceMetaData) GetNamespaceName() string {
	return m.NamespaceName
}

// SetNamespaceName sets the NamespaceName for the NamespaceMetaData
func (m *NamespaceMetaData) SetNamespaceName(namespaceName string) *NamespaceMetaData {
	m.NamespaceName = namespaceName
	return m
}

// GetNamespaceUID returns the namespace_manager.Namespace UID
func (m *NamespaceMetaData) GetNamespaceUID() string {
	return m.GetNamespaceName()
}

// NewNamespaceMetaData creates a new instance of the NamespaceMetaData
func NewNamespaceMetaData(metaData *ClusterMetaData, namespaceName string) *NamespaceMetaData {
	newNamespaceMetaData := &NamespaceMetaData{}
	newNamespaceMetaData.SetClusterMetaData(metaData)
	newNamespaceMetaData.SetNamespaceName(namespaceName)
	return newNamespaceMetaData
}

// NewDefaultNamespaceMetaData creates a new instance of the NamespaceMetaData with default values
func NewDefaultNamespaceMetaData() *NamespaceMetaData {
	return NewNamespaceMetaData(NewDefaultClusterMetaData(), DefaultNamespaceName)
}
