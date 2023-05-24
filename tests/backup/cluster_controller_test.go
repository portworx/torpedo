package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/portworx/torpedo/drivers/backup/controllers/cluster"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

// This testcase demonstrates the various functionalities of the ClusterController
var _ = Describe("{ClusterControllerDemo}", func() {
	var (
		testRailId           int
		clusterControllerMap map[string]*cluster.ClusterController
	)

	JustBeforeEach(func() {
		testRailId = 31313

		Step("Add source-cluster controller to cluster-controller-map", func() {
			// source-cluster controller will help manage the source cluster
			err := cluster.AddSourceClusterControllerToMap(&clusterControllerMap, testRailId)
			log.FailOnError(err, "failed to add source-cluster controller to cluster-controller-map")
		})

		Step("Add destination-cluster controller to cluster-controller-map", func() {
			// destination-cluster controller will help manage the destination cluster
			err := cluster.AddDestinationClusterControllerToMap(&clusterControllerMap, testRailId)
			log.FailOnError(err, "failed to add destination-cluster controller to cluster-controller-map")
		})
	})

	It("Cluster Controller Demo", func() {

		Step("Schedule each application on a unique namespace in source-cluster", func() {
			for _, appKey := range Inst().AppList {
				namespace := fmt.Sprintf("%s-app-namespace-step-1", appKey)

				err := clusterControllerMap["source-cluster"].Application(appKey).ScheduleOnNamespace(namespace)
				log.FailOnError(err, fmt.Sprintf("failed to schedule the application %s on the namespace %s", appKey, namespace))

				err = clusterControllerMap["source-cluster"].Namespace(namespace).Application(appKey).Validate()
				log.FailOnError(err, fmt.Sprintf("failed to validate the application %s on the namespace %s", appKey, namespace))

				err = clusterControllerMap["source-cluster"].Namespace(namespace).Application(appKey).Destroy()
				log.FailOnError(err, fmt.Sprintf("failed to destroy the application %s on the namespace %s", appKey, namespace))
			}
		})

		Step("Schedule all applications on a single namespace in destination-cluster", func() {
			namespace := "all-app-list-namespace-step-2"

			err := clusterControllerMap["destination-cluster"].MultipleApplications(Inst().AppList).ScheduleOnNamespace(namespace)
			log.FailOnError(err, "failed to schedule all applications on a single namespace")

			err = clusterControllerMap["destination-cluster"].Namespace(namespace).Validate()
			log.FailOnError(err, "failed to validate all the applications in the namespace %s", namespace)

			err = clusterControllerMap["destination-cluster"].Namespace(namespace).Destroy()
			log.FailOnError(err, "failed to destroy all the applications in the namespace %s", namespace)
		})

		Step("Schedule each application twice in source-cluster", func() {
			for _, appKey := range Inst().AppList {
				namespacePrefix := fmt.Sprintf("%s-app-namespace-step-3", appKey)

				namespaces, err := clusterControllerMap["source-cluster"].Application(appKey).ScheduleOnPrefixedNamespaces(namespacePrefix, 2)
				log.FailOnError(err, fmt.Sprintf("failed to schedule the application %s on prefixed namespace %s", appKey, namespacePrefix))

				for _, namespace := range namespaces {
					err = clusterControllerMap["source-cluster"].Namespace(namespace).Application(appKey).Validate()
					log.FailOnError(err, fmt.Sprintf("failed to validate the application %s on the namespace %s", appKey, namespace))

					err = clusterControllerMap["source-cluster"].Namespace(namespace).Application(appKey).Destroy()
					log.FailOnError(err, fmt.Sprintf("failed to destroy the application %s on the namespace %s", appKey, namespace))
				}
			}
		})

		Step("Schedule each application twice in the same namespace within the destination-cluster", func() {
			for _, appKey := range Inst().AppList {
				namespace := fmt.Sprintf("%s-app-namespace-step-4", appKey)

				err := clusterControllerMap["destination-cluster"].Application(appKey).ScheduleOnNamespace(namespace)
				log.FailOnError(err, fmt.Sprintf("failed to schedule the application %s on the namespace %s", appKey, namespace))

				err = clusterControllerMap["destination-cluster"].Namespace(namespace).Application(appKey).Validate()
				log.FailOnError(err, fmt.Sprintf("failed to validate the application %s on the namespace %s", appKey, namespace))
			}

			for _, appKey := range Inst().AppList {
				namespace := fmt.Sprintf("%s-app-namespace-step-4", appKey)

				err := clusterControllerMap["destination-cluster"].Application(appKey).ScheduleOnNamespace(namespace)
				log.FailOnError(err, fmt.Sprintf("failed to schedule the application %s on the namespace %s", appKey, namespace))

				err = clusterControllerMap["destination-cluster"].Namespace(namespace).Application(appKey).Validate()
				log.FailOnError(err, fmt.Sprintf("failed to validate the application %s on the namespace %s", appKey, namespace))
			}
		})
	})

	JustAfterEach(func() {
		// To destroy the applications scheduled in Step 4
		for clusterName, clusterController := range clusterControllerMap {
			err := clusterController.Cleanup()
			log.FailOnError(err, fmt.Sprintf("%s", clusterName))
		}
	})
})
