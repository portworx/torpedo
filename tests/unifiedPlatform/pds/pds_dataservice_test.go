package pds

//
//import (
//	. "github.com/onsi/ginkgo"
//	dslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs/dataservice"
//	"github.com/portworx/torpedo/pkg/log"
//	. "github.com/portworx/torpedo/tests"
//	platformTests "github.com/portworx/torpedo/tests/unifiedPlatform/platform"
//)
//
//var _ = Describe("{DeployDataServicesOnDemand}", func() {
//	steplog := "Data service deployment"
//	JustBeforeEach(func() {
//		StartTorpedoTest("DeployDataService", "Deploy data services", nil, 0)
//	})
//
//	log.InfoD(steplog)
//	It("Deploy and Validate DataService", func() {
//		for _, ds := range platformTests.Params.DataServiceToTest {
//			_, err := dslibs.DeployDataService(ds)
//			log.FailOnError(err, "Error while deploying ds %v\n")
//		}
//	})
//
//	JustAfterEach(func() {
//		defer EndTorpedoTest()
//	})
//})
