package tests

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	"strings"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
)

var _ = Describe("{PerformRestoreToSameCluster}", func() {
	var (
		deployment          *automationModels.PDSDeploymentResponse
		latestBackupUid     string
		pdsBackupConfigName string
		restoreNamespace    string
		restoreName         string
		err                 error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("PerformRestoreToSameCluster", "Deploy data services and perform backup and restore on the same cluster", nil, 0)
		restoreNamespace = "restore-" + RandomString(5)
		restoreName = "restore-" + RandomString(5)

	})

	It("Deploy data services and perform backup and restore on the same cluster", func() {

		for _, ds := range NewPdsParams.DataServiceToTest {

			Step("Deploy dataservice", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version)
				log.FailOnError(err, "Error while deploying ds")
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)

			})

			Step("Create Adhoc backup config of the existing deployment", func() {
				pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
				bkpConfigResponse, err := WorkflowPDSBackupConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backupConfig")
				log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)
			})

			Step("Get the latest backup detail for the deployment", func() {
				backupResponse, err := WorkflowPDSBackup.GetLatestBackup(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backup")
				latestBackupUid = *backupResponse.Meta.Uid
				log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
				err = WorkflowPDSBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
				log.FailOnError(err, "Error occured while waiting for backup to complete")
			})

			Step("Create Restore from the latest backup Id", func() {
				defer func() {
					err := SetSourceKubeConfig()
					log.FailOnError(err, "failed to switch context to source cluster")
				}()
				CheckforClusterSwitch()
				WorkflowPDSRestore.SourceNamespace = WorkflowDataService.NamespaceName
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUid, restoreNamespace)
				log.FailOnError(err, "Restore Failed")
				log.Infof("All restores - [%+v]", WorkflowPDSRestore.Restores)
				log.Infof("Restore Created Name - [%s], UID - [%s]", *WorkflowPDSRestore.Restores[restoreName].Meta.Name, *WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
			})
		}

	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})

var _ = Describe("{PerformRestoreToDifferentClusterSameProject}", func() {
	var (
		deployment          *automationModels.PDSDeploymentResponse
		latestBackupUid     string
		pdsBackupConfigName string
		restoreNamespace    string
		restoreName         string
		err                 error
	)
	JustBeforeEach(func() {
		StartPDSTorpedoTest("PerformRestoreToDifferentClusterSameProject", "Deploy data services and perform backup and restore on a different cluster on the same project", nil, 0)

		restoreNamespace = "restore-" + RandomString(5)
		restoreName = "restore-" + RandomString(5)
	})

	It("Deploy data services and perform backup and restore on the different cluster", func() {

		for _, ds := range NewPdsParams.DataServiceToTest {

			Step("Deploy dataservice", func() {

				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version)
				log.FailOnError(err, "Error while deploying ds")
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)

			})

			Step("Create Adhoc backup config of the existing deployment", func() {
				pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
				bkpConfigResponse, err := WorkflowPDSBackupConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backupConfig")
				log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
			})

			Step("Get the latest backup detail for the deployment", func() {
				backupResponse, err := WorkflowPDSBackup.GetLatestBackup(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backup")
				latestBackupUid = *backupResponse.Meta.Uid
				log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
				err = WorkflowPDSBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
				log.FailOnError(err, "Error occured while waiting for backup to complete")
			})

			Step("Create Restore from the latest backup Id", func() {
				WorkflowPDSRestore.Destination = &WorkflowNamespaceDestination
				CheckforClusterSwitch()
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUid, restoreNamespace)
				log.FailOnError(err, "Restore Failed")

				log.Infof("Restore created successfully with ID - [%s]", *WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
			})
		}

	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()

	})
})

