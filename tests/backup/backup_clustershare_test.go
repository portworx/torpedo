package tests

import (
	context1 "context"
	"encoding/base64"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/pure-px/torpedo/drivers"
	"github.com/pure-px/torpedo/drivers/backup"
	"github.com/pure-px/torpedo/drivers/backup/portworx"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
	"io/ioutil"
	appsV1 "k8s.io/api/apps/v1"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

// This testcase verifies cluster share feature with large number of users and larger number of clusters
var _ = Describe("{ClusterShareWithLargeNumberOfUsersAndClusters}", Label(TestCaseLabelsMap[ClusterShareWithLargeNumberOfUsersAndClusters]...), func() {
	var (
		scheduledAppContexts               []*scheduler.Context
		adminCloudCredName                 string
		adminCloudCredentialUID            string
		backupLocationUID                  string
		backupLocationName                 string
		numberOfSharedUsers                int
		numberOfGroups                     int
		providers                          []string
		schedulePolicyName                 string
		schedulePolicyUID                  string
		schedulePolicyInterval             = int64(15)
		labelSelectors                     map[string]string
		randomRole                         backup.PxBackupRole
		sharedGroups                       []string
		userIDMap                          map[string]string
		userBackupsFromSharedClusterMap    map[string]string
		userRestoreFromSharedClusterMap    map[string]string
		backupNamespaceMap                 map[string]string
		backupLocationMap                  map[string]string
		userNamespaceMap                   map[string]string
		scheduleNameUserMap                map[string]string
		userClusterMap                     map[string]map[string]string
		sharedUserClusterMap               map[string]map[string]string
		masterUserBackupNames              []string
		backedUpNamespaces                 []string
		sharedUsers                        []string
		numberOfUsers                      int
		nonAdminUsers                      []string
		clusterSharedUsersList             []string
		restoredNamespaces                 []string
		iter2ClusterUserMap                map[string]map[string]string
		invalidKubeConfig                  = "\"\""
		inActiveClusterUsers               []string
		configPath                         string
		mu                                 sync.RWMutex
		userCtx                            map[string]context1.Context
		usersToBeDeleted                   []string
		userScheduleFromSharedClusterMap   map[string]string
		numberOfPrimaryBackups             int
		wg                                 sync.WaitGroup
		clusterShareOnPremCluster          []string
		isClusterShareSupportedForAllRoles bool
		roles                              []backup.PxBackupRole
		unsupportedUsers                   []string
		clusterSharePermissionError        = "PermissionDenied"
		numOfUnsupportedUsers              int
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("ClusterShareWithLargeNumberOfUsersAndClusters", "Verifies cluster share with large number of users and larger number of clusters", nil, 301138, Ak, Q3FY25)
		providers = GetBackupProviders()
		// checking if cluster share is supported for all roles
		clusterShareOnPremCluster = []string{"vanilla", "openshift"}
		if Contains(clusterShareOnPremCluster, GetClusterProvider()) {
			isClusterShareSupportedForAllRoles = true
		} else {
			isClusterShareSupportedForAllRoles = false
		}
		// keeping the user  and namespace count same , so that each user will take backup of each namespace
		numOfNamespace := 8
		numberOfUsers = 8
		numOfUnsupportedUsers = 4
		// initializing the count of user and groups where cluster share will be tested.
		numberOfSharedUsers, _ = strconv.Atoi(GetEnv(UsersToBeCreated, "20"))
		numberOfGroups, _ = strconv.Atoi(GetEnv(GroupsToBeCreated, "3"))
		numberOfPrimaryBackups, _ = strconv.Atoi(GetEnv(MaxBackupsToBeCreated, "2"))

		userBackupsFromSharedClusterMap = make(map[string]string)
		userRestoreFromSharedClusterMap = make(map[string]string)
		iter2ClusterUserMap = make(map[string]map[string]string)
		backupNamespaceMap = make(map[string]string)
		backupLocationMap = make(map[string]string)
		labelSelectors = make(map[string]string)
		userNamespaceMap = make(map[string]string)
		scheduleNameUserMap = make(map[string]string)
		userClusterMap = make(map[string]map[string]string)
		sharedUserClusterMap = make(map[string]map[string]string)
		userCtx = make(map[string]context1.Context)
		userScheduleFromSharedClusterMap = make(map[string]string)
		userIDMap = make(map[string]string)

		// Schedule applications
		log.InfoD("scheduling applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		for i := 0; i < numOfNamespace; i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, appCtx := range appContexts {
				appCtx.ReadinessTimeout = AppReadinessTimeout
				scheduledAppContexts = append(scheduledAppContexts, appCtx)
				appNamespace := appCtx.ScheduleOptions.Namespace
				backedUpNamespaces = append(backedUpNamespaces, appNamespace)
			}
		}
	})

	It("Verifies cluster share with large number of users and larger number of clusters", func() {
		defer func() {
			log.InfoD("switching to default context")
			err := SetClusterContext("")
			log.FailOnError(err, "failed to SetClusterContext to default cluster")
		}()

		Step("Validating applications", func() {
			log.InfoD("Validating applications")
			ValidateApplications(scheduledAppContexts)
		})

		Step("Create schedule policy from admin and make it public", func() {
			log.InfoD("Create schedule policy from admin and make it public")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			schedulePolicyName = fmt.Sprintf("%s-%v", "periodic-schedule-policy", RandomString(5))
			schedulePolicyUID = uuid.New()
			err = CreateBackupScheduleIntervalPolicy(5, schedulePolicyInterval, 5, schedulePolicyName, schedulePolicyUID, BackupOrgID, ctx, false, false)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of schedule policy %s", schedulePolicyName))
			log.InfoD("Update SchedulePolicy - %s ownership as public ", schedulePolicyName)
			err = AddSchedulePolicyOwnership(schedulePolicyName, schedulePolicyUID, nil, nil, Invalid, Read, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying updation of ownership for schedulepolicy - %s", schedulePolicyName))
		})

		Step("Creating backup location and cloud setting as admin and make it public", func() {
			log.InfoD("Creating backup location and cloud setting as admin and make it public")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, provider := range providers {
				adminCloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, RandomString(5))
				backupLocationName = fmt.Sprintf("%s-%s-bl-%v", provider, getGlobalBucketName(provider), RandomString(5))
				adminCloudCredentialUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocationName
				err := CreateCloudCredential(provider, adminCloudCredName, adminCloudCredentialUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", adminCloudCredName, BackupOrgID, provider))
				if provider != drivers.ProviderNfs {
					log.Infof("Update CloudAccount - %s ownership as public", adminCloudCredName)
					err = AddCloudCredentialOwnership(adminCloudCredName, adminCloudCredentialUID, nil, nil, Invalid, Read, ctx, BackupOrgID)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying updation of owbership for CloudCredential- %s", adminCloudCredName))
				}
				err = CreateBackupLocation(provider, backupLocationName, backupLocationUID, adminCloudCredName, adminCloudCredentialUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, "Creating backup location")
				err = AddBackupLocationOwnership(backupLocationName, backupLocationUID, nil, nil, Invalid, Read, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying updation of ownership for backuplocation - %s", backupLocationName))
			}
		})

		Step(fmt.Sprintf("Create %d users with random roles,cluster will be shared from this users", numberOfUsers), func() {
			log.InfoD(fmt.Sprintf("Creating %d users with random roles, cluster will be shared from this users ", numberOfUsers))
			if isClusterShareSupportedForAllRoles {
				roles = []backup.PxBackupRole{backup.ApplicationOwner, backup.InfrastructureOwner, backup.SuperAdmin, backup.ApplicationUser}
			} else {
				roles = []backup.PxBackupRole{backup.SuperAdmin, backup.InfrastructureOwner}
			}
			for i, user := range CreateUsers(numberOfUsers) {
				randomRole := roles[rand.Intn(len(roles))]
				err := backup.AddRoleToUser(user, randomRole, fmt.Sprintf("Adding %v role to %s", randomRole, user))
				log.FailOnError(err, "failed to add role %s to the user %s", randomRole, user)
				log.Infof(fmt.Sprintf("User %s is created with role %v", user, randomRole))
				nonAdminUsers = append(nonAdminUsers, user)
				userNamespaceMap[user] = backedUpNamespaces[i]
				userCtx[user], err = backup.GetNonAdminCtx(user, CommonPassword)
			}
		})

		Step("Registering application cluster for  creating backup object from each user", func() {
			log.InfoD("Registering application cluster for  creating backup object from each user")
			createClusterFromUser := func(user string) {
				err := CreateApplicationClusters(BackupOrgID, "", "", userCtx[user])
				dash.VerifyFatal(err, nil, "Creating source and destination cluster")
				for _, clusterName := range []string{SourceClusterName, DestinationClusterName} {
					userClusterUID, err := Inst().Backup.GetClusterUID(userCtx[user], BackupOrgID, clusterName)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", clusterName))
					if userClusterMap[user] == nil {
						userClusterMap[user] = make(map[string]string)
					}
					userClusterMap[user][clusterName] = userClusterUID
					log.Infof("Updated userClusterMap for user [%s] with cluster [%s] and uid [%s]", user, clusterName, userClusterUID)
				}
			}
			err := TaskHandler(nonAdminUsers[:2], createClusterFromUser, Sequential)
			log.FailOnError(err, "failed to create application cluster from user")
		})

		Step("Register duplicate cluster object for each user for sharing", func() {
			log.InfoD("Register duplicate cluster object for each user for sharing")
			createDuplicateClusterFromUser := func(user string) {
				log.Infof("Creating duplicate cluster object for user %s using sourceClusterConfigPath", user)
				clusterSuffix := fmt.Sprintf("%s-%s-%s", user, "cluster", RandomString(5))
				_, err := CreateDuplicateApplicationClusters(BackupOrgID, userCtx[user], SourceClusterName, 5, clusterSuffix)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating duplicate cluster object for user [%s]", user))
				log.Infof("Creating duplicate cluster object for user %s using DestinationClusterConfigPath", user)
				clusterSuffix = fmt.Sprintf("%s-%s-%s", user, "cluster", RandomString(5))
				_, err = CreateDuplicateApplicationClusters(BackupOrgID, userCtx[user], DestinationClusterName, 5, clusterSuffix)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating duplicate cluster object for user [%s]", user))
			}
			err := TaskHandler(nonAdminUsers[:2], createDuplicateClusterFromUser, Sequential)
			log.FailOnError(err, "failed to create duplicate cluster objects from user")
		})

		Step("Taking manual backups of application from source cluster for each user", func() {
			log.InfoD("Taking manual backups of applications from source cluster for each user")
			createManualBackupsFromUser := func(user string) {
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{userNamespaceMap[user]})
				BackupNamePrefix := fmt.Sprintf("%s-%s-%s", "manual-backup", user, RandomString(5))
				backupNames, err := TakeMultipleBackupsPerDeployment(userCtx[user], BackupOrgID, SourceClusterName, userClusterMap[user][SourceClusterName], numberOfPrimaryBackups, 1, backupLocationName, backupLocationUID, appContextsToBackup, BackupNamePrefix)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupNames))
				masterUserBackupNames = append(masterUserBackupNames, backupNames...)
			}
			err := TaskHandler(nonAdminUsers[:2], createManualBackupsFromUser, Parallel)
			log.FailOnError(err, "failed to create manual backups from user")
		})

		Step("Taking scheduled backup of application from source cluster for each user", func() {
			log.InfoD("Taking scheduled backup of applications from source cluster for each user")
			createScheduleBackupsFromUser := func(user string) {
				scheduleName := fmt.Sprintf("%s-%s-%s", "schedule", user, RandomString(5))
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{userNamespaceMap[user]})
				scheduleBackupName, err := CreateScheduleBackupWithValidation(userCtx[user], scheduleName, SourceClusterName, userClusterMap[user][SourceClusterName], backupLocationName, backupLocationUID, appContextsToBackup, labelSelectors, BackupOrgID, "", "", "", "", schedulePolicyName, schedulePolicyUID)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of schedule backup [%s]", scheduleBackupName))
				mu.Lock()
				defer mu.Unlock()
				scheduleNameUserMap[user] = scheduleName
				err = SuspendAndDeleteSchedule(scheduleName, schedulePolicyName, SourceClusterName, userClusterMap[user][SourceClusterName], BackupOrgID, userCtx[user], false)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Suspend schedule [%s]", scheduleBackupName))
			}
			err := TaskHandler(nonAdminUsers[:2], createScheduleBackupsFromUser, Parallel)
			log.FailOnError(err, "failed to create duplicate cluster objects from user")
		})

		Step("Create Users and Groups for sharing cluster", func() {
			log.InfoD("Creating %d users and %d groups for sharing cluster", numberOfSharedUsers, numberOfGroups)
			var wg sync.WaitGroup
			userChan := make(chan string, numberOfSharedUsers)
			groupChan := make(chan string, numberOfGroups)
			roles = []backup.PxBackupRole{backup.ApplicationOwner, backup.InfrastructureOwner, backup.SuperAdmin, backup.ApplicationUser}
			for i := 1; i <= numberOfSharedUsers; i++ {
				userName := fmt.Sprintf("shareduser%v", i)
				firstName := fmt.Sprintf("FirstName%v", i)
				lastName := fmt.Sprintf("LastName%v", i)
				email := fmt.Sprintf("testuser%v@cnbu.com", i)
				wg.Add(1)
				go func(userName, firstName, lastName, email string) {
					defer GinkgoRecover()
					defer wg.Done()
					err := backup.AddUser(userName, firstName, lastName, email, CommonPassword)
					log.FailOnError(err, "Failed to create user - %s", userName)
					userID, err := backup.FetchIDOfUser(userName)
					log.FailOnError(err, "Failed to fetch user id - %s", userName)
					userIDMap[userName] = userID
					randomRole = roles[rand.Intn(len(roles))]
					log.Infof("Adding role %v to user %s", randomRole, userName)
					err = backup.AddRoleToUser(userName, randomRole, fmt.Sprintf("Adding %v role to %s", randomRole, userName))
					log.FailOnError(err, "failed to add role %s to the user %s", randomRole, userName)
					userChan <- userName
				}(userName, firstName, lastName, email)
			}

			for i := 1; i <= numberOfGroups; i++ {
				groupName := fmt.Sprintf("testGroup%v", i)
				wg.Add(1)
				go func(groupName string) {
					defer GinkgoRecover()
					defer wg.Done()
					err := backup.AddGroup(groupName)
					log.FailOnError(err, "Failed to create group - %v", groupName)
					randomRole = roles[rand.Intn(len(roles))]
					log.Infof("Adding role %v to group %s", randomRole, groupName)
					err = backup.AddRoleToGroup(groupName, randomRole, fmt.Sprintf("Adding %v role to %s", randomRole, groupName))
					log.FailOnError(err, "failed to add role %s to the user %s", randomRole, groupName)
					groupChan <- groupName
				}(groupName)
			}

			wg.Wait()
			close(userChan)
			close(groupChan)

			for userName := range userChan {
				sharedUsers = append(sharedUsers, userName)
				var err error
				userCtx[userName], err = backup.GetNonAdminCtx(userName, CommonPassword)
				log.FailOnError(err, "Fetching user [%s] ctx", userName)
			}

			for groupName := range groupChan {
				sharedGroups = append(sharedGroups, groupName)
			}
		})

		Step("Add Users to Groups", func() {
			log.InfoD("Adding users to groups")
			var wg sync.WaitGroup

			groupAssignments := []struct {
				start, end, groupIndex int
			}{
				{0, 1, 0},
				{1, 3, 1},
				{3, 6, 2},
			}

			for _, assignment := range groupAssignments {
				log.Infof("Adding users %d to %d to group - %v", assignment.start, assignment.end, sharedGroups[assignment.groupIndex])
				for _, user := range sharedUsers[assignment.start:assignment.end] {
					wg.Add(1)
					go func(userName string, groupName string) {
						defer GinkgoRecover()
						defer wg.Done()
						err := backup.AddGroupToUser(userName, groupName)
						log.FailOnError(err, "Failed to assign group to user")
					}(user, sharedGroups[assignment.groupIndex])
				}
				wg.Wait()
			}

			for _, group := range sharedGroups {
				_, err := backup.GetMembersOfGroup(group)
				log.FailOnError(err, "Error fetching members of the group - %v", group)
			}
		})

		Step("Sharing the cluster created by users with newly created groups", func() {
			log.InfoD("Sharing the cluster created by [%d] users with newly created groups [%v]", numberOfUsers, sharedGroups)
			groupShareConfigs := []struct {
				groupName            string
				shareExistingBackups bool
			}{
				{sharedGroups[0], true},
				{sharedGroups[1], false},
				{sharedGroups[2], false},
			}
			for _, groupShareConfig := range groupShareConfigs {
				shareClusterFromUser := func(user string) {
					for _, clusterName := range []string{SourceClusterName, DestinationClusterName} {
						log.Infof("Sharing cluster [%s] with uid [%s] with users from group [%s]", clusterName, userClusterMap[user][clusterName], groupShareConfig.groupName)
						_, err := ShareClusterWithValidation(userCtx[user], clusterName, userClusterMap[user][clusterName], nil, []string{groupShareConfig.groupName}, groupShareConfig.shareExistingBackups)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Sharing cluster [%s] from user [%s] with users from group[%s] ", clusterName, user, groupShareConfig.groupName))
						log.Infof("Creating a map of cluster and shared user")
						usersFromGroup, err := backup.GetMembersOfGroup(groupShareConfig.groupName)
						log.FailOnError(err, "Fetching members of the group")
						for _, userFromGroup := range usersFromGroup {
							if sharedUserClusterMap[userFromGroup] == nil {
								sharedUserClusterMap[userFromGroup] = make(map[string]string)
							}
							sharedUserClusterMap[userFromGroup][clusterName] = userClusterMap[user][clusterName]
						}
					}
				}
				err := TaskHandler(nonAdminUsers[:2], shareClusterFromUser, Parallel)
				log.FailOnError(err, "failed to share cluster object from user each group")
				log.Infof("Updated sharedUserClusterMap with users: [%v] ", sharedUserClusterMap)
			}
		})

		Step("Create manual backup from shared source cluster from shared users", func() {
			log.InfoD("Create manual backup from shared source cluster from shared users")
			clusterSharedUsersList, _ = GetSubsetOfSlice(sharedUsers[:6], 4)
			log.Infof("Creating manual backup from shared source cluster from shared users [%v]", clusterSharedUsersList)
			createManualBackupsFromSharedCluster := func(user string) {
				nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching user [%s] ctx", user)
				BackupName := fmt.Sprintf("%s-%s-%s", "manual-backup", user, RandomString(5))
				userNamespace, _ := GetSubsetOfSlice(backedUpNamespaces, 1)
				backupNamespaceMap[BackupName] = userNamespace[0]
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{userNamespace[0]})
				err = CreateBackupWithValidation(nonAdminCtx, BackupName, SourceClusterName, backupLocationName, backupLocationUID, appContextsToBackup, labelSelectors, BackupOrgID, sharedUserClusterMap[user][SourceClusterName], "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s] from user [%s] on cluster [%s] with uid [%s] ", BackupName, user, SourceClusterName, sharedUserClusterMap[user][SourceClusterName]))
				userBackupsFromSharedClusterMap[user] = BackupName
			}
			err := TaskHandler(clusterSharedUsersList[:2], createManualBackupsFromSharedCluster, Parallel)
			log.FailOnError(err, "failed to create manual backups from user on a shared cluster")
		})

		Step("Create restore for backup created from shared cluster from shared users to destination cluster", func() {
			log.InfoD("Create restore for backup created from shared cluster from shared users to destination cluster")
			createRestoreFromSharedCluster := func(user string) {
				nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching user [%s] ctx", user)
				BackupName := userBackupsFromSharedClusterMap[user]
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{backupNamespaceMap[BackupName]})
				restoreName := fmt.Sprintf("%s-%s-%v", RestoreNamePrefix, BackupName, time.Now().Unix())
				restoreNamespace := fmt.Sprintf("%s-%s", "custom3", backupNamespaceMap[BackupName])
				namespaceMapping := map[string]string{backupNamespaceMap[BackupName]: restoreNamespace}
				restoredNamespaces = append(restoredNamespaces, restoreNamespace)
				log.Infof("Creating restore [%s] for backup [%s] of namespace [%s] from user [%s]", restoreName, BackupName, backupNamespaceMap[BackupName], user)
				err = CreateRestoreWithValidation(nonAdminCtx, restoreName, BackupName, namespaceMapping, make(map[string]string), DestinationClusterName, sharedUserClusterMap[user][DestinationClusterName], BackupOrgID, appContextsToBackup)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s] for backup [%s] of namespace [%s] from user [%s]", restoreName, BackupName, backupNamespaceMap[BackupName], user))
				userRestoreFromSharedClusterMap[user] = restoreName
			}
			err := TaskHandler(clusterSharedUsersList[:2], createRestoreFromSharedCluster, Sequential)
			log.FailOnError(err, "failed to create restore from user on a shared cluster")
		})

		Step("Create scheduled backup from shared source cluster from shared users ", func() {
			log.InfoD("Create scheduled backup from shared source cluster from shared users")
			createScheduleBackupsFromUser := func(user string) {
				nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching user [%s] ctx", user)
				userNamespace := backedUpNamespaces[rand.Intn(len(backedUpNamespaces))]
				scheduleName := fmt.Sprintf("%s-%s-%s", "schedule", user, RandomString(5))
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{userNamespace})
				scheduleBackupName, err := CreateScheduleBackupWithValidation(nonAdminCtx, scheduleName, SourceClusterName, sharedUserClusterMap[user][SourceClusterName], backupLocationName, backupLocationUID, appContextsToBackup, labelSelectors, BackupOrgID, "", "", "", "", schedulePolicyName, schedulePolicyUID)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of schedule backup [%s] form user [%s]", scheduleName, user))
				userScheduleFromSharedClusterMap[user] = scheduleName
				err = SuspendBackupSchedule(scheduleName, schedulePolicyName, BackupOrgID, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Suspend schedule [%s]", scheduleBackupName))
			}
			err := TaskHandler(clusterSharedUsersList[:3], createScheduleBackupsFromUser, Parallel)
			log.FailOnError(err, "failed to create schedule backups from user on a shared cluster")
		})

		Step("Verify admin has access to object created by cluster shared user", func() {
			log.InfoD("Verify admin has access to object created by cluster shared user")
			adminBackups, err := GetAllBackupsAdmin()
			dash.VerifyFatal(err, nil, "Verifying fetching of all backups")
			validateSharedUserBackupsFromAdmin := func(user string) {
				if !IsPresent(adminBackups, userBackupsFromSharedClusterMap[user]) {
					err := fmt.Errorf("backup %s is not listed in admin backup names [%s]", userBackupsFromSharedClusterMap[user], adminBackups)
					log.FailOnError(fmt.Errorf(""), err.Error())
				} else {
					log.Infof("Backup %s is listed in admin backup names [%s]", userBackupsFromSharedClusterMap[user], adminBackups)
				}
			}
			err = TaskHandler(clusterSharedUsersList[:2], validateSharedUserBackupsFromAdmin, Parallel)
			log.FailOnError(err, "failed to validate backup object from admin")
			adminRestores, err := GetAllRestoresAdmin()
			dash.VerifyFatal(err, nil, "Verifying fetching of all restores")
			validateSharedUserRestoresFromAdmin := func(user string) {
				if !IsPresent(adminRestores, userRestoreFromSharedClusterMap[user]) {
					err := fmt.Errorf("restore %s is not listed  in admin restore names [%v]", userRestoreFromSharedClusterMap[user], adminRestores)
					log.FailOnError(fmt.Errorf(""), err.Error())
				} else {
					log.Infof("Restore %s is listed in admin restore names [%v]", userRestoreFromSharedClusterMap[user], adminRestores)
				}
			}
			err = TaskHandler(clusterSharedUsersList[:2], validateSharedUserRestoresFromAdmin, Parallel)
			log.FailOnError(err, "failed to validate restore object from admin")

			adminSchedules, err := GetAllBackupSchedulesAdmin()
			dash.VerifyFatal(err, nil, "Verifying fetching of all schedules")
			validateSharedUserSchedulesFromAdmin := func(user string) {
				if !IsPresent(adminSchedules, userScheduleFromSharedClusterMap[user]) {
					err := fmt.Errorf("schedule %s is not listed in admin schedule names [%v] ", userScheduleFromSharedClusterMap[user], adminSchedules)
					log.FailOnError(fmt.Errorf(""), err.Error())
				} else {
					log.Infof("Schedule %s is listed in admin schedule names [%v]", userScheduleFromSharedClusterMap[user], adminSchedules)
				}
			}
			err = TaskHandler(clusterSharedUsersList[:3], validateSharedUserSchedulesFromAdmin, Parallel)
			log.FailOnError(err, "failed to validate schedule object from admin")

		})

		Step("Verify cluster shared user doesnt have permission to unshare and delete the cluster", func() {
			log.InfoD("Verify cluster shared user doesnt have permission to unshare and delete the cluster")
			groupShareConfigs := []struct {
				groupName            string
				shareExistingBackups bool
			}{
				{sharedGroups[0], true},
				{sharedGroups[1], false},
				{sharedGroups[2], false},
			}
			for _, groupShareConfig := range groupShareConfigs {
				userFromGroup, err := backup.GetMembersOfGroup(groupShareConfig.groupName)
				log.FailOnError(err, "Fetching members of the group")
				validateUnshareClusterFromSharedUser := func(user string) {
					nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
					log.FailOnError(err, "Fetching user [%s] ctx", user)
					isUserAdmin, err := IsAdminCtx(nonAdminCtx)
					log.FailOnError(err, "Fetching if user is admin")
					if !isUserAdmin {
						for _, clusterName := range []string{SourceClusterName, DestinationClusterName} {
							_, err := UnShareCluster(nonAdminCtx, clusterName, sharedUserClusterMap[user][clusterName], nil, []string{groupShareConfig.groupName})
							log.Infof("Unsharing cluster [%s] from user [%s] with users from group[%s] ", clusterName, user, groupShareConfig.groupName)
							if err != nil {
								dash.VerifyFatal(strings.Contains(err.Error(), clusterSharePermissionError), true, fmt.Sprintf("Verifying user [%s] cant unshare the cluster [%s]", user, SourceClusterName))
							}
						}
					} else {
						log.Infof("User [%s] is admin, so skipping the unshare cluster", user)
					}
				}
				err = TaskHandler(userFromGroup, validateUnshareClusterFromSharedUser, Sequential)
				log.FailOnError(err, "failed to validate unshare cluster object from shared user")

				validateDeleteClusterFromSharedUser := func(user string) {
					nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
					log.FailOnError(err, "Fetching user [%s] ctx", user)
					isUserAdmin, err := IsAdminCtx(nonAdminCtx)
					log.FailOnError(err, "Fetching if user is admin")
					if !isUserAdmin {
						for _, clusterName := range []string{SourceClusterName, DestinationClusterName} {
							err = DeleteClusterWithoutScheduleDelete(clusterName, sharedUserClusterMap[user][clusterName], BackupOrgID, nonAdminCtx, false)
							dash.VerifyFatal(strings.Contains(err.Error(), clusterSharePermissionError), true, fmt.Sprintf("Verifying user [%s] cant delete the cluster [%s]", user, clusterName))
						}
					} else {
						log.Infof("User [%s] is admin, so skipping the delete cluster", user)
					}
				}
				err = TaskHandler(userFromGroup, validateDeleteClusterFromSharedUser, Sequential)
				log.FailOnError(err, "failed to validate delete cluster object from shared user")
			}
		})

		Step("Verify deleting the backup,restore,schedule created from shared cluster", func() {
			log.InfoD("Verify deleting the backup,restore,schedule created from shared cluster")
			validateDeleteBackupFromSharedUser := func(user string) {
				nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching user [%s] ctx", user)
				log.Infof("Deleting backup [%s] from user [%s]", userBackupsFromSharedClusterMap[user], user)
				backupUid, err := Inst().Backup.GetBackupUID(nonAdminCtx, userBackupsFromSharedClusterMap[user], BackupOrgID)
				log.FailOnError(err, "Unable to fetch backup UID")
				_, err = DeleteBackup(userBackupsFromSharedClusterMap[user], backupUid, BackupOrgID, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting backup [%s]", userBackupsFromSharedClusterMap[user]))
			}
			err := TaskHandler(clusterSharedUsersList[:2], validateDeleteBackupFromSharedUser, Parallel)
			log.FailOnError(err, "failed to validate delete backup object from shared user")
			validateDeleteRestoreFromSharedUser := func(user string) {
				nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching user [%s] ctx", user)
				log.Infof("Deleting restore [%s] from user [%s]", userRestoreFromSharedClusterMap[user], user)
				err = DeleteRestore(userRestoreFromSharedClusterMap[user], BackupOrgID, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying user [%s] can delete the restore [%s]", user, userRestoreFromSharedClusterMap[user]))
			}
			err = TaskHandler(clusterSharedUsersList[:2], validateDeleteRestoreFromSharedUser, Parallel)
			log.FailOnError(err, "failed to validate delete restore object from shared user")

			validateDeleteScheduleFromSharedUser := func(user string) {
				nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching user [%s] ctx", user)
				log.Infof("Deleting schedule [%s] from user [%s]", userScheduleFromSharedClusterMap[user], user)
				scheduleUid, err := GetScheduleUID(userScheduleFromSharedClusterMap[user], BackupOrgID, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching schedule uid for schedule [%s]", userScheduleFromSharedClusterMap[user]))
				err = DeleteScheduleWithUIDAndWait(userScheduleFromSharedClusterMap[user], scheduleUid, SourceClusterName, sharedUserClusterMap[user][SourceClusterName], BackupOrgID, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying user [%s] can delete the schedule [%s]", user, userScheduleFromSharedClusterMap[user]))
			}
			err = TaskHandler(clusterSharedUsersList[:3], validateDeleteScheduleFromSharedUser, Parallel)
			log.FailOnError(err, "failed to validate delete schedule object from shared user")

		})

		Step("Validate the unshare of cluster from each shared users", func() {
			log.InfoD("Validate the unshare of cluster from each shared users")
			for _, sharedGroup := range sharedGroups {
				unShareClusterFromUser := func(user string) {
					for _, clusterName := range []string{SourceClusterName, DestinationClusterName} {
						nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
						log.FailOnError(err, "Fetching user [%s] ctx", user)
						_, err = UnShareClusterWithValidation(nonAdminCtx, clusterName, userClusterMap[user][clusterName], nil, []string{sharedGroup})
						dash.VerifyFatal(err, nil, fmt.Sprintf("UnSharing cluster [%s] from user [%s] with users from group[%s] ", clusterName, user, sharedGroup))
					}
				}
				err := TaskHandler(nonAdminUsers[:2], unShareClusterFromUser, Parallel)
				log.FailOnError(err, "failed to unshare cluster object from user")
			}
		})

		log.InfoD("Second scenario of cluster share with large number of users")

		Step("Register Application cluster from larger number of users", func() {
			log.InfoD("Register Application cluster from larger number of users")
			createClusterFromUser := func(user string) {
				defer GinkgoRecover()
				nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching user [%s] ctx", user)
				err = CreateApplicationClusters(BackupOrgID, "", "", nonAdminCtx)
				dash.VerifyFatal(err, nil, "Creating source and destination cluster")
				for _, clusterName := range []string{SourceClusterName, DestinationClusterName} {
					userClusterUID, err := Inst().Backup.GetClusterUID(nonAdminCtx, BackupOrgID, clusterName)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", clusterName))
					if iter2ClusterUserMap[user] == nil {
						iter2ClusterUserMap[user] = make(map[string]string)
					}
					iter2ClusterUserMap[user][clusterName] = userClusterUID
					log.Infof("Updated userClusterMap for user [%s] with cluster [%s] and uid [%s]", user, clusterName, userClusterUID)
				}
			}
			err := TaskHandler(nonAdminUsers[2:8], createClusterFromUser, Sequential)
			log.FailOnError(err, "failed to create application cluster from user")
		})

		Step("Create few manual backups for each user", func() {
			log.InfoD("Create few manual backups for each user")
			createManualBackupsFromUser := func(user string) {
				defer GinkgoRecover()
				nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching user [%s] ctx", user)
				userNamespace := backedUpNamespaces[rand.Intn(len(backedUpNamespaces))]
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{userNamespace})
				BackupNamePrefix := fmt.Sprintf("%s-%s-%s", "manual-backup", user, RandomString(5))
				backupNames, err := TakeMultipleBackupsPerDeployment(nonAdminCtx, BackupOrgID, SourceClusterName, iter2ClusterUserMap[user][SourceClusterName], 1, 3, backupLocationName, backupLocationUID, appContextsToBackup, BackupNamePrefix)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupNames))
			}
			err := TaskHandler(nonAdminUsers[2:8], createManualBackupsFromUser, Parallel)
			log.FailOnError(err, "failed to create manual backups from user")
		})

		Step("Create few scheduled backups for each user", func() {
			log.InfoD("Create few scheduled backups for each user")
			createScheduleBackupsFromUser := func(user string) {
				defer GinkgoRecover()
				nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching user [%s] ctx", user)
				userNamespace := backedUpNamespaces[rand.Intn(len(backedUpNamespaces))]
				scheduleName := fmt.Sprintf("%s-%s-%s", "schedule", user, RandomString(5))
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{userNamespace})
				scheduleBackupName, err := CreateScheduleBackupWithValidation(nonAdminCtx, scheduleName, SourceClusterName, iter2ClusterUserMap[user][SourceClusterName], backupLocationName, backupLocationUID, appContextsToBackup, labelSelectors, BackupOrgID, "", "", "", "", schedulePolicyName, schedulePolicyUID)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of schedule backup [%s]", scheduleBackupName))
				err = SuspendBackupSchedule(scheduleName, schedulePolicyName, BackupOrgID, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Suspend schedule [%s]", scheduleName))
			}
			err := TaskHandler(nonAdminUsers[2:8], createScheduleBackupsFromUser, Parallel)
			log.FailOnError(err, "failed to create schedule backups from user")
		})

		Step("For few user make cluster inactive by updating kubeConfig", func() {
			log.InfoD("For few user make cluster inactive by updating kubeConfig")
			inActiveClusterUsers, _ = GetSubsetOfSlice(nonAdminUsers[2:8], 3)
			makeClusterInactive := func(user string) {
				defer GinkgoRecover()
				nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching user [%s] ctx", user)
				for _, clusterName := range []string{SourceClusterName, DestinationClusterName} {
					clusterInspectRequest := &api.ClusterInspectRequest{
						Name:           clusterName,
						Uid:            iter2ClusterUserMap[user][clusterName],
						OrgId:          BackupOrgID,
						IncludeSecrets: true,
					}
					clusterInspectResp, err := Inst().Backup.InspectCluster(nonAdminCtx, clusterInspectRequest)
					log.FailOnError(err, "failed to inspect cluster %s with uid %s", clusterName, iter2ClusterUserMap[user][clusterName])
					clusterUpdateRequest := &api.ClusterUpdateRequest{
						CreateMetadata: &api.CreateMetadata{
							Name:  clusterName,
							Uid:   iter2ClusterUserMap[user][clusterName],
							OrgId: BackupOrgID,
						},
						Kubeconfig:            invalidKubeConfig,
						CloudCredential:       clusterInspectResp.GetCluster().GetCloudCredential(),
						CloudCredentialRef:    clusterInspectResp.GetCluster().GetCloudCredentialRef(),
						PlatformCredentialRef: clusterInspectResp.GetCluster().GetPlatformCredentialRef(),
					}
					log.Infof("Updating kubeconfig for cluster %s from user [%s]", clusterName, user)
					_, err = Inst().Backup.UpdateCluster(nonAdminCtx, clusterUpdateRequest)
					if err != nil {
						clusterStatus, statusError := Inst().Backup.GetClusterStatus(BackupOrgID, clusterName, nonAdminCtx)
						log.FailOnError(statusError, "failed to get cluster status %s", clusterName)
						if clusterStatus != api.ClusterInfo_StatusInfo_Failed {
							log.FailOnError(err, "failed to make cluster %s inactive. Expected status %v but got %v", clusterName, api.ClusterInfo_StatusInfo_Failed, clusterStatus)
						}
					} else {
						err = fmt.Errorf("failed to make cluster %s inactive. Expected error not be nil", clusterName)
						log.FailOnError(fmt.Errorf(""), err.Error())
					}
				}
			}
			err := TaskHandler(inActiveClusterUsers, makeClusterInactive, Parallel)
			log.FailOnError(err, "failed to make cluster inactive")
		})

		Step("Share the cluster with another set of users simultaneously ", func() {
			log.InfoD("Share the cluster with another set of users simultaneously")
			shareConfigs := []struct {
				users                []string
				shareExistingBackups bool
			}{
				{sharedUsers[6:12], true},
				{sharedUsers[12:18], false},
			}

			for _, shareConfig := range shareConfigs {
				shareClusterFromUser := func(user string) {
					defer GinkgoRecover()
					nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
					log.FailOnError(err, "Fetching user [%s] ctx", user)
					_, err = ShareCluster(nonAdminCtx, SourceClusterName, iter2ClusterUserMap[user][SourceClusterName], shareConfig.users, nil, shareConfig.shareExistingBackups)
					log.FailOnError(err, "failed to share cluster from user")
				}
				err := TaskHandler(nonAdminUsers[2:8], shareClusterFromUser, Parallel)
				log.FailOnError(err, "failed to share cluster from user")
			}

			log.Infof("Validate the cluster shared")
			for _, shareConfig := range shareConfigs {
				validateSharedCluster := func(user string) {
					defer GinkgoRecover()
					nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
					log.FailOnError(err, "Fetching user [%s] ctx", user)
					err = ValidateShareCluster(nonAdminCtx, SourceClusterName, iter2ClusterUserMap[user][SourceClusterName], shareConfig.users, nil)
					log.FailOnError(err, "failed to validate shared cluster")
				}
				err := TaskHandler(nonAdminUsers[2:8], validateSharedCluster, Parallel)
				log.FailOnError(err, "failed to validate shared cluster")
			}
		})

		Step("Delete few user in which cluster is shared with", func() {
			log.InfoD("Delete few user in which cluster is shared with")
			usersToBeDeleted, err := GetSubsetOfSlice(sharedUsers[6:18], 6)
			log.FailOnError(err, "failed to get subset of shared users")
			for _, user := range usersToBeDeleted {
				err := backup.DeleteUser(user)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying user [%s] deletion", user))
			}
		})

		Step("unshare the cluster from shared users", func() {
			log.InfoD("unshare the cluster from shared users")
			shareConfigs := []struct {
				users                []string
				shareExistingBackups bool
			}{
				{sharedUsers[6:12], true},
				{sharedUsers[12:18], false},
			}
			log.Infof("The share config for users [%v]", shareConfigs)
			for _, shareConfig := range shareConfigs {
				unshareClusterFromUser := func(user string) {
					userIDs := make([]string, 0)
					defer GinkgoRecover()
					nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
					log.FailOnError(err, "Fetching user [%s] ctx", user)
					for _, sharedUser := range shareConfig.users {
						userIDs = append(userIDs, userIDMap[sharedUser])
					}
					unshareClusterRequest := &api.UnShareClusterRequest{
						OrgId: BackupOrgID,
						ClusterRef: &api.ObjectRef{
							Name: SourceClusterName,
							Uid:  iter2ClusterUserMap[user][SourceClusterName],
						},
						Users:  userIDs,
						Groups: nil,
					}
					_, err = Inst().Backup.UnShareCluster(nonAdminCtx, unshareClusterRequest)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Unsharing cluster [%s] with uid [%s] from user [%s] with users [%v] ", SourceClusterName, iter2ClusterUserMap[user][SourceClusterName], user, shareConfig.users))

				}
				err := TaskHandler(nonAdminUsers[2:8], unshareClusterFromUser, Parallel)
				log.FailOnError(err, "failed to unshare cluster from user")
			}
			for _, shareConfig := range shareConfigs {
				validateUnshareCluster := func(user string) {
					defer GinkgoRecover()
					nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
					log.FailOnError(err, "Fetching user [%s] ctx", user)
					shareConfig.users = RemoveStringItemFromSlice(shareConfig.users, usersToBeDeleted)
					err = ValidateUnShareCluster(nonAdminCtx, SourceClusterName, iter2ClusterUserMap[user][SourceClusterName], shareConfig.users, nil)
					dash.VerifyFatal(err, nil, fmt.Sprintf("validate unshared cluster from user [%s]", user))
				}
				err := TaskHandler(nonAdminUsers[2:8], validateUnshareCluster, Parallel)
				log.FailOnError(err, "failed to validate unshared cluster")
			}
		})

		Step("Validate app-user or app-admin can't share the cluster if they are not owning the cloud credential", func() {
			log.InfoD("Validate app-user or app-admin can't share the cluster if they are not owning the cloud credential")
			if !isClusterShareSupportedForAllRoles {
				log.Infof("Creating user with app-user role and app-admin role")
				for i := 1; i <= numOfUnsupportedUsers; i++ {
					userName := fmt.Sprintf("unsupportedUser%v", i)
					firstName := fmt.Sprintf("FirstName%v", i)
					lastName := fmt.Sprintf("LastName%v", i)
					email := fmt.Sprintf("unsupportedUser%v@cnbu.com", i)
					err := backup.AddUser(userName, firstName, lastName, email, CommonPassword)
					log.FailOnError(err, "Failed to create user - %s", userName)
					unsupportedUsers = append(unsupportedUsers, userName)
				}

				for i, role := range []backup.PxBackupRole{backup.ApplicationUser, backup.ApplicationOwner} {
					err := backup.AddRoleToUser(unsupportedUsers[i], role, fmt.Sprintf("Adding role [%s] to user [%s]", role, unsupportedUsers[i]))
					dash.VerifyFatal(err, nil, fmt.Sprintf("Adding role [%s] to user [%s]", role, unsupportedUsers[i]))
				}

				createClusterFromUser := func(user string) {
					userCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
					log.FailOnError(err, "Fetching user [%s] ctx", user)
					err = CreateApplicationClusters(BackupOrgID, "", "", userCtx)
					dash.VerifyFatal(err, nil, "Creating source and destination cluster")
					for _, clusterName := range []string{SourceClusterName, DestinationClusterName} {
						userClusterUID, err := Inst().Backup.GetClusterUID(userCtx, BackupOrgID, clusterName)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", clusterName))
						if userClusterMap[user] == nil {
							userClusterMap[user] = make(map[string]string)
						}
						userClusterMap[user][clusterName] = userClusterUID
						log.Infof("Updated userClusterMap for user [%s] with cluster [%s] and uid [%s]", user, clusterName, userClusterUID)
					}
				}
				err := TaskHandler(unsupportedUsers, createClusterFromUser, Sequential)
				log.FailOnError(err, "failed to create application cluster from user")

				log.Infof("Sharing the cluster from user [%s] with users [%s]", unsupportedUsers[0], sharedUsers[18:20])
				nonAdminCtx, err := backup.GetNonAdminCtx(unsupportedUsers[0], CommonPassword)
				log.FailOnError(err, "Fetching user [%s] ctx", unsupportedUsers[0])
				_, err = ShareCluster(nonAdminCtx, SourceClusterName, userClusterMap[unsupportedUsers[0]][SourceClusterName], sharedUsers[18:20], nil, true)
				if err != nil {
					log.Infof("The actual error is [%s]", err.Error())
					dash.VerifyFatal(strings.Contains(err.Error(), clusterSharePermissionError), true, fmt.Sprintf("Verifying user [%s] cant share the cluster [%s]", unsupportedUsers[0], SourceClusterName))
				} else {
					log.FailOnError(err, fmt.Sprintf("Expected error, but got none for user [%s] attempting to share cluster [%s]", unsupportedUsers[0], SourceClusterName))
				}
				log.Infof("Sharing the cluster from user [%s] with groups [%s]", unsupportedUsers[1], sharedGroups[0:2])
				nonAdminCtx, err = backup.GetNonAdminCtx(unsupportedUsers[1], CommonPassword)
				log.FailOnError(err, "Fetching user [%s] ctx", unsupportedUsers[1])
				_, err = ShareCluster(nonAdminCtx, SourceClusterName, userClusterMap[unsupportedUsers[1]][SourceClusterName], nil, sharedGroups[0:2], true)
				if err != nil {
					log.Infof("The actual error is [%s]", err.Error())
					dash.VerifyFatal(strings.Contains(err.Error(), clusterSharePermissionError), true, fmt.Sprintf("Verifying user [%s] cant share the cluster [%s]", unsupportedUsers[1], SourceClusterName))
				} else {
					log.FailOnError(err, fmt.Sprintf("Expected error, but got none for user [%s] attempting to share cluster [%s]", unsupportedUsers[1], SourceClusterName))
				}
			} else {
				log.Infof("Skipping the Validation, As the cluster is of the provider type -[%s], the cluster share operation is supported for all roles", GetClusterProvider())
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
		log.Infof("Revert the cluster in inactive state to active state for backup delete")
		updateCluster := func(user string) {
			userCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", user)
			for _, clusterName := range []string{SourceClusterName, DestinationClusterName} {
				if clusterName == SourceClusterName {
					configPath, _ = GetSourceClusterConfigPath()
				} else if clusterName == DestinationClusterName {
					configPath, _ = GetDestinationClusterConfigPath()
				}
				log.Infof("Updating cluster %s with uid %s", clusterName, iter2ClusterUserMap[user][clusterName])
				_, err = UpdateClusterWithKubeConfig(clusterName, iter2ClusterUserMap[user][clusterName], configPath, userCtx)
				log.FailOnError(err, "failed to update cluster %s with uid %s", clusterName, iter2ClusterUserMap[user][clusterName])
			}
		}
		err := TaskHandler(inActiveClusterUsers, updateCluster, Parallel)
		log.FailOnError(err, "failed to update cluster")
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "failed to get admin context")
		log.InfoD("Deleting the backups")
		allBackups, err := GetAllBackupsAdmin()
		dash.VerifySafely(err, nil, "Verifying fetching of all backups")
		for _, bkp := range allBackups {
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
		log.Infof("Cleanup any clusters, since it can have inactive and backup schedules active")
		err = DeleteAllAdminClusters()
		dash.VerifySafely(err, nil, "Verifying deletion of all admin clusters")
		err = CleanupAllUserAndGroups()
		dash.VerifySafely(err, nil, "Verifying cleanup all user and groups")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		log.Info("Destroying scheduled apps on source cluster")
		DestroyApps(scheduledAppContexts, opts)
		CleanupCloudSettingsAndClusters(backupLocationMap, adminCloudCredName, adminCloudCredentialUID, ctx)
	})
})

