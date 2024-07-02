package tests

import (
	"fmt"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/sched-ops/k8s/storage"
	"github.com/portworx/torpedo/drivers/scheduler/k8s"
	"github.com/portworx/torpedo/drivers/scheduler/rke"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	storageApi "k8s.io/api/storage/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

// EnableNsAndClusterLevelPSAWithBackupAndRestore verifies backup and restore of applications with namespace and cluster level PSA enabled on Vanilla Cluster
var _ = Describe("{EnableNsAndClusterLevelPSAWithBackupAndRestore}", Label(TestCaseLabelsMap[EnableNsAndClusterLevelPSAWithBackupAndRestore]...), func() {
	var (
		backupNames          []string
		backupNames2         []string
		restoreNames         []string
		appList              = Inst().AppList
		scheduledAppContexts []*scheduler.Context
		bkpNamespaces        []string
		label                map[string]string
		preRuleNameList      []string
		postRuleNameList     []string
		providers            []string
		sourceScNameList     []string
		cloudCredName        string
		cloudCredUID         string
		backupLocationUID    string
		backupLocationMap    map[string]string
		sourceClusterUid     string
		scName               string
		params               map[string]string
		backupNSMap          map[string]string
		controlChannel       chan string
		errorGroup           *errgroup.Group
		backupNamesAllNs     []string
		restoredNamespaces   []string
	)
	storageClassMapping := make(map[string]string)
	AppContextsMapping := make(map[string]*scheduler.Context)
	providers = GetBackupProviders()
	scheduledAppContexts = make([]*scheduler.Context, 0)
	bkpNamespaces = make([]string, 0)
	preRuleNameList = make([]string, 0)
	postRuleNameList = make([]string, 0)
	originalList := Inst().AppList
	label = make(map[string]string)
	backupLocationMap = make(map[string]string)
	psaFlag := false
	backupNSMap = make(map[string]string)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("EnableNsAndClusterLevelPSAWithBackupAndRestore", "Enable Namespace and cluster level PSA with Backup and Restore", nil, 299243, Kshithijiyer, Q2FY25)

		log.InfoD("Deploy applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		psaApp := make([]string, 0)
		for _, psalevel := range []string{"restricted", "baseline", "privileged"} {
			if psalevel == "restricted" {
				appList := Inst().AppList
				log.InfoD("The app list at the start of the testcase is %v", Inst().AppList)
				for _, app := range appList {
					psaApp = append(psaApp, PSAAppMap[app])
				}
				log.Infof("The PSA app list is %v", psaApp)
				Inst().AppList = psaApp
			}
			label["pod-security.kubernetes.io/enforce"] = psalevel

			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				taskName := fmt.Sprintf("%s-%d-%s-%d", TaskNamePrefix, 0, psalevel, i)
				namespace := fmt.Sprintf("%s-%d", psalevel, i)
				err := CreateNamespaceAndAssignLabels(namespace, label)
				dash.VerifyFatal(err, nil, "Creating namespace and assigning labels")
				appContexts := ScheduleApplicationsOnNamespace(namespace, taskName)
				for _, ctx := range appContexts {
					ctx.ReadinessTimeout = AppReadinessTimeout
					namespace := GetAppNamespace(ctx, taskName)
					bkpNamespaces = append(bkpNamespaces, namespace)
					scheduledAppContexts = append(scheduledAppContexts, ctx)
					AppContextsMapping[namespace] = ctx
				}
			}
			if psalevel == "restricted" {
				Inst().AppList = originalList
			}
		}
	})

	It("Enable Namespace and cluster level PSA with Backup and Restore", func() {

		Step("Validating applications", func() {
			log.InfoD("Validating applications")
			ctx, _ := backup.GetAdminCtxFromSecret()
			controlChannel, errorGroup = ValidateApplicationsStartData(scheduledAppContexts, ctx)
		})

		Step("Creating rules for backup", func() {
			log.InfoD("Creating pre rule for deployed apps")
			for i := 0; i < len(appList); i++ {
				preRuleStatus, ruleName, err := Inst().Backup.CreateRuleForBackup(appList[i], BackupOrgID, "pre")
				log.FailOnError(err, "Creating pre rule for deployed apps failed")
				dash.VerifyFatal(preRuleStatus, true, fmt.Sprintf("Verifying pre rule %s for backup", ruleName))
				if ruleName != "" {
					preRuleNameList = append(preRuleNameList, ruleName)
				}
			}

			log.InfoD("Creating post rule for deployed apps")
			for i := 0; i < len(appList); i++ {
				postRuleStatus, ruleName, err := Inst().Backup.CreateRuleForBackup(appList[i], BackupOrgID, "post")
				log.FailOnError(err, "Creating post rule for deployed apps failed")
				dash.VerifyFatal(postRuleStatus, true, fmt.Sprintf("Verifying post rule %s for backup", ruleName))
				if ruleName != "" {
					postRuleNameList = append(postRuleNameList, ruleName)
				}
			}
		})

		Step("Creating backup location and cloud setting", func() {
			log.InfoD("Creating backup location and cloud setting")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, RandomString(10))
				backupLocationName := fmt.Sprintf("%s-%s-bl-%v", provider, getGlobalBucketName(provider), time.Now().Unix())
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocationName
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocation(provider, backupLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, "Creating backup location")
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
		})

		Step("Taking backup of all the namespaces created with namespace level PSA", func() {
			log.InfoD("Taking backup of all the namespaces created with namespace level PSA")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			var wg sync.WaitGroup
			var mutex sync.Mutex
			labelSelectors := make(map[string]string)
			for backupLocationUID, backupLocationName := range backupLocationMap {
				for _, namespace := range bkpNamespaces {
					wg.Add(1)
					go func(namespace string) {
						defer wg.Done()
						defer GinkgoRecover()
						backupName := fmt.Sprintf("%s-%s", BackupNamePrefix, RandomString(10))
						appContextsToBackup := FilterAppContextsByNamespace(scheduledAppContexts, []string{namespace})
						preRuleUid, preRule := "", ""
						if len(preRuleNameList) > 0 {
							preRuleUid, err = Inst().Backup.GetRuleUid(BackupOrgID, ctx, preRuleNameList[0])
							log.FailOnError(err, "Fetching pre backup rule [%s] uid", preRuleNameList[0])
							preRule = preRuleNameList[0]
						}
						postRuleUid, postRule := "", ""
						if len(postRuleNameList) > 0 {
							postRuleUid, err = Inst().Backup.GetRuleUid(BackupOrgID, ctx, postRuleNameList[0])
							log.FailOnError(err, "Fetching post backup rule [%s] uid", postRuleNameList[0])
							postRule = postRuleNameList[0]
						}
						err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, backupLocationName, backupLocationUID, appContextsToBackup, labelSelectors, BackupOrgID, sourceClusterUid, preRule, preRuleUid, postRule, postRuleUid)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s] of namespace [%s]", backupName, namespace))
						mutex.Lock()
						backupNames = append(backupNames, backupName)
						backupNSMap[backupName] = namespace
						mutex.Unlock()
					}(namespace)
				}
			}
			wg.Wait()
		})

		Step("Create new storage class for restore", func() {
			log.InfoD("Getting storage class of the source cluster")
			for _, appNamespaces := range bkpNamespaces {
				pvcs, err := core.Instance().GetPersistentVolumeClaims(appNamespaces, make(map[string]string))
				log.FailOnError(err, "Getting PVC on source cluster")
				singlePvc := pvcs.Items[0]
				storageClass, err := core.Instance().GetStorageClassForPVC(&singlePvc)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Getting SC %v from PVC in source cluster",
					storageClass.Name))
				sourceScNameList = append(sourceScNameList, storageClass.Name)
			}

			err := SetDestinationKubeConfig()
			dash.VerifyFatal(err, nil, "Setting destination kubeconfig")

			for _, sc := range sourceScNameList {
				scName = fmt.Sprintf("replica-sc-%v", RandomString(3))
				v1obj := metaV1.ObjectMeta{
					Name: scName,
				}
				reclaimPolicyDelete := v1.PersistentVolumeReclaimDelete
				bindMode := storageApi.VolumeBindingImmediate
				scObj := storageApi.StorageClass{
					ObjectMeta:        v1obj,
					Provisioner:       k8s.CsiProvisioner,
					Parameters:        params,
					ReclaimPolicy:     &reclaimPolicyDelete,
					VolumeBindingMode: &bindMode,
				}
				_, err = storage.Instance().CreateStorageClass(&scObj)
				log.FailOnError(err, "Creating sc on dest cluster")
				storageClassMapping[sc] = scName
			}

			err = SetSourceKubeConfig()
			dash.VerifyFatal(err, nil, "Setting source kubeconfig")
		})

		Step("Default restore of applications by replacing the existing resources with NS level PSA", func() {
			log.InfoD("Default restore of applications by replacing the existing resources with NS level PSA")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			var wg sync.WaitGroup
			var mutex sync.Mutex
			for _, backupName := range backupNames {
				wg.Add(1)
				go func(backupName string) {
					defer wg.Done()
					defer GinkgoRecover()
					defaultRestoreName := fmt.Sprintf("%s-%s-default", RestoreNamePrefix, backupName)
					err = CreateRestoreWithReplacePolicyWithValidation(defaultRestoreName, backupName, make(map[string]string), SourceClusterName, BackupOrgID, ctx, make(map[string]string), ReplacePolicyDelete, scheduledAppContexts)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creating default restore for backup with replace policy [%s]", defaultRestoreName))
					mutex.Lock()
					restoreNames = append(restoreNames, defaultRestoreName)
					mutex.Unlock()
				}(backupName)
			}
			wg.Wait()
		})

		Step("Restore of applications with NS and StorageClass mapping with NS level PSA", func() {
			log.InfoD("Restore of applications with NS and StorageClass mapping with NS level PSA")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			for backupName, backupNamespace := range backupNSMap {
				appContextsToRestore := make([]*scheduler.Context, 0)
				namespaceMapping := make(map[string]string)
				customRestoreName := fmt.Sprintf("%s-%s-custom-ns-sc", RestoreNamePrefix, backupName)
				namespaceMapping[backupNamespace] = backupNamespace + "-restored-1"
				restoredNamespaces = append(restoredNamespaces, backupNamespace+"-restored-1")
				appContextsToRestore = FilterAppContextsByNamespace(scheduledAppContexts, []string{backupNamespace})
				err = CreateRestoreWithReplacePolicyWithValidation(customRestoreName, backupName, namespaceMapping, DestinationClusterName, BackupOrgID, ctx, storageClassMapping, ReplacePolicyDelete, appContextsToRestore)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Create restore with NS and StorageClass mapping with NS level PSA"))
				restoreNames = append(restoreNames, customRestoreName)
			}
		})

		for _, psaLevel := range []string{"restricted", "baseline", "privileged"} {
			Step(fmt.Sprintf("Setup PSA level to %s", psaLevel), func() {

				err := SwitchBothKubeConfigANDContext("destination")
				dash.VerifyFatal(err, nil, "Setting destination kubeconfig and context")

				err = ConfigureClusterLevelPSA(psaLevel, []string{})
				dash.VerifyFatal(err, nil, "Setting cluster level PSA configuration")

				err = VerifyClusterlevelPSA()
				dash.VerifyFatal(err, nil, "Verify cluster level PSA configuration")

				err = SwitchBothKubeConfigANDContext("source")
				dash.VerifyFatal(err, nil, "Setting source kubeconfig")

				err = ConfigureClusterLevelPSA(psaLevel, []string{})
				dash.VerifyFatal(err, nil, "Setting cluster level PSA configuration")

				err = VerifyClusterlevelPSA()
				dash.VerifyFatal(err, nil, "Verify cluster level PSA configuration")
				psaFlag = true

			})

			Step(fmt.Sprintf("Taking backup of all the namespaces created with namespace level PSA and Cluster level PSA Set to %s", psaLevel), func() {
				log.InfoD(fmt.Sprintf("Taking backup of all the namespaces created with namespace level PSA and Cluster level PSA Set to %s", psaLevel))

				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-central-admin ctx")

				var wg sync.WaitGroup
				var mutex sync.Mutex
				labelSelectors := make(map[string]string)
				for backupLocationUID, backupLocationName := range backupLocationMap {
					for _, namespace := range bkpNamespaces {
						wg.Add(1)
						go func(namespace string) {
							defer wg.Done()
							defer GinkgoRecover()
							backupName := fmt.Sprintf("%s-%s-%s", BackupNamePrefix, RandomString(10), psaLevel)
							preRuleUid, preRule := "", ""
							if len(preRuleNameList) > 0 {
								preRuleUid, err = Inst().Backup.GetRuleUid(BackupOrgID, ctx, preRuleNameList[0])
								log.FailOnError(err, "Fetching pre backup rule [%s] uid", preRuleNameList[0])
								preRule = preRuleNameList[0]
							}
							postRuleUid, postRule := "", ""
							if len(postRuleNameList) > 0 {
								postRuleUid, err = Inst().Backup.GetRuleUid(BackupOrgID, ctx, postRuleNameList[0])
								log.FailOnError(err, "Fetching post backup rule [%s] uid", postRuleNameList[0])
								postRule = postRuleNameList[0]
							}
							err = CreateBackup(backupName, SourceClusterName, backupLocationName, backupLocationUID, bkpNamespaces, labelSelectors, BackupOrgID, sourceClusterUid, preRule, preRuleUid, postRule, postRuleUid, ctx)
							dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s] on namespace [%s]", backupName, namespace))
							err := BackupSuccessCheck(backupName, BackupOrgID, MaxWaitPeriodForBackupCompletionInMinutes*time.Minute, 30*time.Second, ctx)
							dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying backup %s success state", backupName))
							mutex.Lock()
							backupNames2 = append(backupNames2, backupName)
							backupNSMap[backupName] = namespace
							mutex.Unlock()
						}(namespace)
					}
				}
				wg.Wait()
			})

			Step(fmt.Sprintf("Default restore of applications by replacing the existing resources with NS level PSA and Cluster level PSA Set to %s", psaLevel), func() {
				log.InfoD(fmt.Sprintf("Default restore of applications by replacing the existing resources with NS level PSA and Cluster level PSA Set to %s", psaLevel))
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-central-admin ctx")
				var wg sync.WaitGroup
				var mutex sync.Mutex
				for _, backupName := range backupNames {
					wg.Add(1)
					go func(backupName string) {
						defer wg.Done()
						defer GinkgoRecover()
						defaultRestoreName := fmt.Sprintf("%s-%s-%s-default-2", RestoreNamePrefix, backupName, psaLevel)
						err = CreateRestoreWithReplacePolicyWithValidation(defaultRestoreName, backupName, make(map[string]string), SourceClusterName, BackupOrgID, ctx, make(map[string]string), ReplacePolicyDelete, scheduledAppContexts)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Creating default restore for manual backup with replace policy [%s]", defaultRestoreName))
						mutex.Lock()
						restoreNames = append(restoreNames, defaultRestoreName)
						mutex.Unlock()
					}(backupName)
				}
				wg.Wait()
			})

			Step(fmt.Sprintf("Restore of applications with NS and StorageClass mapping with NS level PSA with cluster level Set to %s", psaLevel), func() {
				log.InfoD(fmt.Sprintf("Restore of applications with NS and StorageClass mapping with NS level PSA with cluster level Set to %s", psaLevel))
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-central-admin ctx")

				for backupName, backupNamespace := range backupNSMap {
					// If cluster level is restricted and namespace level is baseline or privileged skip the restore as the apps won't come up
					if psaLevel == "restricted" && (strings.Contains(backupNamespace, "baseline") || strings.Contains(backupNamespace, "privileged")) {
						continue
					}
					appContextsToRestore := make([]*scheduler.Context, 0)
					namespaceMapping := make(map[string]string)
					customRestoreName := fmt.Sprintf("%s-%s-%s-custom-ns-sc-2", RestoreNamePrefix, backupName, psaLevel)
					namespaceMapping[backupNamespace] = backupNamespace + "-restored-2"
					restoredNamespaces = append(restoredNamespaces, backupNamespace+"-restored-2")
					appContextsToRestore = FilterAppContextsByNamespace(scheduledAppContexts, []string{backupNamespace})

					err = CreateRestoreWithReplacePolicyWithValidation(customRestoreName, backupName, namespaceMapping, DestinationClusterName, BackupOrgID, ctx, storageClassMapping, ReplacePolicyDelete, appContextsToRestore)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Create restore with NS and StorageClass mapping with NS level PSA [%s]", customRestoreName))

					restoreNames = append(restoreNames, customRestoreName)
				}
			})

			Step(fmt.Sprintf("Restore of applications with NS and StorageClass mapping with NS level PSA with cluster level Set to %s with pre-exisitng namespace", psaLevel), func() {
				log.InfoD(fmt.Sprintf("Restore of applications with NS and StorageClass mapping with NS level PSA with cluster level Set to %s with pre-exisitng namespace", psaLevel))
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-central-admin ctx")
				for backupName, backupNamespace := range backupNSMap {
					appContextsToRestore := make([]*scheduler.Context, 0)
					namespaceMapping := make(map[string]string)
					customRestoreName := fmt.Sprintf("%s-%s-%s-custom-ns-sc-pre-existing", RestoreNamePrefix, backupName, psaLevel)
					namespaceMapping[backupNamespace] = backupNamespace + "-restored-with-labels"
					restoredNamespaces = append(restoredNamespaces, backupNamespace+"-restored-with-labels")
					label["pod-security.kubernetes.io/enforce"] = strings.Split(backupNamespace, "-")[0]
					err = CreateNamespaceAndAssignLabels(backupNamespace+"-restored-with-labels", label)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creating namespace and assigning labels"))
					appContextsToRestore = FilterAppContextsByNamespace(scheduledAppContexts, []string{backupNamespace})
					err = CreateRestoreWithReplacePolicyWithValidation(customRestoreName, backupName, namespaceMapping, DestinationClusterName, BackupOrgID, ctx, storageClassMapping, ReplacePolicyDelete, appContextsToRestore)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Create restore with NS and StorageClass mapping with NS level PSA [%s]", customRestoreName))
					restoreNames = append(restoreNames, customRestoreName)
				}
			})

			Step("Revert Cluster Level PSA settings", func() {
				err := SetDestinationKubeConfig()
				dash.VerifyFatal(err, nil, "Setting destination kubeconfig")
				err = RevertClusterLevelPSA()
				dash.VerifyFatal(err, nil, "Revert cluster level PSA configuration")

				err = SetSourceKubeConfig()
				dash.VerifyFatal(err, nil, "Setting source kubeconfig")
				err = RevertClusterLevelPSA()
				dash.VerifyFatal(err, nil, "Revert cluster level PSA configuration")
				psaFlag = false
			})

			Step("Taking backup of all the namespaces created with namespace level PSA and no cluster level settings in a single backup", func() {
				log.InfoD("Taking backup of all the namespaces created with namespace level PSA and no cluster level settings in a single backup")
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-central-admin ctx")
				labelSelectors := make(map[string]string)

				for backupLocationUID, backupLocationName := range backupLocationMap {
					backupName := fmt.Sprintf("%s-%v-%s", BackupNamePrefix, RandomString(10), psaLevel)
					err = CreateBackup(backupName, SourceClusterName, backupLocationName, backupLocationUID, bkpNamespaces, labelSelectors, BackupOrgID, sourceClusterUid, "", "", "", "", ctx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of backup [%s]", backupName))
					err := BackupSuccessCheck(backupName, BackupOrgID, MaxWaitPeriodForBackupCompletionInMinutes*time.Minute, 30*time.Second, ctx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying backup %s success state", backupName))
					backupNamesAllNs = append(backupNamesAllNs, backupName)
				}
			})

			Step("Default restore of applications by replacing the existing resources with NS level PSA and no cluster level setting", func() {
				log.InfoD("Default restore of applications by replacing the existing resources with NS level PSA and no cluster level setting")
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-central-admin ctx")

				for _, backupName := range backupNamesAllNs {
					defaultRestoreName := fmt.Sprintf("%s-%s-%s-%s", RestoreNamePrefix, backupName, psaLevel, RandomString(10))
					err = CreateRestoreWithReplacePolicyWithValidation(defaultRestoreName, backupName, make(map[string]string), SourceClusterName, BackupOrgID, ctx, make(map[string]string), ReplacePolicyDelete, scheduledAppContexts)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creating default restore for manual backup with replace policy [%s]", defaultRestoreName))
					restoreNames = append(restoreNames, defaultRestoreName)
				}
			})
		}
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)

		defer func() {
			err := SetSourceKubeConfig()
			dash.VerifyFatal(err, nil, "Setting source kubeconfig")
		}()

		// Make sure to revert the cluster level PSA settings
		defer func() {
			if psaFlag {
				err := SetDestinationKubeConfig()
				dash.VerifyFatal(err, nil, "Setting destination kubeconfig")
				err = RevertClusterLevelPSA()
				dash.VerifyFatal(err, nil, "Revert cluster level PSA configuration")

				err = SetSourceKubeConfig()
				dash.VerifyFatal(err, nil, "Setting source kubeconfig")
				err = RevertClusterLevelPSA()
				dash.VerifyFatal(err, nil, "Revert cluster level PSA configuration")
			}

			log.InfoD("Setting the original app list back post testcase")
			Inst().AppList = originalList
		}()
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		if len(preRuleNameList) > 0 {
			for _, ruleName := range preRuleNameList {
				err := Inst().Backup.DeleteRuleForBackup(BackupOrgID, ruleName)
				dash.VerifySafely(err, nil, fmt.Sprintf("Deleting backup pre rules [%s]", ruleName))
			}
		}
		if len(postRuleNameList) > 0 {
			for _, ruleName := range postRuleNameList {
				err := Inst().Backup.DeleteRuleForBackup(BackupOrgID, ruleName)
				dash.VerifySafely(err, nil, fmt.Sprintf("Deleting backup post rules [%s]", ruleName))
			}
		}
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true

		log.Info("Destroying scheduled apps on source cluster")
		err = DestroyAppsWithData(scheduledAppContexts, opts, controlChannel, errorGroup)
		log.FailOnError(err, "Data validations failed")

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

		err = DeleteNamespaces(restoredNamespaces)
		log.FailOnError(err, "failed to delete restored namespaces")

		log.InfoD("switching to default context")
		err = SetClusterContext("")
		log.FailOnError(err, "failed to SetClusterContext to default cluster")

		backupDriver := Inst().Backup
		log.Info("Deleting backed up namespaces")
		for _, backupName := range backupNames {
			backupUID, err := backupDriver.GetBackupUID(ctx, backupName, BackupOrgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupName)
			backupDeleteResponse, err := DeleteBackup(backupName, backupUID, BackupOrgID, ctx)
			log.FailOnError(err, "Backup [%s] could not be deleted", backupName)
			dash.VerifyFatal(backupDeleteResponse.String(), "", fmt.Sprintf("Verifying [%s] backup deletion is successful", backupName))
		}

		log.Info("Deleting restored namespaces")
		for _, restoreName := range restoreNames {
			err = DeleteRestore(restoreName, BackupOrgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting Restore [%s]", restoreName))
		}

		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)

	})
})

