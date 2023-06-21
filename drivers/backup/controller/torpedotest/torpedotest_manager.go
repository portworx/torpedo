package torpedotest

import (
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
)

type TorpedoTestMetaData struct {
	TestName          string
	TestDescription   string
	TestAuthor        string
	TestRailID        int
	TestRunIDForSuite int
	TestTags          map[string]string
	TestLogger        *lumberjack.Logger
}

func (m *TorpedoTestMetaData) GetTestName() string {
	return m.TestName
}

func (m *TorpedoTestMetaData) SetTestName(name string) {
	m.TestName = name
}

func (m *TorpedoTestMetaData) GetTestDescription() string {
	return m.TestDescription
}

func (m *TorpedoTestMetaData) SetTestDescription(description string) {
	m.TestDescription = description
}

func (m *TorpedoTestMetaData) GetTestAuthor() string {
	return m.TestAuthor
}

func (m *TorpedoTestMetaData) SetTestAuthor(author string) {
	m.TestAuthor = author
}

func (m *TorpedoTestMetaData) GetTestRailID() int {
	return m.TestRailID
}

func (m *TorpedoTestMetaData) SetTestRailID(id int) {
	m.TestRailID = id
}

func (m *TorpedoTestMetaData) GetTestRunIDForSuite() int {
	return m.TestRunIDForSuite
}

func (m *TorpedoTestMetaData) SetTestRunIDForSuite(id int) {
	m.TestRunIDForSuite = id
}

func (m *TorpedoTestMetaData) GetTestTags() map[string]string {
	return m.TestTags
}

func (m *TorpedoTestMetaData) SetTestTags(tags map[string]string) {
	m.TestTags = tags
}

func (m *TorpedoTestMetaData) GetTestLogger() *lumberjack.Logger {
	return m.TestLogger
}

func (m *TorpedoTestMetaData) SetTestLogger(logger *lumberjack.Logger) {
	m.TestLogger = logger
}

func (m *TorpedoTestMetaData) GetTorpedoTestUid() string {
	return fmt.Sprintf("%d", m.GetTestRailID())
}

func NewTorpedoTest() *TorpedoTestMetaData {
	newTorpedoTest := &TorpedoTestMetaData{}
	newTorpedoTest.SetTestName("")
	newTorpedoTest.SetTestDescription("")
	newTorpedoTest.SetTestAuthor("")
	newTorpedoTest.SetTestRailID(0)
	newTorpedoTest.SetTestRunIDForSuite(0)
	newTorpedoTest.SetTestTags(make(map[string]string, 0))
	newTorpedoTest.SetTestLogger(nil)
	return newTorpedoTest
}

type TorpedoTestManager struct {
	TorpedoTestMap           map[string]*TorpedoTestMetaData
	CompletedTorpedoTestsMap map[string][]*TorpedoTestMetaData
}

func (m *TorpedoTestManager) GetTorpedoTestMap() map[string]*TorpedoTestMetaData {
	return m.TorpedoTestMap
}

func (m *TorpedoTestManager) SetTorpedoTestMap(testMap map[string]*TorpedoTestMetaData) {
	m.TorpedoTestMap = testMap
}

func (m *TorpedoTestManager) GetCompletedTorpedoTestsMap() map[string][]*TorpedoTestMetaData {
	return m.CompletedTorpedoTestsMap
}

func (m *TorpedoTestManager) SetCompletedTorpedoTestsMap(completedTestsMap map[string][]*TorpedoTestMetaData) {
	m.CompletedTorpedoTestsMap = completedTestsMap
}
