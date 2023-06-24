package pxbackuptest

import (
	"github.com/portworx/torpedo/drivers/backup/utils"
	"gopkg.in/natefinch/lumberjack.v2"
	"sync"
)

// PxBackupTestMetaData represents the metadata for a PxBackupTest
type PxBackupTestMetaData struct {
	TestId string
}

// GetTestId returns the TestId associated with the PxBackupTestMetaData
func (m *PxBackupTestMetaData) GetTestId() string {
	return m.TestId
}

// SetTestId sets the TestId for the PxBackupTestMetaData
func (m *PxBackupTestMetaData) SetTestId(id string) {
	m.TestId = id
}

// GetTestUid returns the PxBackupTest uid
func (m *PxBackupTestMetaData) GetTestUid() string {
	return m.GetTestId()
}

// NewPxBackupTestMetaData creates a new instance of the PxBackupTestMetaData
func NewPxBackupTestMetaData() *PxBackupTestMetaData {
	newPxBackupTestMetaData := &PxBackupTestMetaData{}
	newPxBackupTestMetaData.SetTestId("")
	return newPxBackupTestMetaData
}

// PxBackupTestConfig represents the configuration for a PxBackupTest
type PxBackupTestConfig struct {
	PxBackupTestMetaData   *PxBackupTestMetaData
	PxBackupTestController *PxBackupTestController
}

// GetPxBackupTestMetaData returns the PxBackupTestMetaData associated with the PxBackupTestConfig
func (c *PxBackupTestConfig) GetPxBackupTestMetaData() *PxBackupTestMetaData {
	return c.PxBackupTestMetaData
}

// SetPxBackupTestMetaData sets the PxBackupTestMetaData for the PxBackupTestConfig
func (c *PxBackupTestConfig) SetPxBackupTestMetaData(metaData *PxBackupTestMetaData) {
	c.PxBackupTestMetaData = metaData
}

// GetPxBackupTestController returns the PxBackupTestController associated with the PxBackupTestConfig
func (c *PxBackupTestConfig) GetPxBackupTestController() *PxBackupTestController {
	return c.PxBackupTestController
}

// SetPxBackupTestController sets the PxBackupTestController for the PxBackupTestConfig
func (c *PxBackupTestConfig) SetPxBackupTestController(controller *PxBackupTestController) {
	c.PxBackupTestController = controller
}

// NewPxBackupTestConfig creates a new instance of the PxBackupTestConfig
func NewPxBackupTestConfig() *PxBackupTestConfig {
	newPxBackupTestConfig := &PxBackupTestConfig{}
	PxBackupTestMetaData := NewPxBackupTestMetaData()
	newPxBackupTestConfig.SetPxBackupTestMetaData(PxBackupTestMetaData)
	newPxBackupTestConfig.SetPxBackupTestController(nil)
	return newPxBackupTestConfig
}

// PxBackupTest represents an PxBackupTest
type PxBackupTest struct {
	TestName          string
	TestDescription   string
	TestMaintainer    utils.TestMaintainer
	TestRailID        int
	TestRunIDForSuite int
	TestTags          map[string]string
	TestLogger        *lumberjack.Logger
}

// GetTestName returns the TestName associated with the PxBackupTest
func (t *PxBackupTest) GetTestName() string {
	return t.TestName
}

// SetTestName sets the TestName for the PxBackupTest
func (t *PxBackupTest) SetTestName(name string) {
	t.TestName = name
}

// GetTestDescription returns the TestDescription associated with the PxBackupTest
func (t *PxBackupTest) GetTestDescription() string {
	return t.TestDescription
}

// SetTestDescription sets the TestDescription for the PxBackupTest
func (t *PxBackupTest) SetTestDescription(description string) {
	t.TestDescription = description
}

// GetTestMaintainer returns the TestMaintainer associated with the PxBackupTest
func (t *PxBackupTest) GetTestMaintainer() utils.TestMaintainer {
	return t.TestMaintainer
}

// SetTestMaintainer sets the TestMaintainer for the PxBackupTest
func (t *PxBackupTest) SetTestMaintainer(maintainer utils.TestMaintainer) {
	t.TestMaintainer = maintainer
}

// GetTestRailID returns the TestRailID associated with the PxBackupTest
func (t *PxBackupTest) GetTestRailID() int {
	return t.TestRailID
}

// SetTestRailID sets the TestRailID for the PxBackupTest
func (t *PxBackupTest) SetTestRailID(id int) {
	t.TestRailID = id
}

// GetTestRunIDForSuite returns the TestRunIDForSuite associated with the PxBackupTest
func (t *PxBackupTest) GetTestRunIDForSuite() int {
	return t.TestRunIDForSuite
}

// SetTestRunIDForSuite sets the TestRunIDForSuite for the PxBackupTest
func (t *PxBackupTest) SetTestRunIDForSuite(id int) {
	t.TestRunIDForSuite = id
}

// GetTestTags returns the TestTags associated with the PxBackupTest
func (t *PxBackupTest) GetTestTags() map[string]string {
	return t.TestTags
}

