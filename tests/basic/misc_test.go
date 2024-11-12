package tests

import (
	ctxt "context"
	"fmt"

	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"
	storageApi "k8s.io/api/storage/v1"
	"k8s.io/utils/strings/slices"

	"math/rand"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/pure-px/torpedo/drivers/scheduler/k8s"
	"github.com/pure-px/torpedo/drivers/volume"
	"github.com/pure-px/torpedo/drivers/volume/portworx"
	"github.com/pure-px/torpedo/drivers/volume/portworx/schedops"
	corev1 "k8s.io/api/core/v1"

	"github.com/libopenstorage/openstorage/api"
	opsapi "github.com/libopenstorage/openstorage/api"
	"github.com/pure-px/torpedo/pkg/kvdbutils"
	"github.com/pure-px/torpedo/pkg/log"
	"github.com/pure-px/torpedo/pkg/osutils"
	"github.com/pure-px/torpedo/pkg/pureutils"

	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/sched-ops/task"
	"github.com/pure-px/torpedo/drivers/node"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/pkg/testrailuttils"
	. "github.com/pure-px/torpedo/tests"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var k8sCore = core.Instance()

// This test performs basic test of starting an application and destroying it (along with storage)
var _ = Describe("{SetupTeardown}", Label("p1", "positive", "px_vol_ops", "shared_v4"), func() {
	var testrailID = 35258
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/35258
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("SetupTeardown", "Validate setup tear down", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	var contexts []*scheduler.Context

	It("has to setup, validate and teardown apps", func() {
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
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

func pureWriteRoutine(ctx *scheduler.Context, podName string, dataDir string, shouldStop *bool, errOutChan chan error) {
	for {
		if *shouldStop {
			return
		}
		// Proceed to the write

		// Get the current time in unix timestamp
		filename := fmt.Sprintf("purewritetest-%s-%d", podName, time.Now().Unix())
		log.Debugf("Writing to pod '%s' in namespace '%s' in data dir '%s'. This will be filename %s", podName, ctx.App.NameSpace, dataDir, filename)
		cmdArgs := []string{"exec", "-it", podName, "-n", ctx.App.NameSpace, "--", "touch", path.Join(dataDir, filename)}
		err = osutils.Kubectl(cmdArgs)
		if err != nil {
			log.Errorf("Error writing to pod '%s' in namespace '%s' in data dir '%s': %v", podName, ctx.App.NameSpace, dataDir, err)
			errOutChan <- fmt.Errorf("Error writing to pod '%s' in namespace '%s' in data dir '%s': %v", podName, ctx.App.NameSpace, dataDir, err)
			return
		}

		// Sleep for 10 seconds
		time.Sleep(10 * time.Second)
	}
}

// GetPxPIDMap returns a map of node ID to PX rpocess PID
func GetPxPIDMap(nodes []node.Node) (map[string]string, error) {
	pxPIDMap := make(map[string]string)
	getPxPidCmd := "pidof px"

	for _, n := range nodes {
		err := Inst().V.WaitForPxPodsToBeUp(n)
		if err != nil {
			return nil, fmt.Errorf("failed to wait for PX pod to be up in node [%s]. Err: [%v]", n.Name, err)
		}
		output, err := Inst().N.RunCommand(n, getPxPidCmd, node.ConnectionOpts{Timeout: 30 * time.Second, TimeBeforeRetry: 20 * time.Second, Sudo: true})
		if err != nil {
			return nil, fmt.Errorf("failed to get PX PID on node [%s]. Err: [%v]", n.Name, err)
		}
		pxPIDMap[n.Id] = output
	}
	return pxPIDMap, nil
}

func StartPureBackgroundWriteRoutines() func() {
	pureStopWriteRoutine := false
	pureErrOutChan := make(chan error, 1) // We only need one failure to fail the entire test: no reason to store more than we need

	Step("start write routine on all Pure volumes to ensure data continuity", func() {
		for _, ctx := range contexts {
			var vols []*volume.Volume
			Step(fmt.Sprintf("get %s app's volumes", ctx.App.Key), func() {
				vols, err = Inst().S.GetVolumes(ctx)
				log.FailOnError(err, "Failed to get volumes for app %s", ctx.App.Key)
			})

			podNames := map[string]bool{}
			for _, vol := range vols {
				pods, err := Inst().S.GetPodsForPVC(vol.Name, vol.Namespace)
				log.FailOnError(err, "Failed to get pods for PVC %s in app %s", vol.Name, ctx.App.Key)
				for _, pod := range pods {
					podNames[pod.Name] = true
				}
			}

			dataDir, _ := pureutils.GetAppDataDir(ctx.App.NameSpace)
			for podName := range podNames {
				// Start a routine that will repeatedly write to this pod's data directory until we send something to the stop channel
				// If any errors occur, they will be sent to the err channel, and checked at the end of the test
				log.Infof("Starting background write routine for pod '%s' in namespace '%s' in data dir '%s'", podName, ctx.App.NameSpace, dataDir)
				go pureWriteRoutine(ctx, podName, dataDir, &pureStopWriteRoutine, pureErrOutChan)
			}
		}
	})
	return func() {
		// Finish up the write routines from earlier
		// First, check for any errors
		var err error
		select {
		case err = <-pureErrOutChan:
			log.FailOnError(err, "Error writing to Pure volume")
		default:
			log.Infof("No errors found in error channel for Pure volume write validation")
		}
		// Then, close all the routines out so they stop writing
		pureStopWriteRoutine = true
	}
}

// Volume Driver Plugin is down, unavailable - and the client container should not be impacted.
var _ = Describe("{VolumeDriverDown}", Label("p0", "negative", "px_vol_ops", "shared_v4", "px_restart"), func() {
	var testrailID = 35259
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/35259
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("VolumeDriverDown", "Validate volume driver down", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "has to schedule apps and stop volume driver on app nodes"
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("voldriverdown-%d", i))...)
		}

		ValidateApplications(contexts)

		var pureCleanupFunction func()
		if Inst().V.String() == portworx.PureDriverName {
			pureCleanupFunction = StartPureBackgroundWriteRoutines()
		}

		Step("get nodes bounce volume driver", func() {
			for _, appNode := range node.GetStorageDriverNodes() {
				stepLog = fmt.Sprintf("stop volume driver %s on node: %s",
					Inst().V.String(), appNode.Name)
				Step(stepLog,
					func() {
						log.InfoD(stepLog)
						StopVolDriverAndWait([]node.Node{appNode})
					})

				stepLog = fmt.Sprintf("starting volume %s driver on node %s",
					Inst().V.String(), appNode.Name)
				Step(stepLog,
					func() {
						log.InfoD(stepLog)
						StartVolDriverAndWait([]node.Node{appNode})
					})

				stepLog = "Giving few seconds for volume driver to stabilize"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					time.Sleep(20 * time.Second)
				})

				Step("validate apps", func() {
					for _, ctx := range contexts {
						ValidateContext(ctx)
					}
				})
			}

			if pureCleanupFunction != nil {
				pureCleanupFunction() // Checks for any errors during the background writes and fails the test if any occurred
				return
			}

			err := ValidateDataIntegrity(&contexts)
			log.FailOnError(err, "error validating data integrity")
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

// Volume Driver Plugin is down, unavailable on the nodes where the volumes are
// attached - and the client container should not be impacted.
var _ = Describe("{VolumeDriverDownAttachedNode}", Label("p0", "negative", "px_vol_ops", "shared_v4", "error_injection", "px_restart"), func() {
	var testrailID = 35260
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/35260
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("VolumeDriverDownAttachedNode", "Validate Volume drive down on an volume attached node", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "has to schedule apps and stop volume driver on nodes where volumes are attached"
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("voldriverdownattachednode-%d", i))...)
		}

		ValidateApplications(contexts)

		stepLog = "get nodes where app is running and restart volume driver"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, ctx := range contexts {
				appNodes, err := Inst().S.GetNodesForApp(ctx)
				dash.VerifySafely(err, nil, fmt.Sprintf("Verify Get nodes for app %s", ctx.App.Key))
				for _, appNode := range appNodes {
					stepLog = fmt.Sprintf("stop volume driver %s on app %s's node: %s",
						Inst().V.String(), ctx.App.Key, appNode.Name)
					Step(stepLog,
						func() {
							StopVolDriverAndWait([]node.Node{appNode})
						})

					stepLog = fmt.Sprintf("starting volume %s driver on app %s's node %s",
						Inst().V.String(), ctx.App.Key, appNode.Name)
					Step(stepLog,
						func() {
							StartVolDriverAndWait([]node.Node{appNode})
						})

					stepLog = "Giving few seconds for volume driver to stabilize"
					Step(stepLog, func() {
						log.InfoD("Giving few seconds for volume driver to stabilize")
						time.Sleep(20 * time.Second)
					})

					stepLog = fmt.Sprintf("validate app %s", ctx.App.Key)
					Step(stepLog, func() {
						log.InfoD(stepLog)
						ValidateContext(ctx)
					})
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
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

// Volume Driver Plugin has crashed - and the client container should not be impacted.
var _ = Describe("{VolumeDriverCrash}", Label("p0", "negative", "px_vol_ops", "shared_v4", "error_injection", "px_crash"), func() {
	var testrailID = 35261
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/35261
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("VolumeDriverCrash", "Validate PX after volume driver crash", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "has to schedule apps and crash volume driver on app nodes"
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("voldrivercrash-%d", i))...)
		}

		ValidateApplications(contexts)

		var pureCleanupFunction func()
		if Inst().V.String() == portworx.PureDriverName {
			pureCleanupFunction = StartPureBackgroundWriteRoutines()
		}

		stepLog = "crash volume driver in all nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			for _, appNode := range node.GetStorageDriverNodes() {
				stepLog = fmt.Sprintf("crash volume driver %s on node: %v",
					Inst().V.String(), appNode.Name)
				Step(stepLog,
					func() {
						log.InfoD(stepLog)
						if Inst().V.IsPxLiteCluster() {
							CrashPXDaemonAndWait([]node.Node{appNode})
						} else {
							CrashVolDriverAndWait([]node.Node{appNode})
						}
					})
			}
		})

		if pureCleanupFunction != nil {
			pureCleanupFunction() // Checks for any errors during the background writes and fails the test if any occurred
			return
		}

		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		ValidateAndDestroy(contexts, opts)
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

// Volume driver plugin is down and the client container gets terminated.
// There is a lost unmount call in this case. When the volume driver is
// back up, we should be able to detach and delete the volume.
var _ = Describe("{VolumeDriverAppDown}", Label("p1", "negative", "px_ops"), func() {
	var testrailID = 35262
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/35262
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("VolumeDriverAppDown", "Validate volume driver down and app deletion", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "has to schedule apps, stop volume driver on app nodes and destroy apps"

	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("voldriverappdown-%d", i))...)
		}

		ValidateApplications(contexts)

		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		stepLog = "get nodes for all apps in test and bounce volume driver"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, ctx := range contexts {
				appNodes, err := Inst().S.GetNodesForApp(ctx)
				log.FailOnError(err, "Failed to get nodes for the app %s", ctx.App.Key)
				appNode := appNodes[r.Intn(len(appNodes))]
				stepLog = fmt.Sprintf("stop volume driver %s on app %s's nodes: %v",
					Inst().V.String(), ctx.App.Key, appNode)
				Step(stepLog, func() {
					StopVolDriverAndWait([]node.Node{appNode})
				})

				stepLog = fmt.Sprintf("destroy app: %s", ctx.App.Key)
				Step(stepLog, func() {
					err = Inst().S.Destroy(ctx, nil)
					dash.VerifyFatal(err, nil, "Verify App delete")
					stepLog = "wait for few seconds for app destroy to trigger"
					Step(stepLog, func() {
						log.InfoD(stepLog)
						time.Sleep(10 * time.Second)
					})
				})

				stepLog = "restarting volume driver"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					StartVolDriverAndWait([]node.Node{appNode})
				})

				stepLog = fmt.Sprintf("wait for destroy of app: %s", ctx.App.Key)
				Step(stepLog, func() {
					log.InfoD(stepLog)
					err = Inst().S.WaitForDestroy(ctx, Inst().DestroyAppTimeout)
					dash.VerifySafely(err, nil, fmt.Sprintf("Verify App %s deletion", ctx.App.Key))
				})

				err = DeleteVolumesAndWait(ctx, &scheduler.VolumeOptions{})
				dash.VerifySafely(err, nil, fmt.Sprintf("%s's volume deleted successfully?", ctx.App.Key))

			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

// This test deletes all tasks of an application and checks if app converges back to desired state
var _ = Describe("{AppScaleUpAndDown}", Label("p1", "positive", "px_vol_ops"), func() {
	var testrailID = 35263
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/35264
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("AppScaleUpAndDown", "Validate app after tasks are deleted", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "has to schedule app and delete app tasks"
	It(stepLog, func() {
		log.InfoD(stepLog)
		var err error
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("apptasksdown-%d", i))...)
		}

		ValidateApplications(contexts)

		stepLog = "delete all application tasks"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			// Add interval based sleep here to check what time we will exit out of this delete task loop
			minRunTime := Inst().MinRunTimeMins
			timeout := (minRunTime) * 60
			// set frequency mins depending on the chaos level
			var frequency int
			switch Inst().ChaosLevel {
			case 5:
				frequency = 1
			case 4:
				frequency = 3
			case 3:
				frequency = 5
			case 2:
				frequency = 7
			case 1:
				frequency = 10
			default:
				frequency = 10

			}
			if minRunTime == 0 {
				for _, ctx := range contexts {
					stepLog = fmt.Sprintf("delete tasks for app: %s", ctx.App.Key)
					Step(stepLog, func() {
						err = Inst().S.DeleteTasks(ctx, nil)
						if err != nil {
							PrintDescribeContext(ctx)
						}
						dash.VerifyFatal(err, nil, fmt.Sprintf("validate delete tasks for app: %s", ctx.App.Key))
					})

					ValidateContext(ctx)
				}
			} else {
				start := time.Now().Local()
				for int(time.Since(start).Seconds()) < timeout {
					for _, ctx := range contexts {
						stepLog = fmt.Sprintf("delete tasks for app: %s", ctx.App.Key)
						Step(stepLog, func() {
							err = Inst().S.DeleteTasks(ctx, nil)
							if err != nil {
								PrintDescribeContext(ctx)
							}
							dash.VerifyFatal(err, nil, fmt.Sprintf("validate delete tasks for app: %s", ctx.App.Key))
						})

						ValidateContext(ctx)
					}
					stepLog = fmt.Sprintf("Sleeping for given duration %d", frequency)
					Step(stepLog, func() {
						log.InfoD(stepLog)
						d := time.Duration(frequency)
						time.Sleep(time.Minute * d)
					})
				}
			}
		})

		Step("teardown all apps", func() {
			for _, ctx := range contexts {
				TearDownContext(ctx, nil)
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

// This test scales up and down an application and checks if app has actually scaled accordingly
var _ = Describe("{AppScaleUpAndDown}", Label("p1", "positive", "px_vol_ops"), func() {
	var testrailID = 35264
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/35264
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("AppScaleUpAndDown", "Validate Apps sclae up and scale down", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "has to scale up and scale down the app"
	It(stepLog, func() {
		log.InfoD("has to scale up and scale down the app")
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("applicationscaleupdown-%d", i))...)
		}

		ValidateApplications(contexts)

		stepLog = "Scale up and down all app"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, ctx := range contexts {
				stepLog = fmt.Sprintf("scale up app: %s by %d ", ctx.App.Key, len(node.GetWorkerNodes()))
				Step(stepLog, func() {
					log.InfoD(stepLog)
					applicationScaleUpMap, err := Inst().S.GetScaleFactorMap(ctx)
					log.FailOnError(err, "Failed to get application scale up factor map")
					//Scaling up by number of storage-nodes
					workerStorageNodes := int32(len(node.GetStorageNodes()))
					for name, scale := range applicationScaleUpMap {
						// limit scale up to the number of worker nodes
						if scale < workerStorageNodes {
							applicationScaleUpMap[name] = workerStorageNodes
						}
					}
					err = Inst().S.ScaleApplication(ctx, applicationScaleUpMap)
					dash.VerifyFatal(err, nil, "Validate application scale up")
				})

				stepLog = "Giving few seconds for scaled up applications to stabilize"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					time.Sleep(10 * time.Second)
				})

				ValidateContext(ctx)

				stepLog = fmt.Sprintf("scale down app %s by 1", ctx.App.Key)
				Step(stepLog, func() {
					log.InfoD(stepLog)
					applicationScaleDownMap, err := Inst().S.GetScaleFactorMap(ctx)
					log.FailOnError(err, "Failed to get application scale down factor map")

					for name, scale := range applicationScaleDownMap {
						applicationScaleDownMap[name] = scale - 1
					}
					err = Inst().S.ScaleApplication(ctx, applicationScaleDownMap)
					dash.VerifyFatal(err, nil, "Validate application scale down")
				})

				stepLog = "Giving few seconds for scaled up applications to stabilize"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					time.Sleep(10 * time.Second)
				})

				ValidateContext(ctx)
			}
		})

		Step("teardown all apps", func() {
			for _, ctx := range contexts {
				TearDownContext(ctx, nil)
			}
		})

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{CordonDeployDestroy}", Label("p1", "negative", "px_ops"), func() {
	var testrailID = 54373
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/54373
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("CordonDeployDestroy", "Validate Cordon node and destroy app", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	var contexts []*scheduler.Context

	stepLog := "has to cordon all nodes but one, deploy and destroy app"
	It(stepLog, func() {
		log.InfoD(stepLog)
		stepLog = "Cordon all nodes but one"

		Step(stepLog, func() {
			log.InfoD(stepLog)
			nodes := node.GetStorageDriverNodes()
			for _, node := range nodes[1:] {
				err := Inst().S.DisableSchedulingOnNode(node)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Validate disable scheduling on node %s", node.Name))

			}
		})
		stepLog = "Deploy applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			contexts = make([]*scheduler.Context, 0)

			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("cordondeploydestroy-%d", i))...)
			}
			ValidateApplications(contexts)

		})
		stepLog = "Destroy apps"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForDestroy] = false
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = false
			for _, ctx := range contexts {
				err := Inst().S.Destroy(ctx, opts)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Validate App %s detroy init", ctx.App.Key))
			}
		})
		Step("Validate destroy", func() {
			for _, ctx := range contexts {
				err := Inst().S.WaitForDestroy(ctx, Inst().DestroyAppTimeout)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Validate App %s detroy", ctx.App.Key))

			}
		})
		Step("teardown all apps", func() {
			for _, ctx := range contexts {
				TearDownContext(ctx, nil)
			}
		})
		Step("Uncordon all nodes", func() {
			nodes := node.GetStorageDriverNodes()
			for _, node := range nodes {
				err := Inst().S.EnableSchedulingOnNode(node)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Validate enable scheduling on node %s", node.Name))
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)

	})
})

