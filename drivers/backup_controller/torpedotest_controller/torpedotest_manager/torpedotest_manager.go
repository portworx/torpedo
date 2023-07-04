package torpedotest_manager

import (
	"github.com/portworx/torpedo/drivers/backup_controller/backup_utils"
	. "github.com/portworx/torpedo/drivers/backup_controller/torpedotest_controller/torpedotest_metadata"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	DefaultTestName          = ""
	DefaultTestDescription   = ""
	DefaultTestMaintainer    = ""
	DefaultTestRailID        = 0
	DefaultTestRunIDForSuite = 0
)

var (
	DefaultTestTags                      = make(map[string]string, 0)
	DefaultTestLogger *lumberjack.Logger = nil
)

// TorpedoTest represents an TorpedoTest
type TorpedoTest struct {
	TestName          string
	TestDescription   string
	TestMaintainer    backup_utils.TestMaintainer
	TestRailID        int
	TestRunIDForSuite int
	TestTags          map[string]string
	TestLogger        *lumberjack.Logger
}

// GetTestName returns the TestName associated with the TorpedoTest
func (t *TorpedoTest) GetTestName() string {
	return t.TestName
}

// SetTestName sets the TestName for the TorpedoTest
func (t *TorpedoTest) SetTestName(testName string) *TorpedoTest {
	t.TestName = testName
	return t
}

// GetTestDescription returns the TestDescription associated with the TorpedoTest
func (t *TorpedoTest) GetTestDescription() string {
	return t.TestDescription
}

// SetTestDescription sets the TestDescription for the TorpedoTest
func (t *TorpedoTest) SetTestDescription(description string) *TorpedoTest {
	t.TestDescription = description
	return t
}

// GetTestMaintainer returns the TestMaintainer associated with the TorpedoTest
func (t *TorpedoTest) GetTestMaintainer() backup_utils.TestMaintainer {
	return t.TestMaintainer
}

// SetTestMaintainer sets the TestMaintainer for the TorpedoTest
func (t *TorpedoTest) SetTestMaintainer(maintainer backup_utils.TestMaintainer) *TorpedoTest {
	t.TestMaintainer = maintainer
	return t
}

// GetTestRailID returns the TestRailID associated with the TorpedoTest
func (t *TorpedoTest) GetTestRailID() int {
	return t.TestRailID
}

// SetTestRailID sets the TestRailID for the TorpedoTest
func (t *TorpedoTest) SetTestRailID(id int) *TorpedoTest {
	t.TestRailID = id
	return t
}

// GetTestRunIDForSuite returns the TestRunIDForSuite associated with the TorpedoTest
func (t *TorpedoTest) GetTestRunIDForSuite() int {
	return t.TestRunIDForSuite
}

// SetTestRunIDForSuite sets the TestRunIDForSuite for the TorpedoTest
func (t *TorpedoTest) SetTestRunIDForSuite(id int) *TorpedoTest {
	t.TestRunIDForSuite = id
	return t
}

// GetTestTags returns the TestTags associated with the TorpedoTest
func (t *TorpedoTest) GetTestTags() map[string]string {
	return t.TestTags
}

// SetTestTags sets the TestTags for the TorpedoTest
func (t *TorpedoTest) SetTestTags(tags map[string]string) *TorpedoTest {
	t.TestTags = tags
	return t
}

// GetTestLogger returns the TestLogger associated with the TorpedoTest
func (t *TorpedoTest) GetTestLogger() *lumberjack.Logger {
	return t.TestLogger
}

// SetTestLogger sets the TestLogger for the TorpedoTest
func (t *TorpedoTest) SetTestLogger(logger *lumberjack.Logger) *TorpedoTest {
	t.TestLogger = logger
	return t
}

