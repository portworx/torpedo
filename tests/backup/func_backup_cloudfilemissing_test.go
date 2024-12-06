package tests

import (
	"fmt"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/pure-px/torpedo/drivers"
	"github.com/pure-px/torpedo/drivers/backup"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
)

// This test case verifies if we can Add only three S3 backuplocation and take two backups as part of each BL and then delete one backup from each BL and verify
var _ = Describe("{AddOnlyThreeS3BackupLocationAndTakeTwoBackupsAsPartOfEachBLAndThenDeleteOneBackupFromEachBLAndVerify}", Label(TestCaseLabelsMap[DeleteS3BackupFilesVerifyCloudBackupMissing]...), func() {
	var (
		backupName           string
		scheduledAppContexts []*scheduler.Context
		sourceClusterUid     string
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		cloudCredName        string
		cloudCredUID         string
		bkpNamespaces        []string
		s3CloudCredName      string
		s3BackupLocationName string
		s3CloudCredUID       string
		s3BackupLocationUID  string
		providers            []string
		destClusterUid       string
		s3BucketName         string
	)
	bkpNamespaces = make([]string, 0)
	labelSelectors := make(map[string]string)
	backupLocationMap := make(map[string]string)
	backupNames := make([]string, 0)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyAddOnlyThreeS3BackupLocationAndTakeTwoBackupsAsPartOfEachBLAndThenDeleteOneBackupFromEachBLAndVerify",
			"Add only three S3 backuplocation and take two backups as part of each BL and then delete one backup from each BL and verify", nil, 300405, ABadgujar, Q2FY24)
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
		providers = GetBackupProviders()
	})
	It("Add only three S3 backuplocation and take two backups as part of each BL and then delete one backup from each BL and verify", func() {
		//2.Step-Validate Deployed Applications
		Step("Validate applications", func() {
			log.InfoD("Validating apps")
			ValidateApplications(scheduledAppContexts)
		})

		//3.Step-Creating cloud setting for aws and backup location for S3
		Step("Creating cloud setting for aws and backup location for S3", func() {
			log.InfoD("Creating cloud setting for aws and backup location for S3")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, provider := range providers {
				if provider == drivers.ProviderAws {
					s3CloudCredName = fmt.Sprintf("%s-%s-%v", "cred", "s3", RandomString(4))
					s3CloudCredUID = uuid.New()
					s3BucketName = getGlobalBucketName(provider)
					err = CreateCloudCredential(provider, s3CloudCredName, s3CloudCredUID, BackupOrgID, ctx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", s3CloudCredName, BackupOrgID, "AWS"))
					for i := 1; i <= 3; i++ {
						s3BackupLocationName = fmt.Sprintf("%s-%s-%v", "aws-s3", getGlobalBucketName(provider), i)
						s3BackupLocationUID = uuid.New()
						err = CreateS3BackupLocation(s3BackupLocationName, s3BackupLocationUID, s3CloudCredName, s3CloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of S3 backup location [%s]", s3BackupLocationName))
						backupLocationMap[s3BackupLocationUID] = s3BackupLocationName
					}
				}
			}
		})

		//4.Step-Registering Cluster for Backup Admin Context
		Step("Register cluster for backup", func() {
			adminContext, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching admin user ctx")
			err = CreateApplicationClusters(BackupOrgID, "", "", adminContext)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, adminContext)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			sourceClusterUid, err = Inst().Backup.GetClusterUID(adminContext, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
			destClusterUid, err = Inst().Backup.GetClusterUID(adminContext, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))
		})

		//5.Step-Create 2 Backups per backup location in all backup locations in Admin Context ,Delete 1 Backup File and check for cloud file missing,create restore for other one repeat for all backup locations
		Step("Taking backup of applications", func() {
			//Common Steps for all backups
			adminContext, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching admin user ctx")
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)

			for backupLocationUIDIteration, bkplocationNameIteration := range backupLocationMap {
				backupName = fmt.Sprintf("%s-%s-%v", BackupNamePrefix, bkplocationNameIteration, 1)
				err = CreateBackupWithValidation(adminContext, backupName, SourceClusterName, bkplocationNameIteration, backupLocationUIDIteration, appContextsToBackup, labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupName))
				backupNames = append(backupNames, backupName)
				tempBackupNames := make([]string, 0)
				tempBackupNames = append(tempBackupNames, backupName)
				//Delete backup file from S3 bucket
				err = DeleteFilesFromS3Bucket(s3BucketName, "metadata.json")
				log.FailOnError(err, fmt.Sprintf("Faced error while deleting the s3 backup files from bucket [%s] and backup [%s]", s3BucketName, backupName))
				log.InfoD("Deletion of backup files successful for backup - %s", backupName)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Deletion of backup files successful for backup [%s]", backupName))
				//Make 2nd S3 Backup
				backupName = fmt.Sprintf("%s-%s-%v", BackupNamePrefix, bkplocationNameIteration, 2)
				err = CreateBackupWithValidation(adminContext, backupName, SourceClusterName, bkplocationNameIteration, backupLocationUIDIteration, appContextsToBackup, labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupName))
				tempBackupNames = append(tempBackupNames, backupName)

				//Verify the backups are in CloudBackupMissing state after backup files deletion
				log.InfoD("Verify the backups are in CloudBackupMissing state after backup files deletion")
				var wg sync.WaitGroup
				adminContext, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching admin user ctx")
				for i, backupName := range tempBackupNames {
					if i%2 == 0 {
						wg.Add(1)
						go func(backupName string) {
							defer GinkgoRecover()
							defer wg.Done()
							bkpUid, err := Inst().Backup.GetBackupUID(adminContext, backupName, BackupOrgID)
							log.FailOnError(err, "Fetching backup uid")
							backupInspectRequest := &api.BackupInspectRequest{
								Name:  backupName,
								Uid:   bkpUid,
								OrgId: BackupOrgID,
							}
							requiredStatus := api.BackupInfo_StatusInfo_CloudBackupMissing
							backupCloudBackupMissingCheckFunc := func() (interface{}, bool, error) {
								resp, err := Inst().Backup.InspectBackup(adminContext, backupInspectRequest)
								if err != nil {
									return "", false, err
								}
								actual := resp.GetBackup().GetStatus().Status
								if actual == requiredStatus {
									return "", false, nil
								}
								return "", true, fmt.Errorf("backup status for [%s] expected was [%v] but got [%s]", backupName, requiredStatus, actual)
							}
							_, err = DoRetryWithTimeoutWithGinkgoRecover(backupCloudBackupMissingCheckFunc, 20*time.Minute, 30*time.Second)
							dash.VerifyFatal(err, nil, fmt.Sprintf("Verfiying backup %s is in CloudBackup missing state", backupName))
						}(backupName)
					} else { //Step-Verify if the restores for the other aws backup is possible , so that we know cloudbackup objects are not missing for it
						restoreName := fmt.Sprintf("%s-%s", RestoreNamePrefix, backupName)
						appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
						err = CreateRestoreWithValidation(adminContext, restoreName, backupName, make(map[string]string), make(map[string]string), DestinationClusterName, destClusterUid, BackupOrgID, appContextsToBackup)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s]", restoreName))
					}
				}
				wg.Wait()

			}
		})

	})
	JustAfterEach(func() {

		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		log.InfoD("Deleting the deployed apps after the testcase")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)

		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)

	})
})

