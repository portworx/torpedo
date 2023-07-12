package cluster_manager

import (
	. "github.com/portworx/torpedo/drivers/pxbackup/controller/cl_controller/cluster_manager/cluster"
	. "github.com/portworx/torpedo/drivers/pxbackup/controller_utils/entity/entity_config/entity_manager"
)

// ClusterManager represents a manager for a Cluster
type ClusterManager EntityManager[*Cluster]
