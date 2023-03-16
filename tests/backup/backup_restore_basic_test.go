package tests

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/backup/portworx"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	v1 "k8s.io/api/core/v1"

	semver "github.com/blang/semver"
	"github.com/portworx/torpedo/drivers/scheduler/k8s"
)

// BasicSelectiveRestore selects random backed-up apps and restores them
var _ = Describe("{BasicSelectiveRestore}", func() {
	var (
		backupName        string
		contexts          []*scheduler.Context
		appContexts       []*scheduler.Context
		bkpNamespaces     []string
		clusterUid        string
		clusterStatus     api.ClusterInfo_StatusInfo_Status
		restoreName       string
		cloudCredName     string
		cloudCredUID      string
		backupLocationUID string
		bkpLocationName   string
		numDeployments    int
		providers         []string
		backupLocationMap map[string]string
		labelSelectors    map[string]string
	)
	JustBeforeEach(func() {
		backupName = fmt.Sprintf("%s-%v", BackupNamePrefix, time.Now().Unix())
		bkpNamespaces = make([]string, 0)
		restoreName = fmt.Sprintf("%s-%v", RestoreNamePrefix, time.Now().Unix())
		backupLocationMap = make(map[string]string)
		labelSelectors = make(map[string]string)

		numDeployments = 6 // For this test case to have relevance, it is necessary to raise the number of deployments.
		providers = getProviders()

		StartTorpedoTest("BasicSelectiveRestore", "All namespace backup and restore selective namespaces", nil, 83717)
		log.InfoD(fmt.Sprintf("App list %v", Inst().AppList))
		contexts = make([]*scheduler.Context, 0)
		log.InfoD("Starting to deploy applications")
		for i := 0; i < numDeployments; i++ {
			log.InfoD(fmt.Sprintf("Iteration %v of deploying applications", i))
			taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
			appContexts = ScheduleApplications(taskName)
			contexts = append(contexts, appContexts...)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = appReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
			}
		}
	})
	It("Selective Restore From a Basic Backup", func() {

		Step("Validating deployed applications", func() {
			log.InfoD("Validating deployed applications")
			ValidateApplications(contexts)
		})
		Step("Creating backup location and cloud setting", func() {
			log.InfoD("Creating backup location and cloud setting")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
				bkpLocationName = fmt.Sprintf("%s-%s-bl", provider, getGlobalBucketName(provider))
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = bkpLocationName
				CreateCloudCredential(provider, cloudCredName, cloudCredUID, orgID)
				err := CreateBackupLocation(provider, bkpLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), orgID, "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", bkpLocationName))
			}
		})
		Step("Registering cluster for backup", func() {
			log.InfoD("Registering cluster for backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = CreateSourceAndDestClusters(orgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, clusterUid = Inst().Backup.RegisterBackupCluster(orgID, SourceClusterName, "")
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying backup cluster with uid: [%s]", clusterUid))
		})
		Step("Taking backup of multiple namespaces", func() {
			log.InfoD(fmt.Sprintf("Taking backup of multiple namespaces [%v]", bkpNamespaces))
			ctx, err := backup.GetAdminCtxFromSecret()
			dash.VerifyFatal(err, nil, "Getting context")
			err = CreateBackup(backupName, SourceClusterName, bkpLocationName, backupLocationUID, bkpNamespaces,
				labelSelectors, orgID, clusterUid, "", "", "", "", ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying [%s] backup creation", backupName))
		})
		Step("Selecting random backed-up apps and restoring them", func() {
			log.InfoD("Selecting random backed-up apps and restoring them")
			selectedBkpNamespaces, err := GetSubsetOfSlice(bkpNamespaces, len(bkpNamespaces)/2)
			log.FailOnError(err, "Getting a subset of backed-up namespaces")
			selectedBkpNamespaceMapping := make(map[string]string)
			for _, namespace := range selectedBkpNamespaces {
				selectedBkpNamespaceMapping[namespace] = namespace
			}
			log.InfoD("Selected application namespaces to restore: [%v]", selectedBkpNamespaces)
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = CreateRestore(restoreName, backupName, selectedBkpNamespaceMapping, destinationClusterName, orgID, ctx, make(map[string]string))
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s]", restoreName))
		})
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(contexts)
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		log.InfoD("Deleting deployed applications")
		ValidateAndDestroy(contexts, opts)

		backupDriver := Inst().Backup
		backupUID, err := backupDriver.GetBackupUID(ctx, backupName, orgID)
		log.FailOnError(err, "Failed while trying to get backup UID for - [%s]", backupName)

		log.InfoD("Deleting backup")
		_, err = DeleteBackup(backupName, backupUID, orgID, ctx)
		dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting backup [%s]", backupName))

		log.InfoD("Deleting restore")
		log.InfoD(fmt.Sprintf("Backup name [%s]", restoreName))
		err = DeleteRestore(restoreName, orgID, ctx)
		dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting restore [%s]", restoreName))

		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})