// This test case verifies if we can Add multiple NFS backup location and then take multiple backup as part of each BL and delete some backups from NFS location
var _ = Describe("{AddMultipleNFSBackupLocationAndTakeMultipleBackupAsPartOfEachBLAndDeleteSomeBackupsFromNFSLocation}", Label(TestCaseLabelsMap[RemoveJSONFilesFromNFSBackupLocation]...), func() {
	var (
		backupName           string
		scheduledAppContexts []*scheduler.Context
		sourceClusterUid     string
		destClusterUid       string
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		cloudCredName        string
		cloudCredUID         string
		backupLocationUID    string
		bkpLocationName      string
		globalBucket         string
		bkpNamespaces        []string
	)
	bkpNamespaces = make([]string, 0)
	labelSelectors := make(map[string]string)
	backupLocationMap := make(map[string]string)
	backupNames := make([]string, 0)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyAddMultipleNFSBackupLocationAndTakeMultipleBackupAsPartOfEachBLAndDeleteSomeBackupsFromNFSLocation",
			"Add multiple NFS backup location and take multiple backup as part of each BL and delete some backups from NFS location", nil, 300402, ABadgujar, Q2FY24)
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
	It("Add multiple NFS backup location and take multiple backup as part of each BL and delete some backups from NFS location", func() {
		//2.Step-Validate Deployed Applications
		Step("Validate applications", func() {
			log.InfoD("Validating apps")
			ValidateApplications(scheduledAppContexts)
		})

		//3.Step-Create 2 NFS Backup Locations
		Step("Create 2 NFS Backup Locations", func() {

			globalBucket = getGlobalBucketName(drivers.ProviderNfs)

			log.InfoD("Creating NFS backup location-1")
			bkpLocationName = fmt.Sprintf("%s-%s-%v-1", "nfs", globalBucket, RandomString(6))
			backupLocationUID = uuid.New()
			backupLocationMap[backupLocationUID] = bkpLocationName
			err := CreateNFSBackupLocation(bkpLocationName, backupLocationUID, BackupOrgID, " ", globalBucket, true)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creating NFS backup location %s", bkpLocationName))

			log.InfoD("Creating NFS backup location-2")
			bkpLocationName = fmt.Sprintf("%s-%s-%v-2", "nfs", globalBucket, RandomString(6))
			backupLocationUID = uuid.New()
			backupLocationMap[backupLocationUID] = bkpLocationName
			err = CreateNFSBackupLocation(bkpLocationName, backupLocationUID, BackupOrgID, " ", globalBucket, true)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creating NFS backup location %s", bkpLocationName))

		})

		//4.Step-Registering Cluster for Backup Admin Context
		Step("Register cluster for backup", func() {
			adminContext, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching admin user ctx")
			err = CreateApplicationClusters(BackupOrgID, "", "", adminContext)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, adminContext)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			sourceClusterUid, err = Inst().Backup.GetClusterUID(adminContext, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
			destClusterUid, err = Inst().Backup.GetClusterUID(adminContext, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))
		})

		//5.Step-Create 2 Backups per backup location in all backup locations in Admin Context
		Step("Taking backup of applications", func() {
			//Common Steps for all backups
			adminContext, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching admin user ctx")
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)

			for backupLocationUIDIteration, bkplocationNameIteration := range backupLocationMap {
				for noOfBackups := 1; noOfBackups <= 2; noOfBackups++ {
					backupName = fmt.Sprintf("%s-%s-%v", BackupNamePrefix, bkplocationNameIteration, noOfBackups)
					err = CreateBackupWithValidation(adminContext, backupName, SourceClusterName, bkplocationNameIteration, backupLocationUIDIteration, appContextsToBackup, labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "")
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupName))
					backupNames = append(backupNames, backupName)
				}
			}
		})

		//6.Remove the JSON files from the NFS backup location for half the backups
		Step("Remove the JSON files from the NFS backup location for half the backups", func() {
			log.InfoD("Remove the JSON files from the NFS backup location for half the backups")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for i, backupName := range backupNames {
				if i%2 == 0 {
					backupUID, err := Inst().Backup.GetBackupUID(ctx, backupName, BackupOrgID)
					log.FailOnError(err, fmt.Sprintf("Getting UID for backup %v", backupName))
					backupInspectRequest := &api.BackupInspectRequest{
						Name:  backupName,
						Uid:   backupUID,
						OrgId: BackupOrgID,
					}
					resp, err := Inst().Backup.InspectBackup(ctx, backupInspectRequest)
					log.FailOnError(err, fmt.Sprintf("error inspecting backup %v", backupName))
					currentBackupPath := globalBucket + "/" + resp.Backup.BackupPath
					log.Infof("Deleting the JSON files from the NFS backup location for backup %v", backupName)
					err = DeleteFilesFromNFSLocation(currentBackupPath, "*.json")
					log.FailOnError(err, fmt.Sprintf("Faced error while deleting the JSON files from path [%s]", currentBackupPath))
				}
			}
		})

		//7.Step-Verify the backups are in CloudBackupMissing state after JSON file deletion
		Step("Verify the backups are in CloudBackupMissing state after JSON file deletion", func() {
			log.InfoD("Verify the backups are in CloudBackupMissing state after JSON file deletion")
			var wg sync.WaitGroup
			adminContext, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching admin user ctx")
			for i, backupName := range backupNames {
				if i%2 == 0 {
					wg.Add(1)
					go func(backupName string) {
						defer GinkgoRecover()
						defer wg.Done()
						bkpUid, err := Inst().Backup.GetBackupUID(adminContext, backupName, BackupOrgID)
						log.FailOnError(err, "Fetching backup uid")
						backupInspectRequest := &api.BackupInspectRequest{
							Name:  backupName,
							Uid:   bkpUid,
							OrgId: BackupOrgID,
						}
						requiredStatus := api.BackupInfo_StatusInfo_CloudBackupMissing
						backupCloudBackupMissingCheckFunc := func() (interface{}, bool, error) {
							resp, err := Inst().Backup.InspectBackup(adminContext, backupInspectRequest)
							if err != nil {
								return "", false, err
							}
							actual := resp.GetBackup().GetStatus().Status
							if actual == requiredStatus {
								return "", false, nil
							}
							return "", true, fmt.Errorf("backup status for [%s] expected was [%v] but got [%s]", backupName, requiredStatus, actual)
						}
						_, err = DoRetryWithTimeoutWithGinkgoRecover(backupCloudBackupMissingCheckFunc, 20*time.Minute, 30*time.Second)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Verfiying backup %s is in CloudBackup missing state", backupName))
					}(backupName)
				}
			}
			wg.Wait()
		})

		//8.Step-Verify if the restores for the other 2 backups is possible , so that we know cloudbackup objects are not missing for them
		Step("Verify if the restores for the other 2 backups is possible , so that we know cloudbackup objects are not missing for them", func() {
			log.InfoD("Verify if the restores for the other 2 backups is possible , so that we know cloudbackup objects are not missing for them")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for i, backupName := range backupNames {
				if i%2 == 1 {
					restoreName := fmt.Sprintf("%s-%s", RestoreNamePrefix, backupName)
					appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
					err = CreateRestoreWithValidation(ctx, restoreName, backupName, make(map[string]string), make(map[string]string), DestinationClusterName, destClusterUid, BackupOrgID, appContextsToBackup)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s]", restoreName))
				}
			}
		})
	})
	JustAfterEach(func() {

		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		log.InfoD("Deleting the deployed apps after the testcase")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)
		providers := GetBackupProviders()

		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
		log.InfoD("Delete the local bucket created")
		for _, provider := range providers {
			DeleteBucket(provider, globalBucket)
			log.Infof("bucket deleted - %s", globalBucket)
		}
	})
})