// RestoreFromHigherPrivilegedNamespaceToLower verifies restore of applications from higher privileged to lower privileged at namespace level PSA
var _ = Describe("{RestoreFromHigherPrivilegedNamespaceToLower}", Label(TestCaseLabelsMap[RestoreFromHigherPrivilegedNamespaceToLower]...), func() {
	var (
		backupNames                    []string
		scheduledAppContexts           []*scheduler.Context
		label                          map[string]string
		providers                      []string
		restrictedNamespaceList        []string
		baselineNamespaceList          []string
		cloudCredName                  string
		cloudCredUID                   string
		backupLocationUID              string
		backupLocationMap              map[string]string
		sourceClusterUid               string
		controlChannel                 chan string
		errorGroup                     *errgroup.Group
		restrictedScheduledAppContexts []*scheduler.Context
		baselineScheduledAppContexts   []*scheduler.Context
		preRuleNameMultiApplication    string
		postRuleNameMultiApplication   string
		preRuleUidMultiApplication     string
		postRuleUidMultiApplication    string
		originalAppList                []string
	)
	AppContextsMapping := make(map[string]*scheduler.Context)
	providers = GetBackupProviders()
	label = make(map[string]string)
	backupLocationMap = make(map[string]string)
	scheduledAppContexts = make([]*scheduler.Context, 0)
	appPrivilegeToBkpMap := make(map[string]string)
	appPrivilegeToRestoreMap := make(map[string][]string)
	appPrivilegeToNsMap := make(map[string]string)
	restrictedScheduledAppContexts = make([]*scheduler.Context, 0)
	baselineScheduledAppContexts = make([]*scheduler.Context, 0)
	mulAppScheduledAppContexts := make([]*scheduler.Context, 0)
	mulAppRestrictedNamespaceList := make([]string, 0)

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("RestoreFromHigherPrivilegedNamespaceToLower", "Restore from higher Privileged to lower Privileged namespace", nil, 299239, Sn, Q2FY25)

		pipelineAppList := Inst().AppList
		Inst().AppList = []string{"postgres-backup", "mysql-backup"}
		originalAppList := Inst().AppList
		//Resetting the pipeline app list
		defer func() {
			Inst().AppList = pipelineAppList
		}()

		log.InfoD("Deploy applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		psaApp := make([]string, 0)

		// Deploy multiple applications on multiple namespace on restricted and baseline namespaces
		for _, psaLevel := range []string{"restricted", "baseline"} {

			if strings.Contains(psaLevel, "restricted") {
				log.InfoD("The app list at the start of the testcase is %v", originalAppList)
				for _, app := range originalAppList {
					psaApp = append(psaApp, PSAAppMap[app])
				}
				Inst().AppList = psaApp
				log.Infof("The PSA app list for restricted namespace is %v", psaApp)
			}
			label["pod-security.kubernetes.io/enforce"] = psaLevel
			err := Inst().S.RescanSpecs(Inst().SpecDir, Inst().V.String())
			log.FailOnError(err, "Failed to rescan specs from %s for storage provider %s", Inst().SpecDir, Inst().V.String())

			for i := 0; i < len(originalAppList); i++ {
				taskName := fmt.Sprintf("%s-%d-%s", psaLevel, i, RandomString(10))
				namespace := fmt.Sprintf("%s-%d-%s", psaLevel, i, RandomString(10))
				appPrivilegeToNsMap[psaLevel] = namespace
				err = CreateNamespaceAndAssignLabels(namespace, label)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating namespace [%s] and assigning label [%s]", namespace, label))
				appContexts := ScheduleApplicationsOnNamespace(namespace, taskName)
				for _, ctx := range appContexts {
					ctx.ReadinessTimeout = AppReadinessTimeout
					namespace := GetAppNamespace(ctx, taskName)
					scheduledAppContexts = append(scheduledAppContexts, ctx)
					AppContextsMapping[namespace] = ctx
					if strings.Contains(namespace, "restricted") {
						restrictedScheduledAppContexts = append(restrictedScheduledAppContexts, ctx)
						restrictedNamespaceList = append(restrictedNamespaceList, namespace)
					} else if strings.Contains(namespace, "baseline") {
						baselineScheduledAppContexts = append(baselineScheduledAppContexts, ctx)
						baselineNamespaceList = append(baselineNamespaceList, namespace)
					}
				}
			}

			baselineNamespaceList = GetUniqueElementsFromList(baselineNamespaceList)
			restrictedNamespaceList = GetUniqueElementsFromList(restrictedNamespaceList)

			if strings.Contains(psaLevel, "restricted") {
				Inst().AppList = originalAppList
			}
		}

		// Iterate over each application list to deploy single application on multiple namespace
		for i, app := range originalAppList {
			Inst().AppList = []string{app}
			// Generate a unique namespace name
			namespace := fmt.Sprintf("%s-%d-%s", app, i, RandomString(10))

			// Create namespace and assign labels
			err := CreateNamespaceAndAssignLabels(namespace, BaselinePSALabel)
			dash.VerifyFatal(err, nil, "Creating namespace and assigning labels")

			// Schedule application on namespace
			taskName := fmt.Sprintf("%s-%d-%s", app, i, RandomString(10))
			appContexts := ScheduleApplicationsOnNamespace(namespace, taskName)

			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = AppReadinessTimeout

				// Append context to the list of scheduled app contexts
				mulAppScheduledAppContexts = append(mulAppScheduledAppContexts, ctx)

				// Store namespace in restrictedNamespaceList
				mulAppRestrictedNamespaceList = append(mulAppRestrictedNamespaceList, namespace)
			}
		}
	})
	It("Restore from higher Privileged to lower Privileged namespace", func() {

		Step("Validating applications", func() {
			log.InfoD("Validating applications")
			ctx, _ := backup.GetAdminCtxFromSecret()
			controlChannel, errorGroup = ValidateApplicationsStartData(scheduledAppContexts, ctx)
		})

		Step(fmt.Sprintf("Create pre and post exec rules for multiple applications"), func() {
			log.InfoD("Create pre and post exec rules for multiple applications")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-admin ctx")
			preRuleNameMultiApplication, postRuleNameMultiApplication, err = CreateRuleForBackupWithMultipleApplications(BackupOrgID, originalAppList, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of pre and post exec rules for applications for multiple applications"))
			if preRuleNameMultiApplication != "" {
				preRuleUidMultiApplication, err = Inst().Backup.GetRuleUid(BackupOrgID, ctx, preRuleNameMultiApplication)
				log.FailOnError(err, "Fetching pre backup rule [%s] uid", postRuleNameMultiApplication)
				log.Infof("Pre backup rule [%s] uid: [%s]", postRuleNameMultiApplication, preRuleUidMultiApplication)
			}
			if postRuleNameMultiApplication != "" {
				postRuleUidMultiApplication, err = Inst().Backup.GetRuleUid(BackupOrgID, ctx, postRuleNameMultiApplication)
				log.FailOnError(err, "Fetching post backup rule [%s] uid", postRuleNameMultiApplication)
				log.Infof("Post backup rule [%s] uid: [%s]", postRuleNameMultiApplication, postRuleUidMultiApplication)
			}
		})

		Step("Creating backup location and cloud setting", func() {
			log.InfoD("Creating backup location and cloud setting")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, RandomString(10))
				backupLocationName := fmt.Sprintf("%s-%s-bl-%v", provider, getGlobalBucketName(provider), time.Now().Unix())
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocationName
				err := CreateCloudCredential(provider, cloudCredName, cloudCredUID, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s] with [%s] as provider", cloudCredName, BackupOrgID, provider))
				err = CreateBackupLocation(provider, backupLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, "Creating backup location")
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
		})

		Step("Taking backup of multiple namespaces which is associated with baseline level PSA", func() {
			log.InfoD("Taking backup of multiple namespaces which is associated with baseline level PSA")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			for backupLocationUID, backupLocationName := range backupLocationMap {
				backupName := fmt.Sprintf("%s-%s", BackupNamePrefix, RandomString(10))
				err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, backupLocationName, backupLocationUID, mulAppScheduledAppContexts, make(map[string]string), BackupOrgID, sourceClusterUid, preRuleNameMultiApplication, preRuleUidMultiApplication, postRuleNameMultiApplication, postRuleUidMultiApplication)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				appPrivilegeToBkpMap["baseline-mul-ns-single-app"] = backupName
				backupNames = append(backupNames, backupName)
			}
		})

		Step("Taking backup of namespace which is associated with restricted level PSA", func() {
			log.InfoD("Taking backup of namespace which is associated with restricted level PSA")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			for backupLocationUID, backupLocationName := range backupLocationMap {
				backupName := fmt.Sprintf("%s-%s", BackupNamePrefix, RandomString(10))
				err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, backupLocationName, backupLocationUID, restrictedScheduledAppContexts, make(map[string]string), BackupOrgID, sourceClusterUid, preRuleNameMultiApplication, preRuleUidMultiApplication, postRuleNameMultiApplication, postRuleUidMultiApplication)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				appPrivilegeToBkpMap["restricted"] = backupName
				backupNames = append(backupNames, backupName)
			}
		})

		Step("Taking backup of namespace which is associated with baseline level PSA", func() {
			log.InfoD("Taking backup of namespace which is associated with baseline level PSA")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			for backupLocationUID, backupLocationName := range backupLocationMap {
				backupName := fmt.Sprintf("%s-%s", BackupNamePrefix, RandomString(10))
				err = CreateBackupWithValidation(ctx, backupName, SourceClusterName, backupLocationName, backupLocationUID, baselineScheduledAppContexts, make(map[string]string), BackupOrgID, sourceClusterUid, preRuleNameMultiApplication, preRuleUidMultiApplication, postRuleNameMultiApplication, postRuleUidMultiApplication)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation and Validation of backup [%s]", backupName))
				appPrivilegeToBkpMap["baseline"] = backupName
				backupNames = append(backupNames, backupName)
			}
		})

		Step("Create namespace with different privileges for performing restores on destination cluster", func() {
			err := SetDestinationKubeConfig()
			log.FailOnError(err, "Switching context to destination cluster failed")
			for _, psalevel := range []string{"restricted", "baseline", "privileged"} {
				psaNameSpaceList := make([]string, 0)

				for i := 0; i < len(originalAppList); i++ {
					namespace := fmt.Sprintf("%s-%s", psalevel, RandomString(10))
					label["pod-security.kubernetes.io/enforce"] = psalevel
					err := CreateNamespaceAndAssignLabels(namespace, label)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creating namespace [%s] and assigning label [%s]", namespace, label))
					psaNameSpaceList = append(psaNameSpaceList, namespace)
				}

				// Assign psaNameSpaceList to appPrivilegeToRestoreMap[psalevel]
				appPrivilegeToRestoreMap[psalevel] = psaNameSpaceList
			}

			// Switch context back to source cluster
			err = SetSourceKubeConfig()
			log.FailOnError(err, "Switching context to source cluster failed")
		})

		Step("Perform a custom restore of the backup taken from the namespace in baseline mode to a namespace in restricted mode on a different cluster.", func() {
			log.InfoD("Perform a custom restore of the backup taken from the namespace in baseline mode to a namespace in restricted mode on a different cluster.")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Unable to fetch px-central-admin ctx")

			restoreName := fmt.Sprintf("%s-%s", "test-restore", RandomString(10))
			namespaceMapping := make(map[string]string)

			// Populate namespaceMapping with mappings for baseline to restricted namespace
			for i := range originalAppList {
				namespaceMapping[baselineNamespaceList[i]] = appPrivilegeToRestoreMap["restricted"][i]
			}

			// Define other parameters as needed for CreateRestoreWithValidation
			err = CreateRestoreWithValidation(ctx, restoreName, appPrivilegeToBkpMap["baseline"], namespaceMapping, make(map[string]string), DestinationClusterName, BackupOrgID, baselineScheduledAppContexts)
			dash.VerifyFatal(strings.Contains(err.Error(), "failed to meet the pod security standard"), true, fmt.Sprintf("Creating restore [%s] from backup [%s] taken on baseline namespace and restore to restricted namespace failed as expected", restoreName, appPrivilegeToBkpMap["baseline"]))
		})

		Step("Remove restricted label from the namespace and add baseline label", func() {
			err := SetDestinationKubeConfig()
			log.FailOnError(err, "Switching context to destination cluster failed")
			log.InfoD("Remove restricted label from the namespace [%s]", appPrivilegeToRestoreMap["restricted"])
			for i := range originalAppList {
				err = Inst().S.RemoveNamespaceLabel(appPrivilegeToRestoreMap["restricted"][i], RestrictedPSALabel)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Removing label [%v] from namespace [%v]", RestrictedPSALabel, appPrivilegeToRestoreMap["restricted"][i]))
				err = Inst().S.AddNamespaceLabel(appPrivilegeToRestoreMap["restricted"][i], BaselinePSALabel)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Adding label [%v] to namespace [%v]", BaselinePSALabel, appPrivilegeToRestoreMap["restricted"][i]))
			}
			// Switch context back to source cluster
			err = SetSourceKubeConfig()
			log.FailOnError(err, "Switching context to source cluster failed")
		})

		Step("Perform a custom restore of the backup taken from the namespace in baseline mode to a namespace which is replaced restricted with baseline mode on a different cluster with replace option", func() {
			log.InfoD("Perform a custom restore of the backup taken from the namespace in baseline mode to a namespace which is replaced restricted with baseline mode on a different cluster with replace option")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Unable to fetch px-central-admin ctx")
			restoreName := fmt.Sprintf("%s-%s", "test-restore", RandomString(10))
			namespaceMapping := make(map[string]string)

			// Populate namespaceMapping with mappings for baseline to restricted namespace
			for i := range originalAppList {
				namespaceMapping[baselineNamespaceList[i]] = appPrivilegeToRestoreMap["restricted"][i]
			}
			err = CreateRestoreWithReplacePolicyWithValidation(restoreName, appPrivilegeToBkpMap["baseline"], namespaceMapping, DestinationClusterName, BackupOrgID, ctx, make(map[string]string), 2, baselineScheduledAppContexts)
			log.Infof("error while validation of restore")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s] from backup [%s]", restoreName, appPrivilegeToBkpMap["baseline"]))
		})

		Step("Perform a custom restore of the backup taken from the namespace in baseline mode to a namespace with restricted mode with replace option on different cluster", func() {
			log.InfoD("Perform a custom restore of the backup taken from the namespace in baseline mode to a namespace with restricted mode with replace option on different cluster")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Unable to fetch px-central-admin ctx")
			restrictedNamespaceList := make([]string, 0)

			for i := 0; i < len(originalAppList); i++ {
				namespace := fmt.Sprintf("%s-%s", "restricted-ns-1", RandomString(10))
				err = CreateNamespaceAndAssignLabels(namespace, RestrictedPSALabel)
				dash.VerifyFatal(err, nil, "Creating namespace and assigning labels")
				restrictedNamespaceList = append(restrictedNamespaceList, namespace)
			}

			restoreName := fmt.Sprintf("%s-%s", "test-restore", RandomString(10))
			namespaceMapping := make(map[string]string)

			// Populate namespaceMapping with mappings for baseline to restricted namespace with replace option
			for i := range originalAppList {
				namespaceMapping[mulAppRestrictedNamespaceList[i]] = restrictedNamespaceList[i]
			}
			err = CreateRestoreWithReplacePolicyWithValidation(restoreName, appPrivilegeToBkpMap["baseline-mul-ns-single-app"], namespaceMapping, DestinationClusterName, BackupOrgID, ctx, make(map[string]string), 2, mulAppScheduledAppContexts)
			log.Infof("error while validation of restore")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s] from backup [%s]", restoreName, appPrivilegeToBkpMap["baseline-mul-ns-single-app"]))
		})

		Step("Perform a custom restore of the backup taken from the namespace in restricted mode to a namespace in baseline mode on a different cluster", func() {
			log.InfoD("Perform a custom restore of the backup taken from the namespace in restricted mode to a namespace in baseline mode on a different cluster")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Unable to fetch px-central-admin ctx")
			restoreName := fmt.Sprintf("%s-%s", "test-restore", RandomString(10))
			namespaceMapping := make(map[string]string)

			// Populate namespaceMapping with mappings for baseline to restricted namespace
			for i := range originalAppList {
				namespaceMapping[restrictedNamespaceList[i]] = appPrivilegeToRestoreMap["baseline"][i]
			}
			err = CreateRestoreWithValidation(ctx, restoreName, appPrivilegeToBkpMap["restricted"], namespaceMapping, make(map[string]string), DestinationClusterName, BackupOrgID, restrictedScheduledAppContexts)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s] from backup [%s]", restoreName, appPrivilegeToBkpMap["restricted"]))
		})

		Step("Perform a custom restore of the backup taken from the multiple application on multiple namespace in baseline mode to a namespace in privileged mode on a different cluster", func() {
			log.InfoD("Perform a custom restore of the backup taken from the multiple application on multiple namespace in baseline mode to a namespace in privileged mode on a different cluster")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Unable to fetch px-central-admin ctx")
			restoreName := fmt.Sprintf("%s-%s", "test-restore", RandomString(10))
			namespaceMapping := make(map[string]string)

			// Populate namespaceMapping with mappings for baseline to privileged namespace
			for i := range originalAppList {
				namespaceMapping[baselineNamespaceList[i]] = appPrivilegeToRestoreMap["privileged"][i]
			}
			err = CreateRestoreWithValidation(ctx, restoreName, appPrivilegeToBkpMap["baseline"], namespaceMapping, make(map[string]string), DestinationClusterName, BackupOrgID, baselineScheduledAppContexts)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s] from backup [%s]", restoreName, appPrivilegeToBkpMap["baseline"]))
		})
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		err := SetSourceKubeConfig()
		log.FailOnError(err, "Switching context to source cluster failed")

		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")

		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true

		log.Info("Destroying scheduled apps on source cluster")
		err = DestroyAppsWithData(scheduledAppContexts, opts, controlChannel, errorGroup)
		log.FailOnError(err, "Data validations failed")

		backupDriver := Inst().Backup
		log.Info("Deleting backups")
		for _, backupName := range backupNames {
			backupUID, err := backupDriver.GetBackupUID(ctx, backupName, BackupOrgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupName)
			backupDeleteResponse, err := DeleteBackup(backupName, backupUID, BackupOrgID, ctx)
			log.FailOnError(err, "Backup [%s] could not be deleted", backupName)
			dash.VerifyFatal(backupDeleteResponse.String(), "", fmt.Sprintf("Verifying [%s] backup deletion is successful", backupName))
			err = DeleteBackupAndWait(backupName, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("waiting for backup [%s] deletion", backupName))
		}

		log.Info("Deleting rules")
		if preRuleNameMultiApplication != "" {
			err = Inst().Backup.DeleteRuleForBackup(BackupOrgID, preRuleNameMultiApplication)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting pre exec rule %s ", preRuleNameMultiApplication))
		}
		if postRuleNameMultiApplication != "" {
			err = Inst().Backup.DeleteRuleForBackup(BackupOrgID, postRuleNameMultiApplication)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting post exec rule %s ", postRuleNameMultiApplication))
		}

		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})

