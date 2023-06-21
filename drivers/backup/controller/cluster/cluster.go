package cluster

import (
	"sync"
)

// ClusterMetaData represents the metadata for a Cluster
type ClusterMetaData struct {
	ConfigPath string
}

// GetConfigPath returns the ConfigPath associated with the ClusterMetaData
func (m *ClusterMetaData) GetConfigPath() string {
	return m.ConfigPath
}

// SetConfigPath sets the ConfigPath string for the ClusterMetaData
func (m *ClusterMetaData) SetConfigPath(configPath string) {
	m.ConfigPath = configPath
}

// GetClusterUid returns the Cluster uid
func (m *ClusterMetaData) GetClusterUid() string {
	return m.GetConfigPath()
}

// NewClusterMetaData creates a new instance of the ClusterMetaData
func NewClusterMetaData() *ClusterMetaData {
	newClusterMetaData := &ClusterMetaData{}
	newClusterMetaData.SetConfigPath(GlobalInClusterConfigPath)
	return newClusterMetaData
}

// ClusterConfig represents the configuration for a Cluster
type ClusterConfig struct {
	ClusterMetaData   *ClusterMetaData
	InCluster         bool
	ClusterController *ClusterController
}

// GetClusterMetaData returns the ClusterMetaData associated with the ClusterConfig
func (c *ClusterConfig) GetClusterMetaData() *ClusterMetaData {
	return c.ClusterMetaData
}

// SetClusterMetaData sets the ClusterMetaData for the ClusterConfig
func (c *ClusterConfig) SetClusterMetaData(clusterMetaData *ClusterMetaData) {
	c.ClusterMetaData = clusterMetaData
}

// GetInCluster returns the flag indicating whether a Cluster is InCluster in the ClusterConfig
func (c *ClusterConfig) GetInCluster() bool {
	return c.InCluster
}

// SetInCluster sets the flag indicating whether a Cluster is InCluster in the ClusterConfig
func (c *ClusterConfig) SetInCluster(inCluster bool) *ClusterConfig {
	c.InCluster = inCluster
	return c
}

// GetClusterController returns the ClusterController associated with the ClusterConfig
func (c *ClusterConfig) GetClusterController() *ClusterController {
	return c.ClusterController
}

// SetClusterController sets the ClusterController for the ClusterConfig
func (c *ClusterConfig) SetClusterController(clusterController *ClusterController) {
	c.ClusterController = clusterController
}

func (c *ClusterConfig) Register(hyperConverged bool) error {
	configPath := c.GetClusterMetaData().GetConfigPath()
	if c.GetInCluster() {
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

// NewClusterConfig creates a new instance of the ClusterConfig
func NewClusterConfig() *ClusterConfig {
	newClusterConfig := &ClusterConfig{}
	newClusterConfig.SetClusterMetaData(nil)
	newClusterConfig.SetInCluster(false)
	newClusterConfig.SetClusterController(nil)
	return newClusterConfig
}

// Cluster represents a Cluster
type Cluster struct {
	ContextManager   *ContextManager
	RequestManager   *RequestManager
	NamespaceManager *NamespaceManager
}

// GetContextManager returns the ContextManager associated with the Cluster
func (c *Cluster) GetContextManager() *ContextManager {
	return c.ContextManager
}

// SetContextManager sets the ContextManager for the Cluster
func (c *Cluster) SetContextManager(manager *ContextManager) {
	c.ContextManager = manager
}

// GetRequestManager returns the RequestManager associated with the Cluster
func (c *Cluster) GetRequestManager() *RequestManager {
	return c.RequestManager
}

// SetRequestManager sets the RequestManager for the Cluster
func (c *Cluster) SetRequestManager(manager *RequestManager) {
	c.RequestManager = manager
}

// GetNamespaceManager returns the NamespaceManager associated with the Cluster
func (c *Cluster) GetNamespaceManager() *NamespaceManager {
	return c.NamespaceManager
}

// SetNamespaceManager sets the ContextManager for the Cluster
func (c *Cluster) SetNamespaceManager(manager *NamespaceManager) {
	c.NamespaceManager = manager
}

func (c *Cluster) ProcessRequest() {
	//err = c.GetContextManager().SwitchContext()
	//if err != nil {
	//	return nil, utils.ProcessError(err)
	//}
}

// NewCluster creates a new instance of the Cluster
func NewCluster() *Cluster {
	newCluster := &Cluster{}
	newCluster.SetContextManager(NewContextManager())
	newCluster.SetNamespaceManager(NewNamespaceManager())
	newCluster.SetRequestManager(NewRequestManager())
	return newCluster
}

// ClusterManager represents a manager for Cluster
type ClusterManager struct {
	sync.RWMutex
	ClusterMap         map[string]*Cluster
	RemovedClustersMap map[string][]*Cluster
}

// GetClusterMap returns the ClusterMap of the ClusterManager
func (m *ClusterManager) GetClusterMap() map[string]*Cluster {
	m.RLock()
	defer m.RUnlock()
	return m.ClusterMap
}

// SetClusterMap sets the ClusterMap of the ClusterManager
func (m *ClusterManager) SetClusterMap(clusterMap map[string]*Cluster) {
	m.Lock()
	defer m.Unlock()
	m.ClusterMap = clusterMap
}

// GetRemovedClustersMap returns the RemovedClustersMap of the ClusterManager
func (m *ClusterManager) GetRemovedClustersMap() map[string][]*Cluster {
	m.RLock()
	defer m.RUnlock()
	return m.RemovedClustersMap
}

// SetRemovedClustersMap sets the RemovedClustersMap of the ClusterManager
func (m *ClusterManager) SetRemovedClustersMap(removedClustersMap map[string][]*Cluster) {
	m.Lock()
	defer m.Unlock()
	m.RemovedClustersMap = removedClustersMap
}

// GetCluster returns the Cluster with the given Cluster uid
func (m *ClusterManager) GetCluster(clusterUid string) *Cluster {
	m.RLock()
	defer m.RUnlock()
	return m.GetClusterMap()[clusterUid]
}

// IsClusterPresent checks if the Cluster with the given Cluster uid is present
func (m *ClusterManager) IsClusterPresent(clusterUid string) bool {
	m.RLock()
	defer m.RUnlock()
	_, isPresent := m.GetClusterMap()[clusterUid]
	return isPresent
}

// SetCluster sets the Cluster with the given Cluster uid
func (m *ClusterManager) SetCluster(clusterUid string, cluster *Cluster) {
	m.Lock()
	defer m.Unlock()
	m.GetClusterMap()[clusterUid] = cluster
}

// DeleteCluster deletes the Cluster with the given Cluster uid
func (m *ClusterManager) DeleteCluster(clusterUid string) {
	m.Lock()
	defer m.Unlock()
	delete(m.GetClusterMap(), clusterUid)
}

// RemoveCluster removes the Cluster with the given Cluster uid
func (m *ClusterManager) RemoveCluster(clusterUid string) {
	m.Lock()
	defer m.Unlock()
	m.GetRemovedClustersMap()[clusterUid] = append(m.GetRemovedClustersMap()[clusterUid], m.GetCluster(clusterUid))
}

// NewClusterManager creates a new instance of the ClusterManager
func NewClusterManager() *ClusterManager {
	newClusterManager := &ClusterManager{}
	newClusterManager.SetClusterMap(make(map[string]*Cluster, 0))
	newClusterManager.SetRemovedClustersMap(make(map[string][]*Cluster, 0))
	return newClusterManager
}
