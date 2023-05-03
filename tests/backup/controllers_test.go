package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/torpedo/drivers/backup/portworx"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	"github.com/portworx/torpedo/tests/backup/controllers/pxbackup"
	"github.com/portworx/torpedo/tests/backup/utils"
	"strings"
	"time"
)

// This testcase verifies basic backup rule,backup location, cloud setting
var _ = Describe("{PxbCBasicBackupCreation}", func() {
	var (
		controllers     map[string]*pxbackup.PxbController
		appList         = Inst().AppList
		contexts        []*scheduler.Context
		appContexts     []*scheduler.Context
		bkpNamespaces   []string
		clusterStatus   api.ClusterInfo_StatusInfo_Status
		bkpLocationName string
		providers       []string
		backupName      string
		backupNames     []string
	)

	JustBeforeEach(func() {
		controllers = make(map[string]*pxbackup.PxbController, 0)
		bkpNamespaces = make([]string, 0)
		providers = getProviders()
		StartTorpedoTest("Backup: BasicBackupCreation", "Deploying backup", nil, 0)
		log.InfoD("Verifying if the pre/post rules for the required apps are present in the AppParameters or not")
		for i := 0; i < len(appList); i++ {
			if Contains(preRuleApp, appList[i]) {
				if _, ok := portworx.AppParameters[appList[i]]["pre"]; ok {
					dash.VerifyFatal(ok, true, fmt.Sprintf("Pre Rule details mentioned for the app [%s]", appList[i]))
				}
			}
			if Contains(postRuleApp, appList[i]) {
				if _, ok := portworx.AppParameters[appList[i]]["post"]; ok {
					dash.VerifyFatal(ok, true, fmt.Sprintf("Post Rule details mentioned for the app [%s]", appList[i]))
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
	It("Basic Backup Creation Pxb Controller", func() {
		Step("Validating applications", func() {
			log.InfoD("Validating applications")
			ValidateApplications(contexts)
		})
		Step("Setting up px-backup controllers", func() {
			err := pxbackup.SetControllers(&controllers, nil)
			log.FailOnError(err, "Setting up px-backup controllers failed")
		})
		Step("Creating rules for backup", func() {
			log.InfoD("Creating rules for backup")
			log.InfoD("Creating pre rule for deployed apps")
			for i := 0; i < len(appList); i++ {
				rulesInfo, err := utils.GetPreRulesInfoFromAppParameters(appList[i])
				if err == nil {
					preRuleName := fmt.Sprintf("pre-rule-for-%s-%v", appList[i], time.Now().Unix())
					err := controllers["admin"].Rule(preRuleName, rulesInfo).Add()
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying addition of pre rule [%s]", preRuleName))
				}
			}
			log.InfoD("Creating post rule for deployed apps")
			for i := 0; i < len(appList); i++ {
				rulesInfo, err := utils.GetPostRulesInfoFromAppParameters(appList[i])
				if err == nil {
					postRuleName := fmt.Sprintf("post-rule-for-%s-%v", appList[i], time.Now().Unix())
					err := controllers["admin"].Rule(postRuleName, rulesInfo).Add()
					dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying addition of post rule [%s]", postRuleName))
				}
			}
		})
		Step("Creating backup location and cloud setting", func() {
			log.InfoD("Creating backup location and cloud setting")
			for _, cloudProvider := range providers {
				cloudCredName := fmt.Sprintf("test-%s-ca-%v", cloudProvider, time.Now().Unix())
				err := controllers["admin"].CloudAccount(cloudCredName, cloudProvider).Add()
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying addtion of cloud account [%s]", cloudCredName))
				bkpLocationName = fmt.Sprintf("test-bl-with-%s-%v", cloudCredName, time.Now().Unix())
				err = controllers["admin"].BackupLocation(bkpLocationName, cloudCredName, getGlobalBucketName(cloudProvider)).Add()
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying addtion of backup location [%s]", bkpLocationName))
			}
		})
		Step("Creating backup schedule policies", func() {
			log.InfoD("Creating backup schedule policies")
			log.InfoD("Creating backup interval schedule policy")
			intervalName := fmt.Sprintf("%s-%v", "interval", time.Now().Unix())
			err := controllers["admin"].IntervalSchedulePolicy(intervalName, 5, 15, 2).Add()
			dash.VerifyFatal(err, nil, "Creating interval schedule policy")

			log.InfoD("Creating backup daily schedule policy")
			dailyName := fmt.Sprintf("%s-%v", "daily", time.Now().Unix())
			err = controllers["admin"].DailySchedulePolicy(dailyName, 1, "9:00AM", 2).Add()
			dash.VerifyFatal(err, nil, "Creating daily schedule policy")

			log.InfoD("Creating backup weekly schedule policy")
			weeklyName := fmt.Sprintf("%s-%v", "weekly", time.Now().Unix())
			err = controllers["admin"].WeeklySchedulePolicy(weeklyName, 1, pxbackup.Friday, "9:10AM", 2).Add()
			dash.VerifyFatal(err, nil, "Creating weekly schedule policy")

			log.InfoD("Creating backup monthly schedule policy")
			monthlyName := fmt.Sprintf("%s-%v", "monthly", time.Now().Unix())
			err = controllers["admin"].MonthlySchedulePolicy(monthlyName, 1, 29, "9:20AM", 2).Add()
			dash.VerifyFatal(err, nil, "Creating monthly schedule policy")
		})
		Step("Registering cluster for backup", func() {
			log.InfoD("Registering cluster for backup")
			kubeconfigsList := utils.GetKubeconfigsFromEnv()

			srcClusterKubeconfigPath, err := utils.GetClusterConfigPath(kubeconfigsList[0], utils.DefaultConfigMapName, utils.DefaultConfigMapNamespace)
			log.FailOnError(err, "Getting source cluster kubeconfig path failed")
			err = controllers["admin"].Cluster(SourceClusterName, srcClusterKubeconfigPath).Add()
			dash.VerifyFatal(err, nil, "Verifying addition of source cluster")
			clusterStatus, err = controllers["admin"].WaitForClusterCompletion(SourceClusterName)
			log.FailOnError(err, "Waiting for source cluster failed")
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, "Verifying success of source cluster")

			dstClusterKubeconfigPath, err := utils.GetClusterConfigPath(kubeconfigsList[1], utils.DefaultConfigMapName, utils.DefaultConfigMapNamespace)
			log.FailOnError(err, "Getting destination cluster kubeconfig path failed")
			err = controllers["admin"].Cluster(destinationClusterName, dstClusterKubeconfigPath).Add()
			dash.VerifyFatal(err, nil, "Verifying addition of source cluster")
			clusterStatus, err = controllers["admin"].WaitForClusterCompletion(destinationClusterName)
			log.FailOnError(err, "Waiting for destination cluster failed")
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, "Verifying success of destination cluster")
		})
		Step("Taking backup of all namespaces", func() {
			log.InfoD("Taking backup of all namespaces")
			for _, namespace := range bkpNamespaces {
				backupName = fmt.Sprintf("%s-%s-%s", BackupNamePrefix, namespace, RandomString(4))
				for strings.Contains(strings.Join(backupNames, ","), backupName) {
					backupName = fmt.Sprintf("%s-%s-%s", BackupNamePrefix, namespace, RandomString(4))
				}
				backupNames = append(backupNames, backupName)
				err := controllers["admin"].Backup(backupName, bkpLocationName, SourceClusterName, []string{namespace}).Create()
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of backup [%s]", backupName))
				status, err := controllers["admin"].WaitForBackupCompletion(backupName)
				log.FailOnError(err, "Waiting for backup [%s] failed", backupName)
				dash.VerifyFatal(status, api.BackupInfo_StatusInfo_Success, fmt.Sprintf("Verifying success of backup [%s]", backupName))
			}
		})
		Step("Restoring the backed up namespaces", func() {
			log.InfoD("Restoring the backed up namespaces")
			var restoreNames []string
			for index, namespace := range bkpNamespaces {
				restoreName := fmt.Sprintf("%s-%s-%s", "test-restore", namespace, RandomString(4))
				for strings.Contains(strings.Join(restoreNames, ","), restoreName) {
					restoreName = fmt.Sprintf("%s-%s-%s", "test-restore", namespace, RandomString(4))
				}
				restoreNames = append(restoreNames, restoreName)
				err := controllers["admin"].Restore(restoreName, backupNames[index], destinationClusterName).Create()
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s]", restoreName))
				status, err := controllers["admin"].WaitForRestoreCompletion(restoreName)
				log.FailOnError(err, "Waiting for restore [%s] failed", restoreName)
				dash.VerifyFatal(status, api.RestoreInfo_StatusInfo_Success, fmt.Sprintf("Verifying success of restore [%s]", restoreName))
			}
		})
	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(contexts)
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		log.Info("Deleting deployed namespaces")
		ValidateAndDestroy(contexts, opts)
	})
})
