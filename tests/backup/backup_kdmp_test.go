package tests

import (
	"context"
	"fmt"
	"github.com/pure-px/sched-ops/task"
	"github.com/pure-px/torpedo/drivers"
	"github.com/pure-px/torpedo/drivers/node"
	"k8s.io/apimachinery/pkg/api/errors"
	"math/rand"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	"golang.org/x/sync/errgroup"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/pure-px/sched-ops/k8s/core"
	"github.com/pure-px/sched-ops/k8s/storage"
	"github.com/pure-px/torpedo/drivers/backup"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/drivers/volume/portworx/schedops"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
	storagev1 "k8s.io/api/storage/v1"
)

// This testcase verifies backup and restore of applications by excluding files and directories from mountPath.
var _ = Describe("{ExcludeDirectoryFileBackup}", Label(TestCaseLabelsMap[ExcludeDirectoryFileBackup]...), func() {
	var (
		backupName                    string
		scheduledAppContexts          []*scheduler.Context
		AppContextsMapping            = make(map[string]*scheduler.Context)
		namespace                     string
		bkpNamespaces                 = make([]string, 0)
		backupNames                   = make([]string, 0)
		restoreNames                  = make([]string, 0)
		scheduleNames                 = make([]string, 0)
		clusterUid                    string
		destClusterUid                string
		clusterStatus                 api.ClusterInfo_StatusInfo_Status
		restoreName                   string
		cloudCredName                 string
		cloudCredUID                  string
		backupLocationUID             string
		bkpLocationName               string
		numDeployments                int
		providers                     []string
		backupLocationMap             = make(map[string]string)
		labelSelectors                = make(map[string]string)
		storageClassExcludeFileDirMap = make(map[*storagev1.StorageClass][]string)
		mountPathExcludeFileDirMap    = make(map[string][]string)
		existingFileDirMountPathMap   = make(map[string][]string)
		fileListMountMap              = make(map[string][]string)
		dirListMountMap               = make(map[string][]string)
		masterDirFileList             = make(map[string][]string)
		finalFileList                 = make(map[string][]string)
		podScMountPathMap             = make(map[string]map[string]*storagev1.StorageClass)
		backupNamespaceMap            = make(map[string]string)
		restoredNamespaces            = make([]string, 0)
		excludeList                   string
		fileList                      []string
		dirList                       []string
		preRuleName                   string
		postRuleName                  string
		preRuleUid                    string
		postRuleUid                   string
		periodicSchedulePolicyName    string
		periodicSchedulePolicyUid     string
		testAppList                   []string
		mutex                         sync.Mutex
		wg                            sync.WaitGroup
	)
	JustBeforeEach(func() {
		numDeployments = 1
		providers = GetBackupProviders()

		StartPxBackupTorpedoTest("ExcludeDirectoryFileBackup", "Excludes mentioned directories or files from backed-up apps and restores them", nil, 93691, Ak, Q4FY24)

		log.InfoD(fmt.Sprintf("App list %v", Inst().AppList))
		scheduledAppContexts = make([]*scheduler.Context, 0)
		testAppList = []string{"vdbench-fbda-one-vol"}
		actualAppList := Inst().AppList
		defer func() {
			Inst().AppList = actualAppList
			err := Inst().S.RescanSpecs(Inst().SpecDir, Inst().V.String())
			log.FailOnError(err, "Failed while rescanning specs")
		}()
		Inst().AppList = testAppList
		err := Inst().S.RescanSpecs(Inst().SpecDir, Inst().V.String())
		log.FailOnError(err, "Failed to rescan specs from %s for storage provider %s", Inst().SpecDir, Inst().V.String())
		log.InfoD("Starting to deploy applications")
		for i := 0; i < numDeployments; i++ {
			log.InfoD(fmt.Sprintf("Iteration %v of deploying applications", i))
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = AppReadinessTimeout
				namespace = GetAppNamespace(ctx, taskName)
				scheduledAppContexts = append(scheduledAppContexts, ctx)
				AppContextsMapping[namespace] = ctx
				bkpNamespaces = append(bkpNamespaces, namespace)
			}
		}

	})
	It("Excludes directories or files From a Backup", func() {

		Step("Validating deployed applications", func() {
			log.InfoD("Validating deployed applications")
			ValidateApplications(scheduledAppContexts)
		})

		Step("Getting mountpath and associated storageClass for containers in deployed application", func() {
			log.InfoD("Getting mountpath associated storageClass for containers in deployed application")
			for _, namespace := range bkpNamespaces {
				pods, err := core.Instance().GetPods(namespace, nil)
				dash.VerifyFatal(err, nil, fmt.Sprintf("getting pods from namespace [%s] ", namespace))
				for _, pod := range pods.Items {
					scMountPathMap, err := schedops.GetContainerPVCMountMapWithSC(pod)
					dash.VerifyFatal(err, nil, fmt.Sprintf("getting storage class and mountpath mapping for pod [%s] ", pod.Name))
					podScMountPathMap[pod.Name] = scMountPathMap
				}
			}
		})

		Step("Fetch the existing directories and files within mountPath before writing files and directories", func() {
			log.InfoD("Fetch the existing directories and files within mountPath before writing files and directories")
			for _, namespace := range bkpNamespaces {
				pods, err := core.Instance().GetPods(namespace, nil)
				dash.VerifyFatal(err, nil, fmt.Sprintf("getting pods from namespace [%s] ", namespace))
				for _, pod := range pods.Items {
					containerPaths := schedops.GetContainerPVCMountMap(pod)
					existingFileDirList := make([]string, 0)
					for containerName, mountPaths := range containerPaths {
						for _, mountPath := range mountPaths {
							log.Infof(fmt.Sprintf("Fetch the existing directories and files within mountPath [%s] before writing files and directories", mountPath))
							existingFileList, existingDirList, err := FetchFilesAndDirectoriesFromPod(pod, containerName, mountPath, nil)
							existingFileDirList = append(existingFileDirList, existingFileList...)
							existingFileDirList = append(existingFileDirList, existingDirList...)
							existingFileDirMountPathMap[mountPath] = existingFileDirList
							dash.VerifyFatal(err, nil, fmt.Sprintf("fetching files and directory from mountpath [%s] for pod [%s]", mountPath, pod.Name))
						}
					}
				}
			}
		})

		Step("Create nested directories and files into container mountPath for applications", func() {
			log.InfoD("Create nested directories and files into container mountPath for applications")
			for _, namespace := range bkpNamespaces {
				pods, err := core.Instance().GetPods(namespace, nil)
				dash.VerifyFatal(err, nil, fmt.Sprintf("getting pods from namespace [%s] ", namespace))
				for _, pod := range pods.Items {
					containerPaths := schedops.GetContainerPVCMountMap(pod)
					for containerName, mountPaths := range containerPaths {
						for _, mountPath := range mountPaths {
							sc := podScMountPathMap[pod.Name][mountPath]
							DirectoryConfig := PodDirectoryConfig{
								BasePath:          mountPath,
								Depth:             100,
								Levels:            1,
								FilesPerDirectory: 100,
							}
							log.Infof(fmt.Sprintf("creating nested directories and files within mountPath [%s] with depth [%d] , level [%d] and FilesPerDirectory [%d]", DirectoryConfig.BasePath, DirectoryConfig.Depth, DirectoryConfig.Levels, DirectoryConfig.FilesPerDirectory))
							err = CreateNestedDirectoriesWithFilesInPod(pod, containerName, DirectoryConfig)
							dash.VerifyFatal(err, nil, fmt.Sprintf("creating nested directories and files at mountpath [%v] for pod [%v] in namespace [%v]", mountPath, pod.Name, pod.Namespace))
							log.Infof(fmt.Sprintf("Fetching files and directories from path [%s] by excluding existing directories", DirectoryConfig.BasePath))
							fileList, dirList, err = FetchFilesAndDirectoriesFromPod(pod, containerName, mountPath, existingFileDirMountPathMap[mountPath])
							dash.VerifyFatal(err, nil, fmt.Sprintf("fetching files and directory from mountpath [%s] for pod [%s]", mountPath, pod.Name))
							log.Infof(fmt.Sprintf("the list of files created in mountPath [%v] for pod [%v]: %v", mountPath, pod.Name, fileList))
							log.Infof(fmt.Sprintf("the list of directories created in mountPath [%v] for pod [%v]: %v", mountPath, pod.Name, dirList))
							fileListMountMap[mountPath] = fileList
							dirListMountMap[mountPath] = dirList
							masterDirFileList[mountPath] = append(masterDirFileList[mountPath], fileList...)
							masterDirFileList[mountPath] = append(masterDirFileList[mountPath], dirList...)
							log.Infof(fmt.Sprintf("creating files within mountPath [%s] with extensions", mountPath))
							fileConfig := PodDirectoryConfig{
								BasePath:          mountPath,
								FilesPerDirectory: 1,
								FileName:          "test.yaml",
							}
							_, err = CreateFilesInPodDirectory(pod, containerName, fileConfig)
							dash.VerifyFatal(err, nil, fmt.Sprintf("creating files within mountPath [%s] with extensions and add to exclude list", mountPath))
							storageClassExcludeFileDirMap[sc] = append(storageClassExcludeFileDirMap[sc], fileConfig.FileName)
							mountPathExcludeFileDirMap[mountPath] = append(mountPathExcludeFileDirMap[mountPath], fileConfig.FileName)
							log.Infof(fmt.Sprintf("creating file within mountPath [%s] with hidden type", mountPath))
							fileConfig = PodDirectoryConfig{
								BasePath:          mountPath,
								FilesPerDirectory: 1,
								FileName:          ".hiddenfile",
							}
							_, err = CreateFilesInPodDirectory(pod, containerName, fileConfig)
							dash.VerifyFatal(err, nil, fmt.Sprintf("creating files within mountPath [%s] with hidden type and add to exclude list", mountPath))
							storageClassExcludeFileDirMap[sc] = append(storageClassExcludeFileDirMap[sc], fileConfig.FileName)
							mountPathExcludeFileDirMap[mountPath] = append(mountPathExcludeFileDirMap[mountPath], fileConfig.FileName)
							log.Infof(fmt.Sprintf("creating file within mountPath [%s] with valid special chars ", mountPath))
							fileConfig = PodDirectoryConfig{
								BasePath:          mountPath,
								FilesPerDirectory: 1,
								FileName:          "myn@meisunkn*wn",
							}
							_, err = CreateFilesInPodDirectory(pod, containerName, fileConfig)
							dash.VerifyFatal(err, nil, fmt.Sprintf("creating files within mountPath [%s] valid special chars and add to exclude list", mountPath))
							storageClassExcludeFileDirMap[sc] = append(storageClassExcludeFileDirMap[sc], fileConfig.FileName)
							mountPathExcludeFileDirMap[mountPath] = append(mountPathExcludeFileDirMap[mountPath], fileConfig.FileName)

							log.Infof(fmt.Sprintf("creating file within mountPath [%s] with maximum name length (255 characters)", mountPath))
							fileConfig = PodDirectoryConfig{
								BasePath:          mountPath,
								FilesPerDirectory: 1,
								FileName:          fmt.Sprintf("%s.txt", RandomString(251)),
							}
							_, err = CreateFilesInPodDirectory(pod, containerName, fileConfig)
							dash.VerifyFatal(err, nil, fmt.Sprintf("creating files within mountPath [%s] with maximum name length (255 characters and add to exclude list)", mountPath))
							storageClassExcludeFileDirMap[sc] = append(storageClassExcludeFileDirMap[sc], fileConfig.FileName)
							mountPathExcludeFileDirMap[mountPath] = append(mountPathExcludeFileDirMap[mountPath], fileConfig.FileName)
							log.Infof(fmt.Sprintf("creating file within mountPath [%s] with hardlink and symbolic link and add to exclude list", mountPath))
							fileConfig = PodDirectoryConfig{
								BasePath:           mountPath,
								FilesPerDirectory:  1,
								FileName:           "linkFile.txt",
								CreateSymbolicLink: true,
								CreateHardLink:     true,
							}
							files, err := CreateFilesInPodDirectory(pod, containerName, fileConfig)
							dash.VerifyFatal(err, nil, fmt.Sprintf("creating file within mountPath [%s] with hardlink and symbolic link and add to exclude list", mountPath))
							storageClassExcludeFileDirMap[sc] = append(storageClassExcludeFileDirMap[sc], files...)
							mountPathExcludeFileDirMap[mountPath] = append(mountPathExcludeFileDirMap[mountPath], files...)
						}
					}
				}
			}
		})
		Step("Update KDMP config map on source cluster by formatting storage class name and random files and directories as a string", func() {
			log.InfoD("Update KDMP config map on source cluster by formatting storage class name and random files and directories as a string")
			for _, namespace := range bkpNamespaces {
				pods, err := core.Instance().GetPods(namespace, nil)
				dash.VerifyFatal(err, nil, fmt.Sprintf("getting pods from namespace [%s] ", namespace))
				for _, pod := range pods.Items {
					containerPaths := schedops.GetContainerPVCMountMap(pod)
					for _, mountPaths := range containerPaths {
						for _, mountPath := range mountPaths {
							excludeFileDirList := make([]string, 0)
							log.Infof(fmt.Sprintf("Fetch some random directories from created list %v", dirListMountMap[mountPath]))
							randomDirs, err := GetRandomSubset(dirListMountMap[mountPath], 50)
							dash.VerifyFatal(err, nil, fmt.Sprintf("Getting random directories from the list"))
							log.Infof(fmt.Sprintf("the list of directories randomly selected from mountPath- %v : %v", mountPath, randomDirs))
							log.Infof(fmt.Sprintf("Fetch some random files from created list %v", fileListMountMap[mountPath]))
							randomFiles, err := GetRandomSubset(fileListMountMap[mountPath], 100)
							dash.VerifyFatal(err, nil, fmt.Sprintf("Getting random files from the list"))
							log.Infof(fmt.Sprintf("the list of files randomly selected from mountPath- %v : %v", mountPath, randomFiles))
							excludeFileDirList = append(excludeFileDirList, randomDirs...)
							excludeFileDirList = append(excludeFileDirList, randomFiles...)
							sc := podScMountPathMap[pod.Name][mountPath]
							storageClassExcludeFileDirMap[sc] = append(storageClassExcludeFileDirMap[sc], excludeFileDirList...)
							mountPathExcludeFileDirMap[mountPath] = append(mountPathExcludeFileDirMap[mountPath], excludeFileDirList...)
						}
					}
				}
			}
			log.Infof("create formatted string with storage class name and exclude file and directories list")
			excludeList = GetExcludeFileListValue(storageClassExcludeFileDirMap)
			err := UpdateKDMPConfigMap("KDMP_EXCLUDE_FILE_LIST", excludeList)
			dash.VerifyFatal(err, nil, fmt.Sprintf("updating KDMP config map"))
		})

		Step("Creating backup location and cloud setting", func() {
			log.InfoD("Creating backup location and cloud setting")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
				bkpLocationName = fmt.Sprintf("%s-%s-bl", provider, getGlobalBucketName(provider))
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = bkpLocationName
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocation(provider, bkpLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", bkpLocationName))
			}
		})
		Step("Registering cluster for backup", func() {
			log.InfoD("Registering cluster for backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
			destClusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))
		})

		Step("Create new storage class on destination cluster with similar spec as source cluster", func() {
			log.InfoD("Create new storage class on destination cluster with similar spec as source cluster")
			defer func() {
				err := SetSourceKubeConfig()
				log.FailOnError(err, "Unable to switch context to source cluster [%s]", SourceClusterName)
			}()
			log.InfoD("Switching cluster context to destination cluster")
			err := SetDestinationKubeConfig()
			log.FailOnError(err, "Failed to set destination kubeconfig")

			for _, scMountPathMap := range podScMountPathMap {
				for _, storageClass := range scMountPathMap {
					isScpresent, err := IsStorageClassPresent(storageClass.Name)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Checking storageClass %s present on cluster", storageClass.Name))
					if !isScpresent {
						v1obj := metaV1.ObjectMeta{
							Name: storageClass.Name,
						}
						scObj := storagev1.StorageClass{
							ObjectMeta:           v1obj,
							Provisioner:          storageClass.Provisioner,
							Parameters:           storageClass.Parameters,
							ReclaimPolicy:        storageClass.ReclaimPolicy,
							VolumeBindingMode:    storageClass.VolumeBindingMode,
							MountOptions:         storageClass.MountOptions,
							AllowVolumeExpansion: storageClass.AllowVolumeExpansion,
						}
						_, err = storage.Instance().CreateStorageClass(&scObj)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Creating new storage class %v on destination cluster %s", storageClass.Name, DestinationClusterName))
					} else {
						log.Infof(fmt.Sprintf("storageClass %s already present on cluster , hence skipping creation", storageClass.Name))
					}
				}
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

		Step(fmt.Sprintf("Create schedule policy for backup schedules"), func() {
			log.InfoD("Create schedule policy for backup schedules")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			periodicSchedulePolicyName = fmt.Sprintf("%s-%s", "periodic", RandomString(5))
			periodicSchedulePolicyUid = uuid.New()
			periodicSchedulePolicyInterval := int64(15)
			err = CreateBackupScheduleIntervalPolicy(5, periodicSchedulePolicyInterval, 5, periodicSchedulePolicyName, periodicSchedulePolicyUid, BackupOrgID, ctx, false, false)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of periodic schedule policy of interval [%v] minutes named [%s] ", periodicSchedulePolicyInterval, periodicSchedulePolicyName))

		})

		Step("Taking manual backup of namespaces with rules", func() {
			log.InfoD(fmt.Sprintf("Taking manual backup of namespaces with rules"))
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, namespace := range bkpNamespaces {
				backupName = fmt.Sprintf("%s-%v", BackupNamePrefix, RandomString(3))
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
				err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, bkpLocationName, backupLocationUID, appContextsToBackup, labelSelectors, BackupOrgID, clusterUid, preRuleName, preRuleUid, postRuleName, postRuleUid)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				backupNames = append(backupNames, backupName)
				backupNamespaceMap[backupName] = namespace
			}
		})

		Step("Taking schedule backup of namespaces with rules", func() {
			log.InfoD(fmt.Sprintf("Taking schedule backup of namespaces with rules"))
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, namespace := range bkpNamespaces {
				scheduleName := fmt.Sprintf("%s-sch-rules-%s", BackupNamePrefix, RandomString(3))
				log.InfoD("Creating a schedule backup of namespace [%s] without pre and post exec rules", namespace)
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
				scheduleBackupName, err := CreateScheduleBackupWithValidation(ctx, scheduleName, SourceClusterName, clusterUid, bkpLocationName, backupLocationUID, appContextsToBackup,
					labelSelectors, BackupOrgID, "", "", "", "", periodicSchedulePolicyName, periodicSchedulePolicyUid, false)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of backup [%s]", scheduleBackupName))
				err = SuspendBackupSchedule(scheduleName, periodicSchedulePolicyName, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Suspending Backup Schedule [%s] ", scheduleName))
				backupNames = append(backupNames, scheduleBackupName)
				scheduleNames = append(scheduleNames, scheduleName)
				backupNamespaceMap[scheduleBackupName] = namespace
			}
		})

		Step("Taking restore of backups created", func() {
			log.InfoD("Taking restore of backups created")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			restoreSingleNSBackupInVariousWaysTask := func(index int, backupName string) {
				restoreConfigs := []struct {
					namePrefix          string
					namespaceMapping    map[string]string
					storageClassMapping map[string]string
					replacePolicy       ReplacePolicyType
				}{
					{
						"test-custom-restore-single-ns",
						map[string]string{backupNamespaceMap[backupName]: fmt.Sprintf("cust1-%s-%d", backupNamespaceMap[backupName], index)},
						make(map[string]string),
						ReplacePolicyRetain,
					},
					{
						"test-replace-restore-single-ns",
						map[string]string{backupNamespaceMap[backupName]: fmt.Sprintf("cust1-rep-%s-%d", backupNamespaceMap[backupName], index)},
						make(map[string]string),
						ReplacePolicyDelete,
					},
				}
				for _, config := range restoreConfigs {
					restoreName := fmt.Sprintf("%s-%s", config.namePrefix, RandomString(3))
					log.InfoD("Restoring backup [%s] in cluster [%s] with restore [%s] and namespace mapping %v", backupName, DestinationClusterName, restoreName, config.namespaceMapping)
					if config.replacePolicy == ReplacePolicyRetain {
						appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{backupNamespaceMap[backupName]})
						restoredNamespaces = append(restoredNamespaces, config.namespaceMapping[backupNamespaceMap[backupName]])
						err = CreateRestoreWithValidation(ctx, restoreName, backupName, config.namespaceMapping, config.storageClassMapping, DestinationClusterName, destClusterUid, BackupOrgID, appContextsToBackup)
					} else if config.replacePolicy == ReplacePolicyDelete {
						restoredNamespaces = append(restoredNamespaces, config.namespaceMapping[backupNamespaceMap[backupName]])
						err = CreateRestoreWithReplacePolicy(restoreName, backupName, config.namespaceMapping, DestinationClusterName, BackupOrgID, ctx, config.storageClassMapping, config.replacePolicy)
					}
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying restoration [%s] of single namespace backup [%s] in cluster", restoreName, backupName))
					restoreNames = SafeAppend(&mutex, restoreNames, restoreName).([]string)
				}
			}
			_ = TaskHandler(backupNames, restoreSingleNSBackupInVariousWaysTask, Sequential)
		})

		Step("List files and directories from the mount path after restore and verify excluded items are not present,iteration 1", func() {
			log.InfoD("List files and directories from the mount path after restore and verify excluded items are not present,iteration 1")

			defer func() {
				err := SetSourceKubeConfig()
				log.FailOnError(err, "Unable to switch context to source cluster [%s]", SourceClusterName)
			}()

			err := SetDestinationKubeConfig()
			log.FailOnError(err, "Switching context to destination cluster failed")

			for _, restoredNamespace := range restoredNamespaces {
				pods, err := core.Instance().GetPods(restoredNamespace, nil)
				dash.VerifyFatal(err, nil, fmt.Sprintf("getting pods from namespace [%s] ", restoredNamespace))
				for _, pod := range pods.Items {
					log.Infof(fmt.Sprintf("verifying the files and directories for pod [%s] in restored namespace [%s] ", pod.Name, restoredNamespace))
					containerPaths := schedops.GetContainerPVCMountMap(pod)
					for containerName, mountPaths := range containerPaths {
						for _, mountPath := range mountPaths {
							restoredCombinedList := make([]string, 0)
							fileList, dirList, err := FetchFilesAndDirectoriesFromPod(pod, containerName, mountPath, existingFileDirMountPathMap[mountPath])
							dash.VerifyFatal(err, nil, fmt.Sprintf("fetching files and directory from mountpath [%s] for pod [%s]", mountPath, pod.Name))
							log.Infof(fmt.Sprintf("The list of files created in mountPath [%v] for pod [%v]: %v", mountPath, pod.Name, fileList))
							log.Infof(fmt.Sprintf("The list of directories created in mountPath [%v] for pod [%v]: %v", mountPath, pod.Name, dirList))
							restoredCombinedList = append(restoredCombinedList, fileList...)
							restoredCombinedList = append(restoredCombinedList, dirList...)
							log.Infof(fmt.Sprintf("the list of combined directories and files after restore: %v", restoredCombinedList))
							for _, item := range mountPathExcludeFileDirMap[mountPath] {
								if item != "" {
									if !IsPresent(restoredCombinedList, item) {
										log.Infof(fmt.Sprintf("the item file/directory [%s] is not present in the mountPath[%s] for pod [%s] in namespace [%s]", item, mountPath, pod.Name, restoredNamespace))
									} else {
										err := fmt.Errorf("the item file/directory[%s] is still present in mountPath [%s] for pod [%s] in namespace [%s]", item, mountPath, pod.Name, restoredNamespace)
										dash.VerifyFatal(err, nil, fmt.Sprintf("%v", err))
									}
								}
							}
						}
					}
				}
			}
		})

		Step("Create second iteration of nested directories and files into container mountPath for applications", func() {
			log.InfoD("Create second iteration of  nested directories and files into container mountPath for applications")
			for _, namespace := range bkpNamespaces {
				pods, err := core.Instance().GetPods(namespace, nil)
				dash.VerifyFatal(err, nil, fmt.Sprintf("getting pods from namespace [%s] ", namespace))
				for _, pod := range pods.Items {
					containerPaths := schedops.GetContainerPVCMountMap(pod)
					for containerName, mountPaths := range containerPaths {
						for _, mountPath := range mountPaths {
							DirectoryConfig := PodDirectoryConfig{
								BasePath:          mountPath,
								Depth:             100,
								Levels:            1,
								FilesPerDirectory: 100,
							}
							log.Infof(fmt.Sprintf("creating nested directories and files within mountPath [%s] with depth [%d] , level [%d] and FilesPerDirectory [%d]", DirectoryConfig.BasePath, DirectoryConfig.Depth, DirectoryConfig.Levels, DirectoryConfig.FilesPerDirectory))
							err = CreateNestedDirectoriesWithFilesInPod(pod, containerName, DirectoryConfig)
							dash.VerifyFatal(err, nil, fmt.Sprintf("creating nested directories and files at mountpath [%v] for pod [%v] in namespace [%v]", mountPath, pod.Name, pod.Namespace))
							log.Infof(fmt.Sprintf("Fetching files and directories from path [%s] by excluding existing directories", DirectoryConfig.BasePath))
							fileList, dirList, err := FetchFilesAndDirectoriesFromPod(pod, containerName, mountPath, append(existingFileDirMountPathMap[mountPath], masterDirFileList[mountPath]...))
							dash.VerifyFatal(err, nil, fmt.Sprintf("fetching files and directory from mountpath [%s] for pod [%s]", mountPath, pod.Name))
							log.Infof(fmt.Sprintf("the list of files created in mountPath [%v] for pod [%v]: %v", mountPath, pod.Name, fileList))
							log.Infof(fmt.Sprintf("the list of directories created in mountPath [%v] for pod [%v]: %v", mountPath, pod.Name, dirList))
							fileListMountMap[mountPath] = fileList
							dirListMountMap[mountPath] = dirList
							masterDirFileList[mountPath] = append(masterDirFileList[mountPath], fileList...)
							masterDirFileList[mountPath] = append(masterDirFileList[mountPath], dirList...)
						}
					}
				}
			}
		})

		Step("Update KDMP config map by selecting random files and directories from the second iteration", func() {
			log.InfoD("Update KDMP config map by selecting random files and directories from the second iteration")
			for _, namespace := range bkpNamespaces {
				pods, err := core.Instance().GetPods(namespace, nil)
				dash.VerifyFatal(err, nil, fmt.Sprintf("getting pods from namespace [%s] ", namespace))
				for _, pod := range pods.Items {
					containerPaths := schedops.GetContainerPVCMountMap(pod)
					for _, mountPaths := range containerPaths {
						for _, mountPath := range mountPaths {
							excludeFileDirList := make([]string, 0)
							log.Infof(fmt.Sprintf("Fetch some random directories from created list %v", dirListMountMap[mountPath]))
							randomDirs, err := GetRandomSubset(dirListMountMap[mountPath], 100)
							dash.VerifyFatal(err, nil, fmt.Sprintf("Getting random directories from the list"))
							log.Infof(fmt.Sprintf("the list of directories randomly selected from mountPath- %v : %v", mountPath, randomDirs))
							log.Infof(fmt.Sprintf("Fetch some random files from created list %v", fileListMountMap[mountPath]))
							randomFiles, err := GetRandomSubset(fileListMountMap[mountPath], 500)
							dash.VerifyFatal(err, nil, fmt.Sprintf("Getting random files from the list"))
							log.Infof(fmt.Sprintf("the list of files randomly selected from mountPath- %v : %v", mountPath, randomFiles))
							excludeFileDirList = append(excludeFileDirList, randomDirs...)
							excludeFileDirList = append(excludeFileDirList, randomFiles...)
							sc := podScMountPathMap[pod.Name][mountPath]
							storageClassExcludeFileDirMap[sc] = append(storageClassExcludeFileDirMap[sc], excludeFileDirList...)
							mountPathExcludeFileDirMap[mountPath] = append(mountPathExcludeFileDirMap[mountPath], excludeFileDirList...)
						}
					}
				}
			}
			log.Infof("create new formatted string with storage class name and exclude file ,directories list ")
			excludeList = GetExcludeFileListValue(storageClassExcludeFileDirMap)
			err := UpdateKDMPConfigMap("KDMP_EXCLUDE_FILE_LIST", excludeList)
			dash.VerifyFatal(err, nil, fmt.Sprintf("updating KDMP config map"))
		})

		Step("Taking manual backup of namespaces with rules ,iteration 2", func() {
			log.InfoD(fmt.Sprintf("Taking manual backup of namespaces with rules, iteration 2"))
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			backupNames := make([]string, 0)
			for _, namespace := range bkpNamespaces {
				backupName = fmt.Sprintf("%s-%v", BackupNamePrefix, RandomString(3))
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
				err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, bkpLocationName, backupLocationUID, appContextsToBackup, labelSelectors, BackupOrgID, clusterUid, preRuleName, preRuleUid, postRuleName, postRuleUid)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				backupNames = append(backupNames, backupName)
				backupNamespaceMap[backupName] = namespace
			}
		})

		Step("Taking restore of backups after ,iteration 2", func() {
			log.InfoD("Taking restore of backups after ,iteration 2")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			restoredNamespaces = make([]string, 0)
			for _, backupName := range backupNames {
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{backupNamespaceMap[backupName]})
				restoreName = fmt.Sprintf("%s-%v", RestoreNamePrefix, RandomString(3))
				restoreNamespace := "custom2-" + backupNamespaceMap[backupName]
				namespaceMapping := map[string]string{backupNamespaceMap[backupName]: restoreNamespace}
				restoredNamespaces = append(restoredNamespaces, restoreNamespace)
				err = CreateRestoreWithValidation(ctx, restoreName, backupName, namespaceMapping, make(map[string]string), DestinationClusterName, destClusterUid, BackupOrgID, appContextsToBackup)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s]", restoreName))
				restoreNames = append(restoreNames, restoreName)
			}
		})

		Step("List files and directories from the mount path after restore and verify excluded items are not present ,iteration 2", func() {
			log.InfoD("List files and directories from the mount path after restore and verify excluded items are not present ,iteration 2")

			defer func() {
				err := SetSourceKubeConfig()
				log.FailOnError(err, "Unable to switch context to source cluster [%s]", SourceClusterName)
			}()

			err := SetDestinationKubeConfig()
			log.FailOnError(err, "Switching context to destination cluster failed")

			for _, restoredNamespace := range restoredNamespaces {
				pods, err := core.Instance().GetPods(restoredNamespace, nil)
				dash.VerifyFatal(err, nil, fmt.Sprintf("getting pods from namespace [%s] ", restoredNamespace))
				for _, pod := range pods.Items {
					log.Infof(fmt.Sprintf("verifying the files and directories for pod [%s] in restored namespace [%s] ", pod.Name, restoredNamespace))
					containerPaths := schedops.GetContainerPVCMountMap(pod)
					for containerName, mountPaths := range containerPaths {
						for _, mountPath := range mountPaths {
							restoredCombinedList := make([]string, 0)
							fileList, dirList, err := FetchFilesAndDirectoriesFromPod(pod, containerName, mountPath, existingFileDirMountPathMap[mountPath])
							dash.VerifyFatal(err, nil, fmt.Sprintf("fetching files and directory from mountpath [%s] for pod [%s]", mountPath, pod.Name))
							log.Infof(fmt.Sprintf("The list of files created in mountPath [%v] for pod [%v]: %v", mountPath, pod.Name, fileList))
							log.Infof(fmt.Sprintf("The list of directories created in mountPath [%v] for pod [%v]: %v", mountPath, pod.Name, dirList))
							restoredCombinedList = append(restoredCombinedList, fileList...)
							restoredCombinedList = append(restoredCombinedList, dirList...)
							log.Infof(fmt.Sprintf("the list of combined directories and files after restore: %v", restoredCombinedList))
							for _, item := range mountPathExcludeFileDirMap[mountPath] {
								if item != "" {
									if !IsPresent(restoredCombinedList, item) {
										log.Infof(fmt.Sprintf("the item(file/directory) [%s] is not present in the mountPath[%s] for pod [%s] in namespace [%s]", item, mountPath, pod.Name, restoredNamespace))
									} else {
										err := fmt.Errorf("the item(file/directory) [%s] is still present in mountPath [%s] for pod [%s] in namespace [%s]", item, mountPath, pod.Name, restoredNamespace)
										dash.VerifyFatal(err, nil, fmt.Sprintf("%v", err))
									}
								}
							}
						}
					}
				}
			}
		})

		Step("Update KDMP config map to not exclude any files or directories", func() {
			log.InfoD("Update KDMP config map to not exclude any files or directories")
			log.Infof(fmt.Sprintf("upating KDMP_EXCLUDE_FILE_LIST to nil"))
			excludeList = ""
			err := UpdateKDMPConfigMap("KDMP_EXCLUDE_FILE_LIST", excludeList)
			dash.VerifyFatal(err, nil, fmt.Sprintf("updating KDMP config map"))
		})

		Step("Fetch the directories and files from container mountPath for applications", func() {
			log.InfoD("Fetch the directories and files from container mountPath for applications")
			for _, namespace := range bkpNamespaces {
				pods, err := core.Instance().GetPods(namespace, nil)
				dash.VerifyFatal(err, nil, fmt.Sprintf("getting pods from namespace [%s] ", namespace))
				for _, pod := range pods.Items {
					containerPaths := schedops.GetContainerPVCMountMap(pod)
					for containerName, mountPaths := range containerPaths {
						for _, mountPath := range mountPaths {
							log.Infof(fmt.Sprintf("Fetching files and directories from path [%s] by excluding existing directories", mountPath))
							fileList, dirList, err := FetchFilesAndDirectoriesFromPod(pod, containerName, mountPath, existingFileDirMountPathMap[mountPath])
							dash.VerifyFatal(err, nil, fmt.Sprintf("fetching files and directory from mountpath [%s] for pod [%s]", mountPath, pod.Name))
							log.Infof(fmt.Sprintf("the list of files created in mountPath [%v] for pod [%v]: %v", mountPath, pod.Name, fileList))
							log.Infof(fmt.Sprintf("the list of directories created in mountPath [%v] for pod [%v]: %v", mountPath, pod.Name, dirList))
							fileListMountMap[mountPath] = fileList
							dirListMountMap[mountPath] = dirList
							finalFileList[mountPath] = append(finalFileList[mountPath], fileList...)
							finalFileList[mountPath] = append(finalFileList[mountPath], dirList...)
						}
					}
				}
			}
		})

		Step("Taking manual backup of namespaces with rules without excluding any files or directories", func() {
			log.InfoD(fmt.Sprintf("Taking manual backup of namespaces with rules without excluding any files or directories"))
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			backupNames = make([]string, 0)
			for _, namespace := range bkpNamespaces {
				backupName = fmt.Sprintf("%s-%v", BackupNamePrefix, RandomString(3))
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
				err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, bkpLocationName, backupLocationUID, appContextsToBackup, labelSelectors, BackupOrgID, clusterUid, preRuleName, preRuleUid, postRuleName, postRuleUid)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				backupNames = append(backupNames, backupName)
				backupNamespaceMap[backupName] = namespace
			}
		})

		Step("Taking restore of backups without excluding any files or directories", func() {
			log.InfoD("Taking restore of backups without excluding any files or directories")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			restoredNamespaces = make([]string, 0)
			for _, backupName := range backupNames {
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{backupNamespaceMap[backupName]})
				restoreName = fmt.Sprintf("%s-%v", RestoreNamePrefix, RandomString(3))
				restoreNamespace := "cust3-" + backupNamespaceMap[backupName]
				namespaceMapping := map[string]string{backupNamespaceMap[backupName]: restoreNamespace}
				restoredNamespaces = append(restoredNamespaces, restoreNamespace)
				err = CreateRestoreWithValidation(ctx, restoreName, backupName, namespaceMapping, make(map[string]string), DestinationClusterName, destClusterUid, BackupOrgID, appContextsToBackup)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s]", restoreName))
				restoreNames = append(restoreNames, restoreName)
			}
		})

		Step("List files and directories from the mount path after restore and verify all files and directories are present", func() {
			log.InfoD("List files and directories from the mount path after restore and verify all files and directories are present")

			defer func() {
				err := SetSourceKubeConfig()
				log.FailOnError(err, "Unable to switch context to source cluster [%s]", SourceClusterName)
			}()

			err := SetDestinationKubeConfig()
			log.FailOnError(err, "Switching context to destination cluster failed")

			for _, restoredNamespace := range restoredNamespaces {
				pods, err := core.Instance().GetPods(restoredNamespace, nil)
				dash.VerifyFatal(err, nil, fmt.Sprintf("getting pods from namespace [%s] ", restoredNamespace))
				for _, pod := range pods.Items {
					log.Infof(fmt.Sprintf("verifying all the files and directories created above for pod [%s] in restored namespace [%s] ", pod.Name, restoredNamespace))
					containerPaths := schedops.GetContainerPVCMountMap(pod)
					for containerName, mountPaths := range containerPaths {
						for _, mountPath := range mountPaths {
							restoredCombinedList := make([]string, 0)
							fileList, dirList, err := FetchFilesAndDirectoriesFromPod(pod, containerName, mountPath, existingFileDirMountPathMap[mountPath])
							dash.VerifyFatal(err, nil, fmt.Sprintf("fetching files and directory from mountpath [%s] for pod [%s]", mountPath, pod.Name))
							log.Infof(fmt.Sprintf("The list of files created in mountPath [%v] for pod [%v]: %v", mountPath, pod.Name, fileList))
							log.Infof(fmt.Sprintf("The list of directories created in mountPath [%v] for pod [%v]: %v", mountPath, pod.Name, dirList))
							restoredCombinedList = append(restoredCombinedList, fileList...)
							restoredCombinedList = append(restoredCombinedList, dirList...)
							log.Infof(fmt.Sprintf("the list of combined directories and files after restore: %v", restoredCombinedList))
							for _, item := range mountPathExcludeFileDirMap[mountPath] {
								if item != "" {
									if !IsPresent(restoredCombinedList, item) {
										err := fmt.Errorf("item(file/directory) [%s] is not present in mountPath [%s] for pod [%s] in namespace [%s]", item, mountPath, pod.Name, restoredNamespace)
										dash.VerifyFatal(err, nil, fmt.Sprintf("%v", err))
									}
								}
							}
						}
					}
				}
			}
		})

	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")

		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		log.InfoD("Deleting deployed applications")
		DestroyApps(scheduledAppContexts, opts)

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
		for _, restoreName := range restoreNames {
			err = DeleteRestore(restoreName, BackupOrgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting restore [%s]", restoreName))
		}

		for _, scheduleName := range scheduleNames {
			err = DeleteSchedule(scheduleName, SourceClusterName, BackupOrgID, ctx, true)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting schedule [%s]", scheduleName))
		}

		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})

