package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
	"strings"
)

var _ = Describe("{PerformRestoreToSameCluster}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("PerformRestoreToSameCluster", "Deploy data services and perform backup and restore on the same cluster", nil, 0)
	})
	var (
		workflowDataService  pds.WorkflowDataService
		workflowBackUpConfig pds.WorkflowPDSBackupConfig
		workflowRestore      pds.WorkflowPDSRestore
		deployment           *automationModels.PDSDeploymentResponse
		restoreDeployment    *automationModels.PDSRestoreResponse
		pdsBackupConfigName  string
		err                  error
	)

	It("Deploy, Validate and RunWorkloads on DataService", func() {
		Step("Create a PDS Namespace", func() {
			Namespace = strings.ToLower("pds-test-ns-" + utilities.RandString(5))
			WorkflowNamespace.TargetCluster = WorkflowTargetCluster
			WorkflowNamespace.Namespaces = make(map[string]string)
			workflowNamespace, err := WorkflowNamespace.CreateNamespaces(Namespace)
			log.FailOnError(err, "Unable to create namespace")
			log.Infof("Namespaces created - [%s]", workflowNamespace.Namespaces)
			log.Infof("Namespace id - [%s]", workflowNamespace.Namespaces[Namespace])
		})

		for _, ds := range NewPdsParams.DataServiceToTest {
			workflowDataService.Namespace = WorkflowNamespace
			workflowDataService.NamespaceName = Namespace
			deployment, err = workflowDataService.DeployDataService(ds, ds.Image, ds.Version)
			log.FailOnError(err, "Error while deploying ds")
		}

		defer func() {
			Step("Delete DataServiceDeployment", func() {
				err := workflowDataService.DeleteDeployment()
				log.FailOnError(err, "Error while deleting dataservice")
			})
		}()

		stepLog := "Running Workloads before taking backups"
		Step(stepLog, func() {
			err := workflowDataService.RunDataServiceWorkloads(NewPdsParams)
			log.FailOnError(err, "Error while running workloads on ds")
		})
	})

	It("Perform adhoc backup, restore and validate them", func() {
		var bkpConfigResponse *automationModels.PDSBackupConfigResponse
		workflowBackUpConfig.WorkflowDataService = workflowDataService
		workflowBackUpConfig.WorkflowBackupLocation = WorkflowbkpLoc
		pdsBackupConfigName = strings.ToLower("pds-qa-bkpConfig-" + utilities.RandString(5))

		Step("Take Backup and validate", func() {
			bkpConfigResponse, err = workflowBackUpConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
			log.FailOnError(err, "Error occured while creating backupConfig")
			log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", bkpConfigResponse.Create.Meta.Name, bkpConfigResponse.Create.Meta.Uid)
		})

		defer func() {
			Step("Delete Backups", func() {
				err = workflowBackUpConfig.DeleteBackupConfig(pdsBackupConfigName)
				log.FailOnError(err, "Error while deleting BackupConfig [%s]", pdsBackupConfigName)
			})
		}()

		Step("Perform Restore and validate", func() {
			workflowRestore.WorkflowDataService = workflowDataService
			backupUid := *bkpConfigResponse.Create.Meta.Uid
			deploymentName := *deployment.Create.Meta.Name
			cloudSnapId := ""
			// Set the DestClusterId same as the current ClusterId
			workflowRestore.Destination.DestinationClusterId = WorkflowTargetCluster.ClusterUID

			log.FailOnError(err, "failed while registering destination target cluster")

			workflowRestore.WorkflowBackupLocation = WorkflowbkpLoc
			restoreDeployment, err = workflowRestore.CreateRestore(backupUid, deploymentName, cloudSnapId)
			log.FailOnError(err, "Error while taking restore")
			log.Debugf("Restored DeploymentName: [%s]", restoreDeployment.Create.Meta.Name)
		})

		defer func() {
			Step("Delete RestoredDeployment", func() {
				err := workflowRestore.DeleteRestore(*restoreDeployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while deleting restore")
			})
		}()

		Step("Validate md5hash for the restored deployments", func() {
			err := workflowDataService.ValidateDataServiceWorkloads(NewPdsParams, restoreDeployment)
			log.FailOnError(err, "Error occured in ValidateDataServiceWorkloads method")
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{PerformRestoreToDifferentClusterSameProject}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("PerformRestoreToDifferentClusterSameProject", "Deploy data services and perform backup and restore on a different cluster on the same project", nil, 0)
	})
	var (
		workflowDataService  pds.WorkflowDataService
		workflowBackUpConfig pds.WorkflowPDSBackupConfig
		workflowRestore      pds.WorkflowPDSRestore
		deployment           *automationModels.PDSDeploymentResponse
		restoreDeployment    *automationModels.PDSRestoreResponse
		pdsBackupConfigName  string
		err                  error
	)

	It("Deploy, Validate and RunWorkloads on DataService", func() {
		Step("Create a PDS Namespace", func() {
			Namespace = strings.ToLower("pds-test-ns-" + utilities.RandString(5))
			WorkflowNamespace.TargetCluster = WorkflowTargetCluster
			WorkflowNamespace.Namespaces = make(map[string]string)
			workflowNamespace, err := WorkflowNamespace.CreateNamespaces(Namespace)
			log.FailOnError(err, "Unable to create namespace")
			log.Infof("Namespaces created - [%s]", workflowNamespace.Namespaces)
			log.Infof("Namespace id - [%s]", workflowNamespace.Namespaces[Namespace])
		})

		for _, ds := range NewPdsParams.DataServiceToTest {
			workflowDataService.Namespace = WorkflowNamespace
			workflowDataService.NamespaceName = Namespace
			deployment, err = workflowDataService.DeployDataService(ds, ds.Image, ds.Version)
			log.FailOnError(err, "Error while deploying ds")
		}

		defer func() {
			Step("Delete DataServiceDeployment", func() {
				err := workflowDataService.DeleteDeployment()
				log.FailOnError(err, "Error while deleting dataservice")
			})
		}()

		stepLog := "Running Workloads before taking backups"
		Step(stepLog, func() {
			err := workflowDataService.RunDataServiceWorkloads(NewPdsParams)
			log.FailOnError(err, "Error while running workloads on ds")
		})
	})

	It("Perform adhoc backup, restore and validate them", func() {
		var bkpConfigResponse *automationModels.PDSBackupConfigResponse
		workflowBackUpConfig.WorkflowDataService = workflowDataService
		workflowBackUpConfig.WorkflowBackupLocation = WorkflowbkpLoc
		pdsBackupConfigName = strings.ToLower("pds-qa-bkpConfig-" + utilities.RandString(5))

		Step("Take Backup and validate", func() {
			bkpConfigResponse, err = workflowBackUpConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
			log.FailOnError(err, "Error occured while creating backupConfig")
			log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", bkpConfigResponse.Create.Meta.Name, bkpConfigResponse.Create.Meta.Uid)
		})

		defer func() {
			Step("Delete Backups", func() {
				err = workflowBackUpConfig.DeleteBackupConfig(pdsBackupConfigName)
				log.FailOnError(err, "Error while deleting BackupConfig [%s]", pdsBackupConfigName)
			})
		}()

		Step("Perform Restore and validate", func() {
			workflowRestore.WorkflowDataService = workflowDataService
			backupUid := *bkpConfigResponse.Create.Meta.Uid
			deploymentName := *deployment.Create.Meta.Name
			cloudSnapId := ""

			//Set the context to  the destination clusterId
			err = SetDestinationKubeConfig()
			log.FailOnError(err, "failed while setting dest cluster path")

			destTargetCluster, err := WorkflowTargetCluster.RegisterToControlPlane(true)
			workflowRestore.Destination.DestinationClusterId = destTargetCluster.DestinationClusterId

			log.FailOnError(err, "failed while registering destination target cluster")

			workflowRestore.WorkflowBackupLocation = WorkflowbkpLoc
			restoreDeployment, err = workflowRestore.CreateRestore(backupUid, deploymentName, cloudSnapId)
			log.FailOnError(err, "Error while taking restore")
			log.Debugf("Restored DeploymentName: [%s]", restoreDeployment.Create.Meta.Name)
		})

		defer func() {
			Step("Delete RestoredDeployment", func() {
				err := workflowRestore.DeleteRestore(*restoreDeployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while deleting restore")
			})
		}()

		Step("Validate md5hash for the restored deployments", func() {
			err := workflowDataService.ValidateDataServiceWorkloads(NewPdsParams, restoreDeployment)
			log.FailOnError(err, "Error occured in ValidateDataServiceWorkloads method")
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{UpgradeDataServiceImageAndVersionWithBackUpRestore}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("UpgradeDataServiceImageAndVersionWithBackUpRestore", "Upgrade Data Service Version and Image", nil, 0)
	})
	var (
		workflowDataservice  pds.WorkflowDataService
		workFlowTemplates    pds.CustomTemplates
		workflowBackUpConfig pds.WorkflowPDSBackupConfig
		workflowRestore      pds.WorkflowPDSRestore
		deployment           *automationModels.PDSDeploymentResponse
		updatedDeployment    *automationModels.PDSDeploymentResponse
		restoreDeployment    *automationModels.PDSRestoreResponse
		bkpConfigResponse    *automationModels.PDSBackupConfigResponse
		pdsBackupConfigName  string
		oldDeploymentId      string
		newDeploymentId      string
		err                  error
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

			serviceConfigId, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams, false)
			log.FailOnError(err, "Unable to create Custom Templates for PDS")
			workflowDataservice.PDSTemplates.ServiceConfigTemplateId = serviceConfigId
			workflowDataservice.PDSTemplates.StorageTemplateId = stConfigId
			workflowDataservice.PDSTemplates.ResourceTemplateId = resConfigId
		})

		for _, ds := range NewPdsParams.DataServiceToTest {
			workflowDataservice.Namespace = WorkflowNamespace
			workflowDataservice.NamespaceName = Namespace
			deployment, err = workflowDataservice.DeployDataService(ds, ds.OldImage, ds.OldVersion)
			log.FailOnError(err, "Error while deploying ds")
		}

		defer func() {
			Step("Delete DataServiceDeployment", func() {
				err := workflowDataservice.DeleteDeployment()
				log.FailOnError(err, "Error while deleting dataservice")
			})
		}()

		stepLog := "Running Workloads before upgrading the ds image"
		Step(stepLog, func() {
			err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
			log.FailOnError(err, "Error while running workloads on ds")
		})
	})

	It("Perform adhoc backup of old deployments", func() {
		workflowBackUpConfig.WorkflowDataService = workflowDataservice
		workflowBackUpConfig.WorkflowBackupLocation = WorkflowbkpLoc
		pdsBackupConfigName = strings.ToLower("pds-qa-bkpConfig-" + utilities.RandString(5))

		Step("Take Backup and validate", func() {
			bkpConfigResponse, err = workflowBackUpConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
			log.FailOnError(err, "Error occured while creating backupConfig")
			log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", bkpConfigResponse.Create.Meta.Name, bkpConfigResponse.Create.Meta.Uid)
		})

		defer func() {
			Step("Delete Backups", func() {
				err = workflowBackUpConfig.DeleteBackupConfig(pdsBackupConfigName)
				log.FailOnError(err, "Error while deleting BackupConfig [%s]", pdsBackupConfigName)
			})
		}()
	})

	It("Upgrade DataService Version and Image", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			updatedDeployment, err = workflowDataservice.UpdateDataService(ds, oldDeploymentId, ds.Image, ds.Version)
			log.FailOnError(err, "Error while updating ds")
			log.Debugf("Updated Deployment Id [%s]", *updatedDeployment.Update.Meta.Uid)
		}

		stepLog := "Running Workloads after upgrading the ds image"
		Step(stepLog, func() {
			err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
			log.FailOnError(err, "Error while running workloads on ds")
		})
	})

	It("Restore the old deployment and upgrade the restored deployment", func() {
		Step("Restore the old deployment", func() {
			workflowRestore.WorkflowDataService = workflowDataservice
			backupUid := *bkpConfigResponse.Create.Meta.Uid
			deploymentName := *deployment.Create.Meta.Name
			cloudSnapId := ""
			// Set the DestClusterId same as the current ClusterId
			workflowRestore.Destination.DestinationClusterId = WorkflowTargetCluster.ClusterUID
			workflowRestore.WorkflowBackupLocation = WorkflowbkpLoc
			restoreDeployment, err = workflowRestore.CreateRestore(backupUid, deploymentName, cloudSnapId)
			log.FailOnError(err, "Error while taking restore")
			newDeploymentId = restoreDeployment.Create.Config.DestinationReferences.DeploymentId
			log.Debugf("Restored DeploymentName: [%s]", restoreDeployment.Create.Meta.Name)
		})

		defer func() {
			Step("Delete RestoredDeployment", func() {
				err := workflowRestore.DeleteRestore(*restoreDeployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while deleting restore")
			})
		}()

		Step("Update restored DataService Version and Image", func() {
			for _, ds := range NewPdsParams.DataServiceToTest {
				updatedRestoredDeployment, err := workflowDataservice.UpdateDataService(ds, newDeploymentId, ds.Image, ds.Version)
				log.FailOnError(err, "Error while updating ds")
				log.Debugf("Updated Deployment Id [%s]", *updatedRestoredDeployment.Update.Meta.Uid)
			}

			stepLog := "Running Workloads after upgrading the ds image"
			Step(stepLog, func() {
				err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
				log.FailOnError(err, "Error while running workloads on ds")
			})
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{PerformRestoreToDifferentClusterProject}", func() {
	var (
		workflowDataService       pds.WorkflowDataService
		workflowBackUpConfig      pds.WorkflowPDSBackupConfig
		workflowRestore           pds.WorkflowPDSRestore
		deployment                *automationModels.PDSDeploymentResponse
		restoreDeployment         *automationModels.PDSRestoreResponse
		workflowProjectDest       platform.WorkflowProject       // Workflow for destination project
		workflowTargetClusterDest platform.WorkflowTargetCluster // Workflow for destination target cluster
		pdsBackupConfigName       string
		err                       error
	)

	JustBeforeEach(func() {
		StartTorpedoTest("PerformRestoreToDifferentClusterProject", "Deploy data services and perform backup and restore on the different cluster from different project", nil, 0)
	})

	It("Deploy, Validate and RunWorkloads on DataService", func() {

		Step("Create Project for destination", func() {
			workflowProjectDest.Platform = WorkflowPlatform
			workflowProjectDest.ProjectName = fmt.Sprintf("project-%s", utilities.RandomString(5))
			workflowProject, err := workflowProjectDest.CreateProject()
			log.FailOnError(err, "Unable to create project")
			log.Infof("Project created with ID - [%s]", workflowProject.ProjectId)
		})

		Step("Register Target Cluster for destination", func() {
			workflowTargetClusterDest.Project = workflowProjectDest
			log.Infof("Tenant ID [%s]", workflowTargetClusterDest.Project.Platform.TenantId)
			workflowTargetCluster, err := workflowTargetClusterDest.RegisterToControlPlane(false)
			log.FailOnError(err, "Unable to register target cluster")
			log.Infof("Target cluster registered with uid - [%s]", workflowTargetCluster.ClusterUID)
		})

		Step("Create a PDS Namespace", func() {
			Namespace = strings.ToLower("pds-test-ns-" + utilities.RandString(5))
			WorkflowNamespace.TargetCluster = WorkflowTargetCluster
			WorkflowNamespace.Namespaces = make(map[string]string)
			workflowNamespace, err := WorkflowNamespace.CreateNamespaces(Namespace)
			log.FailOnError(err, "Unable to create namespace")
			log.Infof("Namespaces created - [%s]", workflowNamespace.Namespaces)
			log.Infof("Namespace id - [%s]", workflowNamespace.Namespaces[Namespace])
		})

		for _, ds := range NewPdsParams.DataServiceToTest {
			workflowDataService.Namespace = WorkflowNamespace
			workflowDataService.NamespaceName = Namespace
			deployment, err = workflowDataService.DeployDataService(ds, ds.Image, ds.Version)
			log.FailOnError(err, "Error while deploying ds")
		}

		defer func() {
			Step("Delete DataServiceDeployment", func() {
				err := workflowDataService.DeleteDeployment()
				log.FailOnError(err, "Error while deleting dataservice")
			})
		}()

		stepLog := "Running Workloads before taking backups"
		Step(stepLog, func() {
			err := workflowDataService.RunDataServiceWorkloads(NewPdsParams)
			log.FailOnError(err, "Error while running workloads on ds")
		})
	})

	It("Perform adhoc backup, restore and validate them", func() {
		var bkpConfigResponse *automationModels.PDSBackupConfigResponse
		workflowBackUpConfig.WorkflowDataService = workflowDataService
		workflowBackUpConfig.WorkflowBackupLocation = WorkflowbkpLoc
		pdsBackupConfigName = strings.ToLower("pds-qa-bkpConfig-" + utilities.RandString(5))

		Step("Take Backup and validate", func() {
			bkpConfigResponse, err = workflowBackUpConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
			log.FailOnError(err, "Error occured while creating backupConfig")
			log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", bkpConfigResponse.Create.Meta.Name, bkpConfigResponse.Create.Meta.Uid)
		})

		defer func() {
			Step("Delete Backups", func() {
				err = workflowBackUpConfig.DeleteBackupConfig(pdsBackupConfigName)
				log.FailOnError(err, "Error while deleting BackupConfig [%s]", pdsBackupConfigName)
			})
		}()

		Step("Perform Restore on destination cluster and validate", func() {
			workflowRestore.WorkflowDataService = workflowDataService
			backupUid := *bkpConfigResponse.Create.Meta.Uid
			deploymentName := *deployment.Create.Meta.Name
			cloudSnapId := ""
			// Set the DestClusterId same as the current ClusterId
			// Creating restore on target cluster
			workflowRestore.Destination = workflowTargetClusterDest
			workflowRestore.WorkflowBackupLocation = WorkflowbkpLoc
			restoreDeployment, err = workflowRestore.CreateRestore(backupUid, deploymentName, cloudSnapId)
			log.FailOnError(err, "Error while taking restore")
			log.Debugf("Restored DeploymentName: [%s]", restoreDeployment.Create.Meta.Name)
		})

		defer func() {
			Step("Delete RestoredDeployment", func() {
				err := workflowRestore.DeleteRestore(*restoreDeployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while deleting restore")
			})
		}()

		Step("Validate md5hash for the restored deployments", func() {
			err := workflowDataService.ValidateDataServiceWorkloads(NewPdsParams, restoreDeployment)
			log.FailOnError(err, "Error occured in ValidateDataServiceWorkloads method")
		})

	})

	JustAfterEach(func() {

		err = workflowTargetClusterDest.DeregisterFromControlPlane()
		log.FailOnError(err, "Unable to clean target cluster")
		err = workflowProjectDest.DeleteProject()
		log.FailOnError(err, "Unable to clean target project")
		defer EndTorpedoTest()
	})
})

var _ = Describe("{PerformRestoreAfterPVCResize}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("PerformRestoreAfterPVCResize", "Deploy data services, increase PVC Size and perform backup and restore on the same cluster", nil, 0)
	})
	var (
		workflowDataService  pds.WorkflowDataService
		workflowBackUpConfig pds.WorkflowPDSBackupConfig
		workflowRestore      pds.WorkflowPDSRestore
		deployment           *automationModels.PDSDeploymentResponse
		restoreDeployment    *automationModels.PDSRestoreResponse

		workFlowTemplates pds.CustomTemplates
		tempList          []string

		pdsBackupConfigName string
		err                 error
	)

	It("Deploy, Validate and RunWorkloads on DataService", func() {
		Step("Create a PDS Namespace", func() {
			Namespace = strings.ToLower("pds-test-ns-" + utilities.RandString(5))
			WorkflowNamespace.TargetCluster = WorkflowTargetCluster
			WorkflowNamespace.Namespaces = make(map[string]string)
			workflowNamespace, err := WorkflowNamespace.CreateNamespaces(Namespace)
			log.FailOnError(err, "Unable to create namespace")
			log.Infof("Namespaces created - [%s]", workflowNamespace.Namespaces)
			log.Infof("Namespace id - [%s]", workflowNamespace.Namespaces[Namespace])
		})

		for _, ds := range NewPdsParams.DataServiceToTest {
			workflowDataService.Namespace = WorkflowNamespace
			workflowDataService.NamespaceName = Namespace
			serviceConfigId, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams, false)
			log.FailOnError(err, "Unable to create Custom Templates for PDS")
			workflowDataService.PDSTemplates.ServiceConfigTemplateId = serviceConfigId
			workflowDataService.PDSTemplates.StorageTemplateId = stConfigId
			workflowDataService.PDSTemplates.ResourceTemplateId = resConfigId
			tempList = append(tempList, serviceConfigId, stConfigId, resConfigId)
			deployment, err = workflowDataService.DeployDataService(ds, ds.Image, ds.Version)
			log.FailOnError(err, "Error while deploying ds")
		}

		defer func() {
			Step("Delete DataServiceDeployment", func() {
				err := workflowDataService.DeleteDeployment()
				log.FailOnError(err, "Error while deleting dataservice")
			})
		}()

		defer func() {
			Step("Delete created Templates", func() {
				err := workFlowTemplates.DeleteCreatedCustomPdsTemplates(tempList)
				log.FailOnError(err, "Unable to delete Custom Templates for PDS")
			})
		}()
		stepLog := "Running Workloads before taking backups"
		Step(stepLog, func() {
			err := workflowDataService.RunDataServiceWorkloads(NewPdsParams)
			log.FailOnError(err, "Error while running workloads on ds")
		})
	})
	It("Perform adhoc backup, restore before PVC Resize and validate them", func() {
		var bkpConfigResponse *automationModels.PDSBackupConfigResponse
		workflowBackUpConfig.WorkflowDataService = workflowDataService
		pdsBackupConfigName = strings.ToLower("pds-qa-bkpConfig-" + utilities.RandString(5))

		Step("Take Backup and validate", func() {
			bkpConfigResponse, err = workflowBackUpConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
			log.FailOnError(err, "Error occured while creating backupConfig")
			log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", bkpConfigResponse.Create.Meta.Name, bkpConfigResponse.Create.Meta.Uid)
		})

		defer func() {
			Step("Delete Backups", func() {
				err = workflowBackUpConfig.DeleteBackupConfig(pdsBackupConfigName)
				log.FailOnError(err, "Error while deleting BackupConfig [%s]", pdsBackupConfigName)
			})
		}()

		Step("Perform Restore and validate", func() {
			workflowRestore.WorkflowDataService = workflowDataService
			backupUid := *bkpConfigResponse.Create.Meta.Uid
			deploymentName := *deployment.Create.Meta.Name
			cloudSnapId := ""
			// Set the DestClusterId same as the current ClusterId
			workflowRestore.Destination.DestinationClusterId = WorkflowTargetCluster.ClusterUID

			log.FailOnError(err, "failed while registering destination target cluster")

			workflowRestore.WorkflowBackupLocation = WorkflowbkpLoc
			restoreDeployment, err = workflowRestore.CreateRestore(backupUid, deploymentName, cloudSnapId)
			log.FailOnError(err, "Error while taking restore")
			log.Debugf("Restored DeploymentName: [%s]", restoreDeployment.Create.Meta.Name)
		})

		defer func() {
			Step("Delete RestoredDeployment", func() {
				err := workflowRestore.DeleteRestore(*restoreDeployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while deleting restore")
			})
		}()

		Step("Validate md5hash for the restored deployments", func() {
			err := workflowDataService.ValidateDataServiceWorkloads(NewPdsParams, restoreDeployment)
			log.FailOnError(err, "Error occured in ValidateDataServiceWorkloads method")
		})

	})

	It("Increase PVC Size by 1 GB of DataService from K8s", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			log.InfoD("Dataservice on which the PVC needs to be resized is- [%v]", ds.Name)
			err = workflowDataService.IncreasePvcSizeBy1gb(workflowDataService.NamespaceName, workflowDataService.DataServiceDeployment, 1)
			log.FailOnError(err, "Failing while Increasing the PVC name...")

		}
		stepLog := "Validate the deployment after PVC Resize"
		Step(stepLog, func() {
			//Validate deployment function call here
		})
		stepLog = "Running Workloads after Resize of PVC"
		Step(stepLog, func() {
			err := workflowDataService.RunDataServiceWorkloads(NewPdsParams)
			log.FailOnError(err, "Error while running workloads on ds")
		})
	})
	It("Perform adhoc backup, restore after PVC Resize and validate them", func() {
		var bkpConfigResponse *automationModels.PDSBackupConfigResponse
		workflowBackUpConfig.WorkflowDataService = workflowDataService
		pdsBackupConfigName = strings.ToLower("pds-qa-bkpConfig-" + utilities.RandString(5))

		Step("Take Backup and validate", func() {
			bkpConfigResponse, err = workflowBackUpConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
			log.FailOnError(err, "Error occured while creating backupConfig")
			log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", bkpConfigResponse.Create.Meta.Name, bkpConfigResponse.Create.Meta.Uid)
		})

		defer func() {
			Step("Delete Backups", func() {
				err = workflowBackUpConfig.DeleteBackupConfig(pdsBackupConfigName)
				log.FailOnError(err, "Error while deleting BackupConfig [%s]", pdsBackupConfigName)
			})
		}()

		Step("Perform Restore and validate", func() {
			workflowRestore.WorkflowDataService = workflowDataService
			backupUid := *bkpConfigResponse.Create.Meta.Uid
			deploymentName := *deployment.Create.Meta.Name
			cloudSnapId := ""
			// Set the DestClusterId same as the current ClusterId
			workflowRestore.Destination.DestinationClusterId = WorkflowTargetCluster.ClusterUID

			log.FailOnError(err, "failed while registering destination target cluster")

			workflowRestore.WorkflowBackupLocation = WorkflowbkpLoc
			restoreDeployment, err = workflowRestore.CreateRestore(backupUid, deploymentName, cloudSnapId)
			log.FailOnError(err, "Error while taking restore")
			log.Debugf("Restored DeploymentName: [%s]", restoreDeployment.Create.Meta.Name)
		})

		defer func() {
			Step("Delete RestoredDeployment", func() {
				err := workflowRestore.DeleteRestore(*restoreDeployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while deleting restore")
			})
		}()

		Step("Validate md5hash for the restored deployments", func() {
			err := workflowDataService.ValidateDataServiceWorkloads(NewPdsParams, restoreDeployment)
			log.FailOnError(err, "Error occured in ValidateDataServiceWorkloads method")
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
