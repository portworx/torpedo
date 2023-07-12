package cluster_config

import (
	. "github.com/portworx/torpedo/drivers/pxbackup/controller/cl_controller/cluster_manager/cluster"
)

// Register registers Cluster with the given ClusterSpec and UID
func (c *ClusterConfig) Register(clusterUID string) error {
	c.GetClusterManager().Set(clusterUID, NewDefaultCluster(c.GetClusterSpec()))
	return nil
}
