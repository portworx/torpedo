package cluster

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/backup/controller/cluster/driverapi"
	"github.com/portworx/torpedo/drivers/backup/utils"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/scheduler/k8s"
	"github.com/portworx/torpedo/drivers/scheduler/spec"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/tests"
	appsapi "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"reflect"
	"sync"
	"time"
)

const (
	GlobalAuthTokenParam = "auth-token" // copy of the const `authTokenParam` declared in the common.go file of the tests package
)

// AppMetaData represents the metadata for an App
type AppMetaData struct {
	AppKey     string
	Identifier string
}

// GetAppKey returns the AppKey associated with the AppMetaData
func (m *AppMetaData) GetAppKey() string {
	return m.AppKey
}

// SetAppKey sets the AppKey for the AppMetaData
func (m *AppMetaData) SetAppKey(appKey string) {
	m.AppKey = appKey
}

// GetIdentifier returns the Identifier associated with the AppMetaData
func (m *AppMetaData) GetIdentifier() string {
	return m.Identifier
}

// SetIdentifier sets the Identifier for the AppMetaData
func (m *AppMetaData) SetIdentifier(identifier string) {
	m.Identifier = identifier
}

// IsIdentifierPresent checks if an Identifier is present in the AppMetaData
func (m *AppMetaData) IsIdentifierPresent() bool {
	return m.GetIdentifier() != ""
}

// GetAppSuffix returns the App suffix based on the Identifier in the AppMetaData
func (m *AppMetaData) GetAppSuffix() string {
	if !m.IsIdentifierPresent() {
		return ""
	}
	return fmt.Sprintf("-%s", m.GetIdentifier())
}

// GetApp returns the App associated with the metadata
func (m *AppMetaData) GetApp() string {
	return m.GetAppKey() + m.GetAppSuffix()
}

// GetAppUid returns the App uid
func (m *AppMetaData) GetAppUid() string {
	return m.GetApp()
}

