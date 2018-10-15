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
	It("has to decomission a node and check if node was decommissioned successfuly", func() {
		var contexts []*scheduler.Context
		for i := 0; i < Inst().ScaleFactor; i++ {
			contexts = append(contexts, ScheduleAndValidate(fmt.Sprintf("%s-%d", testName, i))...)
		}

		Step("get nodes for all apps in test, select randomly a node and decommission it", func() {
			for _, ctx := range contexts {
				var appNode node.Node

				Step(fmt.Sprintf("get nodes where %s app is running and select a node", ctx.App.Key), func() {
					appNodes, err := Inst().S.GetNodesForApp(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(appNodes).NotTo(BeEmpty())
					randNode := rand.Intn(len(appNodes))
					appNode = appNodes[randNode]
				})

				Step(fmt.Sprintf("decomission the node"), func() {
					err := Inst().S.DecommissionNode(appNode)
					Expect(err).NotTo(HaveOccurred())
					Step("wait for the node to be decomissioned", func() {
						time.Sleep(6 * time.Minute)
					})

					Step(fmt.Sprintf("check if the node was decomissioned"), func() {
						ValidateContext(ctx)
					})
				})
			}
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
