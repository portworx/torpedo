package cluster

type ClusterController struct {
	ClusterManager *ClusterManager
}

func (c *ClusterController) Cluster(configPath string) *ClusterConfig {
	clusterConfig := NewClusterConfig()
	clusterConfig.SetClusterMetaData(NewClusterMetaData(configPath))
	clusterConfig.SetIsInCluster(false)
	clusterConfig.SetClusterController(c)
	return clusterConfig
}

func NewClusterController() *ClusterController {
	return &ClusterController{
		ClusterManager: NewClusterManager(),
	}
}
