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

var _ = Describe("{AccountsCRUD}", func() {
	steplog := "Accounts CRUD"
	JustBeforeEach(func() {
		StartTorpedoTest("ListAccounts", "validate dns endpoitns", nil, 0)
		err := pdslib.InitUnifiedApiComponents(os.Getenv("CONTROL_PLANE_URL"))
		log.FailOnError(err, "error while initialising api components")
	})

	Step(steplog, func() {
		log.InfoD(steplog)
		It("Accounts", func() {
			Step("create accounts", func() {
				acc, err := pdslib.CreateAccountV2("test-account", "qa-test-automation-account", "marunachalam+2@purestorage.com")
				log.FailOnError(err, "error while creating account")
				log.Infof("created account with name %s", *acc.Meta.Name)
			})
			steplog = "ListAccounts"
			Step(steplog, func() {
				log.InfoD(steplog)
				accList, err := pdslib.GetAccountListV2()
				log.FailOnError(err, "error while getting account list")
				for _, acc := range accList {
					log.Infof("Available account %s", *acc.Meta.Name)
					log.Infof("Available account ID %s", *acc.Meta.Uid)
				}
			})
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = BeforeSuite(func() {
	Step("Test", func() {
		log.InfoD("Just a test")
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
