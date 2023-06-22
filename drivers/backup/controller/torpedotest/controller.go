package torpedotest

// TorpedoTestController represents a controller for TorpedoTest
type TorpedoTestController struct {
	TorpedoTestManager *TorpedoTestManager
}

// GetTorpedoTestManager returns the TorpedoTestManager associated with the TorpedoTestController
func (c *TorpedoTestController) GetTorpedoTestManager() *TorpedoTestManager {
	return c.TorpedoTestManager
}

// SetTorpedoTestManager sets the TorpedoTestManager for the TorpedoTestController
func (c *TorpedoTestController) SetTorpedoTestManager(manager *TorpedoTestManager) {
	c.TorpedoTestManager = manager
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
