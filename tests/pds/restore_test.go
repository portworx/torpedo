package tests

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	pdslib "github.com/portworx/torpedo/drivers/pds/lib"
	pdsbkp "github.com/portworx/torpedo/drivers/pds/pdsbackup"
	restoreBkp "github.com/portworx/torpedo/drivers/pds/pdsrestore"
	tc "github.com/portworx/torpedo/drivers/pds/targetcluster"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

var (
	restoreTargetCluster  *tc.TargetCluster
	bkpTarget             *pds.ModelsBackupTarget
	dsEntity              restoreBkp.DSEntity
	bkpJob                *pds.ModelsBackupJobStatusResponse
	restoredDeployment    *pds.ModelsDeployment
	restoredDeployments   []*pds.ModelsDeployment
	deploymentDSentityMap = make(map[*pds.ModelsDeployment]restoreBkp.DSEntity)
)

var _ = Describe("{PerformRestoreToSameCluster}", func() {
	bkpTargetName = bkpTargetName + pdsbkp.RandString(8)
	JustBeforeEach(func() {
		StartTorpedoTest("PerformRestoreToSameCluster", "Perform multiple restore within same cluster.", pdsLabels, 0)
		bkpClient, err = pdsbkp.InitializePdsBackup()
		log.FailOnError(err, "Failed to initialize backup for pds.")
		bkpTarget, err = bkpClient.CreateAwsS3BackupCredsAndTarget(tenantID, fmt.Sprintf("%v-aws", bkpTargetName), deploymentTargetID)
		log.FailOnError(err, "Failed to create S3 backup target.")
		log.InfoD("AWS S3 target - %v created successfully", bkpTarget.GetName())
		awsBkpTargets = append(awsBkpTargets, bkpTarget)
	})

	It("Perform multiple restore within same cluster", func() {
		var deploymentsToBeCleaned []*pds.ModelsDeployment
		stepLog := "Deploy data service and take adhoc backup, deleting the data service should not delete the backups."
		Step(stepLog, func() {
			log.InfoD(stepLog)
			backupSupportedDataServiceNameIDMap, err = bkpClient.GetAllBackupSupportedDataServices()
			log.FailOnError(err, "Error while fetching the backup supported ds.")
			for _, ds := range params.DataServiceToTest {
				_, supported := backupSupportedDataServiceNameIDMap[ds.Name]
				if !supported {
					log.InfoD("Data service: %v doesn't support backup, skipping...", ds.Name)
					continue
				}
				stepLog = "Deploy and validate data service"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					deployment, _, _, err = DeployandValidateDataServices(ds, params.InfraToTest.Namespace, tenantID, projectID)
					deploymentsToBeCleaned = append(deploymentsToBeCleaned, deployment)
					log.FailOnError(err, "Error while deploying data services")

					// TODO: Add workload generation

					dsEntity = restoreBkp.DSEntity{
						Deployment: deployment,
					}
				})
				stepLog = "Perform adhoc backup and validate them"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					log.Infof("Deployment ID: %v, backup target ID: %v", deployment.GetId(), bkpTarget.GetId())
					err = bkpClient.TriggerAndValidateAdhocBackup(deployment.GetId(), bkpTarget.GetId(), "s3")
					log.FailOnError(err, "Failed while performing adhoc backup")
				})
				stepLog = "Perform restore for the backup jobs."
				Step(stepLog, func() {
					log.InfoD(stepLog)
					ctx := pdslib.GetAndExpectStringEnvVar("PDS_RESTORE_TARGET_CLUSTER")
					restoreTarget := tc.NewTargetCluster(ctx)
					restoreClient := restoreBkp.RestoreClient{
						TenantId:             tenantID,
						ProjectId:            projectID,
						Components:           components,
						Deployment:           deployment,
						RestoreTargetCluster: restoreTarget,
					}
					backupJobs, err := restoreClient.Components.BackupJob.ListBackupJobsBelongToDeployment(projectID, deployment.GetId())
					log.FailOnError(err, "Error while fetching the backup jobs for the deployment: %v", deployment.GetClusterResourceName())
					for _, backupJob := range backupJobs {
						log.Infof("[Restoring] Details Backup job name- %v, Id- %v", backupJob.GetName(), backupJob.GetId())
						restoredModel, err := restoreClient.TriggerAndValidateRestore(backupJob.GetId(), params.InfraToTest.Namespace, dsEntity, true, true)
						log.FailOnError(err, "Failed during restore.")
						log.Infof("Validate ")
						restoredDeployment, err = restoreClient.Components.DataServiceDeployment.GetDeployment(restoredModel.GetDeploymentId())
						log.FailOnError(err, fmt.Sprintf("Failed while fetching the restore data service instance: %v", restoredModel.GetClusterResourceName()))
						deploymentsToBeCleaned = append(deploymentsToBeCleaned, restoredDeployment)
						log.InfoD("Restored successfully. Details: Deployment- %v, Status - %v", restoredModel.GetClusterResourceName(), restoredModel.GetStatus())
					}
				})

				Step("Delete Deployments", func() {
					CleanupDeployments(deploymentsToBeCleaned)
				})
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		err := bkpClient.AWSStorageClient.DeleteBucket()
		log.FailOnError(err, "Failed while deleting the bucket")
	})
})

