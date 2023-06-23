package cluster

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/backup/utils"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/scheduler/k8s"
	"github.com/portworx/torpedo/drivers/scheduler/spec"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/tests"
	appsapi "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storageApi "k8s.io/api/storage/v1"
	"k8s.io/utils/pointer"
	"time"
)

type ScheduleAppConfig struct {
	ScheduleOptions *scheduler.ScheduleOptions
	InstanceID      string
}

type ValidateAppConfig struct {
	WaitForRunningTimeout       time.Duration
	WaitForRunningRetryInterval time.Duration
	ValidateVolumeTimeout       time.Duration
	ValidateVolumeRetryInterval time.Duration
}

type TearDownAppConfig struct {
	WaitForDestroy             bool
	WaitForResourceLeakCleanup bool
	SkipClusterScopedObjects   bool
}

func (c *TearDownAppConfig) GetDestroyOptions() map[string]bool {
	return map[string]bool{
		tests.SkipClusterScopedObjects:              c.SkipClusterScopedObjects,
		scheduler.OptionsWaitForDestroy:             c.WaitForDestroy,
		scheduler.OptionsWaitForResourceLeakCleanup: c.WaitForResourceLeakCleanup,
	}
}

type AppConfig struct {
	ClusterMetaData   *ClusterMetaData
	NamespaceMetaData *NamespaceMetaData
	AppMetaData       *AppMetaData
	ScheduleAppConfig *ScheduleAppConfig
	ValidateAppConfig *ValidateAppConfig
	TearDownAppConfig *TearDownAppConfig
	ClusterController *ClusterController
}

//func UpdateSpecNames(appSpec reflect.Value, identifier string) {
//	switch appSpec.Kind() {
//	case reflect.Ptr:
//		UpdateSpecNames(appSpec.Elem(), identifier)
//	case reflect.Struct:
//		for i := 0; i < appSpec.NumField(); i++ {
//			fieldVal := appSpec.Field(i)
//			fieldType := appSpec.Type().Field(i)
//			switch fieldType.Name {
//			case "Name":
//				if fieldVal.Kind() == reflect.String {
//					fieldVal.SetString(fieldVal.String() + identifier)
//				}
//			case "ClaimName":
//				if fieldVal.Kind() == reflect.String {
//					fieldVal.SetString(fieldVal.String() + identifier)
//				}
//			case "StorageClassName":
//				if fieldVal.Kind() == reflect.String {
//					fieldVal.SetString(fieldVal.String() + identifier)
//				}
//			default:
//				UpdateSpecNames(fieldVal, identifier)
//			}
//		}
//	case reflect.Slice:
//		for i := 0; i < appSpec.Len(); i++ {
//			UpdateSpecNames(appSpec.Index(i), identifier)
//		}
//	}
//}

func GetAppSpec(appKey string) (*spec.AppSpec, error) {
	var specFactory *spec.Factory
	var err error
	switch driver := tests.Inst().S.(type) {
	case *k8s.K8s:
		specFactory = driver.SpecFactory
	default:
		specDir := tests.Inst().SpecDir
		storageProvisioner := tests.Inst().V.String()
		parser := tests.Inst().S
		specFactory, err = spec.NewFactory(specDir, storageProvisioner, parser)
		if err != nil {
			debugStruct := struct {
				SpecDir            string
				StorageProvisioner string
				Parser             scheduler.Driver
			}{
				SpecDir:            specDir,
				StorageProvisioner: storageProvisioner,
				Parser:             parser,
			}
			return nil, utils.ProcessError(err, utils.StructToString(debugStruct))
		}
	}
	appSpec, err := specFactory.Get(appKey)
	if err != nil {
		debugStruct := struct {
			AppKey string
		}{
			AppKey: appKey,
		}
		return nil, utils.ProcessError(err, utils.StructToString(debugStruct))
	}
	return appSpec, nil
}

