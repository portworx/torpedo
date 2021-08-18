package tests

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/sirupsen/logrus"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	. "github.com/portworx/torpedo/tests"
	corev1 "k8s.io/api/core/v1"
)

func TestNFSv4(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_shared.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : Shared", specReporters)
}

var _ = BeforeSuite(func() {
	InitInstance()
})

var _ = Describe("{NFSv4NotSupported}", func() {
	var contexts []*scheduler.Context

	It("has to create sv4 svc volume with nfsv4, validate, and teardown apps", func() {
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("nfsv4notsupported-%d", i))...)
		}

		// validatePVCs will make sure they are in pending state.
		// sleep 30 seconds here, to make sure the pending state is not just transient
		time.Sleep(30 * time.Second)

		for _, ctx := range contexts {
			Step(fmt.Sprintf("validate: %s's pvcs", ctx.App.Key), func() {
				validatePVCs(ctx)
			})
		}

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

func validatePVCs(ctx *scheduler.Context) {
	pvcs, err := Inst().S.GetPVCs(ctx)
	if err != nil {
		logrus.Infof("get pvc error %v", err)
	}
	Expect(len(pvcs)).To(Equal(3), "There should be 3 PVCs")
	for _, pvc := range pvcs {
		Expect(pvc.Phase).To(Equal(string(corev1.ClaimPending)), fmt.Sprintf("pvc %v should be in pending phase", pvc.Name))
	}
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
