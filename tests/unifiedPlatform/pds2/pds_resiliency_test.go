package tests

import (
	"strings"

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

var _ = Describe("{ValidatePdsHealthIncaseofFailures}", func() {
	var (
		workflowDataservice pds.WorkflowDataService
		workFlowTemplates   pds.WorkflowPDSTemplates
		workflowResiliency  pds.WorkflowPDSResiliency
		deployment          *automationModels.PDSDeploymentResponse
		templates           []string
		dsNameAndAppTempId  map[string]string
		stConfigId          string
		resConfigId         string
		err                 error
	)

	JustBeforeEach(func() {
		StartTorpedoTest("DeployDataServicesOnDemandAndScaleUp", "Deploy data services and perform scale up", nil, 0)
		workFlowTemplates.Platform = WorkflowPlatform
		workflowDataservice.Namespace = WorkflowNamespace
		workflowDataservice.NamespaceName = PDS_DEFAULT_NAMESPACE
		workflowDataservice.Dash = dash
	})

	It("Deploy and Validate DataService", func() {
		Step("Create Service Configuration, Resource and Storage Templates", func() {
			//dsNameAndAppTempId = workFlowTemplates.CreateAppTemplate(NewPdsParams)
			dsNameAndAppTempId, stConfigId, resConfigId, err = workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams)
			log.FailOnError(err, "Unable to create Custom Templates for PDS")
			workflowDataservice.PDSTemplates.StorageTemplateId = stConfigId
			workflowDataservice.PDSTemplates.ResourceTemplateId = resConfigId
		})

		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService", func() {
				workflowDataservice.PDSTemplates.ServiceConfigTemplateId = dsNameAndAppTempId[ds.Name]
				templates = append(templates, dsNameAndAppTempId[ds.Name], stConfigId, resConfigId)

				log.Debugf("Deploying DataService [%s]", ds.Name)
				deployment, err = workflowDataservice.DeployDataService(ds, ds.Image, ds.Version)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
			})

			//stepLog := "Running Workloads before taking backups"
			//Step(stepLog, func() {
			//	err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
			//	log.FailOnError(err, "Error while running workloads on ds")
			//})

			Step("Delete PdsDeploymentPods and check the deployment health", func() {
				workflowResiliency.WfDataService = &workflowDataservice
				workflowResiliency.ResiliencyFlag = true
				workflowResiliency.InduceFailureAndExecuteResiliencyScenario(ds, deployment, DeletePdsDeploymentPodAndValidateDeploymentHealth)
			})

			//stepLog = "Running Workloads after ScaleUp of DataService"
			//Step(stepLog, func() {
			//	err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
			//	log.FailOnError(err, "Error while running workloads on ds")
			//})
			Step("Delete DataServiceDeployment", func() {
				log.InfoD("Cleaning Up dataservice...")
				err := workflowDataservice.DeleteDeployment(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while deleting dataservice")
			})
		}
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		Step("Delete PDS CustomTemplates", func() {
			log.InfoD("Cleaning Up templates...")
			err := workFlowTemplates.DeleteCreatedCustomPdsTemplates(templates)
			log.FailOnError(err, "Error while deleting dataservice")
		})
	})
})

var _ = Describe("{StopPXDuringStorageResize}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("StopPXDuringStorageResize", "Deploy data services, Run workloads, and Stop PX on the node while Storage resize is happening", nil, 0)
	})
	var (
		workflowResiliency  pds.WorkflowPDSResiliency
		workflowDataservice pds.WorkflowDataService
		workFlowTemplates   pds.WorkflowPDSTemplates
		deployment          *automationModels.PDSDeploymentResponse
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

		for _, ds := range NewPdsParams.DataServiceToTest {
			workflowDataservice.Namespace = WorkflowNamespace
			workflowDataservice.NamespaceName = Namespace

			serviceConfigId, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams)
			log.FailOnError(err, "Unable to create Custom Templates for PDS")
			workflowDataservice.PDSTemplates.ServiceConfigTemplateId = serviceConfigId[ds.Name]
			workflowDataservice.PDSTemplates.StorageTemplateId = stConfigId
			workflowDataservice.PDSTemplates.ResourceTemplateId = resConfigId

			deployment, err = workflowDataservice.DeployDataService(ds, ds.OldImage, ds.OldVersion)
			log.FailOnError(err, "Error while deploying ds")
		}

		stepLog := "Running Workloads before Storage Resize"
		Step(stepLog, func() {
			err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
			log.FailOnError(err, "Error while running workloads on ds")
		})

		//Update Ds With New Values of Resource Templates
		resourceConfigUpdated, err := workFlowTemplates.CreateResourceTemplateWithCustomValue(NewPdsParams, *deployment.Create.Meta.Name, 1)
		log.FailOnError(err, "Unable to create Custom Templates for PDS")

		log.InfoD("Updated Storage Template ID- [updated- %v]", resourceConfigUpdated)
		workflowDataservice.PDSTemplates.ResourceTemplateId = resourceConfigUpdated
		// Run bot Storage Resize and Stop PX concurrently
		for _, ds := range NewPdsParams.DataServiceToTest {
			err := workflowResiliency.InduceFailureAndExecuteResiliencyScenario(ds, deployment, "StopPXDuringStorageResize")
			log.FailOnError(err, "Error while updating ds")
		}
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
