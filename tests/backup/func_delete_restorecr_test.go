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

// This test deletes restore custom resource.
var _ = Describe("{DeleteRestoreCustomResource}", Label(TestCaseLabelsMap[KDMPBackup]...), func() {
	var (
		backupName           string
		scheduledAppContexts []*scheduler.Context
		clusterUid           string
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		cloudCredName        string
		cloudCredUID         string
		backupLocationUID    string
		backupLocationMap    map[string]string
		restoreName          string
		bkpLocationName      string
		err                  error
		ctx                  context.Context
		providers            []string
		bkpNamespaces        []string
		backupNames          []string
	)
	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("VerifyDeleteRestoreCustomResource", "Verify PxRestoreDeleteCR for restore delete and check status.", nil, 91966, Pingle, Q2FY25)
		ctx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		providers = GetBackupProviders()
		scheduledAppContexts = make([]*scheduler.Context, 0)
		backupLocationMap = make(map[string]string)
		bkpNamespaces = make([]string, 0)
		// Schedule an Application
		appContexts := ScheduleApplications(TaskNamePrefix)
		for _, ctx := range appContexts {
			ctx.ReadinessTimeout = AppReadinessTimeout
			namespace, err := backup.GetPxBackupNamespace()
			dash.VerifyFatal(err, nil, fmt.Sprintf("pxbNamespace :: %s", namespace))
			bkpNamespaces = append(bkpNamespaces, namespace)
			scheduledAppContexts = append(scheduledAppContexts, ctx)
		}
	})

	//Test case to delete custom resource while restore in progress
	It("Test to validate and delete CRs before completion of restore", func() {
		numOfBackup := 1
		var customNamespace string
		// 1. Validate application
		Step("Validate applications", func() {
			log.Infof("Validate applications")
			ValidateApplications(scheduledAppContexts)

		})

		// 3. Create backup location and cloud setting
		Step("Creating backup location and cloud setting", func() {
			log.InfoD("Creating backup location and cloud setting")
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

		// 4. Create application cluster for backup.
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

		// 5. Take the backup of bkNamespaces.
		Step("Taking backup of applications", func() {
			for i := 0; i < numOfBackup; i++ {
				backupDriver := Inst().Backup
				backupName = fmt.Sprintf("%s-%s-%v", BackupNamePrefix, bkpNamespaces[0], time.Now().Unix())
				params := map[string]string{"backupName": backupName, "backupOrgID": BackupOrgID, "bkpLocationName": bkpLocationName, "backupLocationUID": backupLocationUID, "clusterName": SourceClusterName, "clusterUid": clusterUid}
				bkpCreateRequest, err := PrepareGenericBackupRequest(params)
				dash.VerifyFatal(err, nil, fmt.Sprintf("decorating data of KDMP backup [%s]", backupName))
				bkpCreateRequest.Namespaces = []string{bkpNamespaces[0]}
				log.InfoD("Backup without check [%s] started at [%s]", backupName, time.Now().Format("2006-01-02 15:04:05"))
				_, err = backupDriver.CreateBackup(ctx, bkpCreateRequest)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of KDMP backup [%s]", backupName))
				err = backupDriver.WaitForBackupCompletion(ctx, backupName, BackupOrgID, MaxWaitPeriodForBackupCompletionInMinutes*time.Minute, 30*time.Second)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Waiting for completion of KDMP backup [%s]", backupName))
				backupNames = append(backupNames, backupName)
			}
		})

		// 7.create restore for the negative test
		Step("Create restore for negative case", func() {
			log.Infof("Restore for negative case started")
			restoreName = fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
			customNamespace = bkpNamespaces[0] + RandomString(4)
			namespaceMapping := map[string]string{bkpNamespaces[0]: customNamespace}
			_, err := CreateRestoreWithoutCheck(restoreName, backupName, namespaceMapping, SourceClusterName, clusterUid, BackupOrgID, ctx)
			log.InfoD("Negative restoring of restore [%s] has failed with name - [%s]", restoreName, err)
			err = RestoreInprogressCheck(restoreName, BackupOrgID, MaxWaitPeriodForRestoreCompletionInMinute*time.Minute, 2*time.Second, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying RestoreInprogressCheck for Restore - %s", restoreName))
		})

		// 8.Delete restore application CR
		Step("Delete restoreCR ", func() {
			log.InfoD("started deleting restore cr...")
			err := DeleteRestoreCRCmd()
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying restoreCR deletion - %s", restoreName))
			err = RestoreFailCheck(restoreName, BackupOrgID, MaxWaitPeriodForRestoreCompletionInMinute*time.Minute, 30*time.Second, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying RestoreFailCheck for DeleteRestoreCR - %s", restoreName))
		})

		// 9.Verfify the failed status for restore from metrics.
		Step("Verify the failed status of the backup-restore from metrics endpoint", func() {
			metricsName := "pxbackup_restore_status"
			pxbNamespace, err := backup.GetPxBackupNamespace()
			dash.VerifyFatal(err, nil, "Getting px-backup namespace")

			allMetricsData, err := RunCurlCmd(pxbNamespace)
			log.Infof("Fetching all metrics data is")
			log.FailOnError(err, "Fetching all metrics data is")

			expected := "4"
			status := GetMetricValue(allMetricsData, metricsName, restoreName)
			log.Infof("pxbackup_restore_status recived:", status)
			dash.VerifyFatal(status, expected, "Verify the failure status of the pxbackup_backup_restore_status from metrics endpoint")
		})

	})

	JustAfterEach(func() {
		// Cleaning up px-backup cluster
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")

		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		log.InfoD("deleting restores")
		err = DeleteRestore(restoreName, BackupOrgID, ctx)
		dash.VerifySafely(err, nil, fmt.Sprintf("deleting Restore [%s]", restoreName))

		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)

		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})
