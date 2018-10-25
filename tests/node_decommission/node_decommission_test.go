package tests

import (
"fmt"
"testing"
"time"

. "github.com/onsi/ginkgo"
. "github.com/onsi/gomega"
"github.com/portworx/torpedo/drivers/node"
"github.com/portworx/torpedo/drivers/scheduler"
. "github.com/portworx/torpedo/tests"
"math/rand"
	"github.com/portworx/sched-ops/task"
)

const (
	defaultTimeout       = 6 * time.Minute
	defaultRetryInterval = 10 * time.Second
)

func TestDecommissionNode(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Torpedo: DecommissionNode")
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
				err := Inst().S.DecommissionNode(nodeToDecommission)
				Expect(err).NotTo(HaveOccurred())
				err = Inst().V.DecommissionNode(nodeToDecommission)
				Expect(err).NotTo(HaveOccurred())
				Step(fmt.Sprintf("check if the node was decommissioned"), func() {
					t := func() (interface{}, bool, error) {
						status, err := Inst().V.DecommissionNodeStatus(nodeToDecommission)
						if err != nil {
							return 0, false, fmt.Errorf("Not able to check node decomission. Cause: %v", err)
						}
						if len(status) == 0 {
							return 0, true, nil
						}
						return 0, false, nil
					}
					decomissioned, err := task.DoRetryWithTimeout(t, defaultTimeout, defaultRetryInterval)
					Expect(err).NotTo(HaveOccurred())
					Expect(decomissioned).NotTo(BeTrue())
				})
			})

		})

		ValidateAndDestroy(contexts, nil)
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