// SetTestTags sets the TestTags for the PxBackupTest
func (t *PxBackupTest) SetTestTags(tags map[string]string) {
	t.TestTags = tags
}

// GetTestLogger returns the TestLogger associated with the PxBackupTest
func (t *PxBackupTest) GetTestLogger() *lumberjack.Logger {
	return t.TestLogger
}

// SetTestLogger sets the TestLogger for the PxBackupTest
func (t *PxBackupTest) SetTestLogger(logger *lumberjack.Logger) {
	t.TestLogger = logger
}

// NewPxBackupTest creates a new instance of the PxBackupTest
func NewPxBackupTest() *PxBackupTest {
	newPxBackupTest := &PxBackupTest{}
	newPxBackupTest.SetTestName("")
	newPxBackupTest.SetTestDescription("")
	newPxBackupTest.SetTestMaintainer("")
	newPxBackupTest.SetTestRailID(0)
	newPxBackupTest.SetTestRunIDForSuite(0)
	newPxBackupTest.SetTestTags(make(map[string]string, 0))
	newPxBackupTest.SetTestLogger(nil)
	return newPxBackupTest
}

// PxBackupTestManager represents a manager for PxBackupTest
type PxBackupTestManager struct {
	sync.RWMutex
	PxBackupTestMap         map[string]*PxBackupTest
	RemovedPxBackupTestsMap map[string][]*PxBackupTest
}

// GetPxBackupTestMap returns the PxBackupTestMap of the PxBackupTestManager
func (m *PxBackupTestManager) GetPxBackupTestMap() map[string]*PxBackupTest {
	m.RLock()
	defer m.RUnlock()
	return m.PxBackupTestMap
}

// SetPxBackupTestMap sets the PxBackupTestMap of the PxBackupTestManager
func (m *PxBackupTestManager) SetPxBackupTestMap(testMap map[string]*PxBackupTest) {
	m.Lock()
	defer m.Unlock()
	m.PxBackupTestMap = testMap
}

// GetRemovedPxBackupTestsMap returns the RemovedPxBackupTestsMap of the PxBackupTestManager
func (m *PxBackupTestManager) GetRemovedPxBackupTestsMap() map[string][]*PxBackupTest {
	m.RLock()
	defer m.RUnlock()
	return m.RemovedPxBackupTestsMap
}

// SetRemovedPxBackupTestsMap sets the RemovedPxBackupTestsMap of the PxBackupTestManager
func (m *PxBackupTestManager) SetRemovedPxBackupTestsMap(removedTestsMap map[string][]*PxBackupTest) {
	m.Lock()
	defer m.Unlock()
	m.RemovedPxBackupTestsMap = removedTestsMap
}

// GetPxBackupTest returns the PxBackupTest with the given PxBackupTest uid
func (m *PxBackupTestManager) GetPxBackupTest(PxBackupTestUid string) *PxBackupTest {
	m.RLock()
	defer m.RUnlock()
	return m.PxBackupTestMap[PxBackupTestUid]
}

// IsPxBackupTestPresent checks if the PxBackupTest with the given PxBackupTest uid is present
func (m *PxBackupTestManager) IsPxBackupTestPresent(PxBackupTestUid string) bool {
	m.RLock()
	defer m.RUnlock()
	_, isPresent := m.PxBackupTestMap[PxBackupTestUid]
	return isPresent
}

// SetPxBackupTest sets the PxBackupTest with the given PxBackupTest uid
func (m *PxBackupTestManager) SetPxBackupTest(PxBackupTestUid string, PxBackupTest *PxBackupTest) {
	m.Lock()
	defer m.Unlock()
	m.PxBackupTestMap[PxBackupTestUid] = PxBackupTest
}

// DeletePxBackupTest deletes the PxBackupTest with the given PxBackupTest uid
func (m *PxBackupTestManager) DeletePxBackupTest(PxBackupTestUid string) {
	m.Lock()
	defer m.Unlock()
	delete(m.PxBackupTestMap, PxBackupTestUid)
}

// RemovePxBackupTest removes the PxBackupTest with the given PxBackupTest uid
func (m *PxBackupTestManager) RemovePxBackupTest(PxBackupTestUid string) {
	m.Lock()
	defer m.Unlock()
	m.RemovedPxBackupTestsMap[PxBackupTestUid] = append(m.RemovedPxBackupTestsMap[PxBackupTestUid], m.PxBackupTestMap[PxBackupTestUid])
	delete(m.PxBackupTestMap, PxBackupTestUid)
}

// NewPxBackupTestManager creates a new instance of the PxBackupTestManager
func NewPxBackupTestManager() *PxBackupTestManager {
	newPxBackupTestManager := &PxBackupTestManager{}
	newPxBackupTestManager.SetPxBackupTestMap(make(map[string]*PxBackupTest, 0))
	newPxBackupTestManager.SetRemovedPxBackupTestsMap(make(map[string][]*PxBackupTest, 0))
	return newPxBackupTestManager
}
