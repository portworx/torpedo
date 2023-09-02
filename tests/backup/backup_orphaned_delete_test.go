package tests

import (
	"fmt"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

// DeleteBackupOfUserSharedRBAC delete backups created by non-admin user from px-admin with shared RBAC resources from px-admin.
var _ = Describe("{DeleteBackupOfUserSharedRBAC}", func() {
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/87560
	var (
		periodicSchedulePolicyName      string
		periodicSchedulePolicyUid       string
		scheduledAppContexts            []*scheduler.Context
		backupLocationUID               string
		credName                        string
		cloudCredUID                    string
		srcClusterUid                   string
		backupLocationName              string
		preRuleName                     string
		postRuleName                    string
		preRuleUid                      string
		postRuleUid                     string
		nsLabels                        map[string]string
		periodicSchedulePolicyInterval  int64
		namespaceLabel                  string
		wg                              sync.WaitGroup
		mutex                           sync.Mutex
		bkpNamespaces                   = make([]string, 0)
		userNames                       = make([]string, 0)
		numOfNS                         = 2
		numOfUsers                      = 6
		timeBetweenConsecutiveBackups   = 10 * time.Second
		namespaceMapping                = make(map[string]string)
		storageClassMapping             = make(map[string]string)
		userIdMap                       = make(map[string]string)
		clusterUidMap                   = make(map[string]map[string]string)
		backupLocationMap               = make(map[string]string)
		singleNamespaceBackupsMap       = make(map[string][]string)
		mutipleNamespaceBackupsMap      = make(map[string][]string)
		mutipleNamespaceLabelBackupsMap = make(map[string][]string)
		scheduleNameMap                 = make(map[string]string)
		restoreNameMap                  = make(map[string]string)
		userBackupNamesMap              = make(map[string][]string)
		userBackupNamesMapFromAdmin     = make(map[string][]string)
		userBackupSchedulesMap          = make(map[string][]string)
		userRestoresMap                 = make(map[string][]string)
		backupDriver                    = Inst().Backup
	)

	JustBeforeEach(func() {
		StartTorpedoTest("DeleteBackupOfUserSharedRBAC",
			"Delete backups,restores,schedules,clusters created by non-admin user with shared RBAC resources from px-admin", nil, 87560)
		log.InfoD("Deploy applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		for i := 0; i < numOfNS; i++ {
			taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = appReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
				scheduledAppContexts = append(scheduledAppContexts, ctx)
			}
		}
		log.InfoD("Created namespaces %v", bkpNamespaces)
	})
	It("Delete backups by user from admin with shared RBAC objects", func() {
		providers := getProviders()
		Step("Validate applications", func() {
			log.InfoD("Validate applications")
			ValidateApplications(scheduledAppContexts)
		})
		Step("Generate and add labels to namespaces", func() {
			log.InfoD("Generate and add labels to namespaces")
			nsLabels = GenerateRandomLabels(1)
			err := AddLabelsToMultipleNamespaces(nsLabels, bkpNamespaces)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Adding labels [%v] to namespaces [%s]", nsLabels, bkpNamespaces))
		})
		Step("Generating namespace label string from label map", func() {
			log.InfoD("Generating namespace label string from label map")
			namespaceLabel = MapToKeyValueString(nsLabels)
			log.Infof("Generated labels [%s]", namespaceLabel)
		})
		Step("Create Users with Different types of roles", func() {
			log.InfoD("Create Users with Different types of roles")
			roles := [3]backup.PxBackupRole{backup.ApplicationOwner, backup.InfrastructureOwner, backup.ApplicationUser}
			for i := 1; i <= numOfUsers/3; i++ {
				for _, role := range roles {
					userName := createUsers(1)[0]
					err := backup.AddRoleToUser(userName, role, fmt.Sprintf("Adding %v role to %s", role, userName))
					log.FailOnError(err, "Failed to add role for user - %s", userName)
					userNames = append(userNames, userName)
					userUID, err := backup.FetchIDOfUser(userName)
					log.FailOnError(err, "Failed to fetch uid for - %s", userName)
					userIdMap[userName] = userUID
				}
			}
		})
		Step(fmt.Sprintf("Adding Credentials and Backup Location from px-admin user"), func() {
			log.InfoD(fmt.Sprintf("Creating cloud credentials and backup location from px-admin user"))
			for _, provider := range providers {
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-admin ctx")
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				credName = fmt.Sprintf("autogenerated-cred-%v", time.Now().Unix())
				err = CreateCloudCredential(provider, credName, cloudCredUID, orgID, ctx)
				log.FailOnError(err, "Failed to create cloud credential - %s", err)
				backupLocationName = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
				err = CreateBackupLocation(provider, backupLocationName, backupLocationUID, credName, cloudCredUID, getGlobalBucketName(provider), orgID, "")
				log.FailOnError(err, "Failed to add backup location %s using provider %s for px-admin user", backupLocationName, provider)
				backupLocationMap[backupLocationUID] = backupLocationName
			}
		})
		Step(fmt.Sprintf("Create schedule policy from px-admin"), func() {
			log.InfoD("Creating a schedule policy from px-admin")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			periodicSchedulePolicyName = fmt.Sprintf("%s-%v", "periodic", time.Now().Unix())
			periodicSchedulePolicyUid = uuid.New()
			periodicSchedulePolicyInterval = 15
			err = CreateBackupScheduleIntervalPolicy(5, periodicSchedulePolicyInterval, 5, periodicSchedulePolicyName, periodicSchedulePolicyUid, orgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of periodic schedule policy of interval [%v] minutes named [%s]", periodicSchedulePolicyInterval, periodicSchedulePolicyName))
			periodicSchedulePolicyUid, err = Inst().Backup.GetSchedulePolicyUid(orgID, ctx, periodicSchedulePolicyName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching uid of periodic schedule policy named [%s]", periodicSchedulePolicyName))
		})

		Step(fmt.Sprintf("Create pre and post exec rules for applications from px-admin"), func() {
			log.InfoD("Create pre and post exec rules for applications from px-admin")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			preRuleName, postRuleName, err = CreateRuleForBackupWithMultipleApplications(orgID, Inst().AppList, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of pre and post exec rules for applications from px-admin"))
			if preRuleName != "" {
				preRuleUid, err = Inst().Backup.GetRuleUid(orgID, ctx, preRuleName)
				log.FailOnError(err, "Fetching pre backup rule [%s] uid", preRuleName)
				log.Infof("Pre backup rule [%s] uid: [%s]", preRuleName, preRuleUid)
			}
			if postRuleName != "" {
				postRuleUid, err = Inst().Backup.GetRuleUid(orgID, ctx, postRuleName)
				log.FailOnError(err, "Fetching post backup rule [%s] uid", postRuleName)
				log.Infof("Post backup rule [%s] uid: [%s]", postRuleName, postRuleUid)
			}
		})

		Step("Update ownership for RBAC resource with non-admin users from px-admin", func() {
			log.InfoD("update ownership for RBAC resource with non-admin users from px-admin")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "fetching px-admin ctx")
			log.InfoD("Update BackupLocation - %s ownership for users - [%v]", backupLocationName, userNames)
			err = UpdateBackupLocationOwnership(backupLocationName, backupLocationUID, userNames, nil, Read, Invalid, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying updation of owbership for backuplocation - %s", backupLocationName))

			log.InfoD("Update SchedulePolicy - %s ownership for users - [%v]", periodicSchedulePolicyName, userNames)
			err = UpdateSchedulePolicyOwnership(periodicSchedulePolicyName, periodicSchedulePolicyUid, userNames, nil, Read, Invalid, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying updation of ownership for schedulepolicy - %s", periodicSchedulePolicyName))

			log.InfoD("Update Application Rules ownership for users - [%v]", userNames)
			if preRuleName != "" {
				err = UpdateRuleOwnership(preRuleName, preRuleUid, userNames, nil, Read, Invalid, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying updation of ownership for pre-rule of application"))
			}
			if postRuleName != "" {
				err = UpdateRuleOwnership(postRuleName, postRuleUid, userNames, nil, Read, Invalid, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying updation of ownership for post-rule of application"))
			}
		})

		Step(fmt.Sprintf("Register source and destination cluster for backup on non-admin user"), func() {
			log.InfoD("Register source and destination cluster for backup on non-admin user")
			for _, nonAdminUserName := range userNames {
				log.InfoD("Registering Source and Destination clusters as user : %s and verifying the status", nonAdminUserName)
				nonAdminCtx, err := backup.GetNonAdminCtx(nonAdminUserName, commonPassword)
				log.FailOnError(err, "Fetching non admin ctx")
				err = CreateApplicationClusters(orgID, "", "", nonAdminCtx)
				log.FailOnError(err, "Failed creating source and destination cluster for user : %s", nonAdminUserName)
				clusterStatus, err := Inst().Backup.GetClusterStatus(orgID, SourceClusterName, nonAdminCtx)
				log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
				dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
				srcClusterUid, err = Inst().Backup.GetClusterUID(nonAdminCtx, orgID, SourceClusterName)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
				destClusterUid, err := Inst().Backup.GetClusterUID(nonAdminCtx, orgID, destinationClusterName)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", destinationClusterName))
				clusterInfo := make(map[string]string)
				clusterInfo[SourceClusterName] = srcClusterUid
				clusterInfo[destinationClusterName] = destClusterUid
				clusterUidMap[nonAdminUserName] = clusterInfo
			}
		})

		Step(fmt.Sprintf("Taking schedule backup of applications as non-admin user with and without rules"), func() {
			log.InfoD("Taking schedule backup of applications as non-admin user with and without rules")
			for _, nonAdminUserName := range userNames {
				time.Sleep(timeBetweenConsecutiveBackups)
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					nonAdminCtx, err := backup.GetNonAdminCtx(nonAdminUserName, commonPassword)
					log.FailOnError(err, "Fetching non admin ctx")
					log.InfoD("Taking schedule backup of single namespace as user : %s without-rules", nonAdminUserName)
					scheduleName := fmt.Sprintf("%s-schedule-single-ns-%s-without-rules-%s", BackupNamePrefix, nonAdminUserName, RandomString(4))
					scheduleNameMap[nonAdminUserName] = scheduleName
					labelSelectors := make(map[string]string, 0)
					log.InfoD("Creating a schedule backup of namespace [%s] without pre and post exec rules", bkpNamespaces[0])
					appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{bkpNamespaces[0]})
					scheduleBackupName, err := CreateScheduleBackupWithValidation(nonAdminCtx, scheduleNameMap[nonAdminUserName], SourceClusterName, backupLocationName, backupLocationUID, appContextsToBackup,
						labelSelectors, orgID, "", "", "", "", periodicSchedulePolicyName, periodicSchedulePolicyUid)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of backup [%s]", scheduleBackupName))
					singleNamespaceBackupsMap[nonAdminUserName] = SafeAppend(&mutex, singleNamespaceBackupsMap[nonAdminUserName], scheduleBackupName).([]string)
					userBackupNamesMap[nonAdminUserName] = SafeAppend(&mutex, userBackupNamesMap[nonAdminUserName], scheduleBackupName).([]string)
					err = suspendBackupSchedule(scheduleNameMap[nonAdminUserName], periodicSchedulePolicyName, orgID, nonAdminCtx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Suspending Backup Schedule [%s] for user [%s]", scheduleNameMap[nonAdminUserName], nonAdminUserName))
				}(nonAdminUserName)
			}
			wg.Wait()

			for _, nonAdminUserName := range userNames {
				time.Sleep(timeBetweenConsecutiveBackups)
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					nonAdminCtx, err := backup.GetNonAdminCtx(nonAdminUserName, commonPassword)
					log.FailOnError(err, "Fetching non admin ctx")
					log.InfoD("Taking schedule backup of mutiple namespace as user : %s with-rules", nonAdminUserName)
					scheduleName := fmt.Sprintf("%s-schedule-multiple-ns-%s-with-rules-%s", BackupNamePrefix, nonAdminUserName, RandomString(4))
					scheduleNameMap[nonAdminUserName] = scheduleName
					labelSelectors := make(map[string]string, 0)
					log.InfoD("Creating a schedule backup of namespace [%s] with pre and post exec rules", bkpNamespaces)
					appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
					scheduleBackupName, err := CreateScheduleBackupWithValidation(nonAdminCtx, scheduleNameMap[nonAdminUserName], SourceClusterName, backupLocationName, backupLocationUID, appContextsToBackup,
						labelSelectors, orgID, preRuleName, preRuleUid, postRuleName, postRuleUid, periodicSchedulePolicyName, periodicSchedulePolicyUid)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of backup [%s]", scheduleBackupName))
					mutipleNamespaceBackupsMap[nonAdminUserName] = SafeAppend(&mutex, mutipleNamespaceBackupsMap[nonAdminUserName], scheduleBackupName).([]string)
					userBackupNamesMap[nonAdminUserName] = SafeAppend(&mutex, userBackupNamesMap[nonAdminUserName], scheduleBackupName).([]string)
					err = suspendBackupSchedule(scheduleNameMap[nonAdminUserName], periodicSchedulePolicyName, orgID, nonAdminCtx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Suspending Backup Schedule [%s] for user [%s]", scheduleNameMap[nonAdminUserName], nonAdminUserName))
				}(nonAdminUserName)
			}
			wg.Wait()
		})

		Step(fmt.Sprintf("Taking namespace label schedule backup of applications with and without rules from non-admin user"), func() {
			log.InfoD("Taking namespace label schedule backup of applications with and without rules from non-admin user")
			for _, nonAdminUserName := range userNames {
				time.Sleep(timeBetweenConsecutiveBackups)
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					nonAdminCtx, err := backup.GetNonAdminCtx(nonAdminUserName, commonPassword)
					log.FailOnError(err, "Fetching non admin ctx")
					log.InfoD("Taking namespace label schedule backup of applications of user : %s ", nonAdminUserName)
					scheduleName := fmt.Sprintf("%s-schedule-nslabel-%s-with-rules-%s", BackupNamePrefix, nonAdminUserName, RandomString(4))
					scheduleNameMap[nonAdminUserName] = scheduleName
					labelSelectors := make(map[string]string, 0)
					log.InfoD("Creating a backup of namespaces [%v] with pre and post exec rules", bkpNamespaces)
					appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
					scheduleBackupName, err := CreateScheduleBackupWithNamespaceLabelWithValidation(nonAdminCtx, scheduleNameMap[nonAdminUserName], SourceClusterName, backupLocationName, backupLocationUID, appContextsToBackup,
						labelSelectors, orgID, preRuleName, preRuleUid, postRuleName, postRuleUid, namespaceLabel, periodicSchedulePolicyName, periodicSchedulePolicyUid)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of backup [%s]", scheduleBackupName))
					mutipleNamespaceLabelBackupsMap[nonAdminUserName] = SafeAppend(&mutex, mutipleNamespaceLabelBackupsMap[nonAdminUserName], scheduleBackupName).([]string)
					userBackupNamesMap[nonAdminUserName] = SafeAppend(&mutex, userBackupNamesMap[nonAdminUserName], scheduleBackupName).([]string)
					err = suspendBackupSchedule(scheduleNameMap[nonAdminUserName], periodicSchedulePolicyName, orgID, nonAdminCtx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Suspending Backup Schedule [%s] for user [%s]", scheduleNameMap[nonAdminUserName], nonAdminUserName))
				}(nonAdminUserName)
			}
			wg.Wait()
		})
		Step("Restore single namespace backups with different configs", func() {
			log.InfoD("Restore single namespace backups with different configs")
			for _, nonAdminUserName := range userNames {
				time.Sleep(timeBetweenConsecutiveBackups)
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					nonAdminCtx, err := backup.GetNonAdminCtx(nonAdminUserName, commonPassword)
					dash.VerifyFatal(err, nil, "Fetching px-central-admin ctx")
					restoreConfigs := []struct {
						namePrefix          string
						namespaceMapping    map[string]string
						storageClassMapping map[string]string
						replacePolicy       ReplacePolicy_Type
					}{
						{
							"test-restore-single-ns",
							make(map[string]string, 0),
							make(map[string]string, 0),
							ReplacePolicy_Retain,
						},
						{
							"test-custom-restore-single-ns",
							map[string]string{bkpNamespaces[0]: "custom-" + bkpNamespaces[0]},
							make(map[string]string, 0),
							ReplacePolicy_Retain,
						},
					}
					for _, config := range restoreConfigs {
						restoreName := fmt.Sprintf("%s-%s-%s", nonAdminUserName, config.namePrefix, RandomString(4))
						restoreNameMap[nonAdminUserName] = restoreName
						log.InfoD("Restoring single namespace backup [%s] in cluster [%s] with restore [%s] and namespace mapping %v for user [%s]", singleNamespaceBackupsMap[nonAdminUserName][0], destinationClusterName, restoreName, config.namespaceMapping, nonAdminUserName)
						err = CreateRestore(restoreNameMap[nonAdminUserName], singleNamespaceBackupsMap[nonAdminUserName][0], config.namespaceMapping, destinationClusterName, orgID, nonAdminCtx, config.storageClassMapping)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying restoration [%s] of single namespace backup [%s] in cluster [%s] by user [%s]", restoreNameMap[nonAdminUserName], singleNamespaceBackupsMap[nonAdminUserName][0], destinationClusterName, nonAdminUserName))
					}
				}(nonAdminUserName)
			}
			wg.Wait()
		})
		Step("Restore a mutiple namespace backups", func() {
			log.InfoD("Restore a mutiple namespace backups")
			for _, nonAdminUserName := range userNames {
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					nonAdminCtx, err := backup.GetNonAdminCtx(nonAdminUserName, commonPassword)
					dash.VerifyFatal(err, nil, "Fetching px-central-admin ctx")
					restoreName := fmt.Sprintf("%s-mutiple-ns-restore-%s", nonAdminUserName, RandomString(4))
					restoreNameMap[nonAdminUserName] = restoreName
					log.InfoD("Restoring mutiple namespace backup [%s] in cluster [%s] with restore name [%s] for user [%s] ", mutipleNamespaceBackupsMap[nonAdminUserName][0], destinationClusterName, restoreNameMap[nonAdminUserName], nonAdminUserName)
					err = CreateRestore(restoreNameMap[nonAdminUserName], mutipleNamespaceBackupsMap[nonAdminUserName][0], namespaceMapping, destinationClusterName, orgID, nonAdminCtx, storageClassMapping)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying restoration [%s] of mutiple namespace schedule backup [%s] in cluster [%s] for user [%s]", restoreNameMap[nonAdminUserName], mutipleNamespaceBackupsMap[nonAdminUserName][0], destinationClusterName, nonAdminUserName))
				}(nonAdminUserName)
			}
			wg.Wait()
		})
		Step("Restore a namespace label backups", func() {
			log.InfoD("Restore a namespace label backups")
			for _, nonAdminUserName := range userNames {
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					nonAdminCtx, err := backup.GetNonAdminCtx(nonAdminUserName, commonPassword)
					dash.VerifyFatal(err, nil, "Fetching px-central-admin ctx")
					restoreName := fmt.Sprintf("%s-mutiple-ns-label-restore-%s", nonAdminUserName, RandomString(4))
					restoreNameMap[nonAdminUserName] = restoreName
					log.InfoD("Restoring mutiple namespace backup [%s] in cluster [%s] with restore name [%s] ", mutipleNamespaceLabelBackupsMap[nonAdminUserName][0], destinationClusterName, restoreNameMap[nonAdminUserName])
					err = CreateRestore(restoreNameMap[nonAdminUserName], mutipleNamespaceLabelBackupsMap[nonAdminUserName][0], namespaceMapping, destinationClusterName, orgID, nonAdminCtx, storageClassMapping)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying restoration [%s] of mutiple namespace schedule backup [%s] in cluster [%s] for user [%s]", restoreNameMap[nonAdminUserName], mutipleNamespaceLabelBackupsMap[nonAdminUserName][0], destinationClusterName, nonAdminUserName))
				}(nonAdminUserName)
			}
			wg.Wait()
		})
		log.InfoD("Deletion of all backups,restores,schedules,clusters of users when user is present in keycloak ")
		Step(fmt.Sprintf("Listing and Deletion of backup of non-admin user from px-admin user"), func() {
			log.InfoD("Listing and Deletion of backup of non-admin user from px-admin user")
			adminCtx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			for _, nonAdminUserName := range userNames[:3] {
				log.InfoD(fmt.Sprintf("Verifying listing backups of non-admin user [%s] from px-admin user", nonAdminUserName))
				userBackupNames, err := GetAllBackupNamesByOwnerID(userIdMap[nonAdminUserName], orgID, adminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of fetching backups [%v] of user [%s] from px-admin user", userBackupNames, nonAdminUserName))
				userBackupNamesMapFromAdmin[nonAdminUserName] = userBackupNames
				log.Infof(fmt.Sprintf("The list of user [%s] backups from px-admin %v", nonAdminUserName, userBackupNamesMapFromAdmin[nonAdminUserName]))
			}
			for _, nonAdminUserName := range userNames[:3] {
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					for _, backupName := range userBackupNamesMapFromAdmin[nonAdminUserName] {
						wg.Add(1)
						go func(backupName string) {
							defer GinkgoRecover()
							defer wg.Done()
							log.InfoD(fmt.Sprintf("Verifying deletion backup [%s] of non-admin user [%s] from px-admin user", backupName, nonAdminUserName))
							backupUID, _ := backupDriver.GetBackupUID(adminCtx, backupName, orgID)
							_, err = DeleteBackup(backupName, backupUID, orgID, adminCtx)
							log.FailOnError(err, "Failed to delete backup - %s", backupName)
							err = DeleteBackupAndWait(backupName, adminCtx)
							log.FailOnError(err, fmt.Sprintf("waiting for backup [%s] deletion", backupName))
						}(backupName)
					}
				}(nonAdminUserName)
			}
			wg.Wait()
		})
		Step(fmt.Sprintf("Deletion of backup schedules of non-admin user from px-admin user"), func() {
			log.InfoD("Deletion of backup schedules of non-admin user from px-admin user")
			adminCtx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			for _, nonAdminUserName := range userNames[:3] {
				log.InfoD(fmt.Sprintf("Verifying listing of backup schedule of non-admin user [%s] from px-admin user", nonAdminUserName))
				userBackupSchedules, err := GetAllBackupSchedulesForUser(nonAdminUserName, commonPassword)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of fetching backup schedules of user from px-admin user"))
				userBackupSchedulesMap[nonAdminUserName] = userBackupSchedules
				log.Infof(fmt.Sprintf("the list of user [%s ]backup schedules from px-admin user %v", nonAdminUserName, userBackupSchedulesMap[nonAdminUserName]))
				for _, backupName := range userBackupNamesMap[nonAdminUserName] {
					if !IsPresent(userBackupNamesMapFromAdmin[nonAdminUserName], backupName) {
						err := fmt.Errorf("backup %s is not listed in backup names %s", backupName, userBackupNamesMapFromAdmin[nonAdminUserName])
						log.FailOnError(fmt.Errorf(""), err.Error())
					}
				}
			}
			for _, nonAdminUserName := range userNames[:3] {
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					for _, backupScheduleName := range userBackupSchedulesMap[nonAdminUserName] {
						log.InfoD(fmt.Sprintf("Verifying deletion of backup schedule [%s] of non-admin user [%s] from px-admin user", backupScheduleName, nonAdminUserName))
						backupScheduleUid, err := GetScheduleUID(backupScheduleName, orgID, adminCtx)
						log.FailOnError(err, fmt.Sprintf("Fetching schedule uid for shedule [%s]", backupScheduleName))
						err = DeleteScheduleWithUIDAndWait(backupScheduleName, backupScheduleUid, SourceClusterName, clusterUidMap[nonAdminUserName][SourceClusterName], orgID, adminCtx)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of deleting backup scheudle [%s] of user [%s] from px-admin user", backupScheduleName, nonAdminUserName))
					}
				}(nonAdminUserName)
			}
			wg.Wait()
		})
		Step(fmt.Sprintf("Deletion of restores of non-admin user from px-admin user"), func() {
			log.InfoD("Deletion of restores of non-admin user from px-admin user")
			adminCtx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			for _, nonAdminUserName := range userNames[:3] {
				log.InfoD(fmt.Sprintf("Verifying  listing of restores of non-admin user [%s] from px-admin user", nonAdminUserName))
				userRestores, err := GetAllRestoresForUser(nonAdminUserName, commonPassword)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of fetching restores of user ffrom px-admin user"))
				userRestoresMap[nonAdminUserName] = userRestores
				log.Infof(fmt.Sprintf("the list of user [%s] restores from px-admin user %v", nonAdminUserName, userRestoresMap[nonAdminUserName]))
			}
			for _, nonAdminUserName := range userNames[:3] {
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					for _, restoreName := range userRestoresMap[nonAdminUserName] {
						log.InfoD(fmt.Sprintf("Verifying  Deletion of restores [%s] of non-admin user [%s] from px-admin user", restoreName, nonAdminUserName))
						restoreUid, _ := Inst().Backup.GetRestoreUID(adminCtx, restoreName, orgID)
						err := DeleteRestoreWithUID(restoreName, restoreUid, orgID, adminCtx)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of deleting restore [%s] of user [%s] from px-admin user", restoreName, nonAdminUserName))
					}
				}(nonAdminUserName)
			}
			wg.Wait()
		})
		Step(fmt.Sprintf("Deletion of clusters of non-admin user from px-admin user"), func() {
			log.InfoD("Deletion of clusters of non-admin user from px-admin user")
			adminCtx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			for _, nonAdminUserName := range userNames[:3] {
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					err = DeleteClusterWithUID(SourceClusterName, clusterUidMap[nonAdminUserName][SourceClusterName], orgID, adminCtx, true)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting cluster %s", SourceClusterName))
					err = DeleteClusterWithUID(destinationClusterName, clusterUidMap[nonAdminUserName][destinationClusterName], orgID, adminCtx, true)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting cluster %s", destinationClusterName))
				}(nonAdminUserName)
			}
			wg.Wait()
		})

		log.InfoD("Deletion of all backups,restores,schedules,clusters of users when user is deleted from keycloak ")
		Step(fmt.Sprintf("Verifying deletion of non-admin user from keycloak"), func() {
			log.InfoD("Verifying deletion of non-admin user from keycloak")
			for _, nonAdminUserName := range userNames[3:6] {
				log.InfoD(fmt.Sprintf("Verifying deletion of user  [%s] from keycloak", nonAdminUserName))
				log.Infof(fmt.Sprintf("Fetching user [%s] backup schedules and restore before user deletion ", nonAdminUserName))
				userBackupSchedules, err := GetAllBackupSchedulesForUser(nonAdminUserName, commonPassword)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of fetching backup schedules of user from px-admin user"))
				userBackupSchedulesMap[nonAdminUserName] = userBackupSchedules
				log.Infof(fmt.Sprintf("the list of user [%s] backup schedules [%s] ", nonAdminUserName, userBackupSchedulesMap[nonAdminUserName]))
				userRestores, err := GetAllRestoresForUser(nonAdminUserName, commonPassword)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of fetching restores of user ffrom px-admin user"))
				userRestoresMap[nonAdminUserName] = userRestores
				log.Infof(fmt.Sprintf("the list of user [%s] restores [%s] ", nonAdminUserName, userRestoresMap[nonAdminUserName]))
				err = backup.DeleteUser(nonAdminUserName)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying the deletion of the user [%s]", nonAdminUserName))
			}
		})
		Step(fmt.Sprintf("Listing and deletion of backup of deleted non-admin user from px-admin user"), func() {
			log.InfoD("Listing and deletion of backup of deleted non-admin user from px-admin user")
			adminCtx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			for _, nonAdminUserName := range userNames[3:6] {
				log.InfoD(fmt.Sprintf("Verifying listing backups of non-admin user [%s] from px-admin user", nonAdminUserName))
				userBackupNames, err := GetAllBackupNamesByOwnerID(userIdMap[nonAdminUserName], orgID, adminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of fetching backups [%v] of user [%s] from px-admin user", userBackupNames, nonAdminUserName))
				userBackupNamesMapFromAdmin[nonAdminUserName] = userBackupNames
				log.Infof(fmt.Sprintf("The list of user [%s] backups from px-admin %v", nonAdminUserName, userBackupNamesMapFromAdmin[nonAdminUserName]))
				for _, backupName := range userBackupNamesMap[nonAdminUserName] {
					if !IsPresent(userBackupNamesMapFromAdmin[nonAdminUserName], backupName) {
						err := fmt.Errorf("backup %s is not listed in backup names %s", backupName, userBackupNamesMapFromAdmin[nonAdminUserName])
						log.FailOnError(fmt.Errorf(""), err.Error())
					}
				}
			}
			for _, nonAdminUserName := range userNames[3:6] {
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					for _, backupName := range userBackupNamesMapFromAdmin[nonAdminUserName] {
						wg.Add(1)
						go func(backupName string) {
							defer GinkgoRecover()
							defer wg.Done()
							log.InfoD(fmt.Sprintf("Verifying deletion backup [%s] of non-admin user [%s] from px-admin user", backupName, nonAdminUserName))
							backupUID, _ := backupDriver.GetBackupUID(adminCtx, backupName, orgID)
							_, err = DeleteBackup(backupName, backupUID, orgID, adminCtx)
							log.FailOnError(err, "Failed to delete backup - %s", backupName)
							err = DeleteBackupAndWait(backupName, adminCtx)
							log.FailOnError(err, fmt.Sprintf("waiting for backup [%s] deletion", backupName))
						}(backupName)
					}
				}(nonAdminUserName)
			}
			wg.Wait()
		})
		Step(fmt.Sprintf("Verifying  deletion of backup schedule of deleted non-admin user from px-admin user"), func() {
			log.InfoD("Verifying  deletion of backup schedule of deleted non-admin user from px-admin user")
			adminCtx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			for _, nonAdminUserName := range userNames[3:6] {
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					for _, backupScheduleName := range userBackupSchedulesMap[nonAdminUserName] {
						log.InfoD(fmt.Sprintf("Verifying deletion of backup schedule [%s] of non-admin user [%s] from px-admin user", backupScheduleName, nonAdminUserName))
						backupScheduleUid, err := GetScheduleUID(backupScheduleName, orgID, adminCtx)
						log.FailOnError(err, fmt.Sprintf("Fetching schedule uid for shedule [%s]", backupScheduleName))
						err = DeleteScheduleWithUIDAndWait(backupScheduleName, backupScheduleUid, SourceClusterName, clusterUidMap[nonAdminUserName][SourceClusterName], orgID, adminCtx)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of deleting backup scheudle [%s] of user [%s] from px-admin user", backupScheduleName, nonAdminUserName))
					}
				}(nonAdminUserName)
			}
			wg.Wait()
		})
		Step(fmt.Sprintf("Verifying  deletion of restore of deleted non-admin user from px-admin user"), func() {
			log.InfoD("Verifying  deletion of restore of deleted non-admin user from px-admin user")
			adminCtx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			for _, nonAdminUserName := range userNames[3:6] {
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					for _, restoreName := range userRestoresMap[nonAdminUserName] {
						log.InfoD(fmt.Sprintf("Verifying  Deletion of restores [%s] of non-admin user [%s] from px-admin user", restoreName, nonAdminUserName))
						restoreUid, _ := Inst().Backup.GetRestoreUID(adminCtx, restoreName, orgID)
						err := DeleteRestoreWithUID(restoreName, restoreUid, orgID, adminCtx)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of deleting restore [%s] of user [%s] from px-admin user", restoreName, nonAdminUserName))
					}
				}(nonAdminUserName)
			}
			wg.Wait()
		})
		Step(fmt.Sprintf("Verifying  deletion of clusters of deleted non-admin user from px-admin user"), func() {
			log.InfoD("Verifying  deletion of clusters of deleted non-admin user from px-admin user")
			adminCtx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			for _, nonAdminUserName := range userNames[3:6] {
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					log.InfoD(fmt.Sprintf("Verifying  deletion of clusters of deleted non-admin user [%s] from px-admin user", nonAdminUserName))
					err = DeleteClusterWithUID(SourceClusterName, clusterUidMap[nonAdminUserName][SourceClusterName], orgID, adminCtx, true)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting cluster %s", SourceClusterName))
					err = DeleteClusterWithUID(destinationClusterName, clusterUidMap[nonAdminUserName][destinationClusterName], orgID, adminCtx, true)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting cluster %s", destinationClusterName))
				}(nonAdminUserName)
			}
			wg.Wait()
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
		log.Infof("Deleting backup schedule policy")
		err = Inst().Backup.DeleteBackupSchedulePolicy(orgID, []string{periodicSchedulePolicyName})
		dash.VerifySafely(err, nil, fmt.Sprintf("Deleting backup schedule policy %s ", []string{periodicSchedulePolicyName}))
		log.Infof("Deleting pre and post exec rules")
		if preRuleName != "" {
			err = Inst().Backup.DeleteRuleForBackup(orgID, preRuleName)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting pre exec rule %s ", preRuleName))
		}
		if postRuleName != "" {
			err = Inst().Backup.DeleteRuleForBackup(orgID, postRuleName)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting post exec rule %s ", postRuleName))
		}
		CleanupCloudSettingsAndClusters(backupLocationMap, credName, cloudCredUID, ctx)
	})
})
