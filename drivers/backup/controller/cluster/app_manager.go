package cluster

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/scheduler"
	"time"
)

const (
	GlobalAuthTokenParam = "auth-token" // copy of the const `authTokenParam` declared in the common.go file of the tests package
)

const (
	// DefaultWaitForRunningTimeout indicates the duration to wait for an app to reach the running state
	DefaultWaitForRunningTimeout = 10 * time.Minute
	// DefaultWaitForRunningRetryInterval indicates the interval between retries when waiting for an app to reach the running state
	DefaultWaitForRunningRetryInterval = 10 * time.Second
	// DefaultValidateVolumeTimeout indicates the duration to wait for volume validation of an app
	DefaultValidateVolumeTimeout = 10 * time.Minute
	// DefaultValidateVolumeRetryInterval indicates the interval between retries when performing volume validation of an app
	DefaultValidateVolumeRetryInterval = 10 * time.Second
)

const (
	// DefaultWaitForDestroy indicates whether to wait for resources to be destroyed during the teardown process
	DefaultWaitForDestroy = true // indicates the value of the `scheduler.OptionsWaitForDestroy` key
	// DefaultWaitForResourceLeakCleanup indicates whether to wait for resource leak cleanup during the teardown process
	DefaultWaitForResourceLeakCleanup = true // indicates the value of the `scheduler.OptionsWaitForResourceLeakCleanup` key
	// DefaultSkipClusterScopedObjects indicates whether to skip cluster-scoped objects during the teardown process
	DefaultSkipClusterScopedObjects = false // indicates the value of the `SkipClusterScopedObject` field in the `scheduler.Context`
)

type AppMetaData struct {
	AppKey     string
	Identifier []string
}

func (m *AppMetaData) HasIdentifier() bool {
	return m.Identifier == nil
}

func (m *AppMetaData) GetSuffix() string {
	if !m.HasIdentifier() {
		return ""
	}
	return fmt.Sprintf("-%s", m.Identifier[0])
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

type AppManager struct {
	Apps        map[string]*App
	RemovedApps map[string][]*App
}

func NewApp(contexts []*scheduler.Context) *App {
	return &App{
		Contexts: contexts,
	}
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
