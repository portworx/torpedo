package namespace_config

import (
	. "github.com/portworx/torpedo/drivers/backup/controller/cluster/cluster_manager/cluster/namespace_manager/namespace"
	. "github.com/portworx/torpedo/drivers/backup/controller/cluster/cluster_manager/cluster/namespace_spec"
	. "github.com/portworx/torpedo/drivers/backup/controller/cluster/cluster_spec"
	. "github.com/portworx/torpedo/drivers/backup/controller/torpedo/torpedo_utils/entity_generics"
)

// NamespaceConfig represents the configuration for a NamespaceSpec
type NamespaceConfig struct {
	ClusterSpec      *ClusterSpec
	NamespaceSpec    *NamespaceSpec
	NamespaceManager *EntityManager[*Namespace]
}

// GetClusterSpec returns the ClusterSpec associated with the NamespaceConfig
func (c *NamespaceConfig) GetClusterSpec() *ClusterSpec {
	return c.ClusterSpec
}

// SetClusterSpec sets the ClusterSpec for the NamespaceConfig
func (c *NamespaceConfig) SetClusterSpec(spec *ClusterSpec) *NamespaceConfig {
	c.ClusterSpec = spec
	return c
}

// GetNamespaceSpec returns the NamespaceSpec associated with the NamespaceConfig
func (c *NamespaceConfig) GetNamespaceSpec() *NamespaceSpec {
	return c.NamespaceSpec
}

// SetNamespaceSpec sets the NamespaceSpec for the NamespaceConfig
func (c *NamespaceConfig) SetNamespaceSpec(spec *NamespaceSpec) *NamespaceConfig {
	c.NamespaceSpec = spec
	return c
}

// GetNamespaceManager returns the NamespaceManager associated with the NamespaceConfig
func (c *NamespaceConfig) GetNamespaceManager() *EntityManager[*Namespace] {
	return c.NamespaceManager
}

// SetNamespaceManager sets the NamespaceManager for the NamespaceConfig
func (c *NamespaceConfig) SetNamespaceManager(manager *EntityManager[*Namespace]) *NamespaceConfig {
	c.NamespaceManager = manager
	return c
}

// NewNamespaceConfig creates a new instance of the NamespaceConfig
func NewNamespaceConfig(clusterSpec *ClusterSpec, namespaceSpec *NamespaceSpec, namespaceManager *EntityManager[*Namespace]) *NamespaceConfig {
	namespaceConfig := &NamespaceConfig{}
	namespaceConfig.SetClusterSpec(clusterSpec)
	namespaceConfig.SetNamespaceSpec(namespaceSpec)
	namespaceConfig.SetNamespaceManager(namespaceManager)
	return namespaceConfig
}
