package tests

import (
	"fmt"
	"github.com/portworx/torpedo/drivers"
	"math/rand"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	"golang.org/x/sync/errgroup"
)

// MultipleBackupLocationWithSameEndpoint Create Backup and Restore for Multiple backup location added using same endpoint.
var _ = Describe("{MultipleBackupLocationWithSameEndpoint}", Label(TestCaseLabelsMap[MultipleBackupLocationWithSameEndpoint]...), func() {
	var (
		scheduledAppContexts          []*scheduler.Context
		backupLocationNameMap         = make(map[int]string)
		backupLocationUIDMap          = make(map[int]string)
		backupLocationMap             = make(map[string]string)
		restoreNsMapping              = make(map[string]map[string]string)
		bkpNamespaces                 []string
		cloudCredName                 string
		cloudCredUID                  string
		clusterUid                    string
		labelSelectors                map[string]string
		wg                            sync.WaitGroup
		userBackupMap                 = make(map[int]map[string]string)
		restoreNames                  []string
		numberOfBackupLocation        = 1000
		numberOfBackups               = 30
		providers                     = GetBackupProviders()
		timeBetweenConsecutiveBackups = 10 * time.Second
		controlChannel                chan string
		errorGroup                    *errgroup.Group
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("MultipleBackupLocationWithSameEndpoint", "Create Backup and Restore for Multiple backup location added using same endpoint", nil, 84902, Ak, Q3FY24)
		log.InfoD("scheduling applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, appCtx := range appContexts {
				appCtx.ReadinessTimeout = AppReadinessTimeout
				scheduledAppContexts = append(scheduledAppContexts, appCtx)
				namespace := GetAppNamespace(appCtx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
			}
		}
	})

	It("Create Backup and Restore for Multiple backup location added using same endpoint", func() {
		Step("Validate applications", func() {
			ctx, _ := backup.GetAdminCtxFromSecret()
			controlChannel, errorGroup = ValidateApplicationsStartData(scheduledAppContexts, ctx)
		})
		Step(fmt.Sprintf("Creating a cloud credentials from px-admin"), func() {
			log.InfoD(fmt.Sprintf("Creating a cloud credentials from px-admin"))
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
				cloudCredUID = uuid.New()
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
			}
		})
		Step(fmt.Sprintf("Creating [%d] backup locations from px-admin", numberOfBackupLocation), func() {
			log.InfoD(fmt.Sprintf("Creating [%d] backup locations from px-admin", numberOfBackupLocation))
			for i := 0; i <= numberOfBackupLocation; i++ {
				for _, provider := range providers {
					log.InfoD(fmt.Sprintf("Creating backup locations with index [%d]", i))
					backupLocationNameMap[i] = fmt.Sprintf("%s-%d-%s", getGlobalBucketName(provider), i, RandomString(6))
					backupLocationUIDMap[i] = uuid.New()
					err := CreateBackupLocation(provider, backupLocationNameMap[i], backupLocationUIDMap[i], cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of backup location [%s]", backupLocationNameMap[i]))
					backupLocationMap[backupLocationUIDMap[i]] = backupLocationNameMap[i]
				}
			}
		})
		Step("Registering cluster for backup from px-admin", func() {
			log.InfoD("Registering cluster for backup from px-admin")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			log.InfoD("Verifying cluster status for both source and destination clusters")
			clusterStatus, err := Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, DestinationClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", DestinationClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", DestinationClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})
		Step(fmt.Sprintf("Taking [%d] backup for the each application from px-admin", numberOfBackups), func() {
			log.InfoD(fmt.Sprintf("Taking [%d] backup for the each application from px-admin", numberOfBackups))
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "failed to fetch ctx for admin")
			createBackup := func(backupName string, namespace string, index int) {
				defer GinkgoRecover()
				defer wg.Done()
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
				err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, backupLocationNameMap[index], backupLocationUIDMap[index], appContextsToBackup, labelSelectors, BackupOrgID, clusterUid, "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation and validation of backup [%s] of namespace (scheduled Context) [%s]", backupName, namespace))
			}
			semaphore := make(chan int, 4)
			for _, namespace := range bkpNamespaces {
				for index := 0; index < numberOfBackups; index++ {
					time.Sleep(timeBetweenConsecutiveBackups)
					backupName := fmt.Sprintf("%s-%s-%s", BackupNamePrefix, backupLocationNameMap[index], RandomString(4))
					userBackupMap[index] = make(map[string]string)
					userBackupMap[index][backupName] = namespace
					wg.Add(1)
					semaphore <- 0
					go func(backupName string, namespace string, index int) {
						defer func() {
							<-semaphore
						}()
						createBackup(backupName, namespace, index)
					}(backupName, namespace, index)
				}
			}
			wg.Wait()
		})

		Step("Taking restore for each backups created from px-admin", func() {
			log.InfoD(fmt.Sprintf("Taking restore for each backups created from px-admin"))
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			var wg sync.WaitGroup
			var mu sync.Mutex
			errors := make([]string, 0)
			for index := 0; index < numberOfBackups; index++ {
				for backupName, namespace := range userBackupMap[index] {
					wg.Add(1)
					go func(backupName, namespace string) {
						defer GinkgoRecover()
						defer wg.Done()
						mu.Lock()
						restoreName := fmt.Sprintf("%s-%s-%s", RestoreNamePrefix, backupName, RandomString(5))
						customNamespace := "custom-" + namespace + RandomString(5)
						namespaceMapping := map[string]string{namespace: customNamespace}
						restoreNsMapping[restoreName] = namespaceMapping
						mu.Unlock()
						err := CreateRestore(restoreName, backupName, namespaceMapping, SourceClusterName, BackupOrgID, ctx, make(map[string]string))
						if err != nil {
							mu.Lock()
							errors = append(errors, fmt.Sprintf("Failed while taking restore [%s]. Error - [%s]", restoreName, err.Error()))
							mu.Unlock()
						}
					}(backupName, namespace)
				}
			}
			wg.Wait()
			dash.VerifyFatal(len(errors), 0, fmt.Sprintf("Creating restores : -\n%s", strings.Join(errors, "}\n{")))
			log.InfoD("All  mapping list %v", restoreNsMapping)

		})

		Step("Validating all restores", func() {
			log.InfoD("Validating all restores")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			var mutex sync.Mutex
			errors := make([]string, 0)
			var wg sync.WaitGroup
			for restoreName, namespaceMapping := range restoreNsMapping {
				wg.Add(1)
				go func(restoreName string, namespaceMapping map[string]string) {
					defer GinkgoRecover()
					defer wg.Done()
					log.InfoD("Validating restore [%s] with namespace mapping", restoreName)
					expectedRestoredAppContext, _ := CloneAppContextAndTransformWithMappings(scheduledAppContexts[0], namespaceMapping, make(map[string]string), true)
					if err != nil {
						mutex.Lock()
						errors = append(errors, fmt.Sprintf("Failed while context tranforming of restore [%s]. Error - [%s]", restoreName, err.Error()))
						mutex.Unlock()
						return
					}
					err = ValidateRestore(ctx, restoreName, BackupOrgID, []*scheduler.Context{expectedRestoredAppContext}, make([]string, 0))
					if err != nil {
						mutex.Lock()
						errors = append(errors, fmt.Sprintf("Failed while validating restore [%s]. Error - [%s]", restoreName, err.Error()))
						mutex.Unlock()
					}
				}(restoreName, namespaceMapping)
			}
			wg.Wait()
			dash.VerifyFatal(len(errors), 0, fmt.Sprintf("Validating restores of individual backups -\n%s", strings.Join(errors, "}\n{")))

		})
		Step("Delete all Backup locations from px-admin", func() {
			log.InfoD("Delete Backup locations from px-admin")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "failed to fetch ctx for admin")
			for backupLocationUID, backupLocationName := range backupLocationMap {
				wg.Add(1)
				go func(backupLocationName, backupLocationUID string) {
					defer GinkgoRecover()
					defer wg.Done()
					err := DeleteBackupLocationWithContext(backupLocationName, backupLocationUID, BackupOrgID, true, ctx)
					Inst().Dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying deletion of backup location [%s]", backupLocationName))
				}(backupLocationName, backupLocationUID)
			}
			wg.Wait()
		})
		Step("Wait for Backup location deletion", func() {
			log.InfoD("Wait for Backup location deletion")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "failed to fetch ctx for admin")
			AllBackupLocationMap, err := GetAllBackupLocations(ctx)
			log.FailOnError(err, "Fetching all backup locations")
			for backupLocationUID, backupLocationName := range AllBackupLocationMap {
				wg.Add(1)
				go func(backupLocationName, backupLocationUID string) {
					defer GinkgoRecover()
					defer wg.Done()
					err := Inst().Backup.WaitForBackupLocationDeletion(ctx, backupLocationName, backupLocationUID, BackupOrgID, BackupLocationDeleteTimeout, BackupLocationDeleteRetryTime)
					Inst().Dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying waiting for backup location [%s] deletion", backupLocationName))
				}(backupLocationName, backupLocationUID)
			}
			wg.Wait()
		})
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		log.InfoD("Deleting the restores")
		for _, restoreName := range restoreNames {
			wg.Add(1)
			go func(restoreName string) {
				defer GinkgoRecover()
				defer wg.Done()
				err = DeleteRestore(restoreName, BackupOrgID, ctx)
				dash.VerifySafely(err, nil, fmt.Sprintf("Deleting restore [%s]", restoreName))
			}(restoreName)
		}
		wg.Wait()
		backupNames, err := GetAllBackupsAdmin()
		dash.VerifySafely(err, nil, fmt.Sprintf("Fetching all backups for admin"))
		for _, backupName := range backupNames {
			wg.Add(1)
			go func(backupName string) {
				defer GinkgoRecover()
				defer wg.Done()
				backupUid, err := Inst().Backup.GetBackupUID(ctx, backupName, BackupOrgID)
				_, err = DeleteBackup(backupName, backupUid, BackupOrgID, ctx)
				dash.VerifySafely(err, nil, fmt.Sprintf("Delete the backup %s ", backupName))
				err = DeleteBackupAndWait(backupName, ctx)
				dash.VerifySafely(err, nil, fmt.Sprintf("waiting for backup [%s] deletion", backupName))
			}(backupName)
		}
		wg.Wait()
		log.InfoD("Deleting the deployed apps after the testcase")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		err = DestroyAppsWithData(scheduledAppContexts, opts, controlChannel, errorGroup)
		log.FailOnError(err, "Data validations failed")
		log.InfoD("Deleting the px-backup objects")
		backupLocationMap, err := GetAllBackupLocations(ctx)
		log.FailOnError(err, "Fetching all backup locations")
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})

