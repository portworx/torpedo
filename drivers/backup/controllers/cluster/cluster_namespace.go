package cluster

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/backup/utils"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/tests"
	"strings"
	"time"
)

// ValidateSchedulerContextConfig holds the necessary configuration to effectively validate scheduler-context
type ValidateSchedulerContextConfig struct {
	waitForRunningTimeout       time.Duration
	waitForRunningRetryInterval time.Duration
	validateVolumeTimeout       time.Duration
	validateVolumeRetryInterval time.Duration
}

// DestroySchedulerContextConfig holds the necessary configuration to effectively destroy scheduler-context
type DestroySchedulerContextConfig struct {
	waitForDestroy             bool
	waitForResourceLeakCleanup bool
	skipClusterScopedObjects   bool
}

// NamespaceConfig holds the necessary configuration for managing namespaces within a ClusterController
type NamespaceConfig struct {
	namespace  string
	appKey     *string
	contexts   []*scheduler.Context
	controller *ClusterController
	*ValidateSchedulerContextConfig
	*DestroySchedulerContextConfig
}

// validateNamespaceConfig validates NamespaceConfig
func validateNamespaceConfig(c *NamespaceConfig) error {
	if c.contexts == nil {
		err := fmt.Errorf("no scheduler-contexts of the namespace [%s] found", c.namespace)
		return utils.ProcessError(err)
	} else if len(c.contexts) == 0 {
		if c.appKey == nil {
			err := fmt.Errorf("no scheduler-contexts of the namespace [%s] found", c.namespace)
			return utils.ProcessError(err)
		} else {
			err := fmt.Errorf("no scheduler-contexts of the namespace [%s] with the app-key [%s] found", c.namespace, *c.appKey)
			return utils.ProcessError(err)
		}
	}
	return nil
}

// Application filters the NamespaceConfig to include only the scheduler-contexts associated with the specified appKey
func (c *NamespaceConfig) Application(appKey string) *NamespaceConfig {
	if c.contexts == nil {
		return c
	}
	c.appKey = &appKey
	appContexts := make([]*scheduler.Context, 0)
	for _, ctx := range c.contexts {
		log.Infof("%s -- %s", ctx.App.Key, appKey+"-")
		if strings.HasPrefix(ctx.App.Key, appKey+"-") {
			appContexts = append(appContexts, ctx)
		}
	}
	c.contexts = appContexts
	return c
}

// Validate iterates through each scheduler-context in the NamespaceConfig and validates it
func (c *NamespaceConfig) Validate() error {
	err := validateNamespaceConfig(c)
	if err != nil {
		debugMessage := fmt.Sprintf("namespace-config: [%v]", c)
		return utils.ProcessError(err, debugMessage)
	}
	err = c.controller.SwitchContext()
	if err != nil {
		return utils.ProcessError(err)
	}
	for _, ctx := range c.contexts {
		log.Infof("Validating application [%s]", ctx.App.Key)
		log.Infof("Waiting for application [%s] to start running", ctx.App.Key)
		err = tests.Inst().S.WaitForRunning(ctx, c.waitForRunningTimeout, c.waitForRunningRetryInterval)
		if err != nil {
			debugMessage := fmt.Sprintf("scheduler-context: [%v]; wait-for-running: timeout [%s], retry-interval [%s]", ctx, c.waitForRunningTimeout, c.waitForRunningRetryInterval)
			return utils.ProcessError(err, debugMessage)
		}
		if len(tests.Inst().TopologyLabels) > 0 {
			log.Infof("Validating application [%s] topology labels", ctx.App.Key)
			err = tests.Inst().S.ValidateTopologyLabel(ctx)
			if err != nil {
				debugMessage := fmt.Sprintf("scheduler-context: [%v]", ctx)
				return utils.ProcessError(err, debugMessage)
			}
		}
		if !ctx.SkipVolumeValidation {
			log.Infof("Validating application [%s] volumes", ctx.App.Key)
			err = tests.Inst().S.ValidateVolumes(ctx, c.validateVolumeTimeout, c.validateVolumeRetryInterval, nil)
			if err != nil {
				debugMessage := fmt.Sprintf("scheduler-context: [%v]; validate-volume: timeout [%s], retry-interval [%s]", ctx, c.validateVolumeTimeout, c.validateVolumeRetryInterval)
				return utils.ProcessError(err, debugMessage)
			}
			volumeParameters, err := tests.Inst().S.GetVolumeParameters(ctx)
			if err != nil {
				debugMessage := fmt.Sprintf("scheduler-context: [%v]", ctx)
				return utils.ProcessError(err, debugMessage)
			}
			for volumeName, volumeParams := range volumeParameters {
				configMap := tests.Inst().ConfigMap
				if configMap != "" {
					volumeParams[GlobalAuthTokenParam], err = tests.Inst().S.GetTokenFromConfigMap(configMap)
					if err != nil {
						debugMessage := fmt.Sprintf("volume: name [%s], params [%v]; config-map: [%v]", volumeName, volumeParams, configMap)
						return utils.ProcessError(err, debugMessage)
					}
				}
				if ctx.RefreshStorageEndpoint {
					volumeParams["refresh-endpoint"] = "true"
				}
				err = tests.Inst().V.ValidateCreateVolume(volumeName, volumeParams)
				if err != nil {
					debugMessage := fmt.Sprintf("volume: name [%s], params [%v]", volumeName, volumeParams)
					return utils.ProcessError(err, debugMessage)
				}
			}
			log.Infof("Validating if application [%s] volumes are setup", ctx.App.Key)
			volumes, err := tests.Inst().S.GetVolumes(ctx)
			if err != nil {
				debugMessage := fmt.Sprintf("scheduler-context: [%v]", ctx)
				return utils.ProcessError(err, debugMessage)
			}
			for _, volume := range volumes {
				err = tests.Inst().V.ValidateVolumeSetup(volume)
				if err != nil {
					debugMessage := fmt.Sprintf("volume: [%v]", volume)
					return utils.ProcessError(err, debugMessage)
				}
			}
		}
	}
	return nil
}

