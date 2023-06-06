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
			log.InfoD("Performing Step 1")
			for _, appKey := range Inst().AppList {
				namespace := fmt.Sprintf("%s-app-namespace-step-1", appKey)

				err := helper.TestCaseControllerCollectionMap[testRailId].GetClusterController("source-cluster").Application(appKey).ScheduleOnNamespace(namespace)
				log.FailOnError(err, fmt.Sprintf("failed to schedule the application %s on the namespace %s", appKey, namespace))

				err = helper.TestCaseControllerCollectionMap[testRailId].GetClusterController("source-cluster").SelectNamespace(namespace).SelectApplication(appKey).Validate()
				log.FailOnError(err, fmt.Sprintf("failed to validate the application %s on the namespace %s", appKey, namespace))

				//err = helper.TestCaseControllerCollectionMap[testRailId].GetClusterController("source-cluster").SelectNamespace(namespace).SelectApplication(appKey).Destroy()
				//log.FailOnError(err, fmt.Sprintf("failed to destroy the application %s on the namespace %s", appKey, namespace))
			}
		})

		//Step("Schedule all applications on a single namespace in destination-cluster", func() {
		//	log.InfoD("Performing Step 2")
		//
		//	namespace := "all-postgres-namespace-step-2"
		//
		//	err := helper.TestCaseControllerCollectionMap[testRailId].GetClusterController("destination-cluster").MultipleApplications(Inst().AppList).ScheduleOnNamespace(namespace)
		//	log.FailOnError(err, "failed to schedule all applications on a single namespace")
		//
		//	err = helper.TestCaseControllerCollectionMap[testRailId].GetClusterController("destination-cluster").SelectNamespace(namespace).Validate()
		//	log.FailOnError(err, "failed to validate all the applications in the namespace %s", namespace)
		//
		//	err = helper.TestCaseControllerCollectionMap[testRailId].GetClusterController("destination-cluster").SelectNamespace(namespace).Destroy()
		//	log.FailOnError(err, "failed to destroy all the applications in the namespace %s", namespace)
		//})
		//
		//Step("Schedule each application twice in source-cluster", func() {
		//	log.InfoD("Performing Step 3")
		//
		//	for _, appKey := range Inst().AppList {
		//		namespacePrefix := fmt.Sprintf("%s-app-namespace-step-3", appKey)
		//
		//		namespaces, err := helper.TestCaseControllerCollectionMap[testRailId].GetClusterController("source-cluster").Application(appKey).ScheduleOnMultipleNamespaces(namespacePrefix, 2)
		//		log.FailOnError(err, fmt.Sprintf("failed to schedule the application %s on prefixed namespace %s", appKey, namespacePrefix))
		//
		//		for _, namespace := range namespaces {
		//			err = helper.TestCaseControllerCollectionMap[testRailId].GetClusterController("source-cluster").SelectNamespace(namespace).SelectApplication(appKey).Validate()
		//			log.FailOnError(err, fmt.Sprintf("failed to validate the application %s on the namespace %s", appKey, namespace))
		//
		//			err = helper.TestCaseControllerCollectionMap[testRailId].GetClusterController("source-cluster").SelectNamespace(namespace).SelectApplication(appKey).Destroy()
		//			log.FailOnError(err, fmt.Sprintf("failed to destroy the application %s on the namespace %s", appKey, namespace))
		//		}
		//	}
		//})
		//
		//Step("Schedule each application twice in the same namespace within the destination-cluster", func() {
		//	log.InfoD("Performing Step 4")
		//
		//	for _, appKey := range Inst().AppList {
		//		namespace := fmt.Sprintf("%s-app-namespace-step-4", appKey)
		//
		//		err := helper.TestCaseControllerCollectionMap[testRailId].GetClusterController("destination-cluster").Application(appKey).ScheduleOnNamespace(namespace)
		//		log.FailOnError(err, fmt.Sprintf("failed to schedule the application %s on the namespace %s", appKey, namespace))
		//
		//		err = helper.TestCaseControllerCollectionMap[testRailId].GetClusterController("destination-cluster").SelectNamespace(namespace).SelectApplication(appKey).Validate()
		//		log.FailOnError(err, fmt.Sprintf("failed to validate the application %s on the namespace %s", appKey, namespace))
		//	}
		//
		//	for _, appKey := range Inst().AppList {
		//		namespace := fmt.Sprintf("%s-app-namespace-step-4", appKey)
		//
		//		err := helper.TestCaseControllerCollectionMap[testRailId].GetClusterController("destination-cluster").Application(appKey).ScheduleOnNamespace(namespace)
		//		log.FailOnError(err, fmt.Sprintf("failed to schedule the application %s on the namespace %s", appKey, namespace))
		//
		//		err = helper.TestCaseControllerCollectionMap[testRailId].GetClusterController("destination-cluster").SelectNamespace(namespace).SelectApplication(appKey).Validate()
		//		log.FailOnError(err, fmt.Sprintf("failed to validate the application %s on the namespace %s", appKey, namespace))
		//	}
		//})
	})

	JustAfterEach(func() {
		defer func() {
			err := helper.EndPxBackupTorpedoTest(testRailId)
			log.FailOnError(err, "end px-backup torpedo test failed")
		}()
		err := helper.TestCaseControllerCollectionMap[testRailId].GetClusterController("source-cluster").Cleanup()
		log.FailOnError(err, fmt.Sprintf("source cluster clean up failed"))
		// To destroy the applications scheduled in Step 4
		//for clusterName, clusterController := range clusterControllerMap {
		//	err := clusterController.Cleanup()
		//	log.FailOnError(err, fmt.Sprintf("%s", clusterName))
		//}
	})
})
