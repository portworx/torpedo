package tests

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

// This testcase verifies backup and restore with non existing and deleted custom stork admin namespaces
var _ = Describe("{BackupandRestoreWithNonExistingAdminNameSpace}", func() {

	var (
		newAdminNamespace    string // New admin namespace to be set as custom admin namespace
		backupName           string
		scheduledAppContexts []*scheduler.Context
		// bkpNamespaces               []string
		// clusterUid                  string
		// clusterStatus               api.ClusterInfo_StatusInfo_Status
		// restoreName                 string
		cloudCredName string
		cloudCredUID  string
		// backupLocationUID           string
		// bkpLocationName             string
		// numDeployments              int
		// providers                   []string
		backupLocationMap map[string]string
		// labelSelectors              map[string]string
		// selectedBkpNamespaceMapping map[string]string
		// multipleRestoreMapping      map[string]string
		restoreNames []string
		// backupNames                 []string
		// periodicSchedulePolicyName  string
		// periodicSchedulePolicyUid   string
		// schPolicyUid                string
		// scheduleName                string
		// scheduleBackupName          string
		scheduleNames []string
	)
	JustBeforeEach(func() {
		newAdminNamespace = StorkNamePrefix + "-" + RandomString(5) // Randomly generating the value for custom admin namespace
		backupName = fmt.Sprintf("%s-%v", BackupNamePrefix, time.Now().Unix())
		// bkpNamespaces = make([]string, 0)
		// restoreName = fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
		backupLocationMap = make(map[string]string)
		// labelSelectors = make(map[string]string)

		// numDeployments = 2 // 2 apps deployed in 2 namespaces
		// providers = getProviders()

		StartPxBackupTorpedoTest("BackupAndRestoreWithNonExistingAdminNamespace", "Bakcup and restore with non existing custom namespace", nil, 83717, ATrivedi, Q4FY24)
		log.InfoD(fmt.Sprintf("App list %v", Inst().AppList))
		scheduledAppContexts = make([]*scheduler.Context, 0)
		log.InfoD("Starting to deploy applications")
		// for i := 0; i < numDeployments; i++ {
		// 	log.InfoD(fmt.Sprintf("Iteration %v of deploying applications", i))
		// 	taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
		// 	appContexts := ScheduleApplications(taskName)
		// 	for _, ctx := range appContexts {
		// 		ctx.ReadinessTimeout = appReadinessTimeout
		// 		namespace := GetAppNamespace(ctx, taskName)
		// 		bkpNamespaces = append(bkpNamespaces, namespace)
		// 		scheduledAppContexts = append(scheduledAppContexts, ctx)
		// 	}
		// }
	})
	It("Backup and restore with non existing custom admin namespace", func() {

		// Step("Validating deployed applications", func() {
		// 	log.InfoD("Validating deployed applications")
		// 	ValidateApplications(scheduledAppContexts)
		// })
		// Step("Creating backup location and cloud setting", func() {
		// 	log.InfoD("Creating backup location and cloud setting")
		// 	ctx, err := backup.GetAdminCtxFromSecret()
		// 	log.FailOnError(err, "Fetching px-central-admin ctx")
		// 	for _, provider := range providers {
		// 		cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
		// 		bkpLocationName = fmt.Sprintf("%s-%s-bl", provider, getGlobalBucketName(provider))
		// 		cloudCredUID = uuid.New()
		// 		backupLocationUID = uuid.New()
		// 		backupLocationMap[backupLocationUID] = bkpLocationName
		// 		err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, orgID, ctx)
		// 		dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, orgID, provider))
		// 		err = CreateBackupLocation(provider, bkpLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), orgID, "", true)
		// 		dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", bkpLocationName))
		// 	}
		// })
		// Step("Creating new admin namespaces", func() {
		// 	log.InfoD("Creating new admin namespace - %v", newAdminNamespace)
		// 	nsSpec := &v1.Namespace{
		// 		ObjectMeta: meta_v1.ObjectMeta{
		// 			Name: newAdminNamespace,
		// 		},
		// 	}
		// 	ns, err := core.Instance().CreateNamespace(nsSpec)
		// 	log.FailOnError(err, "Unable to create namespace")
		// 	log.InfoD("Created Namespace - %v", ns.Name)
		// })
		Step("Modifying Admin Namespace for Stork", func() {
			log.InfoD("Modifying Admin Namespace for Stork to %v", newAdminNamespace)
			_, err := ChangeAdminNamespace(newAdminNamespace)
			log.FailOnError(err, "Unable to update admin namespace")
			log.Infof("Admin namespace updated successfully")
		})
		// Step("Registering cluster for backup", func() {
		// 	log.InfoD("Registering cluster for backup")
		// 	ctx, err := backup.GetAdminCtxFromSecret()
		// 	log.FailOnError(err, "Fetching px-central-admin ctx")
		// 	err = CreateApplicationClusters(orgID, "", "", ctx)
		// 	dash.VerifyFatal(err, nil, "Creating source and destination cluster")
		// 	clusterStatus, err = Inst().Backup.GetClusterStatus(orgID, SourceClusterName, ctx)
		// 	log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
		// 	dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
		// 	clusterUid, err = Inst().Backup.GetClusterUID(ctx, orgID, SourceClusterName)
		// 	dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		// })
		// Step("Taking backup of multiple namespaces", func() {
		// 	log.InfoD(fmt.Sprintf("Taking backup of multiple namespaces [%v]", bkpNamespaces))
		// 	ctx, err := backup.GetAdminCtxFromSecret()
		// 	log.FailOnError(err, "Fetching px-central-admin ctx")

		// 	appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
		// 	err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, bkpLocationName, backupLocationUID, appContextsToBackup, labelSelectors, orgID, clusterUid, "", "", "", "")
		// 	dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
		// 	backupNames = append(backupNames, backupName)
		// })
		// Step("Create schedule policy", func() {
		// 	log.InfoD("Creating a schedule policy")
		// 	ctx, err := backup.GetAdminCtxFromSecret()
		// 	log.FailOnError(err, "Unable to fetch px-central-admin ctx")
		// 	periodicSchedulePolicyName = fmt.Sprintf("%s-%v", "periodic", time.Now().Unix())
		// 	periodicSchedulePolicyUid = uuid.New()
		// 	periodicSchedulePolicyInfo := Inst().Backup.CreateIntervalSchedulePolicy(5, 15, 5)
		// 	err = Inst().Backup.BackupSchedulePolicy(periodicSchedulePolicyName, periodicSchedulePolicyUid, orgID, periodicSchedulePolicyInfo)
		// 	dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of periodic schedule policy of interval 15 minutes named [%s]", periodicSchedulePolicyName))
		// 	periodicSchedulePolicyUid, err = Inst().Backup.GetSchedulePolicyUid(orgID, ctx, periodicSchedulePolicyName)
		// 	dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching uid of periodic schedule policy named [%s]", periodicSchedulePolicyName))
		// })
		// Step("Creating schedule backups for applications", func() {
		// 	log.InfoD("Creating schedule backups")
		// 	ctx, err := backup.GetAdminCtxFromSecret()
		// 	log.FailOnError(err, "Fetching px-central-admin ctx")
		// 	schPolicyUid, _ = Inst().Backup.GetSchedulePolicyUid(orgID, ctx, periodicSchedulePolicyName)
		// 	scheduleName = fmt.Sprintf("%s-schedule-%v", BackupNamePrefix, time.Now().Unix())
		// 	scheduleBackupName, err = CreateScheduleBackupWithValidation(ctx, scheduleName, SourceClusterName, bkpLocationName, backupLocationUID, scheduledAppContexts, labelSelectors, orgID, "", "", "", "", periodicSchedulePolicyName, schPolicyUid)
		// 	dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of schedule backup with schedule name [%s]", scheduleName))
		// 	scheduleNames = append(scheduleNames, scheduleName)
		// 	backupNames = append(backupNames, scheduleBackupName)
		// })
		// Step("Restoring backup of multiple namespaces", func() {
		// 	log.InfoD("Restoring backup of multiple namespaces")
		// 	selectedBkpNamespaceMapping = make(map[string]string)
		// 	multipleRestoreMapping = make(map[string]string)
		// 	for _, namespace := range bkpNamespaces {
		// 		selectedBkpNamespaceMapping[namespace] = namespace
		// 	}
		// 	log.InfoD("Selected application namespaces to restore: [%v]", bkpNamespaces)
		// 	ctx, err := backup.GetAdminCtxFromSecret()
		// 	log.FailOnError(err, "Fetching px-central-admin ctx")
		// 	err = CreateRestore(restoreName, backupName, selectedBkpNamespaceMapping, SourceClusterName, orgID, ctx, make(map[string]string))
		// 	dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s]", restoreName))

		// 	// Restore to custom namespace
		// 	for _, namespace := range bkpNamespaces {
		// 		restoredNameSpace := fmt.Sprintf("%s-%v", backupName, time.Now().Unix())
		// 		multipleRestoreMapping[namespace] = restoredNameSpace
		// 	}
		// 	log.Infof("Custom restore map %v", multipleRestoreMapping)
		// 	customRestoreName := fmt.Sprintf("%s-%v", "multiple-application", time.Now().Unix())
		// 	err = CreateRestore(customRestoreName, backupName, multipleRestoreMapping, SourceClusterName, orgID, ctx, make(map[string]string))
		// 	dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying multiple backup restore [%s]", customRestoreName))
		// 	restoreNames = append(restoreNames, restoreName, customRestoreName)
		// })
		// Step("Restoring scheduled backups", func() {
		// 	log.InfoD("Restoring scheduled backups")
		// 	ctx, err := backup.GetAdminCtxFromSecret()
		// 	log.FailOnError(err, "Fetching px-central-admin ctx")
		// 	selectedBkpNamespaceMapping = make(map[string]string)
		// 	for _, namespace := range bkpNamespaces {
		// 		selectedBkpNamespaceMapping[namespace] = namespace
		// 	}
		// 	restoreName = fmt.Sprintf("%s-%s-%v", restoreNamePrefix, scheduleBackupName, time.Now().Unix())
		// 	err = CreateRestore(restoreName, scheduleBackupName, selectedBkpNamespaceMapping, SourceClusterName, orgID, ctx, nil)
		// 	dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of restoring scheduled backups - %s", restoreName))
		// 	restoreNames = append(restoreNames, restoreName)
		// })
		// Step("Deleting new admin namespace", func() {
		// 	log.Info("Deleting namespace - %v", newAdminNamespace)
		// 	err := DeleteAppNamespace(newAdminNamespace)
		// 	log.FailOnError(err, "Unable to delete admin namespace")
		// 	log.InfoD("Namespace - %v - deleted successfully", newAdminNamespace)
		// })
		// Step("Restoring backup of multiple namespaces after admin namespace removal", func() {
		// 	log.InfoD("Restoring backup of multiple namespaces - Admin namespace will be recreated")
		// 	selectedBkpNamespaceMapping = make(map[string]string)
		// 	multipleRestoreMapping = make(map[string]string)
		// 	restoreName = fmt.Sprintf("%s-%v", restoreNamePrefix, time.Now().Unix())
		// 	for _, namespace := range bkpNamespaces {
		// 		selectedBkpNamespaceMapping[namespace] = namespace
		// 	}
		// 	log.InfoD("Selected application namespaces to restore: [%v]", bkpNamespaces)
		// 	ctx, err := backup.GetAdminCtxFromSecret()
		// 	log.FailOnError(err, "Fetching px-central-admin ctx")
		// 	err = CreateRestore(restoreName, backupName, selectedBkpNamespaceMapping, SourceClusterName, orgID, ctx, make(map[string]string))
		// 	dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s]", restoreName))

		// 	// Restore to custom namespace
		// 	for _, namespace := range bkpNamespaces {
		// 		restoredNameSpace := fmt.Sprintf("%s-%v", backupName, time.Now().Unix())
		// 		multipleRestoreMapping[namespace] = restoredNameSpace
		// 	}
		// 	log.Infof("Custom restore map %v", multipleRestoreMapping)
		// 	customRestoreName := fmt.Sprintf("%s-%v", "multiple-application", time.Now().Unix())
		// 	err = CreateRestore(customRestoreName, backupName, multipleRestoreMapping, SourceClusterName, orgID, ctx, make(map[string]string))
		// 	dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s]", restoreName))
		// })
		// Step("Deleting new admin namespace again", func() {
		// 	log.Info("Deleting namespace - %v", newAdminNamespace)
		// 	err := DeleteAppNamespace(newAdminNamespace)
		// 	log.FailOnError(err, "Unable to delete admin namespace")
		// 	log.InfoD("Namespace - %v - deleted successfully", newAdminNamespace)
		// })
		// Step("Taking backup of multiple namespaces after admin namespace removal", func() { // This step should fail
		// 	backupName = fmt.Sprintf("%s-%v", BackupNamePrefix, time.Now().Unix())
		// 	restoreName = fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
		// 	log.InfoD(fmt.Sprintf("Taking backup of multiple namespaces [%v]", bkpNamespaces))
		// 	ctx, err := backup.GetAdminCtxFromSecret()
		// 	log.FailOnError(err, "Fetching px-central-admin ctx")

		// 	appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
		// 	err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, bkpLocationName, backupLocationUID, appContextsToBackup, labelSelectors, orgID, clusterUid, "", "", "", "")
		// 	dash.VerifyFatal(strings.Contains(err.Error(), "CR"), true, fmt.Sprintf("Backup creation failed due to non existing namespace [%s]. Error : %s", newAdminNamespace, err.Error()))
		// })
		// Step("Creating schedule backups for applications after admin namespace removal", func() {
		// 	log.InfoD("Creating schedule backups - This should recreate the admin namespace")
		// 	ctx, err := backup.GetAdminCtxFromSecret()
		// 	log.FailOnError(err, "Fetching px-central-admin ctx")
		// 	schPolicyUid, _ = Inst().Backup.GetSchedulePolicyUid(orgID, ctx, periodicSchedulePolicyName)
		// 	scheduleName = fmt.Sprintf("%s-schedule-%v", BackupNamePrefix, time.Now().Unix())
		// 	scheduleBackupName, err = CreateScheduleBackupWithValidation(ctx, scheduleName, SourceClusterName, bkpLocationName, backupLocationUID, scheduledAppContexts, labelSelectors, orgID, "", "", "", "", periodicSchedulePolicyName, schPolicyUid)
		// 	dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of schedule backup with schedule name [%s]", scheduleName))
		// 	scheduleNames = append(scheduleNames, scheduleName)
		// 	backupNames = append(backupNames, scheduleBackupName)
		// })
		// Step("Restoring scheduled backups", func() {
		// 	log.InfoD("Restoring scheduled backups")
		// 	ctx, err := backup.GetAdminCtxFromSecret()
		// 	log.FailOnError(err, "Fetching px-central-admin ctx")
		// 	selectedBkpNamespaceMapping = make(map[string]string)
		// 	for _, namespace := range bkpNamespaces {
		// 		selectedBkpNamespaceMapping[namespace] = namespace
		// 	}
		// 	restoreName = fmt.Sprintf("%s-%s-%v", restoreNamePrefix, scheduleBackupName, time.Now().Unix())
		// 	err = CreateRestore(restoreName, scheduleBackupName, selectedBkpNamespaceMapping, SourceClusterName, orgID, ctx, nil)
		// 	dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of restoring scheduled backups - %s", restoreName))
		// 	restoreNames = append(restoreNames, restoreName)
		// })
	})
	JustAfterEach(func() {
		log.InfoD("Just After Each")
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		log.Infof("Deleting backup schedule policy")
		for _, scheduleName := range scheduleNames {
			err = DeleteSchedule(scheduleName, SourceClusterName, orgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Verification of deleting backup schedule - %s", scheduleName))
		}
		log.InfoD("Deleting deployed applications")
		DestroyApps(scheduledAppContexts, opts)
		backupDriver := Inst().Backup
		backupUID, err := backupDriver.GetBackupUID(ctx, backupName, orgID)
		log.FailOnError(err, "Failed while trying to get backup UID for - [%s]", backupName)
		log.InfoD("Deleting backups")
		for _, backups := range restoreNames {
			_, err = DeleteBackup(backups, backupUID, orgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Verifying the deletion of the backup named [%s]", backups))
		}
		log.InfoD("Deleting restore")
		log.InfoD(fmt.Sprintf("Backup name %v", restoreNames))
		for _, restores := range restoreNames {
			err := DeleteRestore(restores, orgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Verifying the deletion of the restore named [%s]", restores))
		}
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})

})
