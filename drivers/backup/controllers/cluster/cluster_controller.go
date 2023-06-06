package cluster

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/backup/utils"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/tests"
	"reflect"
	"time"
)

const (
	GlobalDestroyOperationLabel  = "Destroy"
	GlobalCleanupOperationLabel  = "Cleanup"
	GlobalValidateOperationLabel = "Validate"
)

const (
	// GlobalMaxK8sNamespaceCharLimit is the maximum character limit for a Kubernetes namespace
	GlobalMaxK8sNamespaceCharLimit = 63
)

const (
	// GlobalAuthTokenParam is an exact copy of the authTokenParam constant defined in the common.go file of the tests package
	GlobalAuthTokenParam = "auth-token"
	// GlobalInClusterConfigPath is the config path of the cluster in which Torpedo and Px-Backup are running, as described in the doc string of the SetConfig function in the k8s.go file of the k8s package
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
	// DefaultWaitForDestroy represents the default value of the key OptionsWaitForDestroy defined in the scheduler.go file of the scheduler package
	// DefaultWaitForDestroy indicates whether to wait for resources to be destroyed during the cleanup process
	DefaultWaitForDestroy = true
	// DefaultWaitForResourceLeakCleanup represents the default value of the key OptionsWaitForResourceLeakCleanup defined in the scheduler.go file of the scheduler package
	// DefaultWaitForResourceLeakCleanup indicates whether to wait for resource leak cleanup during the cleanup process
	DefaultWaitForResourceLeakCleanup = true
	// DefaultSkipClusterScopedObjects represents the default value of Context.SkipClusterScopedObject defined in the scheduler.go file of the scheduler package
	// DefaultSkipClusterScopedObjects indicates whether to skip cluster-scoped objects during the cleanup process
	DefaultSkipClusterScopedObjects = false
)

// ClusterController provides wrapper functions to streamline and simplify cluster related tasks
type ClusterController struct {
	*ClusterInfo
	namespaces     map[string]*NamespaceInfo
	appKeyCountMap map[string]int
}

// getNamespaceInfo returns the NamespaceInfo of the specified namespace from the ClusterController and a boolean indicating new creation
func (c *ClusterController) getNamespaceInfo(namespace string) (*NamespaceInfo, bool) {
	namespaceInfo, ok := c.namespaces[namespace]
	if isNew := !ok; isNew {
		return &NamespaceInfo{
			NamespaceStatInfo: &NamespaceStatInfo{},
		}, true
	}
	return namespaceInfo, false
}

// saveNamespaceInfo saves the NamespaceInfo of the specified namespace in the ClusterController
func (c *ClusterController) saveNamespaceInfo(namespace string, namespaceInfo *NamespaceInfo) {
	c.namespaces[namespace] = namespaceInfo
}

// delNamespaceInfo deletes the NamespaceInfo of the specified namespace from the ClusterController
func (c *ClusterController) delNamespaceInfo(namespace string) {
	delete(c.namespaces, namespace)
}

// isNamespaceRecorded returns true if the specified namespace is recorded in the ClusterController
func (c *ClusterController) isNamespaceRecorded(namespace string) bool {
	_, ok := c.namespaces[namespace]
	return ok
}

func (c *ClusterController) saveContext(namespace string, context *scheduler.Context) {
	namespaceInfo, isNew := c.getNamespaceInfo(namespace)
	namespaceInfo.appendContext(context)
	if isNew {
		c.saveNamespaceInfo(namespace, namespaceInfo)
	}
}

func (c *ClusterController) forgetContext(namespace string, context *scheduler.Context) {
	namespaceInfo, isNew := c.getNamespaceInfo(namespace)
	if !isNew {
		namespaceInfo.forgetContext(context)
	}
}

func (c *ClusterController) saveExecutionTime(operation string, startTime time.Time, context *scheduler.Context, err error) {
	resourceMap := make(map[string]string, 0)
	switch operation {
	case GlobalDestroyOperationLabel:
		resourceMap["App"] = context.App.Key
	case GlobalCleanupOperationLabel:
		for _, spec := range context.App.SpecList {
			specType := reflect.ValueOf(spec).Elem()
			kindField := specType.FieldByName("Kind")
			nameField := specType.FieldByName("Name")
			if kindField.IsValid() && nameField.IsValid() {
				resourceMap[kindField.String()] = nameField.String()
			}
		}
	case GlobalValidateOperationLabel:
		resourceMap["App"] = context.App.Key
	}
	namespaceInfo, isNew := c.getNamespaceInfo(context.ScheduleOptions.Namespace)
	if !isNew {
		var operationStatus string
		if err != nil {
			operationStatus = "FAILED"
		}
		namespaceInfo.saveExecutionTime(operation, operationStatus, startTime, context.ScheduleOptions.Namespace, resourceMap)
	}
}

