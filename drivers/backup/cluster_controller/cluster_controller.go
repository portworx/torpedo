package cluster_controller

import . "github.com/portworx/torpedo/drivers/backup/cluster_controller/cluster_manager"

// ClusterController provides wrapper functions to streamline and simplify Cluster related tasks
type ClusterController struct {
	ClusterManager *ClusterManager
}

// GetClusterManager returns the ClusterManager of the ClusterController
func (c *ClusterController) GetClusterManager() *ClusterManager {
	return c.ClusterManager
}

// SetClusterManager sets the ClusterManager of the ClusterController
func (c *ClusterController) SetClusterManager(manager *ClusterManager) *ClusterController {
	c.ClusterManager = manager
	return c
}

// NewClusterController creates a new instance of the ClusterController
func NewClusterController(clusterManager *ClusterManager) *ClusterController {
	newClusterController := &ClusterController{}
	newClusterController.SetClusterManager(clusterManager)
	return newClusterController
}

// NewDefaultClusterController creates a new instance of the ClusterController with default values
func NewDefaultClusterController() *ClusterController {
	return NewClusterController(NewDefaultClusterManager())
}
