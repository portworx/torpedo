package tests

import (
	. "github.com/onsi/ginkgo"
	"github.com/portworx/torpedo/drivers/vcluster"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	"os"
	"path/filepath"
	"time"
)

var _ = Describe("CreatevCluster", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("VclusterOperations", "Create  Vcluster", nil, 0)
	})
	var vclusterNames = []string{"my-vcluster2"}
	It("Create and connect to vclusters and run a sample method", func() {
		steplog := "Create vClusters"
		log.InfoD(steplog)
		Step(steplog, func() {
			for _, name := range vclusterNames {
				currentDir, err := os.Getwd()
				log.FailOnError(err, "Could not get absolute path to current Dir")
				vClusterPath := filepath.Join(currentDir, "..", "deployments", "customconfigs", "vcluster.yaml")
				absPath, err := filepath.Abs(vClusterPath)
				log.FailOnError(err, "Could not get absolute path to vcluster.yaml")
				err = vcluster.CreateVCluster(name, absPath)
				log.FailOnError(err, "Failed to create vCluster")
			}
		})
		steplog = "Wait for all vClusters to come up in Running State"
		log.InfoD(steplog)
		Step(steplog, func() {
			for _, name := range vclusterNames {
				err := vcluster.WaitForVClusterRunning(name, 10*time.Minute)
				log.FailOnError(err, "Vcluster did not come up in time")
				err = vcluster.GetVClusterSecret("vc-my-vcluster2", "vcluster-my-vcluster2")
				log.FailOnError(err, "error while getting the vcluster secret")
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		//for _, name := range vclusterNames {
		//	vcluster.DeleteVCluster(name)
		//}
	})
})