// getAppKeyCount returns the count of occurrences of the specified app-key in the ClusterController
func (c *ClusterController) getAppKeyCount(appKey string) int {
	return c.appKeyCountMap[appKey]
}

// incrementAppKeyCount increments the count of the specified app-key in the ClusterController
func (c *ClusterController) incrementAppKeyCount(appKey string) {
	c.appKeyCountMap[appKey] += 1
}

// getNamespaces returns a slice of all the namespaces recorded in the ClusterController
func (c *ClusterController) getNamespaces() []string {
	namespaces := make([]string, 0)
	for namespace := range c.namespaces {
		namespaces = append(namespaces, namespace)
	}
	return namespaces
}

// SwitchContext switches the cluster context to the cluster specified by the configPath in the ClusterController
func (c *ClusterController) SwitchContext() error {
	err := utils.SwitchClusterContext(c.configPath)
	if err != nil {
		return utils.ProcessError(err, c.ClusterInfo.String())
	}
	return nil
}

// Application initializes ScheduleApplicationsConfig for scheduling single application using the ClusterController
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

// MultipleApplications initializes ScheduleApplicationsConfig for scheduling multiple applications using the ClusterController
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

// SelectNamespace initializes NamespaceConfig for managing specified namespace using the ClusterController
func (c *ClusterController) SelectNamespace(namespace string) *NamespaceConfig {
	namespaceInfo, isNew := c.getNamespaceInfo(namespace)
	if isNew {
		return &NamespaceConfig{
			namespace: namespace,
		}
	}
	return &NamespaceConfig{
		namespace:     namespace,
		NamespaceInfo: namespaceInfo.DeepCopy(),
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

// Cleanup cleans up all the resources created using the ClusterController
func (c *ClusterController) Cleanup() (err error) {
	//startTime := time.Now()
	loopInComplete, loopIndex := true, -1
	namespaces := c.getNamespaces()
	defer func() {
		if loopInComplete && loopIndex != -1 {
			//c.saveExecutionTime(GlobalCleanupOperationLabel, startTime, )
		}
		c.PrintStats()
		err = utils.SwitchClusterContext(GlobalInClusterConfigPath)
	}()
	for index, namespace := range namespaces {
		loopIndex = index
		log.Infof("Cleaning up namespace [%s]", namespace)
		//resources := c.getNamespaceResources(namespace)
		//resourceMap := make(map[string]string, 0)
		//for _, resource := range resources {
		//	if resourceMap[resource.kind] == "" {
		//		resourceMap[resource.kind] += resource.name
		//	} else {
		//		resourceMap[resource.kind] += fmt.Sprintf(",%s", resource.name)
		//	}
		//}
		err = c.SelectNamespace(namespace).SkipValidation().Destroy()
		if err != nil {
			debugMessage := fmt.Sprintf("namespace: [%s]", namespace)
			return utils.ProcessError(err, debugMessage)
		}
		//c.saveExecutionTime("Cleanup", namespace, resourceMap, startTime)
	}
	loopInComplete = false
	return nil
}

// PrintStats prints the destroyExecutionTime and validateExecutionTime for each namespace
func (c *ClusterController) PrintStats() {
	for namespace, namespaceInfo := range c.namespaces {
		log.Infof("Namespace: %s; Number of contexts: %d and Number of destroyed contexts: %d", namespace, len(namespaceInfo.contexts), len(namespaceInfo.forgottenContexts))
		namespaceInfo.NamespaceStatInfo.Print()
	}
	log.Infof("\n")
}

// Cluster returns ClusterInfo initialized with specified cluster id, name, and config-path
func Cluster(id int, name string, configPath string) *ClusterInfo {
	return &ClusterInfo{
		ClusterMetaData: &ClusterMetaData{
			id:         fmt.Sprintf("%d", id),
			name:       name,
			configPath: configPath,
		},
	}
}

// AddClusterControllersToMap adds ClusterController instances to the specified map based on ClusterInfo objects, using ClusterInfo.name as the key
func AddClusterControllersToMap(clusterControllerMap *map[string]*ClusterController, clustersInfo []*ClusterInfo) error {
	if *clusterControllerMap == nil {
		*clusterControllerMap = make(map[string]*ClusterController)
	}
	for _, clusterInfo := range clustersInfo {
		clusterController, err := clusterInfo.NewController()
		if err != nil {
			return utils.ProcessError(err, clusterInfo.String())
		}
		(*clusterControllerMap)[clusterInfo.name] = clusterController
	}
	return nil
}
