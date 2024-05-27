package tests

import (
	"fmt"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	utils "github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
)

const AVAILABLE = "AVAILABLE"
const UNAVAILABLE = "UNAVAILABLE"
const TOMBSTONED = "TOMBSTONED"
const portorxLabel = "platform.portworx.io/pds"

var _ = Describe("{EnableandDisableNamespace}", func() {
	var (
		numberOfNamespacesTobeCreated int
		namespacePrefix               string
		allError                      []string
		oddNamespaces                 []string
		evenNamespaces                []string
		nsLablesRemove                map[string]string
		nsLablesApply                 map[string]string
		TotalToggles                  int
		waitTime                      time.Duration
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("EnableandDisableNamespace", "Enables and Disables pds on a namespace multiple times", nil, 0)
		numberOfNamespacesTobeCreated = 10 // Number of namespaces to be created by the testcase
		TotalToggles = 3                   // Total number of enable/disable toggles for each namespace
		namespacePrefix = "enabledisable-"
		nsLablesRemove = map[string]string{
			portorxLabel: "false",
		}
		nsLablesApply = map[string]string{
			portorxLabel: "true",
		}
		waitTime = 60 * time.Second
	})

	It("Enables and Disables pds on a namespace multiple times", func() {
		Step(fmt.Sprintf("Creating [%d] namespaces with labels", numberOfNamespacesTobeCreated), func() {
			var wg sync.WaitGroup

			log.InfoD("Creating [%d] namespaces with PDS labels present", numberOfNamespacesTobeCreated)

			for i := 0; i < numberOfNamespacesTobeCreated; i++ {
				wg.Add(1)

				nsName := namespacePrefix + RandomString(5) + "-" + strconv.Itoa(i)

				go func(nsName string) {

					defer wg.Done()
					defer GinkgoRecover()

					_, err := WorkflowNamespace.CreateNamespaces(nsName)
					if err != nil {
						allError = append(allError, err.Error())
					}

				}(nsName)

				if i%2 == 0 {
					evenNamespaces = append(evenNamespaces, nsName)
				} else {
					oddNamespaces = append(oddNamespaces, nsName)
				}
			}

			wg.Wait()
			if allError != nil {
				log.Errorf(strings.Join(allError, "\n"))
			}
			dash.VerifyFatal(len(allError), 0, "Verifying namespaces creation")
		})

		Step(fmt.Sprintf("Associating [%s] to [%s] for deployment", evenNamespaces[0], WorkflowProject.ProjectName), func() {
			log.InfoD("Associating [%s] to [%s] for deployment", evenNamespaces[0], WorkflowProject.ProjectName)
			err := WorkflowProject.Associate(
				[]string{},
				[]string{WorkflowNamespace.Namespaces[evenNamespaces[0]]},
				[]string{},
				[]string{},
				[]string{},
				[]string{},
			)
			log.FailOnError(err, "Unable to associate Cluster to Project")
			log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)
		})

		for i := 0; i < TotalToggles; i++ {
			Step("Removing labels from all odd namespaces", func() {
				var wg sync.WaitGroup
				log.InfoD("Executing [%d] toggle", (i + 1))

				log.InfoD("Removing labels from all odd namespaces")

				for _, namespace := range oddNamespaces {
					wg.Add(1)
					go func(ns string) {

						defer wg.Done()
						defer GinkgoRecover()

						_, err := utils.UpdatePDSNamespce(ns, nsLablesRemove)
						if err != nil {
							allError = append(allError, err.Error())
						}
					}(namespace)
				}

				wg.Wait()
				if allError != nil {
					log.Errorf(strings.Join(allError, "\n"))
				}
				dash.VerifyFatal(len(allError), 0, "Verifying update namespaces - Remove label from odd namespaces")
				log.InfoD("Sleeping for 60 seconds for the changes to be updated on control plane")
				time.Sleep(waitTime)
			})

			Step("Validating all current namespaces", func() {
				var wg sync.WaitGroup

				log.InfoD("Validating all current namespaces")
				for _, namespace := range oddNamespaces {
					wg.Add(1)
					go func(nsName string) {

						defer wg.Done()
						defer GinkgoRecover()

						ns, err := WorkflowNamespace.GetNamespace(nsName)
						if err != nil {
							allError = append(allError, fmt.Sprintf("Some error occurred while listing namespace. Error - [%s]", err.Error()))
						} else {
							if *ns.Status.Phase != UNAVAILABLE {
								allError = append(allError, fmt.Sprintf("[%s] is in [%s] state. Expected - [%s]", nsName, *ns.Status.Phase, UNAVAILABLE))
							}
						}
						log.Infof("[%s] - [%s]", nsName, *ns.Status.Phase)
					}(namespace)
				}

				for _, namespace := range evenNamespaces {
					wg.Add(1)
					go func(nsName string) {

						defer wg.Done()
						defer GinkgoRecover()

						ns, err := WorkflowNamespace.GetNamespace(nsName)
						if err != nil {
							allError = append(allError, fmt.Sprintf("Some error occurred while listing namespace. Error - [%s]", err.Error()))
						} else {
							if *ns.Status.Phase != AVAILABLE {
								allError = append(allError, fmt.Sprintf("[%s] is in [%s] state. Expected - [%s]", nsName, *ns.Status.Phase, AVAILABLE))
							}
						}
						log.Infof("[%s] - [%s]", nsName, *ns.Status.Phase)
					}(namespace)
				}

				wg.Wait()
				if allError != nil {
					log.Errorf(strings.Join(allError, "\n"))
				}
				dash.VerifyFatal(len(allError), 0, "Verifying namespaces on control plane")
			})

			Step("Toggling the lable on odd and even namespaces", func() {
				var wg sync.WaitGroup

				log.InfoD("Applying labels to odd namespaces")

				for _, namespace := range oddNamespaces {
					wg.Add(1)
					go func(nsName string) {

						defer wg.Done()
						defer GinkgoRecover()

						_, err := utils.UpdatePDSNamespce(nsName, nsLablesApply)
						if err != nil {
							allError = append(allError, err.Error())
						}
					}(namespace)
				}

				log.InfoD("Removing labels from even namespaces")
				for _, namespace := range evenNamespaces {
					wg.Add(1)
					go func(nsName string) {

						defer wg.Done()
						defer GinkgoRecover()

						_, err := utils.UpdatePDSNamespce(nsName, nsLablesRemove)
						if err != nil {
							allError = append(allError, err.Error())
						}
					}(namespace)
				}

				wg.Wait()
				if allError != nil {
					log.Errorf(strings.Join(allError, "\n"))
				}
				dash.VerifyFatal(len(allError), 0, "Verifying namespace toggle")
				log.InfoD("Sleeping for 60 seconds for the changes to be updated on control plane")
				time.Sleep(waitTime)
			})

			Step("Validating all current namespaces", func() {
				var wg sync.WaitGroup

				log.InfoD("Validating all current namespaces")

				for _, namespace := range evenNamespaces {
					wg.Add(1)
					go func(nsName string) {

						defer wg.Done()
						defer GinkgoRecover()

						ns, err := WorkflowNamespace.GetNamespace(nsName)
						if err != nil {
							allError = append(allError, fmt.Sprintf("Some error occurred while listing namespace. Error - [%s]", err.Error()))
						} else {
							if *ns.Status.Phase != UNAVAILABLE {
								allError = append(allError, fmt.Sprintf("[%s] is in [%s] state. Expected - [%s]", nsName, *ns.Status.Phase, UNAVAILABLE))
							}
						}
						log.Infof("[%s] - [%s]", nsName, *ns.Status.Phase)
					}(namespace)
				}

				for _, namespace := range oddNamespaces {
					wg.Add(1)
					go func(nsName string) {

						defer wg.Done()
						defer GinkgoRecover()

						ns, err := WorkflowNamespace.GetNamespace(nsName)
						if err != nil {
							allError = append(allError, fmt.Sprintf("Some error occurred while listing namespace. Error - [%s]", err.Error()))
						} else {
							if *ns.Status.Phase != AVAILABLE {
								allError = append(allError, fmt.Sprintf("[%s] is in [%s] state. Expected - [%s]", nsName, *ns.Status.Phase, AVAILABLE))
							}
						}
						log.Infof("[%s] - [%s]", nsName, *ns.Status.Phase)
					}(namespace)
				}

				wg.Wait()
				if allError != nil {
					log.Errorf(strings.Join(allError, "\n"))
				}
				dash.VerifyFatal(len(allError), 0, "Verifying namespaces on control plane after toggle")
			})

			Step("Applying label to all namespaces", func() {
				var wg sync.WaitGroup

				log.InfoD("Applying labels to all namespaces")

				for _, namespace := range oddNamespaces {
					wg.Add(1)
					go func(nsName string) {

						defer wg.Done()
						defer GinkgoRecover()

						_, err := utils.UpdatePDSNamespce(nsName, nsLablesApply)
						if err != nil {
							allError = append(allError, err.Error())
						}
					}(namespace)
				}

				for _, namespace := range evenNamespaces {
					wg.Add(1)
					go func(nsName string) {

						defer wg.Done()
						defer GinkgoRecover()

						_, err := utils.UpdatePDSNamespce(nsName, nsLablesApply)
						if err != nil {
							allError = append(allError, err.Error())
						}
					}(namespace)
				}

				wg.Wait()
				if allError != nil {
					log.Errorf(strings.Join(allError, "\n"))
				}
				dash.VerifyFatal(len(allError), 0, "Verifying update to all namespaces")
				log.InfoD("Sleeping for 60 seconds for the changes to be updated on control plane")
				time.Sleep(waitTime)

			})
		}

		Step("Removing labels from all even namespaces", func() {
			var wg sync.WaitGroup

			log.InfoD("Removing labels from all even namespaces")

			for _, namespace := range evenNamespaces {
				wg.Add(1)
				go func(nsName string) {

					defer wg.Done()
					defer GinkgoRecover()

					_, err := utils.UpdatePDSNamespce(nsName, nsLablesRemove)
					if err != nil {
						allError = append(allError, err.Error())
					}
				}(namespace)
			}

			wg.Wait()
			if allError != nil {
				log.Errorf(strings.Join(allError, "\n"))
			}
			dash.VerifyFatal(len(allError), 0, "Verifying update namespaces")
			log.InfoD("Sleeping for 60 seconds for the changes to be updated on control plane")
			time.Sleep(waitTime)
		})

		Step("Validating all even namespaces - Label Removed", func() {
			var wg sync.WaitGroup

			log.InfoD("Validating all current namespaces")
			for _, namespace := range evenNamespaces {
				wg.Add(1)
				go func(nsName string) {

					defer wg.Done()
					defer GinkgoRecover()

					ns, err := WorkflowNamespace.GetNamespace(nsName)
					if err != nil {
						allError = append(allError, fmt.Sprintf("Some error occurred while listing namespace. Error - [%s]", err.Error()))
					} else {
						if *ns.Status.Phase != UNAVAILABLE {
							allError = append(allError, fmt.Sprintf("[%s] is in [%s] state. Expected - [%s]", nsName, *ns.Status.Phase, UNAVAILABLE))
						}
					}
					log.Infof("[%s] - [%s]", nsName, *ns.Status.Phase)
				}(namespace)
			}

			wg.Wait()
			if allError != nil {
				log.Errorf(strings.Join(allError, "\n"))
			}
			dash.VerifyFatal(len(allError), 0, "Verifying update namespaces")
		})

		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService", func() {

				log.Debugf("Deploying DataService [%s]", ds.Name)
				log.InfoD("Deploying dataservice in [%s] namespace", evenNamespaces[0])
				_, err := WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, evenNamespaces[0])
				// TODO: Error message needs to be changed once https://purestorage.atlassian.net/browse/DS-9607 is resolved
				dash.VerifyFatal(err, "not allowed", "Verifying disable namespace usage")
			})
		}

	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})

