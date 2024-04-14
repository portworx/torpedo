package tests

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
	"strings"
)

var _ = Describe("{DeployDataServicesOnDemandAndScaleUp}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("DeployDataServicesOnDemandAndScaleUp", "Deploy data services and perform scale up", nil, 0)
	})
	var (
		workflowDataservice pds.WorkflowDataService
		workFlowTemplates   pds.WorkflowPDSTemplates
		deployment          *automationModels.PDSDeploymentResponse
		updateDeployment    *automationModels.PDSDeploymentResponse
		templates           []string
		err                 error
	)

	It("Deploy,Validate and ScaleUp DataService", func() {
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

			serviceConfigId, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams, ds.Name)
			log.FailOnError(err, "Unable to create Custom Templates for PDS")

			workflowDataservice.PDSTemplates.ServiceConfigTemplateId = serviceConfigId
			workflowDataservice.PDSTemplates.StorageTemplateId = stConfigId
			workflowDataservice.PDSTemplates.ResourceTemplateId = resConfigId

			deployment, err = workflowDataservice.DeployDataService(ds, ds.Image, ds.Version)
			log.FailOnError(err, "Error while deploying ds")
			log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
		}

		defer func() {
			Step("Delete DataServiceDeployment", func() {
				log.InfoD("Cleaning Up dataservice...")
				err := workflowDataservice.DeleteDeployment()
				log.FailOnError(err, "Error while deleting dataservice")
			})
		}()

		defer func() {
			Step("Delete PDS CustomTemplates", func() {
				log.InfoD("Cleaning Up templates...")

				err := workFlowTemplates.DeleteCreatedCustomPdsTemplates(templates)
				log.FailOnError(err, "Error while deleting dataservice")
			})
		}()

		//stepLog := "Running Workloads before taking backups"
		//Step(stepLog, func() {
		//	err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
		//	log.FailOnError(err, "Error while running workloads on ds")
		//})

		Step("ScaleUp DataService", func() {
			log.InfoD("Scaling Up dataservices...")
			for _, ds := range NewPdsParams.DataServiceToTest {
				updateDeployment, err = workflowDataservice.UpdateDataService(ds, *deployment.Create.Meta.Uid, ds.Image, ds.Version)
				log.FailOnError(err, "Error while updating ds")
				log.Debugf("Updated Deployment Id: [%s]", *updateDeployment.Update.Meta.Uid)
			}
		})

		//stepLog = "Running Workloads after ScaleUp of DataService"
		//Step(stepLog, func() {
		//	err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
		//	log.FailOnError(err, "Error while running workloads on ds")
		//})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{UpgradeDataServiceImageAndVersion}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("UpgradeDataServiceImage", "Upgrade Data Service Version and Image", nil, 0)
	})
	var (
		workflowDataservice pds.WorkflowDataService
		workFlowTemplates   pds.WorkflowPDSTemplates
		deployment          *automationModels.PDSDeploymentResponse
	)

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

			serviceConfigId, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams, ds.Name)
			log.FailOnError(err, "Unable to create Custom Templates for PDS")
			workflowDataservice.PDSTemplates.ServiceConfigTemplateId = serviceConfigId
			workflowDataservice.PDSTemplates.StorageTemplateId = stConfigId
			workflowDataservice.PDSTemplates.ResourceTemplateId = resConfigId

			deployment, err = workflowDataservice.DeployDataService(ds, ds.OldImage, ds.OldVersion)
			log.FailOnError(err, "Error while deploying ds")
		}

		stepLog := "Running Workloads before upgrading the ds image"
		Step(stepLog, func() {
			err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
			log.FailOnError(err, "Error while running workloads on ds")
		})
	})

	//TODO: Add backup and restore workflows once we have the workflows ready
	//TODO: Take backup of the old deployment

	It("Upgrade DataService Version and Image", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			_, err := workflowDataservice.UpdateDataService(ds, *deployment.Create.Meta.Uid, ds.Image, ds.Version)
			log.FailOnError(err, "Error while updating ds")
		}

		stepLog := "Running Workloads after upgrading the ds image"
		Step(stepLog, func() {
			err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
			log.FailOnError(err, "Error while running workloads on ds")
		})
	})

	//TODO: Restore the old deployment
	//TODO: Upgrade the restored deployment image to latest

	It("Delete DataServiceDeployment", func() {
		err := workflowDataservice.DeleteDeployment()
		log.FailOnError(err, "Error while deleting data Service")
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{ScaleUpCpuMemLimitsOfDS}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("ScaleUpCpuMemLimitsOfDS", "Deploy a dataservice and scale up its CPU/MEM limits by editing the respective template", nil, 0)
	})
	var (
		workflowDataservice pds.WorkflowDataService
		workFlowTemplates   pds.WorkflowPDSTemplates
		deployment          *automationModels.PDSDeploymentResponse
		err                 error
	)
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

			serviceConfigId, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams, ds.Name)
			log.FailOnError(err, "Unable to create Custom Templates for PDS")
			workflowDataservice.PDSTemplates.ServiceConfigTemplateId = serviceConfigId
			workflowDataservice.PDSTemplates.StorageTemplateId = stConfigId
			workflowDataservice.PDSTemplates.ResourceTemplateId = resConfigId
			log.InfoD("Original Resource Template ID- [resTempId- %v]", resConfigId)

			serviceConfigId, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams, ds.Name)
			log.FailOnError(err, "Unable to create Custom Templates for PDS")

			deployment, err = workflowDataservice.DeployDataService(ds, ds.OldImage, ds.OldVersion)
			log.FailOnError(err, "Error while deploying ds")
		}

		//Update Ds With New Values of Resource Templates
		resConfigIdUpdated, err := workFlowTemplates.CreateResourceTemplateWithCustomValue(NewPdsParams, *deployment.Create.Meta.Name, 1)
		log.FailOnError(err, "Unable to create Custom Templates for PDS")

		log.InfoD("Updated Resource Template ID- [updated- %v]", resConfigIdUpdated)
		workflowDataservice.PDSTemplates.ResourceTemplateId = resConfigIdUpdated
		for _, ds := range NewPdsParams.DataServiceToTest {
			_, err := workflowDataservice.UpdateDataService(ds, *deployment.Create.Meta.Uid, ds.OldImage, ds.OldVersion)
			log.FailOnError(err, "Error while updating ds")
		}
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{IncreasePVCby1gb}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("IncreasePVCby1gb", "Deploy a dataservice and increase it Storage Size by 1gb  by applying new Storage template", nil, 0)
	})
	var (
		workflowDataservice pds.WorkflowDataService
		workFlowTemplates   pds.WorkflowPDSTemplates
		deployment          *automationModels.PDSDeploymentResponse
	)
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

			serviceConfigId, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams, ds.Name)
			log.FailOnError(err, "Unable to create Custom Templates for PDS")
			workflowDataservice.PDSTemplates.ServiceConfigTemplateId = serviceConfigId
			workflowDataservice.PDSTemplates.StorageTemplateId = stConfigId
			workflowDataservice.PDSTemplates.ResourceTemplateId = resConfigId

			log.InfoD("Original Storage Template ID- [resTempId- %v]", stConfigId)
			deployment, err = workflowDataservice.DeployDataService(ds, ds.OldImage, ds.OldVersion)
			log.FailOnError(err, "Error while deploying ds")
		}

		//Update Ds With New Values of Resource Templates
		//Update Ds With New Values of Resource Templates
		resConfigIdUpdated, err := workFlowTemplates.CreateResourceTemplateWithCustomValue(NewPdsParams, *deployment.Create.Meta.Name, 1)
		log.FailOnError(err, "Unable to create Custom Templates for PDS")

		log.InfoD("Updated Resource Template ID- [updated- %v]", resConfigIdUpdated)
		workflowDataservice.PDSTemplates.ResourceTemplateId = resConfigIdUpdated
		for _, ds := range NewPdsParams.DataServiceToTest {
			_, err := workflowDataservice.UpdateDataService(ds, *deployment.Create.Meta.Uid, ds.OldImage, ds.OldVersion)
			log.FailOnError(err, "Error while updating ds")
		}
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
