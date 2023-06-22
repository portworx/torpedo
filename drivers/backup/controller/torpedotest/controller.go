package torpedotest

import "github.com/portworx/torpedo/drivers/backup/controller/cluster"

// TorpedoTestController represents a controller for TorpedoTest
type TorpedoTestController struct {
	TorpedoTestManager *TorpedoTestManager
	ClusterController  *cluster.ClusterController
}

// GetTorpedoTestManager returns the TorpedoTestManager associated with the TorpedoTestController
func (c *TorpedoTestController) GetTorpedoTestManager() *TorpedoTestManager {
	return c.TorpedoTestManager
}

// SetTorpedoTestManager sets the TorpedoTestManager for the TorpedoTestController
func (c *TorpedoTestController) SetTorpedoTestManager(manager *TorpedoTestManager) {
	c.TorpedoTestManager = manager
}

// GetClusterController returns the ClusterController associated with the TorpedoTestController
func (c *TorpedoTestController) GetClusterController() *cluster.ClusterController {
	return c.ClusterController
}

// SetClusterController sets the ClusterController for the TorpedoTestController
func (c *TorpedoTestController) SetClusterController(controller *cluster.ClusterController) {
	c.ClusterController = controller
}

// TorpedoTest creates a new TorpedoTestConfig and configures it
func (c *TorpedoTestController) TorpedoTest(testId string) *TorpedoTestConfig {
	torpedoTestConfig := NewTorpedoTestConfig()
	torpedoTestMetaData := NewTorpedoTestMetaData()
	torpedoTestMetaData.SetTestId(testId)
	torpedoTestConfig.SetTorpedoTestMetaData(torpedoTestMetaData)
	torpedoTestConfig.SetTorpedoTestController(c)
	return torpedoTestConfig
}

// NewTorpedoTestController creates a new instance of the TorpedoTestController
func NewTorpedoTestController() *TorpedoTestController {
	newTorpedoTestController := &TorpedoTestController{}
	torpedoTestManager := NewTorpedoTestManager()
	newTorpedoTestController.SetTorpedoTestManager(torpedoTestManager)
	return newTorpedoTestController
}
