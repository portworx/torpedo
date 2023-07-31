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

var _ = Describe("{PerformRestoreValidatingHA}", func() {
	bkpTargetName = bkpTargetName + pdsbkp.RandString(8)
	JustBeforeEach(func() {
		StartTorpedoTest("PerformRestoreToSameCluster", "Perform multiple restore within same cluster.", pdsLabels, 0)
		bkpClient, err = pdsbkp.InitializePdsBackup()
		log.FailOnError(err, "Failed to initialize backup for pds.")
		bkpTarget, err = bkpClient.CreateAwsS3BackupCredsAndTarget(tenantID, fmt.Sprintf("%v-aws", bkpTargetName), deploymentTargetID)
		log.FailOnError(err, "Failed to create S3 backup target.")
		log.InfoD("AWS S3 target - %v created successfully", bkpTarget.GetName())
		ctx := pdslib.GetAndExpectStringEnvVar("TARGET_KUBECONFIG")
		sourceTarget = tc.NewTargetCluster(ctx)
		ctx = pdslib.GetAndExpectStringEnvVar("PDS_RESTORE_TARGET_CLUSTER")
		restoreTargetCluster = tc.NewTargetCluster(ctx)
	})

	It("Perform multiple restore within same cluster", func() {
		var (
			deploymentsToBeCleaned []*pds.ModelsDeployment
			nsName                 = params.InfraToTest.Namespace
		)
		stepLog := "Deploy data service and take adhoc backup."
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
				})
				stepLog = "Perform adhoc backup before killing deployment pods."
				Step(stepLog, func() {
					log.InfoD(stepLog)
					dsEntity = restoreBkp.DSEntity{
						Deployment: deployment,
					}
					log.Infof("Deployment ID: %v, backup target ID: %v", deployment.GetId(), bkpTarget.GetId())
					err = bkpClient.TriggerAndValidateAdhocBackup(deployment.GetId(), bkpTarget.GetId(), "s3")
					log.FailOnError(err, "Failed while performing adhoc backup")
				})

				stepLog = "Kill set of pods for HA."
				Step(stepLog, func() {
					dbMaster, isNativelyDistributed := GetDbMasterNode(nsName, ds.Name, deployment, sourceTarget)
					log.FailOnError(err, "Failed while fetching db master node.")
					isNativelyDistributed = false
					if !isNativelyDistributed {
						err = sourceTarget.DeleteK8sPods(dbMaster, nsName)
						log.FailOnError(err, "Failed while deleting db master pod.")
						err = dsTest.ValidateDataServiceDeployment(deployment, nsName)
						log.FailOnError(err, "Failed while validating the deployment pods, post pod deletion.")
						newDbMaster, _ := GetDbMasterNode(nsName, ds.Name, deployment, sourceTarget)
						if dbMaster == newDbMaster {
							log.FailOnError(fmt.Errorf("leader node is not reassigned"), fmt.Sprintf("Leader pod %v", dbMaster))
						}
					} else {
						podName, err := sourceTarget.GetAnyPodName(deployment.GetClusterResourceName(), nsName)
						log.FailOnError(err, "Failed while fetching pod for stateful set %v.", deployment.GetClusterResourceName())
						err = sourceTarget.KillPodsInNamespace(params.InfraToTest.Namespace, podName)
						log.FailOnError(err, "Failed while deleting pod.")
						err = dsTest.ValidateDataServiceDeployment(deployment, nsName)
						log.FailOnError(err, "Failed while validating the deployment pods, post pod deletion.")
					}

				})
				stepLog = "Perform adhoc backup and validate them"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					dsEntity = restoreBkp.DSEntity{
						Deployment: deployment,
					}
					log.Infof("Deployment ID: %v, backup target ID: %v", deployment.GetId(), bkpTarget.GetId())
					err = bkpClient.TriggerAndValidateAdhocBackup(deployment.GetId(), bkpTarget.GetId(), "s3")
					log.FailOnError(err, "Failed while performing adhoc backup")
				})

				stepLog = "Perform restore for the backup jobs."
				Step(stepLog, func() {
					log.InfoD(stepLog)
					restoreClient := restoreBkp.RestoreClient{
						TenantId:             tenantID,
						ProjectId:            projectID,
						Components:           components,
						Deployment:           deployment,
						RestoreTargetCluster: restoreTargetCluster,
					}
					backupJobs, err := restoreClient.Components.BackupJob.ListBackupJobsBelongToDeployment(projectID, deployment.GetId())
					log.FailOnError(err, "Error while fetching the backup jobs for the deployment: %v", deployment.GetClusterResourceName())
					for _, backupJob := range backupJobs {
						log.Infof("[Restoring] Details Backup job name- %v, Id- %v", backupJob.GetName(), backupJob.GetId())
						restoredModel, err := restoreClient.TriggerAndValidateRestore(backupJob.GetId(), params.InfraToTest.Namespace, dsEntity, true, true)
						log.FailOnError(err, "Failed during restore.")
						restoredDeployment, err = restoreClient.Components.DataServiceDeployment.GetDeployment(restoredModel.GetDeploymentId())
						log.FailOnError(err, fmt.Sprintf("Failed while fetching the restore data service instance: %v", restoredModel.GetClusterResourceName()))
						deploymentsToBeCleaned = append(deploymentsToBeCleaned, restoredDeployment)
						log.InfoD("Restored successfully. Deployment- %v", restoredModel.GetClusterResourceName())
					}
				})

				// TODO trigger workload for restored deployment

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
