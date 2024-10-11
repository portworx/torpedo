package tests

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	_ "github.com/rancher/norman/clientbase"
	_ "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"golang.org/x/sync/errgroup"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// This test case calculates the time taken to delete a backup with heavy load
var _ = Describe("{TimeTakenToDeleteBackupWithHeavyLoad}", Label(TestCaseLabelsMap[TimeTakenToDeleteBackupWithHeavyLoad]...), func() {
	var (
		controlChannel       chan string
		errorGroup           *errgroup.Group
		scheduledAppContexts []*scheduler.Context
		backupLocation       string
		backupLocationUID    string
		cloudCredUID         string
		bkpNamespaces        []string
		backupName           string
		clusterUid           string
		cloudCredName        string
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		numberOfVolumes      int
		err                  error
		ctx                  context.Context
		CloudCredUIDMap      map[string]string
		backupLocationMap    map[string]string
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("TimeTakenToDeleteBackupWithHeavyLoad", "Calculates the time taken to delete a backup with heavy load", nil, 301290, Sagrawal, Q3FY25)
		CloudCredUIDMap = make(map[string]string)
		backupLocationMap = make(map[string]string)
		bkpNamespaces = make([]string, 0)
		ctx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		log.InfoD("Switching cluster context to destination cluster")
		err := SetDestinationKubeConfig()
		log.FailOnError(err, "Switching context to destination cluster failed")
		log.InfoD("Deploy applications on destination cluster")
		volumeCountForParallelDelete := os.Getenv("VOLUME_COUNT_FOR_PARALLEL_DELETE")
		if volumeCountForParallelDelete == "" {
			volumeCountForParallelDelete = "10"
		}
		numberOfVolumes, err = strconv.Atoi(volumeCountForParallelDelete)
		log.InfoD("The number of PVC to be deployed are %v", numberOfVolumes)
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		namespace := fmt.Sprintf("multiple-volume-ns-%s", RandomString(6))
		bkpNamespaces = append(bkpNamespaces, namespace)
		Inst().AppList = []string{"vdbench-multi-vol"}
		Inst().CustomAppConfig["vdbench-multi-vol"] = scheduler.AppConfig{
			ClaimsCount: numberOfVolumes,
		}
		err = Inst().S.RescanSpecs(Inst().SpecDir, Inst().V.String())
		log.FailOnError(err, "Failed to rescan specs from %s for storage provider %s with claim count %v", Inst().SpecDir, Inst().V.String(), numberOfVolumes)
		appContexts := ScheduleApplicationsOnNamespace(namespace, TaskNamePrefix)
		for _, appCtx := range appContexts {
			appCtx.ReadinessTimeout = AppReadinessTimeout
			scheduledAppContexts = append(scheduledAppContexts, appCtx)
		}
	})
	It("Calculates the time taken to delete a backup with heavy load", func() {
		Step("Validate applications", func() {
			controlChannel, errorGroup = ValidateApplicationsStartData(scheduledAppContexts, ctx)
		})
		Step("Creating cloud credentials", func() {
			log.InfoD("Creating cloud credentials")
			providers := GetBackupProviders()
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, RandomString(5))
				cloudCredUID = uuid.New()
				CloudCredUIDMap[cloudCredUID] = cloudCredName
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
			}
		})

		Step("Register cluster for backup", func() {
			// Here we need to switch context to source cluster as CreateApplicationClusters runs k8s operation of fetching the cm kubeconfigs
			defer func() {
				err := SetDestinationKubeConfig()
				log.FailOnError(err, "Switching context to source cluster")
			}()
			err := SetSourceKubeConfig()
			log.FailOnError(err, "Switching context to source cluster")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creating source %s and destination %s cluster", SourceClusterName, DestinationClusterName))
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, DestinationClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", DestinationClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", DestinationClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))
		})

		Step("Creating backup location", func() {
			log.InfoD("Creating backup location")
			providers := GetBackupProviders()
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, RandomString(5))
				backupLocation = fmt.Sprintf("autogenerated-backup-location-%v", RandomString(5))
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocation
				err := CreateBackupLocationWithContext(provider, backupLocation, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", ctx, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocation))
			}
		})

		Step("Start backup of application to bucket from destination cluster", func() {
			log.InfoD("Start backup of application to bucket from destination cluster")
			backupName = fmt.Sprintf("%s-%v", BackupNamePrefix, RandomString(5))
			appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, bkpNamespaces)
			err = CreateBackupWithValidation(ctx, backupName, DestinationClusterName, backupLocation, backupLocationUID, appContextsToBackup, nil, BackupOrgID, clusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
		})

		Step("Deleting the backup created to measure the time taken to delete backup", func() {
			log.InfoD("Deleting the backup created to measure the time taken to delete backup")
			backupDriver := Inst().Backup
			log.Infof("About to delete backup - %s", backupName)
			backupUID, err := backupDriver.GetBackupUID(ctx, backupName, BackupOrgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupName)
			_, err = DeleteBackup(backupName, backupUID, BackupOrgID, ctx)
			log.FailOnError(err, "Failed to issue delete backup for - %s", backupName)
			err = DeleteBackupAndWait(backupName, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleted backup - [%s]", backupName))
		})

	})
	JustAfterEach(func() {
		defer func() {
			err := SetSourceKubeConfig()
			log.FailOnError(err, "Switching context to source cluster")
			EndPxBackupTorpedoTest(scheduledAppContexts)
		}()
		log.InfoD("Deleting the deployed apps after the testcase")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		err = DestroyAppsWithData(scheduledAppContexts, opts, controlChannel, errorGroup)
		log.FailOnError(err, "App Deletion failed")
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})

