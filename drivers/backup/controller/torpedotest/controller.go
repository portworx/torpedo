package torpedotest

type TorpedoTestController struct {
	TorpedoTestManager *TorpedoTestManager
}

func (c *TorpedoTestController) GetTorpedoTestManager() *TorpedoTestManager {
	return c.TorpedoTestManager
}

func (c *TorpedoTestController) SetTorpedoTestManager(manager *TorpedoTestManager) {
	c.TorpedoTestManager = manager
}

func (c *TorpedoTestController) TorpedoTest(testId string) *TorpedoTestConfig {
	torpedoTestConfig := NewTorpedoTestConfig()
	torpedoTestMetaData := NewTorpedoTestMetaData()
	torpedoTestMetaData.SetTestId(testId)
	torpedoTestConfig.SetTorpedoTestMetaData(torpedoTestMetaData)
	torpedoTestConfig.SetTorpedoTestController(c)
	return torpedoTestConfig
}
