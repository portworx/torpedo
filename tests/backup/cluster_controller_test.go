package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/portworx/torpedo/drivers/backup/helper"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

// This testcase demonstrates the various functionalities of the ClusterController
var _ = Describe("{ClusterControllerDemo}", func() {
	var (
		testRailId int
	)

	JustBeforeEach(func() {
		testRailId = 31313
		err := helper.StartPxBackupTorpedoTest(testRailId, "ClusterControllerDemo", "It tests cluster-controller functionalities", "kphalgun-px", Inst().AppList)
		log.FailOnError(err, "start px-backup torpedo test failed")
	})

	It("Cluster Controller Demo", func() {

		Step("Schedule each application on a unique namespace in source-cluster", func() {
			log.InfoD("Performing Step 1v2")
			for _, appKey := range Inst().AppList {
				namespace := fmt.Sprintf("%s-app-namespace-step-1v3", appKey)

				//err := helper.TestCaseControllerCollectionMap[testRailId].GetClusterController("source-cluster").Namespace(namespace).App(appKey).Schedule()
				//log.FailOnError(err, fmt.Sprintf("failed to schedule the application %s on the namespace %s", appKey, namespace))
				//
				//err := helper.TestCaseControllerCollectionMap[testRailId].GetClusterController("source-cluster").Namespace(namespace).App(appKey, "31313-0-v3").Schedule()
				//log.FailOnError(err, fmt.Sprintf("failed to schedule the application %s on the namespace %s", appKey, namespace))
				//
				//err = helper.TestCaseControllerCollectionMap[testRailId].GetClusterController("source-cluster").Namespace(namespace).App(appKey, "31313-0-v3").Validate()
				//log.FailOnError(err, fmt.Sprintf("failed to validate the application %s on the namespace %s", appKey, namespace))
				//
				//err = helper.TestCaseControllerCollectionMap[testRailId].GetClusterController("source-cluster").Namespace(namespace).App(appKey, "31313-0-v3").TearDown()
				//log.FailOnError(err, fmt.Sprintf("failed to teardown the application %s on the namespace %s", appKey, namespace))

				//err = helper.TestCaseControllerCollectionMap[testRailId].GetClusterController("source-cluster").SelectNamespace(namespace).SelectApplication(appKey).Validate()
				//log.FailOnError(err, fmt.Sprintf("failed to validate the application %s on the namespace %s", appKey, namespace))

				//err = helper.TestCaseControllerCollectionMap[testRailId].GetClusterController("source-cluster").SelectNamespace(namespace).SelectApplication(appKey).Destroy()
				//log.FailOnError(err, fmt.Sprintf("failed to destroy the application %s on the namespace %s", appKey, namespace))
			}
		})
	})

	JustAfterEach(func() {
		defer func() {
			err := helper.EndPxBackupTorpedoTest(testRailId)
			log.FailOnError(err, "end px-backup torpedo test failed")
		}()
	})
})
