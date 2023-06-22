package cluster

type ClusterController struct {
	ClusterManager *ClusterManager
}

func (c *ClusterController) GetClusterManager() *ClusterManager {
	return c.ClusterManager
}

func (c *ClusterController) SetClusterManager(manager *ClusterManager) {
	c.ClusterManager = manager
}

func (c *ClusterController) Cluster(configPath string) *ClusterConfig {
	clusterMetaData := NewClusterMetaData()
	clusterMetaData.SetConfigPath(configPath)
	clusterConfig := NewClusterConfig()
	clusterConfig.SetClusterMetaData(clusterMetaData)
	clusterConfig.SetClusterController(c)
	return clusterConfig
}

func NewClusterController() *ClusterController {
	newClusterController := &ClusterController{}
	newClusterController.SetClusterManager(NewClusterManager())
	return newClusterController
}
