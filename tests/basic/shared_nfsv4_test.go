package tests

import (
	"fmt"
	"time"

	"github.com/portworx/torpedo/drivers/scheduler"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/portworx/torpedo/tests"
)

var _ = Describe("{NFSv4NotSupported}", func() {
	var contexts []*scheduler.Context

	It("validates that NFSv4 is not allowed with sharedv4 service", func() {
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("nfsv4notsupported-%d", i))...)
		}

		// sleep a bit to make sure that the state is not just transient
		time.Sleep(30 * time.Second)
		for _, ctx := range contexts {
			Step(fmt.Sprintf("verifying that no volumes were created for %s", ctx.App.Key), func() {
				validateNoVolumes(ctx)
			})
		}

		for _, ctx := range contexts {
			TearDownContext(ctx, map[string]bool{scheduler.OptionsWaitForResourceLeakCleanup: true})
		}
	})

	JustAfterEach(func() {
		AfterEachTest(contexts)
	})
})

func validateNoVolumes(ctx *scheduler.Context) {
	vols, err := Inst().S.GetVolumes(ctx)
	Expect(err).ShouldNot(HaveOccurred())
	for _, vol := range vols {
		// GetVolumes() actually returns all PVCs (whether bound or not) with vol.ID set to the PV name.
		// Fail only if vol.ID is not empty.
		if vol.ID != "" {
			Fail(fmt.Sprintf("Sharedv4 service volume %v for PVC %v/%v should not have been created with NFSv4",
				vol.ID, vol.Namespace, vol.Name))
		}
	}
}