var _ = Describe("{PerformRestoreToDifferentClusterProject}", func() {
	var (
		destinationProject  platform.WorkflowProject
		latestBackupUid     string
		pdsBackupConfigName string
		restoreNamespace    string
		restoreName         string
		deployment          *automationModels.PDSDeploymentResponse
		err                 error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("PerformRestoreToDifferentClusterProject", "Deploy data services and perform backup and restore on the different cluster from different project", nil, 0)
		restoreNamespace = "namespace-" + RandomString(5)
		destinationProject.Platform = WorkflowPlatform
		restoreName = "restore-" + RandomString(5)
	})

	It("Deploy data services and perform backup and restore on the different cluster", func() {

		for _, ds := range NewPdsParams.DataServiceToTest {

			Step("Deploy dataservice", func() {

				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version)
				log.FailOnError(err, "Error while deploying ds")
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)

			})

			Step("Create Adhoc backup config of the existing deployment", func() {
				pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
				bkpConfigResponse, err := WorkflowPDSBackupConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backupConfig")
				log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
			})

			Step("Get the latest backup detail for the deployment", func() {
				backupResponse, err := WorkflowPDSBackup.GetLatestBackup(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backup")
				latestBackupUid = *backupResponse.Meta.Uid
				log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
				err = WorkflowPDSBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
				log.FailOnError(err, "Error occured while waiting for backup to complete")
			})

			Step("Create Project", func() {
				destinationProject.ProjectName = fmt.Sprintf("project-destination-%s", utilities.RandomString(5))
				_, err := destinationProject.CreateProject()
				log.FailOnError(err, "Unable to create project")
				log.Infof("Project created with ID - [%s]", destinationProject.ProjectId)
			})

			Step("Associate target cluster and restore namespace to Project", func() {
				err := destinationProject.Associate(
					[]string{WorkflowTargetClusterDestination.ClusterUID},
					[]string{},
					[]string{WorkflowCc.CloudCredentials[NewPdsParams.BackUpAndRestore.TargetLocation].ID},
					[]string{WorkflowbkpLoc.BkpLocation.BkpLocationId},
					[]string{},
					[]string{},
				)
				log.FailOnError(err, "Unable to associate Cluster to Project")
				log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)
			})

			Step("Create Restore from the latest backup Id", func() {
				WorkflowPDSRestore.Destination = &WorkflowNamespaceDestination
				CheckforClusterSwitch()
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUid, restoreNamespace)
				log.FailOnError(err, "Restore Failed")
				log.Infof("Restore created successfully with ID - [%s]", WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
			})
		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})

})