// NewAppMetaData creates a new instance of the AppMetaData
func NewAppMetaData() *AppMetaData {
	newAppConfig := &AppMetaData{}
	newAppConfig.SetAppKey("")
	newAppConfig.SetIdentifier("")
	return newAppConfig
}

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
func (c *ValidateAppConfig) SetWaitForRunningRetryInterval(retryInterval time.Duration) {
	c.WaitForRunningRetryInterval = retryInterval
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
func (c *ValidateAppConfig) SetValidateVolumeRetryInterval(retryInterval time.Duration) {
	c.ValidateVolumeRetryInterval = retryInterval
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

// AppConfig represents the configuration for an App
type AppConfig struct {
	ClusterMetaData   *ClusterMetaData
	NamespaceMetaData *NamespaceMetaData
	AppMetaData       *AppMetaData
	ScheduleAppConfig *ScheduleAppConfig
	ValidateAppConfig *ValidateAppConfig
	TearDownAppConfig *TearDownAppConfig
	ClusterController *ClusterController
}

// GetClusterMetaData returns the ClusterMetaData associated with the AppConfig
func (c *AppConfig) GetClusterMetaData() *ClusterMetaData {
	return c.ClusterMetaData
}

// SetClusterMetaData sets the ClusterMetaData for the AppConfig
func (c *AppConfig) SetClusterMetaData(metaData *ClusterMetaData) {
	c.ClusterMetaData = metaData
}

// GetNamespaceMetaData returns the NamespaceMetaData associated with the AppConfig
func (c *AppConfig) GetNamespaceMetaData() *NamespaceMetaData {
	return c.NamespaceMetaData
}

// SetNamespaceMetaData sets the NamespaceMetaData for the AppConfig
func (c *AppConfig) SetNamespaceMetaData(metaData *NamespaceMetaData) {
	c.NamespaceMetaData = metaData
}

// GetAppMetaData returns the AppMetaData associated with the AppConfig
func (c *AppConfig) GetAppMetaData() *AppMetaData {
	return c.AppMetaData
}

// SetAppMetaData sets the AppMetaData for the AppConfig
func (c *AppConfig) SetAppMetaData(metaData *AppMetaData) {
	c.AppMetaData = metaData
}

// GetScheduleAppConfig returns the ScheduleAppConfig associated with the App
func (c *AppConfig) GetScheduleAppConfig() *ScheduleAppConfig {
	return c.ScheduleAppConfig
}

// SetScheduleAppConfig sets the ScheduleAppConfig for the App
func (c *AppConfig) SetScheduleAppConfig(config *ScheduleAppConfig) {
	c.ScheduleAppConfig = config
}

// GetValidateAppConfig returns the ValidateAppConfig associated with the App
func (c *AppConfig) GetValidateAppConfig() *ValidateAppConfig {
	return c.ValidateAppConfig
}

// SetValidateAppConfig sets the ValidateAppConfig for the App
func (c *AppConfig) SetValidateAppConfig(config *ValidateAppConfig) {
	c.ValidateAppConfig = config
}

// GetTearDownAppConfig returns the TearDownAppConfig associated with the App
func (c *AppConfig) GetTearDownAppConfig() *TearDownAppConfig {
	return c.TearDownAppConfig
}

// SetTearDownAppConfig sets the TearDownAppConfig for the App
func (c *AppConfig) SetTearDownAppConfig(config *TearDownAppConfig) {
	c.TearDownAppConfig = config
}

// GetClusterController returns the ClusterController associated with the AppConfig
func (c *AppConfig) GetClusterController() *ClusterController {
	return c.ClusterController
}

// SetClusterController sets the ClusterController for the AppConfig
func (c *AppConfig) SetClusterController(controller *ClusterController) {
	c.ClusterController = controller
}

// NewAppConfig creates a new instance of the AppConfig
func NewAppConfig() *AppConfig {
	newAppConfig := &AppConfig{}
	newAppConfig.SetClusterMetaData(NewClusterMetaData())
	newAppConfig.SetNamespaceMetaData(NewNamespaceMetaData())
	newAppConfig.SetAppMetaData(NewAppMetaData())
	newAppConfig.SetScheduleAppConfig(NewScheduleAppConfig())
	newAppConfig.SetValidateAppConfig(NewValidateAppConfig())
	newAppConfig.SetTearDownAppConfig(NewTearDownAppConfig())
	newAppConfig.SetClusterController(nil)
	return newAppConfig
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

func (c *AppConfig) GetAppSpecWithIdentifier() (*spec.AppSpec, error) {
	// TODO: Associate all apps with an identifier
	if c.GetAppMetaData().IsIdentifierPresent() {
		identifiableAppKeys := []string{"postgres", "postgres-backup"}
		isIdentifiable := false
		for _, identifiableAppKey := range identifiableAppKeys {
			if identifiableAppKey == c.GetAppMetaData().GetAppKey() {
				isIdentifiable = true
			}
		}
		if !isIdentifiable {
			err := fmt.Errorf("app-key [%s] cannot be associated with an identifier yet", c.GetAppMetaData().GetApp())
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
		appSpecWithIdentifier.Key += c.GetAppMetaData().GetAppSuffix()
		for _, spec := range appSpecWithIdentifier.SpecList {
			specType := reflect.ValueOf(spec).Elem()
			nameField := specType.FieldByName("Name")
			if nameField.IsValid() {
				nameField.SetString(nameField.String() + c.AppMetaData.GetAppSuffix())
			}
			if obj, ok := spec.(*appsapi.Deployment); ok {
				numVolumes := len(obj.Spec.Template.Spec.Volumes)
				for i := 0; i < numVolumes; i++ {
					obj.Spec.Template.Spec.Volumes[i].PersistentVolumeClaim.ClaimName += c.AppMetaData.GetAppSuffix()
				}
			}
			if obj, ok := spec.(*corev1.PersistentVolumeClaim); ok {
				*obj.Spec.StorageClassName += c.AppMetaData.GetAppSuffix()
			}
		}
	}
	return appSpecWithIdentifier, nil
}

func (c *AppConfig) CanSchedule() error {
	if !c.GetClusterController().ClusterManager.IsClusterPresent(c.ClusterMetaData.GetClusterUid()) {
		err := fmt.Errorf("cluster specified at [%s] is not present", c.ClusterMetaData.ConfigPath)
		return utils.ProcessError(err)
	}
	cluster := c.ClusterController.ClusterManager.GetCluster(c.ClusterMetaData.GetClusterUid())
	if cluster.NamespaceManager.IsNamespacePresent(c.NamespaceMetaData.GetNamespaceUid()) {
		namespace := cluster.NamespaceManager.GetNamespace(c.NamespaceMetaData.GetNamespaceUid())
		if namespace.AppManager.IsAppPresent(c.AppMetaData.GetAppUid()) {
			err := fmt.Errorf("app [%s] is already present in namespace [%s]", c.AppMetaData.GetAppUid(), c.NamespaceMetaData.GetNamespaceUid())
			return utils.ProcessError(err)
		}
	}
	return nil
}

func (c *AppConfig) Schedule() error {
	err := c.CanSchedule()
	if err != nil {
		return utils.ProcessError(err)
	}
	appSpec, err := c.GetAppSpecWithIdentifier()
	if err != nil {
		return utils.ProcessError(err)
	}
	appScheduleRequest := &driverapi.AppScheduleRequest{
		Apps:            []*spec.AppSpec{appSpec},
		InstanceID:      c.ScheduleAppConfig.InstanceID,
		ScheduleOptions: *c.ScheduleAppConfig.ScheduleOptions,
	}
	cluster := c.GetClusterController().GetClusterManager().GetCluster(c.GetClusterMetaData().GetClusterUid())
	log.Infof("Scheduling app [%s] on namespace [%s]", c.GetAppMetaData().GetApp(), c.GetNamespaceMetaData().GetNamespace())
	resp, err := cluster.ProcessClusterRequest(appScheduleRequest)
	if err != nil {
		return utils.ProcessError(err, utils.StructToString(appScheduleRequest))
	}
	namespaceUid := c.NamespaceMetaData.GetNamespaceUid()
	if !cluster.GetNamespaceManager().IsNamespacePresent(namespaceUid) {
		cluster.NamespaceManager.SetNamespace(namespaceUid, NewNamespace())
	}
	appScheduleResponse := resp.(*driverapi.AppScheduleResponse)
	app := NewApp()
	app.SetContexts(appScheduleResponse.Contexts)
	cluster.GetNamespaceManager().GetNamespace(namespaceUid).GetAppManager().SetApp(c.AppMetaData.GetAppUid(), app)
	return nil
}

func (c *AppConfig) canValidate() error {
	return nil
}

func (c *AppConfig) Validate() error {
	err := c.canValidate()
	if err != nil {
		return utils.ProcessError(err)
	}
	cluster := c.GetClusterController().GetClusterManager().GetCluster(c.GetClusterMetaData().GetClusterUid())
	err = cluster.GetContextManager().SwitchContext()
	if err != nil {
		return utils.ProcessError(err)
	}
	contexts := cluster.GetNamespaceManager().GetNamespace(c.NamespaceMetaData.GetNamespaceUid()).GetAppManager().GetApp(c.AppMetaData.GetAppUid()).GetContexts()
	for _, ctx := range contexts {
		log.Infof("Validating app [%s]", c.GetAppMetaData().GetApp())
		log.Infof("Waiting for app [%s] to start running", c.GetAppMetaData().GetApp())

		timeout := c.GetValidateAppConfig().GetWaitForRunningTimeout()
		retryInterval := c.GetValidateAppConfig().GetWaitForRunningTimeout()
		err = tests.Inst().S.WaitForRunning(ctx, timeout, retryInterval)
		if err != nil {
			debugStruct := struct {
				Cc            *scheduler.Context
				Timeout       time.Duration
				RetryInterval time.Duration
			}{
				Cc:            ctx,
				Timeout:       timeout,
				RetryInterval: retryInterval,
			}
			return utils.ProcessError(err, utils.StructToString(debugStruct))
		}
		if !ctx.SkipVolumeValidation {
			log.Infof("Validating application [%s] volumes", ctx.App.Key)
			timeout = c.GetValidateAppConfig().GetValidateVolumeTimeout()
			retryInterval = c.GetValidateAppConfig().GetValidateVolumeRetryInterval()
			err = tests.Inst().S.ValidateVolumes(ctx, timeout, retryInterval, nil)
			if err != nil {
				debugStruct := struct {
					Cc            *scheduler.Context
					Timeout       time.Duration
					RetryInterval time.Duration
				}{
					Cc:            ctx,
					Timeout:       timeout,
					RetryInterval: retryInterval,
				}
				return utils.ProcessError(err, utils.StructToString(debugStruct))
			}
			volumeParameters, err := tests.Inst().S.GetVolumeParameters(ctx)
			if err != nil {
				debugStruct := struct {
					Ctx *scheduler.Context
				}{
					Ctx: ctx,
				}
				return utils.ProcessError(err, utils.StructToString(debugStruct))
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
				debugStruct := struct {
					Ctx *scheduler.Context
				}{
					Ctx: ctx,
				}
				return utils.ProcessError(err, utils.StructToString(debugStruct))
			}
			for _, vol := range volumes {
				err = tests.Inst().V.ValidateVolumeSetup(vol)
				if err != nil {
					debugStruct := struct {
						Vol *volume.Volume
					}{
						Vol: vol,
					}
					return utils.ProcessError(err, utils.StructToString(debugStruct))
				}
			}
		}
	}
	return nil
}

func (c *AppConfig) canTearDown() error {
	return nil
}

func (c *AppConfig) TearDown() error {
	err := c.canTearDown()
	if err != nil {
		return utils.ProcessError(err)
	}
	err = c.GetClusterController().GetClusterManager().GetContextManager().SwitchContext()
	if err != nil {
		return utils.ProcessError(err)
	}
	log.Infof("Destroying application [%s]", c.AppMetaData.GetUid())
	Contexts := c.ClusterController.NamespaceManager.GetNamespace(c.NamespaceMetaData).AppManager.GetApp(c.AppMetaData).Contexts
	log.Infof("len of tear down Contexts %d", len(Contexts))
	for _, ctx := range Contexts {
		volumeOptions := &scheduler.VolumeOptions{
			SkipClusterScopedObjects: true,
		}
		volumes, err := tests.Inst().S.DeleteVolumes(ctx, volumeOptions)
		if err != nil {
			debugMessage := fmt.Sprintf("scheduler-context: [%v]; volume-options: [%v]", ctx, volumeOptions)
			return utils.ProcessError(err, debugMessage)
		}
		destroyOptions := map[string]bool{
			tests.SkipClusterScopedObjects:              c.SkipClusterScopedObjects,
			scheduler.OptionsWaitForDestroy:             c.WaitForDestroy,
			scheduler.OptionsWaitForResourceLeakCleanup: c.WaitForResourceLeakCleanup,
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
		if !c.TearDownAppConfig.SkipClusterScopedObjects {
			volumeOptions = &scheduler.VolumeOptions{
				SkipClusterScopedObjects: false,
			}
			_, err = tests.Inst().S.DeleteVolumes(ctx, volumeOptions)
			if err != nil {
				debugMessage := fmt.Sprintf("scheduler-context: [%v]; volume-options: [%v]", ctx, volumeOptions)
				return utils.ProcessError(err, debugMessage)
			}
		}
		c.ClusterController.NamespaceManager.GetNamespace(c.NamespaceMetaData).AppManager.RemoveApp(c.AppMetaData)
		log.Infof("present Apps %d", len(c.ClusterController.NamespaceManager.GetNamespace(c.NamespaceMetaData).AppManager.Apps))
		log.Infof("removed Apps %d", len(c.ClusterController.NamespaceManager.GetNamespace(c.NamespaceMetaData).AppManager.RemovedApps))
	}
	return nil
}

// App represents an App
type App struct {
	Contexts []*scheduler.Context
}

// GetContexts returns the Contexts associated with the App
func (a *App) GetContexts() []*scheduler.Context {
	return a.Contexts
}

// SetContexts sets the Contexts for the App
func (a *App) SetContexts(contexts []*scheduler.Context) {
	a.Contexts = contexts
}

// NewApp creates a new instance of the App
func NewApp() *App {
	newApp := &App{}
	newApp.SetContexts(nil)
	return newApp
}

// AppManager represents a manager for App
type AppManager struct {
	sync.RWMutex
	AppMap         map[string]*App
	RemovedAppsMap map[string][]*App
}

// GetAppMap returns the AppMap of the AppManager
func (m *AppManager) GetAppMap() map[string]*App {
	m.RLock()
	defer m.RUnlock()
	return m.AppMap
}

// SetAppMap sets the AppMap of the AppManager
func (m *AppManager) SetAppMap(appMap map[string]*App) {
	m.Lock()
	defer m.Unlock()
	m.AppMap = appMap
}

// GetRemovedAppsMap returns the RemovedAppsMap of the AppManager
func (m *AppManager) GetRemovedAppsMap() map[string][]*App {
	m.RLock()
	defer m.RUnlock()
	return m.RemovedAppsMap
}

// SetRemovedAppsMap sets the RemovedAppsMap of the AppManager
func (m *AppManager) SetRemovedAppsMap(removedAppsMap map[string][]*App) {
	m.Lock()
	defer m.Unlock()
	m.RemovedAppsMap = removedAppsMap
}

// GetApp returns the App with the given App uid
func (m *AppManager) GetApp(appUid string) *App {
	m.RLock()
	defer m.RUnlock()
	return m.GetAppMap()[appUid]
}

// IsAppPresent checks if the App with the given App uid is present
func (m *AppManager) IsAppPresent(appUid string) bool {
	m.RLock()
	defer m.RUnlock()
	_, isPresent := m.GetAppMap()[appUid]
	return isPresent
}

// SetApp sets the App with the given App uid
func (m *AppManager) SetApp(appUid string, app *App) {
	m.Lock()
	defer m.Unlock()
	m.GetAppMap()[appUid] = app
}

// DeleteApp deletes the App with the given App uid
func (m *AppManager) DeleteApp(appUid string) {
	m.Lock()
	defer m.Unlock()
	delete(m.GetAppMap(), appUid)
}

// RemoveApp removes the App with the given App uid
func (m *AppManager) RemoveApp(appUid string) {
	m.Lock()
	defer m.Unlock()
	m.GetRemovedAppsMap()[appUid] = append(m.GetRemovedAppsMap()[appUid], m.GetApp(appUid))
	m.DeleteApp(appUid)
}

// NewAppManager creates a new instance of the AppManager
func NewAppManager() *AppManager {
	newAppManager := &AppManager{}
	newAppManager.SetAppMap(make(map[string]*App, 0))
	newAppManager.SetRemovedAppsMap(make(map[string][]*App, 0))
	return newAppManager
}
