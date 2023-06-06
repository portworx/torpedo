package cluster

import "github.com/portworx/torpedo/drivers/scheduler"

// import (
//
//	"fmt"
//	"github.com/portworx/torpedo/drivers/backup/utils"
//	"github.com/portworx/torpedo/drivers/scheduler"
//	"github.com/portworx/torpedo/pkg/log"
//	"github.com/portworx/torpedo/tests"
//	"strings"
//	"time"
//
// )
//
//	type ExecutionTimeRowData struct {
//		operation       string
//		operationStatus string
//		namespace       string
//		resourceMap     map[string]string
//		executionTime   utils.ExecutionTime
//	}
//
//	type NamespaceStatInfo struct {
//		executionTimeData []*ExecutionTimeRowData
//	}
//
//	func (i *NamespaceStatInfo) Print() {
//		for index, executionTimeRowData := range i.executionTimeData {
//			log.Infof("index: [%d]", index)
//			log.Infof("operation: [%s]", executionTimeRowData.operation)
//			log.Infof("operation status: [%s]", executionTimeRowData.operationStatus)
//			log.Infof("namespace: [%s]", executionTimeRowData.namespace)
//			log.Infof("resources: [%s]", executionTimeRowData.resourceMap)
//			log.Infof("duration: [%s]", executionTimeRowData.executionTime.TotalDuration)
//		}
//	}
//
// // appendContext appends the specified *scheduler.Context from NamespaceInfo
func (i *NamespaceInfo) appendContext(context *scheduler.Context) {
	i.contexts = append(i.contexts, context)
}

// removeContext removes the specified *scheduler.Context from NamespaceInfo
func (i *NamespaceInfo) removeContext(context *scheduler.Context) {
	remainingContexts := make([]*scheduler.Context, 0)
	for _, ctx := range i.contexts {
		if ctx.UID != context.UID {
			remainingContexts = append(remainingContexts, ctx)
		}
	}
	i.contexts = remainingContexts
}

// forgetContext forgets the specified *scheduler.Context
func (i *NamespaceInfo) forgetContext(context *scheduler.Context) {
	ctxIndex := -1
	for index, ctx := range i.contexts {
		if ctx.UID == context.UID {
			ctxIndex = index
			break
		}
	}
	if ctxIndex != -1 {
		i.contexts = append(i.contexts[:ctxIndex], i.contexts[ctxIndex+1:]...)
		i.forgottenContexts = append(i.forgottenContexts, context)
	}
}

//
//func (i *NamespaceInfo) saveExecutionTime(operation string, operationStatus string, startTime time.Time, namespace string, resourceMap map[string]string) {
//	endTime := time.Now()
//	executionTime := utils.NewExecutionTime(startTime, endTime)
//	executionTimeRowData := &ExecutionTimeRowData{
//		operation:       operation,
//		operationStatus: operationStatus,
//		namespace:       namespace,
//		resourceMap:     resourceMap,
//		executionTime:   executionTime,
//	}
//	i.executionTimeData = append(i.executionTimeData, executionTimeRowData)
//}
//

