package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	utils "github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
	"strconv"
	"strings"
	"sync"
	"time"
)

const AVAILABLE = "AVAILABLE"
const UNAVAILABLE = "UNAVAILABLE"
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
		// TODO: This needs to be uncommented once https://purestorage.atlassian.net/browse/DS-9607 is resolved
		//dsNameAndAppTempId            map[string]string
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
		waitTime = 5 * time.Minute
	})

	It("Enables and Disables pds on a namespace multiple times", func() {

		// TODO: This needs to be uncommented once https://purestorage.atlassian.net/browse/DS-9607 is resolved
		//Step("Create Service Configuration, Resource and Storage Templates", func() {
		//	//dsNameAndAppTempId = workFlowTemplates.CreateAppTemplate(NewPdsParams)
		//	dsNameAndAppTempId, _, _, err := WorkflowPDSTemplate.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams)
		//	log.FailOnError(err, "Unable to create Custom Templates for PDS")
		//})

		Step(fmt.Sprintf("Creating [%d] namespaces with labels", numberOfNamespacesTobeCreated), func() {
			var wg sync.WaitGroup

			log.InfoD("Creating [%d] namespaces with PDS labels present", numberOfNamespacesTobeCreated)

			for i := 0; i < numberOfNamespacesTobeCreated; i++ {
				wg.Add(1)

				nsName := namespacePrefix + RandomString(5) + "-" + strconv.Itoa(i)

				go func() {

					defer wg.Done()
					defer GinkgoRecover()

					_, err := WorkflowNamespace.CreateNamespaces(nsName)
					if err != nil {
						allError = append(allError, err.Error())
					}

				}()

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

		for i := 0; i < TotalToggles; i++ {
			Step("Removing labels from all odd namespaces", func() {
				var wg sync.WaitGroup
				log.InfoD("Executing [%d] toggle", (i + 1))

				log.InfoD("Removing labels from all odd namespaces")

				for _, namespace := range oddNamespaces {
					wg.Add(1)
					go func() {

						defer wg.Done()
						defer GinkgoRecover()

						_, err := utils.UpdatePDSNamespce(namespace, nsLablesRemove)
						if err != nil {
							allError = append(allError, err.Error())
						}
					}()
				}

				wg.Wait()
				if allError != nil {
					log.Errorf(strings.Join(allError, "\n"))
				}
				dash.VerifyFatal(len(allError), 0, "Verifying update namespaces - Remove label from odd namespaces")
				log.InfoD("Sleeping for 5 minutes for the changes to be updated on control plane")
				time.Sleep(waitTime)
			})

			Step("Validating all current namespaces", func() {
				var wg sync.WaitGroup

				log.InfoD("Validating all current namespaces")
				for _, namespace := range oddNamespaces {
					wg.Add(1)
					go func() {

						defer wg.Done()
						defer GinkgoRecover()

						ns, err := WorkflowNamespace.GetNamespace(namespace)
						if err != nil {
							allError = append(allError, fmt.Sprintf("Some error occurred while listing namespace. Error - [%s]", err.Error()))
						} else {
							if *ns.Status.Phase != UNAVAILABLE {
								allError = append(allError, fmt.Sprintf("[%s] is in [%s] state. Expected - [%s]", namespace, *ns.Status.Phase, UNAVAILABLE))
							}
						}
						log.Infof("[%s] - [%s]", namespace, *ns.Status.Phase)
					}()
				}

				for _, namespace := range evenNamespaces {
					wg.Add(1)
					go func() {

						defer wg.Done()
						defer GinkgoRecover()

						ns, err := WorkflowNamespace.GetNamespace(namespace)
						if err != nil {
							allError = append(allError, fmt.Sprintf("Some error occurred while listing namespace. Error - [%s]", err.Error()))
						} else {
							if *ns.Status.Phase != AVAILABLE {
								allError = append(allError, fmt.Sprintf("[%s] is in [%s] state. Expected - [%s]", namespace, *ns.Status.Phase, AVAILABLE))
							}
						}
						log.Infof("[%s] - [%s]", namespace, *ns.Status.Phase)
					}()
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
					go func() {

						defer wg.Done()
						defer GinkgoRecover()

						_, err := utils.UpdatePDSNamespce(namespace, nsLablesApply)
						if err != nil {
							allError = append(allError, err.Error())
						}
					}()
				}

				log.InfoD("Removing labels from even namespaces")
				for _, namespace := range evenNamespaces {
					wg.Add(1)
					go func() {

						defer wg.Done()
						defer GinkgoRecover()

						_, err := utils.UpdatePDSNamespce(namespace, nsLablesRemove)
						if err != nil {
							allError = append(allError, err.Error())
						}
					}()
				}

				wg.Wait()
				if allError != nil {
					log.Errorf(strings.Join(allError, "\n"))
				}
				dash.VerifyFatal(len(allError), 0, "Verifying namespace toggle")
				log.InfoD("Sleeping for 5 minutes for the changes to be updated on control plane")
				time.Sleep(waitTime)
			})

			Step("Validating all current namespaces", func() {
				var wg sync.WaitGroup

				log.InfoD("Validating all current namespaces")

				for _, namespace := range evenNamespaces {
					wg.Add(1)
					go func() {

						defer wg.Done()
						defer GinkgoRecover()

						ns, err := WorkflowNamespace.GetNamespace(namespace)
						if err != nil {
							allError = append(allError, fmt.Sprintf("Some error occurred while listing namespace. Error - [%s]", err.Error()))
						} else {
							if *ns.Status.Phase != UNAVAILABLE {
								allError = append(allError, fmt.Sprintf("[%s] is in [%s] state. Expected - [%s]", namespace, *ns.Status.Phase, UNAVAILABLE))
							}
						}
						log.Infof("[%s] - [%s]", namespace, *ns.Status.Phase)
					}()
				}

				for _, namespace := range oddNamespaces {
					wg.Add(1)
					go func() {

						defer wg.Done()
						defer GinkgoRecover()

						ns, err := WorkflowNamespace.GetNamespace(namespace)
						if err != nil {
							allError = append(allError, fmt.Sprintf("Some error occurred while listing namespace. Error - [%s]", err.Error()))
						} else {
							if *ns.Status.Phase != AVAILABLE {
								allError = append(allError, fmt.Sprintf("[%s] is in [%s] state. Expected - [%s]", namespace, *ns.Status.Phase, AVAILABLE))
							}
						}
						log.Infof("[%s] - [%s]", namespace, *ns.Status.Phase)
					}()
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
					go func() {

						defer wg.Done()
						defer GinkgoRecover()

						_, err := utils.UpdatePDSNamespce(namespace, nsLablesApply)
						if err != nil {
							allError = append(allError, err.Error())
						}
					}()
				}

				for _, namespace := range evenNamespaces {
					wg.Add(1)
					go func() {

						defer wg.Done()
						defer GinkgoRecover()

						_, err := utils.UpdatePDSNamespce(namespace, nsLablesApply)
						if err != nil {
							allError = append(allError, err.Error())
						}
					}()
				}

				wg.Wait()
				if allError != nil {
					log.Errorf(strings.Join(allError, "\n"))
				}
				dash.VerifyFatal(len(allError), 0, "Verifying update to all namespaces")
				log.InfoD("Sleeping for 5 minutes for the changes to be updated on control plane")
				time.Sleep(waitTime)

			})
		}

		Step("Removing labels from all even namespaces", func() {
			var wg sync.WaitGroup

			log.InfoD("Removing labels from all even namespaces")

			for _, namespace := range evenNamespaces {
				wg.Add(1)
				go func() {

					defer wg.Done()
					defer GinkgoRecover()

					_, err := utils.UpdatePDSNamespce(namespace, nsLablesRemove)
					if err != nil {
						allError = append(allError, err.Error())
					}
				}()
			}

			wg.Wait()
			if allError != nil {
				log.Errorf(strings.Join(allError, "\n"))
			}
			dash.VerifyFatal(len(allError), 0, "Verifying update namespaces")
			log.InfoD("Sleeping for 5 minutes for the changes to be updated on control plane")
			time.Sleep(waitTime)
		})

		Step("Validating all even namespaces - Label Removed", func() {
			var wg sync.WaitGroup

			log.InfoD("Validating all current namespaces")
			for _, namespace := range evenNamespaces {
				wg.Add(1)
				go func() {

					defer wg.Done()
					defer GinkgoRecover()

					ns, err := WorkflowNamespace.GetNamespace(namespace)
					if err != nil {
						allError = append(allError, fmt.Sprintf("Some error occurred while listing namespace. Error - [%s]", err.Error()))
					} else {
						if *ns.Status.Phase != UNAVAILABLE {
							allError = append(allError, fmt.Sprintf("[%s] is in [%s] state. Expected - [%s]", namespace, *ns.Status.Phase, UNAVAILABLE))
						}
					}
					log.Infof("[%s] - [%s]", namespace, *ns.Status.Phase)
				}()
			}

			wg.Wait()
			if allError != nil {
				log.Errorf(strings.Join(allError, "\n"))
			}
			dash.VerifyFatal(len(allError), 0, "Verifying update namespaces")
		})

		// TODO: This needs to be uncommented once https://purestorage.atlassian.net/browse/DS-9607 is resolved
		//for _, ds := range NewPdsParams.DataServiceToTest {
		//	Step("Deploy DataService", func() {
		//
		//		WorkflowDataService.PDSTemplates = WorkflowPDSTemplate
		//		WorkflowDataService.PDSTemplates.ServiceConfigTemplateId = dsNameAndAppTempId[ds.Name]
		//
		//		log.Debugf("Deploying DataService [%s]", ds.Name)
		//		_, err := WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version)
		// TODO: <FAILURE NEEDS TO BE ADDED HERE> needs to be replaced with actual failure
		//		dash.VerifyFatal(err, "<FAILURE NEEDS TO BE ADDED HERE>", "Verifying disable namespace usage")
		//	})
		//}

	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})
