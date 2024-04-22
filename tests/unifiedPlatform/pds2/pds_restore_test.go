package tests

import (
	"fmt"
	"strings"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
)

var _ = Describe("{PerformRestoreToSameCluster}", func() {
	var (
		workflowDataservice  pds.WorkflowDataService
		workflowBackUpConfig pds.WorkflowPDSBackupConfig
		workflowRestore      pds.WorkflowPDSRestore
		workflowBackup       pds.WorkflowPDSBackup
		workFlowTemplates    pds.WorkflowPDSTemplates
		deployment           *automationModels.PDSDeploymentResponse
		latestBackupUid      string
		pdsBackupConfigName  string
		restoreNamespace     string
	)

	JustBeforeEach(func() {
		StartTorpedoTest("PerformRestoreToSameCluster", "Deploy data services and perform backup and restore on the same cluster", nil, 0)
		workflowDataservice.DataServiceDeployment = make(map[string]string)

		workflowRestore.Destination = WorkflowNamespace
		workflowRestore.WorkflowProject = WorkflowProject
		workflowDataservice.Dash = dash
		restoreNamespace = "restore-" + RandomString(5)
	})

	It("Deploy data services and perform backup and restore on the same cluster", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy dataservice", func() {
				workFlowTemplates.Platform = WorkflowPlatform
				workflowDataservice.Namespace = WorkflowNamespace
				workflowDataservice.NamespaceName = PDS_DEFAULT_NAMESPACE

				serviceConfigId, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams, ds.Name)
				log.FailOnError(err, "Unable to create Custom Templates for PDS")

				workflowDataservice.PDSTemplates.ServiceConfigTemplateId = serviceConfigId
				workflowDataservice.PDSTemplates.StorageTemplateId = stConfigId
				workflowDataservice.PDSTemplates.ResourceTemplateId = resConfigId

				deployment, err = workflowDataservice.DeployDataService(ds, ds.Image, ds.Version)
				log.FailOnError(err, "Error while deploying ds")
				log.Infof("All deployments - [%+v]", workflowDataservice.DataServiceDeployment)

				//stepLog := "Running Workloads on deployment"
				//Step(stepLog, func() {
				//	err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
				//	log.FailOnError(err, "Error while running workloads on ds")
				//})
			})

			Step("Create Adhoc backup config of the existing deployment", func() {
				workflowBackUpConfig.WorkflowDataService = workflowDataservice
				workflowBackUpConfig.WorkflowBackupLocation = WorkflowbkpLoc
				pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
				workflowBackUpConfig.Backups = make(map[string]automationModels.V1BackupConfig)
				bkpConfigResponse, err := workflowBackUpConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Name)
				log.FailOnError(err, "Error occured while creating backupConfig")
				log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
				log.Infof("All deployments - [%+v]", workflowDataservice.DataServiceDeployment)
			})

			Step("Get the latest backup detail for the deployment", func() {
				workflowBackup.WorkflowDataService = workflowDataservice
				log.Infof("All deployments - [%+v]", workflowDataservice.DataServiceDeployment)
				backupResponse, err := workflowBackup.GetLatestBackup(*deployment.Create.Meta.Name)
				log.FailOnError(err, "Error occured while creating backup")
				latestBackupUid = *backupResponse.Meta.Uid
				log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
				err = workflowBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
				log.FailOnError(err, "Error occured while waiting for backup to complete")
			})

			Step("Create a new namespace for restore", func() {
				_, err := WorkflowNamespace.CreateNamespaces(restoreNamespace)
				log.FailOnError(err, "Unable to create namespace")
				log.Infof("Namespaces created - [%s]", WorkflowNamespace.Namespaces)
			})

			Step("Associate restore namespace to Project", func() {
				err := WorkflowProject.Associate(
					[]string{},
					[]string{WorkflowNamespace.Namespaces[restoreNamespace]},
					[]string{},
					[]string{},
					[]string{},
					[]string{},
				)
				log.FailOnError(err, "Unable to associate Cluster to Project")
				log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)
			})

			Step("Create Restore from the latest backup Id", func() {
				restoreName := "restore-" + RandomString(5)
				workflowRestore.Destination = WorkflowNamespace
				workflowRestore.WorkflowProject = WorkflowProject
				_, err := workflowRestore.CreateRestore(restoreName, latestBackupUid, restoreNamespace)
				log.FailOnError(err, "Restore Failed")
				log.Infof("All restores - [%+v]", workflowRestore.Restores)
			})
		}

		JustAfterEach(func() {
			log.Infof("Cleaning up all resources")
			err := workflowBackup.Purge(*deployment.Create.Meta.Name)
			log.FailOnError(err, "Backup cleanup failed")
			err = workflowBackUpConfig.Purge(true)
			log.FailOnError(err, "Backup Configs cleanup failed")
			err = workflowDataservice.DeleteDeployment(*deployment.Create.Meta.Name)
			log.FailOnError(err, "Data Service cleanup failed")
			defer EndTorpedoTest()
		})

	})
})

