package cluster_config

import (
	. "github.com/portworx/torpedo/drivers/pxbackup/controller/cl_controller/cluster_manager/cluster"
	. "github.com/portworx/torpedo/drivers/pxbackup/controller/cl_controller/cluster_spec"
	. "github.com/portworx/torpedo/drivers/pxbackup/controller_utils/entity/entity_config/entity_manager"
)

// ClusterConfig represents the configuration for a ClusterSpec
type ClusterConfig struct {
	ClusterSpec    *ClusterSpec
	ClusterManager *EntityManager[*Cluster]
}

// GetClusterSpec returns the ClusterSpec associated with the ClusterConfig
func (c *ClusterConfig) GetClusterSpec() *ClusterSpec {
	return c.ClusterSpec
}

// SetClusterSpec sets the ClusterSpec for the ClusterConfig
func (c *ClusterConfig) SetClusterSpec(spec *ClusterSpec) *ClusterConfig {
	c.ClusterSpec = spec
	return c
}

// GetClusterManager returns the ClusterManager associated with the ClusterConfig
func (c *ClusterConfig) GetClusterManager() *EntityManager[*Cluster] {
	return c.ClusterManager
}

// SetClusterManager sets the ClusterManager for the ClusterConfig
func (c *ClusterConfig) SetClusterManager(manager *EntityManager[*Cluster]) *ClusterConfig {
	c.ClusterManager = manager
	return c
}

// NewClusterConfig creates a new instance of the ClusterConfig
func NewClusterConfig(clusterSpec *ClusterSpec, clusterManager *EntityManager[*Cluster]) *ClusterConfig {
	clusterConfig := &ClusterConfig{}
	clusterConfig.SetClusterSpec(clusterSpec)
	clusterConfig.SetClusterManager(clusterManager)
	return clusterConfig
}
