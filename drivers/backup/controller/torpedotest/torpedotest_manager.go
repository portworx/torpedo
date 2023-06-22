package torpedotest

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/testrailuttils"
	"github.com/portworx/torpedo/tests"
	"gopkg.in/natefinch/lumberjack.v2"
	"strconv"
	"strings"
	"sync"
)

// TorpedoTestMetaData represents the metadata for a TorpedoTest
type TorpedoTestMetaData struct {
	TestId string
}

// GetTestId returns the TestId associated with the TorpedoTestMetaData
func (m *TorpedoTestMetaData) GetTestId() string {
	return m.TestId
}

// SetTestId sets the TestId for the TorpedoTestMetaData
func (m *TorpedoTestMetaData) SetTestId(id string) {
	m.TestId = id
}

// GetTestUid returns the TorpedoTest uid
func (m *TorpedoTestMetaData) GetTestUid() string {
	return m.GetTestId()
}

// NewTorpedoTestMetaData creates a new instance of the TorpedoTestMetaData
func NewTorpedoTestMetaData() *TorpedoTestMetaData {
	newTorpedoTestMetaData := &TorpedoTestMetaData{}
	newTorpedoTestMetaData.SetTestId("")
	return newTorpedoTestMetaData
}

// TorpedoTestConfig represents the configuration for a TorpedoTest
type TorpedoTestConfig struct {
	TorpedoTestMetaData   *TorpedoTestMetaData
	TorpedoTestController *TorpedoTestController
}

// GetTorpedoTestMetaData returns the TorpedoTestMetaData associated with the TorpedoTestConfig
func (c *TorpedoTestConfig) GetTorpedoTestMetaData() *TorpedoTestMetaData {
	return c.TorpedoTestMetaData
}

// SetTorpedoTestMetaData sets the TorpedoTestMetaData for the TorpedoTestConfig
func (c *TorpedoTestConfig) SetTorpedoTestMetaData(metaData *TorpedoTestMetaData) {
	c.TorpedoTestMetaData = metaData
}

// GetTorpedoTestController returns the TorpedoTestController associated with the TorpedoTestConfig
func (c *TorpedoTestConfig) GetTorpedoTestController() *TorpedoTestController {
	return c.TorpedoTestController
}

// SetTorpedoTestController sets the TorpedoTestController for the TorpedoTestConfig
func (c *TorpedoTestConfig) SetTorpedoTestController(controller *TorpedoTestController) {
	c.TorpedoTestController = controller
}

func (c *TorpedoTestConfig) Start(testRailId int, testName string, testDescription string, testAuthor string, apps []string, tags map[string]string) {
	if testRailId != 0 {
		c.GetTorpedoTestController().GetTorpedoTestManager().IsTorpedoTestPresent()
		if pxBackupTorpedoTestInfo, ok := PxBackupTorpedoTestInfoMap[testRailId]; ok {
			err := fmt.Errorf("the test [%s] shares the same TestRail id as [%s] and has already been executed", testName, pxBackupTorpedoTestInfo.TestName)
			return err
		}
	}
	testTags := map[string]string{
		"author": testAuthor,
		"apps":   strings.Join(apps, ","),
	}
	if len(tags) > 0 {
		for tagKey, tagValue := range tags[0] {
			testTags[tagKey] = tagValue
		}
	}
	testLogger := tests.CreateLogger(fmt.Sprintf("%s.log", testName))
	log.SetTorpedoFileOutput(testLogger)
	tests.Inst().Dash.TestCaseBegin(testName, testDescription, strconv.Itoa(testRailId), testTags)
	var testRunIdForSuite int
	if tests.TestRailSetupSuccessful && testRailId != 0 {
		testRunIdForSuite = testrailuttils.AddRunsToMilestone(testRailId)
	}
	pxBackupTorpedoTestInfo := &PxBackupTorpedoTestInfo{
		TestName:          testName,
		TestDescription:   testDescription,
		TestRailID:        testRailId,
		TestAuthor:        testAuthor,
		TestTags:          testTags,
		TestLogger:        testLogger,
		TestRunIdForSuite: testRunIdForSuite,
	}
	PxBackupTorpedoTestInfoMap[testRailId] = pxBackupTorpedoTestInfo
	return nil
}

