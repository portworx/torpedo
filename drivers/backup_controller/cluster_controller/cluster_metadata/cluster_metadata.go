package cluster_metadata

import . "github.com/portworx/torpedo/drivers/backup_controller/cluster_controller/cluster_utils"

const (
	// DefaultConfigPath is the default config-path for the cluster_manager.Cluster
	DefaultConfigPath = GlobalInClusterConfigPath
)

// ClusterMetaData represents the metadata for cluster_manager.Cluster
type ClusterMetaData struct {
	ConfigPath string
}

// GetConfigPath returns the ConfigPath associated with the ClusterMetaData
func (m *ClusterMetaData) GetConfigPath() string {
	return m.ConfigPath
}

// SetConfigPath sets the ConfigPath for the ClusterMetaData
func (m *ClusterMetaData) SetConfigPath(configPath string) *ClusterMetaData {
	m.ConfigPath = configPath
	return m
}

// GetClusterUID returns the cluster_manager.Cluster UID
func (m *ClusterMetaData) GetClusterUID() string {
	return m.GetConfigPath()
}

// NewClusterMetaData creates a new instance of the ClusterMetaData
func NewClusterMetaData(configPath string) *ClusterMetaData {
	newClusterMetaData := &ClusterMetaData{}
	newClusterMetaData.SetConfigPath(configPath)
	return newClusterMetaData
}

// NewDefaultClusterMetaData creates a new instance of the ClusterMetaData with default values
func NewDefaultClusterMetaData() *ClusterMetaData {
	return NewClusterMetaData(DefaultConfigPath)
}

const (
	// DefaultNamespaceName is the default name for the namespace_manager.Namespace
	DefaultNamespaceName = "default"
)

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

const (
	// DefaultPodName is the default name for the pod_by_name_manager.PodByName
	DefaultPodName = "torpedo"
)

// PodByNameMetaData represents the metadata for pod_by_name_manager.PodByName
type PodByNameMetaData struct {
	NamespaceMetaData *NamespaceMetaData
	PodName           string
}

// GetNamespaceMetaData returns the NamespaceMetaData associated with the PodByNameMetaData
func (m *PodByNameMetaData) GetNamespaceMetaData() *NamespaceMetaData {
	return m.NamespaceMetaData
}

// SetNamespaceMetaData sets the NamespaceMetaData for the PodByNameMetaData
func (m *PodByNameMetaData) SetNamespaceMetaData(metaData *NamespaceMetaData) *PodByNameMetaData {
	m.NamespaceMetaData = metaData
	return m
}

// GetPodName returns the PodName associated with the PodByNameMetaData
func (m *PodByNameMetaData) GetPodName() string {
	return m.PodName
}

// SetPodName sets the PodName for the PodByNameMetaData
func (m *PodByNameMetaData) SetPodName(podName string) *PodByNameMetaData {
	m.PodName = podName
	return m
}

// GetPodByNameUID returns the PodByName UID
func (m *PodByNameMetaData) GetPodByNameUID() string {
	return m.GetPodName()
}

// NewPodByNameMetaData creates a new instance of the PodByNameMetaData
func NewPodByNameMetaData(metaData *NamespaceMetaData, podName string) *PodByNameMetaData {
	newPodByNameMetaData := &PodByNameMetaData{}
	newPodByNameMetaData.SetNamespaceMetaData(metaData)
	newPodByNameMetaData.SetPodName(podName)
	return newPodByNameMetaData
}

// NewDefaultPodByNameMetaData creates a new instance of the PodByNameMetaData with default values
func NewDefaultPodByNameMetaData() *PodByNameMetaData {
	return NewPodByNameMetaData(NewDefaultNamespaceMetaData(), DefaultPodName)
}
