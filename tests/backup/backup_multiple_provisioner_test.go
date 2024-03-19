package tests

import (
	"fmt"
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

// This MultipleProvisionerBackupAndRestore testcase to test Backup and restore of namespaces with multiple provisioners using csi and kdmp
var _ = Describe("{MultipleProvisionerCsiSnapshotDeleteBackupAndRestore}", func() {

	var (
		//backupNames                               []string
		restoreNames         []string
		scheduledAppContexts []*scheduler.Context
		//sourceClusterUid                          string
		cloudCredName      string
		cloudCredUID       string
		backupLocationUID  string
		backupLocationName string
		backupLocationMap  map[string]string
		//labelSelectors                            map[string]string
		providers []string
		//firstBkpLocationName   string
		schedulePolicyName     string
		schedulePolicyUID      string
		scheduleUid            string
		srcClusterUid          string
		schedulePolicyInterval = int64(15)
		//scheduledAppContextsForDefaultVscBackup   []*scheduler.Context
		allAppContext []*scheduler.Context
		//defaultSchBackupName string
		scheduleList []string
		//defaultProvisionerScheduleName            string
		//nonDefaultVscSchBackupName                string
		randomStringLength = 10
		//forceKdmpSchBackupName                    string
		appSpecList []string
		//kdmpScheduleName                          string
		scheduledAppContextsForMultipleAppSinleNs []*scheduler.Context
		multipleProvisionerSameNsScheduleName     string
		multipleNsSchBackupName                   string
		clusterProviderName                       = GetClusterProvider()
		provisionerDefaultSnapshotClassMap        = GetProvisionerDefaultSnapshotMap(clusterProviderName)
		//provisionerSnapshotClassMap               = GetProvisionerSnapshotClassesMap(clusterProviderName)
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("MultipleProvisionerCsiKdmpBackupAndRestore", "Backup and restore of namespaces with multiple provisioners using csi and kdmp", nil, 296724, Sn, Q4FY24)

		backupLocationMap = make(map[string]string)
		//labelSelectors = make(map[string]string)
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

			//sourceClusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
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
				scheduleList = append(scheduleList, multipleProvisionerSameNsScheduleName)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of scheduled backup with schedule name [%s] for backup location %s", multipleNsSchBackupName, backupLocationName))
				err = IsFullBackup(multipleNsSchBackupName, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying if the first schedule backup [%s] for backup location %s is a full backup", multipleNsSchBackupName, backupLocationName))
				scheduleUid, err = Inst().Backup.GetBackupScheduleUID(ctx, multipleProvisionerSameNsScheduleName, BackupOrgID)
				err = DeleteScheduleWithUIDAndWait(multipleProvisionerSameNsScheduleName, scheduleUid, SourceClusterName, srcClusterUid, BackupOrgID, ctx)
				backupUID, err := Inst().Backup.GetBackupUID(ctx, multipleNsSchBackupName, BackupOrgID)
				log.FailOnError(err, fmt.Sprintf("Getting UID for backup %v", multipleNsSchBackupName))
				backupInspectRequest := &api.BackupInspectRequest{
					Name:  multipleNsSchBackupName,
					Uid:   backupUID,
					OrgId: BackupOrgID,
				}
				resp, err := Inst().Backup.InspectBackup(ctx, backupInspectRequest)
				volumeObjlist := resp.Backup.Volumes
				var volumeNames []string
				for _, obj := range volumeObjlist {
					log.Infof("sleep for 60s %v", obj.Name)
					volumeNames = append(volumeNames, obj.Name)
				}
				DeleteSnapshotsForVolumes(volumeNames)
			} else {
				log.InfoD("Skipping this step as provisioner with default volume snapshot class is not found")
			}
		})
		Step("Restoring the backup taken on singe namespace with multiple application deployed with different provisioner", func() {
			if multipleNsSchBackupName != "" {
				log.InfoD("Restoring the backup taken on singe namespace with multiple application deployed with different provisioner")
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-central-admin ctx")
				namespaceMappingMultiApp := make(map[string]string)
				for _, appCtx := range scheduledAppContextsForMultipleAppSinleNs {
					namespaceMappingMultiApp[appCtx.ScheduleOptions.Namespace] = appCtx.ScheduleOptions.Namespace + "-mul-app-snigle-ns"
				}
				restoreName := fmt.Sprintf("%s-%s-%s", "test-restore", "multi-app-single-ns", RandomString(randomStringLength))
				log.InfoD("Restoring namespaces from the [%s] backup", multipleNsSchBackupName)
				err = CreateRestoreWithValidation(ctx, restoreName, multipleNsSchBackupName, namespaceMappingMultiApp, make(map[string]string), DestinationClusterName, BackupOrgID, scheduledAppContextsForMultipleAppSinleNs)
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

		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true

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