// This testcase verifies backup and restore with mentioned valid directories or files from backed-up apps and restores them when invalid,non-existent storageclass and files are there in KDMP exclude list
var _ = Describe("{ExcludeInvalidDirectoryFileBackup}", Label(TestCaseLabelsMap[ExcludeInvalidDirectoryFileBackup]...), func() {
	var (
		backupName                    string
		scheduledAppContexts          []*scheduler.Context
		AppContextsMapping            = make(map[string]*scheduler.Context)
		namespace                     string
		bkpNamespaces                 = make([]string, 0)
		backupNames                   = make([]string, 0)
		restoreNames                  = make([]string, 0)
		scheduleNames                 = make([]string, 0)
		clusterUid                    string
		destClusterUid                string
		clusterStatus                 api.ClusterInfo_StatusInfo_Status
		restoreName                   string
		cloudCredName                 string
		cloudCredUID                  string
		backupLocationUID             string
		bkpLocationName               string
		numDeployments                int
		providers                     []string
		backupLocationMap             = make(map[string]string)
		labelSelectors                = make(map[string]string)
		storageClassExcludeFileDirMap = make(map[*storagev1.StorageClass][]string)
		mountPathExcludeFileDirMap    = make(map[string][]string)
		existingFileDirMountPathMap   = make(map[string][]string)
		fileListMountMap              = make(map[string][]string)
		dirListMountMap               = make(map[string][]string)
		finalFileList                 = make(map[string][]string)
		scMountPathMap                = make(map[string]*storagev1.StorageClass)
		backupNamespaceMap            = make(map[string]string)
		podScMountPathMap             = make(map[string]map[string]*storagev1.StorageClass)
		excludeList                   string
		fileList                      []string
		dirList                       []string
		preRuleName                   string
		postRuleName                  string
		preRuleUid                    string
		postRuleUid                   string
		periodicSchedulePolicyName    string
		periodicSchedulePolicyUid     string
		testAppList                   []string
		mutex                         sync.Mutex
		restoredNamespaces            = make([]string, 0)
		wg                            sync.WaitGroup
	)
	JustBeforeEach(func() {
		numDeployments = 1
		providers = GetBackupProviders()
		StartPxBackupTorpedoTest("ExcludeInvalidDirectoryFileBackup", "Excludes mentioned valid directories or files from backed-up apps and restores them when invalid,non-existent storageclass and files are there in KDMP exclude list", nil, 93692, Ak, Q4FY24)

		log.InfoD(fmt.Sprintf("App list %v", Inst().AppList))
		scheduledAppContexts = make([]*scheduler.Context, 0)
		testAppList = []string{"vdbench-fbda-one-vol"}
		actualAppList := Inst().AppList
		defer func() {
			Inst().AppList = actualAppList
		}()
		Inst().AppList = testAppList
		log.InfoD("Starting to deploy applications")
		for i := 0; i < numDeployments; i++ {
			log.InfoD(fmt.Sprintf("Iteration %v of deploying applications", i))
			log.Infof(fmt.Sprintf("Taskname Prefix is : %s", TaskNamePrefix))
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = AppReadinessTimeout
				namespace = GetAppNamespace(ctx, taskName)
				scheduledAppContexts = append(scheduledAppContexts, ctx)
				AppContextsMapping[namespace] = ctx
				bkpNamespaces = append(bkpNamespaces, namespace)
			}
		}

	})
	It("Update KDMP config map file with invalid directories or files to exclude from the backup", func() {

		Step("Validating deployed applications", func() {
			log.InfoD("Validating deployed applications")
			ValidateApplications(scheduledAppContexts)
		})

		Step("Getting mountpath and associated storageClass for containers in deployed application", func() {
			log.InfoD("Getting mountpath associated storageClass for containers in deployed application")
			for _, namespace := range bkpNamespaces {
				pods, err := core.Instance().GetPods(namespace, nil)
				dash.VerifyFatal(err, nil, fmt.Sprintf("getting pods from namespace [%s] ", namespace))
				for _, pod := range pods.Items {
					scMountPathMap, err := schedops.GetContainerPVCMountMapWithSC(pod)
					dash.VerifyFatal(err, nil, fmt.Sprintf("getting storage class and mountpath mapping for pod [%s] ", pod.Name))
					podScMountPathMap[pod.Name] = scMountPathMap
				}
			}
		})

		Step("Fetch the existing directories and files within mountPath before writing files and directories", func() {
			log.InfoD("Fetch the existing directories and files within mountPath before writing files and directories")
			for _, namespace := range bkpNamespaces {
				pods, err := core.Instance().GetPods(namespace, nil)
				dash.VerifyFatal(err, nil, fmt.Sprintf("getting pods from namespace [%s] ", namespace))
				for _, pod := range pods.Items {
					containerPaths := schedops.GetContainerPVCMountMap(pod)
					existingFileDirList := make([]string, 0)
					for containerName, mountPaths := range containerPaths {
						for _, mountPath := range mountPaths {
							log.Infof(fmt.Sprintf("Fetch the existing directories and files within mountPath [%s] before writing files and directories", mountPath))
							existingFileList, existingDirList, err := FetchFilesAndDirectoriesFromPod(pod, containerName, mountPath, nil)
							existingFileDirList = append(existingFileDirList, existingFileList...)
							existingFileDirList = append(existingFileDirList, existingDirList...)
							existingFileDirMountPathMap[mountPath] = existingFileDirList
							dash.VerifyFatal(err, nil, fmt.Sprintf("fetching files and directory from mountpath [%s] for pod [%s]", mountPath, pod.Name))
						}
					}
				}
			}
		})

		Step("Create nested directories and files into container mountPath for applications", func() {
			log.InfoD("Create nested directories and files into container mountPath for applications")
			for _, namespace := range bkpNamespaces {
				pods, err := core.Instance().GetPods(namespace, nil)
				dash.VerifyFatal(err, nil, fmt.Sprintf("getting pods from namespace [%s] ", namespace))
				for _, pod := range pods.Items {
					containerPaths := schedops.GetContainerPVCMountMap(pod)
					for containerName, mountPaths := range containerPaths {
						for _, mountPath := range mountPaths {
							DirectoryConfig := PodDirectoryConfig{
								BasePath:          mountPath,
								Depth:             100,
								Levels:            1,
								FilesPerDirectory: 100,
							}
							log.Infof(fmt.Sprintf("creating nested directories and files within mountPath [%s] with depth [%d] , level [%d] and FilesPerDirectory [%d]", DirectoryConfig.BasePath, DirectoryConfig.Depth, DirectoryConfig.Levels, DirectoryConfig.FilesPerDirectory))
							err = CreateNestedDirectoriesWithFilesInPod(pod, containerName, DirectoryConfig)
							dash.VerifyFatal(err, nil, fmt.Sprintf("creating nested directories and files at mountpath [%v] for pod [%v] in namespace [%v]", mountPath, pod.Name, pod.Namespace))
							log.Infof(fmt.Sprintf("Fetching files and directories from path [%s] by excluding existing directories", DirectoryConfig.BasePath))
							fileList, dirList, err = FetchFilesAndDirectoriesFromPod(pod, containerName, mountPath, existingFileDirMountPathMap[mountPath])
							dash.VerifyFatal(err, nil, fmt.Sprintf("fetching files and directory from mountpath [%s] for pod [%s]", mountPath, pod.Name))
							log.Infof(fmt.Sprintf("the list of files created in mountPath [%v] for pod [%v]: %v", mountPath, pod.Name, fileList))
							log.Infof(fmt.Sprintf("the list of directories created in mountPath [%v] for pod [%v]: %v", mountPath, pod.Name, dirList))
							fileListMountMap[mountPath] = fileList
							dirListMountMap[mountPath] = dirList
						}
					}
				}
			}
		})

		Step("Update KDMP config map on source cluster by formatting storage class name and random files and directories as a string", func() {
			log.InfoD("Update KDMP config map on source cluster by formatting storage class name and random files and directories as a string")
			for _, namespace := range bkpNamespaces {
				pods, err := core.Instance().GetPods(namespace, nil)
				dash.VerifyFatal(err, nil, fmt.Sprintf("getting pods from namespace [%s] ", namespace))
				for _, pod := range pods.Items {
					containerPaths := schedops.GetContainerPVCMountMap(pod)
					for _, mountPaths := range containerPaths {
						for _, mountPath := range mountPaths {
							excludeFileDirList := make([]string, 0)
							log.Infof(fmt.Sprintf("Fetch some random directories from created list %v", dirListMountMap[mountPath]))
							randomDirs, err := GetRandomSubset(dirListMountMap[mountPath], 50)
							dash.VerifyFatal(err, nil, fmt.Sprintf("Getting random directories from the list"))
							log.Infof(fmt.Sprintf("the list of directories randomly selected from mountPath- %v : %v", mountPath, randomDirs))
							log.Infof(fmt.Sprintf("Fetch some random files from created list %v", fileListMountMap[mountPath]))
							randomFiles, err := GetRandomSubset(fileListMountMap[mountPath], 100)
							dash.VerifyFatal(err, nil, fmt.Sprintf("Getting random files from the list"))
							log.Infof(fmt.Sprintf("the list of files randomly selected from mountPath- %v : %v", mountPath, randomFiles))
							excludeFileDirList = append(excludeFileDirList, randomDirs...)
							excludeFileDirList = append(excludeFileDirList, randomFiles...)
							sc := podScMountPathMap[pod.Name][mountPath]
							storageClassExcludeFileDirMap[sc] = append(storageClassExcludeFileDirMap[sc], excludeFileDirList...)
							mountPathExcludeFileDirMap[mountPath] = append(mountPathExcludeFileDirMap[mountPath], excludeFileDirList...)
						}
					}
				}
			}
			log.Infof("create formatted string with storage class name and exclude file and directories list")
			excludeList = GetExcludeFileListValue(storageClassExcludeFileDirMap)
			err := UpdateKDMPConfigMap("KDMP_EXCLUDE_FILE_LIST", excludeList)
			dash.VerifyFatal(err, nil, fmt.Sprintf("updating KDMP config map"))
		})

		Step("Creating backup location and cloud setting", func() {
			log.InfoD("Creating backup location and cloud setting")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
				bkpLocationName = fmt.Sprintf("%s-%s-bl", provider, getGlobalBucketName(provider))
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = bkpLocationName
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocation(provider, bkpLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", bkpLocationName))
			}
		})
		Step("Registering cluster for backup", func() {
			log.InfoD("Registering cluster for backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
			destClusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))
		})

		Step("Create new storage class on destination cluster with similar spec as source cluster", func() {
			log.InfoD("Create new storage class on destination cluster with similar spec as source cluster")
			defer func() {
				err := SetSourceKubeConfig()
				log.FailOnError(err, "Unable to switch context to source cluster [%s]", SourceClusterName)
			}()
			log.InfoD("Switching cluster context to destination cluster")
			err := SetDestinationKubeConfig()
			log.FailOnError(err, "Failed to set destination kubeconfig")

			for _, scMountPathMap := range podScMountPathMap {
				for _, storageClass := range scMountPathMap {
					isScpresent, err := IsStorageClassPresent(storageClass.Name)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Checking storageClass %s present on cluster", storageClass.Name))
					if !isScpresent {
						v1obj := metaV1.ObjectMeta{
							Name: storageClass.Name,
						}
						scObj := storagev1.StorageClass{
							ObjectMeta:           v1obj,
							Provisioner:          storageClass.Provisioner,
							Parameters:           storageClass.Parameters,
							ReclaimPolicy:        storageClass.ReclaimPolicy,
							VolumeBindingMode:    storageClass.VolumeBindingMode,
							MountOptions:         storageClass.MountOptions,
							AllowVolumeExpansion: storageClass.AllowVolumeExpansion,
						}
						_, err = storage.Instance().CreateStorageClass(&scObj)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Creating new storage class %v on destination cluster %s", storageClass.Name, DestinationClusterName))
					} else {
						log.Infof(fmt.Sprintf("storageClass %s already present on cluster , hence skipping creation", storageClass.Name))
					}
				}
			}
		})

		Step(fmt.Sprintf("Creation of pre and post exec rules for applications "), func() {
			log.InfoD("Creation of pre and post exec rules for applications ")
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

		Step(fmt.Sprintf("Create schedule policy for backup schedules"), func() {
			log.InfoD("Create schedule policy for backup schedules")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			periodicSchedulePolicyName = fmt.Sprintf("%s-%s", "periodic", RandomString(5))
			periodicSchedulePolicyUid = uuid.New()
			periodicSchedulePolicyInterval := int64(15)
			err = CreateBackupScheduleIntervalPolicy(5, periodicSchedulePolicyInterval, 5, periodicSchedulePolicyName, periodicSchedulePolicyUid, BackupOrgID, ctx, false, false)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of periodic schedule policy of interval [%v] minutes named [%s] ", periodicSchedulePolicyInterval, periodicSchedulePolicyName))

		})

		Step("Taking manual backup of namespaces with rules", func() {
			log.InfoD(fmt.Sprintf("Taking manual backup of namespaces with rules"))
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, namespace := range bkpNamespaces {
				backupName = fmt.Sprintf("%s-%v", BackupNamePrefix, RandomString(3))
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
				err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, bkpLocationName, backupLocationUID, appContextsToBackup, labelSelectors, BackupOrgID, clusterUid, preRuleName, preRuleUid, postRuleName, postRuleUid)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				backupNames = append(backupNames, backupName)
				backupNamespaceMap[backupName] = namespace
			}
		})

		Step("Taking schedule backup of namespaces with rules", func() {
			log.InfoD(fmt.Sprintf("Taking schedule backup of namespaces with rules"))
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, namespace := range bkpNamespaces {
				scheduleName := fmt.Sprintf("%s-schedule-with-rules-%s", BackupNamePrefix, RandomString(3))
				log.InfoD("Creating a schedule backup of namespace [%s] without pre and post exec rules", namespace)
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
				scheduleBackupName, err := CreateScheduleBackupWithValidation(ctx, scheduleName, SourceClusterName, clusterUid, bkpLocationName, backupLocationUID, appContextsToBackup,
					labelSelectors, BackupOrgID, preRuleName, preRuleUid, postRuleName, postRuleUid, periodicSchedulePolicyName, periodicSchedulePolicyUid, false)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of backup [%s]", scheduleBackupName))
				err = SuspendBackupSchedule(scheduleName, periodicSchedulePolicyName, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Suspending Backup Schedule [%s] ", scheduleName))
				backupNames = append(backupNames, scheduleBackupName)
				scheduleNames = append(scheduleNames, scheduleName)
				backupNamespaceMap[scheduleBackupName] = namespace
			}
		})

		Step("Taking restore of backups created", func() {
			log.InfoD("Taking restore of backups created")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			restoreSingleNSBackupInVariousWaysTask := func(index int, backupName string) {
				restoreConfigs := []struct {
					namePrefix          string
					namespaceMapping    map[string]string
					storageClassMapping map[string]string
					replacePolicy       ReplacePolicyType
				}{
					{
						"test-custom-restore-single-ns",
						map[string]string{backupNamespaceMap[backupName]: fmt.Sprintf("cust1-%s-%d", backupNamespaceMap[backupName], index)},
						make(map[string]string),
						ReplacePolicyRetain,
					},
					{
						"test-replace-restore-single-ns",
						map[string]string{backupNamespaceMap[backupName]: fmt.Sprintf("cust1-rep-%s-%d", backupNamespaceMap[backupName], index)},
						make(map[string]string),
						ReplacePolicyDelete,
					},
				}
				for _, config := range restoreConfigs {
					restoreName := fmt.Sprintf("%s-%s", config.namePrefix, RandomString(3))
					log.InfoD("Restoring backup [%s] in cluster [%s] with restore [%s] and namespace mapping %v", backupName, DestinationClusterName, restoreName, config.namespaceMapping)
					if config.replacePolicy == ReplacePolicyRetain {
						appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{backupNamespaceMap[backupName]})
						restoredNamespaces = append(restoredNamespaces, config.namespaceMapping[backupNamespaceMap[backupName]])
						err = CreateRestoreWithValidation(ctx, restoreName, backupName, config.namespaceMapping, config.storageClassMapping, DestinationClusterName, destClusterUid, BackupOrgID, appContextsToBackup)
					} else if config.replacePolicy == ReplacePolicyDelete {
						restoredNamespaces = append(restoredNamespaces, config.namespaceMapping[backupNamespaceMap[backupName]])
						err = CreateRestoreWithReplacePolicy(restoreName, backupName, config.namespaceMapping, DestinationClusterName, BackupOrgID, ctx, config.storageClassMapping, config.replacePolicy)
					}
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying restoration [%s] of single namespace backup [%s] in cluster", restoreName, backupName))
					restoreNames = SafeAppend(&mutex, restoreNames, restoreName).([]string)
				}
			}
			_ = TaskHandler(backupNames, restoreSingleNSBackupInVariousWaysTask, Sequential)
		})

		Step("List files and directories from the mount path after restore and verify excluded items are not present", func() {
			log.InfoD("List files and directories from the mount path after restore and verify excluded items are not present")

			defer func() {
				err := SetSourceKubeConfig()
				log.FailOnError(err, "Unable to switch context to source cluster [%s]", SourceClusterName)
			}()

			err := SetDestinationKubeConfig()
			log.FailOnError(err, "Switching context to destination cluster failed")

			for _, restoredNamespace := range restoredNamespaces {
				pods, err := core.Instance().GetPods(restoredNamespace, nil)
				dash.VerifyFatal(err, nil, fmt.Sprintf("getting pods from namespace [%s] ", restoredNamespace))
				for _, pod := range pods.Items {
					log.Infof(fmt.Sprintf("verifying the files and directories for pod [%s] in restored namespace [%s] ", pod.Name, restoredNamespace))
					containerPaths := schedops.GetContainerPVCMountMap(pod)
					for containerName, mountPaths := range containerPaths {
						for _, mountPath := range mountPaths {
							restoredCombinedList := make([]string, 0)
							fileList, dirList, err := FetchFilesAndDirectoriesFromPod(pod, containerName, mountPath, existingFileDirMountPathMap[mountPath])
							dash.VerifyFatal(err, nil, fmt.Sprintf("fetching files and directory from mountpath [%s] for pod [%s]", mountPath, pod.Name))
							log.Infof(fmt.Sprintf("The list of files created in mountPath [%v] for pod [%v]: %v", mountPath, pod.Name, fileList))
							log.Infof(fmt.Sprintf("The list of directories created in mountPath [%v] for pod [%v]: %v", mountPath, pod.Name, dirList))
							restoredCombinedList = append(restoredCombinedList, fileList...)
							restoredCombinedList = append(restoredCombinedList, dirList...)
							log.Infof(fmt.Sprintf("the list of combined directories and files after restore: %v", restoredCombinedList))
							for _, item := range mountPathExcludeFileDirMap[mountPath] {
								if item != "" {
									if !IsPresent(restoredCombinedList, item) {
										log.Infof(fmt.Sprintf("the item file/directory [%s] is not present in the mountPath[%s] for pod [%s] in namespace [%s]", item, mountPath, pod.Name, restoredNamespace))
									} else {
										err := fmt.Errorf("the item file/directory[%s] is still present in mountPath [%s] for pod [%s] in namespace [%s]", item, mountPath, pod.Name, restoredNamespace)
										dash.VerifyFatal(err, nil, fmt.Sprintf("%v", err))
									}
								}
							}
						}
					}
				}
			}
		})

		Step("Update KDMP config map with some non-existent valid and invalid files or directories", func() {
			log.InfoD("Update KDMP config map with some non-existent valid and invalid files or directories")
			excludeFileDirList := make([]string, 0)
			numOfitems := 100
			log.Infof("Creating some random non-existent valid file names and directories and adding to the exclude list")
			for i := 1; i <= numOfitems; i++ {
				fileName := fmt.Sprintf("file-%s.txt", RandomString(10))
				dirName := fmt.Sprintf("directory-%s", RandomString(10))
				excludeFileDirList = append(excludeFileDirList, fileName)
				excludeFileDirList = append(excludeFileDirList, dirName)
			}

			log.Infof("Creating some random non-existent invalid file names and directories and  adding to the exclude list")
			for i := 1; i <= numOfitems; i++ {
				fileName := fmt.Sprintf("/File_Name-%s.docx", RandomString(10))
				dirName := fmt.Sprintf("/My*Directory-%s", RandomString(10))
				excludeFileDirList = append(excludeFileDirList, fileName)
				excludeFileDirList = append(excludeFileDirList, dirName)
			}

			log.Infof("Creating some invalid file names with more than 255 chars which is not existed  and adding to the exclude list")
			for i := 1; i <= numOfitems; i++ {
				fileName := fmt.Sprintf("%s.txt", RandomString(260))
				excludeFileDirList = append(excludeFileDirList, fileName)
			}

			log.Infof("Adding the above list to all the existing storageClass exclude list")
			for _, storageClass := range scMountPathMap {
				storageClassExcludeFileDirMap[storageClass] = append(storageClassExcludeFileDirMap[storageClass], excludeFileDirList...)
			}

			log.Infof("Adding the above list to the an non existing storageClass and add to exclude list")
			v1obj := metaV1.ObjectMeta{
				Name: "px-backup-test-sc",
			}
			nonExistingstorageClass := storagev1.StorageClass{
				ObjectMeta: v1obj,
			}
			storageClassExcludeFileDirMap[&nonExistingstorageClass] = append(storageClassExcludeFileDirMap[&nonExistingstorageClass], excludeFileDirList...)

			log.Infof("create formatted string with storage class name and exclude file and directories list")
			excludeList = GetExcludeFileListValue(storageClassExcludeFileDirMap)
			err := UpdateKDMPConfigMap("KDMP_EXCLUDE_FILE_LIST", excludeList)
			dash.VerifyFatal(err, nil, fmt.Sprintf("updating KDMP config map"))
		})

		Step("Taking manual backup of namespaces with rules with excluding valid and invalid files or directories", func() {
			log.InfoD(fmt.Sprintf("Taking manual backup of namespaces with rules with excluding valid and invalid files or directories"))
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			backupNames = make([]string, 0)
			for _, namespace := range bkpNamespaces {
				backupName = fmt.Sprintf("%s-%v", BackupNamePrefix, RandomString(3))
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
				err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, bkpLocationName, backupLocationUID, appContextsToBackup, labelSelectors, BackupOrgID, clusterUid, preRuleName, preRuleUid, postRuleName, postRuleUid)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				backupNames = append(backupNames, backupName)
				backupNamespaceMap[backupName] = namespace
			}
		})

		Step("Taking restore of backups with excluding valid and invalid files or directories", func() {
			log.InfoD("Taking restore of backups with excluding valid and invalid files or directories")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			restoredNamespaces = make([]string, 0)
			for _, backupName := range backupNames {
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{backupNamespaceMap[backupName]})
				restoreName = fmt.Sprintf("%s-%v", RestoreNamePrefix, RandomString(3))
				restoreNamespace := "custom2-" + backupNamespaceMap[backupName]
				namespaceMapping := map[string]string{backupNamespaceMap[backupName]: restoreNamespace}
				restoredNamespaces = append(restoredNamespaces, restoreNamespace)
				err = CreateRestoreWithValidation(ctx, restoreName, backupName, namespaceMapping, make(map[string]string), DestinationClusterName, destClusterUid, BackupOrgID, appContextsToBackup)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s]", restoreName))
				restoreNames = append(restoreNames, restoreName)
			}
		})

		Step("List files and directories from the mount path after restore and verify excluded items are still not present", func() {
			log.InfoD("List files and directories from the mount path after restore and verify excluded items are still not present")

			defer func() {
				err := SetSourceKubeConfig()
				log.FailOnError(err, "Unable to switch context to source cluster [%s]", SourceClusterName)
			}()

			err := SetDestinationKubeConfig()
			log.FailOnError(err, "Switching context to destination cluster failed")

			for _, restoredNamespace := range restoredNamespaces {
				pods, err := core.Instance().GetPods(restoredNamespace, nil)
				dash.VerifyFatal(err, nil, fmt.Sprintf("getting pods from namespace [%s] ", restoredNamespace))
				for _, pod := range pods.Items {
					log.Infof(fmt.Sprintf("verifying the files and directories for pod [%s] in restored namespace [%s] ", pod.Name, restoredNamespace))
					containerPaths := schedops.GetContainerPVCMountMap(pod)
					for containerName, mountPaths := range containerPaths {
						for _, mountPath := range mountPaths {
							restoredCombinedList := make([]string, 0)
							fileList, dirList, err := FetchFilesAndDirectoriesFromPod(pod, containerName, mountPath, existingFileDirMountPathMap[mountPath])
							dash.VerifyFatal(err, nil, fmt.Sprintf("fetching files and directory from mountpath [%s] for pod [%s]", mountPath, pod.Name))
							log.Infof(fmt.Sprintf("The list of files created in mountPath [%v] for pod [%v]: %v", mountPath, pod.Name, fileList))
							log.Infof(fmt.Sprintf("The list of directories created in mountPath [%v] for pod [%v]: %v", mountPath, pod.Name, dirList))
							restoredCombinedList = append(restoredCombinedList, fileList...)
							restoredCombinedList = append(restoredCombinedList, dirList...)
							log.Infof(fmt.Sprintf("the list of combined directories and files after restore: %v", restoredCombinedList))
							for _, item := range mountPathExcludeFileDirMap[mountPath] {
								if item != "" {
									if !IsPresent(restoredCombinedList, item) {
										log.Infof(fmt.Sprintf("the item file/directory [%s] is not present in the mountPath[%s] for pod [%s] in namespace [%s]", item, mountPath, pod.Name, restoredNamespace))
									} else {
										err := fmt.Errorf("the item file/directory[%s] is still present in mountPath [%s] for pod [%s] in namespace [%s]", item, mountPath, pod.Name, restoredNamespace)
										dash.VerifyFatal(err, nil, fmt.Sprintf("%v", err))
									}
								}
							}
						}
					}
				}
			}
		})

		Step("Update KDMP config map to not exclude any files or directories", func() {
			log.InfoD("Update KDMP config map to not exclude any files or directories")
			log.Infof(fmt.Sprintf("upating KDMP_EXCLUDE_FILE_LIST to nil"))
			excludeList = ""
			err := UpdateKDMPConfigMap("KDMP_EXCLUDE_FILE_LIST", excludeList)
			dash.VerifyFatal(err, nil, fmt.Sprintf("updating KDMP config map"))
		})

		Step("Fetch the directories and files from container mountPath for applications", func() {
			log.InfoD("Fetch the directories and files from container mountPath for applications")
			for _, namespace := range bkpNamespaces {
				pods, err := core.Instance().GetPods(namespace, nil)
				dash.VerifyFatal(err, nil, fmt.Sprintf("getting pods from namespace [%s] ", namespace))
				for _, pod := range pods.Items {
					containerPaths := schedops.GetContainerPVCMountMap(pod)
					for containerName, mountPaths := range containerPaths {
						for _, mountPath := range mountPaths {
							log.Infof(fmt.Sprintf("Fetching files and directories from path [%s] by excluding existing directories", mountPath))
							fileList, dirList, err := FetchFilesAndDirectoriesFromPod(pod, containerName, mountPath, existingFileDirMountPathMap[mountPath])
							dash.VerifyFatal(err, nil, fmt.Sprintf("fetching files and directory from mountpath [%s] for pod [%s]", mountPath, pod.Name))
							log.Infof(fmt.Sprintf("the list of files created in mountPath [%v] for pod [%v]: %v", mountPath, pod.Name, fileList))
							log.Infof(fmt.Sprintf("the list of directories created in mountPath [%v] for pod [%v]: %v", mountPath, pod.Name, dirList))
							fileListMountMap[mountPath] = fileList
							dirListMountMap[mountPath] = dirList
							finalFileList[mountPath] = append(finalFileList[mountPath], fileList...)
							finalFileList[mountPath] = append(finalFileList[mountPath], dirList...)
						}
					}
				}
			}
		})

		Step("Taking manual backup of namespaces with rules without excluding any files or directories", func() {
			log.InfoD(fmt.Sprintf("Taking manual backup of namespaces with rules without excluding any files or directories"))
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			backupNames = make([]string, 0)
			for _, namespace := range bkpNamespaces {
				backupName = fmt.Sprintf("%s-%v", BackupNamePrefix, RandomString(3))
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
				err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, bkpLocationName, backupLocationUID, appContextsToBackup, labelSelectors, BackupOrgID, clusterUid, preRuleName, preRuleUid, postRuleName, postRuleUid)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				backupNames = append(backupNames, backupName)
				backupNamespaceMap[backupName] = namespace
			}
		})

		Step("Taking restore of backups without excluding any files or directories", func() {
			log.InfoD("Taking restore of backups without excluding any files or directories")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			restoredNamespaces = make([]string, 0)
			for _, backupName := range backupNames {
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{backupNamespaceMap[backupName]})
				restoreName = fmt.Sprintf("%s-%v", RestoreNamePrefix, RandomString(3))
				restoreNamespace := "custom3-" + backupNamespaceMap[backupName]
				namespaceMapping := map[string]string{backupNamespaceMap[backupName]: restoreNamespace}
				restoredNamespaces = append(restoredNamespaces, restoreNamespace)
				err = CreateRestoreWithValidation(ctx, restoreName, backupName, namespaceMapping, make(map[string]string), DestinationClusterName, destClusterUid, BackupOrgID, appContextsToBackup)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s]", restoreName))
				restoreNames = append(restoreNames, restoreName)
			}
		})

		Step("List files and directories from the mount path after restore and verify all files and directories are present", func() {
			log.InfoD("List files and directories from the mount path after restore and verify all files and directories are present")

			defer func() {
				err := SetSourceKubeConfig()
				log.FailOnError(err, "Unable to switch context to source cluster [%s]", SourceClusterName)
			}()

			err := SetDestinationKubeConfig()
			log.FailOnError(err, "Switching context to destination cluster failed")

			for _, restoredNamespace := range restoredNamespaces {
				pods, err := core.Instance().GetPods(restoredNamespace, nil)
				dash.VerifyFatal(err, nil, fmt.Sprintf("getting pods from namespace [%s] ", restoredNamespace))
				for _, pod := range pods.Items {
					log.Infof(fmt.Sprintf("verifying all the files and directories created above for pod [%s] in restored namespace [%s] ", pod.Name, restoredNamespace))
					containerPaths := schedops.GetContainerPVCMountMap(pod)
					for containerName, mountPaths := range containerPaths {
						for _, mountPath := range mountPaths {
							restoredCombinedList := make([]string, 0)
							fileList, dirList, err := FetchFilesAndDirectoriesFromPod(pod, containerName, mountPath, existingFileDirMountPathMap[mountPath])
							dash.VerifyFatal(err, nil, fmt.Sprintf("fetching files and directory from mountpath [%s] for pod [%s]", mountPath, pod.Name))
							log.Infof(fmt.Sprintf("The list of files created in mountPath [%v] for pod [%v]: %v", mountPath, pod.Name, fileList))
							log.Infof(fmt.Sprintf("The list of directories created in mountPath [%v] for pod [%v]: %v", mountPath, pod.Name, dirList))
							restoredCombinedList = append(restoredCombinedList, fileList...)
							restoredCombinedList = append(restoredCombinedList, dirList...)
							log.Infof(fmt.Sprintf("the list of combined directories and files after restore: %v", restoredCombinedList))
							for _, item := range mountPathExcludeFileDirMap[mountPath] {
								if item != "" {
									if !IsPresent(restoredCombinedList, item) {
										err := fmt.Errorf("item(file/directory) [%s] is not present in mountPath [%s] for pod [%s] in namespace [%s]", item, mountPath, pod.Name, restoredNamespace)
										dash.VerifyFatal(err, nil, fmt.Sprintf("%v", err))
									}
								}
							}
						}
					}
				}
			}
		})

	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")

		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		log.InfoD("Deleting deployed applications")
		DestroyApps(scheduledAppContexts, opts)

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
		for _, restoreName := range restoreNames {
			err = DeleteRestore(restoreName, BackupOrgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting restore [%s]", restoreName))
		}

		for _, scheduleName := range scheduleNames {
			err = DeleteSchedule(scheduleName, SourceClusterName, BackupOrgID, ctx, true)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting schedule [%s]", scheduleName))
		}

		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)

	})
})

