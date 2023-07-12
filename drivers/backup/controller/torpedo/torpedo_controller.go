package torpedo

import (
	. "github.com/portworx/torpedo/drivers/backup/controller/cluster"
)

// TorpedoController provides wrapper functions to simplify Torpedo related tasks
type TorpedoController struct {
	ClusterController *ClusterController
}

// GetClusterController returns the ClusterController associated with the TorpedoController
func (c *TorpedoController) GetClusterController() *ClusterController {
	return c.ClusterController
}

// SetClusterController sets the ClusterController for the TorpedoController
func (c *TorpedoController) SetClusterController(controller *ClusterController) *TorpedoController {
	c.ClusterController = controller
	return c
}

// NewTorpedoController creates a new instance of the TorpedoController
func NewTorpedoController(clusterController *ClusterController) *TorpedoController {
	torpedoController := &TorpedoController{}
	torpedoController.SetClusterController(clusterController)
	return torpedoController
}

// NewDefaultTorpedoController creates a new instance of the TorpedoController with default values
func NewDefaultTorpedoController() *TorpedoController {
	return NewTorpedoController(NewDefaultClusterController())
}
