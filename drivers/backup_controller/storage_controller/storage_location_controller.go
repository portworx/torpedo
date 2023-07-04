package storage_controller

// StorageLocationController provides wrapper functions to streamline and simplify StorageLocation related tasks
type StorageLocationController struct {
	AWSLocationManager *AWSLocationManager
}

// GetAWSLocationManager returns the AWSLocationManager of the StorageLocationController
func (c *StorageLocationController) GetAWSLocationManager() *AWSLocationManager {
	return c.AWSLocationManager
}

// SetAWSLocationManager sets the AWSLocationManager of the StorageLocationController
func (c *StorageLocationController) SetAWSLocationManager(manager *AWSLocationManager) *StorageLocationController {
	c.AWSLocationManager = manager
	return c
}

// NewStorageLocationController creates a new instance of the StorageLocationController
func NewStorageLocationController(awsLocationManager *AWSLocationManager) *StorageLocationController {
	newStorageLocationController := &StorageLocationController{}
	newStorageLocationController.SetAWSLocationManager(awsLocationManager)
	return newStorageLocationController
}

// NewDefaultStorageLocationController creates a new instance of the StorageLocationController with default values
func NewDefaultStorageLocationController() *StorageLocationController {
	return NewStorageLocationController(NewDefaultAWSLocationManager())
}
