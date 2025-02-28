package tests

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/pure-px/torpedo/drivers"
	"github.com/pure-px/torpedo/drivers/backup"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
)

// Share cluster from admin cxt to any user with backups and validate the restore and delete.
var _ = Describe("{ClusterBackupShareWithUserGroup}", Label(TestCaseLabelsMap[SuperAdmin]...), func() {
	var (
		scheduledAppContexts []*scheduler.Context
		appContextsToBackup  []*scheduler.Context
		bkpNamespaces        []string
		clusterUid           string
		clusterStatus        api.ClusterInfo_StatusInfo_Status

		pxbUsers          []string
		backupName        string
		backupLocationUID string
		cloudCredName     string
		cloudCredUID      string
		bkpLocationName   string
		group1            string
		group2            string
		backupNames       []string
	)

	userContexts := make([]context.Context, 0)
	backupLocationMap := make(map[string]string)
	labelSelectors := make(map[string]string)
	bkpNamespaces = make([]string, 0)
	backupNames = make([]string, 0)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyClusterBackupShareWithUserGroup", "Verify Cluster backup share from admin ctx to any user with backups ad restores", nil, 82930, Pingle, Q2FY25)
		scheduledAppContexts = make([]*scheduler.Context, 0)
		bkpNamespaces = make([]string, 0)

		// Schedule an Application
		appContexts := ScheduleApplications(TaskNamePrefix)
		for _, ctx := range appContexts {
			ctx.ReadinessTimeout = AppReadinessTimeout
			namespace := GetAppNamespace(ctx, TaskNamePrefix)
			bkpNamespaces = append(bkpNamespaces, namespace)
			scheduledAppContexts = append(scheduledAppContexts, ctx)
		}
	})

	//Share cluster from admin ctx to any user and validate delete and restore operations.
	It("Verify cluster share from admin ctx to any user with backups, schedules and restore creation", func() {
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching admin ctx")
		numberOfUsers := 4
		numOfBackup := 2
		// 1. validate application
		Step("Validate applications", func() {
			log.InfoD("Validating applications")
			ValidateApplications(scheduledAppContexts)
		})

		//2. Create 4 user
		Step("Create Users", func() {
			log.InfoD("Creating users testuser")
			pxbUsers = CreateUsers(numberOfUsers)
			log.Infof("Created %v users and users list is %v", numberOfUsers, pxbUsers)
		})

		// 3. Create 2 groups one with full access and one with view only access.
		Step("Create Groups", func() {
			log.InfoD("Creating 2 user group for view only access and one for full access")
			group1 = "group-with-full-access"
			err := backup.AddGroup(group1)
			log.FailOnError(err, "Failed to create user group-1- %v", group1)

			group2 = "group-with-view-access"
			err = backup.AddGroup(group2)
			log.FailOnError(err, "Failed to create user group-2- %v", group2)
		})

		// 4. Add users to the group1 and group2
		Step("Add pxb users to group of pxb-users", func() {
			log.InfoD("Adding users to group1")
			for _, pxbUser := range pxbUsers[0:3] {
				err := backup.AddGroupToUser(pxbUser, group1)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Adding user %s to group %s", pxbUser, group1))
			}

			log.InfoD("Adding users to group2")
			for _, pxbUser := range pxbUsers[2:4] {
				err := backup.AddGroupToUser(pxbUser, group2)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Adding user %s to group %s", pxbUser, group2))
			}

			usersOfGroup, err := backup.GetMembersOfGroup(group1)
			log.FailOnError(err, "Error fetching members of the group - %v", group1)
			log.Infof("Group [%v] contains the following users: \n%v", group1, usersOfGroup)

			usersOfGroup, err = backup.GetMembersOfGroup(group2)
			log.FailOnError(err, "Error fetching members of the group - %v", group2)
			log.Infof("Group [%v] contains the following users: \n%v", group2, usersOfGroup)
		})

		// 5. Create backup location and cloud setting
		Step("Creating backup location and cloud setting", func() {
			log.InfoD("Creating backup location and cloud setting")
			providers := GetBackupProviders()
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cloudcred", provider, time.Now().Unix())
				bkpLocationName = fmt.Sprintf("%s-%s-%v-bl", provider, getGlobalBucketName(provider), time.Now().Unix())
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = bkpLocationName
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocation(provider, bkpLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", bkpLocationName))
			}
		})

		// 6. Create application cluster for backup.
		Step("Register cluster for backup", func() {
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
			log.InfoD("Uid of [%s] cluster is %s", SourceClusterName, clusterUid)
		})

		// 7. Take backup of cluster
		Step("Taking backup of applications", func() {
			for i := 0; i < numOfBackup; i++ {
				backupName := fmt.Sprintf("%s-%s-%v", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
				appContextsToBackup = FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
				err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, bkpLocationName, backupLocationUID, appContextsToBackup, labelSelectors, BackupOrgID, clusterUid, "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				backupNames = append(backupNames, backupName)
			}
		})

		// 8. Share cluster backup with group1 with full access
		Step("Share backup with group having full access", func() {
			log.InfoD("Share all backups with Full Access in source cluster with a group")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = ClusterUpdateBackupShare(SourceClusterName, []string{group1}, nil, FullAccess, true, ctx)
			log.FailOnError(err, "Failed sharing all backups for cluster [%s]", SourceClusterName)
		})

		// 9. Share cluster backup with group2 with view only access
		Step("Share the cluster backups with a user belonging to the group with view only access.", func() {
			log.InfoD("Share the cluster backups with a user belonging to the group with view only access.")
			err = ClusterUpdateBackupShare(SourceClusterName, []string{group2}, nil, ViewOnlyAccess, true, ctx)
			log.FailOnError(err, "Failed to share backup %s", backupName)
		})

		//10. User from full access group can backup should be deletable and restoreable.
		Step("From shared user's context , the backup should be restorable and deletable.", func() {
			log.InfoD("Backup should be restorable and deletable for common and non common both user from full access group.")
			chosenUser, err := backup.GetRandomUserFromGroup(group1)
			log.FailOnError(err, "Failed to get a random user from group [%s]", group1)
			log.Infof("Sharing backup with user - %s", chosenUser)

			// Get Admin Context - needed to share backup and get backup UID
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			// Get user context
			ctxNonAdmin, err := backup.GetNonAdminCtx(chosenUser, CommonPassword)
			log.FailOnError(err, "Fetching non admin ctx")
			userContexts = append(userContexts, ctxNonAdmin)

			// Register Source and Destination cluster
			log.InfoD("Registering Source and Destination clusters from user context")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctxNonAdmin)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			// Start Restore
			userDestClusterUid, err := Inst().Backup.GetClusterUID(ctxNonAdmin, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))
			restoreName := fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
			appContextsToBackup = FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			err = CreateRestoreWithValidation(ctxNonAdmin, restoreName, backupNames[0], make(map[string]string), make(map[string]string), DestinationClusterName, userDestClusterUid, BackupOrgID, appContextsToBackup)
			log.FailOnError(err, "Restoring of backup [%s] has failed with name - [%s]", backupNames[0], restoreName)

			// Restore validation to make sure that the user with Full Access can restore
			log.InfoD("Restoring of backup [%s] was successful with name - [%s]", backupNames[0], restoreName)
			log.Infof("About to delete restore - %s", restoreName)
			err = DeleteRestore(restoreName, BackupOrgID, ctxNonAdmin)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting Restore %s", restoreName))

			// Get Backup UID
			backupDriver := Inst().Backup
			backupUID, err := backupDriver.GetBackupUID(ctx, backupNames[0], BackupOrgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupNames[0])

			// Delete backup to confirm that the user has Full Access
			backupDeleteResponse, err := DeleteBackup(backupNames[0], backupUID, BackupOrgID, ctxNonAdmin)
			log.FailOnError(err, "Backup [%s] could not be deleted by user [%s]", backupNames[0], chosenUser)
			dash.VerifyFatal(backupDeleteResponse.String(), "", fmt.Sprintf("Verifying backup %s deletion", backupNames[0]))
		})

		//11. Validate user from view only group doesnt have permission to backup,delete and restore options
		Step("Validate Restore Access of view Acess backups shared at cluster level", func() {
			log.InfoD("Validate View Access of backups shared at cluster level for a user of a group")
			// Get user from group
			var err error
			backupName := backupNames[1]
			chosenUser := pxbUsers[len(pxbUsers)-1]
			log.FailOnError(err, "Failed to get a random user from group [%s]", group2)
			log.Infof("User chosen to validate View access - %s", chosenUser)

			// Get Admin Context - needed to share backup and get backup UID
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			// Get user context
			ctxNonAdmin, err := backup.GetNonAdminCtx(chosenUser, CommonPassword)
			log.FailOnError(err, "Fetching non admin ctx")

			// Register Source and Destination cluster
			log.InfoD("Registering Source and Destination clusters from user context")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctxNonAdmin)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")

			// Start Restore
			userDestClusterUid, err := Inst().Backup.GetClusterUID(ctxNonAdmin, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))
			restoreName := fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
			appContextsToBackup = FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)

			// Restore validation to make sure that the user with View Access cannot restore
			err = CreateRestoreWithValidation(ctxNonAdmin, restoreName, backupName, make(map[string]string), make(map[string]string), DestinationClusterName, userDestClusterUid, BackupOrgID, appContextsToBackup)
			dash.VerifyFatal(err != nil, true, "doesn't have permission to restore with view only access")

			// Get Backup UID
			backupDriver := Inst().Backup
			backupUID, err := backupDriver.GetBackupUID(ctx, backupName, BackupOrgID)
			log.InfoD("Error msg for GetBackupUID %v:", err)

			// Delete backup to confirm that the user cannot delete the backup
			_, err = DeleteBackup(backupName, backupUID, BackupOrgID, ctxNonAdmin)
			dash.VerifyFatal(err != nil, true, "doesn't have permission to delete with view only access")
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)

		// Clean the all user created
		var wg sync.WaitGroup
		log.Infof("Cleaning up users")
		for _, userName := range pxbUsers {
			wg.Add(1)
			go func(userName string) {
				defer GinkgoRecover()
				defer wg.Done()
				err := backup.DeleteUser(userName)
				log.FailOnError(err, "Error deleting user %v", userName)
			}(userName)
		}
		wg.Wait()

		// Clean up the all group created
		log.Infof("Cleaning up groups")
		err := backup.DeleteGroup(group1)
		log.FailOnError(err, "Error deleting user %v", group1)
		err = backup.DeleteGroup(group2)
		log.FailOnError(err, "Error deleting user %v", group2)

		// Clean up the cluster
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)

	})
})

// Share cluster from admin ctx to user1 and with group with existing backups and validate restore and delete.
var _ = Describe("{ClusterBackupShareToUserAndGroup}", Label(TestCaseLabelsMap[SuperAdmin]...), func() {
	var (
		scheduledAppContexts []*scheduler.Context
		appContextsToBackup  []*scheduler.Context
		bkpNamespaces        []string
		clusterUid           string
		clusterStatus        api.ClusterInfo_StatusInfo_Status

		pxbUsers          []string
		backupName        string
		backupLocationUID string
		cloudCredName     string
		cloudCredUID      string
		bkpLocationName   string
		group1            string
		backupNames       []string
	)

	userContexts := make([]context.Context, 0)
	backupLocationMap := make(map[string]string)
	labelSelectors := make(map[string]string)
	bkpNamespaces = make([]string, 0)
	backupNames = make([]string, 0)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyClusterClusterBackupShareToUserAndGroup", "Verify Cluster backup share from admin ctx to any user or group with backups ad restores", nil, 80153, "Pingle", Q2FY25)
		scheduledAppContexts = make([]*scheduler.Context, 0)
		bkpNamespaces = make([]string, 0)

		// Schedule an Application
		appContexts := ScheduleApplications(TaskNamePrefix)
		for _, ctx := range appContexts {
			ctx.ReadinessTimeout = AppReadinessTimeout
			namespace := GetAppNamespace(ctx, TaskNamePrefix)
			bkpNamespaces = append(bkpNamespaces, namespace)
			scheduledAppContexts = append(scheduledAppContexts, ctx)
		}

	})

	// Share cluster from admin ctx to user1 and with group with existing backups and validate restore and delete.
	It("Verify cluster share from admin ctx to any user or group with backups, schedules and restore creation", func() {
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching admin ctx")
		numberOfUsers := 2
		numOfBackup := 2
		// 1. validate application
		Step("Validate applications", func() {
			log.InfoD("Validating applications")
			ValidateApplications(scheduledAppContexts)
		})

		//2. Create 2 user
		Step("Create Users", func() {
			log.InfoD("Creating users testuser")
			pxbUsers = CreateUsers(numberOfUsers)
			log.Infof("Created %v users and users list is %v", numberOfUsers, pxbUsers)
		})
		// 3. Create groups one with full access.
		Step("Create Groups", func() {
			log.InfoD("Creating user group ")
			group1 = "group1"
			err := backup.AddGroup(group1)
			log.FailOnError(err, "Failed to create user group-1- %v", group1)
		})
		// 4. Add user2 to the group1
		Step("Add user2 to group of pxb-users", func() {
			log.InfoD("Adding user2 to group1")
			err := backup.AddGroupToUser(pxbUsers[1], group1)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Adding user %s to group %s", pxbUsers[1], group1))

			usersOfGroup, err := backup.GetMembersOfGroup(group1)
			log.FailOnError(err, "Error fetching members of the group - %v", group1)
			log.Infof("Group [%v] contains the following users: \n%v", group1, usersOfGroup)
		})

		// 5. Create backup location and cloud setting
		Step("Creating backup location and cloud setting", func() {
			log.InfoD("Creating backup location and cloud setting")
			providers := GetBackupProviders()
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cloudcred", provider, time.Now().Unix())
				bkpLocationName = fmt.Sprintf("%s-%s-%v-bl", provider, getGlobalBucketName(provider), time.Now().Unix())
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = bkpLocationName
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocation(provider, bkpLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", bkpLocationName))
			}
		})

		// 6. Create application cluster for backup.
		Step("Register cluster for backup", func() {
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
			log.InfoD("Uid of [%s] cluster is %s", SourceClusterName, clusterUid)
		})
		// 7. Take the backup of bkpnamespaces
		Step("Taking backup of applications", func() {
			for i := 0; i < numOfBackup; i++ {
				backupName := fmt.Sprintf("%s-%s-%v", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
				appContextsToBackup = FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
				err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, bkpLocationName, backupLocationUID, appContextsToBackup, labelSelectors, BackupOrgID, clusterUid, "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				backupNames = append(backupNames, backupName)
			}
		})

		// 8. Share cluster backup with group1 with full access
		Step("Share backup with group having full access", func() {
			log.InfoD("Share all backups with Full Access in source cluster with a group")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = ClusterUpdateBackupShare(SourceClusterName, []string{group1}, nil, FullAccess, true, ctx)
			log.FailOnError(err, "Failed sharing all backups for cluster [%s]", SourceClusterName)
		})

		// 8. Share cluster backup with user1 with full access
		Step("Share backup with user1 having full access", func() {
			log.InfoD("Share backup with user1 having full access")
			err = ClusterUpdateBackupShare(SourceClusterName, nil, pxbUsers[0:1], FullAccess, true, ctx)
			log.FailOnError(err, "Failed to share backup %s", backupName)
		})

		// 9. User from full access group can backup should be deletable and restoreable.
		Step("From shared user's context , the backup should be restorable and deletable.", func() {
			log.InfoD("Backup should be restorable and deletable for common and non common both user from full access group.")
			chosenUser, err := backup.GetRandomUserFromGroup(group1)
			log.FailOnError(err, "Failed to get a random user from group [%s]", group1)
			log.Infof("Sharing backup with user - %s", chosenUser)

			// Get Admin Context - needed to share backup and get backup UID
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			// Get user context
			ctxNonAdmin, err := backup.GetNonAdminCtx(chosenUser, CommonPassword)
			log.FailOnError(err, "Fetching non admin ctx")
			userContexts = append(userContexts, ctxNonAdmin)

			// Register Source and Destination cluster
			log.InfoD("Registering Source and Destination clusters from user context")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctxNonAdmin)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			// Start Restore
			userDestClusterUid, err := Inst().Backup.GetClusterUID(ctxNonAdmin, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))
			restoreName := fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
			appContextsToBackup = FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			err = CreateRestoreWithValidation(ctxNonAdmin, restoreName, backupNames[0], make(map[string]string), make(map[string]string), DestinationClusterName, userDestClusterUid, BackupOrgID, appContextsToBackup)
			log.FailOnError(err, "Restoring of backup [%s] has failed with name - [%s]", backupNames[0], restoreName)

			// Restore validation to make sure that the user with Full Access can restore
			log.InfoD("Restoring of backup [%s] was successful with name - [%s]", backupNames[0], restoreName)
			log.Infof("About to delete restore - %s", restoreName)
			err = DeleteRestore(restoreName, BackupOrgID, ctxNonAdmin)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting Restore %s", restoreName))

			// Get Backup UID
			backupDriver := Inst().Backup
			backupUID, err := backupDriver.GetBackupUID(ctx, backupNames[0], BackupOrgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupNames[0])

			// Delete backup to confirm that the user has Full Access
			backupDeleteResponse, err := DeleteBackup(backupNames[0], backupUID, BackupOrgID, ctxNonAdmin)
			log.FailOnError(err, "Backup [%s] could not be deleted by user [%s]", backupNames[0], chosenUser)
			dash.VerifyFatal(backupDeleteResponse.String(), "", fmt.Sprintf("Verifying backup %s deletion", backupNames[0]))
		})

		// 10. Test User1 can perform backup , delete and restore operations.
		Step("From shared user's context , the backup should be restorable and deletable.", func() {
			log.InfoD("Backup should be restorable and deletable for common and non common both user from full access group.")
			chosenUser := pxbUsers[0]
			log.Infof("Sharing backup with user - %s", chosenUser)

			// Get Admin Context - needed to share backup and get backup UID
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			// Get user context
			ctxNonAdmin, err := backup.GetNonAdminCtx(chosenUser, CommonPassword)
			log.FailOnError(err, "Fetching non admin ctx")
			userContexts = append(userContexts, ctxNonAdmin)

			// Register Source and Destination cluster
			log.InfoD("Registering Source and Destination clusters from user context")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctxNonAdmin)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			// Start Restore
			userDestClusterUid, err := Inst().Backup.GetClusterUID(ctxNonAdmin, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))
			restoreName := fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
			appContextsToBackup = FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			err = CreateRestoreWithValidation(ctxNonAdmin, restoreName, backupNames[1], make(map[string]string), make(map[string]string), DestinationClusterName, userDestClusterUid, BackupOrgID, appContextsToBackup)
			log.FailOnError(err, "Restoring of backup [%s] has failed with name - [%s]", backupNames[1], restoreName)

			// Restore validation to make sure that the user with Full Access can restore
			log.InfoD("Restoring of backup [%s] was successful with name - [%s]", backupNames[1], restoreName)
			log.Infof("About to delete restore - %s", restoreName)
			err = DeleteRestore(restoreName, BackupOrgID, ctxNonAdmin)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting Restore %s", restoreName))

			// Get Backup UID
			backupDriver := Inst().Backup
			backupUID, err := backupDriver.GetBackupUID(ctx, backupNames[1], BackupOrgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupNames[1])

			// Delete backup to confirm that the user has Full Access
			backupDeleteResponse, err := DeleteBackup(backupNames[1], backupUID, BackupOrgID, ctxNonAdmin)
			log.FailOnError(err, "Backup [%s] could not be deleted by user [%s]", backupNames[1], chosenUser)
			dash.VerifyFatal(backupDeleteResponse.String(), "", fmt.Sprintf("Verifying backup %s deletion", backupNames[1]))
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)

		// Clean up the all users
		var wg sync.WaitGroup
		log.Infof("Cleaning up users")
		for _, userName := range pxbUsers {
			wg.Add(1)
			go func(userName string) {
				defer GinkgoRecover()
				defer wg.Done()
				err := backup.DeleteUser(userName)
				log.FailOnError(err, "Error deleting user %v", userName)
			}(userName)
		}
		wg.Wait()

		// clean up the all groups
		log.Infof("Cleaning up groups")
		err := backup.DeleteGroup(group1)
		log.FailOnError(err, "Error deleting user %v", group1)

		// clean up the clusters
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)

	})
})

