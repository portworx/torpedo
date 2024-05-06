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
	"sync"
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
			workflowDataService.NamespaceName = Namespace
			serviceConfigId, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams)
			log.FailOnError(err, "Unable to create Custom Templates for PDS")
			// workflowDataService.PDSTemplates.ServiceConfigTemplateId = serviceConfigId[ds.Name]
			workflowDataService.PDSTemplates.StorageTemplateId = stConfigId
			workflowDataService.PDSTemplates.ResourceTemplateId = resConfigId
			tempList = append(tempList, serviceConfigId[ds.Name], stConfigId, resConfigId)
			deployment, err = workflowDataService.DeployDataService(ds, ds.Image, ds.Version)
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

		stepLog := "Running Workloads before taking backups"
		Step(stepLog, func() {
			_, err := workflowDataService.RunDataServiceWorkloads(*deployment.Create.Meta.Uid, NewPdsParams)
			log.FailOnError(err, "Error while running workloads on ds")
		})
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
			err = workflowDataService.KillDBMasterNodeToValidateHA(ds.Name, *deployment.Create.Meta.Name)
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
			workflowRestore.Source = &WorkflowNamespace
			backupUid := *bkpConfigResponse.Create.Meta.Uid
			deploymentName := *deployment.Create.Meta.Name
			cloudSnapId := ""
			// Set the DestClusterId same as the current ClusterId
			workflowRestore.Destination.TargetCluster.DestinationClusterId = WorkflowTargetCluster.ClusterUID
			restoreDeployment, err = workflowRestore.CreateRestore(backupUid, deploymentName, cloudSnapId)
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
		workflowDataservice  pds.WorkflowDataService
		workflowBackUpConfig pds.WorkflowPDSBackupConfig
		workflowRestore      pds.WorkflowPDSRestore
		workflowBackup       pds.WorkflowPDSBackup
		workFlowTemplates    pds.WorkflowPDSTemplates
		deployment           *automationModels.PDSDeploymentResponse
		podsToBeDeleted      []string
		latestBackupUid      string
		pdsBackupConfigName  string
		restoreNamespace     string
	)

	JustBeforeEach(func() {
		StartTorpedoTest("PerformRestorePDSPodsDown", "Perform restore while simultaneously deleting backup controller manager & target controller pods.", nil, 0)
		workflowDataservice.DataServiceDeployment = make(map[string]automationModels.DataServiceDetails)

		workflowRestore.Destination = &WorkflowNamespace
		workflowRestore.Source = &WorkflowNamespace
		workflowDataservice.Dash = dash
		restoreNamespace = "pds-restore-namespace-" + RandomString(5)
	})

	It("Perform restore while simultaneously deleting backup controller manager & target controller pods.", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy dataservice", func() {
				workFlowTemplates.Platform = WorkflowPlatform
				workflowDataservice.Namespace = &WorkflowNamespace
				workflowDataservice.NamespaceName = PDS_DEFAULT_NAMESPACE

				_, stConfigId, resConfigId, err := workFlowTemplates.CreatePdsCustomTemplatesAndFetchIds(NewPdsParams)
				log.FailOnError(err, "Unable to create Custom Templates for PDS")

				// workflowDataservice.PDSTemplates.ServiceConfigTemplateId = serviceConfigId[ds.Name]
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
				workflowBackUpConfig.WorkflowDataService = &workflowDataservice
				workflowBackUpConfig.WorkflowBackupLocation = WorkflowbkpLoc
				pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
				workflowBackUpConfig.Backups = make(map[string]automationModels.V1BackupConfig)
				bkpConfigResponse, err := workflowBackUpConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Name)
				log.FailOnError(err, "Error occured while creating backupConfig")
				log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", bkpConfigResponse.Create.Meta.Name, bkpConfigResponse.Create.Meta.Uid)
				log.Infof("All deployments - [%+v]", workflowDataservice.DataServiceDeployment)
			})

			Step("Get the latest backup detail for the deployment", func() {
				workflowBackup.WorkflowDataService = &workflowDataservice
				log.Infof("All deployments - [%+v]", workflowDataservice.DataServiceDeployment)
				backupResponse, err := workflowBackup.GetLatestBackup(*deployment.Create.Meta.Name)
				log.FailOnError(err, "Error occured while creating backup")
				latestBackupUid = *backupResponse.Meta.Uid
				log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
				err = workflowBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
				log.FailOnError(err, "Error occured while waiting for backup to complete")
			})

			Step("Simultaneously fetch and delete backupController pods from the pds namespace", func() {
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()
					log.InfoD("Delete backup controller and Target Controller operator pod")
					podsToBeDeleted = append(podsToBeDeleted, "pds-backups-operator", "pds-target-operator")
					err := workflowDataservice.DeletePDSPods()
					log.FailOnError(err, "Failed While deleting backup controller manager pod.")
				}()
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
				restoreName := "testing_restore_" + RandomString(5)
				workflowRestore.Destination = &WorkflowNamespace
				workflowRestore.Source = &WorkflowNamespace
				_, err := workflowRestore.CreateRestore(restoreName, latestBackupUid, restoreNamespace)
				log.FailOnError(err, "Restore Failed")

				log.Infof("Restore created successfully with ID - [%s]", workflowRestore.Restores[restoreName].Meta.Uid)
			})
		}

		JustAfterEach(func() {
			defer EndTorpedoTest()
		})

	})
})