//
//// NamespaceConfig represents the configuration for managing namespaces within a ClusterController
//type NamespaceConfig struct {
//	namespace string
//	appKey    string
//	*NamespaceInfo
//	isAppKeySet    bool
//	skipValidation bool
//	*DestroySchedulerContextConfig
//	*ValidateSchedulerContextConfig
//	controller *ClusterController
//}
//
//func (c *NamespaceConfig) SkipValidation() *NamespaceConfig {
//	c.skipValidation = true
//	return c
//}
//
//// validate validates NamespaceConfig
//func (c *NamespaceConfig) validate() error {
//	if !c.skipValidation {
//		log.Infof("validate len contexts - %d", len(c.contexts))
//		if !c.controller.isNamespaceRecorded(c.namespace) {
//			err := fmt.Errorf("namespace [%s] is not in records", c.namespace)
//			return utils.ProcessError(err)
//		} else if len(c.contexts) == 0 {
//			if !c.isAppKeySet {
//				err := fmt.Errorf("no scheduler-contexts of the namespace [%s] are in records", c.namespace)
//				return utils.ProcessError(err)
//			} else {
//				err := fmt.Errorf("no scheduler-contexts of the namespace [%s] with the app-key [%s] are in records", c.namespace, c.appKey)
//				return utils.ProcessError(err)
//			}
//		}
//	}
//	return nil
//}
//
//// filterSchedulerContextsByAppKey filters the scheduler-contexts in NamespaceConfig by the specified appKey
//func (c *NamespaceConfig) filterSchedulerContextsByAppKey(appKey string) {
//	var appContexts []*scheduler.Context
//	for _, ctx := range c.contexts {
//		if strings.HasPrefix(ctx.App.Key, appKey+"-") {
//			appContexts = append(appContexts, ctx)
//		}
//	}
//	c.contexts = appContexts
//}
//
//// SelectApplication filters the NamespaceConfig to include only the scheduler-contexts associated with the specified appKey
//func (c *NamespaceConfig) SelectApplication(appKey string) *NamespaceConfig {
//	c.appKey, c.isAppKeySet = appKey, true
//	c.filterSchedulerContextsByAppKey(appKey)
//	return c
//}
//
//// Validate iterates through each scheduler-context in the NamespaceConfig and validates it
//func (c *NamespaceConfig) Validate() (err error) {
//	startTime := time.Now()
//	loopInComplete, loopIndex := true, -1
//	defer func() {
//		if loopInComplete && loopIndex != -1 {
//			c.controller.saveExecutionTime(GlobalValidateOperationLabel, startTime, c.contexts[loopIndex], err)
//		}
//	}()
//	err = c.validate()
//	if err != nil {
//		debugMessage := fmt.Sprintf("namespace-config: [%v]", c)
//		return utils.ProcessError(err, debugMessage)
//	}
//	err = c.controller.SwitchContext()
//	if err != nil {
//		return utils.ProcessError(err)
//	}
//	for index, ctx := range c.contexts {
//		startTime, loopIndex = time.Now(), index
//		log.Infof("Validating application [%s]", ctx.App.Key)
//		log.Infof("Waiting for application [%s] to start running", ctx.App.Key)
//		err = tests.Inst().S.WaitForRunning(ctx, c.waitForRunningTimeout, c.waitForRunningRetryInterval)
//		if err != nil {
//			debugMessage := fmt.Sprintf("scheduler-context: [%v]; wait-for-running: timeout [%s], retry-interval [%s]", ctx, c.waitForRunningTimeout, c.waitForRunningRetryInterval)
//			return utils.ProcessError(err, debugMessage)
//		}
//		if len(tests.Inst().TopologyLabels) > 0 {
//			log.Infof("Validating application [%s] topology labels", ctx.App.Key)
//			err = tests.Inst().S.ValidateTopologyLabel(ctx)
//			if err != nil {
//				debugMessage := fmt.Sprintf("scheduler-context: [%v]", ctx)
//				return utils.ProcessError(err, debugMessage)
//			}
//		}
//		if !ctx.SkipVolumeValidation {
//			log.Infof("Validating application [%s] volumes", ctx.App.Key)
//			err = tests.Inst().S.ValidateVolumes(ctx, c.validateVolumeTimeout, c.validateVolumeRetryInterval, nil)
//			if err != nil {
//				debugMessage := fmt.Sprintf("scheduler-context: [%v]; validate-volume: timeout [%s], retry-interval [%s]", ctx, c.validateVolumeTimeout, c.validateVolumeRetryInterval)
//				return utils.ProcessError(err, debugMessage)
//			}
//			volumeParameters, err := tests.Inst().S.GetVolumeParameters(ctx)
//			if err != nil {
//				debugMessage := fmt.Sprintf("scheduler-context: [%v]", ctx)
//				return utils.ProcessError(err, debugMessage)
//			}
//			for volumeName, volumeParams := range volumeParameters {
//				configMap := tests.Inst().ConfigMap
//				if configMap != "" {
//					volumeParams[GlobalAuthTokenParam], err = tests.Inst().S.GetTokenFromConfigMap(configMap)
//					if err != nil {
//						debugMessage := fmt.Sprintf("volume: name [%s], params [%v]; config-map: [%v]", volumeName, volumeParams, configMap)
//						return utils.ProcessError(err, debugMessage)
//					}
//				}
//				if ctx.RefreshStorageEndpoint {
//					volumeParams["refresh-endpoint"] = "true"
//				}
//				err = tests.Inst().V.ValidateCreateVolume(volumeName, volumeParams)
//				if err != nil {
//					debugMessage := fmt.Sprintf("volume: name [%s], params [%v]", volumeName, volumeParams)
//					return utils.ProcessError(err, debugMessage)
//				}
//			}
//			log.Infof("Validating if application [%s] volumes are setup", ctx.App.Key)
//			volumes, err := tests.Inst().S.GetVolumes(ctx)
//			if err != nil {
//				debugMessage := fmt.Sprintf("scheduler-context: [%v]", ctx)
//				return utils.ProcessError(err, debugMessage)
//			}
//			for _, volume := range volumes {
//				err = tests.Inst().V.ValidateVolumeSetup(volume)
//				if err != nil {
//					debugMessage := fmt.Sprintf("volume: [%s]", volume.String())
//					return utils.ProcessError(err, debugMessage)
//				}
//			}
//		}
//		c.controller.saveExecutionTime(GlobalValidateOperationLabel, startTime, c.contexts[loopIndex], err)
//	}
//	loopInComplete = false
//	return nil
//}
//
//// Destroy iterates through each scheduler-context in the NamespaceConfig and destroys it
//func (c *NamespaceConfig) Destroy() (err error) {
//	startTime := time.Now()
//	loopInComplete, loopIndex := true, -1
//	defer func() {
//		if loopInComplete && loopIndex != -1 {
//			c.controller.saveExecutionTime(GlobalValidateOperationLabel, startTime, c.contexts[loopIndex], err)
//		}
//	}()
//	err = c.validate()
//	if err != nil {
//		debugMessage := fmt.Sprintf("namespace-config: [%v]", c)
//		return utils.ProcessError(err, debugMessage)
//	}
//	err = c.controller.SwitchContext()
//	if err != nil {
//		return utils.ProcessError(err)
//	}
//	for index, ctx := range c.contexts {
//		startTime, loopIndex = time.Now(), index
//		log.Infof("Destroying application [%s]", ctx.App.Key)
//		volumeOptions := &scheduler.VolumeOptions{
//			SkipClusterScopedObjects: true,
//		}
//		volumes, err := tests.Inst().S.DeleteVolumes(ctx, volumeOptions)
//		if err != nil {
//			debugMessage := fmt.Sprintf("scheduler-context: [%v]; volume-options: [%v]", ctx, volumeOptions)
//			return utils.ProcessError(err, debugMessage)
//		}
//		destroyOptions := map[string]bool{
//			tests.SkipClusterScopedObjects:              c.skipClusterScopedObjects,
//			scheduler.OptionsWaitForDestroy:             c.waitForDestroy,
//			scheduler.OptionsWaitForResourceLeakCleanup: c.waitForResourceLeakCleanup,
//		}
//		err = tests.Inst().S.Destroy(ctx, destroyOptions)
//		if err != nil {
//			debugMessage := fmt.Sprintf("scheduler-context: [%v]; destroy-options: [%v]", ctx, destroyOptions)
//			return utils.ProcessError(err, debugMessage)
//		}
//		if !ctx.SkipVolumeValidation {
//			for _, vol := range volumes {
//				err := tests.Inst().V.ValidateDeleteVolume(vol)
//				if err != nil {
//					debugMessage := fmt.Sprintf("volume: [%v]", vol)
//					return utils.ProcessError(err, debugMessage)
//				}
//			}
//		}
//		// because we have already used DeleteVolumes with true
//		if !c.skipClusterScopedObjects {
//			volumeOptions = &scheduler.VolumeOptions{
//				SkipClusterScopedObjects: false,
//			}
//			_, err = tests.Inst().S.DeleteVolumes(ctx, volumeOptions)
//			if err != nil {
//				debugMessage := fmt.Sprintf("scheduler-context: [%v]; volume-options: [%v]", ctx, volumeOptions)
//				return utils.ProcessError(err, debugMessage)
//			}
//		}
//		c.controller.saveExecutionTime(GlobalValidateOperationLabel, startTime, c.contexts[loopIndex], err)
//		c.controller.forgetContext(c.namespace, ctx)
//	}
//	loopInComplete = false
//	return nil
//}
