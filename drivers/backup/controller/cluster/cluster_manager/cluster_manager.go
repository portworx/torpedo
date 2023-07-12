package cluster_manager

import (
	. "github.com/portworx/torpedo/drivers/backup/controller/cluster/cluster_manager/cluster"
	. "github.com/portworx/torpedo/drivers/backup/controller/torpedo/torpedo_utils/entity_generics"
)

// ClusterManager represents a manager for a Cluster
type ClusterManager EntityManager[*Cluster]
