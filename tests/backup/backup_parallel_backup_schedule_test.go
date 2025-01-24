package tests

import (
	"fmt"
	"github.com/pure-px/stork/pkg/k8sutils"
	"github.com/pure-px/torpedo/drivers/node"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/pure-px/sched-ops/k8s/core"
	"github.com/pure-px/sched-ops/k8s/storage"
	"github.com/pure-px/torpedo/drivers/backup"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
	"golang.org/x/sync/errgroup"
	"math/rand"
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
				dash.VerifyFatal(err, nil, "Checking if backup in desired state")
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
			pvcs, err := core.Instance().GetPersistentVolumeClaims(bkpNamespaces[0], make(map[string]string))
			log.FailOnError(err, "Getting PVC on source cluster")
			for _, singlePvc := range pvcs.Items {
				storageClass, err := core.Instance().GetStorageClassForPVC(&singlePvc)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Getting SC %v from PVC in source cluster",
					storageClass.Name))
				oldScName := storageClass.Name
				storageClass.Name += fmt.Sprintf("-new-sc-%s", RandomString(4))
				err = SetDestinationKubeConfig()
				dash.VerifyFatal(err, nil, "Setting destination kube config")
				v1obj := metaV1.ObjectMeta{
					Name: storageClass.Name,
				}
				scObj := &storagev1.StorageClass{
					ObjectMeta:           v1obj,
					Provisioner:          storageClass.Provisioner,
					Parameters:           storageClass.Parameters,
					ReclaimPolicy:        storageClass.ReclaimPolicy,
					VolumeBindingMode:    storageClass.VolumeBindingMode,
					MountOptions:         storageClass.MountOptions,
					AllowVolumeExpansion: storageClass.AllowVolumeExpansion,
				}
				_, err = storage.Instance().CreateStorageClass(scObj)
				dash.VerifyFatal(err, nil, "Creating sc on dest cluster")
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

