package torpedotest

import (
	"github.com/portworx/torpedo/drivers/backup/controller/pxbackuptest"
)

// TorpedoTestController represents a controller for TorpedoTest
type TorpedoTestController struct {
	PxBackupTestController *pxbackuptest.PxBackupTestController
}

// GetPxBackupTestController returns the PxBackupTestController associated with the TorpedoTestController
func (c *TorpedoTestController) GetPxBackupTestController() *pxbackuptest.PxBackupTestController {
	return c.PxBackupTestController
}

// SetPxBackupTestController sets the PxBackupTestController for the TorpedoTestController
func (c *TorpedoTestController) SetPxBackupTestController(controller *pxbackuptest.PxBackupTestController) {
	c.PxBackupTestController = controller
}

// NewTorpedoTestController creates a new instance of the TorpedoTestController
func NewTorpedoTestController() *TorpedoTestController {
	newTorpedoTestController := &TorpedoTestController{}
	newTorpedoTestController.SetPxBackupTestController(pxbackuptest.NewPxBackupTestController())
	return newTorpedoTestController
}
