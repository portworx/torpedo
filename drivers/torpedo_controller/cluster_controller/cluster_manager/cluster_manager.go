package cluster_manager

import (
	. "github.com/portworx/torpedo/drivers/torpedo_controller/cluster_controller/cluster_manager/cluster"
	. "github.com/portworx/torpedo/drivers/torpedo_controller/torpedo_utils/entity_generics"
)

// ClusterManager represents a manager for a Cluster
type ClusterManager EntityManager[*Cluster]
