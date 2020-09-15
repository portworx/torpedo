package tests

import (
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
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
		ValidateApplications(contexts)

		// Install px-central through helm
		helmcontexts := ScheduleHelmApplications(appName)

		// Skipping volume validation until other volume providers are implemented.
		// Also change the app readinessTimeout to 20mins
		for _, ctx := range helmcontexts {
			ctx.SkipVolumeValidation = true
			ctx.ReadinessTimeout = appReadinessTimeout
		}

		ValidateApplications(helmcontexts)

		DeleteHelmApplications(helmcontexts[0].HelmRepo)

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
