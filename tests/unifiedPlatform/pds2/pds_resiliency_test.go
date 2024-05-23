package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	pdsResLib "github.com/portworx/torpedo/drivers/unifiedPlatform/resiliency"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
	"strings"
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
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("KillAgentDuringDeployment", "Kill Px Agent Pod when a DS Deployment is happening", nil, 0)
		WorkflowDataService.SkipValidatation[pds.ValidatePdsDeployment] = true
		WorkflowDataService.SkipValidatation[pds.ValidatePdsWorkloads] = true
	})

	It("Kill Px Agent Pod when a DS Deployment is happening", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {

			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Deployment failed")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)

			})

			Step("Delete PDSPods while deployment", func() {
				log.InfoD("Delete PDSPods while deployment")
				// Global Resiliency TC marker
				pdsResLib.MarkResiliencyTC(true)
				// Type of failure that this TC needs to cover
				failuretype := pdsResLib.TypeOfFailure{
					Type: pdsResLib.KillAgentPodDuringDeployment,
					Method: func() error {
						return WorkflowDataService.DeletePDSPods([]string{"pds-backups", "pds-target"}, PlatformNamespace)
					},
				}

				pdsResLib.DefineFailureType(failuretype)
				err = pdsResLib.InduceFailureAfterWaitingForCondition(&deployment.Create, PDS_DEFAULT_NAMESPACE, 1)
				log.FailOnError(err, fmt.Sprintf("Error happened while executing Kill Agent Pod test for data service %v", *deployment.Create.Status.CustomResourceName))
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
		defer func() {
			delete(WorkflowDataService.SkipValidatation, pds.ValidatePdsDeployment)
			delete(WorkflowDataService.SkipValidatation, pds.ValidatePdsWorkloads)
		}()
	})
})
