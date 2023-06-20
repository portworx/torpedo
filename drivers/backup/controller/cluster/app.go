package cluster

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/backup/utils"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/scheduler/k8s"
	"github.com/portworx/torpedo/drivers/scheduler/spec"
	"github.com/portworx/torpedo/tests"
	appsapi "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"reflect"
	"time"
)

const (
	GlobalAuthTokenParam = "auth-token" // copy of the const `authTokenParam` declared in the common.go file of the tests package
)

// ScheduleAppConfig represents the configuration for scheduling an App
type ScheduleAppConfig struct {
	ScheduleOptions *scheduler.ScheduleOptions
	InstanceID      string
}

// GetScheduleOptions returns the ScheduleOptions associated with the ScheduleAppConfig
func (c *ScheduleAppConfig) GetScheduleOptions() *scheduler.ScheduleOptions {
	return c.ScheduleOptions
}

// SetScheduleOptions sets the ScheduleOptions for the ScheduleAppConfig
func (c *ScheduleAppConfig) SetScheduleOptions(options *scheduler.ScheduleOptions) {
	c.ScheduleOptions = options
}

// GetInstanceID returns the InstanceID associated with the ScheduleAppConfig
func (c *ScheduleAppConfig) GetInstanceID() string {
	return c.InstanceID
}

// SetInstanceID sets the InstanceID for the ScheduleAppConfig
func (c *ScheduleAppConfig) SetInstanceID(instanceID string) {
	c.InstanceID = instanceID
}

// NewScheduleAppConfig creates a new instance of the ScheduleAppConfig
func NewScheduleAppConfig() *ScheduleAppConfig {
	newScheduleAppConfig := &ScheduleAppConfig{}
	newScheduleAppConfig.SetScheduleOptions(nil)
	newScheduleAppConfig.SetInstanceID(tests.Inst().InstanceID)
	return newScheduleAppConfig
}

// ValidateAppConfig represents the configuration for validating an App
type ValidateAppConfig struct {
	WaitForRunningTimeout       time.Duration
	WaitForRunningRetryInterval time.Duration
	ValidateVolumeTimeout       time.Duration
	ValidateVolumeRetryInterval time.Duration
}

// GetWaitForRunningTimeout returns the timeout duration for waiting for an App to be running in the ValidateAppConfig
func (c *ValidateAppConfig) GetWaitForRunningTimeout() time.Duration {
	return c.WaitForRunningTimeout
}

// SetWaitForRunningTimeout sets the timeout duration for waiting for an App to be running in the ValidateAppConfig
func (c *ValidateAppConfig) SetWaitForRunningTimeout(timeout time.Duration) {
	c.WaitForRunningTimeout = timeout
}

// GetWaitForRunningRetryInterval returns the retry interval duration for waiting for an App to be running in the ValidateAppConfig
func (c *ValidateAppConfig) GetWaitForRunningRetryInterval() time.Duration {
	return c.WaitForRunningRetryInterval
}

// SetWaitForRunningRetryInterval sets the retry interval duration for waiting for an App to be running in the ValidateAppConfig
func (c *ValidateAppConfig) SetWaitForRunningRetryInterval(interval time.Duration) {
	c.WaitForRunningRetryInterval = interval
}

// GetValidateVolumeTimeout returns the timeout duration for validating App volumes in the ValidateAppConfig
func (c *ValidateAppConfig) GetValidateVolumeTimeout() time.Duration {
	return c.ValidateVolumeTimeout
}

// SetValidateVolumeTimeout sets the timeout duration for validating App volumes in the ValidateAppConfig
func (c *ValidateAppConfig) SetValidateVolumeTimeout(timeout time.Duration) {
	c.ValidateVolumeTimeout = timeout
}

// GetValidateVolumeRetryInterval returns the retry interval duration for validating App volumes in the ValidateAppConfig
func (c *ValidateAppConfig) GetValidateVolumeRetryInterval() time.Duration {
	return c.ValidateVolumeRetryInterval
}

// SetValidateVolumeRetryInterval sets the retry interval duration for validating App volumes in the ValidateAppConfig
func (c *ValidateAppConfig) SetValidateVolumeRetryInterval(interval time.Duration) {
	c.ValidateVolumeRetryInterval = interval
}

