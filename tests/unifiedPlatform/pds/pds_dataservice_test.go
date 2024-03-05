package tests

import (
	. "github.com/onsi/ginkgo/v2"
	dslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	platformTests "github.com/portworx/torpedo/tests/unifiedPlatform/platform"
)

var _ = Describe("{DeployDataServicesOnDemand}", func() {
	steplog := "Data service deployment"
	JustBeforeEach(func() {
		StartTorpedoTest("DeployDataService", "Deploy data services", nil, 0)
	})

	log.InfoD(steplog)
	It("Deploy and Validate DataService", func() {
		for _, ds := range platformTests.NewPdsParams.DataServiceToTest {
			_, err := dslibs.DeployDataService(ds)
			log.FailOnError(err, "Error while deploying ds %v\n")
		}
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
