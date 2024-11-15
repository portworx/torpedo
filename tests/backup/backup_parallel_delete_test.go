package tests

import (
	"bytes"
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/sched-ops/task"
	"github.com/pure-px/torpedo/drivers/backup"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/drivers/volume"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
	_ "github.com/rancher/norman/clientbase"
	_ "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"golang.org/x/sync/errgroup"
	"os/exec"
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
		numberOfVolumes, _ = strconv.Atoi(GetEnv(VolCountForParallelDelete, "20"))
		log.InfoD("The number of PVC to be deployed are %v", numberOfVolumes)
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		namespace := fmt.Sprintf("multiple-volume-pvc-%s", RandomString(6))
		bkpNamespaces = append(bkpNamespaces, namespace)
		Inst().AppList = []string{"vdbench-multi-dep-multi-volume"}
		for i := 0; i < numberOfVolumes; i++ {
			Inst().CustomAppConfig["vdbench-multi-dep-multi-volume"] = scheduler.AppConfig{
				PvcStart:        i,
				PvcEnd:          i + 1,
				DeploymentCount: i,
			}
			err = Inst().S.RescanSpecs(Inst().SpecDir, Inst().V.String())
			log.FailOnError(err, "Failed to rescan specs from %s for storage provider %s with claim count %v", Inst().SpecDir, Inst().V.String(), numberOfVolumes)
			appContexts := ScheduleApplicationsOnNamespace(namespace, TaskNamePrefix)
			for _, appCtx := range appContexts {
				appCtx.ReadinessTimeout = AppReadinessTimeout
				scheduledAppContexts = append(scheduledAppContexts, appCtx)
			}
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
		sizePerVolume              float64
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
		totalNoOfScheduleBackups = 5
		backupLocationMap = make(map[string]string)
		sizePerVolume = 1
		ctx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		log.InfoD("Switching cluster context to destination cluster")
		err = SetDestinationKubeConfig()
		log.FailOnError(err, "Switching context to destination cluster failed")
		log.InfoD("Deploy applications on destination cluster")
		numberOfVolumes, _ = strconv.Atoi(GetEnv(VolCountForParallelDelete, "20"))
		log.InfoD("The number of PVC to be deployed are %v", numberOfVolumes)
		appList := Inst().AppList
		defer func() {
			Inst().AppList = appList
		}()
		namespace := fmt.Sprintf("multiple-volume-pvc-%s", RandomString(6))
		bkpNamespaces = append(bkpNamespaces, namespace)
		Inst().AppList = []string{"vdbench-multi-dep-multi-volume"}
		for i := 0; i < numberOfVolumes; i++ {
			Inst().CustomAppConfig["vdbench-multi-dep-multi-volume"] = scheduler.AppConfig{
				PvcStart:        i,
				PvcEnd:          i + 1,
				DeploymentCount: i,
			}
			err = Inst().S.RescanSpecs(Inst().SpecDir, Inst().V.String())
			log.FailOnError(err, "Failed to rescan specs from %s for storage provider %s with claim count %v", Inst().SpecDir, Inst().V.String(), numberOfVolumes)
			appContexts := ScheduleApplicationsOnNamespace(namespace, TaskNamePrefix)
			for _, appCtx := range appContexts {
				appCtx.ReadinessTimeout = AppReadinessTimeout
				scheduledAppContexts = append(scheduledAppContexts, appCtx)
			}
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
				return "", false, nil
			}
			_, err = DoRetryWithTimeoutWithGinkgoRecover(noOfScheduleBackups, 5*time.Hour, 15*time.Minute)
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

// This test case calculates the time taken to delete all schedule backup with heavy load with addition and deletion of PVC while schedule is going on
var _ = Describe("{BackupDeletionWithDynamicPVCGeneration}", Label(TestCaseLabelsMap[BackupDeletionWithDynamicPVCGeneration]...), func() {
	var (
		numberOfVolumes              int
		totalIteration               int
		noOfDeploymentToAddInBetween int
		err                          error
		cloudCredUID                 string
		cloudCredName                string
		backupLocationUID            string
		periodicSchedulePolicyName   string
		periodicSchedulePolicyUid    string
		scheduleName                 string
		clusterUid                   string
		backupLocation               string
		sizePerVolume                float64
		errors                       []string
		bkpNamespaces                []string
		appList                      []string
		controlChannel               chan string
		mu                           sync.Mutex
		wg                           sync.WaitGroup
		errorGroup                   *errgroup.Group
		ctx                          context.Context
		volumeList                   []*volume.Volume
		backupLocationMap            map[string]string
		scheduledAppContexts         []*scheduler.Context
		clusterStatus                api.ClusterInfo_StatusInfo_Status
	)

	JustBeforeEach(func() {
		sizePerVolume = 1
		noOfDeploymentToAddInBetween = 1
		backupLocationMap = make(map[string]string)
		StartPxBackupTorpedoTest("BackupDeletionWithDynamicPVCGeneration", "Calculates the time taken to delete schedule backup with addition and deletion of PVC while schedule is going on", nil, 301292, Sagrawal, Q3FY25)
		bkpNamespaces = make([]string, 0)
		ctx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		log.InfoD("Switching cluster context to destination cluster")
		err = SetDestinationKubeConfig()
		log.FailOnError(err, "Switching context to destination cluster failed")
		log.InfoD("Deploy applications on destination cluster")
		numberOfVolumes, _ = strconv.Atoi(GetEnv(VolCountForParallelDelete, "20"))
		log.InfoD("The number of PVC to be deployed are %v", numberOfVolumes)
		appList = Inst().AppList
		namespace := fmt.Sprintf("multiple-volume-add-delete-pvc-%s", RandomString(6))
		bkpNamespaces = append(bkpNamespaces, namespace)
		Inst().AppList = []string{"vdbench-multi-dep-multi-volume"}
		totalIteration = numberOfVolumes
		for i := 0; i < totalIteration; i++ {
			Inst().CustomAppConfig["vdbench-multi-dep-multi-volume"] = scheduler.AppConfig{
				PvcStart:        i,
				PvcEnd:          i + 1,
				DeploymentCount: i,
			}
			err = Inst().S.RescanSpecs(Inst().SpecDir, Inst().V.String())
			log.FailOnError(err, "Failed to rescan specs from %s for storage provider %s with claim count %v", Inst().SpecDir, Inst().V.String(), numberOfVolumes)
			appContexts := ScheduleApplicationsOnNamespace(namespace, TaskNamePrefix)
			for _, appCtx := range appContexts {
				appCtx.ReadinessTimeout = AppReadinessTimeout
				scheduledAppContexts = append(scheduledAppContexts, appCtx)
			}
		}
	})

	It("Calculates the time taken to delete all schedule backup with heavy load with addition & deletion of PVC while schedule is going on", func() {

		Step("Validate applications and verifying the size usage of all the pvc", func() {
			log.InfoD("Validate applications and verifying the size usage of all the pvc")
			controlChannel, errorGroup = ValidateApplicationsStartData(scheduledAppContexts, ctx)
			volumeList, err = GetVolumeFromContexts(scheduledAppContexts)
			dash.VerifyFatal(err, nil, "Fetching the volume of the deployed namespace")
			for _, vol := range volumeList {
				// Ignoring first volume as the I/O are continuously getting deleted and re-written for this PVC
				if vol.Name != "vdbench-pvc" {
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
			dash.VerifyFatal(len(errors), 0, fmt.Sprintf("verifying application volume size: %s", strings.Join(errors, "; ")))
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
				log.FailOnError(err, "Switching context to destination cluster")
			}()
			log.InfoD("Creating a schedule policy")
			periodicSchedulePolicyName = fmt.Sprintf("parallel-delete-%v", RandomString(5))
			periodicSchedulePolicyUid = uuid.New()
			periodicSchedulePolicyInfo := Inst().Backup.CreateIntervalSchedulePolicy(24, 15, 6)
			err := Inst().Backup.BackupSchedulePolicy(periodicSchedulePolicyName, periodicSchedulePolicyUid, BackupOrgID, periodicSchedulePolicyInfo)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of periodic schedule policy of interval 15 minutes named [%s]", periodicSchedulePolicyName))
			periodicSchedulePolicyUid, err = Inst().Backup.GetSchedulePolicyUid(BackupOrgID, ctx, periodicSchedulePolicyName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching uid of periodic schedule policy named [%s]", periodicSchedulePolicyName))
		})

		Step("Creating schedule backups on destination cluster", func() {
			log.InfoD("Creating schedule backup on destination cluster")
			scheduleName = fmt.Sprintf("%s-%v", BackupNamePrefix, RandomString(5))
			labelSelectors := make(map[string]string)
			_, err := CreateScheduleBackupWithValidation(ctx, scheduleName, DestinationClusterName, clusterUid, backupLocation, backupLocationUID, scheduledAppContexts, labelSelectors, BackupOrgID, "", "", "", "", periodicSchedulePolicyName, periodicSchedulePolicyUid)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of schedule backup on destination cluster with schedule name [%s]", scheduleName))
		})

		Step("Adding and deleting few PVC while schedule backup is going on in background", func() {
			log.InfoD("Adding and deleting few PVC while schedule backup is going on in background for %v time", DynamicPvcGenerationTimeOut*time.Hour)
			startTime := time.Now()
			dynamicPvcGeneration := func() (interface{}, bool, error) {
				if time.Since(startTime) >= DynamicPvcGenerationTime*time.Minute {
					return "", false, nil
				}
				log.InfoD("Adding %v PVC while schedule backup is going on", noOfDeploymentToAddInBetween)
				scheduledAppContexts, err = AddNewPvcToDeployedApplication(totalIteration, totalIteration+noOfDeploymentToAddInBetween, scheduledAppContexts)
				if err != nil {
					return "", false, fmt.Errorf("adding PVC while schedule backup is going on: %v", err)
				}
				// Since we have added one more deployment, the no of iteration will increase by noOfDeploymentToAddInBetween
				totalIteration += noOfDeploymentToAddInBetween
				log.InfoD("Deleting few PVC while schedule backup is going on")
				// Getting a random number to choose a random deployment to be deleted
				randomNumber := GenerateRandomNumber(totalIteration, 1)
				destClusterConfigPath, err := GetDestinationClusterConfigPath()
				if err != nil {
					return "", false, fmt.Errorf("getting kubeconfig path for destination cluster %v: %v", destClusterConfigPath, err)
				}
				// Choosing the random PVC
				randomPvc, err := exec.Command("sh", "-c", fmt.Sprintf("kubectl --kubeconfig=%v get pvc -n %v --no-headers | sed -n '%vp' | awk '{print $1}'", destClusterConfigPath, bkpNamespaces[0], randomNumber)).Output()
				if err != nil {
					return "", false, fmt.Errorf("executing cli to get a random PVC from the deployed application:%v", err)
				}

				pvcName := string(bytes.TrimSpace(randomPvc))
				log.InfoD("The pvc to be deleted is %v", pvcName)
				err = DeleteRandomPvcFromDeployedApplication(pvcName, scheduledAppContexts)
				if err != nil {
					return "", false, fmt.Errorf("deleting few PVC while schedule backup is going on in backgroud:%v", err)
				}
				return "", true, fmt.Errorf("dynamic PVC Genertion is still going on")
			}
			_, err = task.DoRetryWithTimeout(dynamicPvcGeneration, DynamicPvcGenerationTimeOut*time.Minute, DynamicPvcGenerationRetryTime*time.Minute)
			dash.VerifyFatal(err, nil, "Adding and deleting PVC while backup is going on")
			log.InfoD("Verifying the status of the latest schedule backup after dynamically generating the PVCs")
			allScheduleBackupNames, err := Inst().Backup.GetAllScheduleBackupNames(ctx, scheduleName, BackupOrgID)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Getting schedule backup for schedule %v", scheduleName))
			log.InfoD("%v schedule backups are %v", len(allScheduleBackupNames), allScheduleBackupNames)

			log.InfoD("Verify the latest backup status %v", allScheduleBackupNames[len(allScheduleBackupNames)-1])
			err = BackupSuccessCheck(allScheduleBackupNames[len(allScheduleBackupNames)-1], BackupOrgID, MaxWaitPeriodForBackupCompletionInMinutes*time.Minute, 30*time.Second, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying schedule backup %v status", allScheduleBackupNames[len(allScheduleBackupNames)-1]))
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
			dash.VerifyFatal(err, nil, fmt.Sprintf("Time take to delete all schedule backups of schedule %v is %v", scheduleName, TimeTakenToDeleteAllScheduleBackups))
		})
	})

	JustAfterEach(func() {
		defer func() {
			Inst().AppList = appList
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