// NewValidateAppConfig creates a new instance of the ValidateAppConfig
func NewValidateAppConfig() *ValidateAppConfig {
	newValidateAppConfig := &ValidateAppConfig{}
	newValidateAppConfig.SetWaitForRunningTimeout(DefaultWaitForRunningTimeout)
	newValidateAppConfig.SetWaitForRunningRetryInterval(DefaultWaitForRunningRetryInterval)
	newValidateAppConfig.SetValidateVolumeTimeout(DefaultValidateVolumeTimeout)
	newValidateAppConfig.SetValidateVolumeRetryInterval(DefaultValidateVolumeRetryInterval)
	return newValidateAppConfig
}

// TearDownAppConfig represents the configuration for tearing down an App
type TearDownAppConfig struct {
	WaitForDestroy             bool
	WaitForResourceLeakCleanup bool
	SkipClusterScopedObjects   bool
}

// GetWaitForDestroy returns the wait flag for waiting for App destruction in the TearDownAppConfig
func (c *TearDownAppConfig) GetWaitForDestroy() bool {
	return c.WaitForDestroy
}

// SetWaitForDestroy sets the wait flag for waiting for App destruction in the TearDownAppConfig
func (c *TearDownAppConfig) SetWaitForDestroy(wait bool) {
	c.WaitForDestroy = wait
}

// GetWaitForResourceLeakCleanup returns the wait flag for waiting for resource leak cleanup in the TearDownAppConfig
func (c *TearDownAppConfig) GetWaitForResourceLeakCleanup() bool {
	return c.WaitForResourceLeakCleanup
}

// SetWaitForResourceLeakCleanup sets the wait flag for waiting for resource leak cleanup in the TearDownAppConfig
func (c *TearDownAppConfig) SetWaitForResourceLeakCleanup(wait bool) {
	c.WaitForResourceLeakCleanup = wait
}

// GetSkipClusterScopedObjects returns the skip flag for skipping cluster-scoped object deletion in the TearDownAppConfig
func (c *TearDownAppConfig) GetSkipClusterScopedObjects() bool {
	return c.SkipClusterScopedObjects
}

// SetSkipClusterScopedObjects sets the skip flag for skipping cluster-scoped object deletion in the TearDownAppConfig
func (c *TearDownAppConfig) SetSkipClusterScopedObjects(skip bool) {
	c.SkipClusterScopedObjects = skip
}