// Share cluster from admin ctx to user1 and with group with existing backups and validate backup are visible in their view and restore .
var _ = Describe("{ClusterBackupShareToUserAndGroupWithRestore}", Label(TestCaseLabelsMap[BackupShare]...), func() {
	var (
		scheduledAppContexts []*scheduler.Context
		appContextsToBackup  []*scheduler.Context
		bkpNamespaces        []string
		clusterUid           string
		clusterStatus        api.ClusterInfo_StatusInfo_Status

		pxbUsers          []string
		backupName        string
		backupLocationUID string
		cloudCredName     string
		cloudCredUID      string
		bkpLocationName   string
		groupName         string
		backupNames       []string
		restoreList       []string
	)

	userContexts := make([]context.Context, 0)
	backupLocationMap := make(map[string]string)
	labelSelectors := make(map[string]string)
	bkpNamespaces = make([]string, 0)
	backupNames = make([]string, 0)
	restoreList = make([]string, 0)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyClusterBackupShareToUserAndGroupWithRestore", "Verify Cluster backup share from admin ctx to any user or group with restores", nil, 80152, Pingle, Q2FY25)
		scheduledAppContexts = make([]*scheduler.Context, 0)
		bkpNamespaces = make([]string, 0)

		// Schedule an Application
		appContexts := ScheduleApplications(TaskNamePrefix)
		for _, ctx := range appContexts {
			ctx.ReadinessTimeout = AppReadinessTimeout
			namespace := GetAppNamespace(ctx, TaskNamePrefix)
			bkpNamespaces = append(bkpNamespaces, namespace)
			scheduledAppContexts = append(scheduledAppContexts, ctx)
		}

	})

	It("Verify cluster share from admin ctx to any user and group with backups, should be visible in their view and validate restore operation.", func() {
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching admin ctx")
		numberOfUsers := 2
		numOfBackup := 2
		// 1. validate application
		Step("Validate applications", func() {
			log.InfoD("Validating applications")
			ValidateApplications(scheduledAppContexts)
		})

		//2. Create 2 user
		Step("Create Users", func() {
			log.InfoD("Creating users testuser")
			pxbUsers = CreateUsers(numberOfUsers)
			log.Infof("Created %v users and users list is %v", numberOfUsers, pxbUsers)
		})
		// 3. Create groups one.
		Step("Create Groups", func() {
			log.InfoD("Creating user group ")
			groupName = fmt.Sprintf("%s-%s", "group", RandomString(5))
			err := backup.AddGroup(groupName)
			log.FailOnError(err, "Failed to create user groupName- %v", groupName)
		})
		// 4. Add user2 to the group
		Step("Add user2 to group of pxb-users", func() {
			log.InfoD("Adding user2 to group")
			err := backup.AddGroupToUser(pxbUsers[1], groupName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Adding user %s to group %s", pxbUsers[1], groupName))

			usersOfGroup, err := backup.GetMembersOfGroup(groupName)
			log.FailOnError(err, "Error fetching members of the group - %v", groupName)
			log.Infof("Group [%v] contains the following users: \n%v", groupName, usersOfGroup)
		})

		// 5. Create backup location and cloud setting
		Step("Creating backup location and cloud setting", func() {
			log.InfoD("Creating backup location and cloud setting")
			providers := GetBackupProviders()
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cloudcred", provider, time.Now().Unix())
				bkpLocationName = fmt.Sprintf("%s-%s-%v-bl", provider, getGlobalBucketName(provider), time.Now().Unix())
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = bkpLocationName
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocation(provider, bkpLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", bkpLocationName))
			}
		})

		// 6. Create application cluster for backup.
		Step("Register cluster for backup", func() {
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
			log.InfoD("Uid of [%s] cluster is %s", SourceClusterName, clusterUid)
		})
		// 7. Take the backup of bkNamespaces.
		Step("Taking backup of applications", func() {
			for i := 0; i < numOfBackup; i++ {
				backupName := fmt.Sprintf("%s-%s-%v", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
				appContextsToBackup = FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
				err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, bkpLocationName, backupLocationUID, appContextsToBackup, labelSelectors, BackupOrgID, clusterUid, "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				backupNames = append(backupNames, backupName)
			}
		})

		// 8. Share cluster backup with group with restore access
		Step("Share backup with group having restore access", func() {
			log.InfoD("Share all backups with restore Access in source cluster with a group")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = ClusterUpdateBackupShare(SourceClusterName, []string{groupName}, nil, RestoreAccess, true, ctx)
			log.FailOnError(err, "Failed sharing all backups for cluster [%s]", SourceClusterName)
		})

		// 8. Share cluster backup with user1 with restore access
		Step("Share backup with user1 having restore access", func() {
			log.InfoD("Share backup with user1 having restore access")
			err = ClusterUpdateBackupShare(SourceClusterName, nil, pxbUsers[0:1], RestoreAccess, true, ctx)
			log.FailOnError(err, "Failed to share backup %s", backupName)
		})

		// 9.Users are able to see shared backup in their view
		Step("Users are able to see shared backup in their view", func() {
			log.InfoD("Users are able to see shared backup in their view")
			chosenUser, err := backup.GetRandomUserFromGroup(groupName)
			log.FailOnError(err, "Failed to get a random user from group [%s]", groupName)
			log.Infof("Sharing backup with user - %s", chosenUser)

			// Get user context
			ctxNonAdmin, err := backup.GetNonAdminCtx(chosenUser, CommonPassword)
			log.FailOnError(err, "Fetching non admin ctx")
			userContexts = append(userContexts, ctxNonAdmin)
			for _, backupName := range backupNames {
				bkpUid, err := Inst().Backup.GetBackupUID(ctxNonAdmin, backupName, BackupOrgID)
				log.FailOnError(err, "Fetching backup uid")
				backupInspectRequest := &api.BackupInspectRequest{
					Name:  backupName,
					Uid:   bkpUid,
					OrgId: BackupOrgID,
				}

				_, err = Inst().Backup.InspectBackup(ctxNonAdmin, backupInspectRequest)
				dash.VerifyFatal(err, nil, "Backup are visible to user")
			}
		})

		// 10. User from restore access group can backup should be restoreable.
		Step("From shared user's context , the backup should be restorable.", func() {
			log.InfoD("Backup should be restorable common and non common both user from restore access group.")
			chosenUser, err := backup.GetRandomUserFromGroup(groupName)
			log.FailOnError(err, "Failed to get a random user from group [%s]", groupName)
			log.Infof("Sharing backup with user - %s", chosenUser)

			// Get user context
			ctxNonAdmin, err := backup.GetNonAdminCtx(chosenUser, CommonPassword)
			log.FailOnError(err, "Fetching non admin ctx")
			userContexts = append(userContexts, ctxNonAdmin)

			// Register Source and Destination cluster
			log.InfoD("Registering Source and Destination clusters from user context")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctxNonAdmin)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			// Start Restore
			userDestClusterUid, err := Inst().Backup.GetClusterUID(ctxNonAdmin, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))
			restoreName := fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
			appContextsToBackup = FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			err = CreateRestoreWithValidation(ctxNonAdmin, restoreName, backupNames[0], make(map[string]string), make(map[string]string), DestinationClusterName, userDestClusterUid, BackupOrgID, appContextsToBackup)
			log.FailOnError(err, "Restoring of backup [%s] has failed with name - [%s]", backupNames[0], restoreName)
			restoreList = append(restoreList, restoreName)

			// Get Backup UID
			backupDriver := Inst().Backup
			backupUID, err := backupDriver.GetBackupUID(ctx, backupNames[0], BackupOrgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupNames[0])

			// Delete backup to confirm that the user cant delete backup with Restore Access
			_, err = DeleteBackup(backupNames[0], backupUID, BackupOrgID, ctxNonAdmin)
			dash.VerifyNotNilFatal(err, "doesn't have permission to delete with restore only access")
		})

		// 11. Test User1 can perform restore opterations.
		Step("From shared user's context , the backup should be restorable.", func() {
			log.InfoD("Backup should be restorable for common and non common both user from Restore access group.")
			chosenUser := pxbUsers[0]
			log.Infof("Sharing backup with user - %s", chosenUser)

			// Get user context
			ctxNonAdmin, err := backup.GetNonAdminCtx(chosenUser, CommonPassword)
			log.FailOnError(err, "Fetching non admin ctx")
			userContexts = append(userContexts, ctxNonAdmin)

			// Register Source and Destination cluster
			log.InfoD("Registering Source and Destination clusters from user context")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctxNonAdmin)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			// Start Restore
			userDestClusterUid, err := Inst().Backup.GetClusterUID(ctxNonAdmin, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))
			restoreName := fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
			appContextsToBackup = FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			err = CreateRestoreWithValidation(ctxNonAdmin, restoreName, backupNames[1], make(map[string]string), make(map[string]string), DestinationClusterName, userDestClusterUid, BackupOrgID, appContextsToBackup)
			log.FailOnError(err, "Restoring of backup [%s] has failed with name - [%s]", backupNames[1], restoreName)
			restoreList = append(restoreList, restoreName)

			// Get Backup UID
			backupDriver := Inst().Backup
			backupUID, err := backupDriver.GetBackupUID(ctx, backupNames[0], BackupOrgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupNames[0])

			// Delete backup to confirm that the user cant delete backup with Restore Access
			_, err = DeleteBackup(backupNames[0], backupUID, BackupOrgID, ctxNonAdmin)
			dash.VerifyNotNilFatal(err, "doesn't have permission to delete with restore only access")
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)

		// Clean up the all users
		var wg sync.WaitGroup
		log.Infof("Cleaning up users")
		for _, userName := range pxbUsers {
			wg.Add(1)
			go func(userName string) {
				defer GinkgoRecover()
				defer wg.Done()
				err := backup.DeleteUser(userName)
				log.FailOnError(err, "Error deleting user %v", userName)
			}(userName)
		}
		wg.Wait()

		// clean up the all groups
		log.Infof("Cleaning up groups")
		err := backup.DeleteGroup(groupName)
		log.FailOnError(err, "Error deleting user %v", groupName)

		// clean up the clusters
		ctx, err := backup.GetAdminCtxFromSecret()
		for _, restoreName := range restoreList {
			err = DeleteRestore(restoreName, BackupOrgID, ctx)
		}
		log.FailOnError(err, "Fetching px-central-admin ctx")
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)

	})
})

// Share Individual backup with user having view-only access and cluster backup with full access, then validate restore and delete
var _ = Describe("{CheckAccessForUserAfterClusterBackupShareWithFullAccess}", Label(TestCaseLabelsMap[ValidateUserAccessLevel]...), func() {
	var (
		backupName           string
		scheduledAppContexts []*scheduler.Context
		userContexts         []context.Context
		appContextsToBackup  []*scheduler.Context
		bkpNamespaces        []string
		clusterUid           string
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		cloudCredName        string
		cloudCredUID         string
		backupLocationUID    string
		providers            []string
		backupLocationMap    map[string]string
		backupLocation       string
		backupNameList       []string
		customUser           string
		backupCount          int
		customRoleName       backup.PxBackupRole = backup.InfrastructureOwner
	)
	JustBeforeEach(func() {
		bkpNamespaces = make([]string, 0)
		backupLocationMap = make(map[string]string)
		userContexts = make([]context.Context, 0)
		providers = GetBackupProviders()

		StartPxBackupTorpedoTest("CheckAccessForUserAfterClusterBackupShareWithFullAccess", "Verify user can perform delete and restore operations after changing access", nil, 80162, Prikumar, Q3FY25)
		log.InfoD("Deploy applications")

		scheduledAppContexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = AppReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
				scheduledAppContexts = append(scheduledAppContexts, ctx)
			}
		}
	})
	// Verify user can perform delete and restore after backup share access change
	It("Validate and check the access level for user", func() {
		Step("Validate applications", func() {
			log.Infof("Validate applications")
			ValidateApplications(scheduledAppContexts)

		})

		// Creating cloud credentials and backup location
		Step("Creating cloud credentials and backup location", func() {
			log.InfoD("Creating cloud credentials and backup location")
			providers = GetBackupProviders()
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
				cloudCredUID = uuid.New()
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				log.InfoD("Created Cloud Credentials with name - %s", cloudCredName)
				backupLocation = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocation
				err = CreateBackupLocation(provider, backupLocation, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %v", backupLocation))
				log.InfoD("Created Backup Location with name - %s", backupLocation)

			}
		})

		// Create cluster for backup
		Step("Register cluster for backup", func() {
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})

		// Create new user
		Step(fmt.Sprintf("Create a new user [%s]", customRoleName), func() {
			customUser = fmt.Sprintf("testuser-%s", RandomString(5))
			firstName := fmt.Sprintf("FirstName-%s", customUser)
			lastName := fmt.Sprintf("LastName-%s", customUser)
			email := fmt.Sprintf("%v@cnbu.com", customUser)
			err := backup.AddUser(customUser, firstName, lastName, email, CommonPassword)
			log.FailOnError(err, "Failed to create user - %s", customUser)
		})

		// Taking multiple backups
		Step("Taking multiple backups of application on source cluster", func() {
			backupCount = 4
			log.InfoD("Taking %v multiple backups of application on source cluster", backupCount)
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			log.InfoD("Taking Backup of application")
			for i := 0; i < backupCount; i++ {
				backupName = fmt.Sprintf("%s-%v-%v", BackupNamePrefix, i+1, time.Now().Unix())
				err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, backupLocation, backupLocationUID, appContextsToBackup, nil, BackupOrgID, clusterUid, "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				backupNameList = append(backupNameList, backupName)
			}
			log.Infof("List of backups - %v", backupNameList)
		})

		// Share single backup with user having viewonly access
		Step("Share backup with user having viewonly access", func() {
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			log.InfoD("Share backup with user having viewonly access")
			err = ShareBackup(backupName, nil, []string{customUser}, ViewOnlyAccess, ctx)
			log.FailOnError(err, "Failed to share backup %s", backupName)
		})

		// Share cluster backup with user having full access
		Step("Trigger cluster backup share with full access", func() {
			log.InfoD("Share all backups with full access")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = ClusterUpdateBackupShare(SourceClusterName, nil, []string{customUser}, FullAccess, true, ctx)
			log.FailOnError(err, "Failed sharing all backups for cluster")
		})

		// Verify user can do restore and delete after full access
		Step("Validate Full Access of backups shared at cluster level", func() {
			// Get user
			log.InfoD("Validate Full Access of backups shared at cluster level for an individual user - %s", customUser)
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			// Get user context
			ctxNonAdmin, err := backup.GetNonAdminCtx(customUser, CommonPassword)
			log.FailOnError(err, "Fetching non admin ctx")
			userContexts = append(userContexts, ctxNonAdmin)

			// Register Source and Destination cluster
			log.InfoD("Registering Source and Destination clusters from user context")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctxNonAdmin)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")

			userDestClusterUid, err := Inst().Backup.GetClusterUID(ctxNonAdmin, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))

			// Start Restore
			restoreName := fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
			appContextsToBackup = FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			log.InfoD(" backup-name: [%s] restore-name: [%s] backup-location: [%s]", backupNameList[0], restoreName, backupLocation)
			err = CreateRestoreWithValidation(ctxNonAdmin, restoreName, backupNameList[0], make(map[string]string), make(map[string]string), DestinationClusterName, userDestClusterUid, BackupOrgID, appContextsToBackup)
			log.FailOnError(err, "Restoring of backup [%s] has failed with name - [%s]", backupNameList[0], restoreName)

			// Restore validation to make sure that the user with Full Access can restore
			log.InfoD("Restoring of backup [%s] was successful with name - [%s]", backupNameList[0], restoreName)
			log.Infof("About to delete restore - %s", restoreName)
			err = DeleteRestore(restoreName, BackupOrgID, ctxNonAdmin)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting Restore %s", restoreName))

			// Get Backup UID
			backupDriver := Inst().Backup
			backupUID, err := backupDriver.GetBackupUID(ctx, backupNameList[0], BackupOrgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupNameList[0])

			// Delete backup to confirm that the user has Full Access
			backupDeleteResponse, err := DeleteBackup(backupNameList[0], backupUID, BackupOrgID, ctxNonAdmin)
			log.FailOnError(err, "Backup [%s] could not be deleted by user [%s]", backupNameList[0], customUser)
			dash.VerifyFatal(backupDeleteResponse.String(), "",
				fmt.Sprintf("Verifying backup [%s] deletion is successful by user [%s]", backupNameList[0], customUser))
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		// Cleaning up applications created
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)

		// Cleaning up px-backup cluster
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})