// This test case calculates the time taken to delete all schedule backup with heavy load after the schedule is suspended
var _ = Describe("{TimeTakenToDeleteScheduleBackupWithHeavyLoadAfterSuspendingTheSchedule}", Label(TestCaseLabelsMap[TimeTakenToDeleteScheduleBackupWithHeavyLoadAfterSuspendingTheSchedule]...), func() {
	var (
		numberOfVolumes            int
		cloudCredUID               string
		cloudCredName              string
		backupLocationUID          string
		periodicSchedulePolicyName string
		periodicSchedulePolicyUid  string
		scheduleName               string
		clusterUid                 string
		err                        error
		backupLocation             string
		sizePerVolume              int64
		errors                     []string
		bkpNamespaces              []string
		controlChannel             chan string
		totalNoOfScheduleBackups   int
		mu                         sync.Mutex
		volumeList                 []*volume.Volume
		wg                         sync.WaitGroup
		errorGroup                 *errgroup.Group
		ctx                        context.Context
		scheduledAppContexts       []*scheduler.Context
		clusterStatus              api.ClusterInfo_StatusInfo_Status
		backupLocationMap          map[string]string
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("TimeTakenToDeleteScheduleBackupWithHeavyLoadAfterSuspendingTheSchedule", "Calculates the time taken to delete all schedule backup with heavy load after the schedule is suspended", nil, 301291, Sagrawal, Q3FY25)
		bkpNamespaces = make([]string, 0)
		totalNoOfScheduleBackups = 12
		backupLocationMap = make(map[string]string)
		sizePerVolume = 45
		ctx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		log.InfoD("Switching cluster context to destination cluster")
		err = SetDestinationKubeConfig()
		log.FailOnError(err, "Switching context to destination cluster failed")
		log.InfoD("Deploy applications on destination cluster")
		volumeCountForParallelDelete := os.Getenv("VOLUME_COUNT_FOR_PARALLEL_DELETE")
		if volumeCountForParallelDelete == "" {
			volumeCountForParallelDelete = "10"
		}
		numberOfVolumes, err = strconv.Atoi(volumeCountForParallelDelete)
		log.InfoD("The number of PVC to be deployed are %v", numberOfVolumes)
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		namespace := fmt.Sprintf("multiple-volume-ns-%s", RandomString(6))
		bkpNamespaces = append(bkpNamespaces, namespace)
		Inst().AppList = []string{"vdbench-multi-vol"}
		Inst().CustomAppConfig["vdbench-multi-vol"] = scheduler.AppConfig{

			ClaimsCount: numberOfVolumes,
		}
		err = Inst().S.RescanSpecs(Inst().SpecDir, Inst().V.String())
		log.FailOnError(err, "Failed to rescan specs from %s for storage provider %s with claim count %v", Inst().SpecDir, Inst().V.String(), numberOfVolumes)
		appContexts := ScheduleApplicationsOnNamespace(namespace, TaskNamePrefix)
		for _, appCtx := range appContexts {
			appCtx.ReadinessTimeout = AppReadinessTimeout
			scheduledAppContexts = append(scheduledAppContexts, appCtx)
		}
	})

	It("Calculates the time taken to delete all schedule backup with heavy load after the schedule is suspended", func() {
		Step("Validate applications and verifying the size usage of all the pvc", func() {
			log.InfoD("Validate applications and verifying the size usage of all the pvc")
			controlChannel, errorGroup = ValidateApplicationsStartData(scheduledAppContexts, ctx)
			volumeList, err = GetVolumeFromContexts(scheduledAppContexts)
			for _, vol := range volumeList {
				// Ignoring first volume as the I/O are continuously getting deleted and re-written for this PVC
				if vol.Name != "vdbench-pvc-1" {
					wg.Add(1)
					go func(volume *volume.Volume) {
						defer wg.Done()
						_, err := WaitForVolToHaveMinimumSize(volume, sizePerVolume, 60*time.Minute, 30*time.Second)
						if err != nil {
							mu.Lock()
							errors = append(errors, fmt.Sprintf("error waiting for volume %s: %v", volume.Name, err))
							mu.Unlock()
						}
					}(vol)
				}
			}
			wg.Wait()
			dash.VerifyFatal(len(errors), 0, fmt.Sprintf("delete backup errors: %s", strings.Join(errors, "; ")))
		})

		Step("Creating cloud credentials", func() {
			log.InfoD("Creating cloud credentials")
			providers := GetBackupProviders()
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, RandomString(5))
				cloudCredUID = uuid.New()
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
			}
		})

		Step("Creating backup location", func() {
			log.InfoD("Creating backup location")
			providers := GetBackupProviders()
			for _, provider := range providers {
				backupLocation = fmt.Sprintf("backup-location-%v", RandomString(5))
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocation
				err := CreateBackupLocationWithContext(provider, backupLocation, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", ctx, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s using cloud cred %v with UID %v", backupLocation, cloudCredName, cloudCredUID))
			}
		})

		Step("Register cluster for backup", func() {
			// Here we need to switch context to source cluster as CreateApplicationClusters runs k8s operation of fetching the cm kubeconfigs
			err := SetSourceKubeConfig()
			log.FailOnError(err, "Switching context to source cluster")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creating source %s and destination %s cluster", SourceClusterName, DestinationClusterName))
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, DestinationClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", DestinationClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", DestinationClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))
		})

		Step("Create schedule policy", func() {
			defer func() {
				err := SetDestinationKubeConfig()
				log.FailOnError(err, "Switching context to source cluster")
			}()
			log.InfoD("Creating a schedule policy")
			periodicSchedulePolicyName = fmt.Sprintf("parallel-delete-%v", RandomString(5))
			periodicSchedulePolicyUid = uuid.New()
			periodicSchedulePolicyInfo := Inst().Backup.CreateIntervalSchedulePolicy(14, 15, 6)
			err := Inst().Backup.BackupSchedulePolicy(periodicSchedulePolicyName, periodicSchedulePolicyUid, BackupOrgID, periodicSchedulePolicyInfo)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of periodic schedule policy of interval 15 minutes named [%s]", periodicSchedulePolicyName))
			periodicSchedulePolicyUid, err = Inst().Backup.GetSchedulePolicyUid(BackupOrgID, ctx, periodicSchedulePolicyName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching uid of periodic schedule policy named [%s]", periodicSchedulePolicyName))
		})

		Step("Creating schedule backups on destination cluster", func() {
			log.InfoD("Creating %v schedule backup on destination cluster", totalNoOfScheduleBackups)
			scheduleName = fmt.Sprintf("%s-%v", BackupNamePrefix, RandomString(5))
			labelSelectors := make(map[string]string)
			_, err := CreateScheduleBackupWithValidation(ctx, scheduleName, DestinationClusterName, clusterUid, backupLocation, backupLocationUID, scheduledAppContexts, labelSelectors, BackupOrgID, "", "", "", "", periodicSchedulePolicyName, periodicSchedulePolicyUid)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of schedule backup on destination cluster with schedule name [%s]", scheduleName))
			log.InfoD("Waiting for %v schedule backups to be created", totalNoOfScheduleBackups)
			noOfScheduleBackups := func() (interface{}, bool, error) {
				allScheduleBackupNames, err := Inst().Backup.GetAllScheduleBackupNames(ctx, scheduleName, BackupOrgID)
				log.InfoD("%v schedule backups are %v", len(allScheduleBackupNames), allScheduleBackupNames)
				if err != nil {
					return "", true, err
				}
				log.InfoD("Verify the latest backup status %v", allScheduleBackupNames[len(allScheduleBackupNames)-1])
				err = BackupSuccessCheck(allScheduleBackupNames[len(allScheduleBackupNames)-1], BackupOrgID, MaxWaitPeriodForBackupCompletionInMinutes*time.Minute, 30*time.Second, ctx)
				if err != nil {
					return "", false, err
				}
				if len(allScheduleBackupNames) < totalNoOfScheduleBackups {
					return "", true, fmt.Errorf("less than %v schedule backups are created", totalNoOfScheduleBackups)
				}
				//log.InfoD("Validating all the schedule backups")
				//for _, backupName := range allScheduleBackupNames {
				//	err := ValidateBackup(ctx, backupName, BackupOrgID, scheduledAppContexts, nil)
				//	if err != nil {
				//		return "", false, err
				//	}
				//}
				return "", false, nil
			}
			_, err = DoRetryWithTimeoutWithGinkgoRecover(noOfScheduleBackups, 24*time.Hour, 15*time.Minute)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Total %v schedule backups are created successfully", totalNoOfScheduleBackups))
		})

		Step("Suspending the schedule policy and deleting all the schedule backups at once", func() {
			log.InfoD("Suspending the schedule policy and deleting all the schedule backups at once")
			err = SuspendBackupSchedule(scheduleName, periodicSchedulePolicyName, BackupOrgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Suspending Backup Schedule [%s] ", scheduleName))
			log.InfoD("Getting list of all schedule backups")
			allScheduleBackupNames, err := Inst().Backup.GetAllScheduleBackupNames(ctx, scheduleName, BackupOrgID)
			dash.VerifyFatal(err, nil, fmt.Sprintf("List of all schedule backups of schedule [%s] is %v", scheduleName, allScheduleBackupNames))
			log.InfoD("Deleting list of all schedule backups at once")
			err = DeleteListOfBackupsAtOnce(ctx, allScheduleBackupNames, false)
			dash.VerifyFatal(err, nil, "Deleting all the schedule backups at once")
		})

		Step("Calculating the time taken to delete all the schedule backups", func() {
			log.InfoD("Calculating the time taken to delete all the schedule backups")
			TimeTakenToDeleteAllScheduleBackups, err := TimeTakenToDeleteListOfAllScheduleBackupObjects(ctx, scheduleName, 24*time.Hour, 5*time.Second)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Time take to delete all %v schedule backups of schedule %v is %v", totalNoOfScheduleBackups, scheduleName, TimeTakenToDeleteAllScheduleBackups))
		})
	})

	JustAfterEach(func() {
		defer func() {
			err := SetSourceKubeConfig()
			log.FailOnError(err, "Switching context to source cluster")
			EndPxBackupTorpedoTest(scheduledAppContexts)
		}()
		log.InfoD("Deleting backup schedule")
		err = DeleteSchedule(scheduleName, DestinationClusterName, BackupOrgID, ctx, true)
		dash.VerifySafely(err, nil, fmt.Sprintf("Verification of deleting backup schedule - %s", scheduleName))
		log.InfoD("Deleting the deployed apps after the testcase")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		err = DestroyAppsWithData(scheduledAppContexts, opts, controlChannel, errorGroup)
		log.FailOnError(err, "App Deletion failed")
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
		err := SetSourceKubeConfig()
		log.FailOnError(err, "Switching context to source cluster")
		log.Infof("Deleting backup schedule policy")
		err = Inst().Backup.DeleteBackupSchedulePolicy(BackupOrgID, []string{periodicSchedulePolicyName})
		dash.VerifySafely(err, nil, fmt.Sprintf("Deleting backup schedule policies %s ", scheduleName))
	})
})
