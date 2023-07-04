package cluster_controller

import (
	. "github.com/portworx/torpedo/drivers/backup_controller/cluster_controller/cluster_manager"
	. "github.com/portworx/torpedo/drivers/backup_controller/cluster_controller/cluster_metadata"
)

// Cluster creates a new cluster_manager.ClusterConfig and configures it
func (c *ClusterController) Cluster(configPath string) *ClusterConfig {
	if c == nil {
		return nil
	}
	clusterManager := c.GetClusterManager()
	clusterMetaData := NewClusterMetaData(configPath)
	return NewClusterConfig(clusterManager, clusterMetaData)
}
