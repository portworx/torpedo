package tests

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/testrailuttils"
	. "github.com/portworx/torpedo/tests"
	"github.com/sirupsen/logrus"
)

func TestBasic(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_ipv6.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : Ipv6", specReporters)
}

var _ = BeforeSuite(func() {
	InitInstance()
})

// Simple single test
var _ = Describe("{HelloTest}", func() {
	// update the testrailID with your test
	var testrailID = 0
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/35258
	var runID int
	JustBeforeEach(func() {
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	It("has to setup, validate and teardown apps", func() {
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("hellotest-%d", i))...)
		}

		ValidateApplications(contexts)

		// for _, ctx := range contexts {
		// 	Step(fmt.Sprintf("sample step", ctx.App.Key), func() {
		// 		// commands here
		//		// common checks:
		//		Expect(err).NotTo(HaveOccurred())
		//		ValidateContext(ctx)
		// 	})
		// }

		// cleanup
		for _, ctx := range contexts {
			TearDownContext(ctx, map[string]bool{scheduler.OptionsWaitForResourceLeakCleanup: true})
		}
	})
	JustAfterEach(func() {
		AfterEachTest(contexts, testrailID, runID)
	})
})

// Group test that share the set up or test similar functionality
// this way reduces duplicate code. Refer to basic/sharedv4_svc_test.go for similar structure.
var _ = Describe("{HelloGroupTest}", func() {
	var testrailID, runID int
	var contexts []*scheduler.Context
	var namespacePrefix string

	JustBeforeEach(func() {
		runID = testrailuttils.AddRunsToMilestone(testrailID)

		// Set up all apps
		contexts = nil
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("%s-%d", namespacePrefix, i))...)
		}

		ValidateApplications(contexts)
	})

	// Template for additional context
	// Context("{}", func() {
	// 	BeforeEach(func() {
	// 		testrailID = 0
	//      namespacePrefix = ""
	// 	})
	//
	// 	JustBeforeEach(func() {
	//		// since the apps are deployed by JustBeforeEach in the outer block,
	// 		// any test-specific changes to the deployed apps should go here.
	// 	})
	//
	// 	It("", func() {
	// 		for _, ctx := range contexts {
	//		// Step("", func() {
	//		// })
	// 		}
	// 	})

	// 	AfterEach(func() {
	// 	})
	// })

	AfterEach(func() {
		Step("destroy apps", func() {
			if CurrentGinkgoTestDescription().Failed {
				logrus.Info("not destroying apps because the test failed\n")
				return
			}
			for _, ctx := range contexts {
				TearDownContext(ctx, map[string]bool{scheduler.OptionsWaitForResourceLeakCleanup: true})
			}
		})
	})

	JustAfterEach(func() {
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = AfterSuite(func() {
	PerformSystemCheck()
	ValidateCleanup()
})

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
