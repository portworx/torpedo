package tests

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
	"strings"
	"sync"
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

var _ = Describe("{RebootAllWorkerNodesDuringDeployment}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		wg         sync.WaitGroup
		err        error
		allError   []error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("RebootAllWorkerNodesDuringDeployment", "Reboots all worker nodes while a data service pod is coming up", nil, 0)
	})

	It("Reboots all worker nodes while a data service pod is coming up", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService", func() {
				log.InfoD("Deploying DataService")
				wg.Add(1)
				go func() {

					defer wg.Done()
					defer GinkgoRecover()

					deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
					if err != nil {
						log.Errorf("Error while deploying dataservice: [%s]", err.Error())
						allError = append(allError, err)
					}
					log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
				}()
			})

			Step("Reboot all worker nodes", func() {
				log.InfoD("Rebooting all worker nodes")
				wg.Add(1)
				go func() {

					defer wg.Done()
					defer GinkgoRecover()

					nodesToReboot := node.GetWorkerNodes()
					log.Infof("Rebooting all worker nodes: %v", len(nodesToReboot))
					err := RebootNodes(nodesToReboot)
					if err != nil {
						log.Errorf("Error while getting worker nodes: [%s]", err.Error())
						allError = append(allError, err)
					}
				}()
				log.Infof("Waiting for node reboot and deployment to complete")
				wg.Wait()
				dash.VerifyFatal(len(allError), 0, "Error while deploying dataservice or rebooting nodes")

			})

			Step("Validate Data Service after node reboot", func() {
				log.InfoD("Validate Data Service after node reboot")
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.Replicas, WorkflowDataService.PDSTemplates.ResourceTemplateId, WorkflowDataService.PDSTemplates.StorageTemplateId, PDS_DEFAULT_NAMESPACE, ds.Version, ds.Image)
				log.FailOnError(err, "Error while Validating dataservice after node reboot node")
			})

			Step("Running Workloads after node reboot", func() {
				log.InfoD("Running Workloads after node reboot")
				_, err := WorkflowDataService.RunDataServiceWorkloads(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while running workloads on ds")
			})

		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})
