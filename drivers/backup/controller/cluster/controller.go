package cluster

type ClusterController struct {
	ClusterManager *ClusterManager
}

func (c *ClusterController) Cluster(configPath string) *ClusterConfig {
	return &ClusterConfig{
		ClusterMetaData: &ClusterMetaData{
			ConfigPath: configPath,
		},
		InCluster:         false,
		ClusterController: c,
	}
}
