package cluster

import (
	"github.com/portworx/torpedo/drivers/backup/utils"
	"reflect"
)

const (
	// GlobalInClusterConfigPath is the config-path of the cluster in which Torpedo and Px-Backup are running
	GlobalInClusterConfigPath = "" // as described in the doc string of the `SetConfig` function in the k8s.go file of the k8s package
)

type Cluster struct {
	ContextManager   *ContextManager
	NamespaceManager *NamespaceManager
}

func (c *Cluster) ProcessClusterRequest(request interface{}) (response interface{}, err error) {
	err = c.ContextManager.SwitchContext()
	if err != nil {
		return nil, utils.ProcessError(err, utils.StructToString(c))
	}
	switch request.(type) {
	case *AppScheduleRequest:
		response, err = ScheduleApp(request.(*AppScheduleRequest))
		if err != nil {
			return nil, utils.ProcessError(err, utils.StructToString(request.(*AppScheduleRequest)))
		}
	}
	return response, nil
}

func NewCluster(configPath string) *Cluster {
	return &Cluster{
		ContextManager: &ContextManager{
			DstConfigPath: configPath,
		},
		NamespaceManager: &NamespaceManager{
			Namespaces:        make(map[string]*Namespace),
			RemovedNamespaces: make(map[string][]*Namespace),
		},
	}
}

type ClusterMetaData struct {
	ConfigPath string
}

func (m *ClusterMetaData) GetConfigPath() string {
	return m.ConfigPath
}

type ClusterConfig struct {
	ClusterMetaData   *ClusterMetaData
	InCluster         bool
	ClusterController *ClusterController
}

func (c *ClusterConfig) Equals(other *ClusterConfig) bool {
	return reflect.DeepEqual(c, other)
}

func (c *ClusterConfig) SetInCluster() *ClusterConfig {
	c.InCluster = true
	return c
}

func (c *ClusterConfig) Register(hyperConverged bool) (string, error) {
	clusterUid, isPresent := c.ClusterController.ClusterManager.IsClusterConfigRecorded(c)
	if isPresent {
		return clusterUid, nil
	}
	configPath := c.ClusterMetaData.ConfigPath
	if c.InCluster {
		configPath = GlobalInClusterConfigPath
	}
	if !hyperConverged {
		// ToDo: handle non hyper-converged cluster
	}
	clusterUid = "6d02ee80-448b-41a6-a866-b98a861d5590"
	c.ClusterController.ClusterManager.AddCluster(clusterUid, c, NewCluster(configPath))
	return clusterUid, nil
}

func (c *ClusterConfig) Namespace(namespace string) *NamespaceConfig {
	return &NamespaceConfig{
		ClusterMetaData:   c.ClusterMetaData,
		NamespaceMetaData: NewNamespaceMetaData(namespace),
		ClusterController: c.ClusterController,
	}
}

type ClusterManager struct {
	ClusterConfigs  map[string]*ClusterConfig
	Clusters        map[string]*Cluster
	RemovedClusters map[string][]*Cluster
}

func (m *ClusterManager) IsClusterConfigRecorded(clConfig *ClusterConfig) (string, bool) {
	for clusterUid, clusterConfig := range m.ClusterConfigs {
		if clusterConfig.Equals(clConfig) {
			return clusterUid, true
		}
	}
	return "", false
}

func (m *ClusterManager) GetCluster(clUid string) *Cluster {
	return m.Clusters[clUid]
}

func (m *ClusterManager) AddCluster(clUid string, clConfig *ClusterConfig, cluster *Cluster) {
	m.Clusters[clUid] = cluster
	m.ClusterConfigs[clUid] = clConfig
}

func (m *ClusterManager) RemoveCluster(clUid string) {
	m.RemovedClusters[clUid] = append(m.RemovedClusters[clUid], m.GetCluster(clUid))
	delete(m.Clusters, clUid)
}