// This testcase verifies super admin role permissions for local user
var _ = Describe("{BackupSuperAdminRoleForLocalUser}", Label(TestCaseLabelsMap[BackupSuperAdminRoleForLocalUser]...), func() {
	var (
		scheduledAppContexts             []*scheduler.Context
		backupLocationMap                map[string]string
		userNamespaceMap                 map[string]string
		labelSelectors                   map[string]string
		schedulePolicyName               string
		schedulePolicyUID                string
		adminCloudCredName               string
		adminCloudCredentialUID          string
		schedulePolicyInterval           = int64(15)
		providers                        []string
		masterUserBackupNames            []string
		backedUpNamespaces               []string
		numberOfUsers                    int
		nonAdminUsers                    []string
		superAdminUsers                  []string
		userBackupsMap                   map[string][]string
		userScheduleFromSharedClusterMap map[string]string
		userClusterMap                   map[string]map[string]string
		userRoleMap                      map[string]backup.PxBackupRole
		userCloudCredNameMap             map[string]string
		userCloudCredUidMap              map[string]string
		userBkpLocationNameMap           map[string]string
		userBkpLocationUidMap            map[string]string
		numOfSuperAdminUsers             int
		wg                               sync.WaitGroup
		mu                               sync.Mutex
		userTobeDemoted                  string
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("BackupSuperAdminRoleForLocalUser", "Validates super admin role permissions for local user", nil, 301144, Ak, Q3FY25)
		providers = GetBackupProviders()
		// keeping the user  and namespace count same , so that each user will take backup of each namespace
		numberOfUsers, _ = strconv.Atoi(GetEnv(UsersToBeCreated, "10"))
		numOfSuperAdminUsers = 5
		numOfNamespace := numberOfUsers + numOfSuperAdminUsers

		backupLocationMap = make(map[string]string)
		userNamespaceMap = make(map[string]string)
		userBackupsMap = make(map[string][]string)
		userScheduleFromSharedClusterMap = make(map[string]string)
		userClusterMap = make(map[string]map[string]string)
		userRoleMap = make(map[string]backup.PxBackupRole)
		userCloudCredNameMap = make(map[string]string)
		userCloudCredUidMap = make(map[string]string)
		userBkpLocationNameMap = make(map[string]string)
		userBkpLocationUidMap = make(map[string]string)
		labelSelectors = make(map[string]string)

		log.InfoD("scheduling applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		// Schedule applications
		for i := 0; i < numOfNamespace; i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, appCtx := range appContexts {
				appCtx.ReadinessTimeout = AppReadinessTimeout
				scheduledAppContexts = append(scheduledAppContexts, appCtx)
				appNamespace := appCtx.ScheduleOptions.Namespace
				backedUpNamespaces = append(backedUpNamespaces, appNamespace)
			}
		}
	})

	It("Validates super admin role permissions for local user", func() {
		defer func() {
			log.InfoD("switching to default context")
			err := SetClusterContext("")
			log.FailOnError(err, "failed to SetClusterContext to default cluster")
		}()

		Step("Validating applications", func() {
			log.InfoD("Validating applications")
			ValidateApplications(scheduledAppContexts)
		})

		Step(fmt.Sprintf("Create %d users with random roles (app.user, app.admin, infra.admin)", numberOfUsers), func() {
			log.InfoD(fmt.Sprintf("Creating %d users with random roles (app.user, app.admin, infra.admin)", numberOfUsers))
			roles := [3]backup.PxBackupRole{backup.ApplicationOwner, backup.InfrastructureOwner, backup.ApplicationUser}
			for _, user := range CreateUsers(numberOfUsers) {
				randomRole := roles[rand.Intn(len(roles))]
				err := backup.AddRoleToUser(user, randomRole, fmt.Sprintf("Adding %v role to %s", randomRole, user))
				log.FailOnError(err, "failed to add role %s to the user %s", randomRole, user)
				log.Infof(fmt.Sprintf("User %s is created with role %v", user, randomRole))
				nonAdminUsers = append(nonAdminUsers, user)
				userRoleMap[user] = randomRole
			}
			for i, nonAdminUser := range nonAdminUsers {
				userNamespaceMap[nonAdminUser] = backedUpNamespaces[i%10]
			}
		})

		Step("Create another set of users to assign super admin role", func() {
			log.InfoD("Create another set of users to assign super admin role")
			roles := [1]backup.PxBackupRole{backup.SuperAdmin}
			for i, user := range CreateUsers(numOfSuperAdminUsers) {
				randomRole := roles[0]
				err := backup.AddRoleToUser(user, randomRole, fmt.Sprintf("Adding %v role to %s", randomRole, user))
				log.FailOnError(err, "failed to add role %s to the user %s", randomRole, user)
				log.Infof(fmt.Sprintf("User %s is created with role %v", user, randomRole))
				superAdminUsers = append(superAdminUsers, user)
				userNamespaceMap[user] = backedUpNamespaces[numberOfUsers+i]
			}
		})

		Step("Creating backup location and cloud setting for each user based on the role", func() {
			log.InfoD("Creating backup location and cloud setting for each user based on the role")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			AddBackupLocationForUser := func(user string) {
				defer GinkgoRecover()
				nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching px-central-admin ctx")
				AddCloudCredAndBackupLocation := func(provider string) {
					defer GinkgoRecover()
					if userRoleMap[user] != backup.InfrastructureOwner {
						adminCloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, RandomString(5))
						adminCloudCredentialUID = uuid.New()
						userCloudCredNameMap[user] = adminCloudCredName
						userCloudCredUidMap[user] = adminCloudCredentialUID
						err := CreateCloudCredential(provider, userCloudCredNameMap[user], userCloudCredUidMap[user], BackupOrgID, ctx)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", userCloudCredNameMap[user], BackupOrgID, provider))
						if provider != drivers.ProviderNfs {
							log.Infof("Update CloudAccount - %s ownership as public", adminCloudCredName)
							err = AddCloudCredentialOwnership(adminCloudCredName, adminCloudCredentialUID, []string{user}, nil, Read, Invalid, ctx, BackupOrgID)
							dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying updation of owbership for CloudCredential- %s", adminCloudCredName))
						}
					} else {
						cloudCredName := fmt.Sprintf("%s-%s-%v", "cred", provider, RandomString(5))
						cloudCredentialUID := uuid.New()
						userCloudCredNameMap[user] = cloudCredName
						userCloudCredUidMap[user] = cloudCredentialUID
						err := CreateCloudCredential(provider, userCloudCredNameMap[user], userCloudCredUidMap[user], BackupOrgID, nonAdminCtx)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", userCloudCredNameMap[user], BackupOrgID, provider))
					}

					if userRoleMap[user] != backup.ApplicationUser {
						userBkpLocationName := fmt.Sprintf("%s-%s-bl-%v", provider, getGlobalBucketName(provider), RandomString(5))
						userBkpLocationUid := uuid.New()
						userBkpLocationNameMap[user] = userBkpLocationName
						userBkpLocationUidMap[user] = userBkpLocationUid
						backupLocationMap[userBkpLocationUid] = userBkpLocationName
						err = CreateBackupLocationWithContext(provider, userBkpLocationNameMap[user], userBkpLocationUidMap[user], userCloudCredNameMap[user], userCloudCredUidMap[user], getGlobalBucketName(provider), BackupOrgID, "", nonAdminCtx, true)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location [%s] , with uid [%s],for user [%s] ", userBkpLocationName, userBkpLocationUid, user))

					} else {
						backupLocationName := fmt.Sprintf("%s-%s-%v", provider, getGlobalBucketName(provider), RandomString(5))
						backupLocationUID := uuid.New()
						userBkpLocationNameMap[user] = backupLocationName
						userBkpLocationUidMap[user] = backupLocationUID
						backupLocationMap[backupLocationUID] = backupLocationName
						err = CreateBackupLocation(provider, backupLocationName, backupLocationUID, userCloudCredNameMap[user], userCloudCredUidMap[user], getGlobalBucketName(provider), BackupOrgID, "", true)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location [%s] , with uid [%s],for user [%s] ", backupLocationName, backupLocationUID, user))
						err = AddBackupLocationOwnership(backupLocationName, backupLocationUID, []string{user}, nil, Read, Invalid, ctx)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying updation of ownership for backuplocation - %s", backupLocationName))
					}
				}
				err = TaskHandler(providers, AddCloudCredAndBackupLocation, Parallel)
				log.FailOnError(err, "failed to create cloud cred and backup location")
			}
			err = TaskHandler(append(nonAdminUsers, superAdminUsers...), AddBackupLocationForUser, Sequential)
			log.FailOnError(err, "failed to create backup location and cloud setting as user")
		})

		Step("Create schedule policy from admin and make it public", func() {
			log.InfoD("Create schedule policy from admin and make it public")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			schedulePolicyName = fmt.Sprintf("%s-%v", "periodic-schedule-policy", RandomString(5))
			schedulePolicyUID = uuid.New()
			err = CreateBackupScheduleIntervalPolicy(5, schedulePolicyInterval, 5, schedulePolicyName, schedulePolicyUID, BackupOrgID, ctx, false, false)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of schedule policy %s", schedulePolicyName))
			log.InfoD("Update SchedulePolicy - %s ownership as public ", schedulePolicyName)
			err = AddSchedulePolicyOwnership(schedulePolicyName, schedulePolicyUID, nil, nil, Invalid, Read, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying updation of ownership for schedulepolicy - %s", schedulePolicyName))
		})

		Step("Registering cluster for backup object from each user", func() {
			log.InfoD("Registering cluster for backup  object from each user")
			createClusterFromUser := func(user string) {
				defer GinkgoRecover()
				nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "failed to fetch user [%s] ctx", user)
				err = CreateApplicationClusters(BackupOrgID, "", "", nonAdminCtx)
				dash.VerifyFatal(err, nil, "Creating source and destination cluster")
				for _, clusterName := range []string{SourceClusterName, DestinationClusterName} {
					userClusterUID, err := Inst().Backup.GetClusterUID(nonAdminCtx, BackupOrgID, clusterName)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", clusterName))
					if userClusterMap[user] == nil {
						userClusterMap[user] = make(map[string]string)
					}
					userClusterMap[user][clusterName] = userClusterUID
					log.Infof("Updated userClusterMap for user [%s] with cluster [%s] and uid [%s]", user, clusterName, userClusterUID)
				}
			}
			err := TaskHandler(append(nonAdminUsers, superAdminUsers...), createClusterFromUser, Sequential)
			log.FailOnError(err, "failed to create application cluster from user")
		})

		Step("Taking manual backups of application from source cluster for each user", func() {
			log.InfoD("Taking manual backups of applications from source cluster for each user")
			createManualBackupsFromUser := func(user string) {
				defer GinkgoRecover()
				nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "failed to fetch user [%s] ctx", user)
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{userNamespaceMap[user]})
				BackupNamePrefix := fmt.Sprintf("%s-%s-%s", "manual-backup", user, RandomString(5))
				backupNames, err := TakeMultipleBackupsPerDeployment(nonAdminCtx, BackupOrgID, SourceClusterName, userClusterMap[user][SourceClusterName], 1, 1, userBkpLocationNameMap[user], userBkpLocationUidMap[user], appContextsToBackup, BackupNamePrefix)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%v]", backupNames))
				masterUserBackupNames = append(masterUserBackupNames, backupNames...)
				userBackupsMap[user] = backupNames
			}
			err := TaskHandler(append(nonAdminUsers, superAdminUsers...), createManualBackupsFromUser, Parallel)
			log.FailOnError(err, "failed to create manual backups from user")
		})

		Step("Taking scheduled backup of application from source cluster for each user", func() {
			log.InfoD("taking scheduled backup of applications from source cluster for each user")
			createScheduleBackupsFromUser := func(user string) {
				defer GinkgoRecover()
				nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "failed to fetch user [%s] ctx", user)
				scheduleName := fmt.Sprintf("%s-%s-%s", "schedule", user, RandomString(5))
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{userNamespaceMap[user]})
				scheduleBackupName, err := CreateScheduleBackupWithValidation(nonAdminCtx, scheduleName, SourceClusterName, userClusterMap[user][SourceClusterName], userBkpLocationNameMap[user], userBkpLocationUidMap[user], appContextsToBackup, labelSelectors, BackupOrgID, "", "", "", "", schedulePolicyName, schedulePolicyUID)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of schedule backup [%s]", scheduleBackupName))
				err = SuspendBackupSchedule(scheduleName, schedulePolicyName, BackupOrgID, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Suspending Backup Schedule [%s] for user [%s]", scheduleName, user))
				userBackupsMap[user] = append(userBackupsMap[user], scheduleBackupName)
				userScheduleFromSharedClusterMap[user] = scheduleName
			}
			err := TaskHandler(nonAdminUsers, createScheduleBackupsFromUser, Parallel)
			log.FailOnError(err, "failed to create duplicate cluster objects from user")
		})

		Step("Validate Super Admin can view and manage share, edit, delete, all clusters and namespaces, those added by other users.", func() {
			log.InfoD("Validate Super Admin can view and manage share, edit, delete, retry all clusters and namespaces, those added by other users.")
			validateClusterAccess := func(superAdminUser string) {
				defer GinkgoRecover()
				log.Infof("Validate the super admin user can view and manage all clusters and namespaces")
				log.InfoD("Validate the super admin can enumerate all the cluster by users")
				superAdminClusters, err := GetAllClusterFromUser(superAdminUser)
				dash.VerifyFatal(err, nil, "Fetching all clusters from user")
				log.Infof("Super admin user [%s] has access to all the clusters [%v]", superAdminUser, superAdminClusters)
				validateEachUserCluster := func(user string) {
					defer GinkgoRecover()
					userClusters, err := GetAllClusterFromUser(user)
					dash.VerifyFatal(err, nil, "Fetching all clusters from user")
					for _, userClusterUID := range userClusters {
						if IsPresent(superAdminClusters, userClusterUID) {
							log.Infof("Super admin user [%s] has access to cluster [%s] of user [%s]", superAdminUser, userClusterUID, user)
						} else {
							err := fmt.Errorf("super admin user [%s] does not have access to cluster [%s] of user [%s]", superAdminUser, userClusterUID, user)
							log.FailOnError(err, "failed to validate cluster access for super admin user")
						}
					}
				}
				err = TaskHandler(append(nonAdminUsers[:5]), validateEachUserCluster, Parallel)
				log.FailOnError(err, "failed to validate cluster access for super admin user")
			}
			superAdminUser, err := GetSubsetOfSlice(superAdminUsers, 2)
			log.FailOnError(err, "failed to get subset of super admin users")
			err = TaskHandler([]string{superAdminUser[0]}, validateClusterAccess, Sequential)
			log.FailOnError(err, "failed to validate cluster access for super admin user")

			validateClusterShare := func(user string) {
				defer GinkgoRecover()
				superAdminUser, err := GetSubsetOfSlice(superAdminUsers, 1)
				log.FailOnError(err, "failed to get subset of super admin users")
				superAdminCtx, err := backup.GetNonAdminCtx(superAdminUser[0], CommonPassword)
				log.FailOnError(err, "Fetching super admin user [%s] ctx", superAdminUsers[0])
				_, err = ShareClusterWithValidation(superAdminCtx, SourceClusterName, userClusterMap[user][SourceClusterName], nonAdminUsers[7:9], nil, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Sharing cluster [%s] with user [%s]", SourceClusterName, user))
			}
			err = TaskHandler(nonAdminUsers[3:4], validateClusterShare, Sequential)
			log.FailOnError(err, "failed to share cluster from user")

		})

		Step("Validate Super Admin cannot share backup not owned by them .", func() {
			log.InfoD("Validate Super Admin cannot share backup not owned by them")
			validateBackupShare := func(superAdminUser string) {
				superAdminCtx, err := backup.GetNonAdminCtx(superAdminUser, CommonPassword)
				log.FailOnError(err, "Fetching super admin user [%s] ctx", superAdminUser)
				user := nonAdminUsers[4]
				userCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching user [%s] ctx", user)
				sharingUserCtx, err := backup.GetNonAdminCtx(nonAdminUsers[8], CommonPassword)
				log.FailOnError(err, "Fetching user [%s] ctx", nonAdminUsers[8])
				_, err = ShareCluster(superAdminCtx, SourceClusterName, userClusterMap[user][SourceClusterName], []string{nonAdminUsers[8]}, nil, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Sharing cluster [%s] from user[%s] to user [%s]", SourceClusterName, user, nonAdminUsers[8]))
				log.Infof("Validate [%s] doesnt not have access to backup of user [%s] since super admin has shared the cluster", nonAdminUsers[8], user)
				backupName := userBackupsMap[user][0]
				backupUid, err := Inst().Backup.GetBackupUID(userCtx, backupName, BackupOrgID)
				backupInspectReq := &api.BackupInspectRequest{
					Name:  backupName,
					Uid:   backupUid,
					OrgId: BackupOrgID,
				}
				_, err = Inst().Backup.InspectBackup(sharingUserCtx, backupInspectReq)
				if err != nil {
					dash.VerifyFatal(strings.Contains(err.Error(), "not found"), true, fmt.Sprintf("Verifying backup [%s] is not found for user [%s]", backupName, nonAdminUsers[8]))
				} else {
					log.Errorf("Backup [%s] is found for user [%s]", backupName, nonAdminUsers[8])
				}
			}
			superAdminUser, err := GetSubsetOfSlice(superAdminUsers, 2)
			log.FailOnError(err, "failed to get subset of super admin users")
			err = TaskHandler([]string{superAdminUser[0]}, validateBackupShare, Sequential)
			log.FailOnError(err, "failed to validate backup share for super admin user")
		})

		Step("Validate Super Admin cannot revoke/delete backup locations with existing backup schedule", func() {
			log.InfoD("Validate Super Admin cannot revoke/delete backup locations with existing backups schedule")
			validateBackupLocationDeleteError := func(user string) {
				superAdminUser, err := GetSubsetOfSlice(superAdminUsers, 1)
				log.FailOnError(err, "failed to get subset of super admin users")
				superAdminCtx, err := backup.GetNonAdminCtx(superAdminUser[0], CommonPassword)
				log.FailOnError(err, "Fetching super admin user [%s] ctx", superAdminUser[0])
				err = DeleteBackupLocationWithContext(userBkpLocationNameMap[user], userBkpLocationUidMap[user], BackupOrgID, false, superAdminCtx)
				if err != nil {
					dash.VerifyFatal(strings.Contains(err.Error(), "has reference to backuplocation"), true, fmt.Sprintf("Deleting backup location [%s] is not supported when it has reference", userBkpLocationNameMap[user]))
				} else {
					log.Errorf("Deleting backup location [%s] is supported when it has reference", userBkpLocationNameMap[user])
				}
			}
			err := TaskHandler(nonAdminUsers[3:5], validateBackupLocationDeleteError, Sequential)
			log.FailOnError(err, "failed to validate backup location delete for super admin user")
		})

		Step("Validate super admin can create custom role and assign to non admin user and super admin user", func() {
			log.InfoD("Validate super admin can create custom role and assign to non admin user and super admin user")
			superAdminUser, err := GetSubsetOfSlice(superAdminUsers, 3)
			log.FailOnError(err, "failed to get subset of super admin users")
			superAdminCtx, err := backup.GetNonAdminCtx(superAdminUser[0], CommonPassword)
			log.FailOnError(err, "Fetching px-central-admin ctx")
			customRoleName := backup.PxBackupRole(fmt.Sprintf("%s-%s", "custom-role", RandomString(5)))
			services := []RoleServices{BackupSchedulePolicy, Cloudcredential, BackupLocation}
			apis := []RoleApis{All}
			err = CreateRole(customRoleName, services, apis, superAdminCtx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of role [%s] by the super admin", customRoleName))
			roles := [4]backup.PxBackupRole{backup.ApplicationOwner, backup.InfrastructureOwner, backup.ApplicationUser, customRoleName}
			assignRoleForNonAdminSuer := func(user string) {
				randomRole := roles[rand.Intn(len(roles))]
				err := backup.AddRoleToUserWithCredentials(user, randomRole, fmt.Sprintf("Adding %v role to %s", randomRole, user), superAdminUser[0], CommonPassword)
				log.FailOnError(err, "failed to add role %s to the user %s", randomRole, user)
			}
			err = TaskHandler(nonAdminUsers[7:9], assignRoleForNonAdminSuer, Sequential)
			log.FailOnError(err, "failed to assign custom role to non admin users")

			assignRoleForSuperAdminUser := func(user string) {
				err := backup.AddRoleToUserWithCredentials(user, customRoleName, fmt.Sprintf("Adding %v role to %s", customRoleName, user), superAdminUser[2], CommonPassword)
				log.FailOnError(err, "failed to add role %s to the user %s", customRoleName, user)
			}
			err = TaskHandler(superAdminUser[1:2], assignRoleForSuperAdminUser, Sequential)
			log.FailOnError(err, "failed to assign custom role to super admin users")
		})

		Step("Validate all user backups is listed for super admin user", func() {
			log.InfoD("Validate all user backups is listed for super admin user")
			validateBackupList := func(user string) {
				superAdminUser, err := GetSubsetOfSlice(superAdminUsers, 1)
				log.FailOnError(err, "failed to get subset of super admin users")
				superAdminCtx, err := backup.GetNonAdminCtx(superAdminUser[0], CommonPassword)
				log.FailOnError(err, "Fetching px-central-admin ctx")
				nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching user [%s] ctx", user)
				userOwnerID, err := portworx.GetSubFromCtx(nonAdminCtx)
				log.FailOnError(err, "failed to fetch user owner id %s", user)
				backupNamesByOwnerID, err := GetAllBackupNamesByOwnerID(userOwnerID, BackupOrgID, superAdminCtx)
				log.FailOnError(err, "failed to fetch backup names with owner id %s from the admin", userOwnerID)
				for _, backupName := range backupNamesByOwnerID {
					if !IsPresent(userBackupsMap[user], backupName) {
						err := fmt.Errorf("backup [%s] not found in the backup list", backupName)
						log.FailOnError(fmt.Errorf(""), err.Error())
					} else {
						log.Infof("Backup [%s] found in the backup list [%s]", backupName, userBackupsMap[user])
					}
				}
			}
			err := TaskHandler(nonAdminUsers[:5], validateBackupList, Sequential)
			log.FailOnError(err, "failed to validate backup list for super admin user")
		})

		Step("Validate super admin can do delete and restore of a non-admin user backup", func() {
			log.InfoD("Validate super admin can do delete and restore of a non-admin user backup")
			validateBackupDelete := func(user string) {
				superAdminUser, err := GetSubsetOfSlice(superAdminUsers, 1)
				log.FailOnError(err, "failed to get subset of super admin users")
				superAdminCtx, err := backup.GetNonAdminCtx(superAdminUser[0], CommonPassword)
				log.FailOnError(err, "Fetching px-central-admin ctx")
				backupName, _ := GetSubsetOfSlice(userBackupsMap[user], 1)
				backupUid, err := Inst().Backup.GetBackupUID(superAdminCtx, backupName[0], BackupOrgID)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching backup uid [%s]", backupName))
				_, err = DeleteBackup(backupName[0], backupUid, BackupOrgID, superAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting backup [%s]", backupName))
			}
			err := TaskHandler(nonAdminUsers[3:4], validateBackupDelete, Parallel)
			log.FailOnError(err, "failed to validate backup delete for super admin user")
			validateBackupRestore := func(user string) {
				superAdminUser, err := GetSubsetOfSlice(superAdminUsers, 1)
				log.FailOnError(err, "failed to get subset of super admin users")
				superAdminCtx, err := backup.GetNonAdminCtx(superAdminUser[0], CommonPassword)
				log.FailOnError(err, "Fetching px-central-admin ctx")
				backupName, _ := GetSubsetOfSlice(userBackupsMap[user], 1)
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{userNamespaceMap[user]})
				restoreName := fmt.Sprintf("%s-%s-%v", RestoreNamePrefix, backupName[0], time.Now().Unix())
				restoreNamespace := fmt.Sprintf("%s-%s", "custom3", userNamespaceMap[user])
				namespaceMapping := map[string]string{userNamespaceMap[user]: restoreNamespace}
				err = CreateRestoreWithValidation(superAdminCtx, restoreName, backupName[0], namespaceMapping, nil, SourceClusterName, userClusterMap[user][SourceClusterName], BackupOrgID, appContextsToBackup)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Restoring backup [%s]", backupName))
			}
			err = TaskHandler(nonAdminUsers[4:5], validateBackupRestore, Sequential)
			log.FailOnError(err, "failed to validate backup restore for super admin user")
		})

		Step("Validate Super Admin cannot fetch kubeConfig files of non admin user cluster but can update them.", func() {
			log.InfoD("Validate Super Admin cannot fetch kubeConfig files of non admin user cluster but can update them.")
			validateKubeConfig := func(user string) {
				superAdminUser, err := GetSubsetOfSlice(superAdminUsers, 1)
				log.FailOnError(err, "failed to get subset of super admin users")
				superAdminCtx, err := backup.GetNonAdminCtx(superAdminUser[0], CommonPassword)
				log.FailOnError(err, "Fetching px-central-admin ctx")
				clusterName := SourceClusterName
				clusterUid := userClusterMap[user][clusterName]
				clusterInspectResp, err := Inst().Backup.InspectCluster(superAdminCtx, &api.ClusterInspectRequest{
					OrgId: BackupOrgID,
					Name:  clusterName,
					Uid:   clusterUid,
				})
				dash.VerifyFatal(err, nil, fmt.Sprintf("Inspecting cluster [%s]", clusterName))
				if clusterInspectResp.Cluster.Kubeconfig != "" {
					err := fmt.Errorf("super admin user [%s] can fetch kubeconfig of non admin user cluster [%s]", superAdminUsers[0], clusterName)
					log.FailOnError(fmt.Errorf(""), err.Error())
				} else {
					log.Infof("Super admin user [%s] cannot fetch kubeconfig of non admin user cluster [%s]", superAdminUsers[0], clusterName)
				}
			}
			err := TaskHandler(nonAdminUsers[3:4], validateKubeConfig, Sequential)
			log.FailOnError(err, "failed to validate kube config for super admin user")

			validateUpdateKubeConfig := func(user string) {
				superAdminUser, err := GetSubsetOfSlice(superAdminUsers, 1)
				log.FailOnError(err, "failed to get subset of super admin users")
				superAdminCtx, err := backup.GetNonAdminCtx(superAdminUser[0], CommonPassword)
				log.FailOnError(err, "Fetching px-central-admin ctx")
				clusterName := SourceClusterName
				clusterUid := userClusterMap[user][clusterName]
				clusterInspectResp, err := Inst().Backup.InspectCluster(superAdminCtx, &api.ClusterInspectRequest{
					OrgId: BackupOrgID,
					Name:  clusterName,
					Uid:   clusterUid,
				})

				log.FailOnError(err, "failed to inspect cluster %s with uid %s", clusterName, clusterUid)
				kubeConfigPath, err := GetSourceClusterConfigPath()
				log.FailOnError(err, "failed to get source cluster config path")
				validKubeConfig, err := ioutil.ReadFile(kubeConfigPath)
				log.FailOnError(err, "failed to read kubeconfig file")
				clusterUpdateRequest := &api.ClusterUpdateRequest{
					CreateMetadata: &api.CreateMetadata{
						Name:  clusterName,
						Uid:   clusterUid,
						OrgId: BackupOrgID,
					},
					Kubeconfig:            base64.StdEncoding.EncodeToString(validKubeConfig),
					CloudCredential:       clusterInspectResp.GetCluster().GetCloudCredential(),
					CloudCredentialRef:    clusterInspectResp.GetCluster().GetCloudCredentialRef(),
					PlatformCredentialRef: clusterInspectResp.GetCluster().GetPlatformCredentialRef(),
				}
				log.Infof("Updating kubeconfig for cluster %s from user [%s]", clusterName, user)
				_, err = Inst().Backup.UpdateCluster(superAdminCtx, clusterUpdateRequest)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Updating kubeconfig for cluster [%s] of user [%s] from superAdmin user [%s] ", clusterName, user, superAdminUser[0]))
			}
			err = TaskHandler(nonAdminUsers[3:4], validateUpdateKubeConfig, Sequential)
			log.FailOnError(err, "failed to validate kube config for super admin user")
		})

		Step("Validate ownership remains unchanged when schedules are updated by Super Admin.", func() {
			log.InfoD("Validate ownership remains unchanged when schedules are updated by Super Admin.")

			updateBackupSchedule := func(user string) {
				superAdminUser, err := GetSubsetOfSlice(superAdminUsers, 1)
				log.FailOnError(err, "failed to get subset of super admin users")
				superAdminCtx, err := backup.GetNonAdminCtx(superAdminUser[0], CommonPassword)
				log.FailOnError(err, "Fetching px-central-admin ctx")
				nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching user [%s] ctx", user)
				scheduleName := userScheduleFromSharedClusterMap[user]
				scheduleUID, err := GetScheduleUID(scheduleName, BackupOrgID, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Getting schedule UID for schedule %s", scheduleName))
				userSchedulePolicyName := fmt.Sprintf("%s-%v", "periodic-schedule-policy", RandomString(5))
				userSchedulePolicyUid := uuid.New()
				err = CreateBackupScheduleIntervalPolicy(5, schedulePolicyInterval, 5, userSchedulePolicyName, userSchedulePolicyUid, BackupOrgID, superAdminCtx, false, false)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of schedule policy %s", schedulePolicyName))
				err = UpdateBackupSchedulePolicy(scheduleName, scheduleUID, userSchedulePolicyName, userSchedulePolicyUid, superAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Updating schedule policy [%s]", schedulePolicyName))
				log.Info("Inspect the schedule and verify the ownership")

				scheduleInspectResp, err := Inst().Backup.InspectBackupSchedule(nonAdminCtx, &api.BackupScheduleInspectRequest{
					OrgId: BackupOrgID,
					Name:  scheduleName,
					Uid:   scheduleUID,
				})
				ownerID, err := backup.FetchIDOfUser(user)
				log.FailOnError(err, "failed to fetch user id")
				if scheduleInspectResp.GetBackupSchedule().Ownership.GetOwner() != ownerID {
					err := fmt.Errorf("ownership of the schedule [%s] is changed to super admin user", scheduleName)
					log.FailOnError(fmt.Errorf(""), err.Error())
				} else {
					log.Infof("Ownership of the schedule [%s] is not changed", scheduleName)
				}
			}
			err := TaskHandler(nonAdminUsers[3:4], updateBackupSchedule, Sequential)
			log.FailOnError(err, "failed to ownership remains unchanged when schedules are updated by Super Admin")
		})

		Step("Validate that Super Admin cannot delete clusters or revoke cluster share if backups schedules exist.", func() {
			log.InfoD("Validate that Super Admin cannot delete clusters or revoke cluster share if backups schedules exist.")
			validateClusterDeleteError := func(user string) {
				superAdminUser, err := GetSubsetOfSlice(superAdminUsers, 1)
				log.FailOnError(err, "failed to get subset of super admin users")
				superAdminCtx, err := backup.GetNonAdminCtx(superAdminUser[0], CommonPassword)
				log.FailOnError(err, "Fetching px-central-admin ctx")
				clusterName := SourceClusterName
				clusterUid := userClusterMap[user][clusterName]
				err = DeleteClusterWithoutScheduleDelete(SourceClusterName, clusterUid, BackupOrgID, superAdminCtx, false)
				log.Infof("The expected error is %v", err.Error())
				dash.VerifyFatal(strings.Contains(err.Error(), "cluster deletion failed because of existing backup schedules"), true, fmt.Sprintf("Verifying deletion of cluster [%s] is not supported when it has reference", clusterName))
			}
			err := TaskHandler(nonAdminUsers[3:4], validateClusterDeleteError, Sequential)
			log.FailOnError(err, "failed to validate cluster delete for super admin user")

			log.Info("Validate that Super Admin can delete clusters or revoke cluster share if backups schedules is deleted")
			userScheduleDelete := func(user string) {
				superAdminUser, err := GetSubsetOfSlice(superAdminUsers, 1)
				log.FailOnError(err, "failed to get subset of super admin users")
				superAdminCtx, err := backup.GetNonAdminCtx(superAdminUser[0], CommonPassword)
				log.FailOnError(err, "Fetching px-central-admin ctx")
				nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching user [%s] ctx", user)
				scheduleName := userScheduleFromSharedClusterMap[user]
				scheduleUID, err := GetScheduleUID(scheduleName, BackupOrgID, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Getting schedule UID for schedule %s", scheduleName))
				err = DeleteScheduleWithUIDAndWait(scheduleName, scheduleUID, SourceClusterName, userClusterMap[user][SourceClusterName], BackupOrgID, superAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting schedule [%s]", scheduleName))
			}
			err = TaskHandler(nonAdminUsers[3:5], userScheduleDelete, Sequential)
			log.FailOnError(err, "failed to delete schedule for super admin user")

			validateClusterShareError := func(user string) {
				superAdminUser, err := GetSubsetOfSlice(superAdminUsers, 1)
				log.FailOnError(err, "failed to get subset of super admin users")
				superAdminCtx, err := backup.GetNonAdminCtx(superAdminUser[0], CommonPassword)
				log.FailOnError(err, "Fetching px-central-admin ctx")
				clusterName := SourceClusterName
				clusterUid := userClusterMap[user][clusterName]
				_, err = UnShareCluster(superAdminCtx, clusterName, clusterUid, nonAdminUsers[7:9], nil)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying unsharing of cluster [%s] with user [%s]", clusterName, nonAdminUsers[7:9]))
			}
			err = TaskHandler(nonAdminUsers[3:4], validateClusterShareError, Sequential)
			log.FailOnError(err, "failed to validate cluster share for super admin user")

			validateClusterDeleteError = func(user string) {
				superAdminUser, err := GetSubsetOfSlice(superAdminUsers, 1)
				log.FailOnError(err, "failed to get subset of super admin users")
				superAdminCtx, err := backup.GetNonAdminCtx(superAdminUser[0], CommonPassword)
				log.FailOnError(err, "Fetching px-central-admin ctx")
				clusterName := SourceClusterName
				clusterUid := userClusterMap[user][clusterName]
				err = DeleteClusterWithoutScheduleDelete(SourceClusterName, clusterUid, BackupOrgID, superAdminCtx, false)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting cluster [%s]", clusterName))
			}
			err = TaskHandler(nonAdminUsers[4:5], validateClusterDeleteError, Sequential)
			log.FailOnError(err, "failed to validate cluster delete for super admin user")
		})

		Step("Validate demotion of super Admin role to non-admin role, verify the existing superAdmin has access to backup,schedule,restore objects created by demoted superAdmin, And demoted superAdmin has access to his backup objects created pre and post demotion", func() {
			log.InfoD("Validate demotion of super Admin role to non-admin role,verify the existing superAdmin has access to backup,schedule,restore objects created by demoted superAdmin,And demoted superAdmin has access to his backup objects created pre and post demotion")
			userTobeDemoted = superAdminUsers[2]
			var (
				demotedUserBackups   []string
				demotedUserSchedules []string
				demotedUserRestores  []string
			)

			userCtx, err := backup.GetNonAdminCtx(userTobeDemoted, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", userTobeDemoted)

			log.Infof("As a super admin user [%s] create manual backup on owned cluster and non-owned cluster", userTobeDemoted)
			for _, clusterUser := range []string{userTobeDemoted, superAdminUsers[3]} {
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{userNamespaceMap[userTobeDemoted]})
				manualBackupName := fmt.Sprintf("%s-%s-%s", "manual-backup-cluster", userTobeDemoted, userClusterMap[clusterUser][SourceClusterName])
				err = CreateBackupWithValidation(userCtx, manualBackupName, SourceClusterName, userBkpLocationNameMap[userTobeDemoted], userBkpLocationUidMap[userTobeDemoted], appContextsToBackup, labelSelectors, BackupOrgID, userClusterMap[clusterUser][SourceClusterName], "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s] from user [%s] on cluster [%s] with uid [%s] ", manualBackupName, userTobeDemoted, SourceClusterName, userClusterMap[clusterUser][SourceClusterName]))
				demotedUserBackups = append(demotedUserBackups, manualBackupName)
			}

			log.Infof("As a super admin user [%s] create schedule backup on owned cluster and non-owned cluster ", userTobeDemoted)
			for _, clusterUser := range []string{userTobeDemoted, superAdminUsers[3]} {
				scheduleName := fmt.Sprintf("%s-%s-%s", "schedule-cluster", userTobeDemoted, userClusterMap[clusterUser][SourceClusterName])
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{userNamespaceMap[userTobeDemoted]})
				scheduleBackupName, err := CreateScheduleBackupWithValidation(userCtx, scheduleName, SourceClusterName, userClusterMap[clusterUser][SourceClusterName], userBkpLocationNameMap[userTobeDemoted], userBkpLocationUidMap[userTobeDemoted], appContextsToBackup, labelSelectors, BackupOrgID, "", "", "", "", schedulePolicyName, schedulePolicyUID)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of schedule backup for super admin from owned cluster[%s]", scheduleBackupName))
				demotedUserSchedules = append(demotedUserSchedules, scheduleName)
				demotedUserBackups = append(demotedUserBackups, scheduleBackupName)
			}

			log.Infof("As a super admin user [%s] create restore from owned cluster and non-owned cluster", userTobeDemoted)
			for i, clusterUser := range []string{userTobeDemoted, superAdminUsers[3]} {
				backupName := demotedUserBackups[i]
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{userNamespaceMap[userTobeDemoted]})
				restoreName := fmt.Sprintf("%s-%s-%v", RestoreNamePrefix, backupName, RandomString(5))
				restoreNamespace := fmt.Sprintf("%s-%s", "custom3", userNamespaceMap[userTobeDemoted])
				namespaceMapping := map[string]string{userNamespaceMap[userTobeDemoted]: restoreNamespace}
				err = CreateRestoreWithValidation(userCtx, restoreName, backupName, namespaceMapping, nil, DestinationClusterName, userClusterMap[clusterUser][DestinationClusterName], BackupOrgID, appContextsToBackup)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Restoring backup [%s]", backupName))
				demotedUserRestores = append(demotedUserRestores, restoreName)
			}

			log.Infof("Demote the super admin user [%s] to non admin user", userTobeDemoted)
			err = backup.DeleteRoleFromUser(userTobeDemoted, backup.SuperAdmin, fmt.Sprintf("Deleting %v role from %s", backup.SuperAdmin, userTobeDemoted))
			log.FailOnError(err, "failed to delete role %s from the user %s", backup.SuperAdmin, userTobeDemoted)

			log.Infof("As a demoted super admin user wait for next schedule backup to be created for owned and non-owned cluster")

			for _, userSchedule := range demotedUserSchedules {
				wg.Add(1)
				go func(userSchedule string) {
					defer wg.Done()
					scheduleBackupName, err := GetNextCompletedScheduleBackupName(userCtx, userSchedule, time.Duration(schedulePolicyInterval))
					dash.VerifyFatal(err, nil, fmt.Sprintf("Getting next completed schedule backup name for schedule %s", userSchedule))
					mu.Lock() // Ensure safe write access to shared variable
					demotedUserBackups = append(demotedUserBackups, scheduleBackupName)
					mu.Unlock()
				}(userSchedule)
			}
			wg.Wait()

			log.Infof("As a other super admin validate and delete the objects created by demoted super admin")
			validateSuperAdmin := superAdminUsers[1]
			superAdminCtx, err := backup.GetNonAdminCtx(validateSuperAdmin, CommonPassword)
			log.FailOnError(err, "Fetching super admin user [%s] ctx", validateSuperAdmin)

			log.Infof("Validate and delete manual backups from demoted super admin user [%s] from super admin user [%s]  ", userTobeDemoted, validateSuperAdmin)

			allBackups, err := GetAllBackupsForUser(validateSuperAdmin, CommonPassword)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Getting all backups for user %s", validateSuperAdmin))
			for _, backupName := range demotedUserBackups {
				if !IsPresent(allBackups, backupName) {
					err = fmt.Errorf("backup [%s] not found in the backup list", backupName)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying backup [%s] is found in the backup list", backupName))
				} else {
					log.Infof("Backup [%s] found in the backup list , verifying deletion of backup", backupName)
					backupUid, err := Inst().Backup.GetBackupUID(superAdminCtx, backupName, BackupOrgID)
					log.FailOnError(err, fmt.Sprintf("failed to get backup [%s] with uid [%s]", backupName, backupUid))
					_, err = DeleteBackup(backupName, backupUid, BackupOrgID, superAdminCtx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting backup [%s]", backupName))
				}
			}

			allRestores, err := GetAllRestoresForUser(validateSuperAdmin, CommonPassword)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Getting all restores for user %s", validateSuperAdmin))
			for _, restoreName := range demotedUserRestores {
				if !IsPresent(allRestores, restoreName) {
					err = fmt.Errorf("restore [%s] not found in the restore list", restoreName)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying restore [%s] is found in the restore list", restoreName))
				} else {
					log.Infof("Restore [%s] found in the restore list , verifying deletion of restore", restoreName)
					err = DeleteRestore(restoreName, BackupOrgID, superAdminCtx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting restore [%s]", restoreName))
				}
			}

			allSchedules, err := GetAllBackupSchedulesForUser(validateSuperAdmin, CommonPassword)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Getting all schedules for user %s", validateSuperAdmin))
			for _, scheduleName := range demotedUserSchedules {
				if !IsPresent(allSchedules, scheduleName) {
					err = fmt.Errorf("schedule [%s] not found in the schedule list", scheduleName)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying schedule [%s] is found in the schedule list", scheduleName))
				} else {
					log.Infof("Schedule [%s] found in the schedule list , verifying deletion of schedule", scheduleName)
					scheduleUID, err := GetScheduleUID(scheduleName, BackupOrgID, superAdminCtx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Getting schedule UID for schedule %s", scheduleName))
					err = DeleteScheduleWithUID(scheduleName, scheduleUID, BackupOrgID, superAdminCtx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting schedule [%s]", scheduleName))
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

		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true

		log.Info("Destroying scheduled apps on source cluster")
		DestroyApps(scheduledAppContexts, opts)
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		log.InfoD("Deleting the backups")
		allBackups, err := GetAllBackupsAdmin()
		dash.VerifySafely(err, nil, "Verifying fetching of all backups")
		for _, bkp := range allBackups {
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
		log.Infof("Cleanup any clusters in case of failure, since it can have inactive and backup schedules active")
		err = DeleteAllAdminClusters()
		dash.VerifySafely(err, nil, "Verifying deletion of all admin clusters")
		err = CleanupAllUserAndGroups()
		dash.VerifySafely(err, nil, "Verifying cleanup all user and groups")
		CleanupCloudSettingsAndClusters(backupLocationMap, adminCloudCredName, adminCloudCredentialUID, ctx)
	})
})

// This test case is to validate the cluster share operation while concurrent backup operations are ongoing
var _ = Describe("{ValidateClusterShareWithConcurrentBackupOperations}", Label(TestCaseLabelsMap[ValidateClusterShareWithConcurrentBackupOperations]...), func() {
	var (
		numOfDeployments                    int
		numOfBackupsPerDeployment           int
		numOfAdditionalBackupsPerDeployment int
		snapshotLimit                       int
		numOfRestores                       int
		numOfPrimaryUsers                   int
		numOfSharedUsers                    int
		numberOfGroups                      int
		splitIndexForUsers                  int
		splitIndexForGroups                 int
		cloudAccountName                    string
		cloudAccountUid                     string
		backupLocationName                  string
		backupLocationUid                   string
		randomUser                          string
		primaryUserList                     []string
		sharedUserList                      []string
		groupList                           []string
		firstUserList                       []string
		secondUserList                      []string
		firstGroupList                      []string
		secondGroupList                     []string
		restoreNames                        []string
		backupLocationMap                   map[string]string
		initialBackupMap                    map[string][]string
		firstBackupMap                      map[string][]string
		secondBackupMap                     map[string][]string
		thirdBackupMap                      map[string][]string
		primaryUserClusterMap               map[string]map[string]string
		sharedUserClusterMap                map[string]map[string]string
		backupAppContexts                   []*scheduler.Context
		namespaceAppContextMap              map[string][]*scheduler.Context
		wg                                  sync.WaitGroup
		mu                                  sync.Mutex
		roles                               []backup.PxBackupRole
		clusterShareOnPremCluster           []string
		isClusterShareSupportedForAllRoles  bool
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("ValidateClusterShareWithConcurrentBackupOperations", "TC to verify cluster share with Concurrent Backup and Restore Operations", nil, 301140, Sabrarhussaini, Q3FY25)

		snapshotLimit = 4
		numOfPrimaryUsers = 5
		numOfDeployments = numOfPrimaryUsers
		numOfBackupsPerDeployment = 4
		numOfAdditionalBackupsPerDeployment = 2
		numOfSharedUsers = 4
		numberOfGroups = 4
		splitIndexForUsers = 2
		splitIndexForGroups = 2
		numOfRestores = 2
		initialBackupMap = make(map[string][]string)               // Primary set of backups which are created before clusters are shared
		firstBackupMap = make(map[string][]string)                 // First set of additional backups which are created for cluster share validation with backup share
		secondBackupMap = make(map[string][]string)                // Second set of additional backups which are created for cluster share validation with backup share
		thirdBackupMap = make(map[string][]string)                 // Third set of additional backups which are created for cluster share validation without backup share
		primaryUserClusterMap = make(map[string]map[string]string) // Map to store Primary users, Cluster Names and Cluster UIDs
		sharedUserClusterMap = make(map[string]map[string]string)  // Map to store Secondary users, Shared Cluster Names and Shared Cluster UIDs
		backupAppContexts = make([]*scheduler.Context, 0)
		namespaceAppContextMap = make(map[string][]*scheduler.Context)

		clusterShareOnPremCluster = []string{"vanilla", "openshift"}
		if Contains(clusterShareOnPremCluster, GetClusterProvider()) {
			isClusterShareSupportedForAllRoles = true
		} else {
			isClusterShareSupportedForAllRoles = false
		}

		err := SetSourceKubeConfig()
		log.FailOnError(err, "Switching context to source cluster failed")

		// Schedule applications
		log.Infof("Scheduling applications")
		for i := 0; i < numOfDeployments; i++ {
			taskName := fmt.Sprintf("multiple-%d", i)
			appContexts := ScheduleApplications(taskName)
			for _, appCtx := range appContexts {
				namespace := GetAppNamespace(appCtx, taskName)
				backupAppContexts = append(backupAppContexts, appCtx)
				appCtx.ReadinessTimeout = AppReadinessTimeout
				namespaceAppContextMap[namespace] = append(namespaceAppContextMap[namespace], appCtx)
			}
		}
	})

	It("TC to verify Cluster Sharing with Concurrent Backup and Restore Operations", func() {
		Step("Validating applications ", func() {
			log.InfoD("Validating all the deployed applications")
			ValidateApplications(backupAppContexts)
		})

		Step("Create a set of primary users with different roles", func() {
			log.Infof("Creating a set of %d primary users with different roles", numOfPrimaryUsers)
			primaryUserList = CreateUsers(numOfPrimaryUsers)
			if isClusterShareSupportedForAllRoles {
				roles = []backup.PxBackupRole{backup.ApplicationOwner, backup.InfrastructureOwner, backup.SuperAdmin, backup.ApplicationUser}
			} else {
				roles = []backup.PxBackupRole{backup.SuperAdmin, backup.InfrastructureOwner}
			}
			for i, user := range primaryUserList {
				role := roles[i%len(roles)]
				err := backup.AddRoleToUser(user, role, fmt.Sprintf("Adding %v role to %s", role, user))
				log.FailOnError(err, fmt.Sprintf("failed to add role %s to the user %s", role, user))
			}
		})

		Step("Create a set of secondary users to share clusters with", func() {
			log.Infof("Creating  a set of %d secondary users to share the clusters with", numOfSharedUsers)
			roles = []backup.PxBackupRole{backup.SuperAdmin, backup.ApplicationUser, backup.ApplicationOwner, backup.InfrastructureOwner}
			sharedUserList = CreateUsers(numOfSharedUsers)
			for i, user := range sharedUserList {
				role := roles[i%len(roles)]
				err := backup.AddRoleToUser(user, role, fmt.Sprintf("Adding %v role to %s", role, user))
				log.FailOnError(err, fmt.Sprintf("failed to add role %s to the user %s", role, user))
			}
			firstUserList = sharedUserList[:splitIndexForUsers]
			secondUserList = sharedUserList[splitIndexForUsers:]
		})

		Step("Create Groups to share clusters with", func() {
			log.InfoD("Creating %d groups to share the clusters with", numberOfGroups)
			for i := 1; i <= numberOfGroups; i++ {
				groupName := fmt.Sprintf("pxbGroup%v", i)
				wg.Add(1)
				go func(groupName string) {
					defer GinkgoRecover()
					defer wg.Done()
					err := backup.AddGroup(groupName)
					log.FailOnError(err, fmt.Sprintf("Failed to create group - %v", groupName))
					mu.Lock()
					groupList = append(groupList, groupName)
					mu.Unlock()
				}(groupName)
			}
			wg.Wait()
			firstGroupList = groupList[:splitIndexForGroups]
			secondGroupList = groupList[splitIndexForGroups:]
		})

		Step("Adding Cloud credentials and Backup Location from Px-Admin", func() {
			log.InfoD(fmt.Sprintf("Adding Credentials and Backup Location from px-admin and making it public"))
			providers := GetBackupProviders()
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			for _, provider := range providers {
				cloudAccountUid = uuid.New()
				cloudAccountName = fmt.Sprintf("autogenerated-cred-%v", RandomString(5))
				if provider != drivers.ProviderNfs {
					err = CreateCloudCredential(provider, cloudAccountName, cloudAccountUid, BackupOrgID, ctx)
					log.FailOnError(err, "Failed to create cloud credential")
					err = AddCloudCredentialOwnership(cloudAccountName, cloudAccountUid, nil, nil, Invalid, Read, ctx, BackupOrgID)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying public ownership update for cloud credential %s ", cloudAccountName))
				}
				backupLocationName = fmt.Sprintf("autogenerated-backup-location-%v", RandomString(5))
				backupLocationUid = uuid.New()
				err = CreateBackupLocationWithContext(provider, backupLocationName, backupLocationUid, cloudAccountName, cloudAccountUid, getGlobalBucketName(provider), BackupOrgID, "", ctx, true)
				log.FailOnError(err, fmt.Sprintf("Failed to add backup location %s using provider %s for px-admin user", backupLocationName, provider))
				err = AddBackupLocationOwnership(backupLocationName, backupLocationUid, nil, nil, Invalid, Read, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying public ownership update for backup location %s", backupLocationName))
			}
		})

		Step("Add source and destination clusters for all primary users", func() {
			log.InfoD("Adding source and destination clusters for all primary users")
			for _, user := range primaryUserList {
				primaryUserClusterMap[user] = make(map[string]string)
				nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, fmt.Sprintf("Fetching user [%s] ctx", user))
				log.Infof("Creating source [%s] and destination [%s] clusters for user [%s]", SourceClusterName, DestinationClusterName, user)
				err = CreateApplicationClusters(BackupOrgID, "", "", nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of source [%s] and destination [%s] clusters with user [%s] ctx", SourceClusterName, DestinationClusterName, user))
				srcClusterUid, err := Inst().Backup.GetClusterUID(nonAdminCtx, BackupOrgID, SourceClusterName)
				log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster UID", SourceClusterName))
				destClusterUid, err := Inst().Backup.GetClusterUID(nonAdminCtx, BackupOrgID, DestinationClusterName)
				log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster UID", DestinationClusterName))
				log.Infof("User [%s]: Cluster [%s] UID: [%s]", user, DestinationClusterName, destClusterUid)
				primaryUserClusterMap[user][SourceClusterName] = srcClusterUid
				primaryUserClusterMap[user][DestinationClusterName] = destClusterUid
			}
		})

		Step("Taking initial backups for all primary users", func() {
			log.Infof("Taking a total of %v initial backups for all primary users", numOfDeployments*numOfBackupsPerDeployment)
			for i, user := range primaryUserList {
				wg.Add(1)
				i := i
				go func(user string) {
					defer wg.Done()
					defer GinkgoRecover()
					ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
					log.FailOnError(err, "Failed to fetch context for user")
					backupNameList, err := TakeMultipleBackupsPerDeployment(ctx, BackupOrgID, SourceClusterName, primaryUserClusterMap[user][SourceClusterName], numOfBackupsPerDeployment, snapshotLimit, backupLocationName, backupLocationUid, []*scheduler.Context{backupAppContexts[i]}, "initial-backup")
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creating initial backups for user [%s]", user))
					mu.Lock()
					initialBackupMap[user] = backupNameList
					mu.Unlock()
				}(user)
			}
			wg.Wait()
		})

		Step("Initiate first set of additional backups and restores for all primary users for cluster share validation", func() {
			log.InfoD("Initiating first set of additional backups and restores for all primary users for cluster share validation")
			for i, user := range primaryUserList {
				wg.Add(1)
				i := i
				go func(user string) {
					defer wg.Done()
					defer GinkgoRecover()

					ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
					log.FailOnError(err, fmt.Sprintf("Failed to fetch context for user %s", user))
					localBackupNameList, err := TakeMultipleBackupsPerDeploymentWithoutCheck(ctx, BackupOrgID, SourceClusterName, primaryUserClusterMap[user][SourceClusterName], numOfAdditionalBackupsPerDeployment, snapshotLimit, backupLocationName, backupLocationUid, []*scheduler.Context{backupAppContexts[i]}, "additional-first")
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creating first set of additional backups for user [%s]", user))
					mu.Lock()
					firstBackupMap[user] = localBackupNameList
					mu.Unlock()
					// Initiate restore for the backups created initially by primary users
					for i, backupName := range initialBackupMap[user] {
						if i < numOfRestores {
							restoreName := fmt.Sprintf("%s-restore-%s-%d", user, RandomString(6), i+1)
							_, err := CreateRestoreWithoutCheck(restoreName, backupName, make(map[string]string), DestinationClusterName, primaryUserClusterMap[user][DestinationClusterName], BackupOrgID, ctx)
							dash.VerifyFatal(err, nil, fmt.Sprintf("Creating retore [%s] for user [%s]", restoreName, user))
							restoreNames = append(restoreNames, restoreName)
						}
					}
				}(user)
			}
			wg.Wait()
			log.Info("First set of additional backups and restores have been initiated for all primary users.")
		})

		Step("Share the clusters with backups from each primary user to secondary users and groups", func() {
			log.InfoD("Sharing the clusters with backups from each primary user to secondary users and groups")
			for _, user := range primaryUserList {
				wg.Add(1)

				go func(user string) {
					defer wg.Done()
					defer GinkgoRecover()
					clusters := []string{SourceClusterName, DestinationClusterName}
					for _, clusterName := range clusters {
						ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
						log.FailOnError(err, "Failed to fetch context for user")
						_, err = ShareCluster(ctx, clusterName, primaryUserClusterMap[user][clusterName], sharedUserList, groupList, true)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Sharing of cluster [%s] from user %s", clusterName, user))
						for _, sharedUser := range sharedUserList {
							mu.Lock()
							if sharedUserClusterMap[sharedUser] == nil {
								sharedUserClusterMap[sharedUser] = make(map[string]string)
							}
							sharedUserClusterMap[sharedUser][clusterName] = primaryUserClusterMap[user][clusterName]
							mu.Unlock()
						}
					}
				}(user)
			}
			wg.Wait()
			time.Sleep(30 * time.Second)
		})

		Step("Validate the shared clusters for each secondary users and groups", func() {
			log.InfoD("Validate the shared clusters for each secondary users and groups")
			for _, user := range primaryUserList {
				wg.Add(1)

				go func(user string) {
					defer wg.Done()
					defer GinkgoRecover()
					clusters := []string{SourceClusterName, DestinationClusterName}
					for _, clusterName := range clusters {
						ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
						log.FailOnError(err, "Failed to fetch context for user")
						err = ValidateShareCluster(ctx, clusterName, primaryUserClusterMap[user][clusterName], sharedUserList, groupList)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Validating shared cluster [%s] for user %s", clusterName, user))
					}
				}(user)
			}
			wg.Wait()
		})

		Step("Verify the shared backups for a shared user", func() {
			log.InfoD("Verifying the shared backups for a shared user")
			userIndex := rand.Intn(len(sharedUserList))
			randomUser = sharedUserList[userIndex]
			log.Infof("Validating the backups using the shared user %s", randomUser)
			ctx, err := backup.GetNonAdminCtx(randomUser, CommonPassword)
			log.FailOnError(err, "Fetching user ctx")
			backups := firstBackupMap[primaryUserList[0]]

			for _, backupName := range backups {
				wg.Add(1)
				go func(backupName string) {
					defer wg.Done()
					defer GinkgoRecover()
					err = BackupSuccessCheck(backupName, BackupOrgID, MaxWaitPeriodForBackupCompletionInMinutes*time.Minute, 30*time.Second, ctx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying backup [%s] success check", backupName))
					log.Infof("Backup [%s] created successfully", backupName)
				}(backupName)
			}
			wg.Wait()

			//Validate restore and backup deletion if user is admin
			isUserAdmin, err := IsAdminCtx(ctx)
			log.FailOnError(err, "Verifying if the user selected is an admin")
			if isUserAdmin {
				log.InfoD("Verifying the restore of shared backups for a shared user")
				var collectedAppContexts []*scheduler.Context
				collectedAppContexts, err = GetAppContextsFromBackup(backups[0], BackupOrgID, ctx, namespaceAppContextMap)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Getting app context from backup [%s] for user [%s]", backups[0], randomUser))
				restoreName := fmt.Sprintf("restore-%s-%s", randomUser, RandomString(6))
				err = CreateRestoreWithValidation(ctx, restoreName, backups[0], make(map[string]string), make(map[string]string), DestinationClusterName, sharedUserClusterMap[randomUser][DestinationClusterName], BackupOrgID, collectedAppContexts)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Restoring backup [%s] from user [%s] for cluster [%s] with UID [%s]", restoreName, randomUser, DestinationClusterName, sharedUserClusterMap[randomUser][DestinationClusterName]))
			}

			if isUserAdmin {
				log.InfoD("Verifying deletion of one of the shared backups for the shared user")
				backupUid, err := Inst().Backup.GetBackupUID(ctx, backups[0], BackupOrgID)
				log.FailOnError(err, fmt.Sprintf("Failed to fetch the backup %s uid of the user %s", backups[0], randomUser))
				_, err = DeleteBackup(backups[0], backupUid, BackupOrgID, ctx)
				log.FailOnError(err, fmt.Sprintf("Failed to delete the backup %s of the user %s", backups[0], randomUser))
			}
		})

		Step("Initiate second set of additional backups and restores for all primary users for validation", func() {
			log.InfoD("Initiating second set of additional backups and restores for all primary users validation")

			for i, user := range primaryUserList {
				wg.Add(1)
				i := i
				go func(user string) {
					defer wg.Done()
					defer GinkgoRecover()

					ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
					log.FailOnError(err, fmt.Sprintf("Failed to fetch context for user %s", user))
					localBackupNameList, err := TakeMultipleBackupsPerDeploymentWithoutCheck(ctx, BackupOrgID, SourceClusterName, primaryUserClusterMap[user][SourceClusterName], numOfAdditionalBackupsPerDeployment, snapshotLimit, backupLocationName, backupLocationUid, []*scheduler.Context{backupAppContexts[i]}, "additional-second")
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creating second set of additional backups for user [%s]", user))
					mu.Lock()
					secondBackupMap[user] = localBackupNameList
					mu.Unlock()

					// Initiate restore for the backup
					backups := initialBackupMap[user]
					for i, backupName := range backups {
						if i < numOfRestores {
							restoreName := fmt.Sprintf("%s-restore-%s-%d", user, RandomString(6), i+1)
							_, err := CreateRestoreWithoutCheck(restoreName, backupName, make(map[string]string), DestinationClusterName, primaryUserClusterMap[user][DestinationClusterName], BackupOrgID, ctx)
							dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restores for user [%s]", user))
						}
					}
				}(user)
			}
			log.Info("Second set of additional backups and restores have been initiated for all primary users.")
		})

		Step("Unshare the clusters from primary users for a subset of secondary users and groups", func() {
			log.InfoD("Unsharing the clusters from primary users for a subset of secondary users and groups")
			for _, user := range primaryUserList {
				clusters := []string{SourceClusterName, DestinationClusterName}
				ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Failed to fetch context for user")
				for _, clusterName := range clusters {
					_, err = UnShareClusterWithValidation(ctx, clusterName, primaryUserClusterMap[user][clusterName], firstUserList, firstGroupList)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying unsharing of cluster [%s] for user %s", clusterName, user))
				}
			}
		})

		Step("Verify the newly created backups for the subset of secondary users from whom the clusters were unshared", func() {
			log.InfoD("Verifying the newly created backups for the subset of secondary users from whom the clusters were unshared")
			userIndex := rand.Intn(len(firstUserList))
			randomUser = firstUserList[userIndex]
			fmt.Printf("Randomly selected user from unshared list of users: %s\n", randomUser)
			backups := secondBackupMap[primaryUserList[0]]
			ctx, err := backup.GetNonAdminCtx(randomUser, CommonPassword)
			log.FailOnError(err, "Fetching user ctx")
			for _, backupName := range backups {
				err = BackupSuccessCheck(backupName, BackupOrgID, MaxWaitPeriodForBackupCompletionInMinutes*time.Minute, 30*time.Second, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying backup [%s] success check", backupName))
				log.Infof("Backup [%s] created successfully", backupName)
			}

			//Validate backup deletion if user is admin
			isUserAdmin, err := IsAdminCtx(ctx)
			log.FailOnError(err, "Verifying if the user selected is an admin")
			if isUserAdmin {
				log.InfoD("Verifying the deletion of shared backups for a shared user")
				for _, backupName := range backups {
					wg.Add(1)
					go func(backupName string) {
						defer GinkgoRecover()
						defer wg.Done()
						backupUid, err := Inst().Backup.GetBackupUID(ctx, backupName, BackupOrgID)
						log.FailOnError(err, fmt.Sprintf("Failed to fetch the backup %s uid of the user %s", backupName, randomUser))
						_, err = DeleteBackup(backupName, backupUid, BackupOrgID, ctx)
						log.FailOnError(err, fmt.Sprintf("Failed to delete the backup %s of the user %s", backupName, randomUser))
						err = DeleteBackupAndWait(backupName, ctx)
						log.FailOnError(err, fmt.Sprintf("waiting for backup [%s] deletion", backupName))
					}(backupName)
				}
				wg.Wait()
			}
		})

		Step("Unshare the clusters from px-central admin for other subset of secondary users and groups", func() {
			log.InfoD("Unsharing the clusters from px-central admin for other subset of secondary users and groups")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			adminClusterList, err := GetAllClusterAdmin()
			dash.VerifySafely(err, nil, "Verifying fetching of all clusters")
			for clusterUid, clusterName := range adminClusterList {
				wg.Add(1)
				go func(clusterName, clusterUid string) {
					defer GinkgoRecover()
					defer wg.Done()
					_, err = UnShareClusterWithValidation(ctx, clusterName, clusterUid, secondUserList, secondGroupList)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying unsharing of cluster [%s] from px-admin", clusterName))
				}(clusterName, clusterUid)
			}
			wg.Wait()
		})

		Step("Unshare the existing shared backups from secondary users and groups", func() {
			log.InfoD("Unsharing the existing shared backups from secondary users and groups")
			clusters := []string{SourceClusterName, DestinationClusterName}
			for _, user := range primaryUserList {
				ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Failed to fetch context for user")
				for _, clusterName := range clusters {
					err = ClusterUpdateBackupShare(clusterName, secondGroupList, secondUserList, RestoreAccess, false, ctx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying unsharing of backups for cluster [%s]", clusterName))
				}
			}
		})

		Step("Initiate third set of additional backups for all primary users for validation", func() {
			log.InfoD("Taking third set additional backups for all primary users")
			for i, user := range primaryUserList {
				wg.Add(1)
				i := i
				go func(user string) {
					defer wg.Done()
					defer GinkgoRecover()
					ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
					log.FailOnError(err, "Failed to fetch context for user")
					backupNameList, err := TakeMultipleBackupsPerDeploymentWithoutCheck(ctx, BackupOrgID, SourceClusterName, primaryUserClusterMap[user][SourceClusterName], numOfBackupsPerDeployment, snapshotLimit, backupLocationName, backupLocationUid, []*scheduler.Context{backupAppContexts[i]}, "additional-third")
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creating third set of additional backups for user [%s]", user))
					mu.Lock()
					thirdBackupMap[user] = backupNameList
					mu.Unlock()
				}(user)
			}
			wg.Wait()
			log.Info("Third set of additional backups are created for all primary users.")
		})

		Step("Share the clusters without backups from each primary user to secondary users and groups", func() {
			log.InfoD("Sharing the clusters without backups from each primary user to secondary users and groups")

			for _, user := range primaryUserList {
				wg.Add(1)
				go func(user string) {
					defer wg.Done()
					defer GinkgoRecover()
					ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
					log.FailOnError(err, "Failed to fetch context for user")
					clusterUid, err := Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
					_, err = ShareClusterWithValidation(ctx, SourceClusterName, clusterUid, sharedUserList, groupList, false)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying share of source [%s] cluster without backups for user %s", SourceClusterName, user))
				}(user)
			}
			wg.Wait()
		})

		Step("Verify if the latest backups of primary users are not seen by secondary users after cluster share", func() {
			log.InfoD("Verifying if the latest backups of primary users are not seen by secondary users after cluster share")
			randomUserIndex := rand.Intn(len(secondUserList))
			randomUser := secondUserList[randomUserIndex]
			fmt.Printf("Randomly selected user from shared list of users to verify if backups are not seen: %s\n", randomUser)
			backups := thirdBackupMap[primaryUserList[0]]
			backupNamesForUser, err := GetAllBackupsForUser(randomUser, CommonPassword)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching all backups for user [%s]", randomUser))

			ctx, err := backup.GetNonAdminCtx(randomUser, CommonPassword)
			log.FailOnError(err, "Fetching user ctx")
			isUserAdmin, err := IsAdminCtx(ctx)
			log.FailOnError(err, "Verifying if the user selected is an admin")
			for _, backupName := range backups {
				if IsPresent(backupNamesForUser, backupName) && !isUserAdmin {
					err = fmt.Errorf("backup [%s] found for user [%s]", backupName, randomUser)
					log.FailOnError(err, "failed to validate the permissions for backup")
				} else {
					log.Infof("Backup [%s] not found for user [%s]", backupName, randomUser)
				}
			}
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(backupAppContexts)
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		err = SetSourceKubeConfig()
		log.FailOnError(err, "Switching context to source cluster failed")
		log.InfoD("Deleting all backups")
		adminBackups, err := GetAllBackupsAdmin()
		dash.VerifyFatal(err, nil, "Verifying fetching of all backups")
		for _, bkp := range adminBackups {
			wg.Add(1)
			go func(bkp string) {
				defer wg.Done()
				backupUID, err := Inst().Backup.GetBackupUID(ctx, bkp, BackupOrgID)
				_, err = DeleteBackup(bkp, backupUID, BackupOrgID, ctx)
				dash.VerifySafely(err, nil, fmt.Sprintf("Verifying backup deletion - %s", bkp))
			}(bkp)
		}
		wg.Wait()
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(backupAppContexts, opts)
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudAccountName, cloudAccountUid, ctx)
		log.InfoD("Switching context to destination cluster for clean up")
		err = SetDestinationKubeConfig()
		log.FailOnError(err, "Unable to switch context to destination cluster [%s]", DestinationClusterName)
		DestroyApps(backupAppContexts, opts)
		log.InfoD("Switching back context to Source cluster")
		err = SetSourceKubeConfig()
		log.FailOnError(err, "Unable to switch context to source cluster [%s]", SourceClusterName)
		err = CleanupAllUserAndGroups()
		dash.VerifySafely(err, nil, "Verifying cleanup all user and groups")
	})
})

// This test case is to validate the backup share operation using the shared cluster share
var _ = Describe("{ValidateBackupShareUsingBackupsFromSharedCluster}", Label(TestCaseLabelsMap[ValidateBackupShareUsingBackupsFromSharedCluster]...), func() {
	var (
		numOfPrimaryUsers           int
		numOfDeployments            int
		snapshotLimit               int
		numOfInitialBackups         int
		numberOfGroups              int
		numOfSecondaryUsers         int
		primaryUserList             []string
		secondaryUserList           []string
		groupList                   []string
		adminCloudCredName          string
		adminCloudCredentialUID     string
		userBkpLocationNameMap      map[string]string
		userBkpLocationUidMap       map[string]string
		clusterMap                  map[string]string
		backupLocationMap           map[string]string
		backupMap                   map[string][]string
		backupAppContexts           []*scheduler.Context
		infraAdminRole              backup.PxBackupRole
		wg                          sync.WaitGroup
		adminBackupLocationName     string
		adminBackupLocationUID      string
		providers                   []string
		backupNameSpaces            []string
		userNameSpaceMap            map[string]string
		adminBackupNames            []string
		clusterSharePermissionError = "PermissionDenied"
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("ValidateBackupShareUsingBackupsFromSharedCluster", "TC to verify Backup share for the backups taken from the shared cluster", nil, 301143, Sabrarhussaini, Q3FY25)

		providers = GetBackupProviders()
		numOfPrimaryUsers = 3
		numOfDeployments = numOfPrimaryUsers
		snapshotLimit = 4
		numberOfGroups = 2
		numOfInitialBackups = 4
		numOfSecondaryUsers = 4
		backupMap = make(map[string][]string)
		clusterMap = make(map[string]string)
		userBkpLocationNameMap = make(map[string]string)
		userBkpLocationUidMap = make(map[string]string)
		backupLocationMap = make(map[string]string)
		backupAppContexts = make([]*scheduler.Context, 0)
		infraAdminRole = backup.InfrastructureOwner
		userNameSpaceMap = make(map[string]string)

		log.Infof("Scheduling applications")
		err := SetSourceKubeConfig()
		log.FailOnError(err, "Switching context to source cluster failed")
		log.Infof("Scheduling applications")
		for i := 0; i < numOfDeployments; i++ {
			taskName := fmt.Sprintf("multiple-%d", i)
			appContexts := ScheduleApplications(taskName)
			for _, appCtx := range appContexts {
				namespace := GetAppNamespace(appCtx, taskName)
				backupAppContexts = append(backupAppContexts, appCtx)
				appCtx.ReadinessTimeout = AppReadinessTimeout
				backupNameSpaces = append(backupNameSpaces, namespace)
			}
		}
	})

	It("TC to verify Backup share for the backups taken from the shared cluster", func() {
		Step("Validate deployed applications ", func() {
			log.InfoD("Validating all the deployed applications")
			ValidateApplications(backupAppContexts)
		})

		Step("Create a set of primary users with infra admin role", func() {
			log.InfoD("Creating 3 users with infra-admin role")
			primaryUserList = CreateUsers(numOfPrimaryUsers)
			for i, user := range primaryUserList {
				err := backup.AddRoleToUser(user, infraAdminRole, fmt.Sprintf("Adding %v role to %s", infraAdminRole, user))
				log.FailOnError(err, "failed to add role %s to the user %s", infraAdminRole, user)
				userNameSpaceMap[user] = backupNameSpaces[i]
			}
		})

		Step("Add Cloud credentials and Backup Location for primary users", func() {
			log.InfoD("Adding Credentials and Backup Location from Users for primary Users")
			for _, user := range primaryUserList {
				for _, provider := range providers {
					nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
					log.FailOnError(err, "Fetching user ctx")
					cloudAccountUid := uuid.New()
					cloudAccountName := fmt.Sprintf("cred-%s-%v", user, RandomString(5))
					err = CreateCloudCredential(provider, cloudAccountName, cloudAccountUid, BackupOrgID, nonAdminCtx)
					log.FailOnError(err, "Failed to create cloud credential")
					backupLocationName := fmt.Sprintf("backup-location-%s-%v", user, RandomString(5))
					backupLocationUid := uuid.New()
					err = CreateBackupLocationWithContext(provider, backupLocationName, backupLocationUid, cloudAccountName, cloudAccountUid, getGlobalBucketName(provider), BackupOrgID, "", nonAdminCtx, true)
					log.FailOnError(err, fmt.Sprintf("Failed to add backup location %s using provider %s for %s", backupLocationName, provider, user))
					userBkpLocationNameMap[user] = backupLocationName
					userBkpLocationUidMap[user] = backupLocationUid
					backupLocationMap[backupLocationUid] = backupLocationName
				}
			}
		})

		Step("Add Cloud credentials and Backup Location for admin", func() {
			log.InfoD("Adding Credentials and Backup Location from Users for for admin")
			for _, provider := range providers {
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-central-admin ctx")
				adminCloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, RandomString(5))
				adminBackupLocationName = fmt.Sprintf("%s-%s-bl-%v", provider, getGlobalBucketName(provider), RandomString(5))
				adminCloudCredentialUID = uuid.New()
				adminBackupLocationUID = uuid.New()
				backupLocationMap[adminBackupLocationUID] = adminBackupLocationName
				err = CreateCloudCredential(provider, adminCloudCredName, adminCloudCredentialUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", adminCloudCredName, BackupOrgID, provider))
				err = CreateBackupLocation(provider, adminBackupLocationName, adminBackupLocationUID, adminCloudCredName, adminCloudCredentialUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location  [%s]for admin", adminBackupLocationName))
			}
		})

		Step("Add source and destination clusters for super admin and share it with primary users", func() {
			log.InfoD("Add source and destination clusters for super admin and share it with primary users\"")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of source [%s] and destination [%s] clusters", SourceClusterName, DestinationClusterName))
			for _, clusterName := range []string{SourceClusterName, DestinationClusterName} {
				clusterUid, err := Inst().Backup.GetClusterUID(ctx, BackupOrgID, clusterName)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", clusterName))
				clusterMap[clusterName] = clusterUid
				log.Infof("Sharing cluster - [%s] with primary users", clusterName)
				_, err = ShareClusterWithValidation(ctx, clusterName, clusterUid, primaryUserList, nil, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying share of [%s] cluster", clusterName))
			}
		})

		Step("Take backups for all primary users to whom clusters were shared", func() {
			log.InfoD("Taking %v backups for all primary users to whom the clusters were shared", numOfPrimaryUsers*numOfInitialBackups)
			createBackupFromPrimaryUser := func(user string) {
				ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, fmt.Sprintf("Failed to fetch context for user %s", user))
				log.Infof("Using %s backup location for user %s", userBkpLocationNameMap[user], user)
				appContextToBackup := FilterAppContextsByNamespace(backupAppContexts, []string{userNameSpaceMap[user]})
				backupNameList, err := TakeMultipleBackupsPerDeployment(ctx, BackupOrgID, SourceClusterName, clusterMap[SourceClusterName], numOfInitialBackups, snapshotLimit, userBkpLocationNameMap[user], userBkpLocationUidMap[user], appContextToBackup, "initial-backup")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating initial backups for user [%s]", user))
				backupMap[user] = backupNameList
			}
			err := TaskHandler(primaryUserList, createBackupFromPrimaryUser, Parallel)
			log.FailOnError(err, "failed to create initial backups for primary users")
		})

		Step("Perform restore operation for every primary user to whom clusters were shared", func() {
			log.InfoD("Performing restore operation for every user to whom clusters were shared")
			for _, user := range primaryUserList {
				nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, fmt.Sprintf("Failed to get ctx for user %s ", user))
				backupToRestore, err := GetSubsetOfSlice(backupMap[user], 1)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching backup to restore for user %s", user))
				restoreName := fmt.Sprintf("restore-%s-%s", user, RandomString(4))
				log.Infof("Restoring %s for backup %s for user %s", restoreName, backupToRestore, user)
				appContextToBackup := FilterAppContextsByNamespace(backupAppContexts, []string{userNameSpaceMap[user]})
				err = CreateRestoreWithValidation(nonAdminCtx, restoreName, backupToRestore[0], make(map[string]string), make(map[string]string), DestinationClusterName, clusterMap[DestinationClusterName], BackupOrgID, appContextToBackup)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Restoring backup [%s] from user [%s] for cluster [%s]", backupToRestore, user, DestinationClusterName))
			}
		})

		Step("Perform delete operations for every user to whom clusters were shared", func() {
			log.Infof("Perform delete operations for every user to whom clusters were shared")
			validateBackupDelete := func(user string) {
				nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, fmt.Sprintf("Failed to get ctx for user %s ", user))
				backupNameToDelete, err := GetSubsetOfSlice(backupMap[user], 1)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching backup to delete for user %s", user))
				backupUid, err := Inst().Backup.GetBackupUID(nonAdminCtx, backupNameToDelete[0], BackupOrgID)
				_, err = DeleteBackup(backupNameToDelete[0], backupUid, BackupOrgID, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting backup [%s] from user [%s]", backupNameToDelete, user))
				backupMap[user] = RemoveStringItemFromSlice(backupMap[user], backupNameToDelete)
			}
			err := TaskHandler(primaryUserList, validateBackupDelete, Parallel)
			log.FailOnError(err, "failed to delete backup for primary users")

		})
		Step("Create additional secondary users and groups from Px-Admin", func() {
			log.Infof("Creating a set of %d secondary users and %v groups from Px-Admin", numOfSecondaryUsers, numberOfGroups)
			secondaryUserList = CreateUsers(numOfSecondaryUsers)
			for i := 1; i <= numberOfGroups; i++ {
				groupName := fmt.Sprintf("pxbGroup%v", i)
				wg.Add(1)
				go func(groupName string) {
					defer GinkgoRecover()
					defer wg.Done()
					err := backup.AddGroup(groupName)
					log.FailOnError(err, fmt.Sprintf("Failed to create group - %v", groupName))
					groupList = append(groupList, groupName)
				}(groupName)
			}
			wg.Wait()
		})

		Step("Share the backups created by primary users with ViewOnlyAccess with newly created secondary users and groups ", func() {
			log.Infof("Sharing the backups created by primary users with ViewOnlyAccess with newly created secondary users and groups")
			ctx, err := backup.GetNonAdminCtx(primaryUserList[0], CommonPassword)
			log.FailOnError(err, fmt.Sprintf("Failed to get context for user %s", primaryUserList[0]))
			for _, backupName := range backupMap[primaryUserList[0]] {
				err = ShareBackup(backupName, nil, secondaryUserList, ViewOnlyAccess, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Sharing backup [%s] with access type [%v]", backupName, ViewOnlyAccess))
			}
			log.Infof("Validating the shared backups with ViewOnlyAccess from one of the user")
			for _, backupName := range backupMap[primaryUserList[0]] {
				restoreName := fmt.Sprintf("%s-%s-%v", secondaryUserList[0], RestoreNamePrefix, RandomString(5))
				ValidateSharedBackupWithUsers(secondaryUserList[0], ViewOnlyAccess, backupName, restoreName)
			}
			log.InfoD("Finished verifying access level - ViewOnlyAccess")
		})

		Step("Share the backups created by primary users with RestoreAccess with newly created secondary users and groups ", func() {
			log.Infof("Sharing the backups created by primary users with RestoreAccess with newly created secondary users and groups")
			ctx, err := backup.GetNonAdminCtx(primaryUserList[1], CommonPassword)
			log.FailOnError(err, fmt.Sprintf("Failed to get context for user %s", primaryUserList[1]))
			for _, backupName := range backupMap[primaryUserList[1]] {
				err = ShareBackup(backupName, groupList, secondaryUserList, RestoreAccess, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Sharing backup [%s] with access type [%v]", backupName, RestoreAccess))
			}
			log.Infof("Validating the shared backups with RestoreAccess from one of the user")
			for _, backupName := range backupMap[primaryUserList[1]] {
				restoreName := fmt.Sprintf("%s-%s-%v", secondaryUserList[1], RestoreNamePrefix, RandomString(5))
				ValidateSharedBackupWithUsers(secondaryUserList[1], RestoreAccess, backupName, restoreName)
			}
			log.InfoD("Finished verifying access level - RestoreAccess")
		})

		Step("Share the backups created by primary users with FullAccess with newly created secondary users and groups ", func() {
			log.Infof("Sharing the backups created by primary users with FullAccess with newly created secondary users and groups")
			ctx, err := backup.GetNonAdminCtx(primaryUserList[2], CommonPassword)
			log.FailOnError(err, fmt.Sprintf("Failed to get context for user %s", primaryUserList[2]))

			for _, backupName := range backupMap[primaryUserList[2]] {
				err = ShareBackup(backupName, groupList, secondaryUserList, FullAccess, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Sharing backup [%s] with access type [%v]", backupName, FullAccess))
			}
			log.Infof("Validating the shared backups with FullAccess from one of the user")
			for _, backupName := range backupMap[primaryUserList[2]] {
				restoreName := fmt.Sprintf("%s-%s-%v", secondaryUserList[2], RestoreNamePrefix, RandomString(5))
				ValidateSharedBackupWithUsers(secondaryUserList[2], FullAccess, backupName, restoreName)
			}
			log.InfoD("Finished verifying access level - FullAccess")
		})

		Step("Create backups from admin on his cluster", func() {
			log.InfoD("Creating backups from admin on his cluster")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			backupAppContexts = FilterAppContextsByNamespace(backupAppContexts, []string{userNameSpaceMap[primaryUserList[0]]})
			adminBackupNames, err = TakeMultipleBackupsPerDeployment(ctx, BackupOrgID, SourceClusterName, clusterMap[SourceClusterName], numOfInitialBackups, snapshotLimit, adminBackupLocationName, adminBackupLocationUID, backupAppContexts, "admin-backup")
			dash.VerifyFatal(err, nil, "Creating backups from admin on his cluster")
		})

		Step("Validate Primary user can't share admin backups with secondary users and can shared own backups", func() {
			log.InfoD("Validate Primary user can't share admin backups with secondary users and can shared own backups")
			validateAdminBackupShareByPrimaryUser := func(backupName string) {
				primaryUser, err := GetSubsetOfSlice(primaryUserList, 1)
				dash.VerifyFatal(err, nil, "Fetching primary user")
				nonAdminCtx, err := backup.GetNonAdminCtx(primaryUser[0], CommonPassword)
				dash.VerifyFatal(err, nil, "Fetching non admin ctx")
				err = ShareBackup(backupName, groupList, secondaryUserList, FullAccess, nonAdminCtx)
				dash.VerifyFatal(strings.Contains(err.Error(), clusterSharePermissionError), true, fmt.Sprintf("Verifying user [%s] cant share the backup [%s]", primaryUser[0], backupName))
			}
			err := TaskHandler(adminBackupNames, validateAdminBackupShareByPrimaryUser, Sequential)
			log.FailOnError(err, "failed to validate admin backup share by primary user")

			log.Infof("Validate during cluster level backup share only owned backups are shared to secondary user")

			primaryUser := primaryUserList[0]
			primaryUserCtx, err := backup.GetNonAdminCtx(primaryUser, CommonPassword)
			secondaryUser := secondaryUserList[0]
			log.Infof("Sharing cluster level backups from the shared cluster from primary user [%s] to secondary user [%s]", primaryUser, secondaryUser)
			err = ClusterUpdateBackupShareWithClusterUid(SourceClusterName, clusterMap[SourceClusterName], nil, []string{secondaryUser}, FullAccess, true, primaryUserCtx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying cluster level backup share from primary user [%s] to secondary user [%s]", primaryUser, secondaryUser))
			primaryUserBackups := backupMap[primaryUser]
			secondaryUserBackups, err := GetAllBackupsForUser(secondaryUser, CommonPassword)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching all backups for user [%s]", secondaryUser))
			log.Infof("validate primary user backups are shared with secondary user")
			for _, backupName := range primaryUserBackups {
				if !IsPresent(secondaryUserBackups, backupName) {
					err := fmt.Errorf("backup [%s] is not shared with secondary user [%s]", backupName, secondaryUser)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying backup [%s] is shared with secondary user [%s]", backupName, secondaryUser))
				} else {
					log.Infof("Backup [%s] is shared with secondary user [%s]", backupName, secondaryUser)
					err := ValidateSharedBackupAccess(primaryUserCtx, backupName, FullAccess, secondaryUser)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying shared backup [%s] access for user [%s]", backupName, secondaryUser))
				}
			}
			log.Infof("validate admin backups are not shared with secondary user")
			for _, backupName := range adminBackupNames {
				if IsPresent(secondaryUserBackups, backupName) {
					err := fmt.Errorf("backup [%s] is shared with secondary user [%s]", backupName, secondaryUser)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying backup [%s] is shared with secondary user [%s]", backupName, secondaryUser))
				} else {
					log.Infof("Admin backup [%s] is not shared with secondary user [%s]", backupName, secondaryUser)
				}
			}
			log.Infof("Validate the primary users have restore to admin backups")
			ctx, err := backup.GetAdminCtxFromSecret()
			dash.VerifyFatal(err, nil, "Fetching admin ctx")
			err = ValidateShareClusterBackup(ctx, SourceClusterName, clusterMap[SourceClusterName], primaryUserList, nil)
			dash.VerifyFatal(err, nil, "verifying the primary users have restore access to admin backups")
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(backupAppContexts)
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		err = SetSourceKubeConfig()
		log.FailOnError(err, "Switching context to source cluster failed")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(backupAppContexts, opts)
		log.InfoD("Switching context to destination cluster for clean up")
		err = SetDestinationKubeConfig()
		log.FailOnError(err, "Unable to switch context to destination cluster [%s]", DestinationClusterName)
		DestroyApps(backupAppContexts, opts)
		log.InfoD("Switching back context to Source cluster")
		err = SetSourceKubeConfig()
		log.FailOnError(err, "Unable to switch context to source cluster [%s]", SourceClusterName)
		log.Infof("Cleaning up the backups locations and restores created by test")
		restoreNames, err := GetAllRestoresAdmin()
		dash.VerifySafely(err, nil, "Fetching all the restores")
		for _, restoreName := range restoreNames {
			restoreUid, err := Inst().Backup.GetRestoreUID(ctx, restoreName, BackupOrgID)
			dash.VerifySafely(err, nil, fmt.Sprintf("Fetching restore uid for restore [%s]", restoreName))
			err = DeleteRestoreWithUID(restoreName, restoreUid, BackupOrgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting restore [%s]", restoreName))
		}
		err = CleanupAllUserAndGroups()
		dash.VerifySafely(err, nil, "Verifying cleanup all user and groups")
		CleanupCloudSettingsAndClusters(backupLocationMap, adminCloudCredName, adminCloudCredentialUID, ctx)
	})
})

// This testcase is to validate cluster share operation while bringing down PxBackup pods
var _ = Describe("{ValidateClusterShareWhileBringDownPxBackupPods}", Label(TestCaseLabelsMap[ValidateClusterShareWhileBringDownPxBackupPods]...), func() {
	var (
		cloudAccountName                   string
		cloudAccountUid                    string
		backupLocationName                 string
		backupLocationUid                  string
		pxBackupNS                         string
		primaryUserList                    []string
		firstSharedUserList                []string
		secondSharedUserList               []string
		backupLocationMap                  map[string]string
		numDeployments                     int
		numOfPrimaryUsers                  int
		numOfSharedUsers                   int
		originalDeploymentReplicaCount     int32
		originalStatefulSetReplicaCount    int32
		scaledDownReplica                  int32
		userRoleMap                        map[string]backup.PxBackupRole
		backupDeployment                   *appsV1.Deployment
		statefulSet                        *appsV1.StatefulSet
		wg                                 sync.WaitGroup
		backupAppContexts                  []*scheduler.Context
		err                                error
		userClusterMap                     map[string]map[string]string
		backupNameSpaces                   []string
		userNamespaceMap                   map[string]string
		clusterShareOnPremCluster          []string
		isClusterShareSupportedForAllRoles bool
		userRoles                          []backup.PxBackupRole
	)

	JustBeforeEach(func() {
		numDeployments = 2
		numOfPrimaryUsers = 2
		numOfSharedUsers = 20
		userRoleMap = make(map[string]backup.PxBackupRole)
		scaledDownReplica = int32(0)
		backupAppContexts = make([]*scheduler.Context, 0)
		backupLocationMap = make(map[string]string)
		userClusterMap = make(map[string]map[string]string)
		userNamespaceMap = make(map[string]string)

		clusterShareOnPremCluster = []string{"vanilla", "openshift"}
		if Contains(clusterShareOnPremCluster, GetClusterProvider()) {
			isClusterShareSupportedForAllRoles = true
		} else {
			isClusterShareSupportedForAllRoles = false
		}

		StartPxBackupTorpedoTest("ValidateClusterShareWhileBringDownPxBackupPods", "TC to validate cluster share operation at scale while bringing down Px-Backup pods", nil, 301141, Sabrarhussaini, Q3FY25)
		log.Infof("Scheduling applications")
		err := SetSourceKubeConfig()
		log.FailOnError(err, "Switching context to source cluster failed")
		log.Infof("Scheduling applications")
		for i := 0; i < numDeployments; i++ {
			taskName := fmt.Sprintf("multiple-%d", i)
			appContexts := ScheduleApplications(taskName)
			for _, appCtx := range appContexts {
				backupAppContexts = append(backupAppContexts, appCtx)
				appCtx.ReadinessTimeout = AppReadinessTimeout
				namespace := GetAppNamespace(appCtx, taskName)
				backupNameSpaces = append(backupNameSpaces, namespace)
			}
		}
	})

	It("TC to validate cluster share at scale while bringing down Px-Backup pods", func() {
		Step("Validating applications ", func() {
			log.InfoD("Validating all the deployed applications")
			ValidateApplications(backupAppContexts)
		})

		Step("Create a set of primary users for validation", func() {
			log.Infof("Creating a set of %d primary users with different roles", numOfPrimaryUsers)
			if isClusterShareSupportedForAllRoles {
				userRoles = []backup.PxBackupRole{backup.ApplicationOwner, backup.InfrastructureOwner, backup.SuperAdmin, backup.ApplicationUser}
			} else {
				userRoles = []backup.PxBackupRole{backup.SuperAdmin, backup.InfrastructureOwner}
			}
			userList := CreateUsers(numOfPrimaryUsers + numOfSharedUsers)
			primaryUserList = userList[:numOfPrimaryUsers]
			firstSharedUserList = userList[numOfPrimaryUsers : numOfPrimaryUsers+(numOfSharedUsers/2)]
			secondSharedUserList = userList[numOfPrimaryUsers+(numOfSharedUsers/2) : numOfSharedUsers]
			for i, user := range primaryUserList {
				role := userRoles[i%len(userRoles)]
				err := backup.AddRoleToUser(user, role, fmt.Sprintf("Adding %v role to %s", role, user))
				log.FailOnError(err, fmt.Sprintf("failed to add role %s to the user %s", role, user))
				userRoleMap[user] = role
				userNamespaceMap[user] = backupNameSpaces[i]
			}
			log.Infof("Creating a set of %d secondary users to share the clusters with", numOfSharedUsers)
		})

		Step("Add Cloud credentials and Backup Location from Px-Admin", func() {
			log.InfoD(fmt.Sprintf("Adding Credentials and Backup Location from Px-admin and making it public"))
			providers := GetBackupProviders()
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			for _, provider := range providers {
				cloudAccountUid = uuid.New()
				cloudAccountName = fmt.Sprintf("autogenerated-cred-%v", RandomString(5))
				if provider != drivers.ProviderNfs {
					err = CreateCloudCredential(provider, cloudAccountName, cloudAccountUid, BackupOrgID, ctx)
					log.FailOnError(err, "Failed to create cloud credential")
					err = AddCloudCredentialOwnership(cloudAccountName, cloudAccountUid, nil, nil, Invalid, Read, ctx, BackupOrgID)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying public ownership update for cloud credential %s ", cloudAccountName))
				}
				backupLocationName = fmt.Sprintf("autogenerated-backup-location-%v", RandomString(5))
				backupLocationUid = uuid.New()
				err = CreateBackupLocationWithContext(provider, backupLocationName, backupLocationUid, cloudAccountName, cloudAccountUid, getGlobalBucketName(provider), BackupOrgID, "", ctx, true)
				log.FailOnError(err, fmt.Sprintf("Failed to add backup location %s using provider %s for px-admin user", backupLocationName, provider))
				backupLocationMap[backupLocationUid] = backupLocationName
				err = AddBackupLocationOwnership(backupLocationName, backupLocationUid, nil, nil, Invalid, Read, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying public ownership update for backup location %s", backupLocationName))
			}
		})

		Step("Registering application cluster for  creating backup object from each user", func() {
			log.InfoD("Registering application cluster for  creating backup object from each user")
			createClusterFromUser := func(user string) {
				userCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching non-admin ctx")
				err = CreateApplicationClusters(BackupOrgID, "", "", userCtx)
				dash.VerifyFatal(err, nil, "Creating source and destination cluster")
				for _, clusterName := range []string{SourceClusterName, DestinationClusterName} {
					userClusterUID, err := Inst().Backup.GetClusterUID(userCtx, BackupOrgID, clusterName)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", clusterName))
					if userClusterMap[user] == nil {
						userClusterMap[user] = make(map[string]string)
					}
					userClusterMap[user][clusterName] = userClusterUID
					log.Infof("Updated userClusterMap for user [%s] with cluster [%s] and uid [%s]", user, clusterName, userClusterUID)
				}
			}
			err := TaskHandler(primaryUserList, createClusterFromUser, Sequential)
			log.FailOnError(err, "failed to create application cluster from user")
		})

		Step("Taking manual backups of application from source cluster for each user", func() {
			log.InfoD("Taking manual backups of applications from source cluster for each user")
			createManualBackupsFromUser := func(user string) {
				defer GinkgoRecover()
				nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "failed to fetch user [%s] ctx", user)
				appContextsToBackup := FilterAppContextsByNamespace(backupAppContexts, []string{userNamespaceMap[user]})
				BackupNamePrefix := fmt.Sprintf("%s-%s-%s", "manual-backup", user, RandomString(5))
				backupNames, err := TakeMultipleBackupsPerDeployment(nonAdminCtx, BackupOrgID, SourceClusterName, userClusterMap[user][SourceClusterName], 3, 3, backupLocationName, backupLocationUid, appContextsToBackup, BackupNamePrefix)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%v]", backupNames))
			}
			err := TaskHandler(append(primaryUserList), createManualBackupsFromUser, Parallel)
			log.FailOnError(err, "failed to create manual backups from user")
		})

		Step("Get the original replica count for Px-Backup pods before cluster share", func() {
			log.InfoD("Getting the original replica count for Px-Backup pods before cluster share")
			pxBackupNS, err = backup.GetPxBackupNamespace()
			log.FailOnError(err, "Getting backup namespace")
			log.Infof("Getting the replica factor of px-backup deployment in backup namespace [%s] before cluster share", pxBackupNS)
			backupDeployment, err = apps.Instance().GetDeployment(PxBackupDeployment, pxBackupNS)
			log.FailOnError(err, fmt.Sprintf("Getting px-backup deployment replica in backup namespace %s", pxBackupNS))
			originalDeploymentReplicaCount = *backupDeployment.Spec.Replicas
			log.Infof("Replica count for px-backup pod before cluster share is %v", originalDeploymentReplicaCount)
			log.InfoD("Getting the replica factor of mongodb statefulset in backup namespace [%s] before cluster share", pxBackupNS)
			statefulSet, err = apps.Instance().GetStatefulSet(MongodbStatefulset, pxBackupNS)
			dash.VerifyFatal(err, nil, "Getting mongodb statefulset details")
			originalStatefulSetReplicaCount = *statefulSet.Spec.Replicas
			log.Infof("Number of replica for mongodb pod before cluster share is %v", originalStatefulSetReplicaCount)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Getting mongodb statefulset replica in backup namespace %s", pxBackupNS))
		})

		Step("Initiate the sharing of cluster from primary users with first set of secondary users.", func() {
			log.InfoD("Initiate the sharing from primary users with first set of secondary users.")
			shareClusterFromUser := func(user string) {
				defer GinkgoRecover()
				ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching non-admin ctx")
				for _, clusterName := range []string{SourceClusterName, DestinationClusterName} {
					log.InfoD(fmt.Sprintf("Sharing the [%s] cluster with all users", clusterName))
					_, err := ShareCluster(ctx, clusterName, userClusterMap[user][clusterName], firstSharedUserList, nil, true)
					dash.VerifySafely(err, nil, fmt.Sprintf("Verifying share of [%s] cluster", clusterName))
				}
			}
			err := TaskHandler(primaryUserList, shareClusterFromUser, Parallel)
			log.FailOnError(err, "Failed to share the cluster with first set of secondary users")
		})

		Step("Scale Px-backup deployment replica count to 0 and back to original replica while cluster share is in progress", func() {
			log.InfoD("Scaling Px-backup deployment replica count to 0 while cluster share is in progress")
			err = ScaleDeploymentReplicas(PxBackupDeployment, pxBackupNS, scaledDownReplica, 0, PodStatusTimeOut, PodStatusRetryTime)
			dash.VerifyFatal(err, nil, "Scaling down PxBackup Deployment replica to 0")
			log.InfoD("Scaling PxBackup Deployment to original replica")
			log.Infof("Sleeping for 1 minute")
			time.Sleep(1 * time.Minute)
			err = ScaleDeploymentReplicas(PxBackupDeployment, pxBackupNS, originalDeploymentReplicaCount, originalDeploymentReplicaCount, PodStatusTimeOut, PodStatusRetryTime)
			dash.VerifyFatal(err, nil, "Scaling back PxBackup Deployment to original replica count")
			err = ValidatePxBackupIsReady()
			dash.VerifyFatal(err, nil, "Validating px-backup pod is ready")
		})

		Step("Validate shared cluster for first set of secondary users.", func() {
			log.InfoD("Validating shared cluster for first set of secondary users.")
			validateClusterShareForUsers := func(user string) {
				ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching non-admin ctx")
				for _, clusterName := range []string{SourceClusterName, DestinationClusterName} {
					err = ValidateShareCluster(ctx, clusterName, userClusterMap[user][clusterName], firstSharedUserList, nil)
					log.FailOnError(err, "Validating if the cluster is shared")
				}
			}
			err := TaskHandler(primaryUserList, validateClusterShareForUsers, Parallel)
			log.FailOnError(err, "Failed to validate the shared cluster for users")
		})

		Step("Validate backups for shared clusters for users.", func() {
			log.InfoD("Validating backups for shared clusters for users.")
			validateBackupsForClusterShareForUsers := func(user string) {
				ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching non-admin ctx")
				for _, clusterName := range []string{SourceClusterName, DestinationClusterName} {
					err = ValidateShareClusterBackup(ctx, clusterName, userClusterMap[user][clusterName], firstSharedUserList, nil)
					log.FailOnError(err, "Validating if the backups are shared")
				}
			}
			err := TaskHandler(primaryUserList, validateBackupsForClusterShareForUsers, Sequential)
			log.FailOnError(err, "Failed to validate the shared cluster for users")
		})

		Step("Initiate the sharing of cluster from primary users with second set of secondary users.", func() {
			log.InfoD("Initiate the sharing of cluster from primary users with second set of secondary users.")
			shareClusterFromUser := func(user string) {
				defer GinkgoRecover()
				ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching non-admin ctx")
				for _, clusterName := range []string{SourceClusterName, DestinationClusterName} {
					log.InfoD(fmt.Sprintf("Sharing the [%s] cluster with all users", clusterName))
					_, err := ShareCluster(ctx, clusterName, userClusterMap[user][clusterName], secondSharedUserList, nil, true)
					dash.VerifySafely(err, nil, fmt.Sprintf("Verifying share of [%s] cluster", clusterName))
				}
			}
			err := TaskHandler(primaryUserList, shareClusterFromUser, Parallel)
			log.FailOnError(err, "Failed to share the cluster with first set of secondary users")
		})

		Step("Scale MongoDB statefulset replica to 0 and back to original replica while cluster share is in progress", func() {
			log.InfoD("Scaling MongoDB statefulset replica to 0 while cluster share is in progress")
			err := ScaleStatefulSetReplicas(MongodbStatefulset, pxBackupNS, scaledDownReplica, 0, PodStatusTimeOut, PodStatusRetryTime)
			dash.VerifyFatal(err, nil, "Scaling down MongoDB statefulset replica to 0")
			log.InfoD("Sleeping for 1 minute for the pods be scaled")
			time.Sleep(1 * time.Minute)
			err = ScaleStatefulSetReplicas(MongodbStatefulset, pxBackupNS, originalStatefulSetReplicaCount, 2, PodStatusTimeOut, PodStatusRetryTime)
			dash.VerifyFatal(err, nil, "Scaling back MongoDB statefulset to original replica count")
			err = IsMongoDBReady()
			log.FailOnError(err, "Checking if mongo db pod is in running state")
		})

		Step("Validate shared clusters for second set of secondary users.", func() {
			log.InfoD("Validating shared clusters for second set of secondary users..")
			validateClusterShareForUsers := func(user string) {
				ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching non-admin ctx")
				for _, clusterName := range []string{SourceClusterName, DestinationClusterName} {
					err = ValidateShareCluster(ctx, clusterName, userClusterMap[user][clusterName], secondSharedUserList, nil)
					log.FailOnError(err, "Validating if the cluster is shared for user")
				}
			}
			err := TaskHandler(primaryUserList, validateClusterShareForUsers, Parallel)
			log.FailOnError(err, "Failed to validate the cluster share for users")
		})

		Step("Validate backups for shared clusters for users.", func() {
			log.InfoD("Validating backups for shared clusters for users.")
			validateBackupShareForUsers := func(user string) {
				ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching non-admin ctx")
				for _, clusterName := range []string{SourceClusterName, DestinationClusterName} {
					err = ValidateShareClusterBackup(ctx, clusterName, userClusterMap[user][clusterName], secondSharedUserList, nil)
					log.FailOnError(err, "Validating if the backups are shared for user")
				}
			}
			err := TaskHandler(primaryUserList, validateBackupShareForUsers, Sequential)
			log.FailOnError(err, "Failed to validate the backups for clustered share for users")
		})

		Step("Initiate unshare of shared cluster from first set of secondary users.", func() {
			log.InfoD("Initiating unshare of shared cluster from first set of secondary users.")
			for _, user := range primaryUserList {
				wg.Add(1)
				go func(user string) {
					defer wg.Done()
					defer GinkgoRecover()
					ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
					log.FailOnError(err, "Fetching non-admin ctx")
					for _, clusterName := range []string{SourceClusterName, DestinationClusterName} {
						log.InfoD(fmt.Sprintf("Unsharing the [%s] cluster from all users", clusterName))
						_, err = UnShareCluster(ctx, clusterName, userClusterMap[user][clusterName], firstSharedUserList, nil)
						dash.VerifySafely(err, nil, fmt.Sprintf("Verifying unshare of [%s] cluster", clusterName))
					}
				}(user)
			}
			wg.Wait()
		})

		Step("Scale Px-backup deployment replica count to 0 and back to original replicas while cluster share is in progress", func() {
			log.InfoD("Scaling px-backup deployment replica count to 0 while cluster share is in progress")
			err = ScaleDeploymentReplicas(PxBackupDeployment, pxBackupNS, scaledDownReplica, 0, PodStatusTimeOut, PodStatusRetryTime)
			dash.VerifyFatal(err, nil, "Scaling down PxBackup Deployment replica to 0")
			log.InfoD("Scaling PxBackup Deployment to original replica")
			log.Infof("Sleeping for 1 minute")
			time.Sleep(1 * time.Minute)
			err = ScaleDeploymentReplicas(PxBackupDeployment, pxBackupNS, originalDeploymentReplicaCount, originalDeploymentReplicaCount, PodStatusTimeOut, PodStatusRetryTime)
			dash.VerifyFatal(err, nil, "Scaling back PxBackup Deployment to original replica count")
			err = ValidatePxBackupIsReady()
			dash.VerifyFatal(err, nil, "Validating px-backup pod is ready")
		})

		Step("Validate unshare of shared cluster from first set of secondary users", func() {
			log.InfoD("validate unshare of shared cluster from first set of secondary users")
			validateClusterUnshareForUsers := func(user string) {
				ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching non-admin ctx")
				for _, clusterName := range []string{SourceClusterName, DestinationClusterName} {
					err = ValidateUnShareCluster(ctx, clusterName, userClusterMap[user][clusterName], firstSharedUserList, nil)
					log.FailOnError(err, "Validating if the clusters are unshared")
				}
			}
			err := TaskHandler(primaryUserList, validateClusterUnshareForUsers, Parallel)
			log.FailOnError(err, "Failed to validate the cluster share for users")
		})

		Step("Initiate unshare of shared cluster from second set of secondary users.", func() {
			log.InfoD("Initiating unshare of shared cluster from first set of secondary users")
			for _, user := range primaryUserList {
				wg.Add(1)
				go func(user string) {
					defer wg.Done()
					defer GinkgoRecover()
					ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
					log.FailOnError(err, "Fetching non-admin ctx")
					for _, clusterName := range []string{SourceClusterName, DestinationClusterName} {
						log.InfoD(fmt.Sprintf("Unsharing the [%s] cluster from all users", clusterName))
						_, err = UnShareCluster(ctx, clusterName, userClusterMap[user][clusterName], secondSharedUserList, nil)
						dash.VerifySafely(err, nil, fmt.Sprintf("Verifying unshare of [%s] cluster", clusterName))
					}
				}(user)
			}
			wg.Wait()
		})

		Step("Scale MongoDB statefulset replica to 0 and back to original replicas while cluster unshare is in progress", func() {
			log.InfoD("Scaling MongoDB statefulset replica to 0 while cluster unshare is in progress")
			err := ScaleStatefulSetReplicas(MongodbStatefulset, pxBackupNS, scaledDownReplica, 0, PodStatusTimeOut, PodStatusRetryTime)
			dash.VerifyFatal(err, nil, "Scaling down MongoDB statefulset replica to 0")
			log.InfoD("Scaling MongoDB statefulset to original replica")
			log.Infof("Sleeping for 1 minute")
			time.Sleep(1 * time.Minute)
			err = ScaleStatefulSetReplicas(MongodbStatefulset, pxBackupNS, originalStatefulSetReplicaCount, originalStatefulSetReplicaCount, PodStatusTimeOut, PodStatusRetryTime)
			dash.VerifyFatal(err, nil, "Scaling back MongoDB statefulset to original replica count")
			err = IsMongoDBReady()
			log.FailOnError(err, "Checking if mongo db pod is in running state")
		})

		Step("Validate unshare of cluster from users.", func() {
			log.InfoD("Validating unshare of cluster from users.")
			validateClusterUnshareForUsers := func(user string) {
				ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
				log.FailOnError(err, "Fetching non-admin ctx")
				for _, clusterName := range []string{SourceClusterName, DestinationClusterName} {
					err = ValidateUnShareCluster(ctx, clusterName, userClusterMap[user][clusterName], secondSharedUserList, nil)
					log.FailOnError(err, "Validating if the clusters are unshared")
				}
			}
			err := TaskHandler(primaryUserList, validateClusterUnshareForUsers, Parallel)
			log.FailOnError(err, "Failed to validate the cluster share for users")
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(backupAppContexts)
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		err = SetSourceKubeConfig()
		log.FailOnError(err, "Switching context to source cluster failed")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(backupAppContexts, opts)
		log.InfoD("Deleting the backups")
		allBackups, err := GetAllBackupsAdmin()
		dash.VerifySafely(err, nil, "Verifying fetching of all backups")
		for _, bkp := range allBackups {
			wg.Add(1)
			go func(bkp string) {
				defer wg.Done()
				backupUID, err := Inst().Backup.GetBackupUID(ctx, bkp, BackupOrgID)
				_, err = DeleteBackup(bkp, backupUID, BackupOrgID, ctx)
				dash.VerifySafely(err, nil, fmt.Sprintf("Verifying backup deletion - %s", bkp))
			}(bkp)
		}
		wg.Wait()
		log.InfoD("Switching context to destination cluster for clean up")
		err = SetDestinationKubeConfig()
		log.FailOnError(err, "Unable to switch context to destination cluster [%s]", DestinationClusterName)
		DestroyApps(backupAppContexts, opts)
		log.InfoD("Switching back context to Source cluster")
		err = SetSourceKubeConfig()
		log.FailOnError(err, "Unable to switch context to source cluster [%s]", SourceClusterName)
		err = CleanupAllUserAndGroups()
		dash.VerifySafely(err, nil, "Verifying cleanup all user and groups")
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudAccountName, cloudAccountUid, ctx)
	})
})
