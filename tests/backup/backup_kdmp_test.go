package tests

import (
	"fmt"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/pborman/uuid"

	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/volume/portworx/schedops"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

var _ = Describe("{ExcludeDirectoryFileBackup}", func() {
	var (
		backupName           string
		scheduledAppContexts []*scheduler.Context
		AppContextsMapping   map[string]*scheduler.Context
		namespace            string
		bkpNamespaces        []string
		backupNames          []string
		restoreNames         []string
		scheduleNames        []string

		clusterUid        string
		clusterStatus     api.ClusterInfo_StatusInfo_Status
		restoreName       string
		cloudCredName     string
		cloudCredUID      string
		backupLocationUID string
		bkpLocationName   string

		numDeployments int

		providers         []string
		backupLocationMap map[string]string
		labelSelectors    map[string]string

		storageClassExcludeFileDirMap map[string][]string
		mountPathExcludeFileDirMap    map[string][]string
		existingFileDirMountPathMap   map[string][]string
		podMountPathMap               map[string][]string
		fileListMountMap              map[string][]string
		dirListMountMap               map[string][]string
		masterDirFileList             map[string][]string
		scMountPathMap                map[string]string
		backupNamespaceMap            map[string]string
		formattedFileString           string
		fileList                      []string
		dirList                       []string
		preRuleName                   string
		postRuleName                  string
		preRuleUid                    string
		postRuleUid                   string
		periodicSchedulePolicyName    string
		periodicSchedulePolicyUid     string
		mutex                         sync.Mutex
		restoredNamespaces            []string
		wg                            sync.WaitGroup
	)
	JustBeforeEach(func() {
		backupNames = make([]string, 0)
		restoreNames = make([]string, 0)
		scheduleNames = make([]string, 0)
		bkpNamespaces = make([]string, 0)
		restoredNamespaces = make([]string, 0)
		backupLocationMap = make(map[string]string)
		labelSelectors = make(map[string]string)
		AppContextsMapping = make(map[string]*scheduler.Context)
		storageClassExcludeFileDirMap = make(map[string][]string)
		mountPathExcludeFileDirMap = make(map[string][]string)
		existingFileDirMountPathMap = make(map[string][]string)
		podMountPathMap = make(map[string][]string)
		fileListMountMap = make(map[string][]string)
		masterDirFileList = make(map[string][]string)
		scheduleNames = make([]string, 0)
		dirListMountMap = make(map[string][]string)
		scMountPathMap = make(map[string]string)
		backupNamespaceMap = make(map[string]string)
		numDeployments = 1
		providers = getProviders()

		StartPxBackupTorpedoTest("ExcludeDirectoryFileBackup", "Excludes mentioned directories or files from backed-up apps and restores them", nil, 93691, Ak, Q4FY24)

		log.InfoD(fmt.Sprintf("App list %v", Inst().AppList))
		scheduledAppContexts = make([]*scheduler.Context, 0)
		log.InfoD("Starting to deploy applications")
		for i := 0; i < numDeployments; i++ {
			log.InfoD(fmt.Sprintf("Iteration %v of deploying applications", i))
			taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = appReadinessTimeout
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
					containerMountPathList := make([]string, 0)
					containerPaths := schedops.GetContainerPVCMountMap(pod)
					for containerName, paths := range containerPaths {
						log.Infof(fmt.Sprintf("containerName:%s,paths: %s ", containerName, paths))
						containerMountPathList = append(containerMountPathList, paths...)
					}
					if len(containerMountPathList) > 0 {
						podMountPathMap[pod.Name] = containerMountPathList
						log.Infof(fmt.Sprintf("the list of mountPath within pod [%v] are [%v]", pod.Name, containerMountPathList))
					}
					scMountPathMap, err = schedops.GetContainerPVCMountMapWithSC(pod)
					dash.VerifyFatal(err, nil, fmt.Sprintf("getting storage class and mountpath mapping for pod [%s] ", pod.Name))
				}
			}
		})

		Step("Fetch the existing directories and files within mountPath before writing files and directories", func() {
			log.InfoD("Fetch the existing directories and files within mountPath before writing files and directories")
			for _, namespace := range bkpNamespaces {
				pods, err := core.Instance().GetPods(namespace, nil)
				dash.VerifyFatal(err, nil, fmt.Sprintf("getting pods from namespace [%s] ", namespace))
				for _, pod := range pods.Items {
					if len(podMountPathMap[pod.Name]) > 0 {
						existingFileDirList := make([]string, 0)
						for _, mountPath := range podMountPathMap[pod.Name] {
							log.Infof(fmt.Sprintf("Fetch the existing directories and files within mountPath [%s] before writing files and directories", mountPath))
							existingFileList, existingDirList, err := FetchFilesAndDirectoriesFromPod(pod, mountPath, nil)
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
					if len(podMountPathMap[pod.Name]) > 0 {
						for _, mountPath := range podMountPathMap[pod.Name] {
							scName := scMountPathMap[mountPath]
							DirectoryConfig := PodDirectoryConfig{
								BasePath:          mountPath,
								Depth:             10,
								Levels:            3,
								FilesPerDirectory: 1000,
							}
							log.Infof(fmt.Sprintf("creating nested directories and files within mountPath [%s] with depth [%d] , level [%d] and FilesPerDirectory [%d]", DirectoryConfig.BasePath, DirectoryConfig.Depth, DirectoryConfig.Levels, DirectoryConfig.FilesPerDirectory))
							err = CreateNestedDirectoriesWithFilesInPod(pod, DirectoryConfig)
							dash.VerifyFatal(err, nil, fmt.Sprintf("creating nested directories and files at mountpath [%v] for pod [%v] in namespace [%v]", mountPath, pod.Name, pod.Namespace))
							log.Infof(fmt.Sprintf("Fetching files and directories from path [%s] by excluding existing directories", DirectoryConfig.BasePath))
							fileList, dirList, err = FetchFilesAndDirectoriesFromPod(pod, mountPath, existingFileDirMountPathMap[mountPath])
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
							err = CreateFilesInPodDirectory(pod, fileConfig)
							dash.VerifyFatal(err, nil, fmt.Sprintf("creating files within mountPath [%s] with extensions and add to exclude list", mountPath))
							storageClassExcludeFileDirMap[scName] = append(storageClassExcludeFileDirMap[scName], fileConfig.FileName)
							mountPathExcludeFileDirMap[mountPath] = append(mountPathExcludeFileDirMap[mountPath], fileConfig.FileName)
							log.Infof(fmt.Sprintf("creating file within mountPath [%s] with hidden type", mountPath))
							fileConfig = PodDirectoryConfig{
								BasePath:          mountPath,
								FilesPerDirectory: 1,
								FileName:          ".snapshot",
							}
							err = CreateFilesInPodDirectory(pod, fileConfig)
							dash.VerifyFatal(err, nil, fmt.Sprintf("creating files within mountPath [%s] with hidden type and add to exclude list", mountPath))
							storageClassExcludeFileDirMap[scName] = append(storageClassExcludeFileDirMap[scName], fileConfig.FileName)
							mountPathExcludeFileDirMap[mountPath] = append(mountPathExcludeFileDirMap[mountPath], fileConfig.FileName)
							log.Infof(fmt.Sprintf("creating file within mountPath [%s] with valid special chars ", mountPath))
							fileConfig = PodDirectoryConfig{
								BasePath:          mountPath,
								FilesPerDirectory: 1,
								FileName:          "myn@meisunkn*wn",
							}
							err = CreateFilesInPodDirectory(pod, fileConfig)
							dash.VerifyFatal(err, nil, fmt.Sprintf("creating files within mountPath [%s] valid special chars and add to exclude list", mountPath))
							storageClassExcludeFileDirMap[scName] = append(storageClassExcludeFileDirMap[scName], fileConfig.FileName)
							mountPathExcludeFileDirMap[mountPath] = append(mountPathExcludeFileDirMap[mountPath], fileConfig.FileName)

							log.Infof(fmt.Sprintf("creating file within mountPath [%s] with maximum name length (255 characters)", mountPath))
							fileConfig = PodDirectoryConfig{
								BasePath:          mountPath,
								FilesPerDirectory: 1,
								FileName:          fmt.Sprintf("%s.txt", RandomString(251)),
							}
							err = CreateFilesInPodDirectory(pod, fileConfig)
							dash.VerifyFatal(err, nil, fmt.Sprintf("creating files within mountPath [%s] with maximum name length (255 characters and add to exclude list)", mountPath))
							storageClassExcludeFileDirMap[scName] = append(storageClassExcludeFileDirMap[scName], fileConfig.FileName)
							mountPathExcludeFileDirMap[mountPath] = append(mountPathExcludeFileDirMap[mountPath], fileConfig.FileName)

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
					if len(podMountPathMap[pod.Name]) > 0 {
						for _, mountPath := range podMountPathMap[pod.Name] {
							excludeFileDirList := make([]string, 0)
							log.Infof(fmt.Sprintf("Fetch some random directories from created list %v", dirListMountMap[mountPath]))
							randomDirs, err := GetRandomSubset(dirListMountMap[mountPath], 100)
							dash.VerifyFatal(err, nil, fmt.Sprintf("Getting random directories from the list"))
							log.Infof(fmt.Sprintf("the list of directories randomly selected from mountPath- %v : %v", mountPath, randomDirs))
							log.Infof(fmt.Sprintf("Fetch some random files from created list %v", fileListMountMap[mountPath]))
							randomFiles, err := GetRandomSubset(fileListMountMap[mountPath], 100)
							dash.VerifyFatal(err, nil, fmt.Sprintf("Getting random files from the list"))
							log.Infof(fmt.Sprintf("the list of files randomly selected from mountPath- %v : %v", mountPath, randomFiles))
							excludeFileDirList = append(excludeFileDirList, randomDirs...)
							excludeFileDirList = append(excludeFileDirList, randomFiles...)
							scName := scMountPathMap[mountPath]
							storageClassExcludeFileDirMap[scName] = append(storageClassExcludeFileDirMap[scName], excludeFileDirList...)
							mountPathExcludeFileDirMap[mountPath] = append(mountPathExcludeFileDirMap[mountPath], excludeFileDirList...)
						}
					}
				}
			}
			log.Infof(fmt.Sprintf("create formatted string with storage class name and exclude file and directories list [%v]", storageClassExcludeFileDirMap))
			formattedFileString = GenerateStorageClassFormattedString(storageClassExcludeFileDirMap)
			err := UpdateKDMPConfigMap("KDMP_EXCLUDE_FILE_LIST", formattedFileString)
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
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, orgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, orgID, provider))
				err = CreateBackupLocation(provider, bkpLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), orgID, "", true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", bkpLocationName))
			}
		})
		Step("Registering cluster for backup", func() {
			log.InfoD("Registering cluster for backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = CreateApplicationClusters(orgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(orgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			clusterUid, err = Inst().Backup.GetClusterUID(ctx, orgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid", SourceClusterName))
		})

		Step(fmt.Sprintf("Verify creation of pre and post exec rules for applications "), func() {
			log.InfoD("Verify creation of pre and post exec rules for applications ")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			preRuleName, postRuleName, err = CreateRuleForBackupWithMultipleApplications(orgID, Inst().AppList, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of pre and post exec rules for applications from px-admin"))
			if preRuleName != "" {
				preRuleUid, err = Inst().Backup.GetRuleUid(orgID, ctx, preRuleName)
				log.FailOnError(err, "Fetching pre backup rule [%s] uid", preRuleName)
				log.Infof("Pre backup rule [%s] uid: [%s]", preRuleName, preRuleUid)
			}
			if postRuleName != "" {
				postRuleUid, err = Inst().Backup.GetRuleUid(orgID, ctx, postRuleName)
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
			err = CreateBackupScheduleIntervalPolicy(5, periodicSchedulePolicyInterval, 5, periodicSchedulePolicyName, periodicSchedulePolicyUid, orgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of periodic schedule policy of interval [%v] minutes named [%s] ", periodicSchedulePolicyInterval, periodicSchedulePolicyName))

		})

		Step("Taking manual backup of namespaces with rules", func() {
			log.InfoD(fmt.Sprintf("Taking manual backup of namespaces with rules"))
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, namespace := range bkpNamespaces {
				backupName = fmt.Sprintf("%s-%v", BackupNamePrefix, time.Now().Unix())
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
				err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, bkpLocationName, backupLocationUID, appContextsToBackup, labelSelectors, orgID, clusterUid, preRuleName, preRuleUid, postRuleName, postRuleUid)
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
				scheduleName := fmt.Sprintf("%s-schedule-with-rules-%s", BackupNamePrefix, RandomString(4))
				log.InfoD("Creating a schedule backup of namespace [%s] without pre and post exec rules", namespace)
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
				scheduleBackupName, err := CreateScheduleBackupWithValidation(ctx, scheduleName, SourceClusterName, bkpLocationName, backupLocationUID, appContextsToBackup,
					labelSelectors, orgID, "", "", "", "", periodicSchedulePolicyName, periodicSchedulePolicyUid)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of backup [%s]", scheduleBackupName))
				err = suspendBackupSchedule(scheduleName, periodicSchedulePolicyName, orgID, ctx)
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
						map[string]string{backupNamespaceMap[backupName]: fmt.Sprintf("custom1-%s-%d", backupNamespaceMap[backupName], index)},
						make(map[string]string),
						ReplacePolicyRetain,
					},
					{
						"test-replace-restore-single-ns",
						map[string]string{backupNamespaceMap[backupName]: fmt.Sprintf("custom1-rep-%s-%d", backupNamespaceMap[backupName], index)},
						make(map[string]string),
						ReplacePolicyDelete,
					},
				}
				for _, config := range restoreConfigs {
					restoreName := fmt.Sprintf("%s-%s", config.namePrefix, RandomString(4))
					log.InfoD("Restoring backup [%s] in cluster [%s] with restore [%s] and namespace mapping %v", backupName, destinationClusterName, restoreName, config.namespaceMapping)
					if config.replacePolicy == ReplacePolicyRetain {
						appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{backupNamespaceMap[backupName]})
						restoredNamespaces = append(restoredNamespaces, config.namespaceMapping[backupNamespaceMap[backupName]])
						err = CreateRestoreWithValidation(ctx, restoreName, backupName, config.namespaceMapping, config.storageClassMapping, destinationClusterName, orgID, appContextsToBackup)
					} else if config.replacePolicy == ReplacePolicyDelete {
						restoredNamespaces = append(restoredNamespaces, config.namespaceMapping[backupNamespaceMap[backupName]])
						err = CreateRestoreWithReplacePolicy(restoreName, backupName, config.namespaceMapping, destinationClusterName, orgID, ctx, config.storageClassMapping, config.replacePolicy)
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
					log.Infof(fmt.Sprintf("verfiying the files and directories for pod [%s] in restored namespace [%s] ", pod.Name, restoredNamespace))
					containerMountPathList := make([]string, 0)
					containerPaths := schedops.GetContainerPVCMountMap(pod)
					if len(containerPaths) > 0 {
						for containerName, paths := range containerPaths {
							log.Infof(fmt.Sprintf("containerName:%s,paths: %s ", containerName, paths))
							containerMountPathList = append(containerMountPathList, paths...)
						}

						log.Infof(fmt.Sprintf("the list of mountPath within pod [%v] are [%v]", pod.Name, containerMountPathList))
						for _, containerMountPath := range containerMountPathList {
							restoredCombinedList := make([]string, 0)
							fileList, dirList, err := FetchFilesAndDirectoriesFromPod(pod, containerMountPath, existingFileDirMountPathMap[containerMountPath])
							dash.VerifyFatal(err, nil, fmt.Sprintf("fetching files and directory from mountpath [%s] for pod [%s]", containerMountPath, pod.Name))
							log.Infof(fmt.Sprintf("The list of files created in mountPath [%v] for pod [%v]: %v", containerMountPath, pod.Name, fileList))
							log.Infof(fmt.Sprintf("The list of directories created in mountPath [%v] for pod [%v]: %v", containerMountPath, pod.Name, dirList))
							restoredCombinedList = append(restoredCombinedList, fileList...)
							restoredCombinedList = append(restoredCombinedList, dirList...)
							log.Infof(fmt.Sprintf("the list of combined directories and files after restore: %v", restoredCombinedList))
							for _, item := range mountPathExcludeFileDirMap[containerMountPath] {
								if item != "" {
									if !IsPresent(restoredCombinedList, item) {
										log.Infof(fmt.Sprintf("the item file/directory [%s] is not present in the mountPath[%s] for pod [%s] in namespace [%s]", item, containerMountPath, pod.Name, restoredNamespace))
									} else {
										err := fmt.Errorf("the item file/directory[%s] is still present in mountPath [%s] for pod [%s] in namespace [%s]", item, containerMountPath, pod.Name, restoredNamespace)
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
					if len(podMountPathMap[pod.Name]) > 0 {
						for _, mountPath := range podMountPathMap[pod.Name] {
							DirectoryConfig := PodDirectoryConfig{
								BasePath:          mountPath,
								Depth:             50,
								Levels:            1,
								FilesPerDirectory: 50,
							}
							log.Infof(fmt.Sprintf("creating nested directories and files within mountPath [%s] with depth [%d] , level [%d] and FilesPerDirectory [%d]", DirectoryConfig.BasePath, DirectoryConfig.Depth, DirectoryConfig.Levels, DirectoryConfig.FilesPerDirectory))
							err = CreateNestedDirectoriesWithFilesInPod(pod, DirectoryConfig)
							dash.VerifyFatal(err, nil, fmt.Sprintf("creating nested directories and files at mountpath [%v] for pod [%v] in namespace [%v]", mountPath, pod.Name, pod.Namespace))
							log.Infof(fmt.Sprintf("Fetching files and directories from path [%s] by excluding existing directories", DirectoryConfig.BasePath))
							fileList, dirList, err := FetchFilesAndDirectoriesFromPod(pod, mountPath, append(existingFileDirMountPathMap[mountPath], masterDirFileList[mountPath]...))
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
					if len(podMountPathMap[pod.Name]) > 0 {
						for _, mountPath := range podMountPathMap[pod.Name] {
							excludeFileDirList := make([]string, 0)
							log.Infof(fmt.Sprintf("Fetch some random directories from created list %v", dirListMountMap[mountPath]))
							randomDirs, err := GetRandomSubset(dirListMountMap[mountPath], 10)
							dash.VerifyFatal(err, nil, fmt.Sprintf("Getting random directories from the list"))
							log.Infof(fmt.Sprintf("the list of directories randomly selected from mountPath- %v : %v", mountPath, randomDirs))
							log.Infof(fmt.Sprintf("Fetch some random files from created list %v", fileListMountMap[mountPath]))
							randomFiles, err := GetRandomSubset(fileListMountMap[mountPath], 10)
							dash.VerifyFatal(err, nil, fmt.Sprintf("Getting random files from the list"))
							log.Infof(fmt.Sprintf("the list of files randomly selected from mountPath- %v : %v", mountPath, randomFiles))
							excludeFileDirList = append(excludeFileDirList, randomDirs...)
							excludeFileDirList = append(excludeFileDirList, randomFiles...)
							scName := scMountPathMap[mountPath]
							storageClassExcludeFileDirMap[scName] = append(storageClassExcludeFileDirMap[scName], excludeFileDirList...)
							mountPathExcludeFileDirMap[mountPath] = append(mountPathExcludeFileDirMap[mountPath], excludeFileDirList...)
						}
					}
				}
			}
			log.Infof(fmt.Sprintf("create new formatted string with storage class name and exclude file ,directories list [%v]", storageClassExcludeFileDirMap))
			formattedFileString = GenerateStorageClassFormattedString(storageClassExcludeFileDirMap)
			err := UpdateKDMPConfigMap("KDMP_EXCLUDE_FILE_LIST", formattedFileString)
			dash.VerifyFatal(err, nil, fmt.Sprintf("updating KDMP config map"))
		})

		Step("Taking manual backup of namespaces with rules ,iteration 2", func() {
			log.InfoD(fmt.Sprintf("Taking manual backup of namespaces with rules, iteration 2"))
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			backupNames := make([]string, 0)
			for _, namespace := range bkpNamespaces {
				backupName = fmt.Sprintf("%s-%v", BackupNamePrefix, time.Now().Unix())
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
				err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, bkpLocationName, backupLocationUID, appContextsToBackup, labelSelectors, orgID, clusterUid, preRuleName, preRuleUid, postRuleName, postRuleUid)
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
				restoreName = fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
				restoreNamespace := "custom2-" + backupNamespaceMap[backupName]
				namespaceMapping := map[string]string{backupNamespaceMap[backupName]: restoreNamespace}
				restoredNamespaces = append(restoredNamespaces, restoreNamespace)
				err = CreateRestoreWithValidation(ctx, restoreName, backupName, namespaceMapping, make(map[string]string), destinationClusterName, orgID, appContextsToBackup)
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
					log.Infof(fmt.Sprintf("verfiying the files and directories for pod [%s] in restored namespace [%s] ", pod.Name, restoredNamespace))
					containerMountPathList := make([]string, 0)
					containerPaths := schedops.GetContainerPVCMountMap(pod)
					if len(containerPaths) > 0 {
						for containerName, paths := range containerPaths {
							log.Infof(fmt.Sprintf("containerName:%s,paths: %s ", containerName, paths))
							containerMountPathList = append(containerMountPathList, paths...)
						}

						log.Infof(fmt.Sprintf("the list of mountPath within pod [%v] are [%v]", pod.Name, containerMountPathList))
						for _, containerMountPath := range containerMountPathList {
							restoredCombinedList := make([]string, 0)
							fileList, dirList, err := FetchFilesAndDirectoriesFromPod(pod, containerMountPath, existingFileDirMountPathMap[containerMountPath])
							dash.VerifyFatal(err, nil, fmt.Sprintf("fetching files and directory from mountpath [%s] for pod [%s]", containerMountPath, pod.Name))
							log.Infof(fmt.Sprintf("The list of files created in mountPath [%v] for pod [%v]: %v", containerMountPath, pod.Name, fileList))
							log.Infof(fmt.Sprintf("The list of directories created in mountPath [%v] for pod [%v]: %v", containerMountPath, pod.Name, dirList))
							restoredCombinedList = append(restoredCombinedList, fileList...)
							restoredCombinedList = append(restoredCombinedList, dirList...)
							log.Infof(fmt.Sprintf("the list of combined directories and files after restore: %v", restoredCombinedList))
							for _, item := range mountPathExcludeFileDirMap[containerMountPath] {
								if item != "" {
									if !IsPresent(restoredCombinedList, item) {
										log.Infof(fmt.Sprintf("the item(file/directory) [%s] is not present in the mountPath[%s] for pod [%s] in namespace [%s]", item, containerMountPath, pod.Name, restoredNamespace))
									} else {
										err := fmt.Errorf("the item(file/directory) [%s] is still present in mountPath [%s] for pod [%s] in namespace [%s]", item, containerMountPath, pod.Name, restoredNamespace)
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
			formattedFileString = ""
			err := UpdateKDMPConfigMap("KDMP_EXCLUDE_FILE_LIST", formattedFileString)
			dash.VerifyFatal(err, nil, fmt.Sprintf("updating KDMP config map"))
		})

		Step("Taking manual backup of namespaces with rules without excluding any files or directories", func() {
			log.InfoD(fmt.Sprintf("Taking manual backup of namespaces with rules without excluding any files or directories"))
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			backupNames = make([]string, 0)
			for _, namespace := range bkpNamespaces {
				backupName = fmt.Sprintf("%s-%v", BackupNamePrefix, time.Now().Unix())
				appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
				err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, bkpLocationName, backupLocationUID, appContextsToBackup, labelSelectors, orgID, clusterUid, preRuleName, preRuleUid, postRuleName, postRuleUid)
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
				restoreName = fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
				restoreNamespace := "custom3-" + backupNamespaceMap[backupName]
				namespaceMapping := map[string]string{backupNamespaceMap[backupName]: restoreNamespace}
				restoredNamespaces = append(restoredNamespaces, restoreNamespace)
				err = CreateRestoreWithValidation(ctx, restoreName, backupName, namespaceMapping, make(map[string]string), destinationClusterName, orgID, appContextsToBackup)
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
					log.Infof(fmt.Sprintf("verfiying all the files and directories created above for pod [%s] in restored namespace [%s] ", pod.Name, restoredNamespace))
					containerMountPathList := make([]string, 0)
					containerPaths := schedops.GetContainerPVCMountMap(pod)
					if len(containerPaths) > 0 {
						for containerName, paths := range containerPaths {
							log.Infof(fmt.Sprintf("containerName:%s,paths: %s ", containerName, paths))
							containerMountPathList = append(containerMountPathList, paths...)
						}

						log.Infof(fmt.Sprintf("the list of mountPath within pod [%v] are [%v]", pod.Name, containerMountPathList))
						for _, containerMountPath := range containerMountPathList {
							restoredCombinedList := make([]string, 0)
							fileList, dirList, err := FetchFilesAndDirectoriesFromPod(pod, containerMountPath, existingFileDirMountPathMap[containerMountPath])
							dash.VerifyFatal(err, nil, fmt.Sprintf("fetching files and directory from mountpath [%s] for pod [%s]", containerMountPath, pod.Name))
							log.Infof(fmt.Sprintf("The list of files created in mountPath [%v] for pod [%v]: %v", containerMountPath, pod.Name, fileList))
							log.Infof(fmt.Sprintf("The list of directories created in mountPath [%v] for pod [%v]: %v", containerMountPath, pod.Name, dirList))
							restoredCombinedList = append(restoredCombinedList, fileList...)
							restoredCombinedList = append(restoredCombinedList, dirList...)
							log.Infof(fmt.Sprintf("the list of combined directories and files after restore: %v", restoredCombinedList))
							for _, item := range masterDirFileList[containerMountPath] {
								if item != "" {
									if !IsPresent(restoredCombinedList, item) {
										err := fmt.Errorf("item(file/directory) [%s] is not present in mountPath [%s] for pod [%s] in namespace [%s]", item, containerMountPath, pod.Name, restoredNamespace)
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
				backupUid, err := Inst().Backup.GetBackupUID(ctx, backupName, orgID)
				_, err = DeleteBackup(backupName, backupUid, orgID, ctx)
				dash.VerifySafely(err, nil, fmt.Sprintf("Failed to delete the backup %s ", backupName))
				err = DeleteBackupAndWait(backupName, ctx)
				dash.VerifySafely(err, nil, fmt.Sprintf("waiting for backup [%s] deletion", backupName))
			}(backupName)
		}
		wg.Wait()
		for _, restoreName := range restoreNames {
			err = DeleteRestore(restoreName, orgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting restore [%s]", restoreName))
		}

		for _, scheduleName := range scheduleNames {
			err = DeleteSchedule(scheduleName, SourceClusterName, orgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting schedule [%s]", scheduleName))
		}

		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})
