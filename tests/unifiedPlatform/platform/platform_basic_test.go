package platform

import (
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	pdslib "github.com/portworx/torpedo/drivers/pds/lib"
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
		err := pdslib.InitUnifiedApiComponents(os.Getenv("CONTROL_PLANE_URL"), "")
		log.FailOnError(err, "error while initialising api components")
		accList, err := pdslib.GetAccountListV2()
		log.FailOnError(err, "error while getting account list")
		accID := pdslib.GetPlatformAccountID(accList, "demo-milestone-one")
		err = pdslib.InitUnifiedApiComponents(os.Getenv("CONTROL_PLANE_URL"), accID)
		log.FailOnError(err, "error while initialising api components")
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
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : pds", specReporters)

}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