var _ = Describe("{PerformRestoreToDifferentClusterSameProject}", func() {
	var (
		workflowDataservice  pds.WorkflowDataService
		workflowBackUpConfig pds.WorkflowPDSBackupConfig
		workflowRestore      pds.WorkflowPDSRestore
		workflowBackup       pds.WorkflowPDSBackup
		workFlowTemplates    pds.WorkflowPDSTemplates
		deployment           *automationModels.PDSDeploymentResponse
		destinationCluster   platform.WorkflowTargetCluster
		destinationNamespace platform.WorkflowNamespace
		latestBackupUid      string
		pdsBackupConfigName  string
		restoreNamespace     string
	)

	JustBeforeEach(func() {
		StartTorpedoTest("PerformRestoreToDifferentClusterSameProject", "Deploy data services and perform backup and restore on a different cluster on the same project", nil, 0)
		workflowDataservice.DataServiceDeployment = make(map[string]string)

		workflowRestore.Destination = WorkflowNamespace
		workflowRestore.WorkflowProject = WorkflowProject
		workflowDataservice.Dash = dash
		restoreNamespace = "pds-restore-namespace-" + RandomString(5)
	})

	It("Deploy data services and perform backup and restore on the different cluster", func() {

		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy dataservice", func() {

				workFlowTemplates.Platform = WorkflowPlatform

				workflowDataservice.Namespace = WorkflowNamespace
				workflowDataservice.NamespaceName = PDS_DEFAULT_NAMESPACE

				serviceConfigId, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams, ds.Name)
				log.FailOnError(err, "Unable to create Custom Templates for PDS")

				workflowDataservice.PDSTemplates.ServiceConfigTemplateId = serviceConfigId
				workflowDataservice.PDSTemplates.StorageTemplateId = stConfigId
				workflowDataservice.PDSTemplates.ResourceTemplateId = resConfigId

				deployment, err = workflowDataservice.DeployDataService(ds, ds.Image, ds.Version)
				log.FailOnError(err, "Error while deploying ds")
				log.Infof("All deployments - [%+v]", workflowDataservice.DataServiceDeployment)

				//stepLog := "Running Workloads on deployment"
				//Step(stepLog, func() {
				//	err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
				//	log.FailOnError(err, "Error while running workloads on ds")
				//})
			})

			Step("Create Adhoc backup config of the existing deployment", func() {
				workflowBackUpConfig.WorkflowDataService = workflowDataservice
				workflowBackUpConfig.WorkflowBackupLocation = WorkflowbkpLoc
				pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
				workflowBackUpConfig.Backups = make(map[string]automationModels.V1BackupConfig)
				bkpConfigResponse, err := workflowBackUpConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Name)
				log.FailOnError(err, "Error occured while creating backupConfig")
				log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", bkpConfigResponse.Create.Meta.Name, bkpConfigResponse.Create.Meta.Uid)
				log.Infof("All deployments - [%+v]", workflowDataservice.DataServiceDeployment)
			})

			Step("Get the latest backup detail for the deployment", func() {
				workflowBackup.WorkflowDataService = workflowDataservice
				log.Infof("All deployments - [%+v]", workflowDataservice.DataServiceDeployment)
				backupResponse, err := workflowBackup.GetLatestBackup(*deployment.Create.Meta.Name)
				log.FailOnError(err, "Error occured while creating backup")
				latestBackupUid = *backupResponse.Meta.Uid
				log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
				err = workflowBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
				log.FailOnError(err, "Error occured while waiting for backup to complete")
			})

			Step("Register Destination Target Cluster", func() {
				err := SetDestinationKubeConfig()
				if err != nil {
					log.Infof("Failed to switched to destination cluster")
				}
				destinationCluster.Project = WorkflowProject
				log.Infof("Tenant ID [%s]", destinationCluster.Project.Platform.TenantId)
				_, err = destinationCluster.RegisterToControlPlane(false)
				log.FailOnError(err, "Unable to register target cluster")
				log.Infof("Destination Target cluster registered with uid - [%s]", destinationCluster.ClusterUID)
			})

			Step("Create a new namespace for restore", func() {
				destinationNamespace.TargetCluster = destinationCluster
				destinationNamespace.Namespaces = make(map[string]string)
				_, err := destinationNamespace.CreateNamespaces(restoreNamespace)
				log.FailOnError(err, "Unable to create namespace")
				log.Infof("Namespaces created - [%s]", WorkflowNamespace.Namespaces)
			})

			Step("Associate target cluster and restore namespace to Project", func() {
				err := WorkflowProject.Associate(
					[]string{destinationCluster.ClusterUID},
					[]string{destinationNamespace.Namespaces[restoreNamespace]},
					[]string{},
					[]string{},
					[]string{},
					[]string{},
				)
				log.FailOnError(err, "Unable to associate Cluster to Project")
				log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)
			})

			Step("Create Restore from the latest backup Id", func() {
				restoreName := "testing_restore_" + RandomString(5)
				workflowRestore.Destination = destinationNamespace
				workflowRestore.WorkflowProject = WorkflowProject
				_, err := workflowRestore.CreateRestore(restoreName, latestBackupUid, restoreNamespace)
				log.FailOnError(err, "Restore Failed")

				log.Infof("Restore created successfully with ID - [%s]", workflowRestore.Restores[restoreName].Meta.Uid)
			})
		}

		JustAfterEach(func() {
			defer func() {
				err := SetSourceKubeConfig()
				log.FailOnError(err, "Unable to switch context to source cluster [%s]", SourceClusterName)
			}()

			log.Infof("Cleaning up all resources")
			err := workflowBackup.Purge(*deployment.Create.Meta.Name)
			log.FailOnError(err, "Backup cleanup failed")
			err = workflowBackUpConfig.Purge(true)
			log.FailOnError(err, "Backup Configs cleanup failed")
			err = workflowDataservice.DeleteDeployment(*deployment.Create.Meta.Name)
			log.FailOnError(err, "Data Service cleanup failed")
			err = destinationNamespace.Purge()
			log.FailOnError(err, "Destination namespace cleanup failed")

			defer EndTorpedoTest()

		})

	})
})

