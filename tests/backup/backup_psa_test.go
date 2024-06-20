package tests

import (
	"fmt"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"golang.org/x/sync/errgroup"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

var _ = Describe("{RestoreFromHigherPrivilegedNamespaceToLower}", Label(TestCaseLabelsMap[RestoreFromHigherPrivilegedNamespaceToLower]...), func() {
	var (
		backupNames                    []string
		restoreNames                   []string
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
	)
	AppContextsMapping := make(map[string]*scheduler.Context)
	providers = GetBackupProviders()
	label = make(map[string]string)
	backupLocationMap = make(map[string]string)
	scheduledAppContexts = make([]*scheduler.Context, 0)
	originalAppList := Inst().AppList
	appPrivilegeToBkpMap := make(map[string]string)
	appPrivilegeToRestoreMap := make(map[string][]string)
	appPrivilegeToNsMap := make(map[string]string)
	restrictedScheduledAppContexts = make([]*scheduler.Context, 0)
	baselineScheduledAppContexts = make([]*scheduler.Context, 0)
	mulAppScheduledAppContexts := make([]*scheduler.Context, 0)
	mulAppRestrictedNamespaceList := make([]string, 0)

	JustBeforeEach(func() {
		// TODO: Need to update testcase ID
		StartPxBackupTorpedoTest("RestoreFromHigherPrivilegedNamespaceToLower", "Restore from higher Privileged to lower Privileged namespace", nil, 299239, Sn, Q2FY25)

		log.InfoD("Deploy applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		// Define maps to store unique namespaces
		baselineNamespaceMap := make(map[string]struct{})
		restrictedNamespaceMap := make(map[string]struct{})
		psaApp := make([]string, 0)

		//Deploy multiple applications on multiple namespace on restricted and baseline namespaces
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
				dash.VerifyFatal(err, nil, "Creating namespace and assigning labels")
				appContexts := ScheduleApplicationsOnNamespace(namespace, taskName)
				for _, ctx := range appContexts {
					ctx.ReadinessTimeout = AppReadinessTimeout
					namespace := GetAppNamespace(ctx, taskName)
					scheduledAppContexts = append(scheduledAppContexts, ctx)
					AppContextsMapping[namespace] = ctx
					if strings.Contains(namespace, "restricted") {
						restrictedScheduledAppContexts = append(restrictedScheduledAppContexts, ctx)
						// Add to restrictedNamespaceMap if not already present
						if _, ok := restrictedNamespaceMap[namespace]; !ok {
							restrictedNamespaceMap[namespace] = struct{}{}
						}
					} else if strings.Contains(namespace, "baseline") {
						baselineScheduledAppContexts = append(baselineScheduledAppContexts, ctx)
						// Add to baselineNamespaceMap if not already present
						if _, ok := baselineNamespaceMap[namespace]; !ok {
							baselineNamespaceMap[namespace] = struct{}{}
						}
					}
				}
			}
			// Convert map keys to slice to store baseline namespaces
			baselineNamespaceList = make([]string, 0, len(baselineNamespaceMap))
			for ns := range baselineNamespaceMap {
				baselineNamespaceList = append(baselineNamespaceList, ns)
			}

			// Convert map keys to slice to store restricted namespaces
			restrictedNamespaceList = make([]string, 0, len(restrictedNamespaceMap))
			for ns := range restrictedNamespaceMap {
				restrictedNamespaceList = append(restrictedNamespaceList, ns)
			}

			if strings.Contains(psaLevel, "restricted") {
				Inst().AppList = originalAppList
			}
		}

		// Iterate over each application list to deploy single application on multiple namespace
		for i, app := range originalAppList {
			Inst().AppList = []string{app}
			// Generate a unique namespace name
			namespace := fmt.Sprintf("%s-%d-%s", app, i, RandomString(10))

			// Create labels and assign pod-security level
			label := make(map[string]string)
			label["pod-security.kubernetes.io/enforce"] = "baseline"

			// Create namespace and assign labels
			err := CreateNamespaceAndAssignLabels(namespace, label)
			dash.VerifyFatal(err, nil, "Creating namespace and assigning labels")

			// Schedule application on namespace
			taskName := fmt.Sprintf("%s-%d-%s", app, i, RandomString(10))
			actx := ScheduleApplicationsOnNamespace(namespace, taskName)

			for _, ctx := range actx {
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
				time.Sleep(1 * time.Minute)
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

		Step("Create namespace with different privileges for performing restores", func() {
			for _, psalevel := range []string{"restricted", "baseline", "privileged"} {
				psaNameSpaceList := make([]string, 0)

				for i := 0; i < len(originalAppList); i++ {
					namespace := fmt.Sprintf("%s-%s", psalevel, RandomString(10))
					label["pod-security.kubernetes.io/enforce"] = psalevel
					err := CreateNamespaceAndAssignLabels(namespace, label)
					dash.VerifyFatal(err, nil, "Creating namespace and assigning labels")
					psaNameSpaceList = append(psaNameSpaceList, namespace)
				}

				// Assign psaNameSpaceList to appPrivilegeToRestoreMap[psalevel]
				appPrivilegeToRestoreMap[psalevel] = psaNameSpaceList
			}
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
			if err != nil {
				log.Infof("error expected :")
				log.Infof(err.Error())
			}
			dash.VerifyFatal(strings.Contains(err.Error(), "failed to meet the pod security standard"), true, fmt.Sprintf("Creating restore from backup taken on baseline namespace and restore to restricted namespace failed as expected [%s]", restoreName))
		})

		Step("Remove restricted label from the namespace", func() {
			log.InfoD("Remove restricted label from the namespace [%s]", appPrivilegeToRestoreMap["restricted"])
			label["pod-security.kubernetes.io/enforce"] = "restricted"
			var err error
			for i := range originalAppList {
				err = DeleteLabelsFromNamespace(appPrivilegeToRestoreMap["restricted"][i], []string{"pod-security.kubernetes.io/enforce"})
			}
			dash.VerifyFatal(err, nil, "Deleting label from the namespace")
		})

		Step("Add baseline label to namespace", func() {
			log.InfoD("Add baseline label to the namespace [%s]", appPrivilegeToRestoreMap["restricted"])
			label["pod-security.kubernetes.io/enforce"] = "baseline"
			var err error
			for i := range originalAppList {
				err = Inst().S.AddNamespaceLabel(appPrivilegeToRestoreMap["restricted"][i], label)
			}
			log.FailOnError(err, "Failed to add labels %v to namespace %s", label, appPrivilegeToRestoreMap["restricted"])
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
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore from backup [%s]", restoreName))
		})

		Step("Perform a custom restore of the backup taken from the namespace in baseline mode to a namespace with restricted mode with replace option on different cluster", func() {
			log.InfoD("Perform a custom restore of the backup taken from the namespace in baseline mode to a namespace with restricted mode with replace option on different cluster")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Unable to fetch px-central-admin ctx")
			restrictedNamespaceList := make([]string, 0)

			for i := 0; i < len(originalAppList); i++ {
				namespace := fmt.Sprintf("%s-%s", "restricted-ns-1", RandomString(10))
				label["pod-security.kubernetes.io/enforce"] = "restricted"
				err = CreateNamespaceAndAssignLabels(namespace, label)
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
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore from backup [%s]", restoreName))
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
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore from backup [%s]", restoreName))
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
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore from backup [%s]", restoreName))
		})
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)

		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")

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

		log.InfoD("switching to default context")
		err = SetClusterContext("")
		log.FailOnError(err, "failed to SetClusterContext to default cluster")

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

		log.Info("Deleting restored namespaces")
		for _, restoreName := range restoreNames {
			err = DeleteRestore(restoreName, BackupOrgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting Restore [%s]", restoreName))
		}
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})
