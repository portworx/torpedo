package tests

import (
	"fmt"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/portworx/torpedo/drivers/scheduler"
	. "github.com/portworx/torpedo/tests"
)

var (
	testName      = "AutopilotPoolExpand"
	timeout       = 30 * time.Minute
	retryInterval = 30 * time.Second
)

func TestAutopilotPoolExpand(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_autopilot_pool_expand.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : AutopilotPoolExpand", specReporters)
}

var _ = BeforeSuite(func() {
	InitInstance()
})

// This test performs basic test of starting an application, fills up the volume with data
// which is more than the size of pool and expects the pool to be resized.
var _ = Describe(fmt.Sprintf("{%s}", testName), func() {
	It("has to fill up the storage pool, resize it, validate and teardown apps", func() {
		var contexts []*scheduler.Context
		var err error
		for i := 0; i < Inst().ScaleFactor; i++ {
			Step("schedule applications", func() {
				taskName := fmt.Sprintf("%s-%v", fmt.Sprintf("%s-%d", strings.ToLower(testName), i), Inst().InstanceID)
				contexts, err = Inst().S.Schedule(taskName, scheduler.ScheduleOptions{
					AppKeys:            Inst().AppList,
					StorageProvisioner: Inst().Provisioner,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(contexts).NotTo(BeEmpty())
			})
		}

		for _, ctx := range contexts {
			Step("wait until workload completes on pool", func() {
				err = Inst().S.WaitForRunning(ctx, timeout, retryInterval)
				Expect(err).NotTo(HaveOccurred())
			})
		}

		Step("validating pool and verifying size of it", func() {
			err = Inst().V.ValidateStorage(timeout, retryInterval)
			Expect(err).NotTo(HaveOccurred())
		})

		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		ValidateAndDestroy(contexts, opts)
	})
})

var _ = AfterSuite(func() {
	PerformSystemCheck()
	ValidateCleanup()
})

func init() {
	ParseFlags()
}
