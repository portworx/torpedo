package platform

import (
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	dsUtils "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs/dataservice"
	platformUtils "github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	"os"
	"testing"
)

var _ = BeforeSuite(func() {
	steplog := "Get prerequisite params to run platform tests"
	log.InfoD(steplog)
	Step(steplog, func() {
		log.InfoD("Get Account ID")
		err := platformUtils.InitUnifiedApiComponents(os.Getenv(envControlPlaneUrl), "")
		log.FailOnError(err, "error while initialising api components")
		accList, err := platformUtils.GetAccountListv1()
		log.FailOnError(err, "error while getting account list")
		accID := platformUtils.GetPlatformAccountID(accList, defaultTestAccount)
		log.Infof("AccountID - [%s]", accID)
		err = platformUtils.InitUnifiedApiComponents(os.Getenv(envControlPlaneUrl), accID)
		log.FailOnError(err, "error while initialising api components")

		//Initialising UnifiedApiComponents in ds utils
		err = dsUtils.InitUnifiedApiComponents(os.Getenv(envControlPlaneUrl), accID)
		log.FailOnError(err, "error while initialising api components in ds utils")
	})
})

var _ = AfterSuite(func() {
	log.InfoD("Test Finished")
})

func TestDataService(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_basic.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : platform", specReporters)

}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
