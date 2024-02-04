package tests

import (
	"fmt"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/pkg/log"
	"net/url"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/portworx/torpedo/drivers/scheduler"
	. "github.com/portworx/torpedo/tests"
	"k8s.io/api/core/v1"
)

var _ = Describe("{UpgradeCluster}", func() {
	var contexts []*scheduler.Context

	JustBeforeEach(func() {
		tags := map[string]string{
			"upgradeCluster": "true",
		}
		StartTorpedoTest("UpgradeCluster", "Upgrade cluster test", tags, 0)
	})
	It("upgrade scheduler and ensure everything is running fine", func() {
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("upgradecluster-%d", i))...)
		}

		ValidateApplications(contexts)

		var versions []string
		if len(Inst().SchedUpgradeHops) > 0 {
			versions = strings.Split(Inst().SchedUpgradeHops, ",")
		}
		Expect(versions).NotTo(BeEmpty())
		stopSignal := make(chan struct{})
		var mError error
		go getClusterNodesInfo(stopSignal, &mError)
		defer close(stopSignal)

		for _, version := range versions {
			Step("start scheduler upgrade", func() {
				err := Inst().S.UpgradeScheduler(version)
				Expect(err).NotTo(HaveOccurred())
			})

			dash.VerifyFatal(mError, nil, "validate no parallel upgrade of nodes")

			Step("validate storage components", func() {
				u, err := url.Parse(fmt.Sprintf("%s/%s", Inst().StorageDriverUpgradeEndpointURL, Inst().StorageDriverUpgradeEndpointVersion))
				Expect(err).NotTo(HaveOccurred())
				err = Inst().V.ValidateDriver(u.String(), true)
				Expect(err).NotTo(HaveOccurred())
			})

			Step("validate all apps after upgrade", func() {
				ValidateApplications(contexts)
			})
		}

		Step("destroy apps", func() {
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

func getClusterNodesInfo(stopSignal <-chan struct{}, mError *error) {

	masterNodeStatus := make(map[string]bool)
	workerNodeStatus := make(map[string]bool)
	itr := 1
	for {
		log.Infof("K8s node validation iteration: #%d", itr)
		select {
		case <-stopSignal:
			log.Infof("Exiting node validations routine")
			return
		default:
			nodeList, err := core.Instance().GetNodes()
			if err != nil {
				mError = &err
				return
			}
			masterCount := 0
			workerCount := 0
			for _, k8sNode := range nodeList.Items {
				if strings.Contains(k8sNode.Name, "master") {
					masterNodeStatus[k8sNode.Name] = k8sNode.Spec.Unschedulable
					if k8sNode.Spec.Unschedulable {
						masterCount += 1
					}
				} else {
					workerNodeStatus[k8sNode.Name] = k8sNode.Spec.Unschedulable
					if k8sNode.Spec.Unschedulable {
						workerCount += 1
					}
				}
			}
			log.Infof("Master Map status is %#v", masterNodeStatus)
			log.Infof("Worker Map status is %#v", workerNodeStatus)
			if masterCount > 1 || workerCount > 1 {
				err = fmt.Errorf("multiple controlpnane or worker nodes are Unschedulable at same time,"+
					"controlplane:%#v,worker:%#v", masterNodeStatus, workerNodeStatus)
				mError = &err
				return
			}

			itr++
			time.Sleep(30 * time.Second)
		}
	}

}

func isNodeUpgradeInProgress(node v1.Node) bool {
	log.Infof("Node [%s] Unschedulable status: %v ", node.Name, node.Spec.Unschedulable)
	return node.Spec.Unschedulable
}
