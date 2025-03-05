package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/pure-px/torpedo/drivers"
	"github.com/pure-px/torpedo/drivers/backup"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
	"golang.org/x/sync/errgroup"
	"os"
	"strings"
	"sync"
)

// This testcase covers all types of backup like pxd, csi, csi+ offload, direct kdmp and restore
var _ = Describe("{DifferentTypesOfBackupAndRestore}", Label(TestCaseLabelsMap[DifferentTypesOfBackupAndRestore]...), func() {
	var (
		wg                    sync.WaitGroup
		controlChannel        chan string
		errorGroup            *errgroup.Group
		cloudCredName         string
		cloudCredUID          string
		srcClusterUid         string
		backupLocationName    string
		backupLocationUID     string
		nfsBackupLocationName string
		nfsBackupLocationUID  string
		s3BackupLocationName  string
		s3BackupLocationUID   string
		restoreName           string
		preRuleUid            string
		postRuleUid           string
		preRuleName           string
		postRuleName          string
		providers             []string
		appList               []string
		allBackupTypes        []string
		ListOfBackups         []string
		ListOfRestore         []string
		preRuleAppMap         map[string]string
		postRuleAppMap        map[string]string
		backupLocationMap     map[string]string
		AppNamespaceMapping   map[string]string
		labelSelectors        map[string]string
		scheduledAppContexts  []*scheduler.Context
		AppContextsMapping    map[string]*scheduler.Context
		srcClusterStatus      api.ClusterInfo_StatusInfo_Status
	)
	// postgres-backup will be used for pxd backup , postgres-fada will be used for csi,csi+offload,kdmp backup
	appList = []string{"postgres-fada", "postgres-backup"}
	allBackupTypes = []string{"pxd", "direct_kdmp", "native_csi", "csi_offload_s3"}
	AppNamespaceMapping = make(map[string]string)
	backupLocationMap = make(map[string]string)
	preRuleAppMap = make(map[string]string)
	postRuleAppMap = make(map[string]string)
	AppContextsMapping = make(map[string]*scheduler.Context)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("DifferentTypesOfBackupAndRestore", "This testcase covers different types of backup and restore", nil, 0, Sagrawal, Q1FY26)
		log.InfoD("Deploying applications with storage as pure block")
		pipelineAppList := Inst().AppList
		defer func() {
			Inst().AppList = pipelineAppList
		}()

		log.InfoD("Scheduling applications")
		for _, app := range appList {
			Inst().AppList = []string{app}
			// If else loop is used here so that the app deployment becomes independent of what storage driver is passed from the pipeline
			if app == "postgres-fada" {
				err := Inst().S.RescanSpecs(Inst().SpecDir, "pure")
				log.FailOnError(err, "Failed while rescanning specs")
			} else {
				err := Inst().S.RescanSpecs(Inst().SpecDir, "pxd")
				log.FailOnError(err, "Failed while rescanning specs")
			}
			scheduledAppContexts = make([]*scheduler.Context, 0)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
				appContexts := ScheduleApplications(taskName)
				for _, appCtx := range appContexts {
					scheduledAppContexts = append(scheduledAppContexts, appCtx)
					namespace := GetAppNamespace(appCtx, taskName)
					AppContextsMapping[app] = appCtx
					AppNamespaceMapping[app] = namespace
				}
			}
		}
		log.InfoD("The AppNamespaceMapping is %v", AppNamespaceMapping)
	})

	It("Creating different types of backup and restore", func() {
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")

		Step("Validating the deployed applications", func() {
			log.InfoD("Validating the deployed applications")
			for _, appctx := range AppContextsMapping {
				controlChannel, errorGroup = ValidateApplicationsStartData([]*scheduler.Context{appctx}, ctx)
			}
		})

		Step("Creating rules for backup", func() {
			log.InfoD("Creating pre rule for deployed apps")
			for i := 0; i < len(appList); i++ {
				preRuleStatus, ruleName, err := Inst().Backup.CreateRuleForBackup(appList[i], BackupOrgID, "pre")
				log.FailOnError(err, "Creating pre rule for deployed app [%s] failed", appList[i])
				dash.VerifyFatal(preRuleStatus, true, fmt.Sprintf("Creating pre rule for app %v", appList[i]))
				if ruleName != "" {
					preRuleAppMap[appList[i]] = ruleName
				}
			}
			log.InfoD("Creating post rule for deployed apps")
			for i := 0; i < len(appList); i++ {
				postRuleStatus, ruleName, err := Inst().Backup.CreateRuleForBackup(appList[i], BackupOrgID, "post")
				log.FailOnError(err, "Creating post rule for deployed app [%s] failed", appList[i])
				dash.VerifyFatal(postRuleStatus, true, fmt.Sprintf("Creating post rule for app %v", appList[i]))
				if ruleName != "" {
					postRuleAppMap[appList[i]] = ruleName
				}
			}
		})

		Step("Creating backup location and cloud setting", func() {
			log.InfoD("Creating backup location and cloud setting")
			providers = GetBackupProviders()
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, RandomString(6))
				cloudCredUID = uuid.New()
				backupLocationName = fmt.Sprintf("%s-bl-%v", provider, RandomString(5))
				backupLocationUID = uuid.New()
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocation(provider, backupLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocationName))
				backupLocationMap[backupLocationUID] = backupLocationName
				if provider != drivers.ProviderNfs {
					nfsBackupLocationName = fmt.Sprintf("%s-%v", "nfs", RandomString(5))
					nfsBackupLocationUID = uuid.New()
					err = CreateNFSBackupLocation(nfsBackupLocationName, nfsBackupLocationUID, BackupOrgID, "", getGlobalBucketName(provider), true)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creating NFS backup location [%s]", nfsBackupLocationName))
					backupLocationMap[nfsBackupLocationUID] = nfsBackupLocationName
				} else {
					// Creating cloud cred again because in case of NFS as provider, cloud cred will not be created above
					err = CreateCloudCredential("aws", cloudCredName, cloudCredUID, BackupOrgID, ctx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential:[%s] for org [%s] for s3 backup location", cloudCredName, BackupOrgID))
					s3BackupLocationName = fmt.Sprintf("%s-%v", "s3", RandomString(5))
					s3BackupLocationUID = uuid.New()
					err = CreateS3BackupLocation(s3BackupLocationName, s3BackupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of S3 backup location [%s]", s3BackupLocationName))
					backupLocationMap[s3BackupLocationUID] = s3BackupLocationName
				}
			}
			log.InfoD(" The backupLocationMap is %v", backupLocationMap)
		})

		Step("Registering source and destination clusters for backup", func() {
			log.InfoD("Registering source and destination clusters for backup")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creating source cluster %s and destination cluster %s", SourceClusterName, DestinationClusterName))
			srcClusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(srcClusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			srcClusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})
		// Here backup and restore is done together so that we need not set BACKUP_TYPE again for restore which is required during restore validation
		for _, backupType := range allBackupTypes {
			Step(fmt.Sprintf("Taking %s backup and restoring it", backupType), func() {
				log.InfoD("Taking %s backup and restoring it", backupType)
				err = os.Setenv("BACKUP_TYPE", backupType)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Setting BACKUP_TYPE as %s", backupType))
				appName := "postgres-fada"
				if backupType == "pxd" {
					appName = "postgres-backup"
				}

				if value, exists := preRuleAppMap[appName]; exists {
					preRuleUid, err = Inst().Backup.GetRuleUid(BackupOrgID, ctx, value)
					log.FailOnError(err, "Failed to get UID for rule %s", value)
					preRuleName = value
				}

				if value, exists := postRuleAppMap[appName]; exists {
					postRuleUid, err = Inst().Backup.GetRuleUid(BackupOrgID, ctx, value)
					log.FailOnError(err, "Failed to get UID for rule %s", value)
					postRuleName = value
				}

				formattedBackupType := strings.ReplaceAll(backupType, "_", "-")
				for backupLocationUID, backupLocationName := range backupLocationMap {
					backupName := fmt.Sprintf("%s-%s-%s-%s", BackupNamePrefix, formattedBackupType, backupLocationName, RandomString(5))
					ListOfBackups = append(ListOfBackups, backupName)
					err := CreateBackupWithValidation(ctx, backupName, SourceClusterName, backupLocationName, backupLocationUID, []*scheduler.Context{AppContextsMapping[appName]}, labelSelectors, BackupOrgID, srcClusterUid, preRuleName, preRuleUid, postRuleName, postRuleUid)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s] of namespace  [%s] in backup location %s", backupName, AppNamespaceMapping[appName], backupLocationName))
					restoreName = fmt.Sprintf("%s-%s-%v", RestoreNamePrefix, backupName, RandomString(4))
					restoredNameSpace := fmt.Sprintf("%s-%s-%s", AppNamespaceMapping[appName], formattedBackupType, "restored")
					log.InfoD(" The AppNamespaceMapping[appName] is %v", AppNamespaceMapping[appName])
					namespaceMapping := make(map[string]string)
					namespaceMapping[AppNamespaceMapping[appName]] = restoredNameSpace
					appContextsToBackup := FilterAppContextsByNamespace([]*scheduler.Context{AppContextsMapping[appName]}, []string{AppNamespaceMapping[appName]})
					err = CreateRestoreWithValidation(ctx, restoreName, backupName, namespaceMapping, make(map[string]string), SourceClusterName, srcClusterUid, BackupOrgID, appContextsToBackup)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Restoring  %s backups %s with Ns Mapping - %v", restoreName, backupName, namespaceMapping))
					ListOfRestore = append(ListOfRestore, restoreName)
				}
			})
		}
		log.Infof("The list of backups taken is: %v", ListOfBackups)
		log.Infof("The list of restores is: %v", ListOfRestore)

	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")

		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		log.InfoD("Deleting deployed applications")
		err = DestroyAppsWithData(scheduledAppContexts, opts, controlChannel, errorGroup)
		log.FailOnError(err, "Data validations failed")

		for _, backupName := range ListOfBackups {
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

		for _, restoreName := range ListOfRestore {
			wg.Add(1)
			go func(restoreName string) {
				defer GinkgoRecover()
				defer wg.Done()
				err = DeleteRestore(restoreName, BackupOrgID, ctx)
				dash.VerifySafely(err, nil, fmt.Sprintf("Deleting restore [%s]", restoreName))
			}(restoreName)
		}
		wg.Wait()

		for _, ruleName := range preRuleAppMap {
			err := Inst().Backup.DeleteRuleForBackup(BackupOrgID, ruleName)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting backup pre rules [%s]", ruleName))
		}

		for _, ruleName := range postRuleAppMap {
			err := Inst().Backup.DeleteRuleForBackup(BackupOrgID, ruleName)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting backup post rules [%s]", ruleName))
		}

		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})
