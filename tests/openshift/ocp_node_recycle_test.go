package tests

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/portworx/torpedo/drivers/node"
	. "github.com/portworx/torpedo/tests"
	"github.com/sirupsen/logrus"
)

func TestOCPRecylceNode(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_recycle.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : Recycle", specReporters)
}

var _ = BeforeSuite(func() {
	InitInstance()
})

// Sanity test for OCP Recycle method
var _ = Describe("{RecycleOCPNode}", func() {

	It("has to delete a node and wait for new node to be ready", func() {
		Step("get the worker nodes and delete it", func() {
			workerNodes := node.GetWorkerNodes()
			var delNode = workerNodes[0]
			Step(
				fmt.Sprintf("Listing all nodes before deleting a worker node %s", delNode.Name),
				func() {
					workerNodes := node.GetWorkerNodes()
					for x, wNode := range workerNodes {
						logrus.Infof("WorkerNode[%d] is: [%s]", x, wNode.Name)
					}
				})
			Step(
				fmt.Sprintf("Deleting a node: %s", delNode.Name),
				func() {
					err := Inst().S.RecycleNode(delNode)
					Expect(err).NotTo(HaveOccurred(),
						fmt.Sprintf("Failed to recycle a node [%s]. Error: [%v]", delNode.Name, err))

				})
			Step(
				fmt.Sprintf("Listing all nodes after deleting a worker node %s", delNode.Name),
				func() {
					workerNodes := node.GetWorkerNodes()
					for x, wNode := range workerNodes {
						logrus.Infof("WorkerNode[%d] is: [%s]", x, wNode.Name)
					}
				})
		})
	})
})

var _ = AfterSuite(func() {
	PerformSystemCheck()
	//ValidateCleanup()
})

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