var _ = Describe("{PerformRestoreToDifferentClusterProject}", func() {
	var (
		workflowDataservice  pds.WorkflowDataService
		workflowBackUpConfig pds.WorkflowPDSBackupConfig
		workflowRestore      pds.WorkflowPDSRestore
		workflowBackup       pds.WorkflowPDSBackup
		workFlowTemplates    pds.WorkflowPDSTemplates
		deployment           *automationModels.PDSDeploymentResponse
		destinationProject   platform.WorkflowProject
		destinationCluster   platform.WorkflowTargetCluster
		destinationNamespace platform.WorkflowNamespace
		latestBackupUid      string
		pdsBackupConfigName  string
		restoreNamespace     string
	)

	JustBeforeEach(func() {
		StartTorpedoTest("PerformRestoreToDifferentClusterProject", "Deploy data services and perform backup and restore on the different cluster from different project", nil, 0)
		workflowDataservice.DataServiceDeployment = make(map[string]string)

		workflowRestore.Destination = WorkflowNamespace
		workflowRestore.WorkflowProject = WorkflowProject
		workflowDataservice.Dash = dash
		restoreNamespace = "pds-restore-namespace-" + RandomString(5)
		destinationProject.Platform = WorkflowPlatform
	})

	It("Deploy data services and perform backup and restore on the different cluster", func() {

		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy dataservice", func() {

				workFlowTemplates.Platform = WorkflowPlatform

				workflowDataservice.Namespace = WorkflowNamespace
				workflowDataservice.NamespaceName = PDS_DEFAULT_NAMESPACE

				serviceConfigId, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams, ds.Name)
				log.FailOnError(err, "Unable to create Custom Templates for PDS")

				workflowDataservice.PDSTemplates.ServiceConfigTemplateId = serviceConfigId
				workflowDataservice.PDSTemplates.StorageTemplateId = stConfigId
				workflowDataservice.PDSTemplates.ResourceTemplateId = resConfigId

				deployment, err = workflowDataservice.DeployDataService(ds, ds.Image, ds.Version)
				log.FailOnError(err, "Error while deploying ds")
				log.Infof("All deployments - [%+v]", workflowDataservice.DataServiceDeployment)

				//stepLog := "Running Workloads on deployment"
				//Step(stepLog, func() {
				//	err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
				//	log.FailOnError(err, "Error while running workloads on ds")
				//})
			})

			Step("Create Adhoc backup config of the existing deployment", func() {
				workflowBackUpConfig.WorkflowDataService = workflowDataservice
				workflowBackUpConfig.WorkflowBackupLocation = WorkflowbkpLoc
				pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
				workflowBackUpConfig.Backups = make(map[string]automationModels.V1BackupConfig)
				bkpConfigResponse, err := workflowBackUpConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Name)
				log.FailOnError(err, "Error occured while creating backupConfig")
				log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", bkpConfigResponse.Create.Meta.Name, bkpConfigResponse.Create.Meta.Uid)
				log.Infof("All deployments - [%+v]", workflowDataservice.DataServiceDeployment)
			})

			Step("Get the latest backup detail for the deployment", func() {
				workflowBackup.WorkflowDataService = workflowDataservice
				log.Infof("All deployments - [%+v]", workflowDataservice.DataServiceDeployment)
				backupResponse, err := workflowBackup.GetLatestBackup(*deployment.Create.Meta.Name)
				log.FailOnError(err, "Error occured while creating backup")
				latestBackupUid = *backupResponse.Meta.Uid
				log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
				err = workflowBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
				log.FailOnError(err, "Error occured while waiting for backup to complete")
			})

			Step("Create Project", func() {
				destinationProject.ProjectName = fmt.Sprintf("project-destination-%s", utilities.RandomString(5))
				_, err := destinationProject.CreateProject()
				log.FailOnError(err, "Unable to create project")
				log.Infof("Project created with ID - [%s]", destinationProject.ProjectId)
			})

			Step("Register Destination Target Cluster", func() {
				err := SetDestinationKubeConfig()
				if err != nil {
					log.Infof("Failed to switched to destination cluster")
				}
				destinationCluster.Project = destinationProject
				log.Infof("Tenant ID [%s]", destinationCluster.Project.Platform.TenantId)
				_, err = destinationCluster.RegisterToControlPlane(false)
				log.FailOnError(err, "Unable to register target cluster")
				log.Infof("Destination Target cluster registered with uid - [%s]", destinationCluster.ClusterUID)
			})

			Step("Create a new namespace for restore", func() {
				destinationNamespace.TargetCluster = destinationCluster
				destinationNamespace.Namespaces = make(map[string]string)
				_, err := destinationNamespace.CreateNamespaces(restoreNamespace)
				log.FailOnError(err, "Unable to create namespace")
				log.Infof("Namespaces created - [%s]", WorkflowNamespace.Namespaces)
			})

			Step("Associate target cluster and restore namespace to Project", func() {
				err := destinationProject.Associate(
					[]string{destinationCluster.ClusterUID},
					[]string{destinationNamespace.Namespaces[restoreNamespace]},
					[]string{WorkflowCc.CloudCredentials[NewPdsParams.BackUpAndRestore.TargetLocation].ID},
					[]string{WorkflowbkpLoc.BkpLocation.BkpLocationId},
					[]string{},
					[]string{},
				)
				log.FailOnError(err, "Unable to associate Cluster to Project")
				log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)
			})

			Step("Create Restore from the latest backup Id", func() {
				restoreName := "testing_restore_" + RandomString(5)
				workflowRestore.Destination = destinationNamespace
				workflowRestore.WorkflowProject = WorkflowProject
				_, err := workflowRestore.CreateRestore(restoreName, latestBackupUid, restoreNamespace)
				log.FailOnError(err, "Restore Failed")

				log.Infof("Restore created successfully with ID - [%s]", workflowRestore.Restores[restoreName].Meta.Uid)
			})
		}

		JustAfterEach(func() {
			defer func() {
				err := SetSourceKubeConfig()
				log.FailOnError(err, "Unable to switch context to source cluster [%s]", SourceClusterName)
			}()
			log.Infof("Cleaning up all resources")
			err := workflowBackup.Purge(*deployment.Create.Meta.Name)
			log.FailOnError(err, "Backup cleanup failed")
			err = workflowBackUpConfig.Purge(true)
			log.FailOnError(err, "Backup Configs cleanup failed")
			err = workflowDataservice.DeleteDeployment(*deployment.Create.Meta.Name)
			log.FailOnError(err, "Data Service cleanup failed")
			err = SetDestinationKubeConfig()
			log.FailOnError(err, "Unable to set kubeconfig to destination")
			err = destinationNamespace.Purge()
			log.FailOnError(err, "Destination namespace cleanup failed")
			err = destinationProject.DeleteProject()
			log.FailOnError(err, "Destination Project cleanup failed")
			defer EndTorpedoTest()
		})

	})
})

