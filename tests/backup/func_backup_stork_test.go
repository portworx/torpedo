package tests

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/sched-ops/task"
	"github.com/pure-px/sched-ops/k8s/core"
	"github.com/pure-px/sched-ops/k8s/operator"
	"github.com/pure-px/torpedo/drivers"
	"github.com/pure-px/torpedo/drivers/backup"
	"github.com/pure-px/torpedo/drivers/node"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/applicationbackup"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This testcase verifies if we can add cluster without stork and install stork and retry and get cluster status as online after getting disconnected
var _ = Describe("{AddClusterWithoutStorkAndInstallStorkAndRetry}", Label(TestCaseLabelsMap[StorkInstallClusterStatusCheck]...), func() {

	var (
		scheduledAppContexts []*scheduler.Context
		bkpNamespaces        []string
		sourceClusterUid     string
		cloudCredName        string
		cloudCredUID         string
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		storkInstalledFlag   bool
	)
	storkInstalledFlag = false
	backupLocationMap := make(map[string]string)
	JustBeforeEach(func() {

		bkpNamespaces = make([]string, 0)

		StartPxBackupTorpedoTest("VerifyAddClusterWithoutStorkAndInstallStorkAndRetry", "Add cluster without stork and install stork and retry and get cluster status as online after getting disconnected", nil, 300483, ABadgujar, Q4FY24)
		//1.Step-Deploy Applications in the Cluster
		log.InfoD("Deploy applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = AppReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
				scheduledAppContexts = append(scheduledAppContexts, ctx)
			}
		}

	})
	It("Add cluster without stork and install stork and retry and get cluster status as online after getting disconnected", func() {

		//2.Step-Validate Deployed Applications
		Step("Validate applications", func() {
			log.InfoD("Validating apps")
			ValidateApplications(scheduledAppContexts)
		})

		//3.Remove Stork through Command Line
		Step("Remove Stork through Command Line", func() {
			log.InfoD("Get the stc spec from the cluster")
			stc, err := Inst().V.GetDriver()
			log.FailOnError(err, "Failed to get driver")

			log.InfoD("Check if the stork is enabled in the stc")
			if stc.Spec.Stork.Enabled == false {
				log.InfoD("Stork is Disabled so skipping disabling part")
			} else if stc.Spec.Stork.Enabled == true {
				// Change the stork.enabled flag to false
				stc.Spec.Stork.Enabled = false

				pxOperator := operator.Instance()
				_, err = pxOperator.UpdateStorageCluster(stc)
				log.FailOnError(err, "Failed to update the storage cluster")
				log.InfoD("STC updated: stork.enabled set to false.")
			}

			log.InfoD("Remove Stork CRDs through Command Line")
			masterNode := node.GetMasterNodes()[0]
			cmd := "kubectl delete $(kubectl get crd -o name | grep stork)"
			output, err := RunCmdGetOutput(cmd, masterNode)
			log.FailOnError(err, fmt.Sprintf("Failed to run [%s] command on node [%s], error : [%s]", cmd, masterNode, err))
			log.Infof("Output from command [%s] -\n%s", cmd, output)

			//Sleep for a minute to give some time for all CRDs and deployment to be deleted and then Add Source Cluster
			time.Sleep(time.Minute)
		})

		//4.Step-Create Single Only Source Cluster which should be in Failed State
		Step("Create Single Only Source Cluster", func() {
			log.InfoD("Create Single Only Source Cluster")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			log.InfoD("Entering Cluster Config Step")
			srcClusterConfigPath, err := GetSourceClusterConfigPath()
			log.Infof("Save cluster %s kubeconfig to %s", SourceClusterName, srcClusterConfigPath)
			log.FailOnError(err, fmt.Sprintf("Error in Fetching Source Cluster Config Path - [%v] ", err))

			log.InfoD("Entering Create Cluster Step")
			err = CreateCluster(SourceClusterName, srcClusterConfigPath, BackupOrgID, "", "", ctx)
			log.InfoD("Error While Creating Cluster is - %v", err)
			dash.VerifyFatal(strings.Contains(err.Error(), "failed to validate stork deployment on cluster"), true, "Expected Failure on Creating Source and Destination Cluster")

			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.InfoD("Cluster Status is - %s", clusterStatus)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Failed, fmt.Sprintf("Verifying if [%s] cluster is disconnected/Offline/Failed", SourceClusterName))
			sourceClusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid - %s", SourceClusterName, sourceClusterUid))
		})

		//5.Step-Install Stork through Command Line
		Step("Install Stork through Command Line", func() {
			log.InfoD("Install Stork through Command Line")
			log.InfoD("Check if the stork is enabled in the stc")
			stc, err := Inst().V.GetDriver()
			log.FailOnError(err, "Failed to get driver")
			if stc.Spec.Stork.Enabled == true {
				log.InfoD("Stork is Enabled so skipping enabling part")
			} else if stc.Spec.Stork.Enabled == false {
				// Change the stork.enabled flag to false
				stc.Spec.Stork.Enabled = true
				pxOperator := operator.Instance()
				_, err = pxOperator.UpdateStorageCluster(stc)
				log.FailOnError(err, "Failed to update the storage cluster")
				log.InfoD("STC updated: stork.enabled set to true.")
			}
			storkInstalledFlag = true
			time.Sleep(time.Minute * 3)

		})

		//6.Step-Check Cluster Status After Stork Re-installation
		Step("Check Cluster Status if Online After Stork Re-installation", func() {
			log.InfoD("Check Cluster Status if Online After Stork Re-installation")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			// clusterStatus, err := Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			// log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			log.InfoD("Cluster Status before checking is - %s", clusterStatus)

			clusterStatus := func() (interface{}, bool, error) {
				checkClusterStatus, err := Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
				if err != nil {
					return "", true, err
				}
				if checkClusterStatus == api.ClusterInfo_StatusInfo_Online {
					return "", false, nil
				}
				return "", true, fmt.Errorf("the %s cluster state is not Online yet", SourceClusterName)
			}
			_, err = task.DoRetryWithTimeout(clusterStatus, time.Minute*12, time.Minute)
			dash.VerifySafely(err, nil, "Cluster is verified as Online")
		})

	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Entered Cleanup")

		//Reinstall Stork here too to ensure that in case a test case fails after stork deletion, stork is installed back in that test case
		if storkInstalledFlag == false {
			log.InfoD("Stork flag is %v", storkInstalledFlag)
			log.InfoD("Install Stork through Command Line During Cleanup")

			log.InfoD("Check if the stork is disabled in the stc")
			stc, err := Inst().V.GetDriver()
			log.FailOnError(err, "Failed to get driver")
			if stc.Spec.Stork.Enabled == true {
				log.InfoD("Stork is Enabled so skipping enabling part")
			} else if stc.Spec.Stork.Enabled == false {
				// Change the stork.enabled flag to false
				stc.Spec.Stork.Enabled = true
				pxOperator := operator.Instance()
				_, err = pxOperator.UpdateStorageCluster(stc)
				log.FailOnError(err, "Failed to update the storage cluster")
				log.InfoD("STC updated: stork.enabled set to false.")
			}
			time.Sleep(time.Minute * 3)
		}
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		//Cleanup Cluster
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)

	})

})

