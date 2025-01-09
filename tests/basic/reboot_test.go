package tests

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/libopenstorage/openstorage/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/portworx/sched-ops/task"
	"github.com/pure-px/torpedo/drivers/node"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/drivers/volume"
	"github.com/pure-px/torpedo/pkg/log"
	"github.com/pure-px/torpedo/pkg/testrailuttils"
	. "github.com/pure-px/torpedo/tests"
)

const (
	defaultWaitRebootTimeout     = 5 * time.Minute
	defaultWaitRebootRetry       = 10 * time.Second
	defaultCommandRetry          = 5 * time.Second
	defaultCommandTimeout        = 1 * time.Minute
	defaultTestConnectionTimeout = 15 * time.Minute
	defaultRebootTimeRange       = 5 * time.Minute
)

var (
	volMountedMap = make(map[string][]*volume.Volume)
	RebootTime    = time.Time{}
)

func updateSharedV4Map(context *scheduler.Context) error {
	listOfVolumes, err := Inst().S.GetVolumes(context)
	if err != nil {
		return err
	}
	// Need to reset the map to get new values
	volMountedMap = map[string][]*volume.Volume{}
	for _, vol := range listOfVolumes {
		appVol, _ := Inst().V.InspectVolume(vol.ID)
		if appVol.Spec.Sharedv4 && appVol.Spec.Sharedv4ServiceSpec != nil && appVol.Spec.Sharedv4ServiceSpec.Type == 2 {
			volMountedMap[appVol.AttachedOn] = append(volMountedMap[appVol.AttachedOn], vol)
		}
	}
	return nil
}

func nodeContainsSharedv4Vol(nodeIp string) bool {
	log.Info("Checking if %s has NFS volume", nodeIp)
	_, ok := volMountedMap[nodeIp]
	return ok
}

func validateVolumeTime(volumes []*volume.Volume) error {
	for _, vol := range volumes {
		log.Info("Validating %s volume now", vol.ID)
		appVol, err := Inst().V.InspectVolume(vol.ID)
		if err != nil {
			return err
		}
		timeDifference := appVol.AttachTime.AsTime().Sub(RebootTime).Seconds()
		dash.VerifySafely(timeDifference < 120, true, fmt.Sprintf("%s volume took %.0f secs to move", vol.ID, timeDifference))
	}
	return nil
}

