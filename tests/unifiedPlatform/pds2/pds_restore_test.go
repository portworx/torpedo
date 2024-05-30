package tests

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"

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

	})

	It("Deploy data services and perform backup and restore on the same cluster", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {

			Step("Deploy dataservice", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
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
				restoreNamespace = "restore-" + RandomString(5)
				restoreName = "restore-" + RandomString(5)
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUid, restoreNamespace, *deployment.Create.Meta.Uid)
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
	})

	It("Deploy data services and perform backup and restore on the different cluster", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {

			Step("Deploy dataservice", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
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
				restoreNamespace = "restore-" + RandomString(5)
				restoreName = "restore-" + RandomString(5)
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUid, restoreNamespace, *deployment.Create.Meta.Uid)
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
		destinationProject.Platform = WorkflowPlatform
		WorkflowPDSRestore.Destination = &WorkflowNamespaceDestination
	})

	It("Deploy data services and perform backup and restore on the different cluster", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {

			Step("Deploy dataservice", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
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
				WorkflowTargetClusterDestination.Project = &destinationProject
			})

			Step("Associate target cluster and resources to Project", func() {
				err := destinationProject.Associate(
					[]string{WorkflowTargetClusterDestination.ClusterUID},
					[]string{},
					[]string{WorkflowCc.CloudCredentials[NewPdsParams.BackUpAndRestore.TargetLocation].ID},
					[]string{WorkflowbkpLoc.BkpLocation.BkpLocationId},
					TemplateIds,
					[]string{},
				)
				log.FailOnError(err, "Unable to associate Cluster to Project")
				log.Infof("Associated Resources - [%+v]", destinationProject.AssociatedResources)
			})

			Step("Create Restore from the latest backup Id", func() {
				restoreNamespace = "namespace-" + RandomString(5)
				restoreName = "restore-" + RandomString(5)
				CheckforClusterSwitch()
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUid, restoreNamespace, *deployment.Create.Meta.Uid)
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
		pdsBackupConfigName  string
		restoreNames         []string
		deploymentNamespace  string
		allBackupIds         map[string][]string
		BackupsPerDeployment int
		allErrors            []error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("PerformSimultaneousRestoresDifferentDataService", "Perform multiple backup and restore simultaneously for different dataservices.", nil, 0)
		restoreNames = make([]string, 0)
		deployments = make([]*automationModels.PDSDeploymentResponse, 0)
		allBackupIds = make(map[string][]string)
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
				currDeployment, err := WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, deploymentNamespace)
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
					go func(dep *automationModels.PDSDeploymentResponse) {

						defer wg.Done()
						defer GinkgoRecover()

						pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
						bkpConfigResponse, err := WorkflowPDSBackupConfig.CreateBackupConfig(pdsBackupConfigName, *dep.Create.Meta.Uid)
						if err != nil {
							log.Errorf("Some error occurred while creating backup [%s], Error - [%s]", pdsBackupConfigName, err.Error())
							allErrors = append(allErrors, err)
						}
						log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
					}(deployment)
				}
			}

			wg.Wait()
			dash.VerifyFatal(len(allErrors), 0, "Verifying multiple backup creation")
			log.InfoD("Simultaneous backup config creation succeeded")
		})

		Step("Get the backup detail for the backup configs", func() {
			for _, deployment := range deployments {
				allBackupResponse, err := WorkflowPDSBackup.ListAllBackups(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backup")
				dash.VerifyFatal(len(allBackupResponse), BackupsPerDeployment, fmt.Sprintf("Total number of backups found for [%s] are not consisten with backup configs created.", *deployment.Create.Meta.Name))
				for _, backupResponse := range allBackupResponse {
					log.Infof("Backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
					err = WorkflowPDSBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
					log.FailOnError(err, "Error occured while waiting for backup to complete")
					allBackupIds[*deployment.Create.Meta.Uid] = append(allBackupIds[WorkflowDataService.DataServiceDeployment[*deployment.Create.Meta.Uid].Namespace], *backupResponse.Meta.Uid)
				}
			}

			log.InfoD("Simultaneous backups creation succeeded")
		})

		Step("Creating Simultaneous restores from the dataservices", func() {
			var wg sync.WaitGroup

			for depId, backupIds := range allBackupIds {

				for _, backupId := range backupIds {
					wg.Add(1)

					go func(deploymentId string, backup string) {
						defer wg.Done()
						defer GinkgoRecover()

						restoreName := "restore-" + RandomString(5)
						_, err := WorkflowPDSRestore.CreateRestore(restoreName, backup, restoreName, deploymentId)
						if err != nil {
							log.Errorf("Error occurred while creating [%s], Error - [%s]", restoreName, err.Error())
						} else {
							log.Infof("Restore created successfully with ID - [%s]", WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
							restoreNames = append(restoreNames, restoreName)
						}
					}(depId, backupId)
				}

			}

			wg.Wait()
			dash.VerifyFatal(len(allErrors), 0, "Verifying multiple restore creation")
			log.InfoD("Simultaneous restores succeeded")
		})
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})

})

