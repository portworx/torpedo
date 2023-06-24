package cluster_manager

import (
	"github.com/portworx/torpedo/drivers/backup/controller/cluster_controller"
	"github.com/portworx/torpedo/drivers/backup/controller/cluster_controller/cluster_manager/context_manager"
	"github.com/portworx/torpedo/drivers/backup/controller/cluster_controller/cluster_manager/namespace_manager"
	"github.com/portworx/torpedo/drivers/backup/controller/cluster_controller/cluster_manager/request_manager"
	"github.com/portworx/torpedo/drivers/backup/controller/cluster_controller/cluster_manager/request_manager/schedulerapi"
	"reflect"
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

// SetConfigPath sets the ConfigPath for the ClusterMetaData
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
	newClusterMetaData.SetConfigPath(context_manager.GlobalInClusterConfigPath)
	return newClusterMetaData
}

// ClusterConfig represents the configuration for a Cluster
type ClusterConfig struct {
	ClusterMetaData   *ClusterMetaData
	InCluster         bool
	ClusterController *cluster_controller.ClusterController
}

// GetClusterMetaData returns the ClusterMetaData associated with the ClusterConfig
func (c *ClusterConfig) GetClusterMetaData() *ClusterMetaData {
	return c.ClusterMetaData
}

// SetClusterMetaData sets the ClusterMetaData for the ClusterConfig
func (c *ClusterConfig) SetClusterMetaData(metaData *ClusterMetaData) {
	c.ClusterMetaData = metaData
}

// GetInCluster returns the flag associated with the ClusterConfig indicating whether a Cluster is InCluster
func (c *ClusterConfig) GetInCluster() bool {
	return c.InCluster
}

// SetInCluster sets the flag for the ClusterConfig indicating whether a Cluster is InCluster
func (c *ClusterConfig) SetInCluster(inCluster bool) *ClusterConfig {
	c.InCluster = inCluster
	return c
}

// GetClusterController returns the ClusterController associated with the ClusterConfig
func (c *ClusterConfig) GetClusterController() *cluster_controller.ClusterController {
	return c.ClusterController
}

// SetClusterController sets the ClusterController for the ClusterConfig
func (c *ClusterConfig) SetClusterController(controller *cluster_controller.ClusterController) {
	c.ClusterController = controller
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
	sync.RWMutex
	ContextManager   *context_manager.ContextManager
	RequestManager   *request_manager.RequestManager
	NamespaceManager *namespace_manager.NamespaceManager
}

// GetContextManager returns the ContextManager associated with the Cluster
func (c *Cluster) GetContextManager() *context_manager.ContextManager {
	return c.ContextManager
}

// SetContextManager sets the ContextManager for the Cluster
func (c *Cluster) SetContextManager(manager *context_manager.ContextManager) {
	c.ContextManager = manager
}

// GetRequestManager returns the RequestManager associated with the Cluster
func (c *Cluster) GetRequestManager() *request_manager.RequestManager {
	return c.RequestManager
}

// SetRequestManager sets the RequestManager for the Cluster
func (c *Cluster) SetRequestManager(manager *request_manager.RequestManager) {
	c.RequestManager = manager
}

// GetNamespaceManager returns the NamespaceManager associated with the Cluster
func (c *Cluster) GetNamespaceManager() *namespace_manager.NamespaceManager {
	return c.NamespaceManager
}

// SetNamespaceManager sets the NamespaceManager for the Cluster
func (c *Cluster) SetNamespaceManager(manager *namespace_manager.NamespaceManager) {
	c.NamespaceManager = manager
}

// NewCluster creates a new instance of the Cluster
func NewCluster() *Cluster {
	newCluster := &Cluster{}
	newCluster.SetContextManager(context_manager.NewContextManager())
	newCluster.SetNamespaceManager(namespace_manager.NewNamespaceManager())
	requestManager := request_manager.NewRequestManager()
	requestManager.SetRequestProcessorMap(
		map[request_manager.RequestType]request_manager.RequestProcessor{
			reflect.TypeOf(schedulerapi.ScheduleRequest{}): func(request request_manager.Request) (request_manager.Response, error) {
				return schedulerapi.Schedule(request.(*schedulerapi.ScheduleRequest))
			},
		},
	)
	newCluster.SetRequestManager(requestManager)
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
	return m.ClusterMap[clusterUid]
}

// IsClusterPresent checks if the Cluster with the given Cluster uid is present
func (m *ClusterManager) IsClusterPresent(clusterUid string) bool {
	m.RLock()
	defer m.RUnlock()
	_, isPresent := m.ClusterMap[clusterUid]
	return isPresent
}

// SetCluster sets the Cluster with the given Cluster uid
func (m *ClusterManager) SetCluster(clusterUid string, cluster *Cluster) {
	m.Lock()
	defer m.Unlock()
	m.ClusterMap[clusterUid] = cluster
}

// DeleteCluster deletes the Cluster with the given Cluster uid
func (m *ClusterManager) DeleteCluster(clusterUid string) {
	m.Lock()
	defer m.Unlock()
	delete(m.ClusterMap, clusterUid)
}

// RemoveCluster removes the Cluster with the given Cluster uid
func (m *ClusterManager) RemoveCluster(clusterUid string) {
	m.Lock()
	defer m.Unlock()
	m.RemovedClustersMap[clusterUid] = append(m.RemovedClustersMap[clusterUid], m.ClusterMap[clusterUid])
}

// NewClusterManager creates a new instance of the ClusterManager
func NewClusterManager() *ClusterManager {
	newClusterManager := &ClusterManager{}
	newClusterManager.SetClusterMap(make(map[string]*Cluster, 0))
	newClusterManager.SetRemovedClustersMap(make(map[string][]*Cluster, 0))
	return newClusterManager
}
