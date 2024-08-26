package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/testrailuttils"
	. "github.com/portworx/torpedo/tests"
	"time"
)

var _ = Describe("{CrashOneNode}", func() {
	var testrailID = 35255
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/35255
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("CrashOneNode", "Validate Crash one node", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	var contexts []*scheduler.Context

	stepLog := "has to schedule apps and crash node(s) with volumes"
	It(stepLog, func() {
		log.InfoD(stepLog)
		var err error
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("crashonenode-%d", i))...)
		}

		ValidateApplications(contexts)
		stepLog = "get all nodes and crash one by one"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			nodesToCrash := node.GetStorageDriverNodes()

			// Crash node and check driver status
			stepLog = fmt.Sprintf("crash node one at a time from the pxnode(s)")
			Step(stepLog, func() {
				log.InfoD(stepLog)
				for _, n := range nodesToCrash {
					if n.IsStorageDriverInstalled {
						stepLog = fmt.Sprintf("crash node: %s", n.Name)
						Step(stepLog, func() {
							log.InfoD(stepLog)
							err = Inst().N.CrashNode(n, node.CrashNodeOpts{
								Force: true,
								ConnectionOpts: node.ConnectionOpts{
									Timeout:         defaultCommandTimeout,
									TimeBeforeRetry: defaultCommandRetry,
								},
							})
							dash.VerifySafely(err, nil, "Validate node is crashed")

						})

						stepLog = fmt.Sprintf("wait for node: %s to be back up", n.Name)
						Step(stepLog, func() {
							log.InfoD(stepLog)
							err = Inst().N.TestConnection(n, node.ConnectionOpts{
								Timeout:         defaultTestConnectionTimeout,
								TimeBeforeRetry: defaultWaitRebootRetry,
							})
							dash.VerifyFatal(err, nil, "Validate node is back up")
						})

						stepLog = fmt.Sprintf("wait to scheduler: %s and volume driver: %s to start",
							Inst().S.String(), Inst().V.String())
						Step(stepLog, func() {
							log.InfoD(stepLog)
							err = Inst().S.IsNodeReady(n)
							dash.VerifyFatal(err, nil, "Validate node is ready")
							err = Inst().V.WaitDriverUpOnNode(n, Inst().DriverStartTimeout)
							dash.VerifyFatal(err, nil, "Validate volume is driver up")
						})

						Step("validate apps", func() {
							for _, ctx := range contexts {
								ValidateContext(ctx)
							}
						})
					}
				}
				err = ValidateDataIntegrity(&contexts)
				log.FailOnError(err, "error validating data integrity")
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
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{NodeRebootForOneDay}", func() {
	/* https://purestorage.atlassian.net/browse/PTX-25705
	1. Schedule applications
	2. Reboot node(s) for one day
	3. Validate applications
	4. Destroy applications
	*/

	JustBeforeEach(func() {
		StartTorpedoTest("NodeRebootForOneDay", "Reboot node(s) for one day", nil, 0)
	})

	itLog := "Reboot node(s) for one day"
	It(itLog, func() {
		log.InfoD(itLog)
		var err error
		var rebootsDone = 0
		// Time since test started
	        startTime := time.Now()
		
		contexts := make([]*scheduler.Context, 0)
		stepLog := "schedule applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("noderebootoneday-%d", i))...)
			}
			ValidateApplications(contexts)
		})

		stepLog = "reboot node(s) for one day"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			nodeToReboot := node.GetStorageDriverNodes()[0]
			// Start a timer for 24 hours
			timer := time.NewTimer(24 * time.Hour)
			go func() {
				<-timer.C
				nodeToReboot.IsStorageDriverInstalled = false
			}()
			for {
				select {
				case <-timer.C:
					break
				default:
					{
						err = Inst().N.RebootNode(nodeToReboot, node.RebootNodeOpts{
							Force: false,
							ConnectionOpts: node.ConnectionOpts{
								Timeout:         defaultCommandTimeout,
								TimeBeforeRetry: defaultCommandRetry,
							},
						})
						dash.VerifySafely(err, nil, "Validate node is rebooted")
						// Wait for node to be back up
						err = Inst().N.TestConnection(nodeToReboot, node.ConnectionOpts{
							Timeout:         defaultTestConnectionTimeout,
							TimeBeforeRetry: defaultWaitRebootRetry,
						})

						dash.VerifyFatal(err, nil, "Validate node is back up")
						// Wait for scheduler and volume driver to start
						err = Inst().S.IsNodeReady(nodeToReboot)
						dash.VerifyFatal(err, nil, "Validate node is ready")
						err = Inst().V.WaitDriverUpOnNode(nodeToReboot, Inst().DriverStartTimeout)
						dash.VerifyFatal(err, nil, "Validate volume is driver up")

						ValidateApplications(contexts) // Validate applications
						log.InfoD("Rebooted node %s %d times", nodeToReboot.Name, rebootsDone)
						log.InfoD("Time since test started: %s", time.Since(startTime))
						rebootsDone++
						log.InfoD("Number of times node has been rebooted: %d", rebootsDone)
					}
				}
			}
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})

})
