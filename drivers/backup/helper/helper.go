package helper

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/testrailuttils"
	"github.com/portworx/torpedo/tests"
	"strconv"
	"strings"
)

// StartPxBackupTorpedoTest creates a logger, configures the Aetos Dashboard for the specified test, and initializes controllers
func StartPxBackupTorpedoTest(testRailId int, testName string, testDescription string, testAuthor string, apps []string, tags ...map[string]string) error {
	if testRailId != 0 {
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

// EndPxBackupTorpedoTest ends the specified test and performs cleanup
func EndPxBackupTorpedoTest(testRailId int) error {
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
