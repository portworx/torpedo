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
	"time"
)

// This testcase verifies backup and restore of Kubevirt VMs in different states like Running, Stopped, Restarting
var _ = Describe("{KubevirtVMBackupRestoreWithDifferentStates}", func() {

	var (
		backupNames                []string
		restoreNames               []string
		scheduledAppContexts       []*scheduler.Context
		sourceClusterUid           string
		cloudCredName              string
		cloudCredUID               string
		backupLocationUID          string
		backupLocationName         string
		backupLocationMap          map[string]string
		labelSelectors             map[string]string
		providers                  []string
		restoreNameToAppContextMap map[string]*scheduler.Context
		namespaceMappingMixed      map[string]string
		namespaceMappingRestart    map[string]string
		namespaceMappingStopped    map[string]string
		backupWithVMRestart        string
		restoreWithVMRestart       string
		namespaceWithStoppedVM     []string
		backupWithVMMixed          string
		restoreWithVMMixed         string
		backupWithVMStopped        string
		restoreWithVMStopped       string
	)

	JustBeforeEach(func() {
		StartTorpedoTest("KubevirtVMBackupRestoreWithDifferentStates", "Verify backup and restore of Kubevirt VMs in different states",
			nil, 93011)

		backupLocationMap = make(map[string]string)
		labelSelectors = make(map[string]string)
		restoreNameToAppContextMap = make(map[string]*scheduler.Context)
		namespaceMappingMixed = make(map[string]string)
		namespaceMappingRestart = make(map[string]string)
		namespaceMappingStopped = make(map[string]string)
		namespaceWithStoppedVM = make([]string, 0)
		providers = getProviders()

		log.InfoD("scheduling applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		for i := 0; i < 4; i++ {
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

		Step("Stopping VMs in a few namespaces", func() {
			log.InfoD("Stopping VMs in a few namespaces")
			namespaceToStopVMs := []string{scheduledAppContexts[0].ScheduleOptions.Namespace, scheduledAppContexts[1].ScheduleOptions.Namespace}
			namespaceWithStoppedVM = append(namespaceWithStoppedVM, namespaceToStopVMs...)
			log.InfoD("Stopping VMs in namespaces - [%v]", namespaceWithStoppedVM)
			for _, n := range namespaceWithStoppedVM {
				err := StopAllVMsInNamespace(n, true)
				log.FailOnError(err, "Failed stopping the VMs in namespace - "+n)
			}
		})

		Step("Taking individual backup of each namespace", func() {
			log.InfoD("Taking individual backup of each namespace")
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

		Step("Restoring all individual backups", func() {
			log.InfoD("Restoring all individual backups")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			var wg sync.WaitGroup
			restoreNames = make([]string, 0)
			for i, appCtx := range scheduledAppContexts {
				restoreName := fmt.Sprintf("%s-%s-%v", "auto-restore", appCtx.ScheduleOptions.Namespace, RandomString(6))
				restoreNames = append(restoreNames, restoreName)
				restoreNameToAppContextMap[restoreName] = appCtx
				wg.Add(1)
				go func(restoreName string, appCtx *scheduler.Context, i int) {
					defer GinkgoRecover()
					defer wg.Done()
					log.InfoD("Restoring [%s] namespace from the [%s] backup", appCtx.ScheduleOptions.Namespace, backupNames[i])
					err = CreateRestore(restoreName, backupNames[i], make(map[string]string), destinationClusterName, orgID, ctx, make(map[string]string))
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of restore [%s]", restoreName))
				}(restoreName, appCtx, i)
			}
			wg.Wait()
		})

		Step("Validating restores of individual backups", func() {
			log.InfoD("Validating restores of individual backups")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			defer func() {
				log.InfoD("Switching cluster context back to source cluster")
				err = SetSourceKubeConfig()
				log.FailOnError(err, "Switching context to source cluster failed")
			}()
			err = SetDestinationKubeConfig()
			log.FailOnError(err, "Switching context to destination cluster failed")
			var wg sync.WaitGroup
			for _, restoreName := range restoreNames {
				wg.Add(1)
				go func(restoreName string) {
					defer GinkgoRecover()
					defer wg.Done()
					log.InfoD("Validating restore [%s]", restoreName)
					expectedRestoredAppContext, err := CloneAppContextAndTransformWithMappings(restoreNameToAppContextMap[restoreName], make(map[string]string), make(map[string]string), true)
					log.FailOnError(err, "Failed TransformAppContextWithMappings")
					err = ValidateRestore(ctx, restoreName, orgID, []*scheduler.Context{expectedRestoredAppContext}, make([]string, 0))
					log.FailOnError(err, "Failed while validating restore - "+restoreName)
				}(restoreName)

			}
			wg.Wait()
		})

		Step("Take backup of all namespaces with VMs in Running and Stopped state", func() {
			log.InfoD("Take backup of all namespaces with VMs in Running and Stopped state")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			var namespaces []string
			for _, appCtx := range scheduledAppContexts {
				namespaces = append(namespaces, appCtx.ScheduleOptions.Namespace)
				namespaceMappingMixed[appCtx.ScheduleOptions.Namespace] = appCtx.ScheduleOptions.Namespace + "-mixed"
			}
			backupWithVMMixed = fmt.Sprintf("%s-%s", "auto-backup-mixed", RandomString(6))
			backupNames = append(backupNames, backupWithVMMixed)
			log.InfoD("creating backup [%s] in cluster [%s] (%s), organization [%s], of namespace [%v], in backup location [%s]", backupWithVMMixed, SourceClusterName, sourceClusterUid, orgID, namespaces, backupLocationName)
			err = CreateBackup(backupWithVMMixed, SourceClusterName, backupLocationName, backupLocationUID, namespaces,
				nil, orgID, sourceClusterUid, "", "", "", "", ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupWithVMMixed))
		})

		Step("Restoring backup taken when VMs were in Running and Stopped state", func() {
			log.InfoD("Restoring backup taken when VMs were in Running and Stopped state")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			restoreWithVMMixed = fmt.Sprintf("%s-%s", "auto-restore-mixed", RandomString(6))
			restoreNames = append(restoreNames, restoreWithVMMixed)
			log.InfoD("Restoring the [%s] backup", backupWithVMMixed)
			err = CreateRestoreWithValidation(ctx, restoreWithVMMixed, backupWithVMMixed, namespaceMappingMixed, make(map[string]string), destinationClusterName, orgID, scheduledAppContexts)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of restore %s from backup %s", restoreWithVMMixed, backupWithVMMixed))
		})

		Step("Starting the VMs in the namespace where it was stopped", func() {
			log.InfoD("Starting the VMs in the namespace where it was stopped")
			for _, n := range namespaceWithStoppedVM {
				log.InfoD("Starting the VMs in the namespace [%s] where it was stopped", n)
				err := StartAllVMsInNamespace(n, true)
				log.FailOnError(err, "Failed starting the VMs in namespace - "+n)
			}
		})

		Step("Take backup of all namespaces when VMs are restarting", func() {
			log.InfoD("Take backup of all namespaces when VMs are restarting")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			var wg sync.WaitGroup
			var namespaces []string
			for _, appCtx := range scheduledAppContexts {
				namespaces = append(namespaces, appCtx.ScheduleOptions.Namespace)
				namespaceMappingRestart[appCtx.ScheduleOptions.Namespace] = appCtx.ScheduleOptions.Namespace + "-restart"
			}
			backupWithVMRestart = fmt.Sprintf("%s-%s", "auto-backup-restart", RandomString(6))
			backupNames = append(backupNames, backupWithVMRestart)
			log.InfoD("creating backup [%s] in cluster [%s] (%s), organization [%s], of namespace [%v], in backup location [%s]", backupWithVMRestart, SourceClusterName, sourceClusterUid, orgID, namespaces, backupLocationName)
			_, err = CreateBackupWithoutCheck(ctx, backupWithVMRestart, SourceClusterName, backupLocationName, backupLocationUID, scheduledAppContexts, labelSelectors, orgID, sourceClusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupWithVMRestart))
			for _, n := range namespaces {
				wg.Add(1)
				go func(n string) {
					defer GinkgoRecover()
					defer wg.Done()
					err := RestartAllVMsInNamespace(n, true)
					log.FailOnError(err, "Failed restarting the VMs in namespace - "+n)
				}(n)
			}
			wg.Wait()
			err = backupSuccessCheck(backupWithVMRestart, orgID, maxWaitPeriodForBackupCompletionInMinutes*time.Minute, 30*time.Second, ctx)
			log.FailOnError(err, "Failed while checking success of backup [%s]", backupWithVMRestart)
		})

		Step("Restoring backup taken when VMs were Restarting", func() {
			log.InfoD("Restoring backup taken when VMs were Restarting")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			restoreWithVMRestart = fmt.Sprintf("%s-%s", "auto-restore-restart", RandomString(6))
			restoreNames = append(restoreNames, restoreWithVMRestart)
			log.InfoD("Restoring the [%s] backup", backupWithVMRestart)
			err = CreateRestoreWithValidation(ctx, restoreWithVMRestart, backupWithVMRestart, namespaceMappingRestart, make(map[string]string), destinationClusterName, orgID, scheduledAppContexts)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of restore %s from backup %s", restoreWithVMRestart, backupWithVMRestart))
		})

		Step("Stopping all the VMs in all the namespaces", func() {
			log.InfoD("Stopping all the VMs in all the namespaces")
			for _, appCtx := range scheduledAppContexts {
				err := StopAllVMsInNamespace(appCtx.ScheduleOptions.Namespace, true)
				log.FailOnError(err, "Failed stopping the VMs in namespace - "+appCtx.ScheduleOptions.Namespace)
			}
		})

		Step("Take backup of all namespaces with VMs in Stopped state", func() {
			log.InfoD("Take backup of all namespaces with VMs in Stopped state")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			var namespaces []string
			for _, appCtx := range scheduledAppContexts {
				namespaces = append(namespaces, appCtx.ScheduleOptions.Namespace)
				namespaceMappingStopped[appCtx.ScheduleOptions.Namespace] = appCtx.ScheduleOptions.Namespace + "-stopped"
			}
			backupWithVMStopped = fmt.Sprintf("%s-%s", "auto-backup-stopped", RandomString(6))
			backupNames = append(backupNames, backupWithVMStopped)
			log.InfoD("creating backup [%s] in cluster [%s] (%s), organization [%s], of namespace [%v], in backup location [%s]", backupWithVMStopped, SourceClusterName, sourceClusterUid, orgID, namespaces, backupLocationName)
			err = CreateBackup(backupWithVMStopped, SourceClusterName, backupLocationName, backupLocationUID, namespaces,
				nil, orgID, sourceClusterUid, "", "", "", "", ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s]", backupWithVMStopped))
		})

		Step("Restoring all backups taken when VMs were Stopped", func() {
			log.InfoD("Restoring all backups taken when VMs were Stopped")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			restoreWithVMStopped = fmt.Sprintf("%s-%s", "auto-restore-stopped", RandomString(6))
			restoreNames = append(restoreNames, restoreWithVMStopped)
			log.InfoD("Restoring the [%s] backup", backupWithVMStopped)
			err = CreateRestoreWithValidation(ctx, restoreWithVMStopped, backupWithVMStopped, namespaceMappingStopped, make(map[string]string), destinationClusterName, orgID, scheduledAppContexts)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of restore %s from backup %s", restoreWithVMStopped, backupWithVMStopped))
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
		for _, scheduledAppContext := range scheduledAppContexts {
			restoredAppContext, err := CloneAppContextAndTransformWithMappings(scheduledAppContext, namespaceMappingMixed, make(map[string]string), true)
			if err != nil {
				log.Errorf("TransformAppContextWithMappings: %v", err)
				continue
			}
			restoredAppContexts = append(restoredAppContexts, restoredAppContext)
		}
		for _, scheduledAppContext := range scheduledAppContexts {
			restoredAppContext, err := CloneAppContextAndTransformWithMappings(scheduledAppContext, namespaceMappingStopped, make(map[string]string), true)
			if err != nil {
				log.Errorf("TransformAppContextWithMappings: %v", err)
				continue
			}
			restoredAppContexts = append(restoredAppContexts, restoredAppContext)
		}
		for _, scheduledAppContext := range scheduledAppContexts {
			restoredAppContext, err := CloneAppContextAndTransformWithMappings(scheduledAppContext, namespaceMappingRestart, make(map[string]string), true)
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
