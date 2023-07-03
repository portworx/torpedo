package namespace_manager

import (
	. "github.com/portworx/torpedo/drivers/backup/cluster_controller/cluster_manager/namespace_manager/pod_by_name_manager"
	. "github.com/portworx/torpedo/drivers/backup/cluster_controller/cluster_metadata"
)

// PodByName creates a new pod_by_name_manager.PodByNameConfig and configures it
func (c *NamespaceConfig) PodByName(podName string) *PodByNameConfig {
	if c == nil || c.GetNamespaceMetaData() == nil {
		return nil
	}
	namespaceUID := c.GetNamespaceMetaData().GetNamespaceUID()
	podByNameMetaData := NewPodByNameMetaData(c.GetNamespaceMetaData(), podName)
	if c.GetNamespaceManager() == nil || c.GetNamespaceManager().GetNamespace(namespaceUID) == nil {
		return NewPodByNameConfig(nil, podByNameMetaData)
	}
	podByNameManager := c.GetNamespaceManager().GetNamespace(namespaceUID).GetPodByNameManager()
	return NewPodByNameConfig(podByNameManager, podByNameMetaData)
}