var _ = Describe("{CordonStorageNodesDeployDestroy}", Label("p1", "negative", "px_ops"), func() {
	JustBeforeEach(func() {
		StartTorpedoTest("CordonStorageNodesDeployDestroy", "Validate Cordon storage node , deploy and destroy app", nil, 0)

	})
	var contexts []*scheduler.Context

	stepLog := "has to cordon all storage nodes, deploy and destroy app"
	It(stepLog, func() {
		log.InfoD(stepLog)
		stepLog = "Cordon all storage nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			nodes := node.GetNodes()
			storageNodes := node.GetStorageNodes()
			if len(nodes) == len(storageNodes) {
				stepLog = "No storageless nodes detected. Skipping.."
				log.Warn(stepLog)
				Skip(stepLog)
			}
			for _, n := range storageNodes {
				err := Inst().S.DisableSchedulingOnNode(n)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Validate disable scheduling on node %s", n.Name))
			}
		})
		stepLog = "Deploy applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			contexts = make([]*scheduler.Context, 0)

			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("cordondeploydestroy-%d", i))...)
			}
			ValidateApplications(contexts)

		})
		stepLog = "Destroy apps"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForDestroy] = false
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = false
			for _, ctx := range contexts {
				err := Inst().S.Destroy(ctx, opts)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Validate App %s detroy init", ctx.App.Key))
			}
		})
		Step("Validate destroy", func() {
			for _, ctx := range contexts {
				err := Inst().S.WaitForDestroy(ctx, Inst().DestroyAppTimeout)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Validate App %s detroy", ctx.App.Key))

			}
		})
		Step("teardown all apps", func() {
			for _, ctx := range contexts {
				TearDownContext(ctx, nil)
			}
		})
		Step("Uncordon all nodes", func() {
			nodes := node.GetStorageDriverNodes()
			for _, node := range nodes {
				err := Inst().S.EnableSchedulingOnNode(node)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Validate enable scheduling on node %s", node.Name))
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{SecretsVaultFunctional}", Label("p1", "positive", "px_ops"), func() {
	var testrailID, runID int
	var contexts []*scheduler.Context
	var provider string

	const (
		vaultSecretProvider        = "vault"
		vaultTransitSecretProvider = "vault-transit"
		portworxContainerName      = "portworx"
	)

	BeforeEach(func() {
		StartTorpedoTest("SecretsVaultFunctional", "Validate Secrets Vault", nil, 0)
		isOpBased, _ := Inst().V.IsOperatorBasedInstall()
		if !isOpBased {
			k8sApps := apps.Instance()
			daemonSets, err := k8sApps.ListDaemonSets("kube-system", metav1.ListOptions{
				LabelSelector: "name=portworx",
			})
			log.FailOnError(err, "Failed to get daemon sets list")
			dash.VerifyFatal(len(daemonSets) > 0, true, "Daemon sets returned?")
			dash.VerifyFatal(len(daemonSets[0].Spec.Template.Spec.Containers) > 0, true, "Daemon set container is not empty?")
			usingVault := false
			for _, container := range daemonSets[0].Spec.Template.Spec.Containers {
				if container.Name == portworxContainerName {
					for _, arg := range container.Args {
						if arg == vaultSecretProvider || arg == vaultTransitSecretProvider {
							usingVault = true
							provider = arg
						}
					}
				}
			}
			if !usingVault {
				skipLog := fmt.Sprintf("Skip test for not using %s or %s ", vaultSecretProvider, vaultTransitSecretProvider)
				log.Warn(skipLog)
				Skip(skipLog)
			}
		} else {
			spec, err := Inst().V.GetDriver()
			log.FailOnError(err, "Failed to get storage cluster")
			if *spec.Spec.SecretsProvider != vaultSecretProvider &&
				*spec.Spec.SecretsProvider != vaultTransitSecretProvider {
				Skip(fmt.Sprintf("Skip test for not using %s or %s ", vaultSecretProvider, vaultTransitSecretProvider))
			}
			provider = *spec.Spec.SecretsProvider
		}
	})

	var _ = Describe("{RunSecretsLogin}", Label("p2", "positive", "px_ops"), func() {
		testrailID = 82774
		// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/82774
		JustBeforeEach(func() {
			StartTorpedoTest("RunSecretsLogin", "Test secrets login for vaults", nil, 0)
			runID = testrailuttils.AddRunsToMilestone(testrailID)
		})

		stepLog := "has to run secrets login for vault or vault-transit"

		It(stepLog, func() {
			log.InfoD(stepLog)
			contexts = make([]*scheduler.Context, 0)
			n := node.GetStorageDriverNodes()[0]
			if provider == vaultTransitSecretProvider {
				// vault-transit login with `pxctl secrets vaulttransit login`
				provider = "vaulttransit"
			}
			err := Inst().V.RunSecretsLogin(n, provider)
			dash.VerifyFatal(err, nil, "Validate secrets login")
		})
	})

	AfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{VolumeCreatePXRestart}", Label("p1", "negative", "px_vol_ops", "px_restart"), func() {
	JustBeforeEach(func() {
		StartTorpedoTest("VolumeCreatePXRestart", "Validate restart PX while create and attach", nil, 0)

	})
	contexts := make([]*scheduler.Context, 0)

	stepLog := "Validate volume attachment when px is restarting"
	It(stepLog, func() {
		var createdVolIDs map[string]string
		var err error
		volCreateCount := 10
		stepLog := "Create multiple volumes , attached and restart PX"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			stNodes := node.GetStorageNodes()
			index := rand.Intn(len(stNodes))
			selectedNode := stNodes[index]

			log.InfoD("Creating and attaching %d volumes on node %s", volCreateCount, selectedNode.Name)

			wg := new(sync.WaitGroup)
			wg.Add(1)
			go func(appNode node.Node) {
				createdVolIDs, err = CreateMultiVolumesAndAttach(wg, volCreateCount, selectedNode.Id)
				if err != nil {
					log.Fatalf("Error while creating volumes. Err: %v", err)
				}
			}(selectedNode)
			time.Sleep(2 * time.Second)
			wg.Add(1)
			go func(appNode node.Node) {
				defer wg.Done()
				stepLog = fmt.Sprintf("restart volume driver %s on node: %s", Inst().V.String(), appNode.Name)
				Step(stepLog, func() {
					log.InfoD(stepLog)
					err = Inst().V.RestartDriver(appNode, nil)
					log.FailOnError(err, "Error while restarting volume driver")

				})
			}(selectedNode)
			wg.Wait()

		})

		stepLog = "Validate the created volumes"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			for vol, volPath := range createdVolIDs {
				cVol, err := Inst().V.InspectVolume(vol)
				if err == nil {
					dash.VerifySafely(cVol.State, opsapi.VolumeState_VOLUME_STATE_ATTACHED, fmt.Sprintf("Verify vol %s is attached", cVol.Id))
					dash.VerifySafely(cVol.DevicePath, volPath, fmt.Sprintf("Verify vol %s is has device path", cVol.Id))
				} else {
					log.Fatalf("Error while inspecting volume %s. Err: %v", vol, err)
				}
			}
		})

		stepLog = "Deleting the created volumes"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			for vol := range createdVolIDs {
				log.Infof("Detaching and deleting volume: %s", vol)
				err := Inst().V.DetachVolume(vol)
				if err == nil {
					err = Inst().V.DeleteVolume(vol)
				}
				log.FailOnError(err, "Error while deleting volume %s", vol)

			}
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{AutoFSTrimReplAddWithNoPool0}", Label("p1", "positive", "px_vol_ops"), func() {
	/*
		1. Enable autofstrim, wait until autofstrim is actively trimming.
		2. Run back to back cmd `pxctl c options update --auto-fstrim off` , then  `pxctl c options update --auto-fstrim on`
		3.`pxctl v af status` should report the trim status correctly
		4. All Nodes have pool 0 and pool 1.
		5. Delete pool 0 in one of the node
		6. Create repl-2 volumes on nodes with pool 0 and pool 1.
		7. Check auto-fstrim is working
		8. Do a repl-add on all volumes on the node with only pool 1
		9. Check if auto-fstrim is working
	*/
	var testrailID = 84604
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/84604
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("AutoFSTrimReplAddWithNoPool0", "validate autofstrim with repl add with no pool 0", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context

	stepLog := "has to check the autofstrim status with fast switch and repl add"
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)

		storageDriverNodes := node.GetStorageDriverNodes()
		stepLog = "perform fast switch on the autofstrim option"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			stNode := storageDriverNodes[0]
			clusterOpts, err := Inst().V.GetClusterOpts(stNode, []string{"AutoFstrim"})
			log.FailOnError(err, fmt.Sprintf("error getting AutoFstrim status using node [%s]", stNode.Name))
			if clusterOpts["AutoFstrim"] == "false" {
				EnableAutoFSTrim()
				//making sure autofstrim is actively running
				time.Sleep(1 * time.Minute)
			}

			//switching autofstrim feature
			err = Inst().V.SetClusterOpts(stNode, map[string]string{
				"--auto-fstrim": "off"})
			log.FailOnError(err, "error disabling AutoFstrim status using node [%s]", stNode.Name)
			err = Inst().V.SetClusterOpts(stNode, map[string]string{
				"--auto-fstrim": "on"})
			log.FailOnError(err, "error enabling AutoFstrim status using node [%s]", stNode.Name)
			_, err = Inst().V.GetAutoFsTrimStatus(stNode.DataIp)
			dash.VerifyFatal(err, nil, fmt.Sprintf("verify autofstrim status with fast switch from node [%v]", stNode.Name))

			//Check autofstrim using PXCTL
			opts := node.ConnectionOpts{
				IgnoreError:     false,
				TimeBeforeRetry: defaultRetryInterval,
				Timeout:         defaultTimeout,
				Sudo:            true,
			}

			output, err := Inst().V.GetPxctlCmdOutputConnectionOpts(stNode, "v af status", opts, false)
			log.FailOnError(err, "error checking autofstrim status on node [%s]", stNode.Name)
			log.Infof("autofstrim status: %s", output)
			dash.VerifyFatal(strings.Contains(output, "Filesystem Trim Initializing"), false, "verify no autofstrim initialization issue")

		})

		if !Contains(Inst().AppList, "fio-fstrim") {
			Inst().AppList = append(Inst().AppList, "fio-fstrim")
		}

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("afrepladd-%d", i))...)
		}
		ValidateApplications(contexts)

		stepLog = "setting volumes repl to 2"
		var appVolumes []*volume.Volume
		var selectedCtx *scheduler.Context
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, ctx := range contexts {
				if !strings.Contains(ctx.App.Key, "fio-fstrim") {
					continue
				}
				selectedCtx = ctx

				var err error
				stepLog = fmt.Sprintf("get volumes for %s app", ctx.App.Key)
				Step(stepLog, func() {
					log.InfoD(stepLog)
					appVolumes, err = Inst().S.GetVolumes(ctx)
					log.FailOnError(err, "Failed to get volumes")
					dash.VerifyFatal(len(appVolumes) > 0, true, fmt.Sprintf("Found %d app volmues", len(appVolumes)))
				})

				for _, v := range appVolumes {
					// Check if volumes are Pure FA/FB DA volumes
					isPureVol, err := Inst().V.IsPureVolume(v)
					log.FailOnError(err, "Failed to check is PURE volume")
					if isPureVol {
						log.Warnf("Repl increase on Pure DA Volume [%s] not supported. Skipping this operation", v.Name)
						continue
					}

					currRep, err := Inst().V.GetReplicationFactor(v)
					log.FailOnError(err, "Failed to get Repl factor for vil %s", v.Name)

					//Reduce replication factor
					if currRep == 3 {
						log.Infof("Current replication is 3, reducing before proceeding")
						opts := volume.Options{
							ValidateReplicationUpdateTimeout: validateReplicationUpdateTimeout,
						}
						err = Inst().V.SetReplicationFactor(v, currRep-1, nil, nil, true, opts)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Validate set repl factor to %d", currRep-1))
					}

				}
			}
		})

		stepLog = "has to check the autofstrim status with volume update on node with no pool 0"
		var fsTrimStatuses map[string]opsapi.FilesystemTrim_FilesystemTrimStatus
		var fsUsage map[string]*opsapi.FstrimVolumeUsageInfo
		var err error
		Step(stepLog, func() {
			log.InfoD(stepLog)

			fsTrimStatuses, err = GetAutoFsTrimStatusForCtx(selectedCtx)
			log.FailOnError(err, "error getting autofs status")

			fsUsage, err = GetAutoFstrimUsageForCtx(selectedCtx)
			log.FailOnError(err, "error getting autofs usage")

			volReplNodes := make([]string, 0)
			for k := range fsTrimStatuses {
				selectedVol, err := Inst().V.InspectVolume(k)
				log.FailOnError(err, "error inspecting vol [%s]", k)
				volReplNodes = append(volReplNodes, selectedVol.ReplicaSets[0].Nodes...)
			}

			//Selecting a node not part of selected volume replica set and multiple pools
			stNodes := node.GetStorageNodes()
			var selectedNode node.Node
			for _, n := range stNodes {
				if !Contains(volReplNodes, n.Id) && len(n.Pools) > 1 {
					//verifying if node has pool 0
					var ispoolExist bool
					for _, p := range n.Pools {
						if p.ID == 0 {
							ispoolExist = true
						}
					}
					if ispoolExist {
						selectedNode = n
						break
					}
				}
			}

			if selectedNode.Id == "" {
				log.FailOnError(fmt.Errorf("no node found"), "error identifying node for repl increase and pool deletion")
			}

			stepLog = fmt.Sprintf("delete pool 0 from the node [%s]", selectedNode.Name)
			Step(stepLog, func() {
				poolsBfr, err := Inst().V.ListStoragePools(metav1.LabelSelector{})
				log.FailOnError(err, "Failed to list storage pools")

				log.InfoD(stepLog)
				log.InfoD("Setting pools in maintenance on node %s", selectedNode.Name)
				err = Inst().V.EnterPoolMaintenance(selectedNode)
				log.FailOnError(err, "failed to set pool maintenance mode on node %s", selectedNode.Name)

				time.Sleep(1 * time.Minute)
				expectedStatus := "In Maintenance"
				err = WaitForPoolStatusToUpdate(selectedNode, expectedStatus)
				log.FailOnError(err, fmt.Sprintf("node %s pools are not in status %s", selectedNode.Name, expectedStatus))

				err = Inst().V.DeletePool(selectedNode, "0", true)
				log.FailOnError(err, "failed to delete poolID 0 on node %s", selectedNode.Name)

				err = Inst().V.ExitPoolMaintenance(selectedNode)
				log.FailOnError(err, "failed to exit pool maintenance mode on node %s", selectedNode.Name)

				err = Inst().V.WaitDriverUpOnNode(selectedNode, 5*time.Minute)
				log.FailOnError(err, "volume driver down on node %s", selectedNode.Name)

				expectedStatus = "Online"
				err = WaitForPoolStatusToUpdate(selectedNode, expectedStatus)
				log.FailOnError(err, fmt.Sprintf("node %s pools are not in status %s", selectedNode.Name, expectedStatus))

				poolsAfr, err := Inst().V.ListStoragePools(metav1.LabelSelector{})
				log.FailOnError(err, "Failed to list storage pools")

				dash.VerifySafely(len(poolsBfr) > len(poolsAfr), true, "verify pools count is updated after pools deletion")
			})

			stepLog = fmt.Sprintf("Repl increase of volumes to the node %s", selectedNode.Name)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				appVolumes, err = Inst().S.GetVolumes(selectedCtx)
				log.FailOnError(err, "Failed to get volumes")
				for _, v := range appVolumes {
					appVol, err := Inst().V.InspectVolume(v.ID)
					log.FailOnError(err, fmt.Sprintf("error inspecting volume [%s]", v.ID))

					if _, ok := fsTrimStatuses[appVol.Id]; ok {
						opts := volume.Options{
							ValidateReplicationUpdateTimeout: validateReplicationUpdateTimeout,
						}
						err = Inst().V.SetReplicationFactor(v, 3, []string{selectedNode.Id}, nil, true, opts)
						dash.VerifyFatal(err, nil, fmt.Sprintf("Validate set repl factor to 3 for volume [%s]", v.Name))
					}
				}

			})

			stepLog = fmt.Sprintf("validate autofstrim status for the volumes")
			Step(stepLog, func() {
				newFsTrimStatuses, err := GetAutoFsTrimStatusForCtx(selectedCtx)
				log.FailOnError(err, "error getting autofs status")

				for k := range fsTrimStatuses {
					val, ok := newFsTrimStatuses[k]
					dash.VerifySafely(ok, true, fmt.Sprintf("verify autofstrim started for volume %s", k))
					dash.VerifySafely(val != opsapi.FilesystemTrim_FS_TRIM_FAILED, true, fmt.Sprintf("verify autofstrim for volume %s, current status %v", k, val))
				}
				newFsUsage, err := GetAutoFstrimUsageForCtx(selectedCtx)
				log.FailOnError(err, "error getting autofs usage")

				for k := range fsUsage {
					val, ok := newFsUsage[k]
					dash.VerifySafely(ok, true, fmt.Sprintf("verify autofstrim usage for volume %s", k))
					dash.VerifySafely(val.PerformAutoFstrim, "Enabled", fmt.Sprintf("verify autofstrim for volume %s is not disabled, current status %v", k, val))
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

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

// Add a test to detach all the disks from a node and then reattach them to a different node
var _ = Describe("{NodeDiskDetachAttach}", Label("p1", "positive", "node_ops", "error_injection", "drive_failure"), func() {

	JustBeforeEach(func() {
		StartTorpedoTest("NodeDiskDetachAttach", "Validate disk detach and attach", nil, 0)
	})
	var contexts []*scheduler.Context

	testName := "nodediskdetachattach"
	stepLog := "has to detach all disks from a node and reattach them to a different node"
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("%s-%d", testName, i))...)
		}

		randomNum := rand.New(rand.NewSource(time.Now().Unix()))

		ValidateApplications(contexts)
		// Fetch the node where PX is not started
		nonPXNodes := node.GetPXDisabledNodes()
		// Remove disks from all nodes in nonPXNodes
		for _, n := range nonPXNodes {
			// Fetch the disks attached to the node
			err := Inst().N.RemoveNonRootDisks(n)
			log.FailOnError(err, fmt.Sprintf("Failed to remove disks on node %s", n.Name))
		}
		randNonPxNode := nonPXNodes[randomNum.Intn(len(nonPXNodes))]

		// Fetch a random node from the StorageNodes
		storageNodes := node.GetStorageNodes()
		randomStorageNode := storageNodes[randomNum.Intn(len(storageNodes))]

		// Fetch random storage node and detach the disks attached to that node
		var oldNodeIDtoMatch string
		var newNodeIDtoMatch string
		Step(fmt.Sprintf("detach disks attached from a random storage node %s and attach to non px node %s", randomStorageNode.Name, randNonPxNode.Name), func() {
			log.Infof("Detaching disks from node %s and attaching to node %s", randomStorageNode.Name, randNonPxNode.Name)
			// ToDo - Ensure the above selected node has volumes/apps running on it
			// Store the Node ID of randomStorageNode for future use
			oldNodeIDtoMatch = randomStorageNode.Id
			// Stop PX on the node
			err = k8sCore.AddLabelOnNode(randomStorageNode.Name, schedops.PXServiceLabelKey, "stop")
			log.FailOnError(err, fmt.Sprintf("Failed to add label %s=stop on node %s", schedops.PXServiceLabelKey, randomStorageNode.Name))

			err = Inst().N.MoveDisks(randomStorageNode, randNonPxNode)
			log.FailOnError(err, fmt.Sprintf("Failed to move disks from node %s to node %s", randomStorageNode.Name, randNonPxNode.Name))

			// Add PXEnabled false label to srcVM
			err = k8sCore.AddLabelOnNode(randomStorageNode.Name, schedops.PXEnabledLabelKey, "false")
			log.FailOnError(err, fmt.Sprintf("Failed to add label %s=stop on node %s", schedops.PXServiceLabelKey, randomStorageNode.Name))

			// Start PX on the node
			err = k8sCore.AddLabelOnNode(randNonPxNode.Name, schedops.PXEnabledLabelKey, "true")
			log.FailOnError(err, fmt.Sprintf("Failed to add label %s=true on node %s", schedops.PXEnabledLabelKey, randNonPxNode.Name))
			err = k8sCore.AddLabelOnNode(randNonPxNode.Name, schedops.PXServiceLabelKey, "start")
			log.FailOnError(err, fmt.Sprintf("Failed to add label %s=start on node %s", schedops.PXServiceLabelKey, randNonPxNode.Name))
			err = Inst().V.WaitForPxPodsToBeUp(randNonPxNode)
			log.FailOnError(err, fmt.Sprintf("Failed to wait for PX pods to be up on node %s", randNonPxNode.Name))
			// Refresh the driver endpoints
			err = Inst().S.RefreshNodeRegistry()
			log.FailOnError(err, "error refreshing node registry")
			err = Inst().V.RefreshDriverEndpoints()
			log.FailOnError(err, "error refreshing storage drive endpoints")
			// Wait for the driver to be up on the node
			randNonPxNode, err = node.GetNodeByName(randNonPxNode.Name)
			log.FailOnError(err, fmt.Sprintf("Failed to get node %s", randNonPxNode.Name))
			err = Inst().V.WaitDriverUpOnNode(randNonPxNode, Inst().DriverStartTimeout)
			dash.VerifyFatal(err, nil, "Validate volume is driver up")
			// Verify the node ID is same as earlier stored Node ID
			newNodeIDtoMatch = randNonPxNode.Id

			dash.VerifyFatal(oldNodeIDtoMatch, newNodeIDtoMatch, fmt.Sprintf("Node ID mismatch for node %s after moving the disks", randNonPxNode.Name))
			log.Infof("Node ID matches for node %s [%s == %s]", randNonPxNode.Name, oldNodeIDtoMatch, newNodeIDtoMatch)
		})

		// ToDo - Verify the integrity of apps, cluster and volumes

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
		AfterEachTest(contexts)
	})
})

