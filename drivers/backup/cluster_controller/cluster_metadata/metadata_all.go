package cluster_metadata

// ClusterMetaData represents the metadata for cluster_manager.Cluster
type ClusterMetaData struct {
	ConfigPath string
}

// NamespaceMetaData represents the metadata for namespace_manager.Namespace
type NamespaceMetaData struct {
	ClusterMetaData *ClusterMetaData
	NamespaceName   string
}

// PodByNameMetaData represents the metadata for pod_by_name_manager.PodByName
type PodByNameMetaData struct {
	NamespaceMetaData *NamespaceMetaData
	PodName           string
}
