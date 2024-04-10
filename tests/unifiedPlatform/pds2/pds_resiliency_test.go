package tests

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
	"strings"
)

var _ = Describe("{StopPXDuringStorageResize}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("StopPXDuringStorageResize", "Deploy data services, Run workloads, and Stop PX on the node while Storage resize is happening", nil, 0)
	})
	var (
		workflowResiliency  pds.WorkflowPDSResiliency
		workflowDataservice pds.WorkflowDataService
		workFlowTemplates   pds.WorkflowPDSTemplates
		//deployment          *automationModels.PDSDeploymentResponse
	)
	workflowResiliency.WfDataService = &workflowDataservice
	It("Deploy and Validate DataService", func() {
		Step("Create a PDS Namespace", func() {
			Namespace = strings.ToLower("pds-test-ns-" + utilities.RandString(5))
			WorkflowNamespace.TargetCluster = WorkflowTargetCluster
			workFlowTemplates.Platform = WorkflowPlatform
			WorkflowNamespace.Namespaces = make(map[string]string)
			workflowNamespace, err := WorkflowNamespace.CreateNamespaces(Namespace)
			log.FailOnError(err, "Unable to create namespace")
			log.Infof("Namespaces created - [%s]", workflowNamespace.Namespaces)
			log.Infof("Namespace id - [%s]", workflowNamespace.Namespaces[Namespace])
		})

		serviceConfigId, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams, false)
		log.FailOnError(err, "Unable to create Custom Templates for PDS")
		workflowDataservice.PDSTemplates.ServiceConfigTemplateId = serviceConfigId
		workflowDataservice.PDSTemplates.StorageTemplateId = stConfigId
		workflowDataservice.PDSTemplates.ResourceTemplateId = resConfigId

		log.InfoD("Original Storage Template ID- [resTempId- %v]", stConfigId)

		for _, ds := range NewPdsParams.DataServiceToTest {
			workflowDataservice.Namespace = WorkflowNamespace
			workflowDataservice.NamespaceName = Namespace
			_, err := workflowDataservice.DeployDataService(ds, ds.OldImage, ds.OldVersion)
			log.FailOnError(err, "Error while deploying ds")
		}

		stepLog := "Running Workloads before Storage Resize"
		Step(stepLog, func() {
			err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
			log.FailOnError(err, "Error while running workloads on ds")
		})

		//Update Ds With New Values of Resource Templates
		resourceConfigUpdated, err := workFlowTemplates.IncreaseStorageAndFetchIds(NewPdsParams)
		log.FailOnError(err, "Unable to create Custom Templates for PDS")

		log.InfoD("Updated Storage Template ID- [updated- %v]", resourceConfigUpdated)
		workflowDataservice.PDSTemplates.ResourceTemplateId = resourceConfigUpdated
		// Run bot Storage Resize and Stop PX concurrently
		for _, ds := range NewPdsParams.DataServiceToTest {
			err := workflowResiliency.InduceFailureAndExecuteResiliencyScenario(ds, "StopPXDuringStorageResize")
			log.FailOnError(err, "Error while updating ds")
		}
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
