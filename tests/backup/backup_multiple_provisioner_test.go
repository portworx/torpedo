package tests

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

// This Dummy testcase to test the provisioner support will be removed at the time of merge
var _ = Describe("{RestoreCsiBackupAfterSnapshotDelete}", func() {

	var (
		backupNames                       []string
		restoreNames                      []string
		scheduledAppContexts              []*scheduler.Context
		preRuleNameList                   []string
		postRuleNameList                  []string
		sourceClusterUid                  string
		cloudCredName                     string
		cloudCredUID                      string
		backupLocationUID                 string
		backupLocationName                string
		appList                           []string
		backupLocationMap                 map[string]string
		labelSelectors                    map[string]string
		providers                         []string
		intervalName                      string
		dailyName                         string
		weeklyName                        string
		monthlyName                       string
		provisionerVolumeSnapshotClassMap map[string]string
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("BasicBackupCreationDummyTest", "Deploying backup", nil, 84755, Sagrawal, Q4FY23)

		appList = Inst().AppList
		backupLocationMap = make(map[string]string)
		labelSelectors = make(map[string]string)
		providers = GetBackupProviders()
		intervalName = fmt.Sprintf("%s-%v", "interval", time.Now().Unix())
		dailyName = fmt.Sprintf("%s-%v", "daily", time.Now().Unix())
		weeklyName = fmt.Sprintf("%s-%v", "weekly", time.Now().Unix())
		monthlyName = fmt.Sprintf("%s-%v", "monthly", time.Now().Unix())

		provisionerVolumeSnapshotClassMap = map[string]string{
			"openshift-storage.cephfs.csi.ceph.com": "ocs-storagecluster-cephfsplugin-snapclass",
			"openshift-storage.rbd.csi.ceph.com":    "ocs-storagecluster-rbdsplugin-snapclass",
		}

		taskName := fmt.Sprintf("%s-%v", TaskNamePrefix, Inst().InstanceID)
		appCtx, err := Inst().S.Schedule(taskName, scheduler.ScheduleOptions{
			AppKeys:            []string{"postgres-cephfs-csi"},
			StorageProvisioner: "openshift-storage.cephfs.csi.ceph.com",
		})
		appCtx[0].ReadinessTimeout = AppReadinessTimeout
		scheduledAppContexts = append(scheduledAppContexts, appCtx...)
		log.FailOnError(err, "Failed to schedule %v", appCtx[0].App.Key)

		taskName = fmt.Sprintf("%s-%v", TaskNamePrefix, Inst().InstanceID)
		appCtx, err = Inst().S.Schedule(taskName, scheduler.ScheduleOptions{
			AppKeys:            []string{"postgres-rbd-csi"},
			StorageProvisioner: "openshift-storage.rbd.csi.ceph.com",
		})
		appCtx[0].ReadinessTimeout = AppReadinessTimeout
		scheduledAppContexts = append(scheduledAppContexts, appCtx...)
		log.FailOnError(err, "Failed to schedule %v", appCtx[0].App.Key)
	})

	It("Basic Backup Creation", func() {
		defer func() {
			log.InfoD("switching to default context")
			err := SetClusterContext("")
			log.FailOnError(err, "failed to SetClusterContext to default cluster")
		}()

		Step("Creating rules for backup", func() {
			log.InfoD("Creating rules for backup")
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
			log.InfoD("Creating backup interval schedule policy")
			intervalSchedulePolicyInfo := Inst().Backup.CreateIntervalSchedulePolicy(5, 15, 2)
			intervalPolicyStatus := Inst().Backup.BackupSchedulePolicy(intervalName, uuid.New(), BackupOrgID, intervalSchedulePolicyInfo)
			dash.VerifyFatal(intervalPolicyStatus, nil, "Creating interval schedule policy")

			log.InfoD("Creating backup daily schedule policy")
			dailySchedulePolicyInfo := Inst().Backup.CreateDailySchedulePolicy(1, "9:00AM", 2)
			dailyPolicyStatus := Inst().Backup.BackupSchedulePolicy(dailyName, uuid.New(), BackupOrgID, dailySchedulePolicyInfo)
			dash.VerifyFatal(dailyPolicyStatus, nil, "Creating daily schedule policy")

			log.InfoD("Creating backup weekly schedule policy")
			weeklySchedulePolicyInfo := Inst().Backup.CreateWeeklySchedulePolicy(1, backup.Friday, "9:10AM", 2)
			weeklyPolicyStatus := Inst().Backup.BackupSchedulePolicy(weeklyName, uuid.New(), BackupOrgID, weeklySchedulePolicyInfo)
			dash.VerifyFatal(weeklyPolicyStatus, nil, "Creating weekly schedule policy")

			log.InfoD("Creating backup monthly schedule policy")
			monthlySchedulePolicyInfo := Inst().Backup.CreateMonthlySchedulePolicy(1, 29, "9:20AM", 2)
			monthlyPolicyStatus := Inst().Backup.BackupSchedulePolicy(monthlyName, uuid.New(), BackupOrgID, monthlySchedulePolicyInfo)
			dash.VerifyFatal(monthlyPolicyStatus, nil, "Creating monthly schedule policy")
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
		})

		Step("Taking backup of application from source cluster", func() {
			log.InfoD("taking backup of applications")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			backupNames = make([]string, 0)
			for i, appCtx := range scheduledAppContexts {
				scheduledNamespace := appCtx.ScheduleOptions.Namespace
				backupName := fmt.Sprintf("%s-%s-%v", "autogenerated-backup", scheduledNamespace, time.Now().Unix())
				provisionerVolumeSnapshotClassSubMap := make(map[string]string)
				if value, ok := provisionerVolumeSnapshotClassMap[appCtx.ScheduleOptions.StorageProvisioner]; ok {
					provisionerVolumeSnapshotClassSubMap[appCtx.ScheduleOptions.StorageProvisioner] = value
				}
				log.InfoD("creating backup [%s] in source cluster [%s] (%s), organization [%s], of namespace [%s], in backup location [%s]", backupName, SourceClusterName, sourceClusterUid, BackupOrgID, scheduledNamespace, backupLocationName)
				err := CreateBackupWithValidationWithVscMapping(ctx, backupName, SourceClusterName, backupLocationName, backupLocationUID, scheduledAppContexts[i:i+1], labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "", provisionerVolumeSnapshotClassSubMap)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				backupNames = append(backupNames, backupName)
			}
		})

		Step("Restoring the backed up namespaces", func() {
			log.InfoD("Restoring the backed up namespaces")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for i, appCtx := range scheduledAppContexts {
				scheduledNamespace := appCtx.ScheduleOptions.Namespace
				restoreName := fmt.Sprintf("%s-%s-%s", "test-restore", scheduledNamespace, RandomString(4))
				for strings.Contains(strings.Join(restoreNames, ","), restoreName) {
					restoreName = fmt.Sprintf("%s-%s-%s", "test-restore", scheduledNamespace, RandomString(4))
				}
				log.InfoD("Restoring [%s] namespace from the [%s] backup", scheduledNamespace, backupNames[i])
				err = CreateRestoreWithValidation(ctx, restoreName, backupNames[i], make(map[string]string), make(map[string]string), DestinationClusterName, BackupOrgID, scheduledAppContexts[i:i+1])
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of restore [%s]", restoreName))
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
		}()

		policyList := []string{intervalName, dailyName, weeklyName, monthlyName}
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
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
		err = Inst().Backup.DeleteBackupSchedulePolicy(BackupOrgID, policyList)
		dash.VerifySafely(err, nil, "Deleting backup schedule policies")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true

		log.InfoD("switching to destination context")
		err = SetDestinationKubeConfig()
		log.FailOnError(err, "failed to switch to context to destination cluster")

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