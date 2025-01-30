package tests

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/sched-ops/task"
	"github.com/pure-px/sched-ops/k8s/operator"
	"github.com/pure-px/torpedo/drivers/backup"
	"github.com/pure-px/torpedo/drivers/node"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/log"
	. "github.com/pure-px/torpedo/tests"
)

// This testcase verifies if we can add cluster without stork and install stork and retry and get cluster status as online after getting disconnected
var _ = Describe("{AddClusterWithoutStorkAndInstallStorkAndRetry}", Label(TestCaseLabelsMap[StorkInstallClusterStatusCheck]...), func() {

	var (
		scheduledAppContexts []*scheduler.Context
		bkpNamespaces        []string
		sourceClusterUid     string
		cloudCredName        string
		cloudCredUID         string
		clusterStatus        api.ClusterInfo_StatusInfo_Status
		storkInstalledFlag   bool
	)
	storkInstalledFlag = false
	backupLocationMap := make(map[string]string)
	JustBeforeEach(func() {

		bkpNamespaces = make([]string, 0)

		StartPxBackupTorpedoTest("VerifyAddClusterWithoutStorkAndInstallStorkAndRetry", "Add cluster without stork and install stork and retry and get cluster status as online after getting disconnected", nil, 300483, ABadgujar, Q4FY24)
		//1.Step-Deploy Applications in the Cluster
		log.InfoD("Deploy applications")
		scheduledAppContexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", TaskNamePrefix, i)
			appContexts := ScheduleApplications(taskName)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = AppReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
				scheduledAppContexts = append(scheduledAppContexts, ctx)
			}
		}

	})
	It("Add cluster without stork and install stork and retry and get cluster status as online after getting disconnected", func() {

		//2.Step-Validate Deployed Applications
		Step("Validate applications", func() {
			log.InfoD("Validating apps")
			ValidateApplications(scheduledAppContexts)
		})

		//3.Remove Stork through Command Line
		Step("Remove Stork through Command Line", func() {
			log.InfoD("Get the stc spec from the cluster")
			stc, err := Inst().V.GetDriver()
			log.FailOnError(err, "Failed to get driver")

			log.InfoD("Check if the stork is enabled in the stc")
			if stc.Spec.Stork.Enabled == false {
				log.InfoD("Stork is Disabled so skipping disabling part")
			} else if stc.Spec.Stork.Enabled == true {
				// Change the stork.enabled flag to false
				stc.Spec.Stork.Enabled = false

				pxOperator := operator.Instance()
				_, err = pxOperator.UpdateStorageCluster(stc)
				log.FailOnError(err, "Failed to update the storage cluster")
				log.InfoD("STC updated: stork.enabled set to false.")
			}

			log.InfoD("Remove Stork CRDs through Command Line")
			masterNode := node.GetMasterNodes()[0]
			cmd := "kubectl delete $(kubectl get crd -o name | grep stork)"
			output, err := RunCmdGetOutput(cmd, masterNode)
			log.FailOnError(err, fmt.Sprintf("Failed to run [%s] command on node [%s], error : [%s]", cmd, masterNode, err))
			log.Infof("Output from command [%s] -\n%s", cmd, output)

			//Sleep for a minute to give some time for all CRDs and deployment to be deleted and then Add Source Cluster
			time.Sleep(time.Minute)
		})

		//4.Step-Create Single Only Source Cluster which should be in Failed State
		Step("Create Single Only Source Cluster", func() {
			log.InfoD("Create Single Only Source Cluster")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			log.InfoD("Entering Cluster Config Step")
			srcClusterConfigPath, err := GetSourceClusterConfigPath()
			log.Infof("Save cluster %s kubeconfig to %s", SourceClusterName, srcClusterConfigPath)
			log.FailOnError(err, fmt.Sprintf("Error in Fetching Source Cluster Config Path - [%v] ", err))

			log.InfoD("Entering Create Cluster Step")
			err = CreateCluster(SourceClusterName, srcClusterConfigPath, BackupOrgID, "", "", ctx)
			log.InfoD("Error While Creating Cluster is - %v", err)
			dash.VerifyFatal(strings.Contains(err.Error(), "failed to validate stork deployment on cluster"), true, "Expected Failure on Creating Source and Destination Cluster")

			clusterStatus, err = Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			log.InfoD("Cluster Status is - %s", clusterStatus)
			log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Failed, fmt.Sprintf("Verifying if [%s] cluster is disconnected/Offline/Failed", SourceClusterName))
			sourceClusterUid, err = Inst().Backup.GetClusterUID(ctx, BackupOrgID, SourceClusterName)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Fetching [%s] cluster uid - %s", SourceClusterName, sourceClusterUid))
		})

		//5.Step-Install Stork through Command Line
		Step("Install Stork through Command Line", func() {
			log.InfoD("Install Stork through Command Line")
			log.InfoD("Check if the stork is enabled in the stc")
			stc, err := Inst().V.GetDriver()
			log.FailOnError(err, "Failed to get driver")
			if stc.Spec.Stork.Enabled == true {
				log.InfoD("Stork is Enabled so skipping enabling part")
			} else if stc.Spec.Stork.Enabled == false {
				// Change the stork.enabled flag to false
				stc.Spec.Stork.Enabled = true
				pxOperator := operator.Instance()
				_, err = pxOperator.UpdateStorageCluster(stc)
				log.FailOnError(err, "Failed to update the storage cluster")
				log.InfoD("STC updated: stork.enabled set to true.")
			}
			storkInstalledFlag = true
			time.Sleep(time.Minute * 3)

		})

		//6.Step-Check Cluster Status After Stork Re-installation
		Step("Check Cluster Status if Online After Stork Re-installation", func() {
			log.InfoD("Check Cluster Status if Online After Stork Re-installation")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			// clusterStatus, err := Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
			// log.FailOnError(err, fmt.Sprintf("Fetching [%s] cluster status", SourceClusterName))
			log.InfoD("Cluster Status before checking is - %s", clusterStatus)

			clusterStatus := func() (interface{}, bool, error) {
				checkClusterStatus, err := Inst().Backup.GetClusterStatus(BackupOrgID, SourceClusterName, ctx)
				if err != nil {
					return "", true, err
				}
				if checkClusterStatus == api.ClusterInfo_StatusInfo_Online {
					return "", false, nil
				}
				return "", true, fmt.Errorf("the %s cluster state is not Online yet", SourceClusterName)
			}
			_, err = task.DoRetryWithTimeout(clusterStatus, time.Minute*12, time.Minute)
			dash.VerifySafely(err, nil, "Cluster is verified as Online")
		})

	})
	JustAfterEach(func() {
		defer EndPxBackupTorpedoTest(scheduledAppContexts)
		log.InfoD("Entered Cleanup")

		//Reinstall Stork here too to ensure that in case a test case fails after stork deletion, stork is installed back in that test case
		if storkInstalledFlag == false {
			log.InfoD("Stork flag is %v", storkInstalledFlag)
			log.InfoD("Install Stork through Command Line During Cleanup")

			log.InfoD("Check if the stork is disabled in the stc")
			stc, err := Inst().V.GetDriver()
			log.FailOnError(err, "Failed to get driver")
			if stc.Spec.Stork.Enabled == true {
				log.InfoD("Stork is Enabled so skipping enabling part")
			} else if stc.Spec.Stork.Enabled == false {
				// Change the stork.enabled flag to false
				stc.Spec.Stork.Enabled = true
				pxOperator := operator.Instance()
				_, err = pxOperator.UpdateStorageCluster(stc)
				log.FailOnError(err, "Failed to update the storage cluster")
				log.InfoD("STC updated: stork.enabled set to false.")
			}
			time.Sleep(time.Minute * 3)
		}
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		//Cleanup Cluster
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)

	})

})