var _ = Describe("{UpgradeDataServiceImageAndVersionWithBackUpRestore}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("UpgradeDataServiceImageAndVersionWithBackUpRestore", "Upgrade Data Service Version and Image", nil, 0)
	})
	var (
		workflowDataservice  pds.WorkflowDataService
		workFlowTemplates    pds.WorkflowPDSTemplates
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

		defer func() {
			Step("Delete DataServiceDeployment", func() {
				err := workflowDataservice.DeleteDeployment(*deployment.Create.Meta.Uid)
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

		workFlowTemplates pds.WorkflowPDSTemplates
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
			serviceConfigId, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams, ds.Name)
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
				err := workflowDataService.DeleteDeployment(*deployment.Create.Meta.Uid)
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

		//Step("Perform Restore and validate", func() {
		//	workflowRestore.WorkflowDataService = workflowDataService
		//	backupUid := *bkpConfigResponse.Create.Meta.Uid
		//	deploymentName := *deployment.Create.Meta.Name
		//	cloudSnapId := ""
		//	// Set the DestClusterId same as the current ClusterId
		//	workflowRestore.Destination.DestinationClusterId = WorkflowTargetCluster.ClusterUID
		//
		//	log.FailOnError(err, "failed while registering destination target cluster")
		//
		//	workflowRestore.WorkflowBackupLocation = WorkflowbkpLoc
		//	restoreDeployment, err = workflowRestore.CreateRestore(backupUid, deploymentName, cloudSnapId)
		//	log.FailOnError(err, "Error while taking restore")
		//	log.Debugf("Restored DeploymentName: [%s]", restoreDeployment.Create.Meta.Name)
		//})

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

		//Step("Perform Restore and validate", func() {
		//	workflowRestore.WorkflowDataService = workflowDataService
		//	backupUid := *bkpConfigResponse.Create.Meta.Uid
		//	deploymentName := *deployment.Create.Meta.Name
		//	cloudSnapId := ""
		//	// Set the DestClusterId same as the current ClusterId
		//	workflowRestore.Destination.DestinationClusterId = WorkflowTargetCluster.ClusterUID
		//
		//	log.FailOnError(err, "failed while registering destination target cluster")
		//
		//	workflowRestore.WorkflowBackupLocation = WorkflowbkpLoc
		//	restoreDeployment, err = workflowRestore.CreateRestore(backupUid, deploymentName, cloudSnapId)
		//	log.FailOnError(err, "Error while taking restore")
		//	log.Debugf("Restored DeploymentName: [%s]", restoreDeployment.Create.Meta.Name)
		//})

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

var _ = Describe("{PerformRestoreAfterDataServiceUpdate}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("PerformRestoreAfterDataServiceUpdate", "Perform restore after ds update", nil, 0)
	})
	var (
		workflowDataservice  pds.WorkflowDataService
		workFlowTemplates    pds.WorkflowPDSTemplates
		workflowBackUpConfig pds.WorkflowPDSBackupConfig
		workflowBackup       pds.WorkflowPDSBackup
		deployment           *automationModels.PDSDeploymentResponse
		workflowRestore      pds.WorkflowPDSRestore
		restoreDeployment    *automationModels.PDSRestoreResponse
		pdsBackupConfigName  string
		latestBackupUid      string
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

	It("Perform adhoc backup, restore and validate them before upgrade", func() {
		workflowBackUpConfig.WorkflowDataService = workflowDataservice
		workflowBackUpConfig.WorkflowBackupLocation = WorkflowbkpLoc
		pdsBackupConfigName = strings.ToLower("pds-qa-bkpConfig-" + utilities.RandString(5))

		Step("Take Backup and validate", func() {
			bkpConfigResponse, err := workflowBackUpConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
			log.FailOnError(err, "Error occured while creating backupConfig")
			log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", bkpConfigResponse.Create.Meta.Name, bkpConfigResponse.Create.Meta.Uid)
		})

		defer func() {
			Step("Delete Backups", func() {
				err := workflowBackUpConfig.DeleteBackupConfig(pdsBackupConfigName)
				log.FailOnError(err, "Error while deleting BackupConfig [%s]", pdsBackupConfigName)
			})
		}()

		Step("Get the latest backup id", func() {
			backupResponse, err := workflowBackup.GetLatestBackup(*deployment.Create.Meta.Name)
			log.FailOnError(err, "Error occured while creating backup")
			latestBackupUid = *backupResponse.Meta.Uid
			log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
		})

		Step("Perform Restore on destination cluster and validate", func() {
			restoreName := "pds-restore-before-update-" + RandomString(5)
			workflowRestore.Destination = WorkflowNamespace
			restoreDeployment, err := workflowRestore.CreateRestore(restoreName, latestBackupUid, Namespace)
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
			err := workflowDataservice.ValidateDataServiceWorkloads(NewPdsParams, restoreDeployment)
			log.FailOnError(err, "Error occured in ValidateDataServiceWorkloads method")
		})

	})

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

	It("Perform adhoc backup, restore and validate them after upgrade", func() {
		workflowBackUpConfig.WorkflowDataService = workflowDataservice
		workflowBackUpConfig.WorkflowBackupLocation = WorkflowbkpLoc
		pdsBackupConfigName = strings.ToLower("pds-qa-bkpConfig-" + utilities.RandString(5))

		Step("Take Backup and validate", func() {
			bkpConfigResponse, err := workflowBackUpConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
			log.FailOnError(err, "Error occured while creating backupConfig")
			log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", bkpConfigResponse.Create.Meta.Name, bkpConfigResponse.Create.Meta.Uid)
		})

		defer func() {
			Step("Delete Backups", func() {
				err := workflowBackUpConfig.DeleteBackupConfig(pdsBackupConfigName)
				log.FailOnError(err, "Error while deleting BackupConfig [%s]", pdsBackupConfigName)
			})
		}()

		Step("Get the latest backup id", func() {
			backupResponse, err := workflowBackup.GetLatestBackup(*deployment.Create.Meta.Name)
			log.FailOnError(err, "Error occured while creating backup")
			latestBackupUid = *backupResponse.Meta.Uid
			log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
		})

		Step("Perform Restore on destination cluster and validate", func() {
			restoreName := "pds-restore-before-update-" + RandomString(5)
			workflowRestore.Destination = WorkflowNamespace
			restoreDeployment, err := workflowRestore.CreateRestore(restoreName, latestBackupUid, Namespace)
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
			err := workflowDataservice.ValidateDataServiceWorkloads(NewPdsParams, restoreDeployment)
			log.FailOnError(err, "Error occured in ValidateDataServiceWorkloads method")
		})

	})

	It("Delete DataServiceDeployment", func() {
		err := workflowDataservice.DeleteDeployment(*deployment.Create.Meta.Uid)
		log.FailOnError(err, "Error while deleting data Service")
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{PerformSimultaneousBackupRestoreForMultipleDeployments}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("PerformSimultaneousBackupRestoreForMultipleDeployments", "Perform multiple backup and restore simultaneously for different deployments.", nil, 0)
	})
	var (
		workflowDataservice  pds.WorkflowDataService
		workFlowTemplates    pds.WorkflowPDSTemplates
		workflowBackUpConfig pds.WorkflowPDSBackupConfig
		workflowBackup       pds.WorkflowPDSBackup
		deployment           *automationModels.PDSDeploymentResponse
		workflowRestore      pds.WorkflowPDSRestore
		restoreDeployment    *automationModels.PDSRestoreResponse
		pdsBackupConfigNames []string
		latestBackupUid      string
		numberOfIterations   int
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

	It("Perform adhoc backup, restore and validate - Multiple Backup and Restores", func() {
		workflowBackUpConfig.WorkflowDataService = workflowDataservice
		workflowBackUpConfig.WorkflowBackupLocation = WorkflowbkpLoc
		numberOfIterations = 10

		Step("Start Multiple Backup Simultaneously", func() {
			var wg sync.WaitGroup
			for i := 0; i < numberOfIterations; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()
					pdsBackupConfigName := strings.ToLower("pds-qa-bkpConfig-" + utilities.RandString(5))
					bkpConfigResponse, err := workflowBackUpConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
					log.FailOnError(err, "Error occured while creating backupConfig")
					log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", bkpConfigResponse.Create.Meta.Name, bkpConfigResponse.Create.Meta.Uid)
					pdsBackupConfigNames = append(pdsBackupConfigNames, pdsBackupConfigName)

				}()
			}
			wg.Wait()
			log.Infof("All backups are completed successfully")
		})

		Step("Trigger multiple restores simultaneously", func() {
			// TODO: Keeping restores as sequential, needs to be changed to parallel later as multiple ns will be required
			allBackups, err := workflowBackup.ListAllBackups(*deployment.Create.Meta.Name)
			log.FailOnError(err, "Error occured while creating backup")
			log.Infof("Number of backups - [%d]", len(allBackups))

			for _, eachBackup := range allBackups {
				latestBackupUid = *eachBackup.Meta.Uid
				log.Infof("Current backup ID [%s], Name [%s]", *eachBackup.Meta.Uid, *eachBackup.Meta.Name)
				restoreName := "pds-restore-before-update-" + RandomString(5)
				workflowRestore.Destination = WorkflowNamespace
				restoreDeployment, err := workflowRestore.CreateRestore(restoreName, latestBackupUid, Namespace)
				log.FailOnError(err, "Error while taking restore")
				log.Debugf("Restored DeploymentName: [%s]", restoreDeployment.Create.Meta.Name)
			}

		})

		defer func() {
			Step("Delete Backups", func() {
				for _, eachBackup := range pdsBackupConfigNames {
					err := workflowBackUpConfig.DeleteBackupConfig(eachBackup)
					log.FailOnError(err, "Error while deleting BackupConfig [%s]", eachBackup)
				}
			})
		}()

		defer func() {
			Step("Delete RestoredDeployment", func() {
				err := workflowRestore.DeleteRestore(*restoreDeployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while deleting restore")
			})
		}()

		Step("Validate md5hash for the restored deployments", func() {
			err := workflowDataservice.ValidateDataServiceWorkloads(NewPdsParams, restoreDeployment)
			log.FailOnError(err, "Error occured in ValidateDataServiceWorkloads method")
		})

	})

	It("Delete DataServiceDeployment", func() {
		err := workflowDataservice.DeleteDeployment(*deployment.Create.Meta.Uid)
		log.FailOnError(err, "Error while deleting data Service")
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
