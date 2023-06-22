package torpedotest

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/backup/utils"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/testrailuttils"
	"github.com/portworx/torpedo/tests"
	"strconv"
	"strings"
)

// CanStart checks if TorpedoTest CanStart
func (c *TorpedoTestConfig) CanStart(testRailId int) error {
	torpedoTestUid := c.GetTorpedoTestMetaData().GetTestUid()
	torpedoTestManager := c.GetTorpedoTestController().GetTorpedoTestManager()
	for tTestUid, tTest := range torpedoTestManager.GetTorpedoTestMap() {
		torpedoTestName := torpedoTestManager.GetTorpedoTest(tTestUid).GetTestName()
		if tTestUid == torpedoTestUid {
			err := fmt.Errorf("torpedo-test [name: %s] has the same test ID [%s] and has already started", torpedoTestName, tTestUid)
			return utils.ProcessError(err)
		}
		if testRailId != 0 && tTest.GetTestRailID() == testRailId {
			err := fmt.Errorf("torpedo-test [name: %s] has the same TestRail ID [%d] and has already started", torpedoTestName, testRailId)
			return utils.ProcessError(err)
		}
	}
	return nil
}

// Start creates a logger, configures the Aetos Dashboard for the specified test, and initializes controllers
func (c *TorpedoTestConfig) Start(testName string, testDescription string, testRailId int, testMaintainer utils.TestMaintainer, testApps []string, testTags map[string]string) error {
	if testTags == nil {
		testTags = make(map[string]string, 0)
	}
	testTags["apps"] = strings.Join(testApps, ",")
	testLogger := tests.CreateLogger(fmt.Sprintf("%s.log", testName))
	log.SetTorpedoFileOutput(testLogger)
	tests.Inst().Dash.TestCaseBegin(testName, testDescription, strconv.Itoa(testRailId), testTags)
	torpedoTest := NewTorpedoTest()
	torpedoTest.SetTestName(testName)
	torpedoTest.SetTestDescription(testDescription)
	torpedoTest.SetTestRailID(testRailId)
	torpedoTest.SetTestMaintainer(testMaintainer)
	torpedoTest.SetTestLogger(testLogger)
	torpedoTest.SetTestTags(testTags)
	if tests.TestRailSetupSuccessful && testRailId != 0 {
		testRunIdForSuite := testrailuttils.AddRunsToMilestone(testRailId)
		torpedoTest.SetTestRunIDForSuite(testRunIdForSuite)
	}
	c.GetTorpedoTestController().GetTorpedoTestManager().SetTorpedoTest(c.GetTorpedoTestMetaData().GetTestUid(), torpedoTest)
	return nil
}

// CanEnd checks if TorpedoTest CanEnd
func (c *TorpedoTestConfig) CanEnd() error {
	torpedoTestUid := c.GetTorpedoTestMetaData().GetTestUid()
	if !c.GetTorpedoTestController().GetTorpedoTestManager().IsTorpedoTestPresent(torpedoTestUid) {
		err := fmt.Errorf("torpedo-test [%s] has not started yet", torpedoTestUid)
		return utils.ProcessError(err)
	}
	return nil
}

// End ends the specified test and performs cleanup
func (c *TorpedoTestConfig) End() error {
	torpedoTestManager := c.GetTorpedoTestController().GetTorpedoTestManager()
	torpedoTest := torpedoTestManager.GetTorpedoTest(c.GetTorpedoTestMetaData().GetTestUid())
	tests.CloseLogger(torpedoTest.GetTestLogger())
	tests.Inst().Dash.TestCaseEnd()
	if tests.TestRailSetupSuccessful && torpedoTest.GetTestRailID() != 0 && torpedoTest.GetTestRunIDForSuite() != 0 {
		contexts := make([]*scheduler.Context, 0)
		tests.AfterEachTest(contexts, torpedoTest.GetTestRailID(), torpedoTest.GetTestRunIDForSuite())
	}
	torpedoTestManager.RemoveTorpedoTest(c.GetTorpedoTestMetaData().GetTestUid())
	return nil
}
