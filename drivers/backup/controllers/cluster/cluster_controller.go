package cluster

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/backup/utils"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/tests"
	"time"
)

const (
	// GlobalMaxNamespaceCharLimit is the maximum character limit for a Kubernetes namespace
	GlobalMaxNamespaceCharLimit = 63
)

const (
	// GlobalAuthTokenParam is an exact copy of the `authTokenParam` constant defined in the common.go file of the tests package
	GlobalAuthTokenParam = "auth-token"
	// GlobalInClusterConfigPath is the config path of the cluster in which Torpedo is running, as described in the doc string of the `SetConfig` function in the k8s.go file of the k8s package
	GlobalInClusterConfigPath = ""
)

const (
	// DefaultWaitForRunningTimeout specifies the default duration to wait for an application to reach the running state
	DefaultWaitForRunningTimeout = 10 * time.Minute
	// DefaultWaitForRunningRetryInterval specifies the default interval between retries when waiting for an application to reach the running state
	DefaultWaitForRunningRetryInterval = 10 * time.Second
	// DefaultValidateVolumeTimeout specifies the default duration to wait for volume validation
	DefaultValidateVolumeTimeout = 10 * time.Minute
	// DefaultValidateVolumeRetryInterval specifies the default interval between retries when performing volume validation
	DefaultValidateVolumeRetryInterval = 10 * time.Second
)

const (
	// DefaultWaitForDestroy represents the default value of the key `OptionsWaitForDestroy` defined in the scheduler.go file of the scheduler package
	// DefaultWaitForDestroy indicates whether to wait for resources to be destroyed during the cleanup process
	DefaultWaitForDestroy = true
	// DefaultWaitForResourceLeakCleanup represents the default value of the key `OptionsWaitForResourceLeakCleanup` defined in the scheduler.go file of the scheduler package
	// DefaultWaitForResourceLeakCleanup indicates whether to wait for resource leak cleanup during the cleanup process
	DefaultWaitForResourceLeakCleanup = true
	// DefaultSkipClusterScopedObjects represents the default value of `Context.SkipClusterScopedObject` defined in the scheduler.go file of the scheduler package
	// DefaultSkipClusterScopedObjects indicates whether to skip cluster-scoped objects during the cleanup process
	DefaultSkipClusterScopedObjects = false
)

// NamespaceInfo holds information related to a namespace within a cluster
type NamespaceInfo struct {
	contexts []*scheduler.Context
}

// appendContexts appends the specified slice of *scheduler.Context to slice of *scheduler.Context in NamespaceInfo
func (i *NamespaceInfo) appendContexts(contexts []*scheduler.Context) {
	i.contexts = append(i.contexts, contexts...)
}

// removeContexts removes the specified slice of *scheduler.Context from slice of *scheduler.Context in NamespaceInfo
func (i *NamespaceInfo) removeContexts(contexts []*scheduler.Context) {
	removeContextsMap := make(map[*scheduler.Context]bool)
	for _, ctx := range contexts {
		removeContextsMap[ctx] = true
	}
	remainingContexts := make([]*scheduler.Context, 0)
	for _, ctx := range i.contexts {
		if !removeContextsMap[ctx] {
			remainingContexts = append(remainingContexts, ctx)
		}
	}
	i.contexts = remainingContexts
}

// ClusterController provides wrapper functions to streamline and simplify cluster related tasks
type ClusterController struct {
	*ClusterInfo
	namespaces     map[string]*NamespaceInfo
	appKeyCountMap map[string]int
}

// getNamespaceInfo returns the NamespaceInfo of the specified namespace from the ClusterController
func (c *ClusterController) getNamespaceInfo(namespace string) *NamespaceInfo {
	namespaceInfo, ok := c.namespaces[namespace]
	if !ok {
		return &NamespaceInfo{}
	}
	return namespaceInfo
}

// saveNamespaceInfo saves the NamespaceInfo of the specified namespace in ClusterController
func (c *ClusterController) saveNamespaceInfo(namespace string, namespaceInfo *NamespaceInfo) {
	c.namespaces[namespace] = namespaceInfo
}

// delNamespaceInfo deletes the NamespaceInfo of the specified namespace in ClusterController
func (c *ClusterController) delNamespaceInfo(namespace string) {
	delete(c.namespaces, namespace)
}

// isNamespaceRecorded returns true if the specified namespace is recorded in the ClusterController
func (c *ClusterController) isNamespaceRecorded(namespace string) bool {
	_, ok := c.namespaces[namespace]
	return ok
}

// getAppKeyCount returns the count of occurrences of the specified app-key in the ClusterController
func (c *ClusterController) getAppKeyCount(appKey string) int {
	return c.appKeyCountMap[appKey]
}

// incrementAppKeyCount increments the count of the specified app-key in the ClusterController
func (c *ClusterController) incrementAppKeyCount(appKey string) {
	c.appKeyCountMap[appKey] += 1
}

// getAllNamespaces returns a slice of all the namespaces in the ClusterController
func (c *ClusterController) getAllNamespaces() []string {
	namespaces := make([]string, 0)
	for namespace := range c.namespaces {
		namespaces = append(namespaces, namespace)
	}
	return namespaces
}

// SwitchContext switches the cluster context to the cluster specified by the configPath in ClusterController
func (c *ClusterController) SwitchContext() error {
	err := utils.SwitchClusterContext(c.ConfigPath)
	if err != nil {
		debugMessage := fmt.Sprintf("config-path: [%s]", c.ConfigPath)
		return utils.ProcessError(err, debugMessage)
	}
	return nil
}