// Destroy iterates through each scheduler-context in the NamespaceConfig and destroys it
func (c *NamespaceConfig) Destroy() error {
	err := validateNamespaceConfig(c)
	if err != nil {
		debugMessage := fmt.Sprintf("namespace-config:  [%v]", c)
		return utils.ProcessError(err, debugMessage)
	}
	err = c.controller.SwitchContext()
	if err != nil {
		return utils.ProcessError(err)
	}
	for _, ctx := range c.contexts {
		log.Infof("Destroying application [%s]", ctx.App.Key)
		volumeOptions := &scheduler.VolumeOptions{
			SkipClusterScopedObjects: true,
		}
		volumes, err := tests.Inst().S.DeleteVolumes(ctx, volumeOptions)
		if err != nil {
			debugMessage := fmt.Sprintf("scheduler-context: [%v]; volume-options: [%v]", ctx, volumeOptions)
			return utils.ProcessError(err, debugMessage)
		}
		destroyOptions := map[string]bool{
			tests.SkipClusterScopedObjects:              c.skipClusterScopedObjects,
			scheduler.OptionsWaitForDestroy:             c.waitForDestroy,
			scheduler.OptionsWaitForResourceLeakCleanup: c.waitForResourceLeakCleanup,
		}
		err = tests.Inst().S.Destroy(ctx, destroyOptions)
		if err != nil {
			debugMessage := fmt.Sprintf("scheduler-context: [%v]; destroy-options: [%v]", ctx, destroyOptions)
			return utils.ProcessError(err, debugMessage)
		}
		if !ctx.SkipVolumeValidation {
			for _, vol := range volumes {
				err := tests.Inst().V.ValidateDeleteVolume(vol)
				if err != nil {
					debugMessage := fmt.Sprintf("volume: [%v]", vol)
					return utils.ProcessError(err, debugMessage)
				}
			}
		}
		// because we have already used DeleteVolumes with true
		if !c.skipClusterScopedObjects {
			volumeOptions = &scheduler.VolumeOptions{
				SkipClusterScopedObjects: false,
			}
			_, err = tests.Inst().S.DeleteVolumes(ctx, volumeOptions)
			if err != nil {
				debugMessage := fmt.Sprintf("scheduler-context: [%v]; volume-options: [%v]", ctx, volumeOptions)
				return utils.ProcessError(err, debugMessage)
			}
		}
	}
	namespaceInfo := c.controller.getNamespaceInfo(c.namespace)
	namespaceInfo.removeContexts(c.contexts)
	c.controller.saveNamespaceInfo(c.namespace, namespaceInfo)
	return nil
}
