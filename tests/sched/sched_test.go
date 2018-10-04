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
)

func TestStopStartScheduler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Torpedo: StopStartScheduler")
}

var _ = BeforeSuite(func() {
	InitInstance()
})

var _ = Describe("{StopScheduler}", func() {
	testName := "stopscheduler"
	It("has to stop scheduler service and check if all applications are fine", func() {
		var err error
		var contexts []*scheduler.Context
		for i := 0; i < Inst().ScaleFactor; i++ {
			contexts = append(contexts, ScheduleAndValidate(fmt.Sprintf("%s-%d", testName, i))...)
		}

		Step("get nodes for all apps in test and induce drive failure on one of the nodes", func() {
			for _, ctx := range contexts {
				var appNodes []node.Node

				Step(fmt.Sprintf("get nodes where %s app is running", ctx.App.Key), func() {
					appNodes, err = Inst().S.GetNodesForApp(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(appNodes).NotTo(BeEmpty())
				})

				Step(fmt.Sprintf("stop scheduler service"), func() {
					for _, nodeToStopService := range appNodes {
						err := Inst().S.StopSchedOnNode(nodeToStopService)
						Expect(err).NotTo(HaveOccurred())
					}
					Step("wait for the service to stop", func() {
						time.Sleep(6 * time.Minute)
					})

					Step(fmt.Sprintf("check if apps are running"), func() {
						ValidateContext(ctx)
					})
				})
			}
		})

		ValidateAndDestroy(contexts, nil)
	})
})

var _ = Describe("{StartScheduler}", func() {
	It("has to start scheduler service and check if all applications are fine", func() {
		Step("get nodes for all apps in test and induce drive failure on one of the nodes", func() {
				var appNodes []node.Node

				Step(fmt.Sprintf("get workers node"), func() {
					appNodes = node.GetWorkerNodes()
					Expect(appNodes).NotTo(BeEmpty())
				})

				Step(fmt.Sprintf("start scheduler service"), func() {
					for _, nodeToStartService := range appNodes {
						err := Inst().S.StartSchedOnNode(nodeToStartService)
						Expect(err).NotTo(HaveOccurred())
					}
				})
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