func (c *TorpedoTestConfig) End() {
	if pxBackupTorpedoTestInfo, ok := PxBackupTorpedoTestInfoMap[testRailId]; ok {
		tests.CloseLogger(pxBackupTorpedoTestInfo.testLogger)
		tests.Inst().Dash.TestCaseEnd()
		if tests.TestRailSetupSuccessful && pxBackupTorpedoTestInfo.testRailID != 0 && pxBackupTorpedoTestInfo.testRunIdForSuite != 0 {
			contexts := make([]*scheduler.Context, 0)
			tests.AfterEachTest(contexts, pxBackupTorpedoTestInfo.testRailID, pxBackupTorpedoTestInfo.testRunIdForSuite)
		}
	} else {
		err := fmt.Errorf("no test has been executed with the TestRail id [%d]", testRailId)
		return err
	}
	return nil
}

func NewTorpedoTestConfig() *TorpedoTestConfig {
	newTorpedoTestConfig := &TorpedoTestConfig{}
	newTorpedoTestConfig.SetTorpedoTestMetaData(NewTorpedoTestMetaData())
	newTorpedoTestConfig.SetTorpedoTestController(nil)
	return newTorpedoTestConfig
}

// TorpedoTest represents an TorpedoTest
type TorpedoTest struct {
	TestName          string
	TestDescription   string
	TestAuthor        string
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
func (t *TorpedoTest) SetTestName(name string) {
	t.TestName = name
}

// GetTestDescription returns the TestDescription associated with the TorpedoTest
func (t *TorpedoTest) GetTestDescription() string {
	return t.TestDescription
}

// SetTestDescription sets the TestDescription for the TorpedoTest
func (t *TorpedoTest) SetTestDescription(description string) {
	t.TestDescription = description
}

// GetTestAuthor returns the TestAuthor associated with the TorpedoTest
func (t *TorpedoTest) GetTestAuthor() string {
	return t.TestAuthor
}

// SetTestAuthor sets the TestAuthor for the TorpedoTest
func (t *TorpedoTest) SetTestAuthor(author string) {
	t.TestAuthor = author
}

// GetTestRailID returns the TestRailID associated with the TorpedoTest
func (t *TorpedoTest) GetTestRailID() int {
	return t.TestRailID
}

// SetTestRailID sets the TestRailID for the TorpedoTest
func (t *TorpedoTest) SetTestRailID(id int) {
	t.TestRailID = id
}

// GetTestRunIDForSuite returns the TestRunIDForSuite associated with the TorpedoTest
func (t *TorpedoTest) GetTestRunIDForSuite() int {
	return t.TestRunIDForSuite
}

// SetTestRunIDForSuite sets the TestRunIDForSuite for the TorpedoTest
func (t *TorpedoTest) SetTestRunIDForSuite(id int) {
	t.TestRunIDForSuite = id
}

// GetTestTags returns the TestTags associated with the TorpedoTest
func (t *TorpedoTest) GetTestTags() map[string]string {
	return t.TestTags
}

// SetTestTags sets the TestTags for the TorpedoTest
func (t *TorpedoTest) SetTestTags(tags map[string]string) {
	t.TestTags = tags
}

// GetTestLogger returns the TestLogger associated with the TorpedoTest
func (t *TorpedoTest) GetTestLogger() *lumberjack.Logger {
	return t.TestLogger
}

// SetTestLogger sets the TestLogger for the TorpedoTest
func (t *TorpedoTest) SetTestLogger(logger *lumberjack.Logger) {
	t.TestLogger = logger
}

// NewTorpedoTest creates a new instance of the TorpedoTest
func NewTorpedoTest() *TorpedoTest {
	newTorpedoTest := &TorpedoTest{}
	newTorpedoTest.SetTestName("")
	newTorpedoTest.SetTestDescription("")
	newTorpedoTest.SetTestAuthor("")
	newTorpedoTest.SetTestRailID(0)
	newTorpedoTest.SetTestRunIDForSuite(0)
	newTorpedoTest.SetTestTags(make(map[string]string, 0))
	newTorpedoTest.SetTestLogger(nil)
	return newTorpedoTest
}

// TorpedoTestManager represents a manager for TorpedoTest
type TorpedoTestManager struct {
	sync.RWMutex
	TorpedoTestMap         map[string]*TorpedoTest
	RemovedTorpedoTestsMap map[string][]*TorpedoTest
}

// GetTorpedoTestMap returns the TorpedoTestMap of the TorpedoTestManager
func (m *TorpedoTestManager) GetTorpedoTestMap() map[string]*TorpedoTest {
	m.RLock()
	defer m.RUnlock()
	return m.TorpedoTestMap
}

// SetTorpedoTestMap sets the TorpedoTestMap of the TorpedoTestManager
func (m *TorpedoTestManager) SetTorpedoTestMap(testMap map[string]*TorpedoTest) {
	m.Lock()
	defer m.Unlock()
	m.TorpedoTestMap = testMap
}

// GetRemovedTorpedoTestsMap returns the RemovedTorpedoTestsMap of the TorpedoTestManager
func (m *TorpedoTestManager) GetRemovedTorpedoTestsMap() map[string][]*TorpedoTest {
	m.RLock()
	defer m.RUnlock()
	return m.RemovedTorpedoTestsMap
}

// SetRemovedTorpedoTestsMap sets the RemovedTorpedoTestsMap of the TorpedoTestManager
func (m *TorpedoTestManager) SetRemovedTorpedoTestsMap(removedTestsMap map[string][]*TorpedoTest) {
	m.Lock()
	defer m.Unlock()
	m.RemovedTorpedoTestsMap = removedTestsMap
}

// GetTorpedoTest returns the TorpedoTest with the given TorpedoTest uid
func (m *TorpedoTestManager) GetTorpedoTest(torpedoTestUid string) *TorpedoTest {
	m.RLock()
	defer m.RUnlock()
	return m.GetTorpedoTestMap()[torpedoTestUid]
}

// IsTorpedoTestPresent checks if the TorpedoTest with the given TorpedoTest uid is present
func (m *TorpedoTestManager) IsTorpedoTestPresent(torpedoTestUid string) bool {
	m.RLock()
	defer m.RUnlock()
	_, isPresent := m.GetTorpedoTestMap()[torpedoTestUid]
	return isPresent
}

// SetTorpedoTest sets the TorpedoTest with the given TorpedoTest uid
func (m *TorpedoTestManager) SetTorpedoTest(torpedoTestUid string, torpedoTest *TorpedoTest) {
	m.Lock()
	defer m.Unlock()
	m.GetTorpedoTestMap()[torpedoTestUid] = torpedoTest
}

// DeleteTorpedoTest deletes the TorpedoTest with the given TorpedoTest uid
func (m *TorpedoTestManager) DeleteTorpedoTest(torpedoTestUid string) {
	m.Lock()
	defer m.Unlock()
	delete(m.GetTorpedoTestMap(), torpedoTestUid)
}

// RemoveTorpedoTest removes the TorpedoTest with the given TorpedoTest uid
func (m *TorpedoTestManager) RemoveTorpedoTest(torpedoTestUid string) {
	m.Lock()
	defer m.Unlock()
	m.GetRemovedTorpedoTestsMap()[torpedoTestUid] = append(m.GetRemovedTorpedoTestsMap()[torpedoTestUid], m.GetTorpedoTest(torpedoTestUid))
	m.DeleteTorpedoTest(torpedoTestUid)
}

// NewTorpedoTestManager creates a new instance of the TorpedoTestManager
func NewTorpedoTestManager() *TorpedoTestManager {
	newTorpedoTestManager := &TorpedoTestManager{}
	newTorpedoTestManager.SetTorpedoTestMap(make(map[string]*TorpedoTest, 0))
	newTorpedoTestManager.SetRemovedTorpedoTestsMap(make(map[string][]*TorpedoTest, 0))
	return newTorpedoTestManager
}
