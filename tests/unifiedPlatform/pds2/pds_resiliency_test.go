package tests

import (
	"strings"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
)

const (
	StopPXDuringStorageResize                         = "stop-px-during-storage-resize"
	DeletePdsDeploymentPodAndValidateDeploymentHealth = "delete-pdsDeploymentPods-validate-deployment-health"
)

var _ = Describe("{StopPXDuringStorageResize}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("StopPXDuringStorageResize", "Deploy data services, Run workloads, and Stop PX on the node while Storage resize is happening", nil, 0)
	})
	var (
		workflowResiliency  pds.WorkflowPDSResiliency
		workflowDataservice pds.WorkflowDataService
		workFlowTemplates   pds.WorkflowPDSTemplates
		deployment          *automationModels.PDSDeploymentResponse
		err                 error
	)
	workflowResiliency.WfDataService = &workflowDataservice
	It("Deploy and Validate DataService", func() {
		Step("Create a PDS Namespace", func() {
			Namespace = strings.ToLower("pds-test-ns-" + utilities.RandString(5))
			WorkflowNamespace.TargetCluster = &WorkflowTargetCluster
			workFlowTemplates.Platform = WorkflowPlatform
			WorkflowNamespace.Namespaces = make(map[string]string)
			workflowNamespace, err := WorkflowNamespace.CreateNamespaces(Namespace)
			log.FailOnError(err, "Unable to create namespace")
			log.Infof("Namespaces created - [%s]", workflowNamespace.Namespaces)
			log.Infof("Namespace id - [%s]", workflowNamespace.Namespaces[Namespace])
		})

		for _, ds := range NewPdsParams.DataServiceToTest {
			workflowDataservice.Namespace = &WorkflowNamespace
			deployment, err = workflowDataservice.DeployDataService(ds, ds.OldImage, ds.OldVersion, PDS_DEFAULT_NAMESPACE)
			log.FailOnError(err, "Error while deploying ds")

			//stepLog := "Running Workloads before Storage Resize"
			//Step(stepLog, func() {
			//	err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams, ds.Name)
			//	log.FailOnError(err, "Error while running workloads on ds")
			//})

			//Update Ds With New Values of Resource Templates
			resourceConfigUpdated, err := workFlowTemplates.CreateResourceTemplateWithCustomValue(NewPdsParams)
			log.FailOnError(err, "Unable to create Custom Templates for PDS")

			log.InfoD("Updated Storage Template ID- [updated- %v]", resourceConfigUpdated)
			workflowDataservice.PDSTemplates.ResourceTemplateId = resourceConfigUpdated
			// Run bot Storage Resize and Stop PX concurrently
			err = workflowResiliency.InduceFailureAndExecuteResiliencyScenario(ds, deployment, "StopPXDuringStorageResize")
			log.FailOnError(err, "Error while updating ds")

		}
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{KillAgentDuringDeployment}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
		allError   []error
		wg         sync.WaitGroup
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("KillAgentDuringDeployment", "Kill Px Agent Pod when a DS Deployment is happening", nil, 0)
	})

	It("Kill Px Agent Pod when a DS Deployment is happening", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {

			Step("Deploy DataService", func() {
				wg.Add(1)
				go func() {

					defer wg.Done()
					defer GinkgoRecover()

					deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
					if err != nil {
						log.Errorf("Error while deploying ds: [%s]", err.Error())
						allError = append(allError, err)
					}
					log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
				}()
			})

			Step("Delete PDSPods", func() {

				wg.Add(1)

				go func() {

					defer wg.Done()
					defer GinkgoRecover()

					err := WorkflowDataService.DeletePDSPods([]string{"px-agent"}, PlatformNamespace)
					if err != nil {
						log.Errorf("Error while deleting px-agent pods: [%s]", err.Error())
						allError = append(allError, err)
					}
				}()

				wg.Wait()
				dash.VerifyFatal(len(allError), 0, "Error while deploying ds or rebooting agent")
			})

			Step("Validate Data Service to after px-agent reboot", func() {
				log.InfoD("Validate Data Service to after px-agent reboot")
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.Replicas, WorkflowDataService.PDSTemplates.ResourceTemplateId, WorkflowDataService.PDSTemplates.StorageTemplateId, PDS_DEFAULT_NAMESPACE, ds.Version, ds.Image)
				log.FailOnError(err, "Error while Validating dataservice after px-agent reboot")
			})

			stepLog := "Running Workloads after px-agent reboot"
			Step(stepLog, func() {
				_, err := WorkflowDataService.RunDataServiceWorkloads(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while running workloads on ds")
			})
		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})
