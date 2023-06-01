package tests

import (
	"fmt"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/node/ibm"
	"github.com/portworx/torpedo/drivers/volume/portworx/schedops"
	"github.com/portworx/torpedo/pkg/log"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/testrailuttils"
	. "github.com/portworx/torpedo/tests"
	// https://github.com/kubernetes/client-go/issues/242
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	upgradeTimeoutMins = 90 * time.Minute
)

var _ = Describe("{UpgradeScheduler}", func() {
	var testrailID = 58849
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/58849
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("UpgradeScheduler", "Validate scheduler upgrade", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "upgrade scheduler and ensure everything is running fine"
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)

		intitialNodeCount, err := Inst().N.GetASGClusterSize()
		log.FailOnError(err, "error getting ASG cluster size")

		log.InfoD("Validating cluster size before upgrade. Initial Node Count: [%v]", intitialNodeCount)
		ValidateClusterSize(intitialNodeCount)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("upgradescheduler-%d", i))...)
		}

		ValidateApplications(contexts)

		upgradeHops := strings.Split(Inst().SchedUpgradeHops, ",")
		dash.VerifyFatal(len(upgradeHops) > 0, true, "upgrade hops are provided?")

		for _, schedVersion := range upgradeHops {
			schedVersion = strings.TrimSpace(schedVersion)
			stepLog = fmt.Sprintf("start the upgrade of scheduler to version [%v]", schedVersion)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err := Inst().N.SetClusterVersion(schedVersion, upgradeTimeoutMins)
				log.FailOnError(err, "Failed to set cluster version")
			})

			if IsIksCluster() {
				stepLog = fmt.Sprintf("update IKS cluster master node to version %s", schedVersion)
				Step(stepLog, func() {
					err = waitForIKSMasterUpdate(schedVersion)
					dash.VerifyFatal(err, nil, fmt.Sprintf("verify IKS master update to version %s", schedVersion))
				})
				stepLog = fmt.Sprintf("update IKS cluster worker node to version %s", schedVersion)
				Step(stepLog, func() {
					err = upgradeIKSWorkerNodes(schedVersion, ibm.IksPXWorkerpool)
					dash.VerifyFatal(err, nil, fmt.Sprintf("verify IKS worker nodes update to version %s", schedVersion))
				})

			}
			stepLog = fmt.Sprintf("wait for %s minutes for auto recovery of storage nodes",
				Inst().AutoStorageNodeRecoveryTimeout.String())
			Step(stepLog, func() {
				log.InfoD(fmt.Sprintf("wait for %s minutes for auto recovery of storage nodes",
					Inst().AutoStorageNodeRecoveryTimeout.String()))
				time.Sleep(Inst().AutoStorageNodeRecoveryTimeout)
			})

			err = Inst().S.RefreshNodeRegistry()
			log.FailOnError(err, "Node registry refresh failed")

			err = Inst().V.RefreshDriverEndpoints()
			log.FailOnError(err, "Refresh Driver end points failed")
			stepLog = fmt.Sprintf("validate number of storage nodes after scheduler upgrade to [%s]",
				schedVersion)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				ValidateClusterSize(intitialNodeCount)
			})

			Step("validate all apps after upgrade", func() {
				for _, ctx := range contexts {
					ValidateContext(ctx)
				}
			})
			PerformSystemCheck()
		}
		Step("teardown all apps", func() {
			for _, ctx := range contexts {
				TearDownContext(ctx, nil)
			}
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

func waitForIKSMasterUpdate(schedVersion string) error {

	t := func() (interface{}, bool, error) {

		iksCluster, err := ibm.GetCluster()
		if err != nil {
			return nil, false, err
		}

		if strings.Contains(iksCluster.MasterKubeVersion, "pending") {
			return nil, true, fmt.Errorf("waiting for master update to complete.Current status : %s", iksCluster.MasterKubeVersion)
		}
		if strings.Contains(iksCluster.MasterKubeVersion, schedVersion) {
			return nil, false, nil
		}

		return nil, false, fmt.Errorf("master update to %s failed", schedVersion)
	}
	_, err := task.DoRetryWithTimeout(t, 30*time.Minute, 2*time.Minute)

	return err

}

func upgradeIKSWorkerNodes(schedVersion, poolName string) error {

	workers, err := ibm.GetWorkers()
	if err != nil {
		return err
	}

	for _, worker := range workers {

		if worker.PoolName == poolName {
			sNode, err := node.GetNodeByHostName(worker.WorkerID)
			if err != nil {
				return err
			}
			if err = ibm.ReplaceWorkerNodeWithUpdate(sNode); err != nil {
				return err
			}

			if err = waitForIBMNodeToDelete(sNode); err != nil {
				return err
			}

			if err = waitForIBMNodeTODeploy(); err != nil {
				return err
			}

		}

	}

	//storageDriverNodes := node.GetStorageDriverNodes()
	//for _, sNode := range storageDriverNodes {
	//
	//	if err := ibm.ReplaceWorkerNodeWithUpdate(sNode); err != nil {
	//		return err
	//	}
	//
	//	if err := waitForIBMNodeToDelete(sNode); err != nil {
	//		return err
	//	}
	//
	//	if err := waitForIBMNodeTODeploy(); err != nil {
	//		return err
	//	}
	//
	//}

	workers, err = ibm.GetWorkers()
	if err != nil {
		return err
	}

	for _, worker := range workers {
		if worker.PoolName == poolName && !strings.Contains(worker.KubeVersion.Actual, schedVersion) {
			return fmt.Errorf("node %s has version %s expected %s", worker.WorkerID, worker.KubeVersion.Actual, schedVersion)
		}
	}

	return nil
}

var _ = Describe("{MigratePXCluster}", func() {

	JustBeforeEach(func() {
		StartTorpedoTest("MigratePXCluster", "Validate Worker pool upgrade and PX migration", nil, 0)

	})
	var contexts []*scheduler.Context

	stepLog := "upgrade scheduler, migrate PX  and ensure everything is running fine"
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)

		dash.VerifyFatal(Inst().MigrationWorkerPool != "", true, "verify migration ")
		if Inst().MigrationWorkerPool == "" {
			log.FailOnError(fmt.Errorf("migration worker pool value not passed"), "migration pool name is mandatory for MigratePXCluster ")
		}
		iksWorkers, err := ibm.GetWorkers()
		log.FailOnError(err, "error getting IBM workers")
		isWorkerPoolExist := false
		migrationWorkersMap := make(map[string]ibm.Worker)
		defaultWorkersMap := make(map[string]ibm.Worker)
		for _, iksWorker := range iksWorkers {
			if iksWorker.PoolName == Inst().MigrationWorkerPool {
				isWorkerPoolExist = true
				migrationWorkersMap[iksWorker.WorkerID] = iksWorker
			}
			if iksWorker.PoolName == ibm.IksPXWorkerpool {
				defaultWorkersMap[iksWorker.WorkerID] = iksWorker
			}
		}
		dash.VerifyFatal(isWorkerPoolExist, true, fmt.Sprintf("verify worker pool %s exists.", Inst().MigrationWorkerPool))

		intitialNodeCount, err := Inst().N.GetASGClusterSize()
		log.FailOnError(err, "error getting ASG cluster size")

		initialStNodes := node.GetStorageNodes()

		log.InfoD("Validating cluster size before upgrade. Initial Node Count: [%v]", intitialNodeCount)
		ValidateClusterSize(intitialNodeCount)

		ValidateApplications(contexts)

		upgradeHops := strings.Split(Inst().SchedUpgradeHops, ",")
		dash.VerifyFatal(len(upgradeHops) > 0, true, "upgrade hops are provided?")

		for _, schedVersion := range upgradeHops {
			schedVersion = strings.TrimSpace(schedVersion)
			stepLog = fmt.Sprintf("start the upgrade of scheduler to version [%v]", schedVersion)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err := Inst().N.SetClusterVersion(schedVersion, upgradeTimeoutMins)
				log.FailOnError(err, "Failed to set cluster version")
			})

			if IsIksCluster() {
				stepLog = fmt.Sprintf("update IKS cluster master node to version %s", schedVersion)
				Step(stepLog, func() {
					err = waitForIKSMasterUpdate(schedVersion)
					dash.VerifyFatal(err, nil, fmt.Sprintf("verify IKS master update to version %s", schedVersion))
				})
				stepLog = fmt.Sprintf("update IKS cluster worker node to version %s", schedVersion)
				Step(stepLog, func() {
					err = upgradeIKSWorkerNodes(schedVersion, Inst().MigrationWorkerPool)
					dash.VerifyFatal(err, nil, fmt.Sprintf("verify IKS worker nodes update to version %s", schedVersion))
				})

			}
			stepLog = fmt.Sprintf("wait for %s minutes for auto recovery of storage nodes",
				Inst().AutoStorageNodeRecoveryTimeout.String())
			Step(stepLog, func() {
				log.InfoD(fmt.Sprintf("wait for %s minutes for auto recovery of storage nodes",
					Inst().AutoStorageNodeRecoveryTimeout.String()))
				time.Sleep(Inst().AutoStorageNodeRecoveryTimeout)
			})

			err = Inst().S.RefreshNodeRegistry()
			log.FailOnError(err, "Node registry refresh failed")

			err = Inst().V.RefreshDriverEndpoints()
			log.FailOnError(err, "Refresh Driver end points failed")
			stepLog = fmt.Sprintf("validate number of storage nodes after scheduler upgrade to [%s]",
				schedVersion)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				ValidateClusterSize(intitialNodeCount)
			})

			Step("validate all apps after upgrade", func() {
				for _, ctx := range contexts {
					ValidateContext(ctx)
				}
			})
		}

		Step("enabling px on migration pool nodes", func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("upgradescheduler-%d", i))...)
			}

			for k := range migrationWorkersMap {
				n, err := node.GetNodeByHostName(k)
				log.FailOnError(err, fmt.Sprintf("error getting node with hostname %s", k))
				err = Inst().S.RemoveLabelOnNode(n, schedops.PXEnabledLabelKey)
				log.FailOnError(err, fmt.Sprintf("error removing label %s on node %s", schedops.PXEnabledLabelKey, n.Name))
				err = Inst().V.WaitDriverUpOnNode(n, 5*time.Minute)
				log.FailOnError(err, fmt.Sprintf("wait for px up on node %s", n.Name))
			}

			err = Inst().S.RefreshNodeRegistry()
			log.FailOnError(err, "Node registry refresh failed")

			err = Inst().V.RefreshDriverEndpoints()
			log.FailOnError(err, "Refresh Driver end points failed")
		})

		Step("Initiate worker pool migration", func() {

			for k := range defaultWorkersMap {

				n, err := node.GetNodeByHostName(k)
				log.FailOnError(err, fmt.Sprintf("error getting node with hostname %s", k))
				err = ibm.RemoveWorkerNode(n)
				log.FailOnError(err, fmt.Sprintf("error removing node %s", n.Hostname))
				err = waitForIBMNodeToDelete(n)
				log.FailOnError(err, fmt.Sprintf("node %s node deleted", n.Hostname))
				err = Inst().S.RefreshNodeRegistry()
				log.FailOnError(err, "Node registry refresh failed")

				err = Inst().V.RefreshDriverEndpoints()
				log.FailOnError(err, "Refresh Driver end points failed")
				newStNodes := node.GetStorageNodes()
				dash.VerifyFatal(len(initialStNodes), len(newStNodes), fmt.Sprintf("verfiy storage node count after deleteing %s", n.Name))

			}

		})

		stepLog = fmt.Sprintf("validate number of storage nodes after migration")
		Step(stepLog, func() {
			log.InfoD(stepLog)
			ValidateClusterSize(intitialNodeCount)
		})

		Step("validate all apps after upgrade", func() {
			for _, ctx := range contexts {
				ValidateContext(ctx)
			}
		})
		PerformSystemCheck()

		Step("teardown all apps", func() {
			for _, ctx := range contexts {
				TearDownContext(ctx, nil)
			}
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})
