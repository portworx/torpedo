package backup_controller

import (
	. "github.com/portworx/torpedo/drivers/backup_controller/cluster_controller"
	. "github.com/portworx/torpedo/drivers/backup_controller/storage_controller"
)

// BackupController provides wrapper functions to streamline and simplify Backup related tasks
type BackupController struct {
	ClusterController         *ClusterController
	StorageLocationController *StorageLocationController
}

// GetClusterController returns the ClusterController of the BackupController
func (c *BackupController) GetClusterController() *ClusterController {
	return c.ClusterController
}

// SetClusterController sets the ClusterController of the BackupController
func (c *BackupController) SetClusterController(controller *ClusterController) *BackupController {
	c.ClusterController = controller
	return c
}

// GetStorageLocationController returns the StorageLocationController of the BackupController
func (c *BackupController) GetStorageLocationController() *StorageLocationController {
	return c.StorageLocationController
}

// SetStorageLocationController sets the StorageLocationController of the BackupController
func (c *BackupController) SetStorageLocationController(controller *StorageLocationController) *BackupController {
	c.StorageLocationController = controller
	return c
}

// NewBackupController creates a new instance of the BackupController
func NewBackupController(clusterController *ClusterController, storageLocationController *StorageLocationController) *BackupController {
	backupController := &BackupController{}
	backupController.SetClusterController(clusterController)
	backupController.SetStorageLocationController(storageLocationController)
	return backupController
}

// NewDefaultBackupController creates a new instance of the BackupController with default values
func NewDefaultBackupController() *BackupController {
	return NewBackupController(NewDefaultClusterController(), NewDefaultStorageLocationController())
}
