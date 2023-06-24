package app_manager

import (
	"github.com/portworx/torpedo/drivers/backup/controller/cluster_controller"
	"github.com/portworx/torpedo/drivers/backup/controller/cluster_controller/cluster_manager"
	"github.com/portworx/torpedo/drivers/backup/controller/cluster_controller/cluster_manager/namespace_manager"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/tests"
	"sync"
	"time"
)

// AppMetaData represents the metadata for an App
type AppMetaData struct {
	AppKey string
}

// GetAppKey returns the AppKey associated with the AppMetaData
func (m *AppMetaData) GetAppKey() string {
	return m.AppKey
}

// SetAppKey sets the AppKey for the AppMetaData
func (m *AppMetaData) SetAppKey(appKey string) {
	m.AppKey = appKey
}

// GetApp returns the App associated with the AppMetaData
func (m *AppMetaData) GetApp() string {
	return m.GetAppKey()
}

// GetAppUid returns the App uid
func (m *AppMetaData) GetAppUid() string {
	return m.GetApp()
}

// NewAppMetaData creates a new instance of the AppMetaData
func NewAppMetaData() *AppMetaData {
	newAppConfig := &AppMetaData{}
	newAppConfig.SetAppKey("")
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
	newValidateAppConfig.SetWaitForRunningTimeout(namespace_manager.DefaultWaitForRunningTimeout)
	newValidateAppConfig.SetWaitForRunningRetryInterval(namespace_manager.DefaultWaitForRunningRetryInterval)
	newValidateAppConfig.SetValidateVolumeTimeout(namespace_manager.DefaultValidateVolumeTimeout)
	newValidateAppConfig.SetValidateVolumeRetryInterval(namespace_manager.DefaultValidateVolumeRetryInterval)
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
	newTearDownAppConfig.SetWaitForDestroy(namespace_manager.DefaultWaitForDestroy)
	newTearDownAppConfig.SetWaitForResourceLeakCleanup(namespace_manager.DefaultWaitForResourceLeakCleanup)
	newTearDownAppConfig.SetSkipClusterScopedObjects(namespace_manager.DefaultSkipClusterScopedObjects)
	return newTearDownAppConfig
}

// AppConfig represents the configuration for an App
type AppConfig struct {
	ClusterMetaData   *cluster_manager.ClusterMetaData
	NamespaceMetaData *namespace_manager.NamespaceMetaData
	AppMetaData       *AppMetaData
	ScheduleAppConfig *ScheduleAppConfig
	ValidateAppConfig *ValidateAppConfig
	TearDownAppConfig *TearDownAppConfig
	ClusterController *cluster_controller.ClusterController
}

// GetClusterMetaData returns the ClusterMetaData associated with the AppConfig
func (c *AppConfig) GetClusterMetaData() *cluster_manager.ClusterMetaData {
	return c.ClusterMetaData
}

// SetClusterMetaData sets the ClusterMetaData for the AppConfig
func (c *AppConfig) SetClusterMetaData(metaData *cluster_manager.ClusterMetaData) {
	c.ClusterMetaData = metaData
}

// GetNamespaceMetaData returns the NamespaceMetaData associated with the AppConfig
func (c *AppConfig) GetNamespaceMetaData() *namespace_manager.NamespaceMetaData {
	return c.NamespaceMetaData
}

// SetNamespaceMetaData sets the NamespaceMetaData for the AppConfig
func (c *AppConfig) SetNamespaceMetaData(metaData *namespace_manager.NamespaceMetaData) {
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

// GetScheduleAppConfig returns the ScheduleAppConfig associated with the AppConfig
func (c *AppConfig) GetScheduleAppConfig() *ScheduleAppConfig {
	return c.ScheduleAppConfig
}

// SetScheduleAppConfig sets the ScheduleAppConfig for the AppConfig
func (c *AppConfig) SetScheduleAppConfig(config *ScheduleAppConfig) {
	c.ScheduleAppConfig = config
}

// GetValidateAppConfig returns the ValidateAppConfig associated with the AppConfig
func (c *AppConfig) GetValidateAppConfig() *ValidateAppConfig {
	return c.ValidateAppConfig
}

// SetValidateAppConfig sets the ValidateAppConfig for the AppConfig
func (c *AppConfig) SetValidateAppConfig(config *ValidateAppConfig) {
	c.ValidateAppConfig = config
}

// GetTearDownAppConfig returns the TearDownAppConfig associated with the AppConfig
func (c *AppConfig) GetTearDownAppConfig() *TearDownAppConfig {
	return c.TearDownAppConfig
}

// SetTearDownAppConfig sets the TearDownAppConfig for the AppConfig
func (c *AppConfig) SetTearDownAppConfig(config *TearDownAppConfig) {
	c.TearDownAppConfig = config
}

// GetClusterController returns the ClusterController associated with the AppConfig
func (c *AppConfig) GetClusterController() *cluster_controller.ClusterController {
	return c.ClusterController
}

// SetClusterController sets the ClusterController for the AppConfig
func (c *AppConfig) SetClusterController(controller *cluster_controller.ClusterController) {
	c.ClusterController = controller
}

// NewAppConfig creates a new instance of the AppConfig
func NewAppConfig() *AppConfig {
	newAppConfig := &AppConfig{}
	newAppConfig.SetClusterMetaData(cluster_manager.NewClusterMetaData())
	newAppConfig.SetNamespaceMetaData(namespace_manager.NewNamespaceMetaData())
	newAppConfig.SetAppMetaData(NewAppMetaData())
	newAppConfig.SetScheduleAppConfig(NewScheduleAppConfig())
	newAppConfig.SetValidateAppConfig(NewValidateAppConfig())
	newAppConfig.SetTearDownAppConfig(NewTearDownAppConfig())
	newAppConfig.SetClusterController(nil)
	return newAppConfig
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
	return m.AppMap[appUid]
}

// IsAppPresent checks if the App with the given App uid is present
func (m *AppManager) IsAppPresent(appUid string) bool {
	m.RLock()
	defer m.RUnlock()
	_, isPresent := m.AppMap[appUid]
	return isPresent
}

// SetApp sets the App with the given App uid
func (m *AppManager) SetApp(appUid string, app *App) {
	m.Lock()
	defer m.Unlock()
	m.AppMap[appUid] = app
}

// DeleteApp deletes the App with the given App uid
func (m *AppManager) DeleteApp(appUid string) {
	m.Lock()
	defer m.Unlock()
	delete(m.AppMap, appUid)
}

// RemoveApp removes the App with the given App uid
func (m *AppManager) RemoveApp(appUid string) {
	m.Lock()
	defer m.Unlock()
	m.RemovedAppsMap[appUid] = append(m.RemovedAppsMap[appUid], m.AppMap[appUid])
	m.DeleteApp(appUid)
}

// NewAppManager creates a new instance of the AppManager
func NewAppManager() *AppManager {
	newAppManager := &AppManager{}
	newAppManager.SetAppMap(make(map[string]*App, 0))
	newAppManager.SetRemovedAppsMap(make(map[string][]*App, 0))
	return newAppManager
}
