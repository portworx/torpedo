package pxbackuptest

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

// CanStart checks if PxBackupTest CanStart
func (c *PxBackupTestConfig) CanStart(testRailId int) error {
	pxBackupTestUid := c.GetPxBackupTestMetaData().GetTestUid()
	pxBackupTestManager := c.GetPxBackupTestController().GetPxBackupTestManager()
	for tTestUid, tTest := range pxBackupTestManager.GetPxBackupTestMap() {
		PxBackupTestName := pxBackupTestManager.GetPxBackupTest(tTestUid).GetTestName()
		if tTestUid == pxBackupTestUid {
			err := fmt.Errorf("torpedo-test [name: %s] has the same test ID [%s] and has already started", PxBackupTestName, tTestUid)
			return utils.ProcessError(err)
		}
		if testRailId != 0 && tTest.GetTestRailID() == testRailId {
			err := fmt.Errorf("torpedo-test [name: %s] has the same TestRail ID [%d] and has already started", PxBackupTestName, testRailId)
			return utils.ProcessError(err)
		}
	}
	return nil
}

// Start creates a logger, configures the Aetos Dashboard for the specified test, and initializes controllers
func (c *PxBackupTestConfig) Start(testName string, testDescription string, testRailId int, testMaintainer utils.TestMaintainer, testApps []string, testTags map[string]string) error {
	if testTags == nil {
		testTags = make(map[string]string, 0)
	}
	testTags["apps"] = strings.Join(testApps, ",")
	testLogger := tests.CreateLogger(fmt.Sprintf("%s.log", testName))
	log.SetTorpedoFileOutput(testLogger)
	tests.Inst().Dash.TestCaseBegin(testName, testDescription, strconv.Itoa(testRailId), testTags)
	pxBackupTest := NewPxBackupTest()
	pxBackupTest.SetTestName(testName)
	pxBackupTest.SetTestDescription(testDescription)
	pxBackupTest.SetTestRailID(testRailId)
	pxBackupTest.SetTestMaintainer(testMaintainer)
	pxBackupTest.SetTestLogger(testLogger)
	pxBackupTest.SetTestTags(testTags)
	if tests.TestRailSetupSuccessful && testRailId != 0 {
		testRunIdForSuite := testrailuttils.AddRunsToMilestone(testRailId)
		pxBackupTest.SetTestRunIDForSuite(testRunIdForSuite)
	}
	c.GetPxBackupTestController().GetPxBackupTestManager().SetPxBackupTest(c.GetPxBackupTestMetaData().GetTestUid(), pxBackupTest)
	return nil
}

// CanEnd checks if PxBackupTest CanEnd
func (c *PxBackupTestConfig) CanEnd() error {
	PxBackupTestUid := c.GetPxBackupTestMetaData().GetTestUid()
	if !c.GetPxBackupTestController().GetPxBackupTestManager().IsPxBackupTestPresent(PxBackupTestUid) {
		err := fmt.Errorf("torpedo-test [%s] has not started yet", PxBackupTestUid)
		return utils.ProcessError(err)
	}
	return nil
}

// End ends the specified test and performs cleanup
func (c *PxBackupTestConfig) End() error {
	pxBackupTestManager := c.GetPxBackupTestController().GetPxBackupTestManager()
	pxBackupTest := pxBackupTestManager.GetPxBackupTest(c.GetPxBackupTestMetaData().GetTestUid())
	tests.CloseLogger(pxBackupTest.GetTestLogger())
	tests.Inst().Dash.TestCaseEnd()
	if tests.TestRailSetupSuccessful && pxBackupTest.GetTestRailID() != 0 && pxBackupTest.GetTestRunIDForSuite() != 0 {
		contexts := make([]*scheduler.Context, 0)
		tests.AfterEachTest(contexts, pxBackupTest.GetTestRailID(), pxBackupTest.GetTestRunIDForSuite())
	}
	pxBackupTestManager.RemovePxBackupTest(c.GetPxBackupTestMetaData().GetTestUid())
	return nil
}
