package cluster_utils

import "github.com/portworx/torpedo/drivers/backup/backup_utils"

const (
	// DefaultLogsLocation is default location for the logs
	DefaultLogsLocation = "/testresults/"
	// DefaultPodName is the default name for the pod_by_name_manager.PodByName
	DefaultPodName = "torpedo"
	// DefaultConfigPath is the default config-path for the cluster_manager.Cluster
	DefaultConfigPath = backup_utils.GlobalInClusterConfigPath
	// DefaultNamespaceName is the default name for the namespace_manager.Namespace
	DefaultNamespaceName = "default"
)
