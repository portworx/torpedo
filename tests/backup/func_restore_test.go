package tests

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/pure-px/torpedo/drivers/backup"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
)

// Try to restore when token manager is not reachable
var _ = Describe("{TestRestoreResilienceToKeycloakScale}", Label(TestCaseLabelsMap[PxBackupLabel]...), func() {
	var (
		scheduledAppContexts   []*scheduler.Context
		bkpNamespaces          []string
		clusterUid             string
		clusterStatus          api.ClusterInfo_StatusInfo_Status
		backupName             string
		backupLocationUID      string
		cloudCredName          string
		cloudCredUID           string
		bkpLocationName        string
		backupNames            []string
		providers              []string
		ctx                    context.Context
		backupLocationMap      map[string]string
		err                    error
		restoreName            string
		numOfBackup            int
		pxbNamespace           string
		scaledDownReplicaCount int32
		originalReplicaCount   int32
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyTestRestoreResilienceToKeycloakScale", "Try to restore when token manager is not reachable", nil, 3006567, Pingle, Q3FY25)

		backupLocationMap = make(map[string]string)
		bkpNamespaces = make([]string, 0)
		backupNames = make([]string, 0)
		scheduledAppContexts = make([]*scheduler.Context, 0)
		numOfBackup = 1
		scaledDownReplicaCount = 0
		originalReplicaCount = 1
		ctx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		providers = GetBackupProviders()

		// Schedule an Application
		appContexts := ScheduleApplications(TaskNamePrefix)
		for _, ctx := range appContexts {
			ctx.ReadinessTimeout = AppReadinessTimeout
			namespace := GetAppNamespace(ctx, TaskNamePrefix)
			bkpNamespaces = append(bkpNamespaces, namespace)
			scheduledAppContexts = append(scheduledAppContexts, ctx)
		}
	})

	//Try to restore when token manager is not reachable
	It("Try to restore when token manager is not reachable", func() {
		// 1. validate application
		Step("Validate applications", func() {
			log.InfoD("Validating applications")
			ValidateApplications(scheduledAppContexts)
		})

		// 2. Create cloud credentials and backup location
		Step("Creating cloud credentials and backup location", func() {
			log.InfoD("Creating cloud credentials and backup location")
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

		// 3. Create application cluster for backup
		Step("Register cluster for backup", func() {
			err := CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
			log.InfoD("Uid of [%s] cluster is %s", SourceClusterName, clusterUid)
		})

		// 4. Take backup of applications
		Step("Taking backup of applications", func() {
			for i := 0; i < numOfBackup; i++ {
				backupName = fmt.Sprintf("%s-%s", BackupNamePrefix, RandomString(6))
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
				err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, bkpLocationName, backupLocationUID, appContextsToBackup, nil, BackupOrgID, clusterUid, "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				backupNames = append(backupNames, backupName)
			}
		})

		// 5.create restore
		Step("Restoring the backups application", func() {
			restoreName = fmt.Sprintf("%s-restore", backupName)
			for _, backupName := range backupNames {
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-central-admin ctx")
				_, err = CreateRestoreWithoutCheck(restoreName, backupName, nil, SourceClusterName, clusterUid, BackupOrgID, ctx)
				log.FailOnError(err, "Failed while trying to restore [%s] the backup [%s]", fmt.Sprintf("%s-restore", backupName), backupName)
			}
		})

		// 6. Scale down pxcentral-keycloak
		Step("Scale down pxcentral-keycloak", func() {
			log.Infof("Scaling down %s to %d replicas in namespace %s", KeyCloakStateFulSet, scaledDownReplicaCount, pxbNamespace)
			pxbNamespace, err = backup.GetPxBackupNamespace()
			dash.VerifyFatal(err, nil, "failed to get Px-Backup namespace")
			err = ScaleStatefulSetReplicas(KeyCloakStateFulSet, pxbNamespace, scaledDownReplicaCount, 0, PodStatusTimeOut, PodStatusRetryTime)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Scale of: %s with scale count: %d", KeyCloakStateFulSet, scaledDownReplicaCount))
		})

		// 7. Scale up pxcentral-keycloak to actual state
		Step("Scale up pxcentral-keycloak", func() {
			log.Infof("Scaling up %s to %d replicas in namespace %s", KeyCloakStateFulSet, originalReplicaCount, pxbNamespace)
			err = ScaleStatefulSetReplicas(KeyCloakStateFulSet, pxbNamespace, originalReplicaCount, 1, PodStatusTimeOut, PodStatusRetryTime)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Scale of: %s with scale count: %d", KeyCloakStateFulSet, originalReplicaCount))
		})

		// 8. Success check for restore
		Step("Success check for restore", func() {
			log.Infof("Checking success state for the backup %s", backupName)
			err := RestoreSuccessCheck(restoreName, BackupOrgID, MaxWaitPeriodForRestoreCompletionInMinute*time.Minute, 30*time.Second, ctx)
			dash.VerifyFatal(err, nil, "Verify restore success check")
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		// Clean up the cluster
		log.InfoD("deleting restores")
		backupUID, err := Inst().Backup.GetBackupUID(ctx, backupName, BackupOrgID)
		log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupName)
		_, err = DeleteBackup(backupName, backupUID, BackupOrgID, ctx)
		dash.VerifySafely(err, nil, fmt.Sprintf("Verifying backup deletion : %v", backupName))

		// Delete restore
		err = DeleteRestore(restoreName, BackupOrgID, ctx)
		dash.VerifySafely(err, nil, fmt.Sprintf("deleting Restore [%s]", restoreName))
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)

	})
})
