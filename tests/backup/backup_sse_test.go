package tests

import (
	//"github.com/aws/aws-sdk-go/aws"
	//"github.com/aws/aws-sdk-go/service/s3"

	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/sched-ops/k8s/storage"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/scheduler/k8s"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//"github.com/portworx/torpedo/drivers/scheduler/k8s"
	//v1 "k8s.io/api/core/v1"
	storageApi "k8s.io/api/storage/v1"
	//metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sync"
	"time"
	//api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/torpedo/drivers/backup"

	"github.com/portworx/torpedo/pkg/log"
	//"github.com/portworx/torpedo/drivers/scheduler"
	. "github.com/portworx/torpedo/tests"
)

var _ = Describe("{sseS3encryption}", func() {

	var (
		scheduledAppContexts []*scheduler.Context
		customBucket         string
		//customBucket1              string
		backupLocationUID          string
		cloudCredUID               string
		cloudCredUID1              string
		backupNames1               []string
		cloudCredUidList           []string
		customBackupLocationName   string
		customBucketsWithOutPolicy []string
		customBucketsWithPolicy    []string
		backupLocationUID1         string
		customBackupLocationName1  string
		credName                   string
		credName1                  string
		customBuckets              []string
		bkpNamespaces              []string
		sourceScName               *storageApi.StorageClass
		scCount                    int
		scNames                    []string
		clusterStatus              api.ClusterInfo_StatusInfo_Status
		clusterUid                 string
		restoreList                []string
		backupNames                []string
	)

	storageClassMapping := make(map[string]string)
	namespaceMap := make(map[string]string)
	params := make(map[string]string)
	k8sStorage := storage.Instance()

	JustBeforeEach(func() {
		StartTorpedoTest("MultipleCustomRestoreSameTimeDiffStorageClassMapping",
			"Issue multiple custom restores at the same time using different storage class mapping", nil, 58052)
		log.InfoD("Deploy applications needed for backup")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		numberOfNameSpace := 1
		for i := 0; i < numberOfNameSpace; i++ {
			taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = appReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
				scheduledAppContexts = append(scheduledAppContexts, ctx)
			}
		}
	})
	It("sse-s3 encryption", func() {
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		providers := getProviders()
		Step("Validate applications", func() {
			log.InfoD("Validate applications")
			ValidateApplications(scheduledAppContexts)
		})
		Step("Register cluster for backup", func() {
			err = CreateApplicationClusters(orgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(orgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, orgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})
		Step("Adding Credentials and Registering Backup Location", func() {
			// Create a bucket with and without deny policy

			// Create bucket BK1 without deny policy false
			// Create bucket BK2 with deny policy true
			// Create BL1 with BK1 SSE false
			// Create BL2 with BK2 SSE true
			// Take backup with BL1
			// Take backup with BL2
			// Perform parallel restore same cluster
			// Perform parallel restore alternate cluster
			// Reboot stork
			// Reboot backup
			// Perform further backup on BK1
			// Perform further backup on BK2
			// ADD new BL and take backup
			// Make SSE True take backup on BL1
			// Remove deny policy and take backup on BL2

			//bucketMap := map[string]bool{
			//	"sse-bucket-with-policy-1": false,
			//	//"sse-bucket-with-policy-2": true,
			//	//"sse-bucket-without-policy": false,
			//}

			// Create bucket BK1 with deny policy true
			bucketWithOutPolicy := "sse-bucket-without-policy"
			for _, provider := range providers {
				customBucket = GetCustomBucketName(provider, bucketWithOutPolicy)
				customBucketsWithOutPolicy = append(customBucketsWithOutPolicy, customBucket)
			}

			// Create bucket BK2 with deny policy true
			bucketWithPolicy := "sse-bucket-with-policy"
			for _, provider := range providers {
				for i := 0; i < 1; i++ {
					customBucket = GetCustomBucketName(provider, bucketWithPolicy)
					//sseDetails, err := s3utils.GetS3SSEDetailsFromEnv()
					//log.FailOnError(err, "Failed to generate s3 bucket policy check for the correctness of policy parameters")
					policy, err := GenerateS3BucketPolicy("DenyNonAES256Uploads", "s3:x-amz-server-side-encryption=AES256", customBucket)
					log.FailOnError(err, "Failed to generate s3 bucket policy check for the correctness of policy parameters")
					err = UpdateS3BucketPolicy(customBucket, policy)
					log.FailOnError(err, "Failed to apply bucket policy")
					customBuckets = append(customBuckets, customBucket)
					log.Infof("Updated S3 bucket policy - %s", globalAWSBucketName)
				}
			}

			// Create BL1 with BK1 SSE false
			for _, provider := range providers {
				for _, customBucket := range customBucketsWithOutPolicy {
					cloudCredUID1 = uuid.New()
					cloudCredUidList = append(cloudCredUidList, cloudCredUID)
					backupLocationUID1 = uuid.New()
					credName1 = fmt.Sprintf("autogenerated-cred-%v", time.Now().Unix())
					err := CreateCloudCredential(provider, credName1, cloudCredUID1, orgID, ctx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", credName, orgID, provider))
					log.InfoD("Created Cloud Credentials with name - %s", credName1)
					customBackupLocationName1 = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
					//err = CreateS3BackupLocationWithSseType(customBackupLocationName, backupLocationUID, credName, cloudCredUID, customBucket, orgID, "", api.S3Config_SSE_S3)
					err = CreateBackupLocation(provider, customBackupLocationName1, backupLocationUID1, credName1, cloudCredUID1, customBucket, orgID, "")
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", customBackupLocationName))
					log.Infof("created backup location successfully")
				}
			}

			// Create BL2 with BK2 SSE true
			for _, provider := range providers {
				cloudCredUID = uuid.New()
				cloudCredUidList = append(cloudCredUidList, cloudCredUID)
				backupLocationUID = uuid.New()
				credName = fmt.Sprintf("autogenerated-cred-%v", time.Now().Unix())
				err := CreateCloudCredential(provider, credName, cloudCredUID, orgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", credName, orgID, provider))
				log.InfoD("Created Cloud Credentials with name - %s", credName)
				customBackupLocationName = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
				err = CreateS3BackupLocationWithSseType(customBackupLocationName, backupLocationUID, credName, cloudCredUID, customBuckets[0], orgID, "", api.S3Config_SSE_S3)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", customBackupLocationName))
				log.Infof("created backup location successfully")
			}

			for _, provider := range providers {
				err = UpdateBackupLocation(provider, customBackupLocationName1, backupLocationUID1, orgID, credName1, cloudCredUID1, ctx, api.S3Config_SSE_S3)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Updation of backuplocation [%s]", customBackupLocationName1))
			}

			log.InfoD("Kill stork when backup in progress")
			log.Infof("Kill stork when backup in progress")
			err := DeletePodWithLabelInNamespace(getPXNamespace(), storkLabel)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Killing stork while backups %s is in progress", backupNames))

			// Take backup with BL1
			log.InfoD("Taking backup of application for different combination of restores")
			backupName := fmt.Sprintf("%s-%s-%v", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{bkpNamespaces[0]})
			log.Infof(clusterUid)
			err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, customBackupLocationName1, backupLocationUID1, appContextsToBackup, make(map[string]string), orgID, clusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
			backupNames1 = append(backupNames1, backupName)

			log.Infof("Restart backup pod")
			log.InfoD("Restart backup pod")
			backupPodLabel := make(map[string]string)
			backupPodLabel["app"] = "px-backup"
			pxbNamespace, err := backup.GetPxBackupNamespace()
			dash.VerifyFatal(err, nil, "Getting px-backup namespace")
			err = DeletePodWithLabelInNamespace(pxbNamespace, backupPodLabel)
			dash.VerifyFatal(err, nil, "Restart backup pod when backup sharing is in-progress")
			err = ValidatePodByLabel(backupPodLabel, pxbNamespace, 5*time.Minute, 30*time.Second)
			log.FailOnError(err, "Checking if px-backup pod is in running state")

			// Take backup with BL2
			log.InfoD("Taking backup of application for different combination of restores")
			backupName = fmt.Sprintf("%s-%s-%v", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
			appContextsToBackup = FilterAppContextsByNamespace(scheduledAppContexts, []string{bkpNamespaces[0]})
			log.Infof(clusterUid)
			err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, customBackupLocationName, backupLocationUID, appContextsToBackup, make(map[string]string), orgID, clusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
			backupNames = append(backupNames, backupName)

			Step("Create new storage class on source cluster for storage class mapping for restore", func() {
				log.InfoD("Create new storage class on source cluster for storage class mapping for restore")
				scCount = 3
				for i := 0; i < scCount; i++ {
					scName := fmt.Sprintf("replica-sc-%d-%v", time.Now().Unix(), i)
					params["repl"] = "2"
					v1obj := metaV1.ObjectMeta{
						Name: scName,
					}
					reclaimPolicyDelete := v1.PersistentVolumeReclaimDelete
					bindMode := storageApi.VolumeBindingImmediate
					scObj := storageApi.StorageClass{
						ObjectMeta:        v1obj,
						Provisioner:       k8s.CsiProvisioner,
						Parameters:        params,
						ReclaimPolicy:     &reclaimPolicyDelete,
						VolumeBindingMode: &bindMode,
					}

					_, err := k8sStorage.CreateStorageClass(&scObj)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creating new storage class %v on source cluster %s", scName, SourceClusterName))
					scNames = append(scNames, scName)
				}
			})

			Step("Multiple restore for same backup in different storage class in same cluster at the same time", func() {
				log.InfoD(fmt.Sprintf("Multiple restore for same backup into %d different storage class in same cluster at the same time", scCount))
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-central-admin ctx")
				pvcs, err := core.Instance().GetPersistentVolumeClaims(bkpNamespaces[0], make(map[string]string))
				singlePvc := pvcs.Items[0]
				sourceScName, err = core.Instance().GetStorageClassForPVC(&singlePvc)
				var wg sync.WaitGroup
				//for _, scName := range scNames {
				for i := 0; i < 1; i++ {
					storageClassMapping[sourceScName.Name] = scNames[i]
					time.Sleep(2)
					namespaceMap[bkpNamespaces[0]] = fmt.Sprintf("new-namespace-%v", time.Now().Unix())
					restoreName := fmt.Sprintf("restore-new-storage-class-%s-%s", scNames[i], RestoreNamePrefix)
					restoreList = append(restoreList, restoreName)
					wg.Add(1)
					go func(scName string) {
						defer GinkgoRecover()
						defer wg.Done()
						err = CreateRestore(restoreName, backupNames1[0], namespaceMap, SourceClusterName, orgID, ctx, storageClassMapping)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Restoring backup %v using storage class %v", backupName, scName))
					}(scNames[i])
				}
				wg.Wait()
			})

			for _, provider := range providers {
				//err = UpdateBackupLocation(provider, customBackupLocationName1, backupLocationUID1, orgID, credName1, cloudCredUID1, customBucket1, ctx, api.S3Config_SSE_S3)
				err = UpdateBackupLocation(provider, customBackupLocationName1, backupLocationUID1, orgID, credName1, cloudCredUID1, ctx, api.S3Config_SSE_S3)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Updation of backuplocation [%s]", customBackupLocationName1))
			}

			// Take backup with BL1
			log.InfoD("Taking backup of application for different combination of restores")
			backupName = fmt.Sprintf("%s-%s-%v", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
			appContextsToBackup = FilterAppContextsByNamespace(scheduledAppContexts, []string{bkpNamespaces[0]})
			log.Infof(clusterUid)
			err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, customBackupLocationName1, backupLocationUID1, appContextsToBackup, make(map[string]string), orgID, clusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
			backupNames = append(backupNames, backupName)

			err = RemoveS3BucketPolicy(customBucketsWithPolicy[0])
			dash.VerifySafely(err, nil, fmt.Sprintf("Verify removal of S3 bucket policy"))

			// Take backup with BL2
			log.InfoD("Taking backup of application for different combination of restores")
			backupName = fmt.Sprintf("%s-%s-%v", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
			appContextsToBackup = FilterAppContextsByNamespace(scheduledAppContexts, []string{bkpNamespaces[0]})
			log.Infof(clusterUid)
			err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, customBackupLocationName, backupLocationUID, appContextsToBackup, make(map[string]string), orgID, clusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
			backupNames = append(backupNames, backupName)

			//for _, provider := range providers {
			//	err = UpdateBackupLocation(provider, customBackupLocationName1, backupLocationUID1, orgID, ctx, api.S3Config_SSE_S3)
			//	dash.VerifyFatal(err, nil, fmt.Sprintf("Updation of backuplocation [%s]", customBackupLocationName1))
			//}
		})
	})
})

var _ = Describe("{sseS3encryption1}", func() {

	var (
		scheduledAppContexts       []*scheduler.Context
		customBucket               string
		backupLocationUID          string
		cloudCredUID               string
		backupNames1               []string
		cloudCredUidList           []string
		customBackupLocationName   string
		customBucketsWithOutPolicy []string
		customBucketsWithPolicy    []string
		backupLocationUID1         string
		customBackupLocationName1  string
		credName                   string
		customBuckets              []string
		bkpNamespaces              []string
		sourceScName               *storageApi.StorageClass
		scCount                    int
		scNames                    []string
		clusterStatus              api.ClusterInfo_StatusInfo_Status
		clusterUid                 string
		restoreList                []string
		backupNames                []string
	)

	storageClassMapping := make(map[string]string)
	namespaceMap := make(map[string]string)
	params := make(map[string]string)
	k8sStorage := storage.Instance()

	JustBeforeEach(func() {
		StartTorpedoTest("MultipleCustomRestoreSameTimeDiffStorageClassMapping",
			"Issue multiple custom restores at the same time using different storage class mapping", nil, 58052)
		log.InfoD("Deploy applications needed for backup")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		numberOfNameSpace := 1
		for i := 0; i < numberOfNameSpace; i++ {
			taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = appReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
				scheduledAppContexts = append(scheduledAppContexts, ctx)
			}
		}
	})
	It("sse-s3 encryption", func() {
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		providers := getProviders()
		Step("Validate applications", func() {
			log.InfoD("Validate applications")
			ValidateApplications(scheduledAppContexts)
		})
		Step("Register cluster for backup", func() {
			err = CreateApplicationClusters(orgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(orgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, orgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})
		Step("Adding Credentials and Registering Backup Location", func() {
			// Create a bucket with and without deny policy

			// Create bucket BK1 without deny policy false
			// Create bucket BK2 with deny policy true
			// Create BL1 with BK1 SSE false
			// Create BL2 with BK2 SSE true
			// Take backup with BL1
			// Take backup with BL2
			// Perform parallel restore same cluster
			// Perform parallel restore alternate cluster
			// Reboot stork
			// Reboot backup
			// Perform further backup on BK1
			// Perform further backup on BK2
			// ADD new BL and take backup
			// Make SSE True take backup on BL1
			// Remove deny policy and take backup on BL2

			//bucketMap := map[string]bool{
			//	"sse-bucket-with-policy-1": false,
			//	//"sse-bucket-with-policy-2": true,
			//	//"sse-bucket-without-policy": false,
			//}

			// Create Bucket1 with deny policy true
			bucketWithOutPolicy := "sse-bucket-without-policy"
			for _, provider := range providers {
				customBucket := GetCustomBucketName(provider, bucketWithOutPolicy)
				customBucketsWithOutPolicy = append(customBucketsWithOutPolicy, customBucket)
			}

			// Create Bucket2 with deny policy true
			bucketWithPolicy := "sse-bucket-with-policy"
			for _, provider := range providers {
				for i := 0; i < 1; i++ {
					customBucket = GetCustomBucketName(provider, bucketWithPolicy)
					//sseDetails, err := s3utils.GetS3SSEDetailsFromEnv()
					//log.FailOnError(err, "Failed to generate s3 bucket policy check for the correctness of policy parameters")
					policy, err := GenerateS3BucketPolicy("DenyNonAES256Uploads", "s3:x-amz-server-side-encryption=AES256", customBucket)
					log.FailOnError(err, "Failed to generate s3 bucket policy check for the correctness of policy parameters")
					err = UpdateS3BucketPolicy(customBucket, policy)
					log.FailOnError(err, "Failed to apply bucket policy")
					customBuckets = append(customBuckets, customBucket)
					log.Infof("Updated S3 bucket policy - %s", globalAWSBucketName)
				}
			}

			// Create BL1 with BK1 SSE false
			for _, provider := range providers {
				for _, customBucket := range customBucketsWithOutPolicy {
					cloudCredUID = uuid.New()
					cloudCredUidList = append(cloudCredUidList, cloudCredUID)
					backupLocationUID1 = uuid.New()
					credName = fmt.Sprintf("autogenerated-cred-%v", time.Now().Unix())
					err := CreateCloudCredential(provider, credName, cloudCredUID, orgID, ctx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", credName, orgID, provider))
					log.InfoD("Created Cloud Credentials with name - %s", credName)
					customBackupLocationName1 = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
					//err = CreateS3BackupLocationWithSseType(customBackupLocationName, backupLocationUID, credName, cloudCredUID, customBucket, orgID, "", api.S3Config_SSE_S3)
					err = CreateBackupLocation(provider, customBackupLocationName1, backupLocationUID1, credName, cloudCredUID, customBucket, orgID, "")
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", customBackupLocationName))
					log.Infof("created backup location successfully")
				}
			}

			// Create BL2 with BK2 SSE true
			for _, provider := range providers {
				cloudCredUID = uuid.New()
				cloudCredUidList = append(cloudCredUidList, cloudCredUID)
				backupLocationUID = uuid.New()
				credName = fmt.Sprintf("autogenerated-cred-%v", time.Now().Unix())
				err := CreateCloudCredential(provider, credName, cloudCredUID, orgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", credName, orgID, provider))
				log.InfoD("Created Cloud Credentials with name - %s", credName)
				customBackupLocationName = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
				err = CreateS3BackupLocationWithSseType(customBackupLocationName, backupLocationUID, credName, cloudCredUID, customBuckets[0], orgID, "", api.S3Config_SSE_S3)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", customBackupLocationName))
				log.Infof("created backup location successfully")
			}

			log.InfoD("Kill stork when backup in progress")
			log.Infof("Kill stork when backup in progress")
			err := DeletePodWithLabelInNamespace(getPXNamespace(), storkLabel)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Killing stork while backups %s is in progress", backupNames))

			// Take backup with BL1
			log.InfoD("Taking backup of application for different combination of restores")
			backupName := fmt.Sprintf("%s-%s-%v", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{bkpNamespaces[0]})
			log.Infof(clusterUid)
			err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, customBackupLocationName1, backupLocationUID1, appContextsToBackup, make(map[string]string), orgID, clusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
			backupNames1 = append(backupNames1, backupName)

			log.Infof("Restart backup pod")
			log.InfoD("Restart backup pod")
			backupPodLabel := make(map[string]string)
			backupPodLabel["app"] = "px-backup"
			pxbNamespace, err := backup.GetPxBackupNamespace()
			dash.VerifyFatal(err, nil, "Getting px-backup namespace")
			err = DeletePodWithLabelInNamespace(pxbNamespace, backupPodLabel)
			dash.VerifyFatal(err, nil, "Restart backup pod when backup sharing is in-progress")
			err = ValidatePodByLabel(backupPodLabel, pxbNamespace, 5*time.Minute, 30*time.Second)
			log.FailOnError(err, "Checking if px-backup pod is in running state")

			// Take backup with BL2
			log.InfoD("Taking backup of application for different combination of restores")
			backupName = fmt.Sprintf("%s-%s-%v", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
			appContextsToBackup = FilterAppContextsByNamespace(scheduledAppContexts, []string{bkpNamespaces[0]})
			log.Infof(clusterUid)
			err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, customBackupLocationName, backupLocationUID, appContextsToBackup, make(map[string]string), orgID, clusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
			backupNames = append(backupNames, backupName)

			Step("Create new storage class on source cluster for storage class mapping for restore", func() {
				log.InfoD("Create new storage class on source cluster for storage class mapping for restore")
				scCount = 3
				for i := 0; i < scCount; i++ {
					scName := fmt.Sprintf("replica-sc-%d-%v", time.Now().Unix(), i)
					params["repl"] = "2"
					v1obj := metaV1.ObjectMeta{
						Name: scName,
					}
					reclaimPolicyDelete := v1.PersistentVolumeReclaimDelete
					bindMode := storageApi.VolumeBindingImmediate
					scObj := storageApi.StorageClass{
						ObjectMeta:        v1obj,
						Provisioner:       k8s.CsiProvisioner,
						Parameters:        params,
						ReclaimPolicy:     &reclaimPolicyDelete,
						VolumeBindingMode: &bindMode,
					}

					_, err := k8sStorage.CreateStorageClass(&scObj)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creating new storage class %v on source cluster %s", scName, SourceClusterName))
					scNames = append(scNames, scName)
				}
			})

			Step("Multiple restore for same backup in different storage class in same cluster at the same time", func() {
				log.InfoD(fmt.Sprintf("Multiple restore for same backup into %d different storage class in same cluster at the same time", scCount))
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-central-admin ctx")
				pvcs, err := core.Instance().GetPersistentVolumeClaims(bkpNamespaces[0], make(map[string]string))
				singlePvc := pvcs.Items[0]
				sourceScName, err = core.Instance().GetStorageClassForPVC(&singlePvc)
				var wg sync.WaitGroup
				//for _, scName := range scNames {
				for i := 0; i < 1; i++ {
					storageClassMapping[sourceScName.Name] = scNames[i]
					time.Sleep(2)
					namespaceMap[bkpNamespaces[0]] = fmt.Sprintf("new-namespace-%v", time.Now().Unix())
					restoreName := fmt.Sprintf("restore-new-storage-class-%s-%s", scNames[i], RestoreNamePrefix)
					restoreList = append(restoreList, restoreName)
					wg.Add(1)
					go func(scName string) {
						defer GinkgoRecover()
						defer wg.Done()
						err = CreateRestore(restoreName, backupNames1[0], namespaceMap, SourceClusterName, orgID, ctx, storageClassMapping)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Restoring backup %v using storage class %v", backupName, scName))
					}(scNames[i])
				}
				wg.Wait()
			})

			for _, provider := range providers {
				err = UpdateBackupLocation(provider, customBackupLocationName1, backupLocationUID1, orgID, credName, cloudCredUID, ctx, api.S3Config_SSE_S3)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Updation of backuplocation [%s]", customBackupLocationName1))
			}

			// Take backup with BL1
			log.InfoD("Taking backup of application for different combination of restores")
			backupName = fmt.Sprintf("%s-%s-%v", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
			appContextsToBackup = FilterAppContextsByNamespace(scheduledAppContexts, []string{bkpNamespaces[0]})
			log.Infof(clusterUid)
			err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, customBackupLocationName1, backupLocationUID1, appContextsToBackup, make(map[string]string), orgID, clusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
			backupNames = append(backupNames, backupName)

			err = RemoveS3BucketPolicy(customBucketsWithPolicy[0])
			dash.VerifySafely(err, nil, fmt.Sprintf("Verify removal of S3 bucket policy"))

			// Take backup with BL2

			log.InfoD("Taking backup of application for different combination of restores")
			backupName = fmt.Sprintf("%s-%s-%v", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
			appContextsToBackup = FilterAppContextsByNamespace(scheduledAppContexts, []string{bkpNamespaces[0]})
			log.Infof(clusterUid)
			err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, customBackupLocationName, backupLocationUID, appContextsToBackup, make(map[string]string), orgID, clusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
			backupNames = append(backupNames, backupName)

			//for _, provider := range providers {
			//	err = UpdateBackupLocation(provider, customBackupLocationName1, backupLocationUID1, orgID, ctx, api.S3Config_SSE_S3)
			//	dash.VerifyFatal(err, nil, fmt.Sprintf("Updation of backuplocation [%s]", customBackupLocationName1))
			//}
		})
	})
})
