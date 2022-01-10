package tests

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/portworx/torpedo/pkg/testrailuttils"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	. "github.com/portworx/torpedo/tests"
	"github.com/sirupsen/logrus"
)

const (
	defaultWaitRebootTimeout     = 5 * time.Minute
	defaultWaitRebootRetry       = 10 * time.Second
	defaultCommandRetry          = 5 * time.Second
	defaultCommandTimeout        = 1 * time.Minute
	defaultTestConnectionTimeout = 15 * time.Minute
)

func TestNodeCrash(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_Reboot.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : Crash", specReporters)
}

var _ = BeforeSuite(func() {
	InitInstance()
})

var _ = Describe("{CrashOneNode}", func() {
	var testrailID = 35255
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/35255
	var runID int
	JustBeforeEach(func() {
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	var contexts []*scheduler.Context

	It("has to schedule apps and crash node(s) with volumes", func() {
		var err error
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("crashonenode-%d", i))...)
		}

		ValidateApplications(contexts)

		Step("get all nodes and crash one by one", func() {
			nodesToCrash := node.GetWorkerNodes()

			// Crash node and check driver status
			Step(fmt.Sprintf("crash node one at a time from the node(s): %v", nodesToCrash), func() {
				for _, n := range nodesToCrash {
					if n.IsStorageDriverInstalled {
						Step(fmt.Sprintf("crash node: %s", n.Name), func() {
							err = Inst().N.CrashNode(n, node.CrashNodeOpts{
								Force: true,
								ConnectionOpts: node.ConnectionOpts{
									Timeout:         defaultCommandTimeout,
									TimeBeforeRetry: defaultCommandRetry,
								},
							})
							Expect(err).NotTo(HaveOccurred())
						})

						Step(fmt.Sprintf("wait for node: %s to be back up", n.Name), func() {
							err = Inst().N.TestConnection(n, node.ConnectionOpts{
								Timeout:         defaultTestConnectionTimeout,
								TimeBeforeRetry: defaultWaitRebootRetry,
							})
							Expect(err).NotTo(HaveOccurred())
						})

						Step(fmt.Sprintf("wait for volume driver to stop on node: %v", n.Name), func() {
							err := Inst().V.WaitDriverDownOnNode(n)
							Expect(err).NotTo(HaveOccurred())
						})

						Step(fmt.Sprintf("wait to scheduler: %s and volume driver: %s to start",
							Inst().S.String(), Inst().V.String()), func() {

							err = Inst().S.IsNodeReady(n)
							Expect(err).NotTo(HaveOccurred())

							err = Inst().V.WaitDriverUpOnNode(n, Inst().DriverStartTimeout)
							Expect(err).NotTo(HaveOccurred())
						})

						Step("validate apps", func() {
							for _, ctx := range contexts {
								ValidateContext(ctx)
							}
						})
					}
				}
			})
		})

		Step("destroy apps", func() {
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		})
	})
	JustAfterEach(func() {
		AfterEachTest(contexts)
	})
})

var _ = Describe("{ReallocateSharedMount}", func() {
	var contexts []*scheduler.Context

	It("has to schedule apps and crash node(s) with shared volume mounts", func() {
		//var err error
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("reallocate-mount-%d", i))...)
		}

		ValidateApplications(contexts)

		Step("get nodes with shared mount and crash them", func() {
			for _, ctx := range contexts {
				vols, err := Inst().S.GetVolumes(ctx)
				Expect(err).NotTo(HaveOccurred())
				for _, vol := range vols {
					if vol.Shared {
						n, err := Inst().V.GetNodeForVolume(vol, defaultCommandTimeout, defaultCommandRetry)
						Expect(err).NotTo(HaveOccurred())
						logrus.Infof("volume %s is attached on node %s [%s]", vol.ID, n.SchedulerNodeName, n.Addresses[0])
						err = Inst().S.DisableSchedulingOnNode(*n)
						Expect(err).NotTo(HaveOccurred())
						err = Inst().V.StopDriver([]node.Node{*n}, false, nil)
						Expect(err).NotTo(HaveOccurred())
						err = Inst().N.CrashNode(*n, node.CrashNodeOpts{
							Force: true,
							ConnectionOpts: node.ConnectionOpts{
								Timeout:         defaultCommandTimeout,
								TimeBeforeRetry: defaultCommandRetry,
							},
						})

						// as we keep the storage driver down on node until we check if the volume, we wait a minute for
						// crash to occur then we force driver to refresh endpoint to pick another storage node which is up
						logrus.Info("wait for 1 minute for node reboot")
						time.Sleep(defaultCommandTimeout)
						ctx.RefreshStorageEndpoint = true
						ValidateContext(ctx)
						n2, err := Inst().V.GetNodeForVolume(vol, defaultCommandTimeout, defaultCommandRetry)
						Expect(err).NotTo(HaveOccurred())
						// the mount should move to another node otherwise fail
						Expect(n2.SchedulerNodeName).NotTo(Equal(n.SchedulerNodeName))
						logrus.Infof("volume %s is now attached on node %s [%s]", vol.ID, n2.SchedulerNodeName, n2.Addresses[0])
						StartVolDriverAndWait([]node.Node{*n})
						err = Inst().S.EnableSchedulingOnNode(*n)
						Expect(err).NotTo(HaveOccurred())
						ValidateApplications(contexts)
					}
				}
			}
		})

		Step("destroy apps", func() {
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		})
	})
	JustAfterEach(func() {
		AfterEachTest(contexts)
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
