package torpedo_controller

import "github.com/portworx/torpedo/drivers/backup/controller/torpedotest"

// TorpedoController represents a controller for Torpedo
type TorpedoController struct {
	TorpedoTestController *torpedotest.TorpedoTestController
}

// GetTorpedoTestController returns the TorpedoTestController associated with the TorpedoController
func (c *TorpedoController) GetTorpedoTestController() *torpedotest.TorpedoTestController {
	return c.TorpedoTestController
}

// SetTorpedoTestController sets the TorpedoTestController for the TorpedoController
func (c *TorpedoController) SetTorpedoTestController(controller *torpedotest.TorpedoTestController) {
	c.TorpedoTestController = controller
}

// NewTorpedoController creates a new instance of the TorpedoController
func NewTorpedoController() *TorpedoController {
	newTorpedoController := &TorpedoController{}
	newTorpedoController.SetTorpedoTestController(torpedotest.NewTorpedoTestController())
	return newTorpedoController
}