// Share Individual backup with group having view-only access and cluster backup with full access, then validate restore and delete
var _ = Describe("{CheckAccessForGroupAfterClusterBackupShareWithFullAccess}", Label(TestCaseLabelsMap[ValidateUserAccessLevel]...), func() {
	var (
		scheduledAppContexts []*scheduler.Context
		appContextsToBackup  []*scheduler.Context
		bkpNamespaces        []string
		clusterUid           string
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		groupName            string
		userNames            []string
		backupName           string
		backupLocationUID    string
		cloudCredName        string
		cloudCredUID         string
		bkpLocationName      string
		backupNameList       []string
	)
	userContexts := make([]context.Context, 0)
	backupLocationMap := make(map[string]string)
	bkpNamespaces = make([]string, 0)
	numberOfUsers := 1
	JustBeforeEach(func() {
		bkpNamespaces = make([]string, 0)
		backupLocationMap = make(map[string]string)
		userContexts = make([]context.Context, 0)

		StartPxBackupTorpedoTest("CheckAccessForGroupAfterClusterBackupShareWithFullAccess", "Verify users from group can perform delete and restore operations after changing access", nil, 80163, Prikumar, Q3FY25)
		log.InfoD("Deploy applications")

		scheduledAppContexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = AppReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
				scheduledAppContexts = append(scheduledAppContexts, ctx)
			}
		}
	})
	// Verify user from group can perform delete and restore after backup share access change
	It("Validate and check the access level for group", func() {
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		Step("Validate applications", func() {
			log.InfoD("Validate applications ")
			ValidateApplications(scheduledAppContexts)
		})

		// Create new user
		Step("Create Users", func() {
			log.InfoD("Creating users")
			userNames = CreateUsers(numberOfUsers)
			log.Infof("Created %v users and users list is %v", numberOfUsers, userNames)
		})

		// Create groups
		Step("Create Groups", func() {
			log.InfoD("Creating group testGroup")
			groupName = fmt.Sprintf("testGroup")
			err := backup.AddGroup(groupName)
			log.FailOnError(err, "Failed to create group - %v", groupName)
		})

		// Add users to group
		Step("Add users to group", func() {
			log.InfoD("Adding user to groups")
			err := backup.AddGroupToUser(userNames[0], groupName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Adding user %s to group %s", userNames[0], groupName))
			usersOfGroup, err := backup.GetMembersOfGroup(groupName)
			log.FailOnError(err, "Error fetching members of the group - %v", groupName)
			log.Infof("Group [%v] contains the following users: \n%v", groupName, usersOfGroup)
		})

		// Create cloud credentials and backup location
		Step("Creating cloud credentials and backup location", func() {
			log.InfoD("Creating cloud credentials and backup location")
			providers := GetBackupProviders()
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cloudcred", provider, time.Now().Unix())
				bkpLocationName = fmt.Sprintf("%s-%s-%v-bl", provider, getGlobalBucketName(provider), time.Now().Unix())
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = bkpLocationName
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocation(provider, bkpLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", bkpLocationName))
			}
		})

		// Create application cluster for backup
		Step("Register cluster for backup", func() {
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
			log.InfoD("Uid of [%s] cluster is %s", SourceClusterName, clusterUid)
		})

		// Taking multiple backups
		Step("Taking multiple backups of application on source cluster", func() {
			backupCount := 4
			log.InfoD("Taking %v multiple backups of application on source cluster", backupCount)
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			log.InfoD("Taking Backup of application")
			for i := 0; i < backupCount; i++ {
				backupName = fmt.Sprintf("%s-%v-%v", BackupNamePrefix, i+1, time.Now().Unix())
				err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, bkpLocationName, backupLocationUID, appContextsToBackup, nil, BackupOrgID, clusterUid, "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				backupNameList = append(backupNameList, backupName)
			}
			log.Infof("List of backups - %v", backupNameList)
		})

		// Share single backup with group having viewonly access
		Step("Share backup with group having viewonly access", func() {
			log.InfoD("Share backup with group having viewonly access")
			err = ShareBackup(backupName, []string{groupName}, nil, ViewOnlyAccess, ctx)
			log.FailOnError(err, "Failed to share backup %s", backupName)
		})

		// Share cluster backup with group having full access
		Step("Trigger cluster backup share with full access", func() {
			log.InfoD("Share all backups with group having full access")
			err = ClusterUpdateBackupShare(SourceClusterName, []string{groupName}, nil, FullAccess, true, ctx)
			log.FailOnError(err, "Failed sharing all backups for cluster")
		})

		// Verify user from group can do restore and delete after full access
		Step("Validate Full Access of user from group shared at cluster level", func() {
			log.InfoD("Validate Full Access of user from group shared at cluster level")

			// Get user context
			ctxNonAdmin, err := backup.GetNonAdminCtx(userNames[0], CommonPassword)
			log.FailOnError(err, "Fetching non admin ctx")
			userContexts = append(userContexts, ctxNonAdmin)

			// Register Source and Destination cluster
			log.InfoD("Registering Source and Destination clusters from user context")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctxNonAdmin)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			userDestClusterUid, err := Inst().Backup.GetClusterUID(ctxNonAdmin, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))
			restoreName := fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())

			// Try restore with user having fullAccess and it should pass
			err = CreateRestoreWithValidation(ctxNonAdmin, restoreName, backupName, make(map[string]string), make(map[string]string), DestinationClusterName, userDestClusterUid, BackupOrgID, appContextsToBackup)
			log.FailOnError(err, "Restoring of backup [%s] has failed with name - [%s]", backupName, restoreName)
			log.InfoD("Restoring of backup [%s] was successful with name - [%s]", backupName, restoreName)
			err = DeleteRestore(restoreName, BackupOrgID, ctxNonAdmin)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting Restore %s", restoreName))

			backupDriver := Inst().Backup
			backupUID, err := backupDriver.GetBackupUID(ctx, backupName, BackupOrgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupName)

			// Delete backup to confirm that the user from group has Full Access
			backupDeleteResponse, err := DeleteBackup(backupName, backupUID, BackupOrgID, ctxNonAdmin)
			log.FailOnError(err, "Backup [%s] could not be deleted by user [%s] with delete response %s", backupName, userNames, backupDeleteResponse)
			dash.VerifyFatal(backupDeleteResponse.String(), "",
				fmt.Sprintf("Verifying backup [%s] deletion is successful by user [%s]", backupName, userNames))

		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)

		// Clean up the all users
		var wg sync.WaitGroup
		log.Infof("Cleaning up users")
		for _, userName := range userNames {
			wg.Add(1)
			go func(userName string) {
				defer GinkgoRecover()
				defer wg.Done()
				err := backup.DeleteUser(userName)
				log.FailOnError(err, "Error deleting user %v", userName)
			}(userName)
		}
		wg.Wait()

		// clean up the all groups
		log.Infof("Cleaning up groups")
		err := backup.DeleteGroup(groupName)
		log.FailOnError(err, "Error deleting group %v", groupName)

		// clean up the clusters
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)

	})
})