// NewTearDownAppConfig creates a new instance of the TearDownAppConfig
func NewTearDownAppConfig() *TearDownAppConfig {
	newTearDownAppConfig := &TearDownAppConfig{}
	newTearDownAppConfig.SetWaitForDestroy(DefaultWaitForDestroy)
	newTearDownAppConfig.SetWaitForResourceLeakCleanup(DefaultWaitForResourceLeakCleanup)
	newTearDownAppConfig.SetSkipClusterScopedObjects(DefaultSkipClusterScopedObjects)
	return newTearDownAppConfig
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
	customAppSpec := utils.DeepCopyAppSpec(appSpec)
	switch tests.Inst().S.(type) {
	case *k8s.K8s:
		customAppSpec.Key += c.AppMetaData.GetSuffix()
		for _, spec := range customAppSpec.SpecList {
			specType := reflect.ValueOf(spec).Elem()
			nameField := specType.FieldByName("Name")
			if nameField.IsValid() {
				nameField.SetString(nameField.String() + c.AppMetaData.GetSuffix())
			}
			if obj, ok := spec.(*appsapi.Deployment); ok {
				numVolumes := len(obj.Spec.Template.Spec.Volumes)
				for i := 0; i < numVolumes; i++ {
					obj.Spec.Template.Spec.Volumes[i].PersistentVolumeClaim.ClaimName += c.AppMetaData.GetSuffix()
				}
			}
			if obj, ok := spec.(*corev1.PersistentVolumeClaim); ok {
				*obj.Spec.StorageClassName += c.AppMetaData.GetSuffix()
			}
		}
	}
	return customAppSpec, nil
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
	//		err := fmt.Errorf("app [%s] is already present in namespace [%s]", c.AppMetaData.GetNamespaceUid(), c.NamespaceMetaData.GetNamespaceUid())
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
	//appSpec, err := c.GetCustomAppSpec()
	//if err != nil {
	//	return utils.ProcessError(err)
	//}
	//appScheduleRequest := &AppScheduleRequest{
	//	Apps:            []*spec.AppSpec{appSpec},
	//	InstanceID:      c.ScheduleAppConfig.InstanceID,
	//	ScheduleOptions: *c.ScheduleAppConfig.ScheduleOptions,
	//}
	//cluster := c.ClusterController.ClusterManager.GetCluster("6d02ee80-448b-41a6-a866-b98a861d5590")
	//cluster = &Cluster{
	//	ContextManager: &ContextManager{
	//		DstConfigPath: "/tmp/source-config",
	//	},
	//	NamespaceManager: &NamespaceManager{
	//		Namespaces:        make(map[string]*Namespace, 0),
	//		RemovedNamespacesMap: make(map[string][]*Namespace, 0),
	//	},
	//}
	//
	//log.Infof("Scheduling app [%s] on namespace [%s]", c.AppMetaData.GetName(), c.NamespaceMetaData.Namespace)
	//resp, err := cluster.ProcessClusterRequest(appScheduleRequest)
	//if err != nil {
	//	return utils.ProcessError(err, utils.StructToString(appScheduleRequest))
	//}
	//if !cluster.NamespaceManager.IsNamespacePresent(c.NamespaceMetaData) {
	//	cluster.NamespaceManager.SetNamespace(c.NamespaceMetaData, NewNamespace())
	//}
	//appScheduleResponse := resp.(*cluster.AppScheduleResponse)
	//cluster.NamespaceManager.GetNamespace(c.NamespaceMetaData).AppManager.AddApp(c.AppMetaData, NewApp(appScheduleResponse.Contexts))
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
//map[string]bool{
//tests.SkipClusterScopedObjects:              c.SkipClusterScopedObjects,
//scheduler.OptionsWaitForDestroy:             c.WaitForDestroy,
//scheduler.OptionsWaitForResourceLeakCleanup: c.WaitForResourceLeakCleanup,
//}
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

type AppMetaData struct {
	AppKey     string
	Identifier []string
}

func (m *AppMetaData) HasIdentifier() bool {
	return m.Identifier == nil
}

func (m *AppMetaData) GetSuffix() string {
	//if !m.HasIdentifier() {
	//	return ""
	//}
	//return fmt.Sprintf("-%s", m.Identifier[0])
	return ""
}

func (m *AppMetaData) GetName() string {
	return m.AppKey + m.GetSuffix()
}

func NewAppMetaData(appKey string, identifier ...string) *AppMetaData {
	return &AppMetaData{
		AppKey:     appKey,
		Identifier: identifier,
	}
}

type App struct {
	Contexts []*scheduler.Context
}

func NewApp(contexts []*scheduler.Context) *App {
	return &App{
		Contexts: contexts,
	}
}

type AppManager struct {
	Apps        map[string]*App
	RemovedApps map[string][]*App
}

func (m *AppManager) GetApp(appMetaData *AppMetaData) *App {
	return m.Apps[appMetaData.GetName()]
}

func (m *AppManager) AddApp(appMetaData *AppMetaData, app *App) {
	m.Apps[appMetaData.GetName()] = app
}

func (m *AppManager) DeleteApp(appMetaData *AppMetaData) {
	delete(m.Apps, appMetaData.GetName())
}

func (m *AppManager) RemoveApp(appMetaData *AppMetaData) {
	m.RemovedApps[appMetaData.GetName()] = append(m.RemovedApps[appMetaData.GetName()], m.GetApp(appMetaData))
	m.DeleteApp(appMetaData)
}

func (m *AppManager) IsAppPresent(appMetaData *AppMetaData) bool {
	_, ok := m.Apps[appMetaData.GetName()]
	return ok
}

func NewAppManager() *AppManager {
	return &AppManager{
		Apps:        make(map[string]*App, 0),
		RemovedApps: make(map[string][]*App, 0),
	}
}
