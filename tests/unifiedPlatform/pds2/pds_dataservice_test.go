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

const (
	PlatformNamespace = "px-system"
)

var _ = Describe("{DeployDataServicesOnDemandAndScaleUp}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("DeployDataServicesOnDemandAndScaleUp", "Deploy data services and perform scale up", nil, 0)
	})

	It("Deploy,Validate and ScaleUp DataService", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
			})

			Step("ScaleUp DataService", func() {
				log.InfoD("Scaling Up dataServices...")
				updateDeployment, err := WorkflowDataService.UpdateDataService(ds, *deployment.Create.Meta.Uid, ds.Image, ds.Version)
				log.FailOnError(err, "Error while updating ds")
				log.Debugf("Updated Deployment Id: [%s]", *updateDeployment.Update.Meta.Uid)
			})

			stepLog := "Running Workloads after ScaleUp of DataService"
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

var _ = Describe("{UpgradeDataServiceImage}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("UpgradeDataServiceImage", "Upgrade Data Service Image", nil, 0)
	})

	It("Deploy, Validate and Upgrade Data service Image", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.OldImage, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
			})

			Step("Upgrade DataService Image", func() {
				_, err := WorkflowDataService.UpdateDataService(ds, *deployment.Create.Meta.Uid, ds.Image, ds.Version)
				log.FailOnError(err, "Error while updating ds")
			})

			stepLog := "Running Workloads after upgrading the ds image"
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

var _ = Describe("{ScaleUpCpuMemLimitsandStorageOfDS}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("ScaleUpCpuMemLimitsandStorageOfDS", "Deploy a dataservice and scale up its CPU/MEM limits and storage size by editing the respective template", nil, 0)
	})

	It("Deploy,Validate and ScaleUp DataService", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
			})

			//stepLog := "Running Workloads before taking backups"
			//Step(stepLog, func() {
			//	err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
			//	log.FailOnError(err, "Error while running workloads on ds")
			//})

			//Update Ds With New Values of Resource Templates
			resConfigIdUpdated, err := WorkflowPDSTemplate.CreateResourceTemplateWithCustomValue(NewPdsParams)
			log.FailOnError(err, "Unable to create Custom Templates for PDS")
			log.InfoD("Updated Resource Template ID- [updated- %v]", resConfigIdUpdated)

			WorkflowDataService.UpdateDeploymentTemplates = true
			_, err = WorkflowDataService.UpdateDataService(ds, *deployment.Create.Meta.Uid, ds.Image, ds.Version)
			log.FailOnError(err, "Error while updating ds")

			//stepLog = "Running Workloads after ScaleUp of DataService"
			//Step(stepLog, func() {
			//	err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
			//	log.FailOnError(err, "Error while running workloads on ds")
			//})
		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
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
		err                 error
	)
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
			_, err = workflowDataservice.RunDataServiceWorkloads(*deployment.Create.Meta.Uid)
			log.FailOnError(err, "Error while running workloads on ds")

			log.InfoD("Compute the PVC usage")
			err = workflowDataservice.CheckPVCStorageFullCondition(workflowDataservice.DataServiceDeployment[*deployment.Create.Meta.Uid].Namespace, *deployment.Create.Status.CustomResourceName, 85)
			log.FailOnError(err, "Error while checking for pvc full condition")

			log.InfoD("Once pvc has reached threshold, increase the ovc by 1gb")
			err = workflowDataservice.IncreasePvcSizeBy1gb(workflowDataservice.DataServiceDeployment[*deployment.Create.Meta.Uid].Namespace, *deployment.Create.Status.CustomResourceName, 1)
			log.FailOnError(err, "Failing while Increasing the PVC name...")

			//log.InfoD("Validate deployment after PVC increase")
			//err = workflowDataservice.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.Replicas, resConfigId, stConfigId, workflowDataservice.DataServiceDeployment[*deployment.Create.Meta.Uid].Namespace, ds.Version, ds.Image)
		}
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{DeletePDSPods}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("DeletePDSPods", "delete pds pods and validate if its coming back online and dataServices are not affected", nil, 0)
	})

	It("Delete pds pods and validate if its coming back online and dataserices are not affected", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
			})

			stepLog := "Running Workloads before deleting pods in Px-System namespace"
			Step(stepLog, func() {
				_, err := WorkflowDataService.RunDataServiceWorkloads(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while running workloads on ds")
			})

			Step("Delete PDSPods", func() {
				err := WorkflowDataService.DeletePDSPods([]string{"pds-backups", "pds-target"}, PlatformNamespace)
				log.FailOnError(err, "Error while deleting pds pods")
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(
					*deployment.Create.Meta.Uid,
					ds,
					ds.Replicas,
					WorkflowDataService.PDSTemplates.ResourceTemplateId,
					WorkflowDataService.PDSTemplates.StorageTemplateId,
					PDS_DEFAULT_NAMESPACE,
					ds.Version,
					ds.Image)
				log.FailOnError(err, "Error while Validating dataservice")
			})

			stepLog = "Running Workloads after deleting pods in Px-System namespace"
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

var _ = Describe("{ValidatePdsHealthIncaseofFailures}", func() {
	var (
		deployment *automationModels.PDSDeploymentResponse
		err        error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("ValidatePdsHealthIncaseofFailures", "Deploy data services and validate PDS health in case of PDS pod deletion", nil, 0)
	})

	It("Deploy data services, Delete Pds Agent pods and perform backup and restore on the same cluster", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {

			steplog := "Deploy dataservice"
			Step(steplog, func() {
				log.InfoD(steplog)
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)
				WorkflowPDSRestore.SourceDeploymentConfigBeforeUpgrade = &deployment.Create.Config.DeploymentTopologies[0]
			})

			steplog = "Restart PDS Agent Pods and Validate if it comes up"
			Step(steplog, func() {
				log.InfoD(steplog)
				err := WorkflowDataService.DeletePDSPods([]string{*deployment.Create.Status.CustomResourceName}, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deleting pds pods")
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(
					*deployment.Create.Meta.Uid,
					ds,
					ds.Replicas,
					WorkflowDataService.PDSTemplates.ResourceTemplateId,
					WorkflowDataService.PDSTemplates.StorageTemplateId,
					PDS_DEFAULT_NAMESPACE,
					ds.Version,
					ds.Image)
				log.FailOnError(err, "Error while Validating dataservice")
			})

			steplog = "ScaleUp DataService"
			Step(steplog, func() {
				log.InfoD(steplog)
				updateDeployment, err := WorkflowDataService.UpdateDataService(ds, *deployment.Create.Meta.Uid, ds.Image, ds.Version)
				log.FailOnError(err, "Error while updating ds")
				log.Debugf("Updated Deployment Id: [%s]", *updateDeployment.Update.Meta.Uid)
			})
		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})
