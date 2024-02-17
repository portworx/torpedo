package pds

import (
	. "github.com/onsi/ginkgo"
	dslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs/dataservice"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

var _ = Describe("{DeployDataservice}", func() {
	steplog := "Data service deployment"
	JustBeforeEach(func() {
		StartTorpedoTest("DeployDataService", "Deploy data services", nil, 0)
	})

	log.InfoD(steplog)
	It("Deploy and Validate Dataservice", func() {

		// TODO: Take the input json struct and pass on it to the workflows/lib func
		_, err := dslibs.DeployDataservice("ns", "newDeployment", "")
		log.FailOnError(err, "Error while deploying ds %v\n")
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
