package tests

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/pure-px/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/pure-px/torpedo/drivers/utilities"
	utils "github.com/pure-px/torpedo/drivers/utilities"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
	. "github.com/pure-px/torpedo/tests/unifiedPlatform"
	"time"
)

const (
	envMinioAccessKey = "AWS_MINIO_ACCESS_KEY_ID"
	envMinioSecretKey = "AWS_MINIO_SECRET_ACCESS_KEY"
	envMinioRegion    = "AWS_MINIO_REGION"
	envMinioEndPoint  = "AWS_MINIO_ENDPOINT"
)

var _ = Describe("{DeleteDataServiceAndValidateBackupAtObjectStore}", func() {
	var (
		deployment          *automationModels.PDSDeploymentResponse
		pdsBackupConfigName string
		timeStamp           time.Time
		allFoldersBefore    []string
		err                 error
		awsClient           utilities.AwsCompatibleStorageClient
	)

	JustBeforeEach(func() {
		StartPDSTorpedoTest("DeleteDataServiceAndValidateBackupAtObjectStore", "Delete the PDS data service should not delete the backups in backend", nil, 0)
		awsClient = utilities.AwsCompatibleStorageClient{
			Endpoint:  utilities.GetEnv(envMinioEndPoint, envMinioEndPoint),
			AccessKey: utilities.GetEnv(envMinioAccessKey, envMinioAccessKey),
			SecretKey: utilities.GetEnv(envMinioSecretKey, envMinioSecretKey),
			Region:    utilities.GetEnv(envMinioRegion, envMinioRegion),
		}
	})

	It("Delete the PDS data service should not delete the backups in backend", func() {
		for _, ds := range NewPdsParams.DataServiceToTest {

			Step("Deploy dataservice", func() {
				deployment, err = WorkflowDataService.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
				log.FailOnError(err, "Error while deploying ds")
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)

			})

			Step("Listing all folders before taking the backup", func() {
				timeStamp = time.Now().UTC()
				allFoldersBefore, err = awsClient.ListFolders(timeStamp, PDSBucketName)
				log.FailOnError(err, "Error while listing all folders before taking the backup")
				log.Infof("All folders before backup - [%v] - [%d]", allFoldersBefore, len(allFoldersBefore))
				time.Sleep(2 * time.Minute)
			})

			Step("Create Adhoc backup config of the existing deployment", func() {
				timeStamp = time.Now().UTC()
				pdsBackupConfigName = "pds-adhoc-backup-" + RandomString(5)
				bkpConfigResponse, err := WorkflowPDSBackupConfig.CreateBackupConfig(pdsBackupConfigName, *deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backupConfig")
				log.Infof("BackupConfigName: [%s], BackupConfigId: [%s]", *bkpConfigResponse.Create.Meta.Name, *bkpConfigResponse.Create.Meta.Uid)
				log.Infof("All deployments - [%+v]", WorkflowDataService.DataServiceDeployment)
			})

			Step("Get the latest backup detail for the deployment", func() {
				backupResponse, err := WorkflowPDSBackup.GetLatestBackup(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error occured while creating backup")
				log.Infof("Latest backup ID [%s], Name [%s]", *backupResponse.Meta.Uid, *backupResponse.Meta.Name)
				err = WorkflowPDSBackup.WaitForBackupToComplete(*backupResponse.Meta.Uid)
				log.FailOnError(err, "Error occured while waiting for backup to complete")
			})

			Step("Delete all the deployments", func() {
				err := WorkflowDataService.DeleteDeployment(*deployment.Create.Meta.Uid)
				log.FailOnError(err, "Error while deleting all deployments")
				//To validate if the backups are getting cleaned up in the target location
				time.Sleep(2 * time.Minute)
			})

			Step("Check for backup in the s3 bucket", func() {
				allFoldersAfter, err := awsClient.ListFolders(timeStamp, PDSBucketName)
				log.FailOnError(err, "Error while checking backup in s3 bucket")
				log.Infof("All folders after backup - [%v] - [%d]", allFoldersAfter, len(allFoldersAfter))
				dash.VerifyFatal(len(allFoldersAfter) > len(allFoldersBefore), true, "Backup is present in the s3 bucket")
			})
		}

	})

	JustAfterEach(func() {
		defer EndPDSTorpedoTest()
		err := utils.DeletePvandPVCs(*deployment.Create.Status.CustomResourceName, false)
		dash.VerifySafely(err, nil, "Error while deleting pv and pvc")
		err = utils.RemoveFinalizersFromAllResources(PDS_DEFAULT_NAMESPACE)
		dash.VerifySafely(err, nil, "Error while removing finalizers from all resources")
	})
})
