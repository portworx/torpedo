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
				userBackupNames, err := GetAllBackupsOfUsersFromAdmin([]string{userIdMap[nonAdminUserName]}, adminCtx)
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
						err = DeleteScheduleWithClusterRef(backupScheduleName, backupScheduleUid, SourceClusterName, clusterUidMap[nonAdminUserName][SourceClusterName], orgID, adminCtx)
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
						err := DeleteRestoreWithUid(restoreName, restoreUid, orgID, adminCtx)
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
					err = DeleteClusterWithUid(SourceClusterName, clusterUidMap[nonAdminUserName][SourceClusterName], orgID, adminCtx, true)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting cluster %s", SourceClusterName))
					err = DeleteClusterWithUid(destinationClusterName, clusterUidMap[nonAdminUserName][destinationClusterName], orgID, adminCtx, true)
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
				userBackupNames, err := GetAllBackupsOfUsersFromAdmin([]string{userIdMap[nonAdminUserName]}, adminCtx)
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
						err = DeleteScheduleWithClusterRef(backupScheduleName, backupScheduleUid, SourceClusterName, clusterUidMap[nonAdminUserName][SourceClusterName], orgID, adminCtx)
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
						err := DeleteRestoreWithUid(restoreName, restoreUid, orgID, adminCtx)
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
					err := DeleteClusterWithUid(SourceClusterName, clusterUidMap[nonAdminUserName][SourceClusterName], orgID, adminCtx, true)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting cluster %s", SourceClusterName))
					err = DeleteClusterWithUid(destinationClusterName, clusterUidMap[nonAdminUserName][destinationClusterName], orgID, adminCtx, true)
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
		schedulePolicyNames, err := backupDriver.GetAllSchedulePolicys(ctx, orgID)
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