func (c *AppConfig) GetCustomAppSpec() (*spec.AppSpec, error) {
	// ToDo: customize all kinds of app specs
	if c.AppMetaData.HasIdentifier() {
		customizableAppKeys := []string{"postgres", "postgres-backup"}
		isCustomizable := false
		for _, customizableAppKey := range customizableAppKeys {
			if c.AppMetaData.AppKey == customizableAppKey {
				isCustomizable = true
			}
		}
		if !isCustomizable {
			err := fmt.Errorf("app-key [%s] is not customizable yet", c.AppMetaData.AppKey)
			return nil, utils.ProcessError(err)
		}
	}
	appSpec, err := GetAppSpec(c.AppMetaData.AppKey)
	if err != nil {
		debugStruct := struct {
			AppKey string
		}{
			AppKey: c.AppMetaData.AppKey,
		}
		return nil, utils.ProcessError(err, utils.StructToString(debugStruct))
	}
	appSpecWithIdentifier := utils.DeepCopyAppSpec(appSpec)
	switch tests.Inst().S.(type) {
	case *k8s.K8s:
		identifier := "-31313"
		appSpecWithIdentifier.Key += identifier
		for _, spec := range appSpecWithIdentifier.SpecList {
			//specType := reflect.ValueOf(spec).Elem()
			//metaField := specType.FieldByName("ObjectMeta")
			//if metaField.IsValid() {
			//	meta := metaField.Interface().(metav1.ObjectMeta)
			//	meta.Name += identifier
			//	metaField.Set(reflect.ValueOf(meta))
			//}
			switch obj := spec.(type) {
			case *appsapi.Deployment:
				obj.ObjectMeta.Name += identifier
				for i := 0; i < len(obj.Spec.Template.Spec.Volumes); i++ {
					if obj.Spec.Template.Spec.Volumes[i].PersistentVolumeClaim != nil {
						obj.Spec.Template.Spec.Volumes[i].PersistentVolumeClaim.ClaimName += identifier
					}
				}
			case *corev1.PersistentVolumeClaim:
				obj.ObjectMeta.Name += identifier
				obj.Spec.StorageClassName = pointer.String(*obj.Spec.StorageClassName + identifier)
			case *corev1.Secret:
				obj.ObjectMeta.Name += identifier
			case *corev1.Service:
				obj.ObjectMeta.Name += identifier
			case *appsapi.StatefulSet:
				obj.ObjectMeta.Name += identifier
			case *storageApi.StorageClass:
				obj.ObjectMeta.Name += identifier
				// Add more cases as needed
			}
		}
	}
	return appSpecWithIdentifier, nil
}

func (c *AppConfig) CanSchedule() error {
	//if !c.ClusterController.ClusterManager.IsClusterConfigRecorded(c.ClusterMetaData) {
	//	err := fmt.Errorf("cluster specified at [%s] is not present", c.ClusterMetaData.ConfigPath)
	//	return utils.ProcessError(err)
	//}
	//cluster := c.ClusterController.ClusterManager.GetCluster(c.ClusterMetaData)
	//if cluster.NamespaceManager.IsNamespacePresent(c.NamespaceMetaData) {
	//	namespace := cluster.NamespaceManager.GetNamespace(c.NamespaceMetaData)
	//	if namespace.AppManager.IsAppPresent(c.AppMetaData) {
	//		err := fmt.Errorf("app [%s] is already present in namespace [%s]", c.AppMetaData.GetName(), c.NamespaceMetaData.GetName())
	//		return utils.ProcessError(err)
	//	}
	//}
	return nil
}

func (c *AppConfig) Schedule() error {
	//err := c.CanSchedule()
	//if err != nil {
	//	return utils.ProcessError(err)
	//}
	appSpec, err := c.GetCustomAppSpec()
	if err != nil {
		return utils.ProcessError(err)
	}
	appScheduleRequest := &AppScheduleRequest{
		Apps:            []*spec.AppSpec{appSpec},
		InstanceID:      c.ScheduleAppConfig.InstanceID,
		ScheduleOptions: *c.ScheduleAppConfig.ScheduleOptions,
	}
	cluster := c.ClusterController.ClusterManager.GetCluster("6d02ee80-448b-41a6-a866-b98a861d5590")
	cluster = &Cluster{
		ContextManager: &ContextManager{
			DstConfigPath: "/tmp/source-config",
		},
		NamespaceManager: &NamespaceManager{
			Namespaces:        make(map[string]*Namespace, 0),
			RemovedNamespaces: make(map[string][]*Namespace, 0),
		},
	}

	log.Infof("Scheduling app [%s] on namespace [%s]", c.AppMetaData.GetName(), c.NamespaceMetaData.Namespace)
	resp, err := cluster.ProcessClusterRequest(appScheduleRequest)
	if err != nil {
		return utils.ProcessError(err, utils.StructToString(appScheduleRequest))
	}
	if !cluster.NamespaceManager.IsNamespacePresent(c.NamespaceMetaData) {
		cluster.NamespaceManager.AddNamespace(c.NamespaceMetaData, NewNamespace())
	}
	appScheduleResponse := resp.(*AppScheduleResponse)
	cluster.NamespaceManager.GetNamespace(c.NamespaceMetaData).AppManager.AddApp(c.AppMetaData, NewApp(appScheduleResponse.Contexts))
	return nil
}