// NewTorpedoTest creates a new instance of the TorpedoTest
func NewTorpedoTest(testName string, testDescription string, testMaintainer backup_utils.TestMaintainer, testRailID int, testRunIDForSuite int, testTags map[string]string, testLogger *lumberjack.Logger) *TorpedoTest {
	newTorpedoTest := &TorpedoTest{}
	newTorpedoTest.SetTestName(testName)
	newTorpedoTest.SetTestDescription(testDescription)
	newTorpedoTest.SetTestMaintainer(testMaintainer)
	newTorpedoTest.SetTestRailID(testRailID)
	newTorpedoTest.SetTestRunIDForSuite(testRunIDForSuite)
	newTorpedoTest.SetTestTags(testTags)
	newTorpedoTest.SetTestLogger(testLogger)
	return newTorpedoTest
}

// NewDefaultTorpedoTest creates a new instance of the TorpedoTest with default values
func NewDefaultTorpedoTest() *TorpedoTest {
	return NewTorpedoTest(DefaultTestName, DefaultTestDescription, DefaultTestMaintainer, DefaultTestRailID, DefaultTestRunIDForSuite, DefaultTestTags, DefaultTestLogger)
}

var (
	DefaultTorpedoTestMap         = make(map[string]*TorpedoTest, 0)
	DefaultRemovedTorpedoTestsMap = make(map[string][]*TorpedoTest, 0)
)

// TorpedoTestManager represents a manager for TorpedoTest
type TorpedoTestManager struct {
	TorpedoTestMap         map[string]*TorpedoTest
	RemovedTorpedoTestsMap map[string][]*TorpedoTest
}

// GetTorpedoTestMap returns the TorpedoTestMap of the TorpedoTestManager
func (m *TorpedoTestManager) GetTorpedoTestMap() map[string]*TorpedoTest {
	return m.TorpedoTestMap
}

// SetTorpedoTestMap sets the TorpedoTestMap of the TorpedoTestManager
func (m *TorpedoTestManager) SetTorpedoTestMap(testMap map[string]*TorpedoTest) *TorpedoTestManager {
	m.TorpedoTestMap = testMap
	return m
}

// GetRemovedTorpedoTestsMap returns the RemovedTorpedoTestsMap of the TorpedoTestManager
func (m *TorpedoTestManager) GetRemovedTorpedoTestsMap() map[string][]*TorpedoTest {
	return m.RemovedTorpedoTestsMap
}

// SetRemovedTorpedoTestsMap sets the RemovedTorpedoTestsMap of the TorpedoTestManager
func (m *TorpedoTestManager) SetRemovedTorpedoTestsMap(removedTestsMap map[string][]*TorpedoTest) *TorpedoTestManager {
	m.RemovedTorpedoTestsMap = removedTestsMap
	return m
}

// GetTorpedoTest returns the TorpedoTest with the given TorpedoTest UID
func (m *TorpedoTestManager) GetTorpedoTest(torpedoTestUID string) *TorpedoTest {
	return m.TorpedoTestMap[torpedoTestUID]
}

// SetTorpedoTest sets the TorpedoTest with the given TorpedoTest UID
func (m *TorpedoTestManager) SetTorpedoTest(torpedoTestUID string, torpedoTest *TorpedoTest) *TorpedoTestManager {
	m.TorpedoTestMap[torpedoTestUID] = torpedoTest
	return m
}

// DeleteTorpedoTest deletes the TorpedoTest with the given TorpedoTest UID
func (m *TorpedoTestManager) DeleteTorpedoTest(torpedoTestUID string) *TorpedoTestManager {
	delete(m.TorpedoTestMap, torpedoTestUID)
	return m
}

// RemoveTorpedoTest removes the TorpedoTest with the given TorpedoTest UID
func (m *TorpedoTestManager) RemoveTorpedoTest(torpedoTestUID string) *TorpedoTestManager {
	if torpedoTest, isPresent := m.TorpedoTestMap[torpedoTestUID]; isPresent {
		m.RemovedTorpedoTestsMap[torpedoTestUID] = append(m.RemovedTorpedoTestsMap[torpedoTestUID], torpedoTest)
		delete(m.TorpedoTestMap, torpedoTestUID)
	}
	return m
}

