package pxbackuptest

// PxBackupTestController represents a controller for PxBackupTest
type PxBackupTestController struct {
	PxBackupTestManager *PxBackupTestManager
}

// GetPxBackupTestManager returns the PxBackupTestManager associated with the PxBackupTestController
func (c *PxBackupTestController) GetPxBackupTestManager() *PxBackupTestManager {
	return c.PxBackupTestManager
}

// SetPxBackupTestManager sets the PxBackupTestManager for the PxBackupTestController
func (c *PxBackupTestController) SetPxBackupTestManager(manager *PxBackupTestManager) {
	c.PxBackupTestManager = manager
}

// PxBackupTest creates a new PxBackupTestConfig and configures it
func (c *PxBackupTestController) PxBackupTest(testId string) *PxBackupTestConfig {
	pxBackupTestConfig := NewPxBackupTestConfig()
	pxBackupTestMetaData := NewPxBackupTestMetaData()
	pxBackupTestMetaData.SetTestId(testId)
	pxBackupTestConfig.SetPxBackupTestMetaData(pxBackupTestMetaData)
	pxBackupTestConfig.SetPxBackupTestController(c)
	return pxBackupTestConfig
}

// NewPxBackupTestController creates a new instance of the PxBackupTestController
func NewPxBackupTestController() *PxBackupTestController {
	newPxBackupTestController := &PxBackupTestController{}
	pxBackupTestManager := NewPxBackupTestManager()
	newPxBackupTestController.SetPxBackupTestManager(pxBackupTestManager)
	return newPxBackupTestController
}