var _ = Describe("{UpgradeDataServiceImageAndScaleUpDsWithBackUpRestore}", func() {
	var (
		deployment            *automationModels.PDSDeploymentResponse
		latestBackupUid       string
		pdsBackupConfigName   string
		restoreNamespace      string
		restoreName           string
		err                   error
		backupIdBeforeUpgrade string
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("UpgradeDataServiceImageAndScaleUpDsWithBackUpRestore", "Upgrade Data Service Image and ScaleUp Ds Replicas", nil, 0)
		restoreNamespace = "restore-" + RandomString(5)
		restoreName = "restore-" + RandomString(5)
	})

	It("Deploy data services and perform backup and restore on the same cluster", func() {

		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy dataservice", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.OldImage, ds.OldVersion, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)
				WorkflowPDSRestore.SourceDeploymentConfigBeforeUpgrade = &deployment.Create.Config.DeploymentTopologies[0]
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
				backupIdBeforeUpgrade = *backupResponse.Meta.Uid
				log.FailOnError(err, "Error occured while waiting for backup to complete")
			})

			Step("Create Restore from the latest backup Id", func() {
				defer func() {
					err := SetSourceKubeConfig()
					log.FailOnError(err, "failed to switch context to source cluster")
				}()
				CheckforClusterSwitch()
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUid, restoreNamespace, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Restore Failed")
				log.Infof("All restores - [%+v]", WorkflowPDSRestore.Restores)
				log.Infof("Restore Created Name - [%s], UID - [%s]", *WorkflowPDSRestore.Restores[restoreName].Meta.Name, *WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
			})

			Step("Upgrade DataService Image", func() {
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

			Step("Create Restore from the backup Ids before upgrade", func() {
				defer func() {
					err := SetSourceKubeConfig()
					log.FailOnError(err, "failed to switch context to source cluster")
				}()
				restoreName = "restr-old-bkp-" + RandomString(5)
				CheckforClusterSwitch()
				WorkflowPDSRestore.Validatation["VALIDATE_RESTORE_AFTER_SRC_DEPLOYMENT_UPGRADE"] = true
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, backupIdBeforeUpgrade, restoreNamespace, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Restore Failed")
				log.Infof("All restores - [%+v]", WorkflowPDSRestore.Restores)
				log.Infof("Restore Created Name - [%s], UID - [%s]", *WorkflowPDSRestore.Restores[restoreName].Meta.Name, *WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
				WorkflowPDSRestore.Validatation["VALIDATE_RESTORE_AFTER_SRC_DEPLOYMENT_UPGRADE"] = false
				delete(WorkflowPDSRestore.Validatation, "VALIDATE_RESTORE_AFTER_SRC_DEPLOYMENT_UPGRADE")
			})

			Step("Create Restore from the latest backup Id after upgrade", func() {
				defer func() {
					err := SetSourceKubeConfig()
					log.FailOnError(err, "failed to switch context to source cluster")
				}()
				restoreName = "restr-latest-bkp-" + RandomString(5)
				CheckforClusterSwitch()
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUid, restoreNamespace, *deployment.Create.Meta.Uid)
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

var _ = Describe("{PerformRestoreAfterPVCResize}", func() {
	var (
		deployment                       *automationModels.PDSDeploymentResponse
		latestBackupUidBeforePVCIncrease string
		latestBackupUidAfterPVCIncrease  string
		pdsBackupConfigName              string
		restoreNamespace                 string
		restoreName                      string
		err                              error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("PerformRestoreAfterPVCResize", "Perform PVC Resize and restore within same cluster.", nil, 0)

	})

	It("Perform PVC Resize and restore within same cluster.", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {

			Step("Deploy dataservice", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)

			})

			Step("Create Adhoc backup config of the existing deployment - Before PVC Increase", func() {
				pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
				bkpConfigResponse, err := WorkflowPDSBackupConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backupConfig")
				log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)
			})

			Step("Get the latest backup detail for the deployment - Before PVC Increase", func() {
				backupResponse, err := WorkflowPDSBackup.GetLatestBackup(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backup")
				latestBackupUidBeforePVCIncrease = *backupResponse.Meta.Uid
				log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
				err = WorkflowPDSBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
				log.FailOnError(err, "Error occured while waiting for backup to complete")
			})

			Step("Create Restore from the latest backup Id  - Before PVC Increase", func() {
				defer func() {
					err := SetSourceKubeConfig()
					log.FailOnError(err, "failed to switch context to source cluster")
				}()
				CheckforClusterSwitch()
				restoreNamespace = "restore-" + RandomString(5)
				restoreName = "restore-" + RandomString(5)
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUidBeforePVCIncrease, restoreNamespace, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Restore Failed")
				log.Infof("All restores - [%+v]", WorkflowPDSRestore.Restores)
				log.Infof("Restore Created Name - [%s], UID - [%s]", *WorkflowPDSRestore.Restores[restoreName].Meta.Name, *WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
			})

			Step("Increase PVC size by 1 GB", func() {
				log.InfoD("Increase PVC size by 1 GB")
				_, err := IncreasePVCSize(WorkflowDataService.DataServiceDeployment[*deployment.Create.Meta.Uid].Namespace, *deployment.Create.Status.CustomResourceName, 1)
				log.FailOnError(err, "Unable to increase size of PVC")
			})

			Step("Create Adhoc backup config of the existing deployment - After PVC Increase", func() {
				pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
				bkpConfigResponse, err := WorkflowPDSBackupConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backupConfig")
				log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)
			})

			Step("Get the latest backup detail for the deployment - After PVC Increase", func() {
				backupResponse, err := WorkflowPDSBackup.GetLatestBackup(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backup")
				latestBackupUidAfterPVCIncrease = *backupResponse.Meta.Uid
				log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
				err = WorkflowPDSBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
				log.FailOnError(err, "Error occured while waiting for backup to complete")
			})

			Step("Create Restore from the latest backup Id  - After PVC Increase", func() {
				defer func() {
					err := SetSourceKubeConfig()
					log.FailOnError(err, "failed to switch context to source cluster")
				}()
				CheckforClusterSwitch()
				restoreNamespace = "restore-" + RandomString(5)
				restoreName = "restore-" + RandomString(5)
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUidAfterPVCIncrease, restoreNamespace, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Restore Failed")
				log.Infof("All restores - [%+v]", WorkflowPDSRestore.Restores)
				log.Infof("Restore Created Name - [%s], UID - [%s]", *WorkflowPDSRestore.Restores[restoreName].Meta.Name, *WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
			})

			Step("Create Restore from the latest backup Id  - After PVC Increase from older backup", func() {
				defer func() {
					err := SetSourceKubeConfig()
					log.FailOnError(err, "failed to switch context to source cluster")
				}()
				CheckforClusterSwitch()
				restoreNamespace = "restore-" + RandomString(5)
				restoreName = "restore-" + RandomString(5)
				_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUidBeforePVCIncrease, restoreNamespace, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Restore Failed")
				log.Infof("All restores - [%+v]", WorkflowPDSRestore.Restores)
				log.Infof("Restore Created Name - [%s], UID - [%s]", *WorkflowPDSRestore.Restores[restoreName].Meta.Name, *WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
			})

			Step("Reboot all deployment pods - After PVC Resize", func() {
				log.InfoD("Reboot all deployment pods - After PVC Resize")
				delploymentName := *deployment.Create.Status.CustomResourceName
				err = WorkflowDataService.DeletePDSPods([]string{delploymentName}, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Deployment pod reboot failed")
			})

			Step("Validate Data Service to after PVC resize", func() {
				log.InfoD("Validate Data Service to after failover")
				err = WorkflowDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.Replicas, WorkflowDataService.PDSTemplates.ResourceTemplateId, WorkflowDataService.PDSTemplates.StorageTemplateId, PDS_DEFAULT_NAMESPACE, ds.Version, ds.Image)
				log.FailOnError(err, "Error while Validating dataservice after scale up")
			})

			Step("Run data load on the deployment", func() {
				_, err := WorkflowDataService.RunDataServiceWorkloads(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while running workloads on ds")
			})

		}

	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})

var _ = Describe("{PerformSimultaneousBackupRestoreForMultipleDeployments}", func() {
	var (
		deployments          []*automationModels.PDSDeploymentResponse
		pdsBackupConfigName  string
		restoreNames         []string
		allBackupIds         map[string][]string
		backupsPerDeployment int
		allErrors            []error
		deploymentCount      int
		wg                   sync.WaitGroup
		restoreCount         int
	)
	JustBeforeEach(func() {
		StartPDSTorpedoTest("PerformSimultaneousBackupRestoreForMultipleDeployments", "Perform multiple backup and restore simultaneously for different deployments.", nil, 0)
		deploymentCount = 2
		backupsPerDeployment = 1
		restoreCount = 3
		allBackupIds = make(map[string][]string)
	})

	It("Perform multiple backup and restore simultaneously for different deployments.", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {

			for i := 0; i < deploymentCount; i++ {
				wg.Add(1)
				go func() {

					var deploymentNamespace string

					defer wg.Done()
					defer GinkgoRecover()

					Step("Create a namespace for PDS", func() {
						deploymentNamespace = fmt.Sprintf("%s-%s", strings.ToLower(ds.Name), RandomString(5))
						_, err := WorkflowNamespace.CreateNamespaces(deploymentNamespace)
						if err != nil {
							log.Errorf("Error occured while creating namespace - [%s]", err.Error())
							allErrors = append(allErrors, err)
							return
						}
						log.Infof("Namespaces created - [%s]", WorkflowNamespace.Namespaces)
					})

					Step("Associate namespace to Project", func() {

						log.InfoD("Asscoaiting [%s]-[%s] to the project", deploymentNamespace, WorkflowNamespace.Namespaces[deploymentNamespace])

						err := WorkflowProject.Associate(
							[]string{},
							[]string{WorkflowNamespace.Namespaces[deploymentNamespace]},
							[]string{},
							[]string{},
							[]string{},
							[]string{},
						)
						if err != nil {
							log.Errorf("Error occured while associating namespace - [%s]", err.Error())
							allErrors = append(allErrors, err)
							return
						}
						log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)
						time.Sleep(30 * time.Second)
					})

					Step("Deploy dataservice", func() {

						log.InfoD("Starting deployment in [%s] namespace", deploymentNamespace)

						WorkflowDataService.PDSTemplates = WorkflowPDSTemplate

						currDeployment, err := WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, deploymentNamespace)
						if err != nil {
							log.Errorf("Error occured while creating deployment on [%s] - [%s]", deploymentNamespace, err.Error())
							allErrors = append(allErrors, err)
							return
						}
						log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)
						deployments = append(deployments, currDeployment)

					})
				}()
			}

			wg.Wait()
			dash.VerifyFatal(len(allErrors), 0, "Verifying parallel deployments")
		}

		Step("Create multiple Adhoc backup config for the existing deployment", func() {

			log.Infof("All Deployments - [%v]", deployments)

			for _, deployment := range deployments {
				for i := 0; i < backupsPerDeployment; i++ {

					wg.Add(1)
					go func(dep *automationModels.PDSDeploymentResponse) {

						defer wg.Done()
						defer GinkgoRecover()

						pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
						bkpConfigResponse, err := WorkflowPDSBackupConfig.CreateBackupConfig(pdsBackupConfigName, *dep.Create.Meta.Uid)
						if err != nil {
							log.Errorf("Some error occurred while creating backup [%s], Error - [%s]", pdsBackupConfigName, err.Error())
							allErrors = append(allErrors, err)
						} else {
							log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
						}
					}(deployment)
				}
			}

			wg.Wait()

			dash.VerifyFatal(len(allErrors), 0, "Verifying multiple backup creation")
			log.InfoD("Simultaneous backup config creation succeeded")
		})

		Step("Get the backup detail for the backup configs", func() {
			for _, deployment := range deployments {
				allBackupResponse, err := WorkflowPDSBackup.ListAllBackups(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while fetching backups")
				dash.VerifyFatal(len(allBackupResponse), backupsPerDeployment, fmt.Sprintf("Total number of backups found for [%s] are not consisten with backup configs created.", *deployment.Create.Meta.Name))
				for _, backupResponse := range allBackupResponse {
					log.Infof("Backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
					err = WorkflowPDSBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
					log.FailOnError(err, "Error occured while waiting for backup to complete")
					allBackupIds[*deployment.Create.Meta.Uid] = append(allBackupIds[WorkflowDataService.DataServiceDeployment[*deployment.Create.Meta.Uid].Namespace], *backupResponse.Meta.Uid)
				}
			}

			log.InfoD("Simultaneous backups creation succeeded")
		})

		Step("Creating Simultaneous restores from the dataservices and triggering parallel backup", func() {
			defer func() {
				delete(WorkflowPDSBackupConfig.SkipValidatation, pds.RunDataBeforeBackup)
			}()

			log.InfoD("Creating parallel restores")

			// Creating parallel restores
			for depId, backupIds := range allBackupIds {

				for _, backupId := range backupIds {

					for i := 0; i < restoreCount; i++ {
						wg.Add(1)
						go func(deploymentId string, backup string) {
							defer wg.Done()
							defer GinkgoRecover()

							restoreName := "restore-" + RandomString(5)
							_, err := WorkflowPDSRestore.CreateRestore(restoreName, backup, restoreName, deploymentId)
							if err != nil {
								log.Errorf("Error occurred while creating [%s], Error - [%s]", restoreName, err.Error())
							} else {
								log.Infof("Restore created successfully with ID - [%s]", WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
								restoreNames = append(restoreNames, restoreName)
							}
						}(depId, backupId)
					}
				}

			}

			log.InfoD("Creating backups parallel with restores")

			WorkflowPDSBackupConfig.SkipValidatation[pds.RunDataBeforeBackup] = true
			// Creating parallel backups
			for _, deployment := range deployments {
				for i := 0; i < backupsPerDeployment; i++ {
					wg.Add(1)
					go func(dep *automationModels.PDSDeploymentResponse) {

						defer wg.Done()
						defer GinkgoRecover()

						pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
						bkpConfigResponse, err := WorkflowPDSBackupConfig.CreateBackupConfig(pdsBackupConfigName, *dep.Create.Meta.Uid)
						if err != nil {
							log.Errorf("Some error occurred while creating backup [%s], Error - [%s]", pdsBackupConfigName, err.Error())
							allErrors = append(allErrors, err)
						}
						log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)

					}(deployment)
				}
			}

			wg.Wait()
			dash.VerifyFatal(len(allErrors), 0, "Verifying multiple backup config/restore creation in parallel")

			log.InfoD("Waiting for all backups to be successful")
			// Validating parallel backup success
			for _, deployment := range deployments {
				allBackupResponse, err := WorkflowPDSBackup.ListAllBackups(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backup")
				dash.VerifyFatal(len(allBackupResponse), backupsPerDeployment*2, fmt.Sprintf("Total number of backups found for [%s] are not consisten with backup configs created.", *deployment.Create.Meta.Name))
				for _, backupResponse := range allBackupResponse {
					log.Infof("Backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
					err = WorkflowPDSBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
					log.FailOnError(err, "Error occured while waiting for backup to complete")
					allBackupIds[WorkflowDataService.DataServiceDeployment[*deployment.Create.Meta.Uid].Namespace] = append(allBackupIds[WorkflowDataService.DataServiceDeployment[*deployment.Create.Meta.Uid].Namespace], *backupResponse.Meta.Uid)
				}
			}

			log.InfoD("Simultaneous backup/restores succeeded")
		})
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})
