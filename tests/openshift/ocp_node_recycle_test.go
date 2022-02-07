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

	if Inst().S.String() != "openshift" {
		fmt.Printf("Failed: This test is not supported for scheduler: [%s]", Inst().S.String())
		Expect(Inst().S.String()).To(Equal("openshift"))
	}

	It("Validing the drives and pools after recyling a node", func() {
		Step("Get the storage nodes and delete it", func() {
			workerNodes := node.GetStorageDriverNodes()
			var delNode = workerNodes[0]
			Step(
				fmt.Sprintf("Listing all nodes before recycling a storage node %s", delNode.Name),
				func() {
					workerNodes := node.GetWorkerNodes()
					for x, wNode := range workerNodes {
						logrus.Infof("WorkerNode[%d] is: [%s] and volDriverID is [%s]", x, wNode.Name, wNode.VolDriverNodeID)
					}
				})
			Step(
				fmt.Sprintf("Recycle a node: %s", delNode.Name),
				func() {
					err := Inst().S.RecycleNode(delNode)
					Expect(err).NotTo(HaveOccurred(),
						fmt.Sprintf("Failed to recycle a node [%s]. Error: [%v]", delNode.Name, err))

				})
			Step(
				fmt.Sprintf("Listing all nodes after recycle a storage node %s", delNode.Name),
				func() {
					workerNodes := node.GetWorkerNodes()
					for x, wNode := range workerNodes {
						logrus.Infof("WorkerNode[%d] is: [%s] and volDriverID is [%s]", x, wNode.Name, wNode.VolDriverNodeID)
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
