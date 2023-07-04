package storage_controller

import (
	. "github.com/portworx/torpedo/drivers/backup_controller/storage_controller/storage_location_metadata"
)

// AWSLocation creates a new aws_location_manager.AWSLocationConfig and configures it
func (c *StorageLocationController) AWSLocation(locationName string) *AWSLocationConfig {
	if c == nil {
		return nil
	}
	awsLocationManager := c.GetAWSLocationManager()
	awsLocationMetaData := NewAWSLocationMetaData(locationName)
	return NewAWSLocationConfig(awsLocationManager, awsLocationMetaData)
}
