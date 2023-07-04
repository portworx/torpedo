package cluster_manager

import (
	. "github.com/portworx/torpedo/drivers/backup_controller/cluster_controller/cluster_manager/namespace_manager"
	. "github.com/portworx/torpedo/drivers/backup_controller/cluster_controller/cluster_metadata"
	. "github.com/portworx/torpedo/drivers/backup_controller/entity_manager"
)

// Cluster represents Cluster
type Cluster struct {
	NamespaceManager *EntityManager[Namespace]
}

//// GetNamespaceManager returns the NamespaceManager associated with the Cluster
//func (c *Cluster) GetNamespaceManager() *NamespaceManager {
//	return c.NamespaceManager
//}
//
//// SetNamespaceManager sets the NamespaceManager for the Cluster
//func (c *Cluster) SetNamespaceManager(manager *NamespaceManager) *Cluster {
//	c.NamespaceManager = manager
//	return c
//}

// NewCluster creates a new instance of the Cluster
func NewCluster(namespaceManager *NamespaceManager) *Cluster {
	newCluster := &Cluster{}
	newCluster.SetNamespaceManager(namespaceManager)
	return newCluster
}

// NewDefaultCluster creates a new instance of the Cluster with default values
func NewDefaultCluster() *Cluster {
	return NewCluster(NewDefaultNamespaceManager())
}

// ClusterManager represents a manager for Cluster
type ClusterManager struct {
	ClusterMap         map[string]*Cluster
	RemovedClustersMap map[string][]*Cluster
}

// GetClusterMap returns the ClusterMap of the ClusterManager
func (m *ClusterManager) GetClusterMap() map[string]*Cluster {
	return m.ClusterMap
}

// SetClusterMap sets the ClusterMap of the ClusterManager
func (m *ClusterManager) SetClusterMap(clusterMap map[string]*Cluster) *ClusterManager {
	m.ClusterMap = clusterMap
	return m
}

// GetRemovedClustersMap returns the RemovedClustersMap of the ClusterManager
func (m *ClusterManager) GetRemovedClustersMap() map[string][]*Cluster {
	return m.RemovedClustersMap
}

// SetRemovedClustersMap sets the RemovedClustersMap of the ClusterManager
func (m *ClusterManager) SetRemovedClustersMap(removedClustersMap map[string][]*Cluster) *ClusterManager {
	m.RemovedClustersMap = removedClustersMap
	return m
}

// GetCluster returns the Cluster with the given Cluster UID
func (m *ClusterManager) GetCluster(clusterUID string) *Cluster {
	return m.ClusterMap[clusterUID]
}

// SetCluster sets the Cluster with the given Cluster UID
func (m *ClusterManager) SetCluster(clusterUID string, cluster *Cluster) *ClusterManager {
	m.ClusterMap[clusterUID] = cluster
	return m
}

// DeleteCluster deletes the Cluster with the given Cluster UID
func (m *ClusterManager) DeleteCluster(clusterUID string) *ClusterManager {
	delete(m.ClusterMap, clusterUID)
	return m
}

// RemoveCluster removes the Cluster with the given Cluster UID
func (m *ClusterManager) RemoveCluster(clusterUID string) *ClusterManager {
	if cluster, isPresent := m.ClusterMap[clusterUID]; isPresent {
		m.RemovedClustersMap[clusterUID] = append(m.RemovedClustersMap[clusterUID], cluster)
		delete(m.ClusterMap, clusterUID)
	}
	return m
}

// IsClusterPresent checks if the Cluster with the given Cluster UID is present
func (m *ClusterManager) IsClusterPresent(clusterUID string) bool {
	_, isPresent := m.ClusterMap[clusterUID]
	return isPresent
}

// IsClusterRemoved checks if the Cluster with the given Cluster UID is removed
func (m *ClusterManager) IsClusterRemoved(clusterUID string) bool {
	_, isPresent := m.RemovedClustersMap[clusterUID]
	return isPresent
}

// IsClusterRecorded checks if the Cluster with the given Cluster UID is recorded
func (m *ClusterManager) IsClusterRecorded(clusterUID string) bool {
	return m.IsClusterPresent(clusterUID) || m.IsClusterRemoved(clusterUID)
}

// NewClusterManager creates a new instance of the ClusterManager
func NewClusterManager(clusterMap map[string]*Cluster, removedClustersMap map[string][]*Cluster) *ClusterManager {
	newClusterManager := &ClusterManager{}
	newClusterManager.SetClusterMap(clusterMap)
	newClusterManager.SetRemovedClustersMap(removedClustersMap)
	return newClusterManager
}

// NewDefaultClusterManager creates a new instance of the ClusterManager with default values
func NewDefaultClusterManager() *ClusterManager {
	return NewClusterManager(make(map[string]*Cluster, 0), make(map[string][]*Cluster, 0))
}

// ClusterConfig represents the configuration for a Cluster
type ClusterConfig struct {
	ClusterManager  *ClusterManager
	ClusterMetaData *ClusterMetaData
}

// GetClusterManager returns the ClusterManager associated with the ClusterConfig
func (c *ClusterConfig) GetClusterManager() *ClusterManager {
	return c.ClusterManager
}

// SetClusterManager sets the ClusterManager for the ClusterConfig
func (c *ClusterConfig) SetClusterManager(manager *ClusterManager) *ClusterConfig {
	c.ClusterManager = manager
	return c
}

// GetClusterMetaData returns the ClusterMetaData associated with the ClusterConfig
func (c *ClusterConfig) GetClusterMetaData() *ClusterMetaData {
	return c.ClusterMetaData
}

// SetClusterMetaData sets the ClusterMetaData for the ClusterConfig
func (c *ClusterConfig) SetClusterMetaData(metaData *ClusterMetaData) *ClusterConfig {
	c.ClusterMetaData = metaData
	return c
}

// NewClusterConfig creates a new instance of the ClusterConfig
func NewClusterConfig(manager *ClusterManager, metaData *ClusterMetaData) *ClusterConfig {
	newClusterConfig := &ClusterConfig{}
	newClusterConfig.SetClusterManager(manager)
	newClusterConfig.SetClusterMetaData(metaData)
	return newClusterConfig
}

// NewDefaultClusterConfig creates a new instance of the ClusterConfig with default values
func NewDefaultClusterConfig() *ClusterConfig {
	return NewClusterConfig(NewDefaultClusterManager(), NewDefaultClusterMetaData())
}