// This test does custom resource backup and restore.
var _ = Describe("{CustomResourceBackupAndRestore}", func() {
	namespaceMapping := make(map[string]string)
	var contexts []*scheduler.Context
	labelSelectors := make(map[string]string)
	CloudCredUIDMap := make(map[string]string)
	var appContexts []*scheduler.Context
	var backupLocation string
	var backupLocationUID string
	var cloudCredUID string
	backupLocationMap := make(map[string]string)
	var bkpNamespaces []string
	var clusterUid string
	var cloudCredName string
	var clusterStatus api.ClusterInfo_StatusInfo_Status
	var backupName string
	var restoreName string
	var backupNames []string
	var restoreNames []string
	bkpNamespaces = make([]string, 0)

	JustBeforeEach(func() {
		StartTorpedoTest("CustomResourceBackupAndRestore", "Create custom resource backup and restore", nil, 58043)
		log.InfoD("Deploy applications")

		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
			appContexts = ScheduleApplications(taskName)
			contexts = append(contexts, appContexts...)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = appReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
			}
		}
	})
	It("Create custom resource backup and restore", func() {
		Step("Validate applications", func() {
			ValidateApplications(contexts)
		})

		Step("Creating cloud credentials", func() {
			log.InfoD("Creating cloud credentials")
			providers := getProviders()
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
				cloudCredUID = uuid.New()
				CloudCredUIDMap[cloudCredUID] = cloudCredName
				CreateCloudCredential(provider, cloudCredName, cloudCredUID, orgID)
			}
		})

		Step("Register cluster for backup", func() {
			ctx, _ := backup.GetAdminCtxFromSecret()
			err := CreateSourceAndDestClusters(orgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, clusterUid = Inst().Backup.RegisterBackupCluster(orgID, SourceClusterName, "")
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying backup cluster %s creation", SourceClusterName))
		})

		Step("Creating backup location", func() {
			log.InfoD("Creating backup location")
			providers := getProviders()
			for _, provider := range providers {
				backupLocation = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocation
				err := CreateBackupLocation(provider, backupLocation, backupLocationUID, cloudCredName, cloudCredUID,
					getGlobalBucketName(provider), orgID, "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", backupLocation))
			}
		})

		Step("Taking backup of applications", func() {
			log.InfoD("Taking backup of applications")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, namespace := range bkpNamespaces {
				backupName = fmt.Sprintf("%s-%s-%v", BackupNamePrefix, namespace, time.Now().Unix())
				backupNames = append(backupNames, backupName)
				err = CreateBackupWithCustomResourceType(backupName, SourceClusterName, backupLocation, backupLocationUID, []string{namespace}, nil, orgID, clusterUid, "", "", "", "", []string{"PersistentVolumeClaim"}, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying backup %s creation with custom resources", backupName))
			}
		})

		Step("Restoring the backed up application", func() {
			log.InfoD("Restoring the backed up application")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for _, namespace := range bkpNamespaces {
				restoreName = fmt.Sprintf("%s-%s-%v", restoreNamePrefix, backupName, time.Now().Unix())
				restoreNames = append(restoreNames, restoreName)
				restoredNameSpace := fmt.Sprintf("%s-%s", namespace, "restored")
				namespaceMapping[namespace] = restoredNameSpace
				err = CreateRestore(restoreName, backupName, namespaceMapping, SourceClusterName, orgID, ctx, make(map[string]string))
				log.FailOnError(err, "Restoring of backup [%s] has failed with name - [%s]", backupName, restoreName)
			}
		})

		Step("Compare PVCs on both namespaces", func() {
			log.InfoD("Compare PVCs on both namespaces")
			for _, namespace := range bkpNamespaces {
				pvcs, _ := core.Instance().GetPersistentVolumeClaims(namespace, labelSelectors)
				restoreNamespace := fmt.Sprintf("%s-%s", namespace, "restored")
				restoredPvcs, _ := core.Instance().GetPersistentVolumeClaims(restoreNamespace, labelSelectors)
				dash.VerifyFatal(len(pvcs.Items), len(restoredPvcs.Items), "Compare number of PVCs")
			}
		})
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(contexts)
		ctx, _ := backup.GetAdminCtxFromSecret()
		log.InfoD("Deleting the deployed apps after the testcase")
		for i := 0; i < len(contexts); i++ {
			opts := make(map[string]bool)
			opts[SkipClusterScopedObjects] = true
			taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
			err := Inst().S.Destroy(contexts[i], opts)
			dash.VerifySafely(err, nil, fmt.Sprintf("Verify destroying app %s", taskName))
		}
		for _, restore := range restoreNames {
			err := DeleteRestore(restore, orgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting Restore %s", restore))
		}
		for _, backupName := range backupNames {
			backupUID, err := Inst().Backup.GetBackupUID(ctx, backupName, orgID)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Getting backup UID for backup %s", backupName))
			_, err = DeleteBackup(backupName, backupUID, orgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting backup - %s", backupName))
		}
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})

// DeleteAllBackupObjects deletes all backed up objects
var _ = Describe("{DeleteAllBackupObjects}", func() {
	var (
		appList           = Inst().AppList
		backupName        string
		contexts          []*scheduler.Context
		preRuleNameList   []string
		postRuleNameList  []string
		appContexts       []*scheduler.Context
		bkpNamespaces     []string
		clusterUid        string
		clusterStatus     api.ClusterInfo_StatusInfo_Status
		restoreName       string
		cloudCredName     string
		cloudCredUID      string
		backupLocationUID string
		bkpLocationName   string
		preRuleName       string
		postRuleName      string
		preRuleUid        string
		postRuleUid       string
	)
	backupLocationMap := make(map[string]string)
	labelSelectors := make(map[string]string)
	bkpNamespaces = make([]string, 0)
	var namespaceMapping map[string]string
	namespaceMapping = make(map[string]string)
	intervalName := fmt.Sprintf("%s-%v", "interval", time.Now().Unix())
	JustBeforeEach(func() {
		StartTorpedoTest("DeleteAllBackupObjects", "Create the backup Objects and Delete", nil, 58088)
		log.InfoD("Verifying if the pre/post rules for the required apps are present in the AppParameters or not ")
		for i := 0; i < len(appList); i++ {
			if Contains(postRuleApp, appList[i]) {
				if _, ok := portworx.AppParameters[appList[i]]["post"]; ok {
					dash.VerifyFatal(ok, true, "Post Rule details mentioned for the apps")
				}
			}
			if Contains(preRuleApp, appList[i]) {
				if _, ok := portworx.AppParameters[appList[i]]["pre"]; ok {
					dash.VerifyFatal(ok, true, "Pre Rule details mentioned for the apps")
				}
			}
		}
		log.InfoD("Deploy applications")
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
			appContexts = ScheduleApplications(taskName)
			contexts = append(contexts, appContexts...)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = appReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
			}
		}
	})
	It("Create backup objects and delete", func() {
		providers := getProviders()

		Step("Validate applications", func() {
			ValidateApplications(contexts)
		})
		Step("Creating rules for backup", func() {
			log.InfoD("Creating pre rule for deployed apps")
			for i := 0; i < len(appList); i++ {
				preRuleStatus, ruleName, err := Inst().Backup.CreateRuleForBackup(appList[i], orgID, "pre")
				log.FailOnError(err, "Creating pre rule %s for deployed apps failed", ruleName)
				dash.VerifyFatal(preRuleStatus, true, "Verifying pre rule for backup")

				if ruleName != "" {
					preRuleNameList = append(preRuleNameList, ruleName)
				}
			}
			log.InfoD("Creating post rule for deployed apps")
			for i := 0; i < len(appList); i++ {
				postRuleStatus, ruleName, err := Inst().Backup.CreateRuleForBackup(appList[i], orgID, "post")
				log.FailOnError(err, "Creating post %s rule for deployed apps failed", ruleName)
				dash.VerifyFatal(postRuleStatus, true, "Verifying Post rule for backup")
				if ruleName != "" {
					postRuleNameList = append(postRuleNameList, ruleName)
				}
			}
		})
		Step("Creating cloud account and backup location", func() {
			log.InfoD("Creating cloud account and backup location")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
				bkpLocationName = fmt.Sprintf("%s-%s-%v", provider, getGlobalBucketName(provider), time.Now().Unix())
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = bkpLocationName
				CreateCloudCredential(provider, cloudCredName, cloudCredUID, orgID)
				err := CreateBackupLocation(provider, bkpLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), orgID, "")
				log.FailOnError(err, "Creating backup location %s failed", bkpLocationName)
			}
		})
		Step("Creating backup schedule policy", func() {
			log.InfoD("Creating a backup schedule policy")
			intervalSchedulePolicyInfo := Inst().Backup.CreateIntervalSchedulePolicy(5, 15, 2)
			intervalPolicyStatus := Inst().Backup.BackupSchedulePolicy(intervalName, uuid.New(), orgID, intervalSchedulePolicyInfo)
			dash.VerifyFatal(intervalPolicyStatus, nil, fmt.Sprintf("Creating interval schedule policy %s", intervalName))
		})
		Step("Register cluster for backup", func() {
			log.InfoD("Register cluster for backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = CreateSourceAndDestClusters(orgID, "", "", ctx)
			log.FailOnError(err, "Creation of source and destination cluster")
			clusterStatus, clusterUid = Inst().Backup.RegisterBackupCluster(orgID, SourceClusterName, "")
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying backup cluster %s", SourceClusterName))
		})
		Step("Taking backup of applications", func() {
			log.InfoD("Taking backup of applications")
			ctx, err := backup.GetAdminCtxFromSecret()
			dash.VerifyFatal(err, nil, "Getting context")
			if len(preRuleNameList) > 0 {
				preRuleUid, err = Inst().Backup.GetRuleUid(orgID, ctx, preRuleNameList[0])
				log.FailOnError(err, "Failed to get UID for rule %s", preRuleNameList[0])
				preRuleName = preRuleNameList[0]
			} else {
				preRuleUid = ""
				preRuleName = ""
			}
			if len(postRuleNameList) > 0 {
				postRuleUid, err = Inst().Backup.GetRuleUid(orgID, ctx, postRuleNameList[0])
				log.FailOnError(err, "Failed to get UID for rule %s", postRuleNameList[0])
				postRuleName = postRuleNameList[0]
			} else {
				postRuleUid = ""
				postRuleName = ""
			}
			for _, namespace := range bkpNamespaces {
				backupName = fmt.Sprintf("%s-%s-%v", BackupNamePrefix, namespace, time.Now().Unix())
				err = CreateBackup(backupName, SourceClusterName, bkpLocationName, backupLocationUID, []string{namespace},
					labelSelectors, orgID, clusterUid, preRuleName, preRuleUid, postRuleName, postRuleUid, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying %s backup creation", backupName))
			}
		})
		Step("Restoring the backed up applications", func() {
			log.InfoD("Restoring the backed up applications")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			restoreName = fmt.Sprintf("%s-%v", "test-restore", time.Now().Unix())
			err = CreateRestore(restoreName, backupName, namespaceMapping, destinationClusterName, orgID, ctx, make(map[string]string))
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying %s backup's restore %s creation", backupName, restoreName))
		})

		Step("Delete the restores", func() {
			log.InfoD("Delete the restores")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = DeleteRestore(restoreName, orgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying restore %s deletion", restoreName))
		})
		Step("Delete the backups", func() {
			log.Infof("Delete the backups")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			backupDriver := Inst().Backup
			backupUID, err := backupDriver.GetBackupUID(ctx, backupName, orgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupName)
			_, err = DeleteBackup(backupName, backupUID, orgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying backup %s deletion", backupName))

		})
		Step("Delete backup schedule policy", func() {
			log.InfoD("Delete backup schedule policy")
			policyList := []string{intervalName}
			err := Inst().Backup.DeleteBackupSchedulePolicy(orgID, policyList)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting backup schedule policies %s ", policyList))
		})
		Step("Delete the pre and post rules", func() {
			log.InfoD("Delete the pre rule")
			if len(preRuleNameList) > 0 {
				for _, ruleName := range preRuleNameList {
					err := Inst().Backup.DeleteRuleForBackup(orgID, ruleName)
					dash.VerifySafely(err, nil, fmt.Sprintf("Deleting  backup pre rules %s", ruleName))
				}
			}
			log.InfoD("Delete the post rules")
			if len(postRuleNameList) > 0 {
				for _, ruleName := range postRuleNameList {
					err := Inst().Backup.DeleteRuleForBackup(orgID, ruleName)
					dash.VerifySafely(err, nil, fmt.Sprintf("Deleting  backup post rules %s", ruleName))
				}
			}
		})
		Step("Delete the backup location and cloud account", func() {
			log.InfoD("Delete the backup location %s and cloud account %s", bkpLocationName, cloudCredName)
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
		})
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(contexts)
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		log.Info(" Deleting deployed applications")
		ValidateAndDestroy(contexts, opts)
	})
})

// This testcase verifies schedule backup creation with a single namespace.
var _ = Describe("{ScheduleBackupCreationSingleNS}", func() {
	var (
		contexts           []*scheduler.Context
		appContexts        []*scheduler.Context
		backupLocationName string
		backupLocationUID  string
		cloudCredUID       string
		bkpNamespaces      []string
		cloudAccountName   string
		backupName         string
		schBackupName      string
		schPolicyUid       string
		restoreName        string
		clusterStatus      api.ClusterInfo_StatusInfo_Status
	)
	var testrailID = 58014 // testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/58014
	namespaceMapping := make(map[string]string)
	labelSelectors := make(map[string]string)
	cloudCredUIDMap := make(map[string]string)
	backupLocationMap := make(map[string]string)
	bkpNamespaces = make([]string, 0)
	timeStamp := strconv.Itoa(int(time.Now().Unix()))
	periodicPolicyName := fmt.Sprintf("%s-%s", "periodic", timeStamp)

	JustBeforeEach(func() {
		StartTorpedoTest("ScheduleBackupCreationSingleNS", "Create schedule backup creation with a single namespace", nil, testrailID)
		log.Info("Application installation")
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
			appContexts = ScheduleApplications(taskName)
			contexts = append(contexts, appContexts...)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = appReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
			}
		}
	})

	It("Schedule Backup Creation with single namespace", func() {
		Step("Validate deployed applications", func() {
			ValidateApplications(contexts)
		})
		providers := getProviders()
		Step("Adding Cloud Account", func() {
			log.InfoD("Adding cloud account")
			for _, provider := range providers {
				cloudAccountName = fmt.Sprintf("%s-%v", provider, timeStamp)
				cloudCredUID = uuid.New()
				cloudCredUIDMap[cloudCredUID] = cloudAccountName
				CreateCloudCredential(provider, cloudAccountName, cloudCredUID, orgID)
			}
		})

		Step("Adding Backup Location", func() {
			log.InfoD("Adding Backup Location")
			for _, provider := range providers {
				cloudAccountName = fmt.Sprintf("%s-%v", provider, timeStamp)
				backupLocationName = fmt.Sprintf("auto-bl-%v", time.Now().Unix())
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocationName
				err := CreateBackupLocation(provider, backupLocationName, backupLocationUID, cloudAccountName, cloudCredUID,
					getGlobalBucketName(provider), orgID, "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of adding backup location - %s", backupLocationName))
			}
		})

		Step("Creating Schedule Policies", func() {
			log.InfoD("Creating Schedule Policies")
			periodicSchedulePolicyInfo := Inst().Backup.CreateIntervalSchedulePolicy(5, 15, 5)
			periodicPolicyStatus := Inst().Backup.BackupSchedulePolicy(periodicPolicyName, uuid.New(), orgID, periodicSchedulePolicyInfo)
			dash.VerifyFatal(periodicPolicyStatus, nil, fmt.Sprintf("Verification of creating periodic schedule policy - %s", periodicPolicyName))
		})

		Step("Adding Clusters for backup", func() {
			log.InfoD("Adding application clusters")
			ctx, _ := backup.GetAdminCtxFromSecret()
			err := CreateSourceAndDestClusters(orgID, "", "", ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of creating source - %s and destination - %s clusters", SourceClusterName, destinationClusterName))
			clusterStatus, _ = Inst().Backup.RegisterBackupCluster(orgID, SourceClusterName, "")
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verification of adding application clusters - %s", SourceClusterName))
		})

		Step("Creating schedule backups", func() {
			log.InfoD("Creating schedule backups")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			schPolicyUid, _ = Inst().Backup.GetSchedulePolicyUid(orgID, ctx, periodicPolicyName)
			for _, namespace := range bkpNamespaces {
				backupName = fmt.Sprintf("%s-%s", BackupNamePrefix, namespace)
				err = CreateScheduleBackup(backupName, SourceClusterName, backupLocationName, backupLocationUID, []string{namespace},
					labelSelectors, orgID, "", "", "", "", periodicPolicyName, schPolicyUid, ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of creating schedule backup with schedule name - %s", backupName))
				schBackupName, err = GetFirstScheduleBackupName(ctx, backupName, orgID)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching the name of the first schedule backup - %s", schBackupName))
			}
		})

		Step("Restoring scheduled backups", func() {
			log.InfoD("Restoring scheduled backups")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			restoreName = fmt.Sprintf("%s-%s", restoreNamePrefix, schBackupName)
			err = CreateRestore(restoreName, schBackupName, namespaceMapping, destinationClusterName, orgID, ctx, nil)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of restoring scheduled backups - %s", restoreName))
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(contexts)
		ctx, _ := backup.GetAdminCtxFromSecret()
		log.InfoD("Clean up objects after test execution")
		log.Info("Deleting backup schedules")
		scheduleUid, err := GetScheduleUID(backupName, orgID, ctx)
		log.FailOnError(err, "Error while getting schedule uid %v", backupName)
		err = DeleteSchedule(backupName, scheduleUid, periodicPolicyName, schPolicyUid, orgID)
		dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of deleting backup schedules - %s", backupName))
		log.Info("Deleting restores")
		err = DeleteRestore(restoreName, orgID, ctx)
		dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of deleting restores - %s", restoreName))
		log.Info("Deleting the deployed apps after test execution")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		for i := 0; i < len(contexts); i++ {
			taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
			err := Inst().S.Destroy(contexts[i], opts)
			dash.VerifySafely(err, nil, fmt.Sprintf("Verify destroying application %s", taskName))
		}
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudAccountName, cloudCredUID, ctx)
	})
})

// This testcase verifies schedule backup creation with all namespaces.
var _ = Describe("{ScheduleBackupCreationAllNS}", func() {
	var (
		contexts           []*scheduler.Context
		appContexts        []*scheduler.Context
		backupLocationName string
		backupLocationUID  string
		cloudCredUID       string
		bkpNamespaces      []string
		cloudAccountName   string
		backupName         string
		schBackupName      string
		schPolicyUid       string
		restoreName        string
		clusterStatus      api.ClusterInfo_StatusInfo_Status
	)
	var testrailID = 58015 // testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/58015
	namespaceMapping := make(map[string]string)
	labelSelectors := make(map[string]string)
	cloudCredUIDMap := make(map[string]string)
	backupLocationMap := make(map[string]string)
	bkpNamespaces = make([]string, 0)
	timeStamp := strconv.Itoa(int(time.Now().Unix()))
	periodicPolicyName := fmt.Sprintf("%s-%s", "periodic", timeStamp)

	JustBeforeEach(func() {
		StartTorpedoTest("ScheduleBackupCreationAllNS", "Create schedule backup creation with all namespaces", nil, testrailID)
		log.Info("Application installation")
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
			appContexts = ScheduleApplications(taskName)
			contexts = append(contexts, appContexts...)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = appReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
			}
		}
	})

	It("Schedule Backup Creation with all namespaces", func() {
		Step("Validate deployed applications", func() {
			ValidateApplications(contexts)
		})
		providers := getProviders()
		Step("Adding Cloud Account", func() {
			log.InfoD("Adding cloud account")
			for _, provider := range providers {
				cloudAccountName = fmt.Sprintf("%s-%v", provider, timeStamp)
				cloudCredUID = uuid.New()
				cloudCredUIDMap[cloudCredUID] = cloudAccountName
				CreateCloudCredential(provider, cloudAccountName, cloudCredUID, orgID)
			}
		})

		Step("Adding Backup Location", func() {
			for _, provider := range providers {
				cloudAccountName = fmt.Sprintf("%s-%v", provider, timeStamp)
				backupLocationName = fmt.Sprintf("auto-bl-%v", time.Now().Unix())
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocationName
				err := CreateBackupLocation(provider, backupLocationName, backupLocationUID, cloudAccountName, cloudCredUID,
					getGlobalBucketName(provider), orgID, "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Adding Backup Location - %s", backupLocationName))
			}
		})

		Step("Creating Schedule Policies", func() {
			log.InfoD("Adding application clusters")
			periodicSchedulePolicyInfo := Inst().Backup.CreateIntervalSchedulePolicy(5, 15, 5)
			periodicPolicyStatus := Inst().Backup.BackupSchedulePolicy(periodicPolicyName, uuid.New(), orgID, periodicSchedulePolicyInfo)
			dash.VerifyFatal(periodicPolicyStatus, nil, fmt.Sprintf("Verification of creating periodic schedule policy - %s", periodicPolicyName))
		})

		Step("Adding Clusters for backup", func() {
			log.InfoD("Adding application clusters")
			ctx, _ := backup.GetAdminCtxFromSecret()
			err := CreateSourceAndDestClusters(orgID, "", "", ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of creating source - %s and destination - %s clusters", SourceClusterName, destinationClusterName))
			clusterStatus, _ = Inst().Backup.RegisterBackupCluster(orgID, SourceClusterName, "")
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verification of adding application clusters - %s", SourceClusterName))
		})

		Step("Creating schedule backups", func() {
			log.InfoD("Creating schedule backups")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			schPolicyUid, _ = Inst().Backup.GetSchedulePolicyUid(orgID, ctx, periodicPolicyName)
			backupName = fmt.Sprintf("%s-schedule-%v", BackupNamePrefix, timeStamp)
			err = CreateScheduleBackup(backupName, SourceClusterName, backupLocationName, backupLocationUID, bkpNamespaces,
				labelSelectors, orgID, "", "", "", "", periodicPolicyName, schPolicyUid, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of creating schedule backup with schedule name - %s", backupName))
			schBackupName, err = GetFirstScheduleBackupName(ctx, backupName, orgID)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching the name of the first schedule backup - %s", schBackupName))
		})

		Step("Restoring scheduled backups", func() {
			log.InfoD("Restoring scheduled backups")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			restoreName = fmt.Sprintf("%s-%s", restoreNamePrefix, schBackupName)
			err = CreateRestore(restoreName, schBackupName, namespaceMapping, destinationClusterName, orgID, ctx, nil)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of restoring scheduled backups - %s", restoreName))
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(contexts)
		ctx, _ := backup.GetAdminCtxFromSecret()
		log.InfoD("Clean up objects after test execution")
		log.Info("Deleting backup schedules")
		scheduleUid, err := GetScheduleUID(backupName, orgID, ctx)
		log.FailOnError(err, "Error while getting schedule uid %v", backupName)
		err = DeleteSchedule(backupName, scheduleUid, periodicPolicyName, schPolicyUid, orgID)
		dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of deleting backup schedules - %s", backupName))
		log.Info("Deleting restores")
		err = DeleteRestore(restoreName, orgID, ctx)
		dash.VerifyFatal(err, nil, fmt.Sprintf("Verification of deleting restores - %s", restoreName))
		log.Info("Deleting the deployed applications after test execution")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		for i := 0; i < len(contexts); i++ {
			taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
			err := Inst().S.Destroy(contexts[i], opts)
			dash.VerifySafely(err, nil, fmt.Sprintf("Verify destroying application %s", taskName))
		}
		log.Info("Deleting backup location, cloud credentials and clusters")
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudAccountName, cloudCredUID, ctx)
	})
})

var _ = Describe("{CustomResourceRestore}", func() {
	var (
		contexts           []*scheduler.Context
		appContexts        []*scheduler.Context
		backupLocationUID  string
		cloudCredUID       string
		bkpNamespaces      []string
		clusterUid         string
		clusterStatus      api.ClusterInfo_StatusInfo_Status
		backupName         string
		credName           string
		cloudCredUidList   []string
		backupLocationName string
		deploymentName     string
		restoreName        string
		backupNames        []string
		restoreNames       []string
	)
	labelSelectors := make(map[string]string)
	namespaceMapping := make(map[string]string)
	newBackupLocationMap := make(map[string]string)
	backupNamespaceMap := make(map[string]string)
	deploymentBackupMap := make(map[string]string)
	bkpNamespaces = make([]string, 0)

	JustBeforeEach(func() {
		StartTorpedoTest("CustomResourceRestore", "Create custom resource restore", nil, 58041)
		log.InfoD("Deploy applications")

		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
			appContexts = ScheduleApplications(taskName)
			contexts = append(contexts, appContexts...)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = appReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
			}
		}
	})
	It("Create custom resource restore", func() {
		providers := getProviders()
		Step("Validate applications", func() {
			log.InfoD("Validate applications")
			ValidateApplications(contexts)
		})

		Step("Creating credentials and backup location", func() {
			log.InfoD("Creating cloud credentials and backup location")
			for _, provider := range providers {
				cloudCredUID = uuid.New()
				cloudCredUidList = append(cloudCredUidList, cloudCredUID)
				backupLocationUID = uuid.New()
				credName = fmt.Sprintf("autogenerated-cred-%v", time.Now().Unix())
				CreateCloudCredential(provider, credName, cloudCredUID, orgID)
				log.InfoD("Created Cloud Credentials with name - %s", credName)
				backupLocationName = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
				newBackupLocationMap[backupLocationUID] = backupLocationName
				err := CreateBackupLocation(provider, backupLocationName, backupLocationUID, credName, cloudCredUID, getGlobalBucketName(provider), orgID, "")
				log.FailOnError(err, "Creating Backup location [%v] failed", backupLocationName)
				log.InfoD("Created Backup Location with name - %s", backupLocationName)
			}
		})
		Step("Register source and destination cluster for backup", func() {
			log.InfoD("Register source and destination cluster for backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx failed")
			err = CreateSourceAndDestClusters(orgID, "", "", ctx)
			log.FailOnError(err, "Creation of Source and destination cluster failed")
			clusterStatus, clusterUid = Inst().Backup.RegisterBackupCluster(orgID, SourceClusterName, "")
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying backup cluster %s", SourceClusterName))
		})
		Step("Taking backup of applications", func() {
			log.InfoD("Taking backup of applications")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx failed")
			for _, namespace := range bkpNamespaces {
				backupName = fmt.Sprintf("%s-%s-%v", BackupNamePrefix, namespace, time.Now().Unix())
				backupNamespaceMap[namespace] = backupName
				err = CreateBackup(backupName, SourceClusterName, backupLocationName, backupLocationUID, []string{namespace}, labelSelectors, orgID, clusterUid, "", "", "", "", ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying backup creation %s", backupName))
				backupNames = append(backupNames, backupName)
			}
		})
		Step("Restoring the backed up application", func() {
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx failed")
			log.InfoD("Restoring backed up applications")
			for _, namespace := range bkpNamespaces {
				backupName := backupNamespaceMap[namespace]
				restoreName = fmt.Sprintf("%s-%s-%v", restoreNamePrefix, backupName, time.Now().Unix())
				restoreNames = append(restoreNames, restoreName)
				restoredNameSpace := fmt.Sprintf("%s-%s", namespace, "restored")
				namespaceMapping[namespace] = restoredNameSpace
				deploymentName, err = CreateCustomRestoreWithPVCs(restoreName, backupName, namespaceMapping, SourceClusterName, orgID, ctx, make(map[string]string), namespace)
				deploymentBackupMap[backupName] = deploymentName
				log.FailOnError(err, "Restoring of backup [%s] has failed with name [%s] in namespace [%s]", backupName, restoreName, restoredNameSpace)
			}
		})

		Step("Validating restored resources", func() {
			log.InfoD("Validating restored resources")
			for _, namespace := range bkpNamespaces {
				restoreNamespace := fmt.Sprintf("%s-%s", namespace, "restored")
				backupName := backupNamespaceMap[namespace]
				deploymentName = deploymentBackupMap[backupName]
				deploymentStatus, err := apps.Instance().DescribeDeployment(deploymentName, restoreNamespace)
				log.FailOnError(err, "unable to fetch deployment status for %v", deploymentName)
				status := deploymentStatus.Conditions[1].Status
				dash.VerifyFatal(status, v1.ConditionTrue, fmt.Sprintf("checking the deployment status for %v in namespace %v", deploymentName, restoreNamespace))
			}
		})
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(contexts)
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx failed")

		//Delete Backup
		log.InfoD("Deleting backup")
		backupDriver := Inst().Backup
		for _, backupName := range backupNames {
			backupUID, err := backupDriver.GetBackupUID(ctx, backupName, orgID)
			dash.VerifySafely(err, nil, fmt.Sprintf("trying to get backup UID for backup %s", backupName))
			log.Infof("About to delete backup - %s", backupName)
			_, err = DeleteBackup(backupName, backupUID, orgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Verifying backup %s deletion is successful", backupName))
		}

		//Delete Restore
		log.InfoD("Deleting restore")
		for _, restoreName := range restoreNames {
			err = DeleteRestore(restoreName, orgID, ctx)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting user restore %s", restoreName))
		}

		log.Infof("Deleting the deployed apps after the testcase")
		for i := 0; i < len(contexts); i++ {
			opts := make(map[string]bool)
			opts[SkipClusterScopedObjects] = true
			taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
			err := Inst().S.Destroy(contexts[i], opts)
			dash.VerifySafely(err, nil, fmt.Sprintf("Verify destroying app %s, Err: %v", taskName, err))
		}

		log.Infof("Deleting backup location, cloud credentials and clusters")
		CleanupCloudSettingsAndClusters(newBackupLocationMap, credName, cloudCredUID, ctx)

	})
})

var _ = Describe("{AllNSBackupWithIncludeNewNSOption}", func() {
	var (
		contexts                   []*scheduler.Context
		cloudCredUID               string
		cloudCredName              string
		backupLocationName         string
		backupLocationUID          string
		backupLocationMap          map[string]string
		periodicSchedulePolicyName string
		periodicSchedulePolicyUid  string
		scheduleName               string
		appNamespaces              []string
		appClusterName             string
		restoreName                string
		nextScheduleBackupName     interface{}
	)

	JustBeforeEach(func() {
		StartTorpedoTest("AllNSBackupWithIncludeNewNSOption", "Verification of schedule backups created with include new namespaces option", nil, 84760)
	})

	It("Validates schedule backups created with include new namespaces option includes newly created namespaces", func() {
		Step("Create cloud credentials and backup locations", func() {
			log.InfoD("Creating cloud credentials and backup locations")
			providers := getProviders()
			backupLocationMap = make(map[string]string)
			for _, provider := range providers {
				cloudCredUID = uuid.New()
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
				log.InfoD("Creating cloud credential named [%s] and uid [%s] using [%s] as provider", cloudCredUID, cloudCredName, provider)
				CreateCloudCredential(provider, cloudCredName, cloudCredUID, orgID)
				backupLocationName = fmt.Sprintf("%s-%s-bl", provider, getGlobalBucketName(provider))
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocationName
				bucketName := getGlobalBucketName(provider)
				err := CreateBackupLocation(provider, backupLocationName, backupLocationUID, cloudCredName, cloudCredUID, bucketName, orgID, "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of backup location named [%s] with uid [%s] of [%s] as provider", backupLocationName, backupLocationUID, provider))
			}
		})
		Step("Configure source and destination clusters with px-central-admin ctx", func() {
			log.InfoD("Configuring source and destination clusters with px-central-admin ctx")
			ctx, err := backup.GetAdminCtxFromSecret()
			dash.VerifyFatal(err, nil, "Fetching px-central-admin ctx")
			err = CreateSourceAndDestClusters(orgID, "", "", ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of source [%s] and destination [%s] clusters with px-central-admin ctx", SourceClusterName, destinationClusterName))
			appClusterName = destinationClusterName
			clusterStatus, clusterUid := Inst().Backup.RegisterBackupCluster(orgID, appClusterName, "")
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying registration of cluster named [%s] with uid [%s] as backup cluster", appClusterName, clusterUid))
		})
		Step("Create schedule policy", func() {
			log.InfoD("Creating a schedule policy")
			ctx, err := backup.GetAdminCtxFromSecret()
			dash.VerifyFatal(err, nil, "Fetching px-central-admin ctx")
			periodicSchedulePolicyName = fmt.Sprintf("%s-%v", "periodic", time.Now().Unix())
			periodicSchedulePolicyUid = uuid.New()
			periodicSchedulePolicyInfo := Inst().Backup.CreateIntervalSchedulePolicy(5, 15, 5)
			err = Inst().Backup.BackupSchedulePolicy(periodicSchedulePolicyName, periodicSchedulePolicyUid, orgID, periodicSchedulePolicyInfo)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of periodic schedule policy of interval 15 minutes named [%s]", periodicSchedulePolicyName))
			periodicSchedulePolicyUid, err = Inst().Backup.GetSchedulePolicyUid(orgID, ctx, periodicSchedulePolicyName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching uid of periodic schedule policy named [%s]", periodicSchedulePolicyName))
		})
		Step("Create schedule backups", func() {
			log.InfoD("Creating a schedule backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			dash.VerifyFatal(err, nil, "Fetching px-central-admin ctx")
			scheduleName = fmt.Sprintf("%s-schedule-%v", BackupNamePrefix, time.Now().Unix())
			namespaces := []string{"*"}
			labelSelectors := make(map[string]string)
			err = CreateScheduleBackup(scheduleName, appClusterName, backupLocationName, backupLocationUID, namespaces,
				labelSelectors, orgID, "", "", "", "", periodicSchedulePolicyName, periodicSchedulePolicyUid, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of schedule backup with schedule name [%s]", scheduleName))
			firstScheduleBackupName, err := GetFirstScheduleBackupName(ctx, scheduleName, orgID)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching the name of the first schedule backup [%s]", firstScheduleBackupName))
		})
		// To ensure applications are deployed after a schedule backup is created
		Step("Schedule applications to create new namespaces", func() {
			log.InfoD("Scheduling applications to create new namespaces")
			contexts = make([]*scheduler.Context, 0)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
				appContexts := ScheduleApplications(taskName)
				contexts = append(contexts, appContexts...)
				for _, ctx := range appContexts {
					ctx.ReadinessTimeout = appReadinessTimeout
					namespace := GetAppNamespace(ctx, taskName)
					log.InfoD("Scheduled application with namespace [%s]", namespace)
					// appNamespaces in this scenario is of newly created namespaces
					appNamespaces = append(appNamespaces, namespace)
				}
			}
		})
		Step("Validate new namespaces", func() {
			log.InfoD("Validating new namespaces")
			ValidateApplications(contexts)
		})
		Step("Verify new application namespaces inclusion in next schedule backup", func() {
			log.InfoD("Verifying new application namespaces inclusion in next schedule backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			dash.VerifyFatal(err, nil, "Fetching px-central-admin ctx")
			allScheduleBackupNames, err := Inst().Backup.GetAllScheduleBackupNames(ctx, scheduleName, orgID)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching names of all schedule backups of schedule named [%s]", scheduleName))
			currentScheduleBackupCount := len(allScheduleBackupNames)
			log.InfoD("Current number of schedule backups is [%v]", currentScheduleBackupCount)
			nextScheduleBackupOrdinal := currentScheduleBackupCount + 1
			log.InfoD("Ordinal of the next schedule backup is [%v]", nextScheduleBackupOrdinal)
			checkOrdinalScheduleBackupCreation := func() (interface{}, bool, error) {
				ordinalScheduleBackupName, err := GetOrdinalScheduleBackupName(ctx, scheduleName, nextScheduleBackupOrdinal, orgID)
				if err != nil {
					return "", true, err
				}
				return ordinalScheduleBackupName, false, nil
			}
			log.InfoD("Waiting for 15 minutes for the next schedule backup to be triggered")
			time.Sleep(15 * time.Minute)
			nextScheduleBackupName, err = task.DoRetryWithTimeout(checkOrdinalScheduleBackupCreation, maxWaitPeriodForBackupCompletionInMinutes*time.Minute, 30*time.Second)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching next schedule backup name of ordinal [%v] of schedule named [%s]", nextScheduleBackupOrdinal, scheduleName))
			log.InfoD("Next schedule backup name [%s]", nextScheduleBackupName.(string))
			_, err = backupSuccessCheck(nextScheduleBackupName.(string), orgID, 0, 0, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying success of next schedule backup named [%s] of schedule named [%s]", nextScheduleBackupName.(string), scheduleName))
			nextScheduleBackupUid, err := Inst().Backup.GetBackupUID(ctx, nextScheduleBackupName.(string), orgID)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching uid of of next schedule backup named [%s] of schedule named [%s]", nextScheduleBackupName, scheduleName))
			backupInspectRequest := &api.BackupInspectRequest{
				Name:  nextScheduleBackupName.(string),
				Uid:   nextScheduleBackupUid,
				OrgId: orgID,
			}
			resp, err := Inst().Backup.InspectBackup(ctx, backupInspectRequest)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Inspecting next schedule backup named [%s] and uid [%s]", nextScheduleBackupName, nextScheduleBackupUid))
			backedUpNamespaces := resp.GetBackup().GetNamespaces()
			log.InfoD("Namespaces in next schedule backup named [%s] and uid [%s] are [%v]", nextScheduleBackupName, nextScheduleBackupUid, backedUpNamespaces)
			for _, namespace := range appNamespaces {
				dash.VerifyFatal(strings.Contains(strings.Join(backedUpNamespaces, ","), namespace), true, fmt.Sprintf("Checking the new application namespace [%s] against the next scheduled backup named [%s]", namespace, nextScheduleBackupName))
			}
		})
		Step("Restore new application namespaces from next schedule backup", func() {
			log.InfoD("Restoring new application namespaces from next schedule backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			dash.VerifyFatal(err, nil, "Fetching px-central-admin ctx")
			namespaceMapping := make(map[string]string)
			// Modifying namespaceMapping to restore only new namespaces
			for _, namespace := range appNamespaces {
				namespaceMapping[namespace] = namespace
			}
			log.InfoD("Namespace mapping used for restoring - %v", namespaceMapping)
			restoreName = fmt.Sprintf("%s-%s", "test-restore", RandomString(4))
			err = CreateRestore(restoreName, nextScheduleBackupName.(string), namespaceMapping, appClusterName, orgID, ctx, make(map[string]string))
			dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s]", restoreName))
		})
	})

	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(contexts)
		ctx, err := backup.GetAdminCtxFromSecret()
		dash.VerifySafely(err, nil, "Fetching px-central-admin ctx")
		scheduleUid, err := GetScheduleUID(scheduleName, orgID, ctx)
		dash.VerifySafely(err, nil, fmt.Sprintf("Fetching uid of schedule named [%s]", scheduleName))
		allScheduleBackupNames, err := Inst().Backup.GetAllScheduleBackupNames(ctx, scheduleName, orgID)
		dash.VerifySafely(err, nil, fmt.Sprintf("Fetching all schedule backup names of schedule named [%s]", scheduleName))
		log.InfoD("Deleting schedule named [%s] along with its backups [%v] and schedule policies [%v]", scheduleName, allScheduleBackupNames, []string{periodicSchedulePolicyName})
		err = DeleteSchedule(scheduleName, scheduleUid, periodicSchedulePolicyName, periodicSchedulePolicyUid, orgID)
		dash.VerifySafely(err, nil, fmt.Sprintf("Verifying deletion of backup schedule named [%s]", scheduleName))
		err = DeleteRestore(restoreName, orgID, ctx)
		dash.VerifySafely(err, nil, fmt.Sprintf("Deleting restore [%s]", restoreName))
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		log.InfoD("Deleting deployed namespaces - %v", appNamespaces)
		ValidateAndDestroy(contexts, opts)
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})

// BackupSyncBasicTest take a good number of backups check if backup sync is working
var _ = Describe("{BackupSyncBasicTest}", func() {
	numberOfBackups, _ := strconv.Atoi(getEnv(maxBackupsToBeCreated, "10"))
	timeBetweenConsecutiveBackups := 4 * time.Second
	backupNames := make([]string, 0)
	numberOfSimultaneousBackups := 20
	var contexts []*scheduler.Context
	labelSelectors := make(map[string]string)
	var backupLocationUID string
	var cloudCredUID string
	var backupName string
	var cloudCredUidList []string
	var appContexts []*scheduler.Context
	var bkpNamespaces []string
	var clusterUid string
	var clusterStatus api.ClusterInfo_StatusInfo_Status
	var customBackupLocationName string
	var credName string
	bkpNamespaces = make([]string, 0)
	backupNamespaceMap := make(map[string]string)
	backupLocationMap := make(map[string]string)

	JustBeforeEach(func() {
		StartTorpedoTest("BackupSyncBasicTest",
			"Validate that the backup sync syncs all the backups present in bucket", nil, 58040)
		log.InfoD("Deploy applications")
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
			appContexts = ScheduleApplications(taskName)
			contexts = append(contexts, appContexts...)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = appReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
			}
		}
	})
	It("Validate that the backup sync syncs all the backups present in bucket", func() {
		providers := getProviders()
		Step("Validate applications and get their labels", func() {
			log.InfoD("Validate applications")
			ValidateApplications(contexts)
		})

		Step("Adding Credentials and Registering Backup Location", func() {
			log.InfoD("Using pre-provisioned bucket. Creating cloud credentials and backup location.")
			for _, provider := range providers {
				cloudCredUID = uuid.New()
				cloudCredUidList = append(cloudCredUidList, cloudCredUID)
				backupLocationUID = uuid.New()
				credName = fmt.Sprintf("autogenerated-cred-%v", time.Now().Unix())
				CreateCloudCredential(provider, credName, cloudCredUID, orgID)
				log.InfoD("Created Cloud Credentials with name - %s", credName)
				customBackupLocationName = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
				err := CreateBackupLocation(provider, customBackupLocationName, backupLocationUID, credName, cloudCredUID, getGlobalBucketName(provider), orgID, "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", customBackupLocationName))
				log.InfoD("Created Backup Location with name - %s", customBackupLocationName)
			}
		})

		Step("Register source and destination cluster for backup", func() {
			log.InfoD("Registering Source and Destination clusters and verifying the status")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = CreateSourceAndDestClusters(orgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, clusterUid = Inst().Backup.RegisterBackupCluster(orgID, SourceClusterName, "")
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying backup cluster %s status", SourceClusterName))
		})

		Step("Taking backup of applications", func() {
			log.InfoD("Taking backup of applications")
			var sem = make(chan struct{}, numberOfSimultaneousBackups)
			var wg sync.WaitGroup
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			log.InfoD("Taking %d backups", numberOfBackups)
			for _, namespace := range bkpNamespaces {
				for i := 0; i < numberOfBackups; i++ {
					sem <- struct{}{}
					time.Sleep(timeBetweenConsecutiveBackups)
					backupName := fmt.Sprintf("%s-%v", BackupNamePrefix, time.Now().Unix())
					backupNames = append(backupNames, backupName)
					wg.Add(1)
					go func(backupName string) {
						defer GinkgoRecover()
						defer wg.Done()
						defer func() { <-sem }()
						err = CreateBackup(backupName, SourceClusterName, customBackupLocationName, backupLocationUID, []string{namespace},
							labelSelectors, orgID, clusterUid, "", "", "", "", ctx)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying backup creation: %s", backupName))
					}(backupName)
				}
				wg.Wait()
			}
			log.Infof("List of backups - %v", backupNames)
		})

		Step("Remove the backup location where backups were taken", func() {
			log.InfoD("Remove backup location where backups were taken")
			// Issue a remove backup location call
			err := DeleteBackupLocation(customBackupLocationName, backupLocationUID, orgID, false)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting backup location %s", customBackupLocationName))

			// Wait until backup location is removed
			backupLocationDeleteStatusCheck := func() (interface{}, bool, error) {
				ctx, err := backup.GetAdminCtxFromSecret()
				status, err := IsBackupLocationPresent(customBackupLocationName, ctx, orgID)
				if err != nil {
					return "", true, fmt.Errorf("backup location %s still present with error %v", customBackupLocationName, err)
				}
				if status == true {
					return "", true, fmt.Errorf("backup location %s is not deleted yet", customBackupLocationName)
				}
				return "", false, nil
			}
			_, err = task.DoRetryWithTimeout(backupLocationDeleteStatusCheck, 3*time.Minute, 30*time.Second)
			dash.VerifySafely(err, nil, fmt.Sprintf("Deleting backup location %s", customBackupLocationName))
		})

		Step("Add the backup location again which had backups", func() {
			log.InfoD("Add the backup location with backups back")
			for _, provider := range providers {
				customBackupLocationName = fmt.Sprintf("autogenerated-backup-location-%v", time.Now().Unix())
				backupLocationMap[backupLocationUID] = customBackupLocationName
				err := CreateBackupLocation(provider, customBackupLocationName, backupLocationUID, credName, cloudCredUID, getGlobalBucketName(provider), orgID, "")
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating backup location %s", customBackupLocationName))
			}
		})
		Step("Taking backup of applications to trigger BackupSync goroutine", func() {
			log.InfoD("Taking backup of applications to trigger BackupSync goroutine")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx failed")
			for _, namespace := range bkpNamespaces {
				backupName = fmt.Sprintf("%s-%s-%v", BackupNamePrefix, namespace, time.Now().Unix())
				backupNamespaceMap[namespace] = backupName
				err = CreateBackup(backupName, SourceClusterName, customBackupLocationName, backupLocationUID, []string{namespace}, labelSelectors, orgID, clusterUid, "", "", "", "", ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying backup creation %s", backupName))
				backupNames = append(backupNames, backupName)
			}
		})

		Step("Check if all backups are synced or not", func() {
			log.InfoD("Check if backups created before are synced or not")

			// Wait for backups to get synced
			checkBackupSync := func() (interface{}, bool, error) {
				fetchedBackupNames, err := GetAllBackupsAdmin()
				// Debug lines tobe removed in the next patch with the fix
				log.InfoD(fmt.Sprintf("The list of backups fetched %s", fetchedBackupNames))
				if err != nil {
					return "", true, fmt.Errorf("unable to fetch backups. Error: %s", err.Error())
				}
				if len(fetchedBackupNames) == len(backupNames) {
					return "", false, nil
				}
				return "", true, fmt.Errorf("expected: %d and actual: %d", len(backupNames), len(fetchedBackupNames))
			}
			_, err := task.DoRetryWithTimeout(checkBackupSync, 100*time.Minute, 30*time.Second)
			log.FailOnError(err, "Wait for BackupSync to complete")
			fetchedBackupNames, err := GetAllBackupsAdmin()
			log.FailOnError(err, "Getting a list of all backups")
			dash.VerifyFatal(len(fetchedBackupNames), len(backupNames), "Comparing the expected and actual number of backups")
			var bkp *api.BackupObject
			backupDriver := Inst().Backup
			bkpEnumerateReq := &api.BackupEnumerateRequest{
				OrgId: orgID}
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx failed")
			curBackups, err := backupDriver.EnumerateBackup(ctx, bkpEnumerateReq)
			for _, bkp = range curBackups.GetBackups() {
				backupInspectRequest := &api.BackupInspectRequest{
					Name:  bkp.Name,
					Uid:   bkp.Uid,
					OrgId: orgID,
				}
				resp, err := backupDriver.InspectBackup(ctx, backupInspectRequest)
				log.FailOnError(err, "Inspect each backup from list")
				actual := resp.GetBackup().GetStatus().Status
				expected := api.BackupInfo_StatusInfo_Success
				dash.VerifyFatal(actual, expected, fmt.Sprintf("Check each backup for success status %s", bkp.Name))
			}
		})
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(contexts)
		log.InfoD("Deleting the deployed apps after the testcase")
		for i := 0; i < len(contexts); i++ {
			opts := make(map[string]bool)
			opts[SkipClusterScopedObjects] = true
			taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
			err := Inst().S.Destroy(contexts[i], opts)
			dash.VerifySafely(err, nil, fmt.Sprintf("Verify destroying app %s, Err: %v", taskName, err))
		}
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")

		log.Infof("Deleting registered clusters for admin context")
		CleanupCloudSettingsAndClusters(backupLocationMap, credName, cloudCredUID, ctx)
	})
})

var _ = Describe("{BackupRestoreCRsOnDifferentK8sVersions}", func() {

	var (
		backupNames                    []string                // backups in px-backup
		restoreNames                   []string                // restores in px-backup
		restoreLaterNames              []string                // restore-laters in px-backup
		sourceClusterAppsContexts      []*scheduler.Context    // Each Context is for one Namespace which corresponds to one App
		destinationClusterAppsContexts []*scheduler.Context    // Each Context is for one Namespace which corresponds to one App
		backupContexts                 []*BackupRestoreContext // Each Context is for one backup in px-backup
		restoreContexts                []*BackupRestoreContext // Each Context is for one restore in px-backup
		restoreLaterContexts           []*BackupRestoreContext // Each Context is for one restore-later in px-backup
		preRuleNameList                []string
		postRuleNameList               []string
		clusterUid                     string
		cloudCredName                  string
		cloudCredUID                   string
		backupLocationUID              string
		backupLocationName             string
	)

	var (
		appList               = Inst().AppList
		sourceNamespaces      = make([]string, 0)
		destinationNamespaces = make([]string, 0)
		namespaceMapping      = make(map[string]string)
		backupLocationMap     = make(map[string]string)
		labelSelectors        = make(map[string]string)
	)

	providers := getProviders()

	JustBeforeEach(func() {

		StartTorpedoTest("BackupRestoreCRsOnDifferentK8sVersions", "Deploy CRs (CRD + webhook); Backup; two simulatanous Restores with one Success and other PartialSuccess. (Backup and Restore on different K8s version)", nil, 83716)

		log.InfoD("verifying if the pre/post rules for the required apps are present in the AppParameters or not")
		for i := 0; i < len(appList); i++ {
			if Contains(postRuleApp, appList[i]) {
				if _, ok := portworx.AppParameters[appList[i]]["post"]; ok {
					dash.VerifyFatal(ok, true, "post rule details mentioned for the apps")
				}
			}
			if Contains(preRuleApp, appList[i]) {
				if _, ok := portworx.AppParameters[appList[i]]["pre"]; ok {
					dash.VerifyFatal(ok, true, "pre rule details mentioned for the apps")
				}
			}
		}

	})

	It("Deploy CRs (CRD + webhook); Backup; two simulatanous Restores with one Success and other PartialSuccess. (Backup and Restore on different K8s version)", func() {

		defer func() {
			log.InfoD("switching to default context")
			err1 := SetClusterContext("")
			log.FailOnError(err1, "failed to SetClusterContext to default cluster")
		}()

		Step("Verify if app used to execute test is a valid/allowed spec (apps) for *this* test", func() {
			log.InfoD("specs (apps) allowed in execution of test: %v", appsWithCRDsAndWebhooks)
			for i := 0; i < len(appList); i++ {
				contains := Contains(appsWithCRDsAndWebhooks, appList[i])
				dash.VerifyFatal(contains, true,
					fmt.Sprintf("app [%s] allowed in execution of this test", appList[i]))
			}
		})

		Step("verify kubernetes version of source and destination cluster", func() {
			var srcVer, destVer semver.Version
			log.InfoD("begin verification kubernetes version of source and destination cluster")

			Step("register cluster for backup", func() {
				log.InfoD("register cluster for backup")
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "fetching px-central-admin ctx")
				err = CreateSourceAndDestClusters(orgID, "", "", ctx)
				dash.VerifyFatal(err, nil, "creating source and destination cluster")
				_, clusterUid = Inst().Backup.RegisterBackupCluster(orgID, SourceClusterName, "")
			})

			Step("Get kubernetes source cluster version", func() {
				log.InfoD("switched context to source")

				sourceClusterConfigPath, err := GetSourceClusterConfigPath()
				log.FailOnError(err, "failed to get kubeconfig path for source cluster. Error: [%v]", err)

				err = Inst().S.SetConfig(sourceClusterConfigPath)
				log.FailOnError(err, "failed to switch to context to source cluster [%v]", sourceClusterConfigPath)

				ver, err := k8s.ClusterVersion()
				log.FailOnError(err, "failed to get source cluster version")
				srcVer, err = semver.Make(ver)
				log.FailOnError(err, "failed to get source cluster version")
			})

			Step("Get kubernetes destination cluster version", func() {
				log.InfoD("switched context to destination")

				destinationClusterConfigPath, err := GetDestinationClusterConfigPath()
				log.FailOnError(err, "failed to get kubeconfig path for destination cluster. Error: [%v]", err)

				err = Inst().S.SetConfig(destinationClusterConfigPath)
				log.FailOnError(err, "failed to switch to context to destination cluster [%v]", destinationClusterConfigPath)

				ver, err := k8s.ClusterVersion()
				log.FailOnError(err, "failed to get destination cluster version")
				destVer, err = semver.Make(ver)
				log.FailOnError(err, "failed to get destination cluster version")
			})

			Step("Compare Source and Destination cluster version numbers", func() {
				log.InfoD("source cluster version: %s ; destination cluster version: %s", srcVer.String(), destVer.String())
				isValid := srcVer.LT(destVer)
				dash.VerifyFatal(isValid, true,
					"source cluster kubernetes version should be lesser than the destination cluster kubernetes version.")
			})

			log.InfoD("switching to default context")
			err := SetClusterContext("")
			log.FailOnError(err, "failed to SetClusterContext to default cluster")
		})

		Step("deploy the applications on Src cluster", func() {
			log.InfoD("deploy the applications on Src cluster")

			Step("deploy applications", func() {
				log.InfoD("deploy applications")

				log.InfoD("switching to source context")
				err := SetSourceKubeConfig()
				log.FailOnError(err, "failed to switch to context to source cluster")

				log.InfoD("ccheduling applications")
				sourceClusterAppsContexts = make([]*scheduler.Context, 0)
				for i := 0; i < Inst().GlobalScaleFactor; i++ {
					taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
					appContexts := ScheduleApplications(taskName)
					for _, appCtx := range appContexts {
						appCtx.ReadinessTimeout = appReadinessTimeout
						namespace := GetAppNamespace(appCtx, taskName)
						// (sourceNamespaces, sourceClusterAppsContexts) will always correspoond
						sourceNamespaces = append(sourceNamespaces, namespace)
						sourceClusterAppsContexts = append(sourceClusterAppsContexts, appCtx)
					}
				}
			})

			Step("Validate applications", func() {
				ValidateApplications(sourceClusterAppsContexts)

				log.InfoD("switching to default context")
				err := SetClusterContext("")
				log.FailOnError(err, "failed to SetClusterContext to default cluster")
			})

			log.InfoD("waiting (for 2 minutes) for any CRs to finish starting up.")
			time.Sleep(time.Minute * 2)
			log.Warnf("no verification is done; it might lead to undetectable errors.")
		})

		Step("Creating rules for backup", func() {
			log.InfoD("creating pre rule for deployed apps")
			for i := 0; i < len(appList); i++ {
				preRuleStatus, ruleName, err := Inst().Backup.CreateRuleForBackup(appList[i], orgID, "pre")
				log.FailOnError(err, "creating pre rule for deployed apps failed")
				dash.VerifyFatal(preRuleStatus, true, "verifying pre rule for backup")

				if ruleName != "" {
					preRuleNameList = append(preRuleNameList, ruleName)
				}
			}
			log.InfoD("Creating post rule for deployed apps")
			for i := 0; i < len(appList); i++ {
				postRuleStatus, ruleName, err := Inst().Backup.CreateRuleForBackup(appList[i], orgID, "post")
				log.FailOnError(err, "creating post rule for deployed apps failed")
				dash.VerifyFatal(postRuleStatus, true, "verifying Post rule for backup")
				if ruleName != "" {
					postRuleNameList = append(postRuleNameList, ruleName)
				}
			}
		})

		Step("Creating bucket, backup location and cloud credentials", func() {
			log.InfoD("Creating bucket, backup location and cloud credentials")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
				backupLocationName = fmt.Sprintf("%s-%s-bl", provider, getGlobalBucketName(provider))
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocationName
				CreateCloudCredential(provider, cloudCredName, cloudCredUID, orgID)
				log.InfoD("creating backup location [%s] with cloud cred [%s]", backupLocationName, cloudCredName)
				err := CreateBackupLocation(provider, backupLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), orgID, "")
				dash.VerifyFatal(err, nil, "creating backup location")
			}
		})

		Step("Taking backup of application from source cluster", func() {
			log.InfoD("taking backup of applications")
			ctx, err := backup.GetAdminCtxFromSecret()
			dash.VerifyFatal(err, nil, "getting context")
			backupNames = make([]string, len(sourceNamespaces))
			backupContexts = make([]*BackupRestoreContext, len(sourceNamespaces))
			for i, namespace := range sourceNamespaces {
				backupName := fmt.Sprintf("%s-%s-%v", BackupNamePrefix, namespace, time.Now().Unix())
				log.InfoD("creating backup [%s] in source cluster [%s] (%s), organization [%s], of namespace [%s], in backup location [%s]", backupName, SourceClusterName, clusterUid, orgID, namespace, backupLocationName)
				backupCtx, err := CreateBackupAndGetBackupCtx(backupName, SourceClusterName, backupLocationName, backupLocationUID, []string{namespace}, labelSelectors, orgID, clusterUid, "", "", "", "", ctx, []*scheduler.Context{sourceClusterAppsContexts[i]})

				dash.VerifyFatal(err, nil, "verifying backup creation")
				backupNames[i] = backupName
				backupContexts[i] = backupCtx
			}
		})

		Step("Restoring the backed up applications on destination cluster", func() {

			log.InfoD("Restoring the backed up applications on destination cluster")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			for i, sourceNamespace := range sourceNamespaces {
				var initialRestoreName, laterRestoreName string

				Step("Restoring the backed up application to namespace of same name on destination cluster", func() {
					log.InfoD("restoring the backed up application to namespace of same name on destination cluster")

					initialRestoreName = fmt.Sprintf("%s-%s-initial-%v", restoreNamePrefix, sourceNamespace, time.Now().Unix())
					restoreNames = append(restoreNames, initialRestoreName)
					destinationNameSpace := sourceNamespace
					destinationNamespaces = append(destinationNamespaces, destinationNameSpace)
					namespaceMapping[sourceNamespace] = destinationNameSpace

					log.InfoD("creating initial-restore [%s] in destination cluster [%s], organization [%s], in namespace [%s]", initialRestoreName, destinationClusterName, orgID, destinationNameSpace)
					_, err = CreateRestoreWithoutCheck(initialRestoreName, backupNames[i], namespaceMapping, destinationClusterName, orgID, ctx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Initiation of restore %s", initialRestoreName))

					restoreInspectRequest := &api.RestoreInspectRequest{
						Name:  initialRestoreName,
						OrgId: orgID,
					}
					restoreInProgressCheck := func() (interface{}, bool, error) {
						resp, err := Inst().Backup.InspectRestore(ctx, restoreInspectRequest)
						restoreResponseStatus := resp.GetRestore().GetStatus()
						if err != nil {
							err := fmt.Errorf("failed getting restore status for - %s; Err: %s", initialRestoreName, err)
							return "", false, err
						}

						// Status should be LATER than InProgress in order for next STEP to execute
						if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_InProgress {
							log.InfoD("restore status of [%s] is [%s]; expected [InProgress].\ncondition fulfilled.", initialRestoreName, restoreResponseStatus.GetStatus())
							return "", false, nil
						} else if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_PartialSuccess {
							err := fmt.Errorf("restore status of [%s] is [%s]; expected [InProgress].\nhelp: check for remnant cluster-level resources on destination cluster.", initialRestoreName, restoreResponseStatus.GetStatus())
							return "", false, err
						} else if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Success {
							err := fmt.Errorf("restore status of [%s] is [%s]; expected [InProgress].\nhelp: check for status frequently", initialRestoreName, restoreResponseStatus.GetStatus())
							return "", false, err
						} else if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Aborted ||
							restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Failed ||
							restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Deleting {
							err := fmt.Errorf("restore status of [%s] is [%s]; expected [InProgress].", initialRestoreName, restoreResponseStatus.GetStatus())
							return "", false, err
						}

						err = fmt.Errorf("restore status of [%s] is [%s]; waiting for [InProgress]...", initialRestoreName, restoreResponseStatus.GetStatus())
						return "", true, err
					}
					_, err = task.DoRetryWithTimeout(restoreInProgressCheck, 10*time.Minute, 5*time.Second)
					dash.VerifyFatal(err, nil, fmt.Sprintf("restore %s is [InProgress]", initialRestoreName))
				})

				var restoreLaterStatuserr error
				var laterRestoreStatus interface{}

				Step("Restoring the backed up application to namespace with different name on destination cluster", func() {
					log.InfoD("Restoring the backed up application to namespace with different name on destination cluster")

					laterRestoreName = fmt.Sprintf("%s-%s-later-%v", restoreNamePrefix, sourceNamespace, time.Now().Unix())
					restoreLaterNames = append(restoreLaterNames, laterRestoreName)
					destinationNameSpace := fmt.Sprintf("%s-%s", sourceNamespace, "later")
					destinationNamespaces = append(destinationNamespaces, destinationNameSpace)
					namespaceMapping := make(map[string]string) //using local version in order to not change mapping as the key is the same
					namespaceMapping[sourceNamespace] = destinationNameSpace

					log.InfoD("creating later-restore [%s] in destination cluster [%s], organization [%s], in namespace [%s]", laterRestoreName, destinationClusterName, orgID, destinationNameSpace)
					_, err = CreateRestoreWithoutCheck(laterRestoreName, backupNames[i], namespaceMapping, destinationClusterName, orgID, ctx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Initiation of restore %s", laterRestoreName))

					restoreInspectRequest := &api.RestoreInspectRequest{
						Name:  laterRestoreName,
						OrgId: orgID,
					}
					restorePartialSuccessCheck := func() (interface{}, bool, error) {
						resp, err := Inst().Backup.InspectRestore(ctx, restoreInspectRequest)
						restoreResponseStatus := resp.GetRestore().GetStatus()
						if err != nil {
							err := fmt.Errorf("failed getting restore status for - %s; Err: %s", laterRestoreName, err)
							return "", false, err
						}

						if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_PartialSuccess || restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Success {
							log.InfoD("restore status of [%s] is [%s]; expected [PartialSuccess] or [Success].\ncondition fulfilled.", laterRestoreName, restoreResponseStatus.GetStatus())
							return restoreResponseStatus.GetStatus(), false, nil
						} else if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Aborted ||
							restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Failed ||
							restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Deleting {
							err := fmt.Errorf("restore status of [%s] is [%s]; expected [PartialSuccess] or [Success].", laterRestoreName, restoreResponseStatus.GetStatus())
							return restoreResponseStatus.GetStatus(), false, err
						}

						err = fmt.Errorf("restore status of [%s] is [%s]; waiting for [PartialSuccess] or [Success]...", laterRestoreName, restoreResponseStatus.GetStatus())
						return "", true, err
					}
					laterRestoreStatus, restoreLaterStatuserr = task.DoRetryWithTimeout(restorePartialSuccessCheck, 10*time.Minute, 30*time.Second)

					// we don't end the test if there is an error here, as we also want to ensure that we look into the status of the following `Step`, so that we have the full details of what went wrong.
					dash.VerifySafely(restoreLaterStatuserr, nil, fmt.Sprintf("status of later restore [%s] is [PartialSuccess] or [Success]", laterRestoreName))

					// We can consider validation and cleanup for [PartialSuccess] and [Success]
					if restoreLaterStatuserr == nil {
						// Validation of Later Restore
						destinationClusterConfigPath, err := GetDestinationClusterConfigPath()
						log.FailOnError(err, "failed to get kubeconfig path for destination cluster. Error: [%v]", err)

						restoreLaterCtx, err := ValidateRestore(laterRestoreName, destinationClusterConfigPath, orgID, ctx, backupContexts[i], namespaceMapping)
						dash.VerifyFatal(err, nil, fmt.Sprintf("validation of restore [%s] is success", laterRestoreName))
						restoreLaterContexts = append(restoreLaterContexts, restoreLaterCtx)
					} else {
						log.Warnf("proceeding to next step, after which the test will be failed.")
					}
				})

				Step("Verifying status of Initial and Later Restore", func() {
					log.InfoD("Step: Verifying status of Initial and Later Restore")

					// getting the status of initial restore
					restoreInspectRequest := &api.RestoreInspectRequest{
						Name:  initialRestoreName,
						OrgId: orgID,
					}
					restoreSuccessCheck := func() (interface{}, bool, error) {
						resp, err := Inst().Backup.InspectRestore(ctx, restoreInspectRequest)
						restoreResponseStatus := resp.GetRestore().GetStatus()
						if err != nil {
							err := fmt.Errorf("failed getting restore status for - %s; Err: %s", initialRestoreName, err)
							return "", false, err
						}

						if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Success || restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_PartialSuccess {
							log.InfoD("restore status of [%s] is [%s]; expected [PartialSuccess] or [Success].\ncondition fulfilled.", initialRestoreName, restoreResponseStatus.GetStatus())
							return restoreResponseStatus.GetStatus(), false, nil
						} else if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Aborted ||
							restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Failed ||
							restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Deleting {
							err := fmt.Errorf("restore status of [%s] is [%s]; expected [PartialSuccess] or [Success].", initialRestoreName, restoreResponseStatus.GetStatus())
							return restoreResponseStatus.GetStatus(), false, err
						}

						err = fmt.Errorf("restore status of [%s] is [%s]; waiting for [PartialSuccess] or [Success]...", initialRestoreName, restoreResponseStatus.GetStatus())
						return "", true, err
					}
					initialRestoreStatus, initialRestoreError := task.DoRetryWithTimeout(restoreSuccessCheck, 10*time.Minute, 30*time.Second)

					dash.VerifyFatal(initialRestoreError, nil, fmt.Sprintf("status of initial restore [%s] is [PartialSuccess] or [Success]", initialRestoreName))

					// Validation of Inital Restore
					destinationClusterConfigPath, err := GetDestinationClusterConfigPath()
					log.FailOnError(err, "failed to get kubeconfig path for destination cluster. Error: [%v]", err)

					restoreCtx, err := ValidateRestore(initialRestoreName, destinationClusterConfigPath, orgID, ctx, backupContexts[i], namespaceMapping)
					dash.VerifyFatal(err, nil, fmt.Sprintf("validation of restore [%s] is success", initialRestoreName))
					restoreContexts = append(restoreContexts, restoreCtx)

					// If Later Restore was an error before, we have to fail the test at this point, having processed the other stage
					dash.VerifyFatal(restoreLaterStatuserr, nil, fmt.Sprintf("status of later restore [%s] is [PartialSuccess] or [Success]", laterRestoreName))

					// Checking actual validity of restore status
					validity := false
					errHelpStr := ""
					log.InfoD("states of (initial,later) restore are [%s,%s]", initialRestoreStatus, laterRestoreStatus)
					if (initialRestoreStatus == api.RestoreInfo_StatusInfo_Success && laterRestoreStatus == api.RestoreInfo_StatusInfo_PartialSuccess) ||
						(initialRestoreStatus == api.RestoreInfo_StatusInfo_PartialSuccess && laterRestoreStatus == api.RestoreInfo_StatusInfo_Success) {
						validity = true
					} else if initialRestoreStatus == api.RestoreInfo_StatusInfo_PartialSuccess && laterRestoreStatus == api.RestoreInfo_StatusInfo_PartialSuccess {
						validity = false
						errHelpStr = "error help: ensure no remnant cluster-level resources on destination cluster."
					} else if initialRestoreStatus == api.RestoreInfo_StatusInfo_Success && laterRestoreStatus == api.RestoreInfo_StatusInfo_Success {
						validity = false
						errHelpStr = "error help: ensure app has cluster-level resources."
					}
					dash.VerifyFatal(validity, true, fmt.Sprintf("states of (initial,later) restore are [Success,PartialSuccess] or [PartialSuccess,Success]. %s", errHelpStr))
				})

			}
		})

	})

	JustAfterEach(func() {

		defer EndTorpedoTest()

		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "fetching px-central-admin ctx")

		// TODO: move this to AfterSuite
		if len(preRuleNameList) > 0 {
			for _, ruleName := range preRuleNameList {
				err := Inst().Backup.DeleteRuleForBackup(orgID, ruleName)
				dash.VerifySafely(err, nil, fmt.Sprintf("deleting backup pre rules %s", ruleName))
			}
		}

		// TODO: move this to AfterSuite
		if len(postRuleNameList) > 0 {
			for _, ruleName := range postRuleNameList {
				err := Inst().Backup.DeleteRuleForBackup(orgID, ruleName)
				dash.VerifySafely(err, nil, fmt.Sprintf("deleting backup post rules %s", ruleName))
			}
		}

		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = false
		log.InfoD("deleting deployed applications for source and destination clusters")

		log.InfoD("switching to source context")
		err = SetSourceKubeConfig()
		log.FailOnError(err, "failed to switch to context to source cluster")

		log.InfoD("deleting deployed applications on source clusters")
		ValidateAndDestroy(sourceClusterAppsContexts, opts)

		log.InfoD("waiting (for 1 minute) for any Resources created by Operator of Custom Resources to finish being destroyed.")
		time.Sleep(time.Minute * 1)
		log.Warn("no verification of destruction is done; it might lead to undetectable errors.")

		log.InfoD("switching to destination context")
		err = SetDestinationKubeConfig()
		log.FailOnError(err, "failed to switch to context to destination cluster")

		destinationClusterAppsContexts = make([]*scheduler.Context, 0)
		// only adding restoreContexts, not restoreLaterContexts
		for _, restoreCtx := range restoreContexts {
			destinationClusterAppsContexts = append(destinationClusterAppsContexts, restoreCtx.schedulerCtxs...)
		}
		log.InfoD("deleting deployed applications (initial restore) on destination clusters")
		ValidateAndDestroy(destinationClusterAppsContexts, opts)

		//TODO: delete restore-later apps
		log.Warn("not deleting deployed applications (restore-later) on destination clusters")

		log.InfoD("waiting (for 1 minute) for any Resources created by Operator of Custom Resources to finish being destroyed.")
		time.Sleep(time.Minute * 1)
		log.Warn("no verification of destruction is done; it might lead to undetectable errors.")

		log.InfoD("switching to default context")
		err = SetClusterContext("")
		log.FailOnError(err, "failed to SetClusterContext to default cluster")

		backupDriver := Inst().Backup

		log.InfoD("deleting backed up namespaces")
		for _, backupName := range backupNames {
			backupUID, err := backupDriver.GetBackupUID(ctx, backupName, orgID)
			log.FailOnError(err, "failed while trying to get backup UID for - %s", backupName)
			backupDeleteResponse, err := DeleteBackup(backupName, backupUID, orgID, ctx)
			log.FailOnError(err, "backup [%s] could not be deleted", backupName)
			dash.VerifyFatal(backupDeleteResponse.String(), "", fmt.Sprintf("verifying [%s] backup deletion is successful", backupName))
		}

		log.InfoD("deleting restores")
		for _, restoreName := range restoreNames {
			err = DeleteRestore(restoreName, orgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting Restore [%s]", restoreName))
		}

		log.InfoD("deleting restore-laters")
		for _, restoreLaterName := range restoreLaterNames {
			err = DeleteRestore(restoreLaterName, orgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("deleting Restore [%s]", restoreLaterName))
		}

		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})