// Application returns a new ScheduleApplicationsConfig instance for scheduling single application with ClusterController
func (c *ClusterController) Application(appKey string) *ScheduleApplicationsConfig {
	return &ScheduleApplicationsConfig{
		ScheduleOptions: &scheduler.ScheduleOptions{
			AppKeys:            []string{appKey},
			StorageProvisioner: tests.Inst().Provisioner,
			Nodes:              c.ClusterInfo.storageLessNodes,
			Labels:             c.ClusterInfo.storageLessNodeLabels,
		},
		instanceID: tests.Inst().InstanceID,
		controller: c,
	}
}

// MultipleApplications returns a new ScheduleApplicationsConfig instance for scheduling multiple applications with ClusterController
func (c *ClusterController) MultipleApplications(appKeys []string) *ScheduleApplicationsConfig {
	return &ScheduleApplicationsConfig{
		ScheduleOptions: &scheduler.ScheduleOptions{
			AppKeys:            appKeys,
			StorageProvisioner: tests.Inst().Provisioner,
			Nodes:              c.ClusterInfo.storageLessNodes,
			Labels:             c.ClusterInfo.storageLessNodeLabels,
		},
		instanceID: tests.Inst().InstanceID,
		controller: c,
	}
}

// Namespace returns a new NamespaceConfig instance for managing the specified namespace with ClusterController
func (c *ClusterController) Namespace(namespace string) *NamespaceConfig {
	if !c.isNamespaceRecorded(namespace) {
		return &NamespaceConfig{
			namespace: namespace,
		}
	}
	namespaceInfo := c.getNamespaceInfo(namespace)
	return &NamespaceConfig{
		namespace: namespace,
		contexts:  namespaceInfo.contexts,
		DestroySchedulerContextConfig: &DestroySchedulerContextConfig{
			waitForDestroy:             DefaultWaitForDestroy,
			waitForResourceLeakCleanup: DefaultWaitForResourceLeakCleanup,
			skipClusterScopedObjects:   DefaultSkipClusterScopedObjects,
		},
		ValidateSchedulerContextConfig: &ValidateSchedulerContextConfig{
			waitForRunningTimeout:       DefaultWaitForRunningTimeout,
			waitForRunningRetryInterval: DefaultWaitForRunningRetryInterval,
			validateVolumeTimeout:       DefaultValidateVolumeTimeout,
			validateVolumeRetryInterval: DefaultValidateVolumeRetryInterval,
		},
		controller: c,
	}
}

// Cleanup cleans up all the cluster resources recorded in the ClusterController
func (c *ClusterController) Cleanup() (err error) {
	defer func() {
		err = utils.SwitchClusterContext(GlobalInClusterConfigPath)
	}()
	namespaces := c.getAllNamespaces()
	for _, namespace := range namespaces {
		log.Infof("Cleaning up namespace [%s]", namespace)
		err = c.Namespace(namespace).Destroy()
		if err != nil {
			debugMessage := fmt.Sprintf("namespace: [%s]", namespace)
			return utils.ProcessError(err, debugMessage)
		}
	}
	c.appKeyCountMap = make(map[string]int, 0)
	return nil
}

// Cluster returns a new ClusterInfo instance with the specified cluster id, name, and config-path
func Cluster(id int, name string, configPath string) *ClusterInfo {
	return &ClusterInfo{
		ClusterMetaData: &ClusterMetaData{
			Id:         fmt.Sprintf("%d", id),
			Name:       name,
			ConfigPath: configPath,
		},
	}
}

// AddClusterControllersToMap adds ClusterController instances to the specified map based on ClusterInfo objects, using `ClusterInfo.Name` as the key
func AddClusterControllersToMap(clusterControllerMap *map[string]*ClusterController, clustersInfo []*ClusterInfo) error {
	if *clusterControllerMap == nil {
		*clusterControllerMap = make(map[string]*ClusterController)
	}
	for _, clusterInfo := range clustersInfo {
		clusterController, err := clusterInfo.GetController()
		if err != nil {
			debugMessage := fmt.Sprintf("cluster: name [%s], config path [%s]", clusterInfo.Name, clusterInfo.ConfigPath)
			return utils.ProcessError(err, debugMessage)
		}

		(*clusterControllerMap)[clusterInfo.Name] = clusterController
	}
	return nil
}

// AddSourceClusterControllerToMap add SourceClusterController to the specified map, using utils.DefaultSourceClusterName as the key
func AddSourceClusterControllerToMap(clusterControllerMap *map[string]*ClusterController, id int) error {
	sourceClusterConfigPath, err := utils.GetSourceClusterConfigPath()
	if err != nil {
		return utils.ProcessError(err)
	}
	clustersInfo := []*ClusterInfo{
		Cluster(id, utils.DefaultSourceClusterName, sourceClusterConfigPath).IsHyperConverged().IsInCluster(),
	}
	return AddClusterControllersToMap(clusterControllerMap, clustersInfo)
}

// AddDestinationClusterControllerToMap add DestinationClusterController to the specified map, using utils.DefaultDestinationClusterName as the key
func AddDestinationClusterControllerToMap(clusterControllerMap *map[string]*ClusterController, id int) error {
	destinationClusterConfigPath, err := utils.GetDestinationClusterConfigPath()
	if err != nil {
		return utils.ProcessError(err)
	}
	clustersInfo := []*ClusterInfo{
		Cluster(id, utils.DefaultDestinationClusterName, destinationClusterConfigPath).IsHyperConverged(),
	}
	return AddClusterControllersToMap(clusterControllerMap, clustersInfo)
}
