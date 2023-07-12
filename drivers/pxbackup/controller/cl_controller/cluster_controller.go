package cl_controller

import (
	. "github.com/portworx/torpedo/drivers/pxbackup/controller/cl_controller/cluster_manager/cluster"
	. "github.com/portworx/torpedo/drivers/pxbackup/controller_utils/entity/entity_config/entity_manager"
)

// ClusterController provides wrapper functions to simplify Cluster related tasks
type ClusterController struct {
	ClusterManager *EntityManager[*Cluster]
}

// GetClusterManager returns the ClusterManager associated with the ClusterController
func (c *ClusterController) GetClusterManager() *EntityManager[*Cluster] {
	return c.ClusterManager
}

// SetClusterManager sets the ClusterManager for the ClusterController
func (c *ClusterController) SetClusterManager(manager *EntityManager[*Cluster]) *ClusterController {
	c.ClusterManager = manager
	return c
}

// NewClusterController creates a new instance of the ClusterController
func NewClusterController(
	clusterManager *EntityManager[*Cluster],
) *ClusterController {
	clusterController := &ClusterController{}
	clusterController.SetClusterManager(clusterManager)
	return clusterController
}

// NewDefaultClusterController creates a new instance of the ClusterController with default values
func NewDefaultClusterController() *ClusterController {
	return NewClusterController(NewDefaultEntityManager[*Cluster]())
}
