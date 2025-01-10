package tests

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/pure-px/stork/pkg/k8sutils"
	"github.com/pure-px/torpedo/drivers/backup"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
				appCtx.ReadinessTimeout = 30 * time.Minute
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
						err := CreateRestore(restoreName, backupName, namespaceMapping, SourceClusterName, clusterUid, BackupOrgID, ctx, make(map[string]string))
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
					err := Inst().Backup.WaitForBackupLocationDeletion(ctx, backupLocationName, backupLocationUID, BackupOrgID, ScaleBackupLocationDeleteTimeout, BackupLocationDeleteRetryTime)
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
		destClusterUid       string
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
			destClusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))
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
			err = CreateRestoreWithValidation(ctx, restoreName, currentBackupName, make(map[string]string), make(map[string]string), DestinationClusterName, destClusterUid, BackupOrgID, scheduledAppContexts)
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

// ShareLargeNumberOfBackupsWithLargeNumberOfUsers shares large number of backups to large number of users
var _ = Describe("{ShareLargeNumberOfBackupsWithLargeNumberOfUsers}", Label(TestCaseLabelsMap[ShareLargeNumberOfBackupsWithLargeNumberOfUsers]...), func() {

	var (
		scheduledAppContexts        []*scheduler.Context
		backupLocationUID           string
		cloudCredUID                string
		cloudCredUidList            []string
		appContextsToBackup         []*scheduler.Context
		bkpNamespaces               []string
		clusterStatus               api.ClusterInfo_StatusInfo_Status
		customBackupLocationName    string
		credName                    string
		chosenUser                  string
		numberOfUsers               int
		numberOfGroups              int
		groupSize                   int
		numberOfBackups             int
		userContexts                []context.Context
		backupLocationMapNew        map[string]string
		users                       []string
		groups                      []string
		backupNames                 []string
		numberOfSimultaneousBackups int
		clusterUid                  string
		chosenUserDestClusterUid    string
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("ShareLargeNumberOfBackupsWithLargeNumberOfUsers",
			"Share large number of backups to large number of users", nil, 82941, Mkoppal, Q4FY23)
		numberOfUsers, _ = strconv.Atoi(GetEnv(UsersToBeCreated, "200"))
		numberOfGroups, _ = strconv.Atoi(GetEnv(GroupsToBeCreated, "100"))
		groupSize, _ = strconv.Atoi(GetEnv(MaxUsersInGroup, "2"))
		numberOfBackups, _ = strconv.Atoi(GetEnv(MaxBackupsToBeCreated, "100"))
		userContexts = make([]context.Context, 0)
		bkpNamespaces = make([]string, 0)
		backupLocationMapNew = make(map[string]string)
		users = make([]string, 0)
		groups = make([]string, 0)
		backupNames = make([]string, 0)
		numberOfSimultaneousBackups = 3

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
	It("Share all backups at cluster level with a user group and revoke it and validate", func() {
		providers := GetBackupProviders()
		Step("Validate applications and get their labels", func() {
			log.InfoD("Validate applications")
			ValidateApplications(scheduledAppContexts)
		})

		Step("Create Users", func() {
			log.InfoD("Creating %d users to be added to the group", numberOfUsers)
			var wg sync.WaitGroup
			var mutex sync.Mutex
			for i := 1; i <= numberOfUsers; i++ {
				userName := fmt.Sprintf("testuser%v", i)
				firstName := fmt.Sprintf("FirstName%v", i)
				lastName := fmt.Sprintf("LastName%v", i)
				email := fmt.Sprintf("testuser%v@cnbu.com", i)
				wg.Add(1)
				go func(userName, firstName, lastName, email string) {
					defer wg.Done()
					err := backup.AddUser(userName, firstName, lastName, email, CommonPassword)
					log.FailOnError(err, "Failed to create user - %s", userName)
					mutex.Lock()
					users = append(users, userName)
					mutex.Unlock()
				}(userName, firstName, lastName, email)
			}
			wg.Wait()
		})

		Step("Create Groups", func() {
			log.InfoD("Creating %d groups", numberOfGroups)
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
					groups = append(groups, groupName)
					mutex.Unlock()
				}(groupName)
			}
			wg.Wait()
		})

		Step("Add users to group", func() {
			log.InfoD("Adding users to groups")
			var wg sync.WaitGroup
			for i, user := range users {
				groupIndex := i / groupSize
				wg.Add(1)
				go func(userName string, groupIndex int) {
					defer GinkgoRecover()
					defer wg.Done()
					err := backup.AddGroupToUser(userName, groups[groupIndex])
					log.FailOnError(err, "Failed to assign group to user")
				}(user, groupIndex)
			}
			wg.Wait()

			// Print the groups
			for _, group := range groups {
				usersOfGroup, err := backup.GetMembersOfGroup(group)
				log.FailOnError(err, "Error fetching members of the group - %v", group)
				log.Infof("Group [%v] contains the following users: \n%v", group, usersOfGroup)
			}
		})

		Step("Adding Credentials and Registering Backup Location", func() {
			log.InfoD("Using pre-provisioned bucket. Creating cloud credentials and backup location.")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, provider := range providers {
				cloudCredUID = uuid.New()
				cloudCredUidList = append(cloudCredUidList, cloudCredUID)
				backupLocationUID = uuid.New()
				credName = fmt.Sprintf("autogenerated-cred-%v", time.Now().Unix())
				err := CreateCloudCredential(provider, credName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", credName, BackupOrgID, provider))
				log.InfoD("Created Cloud Credentials with name - %s", credName)
				customBackupLocationName = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
				err = CreateBackupLocation(provider, customBackupLocationName, backupLocationUID, credName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", customBackupLocationName))
				log.InfoD("Created Backup Location with name - %s", customBackupLocationName)
				backupLocationMapNew[backupLocationUID] = customBackupLocationName
			}
		})

		Step("Register source and destination cluster for backup", func() {
			log.InfoD("Registering Source and Destination clusters and verifying the status")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid [%s]", SourceClusterName, clusterUid))
		})

		Step("Taking backup of applications", func() {
			log.InfoD("Taking backup of applications")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			appContextsToBackup = FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			log.InfoD("Taking %d backups", numberOfBackups)
			backupNames, err = TakeMultipleBackupsPerDeployment(ctx, BackupOrgID, SourceClusterName, clusterUid, numberOfBackups, numberOfSimultaneousBackups, customBackupLocationName, backupLocationUID, appContextsToBackup, BackupNamePrefix)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupNames))
		})

		Step("Share all backups with Full Access in source cluster with a group", func() {
			log.InfoD("Share all backups with Full Access in source cluster with a group")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = ClusterUpdateBackupShare(SourceClusterName, groups, nil, FullAccess, true, ctx)
			log.FailOnError(err, "Failed sharing all backups for cluster [%s]", SourceClusterName)
		})

		Step("Validate Full Access of backups shared at cluster level", func() {
			log.InfoD("Validate Full Access of backups shared at cluster level for a user of a group")
			// Get user from group
			var err error
			chosenUser, err = backup.GetRandomUserFromGroup(groups[rand.Intn(numberOfGroups-1)])
			log.FailOnError(err, "Failed to get a random user from group [%s]", groups[0])
			log.Infof("User chosen to validate full access - %s", chosenUser)

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

			chosenUserDestClusterUid, err = Inst().Backup.GetClusterUID(ctxNonAdmin, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))

			// Start Restore
			backupName := backupNames[rand.Intn(numberOfBackups-1)]
			restoreName := fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
			appContextsToBackup = FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			err = CreateRestoreWithValidation(ctxNonAdmin, restoreName, backupName, make(map[string]string), make(map[string]string), DestinationClusterName, chosenUserDestClusterUid, BackupOrgID, appContextsToBackup)
			log.FailOnError(err, "Restoring of backup [%s] has failed with name - [%s]", backupName, restoreName)

			// Restore validation to make sure that the user with Full Access can restore
			log.InfoD("Restoring of backup [%s] was successful with name - [%s]", backupName, restoreName)
			log.Infof("About to delete restore - %s", restoreName)
			err = DeleteRestore(restoreName, BackupOrgID, ctxNonAdmin)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting Restore %s", restoreName))

			// Get Backup UID
			backupDriver := Inst().Backup
			backupUID, err := backupDriver.GetBackupUID(ctx, backupName, BackupOrgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupName)

			// Delete backup to confirm that the user has Full Access
			backupDeleteResponse, err := DeleteBackup(backupName, backupUID, BackupOrgID, ctxNonAdmin)
			log.FailOnError(err, "Backup [%s] could not be deleted by user [%s]", backupName, chosenUser)
			dash.VerifyFatal(backupDeleteResponse.String(), "",
				fmt.Sprintf("Verifying backup [%s] deletion is successful by user [%s]", backupName, chosenUser))
		})

		Step("Share all backups with Restore Access in source cluster with a group", func() {
			log.InfoD("Share all backups with Full Access in source cluster with a group")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = ClusterUpdateBackupShare(SourceClusterName, groups, nil, RestoreAccess, true, ctx)
			log.FailOnError(err, "Failed sharing all backups for cluster [%s]", SourceClusterName)
		})

		Step("Validate Restore Access of backups shared at cluster level", func() {
			log.InfoD("Validate Restore Access of backups shared at cluster level")
			log.Infof("User chosen to validate restore access - %s", chosenUser)

			// Get Admin Context - needed to share backup and get backup UID
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			// Get user context
			ctxNonAdmin, err := backup.GetNonAdminCtx(chosenUser, CommonPassword)
			log.FailOnError(err, "Fetching non admin ctx")

			// Start Restore
			backupName := backupNames[rand.Intn(numberOfBackups-1)]
			restoreName := fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
			appContextsToBackup = FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			err = CreateRestoreWithValidation(ctxNonAdmin, restoreName, backupName, make(map[string]string), make(map[string]string), DestinationClusterName, chosenUserDestClusterUid, BackupOrgID, appContextsToBackup)
			log.FailOnError(err, "Restoring of backup [%s] has failed with name - [%s]", backupName, restoreName)
			// Restore validation to make sure that the user with Restore Access can restore
			log.InfoD("Restoring of backup [%s] was successful with name - [%s]", backupName, restoreName)
			log.Infof("About to delete restore - %s", restoreName)
			err = DeleteRestore(restoreName, BackupOrgID, ctxNonAdmin)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting Restore %s", restoreName))

			// Get Backup UID
			backupDriver := Inst().Backup
			backupUID, err := backupDriver.GetBackupUID(ctx, backupName, BackupOrgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupName)

			// Delete backup to confirm that the user cannot delete the backup
			_, err = DeleteBackup(backupName, backupUID, BackupOrgID, ctxNonAdmin)
			dash.VerifyFatal(strings.Contains(err.Error(), "doesn't have permission to delete backup"), true, "Verifying backup deletion is not possible")
		})

		Step("Share all backups with View Only Access in source cluster with a group", func() {
			log.InfoD("Share all backups with Full Access in source cluster with a group")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = ClusterUpdateBackupShare(SourceClusterName, groups, nil, ViewOnlyAccess, true, ctx)
			log.FailOnError(err, "Failed sharing all backups for cluster [%s]", SourceClusterName)
		})

		Step("Validate Restore Access of backups shared at cluster level", func() {
			log.InfoD("Validate Restore Access of backups shared at cluster level")
			log.Infof("User chosen to validate restore access - %s", chosenUser)

			// Get Admin Context - needed to share backup and get backup UID
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			// Get user context
			ctxNonAdmin, err := backup.GetNonAdminCtx(chosenUser, CommonPassword)
			log.FailOnError(err, "Fetching non admin ctx")

			// Start Restore
			backupName := backupNames[rand.Intn(numberOfBackups-1)]
			restoreName := fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
			err = CreateRestore(restoreName, backupName, make(map[string]string), DestinationClusterName, chosenUserDestClusterUid, BackupOrgID, ctxNonAdmin, make(map[string]string))

			// Restore validation to make sure that the user with View Access cannot restore
			dash.VerifyFatal(strings.Contains(err.Error(), "doesn't have permission to restore backup"), true, "Verifying backup restore is not possible")

			// Get Backup UID
			backupDriver := Inst().Backup
			backupUID, err := backupDriver.GetBackupUID(ctx, backupName, BackupOrgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupName)

			// Delete backup to confirm that the user cannot delete the backup
			_, err = DeleteBackup(backupName, backupUID, BackupOrgID, ctxNonAdmin)
			dash.VerifyFatal(strings.Contains(err.Error(), "doesn't have permission to delete backup"), true, "Verifying backup deletion is not possible")
		})
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)

		var wg sync.WaitGroup
		log.Infof("Cleaning up users")
		for _, userName := range users {
			wg.Add(1)
			go func(userName string) {
				defer GinkgoRecover()
				defer wg.Done()
				err := backup.DeleteUser(userName)
				log.FailOnError(err, "Error deleting user %v", userName)
			}(userName)
		}
		wg.Wait()

		log.Infof("Cleaning up groups")
		for _, groupName := range groups {
			wg.Add(1)
			go func(groupName string) {
				defer GinkgoRecover()
				defer wg.Done()
				err := backup.DeleteGroup(groupName)
				log.FailOnError(err, "Error deleting user %v", groupName)
			}(groupName)
		}
		wg.Wait()

		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		CleanupCloudSettingsAndClusters(backupLocationMapNew, credName, cloudCredUID, ctx)

		log.Infof("Deleting registered clusters for non-admin context")
		for _, ctxNonAdmin := range userContexts {
			CleanupCloudSettingsAndClusters(make(map[string]string), "", "", ctxNonAdmin)
		}

	})
})