// This TC takes backup of 50 volumes and performs restore
var _ = Describe("{ValidateFiftyVolumeBackups}", Label(TestCaseLabelsMap[ValidateFiftyVolumeBackups]...), func() {
	var (
		scheduledAppContexts []*scheduler.Context
		sourceClusterUid     string
		backupLocationMap    map[string]string
		cloudAccountName     string
		bkpLocationName      string
		cloudCredUID         string
		backupLocationUID    string
		currentBackupName    string
		namespace            string
		backupNameList       []string
		restoreNames         []string
		preRuleName          string
		postRuleName         string
		preRuleUid           string
		postRuleUid          string
		providers            = GetBackupProviders()
		numberOfVolumes      = 50
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("ValidateFiftyVolumeBackups", "To verify backup of 50 volumes and performs restore", nil, 55816, Sabrarhussaini, Q1FY25)
		backupLocationMap = make(map[string]string)
		log.InfoD("scheduling applications")
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		namespace = fmt.Sprintf("multiple-app-ns-%s", RandomString(6))
		Inst().AppList = []string{"postgres-backup-multivol"}
		Inst().CustomAppConfig["postgres-backup-multivol"] = scheduler.AppConfig{
			ClaimsCount: numberOfVolumes,
		}
		err := Inst().S.RescanSpecs(Inst().SpecDir, Inst().V.String())
		log.FailOnError(err, "Failed to rescan specs from %s for storage provider %s", Inst().SpecDir, Inst().V.String())
		appContexts := ScheduleApplicationsOnNamespace(namespace, TaskNamePrefix)
		for _, appCtx := range appContexts {
			appCtx.ReadinessTimeout = AppReadinessTimeout
			scheduledAppContexts = append(scheduledAppContexts, appCtx)
		}
	})

	It("To verify backup of 50 volumes and performs restore", func() {
		Step("Validate applications", func() {
			log.InfoD("Validating applications")
			ValidateApplications(scheduledAppContexts)
		})

		Step("Validate creation of cloud credentials and backup location", func() {
			log.InfoD("Validate creation of cloud credentials and backup location")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, provider := range providers {
				cloudCredUID = uuid.New()
				cloudAccountName = fmt.Sprintf("%s-%s-%v", CredName, provider, RandomString(4))
				log.InfoD("Creating cloud credential named [%s] and uid [%s] using [%s] as provider", cloudAccountName, cloudCredUID, provider)
				err := CreateCloudCredential(provider, cloudAccountName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudAccountName, BackupOrgID, provider))
				bkpLocationName = fmt.Sprintf("%s-%s-%v", provider, getGlobalBucketName(provider), RandomString(4))
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = bkpLocationName
				bucketName := getGlobalBucketName(provider)
				err = CreateBackupLocation(provider, bkpLocationName, backupLocationUID, cloudAccountName, cloudCredUID, bucketName, BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of backup location named [%s] with uid [%s] of [%s] as provider", bkpLocationName, backupLocationUID, provider))
			}
		})

		Step(fmt.Sprintf("Verify creation of pre and post exec rules for applications "), func() {
			log.InfoD("Verify creation of pre and post exec rules for applications ")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			preRuleName, postRuleName, err = CreateRuleForBackupWithMultipleApplications(BackupOrgID, Inst().AppList, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of pre and post exec rules for applications from px-admin"))
			if preRuleName != "" {
				preRuleUid, err = Inst().Backup.GetRuleUid(BackupOrgID, ctx, preRuleName)
				log.FailOnError(err, "Fetching pre backup rule [%s] uid", preRuleName)
				log.Infof("Pre backup rule [%s] uid: [%s]", preRuleName, preRuleUid)
			}
			if postRuleName != "" {
				postRuleUid, err = Inst().Backup.GetRuleUid(BackupOrgID, ctx, postRuleName)
				log.FailOnError(err, "Fetching post backup rule [%s] uid", postRuleName)
				log.Infof("Post backup rule [%s] uid: [%s]", postRuleName, postRuleUid)
			}
		})

		Step("Adding Clusters for backup", func() {
			log.InfoD("Adding Clusters for backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of creating source - %s and destination - %s clusters", SourceClusterName, DestinationClusterName))
			clusterStatus, err := Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			sourceClusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})

		Step("Taking backup of application with 50 volumes on source cluster", func() {
			log.InfoD("Taking backup of application with 50 volumes on source cluster")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			log.InfoD("Taking Backup of application")
			currentBackupName = fmt.Sprintf("%s-%v", BackupNamePrefix, RandomString(10))
			labelSelectors := make(map[string]string)
			err = CreateBackupWithValidation(ctx, currentBackupName, SourceClusterName, bkpLocationName, backupLocationUID, scheduledAppContexts, labelSelectors, BackupOrgID, sourceClusterUid, preRuleName, preRuleUid, postRuleName, postRuleUid)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", currentBackupName))
			backupNameList = append(backupNameList, currentBackupName)
		})

		Step("Restoring backup with 50 volumes on destination cluster", func() {
			log.InfoD("Restoring backup with 50 volumes on destination cluster")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Unable to fetch px-central-admin ctx")
			log.Infof("Backup to be restored - %v", currentBackupName)
			restoreName := fmt.Sprintf("%s-%v", RestoreNamePrefix, RandomString(10))
			err = CreateRestoreWithValidation(ctx, restoreName, currentBackupName, make(map[string]string), make(map[string]string), DestinationClusterName, BackupOrgID, scheduledAppContexts)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s] from backup [%s]", restoreName, currentBackupName))
			restoreNames = append(restoreNames, restoreName)
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		defer func() {
			err := SetSourceKubeConfig()
			log.FailOnError(err, "Unable to switch context to source cluster [%s]", SourceClusterName)
		}()
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		log.InfoD("Deleting the restores")
		for _, restoreName := range restoreNames {
			err = DeleteRestore(restoreName, BackupOrgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting restore [%s]", restoreName))
		}
		log.InfoD("Deleting the deployed apps after the testcase")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		for _, appCntxt := range scheduledAppContexts {
			appCntxt.SkipVolumeValidation = true
		}
		DestroyApps(scheduledAppContexts, opts)
		log.InfoD("Deleting the px-backup objects")
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudAccountName, cloudCredUID, ctx)
		log.InfoD("Switching context to destination cluster for clean up")
		err = SetDestinationKubeConfig()
		log.FailOnError(err, "Unable to switch context to destination cluster [%s]", DestinationClusterName)
		DestroyApps(scheduledAppContexts, opts)
		log.InfoD("Switching back context to Source cluster")
		err = SetSourceKubeConfig()
		log.FailOnError(err, "Unable to switch context to source cluster [%s]", SourceClusterName)
	})
})

