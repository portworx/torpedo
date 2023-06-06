package helper

import (
	"fmt"
	cluster "github.com/portworx/torpedo/drivers/backup/controllers/clusterv2"
	"github.com/portworx/torpedo/drivers/backup/utils"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/testrailuttils"
	"github.com/portworx/torpedo/tests"
	"gopkg.in/natefinch/lumberjack.v2"
	"strconv"
	"strings"
)

type TestCaseControllerCollection struct {
	clusterControllerMap map[string]*cluster.ClusterController
}

func (m *TestCaseControllerCollection) GetClusterController(clusterName string) *cluster.ClusterController {
	return m.clusterControllerMap[clusterName]
}

var (
	PxBackupTorpedoTestInfoMap      = make(map[int]*PxBackupTorpedoTestInfo, 0)
	TestCaseControllerCollectionMap = make(map[int]*TestCaseControllerCollection, 0)
)

// PxBackupTorpedoTestInfo holds information of a particular test
type PxBackupTorpedoTestInfo struct {
	testName          string
	testDescription   string
	testAuthor        string
	testRailID        int
	testRunIdForSuite int
	testTags          map[string]string
	testLogger        *lumberjack.Logger
}

// AddDefaultClusterControllersToMap adds default ClusterController instances associated with the specified testRailId for use by a test case
func AddDefaultClusterControllersToMap(clusterControllerMap *map[string]*cluster.ClusterController, id int) error {
	sourceClusterConfigPath, err := utils.GetSourceClusterConfigPath()
	if err != nil {
		return utils.ProcessError(err)
	}
	destinationClusterConfigPath, err := utils.GetDestinationClusterConfigPath()
	if err != nil {
		return utils.ProcessError(err)
	}
	clustersInfo := []*cluster.ClusterInfo{
		cluster.Cluster(id, utils.DefaultInClusterName, cluster.GlobalInClusterConfigPath).IsHyperConverged(),         // in-cluster
		cluster.Cluster(id, utils.DefaultSourceClusterName, sourceClusterConfigPath).IsHyperConverged().IsInCluster(), // source-cluster
		cluster.Cluster(id, utils.DefaultDestinationClusterName, destinationClusterConfigPath).IsHyperConverged(),     // destination-cluster
	}
	return cluster.AddClusterControllersToMap(clusterControllerMap, clustersInfo)
}

// StartPxBackupTorpedoTest creates a logger, configures the Aetos Dashboard for the specified test, and initializes controllers
func StartPxBackupTorpedoTest(testRailId int, testName string, testDescription string, testAuthor string, apps []string, tags ...map[string]string) error {
	if testRailId != 0 {
		if pxBackupTorpedoTestInfo, ok := PxBackupTorpedoTestInfoMap[testRailId]; ok {
			err := fmt.Errorf("the test [%s] shares the same TestRail id as [%s] and has already been executed", testName, pxBackupTorpedoTestInfo.testName)
			return utils.ProcessError(err)
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
		testName:          testName,
		testDescription:   testDescription,
		testRailID:        testRailId,
		testAuthor:        testAuthor,
		testTags:          testTags,
		testLogger:        testLogger,
		testRunIdForSuite: testRunIdForSuite,
	}
	PxBackupTorpedoTestInfoMap[testRailId] = pxBackupTorpedoTestInfo
	TestCaseControllerCollectionMap[testRailId] = &TestCaseControllerCollection{}
	err := AddDefaultClusterControllersToMap(&TestCaseControllerCollectionMap[testRailId].clusterControllerMap, testRailId)
	if err != nil {
		return utils.ProcessError(err)
	}
	return nil
}

// EndPxBackupTorpedoTest ends the specified test and performs cleanup
func EndPxBackupTorpedoTest(testRailId int) error {
	if pxBackupTorpedoTestInfo, ok := PxBackupTorpedoTestInfoMap[testRailId]; ok {
		//for clusterName, clusterController := range TestCaseControllerCollectionMap[testRailId].clusterControllerMap {
		//	err := clusterController.Cleanup()
		//	debugMessage := fmt.Sprintf("cluster-name: %s", clusterName)
		//	return utils.ProcessError(err, debugMessage)
		//}
		tests.CloseLogger(pxBackupTorpedoTestInfo.testLogger)
		tests.Inst().Dash.TestCaseEnd()
		if tests.TestRailSetupSuccessful && pxBackupTorpedoTestInfo.testRailID != 0 && pxBackupTorpedoTestInfo.testRunIdForSuite != 0 {
			contexts := make([]*scheduler.Context, 0)
			tests.AfterEachTest(contexts, pxBackupTorpedoTestInfo.testRailID, pxBackupTorpedoTestInfo.testRunIdForSuite)
		}
	} else {
		err := fmt.Errorf("no test has been executed with the TestRail id [%d]", testRailId)
		return utils.ProcessError(err)
	}
	return nil
}
