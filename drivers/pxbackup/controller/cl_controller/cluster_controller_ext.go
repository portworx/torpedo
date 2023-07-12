package cl_controller

import (
	. "github.com/portworx/torpedo/drivers/pxbackup/controller/cl_controller/cluster_manager/cluster"
	. "github.com/portworx/torpedo/drivers/pxbackup/controller/cl_controller/cluster_manager/cluster_config"
	. "github.com/portworx/torpedo/drivers/pxbackup/controller/cl_controller/cluster_spec"
)

// ClusterSpec creates a new cluster_config.ClusterConfig and configures it
func (c *ClusterController) ClusterSpec(configPath string) *ClusterConfig {
	return NewClusterConfig(NewDefaultClusterSpec(configPath), c.ClusterManager)
}

// Cluster returns the Cluster with the given Cluster UID
func (c *ClusterController) Cluster(clusterUID string) *Cluster {
	return c.GetClusterManager().Get(clusterUID)
}
