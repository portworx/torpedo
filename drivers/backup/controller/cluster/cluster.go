package cluster

import (
	"github.com/portworx/torpedo/drivers/backup/utils"
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

func (m *ClusterMetaData) GetClusterUid() string {
	return m.GetConfigPath()
}

func NewClusterMetaData() *ClusterMetaData {
	newClusterMetaData := &ClusterMetaData{}
	newClusterMetaData.SetConfigPath(GlobalInClusterConfigPath)
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
	cluster := NewCluster()
	cluster.GetContextManager().SetDstConfigPath(configPath)
	c.ClusterController.ClusterManager.SetCluster(c.GetClusterMetaData().GetClusterUid(), cluster)
	return nil
}

func (c *ClusterConfig) Namespace(namespace string) *NamespaceConfig {
	namespaceMetaData := NewNamespaceMetaData()
	namespaceMetaData.SetNamespace(namespace)
	return &NamespaceConfig{
		ClusterMetaData:   c.ClusterMetaData,
		NamespaceMetaData: namespaceMetaData,
		ClusterController: c.ClusterController,
	}
}

func NewClusterConfig() *ClusterConfig {
	newClusterConfig := &ClusterConfig{}
	newClusterConfig.SetClusterMetaData(nil)
	newClusterConfig.SetIsInCluster(false)
	newClusterConfig.SetClusterController(nil)
	return newClusterConfig
}

type Cluster struct {
	ContextManager   *ContextManager
	NamespaceManager *NamespaceManager
}

func (c *Cluster) GetContextManager() *ContextManager {
	return c.ContextManager
}

func (c *Cluster) SetContextManager(manager *ContextManager) {
	c.ContextManager = manager
}

func (c *Cluster) GetNamespaceManager() *NamespaceManager {
	return c.NamespaceManager
}

func (c *Cluster) SetNamespaceManager(manager *NamespaceManager) {
	c.NamespaceManager = manager
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

func (m *ClusterManager) GetCluster(clusterUid string) *Cluster {
	return m.Clusters[clusterUid]
}

func (m *ClusterManager) SetCluster(clusterUid string, cluster *Cluster) {
	m.Clusters[clusterUid] = cluster
}

func (m *ClusterManager) DeleteCluster(clusterUid string) {
	delete(m.Clusters, clusterUid)
}

func (m *ClusterManager) RemoveCluster(clusterUid string) {
	m.RemovedClusters[clusterUid] = append(m.RemovedClusters[clusterUid], m.GetCluster(clusterUid))
}

func (m *ClusterManager) IsClusterPresent(clusterUid string) bool {
	_, isPresent := m.Clusters[clusterUid]
	return isPresent
}

func NewClusterManager() *ClusterManager {
	return &ClusterManager{
		Clusters:        make(map[string]*Cluster, 0),
		RemovedClusters: make(map[string][]*Cluster, 0),
	}
}