var _ = Describe("{ClusterSharingWithConcurrentBackupOperations}", func() {
	var (
		appNamespaces             []string
		backupAppContexts         []*scheduler.Context
		namespaceAppContextMap    = make(map[string][]*scheduler.Context)
		cloudAccountName          string
		cloudAccountUid           string
		backupLocationName        string
		backupLocationUid         string
		backupLocationMap         map[string]string
		backupResults             = make(map[string][]string)
		newBackupResults          = make(map[string][]string)
		srcClusterUid             string
		destClusterUid            string
		newBackupMap              = make(map[string][]string)
		newestBackupMap           = make(map[string][]string)
		newBackupNameList         []string
		newestBackupNameList      []string
		numDeployments            = 10
		numOfBackupsPerDeployment = 2
		snapshotLimit             = 4
		numOfPrimaryUsers         = 10
		primaryUserList           []string
		numOfSharedUsers          = 10
		sharedUserList            []string
		userRoleMap               = make(map[string]backup.PxBackupRole)
		numberOfGroups            = 10
		groupList                 = make([]string, 0)
		splitIndexForUsers        = 5
		splitIndexForGroups       = 5
		firstUserList             []string
		secondUserList            []string
		firstGroupList            []string
		secondGroupList           []string
		restoreNames              []string
		numOfRestores             = 5
		firstRandomUser           string
		secondRandomUser          string
		backupDriver              = Inst().Backup
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("ClusterSharingWithConcurrentBackupOperations", "TC to verify Concurrent Backup and Restore Operations with Cluster Sharing", nil, 0, Sabrarhussaini, Q2FY25)
		log.Infof("Scheduling applications")
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		Inst().AppList = []string{"postgres-backup"}
		backupAppContexts = make([]*scheduler.Context, 0)
		appNamespaces = make([]string, 0)
		err := SetSourceKubeConfig()
		log.FailOnError(err, "Switching context to source cluster failed")
		log.Infof("Scheduling applications")
		for i := 0; i < numDeployments; i++ {
			taskName := fmt.Sprintf("multiple-%d", i)
			appContexts := ScheduleApplications(taskName)
			for _, appCtx := range appContexts {
				namespace := GetAppNamespace(appCtx, taskName)
				appNamespaces = append(appNamespaces, namespace)
				backupAppContexts = append(backupAppContexts, appCtx)
				appCtx.ReadinessTimeout = AppReadinessTimeout
				namespaceAppContextMap[namespace] = append(namespaceAppContextMap[namespace], appCtx)
			}
		}
	})

	It("TC to verify Concurrent Backup and Restore Operations with Cluster Sharing", func() {
		Step("Validating applications ", func() {
			log.InfoD("Validating applications")
			ValidateApplications(backupAppContexts)
		})

		Step("Create a set of primary users with different roles", func() {
			log.Infof("Creating %d primary users with different roles", numOfPrimaryUsers)
			primaryUserList = CreateUsers(numOfPrimaryUsers)
			roles := []backup.PxBackupRole{
				//backup.SuperAdmin,
				backup.ApplicationUser,
				backup.ApplicationOwner,
				backup.InfrastructureOwner,
			}
			for i, user := range primaryUserList {
				role := roles[i%len(roles)]
				err := backup.AddRoleToUser(user, role, fmt.Sprintf("Adding %v role to %s", role, user))
				log.FailOnError(err, "failed to add role %s to the user %s", role, user)
				userRoleMap[user] = role
			}
			for user, role := range userRoleMap {
				log.Infof("User %s has been assigned role %v", user, role)
			}
		})

		Step("Create a set of users to share clusters with", func() {
			log.Infof("Creating %d secondary users to share the clusters with", numOfSharedUsers)
			sharedUserList = CreateUsers(numOfSharedUsers)
			roles := []backup.PxBackupRole{
				//backup.SuperAdmin,
				backup.ApplicationUser,
				backup.ApplicationOwner,
				backup.InfrastructureOwner,
			}
			for i, user := range sharedUserList {
				role := roles[i%len(roles)]
				err := backup.AddRoleToUser(user, role, fmt.Sprintf("Adding %v role to %s", role, user))
				log.FailOnError(err, "failed to add role %s to the user %s", role, user)
			}
			firstUserList = sharedUserList[:splitIndexForUsers]
			secondUserList = sharedUserList[splitIndexForUsers:]
		})

		Step("Create Groups to share clusters with", func() {
			log.InfoD("Creating %d groups to share the clusters with", numberOfGroups)
			var wg sync.WaitGroup
			var mutex sync.Mutex
			for i := 1; i <= numberOfGroups; i++ {
				groupName := fmt.Sprintf("testGroup%v", i)
				wg.Add(1)
				go func(groupName string) {
					defer GinkgoRecover()
					defer wg.Done()
					err := backup.AddGroup(groupName)
					log.FailOnError(err, "Failed to create group - %v", groupName)
					mutex.Lock()
					groupList = append(groupList, groupName)
					mutex.Unlock()
				}(groupName)
			}
			wg.Wait()
			firstGroupList = groupList[:splitIndexForGroups]
			secondGroupList = groupList[splitIndexForGroups:]
		})

		Step(fmt.Sprintf("Adding Credentials and BackupLocation from px-admin user and making it public"), func() {
			log.InfoD(fmt.Sprintf("Adding Credentials and BackupLocation from px-admin user and making it public"))
			providers := GetBackupProviders()
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			for _, provider := range providers {
				cloudAccountUid = uuid.New()
				cloudAccountName = fmt.Sprintf("autogenerated-cred-%v", RandomString(5))
				if provider != drivers.ProviderNfs {
					err = CreateCloudCredential(provider, cloudAccountName, cloudAccountUid, BackupOrgID, ctx)
					log.FailOnError(err, "Failed to create cloud credential - %s", err)
					err = AddCloudCredentialOwnership(cloudAccountName, cloudAccountUid, nil, nil, Invalid, Read, ctx, BackupOrgID)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying public ownership update for cloud credential %s ", cloudAccountName))
				}
				backupLocationName = fmt.Sprintf("autogenerated-backup-location-%v", RandomString(5))
				backupLocationUid = uuid.New()
				err = CreateBackupLocationWithContext(provider, backupLocationName, backupLocationUid, cloudAccountName, cloudAccountUid, getGlobalBucketName(provider), BackupOrgID, "", ctx, true)
				log.FailOnError(err, "Failed to add backup location %s using provider %s for px-admin user", backupLocationName, provider)
				err = AddBackupLocationOwnership(backupLocationName, backupLocationUid, nil, nil, Invalid, Read, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying public ownership update for backup location %s", backupLocationName))
			}
		})

		Step("Create source and destination clusters for all primary users", func() {
			log.InfoD("Creating source and destination clusters for all primary users")
			var wg sync.WaitGroup
			for _, user := range primaryUserList {
				wg.Add(1)
				go func(user string) {
					defer wg.Done()
					nonAdminCtx, err := backup.GetNonAdminCtx(user, CommonPassword)
					if err != nil {
						log.Errorf("Failed to fetch user %s ctx: %v", user, err)
						return
					}
					log.Infof("Creating source [%s] and destination [%s] clusters for user [%s]", SourceClusterName, DestinationClusterName, user)
					err = CreateApplicationClusters(BackupOrgID, "", "", nonAdminCtx)
					if err != nil {
						log.Errorf("Failed to create source [%s] and destination [%s] clusters with user [%s] ctx: %v", SourceClusterName, DestinationClusterName, user, err)
						return
					}
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of source [%s] and destination [%s] clusters with user [%s] ctx", SourceClusterName, DestinationClusterName, user))
					srcClusterStatus, err := Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, nonAdminCtx)
					if err != nil {
						log.Errorf("Failed to fetch [%s] cluster status for user [%s]: %v", SourceClusterName, user, err)
						return
					}
					dash.VerifyFatal(srcClusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online for user [%s]", SourceClusterName, user))
					srcClusterUid, err = Inst().Backup.GetClusterUID(nonAdminCtx, BackupOrgID, SourceClusterName)
					if err != nil {
						log.Errorf("Failed to fetch [%s] cluster UID for user [%s]: %v", SourceClusterName, user, err)
						return
					}
					log.Infof("User [%s]: Cluster [%s] UID: [%s]", user, SourceClusterName, srcClusterUid)
					dstClusterStatus, err := Inst().Backup.GetClusterStatus(BackupOrgID, DestinationClusterName, nonAdminCtx)
					if err != nil {
						log.Errorf("Failed to fetch [%s] cluster status for user [%s]: %v", DestinationClusterName, user, err)
						return
					}
					dash.VerifyFatal(dstClusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online for user [%s]", DestinationClusterName, user))
					destClusterUid, err = Inst().Backup.GetClusterUID(nonAdminCtx, BackupOrgID, DestinationClusterName)
					if err != nil {
						log.Errorf("Failed to fetch [%s] cluster UID for user [%s]: %v", DestinationClusterName, user, err)
						return
					}
					log.Infof("User [%s]: Cluster [%s] UID: [%s]", user, DestinationClusterName, destClusterUid)
				}(user)
			}
			wg.Wait()
		})

		Step("Taking multiple backups for all primary users", func() {
			log.InfoD("Taking backup for all primary users")
			var wg sync.WaitGroup
			var mu sync.Mutex

			for _, user := range primaryUserList {
				wg.Add(1)
				go func(user string) {
					defer wg.Done()
					ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
					log.FailOnError(err, "Failed to fetch context for user")
					backupNameList, err := TakeMultipleBackupsPerDeployment(ctx, BackupOrgID, SourceClusterName, numOfBackupsPerDeployment, snapshotLimit, backupLocationName, backupLocationUid, backupAppContexts, BackupNamePrefix)
					if err != nil {
						log.Errorf("Failed to take backups for user %s: %v", user, err)
						return
					}
					mu.Lock()
					backupResults[user] = backupNameList
					mu.Unlock()
				}(user)
			}
			wg.Wait()
			for user, backupNameList := range backupResults {
				log.Infof("User %s: All backups taken: %v", user, backupNameList)
			}
		})

		Step("Initiate additional backups and restores for all primary users for validation", func() {
			log.InfoD("Initiating additional backups and restores for all primary users validation")
			var wg sync.WaitGroup

			for _, user := range primaryUserList {
				wg.Add(1)
				go func(user string) {
					defer wg.Done()
					defer GinkgoRecover()

					ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
					if err != nil {
						log.Errorf("Failed to fetch context for user %s: %v", user, err)
						return
					}

					for _, scheduledAppContext := range backupAppContexts {
						var innerWg sync.WaitGroup
						semaphore := make(chan struct{}, snapshotLimit)

						for i := 0; i < numOfBackupsPerDeployment; i++ {
							innerWg.Add(1)
							go func(i int, scheduledAppContext *scheduler.Context) {
								defer innerWg.Done()
								defer GinkgoRecover()

								semaphore <- struct{}{}
								defer func() { <-semaphore }()
								currentBackupName := fmt.Sprintf("%s-%s-%d", BackupNamePrefix, RandomString(8), i+1)
								_, err := CreateBackupByNamespacesWithoutCheck(currentBackupName, SourceClusterName, backupLocationName, backupLocationUid, []string{scheduledAppContext.ScheduleOptions.Namespace}, make(map[string]string), BackupOrgID, srcClusterUid, "", "", "", "", ctx)
								if err != nil {
									log.Errorf("Failed to create backup %s: %v", currentBackupName, err)
									return
								}
								log.Infof("Backup %s triggered for user %s", currentBackupName, user)
								newBackupNameList = append(newBackupNameList, currentBackupName)
							}(i, scheduledAppContext)
						}
						newBackupMap[user] = newBackupNameList
						innerWg.Wait()
					}

					// Initiate restore for the backup
					backups, exists := backupResults[user]
					if !exists {
						log.Errorf("No backups found for user %s", user)
						return
					}
					for i, backupName := range backups {
						if i >= numOfRestores {
							break
						}
						restoreName := fmt.Sprintf("%s-restore-%d", user, i+1)
						_, err := CreateRestoreWithoutCheck(restoreName, backupName, make(map[string]string), DestinationClusterName, BackupOrgID, ctx)
						log.FailOnError(err, "Failed to create restore %s for user %s: %v", restoreName, user)
						log.Infof("Successfully created restore %s for user %s", restoreName, user)

						restoreNames = append(restoreNames, restoreName)
					}
				}(user)
			}
			log.Info("All additional backups and restores have been initiated for all primary users.")
		})

		Step("Share the clusters from each user with other users and groups", func() {
			log.InfoD("Sharing the clusters from each user with other users and groups")
			var wg sync.WaitGroup

			for _, user := range primaryUserList {
				wg.Add(1)
				go func(user string) {
					defer wg.Done()
					ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
					log.FailOnError(err, "Failed to fetch context for user")
					clusterUid, err := Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
					_, err = ShareClusterWithValidation(ctx, SourceClusterName, clusterUid, sharedUserList, groupList, true)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying share of source [%s] cluster for user %s", SourceClusterName, user))
				}(user)
			}
			wg.Wait()
		})

		Step("Verify the shared backups for a random user", func() {
			log.InfoD("Verifying the shared backups for the user")
			firstUserIndex := rand.Intn(len(sharedUserList))
			firstRandomUser = sharedUserList[firstUserIndex]
			fmt.Printf("Randomly selected user: %s\n", firstRandomUser)
			for {
				secondUserIndex := rand.Intn(len(sharedUserList))
				if secondUserIndex != firstUserIndex {
					secondRandomUser = sharedUserList[secondUserIndex]
					break
				}
			}
			ctx, err := backup.GetNonAdminCtx(firstRandomUser, CommonPassword)
			log.FailOnError(err, "Fetching user ctx")
			backups, exists := newBackupMap[primaryUserList[0]]
			if !exists {
				log.Errorf("No backups found for user %s", primaryUserList[0])
			}
			for _, backupName := range backups {
				err = BackupSuccessCheck(backupName, BackupOrgID, MaxWaitPeriodForBackupCompletionInMinutes*time.Minute, 30*time.Second, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying backup [%s] success check", backupName))
				log.Infof("Backup [%s] created successfully", backupName)
			}
			// Validate restore for one of the backup
			backupUID, err := Inst().Backup.GetBackupUID(ctx, backups[0], BackupOrgID)
			log.FailOnError(err, fmt.Sprintf("Getting UID for backup %v", backups[0]))
			backupInspectRequest := &api.BackupInspectRequest{
				Name:  backups[0],
				Uid:   backupUID,
				OrgId: BackupOrgID,
			}
			resp, err := backupDriver.InspectBackup(ctx, backupInspectRequest)
			log.FailOnError(err, "Inspect each backup from list")
			namespaces := resp.GetBackup().GetNamespaces()
			var collectedAppContexts []*scheduler.Context
			for _, namespace := range namespaces {
				if appContexts, exists := namespaceAppContextMap[namespace]; exists {
					collectedAppContexts = append(collectedAppContexts, appContexts...)
				} else {
					fmt.Printf("No app contexts found for namespace: %s\n", namespace)
				}
			}
			restoreName := fmt.Sprintf("%s-restore", firstRandomUser)
			// double-check the implementation
			err = CreateRestoreWithValidation(ctx, restoreName, backups[0], make(map[string]string), make(map[string]string), DestinationClusterName, BackupOrgID, collectedAppContexts)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Failed to restore backup [%s", restoreName))
			//Validate backup deletion
			var wg sync.WaitGroup
			for _, backupName := range backups {
				wg.Add(1)
				go func(backupName string) {
					defer GinkgoRecover()
					defer wg.Done()
					backupUid, err := Inst().Backup.GetBackupUID(ctx, backupName, BackupOrgID)
					log.FailOnError(err, "Failed to fetch the backup %s uid of the user %s", backupName, firstRandomUser)
					_, err = DeleteBackup(backupName, backupUid, BackupOrgID, ctx)
					log.FailOnError(err, "Failed to delete the backup %s of the user %s", backupName, firstRandomUser)
				}(backupName)
			}
			wg.Wait()
			// validate the backup deletion from second random user
			ctx, err = backup.GetNonAdminCtx(secondRandomUser, CommonPassword)
			log.FailOnError(err, "Fetching user ctx")
			for _, backupName := range backups {
				wg.Add(1)
				go func(backupName string) {
					defer GinkgoRecover()
					defer wg.Done()
					err = DeleteBackupAndWait(backupName, ctx)
					log.FailOnError(err, fmt.Sprintf("waiting for backup [%s] deletion", backupName))
				}(backupName)
			}
			wg.Wait()
		})

		Step("Initiate additional backups and restores for all primary users for validation", func() {
			log.InfoD("Initiating additional backups and restores for all primary users validation")
			var wg sync.WaitGroup

			for _, user := range primaryUserList {
				wg.Add(1)
				go func(user string) {
					defer wg.Done()
					defer GinkgoRecover()

					ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
					log.FailOnError(err, "Failed to fetch context for user")

					for _, scheduledAppContext := range backupAppContexts {
						var innerWg sync.WaitGroup
						semaphore := make(chan struct{}, snapshotLimit)

						for i := 0; i < numOfBackupsPerDeployment; i++ {
							innerWg.Add(1)
							go func(i int, scheduledAppContext *scheduler.Context) {
								defer innerWg.Done()
								defer GinkgoRecover()

								semaphore <- struct{}{}
								defer func() { <-semaphore }()
								currentBackupName := fmt.Sprintf("%s-%s-%d", BackupNamePrefix, RandomString(8), i+1)
								_, err := CreateBackupByNamespacesWithoutCheck(currentBackupName, SourceClusterName, backupLocationName, backupLocationUid, []string{scheduledAppContext.ScheduleOptions.Namespace}, make(map[string]string), BackupOrgID, srcClusterUid, "", "", "", "", ctx)
								if err != nil {
									log.Errorf("Failed to create backup %s: %v", currentBackupName, err)
									return
								}
								log.Infof("Backup %s triggered for user %s", currentBackupName, user)
								newestBackupNameList = append(newestBackupNameList, currentBackupName)
							}(i, scheduledAppContext)
						}
						newestBackupMap[user] = newestBackupNameList
						innerWg.Wait()
					}

					// Initiate restore for the backup
					backups, exists := backupResults[user]
					if !exists {
						log.Errorf("No backups found for user %s", user)
						return
					}
					for i, backupName := range backups {
						if i >= numOfRestores {
							break
						}
						restoreName := fmt.Sprintf("%s-restore-%d", user, i+1)
						_, err := CreateRestoreWithoutCheck(restoreName, backupName, make(map[string]string), DestinationClusterName, BackupOrgID, ctx)
						if err != nil {
							log.Errorf("Failed to create restore %s for user %s: %v", restoreName, user, err)
						} else {
							log.Infof("Successfully created restore %s for user %s", restoreName, user)
						}
					}
				}(user)
			}
			log.Info("All additional backups and restores have been initiated for all primary users.")
		})

		Step("Unshare the clusters for a set of users and groups", func() {
			log.InfoD("Unsharing the clusters from each user with other users and groups")
			var wg sync.WaitGroup

			for _, user := range primaryUserList {
				wg.Add(1)
				go func(user string) {
					defer wg.Done()
					ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
					log.FailOnError(err, "Failed to fetch context for user")
					clusterUid, err := Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
					_, err = UnShareClusterWithValidation(ctx, SourceClusterName, clusterUid, firstUserList, firstGroupList)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying share of source [%s] cluster for user %s", SourceClusterName, user))
				}(user)
			}
			wg.Wait()
		})

		Step("Verify the newly shared backups", func() {
			log.InfoD("Verifying the newly shared backups")
			firstUserIndex := rand.Intn(len(firstUserList))
			firstRandomUser = firstUserList[firstUserIndex]
			fmt.Printf("Randomly selected user from unshared list of users: %s\n", firstRandomUser)
			backups := newBackupMap[primaryUserList[0]]
			ctx, err := backup.GetNonAdminCtx(firstRandomUser, CommonPassword)
			log.FailOnError(err, "Fetching user ctx")
			for _, backupName := range backups {
				err = BackupSuccessCheck(backupName, BackupOrgID, MaxWaitPeriodForBackupCompletionInMinutes*time.Minute, 30*time.Second, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying backup [%s] success check", backupName))
				log.Infof("Backup [%s] created successfully", backupName)
			}
			//Validate backup deletion
			var wg sync.WaitGroup
			for _, backupName := range backups {
				wg.Add(1)
				go func(backupName string) {
					defer GinkgoRecover()
					defer wg.Done()
					backupUid, err := Inst().Backup.GetBackupUID(ctx, backupName, BackupOrgID)
					log.FailOnError(err, "Failed to fetch the backup %s uid of the user %s", backupName, firstRandomUser)
					_, err = DeleteBackup(backupName, backupUid, BackupOrgID, ctx)
					log.FailOnError(err, "Failed to delete the backup %s of the user %s", backupName, firstRandomUser)
					err = DeleteBackupAndWait(backupName, ctx)
					log.FailOnError(err, fmt.Sprintf("waiting for backup [%s] deletion", backupName))
				}(backupName)
			}
			wg.Wait()
		})

		Step("Unshare the clusters from admin for other set of users and groups", func() {
			log.InfoD("Unsharing the clusters from admin for other set of users and groups")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			clusterEnumerateRequest := &api.ClusterEnumerateRequest{
				OrgId:          BackupOrgID,
				IncludeSecrets: false,
			}
			clusterObjs, err := Inst().Backup.EnumerateCluster(ctx, clusterEnumerateRequest)
			log.FailOnError(err, "Fetching cluster objects")
			for _, clusterObj := range clusterObjs.GetClusters() {
				clusterUid := clusterObj.GetUid()
				dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
				_, err = UnShareClusterWithValidation(ctx, SourceClusterName, clusterUid, secondUserList, secondGroupList)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying un-share of source [%s] cluster for px-admin", SourceClusterName))
			}
		})

		Step("Initiate additional backups for all primary users for validation", func() {
			log.InfoD("Taking additional backups for all primary users")
			var wg sync.WaitGroup
			var mu sync.Mutex

			for _, user := range primaryUserList {
				wg.Add(1)
				go func(user string) {
					defer wg.Done()
					ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
					log.FailOnError(err, "Failed to fetch context for user")
					backupNameList, err := TakeMultipleBackupsPerDeployment(ctx, BackupOrgID, SourceClusterName, numOfBackupsPerDeployment, snapshotLimit, backupLocationName, backupLocationUid, backupAppContexts, BackupNamePrefix)
					if err != nil {
						log.Errorf("Failed to take backups for user %s: %v", user, err)
						return
					}
					mu.Lock()
					newBackupResults[user] = backupNameList
					mu.Unlock()
				}(user)
			}
			wg.Wait()
			for user, backupNameList := range newBackupResults {
				log.Infof("User %s: All backups taken: %v", user, backupNameList)
			}
		})

		Step("Share the clusters from each user with other users and groups", func() {
			log.InfoD("Sharing the clusters from each user with other users and groups")
			var wg sync.WaitGroup

			for _, user := range primaryUserList {
				wg.Add(1)
				go func(user string) {
					defer wg.Done()
					ctx, err := backup.GetNonAdminCtx(user, CommonPassword)
					log.FailOnError(err, "Failed to fetch context for user")
					clusterUid, err := Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
					_, err = ShareClusterWithValidation(ctx, SourceClusterName, clusterUid, sharedUserList, groupList, false)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying share of source [%s] cluster for user %s", SourceClusterName, user))
				}(user)
			}
			wg.Wait()
		})

		Step("Verify the if the latest backups are not seen", func() {
			log.InfoD("Verify the if the latest backups are not seen")
			randomUserIndex := rand.Intn(len(firstUserList))
			randomUser := firstUserList[randomUserIndex]
			fmt.Printf("Randomly selected user from shared list of users: %s\n", randomUser)
			backups := newBackupResults[primaryUserList[0]]
			ctx, err := backup.GetNonAdminCtx(firstRandomUser, CommonPassword)
			log.FailOnError(err, "Fetching user ctx")
			bkpEnumerateReq := &api.BackupEnumerateRequest{
				OrgId: BackupOrgID}
			curBackups, err := backupDriver.EnumerateBackup(ctx, bkpEnumerateReq)
			log.FailOnError(err, "Fetching backups for user")
			nonePresent := true
			for _, backup := range backups {
				for _, current := range curBackups.GetBackups() {
					if backup == current.Name {
						nonePresent = false
						break
					}
				}
				if !nonePresent {
					break
				}
			}
			if nonePresent {
				log.InfoD("None of the backups are shared.")
			} else {
				log.Errorf("Backups are shared with the cluster.")
			}
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
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudAccountName, cloudAccountUid, ctx)
		log.InfoD("Deleting the backups")
		bkpEnumerateReq := &api.BackupEnumerateRequest{
			OrgId: BackupOrgID}
		allBackups, err := backupDriver.EnumerateBackup(ctx, bkpEnumerateReq)
		var wg sync.WaitGroup
		for _, bkp := range allBackups.GetBackups() {
			wg.Add(1)
			go func(bkp *api.BackupObject) {
				defer wg.Done()
				backupUID, err := Inst().Backup.GetBackupUID(ctx, bkp.Name, BackupOrgID)
				_, err = DeleteBackup(bkp.Name, backupUID, BackupOrgID, ctx)
				dash.VerifySafely(err, nil, fmt.Sprintf("Verifying backup deletion - %s", bkp.Name))
			}(bkp)
		}
		wg.Wait()
		log.InfoD("Deleting the restores")
		for _, restoreName := range restoreNames {
			err = DeleteRestore(restoreName, BackupOrgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting restore [%s]", restoreName))
		}
		log.InfoD("Switching context to destination cluster for clean up")
		err = SetDestinationKubeConfig()
		log.FailOnError(err, "Unable to switch context to destination cluster [%s]", DestinationClusterName)
		DestroyApps(backupAppContexts, opts)
		log.InfoD("Switching back context to Source cluster")
		err = SetSourceKubeConfig()
		log.FailOnError(err, "Unable to switch context to source cluster [%s]", SourceClusterName)

	})
})