var _ = Describe("{DeployApps}", Label("p2", "positive", "px_ops"), func() {

	JustBeforeEach(func() {
		StartTorpedoTest("DeployApps", "Validate Apps deployment", nil, 0)

	})

	var contexts []*scheduler.Context

	It("has to deploy and  validate  apps", func() {
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("deployapps-%d", i))...)
		}
		ValidateApplications(contexts)

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{RestartMultipleStorageNodeOneKVDBMaster}", Label("p1", "negative", "pc_ops", "kvdb_ops", "error_injection", "px_restart", "KVDBFailover"), func() {
	/*
		Restart Multiple Storage Nodes with one KVDB Master in parallel and wait for the node to come back online
		https://portworx.atlassian.net/browse/PTX-17618
	*/
	JustBeforeEach(func() {
		StartTorpedoTest("RestartMultipleStorageNodeOneKVDBMaster",
			"Restart Multiple Storage Nodes with one KVDB Master",
			nil, 0)
	})
	var contexts []*scheduler.Context
	stepLog := "Expand multiple pool in the cluster at once in parallel"
	It(stepLog, func() {
		contexts = make([]*scheduler.Context, 0)
		var wg sync.WaitGroup

		listOfStorageNodes := node.GetStorageNodes()
		// Test Needs minimum of 3 nodes other than 3 KVDB Member nodes
		// so that few storage nodes (except kvdb nodes ) can be restarted
		dash.VerifyFatal(len(listOfStorageNodes) >= 6, true, "Test Needs minimum of 6 Storage Nodes")

		// assuming that there are minimum number of 3 nodes minus kvdb member nodes , we pick atleast 50% of the nodes for restating
		var nodesToReboot []node.Node
		getKVDBNodes, err := GetAllKvdbNodes()
		log.FailOnError(err, "failed to get list of all kvdb nodes")

		// Verifying if we have kvdb quorum set
		dash.VerifyFatal(len(getKVDBNodes) == 3, true, "missing required kvdb member nodes")

		// Get 50 % of other nodes for restart
		nodeCountsForRestart := (len(listOfStorageNodes) - len(getKVDBNodes)) / 2
		log.InfoD("total nodes picked for rebooting [%v]", nodeCountsForRestart)

		isKVDBNode := func(n node.Node) (bool, bool) {
			for _, eachKvdb := range getKVDBNodes {
				if n.Id == eachKvdb.ID {
					if eachKvdb.Leader == true {
						return true, true
					} else {
						return true, false
					}
				}
			}
			return false, false
		}

		count := 0
		// Add one KVDB node to the List
		for _, each := range listOfStorageNodes {
			kvdbNode, master := isKVDBNode(each)
			if kvdbNode == true && master == true {
				nodesToReboot = append(nodesToReboot, each)
				count = count + 1
			}
		}
		// Add nodes which are not KVDB Nodes
		for _, each := range listOfStorageNodes {
			kvdbNode, _ := isKVDBNode(each)
			if kvdbNode == false {
				if count <= nodeCountsForRestart {
					nodesToReboot = append(nodesToReboot, each)
					count = count + 1
				}
			}
		}

		for _, eachNode := range nodesToReboot {
			log.InfoD("Selected Node [%v] for Restart", eachNode.Name)
		}

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("rebootmulparallel-%d", i))...)
		}
		ValidateApplications(contexts)
		defer appsValidateAndDestroy(contexts)

		// Initiate all node reboot at once using Go Routines
		wg.Add(len(nodesToReboot))

		rebootNode := func(n node.Node) {
			defer wg.Done()
			defer GinkgoRecover()
			log.InfoD("Rebooting Node [%v]", n.Name)

			err := Inst().N.RebootNode(n, node.RebootNodeOpts{
				Force: true,
				ConnectionOpts: node.ConnectionOpts{
					Timeout:         1 * time.Minute,
					TimeBeforeRetry: 5 * time.Second,
				},
			})
			log.FailOnError(err, "failed to reboot Node [%v]", n.Name)

		}

		// Initiating Go Routing to reboot all the nodes at once
		rebootAllNodes := func() {
			for _, each := range nodesToReboot {
				log.InfoD("Node to Reboot [%v]", each.Name)
				go rebootNode(each)
			}
			wg.Wait()

			// Wait for connection to come back online after reboot
			for _, each := range nodesToReboot {
				err = Inst().N.TestConnection(each, node.ConnectionOpts{
					Timeout:         15 * time.Minute,
					TimeBeforeRetry: 10 * time.Second,
				})

				err = Inst().S.IsNodeReady(each)
				log.FailOnError(err, "Node [%v] is not in ready state", each.Name)

				err = Inst().V.WaitDriverUpOnNode(each, Inst().DriverStartTimeout)
				log.FailOnError(err, "failed waiting for driver up on Node[%v]", each.Name)
			}
		}

		// Reboot all the Nodes at once
		rebootAllNodes()

		// Verifications
		getKVDBNodes, err = GetAllKvdbNodes()
		log.FailOnError(err, "failed to get list of all kvdb nodes")
		dash.VerifyFatal(len(getKVDBNodes) == 3, true, "missing required kvdb member nodes after node reboot")

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{KvdbFailoverSnapVolCreateDelete}", Label("p1", "negative", "kvdb_ops", "px_restart", "KVDBFailover"), func() {
	/*
		KVDB failover when lots of snap create/delete, volume inspect requests are coming
		https://portworx.atlassian.net/browse/PTX-17729
	*/
	JustBeforeEach(func() {
		StartTorpedoTest("KvdbFailoverSnapVolCreateDelete",
			"KVDB failover when lot of snap create/delete, volume inspect requests are coming",
			nil, 0)
	})
	var contexts []*scheduler.Context
	stepLog := "Expand multiple pool in the cluster at once in parallel"
	It(stepLog, func() {
		contexts = make([]*scheduler.Context, 0)
		var wg sync.WaitGroup
		wg.Add(4)
		var volumesCreated []string
		var snapshotsCreated []string

		terminate := false

		stopRoutine := func() {
			if !terminate {
				terminate = true
				wg.Done()
				for _, each := range volumesCreated {
					if IsVolumeExits(each) {
						log.FailOnError(Inst().V.DeleteVolume(each), "volume deletion failed on the cluster with volume ID [%s]", each)
					}

				}
				for _, each := range snapshotsCreated {
					if IsVolumeExits(each) {
						log.FailOnError(Inst().V.DeleteVolume(each), "Snapshot Volume deletion failed on the cluster with ID [%s]", each)
					}
				}
			}
		}
		defer stopRoutine()

		go func() {
			defer wg.Done()
			defer GinkgoRecover()

			// Volume Create continuously
			for {
				if terminate {
					break
				}
				// Create Volume on the Cluster
				uuidObj := uuid.New()
				VolName := fmt.Sprintf("volume_%s", uuidObj.String())
				Size := uint64(rand.Intn(10) + 1)   // Size of the Volume between 1G to 10G
				haUpdate := int64(rand.Intn(3) + 1) // Size of the HA between 1 and 3

				volId, err := Inst().V.CreateVolume(VolName, Size, int64(haUpdate))
				log.FailOnError(err, "volume creation failed on the cluster with volume name [%s]", VolName)
				log.InfoD("Volume created with name [%s] having id [%s]", VolName, volId)

				volumesCreated = append(volumesCreated, volId)
			}
		}()

		inspectDeleteVolume := func(volumeId string) error {
			defer GinkgoRecover()
			if IsVolumeExits(volumeId) {
				// inspect volume
				appVol, err := Inst().V.InspectVolume(volumeId)
				if err != nil {
					stopRoutine()
					return err
				}

				err = Inst().V.DeleteVolume(appVol.Id)
				if err != nil {
					stopRoutine()
					return err
				}
			}
			return nil
		}

		go func() {
			defer wg.Done()
			defer GinkgoRecover()

			// Create Snapshots on Volumes continuously
			for {
				if terminate {
					break
				}
				if len(volumesCreated) > 5 {
					for _, eachVol := range volumesCreated {
						uuidCreated := uuid.New()
						snapshotName := fmt.Sprintf("snapshot_%s_%s", eachVol, uuidCreated.String())

						snapshotResponse, err := Inst().V.CreateSnapshot(eachVol, snapshotName)
						if err != nil {
							stopRoutine()
							log.FailOnError(err, "error Creating Snapshot [%s]", eachVol)
						}

						snapshotsCreated = append(snapshotsCreated, snapshotResponse.GetSnapshotId())
						log.InfoD("Snapshot [%s] created with ID [%s]", snapshotName, snapshotResponse.GetSnapshotId())

						err = inspectDeleteVolume(eachVol)
						log.FailOnError(err, "Inspect and Delete Volume failed on cluster with Volume ID [%v]", eachVol)

						// Remove the first element
						for i := 0; i < len(volumesCreated)-1; i++ {
							volumesCreated[i] = volumesCreated[i+1]
						}
						// Resize the array by truncating the last element
						volumesCreated = volumesCreated[:len(volumesCreated)-1]
					}
				}
			}
		}()

		go func() {
			defer wg.Done()
			defer GinkgoRecover()

			// Delete Snapshots on Volumes continuously
			for {
				if terminate {
					break
				}
				if len(snapshotsCreated) > 5 {
					for _, each := range snapshotsCreated {
						err := inspectDeleteVolume(each)
						log.FailOnError(err, "Inspect and Delete Snapshot failed on cluster with snapshot ID [%v]", each)

						// Remove the first element
						for i := 0; i < len(snapshotsCreated)-1; i++ {
							snapshotsCreated[i] = snapshotsCreated[i+1]
						}
						// Resize the array by truncating the last element
						snapshotsCreated = snapshotsCreated[:len(snapshotsCreated)-1]
					}
				}
			}
		}()

		for i := 0; i < 6; i++ {
			// Wait for KVDB Members to be online
			err := WaitForKVDBMembers()
			if err != nil {
				stopRoutine()
				log.FailOnError(err, "failed waiting for KVDB members to be active")
			}

			// Kill KVDB Master Node
			masterNode, err := GetKvdbMasterNode()
			if err != nil {
				stopRoutine()
				log.FailOnError(err, "failed getting details of KVDB master node")
			}

			// Get KVDB Master PID
			pid, err := GetKvdbMasterPID(*masterNode)
			if err != nil {
				stopRoutine()
				log.FailOnError(err, "failed getting PID of KVDB master node")
			}

			log.InfoD("KVDB Master is [%v] and PID is [%v]", masterNode.Name, pid)

			// Kill kvdb master PID for regular intervals
			err = KillKvdbMemberUsingPid(*masterNode)
			if err != nil {
				stopRoutine()
				log.FailOnError(err, "failed to kill KVDB Node")
			}

			// Wait for some time after killing kvdb master Node
			time.Sleep(5 * time.Minute)
		}

		terminate = true
		wg.Wait()
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

// Kubelet stopped on the nodes - and the client container should not be impacted.
var _ = Describe("{StopKubeletOnNodes}", Label("p0", "negative", "px_ops", "error_injection", "node_ops"), func() {

	JustBeforeEach(func() {
		StartTorpedoTest("StopKubeletOnNodes", "Validate PX after kubelet is restarted", nil, 0)

	})
	var contexts []*scheduler.Context

	stepLog := "has to schedule apps and stop kubelet app nodes"
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("kubeletrsrt-%d", i))...)
		}

		ValidateApplications(contexts)

		var pureCleanupFunction func()
		if Inst().V.String() == portworx.PureDriverName {
			pureCleanupFunction = StartPureBackgroundWriteRoutines()
		}

		stepLog = "restart kubelet on all storage driver nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, appNode := range node.GetStorageDriverNodes() {
				stepLog = fmt.Sprintf("restart kubelet %s on node: %v",
					Inst().V.String(), appNode.Name)
				Step(stepLog,
					func() {
						log.InfoD(stepLog)
						RestartKubelet([]node.Node{appNode})
					})
			}
		})

		stepLog = "Validate PX on all nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, node := range node.GetStorageDriverNodes() {
				status, err := IsPxRunningOnNode(&node)
				log.FailOnError(err, fmt.Sprintf("Failed to check if PX is running on node [%s]", node.Name))
				dash.VerifySafely(status, true, fmt.Sprintf("PX is not running on node [%s]", node.Name))
			}
		})

		if pureCleanupFunction != nil {
			pureCleanupFunction() // Checks for any errors during the background writes and fails the test if any occurred
			return
		}

		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		ValidateAndDestroy(contexts, opts)
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

