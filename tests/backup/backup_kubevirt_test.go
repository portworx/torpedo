package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	"sync"
)

// This testcase verifies backup and restore of Kubevirt VMs in different states like Running, Stopped, Restarting
var _ = Describe("{KubevirtVMBackupRestoreWithDifferentStates}", func() {

	var (
		backupNames          []string
		restoreNames         []string
		scheduledAppContexts []*scheduler.Context
		sourceClusterUid     string
		cloudCredName        string
		cloudCredUID         string
		backupLocationUID    string
		backupLocationName   string
		backupLocationMap    map[string]string
		labelSelectors       map[string]string
		providers            []string
	)

	JustBeforeEach(func() {
		StartTorpedoTest("KubevirtVMBackupRestoreWithDifferentStates", "Verify backup and restore of Kubevirt VMs in different states",
			nil, 93011)

		backupLocationMap = make(map[string]string)
		labelSelectors = make(map[string]string)
		providers = getProviders()

		log.InfoD("scheduling applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		for i := 0; i < 3; i++ {
			taskName := fmt.Sprintf("%d-%d", 93011, i)
			appContexts := ScheduleApplications(taskName)
			for _, appCtx := range appContexts {
				appCtx.ReadinessTimeout = appReadinessTimeout
				scheduledAppContexts = append(scheduledAppContexts, appCtx)
			}
		}
	})

	It("Basic Backup Creation", func() {
		defer func() {
			log.InfoD("switching to default context")
			err := SetClusterContext("")
			log.FailOnError(err, "failed to SetClusterContext to default cluster")
		}()

		Step("Validating applications", func() {
			log.InfoD("Validating applications")
			ValidateApplications(scheduledAppContexts)
		})

		Step("Creating backup location and cloud setting", func() {
			log.InfoD("Creating backup location and cloud setting")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, RandomString(6))
				backupLocationName = fmt.Sprintf("%s-%v", getGlobalBucketName(provider), RandomString(6))
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocationName
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, orgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, orgID, provider))
				err = CreateBackupLocation(provider, backupLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), orgID, "")
				dash.VerifyFatal(err, nil, "Creating backup location")
			}
		})

		Step("Registering cluster for backup", func() {
			log.InfoD("Registering cluster for backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			err = CreateApplicationClusters(orgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")

			clusterStatus, err := Inst().Backup.GetClusterStatus(orgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))

			sourceClusterUid, err = Inst().Backup.GetClusterUID(ctx, orgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))

			clusterStatus, err = Inst().Backup.GetClusterStatus(orgID, destinationClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", destinationClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", destinationClusterName))
		})

		Step("Changing the states of Kubevirt Virtual Machines", func() {
			namespaceToStopVMs := scheduledAppContexts[1].ScheduleOptions.Namespace
			err := StopAllVMsInNamespace(namespaceToStopVMs, true)
			log.FailOnError(err, "Failed stopping the VMs in namespace - "+namespaceToStopVMs)

		})

		Step("Taking backup of application from source cluster", func() {
			log.InfoD("Taking backup of applications")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			var wg sync.WaitGroup

			backupNames = make([]string, 0)
			for _, appCtx := range scheduledAppContexts {
				backupName := fmt.Sprintf("%s-%s-%v", "auto-backup", appCtx.ScheduleOptions.Namespace, RandomString(6))
				backupNames = append(backupNames, backupName)
				wg.Add(1)
				go func(backupName string, appCtx *scheduler.Context) {
					defer GinkgoRecover()
					defer wg.Done()
					log.InfoD("creating backup [%s] in source cluster [%s] (%s), organization [%s], of namespace [%s], in backup location [%s]", backupName, SourceClusterName, sourceClusterUid, orgID, appCtx.ScheduleOptions.Namespace, backupLocationName)
					err := CreateBackupWithValidation(ctx, backupName, SourceClusterName, backupLocationName, backupLocationUID, []*scheduler.Context{appCtx}, labelSelectors, orgID, sourceClusterUid, "", "", "", "")
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				}(backupName, appCtx)
			}
			wg.Wait()
		})

		Step("Restoring the backed up namespaces", func() {
			log.InfoD("Restoring the backed up namespaces")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			var wg sync.WaitGroup

			restoreNames = make([]string, 0)
			for i, appCtx := range scheduledAppContexts {
				restoreName := fmt.Sprintf("%s-%s-%v", "auto-restore", appCtx.ScheduleOptions.Namespace, RandomString(6))
				restoreNames = append(restoreNames, restoreName)
				wg.Add(1)
				go func(backupName string, appCtx *scheduler.Context, i int) {
					defer GinkgoRecover()
					defer wg.Done()
					log.InfoD("Restoring [%s] namespace from the [%s] backup", appCtx.ScheduleOptions.Namespace, backupNames[i])
					err = CreateRestoreWithValidation(ctx, restoreName, backupNames[i], make(map[string]string), make(map[string]string), destinationClusterName, orgID, []*scheduler.Context{appCtx})
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of restore [%s]", restoreName))
				}(restoreName, appCtx, i)
			}
			wg.Wait()
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
		dash.VerifySafely(err, nil, "Deleting backup schedule policies")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true

		log.Info("Destroying scheduled apps on source cluster")
		DestroyApps(scheduledAppContexts, opts)

		log.InfoD("switching to destination context")
		err = SetDestinationKubeConfig()
		log.FailOnError(err, "failed to switch to context to destination cluster")

		log.InfoD("Destroying restored apps on destination clusters")
		restoredAppContexts := make([]*scheduler.Context, 0)
		for _, scheduledAppContext := range scheduledAppContexts {
			restoredAppContext, err := CloneAppContextAndTransformWithMappings(scheduledAppContext, make(map[string]string), make(map[string]string), true)
			if err != nil {
				log.Errorf("TransformAppContextWithMappings: %v", err)
				continue
			}
			restoredAppContexts = append(restoredAppContexts, restoredAppContext)
		}
		DestroyApps(restoredAppContexts, opts)

		log.InfoD("switching to default context")
		err = SetClusterContext("")
		log.FailOnError(err, "failed to SetClusterContext to default cluster")

		backupDriver := Inst().Backup
		log.Info("Deleting backed up namespaces")
		for _, backupName := range backupNames {
			backupUID, err := backupDriver.GetBackupUID(ctx, backupName, orgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupName)
			backupDeleteResponse, err := DeleteBackup(backupName, backupUID, orgID, ctx)
			log.FailOnError(err, "Backup [%s] could not be deleted", backupName)
			dash.VerifyFatal(backupDeleteResponse.String(), "", fmt.Sprintf("Verifying [%s] backup deletion is successful", backupName))
		}
		log.Info("Deleting restored namespaces")
		for _, restoreName := range restoreNames {
			err = DeleteRestore(restoreName, orgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting Restore [%s]", restoreName))
		}
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})
