package cluster

import (
	. "github.com/portworx/torpedo/drivers/backup/controller/cluster/cluster_manager/cluster"
	. "github.com/portworx/torpedo/drivers/backup/controller/torpedo/torpedo_utils/entity_generics"
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
	return NewClusterController(NewDefaultManager[*Cluster]())
}