// PSALowerPrivilegeToHigherPrivilegeWithProjectMapping verifies different backup and restore operations from lower privileges to higher privileges with project mapping
var _ = Describe("{PSALowerPrivilegeToHigherPrivilegeWithProjectMapping}", Label(TestCaseLabelsMap[PSALowerPrivilegeToHigherPrivilegeWithProjectMapping]...), func() {
	var (
		err                            error
		wg                             sync.WaitGroup
		mutex                          sync.Mutex
		backupLocationUID              string
		backupLocationName             string
		credName                       string
		credUid                        string
		numOfDeployments               int
		srcClusterUid                  string
		clusterStatus                  api.ClusterInfo_StatusInfo_Status
		preRuleName                    string
		postRuleName                   string
		preRuleUid                     string
		postRuleUid                    string
		backupName                     string
		periodicPolicyName             string
		schPolicyUid                   string
		actualAppList                  []string
		testAppList                    []string
		allNSScheduleBackup            string
		baselineScheduleBackup         string
		restrictedScheduleBackup       string
		backupList                     []string
		schBackupList                  []string
		restrictedManualBackupName     string
		baselineManualBackupName       string
		scheduleNames                  = make([]string, 0)
		scheduledAppContexts           []*scheduler.Context
		privilegeScheduledAppContexts  []*scheduler.Context
		baselineScheduledAppContexts   []*scheduler.Context
		restrictedScheduledAppContexts []*scheduler.Context
	)

	AppContextsMapping := make(map[string]*scheduler.Context)
	backupLocationMap := make(map[string]string)
	bkpNamespaces := make([]string, 0)
	nsLabel := make(map[string]string)
	projectLabel := make(map[string]string)
	projectAnnotation := make(map[string]string)
	ctx, err := backup.GetAdminCtxFromSecret()
	log.FailOnError(err, "Getting admin context from secret")
	scheduledAppContexts = make([]*scheduler.Context, 0)
	privilegeScheduledAppContexts = make([]*scheduler.Context, 0)
	baselineScheduledAppContexts = make([]*scheduler.Context, 0)
	restrictedScheduledAppContexts = make([]*scheduler.Context, 0)
	sourceNamespaceProjectMapping := make(map[string]string)
	sourceNamespaceProjectUIDMapping := make(map[string]string)
	namespaceMapRestrictedToBaseline := make(map[string]string)
	namespaceMapBaselineToPrivilege := make(map[string]string)
	namespaceMapRestrictedToPrivilege := make(map[string]string)
	allNamespaceMap := make(map[string]string)
	restoreProjectNameMapping := make(map[string]string)
	restoreProjectUIDMapping := make(map[string]string)
	periodicPolicyName = fmt.Sprintf("%s-%s", "periodic", RandomString(5))

	JustBeforeEach(func() {
		StartPxBackupTorpedoTest("PSALowerPrivilegeToHigherPrivilegeWithProjectMapping", "Verify PSA backup in lower privilege mode and restore in higher privilege mode with project mapping", nil, 299238, Vpinisetti, Q2FY25)
		numOfDeployments = Inst().GlobalScaleFactor
		actualAppList = Inst().AppList
		testAppList = []string{"postgres-backup", "mysql-backup"} // mysql-backup will be added.
		log.InfoD("Deploy applications")
		psaApp := make([]string, 0)
		for _, psaLevel := range []string{"restricted", "baseline"} {
			if psaLevel == "restricted" {
				for _, app := range testAppList {
					psaApp = append(psaApp, PSAAppMap[app])
				}
				Inst().AppList = psaApp
			}
			nsLabel["pod-security.kubernetes.io/enforce"] = psaLevel

			log.InfoD("Deploying all provided applications in a single namespace")
			for i := 0; i < numOfDeployments; i++ {
				taskName := fmt.Sprintf("%s-%d-%d-%v", TaskNamePrefix, 299238, i, RandomString(3))
				namespace := fmt.Sprintf("%s-ns-multiapp-%v", psaLevel, taskName)
				err := CreateNamespaceAndAssignLabels(namespace, nsLabel)
				dash.VerifyFatal(err, nil, "Creating namespace and assigning labels")
				appContexts := ScheduleApplicationsOnNamespace(namespace, taskName)
				for _, appCtx := range appContexts {
					appCtx.ReadinessTimeout = AppReadinessTimeout
					namespace := GetAppNamespace(appCtx, taskName)
					bkpNamespaces = append(bkpNamespaces, namespace)
					scheduledAppContexts = append(scheduledAppContexts, appCtx)
					AppContextsMapping[namespace] = appCtx
					if strings.Contains(namespace, "restricted") {
						restrictedScheduledAppContexts = append(restrictedScheduledAppContexts, appCtx)
					} else if strings.Contains(namespace, "baseline") {
						baselineScheduledAppContexts = append(baselineScheduledAppContexts, appCtx)
					} else {
						privilegeScheduledAppContexts = append(privilegeScheduledAppContexts, appCtx)
					}
				}
			}

			log.InfoD("Deploying all provided applications in individual namespaces")
			for _, app := range testAppList {
				if psaLevel == "restricted" {
					Inst().AppList = []string{PSAAppMap[app]}
					log.Infof("The restricted PSA app list is %v", Inst().AppList)
				}
				taskName := fmt.Sprintf("%s-%s-%v", TaskNamePrefix, app, RandomString(3))
				namespace := fmt.Sprintf("%s-ns-singleapp-%v", psaLevel, taskName)
				err := CreateNamespaceAndAssignLabels(namespace, nsLabel)
				dash.VerifyFatal(err, nil, "Creating namespace and assigning labels")
				appContexts := ScheduleApplicationsOnNamespace(namespace, taskName)
				for _, appCtx := range appContexts {
					appCtx.ReadinessTimeout = AppReadinessTimeout
					namespace := GetAppNamespace(appCtx, taskName)
					bkpNamespaces = append(bkpNamespaces, namespace)
					scheduledAppContexts = append(scheduledAppContexts, appCtx)
					AppContextsMapping[namespace] = appCtx
					if strings.Contains(namespace, "restricted") {
						restrictedScheduledAppContexts = append(restrictedScheduledAppContexts, appCtx)
					} else if strings.Contains(namespace, "baseline") {
						baselineScheduledAppContexts = append(baselineScheduledAppContexts, appCtx)
					} else {
						privilegeScheduledAppContexts = append(privilegeScheduledAppContexts, appCtx)
					}
				}
			}
			if psaLevel == "restricted" {
				log.Infof("The app list at the end of the iteration %s is %v", psaLevel, Inst().AppList)
				Inst().AppList = testAppList
			}
		}
		projectLabel[RandomString(9)] = RandomString(9)
		projectAnnotation[RandomString(9)] = RandomString(9)
	})

	It("Take backup with restricted PSA namespace and restore it in privilege PSA namespace", func() {
		defer func() {
			log.InfoD("Switching to default context")
			err := SetClusterContext("")
			log.FailOnError(err, "Failed to set ClusterContext to default cluster")
		}()

		Step("Validate applications", func() {
			log.InfoD("Validating applications")
			ValidateApplications(scheduledAppContexts)
		})

		Step("Create cloud credentials and backup location", func() {
			log.InfoD("Creating cloud credentials and backup location")
			backupLocationProviders := GetBackupProviders()
			for _, provider := range backupLocationProviders {
				credName = fmt.Sprintf("%s-cred-%v", provider, RandomString(10))
				credUid = uuid.New()
				err := CreateCloudCredential(provider, credName, credUid, BackupOrgID, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of cloud credential named [%s] for org [%s]  as provider %s", credName, BackupOrgID, provider))
				backupLocationName = fmt.Sprintf("%s-backup-location-%v", provider, RandomString(10))
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocationName
				err = CreateBackupLocation(provider, backupLocationName, backupLocationUID, credName, credUid, getGlobalBucketName(provider), BackupOrgID, "", true)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocationName))
			}
		})

		Step("Registering cluster for backup", func() {
			log.InfoD("Registering cluster for backup")
			err = CreateApplicationClusters(BackupOrgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying if [%s] cluster is online", SourceClusterName))
			srcClusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid %s", SourceClusterName, srcClusterUid))
		})

		Step("Create schedule policies", func() {
			log.InfoD("Creating schedule policies")
			periodicSchedulePolicyInfo := Inst().Backup.CreateIntervalSchedulePolicy(5, 40, 2)
			periodicPolicyStatus := Inst().Backup.BackupSchedulePolicy(periodicPolicyName, uuid.New(), BackupOrgID, periodicSchedulePolicyInfo)
			dash.VerifyFatal(periodicPolicyStatus, nil, fmt.Sprintf("Creation of periodic schedule policy - %s", periodicPolicyName))
		})

		Step("Creation of pre and post exec rules for given applications", func() {
			log.InfoD("Creation of pre and post exec rules for given applications ")
			preRuleName, postRuleName, err = CreateRuleForBackupWithMultipleApplications(BackupOrgID, Inst().AppList, ctx)
			dash.VerifyFatal(err, nil, "Verifying creation of pre and post exec rules for given applications from px-admin")
			if preRuleName != "" {
				preRuleUid, err = Inst().Backup.GetRuleUid(BackupOrgID, ctx, preRuleName)
				log.FailOnError(err, "Fetching pre backup rule [%s] uid", preRuleName)
				log.InfoD("Pre backup rule [%s] uid: [%s]", preRuleName, preRuleUid)
			}
			if postRuleName != "" {
				postRuleUid, err = Inst().Backup.GetRuleUid(BackupOrgID, ctx, postRuleName)
				log.FailOnError(err, "Fetching post backup rule [%s] uid", postRuleName)
				log.InfoD("Post backup rule [%s] uid: [%s]", postRuleName, postRuleUid)
			}
		})

		Step("Creating namespaces and adding them to rancher projects on source cluster", func() {
			log.InfoD("Creating namespaces and adding them to rancher projects on source cluster")
			for i := 0; i < len(bkpNamespaces); i++ {
				projectName := fmt.Sprintf("project-%v-%d", RandomString(5), i)
				_, err = Inst().S.(*rke.Rancher).CreateRancherProject(projectName, RancherProjectDescription, "vpinisetti-62", projectLabel, projectAnnotation)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating rancher project %s", projectName))
				projectID, err := Inst().S.(*rke.Rancher).GetProjectID(projectName)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Getting Project ID - %s for project %s", projectID, projectName))
				log.Infof("Adding namespace to source project and taking backup of it")
				err = Inst().S.(*rke.Rancher).AddNamespacesToProject(projectName, []string{bkpNamespaces[i]})
				dash.VerifyFatal(err, nil, fmt.Sprintf("Adding namespace %s to project %s", bkpNamespaces[i], projectName))
				err = Inst().S.(*rke.Rancher).ValidateProjectOfNamespaces(projectName, []string{bkpNamespaces[i]})
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying project %s of namespace %s", projectName, bkpNamespaces[i]))
				sourceNamespaceProjectMapping[bkpNamespaces[i]] = projectName
				log.Infof("The source namespace project mapping is %v", sourceNamespaceProjectMapping)
				sourceNamespaceProjectUIDMapping[bkpNamespaces[i]] = projectID
				log.Infof("The source namespace project UID mapping is %v", sourceNamespaceProjectUIDMapping)
			}
		})

		wg.Add(1)
		go func() {
			defer GinkgoRecover()
			defer wg.Done()
			Step("Creating manual backup with restricted namespaces", func() {
				log.InfoD("Creating manual backup with restricted namespaces")
				restrictedManualBackupName = fmt.Sprintf("%s-%v", "bkp-restricted", RandomString(5))
				err = CreateBackupWithValidation(ctx, restrictedManualBackupName, SourceClusterName, backupLocationName, backupLocationUID, restrictedScheduledAppContexts, nil, BackupOrgID, srcClusterUid, preRuleName, preRuleUid, postRuleName, postRuleUid)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of manual backup %v with restricted namespaces", restrictedManualBackupName))
				mutex.Lock()
				backupList = append(backupList, restrictedManualBackupName)
				mutex.Unlock()
				log.InfoD("Backup list after manual restricted backup %v", backupList)
			})
		}()

		wg.Add(1)
		go func() {
			defer GinkgoRecover()
			defer wg.Done()
			Step("Creating manual backup with baseline namespaces", func() {
				log.InfoD("Creating manual backup with baseline namespaces")
				baselineManualBackupName = fmt.Sprintf("%s-%v", "bkp-baseline", RandomString(5))
				err = CreateBackupWithValidation(ctx, baselineManualBackupName, SourceClusterName, backupLocationName, backupLocationUID, baselineScheduledAppContexts, nil, BackupOrgID, srcClusterUid, preRuleName, preRuleUid, postRuleName, postRuleUid)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of manual backup %v with basline namespaces", backupName))
				mutex.Lock()
				backupList = append(backupList, baselineManualBackupName)
				mutex.Unlock()
				log.InfoD("Backup list after manual baseline backup %v", backupList)
			})
		}()
		wg.Wait()

		Step("Creating namespaces and rancher projects on destination cluster", func() {
			log.InfoD("Creating namespaces and rancher projects on destination cluster")
			err = SetDestinationKubeConfig()
			log.FailOnError(err, "Switching context to destination cluster failed")
			var restoredNamespaceList []string
			for i := 0; i < len(bkpNamespaces); i++ {
				log.InfoD("Actual namespaces to be created on destination cluster is %v", bkpNamespaces)
				namespace := fmt.Sprintf("restore-%v-%v", bkpNamespaces[i], RandomString(3))
				if strings.Contains(namespace, "restricted") {
					namespaceMapRestrictedToBaseline[bkpNamespaces[i]] = namespace
					err1 := CreateNamespaceAndAssignLabels(namespace, BaselinePSALabel)
					dash.VerifyFatal(err1, nil, fmt.Sprintf("Created namespace %s and assigned PSA label %s", namespace, BaselinePSALabel))
					log.Infof("Restricted to baseline PSA map is %v", namespaceMapRestrictedToBaseline)
					namespaceMapRestrictedToPrivilege[bkpNamespaces[i]] = namespace
					err2 := CreateNamespaceAndAssignLabels(namespace, PrivilegedPSALabel)
					dash.VerifyFatal(err2, nil, fmt.Sprintf("Created namespace %s and assigned PSA label %s", namespace, PrivilegedPSALabel))
					log.Infof("Restricted to privilege PSA map is %v", namespaceMapRestrictedToPrivilege)
				} else if strings.Contains(namespace, "baseline") {
					namespaceMapBaselineToPrivilege[bkpNamespaces[i]] = namespace
					err := CreateNamespaceAndAssignLabels(namespace, PrivilegedPSALabel)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Creating namespace %s", namespace))
					log.Infof("Baseline to Privilege PSA map %v", namespaceMapBaselineToPrivilege)
				}
				allNamespaceMap = make(map[string]string)
				for key, value := range namespaceMapRestrictedToBaseline {
					allNamespaceMap[key] = value
				}
				for key, value := range namespaceMapBaselineToPrivilege {
					allNamespaceMap[key] = value
				}
				restoredNamespaceList = append(restoredNamespaceList, namespace)
				log.InfoD("Created namespace list on destination cluster %v", restoredNamespaceList)
				projectName := fmt.Sprintf("project-%v-%d", RandomString(5), i)
				_, err = Inst().S.(*rke.Rancher).CreateRancherProject(projectName, RancherProjectDescription, "vpinisetti-63", projectLabel, projectAnnotation)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating rancher project %s", projectName))
				projectID, err := Inst().S.(*rke.Rancher).GetProjectID(projectName)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Getting Project ID - %s for project %s", projectID, projectName))
				restoreProjectUIDMapping[sourceNamespaceProjectUIDMapping[bkpNamespaces[i]]] = projectID
				log.Infof("Project UID mapping to restore : %v", restoreProjectUIDMapping)
				restoreProjectNameMapping[sourceNamespaceProjectMapping[bkpNamespaces[i]]] = projectName
				log.Infof("Project name mapping to restore : %v", restoreProjectNameMapping)
				err = Inst().S.(*rke.Rancher).AddNamespacesToProject(projectName, []string{namespace})
				dash.VerifyFatal(err, nil, fmt.Sprintf("Adding namespace %s to project %s", namespace, projectName))
				err = Inst().S.(*rke.Rancher).ValidateProjectOfNamespaces(projectName, []string{namespace})
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying project %s of namespace %s", projectName, namespace))
			}
		})

		Step("Restore restricted to baseline namespaces with namespace & project mappings", func() {
			log.InfoD("Restore restricted to baseline namespaces with namespace & project mappings")
			err = SetSourceKubeConfig()
			log.FailOnError(err, "Switching context to source cluster failed")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			log.Infof("The restricted backup which is going to restore is %v", restrictedManualBackupName)
			restoreName := fmt.Sprintf("restore-rtob-%v", RandomString(5))
			err = CreateRestoreWithProjectMapping(restoreName, restrictedManualBackupName, namespaceMapRestrictedToBaseline, DestinationClusterName, BackupOrgID, ctx, nil, restoreProjectUIDMapping, restoreProjectNameMapping)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of restore with %s from backup %s", restoreName, restrictedManualBackupName))
		})

		Step("Restore baseline to privilege namespaces with namespace & project mappings", func() {
			log.InfoD("Restore baseline to privilege namespaces with namespace & project mappings")
			log.Infof("The baseline backup which is going to restore is %v", baselineManualBackupName)
			restoreName := fmt.Sprintf("restore-btop-%v", RandomString(5))
			err = CreateRestoreWithProjectMapping(restoreName, baselineManualBackupName, namespaceMapBaselineToPrivilege, DestinationClusterName, BackupOrgID, ctx, nil, restoreProjectUIDMapping, restoreProjectNameMapping)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of restore with %s from backup %s", restoreName, baselineManualBackupName))
		})

		Step("Creating schedule backup with all restricted & baseline namespaces", func() {
			log.InfoD("Creating schedule backup with all restricted & baseline namespaces")
			schPolicyUid, _ = Inst().Backup.GetSchedulePolicyUid(BackupOrgID, ctx, periodicPolicyName)
			schBackupName := fmt.Sprintf("schbkp-all-ns-%v", RandomString(4))
			scheduleNames = append(scheduleNames, schBackupName)
			labelSelectors := make(map[string]string)
			log.InfoD("Creating a schedule backup [%s] with namespaces [%s]", backupName, bkpNamespaces)
			allNSScheduleBackup, err = CreateScheduleBackupWithValidation(ctx, schBackupName, SourceClusterName, backupLocationName, backupLocationUID, scheduledAppContexts, labelSelectors, BackupOrgID, preRuleName, preRuleUid, postRuleName, postRuleUid, periodicPolicyName, schPolicyUid)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of schedule backup [%s] with all namespaces contains both restricted and baseline apps.", allNSScheduleBackup))
			schBackupList = append(schBackupList, allNSScheduleBackup)
			log.Infof("Schedule backup list after all namespaces backup %v", schBackupList)
		})

		wg.Add(1)
		go func() {
			defer GinkgoRecover()
			defer wg.Done()
			Step("Creating schedule backup with restricted namespaces", func() {
				log.InfoD("Creating schedule backup with restricted namespaces")
				schPolicyUid, _ = Inst().Backup.GetSchedulePolicyUid(BackupOrgID, ctx, periodicPolicyName)
				schBackupName := fmt.Sprintf("schbkp-restricted-%v", RandomString(5))
				labelSelectors := make(map[string]string)
				restrictedScheduleBackup, err = CreateScheduleBackupWithValidation(ctx, schBackupName, SourceClusterName, backupLocationName, backupLocationUID, restrictedScheduledAppContexts, labelSelectors, BackupOrgID, preRuleName, preRuleUid, postRuleName, postRuleUid, periodicPolicyName, schPolicyUid)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of schedule backup %v with restricted namespaces", restrictedScheduleBackup))
				mutex.Lock()
				scheduleNames = append(scheduleNames, schBackupName)
				schBackupList = append(schBackupList, restrictedScheduleBackup)
				mutex.Unlock()
				log.Infof("Schedule backup list after schedule backup with restricted namespaces - %v", schBackupList)
			})
		}()

		wg.Add(1)
		go func() {
			defer GinkgoRecover()
			defer wg.Done()
			Step("Creating schedule backup with baseline namespaces", func() {
				log.InfoD("Creating schedule backup with baseline namespaces")
				schPolicyUid, _ = Inst().Backup.GetSchedulePolicyUid(BackupOrgID, ctx, periodicPolicyName)
				schBackupName := fmt.Sprintf("schbkp-baseline-%v", RandomString(5))
				labelSelectors := make(map[string]string)
				baselineScheduleBackup, err = CreateScheduleBackupWithValidation(ctx, schBackupName, SourceClusterName, backupLocationName, backupLocationUID, baselineScheduledAppContexts, labelSelectors, BackupOrgID, preRuleName, preRuleUid, postRuleName, postRuleUid, periodicPolicyName, schPolicyUid)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of schedule backup %v with baseline namespaces", baselineScheduleBackup))
				mutex.Lock()
				scheduleNames = append(scheduleNames, schBackupName)
				schBackupList = append(schBackupList, baselineScheduleBackup)
				mutex.Unlock()
				log.Infof("Schedule backup list after schedule backup with baseline namespaces - %v", schBackupList)
			})
		}()
		wg.Wait()

		Step("Custom restore of restricted namespaces to destination cluster", func() {
			log.InfoD("Custom restore of restricted namespaces to destination cluster")
			restoreName := fmt.Sprintf("restore-restricted-%v", RandomString(5))
			err = CreateRestoreWithProjectMapping(restoreName, restrictedScheduleBackup, namespaceMapRestrictedToPrivilege, DestinationClusterName, BackupOrgID, ctx, nil, restoreProjectUIDMapping, restoreProjectNameMapping)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Custom restore %s from restricted namespace backup %s", restoreName, restrictedScheduleBackup))
		})

		Step("Default restore of baseline namespaces to destination cluster", func() {
			log.InfoD("Default restore of baseline namespaces to destination cluster")
			restoreName := fmt.Sprintf("restore-baseline-%v", RandomString(5))
			err = CreateRestoreWithValidation(ctx, restoreName, baselineScheduleBackup, make(map[string]string), make(map[string]string), DestinationClusterName, BackupOrgID, baselineScheduledAppContexts)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Default restore %s from baseline namespace backup %s", restoreName, baselineScheduleBackup))
		})

		Step("Restore all namespace backup to destination cluster with namespace & project mappings", func() {
			log.InfoD("Restore all namespace backup to destination cluster with namespace & project mappings")
			log.Infof("All namespace backup which is going to restore is %v", allNSScheduleBackup)
			restoreName := fmt.Sprintf("restore-all-%v", RandomString(5))
			err = CreateRestoreOnRancherWithoutCheck(restoreName, allNSScheduleBackup, allNamespaceMap, DestinationClusterName, BackupOrgID, ctx, nil, restoreProjectUIDMapping, restoreProjectNameMapping, 2)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creation of restore with replace option %s from backup %s", restoreName, allNSScheduleBackup))
			err = RestoreSuccessCheck(restoreName, BackupOrgID, MaxWaitPeriodForRestoreCompletionInMinute*time.Minute, RestoreJobProgressRetryTime*time.Minute, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying restore %s with all namespace backup with project mappings", restoreName))
		})
	})

	JustAfterEach(func() {
		log.InfoD("Cleaning up the resources after the test execution")
		defer func() {
			Inst().AppList = actualAppList
		}()
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		defer func() {
			log.Infof("switching to default context")
			err := SetClusterContext("")
			log.FailOnError(err, "Failed to SetClusterContext to default cluster")
		}()
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		log.Infof("Deleting the backups as part of clean up")
		allBackups, err := GetAllBackupsAdmin()
		dash.VerifySafely(err, nil, "Verifying fetching of all backups")
		for _, backupName := range allBackups {
			backupUID, err := Inst().Backup.GetBackupUID(ctx, backupName, BackupOrgID)
			dash.VerifySafely(err, nil, fmt.Sprintf("Getting backup UID for backup %s", backupName))
			_, err = DeleteBackup(backupName, backupUID, BackupOrgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Verifying backup deletion - %s", backupName))
		}
		log.Info("Destroying scheduled apps on source cluster")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		DestroyApps(scheduledAppContexts, opts)
		log.Info("Deleting schedules")
		for _, scheduleName := range scheduleNames {
			err = DeleteSchedule(scheduleName, SourceClusterName, BackupOrgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting schedule [%s]", scheduleName))
		}
		log.Infof("Deleting pre & post exec rules")
		allRules, _ := Inst().Backup.GetAllRules(ctx, BackupOrgID)
		for _, ruleName := range allRules {
			err := DeleteRule(ruleName, BackupOrgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Verifying deletion of rule [%s]", ruleName))
		}
		CleanupCloudSettingsAndClusters(backupLocationMap, credName, credUid, ctx)
	})
})
