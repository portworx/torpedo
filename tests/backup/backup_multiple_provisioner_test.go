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
	"math/rand"
)

// This MultipleProvisionerBackupAndRestore testcase to test the provisioner support will be removed at the time of merge
var _ = Describe("{MultipleProvisionerCsiKdmpBackupAndRestore}", func() {

	var (
		backupNames                               []string
		restoreNames                              []string
		scheduledAppContexts                      []*scheduler.Context
		sourceClusterUid                          string
		cloudCredName                             string
		cloudCredUID                              string
		backupLocationUID                         string
		backupLocationName                        string
		backupLocationMap                         map[string]string
		labelSelectors                            map[string]string
		providers                                 []string
		firstBkpLocationName                      string
		schedulePolicyName                        string
		schedulePolicyUID                         string
		scheduleUid                               string
		srcClusterUid                             string
		schedulePolicyInterval                    = int64(15)
		scheduledAppContextsForDefaultVscBackup   []*scheduler.Context
		scheduledAppContextsForCustomVscBackup    []*scheduler.Context
		allAppContext                             []*scheduler.Context
		defaultSchBackupName                      string
		scheduleList                              []string
		defaultProvisionerScheduleName            string
		nonDefaultVscSchBackupName                string
		randomStringLength                        = 10
		forceKdmpSchBackupName                    string
		appSpecList                               []string
		kdmpScheduleName                          string
		scheduledAppContextsForMultipleAppSinleNs []*scheduler.Context
		multipleProvisionerSameNsScheduleName     string
		multipleNsSchBackupName                   string
		clusterProviderName                       = GetClusterProvider()
		provisionerDefaultSnapshotClassMap        = GetProvisionerDefaultSnapshotMap(clusterProviderName)
		provisionerSnapshotClassMap               = GetProvisionerSnapshotClassesMap(clusterProviderName)
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("MultipleProvisionerCsiKdmpBackupAndRestore", "Backup and restore of namespaces with multiple provisioners using csi and kdmp", nil, 296724, Sn, Q4FY24)

		backupLocationMap = make(map[string]string)
		labelSelectors = make(map[string]string)
		providers = GetBackupProviders()

		// Deploy application for Default backup
		appSpecList = []string{"postgres"}
		applicationSpecIndex := rand.Intn(len(appSpecList))
		applicationSpec := appSpecList[applicationSpecIndex]

		// Deploy multiple application in a single namespace using different provisioner
		taskName := fmt.Sprintf("%s-%s", TaskNamePrefix, RandomString(randomStringLength))
		for provisioner, _ := range provisionerDefaultSnapshotClassMap {
			appSpec, err := GetApplicationSpecForProvisioner(clusterProviderName, provisioner, applicationSpec)
			log.FailOnError(err, fmt.Sprintf("Fetching application spec for provisioner %s", provisioner))
			appContexts := ScheduleApplicationsWithScheduleOptions(taskName, appSpec, provisioner)
			appContexts[0].ReadinessTimeout = AppReadinessTimeout
			scheduledAppContextsForMultipleAppSinleNs = append(scheduledAppContextsForMultipleAppSinleNs, appContexts...)
			allAppContext = append(allAppContext, appContexts...)
		}

		// Deploy multiple application in multiple namespace for default backup
		for provisioner, _ := range provisionerDefaultSnapshotClassMap {
			appSpec, err := GetApplicationSpecForProvisioner(clusterProviderName, provisioner, applicationSpec)
			log.FailOnError(err, fmt.Sprintf("Fetching application spec for provisioner %s", provisioner))
			taskName := fmt.Sprintf("%s-%v", TaskNamePrefix, Inst().InstanceID)
			appCtx := ScheduleApplicationsWithScheduleOptions(taskName, appSpec, provisioner)
			appCtx[0].ReadinessTimeout = AppReadinessTimeout
			scheduledAppContextsForDefaultVscBackup = append(scheduledAppContextsForDefaultVscBackup, appCtx...)
			allAppContext = append(allAppContext, appCtx...)
		}

		// Deploy multiple application in multiple namespace for custom backup
		for provisioner, _ := range provisionerSnapshotClassMap {
			appSpec, err := GetApplicationSpecForProvisioner(clusterProviderName, provisioner, applicationSpec)
			log.FailOnError(err, fmt.Sprintf("Fetching application spec for provisioner %s", provisioner))
			taskName := fmt.Sprintf("%s-%v", TaskNamePrefix, Inst().InstanceID)
			appCtx := ScheduleApplicationsWithScheduleOptions(taskName, appSpec, provisioner)
			appCtx[0].ReadinessTimeout = AppReadinessTimeout
			scheduledAppContextsForCustomVscBackup = append(scheduledAppContextsForCustomVscBackup, appCtx...)
			allAppContext = append(allAppContext, appCtx...)
		}
	})

	It("Backup and restore of namespaces with multiple provisioners using csi and kdmp", func() {
		Step("Validate deployed applications", func() {
			ValidateApplications(allAppContext)
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
		Step("Create schedule policy", func() {
			log.InfoD("Creating schedule policy")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			schedulePolicyName = fmt.Sprintf("%s-%v", "periodic-schedule-policy", RandomString(randomStringLength))
			schedulePolicyUID = uuid.New()
			err = CreateBackupScheduleIntervalPolicy(5, schedulePolicyInterval, 5, schedulePolicyName, schedulePolicyUID, BackupOrgID, ctx, false, false)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of schedule policy %s", schedulePolicyName))
		})
		Step(fmt.Sprintf("Creating schedule backup for application deployed using multiple provisioner with default volume snapshot class"), func() {
			log.InfoD("Creating schedule backup for multiple provisioner with default volume snapshot class")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			if len(provisionerDefaultSnapshotClassMap) > 0 {
				// Modify all provisioner to select default volume snapshot class
				provisionerSelectDefaultVolumeSnapshotClass := make(map[string]string)
				for key := range provisionerDefaultSnapshotClassMap {
					provisionerSelectDefaultVolumeSnapshotClass[key] = "default"
				}
				multipleProvisionerSameNsScheduleName = fmt.Sprintf("multiple-provisioner-same-namespace-schedule-%v", RandomString(randomStringLength))
				multipleNsSchBackupName, err = CreateScheduleBackupWithValidationWithVscMapping(ctx, multipleProvisionerSameNsScheduleName, SourceClusterName, backupLocationName, backupLocationUID, scheduledAppContextsForMultipleAppSinleNs, make(map[string]string), BackupOrgID, "", "", "", "", schedulePolicyName, schedulePolicyUID, provisionerSelectDefaultVolumeSnapshotClass, false)
				scheduleList = append(scheduleList, multipleNsSchBackupName)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of scheduled backup with schedule name [%s] for backup location %s", multipleNsSchBackupName, backupLocationName))
				err = IsFullBackup(multipleNsSchBackupName, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying if the first schedule backup [%s] for backup location %s is a full backup", multipleNsSchBackupName, backupLocationName))
				scheduleUid, err = Inst().Backup.GetBackupScheduleUID(ctx, multipleNsSchBackupName, BackupOrgID)
				err = DeleteScheduleWithUIDAndWait(multipleNsSchBackupName, scheduleUid, SourceClusterName, srcClusterUid, BackupOrgID, ctx)
			} else {
				log.InfoD("Skipping this step as provisioner with default volume snapshot class is not found")
			}
		})
		Step(fmt.Sprintf("Creating schedule backup for each namespace having application deployed with different provisioner which has default volume snapshot class"), func() {
			log.InfoD("Creating schedule backup for multiple provisioner which has default volume snapshot class")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			if len(provisionerDefaultSnapshotClassMap) > 0 {
				// Modify all provisioner to select default volume snapshot class
				provisionerSelectDefaultVolumeSnapshotClass := make(map[string]string)
				for key := range provisionerDefaultSnapshotClassMap {
					provisionerSelectDefaultVolumeSnapshotClass[key] = "default"
				}
				defaultProvisionerScheduleName = fmt.Sprintf("default-provisioner-schedule-%v", RandomString(randomStringLength))
				defaultSchBackupName, err = CreateScheduleBackupWithValidationWithVscMapping(ctx, defaultProvisionerScheduleName, SourceClusterName, backupLocationName, backupLocationUID, scheduledAppContextsForDefaultVscBackup, make(map[string]string), BackupOrgID, "", "", "", "", schedulePolicyName, schedulePolicyUID, provisionerSelectDefaultVolumeSnapshotClass, false)
				scheduleList = append(scheduleList, defaultSchBackupName)

				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of scheduled backup with schedule name [%s] for backup location %s", defaultSchBackupName, backupLocationName))
				err = IsFullBackup(defaultSchBackupName, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying if the first schedule backup [%s] for backup location %s is a full backup", defaultSchBackupName, backupLocationName))
				scheduleUid, err = Inst().Backup.GetBackupScheduleUID(ctx, defaultSchBackupName, BackupOrgID)
				err = DeleteScheduleWithUIDAndWait(defaultSchBackupName, scheduleUid, SourceClusterName, srcClusterUid, BackupOrgID, ctx)
			} else {
				log.InfoD("Skipping this step as provisioner with default volume snapshot class is not found")
			}
		})
		Step(fmt.Sprintf("Creating schedule backup for multiple provisioner with non default volume snapshot class"), func() {
			log.InfoD("Creating schedule backup for multiple provisioner with non default volume snapshot class")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			if len(provisionerSnapshotClassMap) > 0 {
				nonDefaultProvisionerScheduleName := fmt.Sprintf("default-provisioner-schedule-%v", RandomString(randomStringLength))
				nonDefaultVscSchBackupName, err = CreateScheduleBackupWithValidationWithVscMapping(ctx, nonDefaultProvisionerScheduleName, SourceClusterName, backupLocationName, backupLocationUID, scheduledAppContextsForDefaultVscBackup, make(map[string]string), BackupOrgID, "", "", "", "", schedulePolicyName, schedulePolicyUID, provisionerSnapshotClassMap, false)
				scheduleList = append(scheduleList, nonDefaultProvisionerScheduleName)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of scheduled backup with schedule name [%s] for backup location %s", nonDefaultVscSchBackupName, backupLocationName))
				err = IsFullBackup(nonDefaultVscSchBackupName, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying if the first schedule backup [%s] for backup location %s is a full backup", nonDefaultVscSchBackupName, backupLocationName))
				scheduleUid, err = Inst().Backup.GetBackupScheduleUID(ctx, nonDefaultVscSchBackupName, BackupOrgID)
				err = DeleteScheduleWithUIDAndWait(nonDefaultVscSchBackupName, scheduleUid, SourceClusterName, srcClusterUid, BackupOrgID, ctx)
			} else {
				log.InfoD("Skipping this step as provisioner with non-default volumeSnapshotClass is not found")
			}
		})
		Step(fmt.Sprintf("Creating schedule backup for multiple provisioner with forced kdmp option"), func() {
			log.InfoD("Creating schedule backup for multiple provisioner with forced kdmp option")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			provisionerNonDefaultSnapshotClassMap := map[string]string{}
			kdmpScheduleName = fmt.Sprintf("default-provisioner-schedule-%v", RandomString(randomStringLength))
			forceKdmpSchBackupName, err = CreateScheduleBackupWithValidationWithVscMapping(ctx, kdmpScheduleName, SourceClusterName, backupLocationName, backupLocationUID, scheduledAppContextsForDefaultVscBackup, make(map[string]string), BackupOrgID, "", "", "", "", schedulePolicyName, schedulePolicyUID, provisionerNonDefaultSnapshotClassMap, true)
			scheduleList = append(scheduleList, kdmpScheduleName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of scheduled backup with schedule name [%s] for backup location %s", nonDefaultVscSchBackupName, firstBkpLocationName))
			err = IsFullBackup(forceKdmpSchBackupName, BackupOrgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching the name of the next schedule backup for schedule: [%s] for backup location %s", nonDefaultVscSchBackupName, firstBkpLocationName))
		})
		Step("Taking manual backup of application from source cluster", func() {
			log.InfoD("taking manual backup of applications")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			backupNames = make([]string, 0)
			if len(provisionerDefaultSnapshotClassMap) > 0 {
				for i, appCtx := range scheduledAppContextsForDefaultVscBackup {
					scheduledNamespace := appCtx.ScheduleOptions.Namespace
					backupName := fmt.Sprintf("%s-%s-%v", "autogenerated-backup", scheduledNamespace, time.Now().Unix())
					provisionerVolumeSnapshotClassSubMap := make(map[string]string)
					if value, ok := provisionerDefaultSnapshotClassMap[appCtx.ScheduleOptions.StorageProvisioner]; ok {
						provisionerVolumeSnapshotClassSubMap[appCtx.ScheduleOptions.StorageProvisioner] = value
					}
					log.InfoD("creating backup [%s] in source cluster [%s] (%s), organization [%s], of namespace [%s], in backup location [%s]", backupName, SourceClusterName, sourceClusterUid, BackupOrgID, scheduledNamespace, backupLocationName)
					err = CreateBackupWithValidationWithVscMapping(ctx, backupName, SourceClusterName, backupLocationName, backupLocationUID, scheduledAppContextsForDefaultVscBackup[i:i+1], labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "", provisionerVolumeSnapshotClassSubMap, false)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
					backupNames = append(backupNames, backupName)
				}
			} else {
				log.InfoD("Skipping this step as provisioner with default volumeSnapshotClass is not found")
			}
		})
		Step("Restoring the backup taken on singe namespace with multiple application deployed with different provisioner", func() {
			if multipleNsSchBackupName != "" {
				log.InfoD("Restoring the backup taken on singe namespace with multiple application deployed with different provisioner")
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-central-admin ctx")
				restoreName := fmt.Sprintf("%s-%s-%s", "test-restore", "multi-app-single-ns", RandomString(randomStringLength))
				log.InfoD("Restoring namespaces from the [%s] backup", multipleNsSchBackupName)
				err = CreateRestoreWithValidation(ctx, restoreName, multipleNsSchBackupName, make(map[string]string), make(map[string]string), DestinationClusterName, BackupOrgID, scheduledAppContextsForMultipleAppSinleNs)
				restoreNames = append(restoreNames, restoreName)
			}
		})
		Step("Restoring the backup taken on multiple namespace with different provisioner", func() {
			if defaultSchBackupName != "" {
				log.InfoD("Restoring the backup taken on multiple namespace with different provisioner")
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-central-admin ctx")
				restoreName := fmt.Sprintf("%s-%s-%s", "test-restore", "multi-ns-different-provisioner", RandomString(randomStringLength))
				log.InfoD("Restoring namespaces from the [%s] backup", defaultSchBackupName)
				err = CreateRestoreWithValidation(ctx, restoreName, defaultSchBackupName, make(map[string]string), make(map[string]string), DestinationClusterName, BackupOrgID, scheduledAppContextsForDefaultVscBackup)
				restoreNames = append(restoreNames, restoreName)
			}
		})
		Step("Restoring the backup taken on multiple provisioner with non default volumeSnapshotClass", func() {
			if nonDefaultVscSchBackupName != "" {
				log.InfoD("Restoring the multiple provisioner with default volume snapshot class")
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-central-admin ctx")
				restoreName := fmt.Sprintf("%s-%s-%s", "test-restore", "non-default-provisioner", RandomString(randomStringLength))
				log.InfoD("Restoring namespaces from the [%s] backup", nonDefaultVscSchBackupName)
				err = CreateRestoreWithValidation(ctx, restoreName, nonDefaultVscSchBackupName, make(map[string]string), make(map[string]string), DestinationClusterName, BackupOrgID, scheduledAppContextsForCustomVscBackup)
				restoreNames = append(restoreNames, restoreName)
			}
		})
		Step("Restoring the backup taken on multiple provisioner with kdmp", func() {
			if forceKdmpSchBackupName != "" {
				log.InfoD("Restoring the multiple provisioner with default volumeSnapshotClass")
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-central-admin ctx")
				restoreName := fmt.Sprintf("%s-%s-%s", "test-restore", "kdmp", RandomString(randomStringLength))
				log.InfoD("Restoring namespaces from the [%s] backup", forceKdmpSchBackupName)
				err = CreateRestoreWithValidation(ctx, restoreName, forceKdmpSchBackupName, make(map[string]string), make(map[string]string), DestinationClusterName, BackupOrgID, scheduledAppContextsForCustomVscBackup)
				restoreNames = append(restoreNames, restoreName)
			}
		})
		Step("Restoring from the manual back backup", func() {
			if len(backupNames) > 0 {
				log.InfoD("Restoring from the manual back backup")
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-central-admin ctx")
				for i, appCtx := range scheduledAppContextsForCustomVscBackup {
					scheduledNamespace := appCtx.ScheduleOptions.Namespace
					restoreName := fmt.Sprintf("%s-%s-%s", "test-restore-manual-backup", scheduledNamespace, RandomString(randomStringLength))
					for strings.Contains(strings.Join(restoreNames, ","), restoreName) {
						restoreName = fmt.Sprintf("%s-%s-%s", "test-restore-manual-backup", scheduledNamespace, RandomString(randomStringLength))
					}
					log.InfoD("Restoring [%s] namespace from the [%s] backup", scheduledNamespace, backupNames[i])
					err = CreateRestoreWithValidation(ctx, restoreName, backupNames[i], make(map[string]string), make(map[string]string), DestinationClusterName, BackupOrgID, scheduledAppContexts[i:i+1])
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of restore [%s]", restoreName))
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
		}()

		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true

		for _, schName := range scheduleList {
			scheduleUid, err = Inst().Backup.GetBackupScheduleUID(ctx, schName, BackupOrgID)
			err = DeleteScheduleWithUIDAndWait(schName, scheduleUid, SourceClusterName, srcClusterUid, BackupOrgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting schedule %s for backup location %s", defaultSchBackupName, firstBkpLocationName))
		}

		log.InfoD("Deleting the deployed apps after the testcase")
		DestroyApps(allAppContext, opts)

		log.InfoD("switching to default context")
		err = SetClusterContext("")
		log.FailOnError(err, "failed to SetClusterContext to default cluster")

		// Delete restores
		log.Info("Delete restores")
		for _, restoreName := range restoreNames {
			err = DeleteRestore(restoreName, BackupOrgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting restore [%s]", restoreName))
		}
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})
