package tests

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/scheduler"

	. "github.com/onsi/ginkgo"
	. "github.com/portworx/torpedo/tests"
)

var _ = Describe("{NFSServerFailover}", func() {
	var contexts []*scheduler.Context

	It("has to setup, validate, fail nfs server, validate, and teardown apps", func() {
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("setupteardown-%d", i))...)
		}

		ValidateApplications(contexts)

		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

		for _, ctx := range contexts {
			TearDownContext(ctx, opts)
		}
	})
	JustAfterEach(func() {
		AfterEachTest(contexts)
	})
})