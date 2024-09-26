package tests

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	_ "github.com/rancher/norman/clientbase"
	_ "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"golang.org/x/sync/errgroup"
	"os"
	"strconv"
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
	)
	CloudCredUIDMap := make(map[string]string)
	backupLocationMap := make(map[string]string)
	timeStamp := time.Now().Unix()
	bkpNamespaces = make([]string, 0)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("TimeTakenToDeleteBackupWithHeavyLoad", "Calculates the time taken to delete a backup with heavy load", nil, 301290, Sagrawal, Q3FY25)
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
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, timeStamp)
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
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, timeStamp)
				backupLocation = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocation
				err := CreateBackupLocationWithContext(provider, backupLocation, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", ctx, true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocation))
			}
		})

		Step("Start backup of application to bucket from destination cluster", func() {
			log.InfoD("Start backup of application to bucket from destination cluster")
			backupName = fmt.Sprintf("%s-%v", BackupNamePrefix, time.Now().Unix())
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
