package cluster

import (
	"github.com/portworx/torpedo/drivers/backup/utils"
)

const (
	// GlobalInClusterConfigPath is the config-path of the cluster in which Torpedo and Px-Backup are running
	GlobalInClusterConfigPath = "" // as described in the doc string of the `SetConfig` function in the k8s.go file of the k8s package
)

type ClusterMetaData struct {
	ConfigPath string
}

func (m *ClusterMetaData) GetConfigPath() string {
	return m.ConfigPath
}

func (m *ClusterMetaData) SetConfigPath(configPath string) {
	m.ConfigPath = configPath
}

func (m *ClusterMetaData) GetClusterName() string {
	return m.GetConfigPath()
}

func NewClusterMetaData(configPath string) *ClusterMetaData {
	newClusterMetaData := &ClusterMetaData{}
	newClusterMetaData.SetConfigPath(configPath)
	return newClusterMetaData
}

type ClusterConfig struct {
	ClusterMetaData   *ClusterMetaData
	IsInCluster       bool
	ClusterController *ClusterController
}

func (c *ClusterConfig) GetClusterMetaData() *ClusterMetaData {
	return c.ClusterMetaData
}

func (c *ClusterConfig) SetClusterMetaData(clusterMetaData *ClusterMetaData) {
	c.ClusterMetaData = clusterMetaData
}

func (c *ClusterConfig) GetIsInCluster() bool {
	return c.IsInCluster
}

func (c *ClusterConfig) SetIsInCluster(isInCluster bool) *ClusterConfig {
	c.IsInCluster = isInCluster
	return c
}

func (c *ClusterConfig) GetClusterController() *ClusterController {
	return c.ClusterController
}

func (c *ClusterConfig) SetClusterController(clusterController *ClusterController) {
	c.ClusterController = clusterController
}

func (c *ClusterConfig) Register(hyperConverged bool) error {
	configPath := c.GetClusterMetaData().GetConfigPath()
	if c.IsInCluster {
		configPath = GlobalInClusterConfigPath
	}
	if !hyperConverged {
		// ToDo: handle non hyper-converged cluster
	}
	clusterName, newCluster := c.GetClusterMetaData().GetClusterName(), NewCluster()
	newCluster.ContextManager.SetDstConfigPath(configPath)
	c.ClusterController.ClusterManager.SetCluster(clusterName, newCluster)
	return nil
}

func (c *ClusterConfig) Namespace(namespace string) *NamespaceConfig {
	return &NamespaceConfig{
		ClusterMetaData:   c.ClusterMetaData,
		NamespaceMetaData: NewNamespaceMetaData(namespace),
		ClusterController: c.ClusterController,
	}
}

func NewClusterConfig() *ClusterConfig {
	return &ClusterConfig{}
}

type Cluster struct {
	ContextManager   *ContextManager
	NamespaceManager *NamespaceManager
}

func (c *Cluster) ProcessClusterRequest(request interface{}) (response interface{}, err error) {
	err = c.ContextManager.SwitchContext()
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	//switch request.(type) {
	//case *AppScheduleRequest:
	//	response, err = ScheduleApp(request.(*AppScheduleRequest))
	//	if err != nil {
	//		return nil, utils.ProcessError(err, utils.StructToString(request.(*AppScheduleRequest)))
	//	}
	//}
	return response, err
}

func NewCluster() *Cluster {
	return &Cluster{
		ContextManager:   NewContextManager(),
		NamespaceManager: NewNamespaceManager(),
	}
}

type ClusterManager struct {
	Clusters        map[string]*Cluster
	RemovedClusters map[string][]*Cluster
}

func (m *ClusterManager) GetCluster(clusterName string) *Cluster {
	return m.Clusters[clusterName]
}

func (m *ClusterManager) SetCluster(clusterName string, cluster *Cluster) {
	m.Clusters[clusterName] = cluster
}

func (m *ClusterManager) DeleteCluster(clusterName string) {
	delete(m.Clusters, clusterName)
}

func (m *ClusterManager) RemoveCluster(clusterName string) {
	m.RemovedClusters[clusterName] = append(m.RemovedClusters[clusterName], m.GetCluster(clusterName))
}

func (m *ClusterManager) IsClusterPresent(clusterName string) bool {
	_, isPresent := m.Clusters[clusterName]
	return isPresent
}

func NewClusterManager() *ClusterManager {
	return &ClusterManager{
		Clusters:        make(map[string]*Cluster, 0),
		RemovedClusters: make(map[string][]*Cluster, 0),
	}
}
