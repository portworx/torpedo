package tests

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/libopenstorage/openstorage/api"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	. "github.com/portworx/torpedo/tests"
)

const (
	defaultTimeout       = 6 * time.Minute
	defaultRetryInterval = 10 * time.Second
)

func TestDecommissionNode(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_DecommissionNode.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo: DecommissionNode", specReporters)
}

var _ = BeforeSuite(func() {
	InitInstance()
})

var _ = Describe("{DecommissionNode}", func() {
	testName := "decommissionnode"
	It("has to decommission a node and check if node was decommissioned successfuly", func() {
		var contexts []*scheduler.Context
		for i := 0; i < Inst().ScaleFactor; i++ {
			contexts = append(contexts, ScheduleAndValidate(fmt.Sprintf("%s-%d", testName, i))...)
		}

		Step("pick a random node and decommission it", func() {

			var nodeToDecommission node.Node
			Step(fmt.Sprintf("pick a node to decommission"), func() {
				workerNodes := node.GetWorkerNodes()
				Expect(workerNodes).NotTo(BeEmpty())
				randNode := rand.Intn(len(workerNodes))
				nodeToDecommission = workerNodes[randNode]
			})

			Step(fmt.Sprintf("decommission the node"), func() {
				err := Inst().S.PrepareNodeToDecommission(nodeToDecommission)
				Expect(err).NotTo(HaveOccurred())
				err = Inst().V.DecommissionNode(nodeToDecommission)
				Expect(err).NotTo(HaveOccurred())
				Step(fmt.Sprintf("check if the node was decommissioned"), func() {
					t := func() (interface{}, bool, error) {
						status, err := Inst().V.GetNodeStatus(nodeToDecommission)
						if err != nil {
							return false, false, err
						}
						if *status == api.Status_STATUS_NONE {
							return true, false, nil
						}
						return false, true, fmt.Errorf("Node %s not decomissioned yet", nodeToDecommission.Name)
					}
					decommissioned, err := task.DoRetryWithTimeout(t, defaultTimeout, defaultRetryInterval)
					Expect(err).NotTo(HaveOccurred())
					Expect(decommissioned).To(BeTrue())
				})
			})
			Step(fmt.Sprintf("Rejoin node"), func() {
				err := Inst().V.RejoinNode(nodeToDecommission)
				Expect(err).NotTo(HaveOccurred())
				err = Inst().V.WaitDriverUpOnNode(nodeToDecommission)
				Expect(err).NotTo(HaveOccurred())
			})

		})

		Step("destroy apps", func() {
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		})

	})
})

var _ = AfterSuite(func() {
	PerformSystemCheck()
	CollectSupport()
	ValidateCleanup()
})

func init() {
	ParseFlags()
}