// RestartMultipleComponentsWhenParallelBackupScheduleInProgress verifies parallel backup schedule when we restart various components like stork and PX.
var _ = Describe("{RestartMultipleComponentsWhenParallelBackupScheduleInProgress}", Label(TestCaseLabelsMap[RestartMultipleComponentsWhenParallelBackupScheduleInProgress]...), func() {

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
		cloudCredName        string
		cloudCredUID         string
		backupLocationUID    string
		backupLocationName   string
		appList              []string
		originalAppList      []string
		backupLocationMap    map[string]string
		labelSelectors       map[string]string
		GlobalOrdinalCounter int
		providers            []string
		intervalName         string
		intervalUid          string
		controlChannel       chan string
		errorGroup           *errgroup.Group
		ordinalCounterList   []int
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("RestartMultipleComponentsWhenParallelBackupScheduleInProgress", "Restarts Various components when Parallel Backup Schedule in-Progress", nil, 304648, Kshithijiyer, Q4FY25)

		originalAppList = Inst().AppList
		Inst().AppList = []string{"mysql-backup-data"}
		appList = Inst().AppList
		backupLocationMap = make(map[string]string)
		labelSelectors = make(map[string]string)
		providers = GetBackupProviders()
		intervalName = fmt.Sprintf("%s-%v", "interval", time.Now().Unix())
		numberOfDeployments := float64(3)
		GlobalOrdinalCounter = 0

		log.InfoD("scheduling applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		appScaleFactor := math.Max(float64(Inst().GlobalScaleFactor), numberOfDeployments)
		for i := 0; i < int(appScaleFactor); i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, appCtx := range appContexts {
				appCtx.ReadinessTimeout = 120 * time.Minute
				namespace := GetAppNamespace(appCtx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
				scheduledAppContexts = append(scheduledAppContexts, appCtx)
			}
		}

	})

	It("Restart Various components when Parallel Backup in-Progress", func() {
		defer func() {
			log.InfoD("switching to default context")
			err := SetClusterContext("")
			log.FailOnError(err, "failed to SetClusterContext to default cluster")
		}()

		Step("Validating applications", func() {
			log.InfoD("Validating applications")
			ctx, _ := backup.GetAdminCtxFromSecret()
			controlChannel, errorGroup = ValidateApplicationsStartData(scheduledAppContexts, ctx)
		})

		Step("Setting cluster-wide average network bandwidth to 60MiB/s", func() {
			err := ThrottleNetworkSpeed(60)
			dash.VerifyFatal(err, nil, "Setting cluster-wide average network bandwidth to 60MiB/s")

			err = SetDestinationKubeConfig()
			log.FailOnError(err, "Switching context to destination cluster failed")

			err = ThrottleNetworkSpeed(60)
			dash.VerifyFatal(err, nil, "Setting cluster-wide average network bandwidth to 60MiB/s")

			defer func() {
				err = SetSourceKubeConfig()
				log.FailOnError(err, "Unable to switch context to source cluster [%s]", SourceClusterName)
			}()
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
			intervalSchedulePolicyInfo := Inst().Backup.CreateIntervalSchedulePolicy(30, 15, 7)
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

			_, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, DestinationClusterName)
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
			} else {
				preRuleUid = ""
				preRuleName = ""
			}
			if len(postRuleNameList) > 0 {
				postRuleUid, err = Inst().Backup.GetRuleUid(BackupOrgID, ctx, postRuleNameList[0])
				log.FailOnError(err, "Failed to get UID for rule %s", postRuleNameList[0])
				postRuleName = postRuleNameList[0]
			} else {
				postRuleUid = ""
				postRuleName = ""
			}
			for backupLocationUID, backupLocationName := range backupLocationMap {
				scheduleBackupName := fmt.Sprintf("%s-%v", BackupNamePrefix, time.Now().Unix())
				resp, err := CreateScheduleBackupWithoutCheck(scheduleBackupName, SourceClusterName, sourceClusterUid, backupLocationName, backupLocationUID, bkpNamespaces, labelSelectors, BackupOrgID, preRuleName, preRuleUid, postRuleName, postRuleUid, intervalName, intervalUid, ctx, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of schedule backup with schedule name [%s]", scheduleBackupName))
				dash.VerifyFatal(resp.BackupSchedule.ParallelBackup, true, "Verifying if parallelBackup var on Backup Schedule is set to True or not")
				err = WaitTillScheduleBackupInDesiredState(scheduleBackupName, BackupOrgID, ctx, 1, api.BackupInfo_StatusInfo_InProgress)
				dash.VerifyFatal(err, nil, "Checking if first backup in desired state")
				err = ValidateNumberOfParallelScheduledBackups(scheduleBackupName, BackupOrgID, 15, ctx, 1)
				dash.VerifyFatal(err, nil, "Checking if first set of backup is triggered or not")
				scheduleNameList = append(scheduleNameList, scheduleBackupName)
				ordinalCounterList = append(ordinalCounterList, 1)
			}
		})

		Step("Kill stork when backup local snaps in progress", func() {
			log.InfoD("Kill stork when backup local snaps in progress")
			storkNamespace, err := k8sutils.GetStorkPodNamespace()
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching stork namespace %s", storkNamespace))
			err = DeletePodWithWithoutLabelInNamespace(storkNamespace, StorkLabel, false)
			dash.VerifyFatal(err, nil, "Killing stork while backups is in progress")
		})

		Step("Check if the first backup is successful or not", func() {
			log.InfoD("Check if the first backup is successful or not")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, scheduleBackupName := range scheduleNameList {
				backupName, err := GetOrdinalScheduleBackupName(ctx, scheduleBackupName, 1, BackupOrgID)
				dash.VerifyFatal(err, nil, "Fetching the first scheduled backup name")
				err = BackupSuccessCheckPartialSuccessCheck(backupName, BackupOrgID, 120*time.Minute, 5*time.Minute, ctx)
				dash.VerifyFatal(err, nil, "Checking if backup is successful or not")
			}
		})

		Step("Wait for the second backup local snapshots to completed", func() {
			log.InfoD("Wait for the second set local snapshots to completed")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, scheduleBackupName := range scheduleNameList {
				scheduleBackupNames, err := Inst().Backup.GetAllScheduleBackupNames(ctx, scheduleBackupName, BackupOrgID)
				dash.VerifyFatal(err, nil, "Get all schedule backup names")
				ordinalCounter := len(scheduleBackupNames) + 2
				err = WaitTillScheduleBackupInDesiredState(scheduleBackupName, BackupOrgID, ctx, ordinalCounter, api.BackupInfo_StatusInfo_InProgress)
				dash.VerifyFatal(err, nil, "Checking if second backup in desired state")
				backupName, err := GetOrdinalScheduleBackupName(ctx, scheduleBackupName, ordinalCounter, BackupOrgID)
				dash.VerifyFatal(err, nil, "Fetching the second scheduled backup name")
				snapshotCount, err := GetSnapshotCount(scheduleBackupName, BackupOrgID, ctx, ordinalCounter)
				dash.VerifyFatal(err, nil, "Fetching the number of expected local snapshots")
				err = ValidateLocalSnapshotCompleted(backupName, BackupOrgID, snapshotCount, ctx)
				dash.VerifyFatal(err, nil, "Checking if local Snapshots are created or not")
				GlobalOrdinalCounter = ordinalCounter
				ordinalCounterList = append(ordinalCounterList, GlobalOrdinalCounter)

			}
		})

		Step("Kill stork when backup upload in progress", func() {
			log.InfoD("Kill stork when backup upload in progress")
			storkNamespace, err := k8sutils.GetStorkPodNamespace()
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching stork namespace %s", storkNamespace))
			err = DeletePodWithWithoutLabelInNamespace(storkNamespace, StorkLabel, false)
			dash.VerifyFatal(err, nil, "Killing stork while backups is in progress")
		})

		Step("Check if the second backup is successful or not", func() {
			log.InfoD("Check if the second backup is successful or not")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, scheduleBackupName := range scheduleNameList {
				backupName, err := GetOrdinalScheduleBackupName(ctx, scheduleBackupName, GlobalOrdinalCounter, BackupOrgID)
				dash.VerifyFatal(err, nil, "Fetching the second scheduled backup name")
				err = BackupSuccessCheckPartialSuccessCheck(backupName, BackupOrgID, 120*time.Minute, 5*time.Minute, ctx)
				dash.VerifyFatal(err, nil, "Checking if backup is successful or not")
			}
		})

		Step("Wait for the third backup local snapshots to completed", func() {
			log.InfoD("Wait for the third set local snapshots to completed")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, scheduleName := range scheduleNameList {
				scheduleBackupNames, err := Inst().Backup.GetAllScheduleBackupNames(ctx, scheduleName, BackupOrgID)
				dash.VerifyFatal(err, nil, "Get all schedule backup names")
				ordinalCounter := len(scheduleBackupNames) + 1
				err = WaitTillScheduleBackupInDesiredState(scheduleName, BackupOrgID, ctx, ordinalCounter, api.BackupInfo_StatusInfo_InProgress)
				dash.VerifyFatal(err, nil, "Checking if third backup in desired state")
				backupName, err := GetOrdinalScheduleBackupName(ctx, scheduleName, ordinalCounter, BackupOrgID)
				dash.VerifyFatal(err, nil, "Fetching the third scheduled backup name")
				snapshotCount, err := GetSnapshotCount(scheduleName, BackupOrgID, ctx, ordinalCounter)
				dash.VerifyFatal(err, nil, "Fetching the number of expected local snapshots")
				err = ValidateLocalSnapshotCompleted(backupName, BackupOrgID, snapshotCount, ctx)
				dash.VerifyFatal(err, nil, "Checking if local Snapshots are created or not")
				GlobalOrdinalCounter = ordinalCounter
			}
		})

		Step(fmt.Sprintf("Restart volume driver nodes starts"), func() {
			log.InfoD("Restart PX on nodes")
			storageNodes := node.GetWorkerNodes()
			for index := range storageNodes {
				// Just restart storage driver on one of the node where volume backup is in progress
				err := Inst().V.RestartDriver(storageNodes[index], nil)
				log.FailOnError(err, fmt.Sprintf("Failed to Restart driver %s", storageNodes[index]))
				err = Inst().V.WaitDriverUpOnNode(storageNodes[index], time.Minute*10)
				dash.VerifyFatal(err, nil, "Validate volume is up")
			}
		})

		Step("Check if the third backup is successful or not", func() {
			log.InfoD("Check if the third backup is successful or not")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, scheduleBackupName := range scheduleNameList {
				backupName, err := GetOrdinalScheduleBackupName(ctx, scheduleBackupName, GlobalOrdinalCounter, BackupOrgID)
				dash.VerifyFatal(err, nil, "Fetching the third scheduled backup name")
				err = BackupSuccessCheckPartialSuccessCheck(backupName, BackupOrgID, 120*time.Minute, 5*time.Minute, ctx)
				dash.VerifyFatal(err, nil, "Checking if backup is successful or not")
			}
		})

		Step("Wait for the fourth backup local snapshots to completed", func() {
			log.InfoD("Wait for the fourth set local snapshots to completed")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, scheduleName := range scheduleNameList {
				scheduleBackupNames, err := Inst().Backup.GetAllScheduleBackupNames(ctx, scheduleName, BackupOrgID)
				dash.VerifyFatal(err, nil, "Get all schedule backup names")
				ordinalCounter := len(scheduleBackupNames) + 1
				err = WaitTillScheduleBackupInDesiredState(scheduleName, BackupOrgID, ctx, ordinalCounter, api.BackupInfo_StatusInfo_InProgress)
				dash.VerifyFatal(err, nil, "Checking if fourth backup in desired state")
				backupName, err := GetOrdinalScheduleBackupName(ctx, scheduleName, ordinalCounter, BackupOrgID)
				dash.VerifyFatal(err, nil, "Fetching the fourth scheduled backup name")
				snapshotCount, err := GetSnapshotCount(scheduleName, BackupOrgID, ctx, ordinalCounter)
				dash.VerifyFatal(err, nil, "Fetching the number of expected local snapshots")
				err = ValidateLocalSnapshotCompleted(backupName, BackupOrgID, snapshotCount, ctx)
				dash.VerifyFatal(err, nil, "Checking if local Snapshots are created or not")
				GlobalOrdinalCounter = ordinalCounter
			}
		})

		Step(fmt.Sprintf("Restart volume driver nodes starts"), func() {
			log.InfoD("Restart PX on nodes")
			storageNodes := node.GetWorkerNodes()
			for index := range storageNodes {
				// Just restart storage driver on one of the node where volume backup is in progress
				err := Inst().V.RestartDriver(storageNodes[index], nil)
				log.FailOnError(err, fmt.Sprintf("Failed to Restart driver %s", storageNodes[index]))
				err = Inst().V.WaitDriverUpOnNode(storageNodes[index], time.Minute*10)
				dash.VerifyFatal(err, nil, "Validate volume is up")
			}
		})

		Step("Check if the fourth backup is successful or not", func() {
			log.InfoD("Check if the fourth backup is successful or not")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, scheduleBackupName := range scheduleNameList {
				backupName, err := GetOrdinalScheduleBackupName(ctx, scheduleBackupName, GlobalOrdinalCounter, BackupOrgID)
				dash.VerifyFatal(err, nil, "Fetching the fourth scheduled backup name")
				err = BackupSuccessCheckPartialSuccessCheck(backupName, BackupOrgID, 120*time.Minute, 5*time.Minute, ctx)
				dash.VerifyFatal(err, nil, "Checking if backup is successful or not")
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
				scheduleBackupNames, err := Inst().Backup.GetAllScheduleBackupNames(ctx, scheduleBackupName, BackupOrgID)
				dash.VerifyFatal(err, nil, "Get all schedule backup names")
				for _, backupName := range scheduleBackupNames {
					err = BackupSuccessCheckPartialSuccessCheck(backupName, BackupOrgID, 120*time.Minute, 5*time.Minute, ctx)
					dash.VerifyFatal(err, nil, "Checking if backup is successful or not")
				}
			}
		})

		Step("Restore from default ns with replace backup", func() {
			log.InfoD("Restore from schedule backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "failed to fetch px-admin ctx")
			for _, scheduleName := range scheduleNameList {
				for _, ordinalCounter := range ordinalCounterList {
					backupName, err := GetOrdinalScheduleBackupName(ctx, scheduleName, ordinalCounter, BackupOrgID)
					dash.VerifyFatal(err, nil, "Fetching ordinal backup name for restoring")
					appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
					restoreName := fmt.Sprintf("%s-restore-%v-%s", backupName, time.Now().Unix(), SourceClusterName)
					err = CreateRestoreWithReplacePolicyWithValidation(restoreName, backupName, make(map[string]string), SourceClusterName, BackupOrgID, ctx, make(map[string]string), ReplacePolicyDelete, appContextsToBackup)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore %s", restoreName))
					restoreNames = append(restoreNames, restoreName)
				}
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

		err = ThrottleNetworkSpeed(0)
		log.FailOnError(err, "Setting cluster-wide average network bandwidth to default value")

		err = Inst().Backup.DeleteBackupSchedulePolicy(BackupOrgID, policyList)
		dash.VerifySafely(err, nil, "Deleting backup schedule policies")
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

// This testcase verify parallel schedule backup success when schedule backups are partial success
var _ = Describe("{ParallelScheduleWithPartialBackup}", Label(TestCaseLabelsMap[ParallelScheduleWithPartialBackup]...), func() {

	var (
		scheduledAppContexts    []*scheduler.Context
		sourceClusterUid        string
		cloudCredName           string
		cloudCredUID            string
		backupLocationUID       string
		backupLocationName      string
		backupLocationMap       map[string]string
		labelSelectors          map[string]string
		providers               []string
		failedPvcs              []*corev1.PersistentVolumeClaim
		firstScheduleBackupName string
		namespaceMapping        map[string]string
		schedulePolicyName      string
		schedulePolicyUID       string
		schedulePolicyInterval  int64
		destClusterUid          string
		scheduleName            string
		partialSuccessBackups   []string
		successBackups          []string
		namespaces              []string
		thirdScheduleBackupName string
		appList                 []string
		controlChannel          chan string
		errorGroup              *errgroup.Group
		appScaleFactor          float64
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("ParallelScheduleWithPartialBackup", "Verify parallel schedule backup success when schedule backups are partial success", nil, 304645, Ak, Q4FY25)

		backupLocationMap = make(map[string]string)
		namespaceMapping = make(map[string]string)
		labelSelectors = make(map[string]string)
		providers = GetBackupProviders()
		schedulePolicyInterval = 15

		log.InfoD("scheduling applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		appList = Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = []string{"mysql-backup-data"}
		numberOfDeployments := float64(1)
		appScaleFactor = math.Max(float64(Inst().GlobalScaleFactor), numberOfDeployments)
		for i := 0; i < int(appScaleFactor); i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, appCtx := range appContexts {
				appCtx.ReadinessTimeout = 60 * time.Minute
				scheduledAppContexts = append(scheduledAppContexts, appCtx)
				namespaces = append(namespaces, appCtx.ScheduleOptions.Namespace)
			}
		}

	})

	It("Verify parallel schedule backup success when schedule backups are partial success", func() {
		defer func() {
			log.InfoD("switching to default context")
			err := SetClusterContext("")
			log.FailOnError(err, "failed to SetClusterContext to default cluster")
		}()

		Step("Validating applications", func() {
			log.InfoD("Validating applications")
			ctx, _ := backup.GetAdminCtxFromSecret()
			controlChannel, errorGroup = ValidateApplicationsStartData(scheduledAppContexts, ctx)
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

		Step("Create schedule policy", func() {
			log.InfoD("Creating schedule policy")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			schedulePolicyName = fmt.Sprintf("%s-%v", "periodic-schedule-policy", RandomString(5))
			schedulePolicyUID = uuid.New()
			err = CreateBackupScheduleIntervalPolicy(5, schedulePolicyInterval, 5, schedulePolicyName, schedulePolicyUID, BackupOrgID, ctx, false, false)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of schedule policy %s", schedulePolicyName))
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

		Step("Making a list of PVCs to fail", func() {
			log.InfoD("Making a list of PVCs to fail")
			k8sCore := core.Instance()
			for _, appCtx := range scheduledAppContexts {
				scheduledNamespace := appCtx.ScheduleOptions.Namespace
				pvcList, err := k8sCore.GetPersistentVolumeClaims(scheduledNamespace, make(map[string]string))
				log.FailOnError(err, fmt.Sprintf("error getting PVC list for namespace %s", scheduledNamespace))
				pvcs := pvcList.Items
				log.Infof("PVCs to fail in namespace %s:", scheduledNamespace)
				for j := 0; j < len(pvcs)-1; j++ {
					log.Infof("PVC Name: [%s], PVC Volume name: [%s], PVC Namespace: [%s]", pvcs[j].Name, pvcs[j].Spec.VolumeName, pvcs[j].Namespace)
					failedPvcs = append(failedPvcs, &pvcs[j])
				}
				appCtx.SkipPodValidation = true
			}
		})

		Step("Taking first schedule backup of application from source cluster", func() {
			log.InfoD("Taking first schedule backup of application from source cluster")
			var wg sync.WaitGroup
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			wg.Add(2)
			// Go routine to stop backups for the PVCs
			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				var wg1 sync.WaitGroup
				wg1.Add(len(failedPvcs))
				for _, pvc := range failedPvcs {
					go func(pvc *corev1.PersistentVolumeClaim) {
						defer GinkgoRecover()
						defer wg1.Done()
						log.Infof("Stopping all cs backups for %s [%s] in namespace %s", pvc.Name, pvc.Spec.VolumeName, pvc.Namespace)
						err = StopCloudsnapBackup(pvc.Name, pvc.Namespace)
						log.FailOnError(err, fmt.Sprintf("error stopping all cs backups for %s [%s] in namespace %s", pvc.Name, pvc.Spec.VolumeName, pvc.Namespace))
					}(pvc)
				}
				wg1.Wait()
			}()

			// Go routine to create schedule backup
			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				scheduleName = fmt.Sprintf("%s-%s", "autogenerated-schedule", RandomString(4))
				log.InfoD("creating schedule backup [%s] in source cluster [%s] (%s), organization [%s], in backup location [%s]", scheduleName, SourceClusterName, sourceClusterUid, BackupOrgID, backupLocationName)
				resp, err := CreateScheduleBackupWithoutCheckWithVscMapping(scheduleName, SourceClusterName, sourceClusterUid, backupLocationName, backupLocationUID, namespaces, labelSelectors, BackupOrgID, "", "", "", "", schedulePolicyName, schedulePolicyUID, ctx, nil, false, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of partial schedule backup for schedule [%s] ", scheduleName))
				firstScheduleBackupName, err = GetFirstScheduleBackupName(ctx, scheduleName, BackupOrgID)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of first schedule backup [%s]", firstScheduleBackupName))
				dash.VerifyFatal(resp.BackupSchedule.ParallelBackup, true, "Verifying if parallelBackup var on Backup Schedule is set to True or not")
				err = ValidateLocalSnapshotCompletedForSuccessfulVolumes(firstScheduleBackupName, BackupOrgID, 1, failedPvcs, ctx)
				dash.VerifyFatal(err, nil, "Checking if local Snapshots are created or not")
				partialSuccessBackups = append(partialSuccessBackups, firstScheduleBackupName)
			}()
			wg.Wait()
		})

		Step("Waiting for second schedule backup of application from source cluster", func() {
			log.InfoD("Waiting for second schedule backup of application from source cluster")
			var wg sync.WaitGroup
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = WaitTillScheduleBackupInDesiredState(scheduleName, BackupOrgID, ctx, 2, api.BackupInfo_StatusInfo_InProgress)
			dash.VerifyFatal(err, nil, "Checking if second scheduled backup  in desired state")
			secondScheduleBackupName, err := GetOrdinalScheduleBackupName(ctx, scheduleName, 2, BackupOrgID)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of second partial parallel schedule backup [%s] has been created within schedule policy interval [%v]", secondScheduleBackupName, schedulePolicyInterval))
			err = ValidateLocalSnapshotCompletedForSuccessfulVolumes(secondScheduleBackupName, BackupOrgID, 1, failedPvcs, ctx)
			dash.VerifyFatal(err, nil, "Checking if local Snapshots are created or not")
			wg.Add(1)
			// Go routine to stop backups for the PVCs
			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				var wg1 sync.WaitGroup
				wg1.Add(len(failedPvcs))
				for _, pvc := range failedPvcs {
					go func(pvc *corev1.PersistentVolumeClaim) {
						defer GinkgoRecover()
						defer wg1.Done()
						log.Infof("Stopping all cs backups for %s [%s] in namespace %s", pvc.Name, pvc.Spec.VolumeName, pvc.Namespace)
						err = StopCloudsnapBackup(pvc.Name, pvc.Namespace)
						log.FailOnError(err, fmt.Sprintf("error stopping all cs backups for %s [%s] in namespace %s", pvc.Name, pvc.Spec.VolumeName, pvc.Namespace))
					}(pvc)
				}
				wg1.Wait()
			}()
			wg.Wait()
			partialSuccessBackups = append(partialSuccessBackups, secondScheduleBackupName)
		})

		Step("waiting for the third schedule backup to complete with success state", func() {
			log.InfoD("waiting for the third schedule backup to complete with success state")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = WaitTillScheduleBackupInDesiredState(scheduleName, BackupOrgID, ctx, 3, api.BackupInfo_StatusInfo_InProgress)
			dash.VerifyFatal(err, nil, "Checking if third scheduled backup in desired state")
			thirdScheduleBackupName, err = GetOrdinalScheduleBackupName(ctx, scheduleName, 3, BackupOrgID)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of third parallel schedule backup [%s] has been created within schedule policy interval [%v]", thirdScheduleBackupName, schedulePolicyInterval))
			successBackups = append(successBackups, thirdScheduleBackupName)
			err = ValidateLocalSnapshotCompleted(thirdScheduleBackupName, BackupOrgID, 1, ctx)
			dash.VerifyFatal(err, nil, "Checking if local Snapshots are created or not")
			err = SuspendBackupSchedule(scheduleName, schedulePolicyName, BackupOrgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Suspend schedule [%s]", scheduleName))
		})

		Step("check if the first 2 backups are in partial backup success state and the last one should be in success state", func() {
			log.InfoD("check if the first 2 backups are in partial backup success state and the last one should be in success state")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, partialSuccessBackup := range partialSuccessBackups {
				err = BackupWithPartialSuccessCheckWithValidation(ctx, partialSuccessBackup, scheduledAppContexts, BackupOrgID, MaxWaitPeriodForBackupCompletionInMinutes*5*time.Minute, 30*time.Second, nil, failedPvcs)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Checking partial backup success for [%s]", partialSuccessBackup))
			}
			for _, successBackup := range successBackups {
				err = BackupSuccessCheckWithValidation(ctx, successBackup, scheduledAppContexts, BackupOrgID, MaxWaitPeriodForBackupCompletionInMinutes*5*time.Minute, 30*time.Second)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Checking backup success for [%s]", successBackup))
			}
		})

		Step("Restoring the partial backup with replace option", func() {
			log.InfoD("Restoring the partial backup with replace option")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			backupName := partialSuccessBackups[rand.Intn(len(partialSuccessBackups))]
			restoreName := fmt.Sprintf("%s-%s-%s", "restore-replace", backupName, RandomString(4))
			log.InfoD("Restoring from the [%s] backup with namespaceMapping [%v]", restoreName, namespaceMapping)
			err = CreatePartialRestoreWithReplacePolicyWithValidation(restoreName, backupName, namespaceMapping, DestinationClusterName, BackupOrgID, ctx, make(map[string]string), 2, scheduledAppContexts, failedPvcs)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of restore with partial backup with replace option [%s]", restoreName))

		})

		Step("Restoring the partial backups with retain option", func() {
			log.InfoD("Restoring the partial backups  with retain option")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			backupName := partialSuccessBackups[rand.Intn(len(partialSuccessBackups))]
			restoreName := fmt.Sprintf("%s-%s-%s", "restore-retain", backupName, RandomString(4))
			log.InfoD("Restoring from the [%s] backup with namespaceMapping [%v]", restoreName, namespaceMapping)
			err = CreatePartialRestoreWithReplacePolicyWithValidation(restoreName, firstScheduleBackupName, namespaceMapping, DestinationClusterName, BackupOrgID, ctx, make(map[string]string), 1, scheduledAppContexts, failedPvcs)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of restore partial backups with retain option [%s]", restoreName))

		})

		Step("Restoring the partial backups with namespace mapping", func() {
			log.InfoD("Restoring the partial backups with namespace mapping")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			backupName := partialSuccessBackups[rand.Intn(len(partialSuccessBackups))]
			for _, appCtx := range scheduledAppContexts {
				namespaceMapping[appCtx.ScheduleOptions.Namespace] = appCtx.ScheduleOptions.Namespace + RandomString(4)
			}
			restoreName := fmt.Sprintf("%s-%s-%s", "restore", backupName, RandomString(4))
			log.InfoD("Restoring from the [%s] backup with namespaceMapping [%v]", restoreName, namespaceMapping)
			err = CreatePartialRestoreWithValidation(ctx, restoreName, backupName, namespaceMapping, make(map[string]string), DestinationClusterName, destClusterUid, BackupOrgID, scheduledAppContexts, failedPvcs)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of restore with namespace mapping[%s]", restoreName))

		})

		Step("Restoring the success backup with replace option", func() {
			log.InfoD("Restoring the  success  backup with replace option")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			restoreName := fmt.Sprintf("%s-%s-%s", "restore-replace", thirdScheduleBackupName, RandomString(4))
			log.InfoD("Restoring from the [%s] backup with namespaceMapping [%v]", restoreName, namespaceMapping)
			err = CreateRestoreWithReplacePolicyWithValidation(restoreName, thirdScheduleBackupName, namespaceMapping, DestinationClusterName, BackupOrgID, ctx, make(map[string]string), 2, scheduledAppContexts)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of restore with success backup with replace option [%s]", restoreName))

		})

		Step("Restoring the success backups with retain option", func() {
			log.InfoD("Restoring the  success backups with retain option")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			restoreName := fmt.Sprintf("%s-%s-%s", "restore-retain", thirdScheduleBackupName, RandomString(4))
			log.InfoD("Restoring from the [%s] backup with namespaceMapping [%v]", restoreName, namespaceMapping)
			err = CreateRestoreWithReplacePolicyWithValidation(restoreName, thirdScheduleBackupName, namespaceMapping, DestinationClusterName, BackupOrgID, ctx, make(map[string]string), 1, scheduledAppContexts)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of restore success backups with retain option [%s]", restoreName))

		})

		Step("Restoring the success backups with namespace mapping", func() {
			log.InfoD("Restoring the success backups with namespace mapping")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, appCtx := range scheduledAppContexts {
				namespaceMapping[appCtx.ScheduleOptions.Namespace] = appCtx.ScheduleOptions.Namespace + RandomString(4)
			}
			restoreName := fmt.Sprintf("%s-%s-%s", "restore", thirdScheduleBackupName, RandomString(4))
			log.InfoD("Restoring from the [%s] backup with namespaceMapping [%v]", restoreName, namespaceMapping)
			err = CreateRestoreWithValidation(ctx, restoreName, thirdScheduleBackupName, namespaceMapping, make(map[string]string), DestinationClusterName, destClusterUid, BackupOrgID, scheduledAppContexts)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of restore with namespace mapping[%s]", restoreName))
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		err = DeleteSchedule(scheduleName, SourceClusterName, BackupOrgID, ctx, true)
		dash.VerifySafely(err, nil, fmt.Sprintf("Deleting schedule [%s]", scheduleName))

		err = Inst().Backup.DeleteBackupSchedulePolicy(BackupOrgID, []string{schedulePolicyName})
		log.FailOnError(err, "Deleting backup schedule policies")

		log.Info("Destroying scheduled apps on source cluster")
		err = DestroyAppsWithData(scheduledAppContexts, opts, controlChannel, errorGroup)
		log.FailOnError(err, "Data validations failed")

		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})

// This testcase verify parallel schedule backup when network latency in induced.
var _ = Describe("{ValidateParallelScheduleWithNetworkLatency}", Label(TestCaseLabelsMap[ValidateParallelScheduleWithNetworkLatency]...), func() {

	var (
		scheduledAppContexts    []*scheduler.Context
		sourceClusterUid        string
		cloudCredName           string
		cloudCredUID            string
		backupLocationUID       string
		backupLocationName      string
		backupLocationMap       map[string]string
		labelSelectors          map[string]string
		providers               []string
		firstScheduleBackupName string
		namespaceMapping        map[string]string
		schedulePolicyName      string
		schedulePolicyUID       string
		schedulePolicyInterval  int64
		destClusterUid          string
		scheduleName            string
		backupNames             []string
		wg                      sync.WaitGroup
		namespaces              []string
		thirdScheduleBackupName string
		appNodesMap             map[string][]node.Node
		networkErrorDelay       int
		restoreNsMapping        map[string]map[string]string
		controlChannel          chan string
		errorGroup              *errgroup.Group
		appScaleFactor          float64
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("ValidateParallelScheduleWithNetworkLatency", "Verify parallel schedule backup success when network latency in induced.", nil, 304647, Ak, Q4FY25)

		backupLocationMap = make(map[string]string)
		namespaceMapping = make(map[string]string)
		labelSelectors = make(map[string]string)
		restoreNsMapping = make(map[string]map[string]string)
		appNodesMap = make(map[string][]node.Node)
		providers = GetBackupProviders()
		schedulePolicyInterval = 15
		networkErrorDelay = 2000
		numberOfDeployments := float64(1)
		appScaleFactor = math.Max(float64(Inst().GlobalScaleFactor), numberOfDeployments)
		log.InfoD("scheduling applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = []string{"mysql-backup-large-data"}
		for i := 0; i < int(appScaleFactor); i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, appCtx := range appContexts {
				appCtx.ReadinessTimeout = 300 * time.Minute
				scheduledAppContexts = append(scheduledAppContexts, appCtx)
				namespaces = append(namespaces, appCtx.ScheduleOptions.Namespace)
				appNodes, _ := Inst().S.GetNodesForApp(appCtx)
				appNodesMap[appCtx.ScheduleOptions.Namespace] = appNodes
			}
		}
	})

	It("Verify parallel schedule backup success when network latency in induced", func() {
		defer func() {
			log.InfoD("switching to default context")
			err := SetClusterContext("")
			log.FailOnError(err, "failed to SetClusterContext to default cluster")
		}()

		Step("Validating applications", func() {
			log.InfoD("Validating applications")
			ctx, _ := backup.GetAdminCtxFromSecret()
			controlChannel, errorGroup = ValidateApplicationsStartData(scheduledAppContexts, ctx)
		})

		Step("Setting cluster-wide average network bandwidth to 450MiB/s", func() {
			err := ThrottleNetworkSpeed(450)
			dash.VerifyFatal(err, nil, "Setting cluster-wide average network bandwidth to 450MiB/s")

			err = SetDestinationKubeConfig()
			log.FailOnError(err, "Switching context to destination cluster failed")

			err = ThrottleNetworkSpeed(450)
			dash.VerifyFatal(err, nil, "Setting cluster-wide average network bandwidth to 450MiB/s")

			defer func() {
				err = SetSourceKubeConfig()
				log.FailOnError(err, "Unable to switch context to source cluster [%s]", SourceClusterName)
			}()
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

		Step("Create schedule policy", func() {
			log.InfoD("Creating schedule policy")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			schedulePolicyName = fmt.Sprintf("%s-%v", "periodic-schedule-policy", RandomString(5))
			schedulePolicyUID = uuid.New()
			err = CreateBackupScheduleIntervalPolicy(5, schedulePolicyInterval, 5, schedulePolicyName, schedulePolicyUID, BackupOrgID, ctx, false, false)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of schedule policy %s", schedulePolicyName))
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

		Step("Taking first schedule backup of application from source cluster", func() {
			log.InfoD("Taking first schedule backup of application from source cluster")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			scheduleName = fmt.Sprintf("%s-%s", "autogenerated-schedule", RandomString(4))
			log.InfoD("creating schedule backup [%s] in source cluster [%s] (%s), organization [%s], in backup location [%s]", scheduleName, SourceClusterName, sourceClusterUid, BackupOrgID, backupLocationName)
			resp, err := CreateScheduleBackupWithoutCheck(scheduleName, SourceClusterName, sourceClusterUid, backupLocationName, backupLocationUID, namespaces, labelSelectors, BackupOrgID, "", "", "", "", schedulePolicyName, schedulePolicyUID, ctx, true)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of schedule backup for schedule [%s] ", scheduleName))
			firstScheduleBackupName, err = GetFirstScheduleBackupName(ctx, scheduleName, BackupOrgID)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of first parallel schedule backup [%s]", firstScheduleBackupName))
			dash.VerifyFatal(resp.BackupSchedule.ParallelBackup, true, "Verifying if parallelBackup var on Backup Schedule is set to True or not")
			backupNames = append(backupNames, firstScheduleBackupName)
			err = ValidateLocalSnapshotCompleted(firstScheduleBackupName, BackupOrgID, 1, ctx)
			dash.VerifyFatal(err, nil, "Checking if first local Snapshots are created or not")
		})

		Step("Waiting for second schedule backup of application from source cluster when network delay is injected", func() {
			log.InfoD("Waiting for second schedule backup of application from source cluster when network delay is injected")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, currentNamespace := range namespaces {
				log.Infof(fmt.Sprintf("Adding a delay of %dms to the applications nodes", networkErrorDelay))
				err = Inst().N.InjectNetworkErrorWithRebootFallback(appNodesMap[currentNamespace], "delay", "add", 0, networkErrorDelay)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Adding a delay of %dms to nodes", networkErrorDelay))
			}
			err = WaitTillScheduleBackupInDesiredState(scheduleName, BackupOrgID, ctx, 2, api.BackupInfo_StatusInfo_InProgress)
			dash.VerifyFatal(err, nil, "Checking if second backup in desired state")
			secondScheduleBackupName, err := GetOrdinalScheduleBackupName(ctx, scheduleName, 2, BackupOrgID)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of second parallel schedule backup [%s] is been created at schedule policy interval [%v]", secondScheduleBackupName, schedulePolicyInterval))
			backupNames = append(backupNames, secondScheduleBackupName)
			err = ValidateLocalSnapshotCompleted(secondScheduleBackupName, BackupOrgID, 1, ctx)
			dash.VerifyFatal(err, nil, "Checking if second local Snapshots are created or not")
		})

		Step("waiting for the third schedule backup to complete with success state", func() {
			log.InfoD("waiting for the third schedule backup to complete with success state")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = WaitTillScheduleBackupInDesiredState(scheduleName, BackupOrgID, ctx, 3, api.BackupInfo_StatusInfo_InProgress)
			dash.VerifyFatal(err, nil, "Checking if third backup in desired state")
			thirdScheduleBackupName, err = GetOrdinalScheduleBackupName(ctx, scheduleName, 3, BackupOrgID)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of third parallel schedule backup [%s] is been created at schedule policy interval [%v]", thirdScheduleBackupName, schedulePolicyInterval))
			backupNames = append(backupNames, thirdScheduleBackupName)
			err = ValidateLocalSnapshotCompleted(thirdScheduleBackupName, BackupOrgID, 1, ctx)
			dash.VerifyFatal(err, nil, "Checking if third local Snapshots are created or not")
			err = SuspendBackupSchedule(scheduleName, schedulePolicyName, BackupOrgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Suspend schedule [%s]", scheduleName))
			backupNames = append(backupNames, thirdScheduleBackupName)
		})

		Step("check the backup success state for all schedule backups", func() {
			log.InfoD("check the backup success state for all schedule backups")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, successBackup := range backupNames {
				err = BackupSuccessCheckWithValidation(ctx, successBackup, scheduledAppContexts, BackupOrgID, MaxWaitPeriodForBackupCompletionInMinutes*10*time.Minute, 30*time.Second)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Checking backup success for [%s]", successBackup))
			}
		})

		Step("Restoring the success backup with replace option", func() {
			log.InfoD("Restoring the  success  backup with replace option")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			backupName := backupNames[rand.Intn(len(backupNames))]
			for _, appCtx := range scheduledAppContexts {
				namespaceMapping[appCtx.ScheduleOptions.Namespace] = appCtx.ScheduleOptions.Namespace
			}
			restoreName := fmt.Sprintf("%s-%s-%s", "restore-replace", backupName, RandomString(4))
			log.InfoD("Restoring from the [%s] backup with namespaceMapping [%v]", restoreName, namespaceMapping)
			err = CreateRestoreWithReplacePolicyWithoutCheck(restoreName, backupName, namespaceMapping, DestinationClusterName, BackupOrgID, ctx, make(map[string]string), 2)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of restore with success backup with replace option [%s]", restoreName))
			restoreNsMapping[restoreName] = namespaceMapping

		})

		Step("Restoring the success backups with namespace mapping", func() {
			log.InfoD("Restoring the success backups with namespace mapping")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			backupName := backupNames[rand.Intn(len(backupNames))]
			namespaceMapping := make(map[string]string)
			for _, appCtx := range scheduledAppContexts {
				namespaceMapping[appCtx.ScheduleOptions.Namespace] = appCtx.ScheduleOptions.Namespace + RandomString(4)
			}
			restoreName := fmt.Sprintf("%s-%s-%s", "restore", backupName, RandomString(4))
			log.InfoD("Restoring from the [%s] backup with namespaceMapping [%v]", restoreName, namespaceMapping)
			_, err = CreateRestoreWithoutCheck(restoreName, backupName, namespaceMapping, DestinationClusterName, destClusterUid, BackupOrgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of restore[%s] with namespace mapping[%s]", restoreName, namespaceMapping))
			restoreNsMapping[restoreName] = namespaceMapping
		})

		Step("Validating all restores", func() {
			log.InfoD("Validating all restores")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			var mutex sync.Mutex
			errors := make([]string, 0)
			var wg sync.WaitGroup
			originalClusterConfigPath := CurrentClusterConfigPath
			if clusterConfigPath, ok := ClusterConfigPathMap[DestinationClusterName]; !ok {
				err = fmt.Errorf("switching cluster context: couldn't find clusterConfigPath for cluster [%s]", DestinationClusterName)
				log.FailOnError(err, "Failed switching cluster context to cluster [%s]", DestinationClusterName)
			} else {
				log.InfoD("Switching cluster context to cluster [%s]", DestinationClusterName)
				err = SetClusterContext(clusterConfigPath)
				if err != nil {
					log.FailOnError(err, "Failed switching cluster context to cluster [%s]", DestinationClusterName)
				}
			}
			defer func() {
				log.InfoD("Switching cluster context back to cluster path [%s]", originalClusterConfigPath)
				err := SetClusterContext(originalClusterConfigPath)
				if err != nil {
					log.FailOnError(err, "Failed switching cluster context back to cluster path [%s]", originalClusterConfigPath)
				}
			}()
			for restoreName, namespaceMapping := range restoreNsMapping {
				wg.Add(1)
				go func(restoreName string, namespaceMapping map[string]string) {
					defer wg.Done()
					log.InfoD("Validating restore [%s] with namespace mapping [%v]", restoreName, namespaceMapping)
					expectedRestoredAppContext, err := CloneAppContextAndTransformWithMappings(scheduledAppContexts[0], namespaceMapping, make(map[string]string), true)
					if err != nil {
						mutex.Lock()
						errors = append(errors, fmt.Sprintf("Failed while context tranforming of restore [%s]. Error - [%s]", restoreName, err.Error()))
						mutex.Unlock()
						return
					}
					err = RestoreSuccessCheck(restoreName, BackupOrgID, MaxWaitPeriodForRestoreCompletionInMinute*10*time.Minute, 30*time.Second, ctx)
					if err != nil {
						mutex.Lock()
						errors = append(errors, fmt.Sprintf("Failed while checking restore [%s]. Error - [%s]", restoreName, err.Error()))
						mutex.Unlock()
						return
					}
					err = ValidateRestore(ctx, restoreName, BackupOrgID, []*scheduler.Context{expectedRestoredAppContext}, make([]string, 0))
					if err != nil {
						mutex.Lock()
						errors = append(errors, fmt.Sprintf("Failed while validating restore [%s]. Error - [%s]", restoreName, err.Error()))
						mutex.Unlock()
						return
					}
				}(restoreName, namespaceMapping)
			}
			wg.Wait()
			dash.VerifyFatal(len(errors), 0, fmt.Sprintf("Validating restores of individual backups -\n%s", strings.Join(errors, "}\n{")))
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		err = ThrottleNetworkSpeed(0)
		log.FailOnError(err, "Setting cluster-wide average network bandwidth to default value")
		for _, currentNamespace := range namespaces {
			log.Infof(fmt.Sprintf("Removing the delay of %dms to the applications nodes", networkErrorDelay))
			err = Inst().N.InjectNetworkErrorWithRebootFallback(appNodesMap[currentNamespace], "delay", "del", 0, networkErrorDelay)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Removing the delay of %dms to nodes", networkErrorDelay))
		}
		err = DeleteSchedule(scheduleName, SourceClusterName, BackupOrgID, ctx, false)
		dash.VerifySafely(err, nil, fmt.Sprintf("Deleting schedule [%s]", scheduleName))
		backupNames, err := GetAllBackupsAdmin()
		dash.VerifySafely(err, nil, "Fetching all backups")
		for _, bkp := range backupNames {
			wg.Add(1)
			go func(bkp string) {
				defer wg.Done()
				backupUID, err := Inst().Backup.GetBackupUID(ctx, bkp, BackupOrgID)
				_, err = DeleteBackup(bkp, backupUID, BackupOrgID, ctx)
				dash.VerifySafely(err, nil, fmt.Sprintf("Verifying backup deletion - %s", bkp))
				err = DeleteBackupAndWait(bkp, ctx)
				dash.VerifySafely(err, nil, fmt.Sprintf("Verifying backup deletion wait- %s", bkp))
			}(bkp)
		}
		wg.Wait()

		log.Info("Destroying scheduled apps on source cluster")
		err = DestroyAppsWithData(scheduledAppContexts, opts, controlChannel, errorGroup)
		log.FailOnError(err, "Data validations failed")

		log.InfoD("switching to destination context")
		err = SetDestinationKubeConfig()
		log.FailOnError(err, "failed to switch to context to destination cluster")

		err = ThrottleNetworkSpeed(0)
		log.FailOnError(err, "Setting cluster-wide average network bandwidth to default value")

		log.InfoD("Destroying restored apps on destination clusters")
		restoredAppContexts := make([]*scheduler.Context, 0)
		for _, namespaceMapping := range restoreNsMapping {
			restoredAppContext, err := CloneAppContextAndTransformWithMappings(scheduledAppContexts[0], namespaceMapping, make(map[string]string), true)
			if err != nil {
				log.Errorf("TransformAppContextWithMappings: %v", err)
				continue
			}
			restoredAppContexts = append(restoredAppContexts, restoredAppContext)
		}
		DestroyApps(restoredAppContexts, opts)
		err = SetClusterContext("")
		log.FailOnError(err, "failed to SetClusterContext to default cluster")
		log.Info("Deleting restored namespaces")
		for restoreName, _ := range restoreNsMapping {
			err = DeleteRestore(restoreName, BackupOrgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting Restore [%s]", restoreName))
		}
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})
