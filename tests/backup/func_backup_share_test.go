package tests

import (
	"context"
	"fmt"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
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