// Modify cluster backup share after the previous backup share is done.
var _ = Describe("{ModifyClusterBackupSharePostPreviousCompletion}", Label(TestCaseLabelsMap[ValidateUserAccessLevel]...), func() {
	var (
		backupName           string
		scheduledAppContexts []*scheduler.Context
		userContexts         []context.Context
		bkpNamespaces        []string
		clusterUid           string
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		cloudCredName        string
		cloudCredUID         string
		backupLocationUID    string
		providers            []string
		backupLocationMap    map[string]string
		backupLocation       string
		backupNameList       []string
		customUser           string
		backupCount          int
		customRoleName       backup.PxBackupRole = backup.InfrastructureOwner
	)
	JustBeforeEach(func() {
		bkpNamespaces = make([]string, 0)
		backupLocationMap = make(map[string]string)
		userContexts = make([]context.Context, 0)
		providers = GetBackupProviders()

		StartPxBackupTorpedoTest("VerifyModifyClusterBackupSharePostPreviousCompletion", "Modify cluster backup share after the previous backup share is done.", nil, 80161, Pingle, Q3FY25)
		log.InfoD("Deploy applications")

		scheduledAppContexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = AppReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
				scheduledAppContexts = append(scheduledAppContexts, ctx)
			}
		}
	})
	//1. Verify user can perform delete and restore after access level change
	It("Validate user can perform delete and restore after access level change", func() {
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")

		Step("Validate applications", func() {
			log.Infof("Validate applications")
			ValidateApplications(scheduledAppContexts)

		})

		//2. Creating cloud credentials and backup location
		Step("Creating cloud credentials and backup location", func() {
			log.InfoD("Creating cloud credentials and backup location")
			providers = GetBackupProviders()
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
				cloudCredUID = uuid.New()
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				log.InfoD("Created Cloud Credentials with name - %s", cloudCredName)
				backupLocation = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocation
				err = CreateBackupLocation(provider, backupLocation, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %v", backupLocation))
				log.InfoD("Created Backup Location with name - %s", backupLocation)

			}
		})

		//3. Create cluster for backup
		Step("Register cluster for backup", func() {
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})

		//4. Create new user
		Step(fmt.Sprintf("Create a new user [%s]", customRoleName), func() {
			customUser = fmt.Sprintf("testuser-%s", RandomString(5))
			firstName := fmt.Sprintf("FirstName-%s", customUser)
			lastName := fmt.Sprintf("LastName-%s", customUser)
			email := fmt.Sprintf("%v@cnbu.com", customUser)
			err := backup.AddUser(customUser, firstName, lastName, email, CommonPassword)
			log.FailOnError(err, "Failed to create user - %s", customUser)
		})

		//5. Taking backup
		Step("Taking backups of application on source cluster", func() {
			backupCount = 1
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			log.InfoD("Taking Backup of application")
			for i := 0; i < backupCount; i++ {
				backupName = fmt.Sprintf("%s-%v-%v", BackupNamePrefix, i+1, time.Now().Unix())
				err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, backupLocation, backupLocationUID, appContextsToBackup, nil, BackupOrgID, clusterUid, "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				backupNameList = append(backupNameList, backupName)
			}
			log.Infof("List of backups - %v", backupNameList)
		})

		//6. Share cluster backup to user with view access
		Step("Share backup to user with View only access", func() {
			log.InfoD("Share backup to user with view access")
			err = ClusterUpdateBackupShare(SourceClusterName, nil, []string{customUser}, ViewOnlyAccess, true, ctx)
			log.FailOnError(err, "Failed to share backup %v with view access", backupName)
		})

		//7. Share cluster backup to user with full access
		Step("Share backup to user with full access", func() {
			log.InfoD("Share backup to user with full access")
			err = ClusterUpdateBackupShare(SourceClusterName, nil, []string{customUser}, FullAccess, true, ctx)
			log.FailOnError(err, "Failed to share backup %s with full access", backupName)
		})

		//8. Verify that backups are modifying with correct backupshare based on the cluster level sharing.
		Step("Validate Full Access of backups shared at cluster level", func() {
			// Get user
			log.InfoD("Validate Full Access of backups shared at cluster level for an individual user - %s", customUser)

			ctxNonAdmin, err := backup.GetNonAdminCtx(customUser, CommonPassword)
			log.FailOnError(err, "Fetching non admin ctx")
			userContexts = append(userContexts, ctxNonAdmin)

			// Get Backup UID
			backupDriver := Inst().Backup
			backupUID, err := backupDriver.GetBackupUID(ctx, backupNameList[0], BackupOrgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupNameList[0])

			// Delete backup to confirm that the user has Full Access
			backupDeleteResponse, err := DeleteBackup(backupNameList[0], backupUID, BackupOrgID, ctxNonAdmin)
			log.FailOnError(err, "Backup [%s] could not be deleted by user [%s]", backupNameList[0], customUser)
			dash.VerifyFatal(backupDeleteResponse.String(), "",
				fmt.Sprintf("Verifying backup [%s] deletion is successful by user [%s]", backupNameList[0], customUser))
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		// Cleaning up applications created
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)

		// Cleaning up px-backup cluster
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})

// Share cluster from user1 to user2 with full access and try to delete backups from admin ctx
var _ = Describe("{ClusterBackupShareFromUserToUserAndDeleteBkpFromAdmin}", Label(TestCaseLabelsMap[BackupShare]...), func() {
	var (
		scheduledAppContexts []*scheduler.Context
		appContextsToBackup  []*scheduler.Context
		bkpNamespaces        []string
		clusterUid           string
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		pxbUsers             []string
		backupName           string
		backupLocationUID    string
		cloudCredName        string
		cloudCredUID         string
		bkpLocationName      string
		backupNames          []string
		infraAdminUsers      []string
		infraAdminRole       backup.PxBackupRole = backup.InfrastructureOwner
	)
	backupLocationMap := make(map[string]string)
	labelSelectors := make(map[string]string)
	bkpNamespaces = make([]string, 0)
	backupNames = make([]string, 0)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("ClusterBackupShareFromUserToUserAndDeleteBkpFromAdmin", "Verify Cluster backup share from user to user with full access and try to delete backups from admin", nil, 300837, Prikumar, Q3FY25)
		scheduledAppContexts = make([]*scheduler.Context, 0)
		bkpNamespaces = make([]string, 0)

		// Schedule an Application
		appContexts := ScheduleApplications(TaskNamePrefix)
		for _, ctx := range appContexts {
			ctx.ReadinessTimeout = AppReadinessTimeout
			namespace := GetAppNamespace(ctx, TaskNamePrefix)
			bkpNamespaces = append(bkpNamespaces, namespace)
			scheduledAppContexts = append(scheduledAppContexts, ctx)
		}
	})

	// Share cluster backup from user to user and try to delete backups from admin ctx
	It("Verify cluster backup share from user1 to user2 with full access and delete backups from admin", func() {
		numberOfUsers := 2
		numOfBackup := 2
		// validate application
		Step("Validate applications", func() {
			log.InfoD("Validating applications")
			ValidateApplications(scheduledAppContexts)
		})

		// Create 2 users
		Step("Create Users", func() {
			log.InfoD("Creating users")
			pxbUsers = CreateUsers(numberOfUsers)
			log.Infof("Created %v users and list is %v", numberOfUsers, pxbUsers)
		})

		// Adding role to users
		Step("Adding role to users", func() {
			log.InfoD(fmt.Sprintf("Creating %d users with %s role", numberOfUsers, infraAdminRole))
			for _, user := range pxbUsers {
				err := backup.AddRoleToUser(user, infraAdminRole, fmt.Sprintf("Adding %v role to %s", infraAdminRole, user))
				log.FailOnError(err, "failed to add role %s to the user %s", infraAdminRole, user)
				infraAdminUsers = append(infraAdminUsers, user)
			}
		})

		// Create cloud credentials and backup location
		Step("Creating cloud credentials and backup location", func() {
			log.InfoD("Creating cloud credentials and backup location")
			nonAdminCtx, err := backup.GetNonAdminCtx(infraAdminUsers[0], CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", infraAdminUsers[0])
			providers := GetBackupProviders()
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cloudcred", provider, time.Now().Unix())
				bkpLocationName = fmt.Sprintf("%s-%s-%v-bl", provider, getGlobalBucketName(provider), time.Now().Unix())
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = bkpLocationName
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocationWithContext(provider, bkpLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", nonAdminCtx, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", bkpLocationName))
			}
		})

		// Create application cluster for backup
		Step("Register cluster for backup", func() {
			nonAdminCtx, err := backup.GetNonAdminCtx(infraAdminUsers[0], CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", infraAdminUsers[0])
			err = CreateApplicationClusters(BackupOrgID, "", "", nonAdminCtx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, nonAdminCtx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(nonAdminCtx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
			log.InfoD("Uid of [%s] cluster is %s", SourceClusterName, clusterUid)
		})

		// Taking backup of applications
		Step("Taking backup of applications", func() {
			nonAdminCtx, err := backup.GetNonAdminCtx(infraAdminUsers[0], CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", infraAdminUsers[0])
			for i := 0; i < numOfBackup; i++ {
				backupName = fmt.Sprintf("%s-%s-%v", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
				appContextsToBackup = FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
				err := CreateBackupWithValidation(nonAdminCtx, backupName, SourceClusterName, bkpLocationName, backupLocationUID, appContextsToBackup, labelSelectors, BackupOrgID, clusterUid, "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				backupNames = append(backupNames, backupName)
			}
			log.InfoD("List of backups [%v]", backupNames)
		})

		// Share cluster backup with user2 having full access
		Step("Trigger cluster backup share to user2 with full access", func() {
			nonAdminCtx, err := backup.GetNonAdminCtx(infraAdminUsers[0], CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", infraAdminUsers[0])
			err = ClusterUpdateBackupShare(SourceClusterName, nil, []string{infraAdminUsers[1]}, FullAccess, true, nonAdminCtx)
			log.FailOnError(err, "Failed sharing all backups for cluster")
		})

		// Verify admin can delete backups after cluster backup share with user2
		Step("Verify admin can delete backups after cluster backup share with user2", func() {
			log.InfoD("Backup should be deletable from admin")

			// Get Admin Context
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			// Get Backup UID
			backupDriver := Inst().Backup
			backupUID, err := backupDriver.GetBackupUID(ctx, backupNames[0], BackupOrgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupNames[0])

			// Verify admin can delete the backup
			backupDeleteResponse, err := DeleteBackup(backupNames[0], backupUID, BackupOrgID, ctx)
			log.FailOnError(err, "Backup [%s] could not be deleted by admin [%v]", backupNames[0], ctx)
			dash.VerifyFatal(backupDeleteResponse.String(), "", fmt.Sprintf("Verifying backup %s deletion", backupNames[0]))
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)

		// Clean up the all users
		var wg sync.WaitGroup
		log.Infof("Cleaning up users")
		for _, userName := range pxbUsers {
			wg.Add(1)
			go func(userName string) {
				defer GinkgoRecover()
				defer wg.Done()
				err := backup.DeleteUser(userName)
				log.FailOnError(err, "Error deleting user %v", userName)
			}(userName)
		}
		wg.Wait()

		// clean up the clusters
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)

	})
})

// Individual backup share to group from a normal user with full access, try to delete it from admin user
var _ = Describe("{ShareBackupWithFullAccessToGroupAndAdminDeletionAttempt}", Label(TestCaseLabelsMap[BackupShare]...), func() {
	var (
		backupName           string
		scheduledAppContexts []*scheduler.Context
		bkpNamespaces        []string
		clusterUid           string
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		cloudCredName        string
		cloudCredUID         string
		backupLocationUID    string
		providers            []string
		backupLocationMap    map[string]string
		backupLocation       string
		customRoleName       backup.PxBackupRole = backup.InfrastructureOwner
		userList             []string
		user1Ctx             context.Context
		adminCtx             context.Context
		groupName            string
		restoreName          string
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyShareBackupWithFullAccessToGroupAndAdminDeletionAttempt", "Verify admin can perform delete operations.", nil, 300840, Pingle, Q3FY25)
		bkpNamespaces = make([]string, 0)
		backupLocationMap = make(map[string]string)

		log.InfoD("Deploy applications")

		scheduledAppContexts = make([]*scheduler.Context, 0)
		// Schedule an Application
		appContexts := ScheduleApplications(TaskNamePrefix)
		for _, ctx := range appContexts {
			ctx.ReadinessTimeout = AppReadinessTimeout
			namespace := GetAppNamespace(ctx, TaskNamePrefix)
			bkpNamespaces = append(bkpNamespaces, namespace)
			scheduledAppContexts = append(scheduledAppContexts, ctx)
		}
	})

	// Verify admin can perform delete after backup share to group.
	It("Verify admin can perform delete after backup share", func() {
		// 1. Validate applications
		Step("Validate applications", func() {
			log.Infof("Validate applications")
			ValidateApplications(scheduledAppContexts)

		})

		//2. Create 2 users with infra-admin role
		Step(fmt.Sprintf("Create a new user [%s]", customRoleName), func() {
			userList = CreateUsers(2)
			for _, user := range userList {
				err := backup.AddRoleToUser(user, customRoleName, fmt.Sprintf("Adding %v role to %s", customRoleName, user))
				log.FailOnError(err, fmt.Sprintf("failed to add role %s to the user %s", customRoleName, user))
			}
		})

		// 3. Create groups one.
		Step("Create Groups", func() {
			log.InfoD("Creating user group ")
			groupName = fmt.Sprintf("%s-%s", "group", RandomString(5))
			err := backup.AddGroup(groupName)
			log.FailOnError(err, "Failed to create user groupName- %v", groupName)
		})
		// 4. Add user2 to the group
		Step("Add user2 to group", func() {
			log.InfoD("Adding user2 to group")
			err := backup.AddGroupToUser(userList[1], groupName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Adding user %s to group %s", userList[1], groupName))

			usersOfGroup, err := backup.GetMembersOfGroup(groupName)
			log.FailOnError(err, "Error fetching members of the group - %v", groupName)
			log.Infof("Group [%v] contains the following users: \n%v", groupName, usersOfGroup)
		})

		//3. common init
		Step("Common initialization.", func() {
			var err error
			providers = GetBackupProviders()
			user1Ctx, err = backup.GetNonAdminCtx(userList[0], CommonPassword)
			log.FailOnError(err, "Fetching px-central-infra-admin user1 ctx")
			adminCtx, err = backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

		})

		//4. Create cloud credentials.
		Step("Creating cloud credentials", func() {
			log.InfoD("Creating cloud credentials")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
				cloudCredUID = uuid.New()
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, user1Ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
			}
		})

		//5. Create backup location
		Step("Creating backup location", func() {
			log.InfoD("Creating backup location")
			for _, provider := range providers {
				backupLocation = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocation
				err := CreateBackupLocationWithContext(provider, backupLocation, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", user1Ctx, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocation))
			}
		})

		//6. Create cluster for backup
		Step("Register cluster for backup", func() {
			err := CreateApplicationClusters(BackupOrgID, "", "", user1Ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, user1Ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(user1Ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})

		//7. Take backup
		Step("Taking backup of applications", func() {
			log.InfoD("Taking Backup of application")
			backupName = fmt.Sprintf("%s-%v", BackupNamePrefix, time.Now().Unix())
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			err := CreateBackupWithValidation(user1Ctx, backupName, SourceClusterName, backupLocation, backupLocationUID, appContextsToBackup, nil, BackupOrgID, clusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupName))
		})

		//8. Share backup to group with full access
		Step("Share backup to group with full access", func() {
			log.InfoD("Share backup with group having full access")
			err := ShareBackup(backupName, []string{groupName}, nil, FullAccess, user1Ctx)
			log.FailOnError(err, "Failed to share backup with group %s", backupName)
		})

		//9. Validate access from user2
		Step("Validate access from user2", func() {
			restoreName = fmt.Sprintf("%s-%v", RestoreNamePrefix, RandomString(5))
			ValidateSharedBackupWithUsers(userList[1], FullAccess, backupName, restoreName)
		})

		//10. Delete backup from admin user
		Step("Verify delete backup from admin ctx", func() {
			backupDriver := Inst().Backup
			backupUID, err := backupDriver.GetBackupUID(user1Ctx, backupName, BackupOrgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - [%s]", backupName)

			log.InfoD("Deleting backup")
			err = DeleteBackupAndWaitForCompletion(backupName, backupUID, BackupOrgID, adminCtx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting backup [%s]", backupName))

		})

	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		// Cleaning up applications created
		err := DeleteRestore(restoreName, BackupOrgID, adminCtx)
		dash.VerifySafely(err, nil, fmt.Sprintf("Deleting restore [%s]", restoreName))

		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)

		// Clean up the all users
		var wg sync.WaitGroup
		log.Infof("Cleaning up users")
		for _, userName := range userList {
			wg.Add(1)
			go func(userName string) {
				defer GinkgoRecover()
				defer wg.Done()
				err := backup.DeleteUser(userName)
				log.FailOnError(err, "Error deleting user %v", userName)
			}(userName)
		}
		wg.Wait()

		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, user1Ctx)
	})
})

// Individual backup share to a normal user from a normal user with full access, try to delete it from admin user
var _ = Describe("{ShareBackupWithFullAccessAndAdminDeletionAttempt}", Label(TestCaseLabelsMap[BackupShare]...), func() {
	var (
		backupName           string
		scheduledAppContexts []*scheduler.Context
		bkpNamespaces        []string
		clusterUid           string
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		cloudCredName        string
		cloudCredUID         string
		backupLocationUID    string
		providers            []string
		backupLocationMap    map[string]string
		backupLocation       string
		customRoleName       backup.PxBackupRole = backup.InfrastructureOwner
		userList             []string
		user1Ctx             context.Context
		user2Ctx             context.Context
		adminCtx             context.Context
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyShareBackupWithFullAccessAndAdminDeletionAttempt", "Verify admin can perform delete operations.", nil, 300840, Pingle, Q3FY25)
		bkpNamespaces = make([]string, 0)
		backupLocationMap = make(map[string]string)

		log.InfoD("Deploy applications")

		scheduledAppContexts = make([]*scheduler.Context, 0)
		// Schedule an Application
		appContexts := ScheduleApplications(TaskNamePrefix)
		for _, ctx := range appContexts {
			ctx.ReadinessTimeout = AppReadinessTimeout
			namespace := GetAppNamespace(ctx, TaskNamePrefix)
			bkpNamespaces = append(bkpNamespaces, namespace)
			scheduledAppContexts = append(scheduledAppContexts, ctx)
		}
	})

	// Verify admin can perform delete after backup share.
	It("Verify admin can perform delete after backup share", func() {
		// 1. Validate applications
		Step("Validate applications", func() {
			log.Infof("Validate applications")
			ValidateApplications(scheduledAppContexts)

		})

		//2. Create 2 users with infra-admin role
		Step(fmt.Sprintf("Create a new user [%s]", customRoleName), func() {
			userList = CreateUsers(2)
			for _, user := range userList {
				err := backup.AddRoleToUser(user, customRoleName, fmt.Sprintf("Adding %v role to %s", customRoleName, user))
				log.FailOnError(err, fmt.Sprintf("failed to add role %s to the user %s", customRoleName, user))
			}
		})
		//3. common init
		Step("Common initialization.", func() {
			var err error
			providers = GetBackupProviders()
			user1Ctx, err = backup.GetNonAdminCtx(userList[0], CommonPassword)
			log.FailOnError(err, "Fetching px-central-infra-admin user1 ctx")
			user2Ctx, err = backup.GetNonAdminCtx(userList[1], CommonPassword)
			log.FailOnError(err, "Fetching px-central-infra-admin ctx2")
			adminCtx, err = backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

		})

		//4. Create cloud credentials.
		Step("Creating cloud credentials", func() {
			log.InfoD("Creating cloud credentials")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
				cloudCredUID = uuid.New()
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, user1Ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
			}
		})

		//5. Create backup location
		Step("Creating backup location", func() {
			log.InfoD("Creating backup location")
			for _, provider := range providers {
				backupLocation = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocation
				err := CreateBackupLocationWithContext(provider, backupLocation, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", user1Ctx, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocation))
			}
		})

		//6. Create cluster for backup
		Step("Register cluster for backup", func() {
			err := CreateApplicationClusters(BackupOrgID, "", "", user1Ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, user1Ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(user1Ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})

		//7. Take backup
		Step("Taking backup of applications", func() {
			log.InfoD("Taking Backup of application")
			backupName = fmt.Sprintf("%s-%v", BackupNamePrefix, time.Now().Unix())
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			_, err := CreateBackupWithoutCheck(user1Ctx, backupName, SourceClusterName, backupLocation, backupLocationUID, appContextsToBackup, nil, BackupOrgID, clusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupName))
		})

		//8. Share backup to user2 with full access
		Step("Share backup to user2 with full access", func() {
			log.InfoD("Share backup with user2 having full access")
			err := ShareBackup(backupName, nil, []string{userList[1]}, FullAccess, user1Ctx)
			log.FailOnError(err, "Failed to share backup %s", backupName)
		})

		//9. Enumerate backup from user2
		Step("Enumerate backup from user2", func() {
			bkpEnumerateReq := &api.BackupEnumerateRequest{OrgId: BackupOrgID}
			backupResponse, err := Inst().Backup.EnumerateBackup(user2Ctx, bkpEnumerateReq)
			dash.VerifyFatal(err, nil, "Enumerate backup from user2 ctx")

			expected := strconv.Itoa(FullAccess)
			actual := ""
			log.Infof("len of Backup object is:%d", len(backupResponse.GetBackups()))
			log.Infof("backushared data: %v", backupResponse.GetBackups()[0].BackupShare.Collaborators)

			accessType := backupResponse.GetBackups()[0].BackupShare.Collaborators[0].Access
			actual = strconv.Itoa(int(accessType))
			dash.VerifyFatal(actual, expected, "Enumerate specific backup from user2 ctx")
		})

		//10. Delete backup from admin user
		Step("Verify delete backup from admin ctx", func() {
			backupDriver := Inst().Backup
			backupUID, err := backupDriver.GetBackupUID(user1Ctx, backupName, BackupOrgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - [%s]", backupName)

			log.InfoD("Deleting backup")
			err = DeleteBackupAndWaitForCompletion(backupName, backupUID, BackupOrgID, adminCtx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting backup [%s]", backupName))

		})

	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		// Cleaning up applications created
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)

		// Clean up the all users
		var wg sync.WaitGroup
		log.Infof("Cleaning up users")
		for _, userName := range userList {
			wg.Add(1)
			go func(userName string) {
				defer GinkgoRecover()
				defer wg.Done()
				err := backup.DeleteUser(userName)
				log.FailOnError(err, "Error deleting user %v", userName)
			}(userName)
		}
		wg.Wait()

		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, user1Ctx)
	})
})

// This test case verifies if Backup can be shared at Cluster backup share with existing backups with view, restorable and full access from multiple backuplocations which the user owns.
var _ = Describe("{ClusterBackupShareWithExistingBackupsWithViewRestorableAndFullAccessFromMultipleBackupLocationsUserOwned}", Label(TestCaseLabelsMap[ValidateUserAccessLevel]...), func() {
	var (
		backupName           string
		scheduledAppContexts []*scheduler.Context
		sourceClusterUid     string
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		cloudCredName        string
		cloudCredUID         string
		backupLocationUID    string
		backupLocation       string
		group1               string
		bkpNamespaces        []string
		allInfraAdminUsers   []string
		allNormalUsers       []string
		testAdminUserName    string
		secondUserName       string
		thirdUserName        string
		firstUserName        string
		bl1                  string
		bl2                  string
		blUID1               string
		blUID2               string
		restoreNames         []string
	)
	bkpNamespaces = make([]string, 0)
	labelSelectors := make(map[string]string)
	backupLocationMap := make(map[string]string)
	backupNames := make([]string, 0)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyClusterBackupShareWithExistingBackupsWithViewRestorableAndFullAccessFromMultipleBackupLocationsUserOwned",
			"Share Backup at cluster level with view, restorable and Full access from Multiple Backup Locations to different users and groups owned by User", nil, 80156, ABadgujar, Q2FY24)
		//1.Step-Deploy Applications in the Cluster
		log.InfoD("Deploy applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = AppReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
				scheduledAppContexts = append(scheduledAppContexts, ctx)
			}
		}
	})
	It("Share Backup at cluster level with view, restorable and Full access from Multiple Backup Locations owned by User to different users and groups", func() {
		//2.Step-Validate Deployed Applications
		Step("Validate applications", func() {
			log.InfoD("Validating apps")
			ValidateApplications(scheduledAppContexts)
		})

		//3.Step-Create 4 Users out of which 1 user has Infra Admin Role
		Step("Create 1 User with Infra Admin Role", func() {
			log.InfoD("Create 1 User with Infra Admin Role")
			role := backup.InfrastructureOwner
			allInfraAdminUsers = CreateUsers(1)
			//Assign Infra Admin Role to testadmin username
			testAdminUserName = allInfraAdminUsers[0]
			err := backup.AddRoleToUser(testAdminUserName, role, fmt.Sprintf("Adding %v role to %s", role, testAdminUserName))
			log.FailOnError(err, "Failed to add role for user - %s", testAdminUserName)
			allNormalUsers = CreateUsers(3)
			firstUserName = allNormalUsers[0]
			secondUserName = allNormalUsers[1]
			thirdUserName = allNormalUsers[2]
		})

		//4.Step-Create a group with full access
		Step("Create Group", func() {
			log.InfoD("Creating 1 user group for full access")
			group1 = "group-with-full-access"
			err := backup.AddGroup(group1)
			log.FailOnError(err, "Failed to create user group-1- %v", group1)
		})

		//5.Step-Add user thirdUserName to the group1
		Step("Add third User to group1", func() {
			log.InfoD("Adding third User to group1")
			err := backup.AddGroupToUser(thirdUserName, group1)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Adding user %s to group %s", thirdUserName, group1))
			usersOfGroup, err := backup.GetMembersOfGroup(group1)
			log.FailOnError(err, "Error fetching members of the group - %v", group1)
			log.Infof("Group [%v] contains the following users: \n%v", group1, usersOfGroup)
		})

		//6.Step-Creating Cloud Credentials with testadmin context and 2 Backup Locations
		Step("Creating backup location and cloud setting in TestAdmin User Context", func() {
			log.InfoD("Creating backup location and cloud setting in TestAdmin User Context")
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching TestAdmin User context")
			providers := GetBackupProviders()
			//For Backup Location 1
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cloudcred", provider, time.Now().Unix())
				backupLocation = fmt.Sprintf("%v-bl-ta-1", time.Now().Unix())
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocation
				blUID1 = backupLocationUID
				bl1 = backupLocation
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocationWithContext(provider, backupLocation, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", nonAdminCtx, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocation))

			}
			//For Backup Location 2
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cloudcred", provider, time.Now().Unix())
				backupLocation = fmt.Sprintf("%v-bl-ta-2", time.Now().Unix())
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocation
				blUID2 = backupLocationUID
				bl2 = backupLocation
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocationWithContext(provider, backupLocation, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", nonAdminCtx, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocation))
			}
		})

		//7.Share Ownership of the two Backup locations with secondusername and thirdusername through group1
		Step("Share Ownership of the Two Backup locations with secondusername and thirdusername through group1", func() {
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching TestAdmin User context")
			//Add Ownership to 2nd user and third user through group1
			for backupLocationUIDIteration, backupLocationIteration := range backupLocationMap {
				err = AddBackupLocationOwnership(backupLocationIteration, backupLocationUIDIteration, []string{secondUserName}, []string{group1}, Read, Invalid, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying updation of ownership for backuplocation - %s", backupLocationIteration))
			}
		})
		//8.Step-Registering Cluster for Backup testAdmin Context
		Step("Register cluster for backup", func() {
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching user context")
			err = CreateApplicationClusters(BackupOrgID, "", "", nonAdminCtx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, nonAdminCtx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			sourceClusterUid, err = Inst().Backup.GetClusterUID(nonAdminCtx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})

		//9.Step-Create Backups in all backup locations in testadminUser Context
		Step("Taking backup of applications", func() {
			//Common Steps for all backups
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", testAdminUserName)
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)

			//Backup1 Taken with BackupLocation1 of TestAdmin Context
			backupName = fmt.Sprintf("%s-%s-%v", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
			err = CreateBackupWithValidation(nonAdminCtx, backupName, SourceClusterName, bl1, blUID1, appContextsToBackup, labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupName))
			backupNames = append(backupNames, backupName)

			//Backup2 Taken with BackupLocation2 of TestAdmin Context
			backupName = fmt.Sprintf("%s-%s-%v", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
			err = CreateBackupWithValidation(nonAdminCtx, backupName, SourceClusterName, bl2, blUID2, appContextsToBackup, labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupName))
			backupNames = append(backupNames, backupName)
		})

		//10.Share All Backups with User1 with View Only access
		Step("Share the backups created by testAdmin User with ViewOnlyAccess with User1", func() {
			log.Infof("Share the backups created by testAdmin User with ViewOnlyAccess with User1")
			//For Sharing Backup Context needed is of Sharing Party i.e TestAdmin User
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", testAdminUserName)

			//Cluster Backup Share of all Backups with ViewOnly Access
			err = ClusterUpdateBackupShare(SourceClusterName, nil, []string{firstUserName}, ViewOnlyAccess, true, nonAdminCtx)
			log.FailOnError(err, "Failed sharing all backups for cluster [%s]", SourceClusterName)

			//Validate the access of Shared Backup
			log.Infof("Validating the shared backups with ViewOnlyAccess from one of the user")
			userCtx, err := backup.GetNonAdminCtx(firstUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", firstUserName)
			bkVar := 0
			// Start Restore and confirm that user cannot restore
			for _, backupName = range backupNames {
				bkVar++
				ctx, err := backup.GetAdminCtxFromSecret()
				Inst().Dash.VerifyFatal(err, nil, "Fetching px-central-admin ctx")
				log.InfoD("Registering Source and Destination clusters from user context")
				err = CreateApplicationClusters(BackupOrgID, "", "", userCtx)
				Inst().Dash.VerifyFatal(err, nil, "Creating source and destination cluster")
				destClusterUid, err := Inst().Backup.GetClusterUID(userCtx, BackupOrgID, DestinationClusterName)
				Inst().Dash.VerifyFatal(err, nil, "Getting destination cluster UID")
				backupDriver := Inst().Backup
				restoreName := fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
				restoreNames = append(restoreNames, restoreName)
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
				Inst().Dash.VerifyFatal(err, nil, "Getting destination cluster UID")
				log.InfoD("Validating if user [%s] with access [%v] can restore and delete backup %s or not", thirdUserName, BackupAccessKeyValue[1], backupName)
				err = CreateRestoreWithValidation(userCtx, restoreName, backupName, make(map[string]string), make(map[string]string), DestinationClusterName, destClusterUid, BackupOrgID, appContextsToBackup)
				if err != nil {
					log.Infof("Error returned - %s", err.Error())
				}
				// Restore validation to make sure that the user with View Access cannot restore
				dash.VerifyFatal(strings.Contains(err.Error(), "doesn't have permission to restore backup") ||
					strings.Contains(err.Error(), "failed to retrieve backup location"), true, fmt.Sprintf("Verifying backup restore is not possible for viewonlyUser-iteration-%v", bkVar))
				backupUID, err := backupDriver.GetBackupUID(ctx, backupName, BackupOrgID)
				Inst().Dash.VerifyFatal(err, nil, fmt.Sprintf("Getting backup UID for- %s", backupName))
				// Delete backup to confirm that the user has ViewOnlyAccess and cannot delete backup
				_, err = DeleteBackup(backupName, backupUID, BackupOrgID, userCtx)
				log.Infof("The expected error returned is %v", err)
				Inst().Dash.VerifyFatal(strings.Contains(err.Error(), "doesn't have permission to delete backup"), true, "Verifying backup deletion is not possible")
			}
			log.InfoD("Finished verifying access level - ViewOnlyAccess")

		})

		//11.Share All Backups with User2 with Restore Only access
		Step("Share the backups created by testAdmin User with Restore Only Access with User2", func() {
			log.Infof("Share the backups created by testAdmin User with Restore Only Access with User2")
			userCtx, err := backup.GetNonAdminCtx(secondUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", secondUserName)
			//For Sharing Backup Context needed is of Sharing Party i.e TestAdmin User
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", testAdminUserName)

			//Cluster Backup Share of all Backups with Restore Access
			err = ClusterUpdateBackupShare(SourceClusterName, nil, []string{secondUserName}, RestoreAccess, true, nonAdminCtx)
			log.FailOnError(err, "Failed sharing all backups for cluster [%s]", SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying share of all backups of cluster [%s] with [%#v] access level to user [%s]", SourceClusterName, RestoreAccess, secondUserName))
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			bkVar := 0
			//Validate the access of Shared Backups
			log.Infof("Validating the shared backups with RestoreAccess from one of the user")
			for _, backupName = range backupNames {
				bkVar++
				ctx, err := backup.GetAdminCtxFromSecret()
				Inst().Dash.VerifyFatal(err, nil, "Fetching px-central-admin ctx")
				log.InfoD("Registering Source and Destination clusters from user context")
				err = CreateApplicationClusters(BackupOrgID, "", "", userCtx)
				Inst().Dash.VerifyFatal(err, nil, "Creating source and destination cluster")
				destClusterUid, err := Inst().Backup.GetClusterUID(userCtx, BackupOrgID, DestinationClusterName)
				Inst().Dash.VerifyFatal(err, nil, "Getting destination cluster UID")
				backupDriver := Inst().Backup
				restoreName := fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
				restoreNames = append(restoreNames, restoreName)
				err = CreateRestoreWithValidation(userCtx, restoreName, backupName, make(map[string]string), make(map[string]string), DestinationClusterName, destClusterUid, BackupOrgID, appContextsToBackup)
				Inst().Dash.VerifyFatal(err, nil, "Verifying that restore is possible")
				// Try to delete the backup with user having RestoreAccess, and it should not pass
				backupUID, err := backupDriver.GetBackupUID(ctx, backupName, BackupOrgID)
				Inst().Dash.VerifyFatal(err, nil, fmt.Sprintf("Getting backup UID for- %s", backupName))
				// Delete backup to confirm that the user has Restore Access and delete backup should fail
				_, err = DeleteBackup(backupName, backupUID, BackupOrgID, userCtx)
				log.Infof("The expected error returned is %v", err)
				Inst().Dash.VerifyFatal(strings.Contains(err.Error(), "doesn't have permission to delete backup") ||
					strings.Contains(err.Error(), "failed to retrieve backup location"), true, fmt.Sprintf("Verifying backup deletion is not possible for restoreOnlyUser-iteration-%v", bkVar))
			}
			log.InfoD("Finished verifying access level - RestoreAccess")
		})

		//12.Share All Backups with Group1 with Full Access
		Step("Share the backups created by testAdmin User with Full Access with Group1", func() {
			log.Infof("Share the backups created by testAdmin User with Full Access with Group1")
			//For Sharing Backup Context needed is of Sharing Party i.e TestAdmin User
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", testAdminUserName)
			//Cluster Backup Share of all Backups with full Access
			err = ClusterUpdateBackupShare(SourceClusterName, []string{group1}, nil, FullAccess, true, nonAdminCtx)
			log.FailOnError(err, "Failed sharing all backups for cluster [%s]", SourceClusterName)
			//Validate the access of Shared Backups
			log.Infof("Validating the shared backups with FullAccess from one of the user")
			for _, backupName = range backupNames {
				restoreName := fmt.Sprintf("%s-%s-%v", thirdUserName, RestoreNamePrefix, RandomString(5))
				restoreNames = append(restoreNames, restoreName)
				ValidateSharedBackupWithUsers(thirdUserName, FullAccess, backupName, restoreName)
			}
			log.InfoD("Finished verifying access level - FullAccess")
		})
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		// Cleaning up applications created
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)

		// Clean the all user created
		var wg sync.WaitGroup
		log.Infof("Cleaning up users")
		for _, userName := range allInfraAdminUsers {
			wg.Add(1)
			go func(userName string) {
				defer GinkgoRecover()
				defer wg.Done()
				err := backup.DeleteUser(userName)
				log.FailOnError(err, "Error deleting user %v", userName)
			}(userName)
		}
		for _, userName := range allNormalUsers {
			wg.Add(1)
			go func(userName string) {
				defer GinkgoRecover()
				defer wg.Done()
				err := backup.DeleteUser(userName)
				log.FailOnError(err, "Error deleting user %v", userName)
			}(userName)
		}
		wg.Wait()

		// Clean up the all group created
		log.Infof("Cleaning up groups")
		err := backup.DeleteGroup(group1)
		log.FailOnError(err, "Error deleting user %v", group1)

		// Clean up the cluster
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		for _, restoreNameIteration := range restoreNames {
			err = DeleteRestore(restoreNameIteration, BackupOrgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting restore [%s]", restoreNameIteration))
		}
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)

	})
})

// This test case verifies if Backup can be shared at Cluster backup share with existing backups with full access from multiple backuplocations which the user owns and restored in different namespace
var _ = Describe("{ClusterBackupShareWithExistingBackupsWithFullAccessFromMultipleBackupLocationsWhichTheUserOwnsDifferentNamespace.}", Label(TestCaseLabelsMap[ValidateUserAccessLevel]...), func() {
	var (
		backupName           string
		scheduledAppContexts []*scheduler.Context
		sourceClusterUid     string
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		cloudCredName        string
		cloudCredUID         string
		backupLocationUID    string
		backupLocation       string
		bkpNamespaces        []string
		allInfraAdminUsers   []string
		allNormalUsers       []string
		testAdminUserName    string
		firstUserName        string
		secondUserName       string
		bl1                  string
		bl2                  string
		blUID1               string
		blUID2               string
		group1               string
	)
	bkpNamespaces = make([]string, 0)
	labelSelectors := make(map[string]string)
	backupLocationMap := make(map[string]string)
	backupNames := make([]string, 0)
	namespaceMapping := make(map[string]string)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyClusterBackupShareWithExistingBackupsWithFullAccessFromMultipleBackupLocationsWhichTheUserOwnsDifferentNamespace",
			"Share Backup at cluster level with existing backups with full access from multiple backuplocations which the user owns and restore from different namespace", nil, 80155, ABadgujar, Q2FY24)
		//1.Step-Deploy Applications in the Cluster
		log.InfoD("Deploy applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = AppReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
				scheduledAppContexts = append(scheduledAppContexts, ctx)
			}
		}
	})
	It("Share Backup at cluster level with existing backups with full access from multiple backuplocations which the user owns and restore from different namespace", func() {
		//2.Step-Validate Deployed Applications
		Step("Validate applications", func() {
			log.InfoD("Validating apps")
			ValidateApplications(scheduledAppContexts)
		})

		//3.Step-Create 3 Users out of which 1 user has Infra Admin Role
		Step("Create 1 User with Infra Admin Role", func() {
			log.InfoD("Create 1 User with Infra Admin Role")
			role := backup.InfrastructureOwner
			allInfraAdminUsers = CreateUsers(1)
			//Assign Infra Admin Role to testadmin username
			testAdminUserName = allInfraAdminUsers[0]
			err := backup.AddRoleToUser(testAdminUserName, role, fmt.Sprintf("Adding %v role to %s", role, testAdminUserName))
			log.FailOnError(err, "Failed to add role for user - %s", testAdminUserName)
			allNormalUsers = CreateUsers(2)
			firstUserName = allNormalUsers[0]
			secondUserName = allNormalUsers[1]
		})

		//4.Step-Create a group with full access
		Step("Create Group", func() {
			log.InfoD("Creating 1 user group for full access")
			group1 = "group-with-full-access"
			err := backup.AddGroup(group1)
			log.FailOnError(err, "Failed to create user group-1- %v", group1)
		})

		//5.Step-Add user secondUserName to the group1
		Step("Add second User to group1", func() {
			log.InfoD("Adding second User to group1")
			err := backup.AddGroupToUser(secondUserName, group1)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Adding user %s to group %s", secondUserName, group1))
			usersOfGroup, err := backup.GetMembersOfGroup(group1)
			log.FailOnError(err, "Error fetching members of the group - %v", group1)
			log.Infof("Group [%v] contains the following users: \n%v", group1, usersOfGroup)
		})

		//6.Step-Creating Cloud setting in TestAdmin User Context
		Step("Creating cloud setting in TestAdmin User Context", func() {
			log.InfoD("Creating cloud setting as admin and share it to testadmin user")
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching TestAdmin User context")
			providers := GetBackupProviders()
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v-a-cc", "cloudcred", provider, time.Now().Unix())
				cloudCredUID = uuid.New()
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				if provider != drivers.ProviderNfs {
					log.Infof("Shared %v Cloud Credentials to User - %v", cloudCredName, testAdminUserName)
					//share cloud credential
					err = AddCloudCredentialOwnership(cloudCredName, cloudCredUID, []string{firstUserName}, []string{group1}, Read, Invalid, nonAdminCtx, BackupOrgID)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying updation of ownership for CloudCredential- %s", cloudCredName))
				}
			}
		})

		//7.Step-Creating 2 backup location in TestAdmin User Context
		Step("Creating 2 backup location and in TestAdmin User Context", func() {
			log.InfoD("Creating backup location in TestAdmin User Context")
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching TestAdmin User context")
			providers := GetBackupProviders()

			//For Backup Location 1
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v-ta-cc", "cloudcred", provider, time.Now().Unix())
				backupLocation = fmt.Sprintf("%v-bl-ta-2", time.Now().Unix())
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocation
				blUID2 = backupLocationUID
				bl2 = backupLocation
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocationWithContext(provider, backupLocation, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", nonAdminCtx, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocation))
			}

			//For Backup Location 2
			for _, provider := range providers {
				backupLocation = fmt.Sprintf("%v-bl-ta-1", time.Now().Unix())
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocation
				blUID1 = backupLocationUID
				bl1 = backupLocation
				err = CreateBackupLocationWithContext(provider, backupLocation, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", nonAdminCtx, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocation))

			}

		})

		//8.Share Ownership of the two Backup locations and cloud credentials with AllNormalUsers
		Step("Share Ownership of the Two Backup locations with AllNormalUsers", func() {
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching TestAdmin User context")
			//Add Ownership to user1 and group1
			for backupLocationUIDIteration, backupLocationIteration := range backupLocationMap {
				err = AddBackupLocationOwnership(backupLocationIteration, backupLocationUIDIteration, []string{firstUserName}, []string{group1}, Read, Invalid, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying updation of ownership for backuplocation - %s", backupLocationIteration))
			}
		})

		//9.Step-Registering Cluster for Backup testAdmin Context
		Step("Register cluster for backup", func() {
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching user context")
			err = CreateApplicationClusters(BackupOrgID, "", "", nonAdminCtx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, nonAdminCtx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			sourceClusterUid, err = Inst().Backup.GetClusterUID(nonAdminCtx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})

		//10.Step-Create Backups in all backup locations in testadminUser Context
		Step("Taking backup of applications", func() {
			//Common Steps for all backups
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", testAdminUserName)
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)

			//Backup1 Taken with BackupLocation1 of TestAdmin Context
			backupName = fmt.Sprintf("%s-%s-%v-backup1", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
			err = CreateBackupWithValidation(nonAdminCtx, backupName, SourceClusterName, bl1, blUID1, appContextsToBackup, labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupName))
			backupNames = append(backupNames, backupName)

			//Backup2 Taken with BackupLocation2 of TestAdmin Context
			backupName = fmt.Sprintf("%s-%s-%v-backup2", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
			err = CreateBackupWithValidation(nonAdminCtx, backupName, SourceClusterName, bl2, blUID2, appContextsToBackup, labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupName))
			backupNames = append(backupNames, backupName)
		})

		//11.Share All Backups with User1 and group1 with Full Access
		Step("Share All Backups with User1 and group1 with Full Access", func() {
			log.Infof("Share All Backups with User1 and group1 with Full Access")
			//For Sharing Backup Context needed is of Sharing Party i.e TestAdmin User
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", testAdminUserName)
			//Cluster Backup Share of all Backups with full Access
			err = ClusterUpdateBackupShare(SourceClusterName, []string{group1}, []string{firstUserName}, FullAccess, true, nonAdminCtx)
			log.FailOnError(err, "Failed sharing all backups for cluster [%s]", SourceClusterName)
			//Validate the access of Shared Backups
		})

		//12.Restore to a new namespace and see if access is there in firstUserName context
		Step("Restore to a new namespace and see if access is there", func() {
			log.Infof("Restore to a new namespace and see if access is there")

			userCtx, err := backup.GetNonAdminCtx(firstUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", testAdminUserName)

			Inst().Dash.VerifyFatal(err, nil, "Fetching px-central-admin ctx")
			log.InfoD("Registering Source and Destination clusters from user context")
			err = CreateApplicationClusters(BackupOrgID, "", "", userCtx)
			Inst().Dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			destClusterUid, err := Inst().Backup.GetClusterUID(userCtx, BackupOrgID, DestinationClusterName)
			Inst().Dash.VerifyFatal(err, nil, "Getting destination cluster UID")
			restoreName := fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
			// Create Restore in a different namespace
			for _, namespace := range bkpNamespaces {
				restoredNameSpace := fmt.Sprintf("%s-%s", RandomString(10), "restored")
				namespaceMapping[namespace] = restoredNameSpace
			}

			log.InfoD("Namespace mapping is %v:", namespaceMapping)
			err = CreateRestore(restoreName, backupNames[0], namespaceMapping, DestinationClusterName, destClusterUid, BackupOrgID, userCtx, make(map[string]string))
			Inst().Dash.VerifyFatal(err, nil, "Verifying that restore is possible")
		})

		//13.Delete from Group User context to confirm full access
		Step("Delete from Group User context to confirm full access", func() {
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Error Fetching User context")
			userCtx, err := backup.GetNonAdminCtx(secondUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", testAdminUserName)
			backupDriver := Inst().Backup

			backupUID, err := backupDriver.GetBackupUID(ctx, backupNames[1], BackupOrgID)
			Inst().Dash.VerifyFatal(err, nil, fmt.Sprintf("Getting backup UID for- %s", backupNames[1]))
			// Delete backup to confirm that the user has full access
			_, err = DeleteBackup(backupNames[1], backupUID, BackupOrgID, userCtx)
			Inst().Dash.VerifyFatal(err, nil, "Verifying that delete backup is possible")
		})
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		// Cleaning up applications created
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)

		// Clean the all user created
		var wg sync.WaitGroup
		log.Infof("Cleaning up users")
		for _, userName := range allInfraAdminUsers {
			wg.Add(1)
			go func(userName string) {
				defer GinkgoRecover()
				defer wg.Done()
				err := backup.DeleteUser(userName)
				log.FailOnError(err, "Error deleting user %v", userName)
			}(userName)
		}
		for _, userName := range allNormalUsers {
			wg.Add(1)
			go func(userName string) {
				defer GinkgoRecover()
				defer wg.Done()
				err := backup.DeleteUser(userName)
				log.FailOnError(err, "Error deleting user %v", userName)
			}(userName)
		}
		wg.Wait()

		// Clean up the all group created
		log.Infof("Cleaning up groups")
		err := backup.DeleteGroup(group1)
		log.FailOnError(err, "Error deleting user %v", group1)

		// Clean up the cluster
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)

	})
})

// This test case verifies if Backup can be shared at Cluster backup share with existing backups with view, restorable and full access from multiple backuplocations from which some cloud credentials are owned by the user.
var _ = Describe("{ClusterBackupShareWithExistingBackupsWithViewRestorableAndFullAccessFromMultipleBackupLocationsCloudCredentialsUserOwned}", Label(TestCaseLabelsMap[ValidateUserAccessLevel]...), func() {
	var (
		backupName           string
		scheduledAppContexts []*scheduler.Context
		sourceClusterUid     string
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		cloudCredName        string
		cloudCredUID         string
		backupLocationUID    string
		backupLocation       string
		bkpNamespaces        []string
		allInfraAdminUsers   []string
		allNormalUsers       []string
		testAdminUserName    string
		firstUserName        string
		bl1                  string
		bl2                  string
		blUID1               string
		blUID2               string
		restoreNames         []string
	)
	bkpNamespaces = make([]string, 0)
	labelSelectors := make(map[string]string)
	backupLocationMap := make(map[string]string)
	backupNames := make([]string, 0)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyClusterBackupShareWithExistingBackupsWithViewRestorableAndFullAccessFromMultipleBackupLocationsCloudCredentialsUserOwned",
			"Share Backup at cluster level with view, restorable and Full access from Multiple Backup Locations from which some cloud credentials are owned by the user.", nil, 80158, ABadgujar, Q2FY24)
		//1.Step-Deploy Applications in the Cluster
		log.InfoD("Deploy applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = AppReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
				scheduledAppContexts = append(scheduledAppContexts, ctx)
			}
		}
	})
	It("Share Backup at cluster level with view, restorable and Full access from Multiple Backup Locations from which some cloud credentials are owned by the user.", func() {
		//2.Step-Validate Deployed Applications
		Step("Validate applications", func() {
			log.InfoD("Validating apps")
			ValidateApplications(scheduledAppContexts)
		})

		//3.Step-Create 2 Users out of which 1 user has Infra Admin Role
		Step("Create 1 User with Infra Admin Role", func() {
			log.InfoD("Create 1 User with Infra Admin Role")
			role := backup.InfrastructureOwner
			allInfraAdminUsers = CreateUsers(1)
			//Assign Infra Admin Role to testadmin username
			testAdminUserName = allInfraAdminUsers[0]
			err := backup.AddRoleToUser(testAdminUserName, role, fmt.Sprintf("Adding %v role to %s", role, testAdminUserName))
			log.FailOnError(err, "Failed to add role for user - %s", testAdminUserName)
			allNormalUsers = CreateUsers(1)
			firstUserName = allNormalUsers[0]
		})

		//4.Step-Creating Cloud Credentials with admin context and share it to testadmin user
		Step("Creating backup location and cloud setting in TestAdmin User Context", func() {
			log.InfoD("Creating backup location and cloud setting as admin and share it to testadmin user")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			providers := GetBackupProviders()
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v-a-cc", "cloudcred", provider, time.Now().Unix())
				cloudCredUID = uuid.New()
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				if provider != drivers.ProviderNfs {
					log.Infof("Shared %v Cloud Credentials to User - %v", cloudCredName, testAdminUserName)
					err = AddCloudCredentialOwnership(cloudCredName, cloudCredUID, []string{testAdminUserName}, nil, Read, Invalid, ctx, BackupOrgID)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying updation of ownership for CloudCredential- %s", cloudCredName))
				}
			}

			//5.Step-Creating 2 backup location and cloud setting in TestAdmin User Context
			Step("Creating 2 backup location and cloud setting in TestAdmin User Context", func() {
				log.InfoD("Creating backup location and cloud setting in TestAdmin User Context")
				nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
				log.FailOnError(err, "Fetching TestAdmin User context")
				providers := GetBackupProviders()
				//For Backup Location 1 to make Backuplocation using cloud credential shared by Admin
				for _, provider := range providers {
					backupLocation = fmt.Sprintf("%v-bl-ta-1", time.Now().Unix())
					backupLocationUID = uuid.New()
					backupLocationMap[backupLocationUID] = backupLocation
					blUID1 = backupLocationUID
					bl1 = backupLocation
					err = CreateBackupLocationWithContext(provider, backupLocation, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", nonAdminCtx, true)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocation))

				}
				//For Backup Location 2 make own cloud credential and make Backup location
				for _, provider := range providers {
					cloudCredName = fmt.Sprintf("%s-%s-%v-ta-cc", "cloudcred", provider, time.Now().Unix())
					backupLocation = fmt.Sprintf("%v-bl-ta-2", time.Now().Unix())
					cloudCredUID = uuid.New()
					backupLocationUID = uuid.New()
					backupLocationMap[backupLocationUID] = backupLocation
					blUID2 = backupLocationUID
					bl2 = backupLocation
					err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, nonAdminCtx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
					err = CreateBackupLocationWithContext(provider, backupLocation, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", nonAdminCtx, true)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocation))
				}
			})

			//6.Step-Registering Cluster for Backup testAdmin Context
			Step("Register cluster for backup", func() {
				nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
				log.FailOnError(err, "Fetching user context")
				err = CreateApplicationClusters(BackupOrgID, "", "", nonAdminCtx)
				dash.VerifyFatal(err, nil, "Creating source and destination cluster")
				clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, nonAdminCtx)
				log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
				dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
				sourceClusterUid, err = Inst().Backup.GetClusterUID(nonAdminCtx, BackupOrgID, SourceClusterName)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
			})

			//7.Step-Create Backups in all backup locations in testadminUser Context
			Step("Taking backup of applications", func() {
				//Common Steps for all backups
				nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
				log.FailOnError(err, "Fetching user [%s] ctx", testAdminUserName)
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)

				//Backup1 Taken with BackupLocation1 of TestAdmin Context
				backupName = fmt.Sprintf("%s-%s-%v-notowncc", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
				err = CreateBackupWithValidation(nonAdminCtx, backupName, SourceClusterName, bl1, blUID1, appContextsToBackup, labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupName))
				backupNames = append(backupNames, backupName)

				//Backup2 Taken with BackupLocation2 of TestAdmin Context
				backupName = fmt.Sprintf("%s-%s-%v", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
				err = CreateBackupWithValidation(nonAdminCtx, backupName, SourceClusterName, bl2, blUID2, appContextsToBackup, labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupName))
				backupNames = append(backupNames, backupName)
			})

			//8.Share All Backups with User1 with Full Access and validate access
			Step("Share All Backups with User1 with Full Access and validate access", func() {
				log.Infof("Share All Backups with User1 with Full Access and validate access should be viewonly for 1 backup and full for 2nd")
				//For Sharing Backup Context needed is of Sharing Party i.e TestAdmin User
				nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
				log.FailOnError(err, "Fetching user [%s] ctx", testAdminUserName)
				//Cluster Backup Share of all Backups with full Access
				err = ClusterUpdateBackupShare(SourceClusterName, nil, []string{firstUserName}, FullAccess, true, nonAdminCtx)
				log.FailOnError(err, "Failed sharing all backups for cluster [%s]", SourceClusterName)
				//Validate the access of Shared Backups
				log.Infof("Validating the shared backups with FullAccess from one of the user")
				for _, backupName = range backupNames {
					if strings.Contains(backupName, "notowncc") == true {
						restoreName := fmt.Sprintf("%s-%s-%v-notowncc", firstUserName, RestoreNamePrefix, RandomString(5))
						restoreNames = append(restoreNames, restoreName)
						ValidateSharedBackupWithUsers(firstUserName, ViewOnlyAccess, backupName, restoreName)
					} else {
						restoreName := fmt.Sprintf("%s-%s-%v", firstUserName, RestoreNamePrefix, RandomString(5))
						restoreNames = append(restoreNames, restoreName)
						ValidateSharedBackupWithUsers(firstUserName, FullAccess, backupName, restoreName)
					}

				}
				log.InfoD("Finished verifying access level - ViewOnly and FullAccess")
			})
		})

	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		// Cleaning up applications created
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)

		// Clean the all user created
		var wg sync.WaitGroup
		log.Infof("Cleaning up users")
		for _, userName := range allInfraAdminUsers {
			wg.Add(1)
			go func(userName string) {
				defer GinkgoRecover()
				defer wg.Done()
				err := backup.DeleteUser(userName)
				log.FailOnError(err, "Error deleting user %v", userName)
			}(userName)
		}
		for _, userName := range allNormalUsers {
			wg.Add(1)
			go func(userName string) {
				defer GinkgoRecover()
				defer wg.Done()
				err := backup.DeleteUser(userName)
				log.FailOnError(err, "Error deleting user %v", userName)
			}(userName)
		}
		wg.Wait()

		// Clean up the cluster
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		for _, restoreNameIteration := range restoreNames {
			err = DeleteRestore(restoreNameIteration, BackupOrgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting restore [%s]", restoreNameIteration))
		}
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})

// This test case verifies if Backup can be shared at Cluster backup share with existing backups with Restore access from multiple backuplocations which the user owns and restored in different namespace
var _ = Describe("{ClusterBackupShareWithExistingBackupsWithRestoreAccessFromMultipleBackupLocationsWhichTheUserOwnsDifferentNamespace.}", Label(TestCaseLabelsMap[ValidateUserAccessLevel]...), func() {
	var (
		backupName           string
		scheduledAppContexts []*scheduler.Context
		sourceClusterUid     string
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		cloudCredName        string
		cloudCredUID         string
		backupLocationUID    string
		backupLocation       string
		bkpNamespaces        []string
		allInfraAdminUsers   []string
		allNormalUsers       []string
		testAdminUserName    string
		firstUserName        string
		secondUserName       string
		bl1                  string
		bl2                  string
		blUID1               string
		blUID2               string
		group1               string
		restoreNames         []string
	)
	bkpNamespaces = make([]string, 0)
	labelSelectors := make(map[string]string)
	backupLocationMap := make(map[string]string)
	backupNames := make([]string, 0)
	namespaceMapping := make(map[string]string)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyClusterBackupShareWithExistingBackupsWithRestoreAccessFromMultipleBackupLocationsWhichTheUserOwnsDifferentNamespace",
			"Share Backup at cluster level with existing backups with Restore access from multiple backuplocations which the user owns and restore from different namespace", nil, 80155, ABadgujar, Q2FY24)
		//1.Step-Deploy Applications in the Cluster
		log.InfoD("Deploy applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = AppReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
				scheduledAppContexts = append(scheduledAppContexts, ctx)
			}
		}
	})
	It("Share Backup at cluster level with existing backups with Restore access from multiple backuplocations which the user owns and restore from different namespace", func() {
		//2.Step-Validate Deployed Applications
		Step("Validate applications", func() {
			log.InfoD("Validating apps")
			ValidateApplications(scheduledAppContexts)
		})

		//3.Step-Create 3 Users out of which 1 user has Infra Admin Role
		Step("Create 1 User with Infra Admin Role", func() {
			log.InfoD("Create 1 User with Infra Admin Role")
			role := backup.InfrastructureOwner
			allInfraAdminUsers = CreateUsers(1)
			//Assign Infra Admin Role to testadmin username
			testAdminUserName = allInfraAdminUsers[0]
			err := backup.AddRoleToUser(testAdminUserName, role, fmt.Sprintf("Adding %v role to %s", role, testAdminUserName))
			log.FailOnError(err, "Failed to add role for user - %s", testAdminUserName)
			allNormalUsers = CreateUsers(2)
			firstUserName = allNormalUsers[0]
			secondUserName = allNormalUsers[1]
		})

		//4.Step-Create a group with full access
		Step("Create Group", func() {
			log.InfoD("Creating 1 user group for full access")
			group1 = "group-with-full-access"
			err := backup.AddGroup(group1)
			log.FailOnError(err, "Failed to create user group-1- %v", group1)
		})

		//5.Step-Add user secondUserName to the group1
		Step("Add second User to group1", func() {
			log.InfoD("Adding second User to group1")
			err := backup.AddGroupToUser(secondUserName, group1)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Adding user %s to group %s", secondUserName, group1))
			usersOfGroup, err := backup.GetMembersOfGroup(group1)
			log.FailOnError(err, "Error fetching members of the group - %v", group1)
			log.Infof("Group [%v] contains the following users: \n%v", group1, usersOfGroup)
		})

		//6.Step-Creating Cloud setting in TestAdmin User Context
		Step("Creating cloud setting in TestAdmin User Context", func() {
			log.InfoD("Creating cloud setting as admin and share it to testadmin user")
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching TestAdmin User context")
			providers := GetBackupProviders()
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v-a-cc", "cloudcred", provider, time.Now().Unix())
				cloudCredUID = uuid.New()
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				if provider != drivers.ProviderNfs {
					log.Infof("Shared %v Cloud Credentials to User - %v", cloudCredName, testAdminUserName)
					//share cloud credential
					err = AddCloudCredentialOwnership(cloudCredName, cloudCredUID, []string{firstUserName}, []string{group1}, Read, Invalid, nonAdminCtx, BackupOrgID)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying updation of ownership for CloudCredential- %s", cloudCredName))
				}
			}
		})

		//7.Step-Creating 2 backup location in TestAdmin User Context
		Step("Creating 2 backup location and in TestAdmin User Context", func() {
			log.InfoD("Creating backup location in TestAdmin User Context")
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching TestAdmin User context")
			providers := GetBackupProviders()

			//For Backup Location 1
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v-ta-cc", "cloudcred", provider, time.Now().Unix())
				backupLocation = fmt.Sprintf("%v-bl-ta-2", time.Now().Unix())
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocation
				blUID2 = backupLocationUID
				bl2 = backupLocation
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocationWithContext(provider, backupLocation, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", nonAdminCtx, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocation))
			}

			//For Backup Location 2
			for _, provider := range providers {
				backupLocation = fmt.Sprintf("%v-bl-ta-1", time.Now().Unix())
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocation
				blUID1 = backupLocationUID
				bl1 = backupLocation
				err = CreateBackupLocationWithContext(provider, backupLocation, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", nonAdminCtx, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocation))

			}

		})

		//8.Share Ownership of the two Backup locations and cloud credentials with AllNormalUsers
		Step("Share Ownership of the Two Backup locations with AllNormalUsers", func() {
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching TestAdmin User context")
			//Add Ownership to user1 and group1
			for backupLocationUIDIteration, backupLocationIteration := range backupLocationMap {
				err = AddBackupLocationOwnership(backupLocationIteration, backupLocationUIDIteration, []string{firstUserName}, []string{group1}, Read, Invalid, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying updation of ownership for backuplocation - %s", backupLocationIteration))
			}
		})

		//9.Step-Registering Cluster for Backup testAdmin Context
		Step("Register cluster for backup", func() {
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching user context")
			err = CreateApplicationClusters(BackupOrgID, "", "", nonAdminCtx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, nonAdminCtx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			sourceClusterUid, err = Inst().Backup.GetClusterUID(nonAdminCtx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})

		//10.Step-Create Backups in all backup locations in testadminUser Context
		Step("Taking backup of applications", func() {
			//Common Steps for all backups
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", testAdminUserName)
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)

			//Backup1 Taken with BackupLocation1 of TestAdmin Context
			backupName = fmt.Sprintf("%s-%s-%v-backup1", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
			err = CreateBackupWithValidation(nonAdminCtx, backupName, SourceClusterName, bl1, blUID1, appContextsToBackup, labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupName))
			backupNames = append(backupNames, backupName)

			//Backup2 Taken with BackupLocation2 of TestAdmin Context
			backupName = fmt.Sprintf("%s-%s-%v-backup2", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
			err = CreateBackupWithValidation(nonAdminCtx, backupName, SourceClusterName, bl2, blUID2, appContextsToBackup, labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupName))
			backupNames = append(backupNames, backupName)
		})

		//11.Share All Backups with User1 and group1 with Restore Access
		Step("Share All Backups with User1 and group1 with Restore Access", func() {
			log.Infof("Share All Backups with User1 and group1 with Restore Access")
			//For Sharing Backup Context needed is of Sharing Party i.e TestAdmin User
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", testAdminUserName)
			//Cluster Backup Share of all Backups with Restore Access
			err = ClusterUpdateBackupShare(SourceClusterName, []string{group1}, []string{firstUserName}, RestoreAccess, true, nonAdminCtx)
			log.FailOnError(err, "Failed sharing all backups for cluster [%s]", SourceClusterName)
		})

		//12.Restore to a new namespace and see if access is there in firstUserName context
		Step("Restore to a new namespace and see if access is there", func() {
			log.Infof("Restore to a new namespace and see if access is there")

			userCtx, err := backup.GetNonAdminCtx(firstUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", testAdminUserName)

			Inst().Dash.VerifyFatal(err, nil, "Fetching px-central-admin ctx")
			log.InfoD("Registering Source and Destination clusters from user context")
			err = CreateApplicationClusters(BackupOrgID, "", "", userCtx)
			Inst().Dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			destClusterUid, err := Inst().Backup.GetClusterUID(userCtx, BackupOrgID, DestinationClusterName)
			Inst().Dash.VerifyFatal(err, nil, "Getting destination cluster UID")
			restoreName := fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
			restoreNames = append(restoreNames, restoreName)
			// Create Restore in a different namespace
			for _, namespace := range bkpNamespaces {
				restoredNameSpace := fmt.Sprintf("%s-%s", RandomString(10), "restored")
				namespaceMapping[namespace] = restoredNameSpace
			}

			log.InfoD("Namespace mapping is %v:", namespaceMapping)
			//Validate restore access of shared backup
			err = CreateRestore(restoreName, backupNames[0], namespaceMapping, DestinationClusterName, destClusterUid, BackupOrgID, userCtx, make(map[string]string))
			Inst().Dash.VerifyFatal(err, nil, "Verifying that restore is possible")
		})

		//13.Delete from Group User context to confirm full access
		Step("Delete from Group User context to confirm full access", func() {
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Error Fetching User context")
			userCtx, err := backup.GetNonAdminCtx(secondUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", testAdminUserName)
			backupDriver := Inst().Backup

			backupUID, err := backupDriver.GetBackupUID(ctx, backupNames[1], BackupOrgID)
			Inst().Dash.VerifyFatal(err, nil, fmt.Sprintf("Getting backup UID for- %s", backupNames[1]))
			// Delete backup to confirm that the user has Restore Access and delete backup should fail
			_, err = DeleteBackup(backupNames[1], backupUID, BackupOrgID, userCtx)
			Inst().Dash.VerifyFatal(strings.Contains(err.Error(), "doesn't have permission to delete backup"), true, "Verifying backup deletion is not possible")
		})
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		// Cleaning up applications created
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)

		// Clean the all user created
		var wg sync.WaitGroup
		log.Infof("Cleaning up users")
		for _, userName := range allInfraAdminUsers {
			wg.Add(1)
			go func(userName string) {
				defer GinkgoRecover()
				defer wg.Done()
				err := backup.DeleteUser(userName)
				log.FailOnError(err, "Error deleting user %v", userName)
			}(userName)
		}
		for _, userName := range allNormalUsers {
			wg.Add(1)
			go func(userName string) {
				defer GinkgoRecover()
				defer wg.Done()
				err := backup.DeleteUser(userName)
				log.FailOnError(err, "Error deleting user %v", userName)
			}(userName)
		}
		wg.Wait()

		// Clean up the all group created
		log.Infof("Cleaning up groups")
		err := backup.DeleteGroup(group1)
		log.FailOnError(err, "Error deleting user %v", group1)

		// Clean up the cluster
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")

		for _, restoreNameIteration := range restoreNames {
			err = DeleteRestore(restoreNameIteration, BackupOrgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting restore [%s]", restoreNameIteration))
		}
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)

	})
})

// This test case verifies if Backup can be shared at cluster level with view, restorable and Full access from Multiple Backup Locations to different users and groups
var _ = Describe("{ClusterBackupShareWithExistingBackupsWithViewRestorableAndFullAccessFromMultipleBackupLocations}", Label(TestCaseLabelsMap[ValidateUserAccessLevel]...), func() {
	var (
		backupName           string
		scheduledAppContexts []*scheduler.Context
		sourceClusterUid     string
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		cloudCredName        string
		cloudCredUID         string
		backupLocationUID    string
		backupLocation       string
		group1               string
		bkpNamespaces        []string
		allInfraAdminUsers   []string
		allNormalUsers       []string
		restoreNames         []string
		testAdminUserName    string
		thirdUserName        string
		fourthUserName       string
		fifthUserName        string
		secondUserName       string
		bl1                  string
		bl2                  string
		bl3                  string
		blUID1               string
		blUID2               string
		blUID3               string
	)
	bkpNamespaces = make([]string, 0)
	labelSelectors := make(map[string]string)
	backupLocationMap := make(map[string]string)
	backupNames := make([]string, 0)
	secondarybackupNames := make([]string, 0)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyClusterBackupShareWithExistingBackupsWithViewRestorableAndFullAccessFromMultipleBackupLocations",
			"Share Backup at cluster level with view, restorable and Full access from Multiple Backup Locations to different users and groups", nil, 80157, ABadgujar, Q2FY24)
		//1.Step-Deploy Applications in the Cluster
		log.InfoD("Deploy applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = AppReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
				scheduledAppContexts = append(scheduledAppContexts, ctx)
			}
		}
	})
	It("Share Backup at cluster level with view, restorable and Full access from Multiple Backup Locations to different users and groups", func() {
		//2.Step-Validate Deployed Applications
		Step("Validate applications", func() {
			log.InfoD("Validating apps")
			ValidateApplications(scheduledAppContexts)
		})

		//3.Step-Create 4 Users out of which 2 users have Infra Admin Role
		Step("Create 2 Users with Infra Admin Role", func() {
			log.InfoD("Create 2 Users with Infra Admin Role")
			role := backup.InfrastructureOwner
			allInfraAdminUsers = CreateUsers(2)
			for _, userName := range allInfraAdminUsers {
				err := backup.AddRoleToUser(userName, role, fmt.Sprintf("Adding %v role to %s", role, userName))
				log.FailOnError(err, "Failed to add role for user - %s", userName)
				//Append both usernames to Slice of Usernames and store them for future use
				log.FailOnError(err, "Failed to fetch uid for - %s", userName)
			}
			testAdminUserName = allInfraAdminUsers[0]
			secondUserName = allInfraAdminUsers[1]
			allNormalUsers = CreateUsers(3)
			thirdUserName = allNormalUsers[0]
			fourthUserName = allNormalUsers[1]
			fifthUserName = allNormalUsers[2]
		})

		//4.Step-Create a group with full access
		Step("Create Group", func() {
			log.InfoD("Creating 1 user group for full access")
			group1 = "group-with-full-access"
			err := backup.AddGroup(group1)
			log.FailOnError(err, "Failed to create user group-1- %v", group1)
		})

		//5.Step-Add user fifthUserName to the group1
		Step("Add fifth User to group1", func() {
			log.InfoD("Adding fifth User to group1")
			err := backup.AddGroupToUser(fifthUserName, group1)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Adding user %s to group %s", fifthUserName, group1))
			usersOfGroup, err := backup.GetMembersOfGroup(group1)
			log.FailOnError(err, "Error fetching members of the group - %v", group1)
			log.Infof("Group [%v] contains the following users: \n%v", group1, usersOfGroup)
		})

		//6.Step-Creating Cloud Credentials with testadmin context and 2 Backup Locations
		Step("Creating backup location and cloud setting in TestAdmin User Context", func() {
			log.InfoD("Creating backup location in TestAdmin User Context")
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching TestAdmin User context")
			providers := GetBackupProviders()

			//For Backup Location 1
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v-ta-cc", "cloudcred", provider, time.Now().Unix())
				backupLocation = fmt.Sprintf("%v-bl-ta-2", time.Now().Unix())
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocation
				blUID2 = backupLocationUID
				bl2 = backupLocation
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocationWithContext(provider, backupLocation, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", nonAdminCtx, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocation))
			}

			//For Backup Location 2
			for _, provider := range providers {
				backupLocation = fmt.Sprintf("%v-bl-ta-1", time.Now().Unix())
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocation
				blUID1 = backupLocationUID
				bl1 = backupLocation
				err = CreateBackupLocationWithContext(provider, backupLocation, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", nonAdminCtx, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocation))

			}
		})

		//7.Share Ownership of the two Backup locations and cloud credentials with AllNormalUsers
		Step("Share Ownership of the Two Backup locations with AllNormalUsers", func() {
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching TestAdmin User context")
			//Add Ownership to User3 and User4 and User5
			for backupLocationUIDIteration, backupLocationIteration := range backupLocationMap {
				err = AddBackupLocationOwnership(backupLocationIteration, backupLocationUIDIteration, allNormalUsers, nil, Read, Invalid, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying updation of ownership for backuplocation - %s", backupLocationIteration))
			}
		})

		//9.Step-Registering Cluster for Backup testAdmin Context
		Step("Register cluster for backup", func() {
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching user context")
			err = CreateApplicationClusters(BackupOrgID, "", "", nonAdminCtx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, nonAdminCtx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			sourceClusterUid, err = Inst().Backup.GetClusterUID(nonAdminCtx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})

		//10.Step-Create Backups in all backup locations in testadminUser Context
		Step("Taking backup of applications", func() {
			//Common Steps for all backups
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", testAdminUserName)
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)

			//Backup1 Taken with BackupLocation1 of TestAdmin Context
			backupName = fmt.Sprintf("%s-%s-%v", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
			err = CreateBackupWithValidation(nonAdminCtx, backupName, SourceClusterName, bl1, blUID1, appContextsToBackup, labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupName))
			backupNames = append(backupNames, backupName)

			//Backup2 Taken with BackupLocation2 of TestAdmin Context
			backupName = fmt.Sprintf("%s-%s-%v", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
			err = CreateBackupWithValidation(nonAdminCtx, backupName, SourceClusterName, bl2, blUID2, appContextsToBackup, labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupName))
			backupNames = append(backupNames, backupName)

		})

		//13.Share All Backups with User3 with View Only access
		Step("Share the backups created by testAdmin User with ViewOnlyAccess with User3", func() {
			log.Infof("Share the backups created by testAdmin User with ViewOnlyAccess with User3")
			//For Sharing Backup Context needed is of Sharing Party i.e TestAdmin User
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", testAdminUserName)

			//Cluster Backup Share of all Backups with ViewOnly Access
			err = ClusterUpdateBackupShare(SourceClusterName, nil, []string{thirdUserName}, ViewOnlyAccess, true, nonAdminCtx)
			log.FailOnError(err, "Failed sharing all backups for cluster [%s]", SourceClusterName)

			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)

			//Validate the access of Shared Backup
			log.Infof("Validating the shared backups with ViewOnlyAccess from one of the user")
			userCtx, err := backup.GetNonAdminCtx(thirdUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", thirdUserName)
			bkVar := 0
			// Start Restore and confirm that user cannot restore
			for _, backupName = range backupNames {
				bkVar++
				ctx, err := backup.GetAdminCtxFromSecret()
				Inst().Dash.VerifyFatal(err, nil, "Fetching px-central-admin ctx")
				log.InfoD("Registering Source and Destination clusters from user context")
				err = CreateApplicationClusters(BackupOrgID, "", "", userCtx)
				Inst().Dash.VerifyFatal(err, nil, "Creating source and destination cluster")
				destClusterUid, err := Inst().Backup.GetClusterUID(userCtx, BackupOrgID, DestinationClusterName)
				Inst().Dash.VerifyFatal(err, nil, "Getting destination cluster UID")
				backupDriver := Inst().Backup
				restoreName := fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
				restoreNames = append(restoreNames, restoreName)
				Inst().Dash.VerifyFatal(err, nil, "Getting destination cluster UID")
				log.InfoD("Validating if user [%s] with access [%v] can restore and delete backup %s or not", thirdUserName, BackupAccessKeyValue[1], backupName)
				err = CreateRestoreWithValidation(userCtx, restoreName, backupName, make(map[string]string), make(map[string]string), DestinationClusterName, destClusterUid, BackupOrgID, appContextsToBackup)
				if err != nil {
					log.Infof("Error returned - %s", err.Error())
				}
				// Restore validation to make sure that the user with View Access cannot restore
				dash.VerifyFatal(strings.Contains(err.Error(), "doesn't have permission to restore backup") ||
					strings.Contains(err.Error(), "failed to retrieve backup location"), true, fmt.Sprintf("Verifying backup restore is not possible for viewonlyUser-iteration-%v", bkVar))
				backupUID, err := backupDriver.GetBackupUID(ctx, backupName, BackupOrgID)
				Inst().Dash.VerifyFatal(err, nil, fmt.Sprintf("Getting backup UID for- %s", backupName))
				// Delete backup to confirm that the user has ViewOnlyAccess and cannot delete backup
				_, err = DeleteBackup(backupName, backupUID, BackupOrgID, userCtx)
				log.Infof("The expected error returned is %v", err)
				Inst().Dash.VerifyFatal(strings.Contains(err.Error(), "doesn't have permission to delete backup"), true, "Verifying backup deletion is not possible")
			}
			log.InfoD("Finished verifying access level - ViewOnlyAccess")

		})

		//11.Share All Backups with User4 with Restore Only access
		Step("Share the backups created by testAdmin User with Restore Only Access with User4", func() {
			log.Infof("Share the backups created by testAdmin User with Restore Only Access with User4")
			//For Sharing Backup Context needed is of Sharing Party i.e TestAdmin User
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", testAdminUserName)

			//Cluster Backup Share of all Backups with Restore Access
			err = ClusterUpdateBackupShare(SourceClusterName, nil, []string{fourthUserName}, RestoreAccess, true, nonAdminCtx)
			log.FailOnError(err, "Failed sharing all backups for cluster [%s]", SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying share of all backups of cluster [%s] with [%#v] access level to user [%s]", SourceClusterName, RestoreAccess, fourthUserName))
			log.Infof("Validating the shared backups with RestoreAccess from one of the user")
			for _, backupName := range backupNames {
				restoreName := fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
				restoreNames = append(restoreNames, restoreName)
				ValidateSharedBackupWithUsers(fourthUserName, RestoreAccess, backupName, restoreName)
			}
			log.InfoD("Finished verifying access level - RestoreAccess")
		})

		//12.Share All Backups with Group1 with Full Access
		Step("Share the backups created by testAdmin User with Full Access with Group1", func() {
			log.Infof("Share the backups created by testAdmin User with Full Access with Group1")
			//For Sharing Backup Context needed is of Sharing Party i.e TestAdmin User
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", testAdminUserName)
			//Cluster Backup Share of all Backups with full Access
			err = ClusterUpdateBackupShare(SourceClusterName, []string{group1}, nil, FullAccess, true, nonAdminCtx)
			log.FailOnError(err, "Failed sharing all backups for cluster [%s]", SourceClusterName)

			//Validate the access of Shared Backups
			log.Infof("Validating the shared backups with FullAccess from one of the user")
			for _, backupName := range backupNames {

				restoreName := fmt.Sprintf("%s-%s-%v", fifthUserName, RestoreNamePrefix, RandomString(5))
				restoreNames = append(restoreNames, restoreName)
				ValidateSharedBackupWithUsers(fifthUserName, FullAccess, backupName, restoreName)

			}
			log.InfoD("Finished verifying access level - FullAccess")
		})

		//8.Step-Creating Cloud Credentials with User2 context and 1 Backup Location and sharing it with testadmin user
		Step("Creating backup location and cloud setting with User2 context", func() {
			log.InfoD("Creating backup location and cloud setting with User2 context")
			nonAdminCtx, err := backup.GetNonAdminCtx(secondUserName, CommonPassword)
			log.FailOnError(err, "Fetching User 2 context")
			providers := GetBackupProviders()
			//For Backup Location which is created in User2 context and shared with testAdmin User
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cloudcred-su", provider, time.Now().Unix())
				backupLocation = fmt.Sprintf("%v-bl-su-1", time.Now().Unix())
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocation
				blUID3 = backupLocationUID
				bl3 = backupLocation
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocationWithContext(provider, backupLocation, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", nonAdminCtx, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocation))

			}
			err = AddCloudCredentialOwnership(cloudCredName, cloudCredUID, []string{testAdminUserName}, nil, Read, Invalid, nonAdminCtx, BackupOrgID)
			dash.VerifyFatal(err, nil, "couldn't add ownership to Cloud Account")
			err = AddBackupLocationOwnership(bl3, blUID3, []string{testAdminUserName}, nil, Read, Invalid, nonAdminCtx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying updation of ownership for backuplocation - %s", []string{testAdminUserName}))
		})

		//8.Step-Creating Cloud Credentials with User2 context and 1 Backup Location and sharing it with testadmin user
		Step("Taking backup of applications", func() {
			//Common Steps for all backups
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", testAdminUserName)
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			//Backup3 Taken with BackupLocation of TestAdmin Context but with Backup Location which was shared to TestAdmin User by Second User.
			backupName = fmt.Sprintf("%s-%s-sharedbackup3", BackupNamePrefix, bkpNamespaces[0])
			err = CreateBackupWithValidation(nonAdminCtx, backupName, SourceClusterName, bl3, blUID3, appContextsToBackup, labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupName))
			secondarybackupNames = append(secondarybackupNames, backupName)
		})

		//11.Share the backup created by testAdmin User with backuplocation created by secondUser with Restore Only Access with User4
		Step("Share the backup created by testAdmin User with backuplocation created by secondUser with Restore Only Access with User4", func() {
			log.Infof("Share the backup created by testAdmin User with backuplocation created by secondUser with Restore Only Access with User4")
			//For Sharing Backup Context needed is of Sharing Party i.e TestAdmin User
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", testAdminUserName)

			//Cluster Backup Share of all Backups with Restore Access
			err = ClusterUpdateBackupShare(SourceClusterName, nil, []string{fourthUserName}, RestoreAccess, true, nonAdminCtx)
			log.FailOnError(err, "Failed sharing all backups for cluster [%s]", SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying share of all backups of cluster [%s] with [%#v] access level to user [%s]", SourceClusterName, RestoreAccess, fourthUserName))
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)

			log.Infof("Validating the shared backups with RestoreAccess from one of the user")
			userCtx, err := backup.GetNonAdminCtx(fourthUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", fourthUserName)
			for _, backupName := range secondarybackupNames {
				ctx, err := backup.GetAdminCtxFromSecret()
				Inst().Dash.VerifyFatal(err, nil, "Fetching px-central-admin ctx")
				log.InfoD("Registering Source and Destination clusters from user context")
				err = CreateApplicationClusters(BackupOrgID, "", "", userCtx)
				Inst().Dash.VerifyFatal(err, nil, "Creating source and destination cluster")
				destClusterUid, err := Inst().Backup.GetClusterUID(userCtx, BackupOrgID, DestinationClusterName)
				Inst().Dash.VerifyFatal(err, nil, "Getting destination cluster UID")
				backupDriver := Inst().Backup
				restoreName := fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
				restoreNames = append(restoreNames, restoreName)
				err = CreateRestoreWithValidation(userCtx, restoreName, backupName, make(map[string]string), make(map[string]string), DestinationClusterName, destClusterUid, BackupOrgID, appContextsToBackup)
				log.Infof("The expected error returned is %v", err)
				Inst().Dash.VerifyFatal(strings.Contains(err.Error(), "doesn't have permission to restore backup") ||
					strings.Contains(err.Error(), "failed to retrieve backup location"), true, "Verifying backup restore is not possible for restore-only-user4 3rd backup")
				// Try to delete the backup with user having RestoreAccess, and it should not pass
				backupUID, err := backupDriver.GetBackupUID(ctx, backupName, BackupOrgID)
				Inst().Dash.VerifyFatal(err, nil, fmt.Sprintf("Getting backup UID for- %s", backupName))
				// Delete backup to confirm that the user has Restore Access and delete backup should fail
				_, err = DeleteBackup(backupName, backupUID, BackupOrgID, userCtx)
				log.Infof("The expected error returned is %v", err)
				Inst().Dash.VerifyFatal(strings.Contains(err.Error(), "doesn't have permission to delete backup") ||
					strings.Contains(err.Error(), "failed to retrieve backup location"), true, "Verifying backup deletion is not possible for restore-only-user4 3rd backup")
			}
			log.InfoD("Finished verifying access level for shared backup with shared backuplocation - RestoreAccess")
		})

		//12.Share the backup created by testAdmin User with backuplocation created by secondUser with Full Access with Group1
		Step("Share the backup created by testAdmin User with backuplocation created by secondUser with Full Access with Group1", func() {
			log.Infof("Share the backup created by testAdmin User with backuplocation created by secondUser with Full Access with Group1")
			//For Sharing Backup Context needed is of Sharing Party i.e TestAdmin User
			nonAdminCtx, err := backup.GetNonAdminCtx(testAdminUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", testAdminUserName)
			//Cluster Backup Share of all Backups with full Access
			err = ClusterUpdateBackupShare(SourceClusterName, []string{group1}, nil, FullAccess, true, nonAdminCtx)
			log.FailOnError(err, "Failed sharing all backups for cluster [%s]", SourceClusterName)
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			log.Infof("Validating the shared backups with RestoreAccess from one of the user")
			userCtx, err := backup.GetNonAdminCtx(fifthUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", fifthUserName)
			//Validate the access of Shared Backups
			log.Infof("Validating the shared backups with FullAccess from one of the user")
			for _, backupName := range secondarybackupNames {
				ctx, err := backup.GetAdminCtxFromSecret()
				Inst().Dash.VerifyFatal(err, nil, "Fetching px-central-admin ctx")
				log.InfoD("Registering Source and Destination clusters from user context")
				err = CreateApplicationClusters(BackupOrgID, "", "", userCtx)
				Inst().Dash.VerifyFatal(err, nil, "Creating source and destination cluster")
				destClusterUid, err := Inst().Backup.GetClusterUID(userCtx, BackupOrgID, DestinationClusterName)
				Inst().Dash.VerifyFatal(err, nil, "Getting destination cluster UID")
				backupDriver := Inst().Backup
				restoreName := fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
				restoreNames = append(restoreNames, restoreName)
				err = CreateRestoreWithValidation(userCtx, restoreName, backupName, make(map[string]string), make(map[string]string), DestinationClusterName, destClusterUid, BackupOrgID, appContextsToBackup)
				log.Infof("The expected error returned is %v", err)
				Inst().Dash.VerifyFatal(strings.Contains(err.Error(), "doesn't have permission to restore backup") ||
					strings.Contains(err.Error(), "failed to retrieve backup location"), true, "Verifying backup restore is not possible for FullAccess-user5 3rd backup")
				// Try to delete the backup with user having RestoreAccess, and it should not pass
				backupUID, err := backupDriver.GetBackupUID(ctx, backupName, BackupOrgID)
				Inst().Dash.VerifyFatal(err, nil, fmt.Sprintf("Getting backup UID for- %s", backupName))
				// Delete backup to confirm that the user has Restore Access and delete backup should fail
				_, err = DeleteBackup(backupName, backupUID, BackupOrgID, userCtx)
				log.Infof("The expected error returned is %v", err)
				Inst().Dash.VerifyFatal(strings.Contains(err.Error(), "doesn't have permission to delete backup") ||
					strings.Contains(err.Error(), "failed to retrieve backup location"), true, "Verifying backup deletion is not possible for FullAccess-user5 3rd backup")

			}
			log.InfoD("Finished verifying access level for shared backup with shared backuplocation - FullAccess")
		})

	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		// Cleaning up applications created
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)

		// Clean the all user created
		var wg sync.WaitGroup
		log.Infof("Cleaning up users")
		for _, userName := range allInfraAdminUsers {
			wg.Add(1)
			go func(userName string) {
				defer GinkgoRecover()
				defer wg.Done()
				err := backup.DeleteUser(userName)
				log.FailOnError(err, "Error deleting user %v", userName)
			}(userName)
		}
		for _, userName := range allNormalUsers {
			wg.Add(1)
			go func(userName string) {
				defer GinkgoRecover()
				defer wg.Done()
				err := backup.DeleteUser(userName)
				log.FailOnError(err, "Error deleting user %v", userName)
			}(userName)
		}
		wg.Wait()

		// Clean up the all group created
		log.Infof("Cleaning up groups")
		err := backup.DeleteGroup(group1)
		log.FailOnError(err, "Error deleting user %v", group1)

		// Clean up the cluster
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		for _, restoreNameIteration := range restoreNames {
			err = DeleteRestore(restoreNameIteration, BackupOrgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting restore [%s]", restoreNameIteration))
		}
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)

	})
})

var _ = Describe("{ClusterBackupShareWithAnotherUserHavingSameClusterNameOfItsOwn}", Label(TestCaseLabelsMap[ValidateUserAccessLevel]...), func() {
	var (
		backupName           string
		scheduledAppContexts []*scheduler.Context
		sourceClusterUid     string
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		cloudCredName        string
		cloudCredUID         string
		backupLocationUID    string
		backupLocation       string
		bkpNamespaces        []string
		allInfraAdminUsers   []string
		firstUserName        string
		secondUserName       string
		bl1                  string
		blUID1               string
		bl2                  string
		blUID2               string
		restoreNames         []string
		backupUidBackup1     string
		labelSelectors       map[string]string
		backupLocationMap    map[string]string
		backupNames          []string
		firstUserNameCtx     context.Context
		ctx                  context.Context
		secondUserNameCtx    context.Context
		err                  error
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyClusterBackupShareWithAnotherUserHavingSameClusterNameOfItsOwn",
			"Cluster backup share with another user having same cluster name of its own", nil, 300638, ABadgujar, Q2FY24)
		//1.Step-Deploy Applications in the Cluster
		bkpNamespaces = make([]string, 0)
		labelSelectors = make(map[string]string)
		backupLocationMap = make(map[string]string)
		backupNames = make([]string, 0)
		ctx, err = backup.GetAdminCtxFromSecret()
		Inst().Dash.VerifyFatal(err, nil, "Fetching px-central-admin ctx")

		log.InfoD("Deploy applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = AppReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
				scheduledAppContexts = append(scheduledAppContexts, ctx)
			}
		}
	})
	It("Cluster backup share with another user having same cluster name of its own", func() {
		//2.Step-Validate Deployed Applications
		Step("Validate applications", func() {
			log.InfoD("Validating apps")
			ValidateApplications(scheduledAppContexts)
		})

		//3.Step-Create 2 Users and assign infra admin role to both
		Step("Create 2 User with Infra Admin Role", func() {
			log.InfoD("Create 2 User with Infra Admin Role")
			role := backup.InfrastructureOwner
			allInfraAdminUsers = CreateUsers(2)
			//Assign Infra Admin Role to both users
			for _, userName := range allInfraAdminUsers {
				err := backup.AddRoleToUser(userName, role, fmt.Sprintf("Adding %v role to %s", role, userName))
				log.FailOnError(err, "Failed to add role for user - %s", userName)
			}
			firstUserName = allInfraAdminUsers[0]
			firstUserNameCtx, err = backup.GetNonAdminCtx(firstUserName, CommonPassword)
			log.FailOnError(err, "Fetching firstUserName User context")

			secondUserName = allInfraAdminUsers[1]
			secondUserNameCtx, err = backup.GetNonAdminCtx(secondUserName, CommonPassword)
			log.FailOnError(err, "Fetching secondUserName User context")
		})

		//4.Step-Creating Backup Location Cloud setting in firstUserName User Context
		Step("Creating Backup Location Cloud setting in firstUserName User Context", func() {
			log.InfoD("Creating Backup Location Cloud setting in firstUserName User Context")

			providers := GetBackupProviders()
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v-u1", "cloudcred", provider, time.Now().Unix())
				backupLocation = fmt.Sprintf("%v-u1", time.Now().Unix())
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocation
				bl1 = backupLocation
				blUID1 = backupLocationUID
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, firstUserNameCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocationWithContext(provider, bl1, blUID1, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", firstUserNameCtx, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocation))
			}
		})

		//5.Step-Registering Cluster for Backup firstUserName Context
		Step("Registering Cluster for Backup firstUserName Context", func() {
			log.InfoD("Registering Cluster for Backup firstUserName Context")
			err := CreateApplicationClusters(BackupOrgID, "", "", firstUserNameCtx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, firstUserNameCtx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			sourceClusterUid, err = Inst().Backup.GetClusterUID(firstUserNameCtx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})

		//6.Step-Create Backups in all backup locations in firstUserName Context
		Step("Create Backups in all backup locations in firstUserName Context", func() {
			log.InfoD("Create Backups in all backup locations in firstUserName Context")
			//Common Steps for all backups
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)

			//Backup Taken with BackupLocation1 of firstUserName Context
			backupName = fmt.Sprintf("%s-common", BackupNamePrefix)
			err := CreateBackupWithValidation(firstUserNameCtx, backupName, SourceClusterName, bl1, blUID1, appContextsToBackup, labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupName))
			backupNames = append(backupNames, backupName)

			backupUidBackup1, err = Inst().Backup.GetBackupUID(ctx, backupNames[0], BackupOrgID)
			log.FailOnError(err, "Error Fetching Backup UID for backup 1")

		})

		//7.Share Backup with SecondUserName from firstUserName context
		Step("Share Backup with SecondUserName from firstUserName context", func() {
			log.Infof("Share Backup with SecondUserName from firstUserName context")
			//Cluster Backup Share of all Backups with Restore Access

			err := ShareBackup(backupNames[0], nil, []string{secondUserName}, RestoreAccess, firstUserNameCtx)
			log.FailOnError(err, "Failed to share backup %s", backupNames[0])

		})

		//8.Restore to a new namespace and see if access is there for backup of firstUserName in secondUserName context
		Step("Restore to a new namespace and see if access is there for backup of firstUserName in secondUserName context", func() {
			log.Infof("Restore to a new namespace and see if access is there for backup of firstUserName in secondUserName context")
			restoreName := fmt.Sprintf("%s-%s-%v", secondUserName, RestoreNamePrefix, RandomString(5))
			restoreNames = append(restoreNames, restoreName)
			backupUid, err := Inst().Backup.GetBackupUID(ctx, backupNames[0], BackupOrgID)
			log.FailOnError(err, "Error Fetching Backup UID")

			dash.VerifyFatal(backupUid, backupUidBackup1, "Verifying Correct Backup for restore i.e backup UIDs are matching as names are the same")
			log.InfoD("Validation if restore successful for Backup with %v - Backup and %v - Backup UID", backupNames[0], backupUid)
			ValidateSharedBackupWithUsers(secondUserName, RestoreAccess, backupNames[0], restoreName)

		})

		//9.Step-Creating Backup location and Cloud setting in secondUserName User Context
		Step("Creating Backup location and Cloud setting in secondUserName User Context", func() {
			log.InfoD("Creating Backup location and Cloud setting in secondUserName User Context")

			providers := GetBackupProviders()
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v-u2", "cloudcred", provider, time.Now().Unix())
				backupLocation = fmt.Sprintf("%v-u2", time.Now().Unix())
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocation
				bl2 = backupLocation
				blUID2 = backupLocationUID
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, secondUserNameCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocationWithContext(provider, bl2, blUID2, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", secondUserNameCtx, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocation))
			}
		})

		//10.Step-Registering Cluster for Backup secondUserName Context
		Step("Registering Cluster for Backup secondUserName Context", func() {
			log.InfoD("Registering Cluster for Backup secondUserName Context")
			err := CreateApplicationClusters(BackupOrgID, "", "", secondUserNameCtx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, secondUserNameCtx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			sourceClusterUid, err = Inst().Backup.GetClusterUID(secondUserNameCtx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})

		//11.Step-Create Backups in secondUserName Context
		Step("Create Backups in secondUserName Context", func() {
			log.Infof("Create Backups in secondUserName Context")
			//Common Steps for all backups
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)

			//Backup Taken with BackupLocation1 of secondUserName Context
			backupName = fmt.Sprintf("%s-common", BackupNamePrefix)
			err := CreateBackupWithValidation(secondUserNameCtx, backupName, SourceClusterName, bl2, blUID2, appContextsToBackup, labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupName))
			backupNames = append(backupNames, backupName)

		})

		//12.Step-Share Backup with firstUserName from secondUserName context
		Step("Share Backup with firstUserName from secondUserName context", func() {
			log.Infof("Share Backup with firstUserName from secondUserName context")
			//Backup Share of all Backups with Restore Access
			err := ShareBackup(backupNames[1], nil, []string{firstUserName}, RestoreAccess, secondUserNameCtx)
			log.FailOnError(err, "Failed to share backup %s", backupNames[1])
		})

		//13.Step-Restore to a new namespace and see if access is there for backup of secondUserName in firstUserName context
		Step("Restore to a new namespace and see if access is there for backup of secondUserName in firstUserName context", func() {
			log.Infof("Restore to a new namespace and see if access is there for backup of secondUserName in firstUserName context")
			restoreName := fmt.Sprintf("%s-%s-%v", firstUserName, RestoreNamePrefix, RandomString(5))
			restoreNames = append(restoreNames, restoreName)
			ValidateSharedBackupWithUsers(firstUserName, RestoreAccess, backupNames[1], restoreName)
		})

	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		// Cleaning up applications created
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)

		// Clean the all user created

		log.Infof("Cleaning up users")
		err := CleanupAllUserAndGroups()
		dash.VerifySafely(err, nil, "Verifying cleanup all user and groups")

		// Clean up the cluster
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")

		for _, restoreNameIteration := range restoreNames {
			err = DeleteRestore(restoreNameIteration, BackupOrgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting restore [%s]", restoreNameIteration))
		}
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)

	})
})

var _ = Describe("{RestoringFromBothBackupAndAnotherSharedBackupWithSameName}", Label(TestCaseLabelsMap[ValidateUserAccessLevel]...), func() {
	var (
		backupName           string
		scheduledAppContexts []*scheduler.Context
		sourceClusterUid     string
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		cloudCredName        string
		cloudCredUID         string
		backupLocationUID    string
		backupLocation       string
		bkpNamespaces        []string
		allInfraAdminUsers   []string
		firstUserName        string
		secondUserName       string
		bl1                  string
		blUID1               string
		bl2                  string
		blUID2               string
		restoreNames         []string
		backupUidBackup1     string
		labelSelectors       map[string]string
		backupLocationMap    map[string]string
		backupNames          []string
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyRestoringFromBothBackupAndAnotherSharedBackupWithSameName",
			"Restoring from both backup and another shared backup with same name to check if restore is successful", nil, 300639, ABadgujar, Q2FY24)

		bkpNamespaces = make([]string, 0)
		labelSelectors = make(map[string]string)
		backupLocationMap = make(map[string]string)
		backupNames = make([]string, 0)

		//1.Step-Deploy Applications in the Cluster
		log.InfoD("Deploy applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = AppReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
				scheduledAppContexts = append(scheduledAppContexts, ctx)
			}
		}

	})
	It("Restoring from both backup and another shared backup with same name to check if restore is successful", func() {
		//2.Step-Validate Deployed Applications
		Step("Validate applications", func() {
			log.InfoD("Validating apps")
			ValidateApplications(scheduledAppContexts)
		})

		//3.Step-Create 2 Users and assign infra admin role to both
		Step("Create 2 User with Infra Admin Role", func() {
			log.InfoD("Create 2 User with Infra Admin Role")
			role := backup.InfrastructureOwner
			allInfraAdminUsers = CreateUsers(2)
			//Assign Infra Admin Role to both users
			firstUserName = allInfraAdminUsers[0]
			secondUserName = allInfraAdminUsers[1]
			for _, userName := range allInfraAdminUsers {
				err := backup.AddRoleToUser(userName, role, fmt.Sprintf("Adding %v role to %s", role, userName))
				log.FailOnError(err, "Failed to add role for user - %s", userName)
			}
		})

		//4.Step-Creating Backup Location Cloud setting in firstUserName User Context
		Step("Creating Backup Location Cloud setting in firstUserName User Context", func() {
			log.InfoD("Creating Backup Location Cloud setting in firstUserName User Context")
			nonAdminCtx, err := backup.GetNonAdminCtx(firstUserName, CommonPassword)
			log.FailOnError(err, "Fetching TestAdmin User context")
			providers := GetBackupProviders()
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v-u1", "cloudcred", provider, time.Now().Unix())
				backupLocation = fmt.Sprintf("%v-u1", time.Now().Unix())
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocation
				bl1 = backupLocation
				blUID1 = backupLocationUID
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocationWithContext(provider, bl1, blUID1, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", nonAdminCtx, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocation))
			}
		})

		//5.Step-Registering Cluster for Backup firstUserName Context
		Step("Registering Cluster for Backup firstUserName Context", func() {
			log.InfoD("Registering Cluster for Backup firstUserName Context")
			nonAdminCtx, err := backup.GetNonAdminCtx(firstUserName, CommonPassword)
			log.FailOnError(err, "Fetching user context")
			err = CreateApplicationClusters(BackupOrgID, "", "", nonAdminCtx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, nonAdminCtx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			sourceClusterUid, err = Inst().Backup.GetClusterUID(nonAdminCtx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})

		//6.Step-Create Backups in all backup locations in firstUserName Context
		Step("Create Backups in all backup locations in firstUserName Context", func() {
			log.InfoD("Create Backups in all backup locations in firstUserName Context")
			//Common Steps for all backups
			nonAdminCtx, err := backup.GetNonAdminCtx(firstUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", firstUserName)
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)

			//Backup Taken with BackupLocation1 of firstUserName Context
			backupName = fmt.Sprintf("%s-common", BackupNamePrefix)
			err = CreateBackupWithValidation(nonAdminCtx, backupName, SourceClusterName, bl1, blUID1, appContextsToBackup, labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupName))
			backupNames = append(backupNames, backupName)

			ctx, err := backup.GetAdminCtxFromSecret()
			Inst().Dash.VerifyFatal(err, nil, "Fetching px-central-admin ctx")
			backupUidBackup1, err = Inst().Backup.GetBackupUID(ctx, backupNames[0], BackupOrgID)
			log.FailOnError(err, "Error Fetching Backup UID for backup 1")

		})

		//7.Share Cluster and Backup with SecondUserName from firstUserName context
		Step("Share Cluster and Backup with SecondUserName from firstUserName context", func() {
			log.Infof("Share Cluster and Backup with SecondUserName from firstUserName context")
			//For Sharing Backup Context needed is of Sharing Party i.e firstUserName User
			nonAdminCtx, err := backup.GetNonAdminCtx(firstUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", firstUserName)
			//Cluster Backup Share of all Backups with Restore Access
			err = ClusterUpdateBackupShare(SourceClusterName, nil, []string{secondUserName}, RestoreAccess, true, nonAdminCtx)
			log.FailOnError(err, "Failed sharing all backups for cluster [%s]", SourceClusterName)
		})

		//8.Restore to a new namespace and see if access is there for backup of firstUserName in secondUserName context
		Step("Restore to a new namespace and see if access is there for backup of firstUserName in secondUserName context", func() {
			log.Infof("Restore to a new namespace and see if access is there for backup of firstUserName in secondUserName context")
			ctx, err := backup.GetAdminCtxFromSecret()
			Inst().Dash.VerifyFatal(err, nil, "Fetching px-central-admin ctx")
			restoreName := fmt.Sprintf("%s-%s-%v", secondUserName, RestoreNamePrefix, RandomString(5))
			restoreNames = append(restoreNames, restoreName)
			backupUid, err := Inst().Backup.GetBackupUID(ctx, backupNames[0], BackupOrgID)
			log.FailOnError(err, "Error Fetching Backup UID")

			//Since Backup Names are same verify if both different backup are being restored and correct ones too
			dash.VerifyFatal(backupUid, backupUidBackup1, "Verifying Correct Backup for restore i.e backup UIDs are matching as names are the same")

			log.InfoD("Validation if restore successful for Backup with %v - Backup and %v - Backup UID", backupNames[0], backupUid)
			ValidateSharedBackupWithUsers(secondUserName, RestoreAccess, backupNames[0], restoreName)

		})

		//9.Step-Creating Backup location and Cloud setting in secondUserName User Context
		Step("Creating Backup location and Cloud setting in secondUserName User Contextt", func() {
			log.InfoD("Creating Backup location and Cloud setting in secondUserName User Context")
			nonAdminCtx, err := backup.GetNonAdminCtx(secondUserName, CommonPassword)
			log.FailOnError(err, "Fetching secondUserName User context")
			providers := GetBackupProviders()
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v-u2", "cloudcred", provider, time.Now().Unix())
				backupLocation = fmt.Sprintf("%v-u2", time.Now().Unix())
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocation
				bl2 = backupLocation
				blUID2 = backupLocationUID
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, nonAdminCtx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocationWithContext(provider, bl2, blUID2, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", nonAdminCtx, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocation))
			}
		})

		//10.Step-Registering Cluster for Backup secondUserName Context
		Step("Registering Cluster for Backup secondUserName Context", func() {
			log.InfoD("Registering Cluster for Backup secondUserName Context")
			nonAdminCtx, err := backup.GetNonAdminCtx(secondUserName, CommonPassword)
			log.FailOnError(err, "Fetching user context")
			err = CreateApplicationClusters(BackupOrgID, "", "", nonAdminCtx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, nonAdminCtx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			sourceClusterUid, err = Inst().Backup.GetClusterUID(nonAdminCtx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})

		//11.Step-Create Backups in secondUserName Context
		Step("Create Backups in secondUserName Context", func() {
			log.Infof("Create Backups in secondUserName Context")
			//Common Steps for all backups
			nonAdminCtx, err := backup.GetNonAdminCtx(secondUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", secondUserName)
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)

			//Backup Taken with BackupLocation1 of secondUserName Context
			backupName = fmt.Sprintf("%s-common", BackupNamePrefix)
			err = CreateBackupWithValidation(nonAdminCtx, backupName, SourceClusterName, bl2, blUID2, appContextsToBackup, labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupName))
			backupNames = append(backupNames, backupName)

		})

		//12.Step-Share Cluster and Backup with firstUserName from secondUserName context
		Step("Share Cluster and Backup with firstUserName from secondUserName context", func() {
			log.Infof("Share Cluster and Backup with firstUserName from secondUserName context")
			//For Sharing Backup Context needed is of Sharing Party i.e secondUserName User
			nonAdminCtx, err := backup.GetNonAdminCtx(secondUserName, CommonPassword)
			log.FailOnError(err, "Fetching user [%s] ctx", secondUserName)
			//Cluster Backup Share of all Backups with Restore Access
			err = ClusterUpdateBackupShare(SourceClusterName, nil, []string{firstUserName}, RestoreAccess, true, nonAdminCtx)
			log.FailOnError(err, "Failed sharing all backups for cluster [%s]", SourceClusterName)
		})

		//13.Step-Restore to a new namespace and see if access is there for backup of secondUserName in firstUserName context
		Step("Restore to a new namespace and see if access is there for backup of secondUserName in firstUserName context", func() {
			log.Infof("Restore to a new namespace and see if access is there for backup of secondUserName in firstUserName context")
			restoreName := fmt.Sprintf("%s-%s-%v", firstUserName, RestoreNamePrefix, RandomString(5))
			restoreNames = append(restoreNames, restoreName)
			log.InfoD("Validation if restore successful for Backup")
			ValidateSharedBackupWithUsers(firstUserName, RestoreAccess, backupNames[1], restoreName)
		})

	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		log.InfoD("Deleting the deployed apps after the testcase")
		// Cleaning up applications created
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)

		enumerateBackupReq := &api.BackupEnumerateRequest{
			OrgId: BackupOrgID,
		}
		backups, err := Inst().Backup.EnumerateBackup(ctx, enumerateBackupReq)
		dash.VerifyFatal(err, nil, "Enumerating backups")
		// Delete the backups
		for _, backup := range backups.GetBackups() {
			backupUid, err := Inst().Backup.GetBackupUID(ctx, backup.Name, BackupOrgID)
			log.FailOnError(err, "Unable to fetch backup UID")
			err = DeleteBackupAndWaitForCompletion(backup.Name, backupUid, BackupOrgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting backup [%s]", backup))
		}

		log.Infof("Cleaning up users")
		err = CleanupAllUserAndGroups()
		dash.VerifySafely(err, nil, "Verifying cleanup all user and groups")
		// Clean up the cluster

		for _, restoreNameIteration := range restoreNames {
			err = DeleteRestore(restoreNameIteration, BackupOrgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting restore [%s]", restoreNameIteration))
		}
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)

	})
})

// This test case ensures that the error message for unauthorized operations includes the user's details and role when access is denied.
var _ = Describe("{VerifyAccessDeniedErrorMessagesforUnAuthorizedOperations}", Label(TestCaseLabelsMap[ValidateUserAccessLevel]...), func() {
	var (
		customUserCtx  context.Context
		providers      []string
		cloudCredName  string
		cloudCredUID   string
		customUser     string
		firstName      string
		lastName       string
		email          string
		userID         string
		customRoleName backup.PxBackupRole = backup.ApplicationUser
		err            error
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyAccessDeniedErrorMessagesforUnAuthorizedOperations",
			"Validate Error Messages for Unauthorized User Actions", nil, 300687, Nvettaiyan, Q1FY25)
		providers = GetBackupProviders()
	})

	It("Ensure Proper Error Messages For Unauthorized Action Attempts", func() {
		Step("Create a new user", func() {
			log.InfoD("Create a new user")
			customUser = fmt.Sprintf("testuser-%s", RandomString(5))
			firstName = fmt.Sprintf("FirstName-%s", customUser)
			lastName = fmt.Sprintf("LastName-%s", customUser)
			email = fmt.Sprintf("%v@cnbu.com", customUser)
			err = backup.AddUser(customUser, firstName, lastName, email, CommonPassword)
			log.FailOnError(err, "Failed to create user - %s", customUser)
			userID, err = backup.FetchIDOfUser(customUser)
			log.FailOnError(err, "Failed to fetch user id - %s", userID)
			log.InfoD("User Information - [User: %s] Name: %s %s Email: %s", customUser, firstName, lastName, email)
		})

		Step("Add custom role to the user", func() {
			log.InfoD(fmt.Sprintf("Add role to the user [%s]", customRoleName))
			err = backup.AddRoleToUser(customUser, customRoleName, fmt.Sprintf("Adding %v role to %s", customRoleName, customUser))
			log.FailOnError(err, "failed to add role %s to the user %s", customRoleName, customUser)
			log.Infof("username %s common password %s", customUser, CommonPassword)
		})

		Step("Creating backup location and cloud setting", func() {
			log.InfoD("Creating backup location and cloud setting")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cloudcred", provider, RandomString(10))
				cloudCredUID = uuid.New()
				customUserCtx, err = backup.GetNonAdminCtx(customUser, CommonPassword)
				log.FailOnError(err, "Fetching non admin ctx")
				err = CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, customUserCtx)
				if err != nil {
					firstName = strings.TrimSpace(firstName)
					lastName = strings.TrimSpace(lastName)
					dash.VerifyFatal(
						strings.Contains(err.Error(), firstName) && strings.Contains(err.Error(), lastName) && strings.Contains(err.Error(), string(customRoleName)),
						true,
						fmt.Sprintf("Verifying if the error message contains user's first name [%s], last name [%s], and role [%s]. Actual error: %s", firstName, lastName, customRoleName, err.Error()),
					)
				} else {
					log.FailOnError(err, fmt.Sprintf("Unexpected success when creating cloud credential %s — Access denied or missing", cloudCredName))
				}
			}
		})
	})

	JustAfterEach(func() {
		customUserCtx, err = backup.GetNonAdminCtx(customUser, CommonPassword)
		log.FailOnError(err, "Fetching non admin ctx")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		log.Infof("Cleaning up users")
		err := CleanupAllUserAndGroups()
		dash.VerifySafely(err, nil, "Verifying cleanup all user and groups")
		CleanupCloudSettingsAndClusters(nil, cloudCredName, cloudCredUID, customUserCtx)
	})
})
