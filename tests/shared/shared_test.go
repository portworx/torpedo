package tests

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/sirupsen/logrus"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/portworx/torpedo/tests"
)

var _ = Describe("{NFSServerFailover}", func() {
	var contexts []*scheduler.Context

	It("has to setup, validate, fail nfs server, validate, and teardown apps", func() {
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("setupteardown-%d", i))...)
		}
		logrus.Infof("testing ctx size %v", len(contexts))
		for _, ctx := range contexts {
			Step(fmt.Sprintf("get replica sets for app: %s's volumes", ctx.App.Key), func() {
				appVolumes, err := Inst().S.GetVolumes(ctx)
				logrus.Infof("testing appVolumes size %v", len(appVolumes))
				Expect(err).NotTo(HaveOccurred())
				Expect(appVolumes).NotTo(BeEmpty())
				for _, vol := range appVolumes {
					logrus.Infof("testing vol %v", vol)
					replicaSets, err := Inst().V.GetReplicaSets(vol)
					logrus.Infof("testing replicaSets size %v", len(replicaSets))
					logrus.Infof("testing nodes %v", replicaSets[0].Nodes)
					Expect(err).NotTo(HaveOccurred())
					Expect(replicaSets).NotTo(BeEmpty())
				}
			})

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