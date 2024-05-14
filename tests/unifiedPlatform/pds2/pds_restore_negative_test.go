package tests

import (
	"fmt"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	. "github.com/portworx/torpedo/tests/unifiedPlatform"
)

var _ = Describe("{PerformRestoreValidatingHA}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("PerformRestoreValidatingHA", "Deploy data services, Perform restore while validating HA on the same cluster", nil, 0)
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
			workflowDataService.Namespace = &WorkflowNamespace
			deployment, err = workflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
			log.FailOnError(err, "Error while deploying ds")
		}

		defer func() {
			Step("Delete created Templates", func() {
				err := workFlowTemplates.DeleteCreatedCustomPdsTemplates(tempList)
				log.FailOnError(err, "Unable to delete Custom Templates for PDS")
			})
		}()

		defer func() {
			Step("Delete DataServiceDeployment", func() {
				err := workflowDataService.DeleteDeployment(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while deleting dataservice")
			})
		}()

		//stepLog := "Running Workloads before taking backups"
		//Step(stepLog, func() {
		//	err := workflowDataService.RunDataServiceWorkloads(NewPdsParams, "")
		//	log.FailOnError(err, "Error while running workloads on ds")
		//})
	})
	It("Perform adhoc backup before killing deployment pods.", func() {
		var bkpConfigResponse *automationModels.PDSBackupConfigResponse
		workflowBackUpConfig.WorkflowDataService = &workflowDataService
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

	It("Kill set of pods for HA.", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			log.InfoD("Kill set of pods of Dataservice to validate HA- [%v]", ds.Name)
			err = workflowDataService.KillDBMasterNodeToValidateHA(ds.Name, *deployment.Create.Meta.Uid)
			log.FailOnError(err, "Error occured while Killing pods to validate HA")
		}
	})
	It("Perform adhoc backup, restore after killing few pods to validate HA", func() {
		var bkpConfigResponse *automationModels.PDSBackupConfigResponse
		workflowBackUpConfig.WorkflowDataService = &workflowDataService
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
			workflowRestore.Source = &WorkflowDataService
			backupUid := *bkpConfigResponse.Create.Meta.Uid
			deploymentName := *deployment.Create.Meta.Name
			cloudSnapId := ""
			// Set the DestClusterId same as the current ClusterId
			workflowRestore.Destination.TargetCluster.ClusterUID = WorkflowTargetCluster.ClusterUID
			restoreDeployment, err = workflowRestore.CreateRestore(backupUid, deploymentName, cloudSnapId, PDS_DEFAULT_NAMESPACE)
			log.FailOnError(err, "Error while taking restore")
			log.Debugf("Restored DeploymentName: [%s]", restoreDeployment.Create.Meta.Name)
		})

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

var _ = Describe("{PerformRestorePDSPodsDown}", func() {
	var (
		deployment          *automationModels.PDSDeploymentResponse
		latestBackupUid     string
		pdsBackupConfigName string
		restoreNamespace    string
		wg                  sync.WaitGroup
		restoreName         string
		err                 error
		allErrors           []string
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("PerformRestorePDSPodsDown", "Perform restore while simultaneously deleting backup controller manager & target controller pods.", nil, 0)
		restoreNamespace = "restore-" + RandomString(5)
		restoreName = "restore-" + RandomString(5)
	})

	It("Perform restore while simultaneously deleting backup controller manager & target controller pods.", func() {
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
				wg.Add(1)
				log.InfoD("Triggering restore - [%s]", time.Now().Format("2006-01-02 15:04:05"))
				go func() {
					defer wg.Done()
					defer GinkgoRecover()
					WorkflowPDSRestore.Destination = &WorkflowNamespaceDestination
					CheckforClusterSwitch()
					_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUid, restoreNamespace, *deployment.Create.Meta.Uid)
					if err != nil {
						allErrors = append(allErrors, err.Error())
					}
				}()

			})

			Step("Simultaneously fetch and delete backupController pods from the pds namespace", func() {
				log.InfoD("Bringing down PDS related pods from cluster - [%s]", time.Now().Format("2006-01-02 15:04:05"))
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()
					log.Infof("Delete backup controller and Target Controller operator pod")
					err := WorkflowDataService.DeletePDSPods()
					if err != nil {
						allErrors = append(allErrors, err.Error())
					}
				}()

				wg.Wait()
				dash.VerifyFatal(len(allErrors), 0, fmt.Sprintf("Verifying restores with restarted px pods. Error - [%s]", strings.Join(allErrors, "\n")))
				log.Infof("Restore created successfully with ID - [%s]", *WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
			})
		}

	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()

	})

})
