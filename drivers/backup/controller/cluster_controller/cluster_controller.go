package cluster_controller

import "github.com/portworx/torpedo/drivers/backup/controller/cluster_controller/cluster_manager"

type ClusterController struct {
	ClusterManager *cluster_manager.ClusterManager
}

func (c *ClusterController) GetClusterManager() *cluster_manager.ClusterManager {
	return c.ClusterManager
}

func (c *ClusterController) SetClusterManager(manager *cluster_manager.ClusterManager) {
	c.ClusterManager = manager
}

func (c *ClusterController) Cluster(configPath string) *cluster_manager.ClusterConfig {
	clusterMetaData := cluster_manager.NewClusterMetaData()
	clusterMetaData.SetConfigPath(configPath)
	clusterConfig := cluster_manager.NewClusterConfig()
	clusterConfig.SetClusterMetaData(clusterMetaData)
	clusterConfig.SetClusterController(c)
	return clusterConfig
}

func NewClusterController() *ClusterController {
	newClusterController := &ClusterController{}
	newClusterController.SetClusterManager(cluster_manager.NewClusterManager())
	return newClusterController
}
