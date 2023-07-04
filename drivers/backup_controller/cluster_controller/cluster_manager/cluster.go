package cluster_manager

import (
	. "github.com/portworx/torpedo/drivers/backup_controller/cluster_controller/cluster_manager/namespace_manager"
	. "github.com/portworx/torpedo/drivers/backup_controller/cluster_controller/cluster_metadata"
)

// Namespace creates a new namespace_manager.NamespaceConfig and configures it
func (c *ClusterConfig) Namespace(namespaceName string) *NamespaceConfig {
	if c == nil || c.GetClusterMetaData() == nil {
		return nil
	}
	clusterUID := c.GetClusterMetaData().GetClusterUID()
	namespaceMetaData := NewNamespaceMetaData(c.GetClusterMetaData(), namespaceName)
	if c.GetClusterManager() == nil || c.GetClusterManager().GetCluster(clusterUID) == nil {
		return NewNamespaceConfig(nil, namespaceMetaData)
	}
	namespaceManager := c.GetClusterManager().GetCluster(clusterUID).GetNamespaceManager()
	return NewNamespaceConfig(namespaceManager, namespaceMetaData)
}
