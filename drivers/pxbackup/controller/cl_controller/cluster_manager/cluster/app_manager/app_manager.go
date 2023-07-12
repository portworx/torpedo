package app_manager

import (
	. "github.com/portworx/torpedo/drivers/pxbackup/controller/cl_controller/cluster_manager/cluster/app_manager/app"
	. "github.com/portworx/torpedo/drivers/pxbackup/controller/cl_controller/cluster_manager/cluster/app_spec"
	. "github.com/portworx/torpedo/drivers/pxbackup/controller_utils/entity/entity_config/entity_manager"
)

// AppManager represents a manager for an app.App
type AppManager struct {
	*EntityManager[*App[*AppSpec]]
}
