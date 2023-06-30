package backup

import . "github.com/portworx/torpedo/drivers/backup/cluster_controller"

// BackupController provides wrapper functions to streamline and simplify Backup related tasks
type BackupController struct {
	ClusterController *ClusterController
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

// NewBackupController creates a new instance of the BackupController
func NewBackupController(clusterController *ClusterController) *BackupController {
	backupController := &BackupController{}
	backupController.SetClusterController(clusterController)
	return backupController
}

// NewDefaultBackupController creates a new instance of the BackupController with default values
func NewDefaultBackupController() *BackupController {
	return NewBackupController(NewDefaultClusterController())
}
