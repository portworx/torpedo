package cluster

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/backup/utils"
	"github.com/portworx/torpedo/drivers/scheduler/k8s"
	"github.com/portworx/torpedo/drivers/scheduler/spec"
	"github.com/portworx/torpedo/tests"
	appsapi "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"reflect"
)

// // ScheduleApplicationsConfig represents the configuration to effectively schedule applications
//
//	type ScheduleApplicationsConfig struct {
//		*scheduler.ScheduleOptions
//		instanceID string
//		controller *ClusterController
//	}
//
// // validate validates ScheduleApplicationsConfig
//
//	func (c *ScheduleApplicationsConfig) validate() error {
//		if len(c.Namespace) > GlobalMaxK8sNamespaceCharLimit {
//			err := fmt.Errorf("namespace [%s] exceeds the maximum character limit of [%d]", c.Namespace, GlobalMaxK8sNamespaceCharLimit)
//			return utils.ProcessError(err)
//		}
//		return nil
//	}
//
// getAppsSpec returns the slice of AppSpecs associated with the app-keys defined in the ScheduleApplicationsConfig
func (c *ApplicationConfig) getAppsSpec() ([]*spec.AppSpec, error) {
	var appsSpec []*spec.AppSpec
	if len(c.AppKeys) > 0 {
		var specFactory *spec.Factory
		switch driver := tests.Inst().S.(type) {
		case *k8s.K8s:
			specFactory = driver.SpecFactory
		default:
			var err error
			specFactory, err = spec.NewFactory(tests.Inst().SpecDir, tests.Inst().V.String(), tests.Inst().S)
			if err != nil {
				debugMessage := fmt.Sprintf("spec-directory: [%s]; volume-driver: [%s]; scheduler: [%s]", tests.Inst().SpecDir, tests.Inst().V.String(), tests.Inst().S.String())
				return nil, utils.ProcessError(err, debugMessage)
			}
		}
		for _, appKey := range c.AppKeys {
			appSpec, err := specFactory.Get(appKey)
			if err != nil {
				debugMessage := fmt.Sprintf("app-keys [%s]: app-key [%s]", c.AppKeys, appKey)
				return nil, utils.ProcessError(err, debugMessage)
			}
			appsSpec = append(appsSpec, appSpec)
		}
	} else {
		err := fmt.Errorf("the app list cannot be empty")
		return nil, utils.ProcessError(err)
	}
	return appsSpec, nil
}

// getCustomAppsSpec returns a slice of customized AppsSpec by adding unique suffixes for spec names
// getCustomAppsSpec currently supports only postgres and postgres-backup
// TODO: Extend functionality to accommodate other applications
func (c *ApplicationConfig) getCustomAppsSpec() ([]*spec.AppSpec, error) {
	appsSpec, err := c.getAppsSpec()
	if err != nil {
		return nil, utils.ProcessError(err)
	}
	customAppsSpec := make([]*spec.AppSpec, len(appsSpec))
	for index, appSpec := range appsSpec {
		customAppsSpec[index] = utils.DeepCopyAppSpec(appSpec)
	}
	switch tests.Inst().S.(type) {
	case *k8s.K8s:
		for _, appSpec := range customAppsSpec {
			appKeyCount := c.controller.getAppKeyCount(appSpec.Key)
			specNameSuffix := fmt.Sprintf("-%s-%d", c.controller.id, appKeyCount)
			c.controller.incrementAppKeyCount(appSpec.Key)
			appSpec.Key += specNameSuffix
			for _, spec := range appSpec.SpecList {
				specType := reflect.ValueOf(spec).Elem()
				nameField := specType.FieldByName("Name")
				if nameField.IsValid() {
					nameField.SetString(nameField.String() + specNameSuffix)
				}
				if obj, ok := spec.(*appsapi.Deployment); ok {
					numVolumes := len(obj.Spec.Template.Spec.Volumes)
					for i := 0; i < numVolumes; i++ {
						obj.Spec.Template.Spec.Volumes[i].PersistentVolumeClaim.ClaimName += specNameSuffix
					}
				}
				if obj, ok := spec.(*corev1.PersistentVolumeClaim); ok {
					*obj.Spec.StorageClassName += specNameSuffix
				}
			}
		}
	}
	return customAppsSpec, nil
}

//
//// ScheduleOnNamespace schedules applications on the specified namespace
//func (c *ScheduleApplicationsConfig) ScheduleOnNamespace(namespace string) error {
//	err := c.validate()
//	if err != nil {
//		return utils.ProcessError(err)
//	}
//	err = c.controller.SwitchContext()
//	if err != nil {
//		return utils.ProcessError(err)
//	}
//	customAppsSpec, err := c.getCustomAppsSpec()
//	if err != nil {
//		return utils.ProcessError(err)
//	}
//	c.Namespace = namespace
//	log.Infof("Scheduling applications [%s] on namespace [%s]", c.AppKeys, c.Namespace)
//	contexts, err := tests.Inst().S.ScheduleWithCustomAppSpecs(customAppsSpec, c.instanceID, *c.ScheduleOptions)
//	if err != nil {
//		debugMessage := fmt.Sprintf("namespace: [%s]; custom-apps-spec: [%v]; instance-id: [%s]; schedule-options: [%v]", namespace, customAppsSpec, c.instanceID, *c.ScheduleOptions)
//		return utils.ProcessError(err, debugMessage)
//	}
//	for _, ctx := range contexts {
//		c.controller.saveContext(namespace, ctx)
//	}
//	return nil
//}
//
//// ScheduleOnMultipleNamespaces schedules applications on multiple namespaces using the namespace prefix numSchedules times
//func (c *ScheduleApplicationsConfig) ScheduleOnMultipleNamespaces(namespacePrefix string, numSchedules int) ([]string, error) {
//	err := c.controller.SwitchContext()
//	if err != nil {
//		return nil, utils.ProcessError(err)
//	}
//	namespaces := make([]string, 0)
//	for i := 0; i < numSchedules; i++ {
//		namespaceSuffix := fmt.Sprintf("-%s-%d", c.controller.id, i)
//		namespace := namespacePrefix + namespaceSuffix
//		err = c.ScheduleOnNamespace(namespace)
//		if err != nil {
//			debugMessage := fmt.Sprintf("namespace [%s]: prefix [%s], suffix [%s]", namespace, namespacePrefix, namespaceSuffix)
//			return nil, utils.ProcessError(err, debugMessage)
//		}
//		namespaces = append(namespaces, namespace)
//	}
//	return namespaces, nil
//}
