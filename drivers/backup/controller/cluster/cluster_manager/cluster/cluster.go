package cluster

import (
	. "github.com/portworx/torpedo/drivers/backup/controller/cluster/cluster_manager/cluster/app_manager"
	. "github.com/portworx/torpedo/drivers/backup/controller/cluster/cluster_manager/cluster/namespace_manager/namespace"
	. "github.com/portworx/torpedo/drivers/backup/controller/cluster/cluster_spec"
	. "github.com/portworx/torpedo/drivers/backup/controller/generics/entity/entity_manager"
)

// Cluster represents Cluster
type Cluster struct {
	Spec             *ClusterSpec
	AppManager       *AppManager
	NamespaceManager *NamespaceManager
}

// GetClusterSpec returns the Spec associated with the Cluster
func (c *Cluster) GetClusterSpec() *ClusterSpec {
	return c.Spec
}

// SetClusterSpec sets the Spec for the Cluster
func (c *Cluster) SetClusterSpec(spec *ClusterSpec) *Cluster {
	c.Spec = spec
	return c
}

// GetNamespaceManager returns the NamespaceManager associated with the Cluster
func (c *Cluster) GetNamespaceManager() *NamespaceManager {
	return c.NamespaceManager
}

// SetNamespaceManager sets the NamespaceManager for the Cluster
func (c *Cluster) SetNamespaceManager(manager *NamespaceManager) *Cluster {
	c.NamespaceManager = manager
	return c
}

// NewCluster creates a new instance of the Cluster
func NewCluster(clusterSpec *ClusterSpec, namespaceManager *EntityManager[*Namespace]) *Cluster {
	cluster := &Cluster{}
	cluster.SetClusterSpec(clusterSpec)
	cluster.SetNamespaceManager(namespaceManager)
	return cluster
}

// NewDefaultCluster creates a new instance of the Cluster with default values
func NewDefaultCluster(clusterSpec *ClusterSpec) *Cluster {
	return NewCluster(clusterSpec, NewDefaultManager[*Namespace]())
}
