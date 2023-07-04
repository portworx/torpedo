package torpedotest_manager

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/backup_controller/backup_utils"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/testrailuttils"
	"github.com/portworx/torpedo/tests"
	"strconv"
	"strings"
)

// CanStart checks if TorpedoTest CanStart
func (c *TorpedoTestConfig) CanStart(testRailId int) error {
	torpedoTestUid := c.GetTorpedoTestMetaData().GetTestUID()
	torpedoTestManager := c.GetTorpedoTestManager()
	for tTestUid, tTest := range torpedoTestManager.GetTorpedoTestMap() {
		torpedoTestName := torpedoTestManager.GetTorpedoTest(tTestUid).GetTestName()
		if tTestUid == torpedoTestUid {
			err := fmt.Errorf("torpedo-test [name: %s] has the same test ID [%s] and has already started", torpedoTestName, tTestUid)
			return backup_utils.ProcessError(err)
		}
		if testRailId != 0 && tTest.GetTestRailID() == testRailId {
			err := fmt.Errorf("torpedo-test [name: %s] has the same TestRail ID [%d] and has already started", torpedoTestName, testRailId)
			return backup_utils.ProcessError(err)
		}
	}
	return nil
}

// Start starts the TorpedoTest
func (c *TorpedoTestConfig) Start(testName string, testDescription string, testRailId int, testMaintainer backup_utils.TestMaintainer, testApps []string, testTags map[string]string) error {
	if testTags == nil {
		testTags = make(map[string]string, 0)
	}
	testTags["apps"] = strings.Join(testApps, ",")
	testLogger := tests.CreateLogger(fmt.Sprintf("%s.log", testName))
	log.SetTorpedoFileOutput(testLogger)
	tests.Inst().Dash.TestCaseBegin(testName, testDescription, strconv.Itoa(testRailId), testTags)
	testRunIdForSuite := DefaultTestRunIDForSuite
	if tests.TestRailSetupSuccessful && testRailId != 0 {
		testRunIdForSuite = testrailuttils.AddRunsToMilestone(testRailId)
	}
	torpedoTest := NewTorpedoTest(testName, testDescription, testMaintainer, testRailId, testRunIdForSuite, testTags, testLogger)
	c.GetTorpedoTestManager().SetTorpedoTest(c.GetTorpedoTestMetaData().GetTestUID(), torpedoTest)
	return nil
}

// CanEnd checks if TorpedoTest CanEnd
func (c *TorpedoTestConfig) CanEnd() error {
	torpedoTestUid := c.GetTorpedoTestMetaData().GetTestUID()
	if !c.GetTorpedoTestManager().IsTorpedoTestPresent(torpedoTestUid) {
		err := fmt.Errorf("torpedo-test [%s] has not started yet", torpedoTestUid)
		return backup_utils.ProcessError(err)
	}
	return nil
}

// End ends the TorpedoTest
func (c *TorpedoTestConfig) End() error {
	torpedoTestManager := c.GetTorpedoTestManager()
	torpedoTest := torpedoTestManager.GetTorpedoTest(c.GetTorpedoTestMetaData().GetTestUID())
	tests.CloseLogger(torpedoTest.GetTestLogger())
	tests.Inst().Dash.TestCaseEnd()
	if tests.TestRailSetupSuccessful && torpedoTest.GetTestRailID() != 0 && torpedoTest.GetTestRunIDForSuite() != 0 {
		contexts := make([]*scheduler.Context, 0)
		tests.AfterEachTest(contexts, torpedoTest.GetTestRailID(), torpedoTest.GetTestRunIDForSuite())
	}
	torpedoTestManager.RemoveTorpedoTest(c.GetTorpedoTestMetaData().GetTestUID())
	return nil
}
