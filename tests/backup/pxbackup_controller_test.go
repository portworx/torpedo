package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/portworx/torpedo/drivers/backup/controllers/cluster"
	"github.com/portworx/torpedo/drivers/backup/controllers/pxbackup"
	"github.com/portworx/torpedo/drivers/backup/utils"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	"time"
)

// This testcase verifies basic backup rule,backup location, cloud setting
var _ = Describe("{NewBasicBackupCreation}", func() {
	var (
		testRailId       int
		clControllerMap  map[string]*cluster.ClusterController
		pxbControllerMap map[string]*pxbackup.PxBackupController

		//controllers     map[string]*pxbackup.PxBackupController
		//appList         = Inst().AppList
		//contexts        []*scheduler.Context
		//appContexts     []*scheduler.Context
		//bkpNamespaces   []string
		//clusterStatus   api.ClusterInfo_StatusInfo_Status
		//bkpLocationName string
		//providers       []string
		//backupName      string
		//backupNames     []string

		appNamespaces []string
	)

	JustBeforeEach(func() {
		testRailId = 31313
		StartTorpedoTest("NewBasicBackupCreation", "Basic Backup Creation", nil, testRailId)
		Step("Add px-backup and cluster controllers", func() {
			err := pxbackup.AddPxBackupControllersToMap(&pxbControllerMap, nil)
			log.FailOnError(err, "failed to add px-backup controller to pxb-controller-map")

			err = cluster.AddTestCaseClusterControllers(&clControllerMap, testRailId)
			log.FailOnError(err, "failed to add test case cluster controllers to cl-controller-map")
		})
		for _, appKey := range Inst().AppList {
			namespace := fmt.Sprintf("%s-%d", appKey, testRailId)
			err := clControllerMap[utils.DefaultSourceClusterName].Application(appKey).ScheduleOnNamespace(namespace)
			log.FailOnError(err, fmt.Sprintf("failed to schedule application [%s] on [%s]", appKey, utils.DefaultSourceClusterName))
			appNamespaces = append(appNamespaces, namespace)
			err = clControllerMap[utils.DefaultSourceClusterName].Namespace(namespace).Validate()
			log.FailOnError(err, fmt.Sprintf("failed to validate namespace [%s] on [%s]", namespace, utils.DefaultSourceClusterName))
		}
	})
	It("Basic Backup Creation Pxb Controller", func() {
		//Step("Creating rules for backup", func() {
		//	log.InfoD("Creating rules for backup")
		//	log.InfoD("Creating pre rule for deployed apps")
		//	for i := 0; i < len(appList); i++ {
		//		rulesInfo, err := utils.GetPreRulesInfoFromAppParameters(appList[i])
		//		if err == nil {
		//			preRuleName := fmt.Sprintf("pre-rule-for-%s-%v", appList[i], time.Now().Unix())
		//			err := controllers["admin"].Rule(preRuleName, rulesInfo).Add()
		//			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying addition of pre rule [%s]", preRuleName))
		//		}
		//	}
		//	log.InfoD("Creating post rule for deployed apps")
		//	for i := 0; i < len(appList); i++ {
		//		rulesInfo, err := utils.GetPostRulesInfoFromAppParameters(appList[i])
		//		if err == nil {
		//			postRuleName := fmt.Sprintf("post-rule-for-%s-%v", appList[i], time.Now().Unix())
		//			err := controllers["admin"].Rule(postRuleName, rulesInfo).Add()
		//			dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying addition of post rule [%s]", postRuleName))
		//		}
		//	}
		//})
		Step("Creating backup location and cloud setting", func() {
			log.InfoD("Creating backup location and cloud setting")
			for _, cloudProvider := range utils.GetProvidersFromEnv() {
				cloudCredName := fmt.Sprintf("test-%s-ca-%v", cloudProvider, time.Now().Unix())
				err := pxbControllerMap["admin"].CloudAccount(cloudCredName).Add(cloudProvider)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying addtion of cloud account [%s]", cloudCredName))
				//bkpLocationName = fmt.Sprintf("test-bl-with-%s-%v", cloudCredName, time.Now().Unix())
				//err = controllers["admin"].BackupLocation(bkpLocationName, cloudCredName, getGlobalBucketName(cloudProvider)).Add()
				//dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying addtion of backup location [%s]", bkpLocationName))
			}
		})
		//Step("Creating backup schedule policies", func() {
		//	log.InfoD("Creating backup schedule policies")
		//	log.InfoD("Creating backup interval schedule policy")
		//	intervalName := fmt.Sprintf("%s-%v", "interval", time.Now().Unix())
		//	err := controllers["admin"].IntervalSchedulePolicy(intervalName, 5, 15, 2).Add()
		//	dash.VerifyFatal(err, nil, "Creating interval schedule policy")
		//
		//	log.InfoD("Creating backup daily schedule policy")
		//	dailyName := fmt.Sprintf("%s-%v", "daily", time.Now().Unix())
		//	err = controllers["admin"].DailySchedulePolicy(dailyName, 1, "9:00AM", 2).Add()
		//	dash.VerifyFatal(err, nil, "Creating daily schedule policy")
		//
		//	log.InfoD("Creating backup weekly schedule policy")
		//	weeklyName := fmt.Sprintf("%s-%v", "weekly", time.Now().Unix())
		//	err = controllers["admin"].WeeklySchedulePolicy(weeklyName, 1, pxbackup.Friday, "9:10AM", 2).Add()
		//	dash.VerifyFatal(err, nil, "Creating weekly schedule policy")
		//
		//	log.InfoD("Creating backup monthly schedule policy")
		//	monthlyName := fmt.Sprintf("%s-%v", "monthly", time.Now().Unix())
		//	err = controllers["admin"].MonthlySchedulePolicy(monthlyName, 1, 29, "9:20AM", 2).Add()
		//	dash.VerifyFatal(err, nil, "Creating monthly schedule policy")
		//})
		//Step("Registering cluster for backup", func() {
		//	log.InfoD("Registering cluster for backup")
		//	kubeconfigsList := utils.GetKubeconfigsFromEnv()
		//
		//	srcClusterKubeconfigPath, err := utils.GetClusterConfigPath(kubeconfigsList[0], utils.DefaultConfigMapName, utils.DefaultConfigMapNamespace)
		//	log.FailOnError(err, "Getting source cluster kubeconfig path failed")
		//	err = controllers["admin"].Cluster(SourceClusterName, srcClusterKubeconfigPath).Add()
		//	dash.VerifyFatal(err, nil, "Verifying addition of source cluster")
		//	clusterStatus, err = controllers["admin"].WaitForClusterCompletion(SourceClusterName)
		//	log.FailOnError(err, "Waiting for source cluster failed")
		//	dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, "Verifying success of source cluster")
		//
		//	dstClusterKubeconfigPath, err := utils.GetClusterConfigPath(kubeconfigsList[1], utils.DefaultConfigMapName, utils.DefaultConfigMapNamespace)
		//	log.FailOnError(err, "Getting destination cluster kubeconfig path failed")
		//	err = controllers["admin"].Cluster(destinationClusterName, dstClusterKubeconfigPath).Add()
		//	dash.VerifyFatal(err, nil, "Verifying addition of source cluster")
		//	clusterStatus, err = controllers["admin"].WaitForClusterCompletion(destinationClusterName)
		//	log.FailOnError(err, "Waiting for destination cluster failed")
		//	dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, "Verifying success of destination cluster")
		//})
		//Step("Taking backup of all namespaces", func() {
		//	log.InfoD("Taking backup of all namespaces")
		//	for _, namespace := range bkpNamespaces {
		//		backupName = fmt.Sprintf("%s-%s-%s", BackupNamePrefix, namespace, RandomString(4))
		//		for strings.Contains(strings.Join(backupNames, ","), backupName) {
		//			backupName = fmt.Sprintf("%s-%s-%s", BackupNamePrefix, namespace, RandomString(4))
		//		}
		//		backupNames = append(backupNames, backupName)
		//		err := controllers["admin"].Backup(backupName, bkpLocationName, SourceClusterName, []string{namespace}).Create()
		//		dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying creation of backup [%s]", backupName))
		//		status, err := controllers["admin"].WaitForBackupCompletion(backupName)
		//		log.FailOnError(err, "Waiting for backup [%s] failed", backupName)
		//		dash.VerifyFatal(status, api.BackupInfo_StatusInfo_Success, fmt.Sprintf("Verifying success of backup [%s]", backupName))
		//	}
		//})
		//Step("Restoring the backed up namespaces", func() {
		//	log.InfoD("Restoring the backed up namespaces")
		//	var restoreNames []string
		//	for index, namespace := range bkpNamespaces {
		//		restoreName := fmt.Sprintf("%s-%s-%s", "test-restore", namespace, RandomString(4))
		//		for strings.Contains(strings.Join(restoreNames, ","), restoreName) {
		//			restoreName = fmt.Sprintf("%s-%s-%s", "test-restore", namespace, RandomString(4))
		//		}
		//		restoreNames = append(restoreNames, restoreName)
		//		err := controllers["admin"].Restore(restoreName, backupNames[index], destinationClusterName).Create()
		//		dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s]", restoreName))
		//		status, err := controllers["admin"].WaitForRestoreCompletion(restoreName)
		//		log.FailOnError(err, "Waiting for restore [%s] failed", restoreName)
		//		dash.VerifyFatal(status, api.RestoreInfo_StatusInfo_Success, fmt.Sprintf("Verifying success of restore [%s]", restoreName))
		//	}
		//})
	})
	JustAfterEach(func() {
		//for clusterName, clusterController := range clControllerMap {
		//	err := clusterController.Cleanup()
		//	log.FailOnError(err, "failed to clean up cluster [%s]", clusterName)
		//}
		//defer EndPxBackupTorpedoTest(contexts)
		//opts := make(map[string]bool)
		//opts[SkipClusterScopedObjects] = true
		//log.Info("Deleting deployed namespaces")
		//ValidateAndDestroy(contexts, opts)
	})
})
