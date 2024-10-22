package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/pure-px/torpedo/drivers/unifiedPlatform/automationModels"
	pdslibs "github.com/pure-px/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/pure-px/torpedo/drivers/unifiedPlatform/stworkflows/pds"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
	. "github.com/pure-px/torpedo/tests/unifiedPlatform"
	"strconv"
	"strings"
	"sync"
	"time"
)

var _ = Describe("{PerformStorageResizeBy1GbnTimes}", func() {
	var (
		deployment           *automationModels.PDSDeploymentResponse
		templatesToBeDeleted []string
		err                  error
		oldTemplateValue     string
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("PerformStorageResizeBy1GbnTimes", "Perform PVC Resize by 1GB for n times in a loop and validate the updated vol in the storage config.", nil, 0)
	})

	It("Deploy,Validate and ScaleUp DataService", func() {

		defer func() {
			NewPdsParams.ResourceConfiguration.New_Storage_Request = oldTemplateValue
		}()

		for _, ds := range NewPdsParams.DataServiceToTest {
			Step("Deploy DataService", func() {
				WorkflowDataService.SkipValidatation = make(map[string]bool)
				WorkflowDataService.SkipValidatation["VALIDATE_PDS_WORKLOADS"] = true
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
			})

			for i := 1; i <= NewPdsParams.StorageConfigurationsSSIE.Iterations; i++ {
				//Update Ds With New Values of Resource Templates
				oldTemplateValue = NewPdsParams.ResourceConfiguration.New_Storage_Request
				NewPdsParams.ResourceConfiguration.New_Storage_Request = strconv.Itoa(i+1) + "G"
				log.Debugf("Updating the Storage to: [%s]", NewPdsParams.ResourceConfiguration.New_Storage_Request)

				resConfigIdUpdated, err := WorkflowPDSTemplate.CreateResourceTemplateWithCustomValue(NewPdsParams)
				log.FailOnError(err, "Unable to create Custom Templates for PDS")
				log.InfoD("Updated Resource Template ID- [updated- %v]", resConfigIdUpdated)
				templatesToBeDeleted = append(templatesToBeDeleted, resConfigIdUpdated)
				log.Infof("Associate newly created template to the project")
				err = WorkflowProject.Associate(
					[]string{},
					[]string{},
					[]string{},
					[]string{},
					[]string{resConfigIdUpdated},
					[]string{},
				)
				log.FailOnError(err, "Unable to associate Templates to Project")
				log.Infof("Associated Resources - [%+v]", WorkflowProject.AssociatedResources)

				beforeResizePodAge, err := WorkflowDataService.GetPodAgeForDeployment(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "unable to get pods AGE before Storage resize")
				log.InfoD("Pods Age before storage resize is- [%v]Min", beforeResizePodAge)

				//sleeping 30sec before the next update
				time.Sleep(30 * time.Second)

				WorkflowDataService.UpdateDeploymentTemplates = true
				WorkflowDataService.PDSTemplates = WorkflowPDSTemplate
				_, err = WorkflowDataService.UpdateDataService(ds, *deployment.Create.Meta.Uid, ds.Image, ds.Version)
				log.FailOnError(err, "Error while updating ds")

				afterResizePodAge, err := WorkflowDataService.GetPodAgeForDeployment(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "unable to get pods AGE before Storage resize")
				log.InfoD("Pods Age after storage resize is- [%v]Min", afterResizePodAge)

				dash.VerifyFatal(afterResizePodAge > beforeResizePodAge, true, "Validating if the pod restarted after storage size increase")
			}

			stepLog := "Running Workloads after StorageSizeIncrease of DataService"
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

var _ = Describe("{SSIEDeployBackupRestoreAndDeleteDS}", func() {
	var (
		deployment          *automationModels.PDSDeploymentResponse
		latestBackupUid     string
		pdsBackupConfigName string
		restoreNamespace    string
		restoreName         string
		err                 error
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("SSIEDeployBackupRestoreAndDeleteDS", "100 times: Deploy data services and perform backup and restore on the same cluster", nil, 0)

	})

	It("Repeatedly Deploy data services and perform backup and restore on the same cluster", func() {
		for index := 1; index <= NewPdsParams.SSIE.NumIterations; index++ {
			log.Infof("================  Iteration: %d of %d  ================", index, NewPdsParams.SSIE.NumIterations)
			for _, ds := range NewPdsParams.DataServiceToTest {

				Step("Deploy Data Service", func() {
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
					restoreNamespace = "restore-" + RandomString(5)
					restoreName = "restore-" + RandomString(5)
					_, err := WorkflowPDSRestore.CreateRestore(restoreName, latestBackupUid, restoreNamespace, *deployment.Create.Meta.Uid)
					log.FailOnError(err, "Restore Failed")
					log.Infof("All restores - [%+v]", WorkflowPDSRestore.Restores)
					log.Infof("Restore Created Name - [%s], UID - [%s]", *WorkflowPDSRestore.Restores[restoreName].Meta.Name, *WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
				})

				Step("Purging the Data Services", func() {
					log.InfoD("Purging all dataService objects")
					err = WorkflowDataService.Purge(false)
					if err != nil {
						log.FailOnError(err, "error while purging Data Services\n")

					}
					err = WorkflowPDSRestore.RestoredDeployments.Purge(false)
					if err != nil {
						log.FailOnError(err, "error while purging Restored Deployments\n")

					}
				})
			}
		}
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})

var _ = Describe("{ParallelDeploymentBackupAndRestore}", func() {
	var (
		deployments          []*automationModels.PDSDeploymentResponse
		pdsBackupConfigName  string
		restoreNames         []string
		allBackupIds         map[string][]string
		backupsPerDeployment int
		allErrors            []error
		deploymentCount      int
		wg                   sync.WaitGroup
	)
	JustBeforeEach(func() {

		deploymentCount = 25

		if NewPdsParams.SSIE.Deployments != 0 {
			deploymentCount = NewPdsParams.SSIE.Deployments
		}

		StartPDSTorpedoTest("ParallelDeploymentBackupAndRestore", fmt.Sprintf("Deploy %d data service in parallel, perform parallel backup and restores for all", deploymentCount), nil, 0)

		backupsPerDeployment = 1
		allBackupIds = make(map[string][]string)
	})

	It(fmt.Sprintf("Deploy %d data service in parallel, perform parallel backup and restores", deploymentCount), func() {
		WorkflowDataService.SkipValidatation[pds.ValidatePdsWorkloads] = true
		for _, ds := range NewPdsParams.DataServiceToTest {

			for i := 0; i < deploymentCount; i++ {

				wg.Add(1)
				depName := fmt.Sprintf("%s-%s", strings.ToLower(ds.Name), RandomString(5))

				go func(deploymentNamespace string) {

					defer wg.Done()
					defer GinkgoRecover()

					Step("Create a namespace for PDS", func() {
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
				}(depName)
			}

			wg.Wait()
			dash.VerifyFatal(len(allErrors), 0, "Verifying parallel deployments")
		}

		defer func() {
			delete(WorkflowDataService.SkipValidatation, pds.ValidatePdsWorkloads)
		}()

		Step("Create parallel Adhoc backup config for the deployments", func() {

			log.Infof("All Deployments - [%v]", deployments)

			for _, deployment := range deployments {
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
			dash.VerifyFatal(len(allBackupIds), deploymentCount, "Verifying backup creation")
			log.InfoD("Simultaneous backups creation succeeded")
		})

		Step("Creating Simultaneous restores from the backups", func() {
			defer func() {
				delete(WorkflowPDSBackupConfig.SkipValidatation, pds.RunDataBeforeBackup)
				delete(WorkflowPDSRestore.SkipValidation, pds.CheckDataAfterRestore)
				delete(WorkflowPDSRestore.SkipValidation, pds.ValidatePdsRestore)
			}()

			var allErrors = make([]error, 0)

			log.InfoD("Creating parallel restores")
			WorkflowPDSRestore.SkipValidation[pds.CheckDataAfterRestore] = true
			WorkflowPDSRestore.SkipValidation[pds.ValidatePdsRestore] = true

			// Creating parallel restores
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
							allErrors = append(allErrors, err)
						} else {
							log.Infof("Restore created successfully with ID - [%s]", WorkflowPDSRestore.Restores[restoreName].Meta.Uid)
							restoreNames = append(restoreNames, restoreName)
						}

						err = pdslibs.ValidateRestoreDeployment(*WorkflowPDSRestore.Restores[restoreName].Meta.Uid, restoreName)
						if err != nil {
							log.Errorf("Error occurred while validating restore [%s], Error - [%s]", restoreName, err.Error())
							allErrors = append(allErrors, err)
						}

					}(depId, backupId)
				}

			}

			wg.Wait()

			dash.VerifyFatal(len(allErrors), 0, "Verifying restore creation")
			log.InfoD("Simultaneous backup/restores succeeded")
		})
	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
	})
})