var _ = Describe("{PerformSimultaneousRestoresDifferentDataService}", func() {
	var (
		deployments          []*automationModels.PDSDeploymentResponse
		latestBackupUid      string
		pdsBackupConfigName  string
		restoreNames         []string
		deploymentNamespace  string
		allBackupIds         []string
		BackupsPerDeployment int
		allErrors            []error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("PerformSimultaneousRestoresDifferentDataService", "Perform multiple backup and restore simultaneously for different dataservices.", nil, 0)
		restoreNames = make([]string, 0)
		deployments = make([]*automationModels.PDSDeploymentResponse, 0)
		allBackupIds = make([]string, 0)
		BackupsPerDeployment = 1
	})

	It("Perform multiple backup and restore simultaneously for different dataservices", func() {

		for _, ds := range NewPdsParams.DataServiceToTest {

			Step("Create a namespace for PDS", func() {
				deploymentNamespace = fmt.Sprintf("%s-%s", strings.ToLower(ds.Name), RandomString(5))
				_, err := WorkflowNamespace.CreateNamespaces(deploymentNamespace)
				log.FailOnError(err, "Unable to create namespace")
				log.Infof("Namespaces created - [%s]", WorkflowNamespace.Namespaces)
			})

			Step("Associate namespace and cluster to Project", func() {
				err := WorkflowProject.Associate(
					[]string{},
					[]string{WorkflowNamespace.Namespaces[deploymentNamespace]},
					[]string{},
					[]string{},
					[]string{},
					[]string{},
				)
				log.FailOnError(err, "Unable to associate Cluster to Project")
				log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)
			})

			Step("Deploy multiple dataservice", func() {

				currDeployment, err := WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version)
				log.FailOnError(err, "Error while deploying ds")
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)
				deployments = append(deployments, currDeployment)

				//stepLog := "Running Workloads on deployment"
				//Step(stepLog, func() {
				//	err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
				//	log.FailOnError(err, "Error while running workloads on ds")
				//})
			})
		}

		Step("Create multiple Adhoc backup config for the existing deployment", func() {
			var wg sync.WaitGroup

			for _, deployment := range deployments {
				for i := 0; i < BackupsPerDeployment; i++ {
					wg.Add(1)
					go func() {

						defer wg.Done()
						defer GinkgoRecover()

						pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
						bkpConfigResponse, err := WorkflowPDSBackupConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
						if err != nil {
							log.Errorf("Some error occurred while creating backup [%s], Error - [%s]", pdsBackupConfigName, err.Error())
							allErrors = append(allErrors, err)
						}
						log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
					}()
				}
			}

			wg.Wait()
			dash.VerifyFatal(len(allErrors), 0, "Some error occurred while creating backup configs")
			log.InfoD("Simultaneous backup config creation succeeded")
		})

		Step("Get the backup detail for the backup configs", func() {
			for _, deployment := range deployments {
				allBackupResponse, err := WorkflowPDSBackup.ListAllBackups(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backup")
				dash.VerifyFatal(len(allBackupResponse), BackupsPerDeployment, fmt.Sprintf("Total number of backups found for [%s] are not consisten with backup configs created.", *deployment.Create.Meta.Name))
				for _, backupResponse := range allBackupResponse {
					latestBackupUid = *backupResponse.Meta.Uid
					log.Infof("Backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
					err = WorkflowPDSBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
					log.FailOnError(err, "Error occured while waiting for backup to complete")
					allBackupIds = append(allBackupIds, latestBackupUid)
				}
			}

			log.InfoD("Simultaneous backups creation succeeded")
		})

		Step("Creating Simultaneous restores from the dataservices", func() {
			var wg sync.WaitGroup

			for _, backupId := range allBackupIds {

				wg.Add(1)

				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					restoreName := "restore-" + RandomString(5)
					_, err := WorkflowPDSRestore.CreateRestore(restoreName, backupId, restoreName)
					if err != nil {
						log.Errorf("Error occurred while creating [%s], Error - [%s]", restoreName, err.Error())
					}
					log.Infof("Restore created successfully with ID - [%s]", WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
					restoreNames = append(restoreNames, restoreName)
				}()

			}

			wg.Wait()
			dash.VerifyFatal(len(allErrors), 0, "Some error occurred while creating restores")
			log.InfoD("Simultaneous restores succeeded")
		})
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
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
		//workflowRestore      pds.WorkflowPDSRestore
		deployment        *automationModels.PDSDeploymentResponse
		updatedDeployment *automationModels.PDSDeploymentResponse
		//restoreDeployment    *automationModels.PDSRestoreResponse
		bkpConfigResponse   *automationModels.PDSBackupConfigResponse
		pdsBackupConfigName string
		oldDeploymentId     string
		newDeploymentId     string
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

			_, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams)
			log.FailOnError(err, "Unable to create Custom Templates for PDS")
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

		//stepLog := "Running Workloads before upgrading the ds image"
		//Step(stepLog, func() {
		//	err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
		//	log.FailOnError(err, "Error while running workloads on ds")
		//})
	})

	It("Perform adhoc backup of old deployments", func() {
		workflowBackUpConfig.WorkflowDataService = &workflowDataservice
		workflowBackUpConfig.WorkflowBackupLocation = WorkflowbkpLoc
		pdsBackupConfigName = strings.ToLower("pds-qa-bkpConfig-" + utilities.RandString(5))

		Step("Take Backup and validate", func() {
			bkpConfigResponse, err = workflowBackUpConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
			log.FailOnError(err, "Error occured while creating backupConfig")
			log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
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

		//stepLog := "Running Workloads after upgrading the ds image"
		//Step(stepLog, func() {
		//	err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
		//	log.FailOnError(err, "Error while running workloads on ds")
		//})
	})

	It("Restore the old deployment and upgrade the restored deployment", func() {

		//defer func() {
		//	Step("Delete RestoredDeployment", func() {
		//		err := workflowRestore.DeleteRestore(*restoreDeployment.Create.Meta.Uid)
		//		log.FailOnError(err, "Error while deleting restore")
		//	})
		//}()

		Step("Update restored DataService Version and Image", func() {
			for _, ds := range NewPdsParams.DataServiceToTest {
				updatedRestoredDeployment, err := workflowDataservice.UpdateDataService(ds, newDeploymentId, ds.Image, ds.Version)
				log.FailOnError(err, "Error while updating ds")
				log.Debugf("Updated Deployment Id [%s]", *updatedRestoredDeployment.Update.Meta.Uid)
			}

			//stepLog := "Running Workloads after upgrading the ds image"
			//Step(stepLog, func() {
			//	err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
			//	log.FailOnError(err, "Error while running workloads on ds")
			//})
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
		//workflowRestore      pds.WorkflowPDSRestore
		deployment *automationModels.PDSDeploymentResponse
		//	restoreDeployment *automationModels.PDSRestoreResponse

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
			workflowDataService.Namespace = &WorkflowNamespace
			workflowDataService.NamespaceName = Namespace
			serviceConfigId, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams)
			log.FailOnError(err, "Unable to create Custom Templates for PDS")
			workflowDataService.PDSTemplates.StorageTemplateId = stConfigId
			workflowDataService.PDSTemplates.ResourceTemplateId = resConfigId
			tempList = append(tempList, serviceConfigId[ds.Name], stConfigId, resConfigId)
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
		//stepLog := "Running Workloads before taking backups"
		//Step(stepLog, func() {
		//	err := workflowDataService.RunDataServiceWorkloads(NewPdsParams)
		//	log.FailOnError(err, "Error while running workloads on ds")
		//})
	})
	It("Perform adhoc backup, restore before PVC Resize and validate them", func() {
		var bkpConfigResponse *automationModels.PDSBackupConfigResponse
		workflowBackUpConfig.WorkflowDataService = &workflowDataService
		pdsBackupConfigName = strings.ToLower("pds-qa-bkpConfig-" + utilities.RandString(5))

		Step("Take Backup and validate", func() {
			bkpConfigResponse, err = workflowBackUpConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
			log.FailOnError(err, "Error occured while creating backupConfig")
			log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
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

		//defer func() {
		//	Step("Delete RestoredDeployment", func() {
		//		err := workflowRestore.DeleteRestore(*restoreDeployment.Create.Meta.Uid)
		//		log.FailOnError(err, "Error while deleting restore")
		//	})
		//}()

		//Step("Validate md5hash for the restored deployments", func() {
		//	err := workflowDataService.ValidateDataServiceWorkloads(NewPdsParams, restoreDeployment)
		//	log.FailOnError(err, "Error occured in ValidateDataServiceWorkloads method")
		//})

	})

	It("Increase PVC Size by 1 GB of DataService from K8s", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			log.InfoD("Dataservice on which the PVC needs to be resized is- [%v]", ds.Name)
			err = workflowDataService.IncreasePvcSizeBy1gb(workflowDataService.NamespaceName, *deployment.Create.Status.CustomResourceName, 1)
			log.FailOnError(err, "Failing while Increasing the PVC name...")

		}
		stepLog := "Validate the deployment after PVC Resize"
		Step(stepLog, func() {
			//Validate deployment function call here
		})
		//stepLog = "Running Workloads after Resize of PVC"
		//Step(stepLog, func() {
		//	err := workflowDataService.RunDataServiceWorkloads(NewPdsParams)
		//	log.FailOnError(err, "Error while running workloads on ds")
		//})
	})
	It("Perform adhoc backup, restore after PVC Resize and validate them", func() {
		var bkpConfigResponse *automationModels.PDSBackupConfigResponse
		workflowBackUpConfig.WorkflowDataService = &workflowDataService
		pdsBackupConfigName = strings.ToLower("pds-qa-bkpConfig-" + utilities.RandString(5))

		Step("Take Backup and validate", func() {
			bkpConfigResponse, err = workflowBackUpConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
			log.FailOnError(err, "Error occured while creating backupConfig")
			log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
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

		//defer func() {
		//	Step("Delete RestoredDeployment", func() {
		//		err := workflowRestore.DeleteRestore(*restoreDeployment.Create.Meta.Uid)
		//		log.FailOnError(err, "Error while deleting restore")
		//	})
		//}()

		//Step("Validate md5hash for the restored deployments", func() {
		//	err := workflowDataService.ValidateDataServiceWorkloads(NewPdsParams, restoreDeployment)
		//	log.FailOnError(err, "Error occured in ValidateDataServiceWorkloads method")
		//})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{PerformRestoreAfterDataServiceUpdate}", func() {
	var (
		deployment            *automationModels.PDSDeploymentResponse
		latestBackupUid       string
		pdsBackupConfigName   string
		restoreName           string
		err                   error
		backupIdBeforeUpgrade string
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("PerformRestoreAfterDataServiceUpdate", "Perform restore after ds update", nil, 0)

	})

	It("Deploy data services and perform backup and restore on the same cluster", func() {

		for _, ds := range NewPdsParams.DataServiceToTest {

			Step("Deploy dataservice", func() {

				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version)
				log.FailOnError(err, "Error while deploying ds")
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)

			})

			Step("Create Adhoc backup config of the existing deployment", func() {
				pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
				bkpConfigResponse, err := WorkflowPDSBackupConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backupConfig")
				log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)
			})

			Step("Get the latest backup detail for the deployment", func() {
				backupResponse, err := WorkflowPDSBackup.GetLatestBackup(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backup")
				latestBackupUid = *backupResponse.Meta.Uid
				log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
				err = WorkflowPDSBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
				backupIdBeforeUpgrade = *backupResponse.Meta.Uid
				log.FailOnError(err, "Error occured while waiting for backup to complete")
			})

			Step("Create Restore from the latest backup Id", func() {
				defer func() {
					err := SetSourceKubeConfig()
					log.FailOnError(err, "failed to switch context to source cluster")
				}()
				restoreName = "restore-bu-" + RandomString(5)
				CheckforClusterSwitch()
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUid, restoreName)
				log.FailOnError(err, "Restore Failed")
				log.Infof("All restores - [%+v]", WorkflowPDSRestore.Restores)
				log.Infof("Restore Created Name - [%s], UID - [%s]", *WorkflowPDSRestore.Restores[restoreName].Meta.Name, *WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
			})

			Step("Upgrade DataService Version and Image", func() {
				_, err := WorkflowDataService.UpdateDataService(ds, *deployment.Create.Meta.Uid, ds.Image, ds.Version)
				log.FailOnError(err, "Error while updating ds")
			})

			Step("Create Adhoc backup config of the existing deployment after upgrade", func() {
				pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
				bkpConfigResponse, err := WorkflowPDSBackupConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backupConfig")
				log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)
			})

			Step("Get the latest backup detail for the deployment after upgrade", func() {
				backupResponse, err := WorkflowPDSBackup.GetLatestBackup(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backup")
				latestBackupUid = *backupResponse.Meta.Uid
				log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
				err = WorkflowPDSBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
				log.FailOnError(err, "Error occured while waiting for backup to complete")
			})

			Step("Create Restore from the latest backup Id after upgrade", func() {
				defer func() {
					err := SetSourceKubeConfig()
					log.FailOnError(err, "failed to switch context to source cluster")
				}()
				restoreName = "restore-au-" + RandomString(5)
				CheckforClusterSwitch()
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUid, restoreName)
				log.FailOnError(err, "Restore Failed")
				log.Infof("All restores - [%+v]", WorkflowPDSRestore.Restores)
				log.Infof("Restore Created Name - [%s], UID - [%s]", *WorkflowPDSRestore.Restores[restoreName].Meta.Name, *WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
			})

			Step("Create Restore from the backup Ids before upgrade", func() {
				defer func() {
					err := SetSourceKubeConfig()
					log.FailOnError(err, "failed to switch context to source cluster")
				}()
				restoreName = "restore-aubi-" + RandomString(5)
				CheckforClusterSwitch()
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, backupIdBeforeUpgrade, restoreName)
				log.FailOnError(err, "Restore Failed")
				log.Infof("All restores - [%+v]", WorkflowPDSRestore.Restores)
				log.Infof("Restore Created Name - [%s], UID - [%s]", *WorkflowPDSRestore.Restores[restoreName].Meta.Name, *WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
			})

		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
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
		//	restoreDeployment    *automationModels.PDSRestoreResponse
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
			workflowDataservice.Namespace = &WorkflowNamespace
			workflowDataservice.NamespaceName = Namespace

			_, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams)
			log.FailOnError(err, "Unable to create Custom Templates for PDS")
			workflowDataservice.PDSTemplates.StorageTemplateId = stConfigId
			workflowDataservice.PDSTemplates.ResourceTemplateId = resConfigId

			deployment, err = workflowDataservice.DeployDataService(ds, ds.OldImage, ds.OldVersion)
			log.FailOnError(err, "Error while deploying ds")
		}

		//stepLog := "Running Workloads before upgrading the ds image"
		//Step(stepLog, func() {
		//	err := workflowDataservice.RunDataServiceWorkloads(NewPdsParams)
		//	log.FailOnError(err, "Error while running workloads on ds")
		//})
	})

	It("Perform adhoc backup, restore and validate - Multiple Backup and Restores", func() {
		workflowBackUpConfig.WorkflowDataService = &workflowDataservice
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
					log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
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
				workflowRestore.Destination = &WorkflowNamespace
				restoreDeployment, err := workflowRestore.CreateRestore(restoreName, latestBackupUid, Namespace)
				log.FailOnError(err, "Error while taking restore")
				log.Debugf("Restored DeploymentName: [%s]", restoreDeployment.Create.Meta.Name)
			}

		})

		//Step("Validate md5hash for the restored deployments", func() {
		//	err := workflowDataservice.ValidateDataServiceWorkloads(NewPdsParams, restoreDeployment)
		//	log.FailOnError(err, "Error occured in ValidateDataServiceWorkloads method")
		//})

	})

	It("Delete DataServiceDeployment", func() {
		err := workflowDataservice.DeleteDeployment(*deployment.Create.Meta.Uid)
		log.FailOnError(err, "Error while deleting data Service")
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