var _ = Describe("{PerformRestorePostPVCResize}", func() {
	bkpTargetName = bkpTargetName + pdsbkp.RandString(8)
	JustBeforeEach(func() {
		StartTorpedoTest("PerformRestorePostPVCResize", "Perform multiple backup and restore after resize of PVC", pdsLabels, 0)
		bkpClient, err = pdsbkp.InitializePdsBackup()
		log.FailOnError(err, "Failed to initialize backup for pds.")
		bkpTarget, err = bkpClient.CreateAwsS3BackupCredsAndTarget(tenantID, fmt.Sprintf("%v-aws", bkpTargetName), deploymentTargetID)
		log.FailOnError(err, "Failed to create S3 backup target.")
		log.InfoD("AWS S3 target - %v created successfully", bkpTarget.GetName())
	})

	It("Perform restore Post PVC Resize within same cluster", func() {
		var deployments = make(map[PDSDataService]*pds.ModelsDeployment)
		var deploymentsToBeCleaned []*pds.ModelsDeployment
		var depList []*pds.ModelsDeployment
		var dsEntity1 restoreBkp.DSEntity
		var dsEntity2 restoreBkp.DSEntity
		stepLog := "Deploy data service and take adhoc backup, "
		Step(stepLog, func() {
			log.InfoD(stepLog)
			backupSupportedDataServiceNameIDMap, err = bkpClient.GetAllBackupSupportedDataServices()
			log.FailOnError(err, "Error while fetching the backup supported ds.")
			for _, ds := range params.DataServiceToTest {
				_, supported := backupSupportedDataServiceNameIDMap[ds.Name]
				if !supported {
					log.InfoD("Data service: %v doesn't support backup, skipping...", ds.Name)
					continue
				}
				stepLog = "Deploy and validate data service"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					dsDeployment, _, _, err := DeployandValidateDataServices(ds, params.InfraToTest.Namespace, tenantID, projectID)
					depList = append(depList, deployment)
					deploymentsToBeCleaned = append(deploymentsToBeCleaned, deployment)
					log.FailOnError(err, "Error while deploying data services")

					// TODO: Add workload generation

					dsEntity1 = restoreBkp.DSEntity{
						Deployment: dsDeployment,
					}
					log.Infof("Details DSObject- %v, Name - %v, DSEntity - %v", dsDeployment, dsDeployment.GetClusterResourceName(), dsEntity1)
					deploymentDSentityMap[dsDeployment] = dsEntity1
				})
			}
		})
		stepLog = "Perform adhoc backup and validate before PVC resize"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for deployment := range deploymentDSentityMap {
				log.Infof("Deployment ID: %v, backup target ID: %v", deployment.GetId(), bkpTarget.GetId())
				err = bkpClient.TriggerAndValidateAdhocBackup(deployment.GetId(), bkpTarget.GetId(), "s3")
				log.FailOnError(err, "Failed while performing adhoc backup")
			}
		})
		stepLog = "Perform restore for the backup jobs before PVC Resize"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			ctx := pdslib.GetAndExpectStringEnvVar("PDS_RESTORE_TARGET_CLUSTER")
			restoreTarget := tc.NewTargetCluster(ctx)
			restoreClient := restoreBkp.RestoreClient{
				TenantId:             tenantID,
				ProjectId:            projectID,
				Components:           components,
				Deployment:           deployment,
				RestoreTargetCluster: restoreTarget,
			}
			backupJobs1, err := restoreClient.Components.BackupJob.ListBackupJobsBelongToDeployment(projectID, deployment.GetId())
			log.FailOnError(err, "Error while fetching the backup jobs for the deployment: %v", deployment.GetClusterResourceName())
			for _, backupJob := range backupJobs1 {
				log.Infof("[Restoring] Details Backup job name- %v, Id- %v", backupJob.GetName(), backupJob.GetId())
				restoredModel, err := restoreClient.TriggerAndValidateRestore(backupJob.GetId(), params.InfraToTest.Namespace, dsEntity1, true, true)
				log.FailOnError(err, "Failed during restore.")
				log.Infof("Validate ")
				restoredDeployment, err = restoreClient.Components.DataServiceDeployment.GetDeployment(restoredModel.GetDeploymentId())
				log.FailOnError(err, fmt.Sprintf("Failed while fetching the restore data service instance: %v", restoredModel.GetClusterResourceName()))
				deploymentsToBeCleaned = append(deploymentsToBeCleaned, restoredDeployment)
				log.InfoD("Restored successfully. Details: Deployment- %v, Status - %v", restoredModel.GetClusterResourceName(), restoredModel.GetStatus())
			}
		})
		stepLog = "Resize the App PVC by 1 gb"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			ctx, err := Inst().Pds.CreateSchedulerContextForPDSApps(depList)
			log.FailOnError(err, "Unable to create scheduler context")
			err = IncreasePVCby1Gig(ctx)
			log.FailOnError(err, "Failing while Increasing the PVC name...")
			dsEntity2 = restoreBkp.DSEntity{
				Deployment: deployment,
			}
			log.Infof("Details DSObject- %v, Name - %v, DSEntity - %v", deployment, deployment.GetClusterResourceName(), dsEntity2)
			deploymentDSentityMap[deployment] = dsEntity2

			//ToDo: Add a step to take backup after resize.

			Step("Validate Deployments after PVC Resize", func() {
				for ds, deployment := range deployments {
					err = dsTest.ValidateDataServiceDeployment(deployment, namespace)
					log.FailOnError(err, "Error while validating dataservices")
					log.InfoD("Data-service: %v is up and healthy", ds.Name)
				}
			})
		})
		stepLog = "Take backup after resize of PVC"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.InfoD(stepLog)
			for deployment := range deploymentDSentityMap {
				log.Infof("Deployment ID: %v, backup target ID: %v", deployment.GetId(), bkpTarget.GetId())
				err = bkpClient.TriggerAndValidateAdhocBackup(deployment.GetId(), bkpTarget.GetId(), "s3")
				log.FailOnError(err, "Failed while performing adhoc backup")
			}
		})
		stepLog = "Perform restore for the backup jobs after PVC Resize"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			ctx := pdslib.GetAndExpectStringEnvVar("PDS_RESTORE_TARGET_CLUSTER")
			restoreTarget := tc.NewTargetCluster(ctx)
			restoreClient := restoreBkp.RestoreClient{
				TenantId:             tenantID,
				ProjectId:            projectID,
				Components:           components,
				Deployment:           deployment,
				RestoreTargetCluster: restoreTarget,
			}
			backupJobs2, err := restoreClient.Components.BackupJob.ListBackupJobsBelongToDeployment(projectID, deployment.GetId())
			log.FailOnError(err, "Error while fetching the backup jobs for the deployment: %v", deployment.GetClusterResourceName())
			for _, backupJob := range backupJobs2 {
				log.Infof("[Restoring] Details Backup job name- %v, Id- %v", backupJob.GetName(), backupJob.GetId())
				restoredModel, err := restoreClient.TriggerAndValidateRestore(backupJob.GetId(), params.InfraToTest.Namespace, dsEntity2, true, true)
				log.FailOnError(err, "Failed during restore.")
				log.Infof("Validate ")
				restoredDeployment, err = restoreClient.Components.DataServiceDeployment.GetDeployment(restoredModel.GetDeploymentId())
				log.FailOnError(err, fmt.Sprintf("Failed while fetching the restore data service instance: %v", restoredModel.GetClusterResourceName()))
				deploymentsToBeCleaned = append(deploymentsToBeCleaned, restoredDeployment)
				log.InfoD("Restored successfully. Details: Deployment- %v, Status - %v", restoredModel.GetClusterResourceName(), restoredModel.GetStatus())
			}
		})
		Step("Delete Deployments", func() {
			CleanupDeployments(deploymentsToBeCleaned)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		err := bkpClient.AWSStorageClient.DeleteBucket()
		log.FailOnError(err, "Failed while deleting the bucket")
	})
})