// This TC deletes the kopia executor pod while backup and restore are in progress and validates their status
var _ = Describe("{CrashKopiaToolWhenBackUpRestoreInProgress}", Label(TestCaseLabelsMap[CrashKopiaToolWhenBackUpRestoreInProgress]...), func() {

	/*
		Steps:
		1. Schedule applications
		2. Create a backup location and cloud setting
		3. Register source and destination clusters for backup and restore
		4. Create a backup of application from source cluster
		5. Delete the kopia executor pod while the backup is in progress
		6. Verify backup status after deleting the kopia executor pod
		7. Restore the backup on the destination cluster
		8. Delete the kopia executor pod while the restore is in progress
		9. Verify restore status after deleting the kopia executor pod
	*/

	var (
		err                   error
		ctx                   context.Context
		backupName            string
		restoreName           string
		providers             []string
		namespaces            []string
		scheduledAppContexts  []*scheduler.Context
		appContextsToBackup   []*scheduler.Context
		backupLocationMap     map[string]string
		cloudCredName         string
		cloudCredUID          string
		backupLocationName    string
		backupLocationUID     string
		sourceClusterUid      string
		destinationClusterUid string
		controlChannel        chan string
		errorGroup            *errgroup.Group
	)

	JustBeforeEach(func() {
		backupLocationMap = make(map[string]string)
		providers = GetBackupProviders()

		ctx, err = backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")

		StartPxBackupTorpedoTest("CrashKopiaToolWhenBackUpRestoreInProgress", "Crash the Kopia tool when Backup and Restore is in progress", nil, 58078, Dchothani, Q3FY25)

		log.InfoD("scheduling applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, appCtx := range appContexts {
				appCtx.ReadinessTimeout = AppReadinessTimeout
				if !Contains(namespaces, appCtx.ScheduleOptions.Namespace) {
					namespaces = append(namespaces, appCtx.ScheduleOptions.Namespace)
				}
				scheduledAppContexts = append(scheduledAppContexts, appCtx)
			}
		}

		appContextsToBackup = FilterAppContextsByNamespace(scheduledAppContexts, []string{namespaces[0]})
	})

	It("Crash the kopia tool when the backup and restore is in progress", func() {

		Step("Validating deployed applications", func() {
			log.InfoD("Validating deployed applications")

			controlChannel, errorGroup = ValidateApplicationsStartData(scheduledAppContexts, ctx)
		})

		Step("Disk dumping data into the application pods", func() {
			log.InfoD("Disk dumping data into the application pods")
			err = PopulateDataInNamespacePods(namespaces[0], 1024)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Writing data to the pods in namespace [%s]", namespaces[0]))
		})

		Step("Creating backup location and cloud setting", func() {
			log.InfoD("Creating backup location and cloud setting")

			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
				backupLocationName = fmt.Sprintf("%s-%s-bl-%v", provider, getGlobalBucketName(provider), time.Now().Unix())
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

			destinationClusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, DestinationClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", DestinationClusterName))
		})

		Step("Creating backup of application from source cluster", func() {
			log.InfoD("Creating backup of application from source cluster")

			backupName = fmt.Sprintf("%s-%v", BackupNamePrefix, time.Now().Unix())

			_, err = CreateBackupWithoutCheck(ctx, backupName, SourceClusterName, backupLocationName, backupLocationUID, appContextsToBackup, make(map[string]string), BackupOrgID, sourceClusterUid, "", "", "", "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of backup [%s] with namespace [%s]", backupName, namespaces[0]))
		})

		Step("Deleting the kopia executor pod while the backup is in progress", func() {
			log.InfoD("deleting the kopia executor pod while backup %s is in progress", backupName)

			err = DeletePodWhileBackupInProgress(ctx, BackupOrgID, backupName, namespaces[0], KopiaBackupExecutorPodLabel)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Delete kopia executor pod while backup %s in progress", backupName))
		})

		Step("Verifying backup status after deleting the kopia executor pod", func() {
			log.InfoD("Verifying backup status after deleting the kopia executor pod")

			err = BackupSuccessCheckWithValidation(ctx, backupName, appContextsToBackup, BackupOrgID, MaxWaitPeriodForBackupCompletionInMinutes*time.Minute, 30*time.Second, []string{}...)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of backup:[%s] after deleting kopia executor pod", backupName))
		})

		Step("Create storage class on destination cluster for restore", func() {
			log.InfoD("Create storage class on destination cluster for restore")
			pvcs, err := core.Instance().GetPersistentVolumeClaims(namespaces[0], make(map[string]string))
			log.FailOnError(err, "Getting PVCs on source cluster")
			var storageClasses []*storagev1.StorageClass
			for _, singlePvc := range pvcs.Items {
				storageClass, err := core.Instance().GetStorageClassForPVC(&singlePvc)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Getting storage class %v from PVC in source cluster", storageClass.Name))
				storageClasses = append(storageClasses, storageClass)
			}
			defer func() {
				err = SetSourceKubeConfig()
				dash.VerifyFatal(err, nil, "Setting source kubeconfig")
			}()
			err = SetDestinationKubeConfig()
			dash.VerifyFatal(err, nil, "Setting destination kubeconfig")
			for _, sc := range storageClasses {
				sc.ResourceVersion = ""
				_, err = storage.Instance().CreateStorageClass(sc)
				if err != nil && !strings.Contains(err.Error(), "already exists") {
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creating storage class %s on dest cluster", sc.Name))
				}
			}
		})

		Step("Restoring the backup on the destination cluster", func() {
			log.InfoD("Restoring the backup on the destination cluster")

			log.Infof("Switching the context to destination cluster")
			err = SetDestinationKubeConfig()
			log.FailOnError(err, "Switching context to destination cluster failed")

			log.InfoD("Restoring the backup %s", backupName)
			restoreName = fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
			_, err = CreateRestoreWithoutCheck(restoreName, backupName, make(map[string]string), DestinationClusterName, destinationClusterUid, BackupOrgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Restore [%s] from backup %s", restoreName, backupName))
		})

		Step("Deleting the kopia executor pod while the restore is in progress", func() {
			log.InfoD("deleting the kopia executor pod while restore %s is in progress", restoreName)

			err = DeletePodWhileRestoreInProgress(ctx, BackupOrgID, restoreName, namespaces[0], KopiaRestoreExecutorPodLabel)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting kopia executor pod while restore %s is in progress", restoreName))
		})

		Step("Verifying restore status after deleting the kopia executor pod", func() {
			log.InfoD("Verifying restore %s status after deleting kopia executor pod", restoreName)

			err = RestoreSuccessCheck(restoreName, BackupOrgID, MaxWaitPeriodForRestoreCompletionInMinute*time.Minute, RestoreJobProgressRetryTime*time.Minute, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying restore %s taken from backup %v after deleting kopia executor pod", restoreName, backupName))
		})

	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)

		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		log.InfoD("Deleting deployed applications")
		err = DestroyAppsWithData(scheduledAppContexts, opts, controlChannel, errorGroup)
		log.FailOnError(err, "Data validations failed")

		err = SetSourceKubeConfig()
		log.FailOnError(err, "Switching context to destination cluster failed")

		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})

})

