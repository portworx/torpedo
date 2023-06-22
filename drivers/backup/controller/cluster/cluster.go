package cluster

import "github.com/portworx/torpedo/drivers/backup/utils"

// CanRegister checks if Cluster CanRegister
func (c *ClusterConfig) CanRegister() error {
	return nil
}

// Register registers Cluster
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

// Namespace creates a new NamespaceConfig and configures it
func (c *ClusterConfig) Namespace(namespace string) *NamespaceConfig {
	namespaceConfig := NewNamespaceConfig()
	namespaceConfig.SetClusterMetaData(c.GetClusterMetaData())
	namespaceMetaData := NewNamespaceMetaData()
	namespaceMetaData.SetNamespace(namespace)
	namespaceConfig.SetNamespaceMetaData(namespaceMetaData)
	namespaceConfig.SetClusterController(c.GetClusterController())
	return namespaceConfig
}

// ProcessClusterRequest processes Cluster Request
func (c *Cluster) ProcessClusterRequest(request Request) (response Response, err error) {
	c.Lock()
	defer c.Unlock()
	err = c.GetContextManager().SwitchContext()
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	response, err = c.GetRequestManager().ProcessRequest(request)
	if err != nil {
		debugStruct := struct {
			Request Request
		}{
			Request: request,
		}
		return nil, utils.ProcessError(err, utils.StructToString(debugStruct))
	}
	return response, nil
}
