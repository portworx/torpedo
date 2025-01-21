package tests

import (
	"fmt"
	"math"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/sched-ops/k8s/storage"
	"github.com/pure-px/torpedo/drivers/backup"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
	"golang.org/x/sync/errgroup"
)

// This testcase verifies if basic backup and Restore with parallel Backup Schedule
var _ = Describe("{BasicBackupAndRestoreWithParallelBackupSchedule}", Label(TestCaseLabelsMap[BasicBackupAndRestoreWithParallelBackupSchedule]...), func() {

	var (
		backupNames          []string
		restoreNames         []string
		scheduledAppContexts []*scheduler.Context
		preRuleNameList      []string
		postRuleNameList     []string
		bkpNamespaces        []string
		scheduleNameList     []string
		preRuleName          string
		postRuleName         string
		preRuleUid           string
		postRuleUid          string
		sourceClusterUid     string
		destClusterUid       string
		cloudCredName        string
		cloudCredUID         string
		backupLocationUID    string
		backupLocationName   string
		appList              []string
		originalAppList      []string
		backupLocationMap    map[string]string
		labelSelectors       map[string]string
		storageClassMapping  map[string]string
		providers            []string
		intervalName         string
		intervalUid          string
		controlChannel       chan string
		errorGroup           *errgroup.Group
		appScaleFactor       float64
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("BasicBackupAndRestoreWithParallelBackupSchedule", "Validate if Parallel Backups are scheduled or not if Local Snapshot is completed before subsequent triggers", nil, 304644, Kshithijiyer, Q4FY25)

		originalAppList = Inst().AppList
		Inst().AppList = []string{"mysql-backup-data"}
		appList = Inst().AppList
		backupLocationMap = make(map[string]string)
		storageClassMapping = make(map[string]string)
		labelSelectors = make(map[string]string)
		providers = GetBackupProviders()
		intervalName = fmt.Sprintf("%s-%v", "interval", time.Now().Unix())
		numberOfDeployments := float64(5)

		log.InfoD("scheduling applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		appScaleFactor = math.Max(float64(Inst().GlobalScaleFactor), numberOfDeployments)
		for i := 0; i < int(appScaleFactor); i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, appCtx := range appContexts {
				appCtx.ReadinessTimeout = 45 * time.Minute
				namespace := GetAppNamespace(appCtx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
				scheduledAppContexts = append(scheduledAppContexts, appCtx)
			}
		}

	})

	It("Basic Parallel Backup Schedule validation", func() {
		defer func() {
			log.InfoD("switching to default context")
			err := SetClusterContext("")
			log.FailOnError(err, "failed to SetClusterContext to default cluster")
		}()

		Step("Setting cluster-wide average network bandwidth to 90MiB/s", func() {
			err := ThrottleNetworkSpeed(90)
			dash.VerifyFatal(err, nil, "Setting cluster-wide average network bandwidth to 90MiB/s")

			err = SetDestinationKubeConfig()
			log.FailOnError(err, "Switching context to destination cluster failed")

			err = ThrottleNetworkSpeed(90)
			dash.VerifyFatal(err, nil, "Setting cluster-wide average network bandwidth to 90MiB/s")

			defer func() {
				err = SetSourceKubeConfig()
				log.FailOnError(err, "Unable to switch context to source cluster [%s]", SourceClusterName)
			}()
		})

		Step("Validating applications", func() {
			log.InfoD("Validating applications")
			ctx, _ := backup.GetAdminCtxFromSecret()
			controlChannel, errorGroup = ValidateApplicationsStartData(scheduledAppContexts, ctx)
		})

		Step("Creating rules for backup", func() {
			log.InfoD("Creating pre rule for deployed apps")
			for i := 0; i < len(appList); i++ {
				preRuleStatus, ruleName, err := Inst().Backup.CreateRuleForBackup(appList[i], BackupOrgID, "pre")
				log.FailOnError(err, "Creating pre rule for deployed app [%s] failed", appList[i])
				dash.VerifyFatal(preRuleStatus, true, "Verifying pre rule for backup")
				if ruleName != "" {
					preRuleNameList = append(preRuleNameList, ruleName)
				}
			}
			log.InfoD("Creating post rule for deployed apps")
			for i := 0; i < len(appList); i++ {
				postRuleStatus, ruleName, err := Inst().Backup.CreateRuleForBackup(appList[i], BackupOrgID, "post")
				log.FailOnError(err, "Creating post rule for deployed app [%s] failed", appList[i])
				dash.VerifyFatal(postRuleStatus, true, "Verifying Post rule for backup")
				if ruleName != "" {
					postRuleNameList = append(postRuleNameList, ruleName)
				}
			}
		})

		Step("Creating backup location and cloud setting", func() {
			log.InfoD("Creating backup location and cloud setting")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
				backupLocationName = fmt.Sprintf("%s-%s-bl-%v", provider, getGlobalBucketName(provider), time.Now().Unix())
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocationName
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocation(provider, backupLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, "Creating backup location")
			}
		})

		Step("Creating backup schedule policies", func() {
			log.InfoD("Creating backup schedule policies")
			intervalSchedulePolicyInfo := Inst().Backup.CreateIntervalSchedulePolicy(10, 15, 6)
			intervalUid = uuid.New()
			intervalPolicyStatus := Inst().Backup.BackupSchedulePolicy(intervalName, intervalUid, BackupOrgID, intervalSchedulePolicyInfo)
			dash.VerifyFatal(intervalPolicyStatus, nil, "Creating interval schedule policy")
		})

		Step("Registering cluster for backup", func() {
			log.InfoD("Registering cluster for backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")

			clusterStatus, err := Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))

			sourceClusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))

			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, DestinationClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", DestinationClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", DestinationClusterName))

			destClusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))
		})

		Step("Taking backup of application from source cluster", func() {
			log.InfoD("Taking backup of applications")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			if len(preRuleNameList) > 0 {
				preRuleUid, err = Inst().Backup.GetRuleUid(BackupOrgID, ctx, preRuleNameList[0])
				log.FailOnError(err, "Failed to get UID for rule %s", preRuleNameList[0])
				preRuleName = preRuleNameList[0]
			}
			if len(postRuleNameList) > 0 {
				postRuleUid, err = Inst().Backup.GetRuleUid(BackupOrgID, ctx, postRuleNameList[0])
				log.FailOnError(err, "Failed to get UID for rule %s", postRuleNameList[0])
				postRuleName = postRuleNameList[0]
			}
			for backupLocationUID, backupLocationName := range backupLocationMap {
				scheduleBackupName := fmt.Sprintf("%s-%v", BackupNamePrefix, time.Now().Unix())
				resp, err := CreateScheduleBackupWithoutCheck(scheduleBackupName, SourceClusterName, sourceClusterUid, backupLocationName, backupLocationUID, bkpNamespaces, labelSelectors, BackupOrgID, preRuleName, preRuleUid, postRuleName, postRuleUid, intervalName, intervalUid, ctx, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of schedule backup with schedule name [%s]", scheduleBackupName))
				dash.VerifyFatal(resp.BackupSchedule.ParallelBackup, true, "Verifying if parallelBackup var on Backup Schedule is set to True or not")
				scheduleNameList = append(scheduleNameList, scheduleBackupName)
			}
		})

		Step("Checking if Pre rule was executed or not", func() {
			log.InfoD("Checking if Pre rule was executed or not")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, scheduleBackupName := range scheduleNameList {
				backupName, err := GetOrdinalScheduleBackupName(ctx, scheduleBackupName, 1, BackupOrgID)
				dash.VerifyFatal(err, nil, "Fetching the first scheduled backup name")
				err = CheckPreAndPostRuleExecution(PreRule, appScaleFactor, backupName)
				dash.VerifyFatal(err, nil, "Checking if Pre rule was executed or not")
			}
		})

		Step("Checking if local Snapshots are created or not on all volumes or not", func() {
			log.InfoD("Checking if local Snapshots are created or not on all volumes or not")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, scheduleBackupName := range scheduleNameList {
				backupName, err := GetOrdinalScheduleBackupName(ctx, scheduleBackupName, 1, BackupOrgID)
				dash.VerifyFatal(err, nil, "Fetching the first scheduled backup name")
				err = ValidateLocalSnapshotCompleted(backupName, BackupOrgID, 1, ctx)
				dash.VerifyFatal(err, nil, "Checking if local Snapshots are created or not")
			}
		})

		Step("Checking if Post rule was executed or not", func() {
			log.InfoD("Checking if Post rule was executed or not")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, scheduleBackupName := range scheduleNameList {
				backupName, err := GetOrdinalScheduleBackupName(ctx, scheduleBackupName, 1, BackupOrgID)
				dash.VerifyFatal(err, nil, "Fetching the first scheduled backup name")
				err = CheckPreAndPostRuleExecution(PostRule, appScaleFactor, backupName)
				dash.VerifyFatal(err, nil, "Checking if Post rule was executed or not")
			}
		})

		Step("Check if the second set of local snapshots were created or not", func() {
			log.InfoD("Check if the second set of local snapshots were created or not")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, scheduleBackupName := range scheduleNameList {
				err := WaitTillScheduleBackupInDesiredState(scheduleBackupName, BackupOrgID, ctx, 2, api.BackupInfo_StatusInfo_InProgress)
				dash.VerifyFatal(err, nil, "Checking if second backup in desired state")
				err = ValidateNumberOfParallelScheduledBackups(scheduleBackupName, BackupOrgID, 0, ctx, 2)
				dash.VerifyFatal(err, nil, "Checking if second set of backup is triggered or not")
				backupName, err := GetOrdinalScheduleBackupName(ctx, scheduleBackupName, 2, BackupOrgID)
				dash.VerifyFatal(err, nil, "Fetching the second scheduled backup name")
				err = ValidateLocalSnapshotCompleted(backupName, BackupOrgID, 2, ctx)
				dash.VerifyFatal(err, nil, "Checking if local Snapshots are created or not")
			}
		})

		Step("Check if the first backup is successful or not", func() {
			log.InfoD("Check if the first backup is successful or not")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, scheduleBackupName := range scheduleNameList {
				backupName, err := GetOrdinalScheduleBackupName(ctx, scheduleBackupName, 1, BackupOrgID)
				dash.VerifyFatal(err, nil, "Fetching the second scheduled backup name")
				err = BackupSuccessCheck(backupName, BackupOrgID, 120*time.Minute, 5*time.Minute, ctx)
				dash.VerifyFatal(err, nil, "Checking if backup is successful or not")
			}
		})

		Step("Check if the next backup is created or not", func() {
			log.InfoD("Check if the next backup is created or not")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, scheduleName := range scheduleNameList {
				scheduleBackupNames, err := Inst().Backup.GetAllScheduleBackupNames(ctx, scheduleName, BackupOrgID)
				dash.VerifyFatal(err, nil, "Get all schedule backup names")
				ordinalCounter := len(scheduleBackupNames) + 1
				err = WaitTillScheduleBackupInDesiredState(scheduleName, BackupOrgID, ctx, ordinalCounter, api.BackupInfo_StatusInfo_InProgress)
				dash.VerifyFatal(err, nil, "Checking if latest backup in desired state")
				backupName, err := GetOrdinalScheduleBackupName(ctx, scheduleName, ordinalCounter, BackupOrgID)
				dash.VerifyFatal(err, nil, "Fetching the latest scheduled backup name")
				snapshotCount, err := GetSnapshotCount(scheduleName, BackupOrgID, ctx, ordinalCounter)
				dash.VerifyFatal(err, nil, "Fetching the number of expected local snapshots")
				err = ValidateLocalSnapshotCompleted(backupName, BackupOrgID, snapshotCount, ctx)
				dash.VerifyFatal(err, nil, "Checking if local Snapshots are created or not")
			}
		})

		Step("Pause the backup schedule to make sure no new backups are created", func() {
			log.InfoD("Pause the backup schedule to make sure no new backups are created")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, scheduleBackupName := range scheduleNameList {
				err := SuspendBackupSchedule(scheduleBackupName, intervalName, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Pause Backup Schedule :[%s]", scheduleBackupName))
			}
		})

		Step("Check all backups are successful or not", func() {
			log.InfoD("Check all backups are successful or not")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, scheduleBackupName := range scheduleNameList {
				for i := 1; i < 4; i++ {
					backupName, err := GetOrdinalScheduleBackupName(ctx, scheduleBackupName, i, BackupOrgID)
					dash.VerifyFatal(err, nil, "Fetching the ordinal scheduled backup name")
					err = BackupSuccessCheck(backupName, BackupOrgID, 120*time.Minute, 5*time.Minute, ctx)
					dash.VerifyFatal(err, nil, "Checking if backup is successful or not")
					if i > 1 {
						isIncremental, err := AreAllVolumeBackupsFull(BackupOrgID, backupName, ctx)
						dash.VerifyFatal(err, nil, "Checking if backup is incremental or not")
						dash.VerifyFatal(isIncremental, false, "Checking if backup is incremental or not")
					}
				}
			}
		})

		Step("Create new storage class for restore", func() {
			log.InfoD("Getting storage class of the destination cluster")
			err := SetDestinationKubeConfig()
			dash.VerifyFatal(err, nil, "Setting destination kube config")
			pvcs, err := core.Instance().GetPersistentVolumeClaims(bkpNamespaces[0], make(map[string]string))
			log.FailOnError(err, "Getting PVC on source cluster")
			for _, singlePvc := range pvcs.Items {
				storageClass, err := core.Instance().GetStorageClassForPVC(&singlePvc)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Getting SC %v from PVC in source cluster",
					storageClass.Name))
				oldScName := storageClass.Name
				storageClass.Name += fmt.Sprintf("-new-sc-%s", RandomString(4))
				_, err = storage.Instance().CreateStorageClass(storageClass)
				log.FailOnError(err, "Creating sc on dest cluster")
				storageClassMapping[oldScName] = storageClass.Name
			}
			defer func() {
				err = SetSourceKubeConfig()
				dash.VerifyFatal(err, nil, "Setting source kube config")
			}()
		})

		Step("Restoring the scheduled backup with NS mapping on the same cluster", func() {
			log.InfoD("Restoring the scheduled backup with NS mapping on the same cluster")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, scheduleName := range scheduleNameList {
				latestScheduleBackupName, err := GetOrdinalScheduleBackupName(ctx, scheduleName, 1, BackupOrgID)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching the name of the first schedule backup [%s]", latestScheduleBackupName))
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
				restoreName := fmt.Sprintf("%s-restore-%v", latestScheduleBackupName, time.Now().Unix())
				namespaceMapping := make(map[string]string)
				for _, namespace := range bkpNamespaces {
					customNamespace := fmt.Sprintf("%s-%s", namespace, RandomString(4))
					namespaceMapping[namespace] = customNamespace
				}
				err = CreateRestoreWithValidation(ctx, restoreName, latestScheduleBackupName, namespaceMapping, make(map[string]string), SourceClusterName, sourceClusterUid, BackupOrgID, appContextsToBackup)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore %s", restoreName))
				restoreNames = append(restoreNames, restoreName)
			}
		})

		Step("Restoring backups with namespace and storage class mapping", func() {
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			log.InfoD("Restoring schedule backups with namespace and storage class mapping")
			for _, scheduleName := range scheduleNameList {
				latestScheduleBackupName, err := GetOrdinalScheduleBackupName(ctx, scheduleName, 2, BackupOrgID)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching the name of the first schedule backup [%s]", latestScheduleBackupName))
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
				restoreName := fmt.Sprintf("%s-restore-%v", latestScheduleBackupName, time.Now().Unix())
				namespaceMapping := make(map[string]string)
				for _, namespace := range bkpNamespaces {
					customNamespace := fmt.Sprintf("%s-%s", namespace, RandomString(5))
					namespaceMapping[namespace] = customNamespace
				}
				err = CreateRestoreWithValidation(ctx, restoreName, latestScheduleBackupName, namespaceMapping, storageClassMapping, DestinationClusterName, destClusterUid, BackupOrgID, appContextsToBackup)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore %s", restoreName))
				restoreNames = append(restoreNames, restoreName)
			}
		})

		Step("Restore from default ns with replace backup", func() {
			log.InfoD("Restore from schedule backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "failed to fetch px-admin ctx")
			for _, scheduleName := range scheduleNameList {
				latestScheduleBackupName, err := GetOrdinalScheduleBackupName(ctx, scheduleName, 3, BackupOrgID)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching the name of the first schedule backup [%s]", latestScheduleBackupName))
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
				restoreName := fmt.Sprintf("%s-restore-%v-%s", latestScheduleBackupName, time.Now().Unix(), SourceClusterName)
				err = CreateRestoreWithReplacePolicyWithValidation(restoreName, latestScheduleBackupName, make(map[string]string), SourceClusterName, BackupOrgID, ctx, make(map[string]string), ReplacePolicyDelete, appContextsToBackup)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore %s", restoreName))
				restoreNames = append(restoreNames, restoreName)
			}
		})

	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)

		defer func() {
			log.InfoD("switching to default context")
			err := SetClusterContext("")
			log.FailOnError(err, "failed to SetClusterContext to default cluster")
			Inst().AppList = originalAppList
		}()

		policyList := []string{intervalName}
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		for _, scheduleBackupName := range scheduleNameList {
			err = SuspendAndDeleteSchedule(scheduleBackupName, intervalName, SourceClusterName, sourceClusterUid, BackupOrgID, ctx, false)
			log.FailOnError(err, "Suspend and delete backup schedule")
		}

		err = Inst().Backup.DeleteBackupSchedulePolicy(BackupOrgID, policyList)
		log.FailOnError(err, "Deleting backup schedule policies")
		opts := make(map[string]bool)
		if len(preRuleNameList) > 0 {
			for _, ruleName := range preRuleNameList {
				err := Inst().Backup.DeleteRuleForBackup(BackupOrgID, ruleName)
				dash.VerifySafely(err, nil, fmt.Sprintf("Deleting backup pre rules [%s]", ruleName))
			}
		}
		if len(postRuleNameList) > 0 {
			for _, ruleName := range postRuleNameList {
				err := Inst().Backup.DeleteRuleForBackup(BackupOrgID, ruleName)
				dash.VerifySafely(err, nil, fmt.Sprintf("Deleting backup post rules [%s]", ruleName))
			}
		}
		opts[SkipClusterScopedObjects] = true

		log.Info("Destroying scheduled apps on source cluster")
		err = DestroyAppsWithData(scheduledAppContexts, opts, controlChannel, errorGroup)
		log.FailOnError(err, "Data validations failed")

		err = ThrottleNetworkSpeed(0)
		log.FailOnError(err, "Setting cluster-wide average network bandwidth to default value")

		log.InfoD("switching to destination context")
		err = SetDestinationKubeConfig()
		log.FailOnError(err, "failed to switch to context to destination cluster")

		err = ThrottleNetworkSpeed(0)
		log.FailOnError(err, "Setting cluster-wide average network bandwidth to default value")

		log.InfoD("Destroying restored apps on destination clusters")
		restoredAppContexts := make([]*scheduler.Context, 0)
		for _, scheduledAppContext := range scheduledAppContexts {
			restoredAppContext, err := CloneAppContextAndTransformWithMappings(scheduledAppContext, make(map[string]string), make(map[string]string), true)
			if err != nil {
				log.Errorf("TransformAppContextWithMappings: %v", err)
				continue
			}
			restoredAppContexts = append(restoredAppContexts, restoredAppContext)
		}
		DestroyApps(restoredAppContexts, opts)

		log.InfoD("switching to default context")
		err = SetClusterContext("")
		log.FailOnError(err, "failed to SetClusterContext to default cluster")

		backupDriver := Inst().Backup
		log.Info("Deleting backed up namespaces")
		for _, backupName := range backupNames {
			backupUID, err := backupDriver.GetBackupUID(ctx, backupName, BackupOrgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupName)
			backupDeleteResponse, err := DeleteBackup(backupName, backupUID, BackupOrgID, ctx)
			log.FailOnError(err, "Backup [%s] could not be deleted", backupName)
			dash.VerifyFatal(backupDeleteResponse.String(), "", fmt.Sprintf("Verifying [%s] backup deletion is successful", backupName))
		}
		log.Info("Deleting restored namespaces")
		for _, restoreName := range restoreNames {
			err = DeleteRestore(restoreName, BackupOrgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting Restore [%s]", restoreName))
		}

		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})