var _ = Describe("{RebootOneNode}", Label("p1", "negative", "node_ops", "error_injection", "node_reboot"), func() {
	var testrailID = 35266
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/35266
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("RebootOneNode", "Reboot one storage node at a time and validate apps and px", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	It("has to schedule apps and reboot node(s) with volumes", func() {
		log.InfoD("has to schedule apps and reboot node(s) with volumes")
		contexts = make([]*scheduler.Context, 0)

		log.InfoD("Scheduling Applications")

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("rebootonenode-%d", i))...)
		}

		ValidateApplications(contexts)

		Step("get all nodes and reboot one by one", func() {
			log.InfoD("get all nodes and reboot one by one")
			nodesToReboot := node.GetStorageDriverNodes()
			// Reboot node and check driver status
			Step(fmt.Sprintf("reboot node one at a time from the node(s): %v", nodesToReboot), func() {
				log.InfoD("reboot node one at a time from the node(s): %v", nodesToReboot)
				for _, n := range nodesToReboot {
					// Skip rebooting master nodes, even if PX is installed on them to avoid torpedo pod crashing
					if node.IsMasterNode(n) {
						log.InfoD("This node [%s] is master, will skip rebooting it..", n.Name)
						continue
					}
					err := updateSharedV4Map(contexts[0])
					if err != nil {
						dash.VerifyFatal(err, nil, fmt.Sprintf("Error getting sharedv4 svc vols %v", err))
					}
					if n.IsStorageDriverInstalled {
						Step(fmt.Sprintf("reboot node: %s", n.Name), func() {
							log.InfoD("reboot node: %s", n.Name)
							rebootTime, err := Inst().N.RunCommand(n, "date -u +'%a %b %d %T UTC %Y'", node.ConnectionOpts{
								Timeout:         defaultCommandTimeout,
								TimeBeforeRetry: defaultCommandRetry,
								IgnoreError:     false,
								Sudo:            false,
							})
							rebootString := strings.TrimSpace(string(rebootTime))
							dash.VerifyFatal(err, nil, fmt.Sprintf("Got date from node: %s. Err: %v", n.Name, err))
							RebootTime, err = time.Parse(time.UnixDate, rebootString)
							err = Inst().N.RebootNode(n, node.RebootNodeOpts{
								Force: true,
								ConnectionOpts: node.ConnectionOpts{
									Timeout:         defaultCommandTimeout,
									TimeBeforeRetry: defaultCommandRetry,
								},
							})
							dash.VerifyFatal(err, nil, fmt.Sprintf("Reboot node %s successful? Err: %v", n.Name, err))
						})

						Step(fmt.Sprintf("wait for node: %s to be back up", n.Name), func() {
							log.InfoD("wait for node: %s to be back up", n.Name)
							err = Inst().N.TestConnection(n, node.ConnectionOpts{
								Timeout:         defaultTestConnectionTimeout,
								TimeBeforeRetry: defaultWaitRebootRetry,
							})
							dash.VerifyFatal(err, nil, fmt.Sprintf("node %s is up. Err: %v", n.Name, err))
						})

						Step(fmt.Sprintf("Check if node: %s rebooted in last 3 minutes", n.Name), func() {
							log.InfoD("Check if node: %s rebooted in last 3 minutes", n.Name)
							isNodeRebootedAndUp, err := Inst().N.IsNodeRebootedInGivenTimeRange(n, defaultRebootTimeRange)
							log.FailOnError(err, "Check for node: %s rebooted in last 3 minutes", n.Name)
							if !isNodeRebootedAndUp {
								Step(fmt.Sprintf("wait for volume driver to stop on node: %v", n.Name), func() {
									log.InfoD("wait for volume driver to stop on node: %v", n.Name)
									err := Inst().V.WaitDriverDownOnNode(n)
									dash.VerifyFatal(err, nil, fmt.Sprintf("node %s is PX stopped ? Err: %v", n.Name, err))
								})
							}
						})

						Step(fmt.Sprintf("wait to scheduler: %s and volume driver: %s to start",
							Inst().S.String(), Inst().V.String()), func() {
							log.InfoD("wait to scheduler: %s and volume driver: %s to start",
								Inst().S.String(), Inst().V.String())

							err = Inst().S.IsNodeReady(n)
							dash.VerifyFatal(err == nil, true, fmt.Sprintf("node %s is ready? Err: %v", n.Name, err))
							err = Inst().V.WaitDriverUpOnNode(n, Inst().DriverStartTimeout)
							dash.VerifyFatal(err == nil, true, fmt.Sprintf("node %s volume driver is up ? Err: %v", n.Name, err))

						})

						Step(fmt.Sprintf("Validate volume got migrated on time"), func() {
							// If the volume was mounted on this node
							if nodeContainsSharedv4Vol(n.MgmtIp) {
								log.InfoD("Checking node %s to see volume got migrated in <2 mins", n.MgmtIp)
								volumesToCheck := volMountedMap[n.MgmtIp]
								err := validateVolumeTime(volumesToCheck)
								dash.VerifySafely(err == nil, true, "Volumes got migrated on time?")
							} else {
								log.InfoD("Skipping the volume check for node %s since it does not contain shared V4 vol", n.MgmtIp)
							}
						})

						Step(fmt.Sprintf("validate apps"), func() {
							log.InfoD("Validate Apps")
							for _, ctx := range contexts {
								ValidateContext(ctx)
							}
						})
					}
				}
			})
		})

		Step(fmt.Sprintf("Destroying apps"), func() {
			log.InfoD("Destroying Apps")
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

var _ = Describe("{ReallocateSharedMount}", Label("p1", "negative", "node_ops", "error_injection", "node_reboot"), func() {

	var testrailID = 58844
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/58844
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("ReallocateSharedMount", "Validating Px and apps after reallocating shared mounts", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	It("has to schedule apps and reboot node(s) with shared volume mounts", func() {
		log.InfoD("has to schedule apps and reboot node(s) with shared volume mounts")

		//var err error
		contexts = make([]*scheduler.Context, 0)
		log.InfoD("Scheduling Applications")

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("reallocate-mount-%d", i))...)
		}

		log.InfoD("Validating Applications")

		ValidateApplications(contexts)

		Step(fmt.Sprintf("get nodes with shared mount and reboot them"), func() {
			log.InfoD("get nodes with shared mount and reboot them")
			for _, ctx := range contexts {
				vols, err := Inst().S.GetVolumes(ctx)
				Expect(err).NotTo(HaveOccurred())
				for _, vol := range vols {
					if vol.Shared {

						n, err := Inst().V.GetNodeForVolume(vol, defaultCommandTimeout, defaultCommandRetry)
						log.FailOnError(err, "Failed to get node for volume: %s", vol.ID)

						log.InfoD("volume %s is attached on node %s [%s]", vol.ID, n.SchedulerNodeName, n.Addresses[0])

						// Workaround to avoid PWX-24277 for now.
						Step(fmt.Sprintf("wait until volume %v status is Up", vol.ID), func() {
							log.InfoD("wait until volume %v status is Up", vol.ID)
							prevStatus := ""
							Eventually(func() (string, error) {
								connOpts := node.ConnectionOpts{
									Timeout:         defaultCommandTimeout,
									TimeBeforeRetry: defaultCommandRetry,
									Sudo:            true,
								}
								cmd := fmt.Sprintf("pxctl volume inspect %s | grep \"Replication Status\"", vol.ID)
								volStatus, err := Inst().N.RunCommandWithNoRetry(*n, cmd, connOpts)
								if err != nil {
									log.Warnf("failed to get replication state of volume %v: %v", vol.ID, err)
									return "", err
								}
								if volStatus != prevStatus {
									log.InfoD("volume %v: %v", vol.ID, volStatus)
									prevStatus = volStatus
								}
								return volStatus, nil
							}, 30*time.Minute, 10*time.Second).Should(ContainSubstring("Up"),
								"volume %v status is not Up for app %v", vol.ID, ctx.App.Key)
						})

						err = Inst().S.DisableSchedulingOnNode(*n)
						dash.VerifyFatal(err == nil, true, fmt.Sprintf("Disable sceduling on node : %s", n.Name))

						err = Inst().V.StopDriver([]node.Node{*n}, false, nil)
						dash.VerifyFatal(err == nil, true, fmt.Sprintf("Stop volume driver on node : %s success ?", n.Name))

						err = Inst().N.RebootNode(*n, node.RebootNodeOpts{
							Force: true,
							ConnectionOpts: node.ConnectionOpts{
								Timeout:         defaultCommandTimeout,
								TimeBeforeRetry: defaultCommandRetry,
							},
						})
						log.FailOnError(err, "Failed Rebooting node : %s", n.Name)

						// as we keep the storage driver down on node until we check if the volume, we wait a minute for
						// reboot to occur then we force driver to refresh endpoint to pick another storage node which is up
						log.InfoD("wait for %v for node reboot", defaultCommandTimeout)
						time.Sleep(defaultCommandTimeout)

						// Start NFS server to avoid pods stuck in terminating state (PWX-24274)
						err = Inst().N.Systemctl(*n, "nfs-server.service", node.SystemctlOpts{
							Action: "start",
							ConnectionOpts: node.ConnectionOpts{
								Timeout:         5 * time.Minute,
								TimeBeforeRetry: 10 * time.Second,
							}})
						log.FailOnError(err, "Failed starting nfs service on node : %s", n.Name)

						ctx.RefreshStorageEndpoint = true
						n2, err := Inst().V.GetNodeForVolume(vol, defaultCommandTimeout, defaultCommandRetry)
						log.FailOnError(err, "Failed to get node for volume : %s", vol.ID)

						// the mount should move to another node otherwise fail
						log.InfoD("volume %s is now attached on node %s [%s]", vol.ID, n2.SchedulerNodeName, n2.Addresses[0])
						dash.VerifyFatal(n.SchedulerNodeName != n2.SchedulerNodeName, true, "Volume is scheduled on different nodes?")

						StartVolDriverAndWait([]node.Node{*n})
						log.FailOnError(err, "Failed to Start volume driver on node: %s success ?", n.Name)
						err = Inst().S.EnableSchedulingOnNode(*n)
						log.FailOnError(err, "Failed to Enable scheduling on node: %s", n.Name)

						log.InfoD("validating applications")
						ValidateApplications(contexts)
					}
				}
			}
		})

		Step(fmt.Sprintf("Destroy apps"), func() {
			log.InfoD("Destroy apps")
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

var _ = Describe("{ClusterPxRestart}", Label("p1", "negative", "node_ops", "error_injection", "px_ops", "multiple_px_crash"), func() {
	var testrailID = 60042
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/60042
	var runID int

	JustBeforeEach(func() {
		StartTorpedoTest("ClusterPxRestart", "Validate restart of PX in cluster", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	var contexts []*scheduler.Context

	stepLog := "has to restart Px cluster "
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("restartpx-%d", i))...)
		}

		ValidateApplications(contexts)
		stepLog = "get all nodes which has PX installed and restart PX one by one"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			Step("get all PX nodes and restart PX one by one", func() {
				nodesToPXRestart := node.GetStorageDriverNodes()
				loc, _ := time.LoadLocation("UTC")
				t1 := time.Now().In(loc)

				// Reboot node and check driver status
				Step(fmt.Sprintf("restart px one at a time on node(s): %v", nodesToPXRestart), func() {
					for _, n := range nodesToPXRestart {
						if n.IsStorageDriverInstalled {
							Step(fmt.Sprintf("node with Px restart is: %s", n.Name), func() {

								err := Inst().V.RestartDriver(n, nil)
								log.FailOnError(err, fmt.Sprintf("Error occured while Restart PX on node:%v", n.Name))
							})

							Step(fmt.Sprintf("wait for volume driver to restart on node: %v", n.Name), func() {
								err := Inst().V.WaitForPxPodsToBeUp(n)
								log.FailOnError(err, fmt.Sprintf("Error occured while Validating PX restart is done on node:%v", n.Name))
							})

						}
					}
					t2 := time.Now()
					Step("validate apps", func() {
						for _, ctx := range contexts {
							ValidateContext(ctx)
						}
					})
					log.InfoD(fmt.Sprintf("Time taken to reboot all the PX is: %fsecs", t2.Sub(t1).Seconds()))
					dash.VerifyFatal((t2.Sub(t1).Seconds()) <= (float64(len(nodesToPXRestart)*240)), true, fmt.Sprintf("Time taken to reboot all the PX is: %fsecs which is less than the timeout of %fsecs", t2.Sub(t1).Seconds(), float64(len(nodesToPXRestart)*240)))
				})
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

var _ = Describe("{NodeWipeWithNodeReboot}", Label("p1", "negative", "node_ops", "px_ops", "error_injection", "multiple_node_reboot"), func() {

	/*
	   1.  Create volume and do IOs / deploy apps to do IOs
	   2.  Validate and destroy app
	   3.  Prepare node for decommission
	   4.  Decommission node
	   5.  Wipe node
	   6.  Reboot Node with some random Delay
	   7.  After node comes up run the node wipe again and verify it
	   8.  Rejoin node
	   9.  Verify node rejoin
	*/

	var contexts []*scheduler.Context
	BeforeEach(func() {
		StartTorpedoTest("NodeWipeWithNodeReboot", "node wipe with node reboot", nil, 0)
	})

	ItLog := "NodeWipeWithNodeReboot"
	It(ItLog, func() {
		stepLog := "Schedule apps"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			contexts = scheduleApps()
			log.Info("schedule app succeed")
			time.Sleep(5 * time.Minute)
		})

		// appsValidateAndDestroy this will delete all the app including the app we install while setting-up the job
		stepLog = "validate and destroy app"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			appsValidateAndDestroy(contexts)
			time.Sleep(time.Minute)
			log.Info("app validate and destroy succeed")
		})

		workerNodes := node.GetStorageDriverNodes()
		index := rand.Intn(len(workerNodes))
		selectedNode := workerNodes[index]

		stepLog = "Prepare node for decommission"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err := Inst().S.PrepareNodeToDecommission(selectedNode, Inst().Provisioner)
			log.FailOnError(err, fmt.Sprintf("Failed prepare node [%v] for decommission", selectedNode.Name))
			log.Info("prepare node for decommission succeed")
		})

		// decommission node
		stepLog = "Decommission node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = Inst().V.DecommissionNode(&selectedNode)
			log.FailOnError(err, fmt.Sprintf("Failed to decommission node [%v]", selectedNode.Name))
			log.Info("decommission node succeed")
		})
		time.Sleep(1 * time.Minute)

		stepLog = "Wipe node"
		Step(stepLog, func() {
			output, err := Inst().N.RunCommand(selectedNode, "echo Yes | pxctl sv node-wipe -s --all ", node.ConnectionOpts{
				Timeout:         2 * time.Minute,
				TimeBeforeRetry: 10 * time.Second,
			})
			log.FailOnError(err, "Failed to wipe node")
			dash.VerifyFatal(strings.Contains(output, "Wiped node successfully"), true, fmt.Sprintf("node [%s] is Wiped", selectedNode.Name))
			log.Info("Node wipe is succeed")
		})
		sleepTime := rand.Intn(100) + 1
		time.Sleep(time.Second * (time.Duration(sleepTime)))

		stepLog = fmt.Sprintf("Verify reboot after [%d] seconds", sleepTime)
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = Inst().N.RebootNodeAndWait(selectedNode)
			log.FailOnError(err, "Failed to reboot node and wait till it is up")
			log.Info("Verify reboot succeed")
		})

		//Wiped node
		stepLog = "Node wipe after reboot"
		Step(stepLog, func() {
			output, err := runPxctlCommand("sv node-wipe --all", selectedNode, nil)
			log.FailOnError(err, "Failed to wipe node")
			dash.VerifyFatal(strings.Contains(output, "Wiped node successfully"), true, fmt.Sprintf("node [%s] is Wiped", selectedNode.Name))
			log.Info("Node wipe is successfull")
		})

		stepLog = fmt.Sprintf("Rejoin node %s and verify", selectedNode.Name)
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = Inst().V.RejoinNode(&selectedNode)
			log.FailOnError(err, "Failed to rejoin node [%v]", selectedNode)
			log.Info("rejoin node successfull")

		})

		stepLog = "Verify node rejoin"
		Step(stepLog, func() {
			var rejoinedNode *api.StorageNode
			t := func() (interface{}, bool, error) {
				drvNodes, err := Inst().V.GetDriverNodes()
				if err != nil {
					return false, true, err
				}

				for _, n := range drvNodes {
					if n.Hostname == selectedNode.Hostname {
						rejoinedNode = n
						return true, false, nil
					}
				}

				return false, true, fmt.Errorf("node %s not joined yet", selectedNode.Name)
			}
			_, err = task.DoRetryWithTimeout(t, 20*time.Minute, defaultRetryInterval)
			log.FailOnError(err, fmt.Sprintf("error joining the node [%s]", selectedNode.Name))

			err = Inst().S.RefreshNodeRegistry()
			log.FailOnError(err, "error refreshing node registry")
			log.Info("refreshing node registry succeed")

			err = Inst().V.RefreshDriverEndpoints()
			log.FailOnError(err, "error refreshing storage drive endpoints")
			log.Info("refreshing storage drive endpoints succeed")

			selectedNode = node.Node{}
			for _, n := range node.GetStorageDriverNodes() {
				if n.Name == rejoinedNode.Hostname {
					selectedNode = n
					break
				}
			}
			if selectedNode.Name == "" {
				log.FailOnError(fmt.Errorf("rejoined node not found"), fmt.Sprintf("node [%s] not found in the node registry", rejoinedNode.Hostname))
			}
			log.Info("rejoined node found succesfully")

			err = Inst().V.WaitDriverUpOnNode(selectedNode, Inst().DriverStartTimeout)
			log.FailOnError(err, "error refreshing storage drive endpoints")

			log.Info("Validate driver up on rejoined node after rejoining succeed")

		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})