// This test restarts master nodes too and hence this test should be enabled where torpedo is ran using
// ginkgo command on the jenkins agent
var _ = Describe("{KubeClusterRestart}", Label("p1", "negative", "error_injection", "node_ops"), func() {
	var testrailID = 86010
	// testrailID corresponds to: https://portworx.testrail.net/index.php?/cases/view/86010
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("KubeClusterRestart", "Validate shutdown of all the nodes in the k8s cluster and start after 5 mins", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	var contexts []*scheduler.Context

	It("has to setup, validate and teardown apps", func() {

		namespace, err := Inst().V.GetVolumeDriverNamespace()
		log.FailOnError(err, "failed to get volume driver namespace")
		isVsphereSecretExists := false
		isPureSecretExists := false

		if _, err = core.Instance().GetSecret(PX_VSPHERE_SCERET_NAME, namespace); err == nil {
			isVsphereSecretExists = true
		}

		if !isVsphereSecretExists {
			if _, err = core.Instance().GetSecret(PX_PURE_SECRET_NAME, namespace); err == nil {
				isPureSecretExists = true
			}
		}

		if !isVsphereSecretExists && !isPureSecretExists {
			log.InfoD("Skipping the test as it is not on-prem cluster")
			Skip("Skipping the test as it is not on-prem cluster")

		}

		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("setupteardown-%d", i))...)
		}
		ValidateApplications(contexts)

		stepLog = "Powering off all the nodes in cluster"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, vmNode := range node.GetNodes() {
				log.InfoD("Powering off node [%s]", vmNode.Name)
				err := Inst().N.PowerOffVM(vmNode)
				log.FailOnError(err, "error powering off node [%s]", vmNode.Name)
			}
		})

		log.InfoD("Waiting for 5 minutes for the cluster to completely shutdown")
		time.Sleep(5 * time.Minute)
		err = Inst().S.IsNodeReady(node.GetStorageDriverNodes()[0])
		if err != nil {
			log.Infof("Cluster is down as expected")
		}

		stepLog = "Powering on all the nodes in cluster"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, vmNode := range node.GetNodes() {
				log.InfoD("Powering on node [%s]", vmNode.Name)
				err := Inst().N.PowerOnVM(vmNode)
				log.FailOnError(err, "error powering on node [%s]", vmNode.Name)
			}
		})

		masterNodes := node.GetMasterNodes()

		stepLog = "Wait for the master nodes to be back up"

		Step(stepLog, func() {
			for _, masterNode := range masterNodes {
				log.InfoD("Waiting for node [%s] to be back up", masterNode.Name)
				err := Inst().N.TestConnection(masterNode, node.ConnectionOpts{
					Timeout:         15 * time.Minute,
					TimeBeforeRetry: 10 * time.Second,
				})
				log.FailOnError(err, "error while testing node status [%s]", masterNode.Name)
				err = Inst().S.IsNodeReady(masterNode)
				log.FailOnError(err, "error while testing node status [%s]", masterNode.Name)

			}
		})

		stNodes := node.GetStorageDriverNodes()

		stepLog = "Wait for the storage driver nodes to be back up and px ready"
		Step(stepLog, func() {
			for _, stNode := range stNodes {
				log.InfoD("Waiting for node [%s] to be back up", stNode.Name)
				err := Inst().N.TestConnection(stNode, node.ConnectionOpts{
					Timeout:         15 * time.Minute,
					TimeBeforeRetry: 30 * time.Second,
				})
				log.FailOnError(err, "error while testing node status [%s]", stNode.Name)
				err = Inst().S.IsNodeReady(stNode)
				log.FailOnError(err, "error while testing node status [%s]", stNode.Name)
				log.InfoD(fmt.Sprintf("wait to scheduler: %s and volume driver: %s to start on node [%s]",
					Inst().S.String(), Inst().V.String(), stNode.Name))
				err = Inst().V.WaitDriverUpOnNode(stNode, Inst().DriverStartTimeout)
			}
		})

		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		ValidateAndDestroy(contexts, opts)

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{VerifyNoPxRestartDueToPxPodStop}", Label("p1", "positive", "error_injection", "px_ops"), func() {
	JustBeforeEach(func() {
		// https://purestorage.atlassian.net/browse/PTX-24859
		// https://portworx.testrail.net/index.php?/cases/view/300005
		StartTorpedoTest("VerifyNoPxRestartDueToPxPodStop", "Verify that px serivce remain up even if px pod got deleted ", nil, 300005)
	})

	It("Delete px pods and validate px service", func() {
		// Get uptime for px service on each node
		stepLog = "Getting PID of Px process before and after restarting PX pods on all the nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			processPid, err := GetPxPIDMap(node.GetStorageDriverNodes())
			if err != nil {
				log.FailOnError(err, "Failed while getting PID of PX process")
			}

			namespace, err := Inst().S.GetPortworxNamespace()
			log.FailOnError(err, "Error getting portworx namespace")

			//Deleting px pods from all the node
			err = DeletePXPods(namespace)
			log.FailOnError(err, "Error deleting px pods")

			//Capturing PID of PX after stopping PX pods
			processPidPostRestart, err := GetPxPIDMap(node.GetStorageDriverNodes())
			log.FailOnError(err, "Failed while getting PID of PX process")
			log.Infof("Process IDs for px after stopping portworx pod  %s", processPidPostRestart)

			//Verify PID before and after for PX process
			for nodeId, beforePID := range processPid {
				afterPID, _ := processPidPostRestart[nodeId]
				dash.VerifyFatal(beforePID, afterPID, fmt.Sprintf("Validate Process ID of PX process before and after PX pod restart on node %s", nodeId))
			}
		})
	})
	JustAfterEach(func() {
		EndTorpedoTest()

	})
})

