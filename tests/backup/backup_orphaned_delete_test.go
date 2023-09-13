package tests

import (
	"fmt"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/torpedo/drivers"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

// DeleteBackupOfUserNonSharedRBAC delete backups created by user from admin with non-shared RBAC resources from  px-admin.
var _ = Describe("{DeleteBackupOfUserNonSharedRBAC}", func() {
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/87561
	var (
		userNames                      []string
		periodicSchedulePolicyName     string
		periodicSchedulePolicyUid      string
		periodicSchedulePolicyInterval int64
		scheduledAppContexts           []*scheduler.Context
		backupLocationUID              string
		bkpNamespaces                  []string
		adminCredName                  string
		adminCloudCredUID              string
		infraUserCredName              string
		infraUserCloudCredUID          string
		srcClusterUid                  string
		backupLocationName             string
		preRuleName                    string
		postRuleName                   string
		preRuleUid                     string
		postRuleUid                    string
		backupName                     string
		backupLocationMap              map[string]string
		nsLabels                       map[string]string
		namespaceLabel                 string
		appAdminUserNames              []string
		infraAdminUserNames            []string
		mutex                          sync.Mutex
		wg                             sync.WaitGroup
	)
	bkpNamespaces = make([]string, 0)
	userNames = make([]string, 0)
	numOfNS := 2
	numOfUsers := 4
	timeBetweenConsecutiveBackups := 10 * time.Second
	backupLocationMap = make(map[string]string)
	userIdMap := make(map[string]string)
	userCredNameMap := make(map[string]string)
	userCloudCredUIDMap := make(map[string]string)
	namespaceMapping := make(map[string]string)
	storageClassMapping := make(map[string]string)
	clusterUidMap := make(map[string]map[string]string)
	backupLocationNameMap := make(map[string]string)
	backupLocationUidMap := make(map[string]string)
	periodicSchedulePolicyNameMap := make(map[string]string)
	periodicSchedulePolicyUidMap := make(map[string]string)
	preRuleNameMap := make(map[string]string)
	preRuleUidMap := make(map[string]string)
	postRuleNameMap := make(map[string]string)
	postRuleUidMap := make(map[string]string)
	singleNamespaceBackupsMap := make(map[string][]string)
	mutipleNamespaceBackupsMap := make(map[string][]string)
	mutipleNamespaceLabelBackupsMap := make(map[string][]string)
	backupNameMap := make(map[string]string)
	scheduleNameMap := make(map[string]string)
	restoreNameMap := make(map[string]string)
	userBackupNamesMap := make(map[string][]string)
	userBackupSchedulesMap := make(map[string][]string)
	userRestoresMap := make(map[string][]string)
	backupDriver := Inst().Backup

	JustBeforeEach(func() {
		StartTorpedoTest("DeleteBackupOfUserNonSharedRBAC",
			"Delete backups,restores,schedules,clusters created by non-admin user with non-shared RBAC resources from px-admin ", nil, 87561)
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
	It("Delete backups by user from admin with non-shared RBAC objects", func() {
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
			roles := [2]backup.PxBackupRole{backup.ApplicationOwner, backup.InfrastructureOwner}
			for i := 1; i <= numOfUsers/2; i++ {
				for _, role := range roles {
					userName := createUsers(1)[0]
					err := backup.AddRoleToUser(userName, role, fmt.Sprintf("Adding %v role to %s", role, userName))
					log.FailOnError(err, "Failed to add role for user - %s", userName)
					if role == backup.ApplicationOwner {
						appAdminUserNames = append(appAdminUserNames, userName)
					} else {
						infraAdminUserNames = append(infraAdminUserNames, userName)
					}
					userNames = append(userNames, userName)
					userUID, err := backup.FetchIDOfUser(userName)
					log.FailOnError(err, "Failed to fetch uid for - %s", userName)
					userIdMap[userName] = userUID
				}
			}
		})

		Step(fmt.Sprintf("Adding cloud account and backup location from px-admin "), func() {
			log.InfoD(fmt.Sprintf("Adding cloud account and backup location from px-admin"))
			adminCtx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			for _, provider := range providers {
				adminCloudCredUID = uuid.New()
				adminCredName = fmt.Sprintf("%v-cred-%v", provider, RandomString(5))
				err := CreateCloudCredential(provider, adminCredName, adminCloudCredUID, orgID, adminCtx)
				log.FailOnError(err, "Failed to create cloud account for backup location from px-admin user  - %s", err)
				if provider != drivers.ProviderNfs {
					log.InfoD(fmt.Sprintf("Update ownership for cloud account from px-admin to users with role app.admin"))
					log.Infof("Update CloudAccount - %s ownership for users - [%v]", adminCredName, appAdminUserNames)
					err = UpdateCloudCredentialOwnership(adminCredName, adminCloudCredUID, appAdminUserNames, nil, Read, Invalid, adminCtx, orgID)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying updation of owbership for CloudCredential- %s", adminCredName))
					for _, appAdminUserName := range appAdminUserNames {
						userCredNameMap[appAdminUserName] = adminCredName
						userCloudCredUIDMap[appAdminUserName] = adminCloudCredUID
					}
				}
			}
		})

		for _, infraAdminUserName := range infraAdminUserNames {
			Step(fmt.Sprintf("Adding cloud account for backup location for infra-admin user [%s] from px-admin ", infraAdminUserName), func() {
				log.InfoD(fmt.Sprintf("Adding cloud account for backup location for infra-admin user"))
				nonAdminCtx, err := backup.GetNonAdminCtx(infraAdminUserName, commonPassword)
				dash.VerifyFatal(err, nil, "Fetching px-central-admin ctx")
				for _, provider := range providers {
					infraUserCloudCredUID = uuid.New()
					infraUserCredName = fmt.Sprintf("%v-cred-%v", provider, RandomString(5))
					err = CreateCloudCredential(provider, infraUserCredName, infraUserCloudCredUID, orgID, nonAdminCtx)
					log.FailOnError(err, "Failed to create cloud account for backup location from px-admin user  - %s", err)
					userCredNameMap[infraAdminUserName] = infraUserCredName
					userCloudCredUIDMap[infraAdminUserName] = infraUserCloudCredUID
				}
			})
		}

		Step(fmt.Sprintf("Create backup location for non-admin user using the cloud cred "), func() {
			for _, nonAdminUserName := range userNames {
				log.InfoD(fmt.Sprintf("Create backup location for non-admin user [%s] using the cloud cred created ", nonAdminUserName))
				nonAdminCtx, err := backup.GetNonAdminCtx(nonAdminUserName, commonPassword)
				log.FailOnError(err, "Fetching non admin user ctx")
				for _, provider := range providers {
					backupLocationUID = uuid.New()
					backupLocationName = fmt.Sprintf("%s-location-%s", provider, RandomString(5))
					backupLocationNameMap[nonAdminUserName] = backupLocationName
					backupLocationUidMap[nonAdminUserName] = backupLocationUID
					userBucketName := fmt.Sprintf("%s-%s", getGlobalBucketName(provider), RandomString(5))
					err = CreateBackupLocationWithContext(provider, backupLocationNameMap[nonAdminUserName], backupLocationUidMap[nonAdminUserName], userCredNameMap[nonAdminUserName], userCloudCredUIDMap[nonAdminUserName], userBucketName, orgID, "", "", nonAdminCtx)
					log.FailOnError(err, "Failed to add backup location %s using provider %s for non-admin user %s", backupLocationNameMap[nonAdminUserName], provider, nonAdminUserName)
					backupLocationMap[backupLocationUID] = backupLocationName
				}
			}
		})

		Step(fmt.Sprintf("Create schedule policy from non-admin user"), func() {
			for _, nonAdminUserName := range userNames {
				log.InfoD(fmt.Sprintf("Creating a schedule policy from non-admin [%s] user", nonAdminUserName))
				nonAdminCtx, err := backup.GetNonAdminCtx(nonAdminUserName, commonPassword)
				log.FailOnError(err, "Fetching non admin user ctx")
				periodicSchedulePolicyName = fmt.Sprintf("%s-%v-%s", "periodic", time.Now().Unix(), nonAdminUserName)
				periodicSchedulePolicyUid = uuid.New()
				periodicSchedulePolicyInterval = 15
				periodicSchedulePolicyNameMap[nonAdminUserName] = periodicSchedulePolicyName
				periodicSchedulePolicyUidMap[nonAdminUserName] = periodicSchedulePolicyUid
				err = CreateBackupScheduleIntervalPolicy(5, periodicSchedulePolicyInterval, 5, periodicSchedulePolicyNameMap[nonAdminUserName], periodicSchedulePolicyUidMap[nonAdminUserName], orgID, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of periodic schedule policy of interval [%v] minutes named [%s] for user [%s]", periodicSchedulePolicyInterval, periodicSchedulePolicyNameMap[nonAdminUserName], nonAdminUserName))
			}
		})

		Step(fmt.Sprintf("Create pre and post exec rules for applications from non-admin user "), func() {
			for _, nonAdminUserName := range userNames {
				log.InfoD(fmt.Sprintf("Create pre and post exec rules for applications from non-admin user [%s]", nonAdminUserName))
				nonAdminCtx, err := backup.GetNonAdminCtx(nonAdminUserName, commonPassword)
				log.FailOnError(err, "Fetching non admin user ctx")
				preRuleName, postRuleName, err = CreateRuleForBackupWithMultipleApplications(orgID, Inst().AppList, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of pre and post exec rules for applications from non-admin user [%s]", nonAdminUserName))
				if preRuleName != "" {
					preRuleNameMap[nonAdminUserName] = preRuleName
					preRuleUid, err = Inst().Backup.GetRuleUid(orgID, nonAdminCtx, preRuleNameMap[nonAdminUserName])
					log.FailOnError(err, "Fetching pre backup rule [%s] uid", preRuleNameMap[nonAdminUserName])
					preRuleUidMap[nonAdminUserName] = preRuleUid
					log.Infof("Pre backup rule [%s] uid: [%s]", preRuleNameMap[nonAdminUserName], preRuleUidMap[nonAdminUserName])
				}
				if postRuleName != "" {
					postRuleNameMap[nonAdminUserName] = postRuleName
					postRuleUid, err = Inst().Backup.GetRuleUid(orgID, nonAdminCtx, postRuleNameMap[nonAdminUserName])
					log.FailOnError(err, "Fetching post backup rule [%s] uid", postRuleNameMap[nonAdminUserName])
					postRuleUidMap[nonAdminUserName] = postRuleUid
					log.Infof("Post backup rule [%s] uid: [%s]", postRuleNameMap[nonAdminUserName], postRuleUidMap[nonAdminUserName])
				}
			}
		})

		Step(fmt.Sprintf("Register source and destination cluster for backup on non-admin user"), func() {
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

		Step(fmt.Sprintf("Taking manual backup of applications with and without rules from non-admin user"), func() {
			for _, nonAdminUserName := range userNames {
				log.InfoD(fmt.Sprintf("Taking manual backup of applications with and without rules from non-admin user [%s]", nonAdminUserName))
				time.Sleep(timeBetweenConsecutiveBackups)
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					nonAdminCtx, err := backup.GetNonAdminCtx(nonAdminUserName, commonPassword)
					log.FailOnError(err, "Fetching non admin ctx")
					log.InfoD("Taking manual backup of single namespace as user : %s with-rules", nonAdminUserName)
					backupName = fmt.Sprintf("%s-manual-single-ns-%s-with-rules-%s", BackupNamePrefix, nonAdminUserName, RandomString(4))
					backupNameMap[nonAdminUserName] = backupName
					labelSelectors := make(map[string]string, 0)
					log.InfoD("Creating a backup of namespace [%s] with pre and post exec rules", bkpNamespaces[0])
					appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{bkpNamespaces[0]})
					err = CreateBackupWithValidation(nonAdminCtx, backupNameMap[nonAdminUserName], SourceClusterName, backupLocationNameMap[nonAdminUserName], backupLocationUidMap[nonAdminUserName], appContextsToBackup,
						labelSelectors, orgID, clusterUidMap[nonAdminUserName][SourceClusterName], preRuleNameMap[nonAdminUserName], preRuleUidMap[nonAdminUserName], postRuleNameMap[nonAdminUserName], postRuleUidMap[nonAdminUserName])
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of backup [%s]", backupNameMap[nonAdminUserName]))
					singleNamespaceBackupsMap[nonAdminUserName] = SafeAppend(&mutex, singleNamespaceBackupsMap[nonAdminUserName], backupNameMap[nonAdminUserName]).([]string)
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
					log.InfoD("Taking manual backup of mutiple namespace as user : %s without-rules", nonAdminUserName)
					backupName = fmt.Sprintf("%s-manual-multiple-ns-%s-without-rules-%s", BackupNamePrefix, nonAdminUserName, RandomString(4))
					backupNameMap[nonAdminUserName] = backupName
					labelSelectors := make(map[string]string, 0)
					log.InfoD("Creating a backup of namespaces [%v] without pre and post exec rules", bkpNamespaces)
					appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
					err = CreateBackupWithValidation(nonAdminCtx, backupNameMap[nonAdminUserName], SourceClusterName, backupLocationNameMap[nonAdminUserName], backupLocationUidMap[nonAdminUserName], appContextsToBackup,
						labelSelectors, orgID, clusterUidMap[nonAdminUserName][SourceClusterName], "", "", "", "")
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of backup [%s]", backupNameMap[nonAdminUserName]))
					mutipleNamespaceBackupsMap[nonAdminUserName] = SafeAppend(&mutex, mutipleNamespaceBackupsMap[nonAdminUserName], backupNameMap[nonAdminUserName]).([]string)
				}(nonAdminUserName)
			}
			wg.Wait()

		})

		Step(fmt.Sprintf("Taking schedule backup of applications as non-admin user with and without rules"), func() {
			for _, nonAdminUserName := range userNames {
				log.InfoD(fmt.Sprintf("Taking schedule backup of applications with and without rules from non-admin user [%s]", nonAdminUserName))
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
					scheduleBackupName, err := CreateScheduleBackupWithValidation(nonAdminCtx, scheduleNameMap[nonAdminUserName], SourceClusterName, backupLocationNameMap[nonAdminUserName], backupLocationUidMap[nonAdminUserName], appContextsToBackup,
						labelSelectors, orgID, "", "", "", "", periodicSchedulePolicyNameMap[nonAdminUserName], periodicSchedulePolicyUidMap[nonAdminUserName])
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of backup [%s]", scheduleBackupName))
					singleNamespaceBackupsMap[nonAdminUserName] = SafeAppend(&mutex, singleNamespaceBackupsMap[nonAdminUserName], scheduleBackupName).([]string)
					err = suspendBackupSchedule(scheduleNameMap[nonAdminUserName], periodicSchedulePolicyNameMap[nonAdminUserName], orgID, nonAdminCtx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Suspending Backup Schedule [%s] for user [%s]", scheduleNameMap[nonAdminUserName], nonAdminUserName))
				}(nonAdminUserName)
			}
			wg.Wait()
		})

		Step(fmt.Sprintf("Taking namespace label schedule backup of applications with and without rules from non-admin user"), func() {
			for _, nonAdminUserName := range userNames {
				log.InfoD(fmt.Sprintf("Taking namespace label schedule backup of applications with and without rules from non-admin user [%s]", nonAdminUserName))
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
					scheduleBackupName, err := CreateScheduleBackupWithNamespaceLabelWithValidation(nonAdminCtx, scheduleNameMap[nonAdminUserName], SourceClusterName, backupLocationNameMap[nonAdminUserName], backupLocationUidMap[nonAdminUserName], appContextsToBackup,
						labelSelectors, orgID, preRuleNameMap[nonAdminUserName], preRuleUidMap[nonAdminUserName], postRuleNameMap[nonAdminUserName], postRuleUidMap[nonAdminUserName], namespaceLabel, periodicSchedulePolicyNameMap[nonAdminUserName], periodicSchedulePolicyUidMap[nonAdminUserName])
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of backup [%s]", scheduleBackupName))
					mutipleNamespaceLabelBackupsMap[nonAdminUserName] = SafeAppend(&mutex, mutipleNamespaceLabelBackupsMap[nonAdminUserName], scheduleBackupName).([]string)
					err = suspendBackupSchedule(scheduleNameMap[nonAdminUserName], periodicSchedulePolicyNameMap[nonAdminUserName], orgID, nonAdminCtx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Suspending Backup Schedule [%s] for user [%s]", scheduleNameMap[nonAdminUserName], nonAdminUserName))
				}(nonAdminUserName)
			}
			wg.Wait()
		})

		// single namespace backups restore
		Step("Restore single namespace backups with different configs", func() {
			log.InfoD("Restore single namespace backups with different configs")
			for _, nonAdminUserName := range userNames {
				time.Sleep(timeBetweenConsecutiveBackups)
				nonAdminCtx, err := backup.GetNonAdminCtx(nonAdminUserName, commonPassword)
				log.FailOnError(err, "Fetching non admin ctx")
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
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

		// Restore a mutiple namespace backup
		Step("Restore a mutiple namespace backups", func() {
			log.InfoD("Restore a mutiple namespace backups")
			for _, nonAdminUserName := range userNames {
				time.Sleep(timeBetweenConsecutiveBackups)
				nonAdminCtx, err := backup.GetNonAdminCtx(nonAdminUserName, commonPassword)
				log.FailOnError(err, "Fetching non admin ctx")
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					restoreName := fmt.Sprintf("%s-mutiple-ns-restore-%s", nonAdminUserName, RandomString(4))
					restoreNameMap[nonAdminUserName] = restoreName
					log.InfoD("Restoring mutiple namespace backup [%s] in cluster [%s] with restore name [%s] for user [%s] ", mutipleNamespaceBackupsMap[nonAdminUserName][0], destinationClusterName, restoreNameMap[nonAdminUserName], nonAdminUserName)
					err = CreateRestore(restoreNameMap[nonAdminUserName], mutipleNamespaceBackupsMap[nonAdminUserName][0], namespaceMapping, destinationClusterName, orgID, nonAdminCtx, storageClassMapping)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying restoration [%s] of mutiple namespace schedule backup [%s] in cluster [%s] for user [%s]", restoreNameMap[nonAdminUserName], mutipleNamespaceBackupsMap[nonAdminUserName][0], destinationClusterName, nonAdminUserName))
				}(nonAdminUserName)
			}
			wg.Wait()
		})

		// Restore a mutiple namespace backup
		Step("Restore a namespace label backups", func() {
			log.InfoD("Restore a namespace label backups")
			for _, nonAdminUserName := range userNames {
				time.Sleep(timeBetweenConsecutiveBackups)
				nonAdminCtx, err := backup.GetNonAdminCtx(nonAdminUserName, commonPassword)
				log.FailOnError(err, "Fetching non admin ctx")
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					restoreName := fmt.Sprintf("%s-mutiple-ns-label-restore-%s", nonAdminUserName, RandomString(4))
					restoreNameMap[nonAdminUserName] = restoreName
					log.InfoD("Restoring mutiple namespace backup [%s] in cluster [%s] with restore name [%s] ", mutipleNamespaceLabelBackupsMap[nonAdminUserName][0], destinationClusterName, restoreNameMap[nonAdminUserName])
					err = CreateRestore(restoreNameMap[nonAdminUserName], mutipleNamespaceLabelBackupsMap[nonAdminUserName][0], namespaceMapping, destinationClusterName, orgID, nonAdminCtx, storageClassMapping)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying restoration [%s] of mutiple namespace schedule backup [%s] in cluster [%s] for user [%s]", restoreNameMap[nonAdminUserName], mutipleNamespaceLabelBackupsMap[nonAdminUserName][0], destinationClusterName, nonAdminUserName))
				}(nonAdminUserName)
			}
			wg.Wait()
		})

		log.InfoD("Deletion of all backup,restore,schedule,cluster of users when user is present in keycloak ")
		Step(fmt.Sprintf("Listing and Deletion of backup of non-admin user from px-admin user"), func() {
			log.InfoD("Listing and Deletion of backup of non-admin user from px-admin user")
			adminCtx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			for _, nonAdminUserName := range userNames[:2] {
				log.InfoD(fmt.Sprintf("Verifying listing backups of non-admin user [%s] from px-admin user", nonAdminUserName))
				userBackupNames, err := GetAllBackupNamesByOwnerID(userIdMap[nonAdminUserName], orgID, adminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of fetching backups [%v] of user [%s] from px-admin user", userBackupNames, nonAdminUserName))
				userBackupNamesMap[nonAdminUserName] = userBackupNames
				log.Infof(fmt.Sprintf("The list of user [%s] backups from px-admin %v", nonAdminUserName, userBackupNamesMap[nonAdminUserName]))
			}
			for _, nonAdminUserName := range userNames[:2] {
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					for _, backupName := range userBackupNamesMap[nonAdminUserName] {
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
			for _, nonAdminUserName := range userNames[:2] {
				log.InfoD(fmt.Sprintf("Verifying listing of backup schedule of non-admin user [%s] from px-admin user", nonAdminUserName))
				userBackupSchedules, err := GetAllBackupSchedulesForUser(nonAdminUserName, commonPassword)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of fetching backup schedules of user from px-admin user"))
				userBackupSchedulesMap[nonAdminUserName] = userBackupSchedules
				log.Infof(fmt.Sprintf("the list of user [%s ]backup schedules from px-admin user %v", nonAdminUserName, userBackupSchedulesMap[nonAdminUserName]))
			}
			for _, nonAdminUserName := range userNames[:2] {
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
			for _, nonAdminUserName := range userNames[:2] {
				log.InfoD(fmt.Sprintf("Verifying  listing of restores of non-admin user [%s] from px-admin user", nonAdminUserName))
				userRestores, err := GetAllRestoresForUser(nonAdminUserName, commonPassword)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of fetching restores of user ffrom px-admin user"))
				userRestoresMap[nonAdminUserName] = userRestores
				log.Infof(fmt.Sprintf("the list of user [%s] restores from px-admin user %v", nonAdminUserName, userRestoresMap[nonAdminUserName]))
			}
			for _, nonAdminUserName := range userNames[:2] {
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
			for _, nonAdminUserName := range userNames[:2] {
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
			for _, nonAdminUserName := range userNames[2:4] {
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
			for _, nonAdminUserName := range userNames[2:4] {
				log.InfoD(fmt.Sprintf("Verifying listing backups of non-admin user [%s] from px-admin user", nonAdminUserName))
				userBackupNames, err := GetAllBackupNamesByOwnerID(userIdMap[nonAdminUserName], orgID, adminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of fetching backups [%v] of user [%s] from px-admin user", userBackupNames, nonAdminUserName))
				userBackupNamesMap[nonAdminUserName] = userBackupNames
				log.Infof(fmt.Sprintf("The list of user [%s] backups from px-admin  %v", nonAdminUserName, userBackupNamesMap[nonAdminUserName]))
			}
			for _, nonAdminUserName := range userNames[2:4] {
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					for _, backupName := range userBackupNamesMap[nonAdminUserName] {
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
			for _, nonAdminUserName := range userNames[2:4] {
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
			for _, nonAdminUserName := range userNames[2:4] {
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
			for _, nonAdminUserName := range userNames[2:4] {
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					log.InfoD(fmt.Sprintf("Verifying  deletion of clusters of deleted non-admin user [%s] from px-admin user", nonAdminUserName))
					err := DeleteClusterWithUID(SourceClusterName, clusterUidMap[nonAdminUserName][SourceClusterName], orgID, adminCtx, true)
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
		schedulePolicyNames, err := backupDriver.GetAllSchedulePolicies(ctx, orgID)
		for _, schedulePolicyName := range schedulePolicyNames {
			err = Inst().Backup.DeleteBackupSchedulePolicy(orgID, []string{schedulePolicyName})
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting backup schedule policy %s ", []string{schedulePolicyName}))
		}
		ruleNames, err := backupDriver.GetAllRules(ctx, orgID)
		for _, ruleName := range ruleNames {
			err = Inst().Backup.DeleteRuleForBackup(orgID, ruleName)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting  rule %s ", ruleName))
		}
		CleanupCloudSettingsAndClusters(backupLocationMap, adminCredName, adminCloudCredUID, ctx)
	})
})

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
				userBackupNamesMap[nonAdminUserName] = userBackupNames
				log.Infof(fmt.Sprintf("The list of user [%s] backups from px-admin %v", nonAdminUserName, userBackupNamesMap[nonAdminUserName]))
			}
			for _, nonAdminUserName := range userNames[:3] {
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					for _, backupName := range userBackupNamesMap[nonAdminUserName] {
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
				userBackupNamesMap[nonAdminUserName] = userBackupNames
				log.Infof(fmt.Sprintf("The list of user [%s] backups from px-admin %v", nonAdminUserName, userBackupNamesMap[nonAdminUserName]))
			}
			for _, nonAdminUserName := range userNames[3:6] {
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					for _, backupName := range userBackupNamesMap[nonAdminUserName] {
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

// UpdatesBackupOfUserFromAdmin updates backups of non admin user from px-admin with valid/in-valid cloud account.
var _ = Describe("{UpdatesBackupOfUserFromAdmin}", func() {
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/87568
	var scheduledAppContexts []*scheduler.Context
	var backupLocationUID string
	var bkpNamespaces []string
	var credName string
	var cloudCredUID string
	var invalidCloudCredUID string
	var invalidCredName string
	var srcClusterUid string
	var destClusterUid string
	var backupLocationName string
	var nonAdminUserName string
	var providers []string
	var periodicSchedulePolicyName string
	var periodicSchedulePolicyInterval int64
	var periodicSchedulePolicyUid string

	bkpNamespaces = make([]string, 0)
	backupLocationMap := make(map[string]string)
	providers = getProviders()

	JustBeforeEach(func() {
		StartTorpedoTest("UpdatesBackupOfUserFromAdmin",
			"Updates backups of non admin user from px-admin with valid/in-valid account", nil, 87568)
		log.InfoD("Deploy applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
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

	It("Updates Backups and Cluster of user from px-admin", func() {
		Step("Validate applications", func() {
			log.InfoD("Validate applications")
			ValidateApplications(scheduledAppContexts)
		})

		Step("Create a non-admin user", func() {
			log.InfoD(fmt.Sprintf("Create a non-admin user"))
			nonAdminUserName = createUsers(1)[0]
			err := backup.AddRoleToUser(nonAdminUserName, backup.InfrastructureOwner, fmt.Sprintf("Adding %v role to %s", backup.InfrastructureOwner, nonAdminUserName))
			log.FailOnError(err, "Failed to add role for user - %s", nonAdminUserName)
		})

		Step(fmt.Sprintf("Adding Credentials and Backup Location from non-admin user"), func() {
			log.InfoD(fmt.Sprintf("Creating cloud credentials and backup location from non-adminuser"))
			for _, provider := range providers {
				nonAdminCtx, err := backup.GetNonAdminCtx(nonAdminUserName, commonPassword)
				log.FailOnError(err, "Fetching non-admin ctx")
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				credName = fmt.Sprintf("autogenerated-cred-%v", time.Now().Unix())
				err = CreateCloudCredential(provider, credName, cloudCredUID, orgID, nonAdminCtx)
				log.FailOnError(err, "Failed to create cloud credential - %s", err)
				backupLocationName = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
				err = CreateBackupLocationWithContext(provider, backupLocationName, backupLocationUID, credName, cloudCredUID, getGlobalBucketName(provider), orgID, "", "", nonAdminCtx)
				log.FailOnError(err, "Failed to add backup location %s using provider %s for px-admin user", backupLocationName, provider)
				backupLocationMap[backupLocationUID] = backupLocationName
			}
		})

		Step(fmt.Sprintf("Create schedule policy from non-admin user"), func() {
			log.InfoD(fmt.Sprintf("Creating a schedule policy from non-admin [%s] user", nonAdminUserName))
			nonAdminCtx, err := backup.GetNonAdminCtx(nonAdminUserName, commonPassword)
			log.FailOnError(err, "Fetching non admin user ctx")
			periodicSchedulePolicyName = fmt.Sprintf("%s-%v-%s", "periodic", time.Now().Unix(), nonAdminUserName)
			periodicSchedulePolicyUid = uuid.New()
			periodicSchedulePolicyInterval = 15
			err = CreateBackupScheduleIntervalPolicy(5, periodicSchedulePolicyInterval, 5, periodicSchedulePolicyName, periodicSchedulePolicyUid, orgID, nonAdminCtx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of periodic schedule policy of interval [%v] minutes named [%s] for user [%s]", periodicSchedulePolicyInterval, periodicSchedulePolicyName, nonAdminUserName))

		})

		Step(fmt.Sprintf("Register source and destination cluster for backup on %s ", nonAdminUserName), func() {
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
			destClusterUid, err = Inst().Backup.GetClusterUID(nonAdminCtx, orgID, destinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", destinationClusterName))
		})

		Step(fmt.Sprintf("Taking manual backup and schedule backup of applications for user %s", nonAdminUserName), func() {
			log.InfoD(fmt.Sprintf("Taking manual backup and schedule backup of applications for user%s", nonAdminUserName))
			nonAdminCtx, err := backup.GetNonAdminCtx(nonAdminUserName, commonPassword)
			log.FailOnError(err, "Fetching non admin ctx")
			labelSelectors := make(map[string]string, 0)
			backupName := fmt.Sprintf("%s-manual-%s-%s", BackupNamePrefix, nonAdminUserName, RandomString(4))
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			err = CreateBackupWithValidation(nonAdminCtx, backupName, SourceClusterName, backupLocationName, backupLocationUID, appContextsToBackup,
				labelSelectors, orgID, srcClusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of backup [%s]", backupName))

			scheduleName := fmt.Sprintf("%s-schedule-ns-%s-without-rules-%s", BackupNamePrefix, nonAdminUserName, RandomString(4))
			log.InfoD("Creating a schedule backup of namespace [%s] without pre and post exec rules", bkpNamespaces[0])
			appContextsToBackup = FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			scheduleBackupName, err := CreateScheduleBackupWithValidation(nonAdminCtx, scheduleName, SourceClusterName, backupLocationName, backupLocationUID, appContextsToBackup,
				labelSelectors, orgID, "", "", "", "", periodicSchedulePolicyName, periodicSchedulePolicyUid)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of backup [%s]", scheduleBackupName))

		})

		Step("Create invalid credential for cluster and backup object", func() {
			log.InfoD("Create invalid credential for cluster and backup object")
			nonAdminCtx, err := backup.GetNonAdminCtx(nonAdminUserName, commonPassword)
			log.FailOnError(err, "Fetching non-admin ctx")
			invalidCloudCredUID = uuid.New()
			backupLocationUID = uuid.New()
			invalidCredName = fmt.Sprintf("invalid-autogenerated-cred-%v", time.Now().Unix())
			err = createInvalidAWSCloudCredential(invalidCredName, invalidCloudCredUID, orgID, nonAdminCtx)
			log.FailOnError(err, "Failed to create invalid cloud credential - %s", err)
		})

		Step("Verifying listing and updation of backup of non-admin user from px-admin", func() {
			log.InfoD("Verifying listing and updation of backup of non-admin user from px-admin")
			adminCtx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			userUID, err := backup.FetchIDOfUser(nonAdminUserName)
			userBackupNames, err := GetAllBackupNamesByOwnerID(userUID, orgID, adminCtx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of fetching backups [%v] of user [%s] from px-admin user", userBackupNames, nonAdminUserName))
			log.Infof(fmt.Sprintf("The list of user [%s] backups from px-admin %v", nonAdminUserName, userBackupNames))
			for _, backupName := range userBackupNames {
				bkpUid, _ := Inst().Backup.GetBackupUID(adminCtx, backupName, orgID)
				_, err = UpdateBackup(backupName, bkpUid, orgID, invalidCredName, invalidCloudCredUID, adminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of updation of backup [%v] of user [%s] from px-admin user", backupName, nonAdminUserName))
			}
		})

		Step("Verifying deletion of backup of non-admin user from px-admin", func() {
			log.InfoD("Verifying deletion of backup of non-admin user from px-admin")
			adminCtx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching admin ctx")
			userUID, err := backup.FetchIDOfUser(nonAdminUserName)
			userBackupNames, err := GetAllBackupNamesByOwnerID(userUID, orgID, adminCtx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of fetching backups [%v] of user [%s] from px-admin user", userBackupNames, nonAdminUserName))
			log.Infof(fmt.Sprintf("The list of user [%s] backups from px-admin %v", nonAdminUserName, userBackupNames))
			for _, backupName := range userBackupNames {
				backupUID, _ := Inst().Backup.GetBackupUID(adminCtx, backupName, orgID)
				_, err = DeleteBackup(backupName, backupUID, orgID, adminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying deletion backup [%s] of non-admin user [%s] from px-admin user", backupName, nonAdminUserName))
				err = DeleteBackupAndWait(backupName, adminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("waiting for backup [%s] deletion", backupName))
			}
		})

		Step("Verifying deletion of backup schedule of non-admin user from px-admin", func() {
			log.InfoD("Verifying deletion of backup schedule of non-admin user from px-admin")
			adminCtx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching admin ctx")
			userBackupScheduleNames, err := GetAllBackupSchedulesForUser(nonAdminUserName, commonPassword)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of fetching backups [%v] of user [%s] from px-admin user", userBackupScheduleNames, nonAdminUserName))
			log.Infof(fmt.Sprintf("The list of user [%s] backups from px-admin %v", nonAdminUserName, userBackupScheduleNames))
			for _, userBackupScheduleName := range userBackupScheduleNames {
				log.InfoD(fmt.Sprintf("Verifying deletion of backup schedule [%s] of non-admin user [%s] from px-admin user", userBackupScheduleName, nonAdminUserName))
				backupScheduleUid, err := GetScheduleUID(userBackupScheduleName, orgID, adminCtx)
				log.FailOnError(err, fmt.Sprintf("Fetching schedule uid for shedule [%s]", userBackupScheduleName))
				err = DeleteScheduleWithUIDAndWait(userBackupScheduleName, backupScheduleUid, SourceClusterName, srcClusterUid, orgID, adminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of deleting backup scheudle [%s] of user [%s] from px-admin user", userBackupScheduleName, nonAdminUserName))
			}
		})

		Step("Verifying updation of cluster of non-admin user from px-admin", func() {
			log.InfoD("Verifying updation of cluster of non-admin user from px-admin")
			adminCtx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			srcClusterConfigPath, err := GetSourceClusterConfigPath()
			log.FailOnError(err, "Fetching source clusterconfigpath")
			_, err = UpdateCluster(SourceClusterName, srcClusterUid, srcClusterConfigPath, orgID, invalidCredName, invalidCloudCredUID, adminCtx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of updation of cluster [%v] of user [%s] from px-admin user", SourceClusterName, nonAdminUserName))
			dstClusterConfigPath, err := GetDestinationClusterConfigPath()
			log.FailOnError(err, "Fetching destination clusterconfigpath")
			_, err = UpdateCluster(destinationClusterName, destClusterUid, dstClusterConfigPath, orgID, invalidCredName, invalidCloudCredUID, adminCtx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of updation of cluster [%v] of user [%s] from px-admin user", destinationClusterName, nonAdminUserName))
		})

		Step(fmt.Sprintf("Verifying  deletion of clusters of non-admin user from px-admin user"), func() {
			log.InfoD("Verifying  deletion of clusters of non-admin user from px-admin user")
			adminCtx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			err = DeleteClusterWithUID(SourceClusterName, srcClusterUid, orgID, adminCtx, true)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting cluster %s", SourceClusterName))
			err = DeleteClusterWithUID(destinationClusterName, destClusterUid, orgID, adminCtx, true)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting cluster %s", destinationClusterName))
		})
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		ctx, err := backup.GetNonAdminCtx(nonAdminUserName, commonPassword)
		log.FailOnError(err, "Fetching px-central-admin ctx")
		log.InfoD("Deleting the deployed apps after the testcase")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)
		log.Infof("Deleting backup schedule policy")
		err = Inst().Backup.DeleteBackupSchedulePolicy(orgID, []string{periodicSchedulePolicyName})
		dash.VerifySafely(err, nil, fmt.Sprintf("Deleting backup schedule policies %s ", []string{periodicSchedulePolicyName}))
		CleanupCloudSettingsAndClusters(backupLocationMap, credName, cloudCredUID, ctx)
	})
})

// DeleteBackupSharedByMultipleUsersFromAdmin deletes backups non admin user from px-admin when backup is shared by multiple users.
var _ = Describe("{DeleteBackupSharedByMultipleUsersFromAdmin}", func() {
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/87565

	var (
		scheduledAppContexts           []*scheduler.Context
		backupLocationUID              string
		adminCredName                  string
		userName                       string
		adminCloudCredUID              string
		srcClusterUid                  string
		backupLocationName             string
		periodicSchedulePolicyName     string
		periodicSchedulePolicyInterval int64
		periodicSchedulePolicyUid      string
		appAdminUserName               string
		infraAdminUserName             string
		appUserName                    string
		userNames                      []string
		wg                             sync.WaitGroup
		mutex                          sync.Mutex
		bkpNamespaces                  = make([]string, 0)
		timeBetweenConsecutiveBackups  = 10 * time.Second
		backupLocationMap              = make(map[string]string)
		providers                      = getProviders()
		userIdMap                      = make(map[string]string)
		backupLocationUserMap          = make(map[string]string)
		backupLocationUidUserMap       = make(map[string]string)
		clusterUidMap                  = make(map[string]map[string]string)
		scheduleNameMap                = make(map[string]string)
		userBackupSchedulesMap         = make(map[string][]string)
		userCloudCredentialMap         = make(map[string]map[string]string)
		userBackupLocationMap          = make(map[string]map[string]string)
		userBackupsMap                 = make(map[string][]string)
		backupDriver                   = Inst().Backup
	)

	JustBeforeEach(func() {
		StartTorpedoTest("DeleteBackupSharedByMultipleUsersFromAdmin",
			"Delete backups of non admin user from px-admin when backup is shared by multiple users", nil, 87565)
		log.InfoD("Deploy applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
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

	It("Delete backups of non admin user from px-admin", func() {
		Step("Validate applications", func() {
			log.InfoD("Validate applications")
			ValidateApplications(scheduledAppContexts)
		})

		Step("Create a non-admin users to create the backups and restore", func() {
			log.InfoD(fmt.Sprintf("Create a non-admin users to create the backups and restore"))
			roles := [3]backup.PxBackupRole{backup.ApplicationOwner, backup.InfrastructureOwner, backup.ApplicationUser}
			for _, role := range roles {
				userName = createUsers(1)[0]
				err := backup.AddRoleToUser(userName, role, fmt.Sprintf("Adding %v role to %s", role, userName))
				log.FailOnError(err, "Failed to add role for user - %s", userName)
				if role == backup.ApplicationOwner {
					appAdminUserName = userName
				} else if role == backup.InfrastructureOwner {
					infraAdminUserName = userName
				} else {
					appUserName = userName
				}
				userNames = append(userNames, userName)
				userUID, err := backup.FetchIDOfUser(userName)
				log.FailOnError(err, "Failed to fetch uid for - %s", userName)
				userIdMap[userName] = userUID
			}
		})

		Step(fmt.Sprintf("Adding Credentials from px-admin user and sharing with required users"), func() {
			log.InfoD(fmt.Sprintf("Adding Credentials from px-admin user and sharing with required users"))
			adminCtx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			for _, provider := range providers {
				adminCloudCredUID = uuid.New()
				adminCredName = fmt.Sprintf("autogenerated-cred-%v", time.Now().Unix())
				err = CreateCloudCredential(provider, adminCredName, adminCloudCredUID, orgID, adminCtx)
				log.FailOnError(err, "Failed to create cloud credential - %s", err)
				log.Infof("Updating cloud credential ownership for non-admin users")
				if provider != drivers.ProviderNfs {
					err = UpdateCloudCredentialOwnership(adminCredName, adminCloudCredUID, []string{appAdminUserName}, nil, Read, Invalid, adminCtx, orgID)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying ownership update for cloud credential %s to user %s", adminCredName, []string{appAdminUserName}))
				}
			}
		})

		Step(fmt.Sprintf("Adding backup location from px-admin user and sharing with app users"), func() {
			log.InfoD(fmt.Sprintf("Adding backup location  from px-admin user and sharing with app users"))
			adminCtx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			for _, provider := range providers {
				backupLocationName = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
				backupLocationUID = uuid.New()
				err := CreateBackupLocationWithContext(provider, backupLocationName, backupLocationUID, adminCredName, adminCloudCredUID, getGlobalBucketName(provider), orgID, "", "", adminCtx)
				log.FailOnError(err, "Failed to add backup location %s using provider %s for px-admin user", backupLocationName, provider)
				log.Infof("Updating backup location ownership for non-admin users")
				err = UpdateBackupLocationOwnership(backupLocationName, backupLocationUID, []string{appUserName}, nil, Read, Invalid, adminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying ownership update for backup location %s to user %s", backupLocationName, []string{appUserName}))
				backupLocationUserMap[appUserName] = backupLocationName
				backupLocationUidUserMap[appUserName] = backupLocationUID
				backupLocationMap[backupLocationUID] = backupLocationName
			}
		})

		Step(fmt.Sprintf("Adding backup location for app-admin user and shared cloud cred from px-admin"), func() {
			log.InfoD(fmt.Sprintf("Adding backup location for app-admin user and shared cloud cred from px-admin"))
			for _, provider := range providers {
				nonAdminCtx, err := backup.GetNonAdminCtx(appAdminUserName, commonPassword)
				log.FailOnError(err, "Fetching px-admin ctx")
				backupLocationName = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
				backupLocationUID = uuid.New()
				err = CreateBackupLocationWithContext(provider, backupLocationName, backupLocationUID, adminCredName, adminCloudCredUID, getGlobalBucketName(provider), orgID, "", "", nonAdminCtx)
				log.FailOnError(err, "Failed to add backup location %s using provider %s for px-admin user", backupLocationName, provider)
				backupLocationUserMap[appAdminUserName] = backupLocationName
				backupLocationUidUserMap[appAdminUserName] = backupLocationUID
				userBackupLocationMap[appAdminUserName] = map[string]string{backupLocationUID: backupLocationName}
			}
		})

		Step(fmt.Sprintf("Adding cloud credentials and backup location for infra-admin users"), func() {
			log.InfoD(fmt.Sprintf("Adding cloud credentials and backup location for infra-admin users"))
			for _, provider := range providers {
				nonAdminCtx, err := backup.GetNonAdminCtx(infraAdminUserName, commonPassword)
				log.FailOnError(err, "Fetching px-admin ctx")
				cloudCredUID := uuid.New()
				credName := fmt.Sprintf("autogenerated-cred-%v", time.Now().Unix())
				err = CreateCloudCredential(provider, credName, cloudCredUID, orgID, nonAdminCtx)
				log.FailOnError(err, "Failed to create cloud credential - %s", err)
				backupLocationName = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
				backupLocationUID = uuid.New()
				err = CreateBackupLocationWithContext(provider, backupLocationName, backupLocationUID, credName, cloudCredUID, getGlobalBucketName(provider), orgID, "", "", nonAdminCtx)
				log.FailOnError(err, "Failed to add backup location %s using provider %s for px-admin user", backupLocationName, provider)
				backupLocationUserMap[infraAdminUserName] = backupLocationName
				backupLocationUidUserMap[infraAdminUserName] = backupLocationUID
				userCloudCredentialMap[infraAdminUserName] = map[string]string{cloudCredUID: credName}
				userBackupLocationMap[infraAdminUserName] = map[string]string{backupLocationUID: backupLocationName}
			}
		})

		Step(fmt.Sprintf("Create schedule policy from px-admin user"), func() {
			log.InfoD(fmt.Sprintf("Creating a schedule policy from px-admin user"))
			adminCtx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			periodicSchedulePolicyName = fmt.Sprintf("%s-%v", "periodic", time.Now().Unix())
			periodicSchedulePolicyUid = uuid.New()
			periodicSchedulePolicyInterval = 15
			err = CreateBackupScheduleIntervalPolicy(5, periodicSchedulePolicyInterval, 5, periodicSchedulePolicyName, periodicSchedulePolicyUid, orgID, adminCtx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of periodic schedule policy of interval [%v] minutes named [%s] for px-admin ", periodicSchedulePolicyInterval, periodicSchedulePolicyName))
			err = UpdateSchedulePolicyOwnership(periodicSchedulePolicyName, periodicSchedulePolicyUid, userNames, nil, Read, Invalid, adminCtx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying updation of owbership for SchedulePolicy- %s", periodicSchedulePolicyName))

		})

		Step(fmt.Sprintf("Register source and destination cluster for backup on non-admin user"), func() {
			log.InfoD(fmt.Sprintf("Register source and destination cluster for backup on non-admin user"))
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

		Step(fmt.Sprintf("Taking manual backup of applications as non-admin user"), func() {
			log.InfoD(fmt.Sprintf("Taking manual backup of applications as non-admin user"))
			for _, nonAdminUserName := range userNames {
				time.Sleep(timeBetweenConsecutiveBackups)
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					nonAdminCtx, err := backup.GetNonAdminCtx(nonAdminUserName, commonPassword)
					log.FailOnError(err, "Fetching non admin ctx")
					labelSelectors := make(map[string]string, 0)
					backupName := fmt.Sprintf("%s-manual-%s-%s", BackupNamePrefix, nonAdminUserName, RandomString(4))
					appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
					err = CreateBackupWithValidation(nonAdminCtx, backupName, SourceClusterName, backupLocationUserMap[nonAdminUserName], backupLocationUidUserMap[nonAdminUserName], appContextsToBackup,
						labelSelectors, orgID, clusterUidMap[nonAdminUserName][SourceClusterName], "", "", "", "")
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of backup [%s]", backupName))
					userBackupsMap[nonAdminUserName] = SafeAppend(&mutex, userBackupsMap[nonAdminUserName], backupName).([]string)
				}(nonAdminUserName)
			}
			wg.Wait()
		})

		Step(fmt.Sprintf("Taking schedule backup of applications as non-admin user "), func() {
			log.InfoD(fmt.Sprintf("Taking schedule backup of applications as non-admin user"))
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
					scheduleBackupName, err := CreateScheduleBackupWithValidation(nonAdminCtx, scheduleNameMap[nonAdminUserName], SourceClusterName, backupLocationUserMap[nonAdminUserName], backupLocationUidUserMap[nonAdminUserName], appContextsToBackup,
						labelSelectors, orgID, "", "", "", "", periodicSchedulePolicyName, periodicSchedulePolicyUid)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of backup [%s]", scheduleBackupName))
					err = suspendBackupSchedule(scheduleNameMap[nonAdminUserName], periodicSchedulePolicyName, orgID, nonAdminCtx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Suspending Backup Schedule [%s] for user [%s]", scheduleNameMap[nonAdminUserName], nonAdminUserName))
					userBackupsMap[nonAdminUserName] = SafeAppend(&mutex, userBackupsMap[nonAdminUserName], scheduleBackupName).([]string)
				}(nonAdminUserName)
			}
			wg.Wait()
		})

		Step("Sharing of backup of app-user user with app-admin and infra-admin user", func() {
			log.InfoD(fmt.Sprintf("Sharing of backup of app-user user with app-admin and infra-admin user"))
			nonAdminCtx, err := backup.GetNonAdminCtx(appUserName, commonPassword)
			log.FailOnError(err, "Fetching non-admin ctx")
			log.Infof(fmt.Sprintf("Sharing app-user [%s] backups [%v] with app-admin and infra-admin", appUserName, userBackupsMap[appUserName]))
			for _, appUserBackupName := range userBackupsMap[appUserName] {
				err := ShareBackup(appUserBackupName, nil, []string{appAdminUserName, infraAdminUserName}, 1, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of sharing backup [%v] of user [%s] ", appUserBackupName, appUserName))
			}
		})

		Step("Sharing of backup of app-admin user with app-user and infra-admin user", func() {
			log.InfoD(fmt.Sprintf("Sharing of backup of app-admin user with app-user and infra-admin user"))
			nonAdminCtx, err := backup.GetNonAdminCtx(appAdminUserName, commonPassword)
			log.FailOnError(err, "Fetching non-admin ctx")
			log.Infof(fmt.Sprintf("Sharing app-admin user [%s] backups [%v] with app-user and infra-admin", appAdminUserName, userBackupsMap[appAdminUserName]))
			for _, appAdminUserBackupName := range userBackupsMap[appAdminUserName] {
				err := ShareBackup(appAdminUserBackupName, nil, []string{appUserName, infraAdminUserName}, 1, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of sharing backup [%v] of user [%s] ", appAdminUserBackupName, appAdminUserName))
			}
		})

		Step("Sharing of backup of infra-admin user with app-user and app-admin user", func() {
			log.InfoD(fmt.Sprintf("Sharing of backup of infra-admin user with app-user and app-admin user"))
			nonAdminCtx, err := backup.GetNonAdminCtx(infraAdminUserName, commonPassword)
			log.FailOnError(err, "Fetching non-admin ctx")
			log.Infof(fmt.Sprintf("Sharing infra-admin user [%s] backups [%v] with app-admin and app-user", infraAdminUserName, userBackupsMap[infraAdminUserName]))
			for _, infraAdminUserBackupName := range userBackupsMap[infraAdminUserName] {
				err := ShareBackup(infraAdminUserBackupName, nil, []string{appAdminUserName, appUserName}, 3, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of sharing backup [%v] of user [%s] ", infraAdminUserBackupName, infraAdminUserName))
			}
		})

		Step(fmt.Sprintf("Listing and Deletion of shared backup of non-admin users from px-admin user"), func() {
			log.InfoD(fmt.Sprintf("Listing and Deletion of shared backup of non-admin users from px-admin user"))
			adminCtx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			for _, nonAdminUserName := range userNames {
				log.InfoD(fmt.Sprintf("Verifying listing backups of non-admin user [%s] from px-admin user", nonAdminUserName))
				userBackupNames, err := GetAllBackupNamesByOwnerID(userIdMap[nonAdminUserName], orgID, adminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of fetching backups [%v] of user [%s] from px-admin user", userBackupNames, nonAdminUserName))
				userBackupsMap[nonAdminUserName] = userBackupNames
				log.Infof(fmt.Sprintf("The list of user [%s] backups from px-admin [%v]", nonAdminUserName, userBackupsMap[nonAdminUserName]))
			}
			for _, nonAdminUserName := range userNames {
				wg.Add(1)
				go func(nonAdminUserName string) {
					defer GinkgoRecover()
					defer wg.Done()
					for _, backupName := range userBackupsMap[nonAdminUserName] {
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

	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)
		for _, nonAdminUserName := range userNames {
			userBackupSchedules, err := GetAllBackupSchedulesForUser(nonAdminUserName, commonPassword)
			dash.VerifySafely(err, nil, fmt.Sprintf("Verification of fetching backup schedules of user "))
			userBackupSchedulesMap[nonAdminUserName] = userBackupSchedules
		}
		for _, nonAdminUserName := range userNames {
			nonAdminCtx, err := backup.GetNonAdminCtx(nonAdminUserName, commonPassword)
			log.FailOnError(err, "Fetching non admin ctx")
			wg.Add(1)
			go func(nonAdminUserName string) {
				defer GinkgoRecover()
				defer wg.Done()
				for _, backupScheduleName := range userBackupSchedulesMap[nonAdminUserName] {
					log.InfoD(fmt.Sprintf("Verifying deletion of backup schedule [%s] of non-admin user [%s] ", backupScheduleName, nonAdminUserName))
					backupScheduleUid, err := GetScheduleUID(backupScheduleName, orgID, nonAdminCtx)
					log.FailOnError(err, fmt.Sprintf("Fetching schedule uid for shedule [%s]", backupScheduleName))
					err = DeleteScheduleWithUIDAndWait(backupScheduleName, backupScheduleUid, SourceClusterName, clusterUidMap[nonAdminUserName][SourceClusterName], orgID, nonAdminCtx)
					dash.VerifySafely(err, nil, fmt.Sprintf("Verification of deleting backup scheudle [%s] of user [%s] from px-admin user", backupScheduleName, nonAdminUserName))
				}
			}(nonAdminUserName)
		}
		wg.Wait()
		log.Infof("Deleting backup schedule policy")
		err := Inst().Backup.DeleteBackupSchedulePolicy(orgID, []string{periodicSchedulePolicyName})
		dash.VerifySafely(err, nil, fmt.Sprintf("Deleting backup schedule policies %s ", []string{periodicSchedulePolicyName}))
		log.Infof("Cleaning up cluster, backup location and cloud credentials for non admin users")
		for _, nonAdminUserName := range userNames {
			nonAdminCtx, err := backup.GetNonAdminCtx(nonAdminUserName, commonPassword)
			log.FailOnError(err, "Fetching non admin ctx")
			for cloudCredentialUID, cloudCredentialName := range userCloudCredentialMap[nonAdminUserName] {
				CleanupCloudSettingsAndClusters(userBackupLocationMap[nonAdminUserName], cloudCredentialName, cloudCredentialUID, nonAdminCtx)
			}
			err = backup.DeleteUser(nonAdminUserName)
			log.FailOnError(err, "failed to delete user %s", nonAdminUserName)
		}
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching non admin ctx")
		CleanupCloudSettingsAndClusters(backupLocationMap, adminCredName, adminCloudCredUID, ctx)
	})
})
