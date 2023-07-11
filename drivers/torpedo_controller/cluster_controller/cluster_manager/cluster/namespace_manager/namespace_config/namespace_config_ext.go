package namespace_config

import (
	. "github.com/portworx/torpedo/drivers/torpedo_controller/cluster_controller/cluster_manager/cluster/namespace_manager/namespace"
)

// SetNamespace sets the Namespace for the NamespaceConfig
func (c *NamespaceConfig) SetNamespace(namespace string) *NamespaceConfig {
	c.GetNamespaceSpec().SetNamespace(namespace)
	return c
}

// Register registers Namespace with the given NamespaceSpec and UID
func (c *NamespaceConfig) Register(namespaceUID string) error {
	c.GetNamespaceManager().Set(namespaceUID, NewDefaultNamespace(c.GetNamespaceSpec()))
	return nil
}