var _ = Describe("{DeletePDSEnabledNamespace}", func() {
	var (
		deployment     *automationModels.PDSDeploymentResponse
		err            error
		nsLablesRemove map[string]string
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("DeletePDSEnabledNamespace", "Create a namespace, deploy dataservices, delete the namespace and validate", nil, 0)
		nsLablesRemove = map[string]string{
			portorxLabel: "false",
		}
	})

	It("Create a namespace, deploy dataservices, delete the namespace and validate", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
			})

			Step("Removing the label from the namespace", func() {
				log.InfoD("Removing the label from the namespace")
				_, err := utils.UpdatePDSNamespce(PDS_DEFAULT_NAMESPACE, nsLablesRemove)
				log.FailOnError(err, "Error while removing namespace label")
				log.Infof("Namespace [%s] label removed successfully", PDS_DEFAULT_NAMESPACE)
			})

			Step("Validating namespace state from control plane - Disabled Namespace", func() {
				log.InfoD("Validating namespace state from control plane - Disabled Namespace")
				log.Infof("Sleeping for 1 minute for ns changes to be updated on control plane")
				time.Sleep(1 * time.Minute)
				nsModel, err := WorkflowNamespace.GetNamespace(PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while fetching namsepace details from control plane")
				log.Infof("Namespace [%s] is in [%s] state", PDS_DEFAULT_NAMESPACE, *nsModel.Status.Phase)
				dash.VerifyFatal(strings.ToLower(string(*nsModel.Status.Phase)), strings.ToLower(UNAVAILABLE), "Verifying namespace state from control plane")
			})

			Step("Validating dataservice state from control plane", func() {
				log.InfoD("Validating dataservice state from control plane")
				currDepState, err := WorkflowDataService.GetDeployment(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while fetching deployment details from control plane")
				log.InfoD("Deployment [%s] is in [%s] state", *deployment.Create.Meta.Uid, *currDepState.Status.Health)
				//TODO: DataService state checks needs to be added here
			})

			Step("Running Workloads after disabling the namespace", func() {
				log.InfoD("Running Workloads after disabling the namespace")
				_, err := WorkflowDataService.RunDataServiceWorkloads(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while running workloads on ds")
			})

			Step("Delete Namespace", func() {
				log.InfoD("Deleting namespace")
				err := core.Instance().DeleteNamespace(PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deleting namespace")
				log.Infof("Namespace [%s] deletion started successfully", PDS_DEFAULT_NAMESPACE)
			})

			Step("Printing current phase of namespace", func() {
				log.InfoD("Printing current state of namespace")
				ns, err := core.Instance().GetNamespace(PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deleting namespace")
				log.Infof("Namespace [%s] is in [%s] state", PDS_DEFAULT_NAMESPACE, ns.Status.Phase)
			})

			Step("Validating namespace state from control plane", func() {
				log.InfoD("Validating namespace state from control plane")
				log.Infof("Sleeping for 1 minute for ns changes to be updated on control plane")
				time.Sleep(1 * time.Minute)
				nsModel, err := WorkflowNamespace.GetNamespace(PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while fetching namsepace details from control plane")
				log.Infof("Namespace [%s] is in [%s] state", PDS_DEFAULT_NAMESPACE, *nsModel.Status.Phase)
				dash.VerifyFatal(strings.ToLower(string(*nsModel.Status.Phase)), strings.ToLower(TOMBSTONED), "Verifying namespace state from control plane")
			})

			Step("Validating dataservice state from control plane", func() {
				log.InfoD("Validating dataservice state from control plane")
				currDepState, err := WorkflowDataService.GetDeployment(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while fetching deployment details from control plane")
				log.InfoD("Deployment [%s] is in [%s] state", *deployment.Create.Meta.Uid, *currDepState.Status.Phase)
				dash.VerifyFatal(strings.ToLower(string(*currDepState.Status.Phase)), strings.ToLower("DELETED_FROM_CLUSTER"), "Verifying namespace state from control plane")
			})

		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
		WorkflowDataService.Purge(true)
	})
})
