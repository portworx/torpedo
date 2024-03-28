package tests

import (
	"fmt"
	//	"math/rand"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	//	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	. "github.com/portworx/torpedo/tests"
	"github.com/sirupsen/logrus"
)

const (
	defaultCommandRetry   = 5 * time.Second
	defaultCommandTimeout = 1 * time.Minute
)

func TestBasic(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_basic.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : Sharedv4-svc", specReporters)
}

var _ = BeforeSuite(func() {
	InitInstance()
})

// App using repl-3 sharedv4 service volume and volume is in degraded state, should still work with only one functioning replica
var _ = Describe("{DegradeVolumesFailover}", func() {
	var contexts []*scheduler.Context

	It("has to setup apps with a repl-3 sharedv4 service vol and have 2 of the replicas in degraded state", func() {
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("sharedv4-svc-degrade-%d", i))...)
		}

		ValidateApplications(contexts)

		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

		vols, err := Inst().S.GetVolumes(contexts[0])
		Expect(err).NotTo(HaveOccurred())

		for _, vol := range vols {
			currRep, err := Inst().V.GetReplicationFactor(vol)
			Expect(err).NotTo(HaveOccurred())
			logrus.Infof("volume %s has replication factor: %d", vol.ID, currRep)
			n, err := Inst().V.GetNodeForVolume(vol, defaultCommandTimeout, defaultCommandRetry)
			Expect(err).NotTo(HaveOccurred())
			logrus.Infof("volume %s is attached on node %s [%s]", vol.ID, n.SchedulerNodeName, n.Addresses[0])
			replicaSets, err := Inst().V.GetReplicaSets(vol)
			Expect(err).NotTo(HaveOccurred())
			Expect(replicaSets).NotTo(BeEmpty())
			//err = Inst().V.StopDriver([]node.Node{*n}, false, nil)
		}
		//for _, ctx := range contexts {
		//	TearDownContext(ctx, opts)
		//}
	})
	JustAfterEach(func() {
		AfterEachTest(contexts)
	})
})

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	ParseFlags()
	os.Exit(m.Run())
}
