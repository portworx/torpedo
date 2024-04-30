package tests

import (
	dslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
)

var _ = Describe("{DeployDataServicesOnDemandAndScaleUp}", func() {
	var (
		workflowDataservice pds.WorkflowDataService
		workFlowTemplates   pds.WorkflowPDSTemplates
		deployment          *automationModels.PDSDeploymentResponse
		dsNameAndAppTempId  map[string]string
		stConfigId          string
		resConfigId         string
		templates           []string
		err                 error
	)

	JustBeforeEach(func() {
		StartTorpedoTest("DeployDataServicesOnDemandAndScaleUp", "Deploy data services and perform scale up", nil, 0)
		workFlowTemplates.Platform = WorkflowPlatform
		workflowDataservice.Namespace = &WorkflowNamespace
		workflowDataservice.NamespaceName = PDS_DEFAULT_NAMESPACE
		workflowDataservice.Dash = dash
	})

	It("Deploy,Validate and ScaleUp DataService", func() {
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

			Step("ScaleUp DataService", func() {
				log.InfoD("Scaling Up dataServices...")
				updateDeployment, err := workflowDataservice.UpdateDataService(ds, *deployment.Create.Meta.Uid, ds.Image, ds.Version)
				log.FailOnError(err, "Error while updating ds")
				log.Debugf("Updated Deployment Id: [%s]", *updateDeployment.Update.Meta.Uid)
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

var _ = Describe("{UpgradeDataServiceImage}", func() {
	var (
		workflowDataservice pds.WorkflowDataService
		workFlowTemplates   pds.WorkflowPDSTemplates
		deployment          *automationModels.PDSDeploymentResponse
		templates           []string
		dsNameAndAppTempId  map[string]string
		stConfigId          string
		resConfigId         string
		err                 error
	)

	JustBeforeEach(func() {
		StartTorpedoTest("UpgradeDataServiceImage", "Upgrade Data Service Image", nil, 0)
		workFlowTemplates.Platform = WorkflowPlatform
		workflowDataservice.Namespace = &WorkflowNamespace
		workflowDataservice.NamespaceName = PDS_DEFAULT_NAMESPACE
		workflowDataservice.Dash = dash
	})

	It("Deploy, Validate and Upgrade Data service Image", func() {

		Step("Create Service Configuration, Resource and Storage Templates", func() {
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
				deployment, err = workflowDataservice.DeployDataService(ds, ds.OldImage, ds.Version)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
			})

			//stepLog := "Running Workloads before upgrading the ds image"
			//Step(stepLog, func() {
			//	err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
			//	log.FailOnError(err, "Error while running workloads on ds")
			//})

			Step("Upgrade DataService Image", func() {
				_, err := workflowDataservice.UpdateDataService(ds, *deployment.Create.Meta.Uid, ds.Image, ds.Version)
				log.FailOnError(err, "Error while updating ds")

				//stepLog := "Running Workloads after upgrading the ds image"
				//Step(stepLog, func() {
				//	err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
				//	log.FailOnError(err, "Error while running workloads on ds")
				//})
			})

			Step("Delete DataServiceDeployment", func() {
				log.InfoD("Cleaning Up dataservice...")
				err := workflowDataservice.DeleteDeployment(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while deleting dataservice")
			})
		}
	})

	//TODO: Take backup and Restore the deployment once restore issue is resolved

	JustAfterEach(func() {
		defer EndTorpedoTest()
		Step("Delete PDS CustomTemplates", func() {
			log.InfoD("Cleaning Up templates...")
			err := workFlowTemplates.DeleteCreatedCustomPdsTemplates(templates)
			log.FailOnError(err, "Error while deleting dataservice")
		})
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
			workflowDataservice.Namespace = &WorkflowNamespace
			workflowDataservice.NamespaceName = Namespace

			//serviceConfigId, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams, ds.Name)
			//log.FailOnError(err, "Unable to create Custom Templates for PDS")

			workflowDataservice.PDSTemplates.ServiceConfigTemplateId = "tmpl:d1ed6519-fe79-463f-8d8f-8e6aaedb2f79"
			workflowDataservice.PDSTemplates.StorageTemplateId = "tmpl:a584ede7-811e-48bd-b000-ae799e3e084e"
			workflowDataservice.PDSTemplates.ResourceTemplateId = "tmpl:04dab835-1fe2-4526-824f-d7a45694676c"

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
			workflowDataservice.Namespace = &WorkflowNamespace
			workflowDataservice.NamespaceName = Namespace

			serviceConfigId, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams)
			log.FailOnError(err, "Unable to create Custom Templates for PDS")
			workflowDataservice.PDSTemplates.ServiceConfigTemplateId = serviceConfigId[ds.Name]
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

var _ = Describe("{GetPVCFullCondition}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("GetPVCFullCondition", "Deploy a dataservice and fill-up the PVC, Once full, resize the PVC", nil, 0)
	})
	var (
		workflowDataservice pds.WorkflowDataService
		workFlowTemplates   pds.WorkflowPDSTemplates
		deployment          *automationModels.PDSDeploymentResponse
		deployments         = make(map[dslibs.PDSDataService]*automationModels.PDSDeploymentResponse)
		templates           []string
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
			workflowDataservice.Namespace = &WorkflowNamespace
			workflowDataservice.NamespaceName = Namespace

			serviceConfigId, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams)
			log.FailOnError(err, "Unable to create Custom Templates for PDS")
			workflowDataservice.PDSTemplates.ServiceConfigTemplateId = serviceConfigId[ds.Name]
			workflowDataservice.PDSTemplates.StorageTemplateId = stConfigId
			workflowDataservice.PDSTemplates.ResourceTemplateId = resConfigId
			templates = append(templates, serviceConfigId[ds.Name], stConfigId, resConfigId)

			log.InfoD("Original Storage Template ID- [resTempId- %v]", stConfigId)
			deployment, err = workflowDataservice.DeployDataService(ds, ds.OldImage, ds.OldVersion)
			log.FailOnError(err, "Error while deploying ds")
			deployments[ds] = deployment

			defer func() {
				Step("Delete PDS CustomTemplates", func() {
					log.InfoD("Cleaning Up templates...")
					err := workFlowTemplates.DeleteCreatedCustomPdsTemplates(templates)
					log.FailOnError(err, "Error while deleting dataservice")
				})
			}()

			defer func() {
				for _, deployment := range deployments {
					Step("Delete DataServiceDeployment", func() {
						log.InfoD("Cleaning Up dataservice...")
						err := workflowDataservice.DeleteDeployment(*deployment.Create.Meta.Uid)
						log.FailOnError(err, "Error while deleting dataservice")
					})
				}
			}()

			log.InfoD("Running Workloads to fill up the PVC")
			err = workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
			log.FailOnError(err, "Error while running workloads on ds")

			log.InfoD("Compute the PVC usage")
			err = workflowDataservice.CheckPVCStorageFullCondition(workflowDataservice.NamespaceName, *deployment.Create.Status.CustomResourceName, 85)
			log.FailOnError(err, "Error while checking for pvc full condition")

			log.InfoD("Once pvc has reached threshold, increase the ovc by 1gb")
			err = workflowDataservice.IncreasePvcSizeBy1gb(workflowDataservice.NamespaceName, *deployment.Create.Status.CustomResourceName, 1)
			log.FailOnError(err, "Failing while Increasing the PVC name...")

			log.InfoD("Validate deployment after PVC increase")
			err = workflowDataservice.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.Replicas, resConfigId, stConfigId, workflowDataservice.NamespaceName, ds.Version, ds.Image)
		}
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{DeletePDSPods}", func() {
	var (
		workflowDataservice pds.WorkflowDataService
		workFlowTemplates   pds.WorkflowPDSTemplates
		deployment          *automationModels.PDSDeploymentResponse
		dsNameAndAppTempId  map[string]string
		stConfigId          string
		resConfigId         string
		templates           []string
		err                 error
	)

	JustBeforeEach(func() {
		StartTorpedoTest("DeletePDSPods", "delete pds pods and validate if its coming back online and dataServices are not affected", nil, 0)
		workFlowTemplates.Platform = WorkflowPlatform
		workflowDataservice.Namespace = &WorkflowNamespace
		workflowDataservice.NamespaceName = PDS_DEFAULT_NAMESPACE
		workflowDataservice.Dash = dash
	})

	It("Delete pds pods and validate if its coming back online and dataserices are not affected", func() {

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

			//stepLog := "Running Workloads before deleting pods in Px-System namespace"
			//Step(stepLog, func() {
			//	err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
			//	log.FailOnError(err, "Error while running workloads on ds")
			//})

			Step("Delete PDSPods", func() {
				err := workflowDataservice.DeletePDSPods()
				log.FailOnError(err, "Error while deleting pds pods")
				err = workflowDataservice.ValidatePdsDataServiceDeployments(
					*deployment.Create.Meta.Uid,
					ds,
					ds.Replicas,
					workflowDataservice.PDSTemplates.ResourceTemplateId,
					workflowDataservice.PDSTemplates.StorageTemplateId,
					workflowDataservice.NamespaceName,
					ds.Version,
					ds.Image)
				log.FailOnError(err, "Error while Validating dataservice")
			})

			Step("Delete DataServiceDeployment", func() {
				log.InfoD("Cleaning Up dataservice...")
				err := workflowDataservice.DeleteDeployment(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while deleting dataservice")
			})
		}
	})
	JustAfterEach(func() {
		Step("Delete PDS CustomTemplates", func() {
			log.InfoD("Cleaning Up templates...")
			err := workFlowTemplates.DeleteCreatedCustomPdsTemplates(templates)
			log.FailOnError(err, "Error while deleting dataservice")
		})
		defer EndTorpedoTest()
	})
})