// this checks whether the backup and restore pods are scheduled on the labelled node during kdmp backup and restore
var _ = Describe("{BackupAndRestoreJobWithNodeAffinity}", Label(TestCaseLabelsMap[BackupAndRestoreJobWithNodeAffinity]...), func() {
	/*
	   Steps:
	   - Backup:
	       1. Update the kdmp-config map with the pxb_job_node_affinity_label on source cluster
	       2. Label one of the worker nodes of the source cluster with label pxb_job_node_affinity_label=True
	       3. Trigger the KDMP backup
	       4. Filter out the backup pods from the pods created in the namespace during the backup
	       5. Check whether these pods are scheduled on the labelled node
	       6. Verify backup success
	   - Restore:
	       1. Update the kdmp-config map with the pxb_job_node_affinity_label on destination cluster
	       2. Label one of the worker nodes of the destination cluster with label pxb_job_node_affinity_label=True
	       3. Trigger the restore
	       4. Filter out the restore pods from the pods created in the namespace during the restore
	       5. Check whether these pods are scheduled on the labelled node
	       6. Verify restore success
	*/
	var (
		providers                  []string
		scheduledAppContexts       []*scheduler.Context
		cloudCredName              string
		cloudCredUID               string
		backupLocationName         string
		backupLocationUID          string
		s3BackupLocationName       string
		s3BackupLocationUID        string
		nfsBackupLocationName      string
		nfsBackupLocationUID       string
		sourceClusterUid           string
		destClusterUid             string
		labelSelectors             map[string]string
		backupLocationMap          map[string]string
		sourceWorkerNodesList      []node.Node
		destinationWorkerNodesList []node.Node
		randomIndexSource          int
		randomIndexDestination     int
		wg                         sync.WaitGroup
		numberOfPVCsMap            map[string]int
		backupNames                []string
		restoreNames               []string
		backupName                 string
		restoreName                string
		labelToBeAdded             map[string]string
		controlChannel             chan string
		errorGroup                 *errgroup.Group
	)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("BackupAndRestoreJobWithNodeAffinity", "This checks whether kopia/nfs backup and restore pods are scheduled on the labelled node during kdmp backup and restore", nil, 302442, Bht, Q1FY26)
		backupLocationMap = make(map[string]string)
		labelSelectors = make(map[string]string)
		numberOfPVCsMap = make(map[string]int)
		providers = GetBackupProviders()
		labelToBeAdded = map[string]string{"pxb_job_node_affinity_label": "True"}
	})

	It("Kopia/nfs backup and restore job should have node affinity", func() {
		Step("Scheduling applications", func() {
			pipelineAppList := Inst().AppList
			defer func() {
				Inst().AppList = pipelineAppList
			}()
			Inst().AppList = []string{"mysql-backup-data"}
			log.InfoD("scheduling applications")
			scheduledAppContexts = make([]*scheduler.Context, 0)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
				appContexts := ScheduleApplications(taskName)
				for _, appCtx := range appContexts {
					appCtx.ReadinessTimeout = 20 * time.Minute
					scheduledAppContexts = append(scheduledAppContexts, appCtx)
				}
			}
		})

		Step("Validating applications", func() {
			log.InfoD("Validating applications")
			controlChannel, errorGroup = ValidateApplicationsStartData(scheduledAppContexts, context.TODO())
		})

		Step("Getting the number of PVCs created for each application", func() {
			log.InfoD("getting the number of PVCs created for each application before taking backup")
			// this is to know the number of backup pods and restore pods that will be created
			for _, appCtx := range scheduledAppContexts {
				namespace := appCtx.ScheduleOptions.Namespace
				pvcList, err := GetPVCListForNamespace(namespace)
				dash.VerifyFatal(err, nil, fmt.Sprintf("getting pvc list [%s] for namespace [%s] ", pvcList, namespace))
				numberOfPVCsMap[namespace] = len(pvcList)
				log.InfoD("Number of PVCs %d", numberOfPVCsMap[namespace])
			}
		})

		Step("Adding the pxb_job_node_affinity_label to kdmp-config map on source cluster", func() {
			log.InfoD("Adding the pxb_job_node_affinity_label to kdmp-config map on source cluster")
			for k, v := range labelToBeAdded {
				err := UpdateKDMPConfigMap(k, v)
				log.FailOnError(err, fmt.Sprintf("failed to update the kdmp-config map with label [%s/%s]", k, v))
			}
		})

		Step("Labeling one of the worker nodes on source cluster with label pxb_job_node_affinity_label=True", func() {
			log.InfoD("Labeling one of the worker nodes with label pxb_job_node_affinity_label=True")
			//get the worker nodes and the label to be added
			sourceWorkerNodesList = node.GetWorkerNodes()
			//generate a random index
			randomIndexSource = rand.Intn(len(sourceWorkerNodesList))
			//label the node
			err := AddLabelsOnNode(sourceWorkerNodesList[randomIndexSource], labelToBeAdded)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Add label on node %s on source cluster", sourceWorkerNodesList[randomIndexSource].Name))
			log.InfoD("The labelled node on source cluster is %s", sourceWorkerNodesList[randomIndexSource].Name)
		})

		// For restore part, creating the storage class on destination cluster, setting destination cluster kdmp-config, and labeling node on destination cluster
		Step("Create new storage class on destination cluster with similar spec as source cluster", func() {
			log.InfoD("Create new storage class on destination cluster with similar spec as source cluster")
			log.InfoD("Switching cluster context to destination cluster")
			err := SetDestinationKubeConfig()
			log.FailOnError(err, "Failed to set destination kubeconfig")

			for _, appCtx := range scheduledAppContexts {
				for _, spec := range appCtx.App.SpecList {
					switch obj := spec.(type) {
					case *storagev1.StorageClass:
						obj.ResourceVersion = ""
						_, err = storage.Instance().CreateStorageClass(obj)
						if err != nil {
							if errors.IsAlreadyExists(err) {
								log.Warnf("storage class [%s] already present on destination cluster", obj.Name)
							} else {
								log.FailOnError(err, fmt.Sprintf("Failed to create storage class [%s] on destination cluster", obj.Name))
							}
						} else {
							log.InfoD("Created storage class [%s] on destination cluster", obj.Name)
						}
					default:
						log.InfoD("App [%s] has spec of type [%T]", appCtx.App.Key, spec)
					}
				}
			}
		})

		Step("Adding the pxb_job_node_affinity_label to kdmp-config map on destination cluster", func() {
			log.InfoD("Adding the pxb_job_node_affinity_label to kdmp-config map on source cluster")
			for k, v := range labelToBeAdded {
				err := UpdateKDMPConfigMap(k, v)
				log.FailOnError(err, fmt.Sprintf("failed to update the kdmp-config map with label [%s/%s]", k, v))
			}
		})

		Step("Labelling one of the worker nodes on destination cluster with label pxb_job_node_affinity_label=True", func() {
			log.InfoD("Labeling one of the worker nodes on destination cluster with label pxb_job_node_affinity_label=True")
			defer func() {
				log.InfoD("Switching context to source cluster")
				err := SetSourceKubeConfig()
				dash.VerifyFatal(err, nil, "Setting source kubeconfig")
			}()
			destinationWorkerNodesList = node.GetWorkerNodes()
			//generate a random index
			randomIndexDestination = rand.Intn(len(destinationWorkerNodesList))
			err := AddLabelsOnNode(destinationWorkerNodesList[randomIndexDestination], labelToBeAdded)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Add label on node %s on source cluster", destinationWorkerNodesList[randomIndexDestination].Name))
			log.InfoD("The labelled node on destination cluster is %s", destinationWorkerNodesList[randomIndexDestination].Name)
		})

		// switched back to source cluster context
		Step("Creating backup location and cloud setting", func() {
			log.InfoD("Creating backup location and cloud setting")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, RandomString(6))
				cloudCredUID = uuid.New()
				backupLocationName = fmt.Sprintf("%s-%s-bl-%v", provider, getGlobalBucketName(provider), RandomString(6))
				backupLocationUID = uuid.New()
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocation(provider, backupLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, "Creating backup location")
				backupLocationMap[backupLocationUID] = backupLocationName
				if provider != drivers.ProviderNfs {
					nfsBackupLocationName = fmt.Sprintf("%s-%s-%v", "nfs", getGlobalBucketName(provider), RandomString(5))
					nfsBackupLocationUID = uuid.New()
					err = CreateNFSBackupLocation(nfsBackupLocationName, nfsBackupLocationUID, BackupOrgID, "", getGlobalBucketName(provider), true)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creating NFS backup location [%s]", nfsBackupLocationName))
					backupLocationMap[nfsBackupLocationUID] = nfsBackupLocationName
				} else {
					// Creating cloud cred again because in case of NFS as provider, cloud cred will not be created above
					err = CreateCloudCredential("aws", cloudCredName, cloudCredUID, BackupOrgID, ctx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential:[%s] for org [%s] for s3 backup location", cloudCredName, BackupOrgID))
					s3BackupLocationName = fmt.Sprintf("%s-%s-%v", "s3", getGlobalBucketName(provider), RandomString(5))
					s3BackupLocationUID = uuid.New()
					err = CreateS3BackupLocation(s3BackupLocationName, s3BackupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of S3 backup location [%s]", s3BackupLocationName))
					backupLocationMap[s3BackupLocationUID] = s3BackupLocationName
				}
			}
		})

		Step("Registering cluster for backup", func() {
			log.InfoD("Registering cluster for backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

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

		for backupLocationUID, backupLocationName := range backupLocationMap {
			namespaceMapping := make(map[string]string)

			Step("Taking backup of all applications from source cluster", func() {
				log.InfoD("Taking backup of applications")
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-central-admin ctx")

				backupName = fmt.Sprintf("%s-%s-%v", "autogenerated-backup", "for-all-applications", RandomString(5))
				log.InfoD("creating backup [%s] in source cluster [%s] (%s), organization [%s], for all applications, in backup location [%s]", backupName, SourceClusterName, sourceClusterUid, BackupOrgID, backupLocationName)

				_, err = CreateBackupWithoutCheck(ctx, backupName, SourceClusterName, backupLocationName, backupLocationUID, scheduledAppContexts, labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Triggered backup creation, with name [%s]", backupName))
				backupNames = append(backupNames, backupName)
			})

			Step("Verifying whether the backup pods are scheduled on the labelled node ", func() {
				log.InfoD("Verifying whether the backup pods are scheduled on the labelled node")
				// using for loop to iterate through the namespaces in case multiple applications are used
				for _, appCtx := range scheduledAppContexts {
					scheduledNamespace := appCtx.ScheduleOptions.Namespace
					//a hashmap to avoid checking the backup pods already checked
					backupPodsHashMap := make(map[string]bool)
					//A map to store the scheduling details of backup pods
					backupPodsScheduleDetails := make(map[string]map[string]interface{})

					temp := func() (interface{}, bool, error) {
						backupPodsList, err := core.Instance().GetPods(scheduledNamespace, KopiaBackupExecutorPodLabel)
						if err != nil {
							return nil, true, fmt.Errorf("failed to list backup pods with [%v] label in [%s] namespace", KopiaBackupExecutorPodLabel, scheduledNamespace)
						}
						if len(backupPodsList.Items) == 0 {
							return backupPodsList, true, fmt.Errorf("backup pods not created yet")
						}
						log.InfoD("The backup pods are :-\n %v", backupPodsList.Items)
						for _, p := range backupPodsList.Items {
							if _, ok := backupPodsHashMap[p.Name]; !ok {
								backupPodsHashMap[p.Name] = true
								//Check whether the pod p is scheduled on correct node or not
								log.InfoD("Expecting the backup pod [%s] to be scheduled on node [%s]", p.Name, sourceWorkerNodesList[randomIndexSource].Name)
								scheduledNodeName := p.Spec.NodeName
								if scheduledNodeName == sourceWorkerNodesList[randomIndexSource].Name {
									log.InfoD("The backup pod [%s] scheduled on the correct node", p.Name)
									backupPodsScheduleDetails[p.Name] = map[string]interface{}{
										"isScheduledProperly": true,
										"errorMsg":            nil,
									}
								} else {
									log.InfoD("The backup pod [%s] is NOT scheduled on the correct node", p.Name)
									backupPodsScheduleDetails[p.Name] = map[string]interface{}{
										"isScheduledProperly": false,
										"errorMsg":            fmt.Errorf("the backup pod [%s] NOT scheduled on the correct node, instead scheduled on [%s]", p.Name, scheduledNodeName),
									}
								}
							}
						}
						// if the all the backup pods are checked
						if len(backupPodsHashMap) == numberOfPVCsMap[scheduledNamespace] && len(backupPodsScheduleDetails) == numberOfPVCsMap[scheduledNamespace] {
							log.InfoD("Details:\n %+v", backupPodsScheduleDetails)
							// checking if any pods is not scheduled, if not scheduled return the error
							for _, status := range backupPodsScheduleDetails {
								if status["isScheduledProperly"].(bool) != true {
									return nil, false, status["errorMsg"].(error)
								}
							}
							return backupPodsList, false, nil
						} else {
							return backupPodsList, true, fmt.Errorf("waiting for more backup pods")
						}
					}
					_, err := task.DoRetryWithTimeout(temp, time.Duration(numberOfPVCsMap[scheduledNamespace])*30*time.Second, 2*time.Second)
					dash.VerifyFatal(err, nil, "Verifying backup pods are scheduled on correct node")
				}
			})

			Step("Check whether the triggered backup was successful or not", func() {
				log.InfoD("Checking whether the triggered backup was successful")
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-central-admin ctx")
				err = BackupSuccessCheckWithValidation(ctx, backupName, scheduledAppContexts, BackupOrgID, MaxWaitPeriodForBackupCompletionInMinutes*time.Minute, 30*time.Second)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of success and Validation of the backup [%s]", backupName))
			})

			// Restore part
			Step("Restoring applications to destination cluster", func() {
				log.InfoD("Restoring the backed up namespaces")
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-central-admin ctx")

				restoreName = fmt.Sprintf("%s-%s-%v", "test-restore", "of-all-applications", time.Now().Unix())
				// creating the namespaceMapping map
				for _, appCtx := range scheduledAppContexts {
					scheduledNamespace := appCtx.ScheduleOptions.Namespace
					namespaceMapping[scheduledNamespace] = fmt.Sprintf("%s-%v", scheduledNamespace, time.Now().Unix())
				}
				_, err = CreateRestoreWithoutCheck(restoreName, backupName, namespaceMapping, DestinationClusterName, destClusterUid, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Triggered restore with name [%s] for the backup [%s] with the follwing mapping [%s]", restoreName, backupName, namespaceMapping))
				restoreNames = append(restoreNames, restoreName)
			})

			Step("Check whether the restore pods are scheduled on correct node", func() {
				log.InfoD("Checking whether the restore pods are scheduled on correct node")

				log.InfoD("Switching cluster context to destination cluster")
				err := SetDestinationKubeConfig()
				dash.VerifyFatal(err, nil, "Setting destination kubeconfig")

				defer func() {
					log.InfoD("Switching cluster context to source cluster")
					err = SetSourceKubeConfig()
					dash.VerifyFatal(err, nil, "Setting source kubeconfig")
				}()

				for sourceNamespace, destinationNamespace := range namespaceMapping {
					// a hashmap to avoid checking the restore pods already checked
					restorePodHashMap := make(map[string]bool)
					//A map to store the scheduling details of backup pods
					restorePodsScheduleDetails := make(map[string]map[string]interface{})

					temp := func() (interface{}, bool, error) {
						restorePodsList, err := core.Instance().GetPods(destinationNamespace, KopiaRestoreExecutorPodLabel)
						if err != nil {
							return nil, true, fmt.Errorf("failed to list restore pods with [%v] label in [%s] namespace", KopiaRestoreExecutorPodLabel, destinationNamespace)
						}
						if len(restorePodsList.Items) == 0 {
							return restorePodsList, true, fmt.Errorf("restore pods not created yet")
						}
						log.InfoD("The restore pods are :-\n %v", restorePodsList.Items)
						for _, p := range restorePodsList.Items {
							if _, ok := restorePodHashMap[p.Name]; !ok {
								restorePodHashMap[p.Name] = true
								//Check whether the pod p is scheduled on correct node or not
								log.InfoD("Expecting the restore pod [%s] to be scheduled on node [%s]", p.Name, destinationWorkerNodesList[randomIndexDestination].Name)
								scheduledNodeName := p.Spec.NodeName
								if scheduledNodeName == destinationWorkerNodesList[randomIndexDestination].Name {
									log.InfoD("The restore pod [%s] scheduled on the correct node", p.Name)
									restorePodsScheduleDetails[p.Name] = map[string]interface{}{
										"isScheduledProperly": true,
										"errorMsg":            nil,
									}
								} else {
									log.InfoD("The restore pod [%s] is NOT scheduled on the correct node", p.Name)
									restorePodsScheduleDetails[p.Name] = map[string]interface{}{
										"isScheduledProperly": false,
										"errorMsg":            fmt.Errorf("the restore pod [%s] NOT scheduled on the correct node, instead scheduled on [%s]", p.Name, scheduledNodeName),
									}
								}
							}
						}
						// if the all the restore pods are checked
						if len(restorePodHashMap) == numberOfPVCsMap[sourceNamespace] && len(restorePodsScheduleDetails) == numberOfPVCsMap[sourceNamespace] {
							log.InfoD("Details:\n %+v", restorePodsScheduleDetails)
							// checking if any pods is not scheduled, if not scheduled return the error
							for _, status := range restorePodsScheduleDetails {
								if status["isScheduledProperly"].(bool) != true {
									return nil, false, status["errorMsg"].(error)
								}
							}
							return restorePodsList, false, nil
						} else {
							return restorePodsList, true, fmt.Errorf("waiting for more restore pods")
						}
					}
					_, err = task.DoRetryWithTimeout(temp, time.Duration(numberOfPVCsMap[sourceNamespace])*180*time.Second, 2*time.Second)
					dash.VerifyFatal(err, nil, "Verifying restore pods are scheduled on correct node")
				}
			})

			Step("Check whether the triggered restore was successful or not", func() {
				log.InfoD("Checking whether the triggered restore was successful or not")
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-central-admin ctx")
				err = RestoreSuccessCheck(restoreName, BackupOrgID, MaxWaitPeriodForRestoreCompletionInMinute*time.Minute, 30*time.Second, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Inspecting restore success for - [%s]", restoreName))
			})
		}
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		err := SetDestinationKubeConfig()
		dash.VerifyFatal(err, nil, "Setting destination kubeconfig")

		log.InfoD("Disabling the pxb_job_node_affinity_label in kdmp-config map on destination cluster")
		err = UpdateKDMPConfigMap("pxb_job_node_affinity_label", "False")
		dash.VerifyFatal(err, nil, fmt.Sprintf("Disabling the pxb_job_node_affinity_label KDMP config map"))

		// removing the label on node in destination cluster
		err = Inst().S.RemoveLabelOnNode(destinationWorkerNodesList[randomIndexDestination], "pxb_job_node_affinity_label")
		log.FailOnError(err, "Removing the label on destination worker node")

		err = SetSourceKubeConfig()
		dash.VerifyFatal(err, nil, "Setting source kubeconfig")
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")

		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		log.InfoD("Deleting deployed applications")
		err = DestroyAppsWithData(scheduledAppContexts, opts, controlChannel, errorGroup)
		log.FailOnError(err, "Data validations failed")

		log.InfoD("Disabling the pxb_job_node_affinity_label in kdmp-config map on source cluster")
		err = UpdateKDMPConfigMap("pxb_job_node_affinity_label", "False")
		dash.VerifyFatal(err, nil, fmt.Sprintf("Disabling the pxb_job_node_affinity_label KDMP config map"))

		// removing the label on the source node
		err = Inst().S.RemoveLabelOnNode(sourceWorkerNodesList[randomIndexSource], "pxb_job_node_affinity_label")
		log.FailOnError(err, "Removing the label on source worker node")

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

		restoreNames, err := GetAllRestoresAdmin()
		dash.VerifySafely(err, nil, fmt.Sprintf("Fetching all backups for admin"))

		for _, restoreName := range restoreNames {
			err = DeleteRestore(restoreName, BackupOrgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting restore [%s]", restoreName))
		}
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})
