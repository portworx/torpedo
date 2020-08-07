package tests

import (
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/portworx/torpedo/drivers/scheduler"
	. "github.com/portworx/torpedo/tests"
	"github.com/sirupsen/logrus"
)

const (
	defaultTimeout       = 5 * time.Minute
	defaultRetryInterval = 20 * time.Second
	appReadinessTimeout  = 20 * time.Minute
)

func TestPxcentral(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_basic.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : pxcentral", specReporters)
}

var _ = BeforeSuite(func() {
	logrus.Infof("Init instance")
	InitInstance()
})

// This test performs basic test of installing px-central with helm
var _ = Describe("{Installpxcentral}", func() {
	It("has to setup, validate and teardown apps", func() {

		// Install the storage class
		appName := "pxcentral"
		contexts := ScheduleApplications(appName)

		// Skipping volume validation until other volume providers are implemented.
		// Also change the app readinessTimeout to 20mins
		for _, ctx := range contexts {
			ctx.SkipVolumeValidation = true
			ctx.ReadinessTimeout = appReadinessTimeout
		}

		ValidateApplications(contexts)
		logrus.Infof("Successfully validated specs for pxcentral app")

		Step("destroy apps", func() {
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		})
		logrus.Infof("Successfully destroyed pxcentral app")
	})
})

var _ = AfterSuite(func() {
	PerformSystemCheck()
	ValidateCleanup()
})

func TestMain(m *testing.M) {
	ParseFlags()
	os.Exit(m.Run())
}
