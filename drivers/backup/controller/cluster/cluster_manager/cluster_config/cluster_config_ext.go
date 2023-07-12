package cluster_config

import (
	. "github.com/portworx/torpedo/drivers/backup/controller/cluster/cluster_manager/cluster"
)

// SetConfigPath sets the ConfigPath for the ClusterConfig
func (c *ClusterConfig) SetConfigPath(configPath string) *ClusterConfig {
	c.GetClusterSpec().SetConfigPath(configPath)
	return c
}

// SetScheduler sets the Scheduler for the ClusterConfig
func (c *ClusterConfig) SetScheduler(scheduler string) *ClusterConfig {
	c.GetClusterSpec().SetScheduler(scheduler)
	return c
}

// SetHyperconverged sets the Hyperconverged for the ClusterConfig
func (c *ClusterConfig) SetHyperconverged(isHyperconverged bool) *ClusterConfig {
	c.GetClusterSpec().SetHyperconverged(isHyperconverged)
	return c
}

// SetStorageProvisioner sets the StorageProvisioner for the ClusterConfig
func (c *ClusterConfig) SetStorageProvisioner(provisioner string) *ClusterConfig {
	c.GetClusterSpec().SetStorageProvisioner(provisioner)
	return c
}

// Register registers Cluster with the given ClusterSpec and UID
func (c *ClusterConfig) Register(clusterUID string) error {
	c.GetClusterManager().Set(clusterUID, NewDefaultCluster(c.GetClusterSpec()))
	return nil
}
