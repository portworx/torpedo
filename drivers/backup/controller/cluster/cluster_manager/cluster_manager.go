package cluster_manager

import (
	. "github.com/portworx/torpedo/drivers/backup/controller/cluster/cluster_manager/cluster"
	. "github.com/portworx/torpedo/drivers/backup/controller/generics/entity/entity_manager"
)

// ClusterManager represents a manager for a Cluster
type ClusterManager EntityManager[*Cluster]