// Kubelet stopped on the nodes - and the client container should not be impacted.
var _ = Describe("{PerformStorageVMotions}", Label("p0", "positive", "node_ops", "pool_ops"), func() {

	JustBeforeEach(func() {
		StartTorpedoTest("PerformStorageVMotions", "Perform Storage Vmotion and Validate PX", nil, 0)

	})
	var contexts []*scheduler.Context

	stepLog := "has to schedule apps and perform storage vmotion"
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("svmotion-%d", i))...)
		}

		ValidateApplications(contexts)

		stepLog = "Choosing a single Storage Node randomly and performing SV Motion on it"
		var randomIndex int
		var moveAllDisks bool
		Step(stepLog, func() {
			log.InfoD(stepLog)
			workerNodes := node.GetStorageNodes()
			if len(workerNodes) > 0 {
				randomIndex = rand.Intn(len(workerNodes))
				log.Infof("Selected worker node %v for storage vmotion", workerNodes[randomIndex].Name)
			} else {
				log.FailOnError(fmt.Errorf("no worker nodes available for svmotion"), "No worker nodes available")

			}
			stc, err := Inst().V.GetDriver()
			log.FailOnError(err, "Failed to get storage driver")

			selectedIds := make([]string, 0)

			preData, err := Inst().S.GetPXCloudDriveConfigMap(stc)
			log.FailOnError(err, "Failed to get pre-vMotion cloud drive config")
			diskUUIDMap := make(map[string]string)
			moveAllDisks = rand.Intn(2) == 0
			preNodeConfigs := preData[workerNodes[randomIndex].VolDriverNodeID]
			if moveAllDisks {
				log.Infof("Moving all disks on worker node %v", workerNodes[randomIndex].Name)

				for id, _ := range preNodeConfigs.Configs {
					selectedIds = append(selectedIds, id)
				}
			} else {
				log.Infof("Moving only largest sized disk(s) on worker node %v", workerNodes[randomIndex].Name)
				maxSize := int64(0)
				selectedId := ""
				for id, postDrive := range preNodeConfigs.Configs {
					if postDrive.Size > maxSize {
						selectedId = id
					}
				}
				for id, _ := range preNodeConfigs.Configs {
					if id == selectedId {
						selectedIds = append(selectedIds, id)
					}
				}
			}
			for id, preDriveConfig := range preNodeConfigs.Configs {
				if slices.Contains(selectedIds, id) {
					dsName := preDriveConfig.Labels["datastore"]
					driveProps := strings.Split(preDriveConfig.ID, " ")
					if len(driveProps) < 2 {
						log.FailOnError(fmt.Errorf("invalid drive id %v", preDriveConfig.ID), "Invalid drive id")
					}
					dsPath := driveProps[1]

					driveID := fmt.Sprintf("[%s] %s", dsName, dsPath)
					diskUUIDMap[preDriveConfig.DiskUUID] = driveID
				}
			}
			log.Infof("Ndde [%s],Disk UUIDs to be moved %v", workerNodes[randomIndex].Name, diskUUIDMap)

			var envVariables []v1.EnvVar
			envVariables = stc.Spec.CommonConfig.Env
			var prefixName string
			for _, envVar := range envVariables {
				if envVar.Name == "VSPHERE_DATASTORE_PREFIX" {
					prefixName = envVar.Value
					log.Infof("prefixName   %s ", prefixName)
				}
			}

			ctx := ctxt.Background()
			expectedDatastoreMap, err := Inst().N.StorageVmotion(ctx, workerNodes[randomIndex], prefixName, diskUUIDMap)
			dash.VerifyFatal(err, nil, fmt.Sprintf("validate storage vmotion on node [%s]", workerNodes[randomIndex].Name))
			log.Infof("Waiting for 5 seconds for configmap to be updated")
			time.Sleep(5 * time.Second)

			postData, err := Inst().S.GetPXCloudDriveConfigMap(stc)
			log.FailOnError(err, "Failed to get post-vMotion cloud drive config")
			err = ValidateDatastoreUpdate(diskUUIDMap, preData, postData, workerNodes[randomIndex].VolDriverNodeID, expectedDatastoreMap)
			dash.VerifyFatal(err, nil, fmt.Sprintf("validate datastore update in cloud drive config after storage vmotion on node [%s]", workerNodes[randomIndex].Name))
		})

		stepLog = "Validate PX on all nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, node := range node.GetStorageDriverNodes() {
				status, err := IsPxRunningOnNode(&node)
				log.FailOnError(err, fmt.Sprintf("Failed to check if PX is running on node [%s]", node.Name))
				dash.VerifySafely(status, true, fmt.Sprintf("PX is not running on node [%s]", node.Name))
			}
		})

		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
		ValidateAndDestroy(contexts, opts)
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})
var _ = Describe("{DrainAllNodes}", Label("p2", "positive", "node_ops"), func() {
	/*
			1. Schedule apps
		    2. Pick one node
		    3. Drain it
		    4. Wait for all app pods to get drained from node
		    5. Uncordon node
			6. kill the px on uncordon node
		    6. Pick next node
		    7. Do again steps from 3-5
	*/
	var testrailID = 0
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("DrainAllNodes", "Drain the node wait for all app pods get drained from node", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	stepLog := "Drain all nodes from cluster "
	It("has to drain node wait for all app pods get drained from node", func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)
		log.InfoD("Scheduling Applications")
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("drainnode-%d", i))...)
		}
		namespace := contexts[0].App.NameSpace
		log.Infof("Namespace for the app: %s", namespace)
		ValidateApplications(contexts)
		defer appsValidateAndDestroy(contexts)
		Step("get all nodes and drain one by one", func() {
			k8sOps := k8sCore
			log.InfoD("get all nodes and drain one by one")
			nodesTodrain := node.GetStorageDriverNodes()
			Step(fmt.Sprintf("drain node one at a time from the node(s): %v", nodesTodrain), func() {
				log.InfoD("drain node one at a time from the node(s): %v", nodesTodrain)
				for _, n := range nodesTodrain {
					// Skip  master nodes, for drain
					if node.IsMasterNode(n) {
						log.InfoD("This node [%s] is master, will skip for drain it..", n.Name)
						continue
					}
					stepLog := "Getting pods from node for drain"
					podsUsingStorage := make([]corev1.Pod, 0)
					Step("Getting pods from the node", func() {
						log.InfoD(stepLog)
						k8sOps := k8sCore
						podslist, err := k8sOps.GetPodsByNode(n.Name, namespace)
						log.FailOnNoError(err, "Failed to get pods from node")
						pods := podslist
						log.Infof("Retrieved pods list from %v: %v", n.Name, podslist)
						for _, pod := range pods.Items {
							if pod.Name != "" {
								podsUsingStorage = append(podsUsingStorage, pod)
							}

						}

					})
					Step("Drain all the pods from node", func() {
						log.Infof("Starting to drain all pods from node: '%s'", n.Name)
						timeout := 10 * time.Minute
						DefaultRetryInterval := 10 * time.Second
						err := k8sOps.DrainPodsFromNode(n.Name, podsUsingStorage, timeout, DefaultRetryInterval)
						log.FailOnError(err, "Failed to drain pods from node")

					})
					Step("validate pods are not in node", func() {
						time.Sleep(5 * time.Minute)
						pods, err := k8sOps.GetPodsByNode(n.Name, namespace)
						log.FailOnNoError(err, "Failed to get pods on the node")
						var PodItems []corev1.Pod
						for _, pod := range pods.Items {
							if pod.Name != "" {
								PodItems = append(PodItems, pod)
							}
						}
						//The number of items in PodItems
						log.Infof("Total pods found on node %s: %d", n.Name, len(PodItems))
						// Verify that there are no pods in PodItems
						dash.VerifyFatal(len(PodItems) == 0, true, "Validation : pods exist on the node.")
						log.Infof("Validation succeeded: No running pods found on node %s.", n.Name)
					})

					stepLog = "Uncordon the node"
					Step("Uncordon the node", func() {
						log.InfoD(stepLog)
						err := k8sCore.UnCordonNode(n.Name, 1*time.Minute, 5*time.Second)
						log.FailOnError(err, fmt.Sprintf("Verifying uncordon the node %s", n))
					})
					stepLog = fmt.Sprintf("Kill the PX on the node %v", n.Name)
					Step(stepLog, func() {
						log.InfoD(stepLog)
						CrashVolDriverAndWait([]node.Node{n})
					})

				}
			})
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{TrashcanPVCRestoreByAttachingExistingPVCToPod}", Label("staging", "p0", "px_vol_ops"), func() {
	/*
	   Hazel ticket: https://purestorage.atlassian.net/browse/HAZEL-259
	   1. Enable trashcan
	   2. Delete PVC using kubenrnetes command.
	   3. Restore pvc from trashcan.
	   4. Attach this new PVC to kuberntes pod and make sure readIO can be performed.
	*/
	JustBeforeEach(func() {
		StartTorpedoTest("TrashcanPVCRestoreByAttachingExistingPVCToPod", "Trashcan restore using Kubenetes way by attaching existing PVC to kubernetes pod", nil, 0)
	})

	stepLog := "Trashcan PVC restore and validation by attaching restored PVC to Kubernetes pod"
	It(stepLog, func() {
		log.InfoD(stepLog)

		const (
			PvProvisionedByAnnotation = "pv.kubernetes.io/provisioned-by"
			PVProvisionerName         = k8s.CsiProvisioner
			PVProtectionName          = "kubernetes.io/pv-protection"
			VolumeMode                = "Filesystem"
		)
		var (
			appNamespace = fmt.Sprintf("tc-cs-%s", Inst().InstanceID)
			fioPVcName   = "tc-pvc-restore"
			contexts     = make([]*scheduler.Context, 0)
			allPvcList   *v1.PersistentVolumeClaimList
			scForPvc     *storageApi.StorageClass
			restoredVol  *api.Volume
			trashcanVols []string
		)

		stepLog = "Enabling trashcan feature on the cluster"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			currNode := node.GetStorageDriverNodes()[0]
			err := Inst().V.SetClusterOptsWithConfirmation(currNode, map[string]string{
				"--volume-expiration-minutes": "600",
			})
			log.FailOnError(err, fmt.Sprintf("Failed to enable trashcan feature on the node: %v", currNode.Name))
			log.InfoD("Trashcan feature enabled successfully on the node: %v", currNode.Name)
		})

		stepLog = "scheduling application and retrieve PVCs and storageclass created by the application"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplicationsOnNamespace(appNamespace, fmt.Sprintf("trashrec-%d", i))...)
			}
			for _, ctx := range contexts {
				ctx.SkipVolumeValidation = true
				ValidateContext(ctx)
			}

			allPvcList, err = core.Instance().GetPersistentVolumeClaims(appNamespace, nil)
			log.FailOnError(err, fmt.Sprintf("Failed to retrieve PVCs from application namespace: %v", appNamespace))

			scForPvc, err = k8sCore.GetStorageClassForPVC(&allPvcList.Items[0])
			log.FailOnError(err, fmt.Sprintf("Failed to retrieve SC from PVC: %v", allPvcList.Items[0].Name))

			log.InfoD("Application scheduled successfully on namespace %v", appNamespace)
		})

		stepLog = "Creating and deleting a PVC to validate trashcan functionality"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			pvcObj := &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fioPVcName,
					Namespace: appNamespace,
				},
				Spec: v1.PersistentVolumeClaimSpec{
					AccessModes:      []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
					StorageClassName: &scForPvc.Name,
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceStorage: resource.MustParse("5Gi"),
						},
					},
				},
			}
			createdPVC, err := core.Instance().CreatePersistentVolumeClaim(pvcObj)
			log.FailOnError(err, fmt.Sprintf("Failed to create PVC: %v", pvcObj))

			err = Inst().S.WaitForSinglePVCToBound(createdPVC.Name, createdPVC.Namespace, 3)
			log.FailOnError(err, "Failed to wait for pvc %v to bound in namespace %v", createdPVC.Name, createdPVC.Namespace)

			err = core.Instance().DeletePersistentVolumeClaim(createdPVC.Name, createdPVC.Namespace)
			log.FailOnError(err, fmt.Sprintf("Failed to delete PVC: %v from namespace: %v", createdPVC.Name, createdPVC.Namespace))

			time.Sleep(10 * time.Second)
			log.InfoD("PVC %v deleted successfully", createdPVC.Name)
		})

		stepLog = "Validating that volumes exist in the trashcan"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			node := node.GetStorageDriverNodes()[0]
			trashcanVols, err = Inst().V.GetTrashCanVolumeIds(node)
			log.FailOnError(err, fmt.Sprintf("Failed to retrieve trashcan volumes from the node: %v", node))

			log.Infof("trashcan len: %d", len(trashcanVols))
			dash.VerifyFatal(len(trashcanVols) > 0, true, "Volumes should exist in trashcan")
		})

		stepLog = "Restoring volumes from trashcan and validating restoration"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, tID := range trashcanVols {
				if tID != "" {
					restoredVol, err = Inst().V.InspectVolume(tID)
					log.FailOnError(err, fmt.Sprintf("error inspecting volume %s", tID))
					if strings.Contains(restoredVol.Locator.Name, fioPVcName) {
						err = trashcanRestore(restoredVol.Id, fioPVcName)
						log.FailOnError(err, fmt.Sprintf("error restoring volume %s from trashcan", restoredVol.Id))
					}
				}
			}
			log.InfoD("Volume %v restored successfully", restoredVol.Id)
		})

		stepLog = "Creating PersistentVolume from restored volume"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pvObj := &corev1.PersistentVolume{
				TypeMeta: metav1.TypeMeta{Kind: "PersistentVolume"},
				ObjectMeta: metav1.ObjectMeta{
					Name:      restoredVol.Id,
					Namespace: scForPvc.Namespace,
					Annotations: map[string]string{
						PvProvisionedByAnnotation: PVProvisionerName,
					},
					Finalizers: []string{
						PVProtectionName,
					},
				},

				Spec: corev1.PersistentVolumeSpec{
					StorageClassName: scForPvc.Name,
					AccessModes:      []v1.PersistentVolumeAccessMode{v1.ReadWriteMany},
					Capacity: corev1.ResourceList{
						corev1.ResourceName(corev1.ResourceStorage): resource.MustParse("5Gi"),
					},
					ClaimRef: &corev1.ObjectReference{
						APIVersion: "v1",
						Kind:       "PersistentVolumeClaim",
						Name:       fioPVcName,
						Namespace:  scForPvc.Namespace,
					},
					PersistentVolumeReclaimPolicy: v1.PersistentVolumeReclaimDelete,
					PersistentVolumeSource: v1.PersistentVolumeSource{
						PortworxVolume: &v1.PortworxVolumeSource{
							VolumeID: restoredVol.Id,
						},
					},
					VolumeMode: allPvcList.Items[0].Spec.VolumeMode,
				},
			}
			createdPV, err := k8sCore.CreatePersistentVolume(pvObj)
			log.FailOnError(err, fmt.Sprintf("Failled to create PV %v", pvObj))
			log.InfoD("PV created successfully: %v", createdPV)
		})

		stepLog = "Creating PersistentVolumeClaim and Attaching it to PersistentVolume"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			pvcObj := &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fioPVcName,
					Namespace: allPvcList.Items[0].Namespace,
				},
				Spec: v1.PersistentVolumeClaimSpec{
					AccessModes:      []v1.PersistentVolumeAccessMode{v1.ReadWriteMany},
					StorageClassName: &scForPvc.Name,
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceStorage: resource.MustParse("5Gi"),
						},
					},
				},
			}
			pvc, err := core.Instance().CreatePersistentVolumeClaim(pvcObj)
			log.FailOnError(err, fmt.Sprintf("Failled to create PVC %v", pvcObj))
			log.InfoD("PVC created successfully: %v", pvc)
		})

		stepLog = "Attach the restored PVC to the existing deployment"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			depl, err := apps.Instance().ListDeployments(appNamespace, metav1.ListOptions{})
			log.FailOnError(err, "Failed to list deployments")
			fioDeployment := depl.Items[0]

			newVolume := v1.Volume{
				Name: fioPVcName,
				VolumeSource: v1.VolumeSource{
					PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
						ClaimName: fioPVcName,
					},
				},
			}
			newVolumeMount := v1.VolumeMount{
				Name:      fioPVcName,
				MountPath: "/pvc-restore",
			}

			fioDeployment.Spec.Template.Spec.Volumes = append(fioDeployment.Spec.Template.Spec.Volumes, newVolume)

			containers := fioDeployment.Spec.Template.Spec.Containers
			container := &containers[0]
			container.VolumeMounts = append(container.VolumeMounts, newVolumeMount)

			_, err = apps.Instance().UpdateDeployment(&fioDeployment)
			log.FailOnError(err, fmt.Sprintf("Failled to update the deployment: %v", fioDeployment))
		})

		stepLog = "Validating that all application pods are running"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			waitForPodsRunning := func() (interface{}, bool, error) {
				for _, eachContext := range contexts {
					log.Infof("Verifying Context [%v]", eachContext.App.Key)
					err := Inst().S.WaitForRunning(eachContext, 5*time.Minute, 2*time.Second)
					if err != nil {
						return nil, true, err
					}
				}
				return nil, false, nil
			}
			_, err = task.DoRetryWithTimeout(waitForPodsRunning, 5*time.Minute, 10*time.Second)
			dash.VerifyFatal(err == nil, true, "Check all pods are running")

			ValidateApplications(contexts)
		})

		stepLog = "Destroy apps"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			DestroyApps(contexts, nil)

			err = core.Instance().DeletePersistentVolumeClaim(fioPVcName, appNamespace)
			log.FailOnError(err, fmt.Sprintf("Failled to delete PVC: %v namespace: %v", fioPVcName, appNamespace))

			err = core.Instance().DeletePersistentVolume(restoredVol.Id)
			log.FailOnError(err, fmt.Sprintf("Failled to delete PV: %v namespace: %v", restoredVol.Id, appNamespace))
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{VerifyNoIOInterruptionDuringRunFlatState}", Label("staging", "kvdb_ops", "p1", "negative"), func() {
	/*
		   https://purestorage.atlassian.net/browse/HAZEL-269
			1. Verify pxctl status show that cluster is in run flat state.
			2. Verify deployment of new pod is failed but IO is running on existing pods.
			3. When you bring the KVDB nodes back verify cluster is no more in run-flat state and new deployment is
	*/
	JustBeforeEach(func() {
		StartTorpedoTest("VerifyNoIOInterruptionDuringRunFlatState",
			"Simulate run-flat state, ensuring uninterrupted IO operations and new deployments fail during this state", nil, 0)
	})

	itLog := "Testing uninterrupted IO operations and new deployments fail during run flat state"
	It(itLog, func() {
		log.InfoD(itLog)
		var (
			contexts          []*scheduler.Context
			postRunFlatCtx    []*scheduler.Context
			postPxUpCtx       []*scheduler.Context
			selectedKvdbNodes []KvdbNode
			kvdbNodes         []KvdbNode
		)

		stepLog = "Schedule application"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("beforerunflat-%d", i))...)
			}
		})

		ValidateApplications(contexts)
		defer DestroyApps(contexts, nil)

		stepLog = "Stopping Portworx service on selected KVDB nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			kvdbNodes, err = GetAllKvdbNodes()
			log.FailOnError(err, "Failed to retrieve KVDB nodes")

			selectedKvdbNodes = kvdbNodes[1:]
			log.InfoD("Selected KVDB nodes for PX service stop: %v", selectedKvdbNodes)
			for _, kvdbNode := range selectedKvdbNodes {
				nodeDetails, err := node.GetNodeDetailsByNodeID(kvdbNode.ID)
				log.FailOnError(err, "Unable to retrieve node details for NodeID [%v]", kvdbNode.ID)

				StopVolDriverAndWait([]node.Node{nodeDetails})
				log.InfoD("PX service successfully stopped on node: %v", nodeDetails)

			}
		})

		stepLog = " Verify cluster is in run-flat state"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pxNode, err := node.GetNodeDetailsByNodeID(kvdbNodes[0].ID)
			output, err := runCmd("pxctl status", pxNode)
			log.FailOnError(err, "Failed to execute 'pxctl status' on node: %v", pxNode.Name)

			log.Infof("pxctl status output: %v\n", output)
			expect_out := "Volume and node operations may be unavailable but I/O will continue"
			dash.VerifyFatal(strings.Contains(output, expect_out), true, "Is cluster in run-flat state?")
		})

		log.Infof("validate that existing applications remain in running state during run-flat state")
		ValidateApplications(contexts)

		stepLog = " Verify new deployments fail during the run-flat state"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			postRunFlatCtx = append(postRunFlatCtx, ScheduleApplications(fmt.Sprintf("postrunflat"))...)

			for _, ctx := range postRunFlatCtx {
				err := Inst().S.WaitForRunning(ctx, 5*time.Minute, defaultRetryInterval)
				dash.VerifyFatal(err != nil, true, "Is scheduling app failled in run-flat state?")
			}
		})
		defer func() {
			for _, ctx := range postRunFlatCtx {
				ctx.SkipVolumeValidation = true
				TearDownContext(ctx, nil)
			}
		}()

		stepLog = "Starting Portworx on selected KVDB nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, n := range selectedKvdbNodes {
				nodeDetails, err := node.GetNodeDetailsByNodeID(n.ID)
				log.FailOnError(err, "Failed to retrieve node details for NodeID [%v]", n.ID)

				err = Inst().V.StartDriver(nodeDetails)
				log.FailOnError(err, "Failed to start Portworx driver on node %s", nodeDetails.Name)
				err = Inst().V.WaitDriverUpOnNode(nodeDetails, 10*time.Minute)
				log.FailOnError(err, "Failed to waiting for Portworx driver to start on node %s", nodeDetails.Name)

				log.InfoD("Successfully started Portworx on KVDB node: %v", nodeDetails.Name)
			}
		})

		stepLog = "Verify all KVDB nodes are running and in a healthy state"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, kvdbNode := range kvdbNodes {
				nodeInfo, err := node.GetNodeDetailsByNodeID(kvdbNode.ID)
				log.FailOnError(err, "Failed to get details for KVDB node ID: %s", kvdbNode.ID)
				nodeStatus, err := Inst().V.GetNodeStatus(nodeInfo)
				dash.VerifyFatal(*nodeStatus, opsapi.Status_STATUS_OK, fmt.Sprintf("validate PX status on node %s", kvdbNode.ID))
			}

			storagenode := node.GetStorageNodes()
			kvdbMembers, err := Inst().V.GetKvdbMembers(storagenode[0])
			log.FailOnError(err, "Failed to retrieve KVDB members list")

			err = kvdbutils.ValidateKVDBMembers(kvdbMembers)
			log.FailOnError(err, "Failed to validate KVDB members")

			output, err := runCmd("pxctl status", storagenode[0])
			log.FailOnError(err, "Failed to execute pxctl status on node: %v", storagenode[0].Name)
			dash.VerifyFatal(!strings.Contains(output, "Warning"), true, "Output contains warnings. Is the cluster healthy?")
		})

		stepLog = "Validating that the application created during the run-flat state is now in a running state"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			waitForPodsRunning := func() (interface{}, bool, error) {
				for _, eachContext := range postRunFlatCtx {
					log.Infof("Verifying Context [%v]", eachContext.App.Key)
					err := Inst().S.WaitForRunning(eachContext, 5*time.Minute, 2*time.Second)
					if err != nil {
						return nil, true, err
					}
				}
				return nil, false, nil
			}
			_, err = task.DoRetryWithTimeout(waitForPodsRunning, 30*time.Minute, 10*time.Second)
			dash.VerifyFatal(err == nil, true, "Check all pods are running")

			ValidateApplications(postRunFlatCtx)
		})

		stepLog = "Validating application deployments after KVDB quorum is established"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			postPxUpCtx = append(postPxUpCtx, ScheduleApplications(fmt.Sprintf("afterrunflat"))...)
			defer DestroyApps(postPxUpCtx, nil)
			ValidateApplications(postPxUpCtx)
		})
	})
	JustAfterEach(func() {
		EndTorpedoTest()
		AfterEachTest(contexts)
	})
})
