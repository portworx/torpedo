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
	"strings"
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
	2. Reboot 2 node(s) for one day
	3. Validate applications
	4. Destroy applications
	*/

	JustBeforeEach(func() {
		StartTorpedoTest("NodeRebootForOneDay", "Reboot node(s) for one day", nil, 0)
	})

	itLog := "Reboot node(s) for one day"
	It(itLog, func() {
		log.InfoD(itLog)
		contexts := make([]*scheduler.Context, 0)
		stepLog := "schedule applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("noderebootoneday-%d", i))...)
			}
			ValidateApplications(contexts)
		})

		stepLog = "reboot 2 nodes in parallel for one day"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			nodesToReboot := node.GetStorageDriverNodes()[:2] // Get the first two nodes to reboot

			// Start a timer for 24 hours
			timer := time.NewTimer(24 * time.Hour)
			defer timer.Stop()

			semaphore := make(chan struct{}, 2)
			stopChan := make(chan struct{}) // Channel to signal goroutines to stop
			doneChan := make(chan struct{}) // Channel to signal that all work is done

			// Function to reboot a node and perform multipath consistency check
			rebootNode := func(nodeToReboot node.Node) {
				defer GinkgoRecover()
				defer func() {
					stopChan <- struct{}{} // Signal the goroutine to stop
				}()
				poolMultipathMap := make(map[string][]string, 0)
				firstReboot := true
				countBeforeReboot := 0

				for {
					select {
					case <-stopChan:
						return
					case <-timer.C:
						return
					default:
						// Acquire semaphore to control the parallel execution
						semaphore <- struct{}{}
						defer func() {
							<-semaphore // Release semaphore
						}()

						if IsPureCluster() {
							if firstReboot {
								cmd := "pxctl status | awk '/\\/dev\\//'"
								output, err := Inst().N.RunCommand(nodeToReboot, cmd, node.ConnectionOpts{
									Timeout:         60 * time.Minute,
									TimeBeforeRetry: 30 * time.Second,
								})
								log.FailOnError(err, "Error getting multipath status")
								lines := strings.Split(output, "\n")
								for _, line := range lines {
									deviceInfo := strings.Fields(line)
									if len(deviceInfo) > 1 {
										devicePathSize := []string{deviceInfo[1], deviceInfo[2]}
										poolMultipathMap[deviceInfo[0]] = devicePathSize
										countBeforeReboot++
									}
								}
								firstReboot = false
							}

							// Consistency check after reboot
							cmd := "pxctl status | awk '/\\/dev\\//'"
							output, err := Inst().N.RunCommand(nodeToReboot, cmd, node.ConnectionOpts{
								Timeout:         60 * time.Minute,
								TimeBeforeRetry: 30 * time.Second,
							})
							log.FailOnError(err, "Error getting multipath status")
							lines := strings.Split(output, "\n")
							countAfterReboot := 0
							for _, line := range lines {
								deviceInfo := strings.Fields(line)
								if len(deviceInfo) > 1 {
									deviceInfoAfterReboot := []string{deviceInfo[1], deviceInfo[2]}
									if poolMultipathMap[deviceInfo[0]][0] != deviceInfoAfterReboot[0] || poolMultipathMap[deviceInfo[0]][1] != deviceInfoAfterReboot[1] {
										log.InfoD("Device path before reboot")
										for key, value := range poolMultipathMap {
											log.InfoD("pool id: %s device path and size :[%s]", key, value)
										}
										log.InfoD("Device path size before reboot: %s %s", poolMultipathMap[deviceInfo[0]][0], poolMultipathMap[deviceInfo[0]][1])
										log.InfoD("Device path size after reboot: %s %s", deviceInfoAfterReboot[0], deviceInfoAfterReboot[1])
										log.FailOnError(fmt.Errorf("multipath consistency check failed"), "Multipath consistency check failed")
									}
									countAfterReboot++
								}
							}
							if countBeforeReboot != countAfterReboot {
								log.InfoD("There is a mismatch in the number of device paths before and after reboot")
								log.InfoD("Device path before reboot")
								for key, value := range poolMultipathMap {
									log.InfoD("pool id: %s device path and size :[%s]", key, value)
								}
								log.InfoD("Device path size before reboot: %d", countBeforeReboot)
								for _, line := range lines {
									deviceInfo := strings.Fields(line)
									if len(deviceInfo) > 1 {
										log.InfoD("Device path size after reboot: %s %s", deviceInfo[0], deviceInfo[1])
									}
								}
							}
						}

						// Reboot the node
						err := Inst().N.RebootNode(nodeToReboot, node.RebootNodeOpts{
							Force: false,
							ConnectionOpts: node.ConnectionOpts{
								Timeout:         60 * time.Minute,
								TimeBeforeRetry: 30 * time.Second,
							},
						})
						dash.VerifySafely(err, nil, "Validate node is rebooted")

						// Verify that the node is back online
						err = Inst().N.TestConnection(nodeToReboot, node.ConnectionOpts{
							Timeout:         60 * time.Minute,
							TimeBeforeRetry: 30 * time.Second,
						})
						dash.VerifyFatal(err, nil, "Validate node is back up")

						err = Inst().S.IsNodeReady(nodeToReboot)
						dash.VerifyFatal(err, nil, "Validate node is ready")

						err = Inst().V.WaitDriverUpOnNode(nodeToReboot, 60*time.Minute)
						dash.VerifyFatal(err, nil, "Validate volume driver is up")

						log.InfoD("Rebooted node %s", nodeToReboot.Name)
					}
				}
			}

			// Reboot nodes in parallel
			for _, nodeToReboot := range nodesToReboot {
				go rebootNode(nodeToReboot)
				time.Sleep(5 * time.Second) // Small delay between reboots to avoid contention
			}

			// Monitor the application state during reboots
			go func() {
				defer GinkgoRecover()
				for {
					select {
					case <-stopChan:
						return
					default:
						// Ensure that reboots are not happening in parallel
						semaphore <- struct{}{}
						semaphore <- struct{}{}
						ValidateApplications(contexts)
						time.Sleep(1 * time.Minute) // Polling interval
					}
				}
			}()

			// Ensure all nodes complete the reboot process or timer ends
			go func() {
				defer close(doneChan) // Signal that all work is done
				<-timer.C             // Wait for the timer to finish
				close(stopChan)       // Signal the goroutines to stop
			}()

			<-doneChan // Wait for all work to be completed

			// Final comprehensive validation after the 24-hour period
			ValidateApplications(contexts)
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})
