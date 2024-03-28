package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows"
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
		workflowDataService  stworkflows.WorkflowDataService
		workflowBackUpConfig stworkflows.WorkflowPDSBackupConfig
		workflowRestore      stworkflows.WorkflowPDSRestore
		deployment           *automationModels.WorkFlowResponse
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
		pdsBackupConfigName = strings.ToLower("pds-qa-bkpConfig-" + utilities.RandString(5))

		Step("Take Backup and validate", func() {
			bkpConfigResponse, err = workflowBackUpConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.PDSDeployment.V1Deployment.Meta.Uid)
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
			deploymentName := *deployment.PDSDeployment.V1Deployment.Meta.Name
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
		workflowDataService  stworkflows.WorkflowDataService
		workflowBackUpConfig stworkflows.WorkflowPDSBackupConfig
		workflowRestore      stworkflows.WorkflowPDSRestore
		deployment           *automationModels.WorkFlowResponse
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
		pdsBackupConfigName = strings.ToLower("pds-qa-bkpConfig-" + utilities.RandString(5))

		Step("Take Backup and validate", func() {
			bkpConfigResponse, err = workflowBackUpConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.PDSDeployment.V1Deployment.Meta.Uid)
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
			deploymentName := *deployment.PDSDeployment.V1Deployment.Meta.Name
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

var _ = Describe("{PerformRestoreToDifferentClusterProject}", func() {
	var (
		workflowDataService       stworkflows.WorkflowDataService
		workflowBackUpConfig      stworkflows.WorkflowPDSBackupConfig
		workflowRestore           stworkflows.WorkflowPDSRestore
		deployment                *automationModels.WorkFlowResponse
		restoreDeployment         *automationModels.PDSRestoreResponse
		workflowProjectDest       stworkflows.WorkflowProject       // Workflow for destination project
		workflowTargetClusterDest stworkflows.WorkflowTargetCluster // Workflow for destination target cluster
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
		pdsBackupConfigName = strings.ToLower("pds-qa-bkpConfig-" + utilities.RandString(5))

		Step("Take Backup and validate", func() {
			bkpConfigResponse, err = workflowBackUpConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.PDSDeployment.V1Deployment.Meta.Uid)
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
			deploymentName := *deployment.PDSDeployment.V1Deployment.Meta.Name
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
