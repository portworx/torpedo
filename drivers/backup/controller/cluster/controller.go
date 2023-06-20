package cluster

const (
	// GlobalInClusterConfigPath is the config-path of the cluster in which Torpedo and Px-Backup are running
	GlobalInClusterConfigPath = "" // as described in the doc string of the `SetConfig` function in the k8s.go file of the k8s package
)

type ClusterController struct {
	ClusterManager *ClusterManager
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
	return &ClusterController{
		ClusterManager: NewClusterManager(),
	}
}