// This TC takes backup of a namespace with large number of resources and restores it.
var _ = Describe("{LargeResourceNamespaceBackup}", Label(TestCaseLabelsMap[LargeResourceNamespaceBackup]...), func() {
	/*
		1. Deploy large number of secrets and configmaps into a namespace.
		2. Populate the configmaps and secrets with some data.
		3. Create backup location and cloud settings.
		4. Register source and destination clusters.
		5. Create backup of the namespace.
		6. Validate that the backup is a large resource backup.
		7. Restore the backup to the destination cluster.
		8. Restore the backup to the source cluster in different namespace.
		9. Restore the backup to the source cluster in the same namespace.
	*/

	var (
		err                                        error
		ctx                                        context.Context
		scheduledAppContexts                       []*scheduler.Context
		testrailID                                 int
		numberOfResources                          int
		numberOfEntries                            int
		namespace                                  string
		appList                                    []string
		configMapEntries                           map[string]string
		secretEntries                              map[string]string
		providers                                  []string
		cloudCredName                              string
		cloudCredUID                               string
		backupLocationName                         string
		backupLocationUID                          string
		backupLocationMap                          map[string]string
		sourceClusterUid                           string
		destClusterUid                             string
		backupName                                 string
		restoreToDestinationClusterName            string
		restoreToSameClusterDifferentNamespaceName string
		restoreToSameClusterSameNamespaceName      string
		isLargeResourceBackup                      bool
		storkControllerCM                          *corev1.ConfigMap
		oldLargeResourceSizeLimit                  string
		newLargeResourceSizeLimit                  string
	)

	JustBeforeEach(func() {
		testrailID = 85773
		StartPxBackupTorpedoTest("LargeResourceNamespaceBackup", "To verify backup of namespace with large number of resources and restore it", nil, testrailID, Dchothani, Q3FY25)
		ctx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		numberOfResources, _ = strconv.Atoi(GetEnv(NumberOfResources, "1000"))
		numberOfEntries, _ = strconv.Atoi(GetEnv(NumberOfEntries, "1000"))
		newLargeResourceSizeLimit = GetEnv(ReduceLargeResourceSizeLimit, "102400")
		namespace = fmt.Sprintf("namespace-%d-%s", testrailID, RandomString(6))
		backupLocationMap = make(map[string]string)
		providers = GetBackupProviders()
		appList = Inst().AppList
		defer func() {
			log.Infof("Resetting applist and removing the custom app config")
			Inst().AppList = appList
			delete(Inst().CustomAppConfig, "config-maps")
			delete(Inst().CustomAppConfig, "secrets")
			err := Inst().S.RescanSpecs(Inst().SpecDir, Inst().V.String())
			log.FailOnError(err, "Failed while rescanning specs")
		}()
		Inst().AppList = []string{"config-maps", "secrets"}
		Inst().CustomAppConfig["config-maps"] = scheduler.AppConfig{
			ClaimsCount: numberOfResources,
		}
		Inst().CustomAppConfig["secrets"] = scheduler.AppConfig{
			ClaimsCount: numberOfResources,
		}
		err := Inst().S.RescanSpecs(Inst().SpecDir, Inst().V.String())
		log.FailOnError(err, "Failed to rescan specs from %s for storage provider %s", Inst().SpecDir, Inst().V.String())
		log.InfoD("Updating %s in %s to set %s:%s", StorkControllerConfigMap, DefaultStorkDeploymentNamespace, k8sutils.LargeResourceSizeLimitName, newLargeResourceSizeLimit)
		storkControllerCM, err = core.Instance().GetConfigMap(StorkControllerConfigMap, DefaultStorkDeploymentNamespace)
		log.FailOnError(err, fmt.Sprintf("Failed to get %s configmap", StorkControllerConfigMap))
		if _, ok := storkControllerCM.Data[k8sutils.LargeResourceSizeLimitName]; ok {
			oldLargeResourceSizeLimit = storkControllerCM.Data[k8sutils.LargeResourceSizeLimitName]
		}
		storkControllerCM.Data[k8sutils.LargeResourceSizeLimitName] = newLargeResourceSizeLimit
		storkControllerCM, err = core.Instance().UpdateConfigMap(storkControllerCM)
		log.FailOnError(err, fmt.Sprintf("Failed to update %s configmap", StorkControllerConfigMap))
		log.InfoD("scheduling applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, testrailID)
		appContexts := ScheduleApplicationsOnNamespace(namespace, taskName)
		for _, appCtx := range appContexts {
			appCtx.ReadinessTimeout = AppReadinessTimeout
			scheduledAppContexts = append(scheduledAppContexts, appCtx)
		}
	})

	It("Backup a namespace with large number of resources and restore it", func() {
		Step("Validating applications", func() {
			log.InfoD("validating applications")
			ValidateApplications(scheduledAppContexts)
		})

		Step("Populating the configmaps and secrets with data", func() {
			log.InfoD("Populating the configmaps and secrets with data")
			for _, ctx := range scheduledAppContexts {
				if ctx.App.Key == "config-maps" {
					configMapEntries, err = PopulateConfigMapsInContext(ctx, numberOfEntries)
				} else if ctx.App.Key == "secrets" {
					secretEntries, err = PopulateSecretsInContext(ctx, numberOfEntries)
				}
			}
		})

		Step("Creating backup location and cloud setting", func() {
			log.InfoD("Creating backup location and cloud setting")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
				backupLocationName = fmt.Sprintf("%s-%v", getGlobalBucketName(provider), RandomString(6))
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocationName
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocation(provider, backupLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, "Creating backup location")
			}
		})

		Step("Registering clusters for backup", func() {
			log.InfoD("Registering clusters for backup")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err := Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			sourceClusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, DestinationClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", DestinationClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", DestinationClusterName))
			destClusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))
		})

		Step("Creating a backup of the namespace", func() {
			log.InfoD("Creating backup of application from source cluster")
			backupName = fmt.Sprintf("%s-%v", BackupNamePrefix, time.Now().Unix())
			err = CreateBackup(backupName, SourceClusterName, backupLocationName, backupLocationUID, []string{namespace}, nil, BackupOrgID, sourceClusterUid, "", "", "", "", ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s] with namespace [%s]", backupName, namespace))
		})

		Step("Checking whether the backup is a large resource backup", func() {
			log.InfoD("Checking whether the backup [%s] is a large resource backup", backupName)
			isLargeResourceBackup, err = IsLargeResourceBackup(ctx, backupName, BackupOrgID)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Checking the backup [%s] is a large resource backup", backupName))
			dash.VerifyFatal(isLargeResourceBackup, true, fmt.Sprintf("Verifying the backup [%s] is a large resource backup", backupName))
		})

		Step("Restoring the backup on destination cluster", func() {
			log.InfoD("Restoring the backup on the destination cluster")
			restoreToDestinationClusterName = fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
			err = CreateRestore(restoreToDestinationClusterName, backupName, make(map[string]string), DestinationClusterName, destClusterUid, BackupOrgID, ctx, nil)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s] from backup [%s] on the destination cluster", restoreToDestinationClusterName, backupName))
			err = SetDestinationKubeConfig()
			dash.VerifyFatal(err, nil, "switching to destination kubeconfig")
			configMapList, err := core.Instance().ListConfigMap(namespace, metav1.ListOptions{})
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching configmaps from restored namespace [%s] in [%s]", namespace, DestinationClusterName))
			if len(configMapList.Items) < numberOfResources {
				dash.Fatal("All the configmaps are not restored")
			}
			for _, cm := range configMapList.Items {
				// this configmap exists in every namespace; excluding this in validation
				if cm.Name == "kube-root-ca.crt" {
					continue
				}
				_, err = ValidateConfigMapEntries(cm.Name, namespace, configMapEntries)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Validating configmap [%s] entries", cm.Name))
			}
			secretsList, err := core.Instance().ListSecret(namespace, metav1.ListOptions{})
			if len(secretsList.Items) < numberOfResources {
				dash.Fatal("All the secrets are not restored")
			}
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching secrets from restored namespace [%s] in [%s]", namespace, DestinationClusterName))
			for _, cm := range secretsList.Items {
				_, err = ValidateSecretEntries(cm.Name, namespace, secretEntries)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Validating secret [%s] entries", cm.Name))
			}
		})

		Step("Restoring the backup on source cluster with different namespace", func() {
			log.InfoD("Restoring the backup on source cluster with different namespace")
			restoreNamespace := fmt.Sprintf("namespace-%d-%s", testrailID, RandomString(6))
			restoreNamespaceMapping := map[string]string{namespace: restoreNamespace}
			restoreToSameClusterDifferentNamespaceName = fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
			err = CreateRestore(restoreToSameClusterDifferentNamespaceName, backupName, restoreNamespaceMapping, SourceClusterName, sourceClusterUid, BackupOrgID, ctx, nil)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s] from backup [%s] on the source cluster with namespace [%s]", restoreToSameClusterDifferentNamespaceName, backupName, restoreNamespace))
			err = SetSourceKubeConfig()
			dash.VerifyFatal(err, nil, "switching to source kubeconfig")
			configMapList, err := core.Instance().ListConfigMap(restoreNamespace, metav1.ListOptions{})
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching configmaps from restored namespace [%s] in [%s]", restoreNamespace, SourceClusterName))
			if len(configMapList.Items) < numberOfResources {
				dash.Fatal("All the configmaps are not restored")
			}
			for _, cm := range configMapList.Items {
				// this configmap exists in every namespace; excluding this in validation
				if cm.Name == "kube-root-ca.crt" {
					continue
				}
				_, err = ValidateConfigMapEntries(cm.Name, restoreNamespace, configMapEntries)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Validating configmap [%s] entries", cm.Name))
			}
			secretsList, err := core.Instance().ListSecret(restoreNamespace, metav1.ListOptions{})
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching secrets from restored namespace [%s] in [%s]", restoreNamespace, SourceClusterName))
			if len(secretsList.Items) < numberOfResources {
				dash.Fatal("All the secrets are not restored")
			}
			for _, secret := range secretsList.Items {
				_, err = ValidateSecretEntries(secret.Name, restoreNamespace, secretEntries)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Validating secret [%s] entries", secret.Name))
			}
		})

		Step("Restoring the backup on source cluster in same namespace", func() {
			log.InfoD("Restoring the backup on source cluster in same namespace")
			restoreToSameClusterSameNamespaceName = fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
			err = CreateRestore(restoreToSameClusterSameNamespaceName, backupName, nil, SourceClusterName, sourceClusterUid, BackupOrgID, ctx, nil)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s] from backup [%s] on the source cluster in same namespace", restoreToSameClusterDifferentNamespaceName, backupName))
			err = SetSourceKubeConfig()
			dash.VerifyFatal(err, nil, "switching to source kubeconfig")
			configMapList, err := core.Instance().ListConfigMap(namespace, metav1.ListOptions{})
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching configmaps from restored namespace [%s] in [%s]", namespace, SourceClusterName))
			if len(configMapList.Items) < numberOfResources {
				dash.Fatal("All the configmaps are not restored")
			}
			for _, cm := range configMapList.Items {
				// this configmap exists in every namespace; excluding this in validation
				if cm.Name == "kube-root-ca.crt" {
					continue
				}
				_, err = ValidateConfigMapEntries(cm.Name, namespace, configMapEntries)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Validating configmap [%s] entries", cm.Name))
			}
			secretsList, err := core.Instance().ListSecret(namespace, metav1.ListOptions{})
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching secrets from restored namespace [%s] in [%s]", namespace, SourceClusterName))
			if len(secretsList.Items) < numberOfResources {
				dash.Fatal("All the secrets are not restored")
			}
			for _, secret := range secretsList.Items {
				_, err = ValidateSecretEntries(secret.Name, namespace, secretEntries)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Validating secret [%s] entries", secret.Name))
			}
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		defer func() {
			err = SetSourceKubeConfig()
			dash.VerifyFatal(err, nil, "switching to source kubeconfig")
			if oldLargeResourceSizeLimit != "" {
				storkControllerCM.Data[k8sutils.LargeResourceSizeLimitName] = oldLargeResourceSizeLimit
			} else {
				delete(storkControllerCM.Data, k8sutils.LargeResourceSizeLimitName)
			}
			storkControllerCM, err = core.Instance().UpdateConfigMap(storkControllerCM)
			log.FailOnError(err, fmt.Sprintf("Failed to update %s configmap", StorkControllerConfigMap))
		}()
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		log.InfoD("Deleting deployed applications")
		DestroyApps(scheduledAppContexts, opts)
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})