// Restore from backups taken by stork and not done by px backup
var _ = Describe("{RestoreFromBackupsTakenByStorkAndNotDoneByPxBackup}", Label(TestCaseLabelsMap[RestoreFromBackupsTakenByStork]...), func() {

	var (
		scheduledAppContexts []*scheduler.Context
		bkpNamespaces        []string
		backupLocationUID    string
		cloudCredName        string
		cloudCredUID         string
		backupNames          []string
		providers            []string
		backupLocationMap    map[string]string
		secretObj            *corev1.Secret
		restoreNames         []string
		adminCtx             context.Context
		secretName           string
	)
	var (
		backupLocationName = "storkbackuplocation"
		backupName         = "storkbackup"
		timeout            = 10 * time.Minute
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyRestoreFromBackupsTakenByStorkAndNotDoneByPxBackup", "Restore from backups taken by stork and not done by px backup", nil, 300566, ABadgujar, Q1FY25)

		backupLocationMap = make(map[string]string)
		bkpNamespaces = make([]string, 0)
		backupNames = make([]string, 0)
		scheduledAppContexts = make([]*scheduler.Context, 0)
		bkpNamespaces = make([]string, 0)
		var err error

		//1.Step-Deploy Applications in the Cluster
		log.InfoD("Deploy applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = AppReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
				scheduledAppContexts = append(scheduledAppContexts, ctx)
			}
		}

		log.InfoD("Namespace is - %v", bkpNamespaces[0])
		adminCtx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")

		providers = GetBackupProviders()

		for _, provider := range providers {
			if provider == drivers.ProviderAws {
				secretName = "s3secret"
				s3AccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
				dash.VerifyFatal(s3AccessKeyID != "", true, "environment variable AWS_ACCESS_KEY_ID should not be empty")
				s3SecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
				dash.VerifyFatal(s3SecretAccessKey != "", true, "environment variable AWS_ACCESS_KEY_ID should not be empty")
				s3Endpoint := os.Getenv("S3_ENDPOINT")
				dash.VerifyFatal(s3Endpoint != "", true, "environment variable s3Endpoint should not be empty")
				s3Region := os.Getenv("S3_REGION")
				dash.VerifyFatal(s3Region != "", true, "environment variable s3Region should not be empty")

				// Create the secret object

				secretObj = &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: bkpNamespaces[0],
						Name:      secretName, // Name of the secret
						Annotations: map[string]string{
							"stork.libopenstorage.org/skipresource": "true", // Add the annotation
						},
					},
					StringData: map[string]string{
						"region":          s3Region,
						"accessKeyID":     s3AccessKeyID,
						"secretAccessKey": s3SecretAccessKey,
						"endpoint":        s3Endpoint,
						"disableSSL":      "false",
					},
					Type: corev1.SecretTypeOpaque, // Default type for opaque secrets
				}

			} else if provider == drivers.ProviderNfs {
				secretName = "nfssecret"
				serverAddr := os.Getenv("NFS_SERVER_ADDR")
				mountOption := os.Getenv("NFS_MOUNT_OPTION")
				log.InfoD("NFS_MOUNT_OPTION - %v", mountOption)
				path := os.Getenv("NFS_PATH")
				subpath := os.Getenv("NFS_SUB_PATH")

				secretObj = &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: bkpNamespaces[0],
						Name:      secretName, // Name of the secret
						Annotations: map[string]string{
							"stork.libopenstorage.org/skipresource": "true", // Add the annotation
						},
					},
					StringData: map[string]string{
						"nfsServer":       serverAddr,
						"nfsSharePath":    path,
						"nfsMountOptions": mountOption,
						"nfsSubPath":      subpath,
						"disableSSL":      "false",
					},
					Type: corev1.SecretTypeOpaque, // Default type for opaque secrets
				}
			}
		}
		_, err = core.Instance().CreateSecret(secretObj)
		log.FailOnError(err, fmt.Sprintf("secret %v is not getting created in %v namespace", secretObj.Name, secretObj.Namespace))

		log.InfoD("Secret created successfully")

	})

	//Restore from backups taken by stork and not done by px backup
	It("Restore from backups taken by stork and not done by px backup.", func() {

		// 2. validate application
		Step("Validate applications", func() {
			log.InfoD("Validating applications")
			ValidateApplications(scheduledAppContexts)
		})

		// 3. Create Backup Location and Application Backup through Yamls
		Step("Create Backup Location through Yamls", func() {
			log.Infof("Create Backup Location through Yamls")

			backupLocation, err := applicationbackup.CreateCustomBackupLocation(backupLocationName, bkpNamespaces[0], secretName)
			log.FailOnError(err, "Failed to create backup location")
			log.Infof("Backup Location Created - %v", backupLocation)
			appBackup, bkp_create_err := applicationbackup.CreateApplicationBackup(backupName, bkpNamespaces[0], backupLocation)
			log.FailOnError(bkp_create_err, "Failed to create backup")
			log.Infof("Backup is %v", appBackup.Name)
			log.InfoD("backup successful, backup name - %v, backup location - %v", backupName, backupLocationName)
			backupNames = append(backupNames, backupName)

		})

		//4. Creating cloud credentials and backup location with the same name as the backup location created manually in step 2
		Step("Creating cloud credentials and backup location", func() {
			log.InfoD("Creating cloud credentials and backup location")
			backupLocation := "storkbackuplocation"
			backupLocationUID = uuid.New()
			for _, provider := range providers {
				if provider == drivers.ProviderAws {
					cloudCredName = fmt.Sprintf("%s-%s-%v", "cloudcred", provider, time.Now().Unix())
					cloudCredUID = uuid.New()
					backupLocationMap[backupLocationUID] = backupLocation
					err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, adminCtx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
					err = CreateBackupLocationWithContext(provider, backupLocation, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", adminCtx, true)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocation))
				} else if provider == drivers.ProviderNfs {
					err := CreateNFSBackupLocationCustomAddress(backupLocation, backupLocationUID, BackupOrgID, "", os.Getenv("NFS_SUB_PATH"), true)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %v", backupLocation))
					backupLocationMap[backupLocationUID] = backupLocation
					log.InfoD("Created Backup Location with name - %s", backupLocation)
				}
			}

		})

		//5. Check if all backups are synced or not
		Step("Check if backups created before are synced or not", func() {
			log.InfoD("Check if backups created before are synced or not")
			log.InfoD("Backup Name and default Ns are - %v and %v", backupName, bkpNamespaces[0])
			bkp_comp_err := applicationbackup.WaitForAppBackupCompletion(backupName, bkpNamespaces[0], timeout)
			log.FailOnError(bkp_comp_err, "Backup completion failed")
		})

		//6. Restore the backup if synced
		Step("Restore the backup", func() {
			log.InfoD("Restore the backup")
			err := CreateApplicationClusters(BackupOrgID, "", "", adminCtx)
			Inst().Dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			userDestClusterUid, err := Inst().Backup.GetClusterUID(adminCtx, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))
			restoreName := fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			err = CreateRestoreWithValidation(adminCtx, restoreName, backupNames[0], make(map[string]string), make(map[string]string), DestinationClusterName, userDestClusterUid, BackupOrgID, appContextsToBackup)
			dash.VerifyNotNilFatal(err, fmt.Sprintf("Validation of Restore [%s] ", restoreName))
			restoreNames = append(restoreNames, restoreName)
		})

	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)

		log.InfoD("Deleting the deployed apps after the testcase")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)

		//Cleaning up Backup Location
		log.InfoD("Delete Backuplocation created manually during Cleanup")
		err := applicationbackup.DeleteBackupLocationAndSecret(backupLocationName, bkpNamespaces[0], secretName)
		log.FailOnError(err, fmt.Sprintf("Failed to Delete BackupLocation %v", backupLocationName))
		log.InfoD("Deletion of Backuplocation created manually during Cleanup is successful")

		//Cleaning up Backups
		log.InfoD("Delete Backup created manually during Cleanup")
		err = applicationbackup.DeleteApplicationBackup(backupName, bkpNamespaces[0])
		log.FailOnError(err, fmt.Sprintf("Failed to Delete Backup %v", backupName))
		log.InfoD("Deletion of Backup created manually during Cleanup is successful")

		//Cleanup Restores
		log.InfoD("Delete Restore")
		for _, restoreNameIteration := range restoreNames {
			err = DeleteRestore(restoreNameIteration, BackupOrgID, adminCtx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting restore [%s]", restoreNameIteration))
		}

		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, adminCtx)

	})
})
