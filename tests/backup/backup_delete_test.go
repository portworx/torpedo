package tests

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/portworx/torpedo/drivers"
	"golang.org/x/sync/errgroup"

	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

// IssueDeleteOfIncrementalBackupsAndRestore Issues delete of incremental backups in between and tries to restore from
// the newest backup.
var _ = Describe("{IssueDeleteOfIncrementalBackupsAndRestore}", Label(TestCaseLabelsMap[IssueDeleteOfIncrementalBackupsAndRestore]...), func() {
	var (
		credName                 string
		clusterUid               string
		cloudCredUID             string
		fullBackupName           string
		restoreName              string
		backupLocationUID        string
		customBackupLocationName string
		incrementalBackupName    string
		restoreNames             []string
		cloudCredUidList         []string
		namespaceMapping         map[string]string
		scheduledAppContexts     []*scheduler.Context
		clusterStatus            api.ClusterInfo_StatusInfo_Status
		controlChannel           chan string
		errorGroup               *errgroup.Group
	)
	labelSelectors := make(map[string]string)
	backupNames := make([]string, 0)
	incrementalBackupNames := make([]string, 0)
	incrementalBackupNames2 := make([]string, 0)
	var bkpNamespaces = make([]string, 0)
	backupLocationMap := make(map[string]string)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("IssueDeleteOfIncrementalBackupsAndRestore",
			"Issue delete of incremental backups and try to restore the newest backup", nil, 58056, Kshithijiyer, Q1FY24)
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

	It("Issue delete of incremental backups and try to restore the newest backup", func() {
		providers := GetBackupProviders()
		Step("Validate applications", func() {
			log.InfoD("Validate applications")
			ctx, _ := backup.GetAdminCtxFromSecret()
			controlChannel, errorGroup = ValidateApplicationsStartData(scheduledAppContexts, ctx)
		})

		Step("Adding Credentials and Registering Backup Location", func() {
			log.InfoD("Creating cloud credentials and backup location")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, provider := range providers {
				cloudCredUID = uuid.New()
				cloudCredUidList = append(cloudCredUidList, cloudCredUID)
				backupLocationUID = uuid.New()
				credName = fmt.Sprintf("autogenerated-cred-%v", time.Now().Unix())
				err := CreateCloudCredential(provider, credName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", credName, BackupOrgID, provider))
				log.InfoD("Created Cloud Credentials with name - %s", credName)
				cloudCredCreateStatus := func() (interface{}, bool, error) {
					ok, err := IsCloudCredPresent(credName, ctx, BackupOrgID)
					if err != nil {
						return "", true, fmt.Errorf("cloud cred %s is not created with error %v", credName, err)
					}
					if ok {
						return "", false, nil
					}
					return "", true, fmt.Errorf("cloud cred %s is created yet", credName)
				}
				_, err = DoRetryWithTimeoutWithGinkgoRecover(cloudCredCreateStatus, 10*time.Minute, 30*time.Second)
				customBackupLocationName = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
				err = CreateBackupLocation(provider, customBackupLocationName, backupLocationUID, credName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", customBackupLocationName))
				backupLocationMap[backupLocationUID] = customBackupLocationName
				log.InfoD("Created Backup Location with name - %s", customBackupLocationName)
			}
		})

		Step("Register source and destination cluster for backup", func() {
			log.InfoD("Registering Source and Destination clusters and verifying the status")
			// Registering for admin user
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})

		Step("Taking backup of applications", func() {
			log.InfoD("Taking backup of applications")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			// Full backup
			for _, namespace := range bkpNamespaces {
				fullBackupName = fmt.Sprintf("%s-%s-%v", "full-backup", namespace, time.Now().Unix())
				backupNames = append(backupNames, fullBackupName)
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
				err := CreateBackupWithValidation(ctx, fullBackupName, SourceClusterName, customBackupLocationName, backupLocationUID, appContextsToBackup, labelSelectors, BackupOrgID, clusterUid, "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of full backup [%s]", fullBackupName))
			}

			// Incremental backup set 1
			for _, namespace := range bkpNamespaces {
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
				incrementalBackupName, err = CreateBackupUntilIncrementalBackup(ctx, appContextsToBackup[0], customBackupLocationName, backupLocationUID, labelSelectors, BackupOrgID, clusterUid)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating incremental backup [%s]", incrementalBackupName))
				incrementalBackupNames = append(incrementalBackupNames, incrementalBackupName)
			}

			// Incremental backup set 2
			for _, namespace := range bkpNamespaces {
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
				incrementalBackupName, err = CreateBackupUntilIncrementalBackup(ctx, appContextsToBackup[0], customBackupLocationName, backupLocationUID, labelSelectors, BackupOrgID, clusterUid)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating incremental backup [%s]", incrementalBackupName))
				incrementalBackupNames2 = append(incrementalBackupNames2, incrementalBackupName)
			}

			log.InfoD("List of backups - %v", backupNames)
			log.InfoD("List of Incremental backups Set 1 - %v", incrementalBackupNames)
			log.InfoD("List of Incremental backups Set 2 - %v", incrementalBackupNames2)
		})

		Step("Deleting incremental backup", func() {
			log.InfoD("Deleting incremental backups")
			backupDriver := Inst().Backup
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, backupName := range incrementalBackupNames {
				log.Infof("About to delete backup - %s", backupName)
				backupUID, err := backupDriver.GetBackupUID(ctx, backupName, BackupOrgID)
				log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupName)
				_, err = DeleteBackup(backupName, backupUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting backup - [%s]", backupName))
			}
		})
		Step("Restoring the backed up namespaces", func() {
			log.InfoD("Restoring the backed up namespaces")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for i, backupName := range incrementalBackupNames2 {
				restoreName = fmt.Sprintf("%s-%s", backupName, RandomString(4))
				for strings.Contains(strings.Join(restoreNames, ","), restoreName) {
					restoreName = fmt.Sprintf("%s-%s", backupName, RandomString(4))
				}
				log.InfoD("Restoring %s backup", backupName)
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{bkpNamespaces[i]})
				err = CreateRestoreWithValidation(ctx, restoreName, backupName, namespaceMapping, make(map[string]string), DestinationClusterName, BackupOrgID, appContextsToBackup)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s]", restoreName))
				restoreNames = append(restoreNames, restoreName)
			}
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		// Cleaning up applications created
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		err := DestroyAppsWithData(scheduledAppContexts, opts, controlChannel, errorGroup)
		log.FailOnError(err, "Data validations failed")

		// Remove all the restores created
		log.Info("Deleting restored namespaces")
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		for _, restoreName := range restoreNames {
			err := DeleteRestore(restoreName, BackupOrgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting Restore [%s]", restoreName))
		}

		// Cleaning up px-backup cluster
		ctx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		CleanupCloudSettingsAndClusters(backupLocationMap, credName, cloudCredUID, ctx)
	})
})

// DeleteIncrementalBackupsAndRecreateNew Delete Incremental Backups and Recreate
// new ones
var _ = Describe("{DeleteIncrementalBackupsAndRecreateNew}", Label(TestCaseLabelsMap[DeleteIncrementalBackupsAndRecreateNew]...), func() {
	backupNames := make([]string, 0)
	incrementalBackupNames := make([]string, 0)
	incrementalBackupNamesRecreated := make([]string, 0)
	var scheduledAppContexts []*scheduler.Context
	labelSelectors := make(map[string]string)
	var (
		controlChannel chan string
		errorGroup     *errgroup.Group
	)
	var backupLocationUID string
	var cloudCredUID string
	var cloudCredUidList []string
	var clusterUid string
	var clusterStatus api.ClusterInfo_StatusInfo_Status
	var customBackupLocationName string
	var credName string
	var fullBackupName string
	var incrementalBackupName string
	var bkpNamespaces = make([]string, 0)
	backupLocationMap := make(map[string]string)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("DeleteIncrementalBackupsAndRecreateNew",
			"Delete incremental Backups and re-create them", nil, 58039, Kshithijiyer, Q1FY24)
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

	It("Delete incremental Backups and re-create them", func() {
		providers := GetBackupProviders()
		Step("Validate applications", func() {
			log.InfoD("Validate applications")
			ctx, _ := backup.GetAdminCtxFromSecret()
			controlChannel, errorGroup = ValidateApplicationsStartData(scheduledAppContexts, ctx)
		})

		Step("Adding Credentials and Registering Backup Location", func() {
			log.InfoD("Creating cloud credentials and backup location")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, provider := range providers {
				cloudCredUID = uuid.New()
				cloudCredUidList = append(cloudCredUidList, cloudCredUID)
				backupLocationUID = uuid.New()
				credName = fmt.Sprintf("autogenerated-cred-%v", time.Now().Unix())
				err = CreateCloudCredential(provider, credName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating cloud credential %s", credName))
				log.InfoD("Created Cloud Credentials with name - %s", credName)
				customBackupLocationName = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
				err := CreateBackupLocation(provider, customBackupLocationName, backupLocationUID, credName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", customBackupLocationName))
				backupLocationMap[backupLocationUID] = customBackupLocationName
				log.InfoD("Created Backup Location with name - %s", customBackupLocationName)
			}
		})

		Step("Register source and destination cluster for backup", func() {
			log.InfoD("Registering Source and Destination clusters and verifying the status")
			// Registering for admin user
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})

		Step("Taking backup of applications", func() {
			log.InfoD("Taking backup of applications")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			// Full backup
			for _, namespace := range bkpNamespaces {
				fullBackupName = fmt.Sprintf("%s-%s-%v", "full-backup", namespace, time.Now().Unix())
				backupNames = append(backupNames, fullBackupName)
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
				err = CreateBackupWithValidation(ctx, fullBackupName, SourceClusterName, customBackupLocationName, backupLocationUID, appContextsToBackup, labelSelectors, BackupOrgID, clusterUid, "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of full backup [%s]", fullBackupName))
			}

			// Incremental backup
			for _, namespace := range bkpNamespaces {
				incrementalBackupName = fmt.Sprintf("%s-%s-%v", "incremental-backup", namespace, time.Now().Unix())
				incrementalBackupNames = append(incrementalBackupNames, incrementalBackupName)
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
				err = CreateBackupWithValidation(ctx, incrementalBackupName, SourceClusterName, customBackupLocationName, backupLocationUID, appContextsToBackup, labelSelectors, BackupOrgID, clusterUid, "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of incremental backup [%s]", incrementalBackupName))
			}
			log.Infof("List of backups - %v", backupNames)
			log.Infof("List of Incremental backups - %v", incrementalBackupNames)

		})
		Step("Deleting incremental backup", func() {
			log.InfoD("Deleting incremental backups")
			backupDriver := Inst().Backup
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, backupName := range incrementalBackupNames {
				log.Infof("About to delete backup - %s", backupName)
				backupUID, err := backupDriver.GetBackupUID(ctx, backupName, BackupOrgID)
				log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupName)
				_, err = DeleteBackup(backupName, backupUID, BackupOrgID, ctx)
				log.FailOnError(err, "Failed to issue delete backup for - %s", backupName)
				err = DeleteBackupAndWait(backupName, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Deleted backup - [%s]", backupName))
			}
		})
		Step("Taking incremental backups of applications again", func() {
			log.InfoD("Taking incremental backups of applications again")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			// Incremental backup
			for _, namespace := range bkpNamespaces {
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
				incrementalBackupName, err = CreateBackupUntilIncrementalBackup(ctx, appContextsToBackup[0], customBackupLocationName, backupLocationUID, labelSelectors, BackupOrgID, clusterUid)
				dash.VerifyFatal(err, nil, "Creating incremental backup")
				incrementalBackupNamesRecreated = append(incrementalBackupNamesRecreated, incrementalBackupName)
			}
			log.Infof("List of New Incremental backups - %v", incrementalBackupNames)
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		// Cleaning up applications created
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		err := DestroyAppsWithData(scheduledAppContexts, opts, controlChannel, errorGroup)
		log.FailOnError(err, "Data validations failed")

		// Cleaning up px-backup cluster
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		CleanupCloudSettingsAndClusters(backupLocationMap, credName, cloudCredUID, ctx)
	})
})

// DeleteBucketVerifyCloudBackupMissing validates the backup state (CloudBackupMissing) when bucket is deleted.
var _ = Describe("{DeleteBucketVerifyCloudBackupMissing}", Label(TestCaseLabelsMap[DeleteBucketVerifyCloudBackupMissing]...), func() {
	var (
		scheduledAppContexts       []*scheduler.Context
		clusterUid                 string
		cloudAccountUID            string
		cloudAccountName           string
		bkpLocationName            string
		backupLocationUID          string
		backupLocationMap          map[string]string
		periodicSchedulePolicyName string
		periodicSchedulePolicyUid  string
		scheduleName               string
		appNamespaces              []string
		scheduleNames              []string
		backupNames                []string
		localBucketNameMap         map[string]string
		controlChannel             chan string
		errorGroup                 *errgroup.Group
	)

	providers := GetBackupProviders()
	backupLocationMap = make(map[string]string)
	localBucketNameMap = make(map[string]string)
	appContextsToBackupMap := make(map[string][]*scheduler.Context)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("DeleteBucketVerifyCloudBackupMissing", "Validates the backup state (CloudBackupMissing) when bucket is deleted.", nil, 58070, Ak, Q1FY24)
		log.Infof("Deploying applications required for the testcase")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = AppReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				appNamespaces = append(appNamespaces, namespace)
				scheduledAppContexts = append(scheduledAppContexts, ctx)
			}
		}
	})

	It("Delete bucket and Validates the backup state", func() {
		Step("Validate deployed applications", func() {
			log.InfoD("Validate applications")
			ctx, _ := backup.GetAdminCtxFromSecret()
			controlChannel, errorGroup = ValidateApplicationsStartData(scheduledAppContexts, ctx)
		})

		Step("Adding cloud path/bucket", func() {
			log.InfoD("Adding cloud path/bucket")
			for _, provider := range providers {
				bucketNameSuffix := getBucketNameSuffix()
				bucketNamePrefix := fmt.Sprintf("local-%s", provider)
				localBucketName := fmt.Sprintf("%s-%s-%v", bucketNamePrefix, bucketNameSuffix, time.Now().Unix())
				localBucketNameMap[provider] = localBucketName
			}
		})

		Step("Adding cloud account and backup location", func() {
			log.InfoD("Adding cloud account and backup location")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, provider := range providers {
				cloudAccountName = fmt.Sprintf("%s-%s-%v", "cloudcred", provider, time.Now().Unix())
				bkpLocationName = fmt.Sprintf("%s-%s-%v-bl", provider, "local-location", time.Now().Unix())
				cloudAccountUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = bkpLocationName
				err := CreateCloudCredential(provider, cloudAccountName, cloudAccountUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud account named [%s] for org [%s] with [%s] as provider", cloudAccountName, BackupOrgID, provider))
				err = CreateBackupLocation(provider, bkpLocationName, backupLocationUID, cloudAccountName, cloudAccountUID, localBucketNameMap[provider], BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", bkpLocationName))
			}
		})

		Step("Creating Schedule Policies", func() {
			log.InfoD("Creating Schedule Policies")
			periodicSchedulePolicyName = fmt.Sprintf("%s-%v", "periodic", time.Now().Unix())
			periodicSchedulePolicyInfo := Inst().Backup.CreateIntervalSchedulePolicy(5, 15, 5)
			periodicPolicyStatus := Inst().Backup.BackupSchedulePolicy(periodicSchedulePolicyName, uuid.New(), BackupOrgID, periodicSchedulePolicyInfo)
			dash.VerifyFatal(periodicPolicyStatus, nil, fmt.Sprintf("Verification of creating periodic schedule policy - %s", periodicSchedulePolicyName))
		})

		Step("Adding Clusters for backup", func() {
			log.InfoD("Adding Clusters for backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of creating source - %s and destination - %s clusters", SourceClusterName, DestinationClusterName))
			clusterStatus, err := Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})

		Step("Creating schedule backup", func() {
			log.InfoD("Creating schedule backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			periodicSchedulePolicyUid, _ = Inst().Backup.GetSchedulePolicyUid(BackupOrgID, ctx, periodicSchedulePolicyName)
			for _, namespace := range appNamespaces {
				scheduleName = fmt.Sprintf("%s-schedule-%v", BackupNamePrefix, time.Now().Unix())
				scheduleNames = append(scheduleNames, scheduleName)
				labelSelectors := make(map[string]string)
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
				appContextsToBackupMap[scheduleName] = appContextsToBackup

				firstScheduleBackupName, err := CreateScheduleBackupWithValidation(ctx, scheduleName, SourceClusterName, bkpLocationName, backupLocationUID, appContextsToBackup, labelSelectors, BackupOrgID, "", "", "", "", periodicSchedulePolicyName, periodicSchedulePolicyUid)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of schedule backup with schedule name [%s]", scheduleName))
				backupNames = append(backupNames, firstScheduleBackupName)
			}
		})

		Step("Creating a manual backup", func() {
			log.InfoD("Creating a manual backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, namespace := range appNamespaces {
				backupName := fmt.Sprintf("%s-%v", BackupNamePrefix, time.Now().Unix())
				labelSelectors := make(map[string]string)
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
				err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, bkpLocationName, backupLocationUID, appContextsToBackup, labelSelectors, BackupOrgID, clusterUid, "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				backupNames = append(backupNames, backupName)
			}
		})

		Step("Suspending the existing backup schedules", func() {
			log.InfoD("Suspending the existing backup schedules")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, scheduleName := range scheduleNames {
				err = SuspendBackupSchedule(scheduleName, periodicSchedulePolicyName, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of suspending backup schedule - %s", scheduleName))
			}
		})

		Step("Delete the bucket where the backup objects are present", func() {
			log.InfoD("Delete the bucket where the backup objects are present")
			for _, provider := range providers {
				DeleteBucket(provider, localBucketNameMap[provider])
				log.Infof("Sleeping for default 10 minutes for next backup sync service to be triggered")
				time.Sleep(10 * time.Minute)
			}
		})

		Step("Verify the backups are in CloudBackupMissing state after bucket deletion", func() {
			log.InfoD("Verify the backups are in CloudBackupMissing state after bucket deletion")
			var wg sync.WaitGroup
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, backupName := range backupNames {
				wg.Add(1)
				go func(backupName string) {
					defer GinkgoRecover()
					defer wg.Done()
					bkpUid, err := Inst().Backup.GetBackupUID(ctx, backupName, BackupOrgID)
					log.FailOnError(err, "Fetching backup uid")
					backupInspectRequest := &api.BackupInspectRequest{
						Name:  backupName,
						Uid:   bkpUid,
						OrgId: BackupOrgID,
					}
					requiredStatus := api.BackupInfo_StatusInfo_CloudBackupMissing
					backupCloudBackupMissingCheckFunc := func() (interface{}, bool, error) {
						resp, err := Inst().Backup.InspectBackup(ctx, backupInspectRequest)
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
			wg.Wait()
		})

		Step("Resume the existing backup schedules", func() {
			log.InfoD("Resume the existing backup schedules")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, scheduleName := range scheduleNames {
				err = ResumeBackupSchedule(scheduleName, periodicSchedulePolicyName, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of resuming backup schedule - %s", scheduleName))
			}
			log.Infof("Waiting 5 minute for another schedule backup to trigger")
			time.Sleep(5 * time.Minute)
		})

		Step("Get the latest schedule backup and verify the backup status", func() {
			log.InfoD("Get the latest schedule backup and verify the backup status")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, scheduleName := range scheduleNames {
				latestScheduleBkpName, err := GetLatestScheduleBackupName(ctx, scheduleName, BackupOrgID)
				log.FailOnError(err, "Error while getting latest schedule backup name")
				err = BackupSuccessCheckWithValidation(ctx, latestScheduleBkpName, appContextsToBackupMap[scheduleName], BackupOrgID, MaxWaitPeriodForBackupCompletionInMinutes*time.Minute, 30*time.Second)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of success and Validation of latest schedule backup [%s]", latestScheduleBkpName))
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
		err = DestroyAppsWithData(scheduledAppContexts, opts, controlChannel, errorGroup)
		log.FailOnError(err, "Data validations failed")
		for _, scheduleName := range scheduleNames {
			err = DeleteSchedule(scheduleName, SourceClusterName, BackupOrgID, ctx, true)
			dash.VerifySafely(err, nil, fmt.Sprintf("Verification of deleting backup schedule - %s", scheduleName))
		}
		log.Infof("Deleting backup schedule policy")
		err = Inst().Backup.DeleteBackupSchedulePolicy(BackupOrgID, []string{periodicSchedulePolicyName})
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudAccountName, cloudAccountUID, ctx)
		log.InfoD("Delete the local bucket created")
		for _, provider := range providers {
			DeleteBucket(provider, localBucketNameMap[provider])
			log.Infof("local bucket deleted - %s", localBucketNameMap[provider])
		}
	})
})

// DeleteBackupAndCheckIfBucketIsEmpty delete backups and verify if contents are deleted from backup location or not
var _ = Describe("{DeleteBackupAndCheckIfBucketIsEmpty}", Label(TestCaseLabelsMap[DeleteBackupAndCheckIfBucketIsEmpty]...), func() {
	numberOfBackups, _ := strconv.Atoi(GetEnv(MaxBackupsToBeCreated, "10"))
	var (
		scheduledAppContexts     []*scheduler.Context
		backupLocationUID        string
		cloudCredUID             string
		bkpNamespaces            []string
		clusterUid               string
		clusterStatus            api.ClusterInfo_StatusInfo_Status
		customBackupLocationName string
		credName                 string
		customBucketName         string
		controlChannel           chan string
		errorGroup               *errgroup.Group
	)
	timeBetweenConsecutiveBackups := 10 * time.Second
	backupNames := make([]string, 0)
	numberOfSimultaneousBackups := 4
	labelSelectors := make(map[string]string)
	bkpNamespaces = make([]string, 0)
	backupLocationMap := make(map[string]string)
	appContextsToBackupMap := make(map[string][]*scheduler.Context)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("DeleteBackupAndCheckIfBucketIsEmpty",
			"Delete backups and verify if contents are deleted from backup location or not", nil, 58071, Kshithijiyer, Q2FY24)
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
		providers := GetBackupProviders()
		log.Info("Check if backup location is empty or not")
		customBucketName = fmt.Sprintf("custom-bucket-%v", RandomString(5))
		for _, provider := range providers {
			switch provider {
			case drivers.ProviderAws:
				CreateBucket(provider, customBucketName)
				log.Infof("Bucket created with name - %s", GlobalAWSBucketName)
			case drivers.ProviderAzure:
				CreateBucket(provider, customBucketName)
				log.Infof("Bucket created with name - %s", GlobalAzureBucketName)
			case drivers.ProviderGke:
				CreateBucket(provider, customBucketName)
				log.Infof("Bucket created with name - %s", GlobalGCPBucketName)
			}
		}
	})
	It("Delete backups and verify if contents are deleted from backup location or not", func() {
		providers := GetBackupProviders()
		Step("Validate applications", func() {
			log.InfoD("Validate applications")
			ctx, _ := backup.GetAdminCtxFromSecret()
			controlChannel, errorGroup = ValidateApplicationsStartData(scheduledAppContexts, ctx)
		})

		Step("Adding Credentials and Backup Location", func() {
			log.InfoD("Using bucket - [%s]. Creating cloud credentials and backup location.", customBucketName)
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, provider := range providers {
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				credName = fmt.Sprintf("autogenerated-cred-%v", time.Now().Unix())
				err := CreateCloudCredential(provider, credName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", credName, BackupOrgID, provider))
				log.InfoD("Created Cloud Credentials with name - %s", credName)
				customBackupLocationName = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
				err = CreateBackupLocation(provider, customBackupLocationName, backupLocationUID, credName, cloudCredUID, customBucketName, BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", customBackupLocationName))
				backupLocationMap[backupLocationUID] = customBackupLocationName
			}
		})

		Step("Register source and destination cluster for backup", func() {
			log.InfoD("Registering Source and Destination clusters and verifying the status")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})

		Step("Taking backup of applications", func() {
			log.InfoD("Taking backup of applications")
			var sem = make(chan struct{}, numberOfSimultaneousBackups)
			var wg sync.WaitGroup
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			log.InfoD("Taking %d backups", numberOfBackups)
			var mutex sync.Mutex
			for backupLocationUID, backupLocationName := range backupLocationMap {
				for _, namespace := range bkpNamespaces {
					for i := 0; i < numberOfBackups; i++ {
						time.Sleep(timeBetweenConsecutiveBackups)
						backupName := fmt.Sprintf("%s-%v", BackupNamePrefix, time.Now().Unix())
						backupNames = append(backupNames, backupName)
						sem <- struct{}{}
						wg.Add(1)
						go func(backupName, backupLocationName, backupLocationUID, namespace string) {
							defer GinkgoRecover()
							defer wg.Done()
							defer func() { <-sem }()
							appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
							mutex.Lock()
							appContextsToBackupMap[backupName] = appContextsToBackup
							mutex.Unlock()
							err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, backupLocationName, backupLocationUID, appContextsToBackup, labelSelectors, BackupOrgID, clusterUid, "", "", "", "")
							dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
						}(backupName, backupLocationName, backupLocationUID, namespace)
					}
				}
			}
			wg.Wait()
			log.Infof("List of backups - %v", backupNames)
		})

		Step("Delete all the backups taken in the previous step ", func() {
			log.InfoD("Deleting the backups")
			ctx, err := backup.GetAdminCtxFromSecret()
			dash.VerifyFatal(err, nil, "Unable fetch admin context")
			var wg sync.WaitGroup
			for _, backupName := range backupNames {
				wg.Add(1)
				go func(backupName string) {
					defer GinkgoRecover()
					defer wg.Done()
					backupUid, err := Inst().Backup.GetBackupUID(ctx, backupName, BackupOrgID)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching the backup %s uid", backupName))
					_, err = DeleteBackup(backupName, backupUid, BackupOrgID, ctx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying the backup %s deletion", backupName))
				}(backupName)
			}
			wg.Wait()
			for _, backup := range backupNames {
				err := DeleteBackupAndWait(backup, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Waiting for the backup %s to be deleted", backup))
			}
		})
		Step("Check if contents are erased from the backup location or not", func() {
			log.Info("Check if backup location is empty or not")
			for _, provider := range providers {
				result, err := IsBackupLocationEmpty(provider, customBucketName)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Validating contents of bucket [%s] for provider [%s]", customBucketName, provider))
				dash.VerifyFatal(result, true, fmt.Sprintf("Validate if bucket [%s] is empty", customBucketName))
			}
		})
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		err := DestroyAppsWithData(scheduledAppContexts, opts, controlChannel, errorGroup)
		log.FailOnError(err, "Data validations failed")

		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		CleanupCloudSettingsAndClusters(backupLocationMap, credName, cloudCredUID, ctx)
	})
})