// IsTorpedoTestPresent checks if the TorpedoTest with the given TorpedoTest UID is present
func (m *TorpedoTestManager) IsTorpedoTestPresent(torpedoTestUID string) bool {
	_, isPresent := m.TorpedoTestMap[torpedoTestUID]
	return isPresent
}

// IsTorpedoTestRemoved checks if the TorpedoTest with the given TorpedoTest UID is removed
func (m *TorpedoTestManager) IsTorpedoTestRemoved(torpedoTestUID string) bool {
	_, isPresent := m.RemovedTorpedoTestsMap[torpedoTestUID]
	return isPresent
}

// IsTorpedoTestRecorded checks if the TorpedoTest with the given TorpedoTest UID is recorded
func (m *TorpedoTestManager) IsTorpedoTestRecorded(torpedoTestUID string) bool {
	return m.IsTorpedoTestPresent(torpedoTestUID) || m.IsTorpedoTestRemoved(torpedoTestUID)
}

// NewTorpedoTestManager creates a new instance of the TorpedoTestManager
func NewTorpedoTestManager(torpedoTestMap map[string]*TorpedoTest, removedTorpedoTestsMap map[string][]*TorpedoTest) *TorpedoTestManager {
	newTorpedoTestManager := &TorpedoTestManager{}
	newTorpedoTestManager.SetTorpedoTestMap(torpedoTestMap)
	newTorpedoTestManager.SetRemovedTorpedoTestsMap(removedTorpedoTestsMap)
	return newTorpedoTestManager
}

// NewDefaultTorpedoTestManager creates a new instance of the TorpedoTestManager with default values
func NewDefaultTorpedoTestManager() *TorpedoTestManager {
	return NewTorpedoTestManager(DefaultTorpedoTestMap, DefaultRemovedTorpedoTestsMap)
}

// TorpedoTestConfig represents the configuration for a TorpedoTest
type TorpedoTestConfig struct {
	TorpedoTestManager  *TorpedoTestManager
	TorpedoTestMetaData *TorpedoTestMetaData
}

// GetTorpedoTestManager returns the TorpedoTestManager associated with the TorpedoTestConfig
func (c *TorpedoTestConfig) GetTorpedoTestManager() *TorpedoTestManager {
	return c.TorpedoTestManager
}

// SetTorpedoTestManager sets the TorpedoTestManager for the TorpedoTestConfig
func (c *TorpedoTestConfig) SetTorpedoTestManager(manager *TorpedoTestManager) *TorpedoTestConfig {
	c.TorpedoTestManager = manager
	return c
}

// GetTorpedoTestMetaData returns the TorpedoTestMetaData associated with the TorpedoTestConfig
func (c *TorpedoTestConfig) GetTorpedoTestMetaData() *TorpedoTestMetaData {
	return c.TorpedoTestMetaData
}

// SetTorpedoTestMetaData sets the TorpedoTestMetaData for the TorpedoTestConfig
func (c *TorpedoTestConfig) SetTorpedoTestMetaData(metaData *TorpedoTestMetaData) *TorpedoTestConfig {
	c.TorpedoTestMetaData = metaData
	return c
}

// NewTorpedoTestConfig creates a new instance of the TorpedoTestConfig
func NewTorpedoTestConfig(manager *TorpedoTestManager, metaData *TorpedoTestMetaData) *TorpedoTestConfig {
	newTorpedoTestConfig := &TorpedoTestConfig{}
	newTorpedoTestConfig.SetTorpedoTestManager(manager)
	newTorpedoTestConfig.SetTorpedoTestMetaData(metaData)
	return newTorpedoTestConfig
}

// NewDefaultTorpedoTestConfig creates a new instance of the TorpedoTestConfig with default values
func NewDefaultTorpedoTestConfig() *TorpedoTestConfig {
	return NewTorpedoTestConfig(NewDefaultTorpedoTestManager(), NewDefaultTorpedoTestMetaData())
}