//func (c *AppConfig) canValidate() error {
//	return nil
//}
//
//func (c *AppConfig) Validate() error {
//	err := c.canTearDown()
//	if err != nil {
//		return utils.ProcessError(err, c.String())
//	}
//	err = c.ClusterController.SwitchContext()
//	if err != nil {
//		return utils.ProcessError(err)
//	}
//	Contexts := c.ClusterController.NamespaceManager.GetNamespace(c.NamespaceMetaData).AppManager.GetApp(c.AppMetaData).Contexts
//	for _, ctx := range Contexts {
//		log.Infof("Validating application [%s]", ctx.App.Key)
//		log.Infof("Waiting for application [%s] to start running", ctx.App.Key)
//		err = tests.Inst().S.WaitForRunning(ctx, c.ValidateAppConfig.WaitForRunningTimeout, c.ValidateAppConfig.waitForRunningRetryInterval)
//		if err != nil {
//			debugMessage := fmt.Sprintf("scheduler-context: [%v]; wait-for-running: timeout [%s], retry-interval [%s]", ctx, c.ValidateAppConfig.WaitForRunningTimeout, c.ValidateAppConfig.waitForRunningRetryInterval)
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
//			err = tests.Inst().S.ValidateVolumes(ctx, c.ValidateAppConfig.ValidateVolumeTimeout, c.ValidateAppConfig.ValidateVolumeRetryInterval, nil)
//			if err != nil {
//				debugMessage := fmt.Sprintf("scheduler-context: [%v]; validate-volume: timeout [%s], retry-interval [%s]", ctx, c.ValidateAppConfig.ValidateVolumeTimeout, c.ValidateAppConfig.ValidateVolumeRetryInterval)
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
//	}
//	return nil
//}
//
//func (c *AppConfig) canTearDown() error {
//	return nil
//}
//
//func (c *AppConfig) TearDown() error {
//	err := c.canTearDown()
//	if err != nil {
//		return utils.ProcessError(err, c.String())
//	}
//	err = c.ClusterController.SwitchContext()
//	if err != nil {
//		return utils.ProcessError(err)
//	}
//	log.Infof("Destroying application [%s]", c.AppMetaData.GetUid())
//	Contexts := c.ClusterController.NamespaceManager.GetNamespace(c.NamespaceMetaData).AppManager.GetApp(c.AppMetaData).Contexts
//	log.Infof("len of tear down Contexts %d", len(Contexts))
//	for _, ctx := range Contexts {
//		volumeOptions := &scheduler.VolumeOptions{
//			SkipClusterScopedObjects: true,
//		}
//		volumes, err := tests.Inst().S.DeleteVolumes(ctx, volumeOptions)
//		if err != nil {
//			debugMessage := fmt.Sprintf("scheduler-context: [%v]; volume-options: [%v]", ctx, volumeOptions)
//			return utils.ProcessError(err, debugMessage)
//		}
//		destroyOptions := c.TearDownAppConfig.GetDestroyOptions()
//		log.Infof("destroy options %v", destroyOptions)
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
//		if !c.TearDownAppConfig.SkipClusterScopedObjects {
//			volumeOptions = &scheduler.VolumeOptions{
//				SkipClusterScopedObjects: false,
//			}
//			_, err = tests.Inst().S.DeleteVolumes(ctx, volumeOptions)
//			if err != nil {
//				debugMessage := fmt.Sprintf("scheduler-context: [%v]; volume-options: [%v]", ctx, volumeOptions)
//				return utils.ProcessError(err, debugMessage)
//			}
//		}
//		c.ClusterController.NamespaceManager.GetNamespace(c.NamespaceMetaData).AppManager.RemoveApp(c.AppMetaData)
//		log.Infof("present Apps %d", len(c.ClusterController.NamespaceManager.GetNamespace(c.NamespaceMetaData).AppManager.Apps))
//		log.Infof("removed Apps %d", len(c.ClusterController.NamespaceManager.GetNamespace(c.NamespaceMetaData).AppManager.RemovedApps))
//	}
//	return nil
//}
