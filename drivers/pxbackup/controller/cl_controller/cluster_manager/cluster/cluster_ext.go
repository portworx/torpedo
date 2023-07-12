package cluster

import (
	. "github.com/portworx/torpedo/drivers/pxbackup/controller/cl_controller/cluster_manager/cluster/namespace_manager/namespace"
	. "github.com/portworx/torpedo/drivers/pxbackup/controller/cl_controller/cluster_manager/cluster/namespace_manager/namespace_config"
	. "github.com/portworx/torpedo/drivers/pxbackup/controller/cl_controller/cluster_manager/cluster/namespace_spec"
)

// NamespaceSpec creates a new namespace_config.NamespaceConfig and configures it
func (c *Cluster) NamespaceSpec(namespace string) *NamespaceConfig {
	return NewNamespaceConfig(c.GetClusterSpec(), NewDefaultNamespaceSpec(namespace), c.NamespaceManager)
}

// Namespace returns the Namespace with the given Namespace UID
func (c *Cluster) Namespace(namespaceUID string) *Namespace {
	return c.GetNamespaceManager().Get(namespaceUID)
}
