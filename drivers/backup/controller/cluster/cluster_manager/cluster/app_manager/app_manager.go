package app_manager

import (
	. "github.com/portworx/torpedo/drivers/backup/controller/cluster/cluster_manager/cluster/app_manager/app"
	. "github.com/portworx/torpedo/drivers/backup/controller/cluster/cluster_manager/cluster/app_spec"
	. "github.com/portworx/torpedo/drivers/backup/controller/torpedo/torpedo_utils/entity_generics"
)

// AppManager represents a manager for an app.App
type AppManager struct {
	*EntityManager[*App[*AppSpec]]
}
