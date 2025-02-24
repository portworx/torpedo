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

	"github.com/pure-px/torpedo/drivers/scheduler/anthos"
	"github.com/pure-px/torpedo/drivers/scheduler/k8s"
	"github.com/pure-px/torpedo/drivers/scheduler/openshift"
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
	"github.com/pure-px/torpedo/pkg/units"

	. "github.com/onsi/ginkgo/v2"
	"github.com/pure-px/sched-ops/k8s/apps"
	"github.com/pure-px/sched-ops/k8s/core"
	"github.com/pure-px/sched-ops/k8s/operator"
	"github.com/pure-px/sched-ops/task"
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
		// Proceed to write

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

	contexts := make([]*scheduler.Context, 0)
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
		stepLog = "uncordon other nodes and drain the current node and check if pods are failed over to the other node and running fine or not"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			nodes := node.GetStorageDriverNodes()
			for _, node := range nodes[1:] {
				err := Inst().S.EnableSchedulingOnNode(node)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Validate enable scheduling on node %s", node.Name))
			}
			pods, err := k8sCore.GetPodsByNode(nodes[0].Name, "")
			dash.VerifyFatal(err, nil, fmt.Sprintf("Validate get pods on node %s", nodes[0].Name))
			err = k8sCore.DrainPodsFromNode(nodes[0].Name, pods.Items, 5*time.Minute, 10*time.Second)

		})
		stepLog = "Validate Applications after it is failed over to the other nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
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
				log.Infof("Inspect Output [%v]", cVol)
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
		    7. Do again step from 3-5
	*/
	var testrailID = 0
	var runID int
	var contexts = make([]*scheduler.Context, 0)
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
	var (
		contexts = make([]*scheduler.Context, 0)
	)
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
	var (
		contexts = make([]*scheduler.Context, 0)
	)
	JustBeforeEach(func() {
		StartTorpedoTest("VerifyNoIOInterruptionDuringRunFlatState",
			"Simulate run-flat state, ensuring uninterrupted IO operations and new deployments fail during this state", nil, 0)
	})

	itLog := "Testing uninterrupted IO operations and new deployments fail during run flat state"
	It(itLog, func() {
		log.InfoD(itLog)
		var (
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

var _ = Describe("{ValidateKVDBQuorumCheck}", Label("staging", "kvdb_ops", "p1", "negative"), func() {
	/*
		    Ticket id: https://purestorage.atlassian.net/browse/HAZEL-1086
			One of the kvdb nodes is down for over 3 minutes.
			A different node starts to become a kvdb member and adds itself to the etcd cluster as "learner".
			As soon as the new node is seen in kvdb members, stop kvdb on another node
			Validate kvdb is not out of qurum
			Start node which is stopped as step and step 3
			Validate kvdb
	*/

	var testrailID = 0
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("ValidateKVDBQuorumCheck", "Validate Kvdb quorum by adding a new learner node, stopping an existing node, and verifying Kvdb's functionality before and after restarting the stopped node.", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var contexts []*scheduler.Context
	stepLog := "Stop one Kvdb node for 3 minutes, add a learner node, stop another node, validate quorum, restart the stopped node, and verify Kvdb functionality."
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("kvdb-%d", i))...)
		}
		var (
			stopped_px_nodes_ids []string
		)
		ValidateApplications(contexts)
		defer DestroyApps(contexts, nil)
		storagenode := node.GetStorageNodes()
		log.InfoD("Get all KVDB nodes")
		kvdbNodes, err := GetAllKvdbNodes()
		log.FailOnError(err, "Unable to retrieve KVDB nodes")
		stepLog = "Stopping Portworx on one of the KVDB member nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			selectedKvdbNodes := kvdbNodes[0:1]
			log.InfoD("Selected KVDB nodes for PX service stop: %v", selectedKvdbNodes)
			for _, kvdbNode := range selectedKvdbNodes {
				nodeDetails, err := node.GetNodeDetailsByNodeID(kvdbNode.ID)
				log.FailOnError(err, "Unable to retrieve node details for NodeID [%v]", kvdbNode.ID)
				StopVolDriverAndWait([]node.Node{nodeDetails})
				log.InfoD("PX service successfully stopped on node: %v", nodeDetails)
				stopped_px_nodes_ids = append(stopped_px_nodes_ids, kvdbNode.ID)

			}
		})
		stepLog = "Wait for the random amount of time"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			randDuration := time.Duration(rand.Intn(60)+180) * time.Second
			log.Infof("Waiting for a random duration of %v before stopping next KVDB member", randDuration)
			time.Sleep(randDuration)

		})
		stepLog := "Stopping portworx on  another KVDB member node and starting portworx on previous stopped node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			selectedKvdbNodes := kvdbNodes[1:2]
			log.InfoD("Selected KVDB nodes for PX service stop: %v", selectedKvdbNodes)
			for _, kvdbNode := range selectedKvdbNodes {
				nodeDetails, err := node.GetNodeDetailsByNodeID(kvdbNode.ID)
				log.FailOnError(err, "Unable to retrieve node details for NodeID [%v]", kvdbNode.ID)
				StopVolDriverAndWait([]node.Node{nodeDetails})
				log.InfoD("PX service successfully stopped on node: %v", nodeDetails)
				stopped_px_nodes_ids = append(stopped_px_nodes_ids, kvdbNode.ID)
			}
			log.Infof("Waiting for the cluster to stabilize and ensure quorum is not lost...")
			time.Sleep(1 * time.Minute)
			log.InfoD("Check KVDB is in quorum")
			getKVDBNodes, err := GetAllKvdbNodes()
			log.FailOnError(err, "Unable to get kvdb nodes")
			healthyCount := 0
			for _, each := range getKVDBNodes {
				if each.IsHealthy == true {
					healthyCount++
				}
			}
			dash.VerifyFatal(healthyCount >= 2, true, fmt.Sprintf("verify kvdb quorum is not lost.Healthy count: %d, Expected: >=2", healthyCount))
			log.Infof("Starting PX on stopped KVDB member: %s", stopped_px_nodes_ids[0])
			node, err := node.GetNodeDetailsByNodeID(stopped_px_nodes_ids[0])
			err = Inst().V.StartDriver(node)
			log.FailOnError(err, "error starting driver on node %s", node.Name)
			err = Inst().V.WaitDriverUpOnNode(node, 10*time.Minute)
			log.FailOnError(err, "error while waiting for driver up on node %s", node.Name)

		})
		stepLog = "Check KVDB is in quorum"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			getKVDBNodes, err := GetAllKvdbNodes()
			log.FailOnError(err, "Unable to get kvdb nodes")
			healthyCount := 0
			for _, each := range getKVDBNodes {
				if each.IsHealthy == true {
					healthyCount++
				}
			}
			dash.VerifyFatal(healthyCount == 3, true, fmt.Sprintf("verify kvdb quorum is not lost. Healthy count: %d, Expected: 3", healthyCount))
		})
		stepLog = "Restart Portworx on stopped node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			node, err := node.GetNodeDetailsByNodeID(stopped_px_nodes_ids[1])
			err = Inst().V.StartDriver(node)
			log.FailOnError(err, "error starting driver on node %s", node.Name)
			err = Inst().V.WaitDriverUpOnNode(node, 10*time.Minute)
			log.FailOnError(err, "error while waiting for driver up on node %s", node.Name)

		})
		stepLog = "Verified that all KVDB nodes are running and in a healthy state."
		Step(fmt.Sprintf("get kvdb nodes"), func() {
			log.InfoD(stepLog)
			kvdbNodes, err := GetAllKvdbNodes()
			log.FailOnError(err, "Failed to get list of KVDB nodes from the cluster")
			for _, kvdbNode := range kvdbNodes {
				nodeInfo, err := node.GetNodeDetailsByNodeID(kvdbNode.ID)
				log.FailOnError(err, "Unable to get details for node ID: %s", kvdbNode.ID)
				nodeStatus, err := Inst().V.GetNodeStatus(nodeInfo)
				dash.VerifyFatal(*nodeStatus, opsapi.Status_STATUS_OK, fmt.Sprintf("Validate PX status on node %s", kvdbNode.ID))
				kvdbMembers, err := Inst().V.GetKvdbMembers(storagenode[0])
				log.FailOnError(err, "Failed to get kvdb members")
				err = kvdbutils.ValidateKVDBMembers(kvdbMembers)
				log.FailOnError(err, "Failed to validate kvdb members")
				output, err := runCmd("pxctl status", storagenode[0])
				log.FailOnError(err, "Failed to run pxctl status on node: %v", storagenode[0].Name)
				log.Infof("pxctl status output: %v\n", output)
				dash.VerifyFatal(!strings.Contains(output, "Warning"), true, "Output contains warnings. Is the cluster healthy?")
			}

		})
		defer func() {
			log.InfoD("Setting cluster to running by restarting Portworx on stopped nodes")
			for _, n := range stopped_px_nodes_ids {
				node, err := node.GetNodeDetailsByNodeID(n)
				log.FailOnError(err, "Unable to get node details for NodeID [%v]", n)
				err = Inst().V.StartDriver(node)
				log.FailOnError(err, "Error starting driver on node %s", node.Name)
				err = Inst().V.WaitDriverUpOnNode(node, 10*time.Minute)
				log.FailOnError(err, "Error while waiting for driver to be up on node %s", node.Name)
			}
		}()

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{VerifyFstrimWithFastPathVolumes}", Label("staging", "p0", "negative", "px_vol_ops"), func() {
	/*
	   https://purestorage.atlassian.net/browse/HAZEL-1074
	   1. Enabel scheduled FSTrim on the cluster
	   2. Verify scheduled FSTrim with fast path volumes
	*/
	var (
		contexts = make([]*scheduler.Context, 0)
	)
	JustBeforeEach(func() {
		StartTorpedoTest("VerifyFstrimWithFastPathVolumes", "Enable Fstrim Schedule on FastPath Volumes and verify its running", nil, 0)
	})

	stepLog := "Enable Fstrim Schedule on FastPath Volumes and verify its running"
	It(stepLog, func() {
		var (
			storageNodes        []node.Node
			selectedStorageNode node.Node
			volAttachedNode     node.Node
			volumeList          []*opsapi.Volume
		)

		cleanup := func() {
			log.Info("Executing cleanup tasks")
			DestroyApps(contexts, nil)
			_ = Inst().V.SetClusterOpts(selectedStorageNode, map[string]string{"--fstrim-schedule-start": ""})
			err = RemoveLabelsAllNodes(k8s.NodeType, true, false)
			log.FailOnError(err, "error removing label on all nodes")
		}
		defer cleanup()

		stepLog = "Add label on the selected storage node and Schedule application"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			// Select storage node
			storageNodes = node.GetStorageNodes()
			selectedStorageNode = GetRandomNode(storageNodes)
			log.Infof("The Selected node for Fast path label is %v : ", selectedStorageNode.Name)

			// Add label on the selected node
			err = Inst().S.AddLabelOnNode(selectedStorageNode, k8s.NodeType, k8s.FastpathNodeType)
			log.FailOnError(err, fmt.Sprintf("Failed add label on node %s", selectedStorageNode.Name))

			// Deploy application on the selected node
			contexts = make([]*scheduler.Context, 0)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				taskName := fmt.Sprintf("fastpath-%d", i)
				appSpec := "fio-fastpath-repl1"
				provisioner := Inst().Provisioner
				contexts = append(contexts, ScheduleApplicationsWithScheduleOptions(taskName, appSpec, provisioner)...)
			}
		})

		ValidateApplications(contexts)

		stepLog = "Get app volumes and Check fast path is active on the node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, ctx := range contexts {
				stepLog = fmt.Sprintf("get volumes for %s app", ctx.App.Key)
				Step(stepLog, func() {
					log.InfoD(stepLog)
					appVolumes, err := Inst().S.GetVolumes(ctx)
					log.FailOnError(err, "Failed to get volumes for app %s", ctx.App.Key)
					dash.VerifyFatal(len(appVolumes) > 0, true, "App volumes exist?")
					log.Infof("App volumes details: %v ", appVolumes)

					// Loop through the apps and check if the volumes are fastpath active before reboot
					for _, appvolume := range appVolumes {
						log.Infof("current volume : %v", appvolume.Name)
						if strings.Contains(ctx.App.Key, fastpathAppName) {
							err := ValidateFastpathVolume(ctx, opsapi.FastpathStatus_FASTPATH_ACTIVE)
							log.FailOnError(err, "fastpath volume validation failed for the volume %v", appvolume.Name)
						}
					}
				})
			}
		})

		stepLog = fmt.Sprintf("Enabel nodiscard and scheduled fstrim on fast path volumes")
		Step(stepLog, func() {
			log.InfoD(stepLog)

			stepLog = fmt.Sprintf("Enabel nodiscard on fast path volumes")
			Step(stepLog, func() {
				log.InfoD(stepLog)
				for _, ctx := range contexts {
					appVolumes, err := Inst().S.GetVolumes(ctx)
					log.FailOnError(err, "Failed to get volumes for app %s", ctx.App.Key)

					for _, appvolume := range appVolumes {
						apivol, err := Inst().V.InspectVolume(appvolume.ID)
						log.FailOnError(err, "Failed to inspect the volume: %v", appvolume.ID)
						volumeList = append(volumeList, apivol)

						attachedNode := apivol.AttachedOn
						volAttachedNode, err = node.GetNodeByIP(attachedNode)
						log.FailOnError(err, "Failed to get the node details by IP: %v", attachedNode)
						log.Infof("FastPath volume attached on the node: %v", volAttachedNode.Name)

						EnableNodiscardOnVolume(volAttachedNode, appvolume)
					}
				}
			})

			stepLog = fmt.Sprintf("Enabel scheduled fs trim for fast path volumes")
			Step(stepLog, func() {
				log.InfoD(stepLog)
				// Daily: daily=HH:MM, Weekly: weekly=day@hh:mm
				log.Infof("Enable scheduled fs trim on the cluster")
				formattedTime := time.Now().UTC().Add(1 * time.Minute).Format("15:04")
				scheduleStartTime := fmt.Sprintf("daily=%s", formattedTime)
				err := EnableScheduledFSTrim(storageNodes, 10, scheduleStartTime)
				log.FailOnError(err, "failed to enable scheduled fs trim")
			})
		})

		stepLog = fmt.Sprintf("verify FSTrim on fast path valume is running")
		Step(stepLog, func() {
			log.InfoD(stepLog)

			for _, vol := range volumeList {
				attachedNode := vol.AttachedOn
				volAttachedNode, err = node.GetNodeByIP(attachedNode)
				log.FailOnError(err, "Failed to get the node details by IP: %v", attachedNode)

				fsTrimStatuses, err := CheckFSTrimRunningOnNode(volAttachedNode, vol, 30*time.Minute, 2*time.Minute)
				log.FailOnError(err, "error getting fstrim status for the volume %v", vol.Id)

				for vol, status := range fsTrimStatuses {
					dash.VerifySafely(status != opsapi.FilesystemTrim_FS_TRIM_FAILED, true, fmt.Sprintf("verify autofstrim for volume %s, current status %v", vol, status))
				}
			}
		})

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{DeleteVolumeWhenClusterInRunFlatState}", Label("staging", "kvdb_ops", "p1", "negative"), func() {
	/*
		   Ticket ID: https://purestorage.atlassian.net/browse/HAZEL-1041
			Create Apps
			Bring cluster in run flat state
			Delete apps → Should succeed (Use Tear Down Context method)
			Bring up kvdb nodes again

	*/
	JustBeforeEach(func() {
		StartTorpedoTest("DeleteVolumeWhenClusterInRunFlatState",
			"Simulate run-flat state, delete the apps validate it should succeed and bring the nodes again", nil, 0)
	})
	var (
		contexts = make([]*scheduler.Context, 0)
	)

	itLog := "Simulate run-flat state, delete the apps validate it should succeed and bring the nodes again"
	It(itLog, func() {
		log.InfoD(itLog)
		var (
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

		cleanup := func() {
			log.Info("Executing cleanup tasks")
			if len(selectedKvdbNodes) > 0 {
				for _, n := range selectedKvdbNodes {
					nodeDetails, err := node.GetNodeDetailsByNodeID(n.ID)
					log.FailOnError(err, "Failed to retrieve node details for NodeID [%v]", n.ID)

					err = Inst().V.StartDriver(nodeDetails)
					log.FailOnError(err, "Failed to start Portworx driver on node %s", nodeDetails.Name)
					err = Inst().V.WaitDriverUpOnNode(nodeDetails, 10*time.Minute)
					log.FailOnError(err, "Failed to waiting for Portworx driver to start on node %s", nodeDetails.Name)
					log.InfoD("Successfully started Portworx on KVDB node: %v", nodeDetails.Name)
				}
			}
			DestroyApps(contexts, nil)
		}
		defer cleanup()

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

		DestroyApps(contexts, nil)
		log.Infof("Successfully destroyed the applications in the cluster. Context: %v", contexts)

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
	})

	JustAfterEach(func() {
		EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{ValidateKVDBFailOver}", Label("staging", "kvdb_ops", "p1", "negative"), func() {
	/*
		https://purestorage.atlassian.net/browse/HAZEL-1031
		Pre-requisities: Atleast 7 nodes in the cluster
		1.Create some app
		2.Validate kvdb is running on these nodes
		3.Kill the px on the 2 of these kvdb nodes
		4.Validate apps
		5.Once nodes come up, validate apps
	*/

	var testrailID = 0
	var runID int
	JustBeforeEach(func() {
		StartTorpedoTest("ValidateKVDBFailOver", "Identify KVDB nodes, stop PX on two KVDB nodes, validate failover, restart PX, and validate apps..", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	var (
		contexts             []*scheduler.Context
		stopped_px_nodes_ids []string
	)
	stepLog := "Identify KVDB nodes, stop PX on two KVDB nodes, validate failover, restart PX, and validate apps.."
	It(stepLog, func() {
		log.InfoD(stepLog)
		storageNodes := node.GetStorageNodes()
		log.Infof("Storage node in the cluster: [%v]", storageNodes)
		if len(storageNodes) < 7 {
			Skip("At least 7 nodes are required to run the tests")
		}
		log.InfoD("Get all KVDB nodes")
		kvdbNodes, err := GetAllKvdbNodes()
		log.FailOnError(err, "Unable to retrieve KVDB nodes")

		stepLog := "Schedule application"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			contexts = make([]*scheduler.Context, 0)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("kvdbfailover-%d", i))...)
			}
			ValidateApplications(contexts)
		})

		cleanup := func() {
			log.Info("Executing cleanup tasks")
			for _, px_stop_node := range stopped_px_nodes_ids {
				node, err := node.GetNodeDetailsByNodeID(px_stop_node)
				log.FailOnError(err, "Unable to get the node details for the node  [%v]", node)
				err = Inst().V.StartDriver(node)
				log.FailOnError(err, "error starting driver on node %s", node.Name)
				err = Inst().V.WaitDriverUpOnNode(node, 10*time.Minute)
				log.FailOnError(err, "error while waiting for driver up on node %s", node.Name)
				log.Infof("Successfully start the portworx :[%v]", px_stop_node)
			}
			DestroyApps(contexts, nil)
		}
		defer cleanup()
		stepLog = "Kill the Portworx on KVDB member nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			Restart_px_on_two_kvdb_node := kvdbNodes[0:2]
			log.Infof("KVDB nodes for Restarting the Portworx: %v", Restart_px_on_two_kvdb_node)
			errCh := make(chan error, len(Restart_px_on_two_kvdb_node))
			for _, kvdbNode := range Restart_px_on_two_kvdb_node {
				go func(kvdbNode KvdbNode) {
					nodeDetails, err := node.GetNodeDetailsByNodeID(kvdbNode.ID)
					if err != nil {
						errCh <- fmt.Errorf("Unable to retrieve node details for NodeID [%v]: %w", kvdbNode.ID, err)
						return
					}
					log.Info("Stop the volume driver and wait for it to stop completely")
					StopVolDriverAndWait([]node.Node{nodeDetails})
					log.InfoD("PX service successfully stopped on node: %v", nodeDetails)
					stopped_px_nodes_ids = append(stopped_px_nodes_ids, kvdbNode.ID)
					log.Infof("Restarting PX on node: %v", nodeDetails)
					err = Inst().V.StartDriver(nodeDetails)
					if err != nil {
						errCh <- fmt.Errorf("Error starting driver on node %v: %w", nodeDetails.Name, err)
						return
					}
					err = Inst().V.WaitDriverUpOnNode(nodeDetails, 10*time.Minute)
					if err != nil {
						errCh <- fmt.Errorf("Error while waiting for driver up on node %v: %w", nodeDetails.Name, err)
						return
					}

					log.Infof("Successfully restarted Portworx on node: %v", nodeDetails)
					errCh <- nil
				}(kvdbNode)
			}
			for i := 0; i < len(Restart_px_on_two_kvdb_node); i++ {
				err := <-errCh
				if err != nil {
					log.FailOnError(err, "Error occurred during PX service restart")
				}
			}

			log.Infof("All PX services have been successfully restarted on nodes: %v", stopped_px_nodes_ids)
		})

		stepLog = "Check KVDB is in quorum"
		Step(stepLog, func() {
			checkKVDBQuorum := func() (interface{}, bool, error) {
				log.Infof("verifying that the quorum is intact")
				getKVDBNodes, err := GetAllKvdbNodes()
				if err != nil {
					return nil, true, fmt.Errorf("unable to get KVDB nodes: %w", err)
				}
				log.Infof("KVDB node details: %v", getKVDBNodes)
				healthyCount := 0
				for _, each := range getKVDBNodes {
					if each.IsHealthy == true {
						healthyCount++
					}
				}
				if healthyCount == 3 {
					log.Infof("KVDB quorum intact. Healthy count: %d, Expected: 3", healthyCount)
					return nil, false, nil
				}
				log.Errorf("KVDB quorum lost. Healthy count: %d, Expected: 3. Retrying...", healthyCount)
				return nil, true, fmt.Errorf("quorum lost. Healthy count: %d, Expected: 3", healthyCount)
			}
			_, err = task.DoRetryWithTimeout(checkKVDBQuorum, 5*time.Minute, 30*time.Second)
			log.FailOnError(err, "Error occurred during KVDB quorum check")
			log.Infof("KVDB quorum check completed successfully.")

		})

		ValidateApplications(contexts)
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{ValidateSecondKVDBFailOverWithOnlyThreeNodeLabel}", Label("staging", "kvdb_ops", "p1", "negative"), func() {
	/*
		    ticket id: https://purestorage.atlassian.net/browse/HAZEL-1029
			Pre-requisities: Atleast 7 nodes in the cluster
			Label only 3 nodes with metadata=true that are currently not a kvdb member
			Label rest of the nodes as metadata=false that are currently not a kvdb member. We need to restart Px on these nodes.
			Create some apps
			IDentify current 3 nodes that have kvdb cluster
			Long stop Px on all 3 current kvdb nodes
			Validate kvdb cluster has failed over to some other nodes
			Label nodes from Step 4 as metadata=false
			Bring up Px on the nodes from Step 4
			Again long stop Px on any one of the kvdb member
			Validate kvdb cluster have only two members
			Validate apps
	*/

	var testrailID = 0
	var runID int
	var (
		contexts = make([]*scheduler.Context, 0)
	)
	JustBeforeEach(func() {
		StartTorpedoTest("ValidateSecondKVDBFailOverWithOnlyThreeNodeLabel", "KVDB cluster failover and quorum maintenance by stopping and restarting PX on KVDB nodes, verifying cluster health, and ensuring app availability.", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	stepLog := "Label only 3 nodes with px/metadata=true, stop PX on KVDB nodes to validate failover, restart PX, stop one KVDB node again, with two kvdb member and validate app functionality.."
	It(stepLog, func() {
		log.InfoD(stepLog)
		var (
			stopped_px_kvdb_nodes_ids           []string
			nonKvdbNodes                        []node.Node
			selected_node_for_px_stop           KvdbNode
			selectedNodesNotInKvdbForLabelTrue  []node.Node
			selectedNodesNotInKvdbForLabelFalse []node.Node
		)
		storageNodes := node.GetStorageNodes()
		log.Infof("Storage node in the cluster: [%v]", storageNodes)
		if len(storageNodes) < 7 {
			Skip("At least 7 nodes are required to run the tests")
		}

		log.InfoD("Get all KVDB nodes")
		kvdbNodes, err := GetAllKvdbNodes()
		log.FailOnError(err, "Unable to retrieve KVDB nodes")
		log.Infof("Initally kvdb node in the cluster: [%v]", kvdbNodes)

		cleanup := func() {
			log.Info("Executing cleanup tasks")
			if len(selectedNodesNotInKvdbForLabelTrue) > 0 {
				for _, node := range selectedNodesNotInKvdbForLabelTrue {
					err := Inst().S.RemoveLabelOnNode(node, "px/metadata-node")
					log.FailOnError(err, "Failed to remove label  'px/metadata-node=true' from node [%s]", node)
					log.Infof("Successfully removed label from node [%s]", node)
				}
			}
			if len(selectedNodesNotInKvdbForLabelFalse) > 0 {
				for _, node := range selectedNodesNotInKvdbForLabelFalse {
					err := Inst().S.RemoveLabelOnNode(node, "px/metadata-node")
					log.FailOnError(err, "Failed to remove label  'px/metadata-node=false' from node [%s]", node)
					log.Infof("Successfully removed label from node [%s]", node)
				}
			}
			if len(stopped_px_kvdb_nodes_ids) > 0 {
				for _, px_stop_node := range stopped_px_kvdb_nodes_ids {
					node, err := node.GetNodeDetailsByNodeID(px_stop_node)
					log.FailOnError(err, "Unable to get the node details for the node  [%v]", node)
					err = Inst().V.StartDriver(node)
					log.FailOnError(err, "error starting driver on node %s", node.Name)
					err = Inst().V.WaitDriverUpOnNode(node, 10*time.Minute)
					log.FailOnError(err, "error while waiting for driver up on node %s", node.Name)
					log.Infof("Successfully start the portworx :[%v]", px_stop_node)
				}
			}
			if selected_node_for_px_stop.ID != "" {
				node, err := node.GetNodeDetailsByNodeID(selected_node_for_px_stop.ID)
				log.FailOnError(err, "Unable to get the node details for the node  [%v]", node)
				err = Inst().V.StartDriver(node)
				log.FailOnError(err, "error starting driver on node %s", node.Name)
				err = Inst().V.WaitDriverUpOnNode(node, 10*time.Minute)
				log.FailOnError(err, "error while waiting for driver up on node %s", node.Name)
				log.Infof("Successfully start the portworx :[%v]", selected_node_for_px_stop)

			}
			DestroyApps(contexts, nil)
		}
		defer cleanup()

		stepLog = "Getting non KVDB nodes in the cluster"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			kvdbNodeMap := make(map[string]bool)
			for _, kvdbNode := range kvdbNodes {
				kvdbNodeMap[kvdbNode.ID] = true
			}
			for _, storageNode := range storageNodes {
				if _, exists := kvdbNodeMap[storageNode.Id]; !exists {
					nonKvdbNodes = append(nonKvdbNodes, storageNode)
				}
			}
			log.Infof("All storage nodes List are part of KVDB members: [%v]", nonKvdbNodes)

		})
		selectedNodesNotInKvdbForLabelTrue = nonKvdbNodes[0:3]
		selectedNodesNotInKvdbForLabelFalse = nonKvdbNodes[3:len(nonKvdbNodes)]

		stepLog = "Label only 3 nodes with px/metadata-node=true"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.Infof("Selected nodes for labeling: %v", selectedNodesNotInKvdbForLabelTrue)
			for _, Non_kvdb_label_true_node := range selectedNodesNotInKvdbForLabelTrue {
				err := Inst().S.AddLabelOnNode(Non_kvdb_label_true_node, "px/metadata-node", "true")
				log.FailOnError(err, "Failed to add label 'px/metadata-node=true' to node [%s]", Non_kvdb_label_true_node)

			}
		})

		stepLog = "Label remaining nodes with px/metadata-node=false"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.Infof("Selected nodes for labeling: %v", selectedNodesNotInKvdbForLabelFalse)
			for _, Non_kvdb_label_false_node := range selectedNodesNotInKvdbForLabelFalse {
				err := Inst().S.AddLabelOnNode(Non_kvdb_label_false_node, "px/metadata-node", "false")
				log.FailOnError(err, "Failed to add label 'px/metadata-node=true' to node [%s]", Non_kvdb_label_false_node)
				log.Info("Stop the volume driver and wait for it to stop completely")
				StopVolDriverAndWait([]node.Node{Non_kvdb_label_false_node})
				log.InfoD("PX service successfully stopped on node: %v", Non_kvdb_label_false_node)
				log.Infof("Restarting PX on node: %v", Non_kvdb_label_false_node)
				err = Inst().V.StartDriver(Non_kvdb_label_false_node)
				log.FailOnError(err, "error starting driver on node %s", Non_kvdb_label_false_node.Name)
				err = Inst().V.WaitDriverUpOnNode(Non_kvdb_label_false_node, 10*time.Minute)
				log.FailOnError(err, "error while waiting for driver up on node %s", Non_kvdb_label_false_node.Name)
				log.Infof("Successfully restart the portworx :[%v]", Non_kvdb_label_false_node.Name)
			}

		})

		stepLog := "Schedule application"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			contexts = make([]*scheduler.Context, 0)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("kvdbfailover-%d", i))...)
			}
			ValidateApplications(contexts)
		})

		stepLog = "Stopping Portworx on KVDB member node and waiting for labeled node to join"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			log.Infof("KVDB nodes for PX service stop: %v", kvdbNodes)
			for _, kvdbNode := range kvdbNodes {
				nodeDetails, err := node.GetNodeDetailsByNodeID(kvdbNode.ID)
				log.FailOnError(err, "Unable to retrieve node details for NodeID [%v]", kvdbNode.ID)
				StopVolDriverAndWait([]node.Node{nodeDetails})
				log.InfoD("PX service successfully stopped on node: %v", nodeDetails)
				checkKVDBQuorum := func() (interface{}, bool, error) {
					log.Infof("Checking if the labeled node has joined and verifying that the quorum is intact before stopping PX on the next node.")
					getKVDBNodes, err := GetAllKvdbNodes()
					if err != nil {
						return nil, true, fmt.Errorf("unable to get KVDB nodes: %w", err)
					}
					log.Infof("KVDB node details: %v", getKVDBNodes)
					healthyCount := 0
					for _, each := range getKVDBNodes {
						if each.IsHealthy == true {
							healthyCount++
						}
					}
					if healthyCount == 3 {
						log.Infof("KVDB quorum intact. Healthy count: %d, Expected: 3", healthyCount)
						return nil, false, nil
					}
					log.Errorf("KVDB quorum lost. Healthy count: %d, Expected: 3. Retrying...", healthyCount)
					return nil, true, fmt.Errorf("quorum lost. Healthy count: %d, Expected: 3", healthyCount)
				}
				_, err = task.DoRetryWithTimeout(checkKVDBQuorum, 5*time.Minute, 30*time.Second)
				log.FailOnError(err, "Error occurred while checking KVDB quorum")
				stopped_px_kvdb_nodes_ids = append(stopped_px_kvdb_nodes_ids, kvdbNode.ID)
				log.Infof("PX service stopped successfully on KVDB node [%v]", stopped_px_kvdb_nodes_ids)
			}
		})

		stepLog = "Label KVDB nodes with px/metadata-node=false"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.Infof("Selected nodes for labeling: %v", stopped_px_kvdb_nodes_ids)
			for _, kvdb_node_label_false := range stopped_px_kvdb_nodes_ids {
				nodeDetails, err := node.GetNodeDetailsByNodeID(kvdb_node_label_false)
				log.FailOnError(err, "Unable to retrieve node details for NodeID [%v]", kvdb_node_label_false)
				err = Inst().S.AddLabelOnNode(nodeDetails, "px/metadata-node", "false")
				log.FailOnError(err, "Failed to add label 'px/metadata-node=true' to node [%s]", kvdb_node_label_false)
				selectedNodesNotInKvdbForLabelFalse = append(selectedNodesNotInKvdbForLabelFalse, nodeDetails)

			}
		})

		stepLog = "Restart Portworx on stopped kvdb nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, px_stop_node := range stopped_px_kvdb_nodes_ids {
				node, err := node.GetNodeDetailsByNodeID(px_stop_node)
				log.FailOnError(err, "Unable to get the node details for the node  [%v]", node)
				err = Inst().V.StartDriver(node)
				log.FailOnError(err, "error starting driver on node %s", node.Name)
				err = Inst().V.WaitDriverUpOnNode(node, 10*time.Minute)
				log.FailOnError(err, "error while waiting for driver up on node %s", node.Name)
				log.Infof("Successfully start the portworx :[%v]", px_stop_node)
			}

		})

		stepLog = "Check the selected labeled nodes are in kvdb members"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			getKVDBNodes, err := GetAllKvdbNodes()
			log.FailOnError(err, "Unable to get kvdb nodes")
			log.Infof("After kvdb members Labeled : [%v]", getKVDBNodes)
			kvdbNodeMap := make(map[string]bool)
			failedNodes := []string{}
			for _, kvdbNode := range getKVDBNodes {
				kvdbNodeMap[kvdbNode.ID] = true
				log.Infof("KVDB member found: NodeID [%s]", kvdbNode.ID)
			}
			allKvdbMembers := true
			for _, node := range selectedNodesNotInKvdbForLabelTrue {
				if _, exists := kvdbNodeMap[node.Id]; !exists {
					log.Infof("Node [%s] is NOT a KVDB member.", node.Id)
					failedNodes = append(failedNodes, node.Id)
					allKvdbMembers = false
				}
				log.Infof("Node [%s] is a KVDB member.", node.Id)
			}
			if !allKvdbMembers {
				err := fmt.Errorf("Selected nodes %v are not part of the KVDB cluster", failedNodes)
				log.FailOnError(err, "Selected nodes are not part of the KVDB cluster")
			}
			log.InfoD("All selected nodes are KVDB members")
		})

		stepLog = "Initiating second KVDB failover by stopping PX on one of the KVDB member nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			getKVDBNodes, err := GetAllKvdbNodes()
			log.FailOnError(err, "Unable to get KVDB nodes")
			selected_node_for_px_stop = getKVDBNodes[0]
			log.Infof("Selected node for PX stop: %v", selected_node_for_px_stop)
			node_details, err := node.GetNodeDetailsByNodeID(selected_node_for_px_stop.ID)
			log.FailOnError(err, "Unable to retrieve node details for NodeID [%v]", selected_node_for_px_stop.ID)
			StopVolDriverAndWait([]node.Node{node_details})
			log.InfoD("PX service successfully stopped on node: %v", selected_node_for_px_stop)
			healthyCount := 0
			checkKVDBQuorum := func() (interface{}, bool, error) {
				getKVDBNodes, err := GetAllKvdbNodes()
				if err != nil {
					return nil, true, fmt.Errorf("unable to get KVDB nodes: %w", err)
				}
				log.Infof("KVDB node details: %v", getKVDBNodes)
				for _, each := range getKVDBNodes {
					if each.IsHealthy == true {
						healthyCount++
					}
				}
				if healthyCount == 2 {
					log.Infof("KVDB quorum intact. Healthy count: %d, Expected: 2", healthyCount)
					return nil, false, nil
				}
				log.Errorf("KVDB quorum lost. Healthy count: %d, Expected: 2. Retrying...", healthyCount)
				return nil, true, fmt.Errorf("quorum lost. Healthy count: %d, Expected: 2", healthyCount)
			}
			_, err = task.DoRetryWithTimeout(checkKVDBQuorum, 5*time.Minute, 30*time.Second)
			log.FailOnError(err, "Error occurred while checking KVDB quorum")
			dash.VerifyFatal(healthyCount == 2, true, fmt.Sprintf("verify kvdb quorum is not lost.Healthy count: %d, Expected: 2", healthyCount))
		})

		ValidateApplications(contexts)

		stepLog = "Restart Portworx on stopped kvdb node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			node, err := node.GetNodeDetailsByNodeID(selected_node_for_px_stop.ID)
			log.FailOnError(err, "Unable to get the node details for the node  [%v]", node)
			err = Inst().V.StartDriver(node)
			log.FailOnError(err, "error starting driver on node %s", node.Name)
			err = Inst().V.WaitDriverUpOnNode(node, 10*time.Minute)
			log.FailOnError(err, "error while waiting for driver up on node %s", node.Name)
			log.Infof("Successfully start the portworx :[%v]", selected_node_for_px_stop)
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

// Restart all the px nodes after killing leader node of etcd in etcd cluster.
var _ = Describe("{RestartPxNodesAfterKVDBMasterFailure}", Label("p1", "kvdb_ops", "negative", "staging"), func() {
	/*
		Create few apps
		Kill kvdb leader node
		Do app validation
		Restart more than half of Px nodes
		Do app validation
		Wait for all nodes to come up again
		Do app validation
	*/

	var (
		testrailID   = 36147373
		runID        int
		contexts     []*scheduler.Context
		wg           sync.WaitGroup
		storageNodes = []node.Node{}
	)
	JustBeforeEach(func() {
		StartTorpedoTest("RestartPxNodesAfterKVDBMasterFailure", "Restart all the px nodes after killing leader node of etcd in etcd cluster", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	stepLog := "Restart all the px nodes after killing leader node of etcd in etcd cluster"
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("kvdb-%d", i))...)
		}
		ValidateApplications(contexts)
		defer DestroyApps(contexts, nil)

		stepLog = "Killing leader node of etcd(kvdb)"
		Step(stepLog, func() {
			masterNode, err := GetKvdbMasterNode()
			log.FailOnError(err, "failed getting details of KVDB master node")

			pid, err := GetKvdbMasterPID(*masterNode)
			log.FailOnError(err, "failed getting PID of KVDB master node")

			log.InfoD("KVDB Master is [%v] and PID is [%v]", masterNode.Name, pid)
			err = KillKvdbMemberUsingPid(*masterNode)
			log.FailOnError(err, "failed to kill KVDB Node")
			// Wait for KVDB Members to be online
			log.FailOnError(WaitForKVDBMembers(), "failed waiting for KVDB members to be active")
		})

		stepLog = "validate applications after killing KVDB Master node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, ctx := range contexts {
				wg.Add(1)
				go func(c *scheduler.Context) {
					defer wg.Done()
					ValidateContext(c)
				}(ctx)
			}
			wg.Wait()
		})

		stepLog = "Restart more than half of Px nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			storageNodes = node.GetStorageNodes()
			nodesToRestartPx := storageNodes[:(len(storageNodes)/2)+1]
			wg.Add(1)
			go func() {
				defer wg.Done()
				log.Infof("Stop volume driver [%s] on nodes: [%v]", Inst().V.String(), nodesToRestartPx)
				StopVolDriverAndWait(nodesToRestartPx)
				log.Infof("Starting volume driver [%s] on nodes [%v]", Inst().V.String(), nodesToRestartPx)
				StartVolDriverAndWait(nodesToRestartPx)
			}()
			// Wait for both steps to complete
			wg.Wait()
		})

		stepLog = "validate applications after restarting px nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, ctx := range contexts {
				wg.Add(1)
				go func(c *scheduler.Context) {
					defer wg.Done()
					ValidateContext(c)
				}(ctx)
			}
			wg.Wait()
		})

		stepLog = "Verified that all KVDB nodes are running and in a healthy state."
		Step(fmt.Sprintf("Get kvdb nodes"), func() {
			log.InfoD(stepLog)
			kvdbNodes, err := GetAllKvdbNodes()
			log.FailOnError(err, "Failed to get list of KVDB nodes from the cluster")
			for _, kvdbNode := range kvdbNodes {
				nodeInfo, err := node.GetNodeDetailsByNodeID(kvdbNode.ID)
				log.FailOnError(err, "Unable to get details for node ID: %s", kvdbNode.ID)
				nodeStatus, err := Inst().V.GetNodeStatus(nodeInfo)
				dash.VerifyFatal(*nodeStatus, opsapi.Status_STATUS_OK, fmt.Sprintf("validate PX status on node %s", kvdbNode.ID))
			}
			kvdbMembers, err := Inst().V.GetKvdbMembers(storageNodes[0])
			log.FailOnError(err, "Failed to get kvdb members")
			err = kvdbutils.ValidateKVDBMembers(kvdbMembers)
			log.FailOnError(err, "Failed to validate kvdb members")
			output, err := runCmd("pxctl status", storageNodes[0])
			log.FailOnError(err, "Failed to run pxctl status on node: %v", storageNodes[0].Name)
			log.Infof("pxctl status output: %v\n", output)
			dash.VerifyFatal(!strings.Contains(output, "Warning"), true, "Output contains warnings. Is the cluster healthy?")
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{ValidateSecondKVDBFailOver}", Label("staging", "kvdb_ops", "p1", "negative"), func() {
	/*
		    ticket id: https://purestorage.atlassian.net/browse/HAZEL-1030
			Pre-requisities: Atleast 7 nodes in the cluster
			Label 4 nodes with metadata=true & kill px on all these nodes
			Remaning nodes label metadata=false & kill px on all these nodes
			Create apps
			IDentify current 3 nodes that have kvdb cluster
			Long stop Px on all 3 current kvdb nodes
			Validate kvdb cluster has failed over to some other nodes
			Label kvdb node with metadata=false
			Bring up Px on the nodes from Step 4
			Again long stop Px on any one of the kvdb member
			Validate kvdb cluster becomes healthy again as some other node gets chosen as kvdb member
			Bring up Px on failed node from step 7
			Validate apps
	*/

	var testrailID = 0
	var runID int
	var (
		contexts = make([]*scheduler.Context, 0)
	)
	JustBeforeEach(func() {
		StartTorpedoTest(" ValidateSecondKVDBFailOver", "Testing KVDB cluster failover, quorum, and recovery under node failures while ensuring application availability.", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	stepLog := "Label 4 nodes with px/metadata=true, stop PX on KVDB nodes to validate failover, restart PX, stop one KVDB node again, verify fourth node joins the cluster, and validate app functionality."
	It(stepLog, func() {
		log.InfoD(stepLog)
		var (
			stopped_px_kvdb_nodes_ids           []string
			nonKvdbNodes                        []node.Node
			selectedNodesNotInKvdbForLabelTrue  []node.Node
			selectedNodesNotInKvdbForLabelFalse []node.Node
			selected_node_for_px_stop           KvdbNode
			nonKvdbNodeAfterLabel               node.Node
			healthyCount                        int
		)
		storageNodes := node.GetStorageNodes()
		log.Infof("Storage node in the cluster: [%v]", storageNodes)
		if len(storageNodes) < 7 {
			Skip("At least 7 nodes are required to run the tests")
		}
		log.InfoD("Get all KVDB nodes")
		kvdbNodes, err := GetAllKvdbNodes()
		log.FailOnError(err, "Unable to retrieve KVDB nodes")
		log.Infof("Initally kvdb node in the cluster: [%v]", kvdbNodes)

		cleanup := func() {
			log.Info("Executing cleanup tasks")
			if len(selectedNodesNotInKvdbForLabelTrue) > 0 {
				for _, node := range selectedNodesNotInKvdbForLabelTrue {
					err := Inst().S.RemoveLabelOnNode(node, "px/metadata-node")
					log.FailOnError(err, "Failed to remove label  'px/metadata-node=true' from node [%s]", node)
					log.Infof("Successfully removed label from node [%s]", node)
				}
			}
			if len(selectedNodesNotInKvdbForLabelFalse) > 0 {
				for _, node := range selectedNodesNotInKvdbForLabelFalse {
					err := Inst().S.RemoveLabelOnNode(node, "px/metadata-node")
					log.FailOnError(err, "Failed to remove label  'px/metadata-node=false' from node [%s]", node)
					log.Infof("Successfully removed label from node [%s]", node)
				}
			}
			if len(stopped_px_kvdb_nodes_ids) > 0 {
				for _, px_stop_node := range stopped_px_kvdb_nodes_ids {
					node, err := node.GetNodeDetailsByNodeID(px_stop_node)
					log.FailOnError(err, "Unable to get the node details for the node  [%v]", node)
					err = Inst().V.StartDriver(node)
					log.FailOnError(err, "error starting driver on node %s", node.Name)
					err = Inst().V.WaitDriverUpOnNode(node, 10*time.Minute)
					log.FailOnError(err, "error while waiting for driver up on node %s", node.Name)
					log.Infof("Successfully start the portworx :[%v]", px_stop_node)
				}
			}
			if selected_node_for_px_stop.ID != "" {
				node, err := node.GetNodeDetailsByNodeID(selected_node_for_px_stop.ID)
				log.FailOnError(err, "Unable to get the node details for the node  [%v]", node)
				err = Inst().V.StartDriver(node)
				log.FailOnError(err, "error starting driver on node %s", node.Name)
				err = Inst().V.WaitDriverUpOnNode(node, 10*time.Minute)
				log.FailOnError(err, "error while waiting for driver up on node %s", node.Name)
				log.Infof("Successfully start the portworx :[%v]", selected_node_for_px_stop)

			}
			DestroyApps(contexts, nil)
		}
		defer cleanup()

		stepLog = "Getting non KVDB nodes in the cluster"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			kvdbNodeMap := make(map[string]bool)
			for _, kvdbNode := range kvdbNodes {
				kvdbNodeMap[kvdbNode.ID] = true
			}
			for _, storageNode := range storageNodes {
				if _, exists := kvdbNodeMap[storageNode.Id]; !exists {
					nonKvdbNodes = append(nonKvdbNodes, storageNode)
				}
			}
			log.Infof("Storage nodes not part of KVDB cluster: [%v]", nonKvdbNodes)

		})

		selectedNodesNotInKvdbForLabelTrue = nonKvdbNodes[0:4]
		selectedNodesNotInKvdbForLabelFalse = nonKvdbNodes[4:len(nonKvdbNodes)]

		stepLog = "Label the node px/metadata-node=true and kill the px "
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.Infof("Selected nodes for labeling: %v", selectedNodesNotInKvdbForLabelTrue)
			for _, Non_kvdb_Node := range selectedNodesNotInKvdbForLabelTrue {
				err := Inst().S.AddLabelOnNode(Non_kvdb_Node, "px/metadata-node", "true")
				log.FailOnError(err, "Failed to add label 'px/metadata-node=true' to node [%s]", Non_kvdb_Node)
				log.InfoD("Stop PX  service  on the  node: %v", Non_kvdb_Node)
				StopVolDriverAndWait([]node.Node{Non_kvdb_Node})
				log.InfoD("PX service successfully stopped on node: %v", Non_kvdb_Node)
				log.Infof("Restarting PX on node: %v", Non_kvdb_Node)
				err = Inst().V.StartDriver(Non_kvdb_Node)
				log.FailOnError(err, "error starting driver on node %s", Non_kvdb_Node)
				err = Inst().V.WaitDriverUpOnNode(Non_kvdb_Node, 10*time.Minute)
				log.FailOnError(err, "error while waiting for driver up on node %s", Non_kvdb_Node)
				log.Infof("Successfully restart the portworx :[%v]", Non_kvdb_Node.Name)
			}
		})

		stepLog = "Label remaining nodes with px/metadata-node=false and kill the px on those nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.Infof("Selected nodes for labeling: %v", selectedNodesNotInKvdbForLabelFalse)
			for _, Non_kvdb_label_false_node := range selectedNodesNotInKvdbForLabelFalse {
				err := Inst().S.AddLabelOnNode(Non_kvdb_label_false_node, "px/metadata-node", "false")
				log.FailOnError(err, "Failed to add label 'px/metadata-node=true' to node [%s]", Non_kvdb_label_false_node)
				log.Info("Stop the volume driver and wait for it to stop completely")
				StopVolDriverAndWait([]node.Node{Non_kvdb_label_false_node})
				log.InfoD("PX service successfully stopped on node: %v", Non_kvdb_label_false_node)
				log.Infof("Restarting PX on node: %v", Non_kvdb_label_false_node)
				err = Inst().V.StartDriver(Non_kvdb_label_false_node)
				log.FailOnError(err, "error starting driver on node %s", Non_kvdb_label_false_node.Name)
				err = Inst().V.WaitDriverUpOnNode(Non_kvdb_label_false_node, 10*time.Minute)
				log.FailOnError(err, "error while waiting for driver up on node %s", Non_kvdb_label_false_node.Name)
				log.Infof("Successfully restart the portworx :[%v]", Non_kvdb_label_false_node.Name)
			}

		})

		stepLog := "Schedule application"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			contexts = make([]*scheduler.Context, 0)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("kvdbfailover-%d", i))...)
			}
			ValidateApplications(contexts)
		})

		stepLog = "Stopping portworx on KVDB member node and wait labeled node to join"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			kvdbNodes, err := GetAllKvdbNodes()
			log.FailOnError(err, "Unable to retrieve KVDB nodes")
			log.InfoD("KVDB nodes for PX service stop: %v", kvdbNodes)
			for _, kvdbNode := range kvdbNodes {
				nodeDetails, err := node.GetNodeDetailsByNodeID(kvdbNode.ID)
				log.FailOnError(err, "Unable to retrieve node details for NodeID [%v]", kvdbNode.ID)
				StopVolDriverAndWait([]node.Node{nodeDetails})
				log.InfoD("PX service successfully stopped on node: %v", nodeDetails)
				checkKVDBQuorum := func() (interface{}, bool, error) {
					log.Infof("Checking if the labeled node has joined and verifying that the quorum is intact before stopping PX on the next node.")
					getKVDBNodes, err := GetAllKvdbNodes()
					if err != nil {
						return nil, true, fmt.Errorf("unable to get KVDB nodes: %w", err)
					}
					log.Infof("KVDB node details: %v", getKVDBNodes)
					healthyCount := 0
					for _, each := range getKVDBNodes {
						if each.IsHealthy == true {
							healthyCount++
						}
					}
					if healthyCount == 3 {
						log.Infof("KVDB quorum intact. Healthy count: %d, Expected: 3", healthyCount)
						return nil, false, nil
					}
					log.Errorf("KVDB quorum lost. Healthy count: %d, Expected: 3. Retrying...", healthyCount)
					return nil, true, fmt.Errorf("quorum lost. Healthy count: %d, Expected: 3", healthyCount)
				}
				_, err = task.DoRetryWithTimeout(checkKVDBQuorum, 5*time.Minute, 30*time.Second)
				log.FailOnError(err, "Error occurred while checking KVDB quorum")
				stopped_px_kvdb_nodes_ids = append(stopped_px_kvdb_nodes_ids, kvdbNode.ID)
			}
		})

		stepLog = "Label KVDB nodes with px/metadata-node=false"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.Infof("Selected nodes for labeling: %v", stopped_px_kvdb_nodes_ids)
			for _, kvdb_node_label_false := range stopped_px_kvdb_nodes_ids {
				nodeDetails, err := node.GetNodeDetailsByNodeID(kvdb_node_label_false)
				log.FailOnError(err, "Unable to retrieve node details for NodeID [%v]", kvdb_node_label_false)
				err = Inst().S.AddLabelOnNode(nodeDetails, "px/metadata-node", "false")
				log.FailOnError(err, "Failed to add label 'px/metadata-node=true' to node [%s]", kvdb_node_label_false)
				selectedNodesNotInKvdbForLabelFalse = append(selectedNodesNotInKvdbForLabelFalse, nodeDetails)

			}
		})

		stepLog = "Restart Portworx on stopped kvdb nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, px_stop_node := range stopped_px_kvdb_nodes_ids {
				node, err := node.GetNodeDetailsByNodeID(px_stop_node)
				log.FailOnError(err, "Unable to get the node details for the node  [%v]", node)
				err = Inst().V.StartDriver(node)
				log.FailOnError(err, "error starting driver on node %s", node.Name)
				err = Inst().V.WaitDriverUpOnNode(node, 10*time.Minute)
				log.FailOnError(err, "error while waiting for driver up on node %s", node.Name)
				log.Infof("Successfully start the portworx :[%v]", px_stop_node)
			}

		})

		stepLog = "Check if the selected nodes are part of KVDB members"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			getKVDBNodes, err := GetAllKvdbNodes()
			log.FailOnError(err, "Unable to get KVDB nodes")
			log.Infof("After KVDB members labeled: [%v]", getKVDBNodes)

			kvdbNodeMap := make(map[string]bool)
			NonkvdbNodesAfterLabel := []string{}
			kvdbNodesAfterLabeled := []string{}
			for _, kvdbNode := range getKVDBNodes {
				kvdbNodeMap[kvdbNode.ID] = true
				log.Infof("KVDB member found: NodeID [%s]", kvdbNode.ID)
			}
			for _, node := range selectedNodesNotInKvdbForLabelTrue {
				if _, exists := kvdbNodeMap[node.Id]; !exists {
					log.Infof("Node [%s] is NOT a KVDB member.", node.Id)
					NonkvdbNodesAfterLabel = append(NonkvdbNodesAfterLabel, node.Id)
					nonKvdbNodeAfterLabel = node
					log.Infof("Storing the node that is not part of KVDB: [%s] [%s]", nonKvdbNodeAfterLabel.Id, nonKvdbNodeAfterLabel)
				} else {
					log.Infof("Node [%s] is a KVDB member.", node.Id)
					kvdbNodesAfterLabeled = append(kvdbNodesAfterLabeled, node.Id)
				}
			}
			if len(NonkvdbNodesAfterLabel) > 1 {
				err := fmt.Errorf("More than one labeled node is not part of the KVDB cluster: %v", NonkvdbNodesAfterLabel)
				log.FailOnError(err, "Test failed: More than one node is not part of KVDB.")
			}
			dash.VerifyFatal(len(kvdbNodesAfterLabeled) == 3, true, fmt.Sprintf("Labeled nodes are part of KVDB members? %v", kvdbNodesAfterLabeled))
		})

		stepLog = "Initiating second KVDB failover by stopping PX on one of the KVDB member nodes and checking the quorum"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			getKVDBNodes, err := GetAllKvdbNodes()
			log.FailOnError(err, "Unable to get KVDB nodes")
			selected_node_for_px_stop = getKVDBNodes[0]
			log.Infof("Selected node for PX stop: %v", selected_node_for_px_stop.ID)
			node_details, err := node.GetNodeDetailsByNodeID(selected_node_for_px_stop.ID)
			log.FailOnError(err, "Unable to retrieve node details for NodeID [%v]", selected_node_for_px_stop.ID)
			StopVolDriverAndWait([]node.Node{node_details})
			checkKVDBQuorum := func() (interface{}, bool, error) {
				log.Infof("Checking if the labeled node has joined and verifying that the quorum is intact before stopping PX on the next node.")
				getKVDBNodes, err := GetAllKvdbNodes()
				if err != nil {
					return nil, true, fmt.Errorf("unable to get KVDB nodes: %w", err)
				}
				log.Infof("KVDB node details: %v", getKVDBNodes)
				healthyCount = 0
				for _, each := range getKVDBNodes {
					if each.IsHealthy == true {
						healthyCount++
					}
				}
				if healthyCount == 3 {
					log.Infof("KVDB quorum intact. Healthy count: %d, Expected: 3", healthyCount)
					return nil, false, nil
				}
				log.Errorf("KVDB quorum lost. Healthy count: %d, Expected: 3. Retrying...", healthyCount)
				return nil, true, fmt.Errorf("quorum lost. Healthy count: %d, Expected: 3", healthyCount)
			}
			_, err = task.DoRetryWithTimeout(checkKVDBQuorum, 5*time.Minute, 30*time.Second)
			log.FailOnError(err, "Error occurred while checking KVDB quorum")
			dash.VerifyFatal(healthyCount == 3, true, fmt.Sprintf("verify kvdb quorum is not lost.Healthy count: %d, Expected: 3", healthyCount))

		})

		stepLog = "Verified labeled node only joined."
		Step(stepLog, func() {
			getKVDBNodes, err := GetAllKvdbNodes()
			log.FailOnError(err, "Unable to get KVDB nodes")
			kvdbNodeMap := make(map[string]bool)
			for _, kvdbNode := range getKVDBNodes {
				kvdbNodeMap[kvdbNode.ID] = true
				log.Infof("KVDB member found: NodeID [%s]", kvdbNode.ID)
			}
			if _, exists := kvdbNodeMap[nonKvdbNodeAfterLabel.Id]; exists {
				log.Infof("Node [%s] is a KVDB member.", nonKvdbNodeAfterLabel.Id)
			} else {
				log.Infof("Node [%s] is NOT a KVDB member.", nonKvdbNodeAfterLabel.Id)
				log.FailOnError(fmt.Errorf("Node [%s] is not a KVDB member", nonKvdbNodeAfterLabel.Id), "Selected node is not part of the KVDB cluster")
			}
			log.Infof("Labeled node verification completed.")

		})
		stepLog = "Starting Portworx on stopped node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			node, err := node.GetNodeDetailsByNodeID(selected_node_for_px_stop.ID)
			log.FailOnError(err, "Unable to get the node details for the node  [%v]", node)
			err = Inst().V.StartDriver(node)
			log.FailOnError(err, "error starting driver on node %s", node.Name)
			err = Inst().V.WaitDriverUpOnNode(node, 10*time.Minute)
			log.FailOnError(err, "error while waiting for driver up on node %s", node.Name)
			log.Infof("Successfully start the portworx :[%v]", selected_node_for_px_stop)
		})

		ValidateApplications(contexts)

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{RestartPxOnStorageLessNode}", Label("staging", "kvdb_ops", "p1", "negative"), func() {
	/*
		https://purestorage.atlassian.net/browse/HAZEL-1020
		Step 1: Have 7 storage nodes & 3 storageless nodes
		Step 2: Create apps
		Step 3: Kill the PX on one of the kvdb node
		Step 4: Px restart on all storage less nodes
		Step 5: Validate apps
	*/

	var (
		testrailID           = 0
		runID                int
		contexts             []*scheduler.Context
		Kill_px_on_kvdb_node KvdbNode
	)
	JustBeforeEach(func() {
		StartTorpedoTest("RestartPxOnStorageLessNode", "Managing applications and validating them after restarting Px on storageless nodes.", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	stepLog := "Identify KVDB nodes, stop PX on one KVDB node, restart PX on storageless nodes, and validate applications."
	It(stepLog, func() {
		log.InfoD(stepLog)
		storageNodes := node.GetStorageNodes()
		log.Infof("Storage node in the cluster: [%v]", storageNodes)

		storageLessNode := node.GetStorageLessNodes()
		if len(storageLessNode) >= 3 {
			Skip("At least 3 storageless nodes are required to run the tests.")
		}
		log.InfoD("Get all KVDB nodes")
		kvdbNodes, err := GetAllKvdbNodes()
		log.FailOnError(err, "Unable to retrieve KVDB nodes")

		cleanup := func() {
			log.Info("Executing cleanup tasks")
			node, err := node.GetNodeDetailsByNodeID(Kill_px_on_kvdb_node.ID)
			log.FailOnError(err, "Unable to get the node details for the node  [%v]", node)
			err = Inst().V.StartDriver(node)
			log.FailOnError(err, "error starting driver on node %s", node.Name)
			err = Inst().V.WaitDriverUpOnNode(node, 10*time.Minute)
			log.FailOnError(err, "error while waiting for driver up on node %s", node.Name)
			log.Infof("Successfully start the portworx :[%v]", Kill_px_on_kvdb_node)
			DestroyApps(contexts, nil)
		}
		defer cleanup()

		stepLog := "Schedule application"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			contexts = make([]*scheduler.Context, 0)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("kvdbfailover-%d", i))...)
			}
			ValidateApplications(contexts)
		})

		stepLog = "Kill the Portworx one of the  KVDB member node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			Kill_px_on_kvdb_node = kvdbNodes[0]
			nodeDetails, err := node.GetNodeDetailsByNodeID(Kill_px_on_kvdb_node.ID)
			log.FailOnError(err, "Unable to retrieve node details for NodeID [%v]", Kill_px_on_kvdb_node.ID)
			StopVolDriverAndWait([]node.Node{nodeDetails})
			log.InfoD("PX service successfully stopped on node: %v", nodeDetails)
			checkKVDBQuorum := func() (interface{}, bool, error) {
				getKVDBNodes, err := GetAllKvdbNodes()
				if err != nil {
					return nil, true, fmt.Errorf("unable to get KVDB nodes: %w", err)
				}
				log.Infof("KVDB node details: %v", getKVDBNodes)
				healthyCount := 0
				for _, each := range getKVDBNodes {
					if each.IsHealthy == true {
						healthyCount++
					}
				}
				if healthyCount == 2 {
					log.Infof("KVDB quorum intact. Healthy count: %d, Expected: 2", healthyCount)
					return nil, false, nil
				}
				log.Errorf("KVDB quorum not lost. Healthy count: %d, Expected: 2. Retrying...", healthyCount)
				return nil, true, fmt.Errorf("quorum need to lost. Healthy count: %d, Expected: 2", healthyCount)
			}
			_, err = task.DoRetryWithTimeout(checkKVDBQuorum, 5*time.Minute, 30*time.Second)
			log.FailOnError(err, "Error occurred while checking KVDB quorum")
		})

		stepLog = "Restart PX on all storageless node"
		Step(stepLog, func() {
			log.Info(stepLog)
			for _, storage_less_node := range storageLessNode {
				err = Inst().V.RestartDriver(storage_less_node, nil)
				log.FailOnError(err, fmt.Sprintf("Error restarting px on node [%s]", storage_less_node.Name))
				err = Inst().V.WaitForPxPodsToBeUp(storage_less_node)
				log.FailOnError(err, fmt.Sprintf("Error occured while Validating PX restart is done on node:%v", storage_less_node.Name))
				log.Infof("PX restart validated successfully on storageless node: %v", storage_less_node.Name)

			}
		})

		ValidateApplications(contexts)
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{ContinuousPXRestartAndAppValidation}", Label("p0", "negative", "px_vol_ops", "staging"), func() {
	/*
	   ticket id:https://purestorage.atlassian.net/browse/HAZEL-996
	   Create apps
	   Randomly choose a storage node
	   Kill PX on this storage node
	   Wait for PX to come up
	   Validate apps
	   Repeat steps 2 to 5 at least 25 times
	*/
	var (
		testrailID = 0
		runID      int
		contexts   []*scheduler.Context
	)
	JustBeforeEach(func() {
		StartTorpedoTest("ContinuousPXRestartAndAppValidation", "Continuously kills PX on a random storage node, waits for it to restart, and validates apps.", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	itLog := "Continuously kills PX on a random storage node, waits for it to restart, and validates apps"
	It(itLog, func() {
		log.InfoD(itLog)

		log.Info("Get storage nodes")
		storageNodes := node.GetStorageNodes()
		log.Infof("Storage node in the cluster: [%v]", storageNodes)

		cleanup := func() {
			log.InfoD("Cleanup the task")
			for _, n := range storageNodes {
				pxStatus, err := Inst().V.GetPxctlStatus(n)
				if err == nil && pxStatus != api.Status_STATUS_OK.String() {
					err = Inst().V.StartDriver(n)
					log.FailOnError(err, "Failed to start Portworx driver on node %s", n.Name)
					err = Inst().V.WaitDriverUpOnNode(n, 10*time.Minute)
					log.FailOnError(err, "Failed to waiting for Portworx driver to start on node %s", n.Name)
				}
			}
			DestroyApps(contexts, nil)
		}
		defer cleanup()

		stepLog = "Kill PX on a random storage node and validate applications 25 times"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for i := 0; i < 25; i++ {
				pxNode := GetRandomNode(storageNodes)
				log.Infof("Stopping the volume driver on node %v", pxNode)
				StopVolDriverAndWait([]node.Node{pxNode})
				log.Infof("PX service successfully stopped on node: %v", pxNode)
				log.Infof("Restarting PX on node: %v", pxNode)
				err := Inst().V.StartDriver(pxNode)
				log.FailOnError(err, "Error starting PX driver on node %s", pxNode)
				err = Inst().V.WaitDriverUpOnNode(pxNode, 10*time.Minute)
				log.FailOnError(err, "Error waiting for PX driver up on node %s", pxNode)
				log.Infof("PX restarted successfully on node: %v", pxNode)
				ValidateApplications(contexts)
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

// Keep killing the node which is becoming new etcd node.
var _ = Describe("{ValidateKillingNewNodeAfterKVDBFailOver}", Label("p1", "kvdb_ops", "negative", "staging"), func() {
	/*
		JiraID : https://purestorage.atlassian.net/browse/HAZEL-1006
		Create few (10) initial apps
		Trigger KVDB Failover by stopping px on kvdb node
		Once failover is done and new node becomes kvdb member, kill this new node as well
		Create one app & delete any one initial app
		Repeat steps 2 to 4 atleast 10 times
		After Step 5, all initial apps should be deleted and 10 new apps from step 4 should be running
		Validate apps
	*/

	var (
		testrailID            = 36147344
		runID                 int
		preContexts           = make([]*scheduler.Context, 0)
		postContexts          = make([]*scheduler.Context, 0)
		initialAppCount       = 10
		selectedNodeForPxStop KvdbNode
		previousKVDBNodesMap  = make(map[string]bool)
	)
	JustBeforeEach(func() {
		StartTorpedoTest("ValidateKillingNewNodeAfterKVDBFailOver", "Keep killing the node which is becoming new etcd node", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	stepLog := "Restart all the px nodes after killing leader node of etcd in etcd cluster"
	It(stepLog, func() {
		log.InfoD(stepLog)

		storageNodes := node.GetStorageNodes()
		log.Infof("Storage node in the cluster: [%v]", storageNodes)
		if len(storageNodes) < 7 {
			Skip("At least 7 nodes are required to run the tests")
		}

		for i := 0; i < initialAppCount; i++ {
			preContexts = append(preContexts, ScheduleApplications(fmt.Sprintf("kvdb-%d", i))...)
		}

		ValidateApplications(preContexts)
		cleanup := func() {
			log.Info("Executing cleanup tasks")

			if selectedNodeForPxStop.ID != "" {
				node, err := node.GetNodeDetailsByNodeID(selectedNodeForPxStop.ID)
				log.FailOnError(err, "Unable to get the node details for the node  [%v]", node)
				err = Inst().V.StartDriver(node)
				log.FailOnError(err, "error starting driver on node %s", node.Name)
				err = Inst().V.WaitDriverUpOnNode(node, 10*time.Minute)
				log.FailOnError(err, "error while waiting for driver up on node %s", node.Name)
				log.Infof("Successfully start the portworx :[%v]", selectedNodeForPxStop)
			}
			DestroyApps(preContexts, nil)
			DestroyApps(postContexts, nil)
		}
		defer cleanup()

		stopPXOnKVDBNode := func(nodeDetails node.Node) {
			StopVolDriverAndWait([]node.Node{nodeDetails})
			log.InfoD("PX service successfully stopped on node: %v", selectedNodeForPxStop)
			time.Sleep(20 * time.Second)
			healthyCount := 0
			checkKVDBQuorum := func() (interface{}, bool, error) {
				healthyCount = 0
				getKVDBNodes, err := GetAllKvdbNodes()
				if err != nil {
					return nil, true, fmt.Errorf("unable to get KVDB nodes: %w", err)
				}
				log.Infof("KVDB node details: %v", getKVDBNodes)
				for _, each := range getKVDBNodes {
					if each.IsHealthy == true {
						healthyCount++
					}
				}
				if healthyCount == 2 {
					log.Infof("KVDB quorum intact. Healthy count: %d, Expected: 2", healthyCount)
					return nil, false, nil
				}
				log.Errorf("KVDB quorum lost. Healthy count: %d, Expected: 2. Retrying...", healthyCount)
				return nil, true, fmt.Errorf("quorum lost. Healthy count: %d, Expected: 2", healthyCount)
			}
			_, err = task.DoRetryWithTimeout(checkKVDBQuorum, 5*time.Minute, 30*time.Second)
			log.FailOnError(err, "Error occurred while checking KVDB quorum")
			dash.VerifyFatal(healthyCount == 2, true, fmt.Sprintf("verify kvdb quorum is not lost.Healthy count: %d, Expected: 2", healthyCount))
		}

		// Repeat steps 2 to 4 atleast 10 times
		for i := 0; i < initialAppCount; i++ {
			stepLog = "Initiating KVDB failover by stopping PX on one of the KVDB member nodes"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				getKVDBNodes, err := GetAllKvdbNodes()
				log.FailOnError(err, "Unable to get KVDB nodes")
				previousKVDBNodesMap = make(map[string]bool)
				for _, kvdbNode := range getKVDBNodes {
					previousKVDBNodesMap[kvdbNode.ID] = true
				}
				log.Infof("Previous KVDB node details: %v", getKVDBNodes)
				selectedNodeForPxStop = getKVDBNodes[0]
				log.Infof("Selected node for PX stop: %v", selectedNodeForPxStop)
				nodeDetails, err := node.GetNodeDetailsByNodeID(selectedNodeForPxStop.ID)
				log.FailOnError(err, "Unable to retrieve node details for NodeID [%v]", selectedNodeForPxStop.ID)
				stopPXOnKVDBNode(nodeDetails)
			})

			stepLog = "Deleting the new kvdb member node"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				newKVDBNode := node.Node{}
				t := func() (interface{}, bool, error) {
					getKVDBNodes, err := GetAllKvdbNodes()
					if err != nil {
						return nil, true, fmt.Errorf("failed to get KVDB nodes")
					}
					healthyCount := 0
					for _, each := range getKVDBNodes {
						if each.IsHealthy == true {
							healthyCount++
						}
					}
					if healthyCount == 3 {
						log.Infof("KVDB quorum intact. Healthy count: %d, Expected: 3", healthyCount)
						log.Infof("Number of KVDB nodes: %v", len(getKVDBNodes))
						return getKVDBNodes, false, nil
					}
					return nil, true, fmt.Errorf("KVDB quorum not intact. Healthy count: %d, Expected: 3", healthyCount)
				}
				getKVDBNodes, err := task.DoRetryWithTimeout(t, 10*time.Minute, 30*time.Second)
				log.FailOnError(err, "Unable to get KVDB nodes")
				log.Infof("KVDB node details: %v", getKVDBNodes)
				for _, kvdbNode := range getKVDBNodes.([]KvdbNode) {
					log.Infof("New KVDB node details: %v", kvdbNode.ID)
					if _, ok := previousKVDBNodesMap[kvdbNode.ID]; !ok {
						log.InfoD("The New KVDB Node to be killed %v", kvdbNode)
						newKVDBNode, err = node.GetNodeDetailsByNodeID(kvdbNode.ID)
						log.FailOnError(err, "Unable to get node details with ID [%v]", kvdbNode.ID)
						break
					}
				}
				log.Infof("Stop volume driver [%s] on nodes: [%v]", Inst().V.String(), newKVDBNode)
				StopVolDriverAndWait([]node.Node{newKVDBNode})
				log.Infof("Stopped volume driver [%s] on nodes [%v]", Inst().V.String(), newKVDBNode)
				log.Infof("Starting volume driver [%s] on nodes [%v]", Inst().V.String(), newKVDBNode)
				StartVolDriverAndWait([]node.Node{newKVDBNode})
				log.Infof("Started volume driver [%s] on nodes [%v]", Inst().V.String(), newKVDBNode)
			})

			stepLog = "Create one app & deleting one initial app"
			Step(stepLog, func() {
				log.InfoD(stepLog)

				// deleting an app and namespace from the  preContexts
				ns := preContexts[i].App.NameSpace
				TearDownContext(preContexts[i], nil)
				err = k8sCore.DeleteNamespace(ns)
				log.FailOnError(err, "Unable to delete namespace [%v]", ns)
				// schedule a new app in postContexts
				postContexts = append(postContexts, ScheduleApplications(fmt.Sprintf("kvdb-new-%d", i))...)
			})
			stepLog = "Starting the PX driver on the initially stopped Node"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				log.Infof("Selected node for PX start: %v", selectedNodeForPxStop)
				nodeDetails, err := node.GetNodeDetailsByNodeID(selectedNodeForPxStop.ID)
				log.FailOnError(err, "Unable to retrieve node details for NodeID [%v]", selectedNodeForPxStop.ID)
				StartVolDriverAndWait([]node.Node{nodeDetails})
				log.Infof("Started volume driver [%s] on nodes [%v]", Inst().V.String(), nodeDetails)
				selectedNodeForPxStop = KvdbNode{}
			})
		}

		stepLog = "Verify all initial apps are deleted"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, ctx := range preContexts {
				log.Infof("Length of precontexts is %v, ctx value: %v", len(preContexts), ctx.UID)
				err := Inst().S.SelectiveWaitForTermination(ctx, Inst().DestroyAppTimeout, []node.Node{})
				dash.VerifyFatal(err == nil, true, fmt.Sprintf("Initially created app [%v] is not deleted!!", len(preContexts)))
			}
		})

		stepLog = "validate applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			ValidateApplications(postContexts)
		})

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(postContexts, testrailID, runID)
	})
})

// keep restarting etcd node till 3 mins.
var _ = Describe("{RestartKVDBNodeUntilTimeout}", Label("p1", "kvdb_ops", "positive", "node_reboot", "staging"), func() {
	/*
		JiraID : https://purestorage.atlassian.net/browse/HAZEL-1004
		1. Pick any kvdb node
		2. Restart the node
		3. Repeat above 2 steps atleast 10 times
		4. Validate apps
	*/

	var (
		testrailID    = 36147339
		runID         int
		contexts      = make([]*scheduler.Context, 0)
		loopCount     = 10
		nodeToRestart = node.Node{}
	)
	JustBeforeEach(func() {
		StartTorpedoTest("RestartKVDBNodeUntilTimeout", "keep restarting etcd node till 3 mins", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	stepLog := "Restarting Random KVDB nodes in a loop(10 times)"
	It(stepLog, func() {
		log.InfoD(stepLog)

		stepLog = "Checking the Storage node in the cluster"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			storageNodes := node.GetStorageNodes()
			log.Infof("Storage node in the cluster: [%v]", storageNodes)
			if len(storageNodes) < 7 {
				Skip("At least 7 nodes are required to run the tests")
			}
		})

		stepLog = "Scheduling Applications and validating"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("kvdb-%d", i))...)
			}
			ValidateApplications(contexts)
		})

		cleanup := func() {
			log.Info("Executing cleanup tasks")
			DestroyApps(contexts, nil)
		}
		defer cleanup()

		checkKVDBQuorum := func() (interface{}, bool, error) {
			healthyCount := 0
			getKVDBNodes, err := GetAllKvdbNodes()
			if err != nil {
				return nil, true, fmt.Errorf("unable to get KVDB nodes: %w", err)
			}
			log.Infof("KVDB node details: %v", getKVDBNodes)
			for _, each := range getKVDBNodes {
				if each.IsHealthy == true {
					healthyCount++
				}
			}
			if healthyCount == 3 {
				log.Infof("KVDB quorum intact. Healthy count: %d, Expected: 3", healthyCount)
				return getKVDBNodes, false, nil
			}
			log.Errorf("KVDB quorum lost. Healthy count: %d, Expected: 3. Retrying...", healthyCount)
			return nil, true, fmt.Errorf("quorum lost. Healthy count: %d, Expected: 3", healthyCount)
		}

		// Repeat steps 1 & 2 atleast 10 times
		for i := 0; i < loopCount; i++ {
			log.InfoD("Iteration : <<%d>> Restarting KVDB Node", i)

			stepLog = "Picking a KVDB node for restart"
			Step(stepLog, func() {
				log.InfoD(stepLog)

				getKVDBNodes, err := task.DoRetryWithTimeout(checkKVDBQuorum, 5*time.Minute, 20*time.Second)
				log.FailOnError(err, "Unable to get KVDB nodes")
				selectedKVDBNode := getKVDBNodes.([]KvdbNode)[rand.Intn(len(getKVDBNodes.([]KvdbNode)))]
				nodeToRestart, err = node.GetNodeDetailsByNodeID(selectedKVDBNode.ID)
				log.FailOnError(err, "Unable to retrieve node details for NodeID [%v]", selectedKVDBNode.ID)
				log.Infof("KVDB node to Restart : %v", nodeToRestart)
			})

			stepLog = "keep restarting etcd node till 3 mins"
			Step(stepLog, func() {
				log.InfoD(stepLog)

				log.InfoD("Rebooting Node [%v]", nodeToRestart.Name)
				err = Inst().N.RebootNode(nodeToRestart, node.RebootNodeOpts{
					Force: true,
					ConnectionOpts: node.ConnectionOpts{
						Timeout:         1 * time.Minute,
						TimeBeforeRetry: 5 * time.Second,
					},
				})
				log.FailOnError(err, "failed to reboot Node [%v]", nodeToRestart.Name)
				log.InfoD("Restarted the KVDB Node [%v]", nodeToRestart.Name)
			})

			stepLog = "Wait for connection to come back online after reboot"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err = Inst().N.TestConnection(nodeToRestart, node.ConnectionOpts{
					Timeout:         15 * time.Minute,
					TimeBeforeRetry: 10 * time.Second,
				})

				err = Inst().S.IsNodeReady(nodeToRestart)
				log.FailOnError(err, "Node [%v] is not in ready state", nodeToRestart.Name)

				err = Inst().V.WaitDriverUpOnNode(nodeToRestart, Inst().DriverStartTimeout)
				log.FailOnError(err, "failed waiting for driver up on Node[%v]", nodeToRestart.Name)
			})
		}

		stepLog = "validate applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			ValidateApplications(contexts)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{RestartPxOnStorageLessNodeForMultipleTimes}", Label("staging", "kvdb_ops", "p1", "negative"), func() {
	/*
		https://purestorage.atlassian.net/browse/HAZEL-1020
		Step 1: few atleast 3 storageless nodes in the cluster
		Step 2:  top Px on any one storage node - which should be a kvdb node also
		Step 3: Restart Px on atleast 3 storageless nodes and keep killing Px atleast 5 times in quick succession
		Step 4: Validate all apps / nodes status afterward
	*/

	var (
		testrailID           = 0
		runID                int
		contexts             []*scheduler.Context
		Kill_px_on_kvdb_node KvdbNode
	)
	JustBeforeEach(func() {
		StartTorpedoTest("RestartPxOnStorageLessNodeForMultipleTimes", "Managing applications and validating them after Multiple restart Px on storageless nodes.", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	stepLog := "Identify KVDB nodes, stop PX on one KVDB node, multiple restart PX on storageless nodes, and validate applications."
	It(stepLog, func() {
		log.InfoD(stepLog)
		storageNodes := node.GetStorageNodes()
		log.Infof("Storage node in the cluster: [%v]", storageNodes)

		storageLessNode := node.GetStorageLessNodes()
		log.Infof("StorageLess node in the cluster: [%d]", len(storageLessNode))

		if len(storageLessNode) < 3 {
			Skip("At least 3 storageless nodes are required to run the tests.")
		}
		log.InfoD("Get all KVDB nodes")
		kvdbNodes, err := GetAllKvdbNodes()
		log.FailOnError(err, "Unable to retrieve KVDB nodes")

		cleanup := func() {
			log.Info("Executing cleanup tasks")
			node, err := node.GetNodeDetailsByNodeID(Kill_px_on_kvdb_node.ID)
			log.FailOnError(err, "Unable to get the node details for the node  [%v]", node)
			err = Inst().V.StartDriver(node)
			log.FailOnError(err, "error starting driver on node %s", node.Name)
			err = Inst().V.WaitDriverUpOnNode(node, 10*time.Minute)
			log.FailOnError(err, "error while waiting for driver up on node %s", node.Name)
			log.Infof("Successfully start the portworx :[%v]", Kill_px_on_kvdb_node)
			DestroyApps(contexts, nil)
		}
		defer cleanup()

		stepLog := "Schedule application"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			contexts = make([]*scheduler.Context, 0)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("kvdbfailover-%d", i))...)
			}
			ValidateApplications(contexts)
		})

		stepLog = "Kill the Portworx one of the  KVDB member node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			Kill_px_on_kvdb_node = kvdbNodes[0]
			nodeDetails, err := node.GetNodeDetailsByNodeID(Kill_px_on_kvdb_node.ID)
			log.FailOnError(err, "Unable to retrieve node details for NodeID [%v]", Kill_px_on_kvdb_node.ID)
			StopVolDriverAndWait([]node.Node{nodeDetails})
			log.InfoD("PX service successfully stopped on node: %v", nodeDetails)
			checkKVDBQuorum := func() (interface{}, bool, error) {
				getKVDBNodes, err := GetAllKvdbNodes()
				if err != nil {
					return nil, true, fmt.Errorf("unable to get KVDB nodes: %w", err)
				}
				log.Infof("KVDB node details: %v", getKVDBNodes)
				healthyCount := 0
				for _, each := range getKVDBNodes {
					if each.IsHealthy == true {
						healthyCount++
					}
				}
				if healthyCount == 2 {
					log.Infof("KVDB quorum intact. Healthy count: %d, Expected: 2", healthyCount)
					return nil, false, nil
				}
				log.Errorf("KVDB quorum not lost. Healthy count: %d, Expected: 2. Retrying...", healthyCount)
				return nil, true, fmt.Errorf("quorum need to lost. Healthy count: %d, Expected: 2", healthyCount)
			}
			_, err = task.DoRetryWithTimeout(checkKVDBQuorum, 5*time.Minute, 30*time.Second)
			log.FailOnError(err, "Error occurred while checking KVDB quorum")
		})

		stepLog = "Multiple restarts of PX on all storage-less nodes."
		Step(stepLog, func() {
			log.Info(stepLog)
			for i := 0; i < 5; i++ {
				for _, storage_less_node := range storageLessNode {
					log.Infof("Stop volume driver [%s] on nodes: [%v]", Inst().V.String(), storage_less_node)
					StopVolDriverAndWait([]node.Node{storage_less_node})
					log.Infof("Stopped volume driver [%s] on nodes [%v]", Inst().V.String(), storage_less_node)
					log.Infof("Starting volume driver [%s] on nodes [%v]", Inst().V.String(), storage_less_node)
					StartVolDriverAndWait([]node.Node{storage_less_node})
					log.Infof("Started volume driver [%s] on nodes [%v]", Inst().V.String(), storage_less_node)

				}
			}

		})

		ValidateApplications(contexts)
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{AddNewNodeWhenClusterInRunFlatState}", Label("staging", "kvdb_ops", "p1", "negative"), func() {
	/*
				   Ticket ID: https://purestorage.atlassian.net/browse/HAZEL-1048
				   Step 1: Bring system in run flat state
		           Step 2: Add a node to the setup
				   Step 3: Validate node addition is successful
				   Step 4: Validate px status on all nodes (including new one) → pxctl status
				   Step 5: Exit run flat state
				   Step 6: Validate apps
	*/
	JustBeforeEach(func() {
		StartTorpedoTest("AddNewNodeWhenClusterInRunFlatState",
			"Simulate a run-flat state, add new node to cluster, validate PX status, Exit run flat state and validate apps", nil, 0)
	})
	var (
		contexts                      []*scheduler.Context
		selectedKvdbNodes             []KvdbNode
		kvdbNodes                     []KvdbNode
		numOfStorageNodes             int
		maxStorageNodesPerZone        uint32
		updatedMaxStorageNodesPerZone uint32 = 0
		zones                         []string
		SelectednodeDetails           []node.Node
	)

	itLog := "Simulate a run-flat state, add new node to cluster, validate PX status, Exit run flat state and validate apps"
	It(itLog, func() {
		log.InfoD(itLog)
		stepLog = "Schedule application"
		Step(stepLog, func() {
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("runflat-%d", i))...)
			}
		})

		ValidateApplications(contexts)

		cleanup := func() {
			log.Info("Executing cleanup tasks")
			if len(selectedKvdbNodes) > 0 {
				for _, n := range selectedKvdbNodes {
					nodeDetails, err := node.GetNodeDetailsByNodeID(n.ID)
					log.FailOnError(err, "Failed to retrieve node details for NodeID [%v]", n.ID)
					err = Inst().V.StartDriver(nodeDetails)
					log.FailOnError(err, "Failed to start Portworx driver on node %s", nodeDetails.Name)
					err = Inst().V.WaitDriverUpOnNode(nodeDetails, 10*time.Minute)
					log.FailOnError(err, "Failed to waiting for Portworx driver to start on node %s", nodeDetails.Name)
					log.InfoD("Successfully started Portworx on KVDB node: %v", nodeDetails.Name)
				}
			}
			DestroyApps(contexts, nil)
		}
		defer cleanup()
		// Validate total node count
		currentNodeCount, err := Inst().S.GetASGClusterSize()
		log.FailOnError(err, "Failed to Get ASG Cluster Size")
		log.Infof("Current nodes before adding node:[d]", currentNodeCount)

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
				SelectednodeDetails = append(SelectednodeDetails, nodeDetails)

			}
		})

		stepLog = " Verify cluster is in run-flat state"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			pxNode, err := node.GetNodeDetailsByNodeID(kvdbNodes[0].ID)
			output, err := runCmd("pxctl status", pxNode)
			log.FailOnError(err, "Failed to execute 'pxctl status' on node: %v", pxNode.Name)

			log.Infof("pxctl status output: %v\n", output)
			expect_out := "All operations (get/update/delete) are unavailable."
			dash.VerifyFatal(strings.Contains(output, expect_out), true, "Is cluster in run-flat state?")
		})

		Step(stepLog, func() {
			stepLog = "update maxStorageNodesPerZone in storage cluster spec"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				stc, err := Inst().V.GetDriver()

				log.FailOnError(err, "error getting volume driver")
				maxStorageNodesPerZone = *stc.Spec.CloudStorage.MaxStorageNodesPerZone
				numOfStorageNodes = len(node.GetStorageNodes())
				log.Infof("maxStorageNodesPerZone %d", int(maxStorageNodesPerZone))
				log.Infof("numOfStorageNodes %d", numOfStorageNodes)

				actualPerZoneCount := numOfStorageNodes
				if Inst().S.String() != openshift.SchedName && Inst().S.String() != anthos.SchedName {
					zones, err = Inst().S.GetZones()
					dash.VerifyFatal(err, nil, "Verify Get zones")

					actualPerZoneCount = numOfStorageNodes / len(zones)
				}

				if int(maxStorageNodesPerZone) <= actualPerZoneCount {
					updatedMaxStorageNodesPerZone = uint32(actualPerZoneCount + 1)
				}

				if updatedMaxStorageNodesPerZone != 0 {
					stc.Spec.CloudStorage.MaxStorageNodesPerZone = &updatedMaxStorageNodesPerZone
					log.InfoD("updating maxStorageNodesPerZone from %d to %d", maxStorageNodesPerZone, updatedMaxStorageNodesPerZone)
					pxOperator := operator.Instance()
					_, err = pxOperator.UpdateStorageCluster(stc)
					log.FailOnError(err, "error updating storage cluster")

				}
				PrintPxctlStatus()
				expReplicas := len(node.GetStorageDriverNodes()) + 1
				log.InfoD("scaling up the cluster to replicas %d", expReplicas)
				Scale(int64(expReplicas))
				stepLog = fmt.Sprintf("wait for %s minutes for auto recovery of storage nodes",
					Inst().AutoStorageNodeRecoveryTimeout.String())

				Step(stepLog, func() {
					log.InfoD(stepLog)
					time.Sleep(Inst().AutoStorageNodeRecoveryTimeout)
				})
				NodeCountAfterNewNode, err := Inst().S.GetASGClusterSize()
				log.FailOnError(err, "Failed to Get ASG Cluster Size")
				log.Infof("Nodes count After adding new node:[d]", currentNodeCount)
				dash.VerifyFatal(NodeCountAfterNewNode > currentNodeCount, true, "Node is added?")

			})
			stepLog = "Starting Portworx on selected KVDB nodes"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				for _, n := range selectedKvdbNodes {
					nodeDetails, err := node.GetNodeDetailsByNodeID(n.ID)
					log.FailOnError(err, "Failed to retrieve node details for NodeID [%v]", n.ID)
					StartVolDriverAndWait([]node.Node{nodeDetails})
					log.InfoD("Successfully started Portworx on KVDB node: %v", nodeDetails.Name)
				}
			})

			stepLog = "validate PX on all nodes after adding a storage node"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				time.Sleep(10 * time.Minute)
				err = Inst().S.RefreshNodeRegistry()
				log.FailOnError(err, "Verify node registry refresh")
				err = Inst().V.RefreshDriverEndpoints()
				log.FailOnError(err, "Verify driver end points refresh")
				nodes := node.GetStorageDriverNodes()
				for _, n := range nodes {
					log.InfoD("Check PX status on %v", n.Name)
					err := Inst().V.WaitForPxPodsToBeUp(n)
					dash.VerifyFatal(err, nil, fmt.Sprintf("verify px is up on  node %s", n.Name))
				}
				PrintPxctlStatus()
				expectedPerZone := maxStorageNodesPerZone
				if updatedMaxStorageNodesPerZone != 0 {
					expectedPerZone = updatedMaxStorageNodesPerZone
				}
				numOfZones := 1
				if len(zones) != 0 {
					numOfZones = len(zones)
				}
				expectedStorageNodesCount := int(expectedPerZone) * numOfZones

				if expectedStorageNodesCount >= len(node.GetStorageNodes()) {
					expectedStorageNodesCount = len(node.GetStorageNodes())
				}

				updatedStorageNodesCount := len(node.GetStorageNodes())
				dash.VerifyFatal(expectedStorageNodesCount, updatedStorageNodesCount, "verify new storage node is added")
			})

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

			storageNodes := node.GetStorageNodes()
			kvdbMembers, err := Inst().V.GetKvdbMembers(storageNodes[0])
			log.FailOnError(err, "Failed to retrieve KVDB members list")

			err = kvdbutils.ValidateKVDBMembers(kvdbMembers)
			log.FailOnError(err, "Failed to validate KVDB members")

			output, err := runCmd("pxctl status", storageNodes[0])
			log.FailOnError(err, "Failed to execute pxctl status on node: %v", storageNodes[0].Name)
			dash.VerifyFatal(!strings.Contains(output, "Warning"), true, "Output contains warnings. Is the cluster healthy?")
		})

		ValidateApplications(contexts)
	})

	JustAfterEach(func() {
		EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

// Bring up old etcd node at 3rd min while another node starting to bootstrap as new etcd
var _ = Describe("{RecoverOldKVDBDuringBootstrap}", Label("kvdb_ops", "p1", "negative", "staging"), func() {
	/*
		ticket id: https://purestorage.atlassian.net/browse/HAZEL-1001
		Pre-requisities: Atleast 7 nodes in the cluster
		1. Schedule Apps
		2. Bring down one KVDB node for long
		3. Wait for KVDB Failover to happen and a new node to become kvdb member
		4. As soon as you see this happening, bring up failed old kvdb node as well
		5. Validate kvdb cluster has 3 nodes - 2 old and 1 new member.
		6. validate Apps
	*/

	var (
		testrailID = 0
		runID      int
		contexts   []*scheduler.Context
	)
	JustBeforeEach(func() {
		StartTorpedoTest("RecoverOldKVDBDuringBootstrap", "Bring up old etcd node at 3rd min while another node starting to bootstrap as new etcd.", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})
	stepLog := "Bring up old etcd node at 3rd min while another node starting to bootstrap as new etcd."
	It(stepLog, func() {
		log.InfoD(stepLog)
		var (
			selectedNodeForPxStop KvdbNode
			previousKVDBNodesMap  = make(map[string]bool)
		)
		storageNodes := node.GetStorageNodes()
		log.Infof("Storage node in the cluster: [%v]", storageNodes)
		if len(storageNodes) < 7 {
			Skip("At least 7 nodes are required to run the tests")
		}
		log.InfoD("Get all KVDB nodes")
		kvdbNodes, err := GetAllKvdbNodes()
		log.FailOnError(err, "Unable to retrieve KVDB nodes")
		log.Infof("Initally kvdb node in the cluster: [%v]", kvdbNodes)

		cleanup := func() {
			log.Info("Executing cleanup tasks")
			if selectedNodeForPxStop.ID != "" {
				node, err := node.GetNodeDetailsByNodeID(selectedNodeForPxStop.ID)
				log.FailOnError(err, "Unable to get the node details for the node  [%v]", node)
				err = Inst().V.StartDriver(node)
				log.FailOnError(err, "error starting driver on node %s", node.Name)
				err = Inst().V.WaitDriverUpOnNode(node, 10*time.Minute)
				log.FailOnError(err, "error while waiting for driver up on node %s", node.Name)
				log.Infof("Successfully start the portworx :[%v]", selectedNodeForPxStop)

			}
			DestroyApps(contexts, nil)
		}
		defer cleanup()

		stepLog := "Schedule application"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			contexts = make([]*scheduler.Context, 0)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("kvdbfailover-%d", i))...)
			}
			ValidateApplications(contexts)
		})

		stopPXOnKVDBNode := func(nodeDetails node.Node) {
			StopVolDriverAndWait([]node.Node{nodeDetails})
			log.InfoD("PX service successfully stopped on node: %v", selectedNodeForPxStop)
			time.Sleep(20 * time.Second)
			healthyCount := 0
			checkKVDBQuorum := func() (interface{}, bool, error) {
				healthyCount = 0
				getKVDBNodes, err := GetAllKvdbNodes()
				if err != nil {
					return nil, true, fmt.Errorf("unable to get KVDB nodes: %w", err)
				}
				log.Infof("KVDB node details: %v", getKVDBNodes)
				for _, each := range getKVDBNodes {
					if each.IsHealthy == true {
						healthyCount++
					}
				}
				if healthyCount == 2 {
					log.Infof("KVDB quorum intact. Healthy count: %d, Expected: 2", healthyCount)
					return nil, false, nil
				}
				log.Errorf("KVDB quorum lost. Healthy count: %d, Expected: 2. Retrying...", healthyCount)
				return nil, true, fmt.Errorf("quorum lost. Healthy count: %d, Expected: 2", healthyCount)
			}
			_, err = task.DoRetryWithTimeout(checkKVDBQuorum, 5*time.Minute, 30*time.Second)
			log.FailOnError(err, "Error occurred while checking KVDB quorum")
			dash.VerifyFatal(healthyCount == 2, true, fmt.Sprintf("verify kvdb quorum is not lost.Healthy count: %d, Expected: 2", healthyCount))
		}

		stepLog = "Bring down one KVDB node for long"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			getKVDBNodes, err := GetAllKvdbNodes()
			log.FailOnError(err, "Unable to get KVDB nodes")
			previousKVDBNodesMap = make(map[string]bool)
			for _, kvdbNode := range getKVDBNodes {
				previousKVDBNodesMap[kvdbNode.ID] = true
			}
			log.Infof("Previous KVDB node details: %v", getKVDBNodes)
			selectedNodeForPxStop = getKVDBNodes[0]
			log.Infof("Selected node for PX stop: %v", selectedNodeForPxStop)
			nodeDetails, err := node.GetNodeDetailsByNodeID(selectedNodeForPxStop.ID)
			log.FailOnError(err, "Unable to retrieve node details for NodeID [%v]", selectedNodeForPxStop.ID)
			stopPXOnKVDBNode(nodeDetails)
		})

		stepLog = "Wait for KVDB Failover to happen and a new node to become kvdb member"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			t := func() (interface{}, bool, error) {
				getKVDBNodes, err := GetAllKvdbNodes()
				if err != nil {
					return nil, true, fmt.Errorf("failed to get KVDB nodes")
				}
				healthyCount := 0
				for _, each := range getKVDBNodes {
					if each.IsHealthy == true {
						healthyCount++
					}
				}
				if healthyCount == 3 {
					log.Infof("KVDB quorum intact. Healthy count: %d, Expected: 3", healthyCount)
					log.Infof("Number of KVDB nodes: %v", len(getKVDBNodes))
					return getKVDBNodes, false, nil
				}
				return nil, true, fmt.Errorf("KVDB quorum not intact. Healthy count: %d, Expected: 3", healthyCount)
			}
			getKVDBNodes, err := task.DoRetryWithTimeout(t, 10*time.Minute, 30*time.Second)
			log.FailOnError(err, "Unable to get KVDB nodes")
			log.Infof("KVDB node details: %v", getKVDBNodes)
		})

		stepLog = "As soon as you see this happening, bring up failed old kvdb node as well"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.Infof("Selected node for PX start: %v", selectedNodeForPxStop)
			nodeDetails, err := node.GetNodeDetailsByNodeID(selectedNodeForPxStop.ID)
			log.FailOnError(err, "Unable to retrieve node details for NodeID [%v]", selectedNodeForPxStop.ID)
			StartVolDriverAndWait([]node.Node{nodeDetails})
			log.Infof("Started volume driver [%s] on nodes [%v]", Inst().V.String(), nodeDetails)
			selectedNodeForPxStop = KvdbNode{}
		})

		stepLog = "Validate kvdb cluster has 3 nodes - 2 old and 1 new member"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			oldKVDBMembers := 0
			newKVDBMembers := 0
			getKVDBNodes, err := GetAllKvdbNodes()
			log.FailOnError(err, "Unable to get KVDB nodes")
			for _, kvdbNode := range getKVDBNodes {
				if previousKVDBNodesMap[kvdbNode.ID] {
					log.Infof("old KVDB node details: %v", getKVDBNodes)
					oldKVDBMembers++
				} else {
					log.Infof("new KVDB node details: %v", getKVDBNodes)
					newKVDBMembers++
				}
			}
			dash.VerifyFatal(oldKVDBMembers == 2 && newKVDBMembers == 1, true, "New KVDB member is not part of the current KVDB nodes!!")
		})

		stepLog = "Validate Applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			ValidateApplications(contexts)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

// Keep volumes in resync state with one replica being in clean state , transition to no-kvdb quorum state, IOs should continue on clean node
var _ = Describe("{HaUpdateWhenClusterInRunFlatState}", Label("staging", "kvdb_ops", "p1", "positive", "HA_Increase_Decrease"), func() {
	/*
		Ticket ID: https://purestorage.atlassian.net/browse/HAZEL-1051
		Deploys apps and make sure volumes are attached to only non-kvdb nodes.
		Increase the repl on all the volumes and at the same time make 2 kvdb nodes down
		Validate apps are running

	*/
	JustBeforeEach(func() {
		StartTorpedoTest("HaUpdateWhenClusterInRunFlatState",
			"Keep volumes in resync state with one replica being in clean state , transition to no-kvdb quorum state, IOs should continue on clean node", nil, 0)
	})
	var (
		contexts = make([]*scheduler.Context, 0)
	)

	itLog := "Keep volumes in resync state with one replica being in clean state , transition to no-kvdb quorum state, IOs should continue on clean node"
	It(itLog, func() {
		log.InfoD(itLog)
		var (
			selectedKvdbNodes []KvdbNode
			kvdbNodes         []KvdbNode
			nonKvdbNodes      []node.Node
			wg                sync.WaitGroup
			expectedRepl      int64
		)
		stepLog = "Get KVDB and Non-KVDB Nodes"
		Step(stepLog, func() {
			kvdbNodes, err = GetAllKvdbNodes()
			log.FailOnError(err, "Failed to retrieve KVDB nodes")
			storageNodes, err := GetStorageNodes()
			log.FailOnError(err, "Failed to retrieve storage nodes")

			kvdbNodeMap := make(map[string]bool)
			for _, kvdbNode := range kvdbNodes {
				kvdbNodeMap[kvdbNode.ID] = true
			}
			for _, storageNode := range storageNodes {
				if _, exists := kvdbNodeMap[storageNode.Id]; !exists {
					nonKvdbNodes = append(nonKvdbNodes, storageNode)
				}
			}
			log.Infof("All storage nodes List are part of KVDB members: [%v]", nonKvdbNodes)
		})

		stepLog = "Schedule application on Non-KVDB node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				taskName := fmt.Sprintf("beforerunflat-%v", i)
				context, err := Inst().S.Schedule(taskName, scheduler.ScheduleOptions{
					AppKeys: Inst().AppList,
					Nodes:   nonKvdbNodes,
				})
				log.FailOnError(err, "Failed to schedule application of %v namespace", taskName)
				contexts = append(contexts, context...)
			}
		})

		ValidateApplications(contexts)

		type volMap struct {
			ReplSet int64
			volObj  *volume.Volume
		}
		volHAMap := make([]*volMap, 0)

		revertReplica := func() {
			log.InfoD("Reverting Replicas based on volHAMap")
			for _, eachvol := range volHAMap {
				getReplicaSets, err := Inst().V.GetReplicaSets(eachvol.volObj)
				log.FailOnError(err, "Failed to get replication factor on the volume")
				if len(getReplicaSets[0].Nodes) != int(eachvol.ReplSet) {
					log.Infof("Reverting Replication factor on Volume [%v] with ID [%v] to [%v]",
						eachvol.volObj.Name, eachvol.volObj.ID, eachvol.ReplSet)
					err := Inst().V.SetReplicationFactor(eachvol.volObj, eachvol.ReplSet,
						nil, nil, true)
					log.FailOnError(err, "failed to set replication value of Volume [%v]", eachvol.volObj.Name)
				}
			}
		}

		setReplOnVolumes := func(volumeSelected *volume.Volume, wait bool, wg *sync.WaitGroup, setReplErrChan chan error) {
			defer wg.Done()
			defer GinkgoRecover()
			var setRepl int64
			currRepl, err := Inst().V.GetReplicationFactor(volumeSelected)
			if err != nil {
				log.Infof("Failed to get Repl factor for vol %s", volumeSelected.Name)
				setReplErrChan <- err
				return
			}

			if currRepl == 3 {
				log.InfoD("The current repl factor is 3, Decrease HA of all PVCs in this app to (current repl - 1) Before HA Increase!")
				stepLog = "Decrease HA of all PVCs in this app to (current repl - 1)"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					opts := volume.Options{ValidateReplicationUpdateTimeout: validateReplicationUpdateTimeout}
					err = Inst().V.SetReplicationFactor(volumeSelected, currRepl-1, nil, nil, true, opts)
					if err != nil {
						err = fmt.Errorf("err setting repl factor  to %d for  vol : %s", setRepl, volumeSelected.Name)
						log.Infof("error while setting the repl factor %v", err)
						setReplErrChan <- err
					}
				})
			}

			log.InfoD("Increase HA of all PVCs in this app to (current repl + 1)")
			currRepl, err = Inst().V.GetReplicationFactor(volumeSelected)
			if err != nil {
				log.Infof("Failed to get Repl factor for vol %s", volumeSelected.Name)
				setReplErrChan <- err
				return
			}
			opts := volume.Options{ValidateReplicationUpdateTimeout: validateReplicationUpdateTimeout}
			err = Inst().V.SetReplicationFactor(volumeSelected, currRepl+1, nil, nil, wait, opts)
			if err != nil {
				err = fmt.Errorf("err setting repl factor  to %d for  vol : %s", setRepl, volumeSelected.Name)
				log.Infof("error while setting the repl factor %v", err)
				setReplErrChan <- err
			}
		}

		getReplFactors := func(vol *volume.Volume, wg *sync.WaitGroup) {
			defer wg.Done()
			defer GinkgoRecover()
			volDet := volMap{}
			curReplSet, err := Inst().V.GetReplicationFactor(vol)
			log.FailOnError(err, "failed to get replication factor of the volume")
			volDet.volObj = vol
			volDet.ReplSet = curReplSet
			expectedRepl = curReplSet
			if curReplSet != 3 {
				expectedRepl = curReplSet + 1
			}
			log.Infof("Volume [%v] is with HA [%v]", volDet.volObj.Name, volDet.ReplSet)
			volHAMap = append(volHAMap, &volDet)
		}

		for _, eachCtx := range contexts {
			vols, err := Inst().S.GetVolumes(eachCtx)
			log.FailOnError(err, "Failed to get list of Volumes in the cluster")

			for _, eachVol := range vols {

				// Check if volumes are Pure FA/FB DA volumes
				isPureVol, err := Inst().V.IsPureVolume(eachVol)
				log.FailOnError(err, "Failed to check is PURE volume")
				dash.VerifyFatal(isPureVol, false, fmt.Sprintf("Repl increase on Pure DA Volume [%s] not supported.Skiping this operation", eachVol.Name))

				wg.Add(1)
				log.Infof("Get Repl factor for Volume [%v]", eachVol.Name)
				go getReplFactors(eachVol, &wg)
			}
		}
		wg.Wait()

		cleanup := func() {
			log.InfoD("Executing cleanup tasks")
			if len(selectedKvdbNodes) > 0 {
				for _, n := range selectedKvdbNodes {
					nodeDetails, err := node.GetNodeDetailsByNodeID(n.ID)
					log.FailOnError(err, "Failed to retrieve node details for NodeID [%v]", n.ID)

					err = Inst().V.StartDriver(nodeDetails)
					log.FailOnError(err, "Failed to start Portworx driver on node %s", nodeDetails.Name)
					err = Inst().V.WaitDriverUpOnNode(nodeDetails, 10*time.Minute)
					log.FailOnError(err, "Failed to waiting for Portworx driver to start on node %s", nodeDetails.Name)
					log.InfoD("Successfully started Portworx on KVDB node: %v", nodeDetails.Name)
				}
			}
			DestroyApps(contexts, nil)
		}
		defer cleanup()

		stepLog = "perform HA update on Volumes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			batchSize := 5
			errOccured := false
			// Set Repl Factor on all the volumes at ones
			for i := 0; i < len(volHAMap); i += batchSize {
				n := i + batchSize
				if n > len(volHAMap) {
					n = len(volHAMap)
				}
				setReplErrChan := make(chan error, batchSize)
				for _, eachVol := range volHAMap[i:n] {
					wg.Add(1)
					go setReplOnVolumes(eachVol.volObj, false, &wg, setReplErrChan)
				}
				wg.Wait()
				close(setReplErrChan)

				for err := range setReplErrChan {
					log.Errorf("failed to set repl factor to curr+1: %v", err)
					errOccured = true
				}
				if errOccured {
					log.FailOnError(fmt.Errorf("one or more errors occured while setting repl factor to curr+1"), "failed to set repl factor to curr+1 for the volumes")
				}
			}
		})

		stepLog = "Verify volume replica are in resync state"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			// Sleep for some time before checking if any resync to start
			time.Sleep(2 * time.Minute)
			for _, eachVol := range volHAMap {
				if !WaitTillVolumeInResync(eachVol.volObj.ID) {
					log.FailOnError(fmt.Errorf("volume are not in resync state"), "Failed to get Volume in Resync state [%s]", eachVol.volObj.ID)
				}
			}
		})

		stepLog = "Stopping Portworx service on selected KVDB nodes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
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
			expectOut := "Volume and node operations may be unavailable but I/O will continue"
			dash.VerifyFatal(strings.Contains(output, expectOut), true, "Is cluster in run-flat state?")
		})

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

			storageNodes := node.GetStorageNodes()
			kvdbMembers, err := Inst().V.GetKvdbMembers(storageNodes[0])
			log.FailOnError(err, "Failed to retrieve KVDB members list")

			err = kvdbutils.ValidateKVDBMembers(kvdbMembers)
			log.FailOnError(err, "Failed to validate KVDB members")

			output, err := runCmd("pxctl status", storageNodes[0])
			log.FailOnError(err, "Failed to execute pxctl status on node: %v", storageNodes[0].Name)
			dash.VerifyFatal(!strings.Contains(output, "Warning"), true, "Output contains warnings. Is the cluster healthy?")
		})

		stepLog = "Verify HA update is successful"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, vol := range volHAMap {
				err = ValidateReplFactorUpdate(vol.volObj, expectedRepl)
				if err != nil {
					replStatus, err1 := GetVolumeReplicationStatus(vol.volObj)
					log.FailOnError(err1, fmt.Sprintf("failed to get repl status for the vol %s", vol.volObj.Name))
					log.Infof("got replication status for the vol %s as %s", vol.volObj.Name, replStatus)
					log.FailOnError(err, "error in ha-increase after pool resize")
				}
				log.Infof("verified HA update for the vol %s", vol.volObj.ID)
			}
		})

		stepLog = "Revert replica and validate applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			revertReplica()
			ValidateApplications(contexts)
		})
	})

	JustAfterEach(func() {
		EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{VerifyFstrimWithPoolOffline}", Label("staging", "p0", "positive", "px_ops", "pool_ops"), func() {

	/*
	   https://purestorage.atlassian.net/browse/HAZEL-1071
	   1. Enabel scheduled FSTrim on the cluster
	   2. Verify scheduled FSTrim with pool full / offline
	*/

	JustBeforeEach(func() {
		StartTorpedoTest("VerifyFstrimWithPoolOffline", "Verify Fstrim Schedule with pool full / offline", nil, 0)
	})
	var (
		contexts = make([]*scheduler.Context, 0)
	)

	stepLog := "Create volumes, make pool full / offline with scheduled fs trim"
	It(stepLog, func() {
		log.InfoD(stepLog)

		var (
			selectedNode      *node.Node
			secondReplNode    node.Node
			fsTrimRunningNode node.Node
			fsTrimStatuses    map[string]opsapi.FilesystemTrim_FilesystemTrimStatus
			applist           = Inst().AppList
			stNodes           []node.Node
		)

		cleanup := func() {
			log.Info("Executing cleanup tasks")
			DestroyApps(contexts, nil)
			Inst().AppList = applist
			_ = Inst().V.SetClusterOpts(*selectedNode, map[string]string{"--fstrim-schedule-start": ""})
			err = Inst().S.RemoveLabelOnNode(*selectedNode, k8s.NodeType)
			log.FailOnError(err, "error removing label on node [%s]", selectedNode.Name)
			err = Inst().S.RemoveLabelOnNode(secondReplNode, k8s.NodeType)
			log.FailOnError(err, "error removing label on node [%s]", secondReplNode.Name)
		}
		defer cleanup()

		stepLog = fmt.Sprintf("Select nodes and schedule application for FSTrim validation")
		Step(stepLog, func() {
			log.InfoD(stepLog)

			selectedNode = GetNodeWithLeastSize()
			stNodes = node.GetStorageNodes()
			for _, stNode := range stNodes {
				if stNode.Name != selectedNode.Name {
					secondReplNode = stNode
				}
			}

			err = Inst().S.AddLabelOnNode(*selectedNode, k8s.NodeType, k8s.FastpathNodeType)
			log.FailOnError(err, fmt.Sprintf("Failed add label on node %s", selectedNode.Name))
			err = Inst().S.AddLabelOnNode(secondReplNode, k8s.NodeType, k8s.FastpathNodeType)
			log.FailOnError(err, fmt.Sprintf("Failed add label on node %s", secondReplNode.Name))

			Inst().AppList = []string{"fio-fastpath"}
			contexts = make([]*scheduler.Context, 0)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("poolfulfs-%d", i))...)
			}
		})
		ValidateApplications(contexts)

		stepLog = fmt.Sprintf("Enabel nodiscard and scheduled fstrim on volumes")
		Step(stepLog, func() {
			log.InfoD(stepLog)

			stepLog = fmt.Sprintf("Enabel nodiscard on the volumes")
			Step(stepLog, func() {
				log.InfoD(stepLog)
				for _, ctx := range contexts {
					appVolumes, err := Inst().S.GetVolumes(ctx)
					log.FailOnError(err, "Failed to get volumes for app %s", ctx.App.Key)

					for _, appvolume := range appVolumes {
						apivol, err := Inst().V.InspectVolume(appvolume.ID)
						log.FailOnError(err, "Failed to get the volume details for the ID: %v", appvolume.ID)
						attachedNode := apivol.AttachedOn
						fsTrimRunningNode, err = node.GetNodeByIP(attachedNode)
						log.FailOnError(err, "Failed to get the details for the node ID: %v", attachedNode)
						log.Infof("Volume [%v] is attached on the node: %v", appvolume.ID, fsTrimRunningNode.Name)

						EnableNodiscardOnVolume(fsTrimRunningNode, appvolume)
						log.Infof("Enabled nodiscard on the volume [%v]", appvolume.ID)
					}
				}

			})

			stepLog = fmt.Sprintf("Enabel scheduled fs trim on the cluster")
			Step(stepLog, func() {
				log.InfoD(stepLog)
				// Daily: daily=HH:MM, Weekly: weekly=day@hh:mm
				log.Infof("Enable scheduled fs trim on the cluster")
				formattedTime := time.Now().UTC().Add(1 * time.Minute).Format("15:04")
				scheduleStartTime := fmt.Sprintf("daily=%s", formattedTime)
				err := EnableScheduledFSTrim(stNodes, 10, scheduleStartTime)
				log.FailOnError(err, "failed to enable scheduled fs trim")

				fsTrimStatuses, err = Inst().V.GetAutoFsTrimStatus(selectedNode.DataIp)
				log.FailOnError(err, "Failed to get the fstrim status for the node: %v", selectedNode.Name)
			})
		})

		stepLog = fmt.Sprintf("verify the pool to become full and go offline status")
		Step(stepLog, func() {
			log.InfoD(stepLog)

			err = WaitForPoolOffline(*selectedNode)
			log.FailOnError(err, fmt.Sprintf("Failed to make node %s storage down", selectedNode.Name))

			poolsStatus, err := Inst().V.GetNodePoolsStatus(*selectedNode)
			log.FailOnError(err, "error getting pool status on node %s", selectedNode.Name)

			var offlinePoolUUID string
			for i, s := range poolsStatus {
				if s == "Offline" {
					offlinePoolUUID = i
					break
				}
			}
			dash.VerifyFatal(poolsStatus[offlinePoolUUID], "Offline", fmt.Sprintf("verify status for the pool %s, current status %v", offlinePoolUUID, poolsStatus[offlinePoolUUID]))
		})

		stepLog = fmt.Sprintf("verify FSTrim is running after pool status offline")
		Step(stepLog, func() {
			log.InfoD(stepLog)
			newFsTrimStatusesAftrRestart, err := Inst().V.GetAutoFsTrimStatus(selectedNode.DataIp)
			log.FailOnError(err, "error getting fstrim status for the node %v", selectedNode.DataIp)
			for k := range fsTrimStatuses {
				val, ok := newFsTrimStatusesAftrRestart[k]
				dash.VerifySafely(ok, true, fmt.Sprintf("verify fstrim started for volume %s", k))
				dash.VerifySafely(val != opsapi.FilesystemTrim_FS_TRIM_FAILED, true, fmt.Sprintf("verify fstrim for the volume %s, current status is %v", k, val))
			}
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})

})

var _ = Describe("{StorageLessToStorageByShutDownNode}", Label("staging", "p0", "negative", "error_injection", "node_ops"), func() {

	/*
	   https://purestorage.atlassian.net/browse/HAZEL-1065
	   1. Bring down one storage node
	   2. Validate one of the storageless nodes has become storage node by picking up the cloud drives that were attached to the down storage node
	*/

	var (
		testrailID          = 0
		runID               int
		contexts            []*scheduler.Context
		storageNodes        []node.Node
		storageLessNodesIds []string
		selectedNode        node.Node
		isStorageNode       = false
	)
	JustBeforeEach(func() {
		StartTorpedoTest("StorageLessToStorageByShutDownNode", "Transition storage-less to Storage node by stopping one of the storage nodes", nil, testrailID)
		runID = testrailuttils.AddRunsToMilestone(testrailID)
	})

	stepLog := "Transition storage-less to Storage node by stopping one of the storage nodes"
	It(stepLog, func() {
		log.InfoD(stepLog)

		cleanup := func() {
			log.Info("Executing cleanup tasks")
			allNodes := node.GetStorageDriverNodes()
			for _, n := range allNodes {
				err = Inst().S.IsNodeReady(n)
				if err != nil {
					log.InfoD("Powering on node [%s]", n.Name)
					err := Inst().N.PowerOnVM(n)
					log.FailOnError(err, "error powering on node [%s]", n.Name)
				}
				err = Inst().V.WaitDriverUpOnNode(n, 15*time.Minute)
				log.FailOnError(err, "error while waiting for driver up on node %s", n.Name)
			}
			DestroyApps(contexts, nil)
		}
		defer cleanup()

		stepLog = "Validate cluster has at least one storage-less node"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			storageNodes = node.GetStorageNodes()
			selectedNode = GetRandomNode(storageNodes)
			log.InfoD("Check if the cluster has at least one storage-less node before proceeding")
			storageLessNodes := node.GetStorageLessNodes()
			if len(storageLessNodes) == 0 {
				log.InfoD("Skipping the test as it required atleast one storage less node")
				Skip("Skipping the test as it required atleast one storage less node")
			}
			for _, n := range storageLessNodes {
				storageLessNodesIds = append(storageLessNodesIds, n.Name)
			}
		})

		stepLog = "Powering off the storage node in cluster"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			namespace, err := Inst().S.GetPortworxNamespace()
			log.FailOnError(err, "error getting the portworx namespace")

			pxOperator := operator.Instance()
			stcList, err := pxOperator.ListStorageClusters(namespace)
			log.FailOnError(err, "error listing the storage clusters for the namespace %s", namespace)

			stc, err := pxOperator.GetStorageCluster(stcList.Items[0].Name, stcList.Items[0].Namespace)
			log.FailOnError(err, "error getting the storage cluster %s", stcList.Items[0].Name)
			pxCloudDriveConfigMap, err := Inst().S.GetPXCloudDriveConfigMap(stc)
			log.FailOnError(err, "error getting the PX cloud drive config map")

			err = Inst().N.DetachDrivesFromVM(selectedNode.Name, pxCloudDriveConfigMap)
			log.FailOnError(err, "error detaching the drives from the node %s", selectedNode.Name)
			time.Sleep(1 * time.Minute)
			err = Inst().V.RefreshDriverEndpoints()
			log.FailOnError(err, "error refreshing the driver endpoints")

			log.InfoD("Powering off the node [%s]", selectedNode.Name)
			err = Inst().N.PowerOffVM(selectedNode)
			log.FailOnError(err, "error powering off node [%s]", selectedNode.Name)

			time.Sleep(20 * time.Minute)
		})

		stepLog = "Validate if the storageless become storage node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			checkStorageNodeAvailable := func() (interface{}, bool, error) {
				for _, eachId := range storageLessNodesIds {
					stNode, err := node.GetNodeByName(eachId)
					if err != nil {
						return nil, true, err
					}

					output, err := RunCmdGetOutput("pxctl sv pool show", stNode)
					if err != nil {
						return nil, true, err
					}
					log.Infof("Pool status: %v", output)
					if strings.Contains(output, "Pool ID:") && stNode.Status == api.Status_STATUS_OK {
						isStorageNode = true
						break
					}
				}
				if isStorageNode {
					return nil, false, nil
				}
				log.Infof("Storageless node has not transitioned yet, retrying...")
				return nil, true, fmt.Errorf("storageless node transition not completed")
			}
			_, err := task.DoRetryWithTimeout(checkStorageNodeAvailable, 30*time.Minute, 2*time.Minute)
			log.FailOnError(err, "error waiting for storageless node to transition to storage node")
			dash.VerifyFatal(isStorageNode, true, "Verify storageless become storage node")
		})

		stepLog = "Powering on the shout down node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.InfoD("Powering on node [%s]", selectedNode.Name)
			err := Inst().N.PowerOnVM(selectedNode)
			log.FailOnError(err, "error powering on node [%s]", selectedNode.Name)

			log.Infof("Waiting 5 mins to power on the VM [%s]", selectedNode.Name)
			time.Sleep(5 * time.Minute)
			stNode, err := node.GetNodeByName(selectedNode.Name)
			log.FailOnError(err, "error getting the node [%s]", selectedNode.Name)
			err = Inst().V.WaitDriverUpOnNode(stNode, 10*time.Minute)
			log.FailOnError(err, "error while waiting for driver up on node %s", selectedNode.Name)
			log.Infof("Node powered on successfully after storage less to storage node transition")

			err = Inst().S.RefreshNodeRegistry()
			log.FailOnError(err, "error refreshing the node registry")
			log.Infof("Refreshed the node registry Successfully")

			err = Inst().V.RefreshDriverEndpoints()
			log.FailOnError(err, "error refreshing the driver endpoints")
			log.Infof("Refreshed the driver endpoints Successfully")
		})

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts, testrailID, runID)
	})
})

var _ = Describe("{VerifyFstrimWithPVCRresize}", Label("staging", "p0", "positive", "px_vol_ops"), func() {
	/*
	   https://purestorage.atlassian.net/browse/HAZEL-1069

	   1. Enabel scheduled FSTrim on the cluster
	   2. Resize pvc / volume
	   3. Verify scheduled FSTrim continue to happen
	*/

	JustBeforeEach(func() {
		StartTorpedoTest("VerifyFstrimWithPVCRresize", "Verify Fstrim schedule with Volume / pvc resize", nil, 0)
	})
	var (
		contexts = make([]*scheduler.Context, 0)
	)

	stepLog := "Verify FS trim continues after volume / pvc resize"
	It(stepLog, func() {
		log.InfoD(stepLog)

		var (
			storageNodes        []node.Node
			selectedStorageNode node.Node
			appList             = Inst().AppList
			fsTrimTimeout       = 60 * time.Minute
			fsTrimRetryInterval = 2 * time.Minute
			fsTrimStatuses      map[string]opsapi.FilesystemTrim_FilesystemTrimStatus
		)

		stepLog = "Enable Fstrim schedule on the cluster"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			storageNodes = node.GetStorageNodes()
			if len(storageNodes) == 0 {
				log.FailOnError(fmt.Errorf("Storage nodes list empty"), "failed to get storage nodes")
			}
			selectedStorageNode = storageNodes[0]

			// Daily: daily=HH:MM, Weekly: weekly=day@hh:mm
			log.Infof("Enable scheduled fs trim on the cluster")
			formattedTime := time.Now().UTC().Add(1 * time.Minute).Format("15:04")
			scheduleStartTime := fmt.Sprintf("daily=%s", formattedTime)
			err := EnableScheduledFSTrim(storageNodes, 10, scheduleStartTime)
			log.FailOnError(err, "failed to enable scheduled fs trim")

			Inst().AppList = []string{"fio-fstrim"}
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("fspxrestart-%d", i))...)
			}
		})
		ValidateApplications(contexts)

		defer func() {
			DestroyApps(contexts, nil)
			_ = Inst().V.SetClusterOpts(selectedStorageNode, map[string]string{
				"--fstrim-schedule-start": ""})
			Inst().AppList = appList
		}()

		stepLog = "select a node where fs trim running and get the fs trim status"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, ctx := range contexts {
				appVolumes, err := Inst().S.GetVolumes(ctx)
				log.FailOnError(err, "Failed to get volumes list for the application: %v", ctx.App.Key)

				for _, v := range appVolumes {
					selectedVol, err := Inst().V.InspectVolume(v.ID)
					log.FailOnError(err, "Failed to get volumes details using volume ID: %v", v.ID)

					attachedNode := selectedVol.AttachedOn
					log.Infof("Volume attached on node: %v", attachedNode)
					selectedStorageNode, err = node.GetNodeByIP(attachedNode)
					log.FailOnError(err, "Failed to get the node details by nodeIP: %v", attachedNode)

					fsTrimStatuses, err = CheckFSTrimRunningOnNode(selectedStorageNode, selectedVol, fsTrimTimeout, fsTrimRetryInterval)
					log.FailOnError(err, "Failed to get the fstrim status for the node: %v", selectedStorageNode.DataIp)
					break
				}
			}
		})

		stepLog = "Resize pvc for all volumes in the application"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, ctx := range contexts {
				if !strings.Contains(ctx.App.Key, "fio-fstrim") {
					continue
				}
				appVols, err := Inst().S.GetVolumes(ctx)
				log.FailOnError(err, fmt.Sprintf("error getting volumes list for the app [%s]", ctx.App.Key))

				for _, vol := range appVols {
					apiVol, err := Inst().V.InspectVolume(vol.ID)
					log.FailOnError(err, fmt.Sprintf("error getting volume details using volume id: [%s]", vol.ID))

					curSize := apiVol.Spec.Size
					newSize := curSize + (uint64(1) * units.GiB)
					log.Infof("Initiating volume size increase on volume [%s/%v] by size [%v]GiB to [%v]GiB", ctx.App.Key,
						vol.ID, curSize/units.GiB, newSize/units.GiB)

					pvcs, err := GetAllPVCFromNs(ctx.App.NameSpace, nil)
					log.FailOnError(err, fmt.Sprintf("error getting pvc list for the app [%v], in the namespace [%v]", ctx.App.Key, ctx.App.NameSpace))

					for _, pvc := range pvcs {
						log.Debugf("checking pvc:[%s], with vol name [%s]", pvc.Name, vol.Name)
						if pvc.Name == vol.Name {
							log.InfoD("increasing pvc [%s/%s] size to %dGiB", pvc.Namespace, pvc.Name, newSize/units.GiB)
							_, err = Inst().S.ResizePVC(ctx, &pvc, uint64(1))
							log.FailOnError(err, fmt.Sprintf("error getting while resize pvc [%s]", pvc.Name))
						}

					}

					// Wait for 2 seconds for Volume to update stats
					time.Sleep(2 * time.Second)
					volumeInspect, err := Inst().V.InspectVolume(vol.ID)
					log.FailOnError(err, fmt.Sprintf("error getting volume details using volume id: [%s]", vol.ID))

					updatedSize := volumeInspect.Spec.Size
					dash.VerifyFatal(updatedSize > curSize, true, fmt.Sprintf("verify pvc size increased. current size: %vGiB, updated size: %vGiB", curSize, updatedSize))
				}
				log.Infof("Volumes [%v] for the app [%v] successfully resized", appVols, ctx.App.Key)
			}
		})

		stepLog = "verifiy that FSTrim schedule running after volume / pvc resize"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			newFsTrimStatusesAftrRestart, err := Inst().V.GetAutoFsTrimStatus(selectedStorageNode.DataIp)
			log.FailOnError(err, "error getting autofs status")
			for k := range fsTrimStatuses {
				val, ok := newFsTrimStatusesAftrRestart[k]
				dash.VerifySafely(ok, true, fmt.Sprintf("verify autofstrim started for volume %s", k))
				dash.VerifySafely(val != opsapi.FilesystemTrim_FS_TRIM_FAILED, true, fmt.Sprintf("verify fstrim status for volume %s, current status %v", k, val))
			}
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{VerifyFstrimWithPXRestartAndHAIncrease}", Label("staging", "p0", "positive", "px_vol_ops", "MiniScale", "HA_Increase_Decrease"), func() {
	/*
	   https://purestorage.atlassian.net/browse/HAZEL-1075

	   1. Enable scheduled FSTrim on the cluster
	   2. Restart PX and Increase HA on some nodes
	   3. Verify scheduled FSTrim continue to happen
	*/

	var (
		contexts                    []*scheduler.Context
		storageNodes                []node.Node
		selectedStorageNode         node.Node
		selectedNodeForRestart      node.Node
		selectedVol                 *opsapi.Volume
		volAttachedNodeAfterRestart string
		fsTrimStatusesBfrRestart    map[string]opsapi.FilesystemTrim_FilesystemTrimStatus
		appList                     = Inst().AppList
		fsTrimTimeout               = 60 * time.Minute
		fsTrimRetryInterval         = 2 * time.Minute
	)

	JustBeforeEach(func() {
		StartTorpedoTest("VerifyFstrimWithPXRestartAndHAIncrease", "Restart PX while Trim is in progress and HA increase is in progress on some of the nodes", nil, 0)
	})

	stepLog := "Restart PX while Trim is in progress and HA increase is in progress on some of the nodes"
	It(stepLog, func() {
		log.InfoD(stepLog)

		defer func() {
			DestroyApps(contexts, nil)
			_ = Inst().V.SetClusterOpts(selectedStorageNode, map[string]string{
				"--fstrim-schedule-start": ""})
			Inst().AppList = appList
		}()

		stepLog = "Enable Fstrim schedule on the cluster and schedule application"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			storageNodes = node.GetStorageNodes()
			if len(storageNodes) == 0 {
				log.FailOnError(fmt.Errorf("Storage nodes list empty"), "failed to get storage nodes")
			}
			selectedStorageNode = storageNodes[0]

			// Daily: daily=HH:MM, Weekly: weekly=day@hh:mm
			log.Infof("Enable scheduled fs trim on the cluster")
			formattedTime := time.Now().UTC().Add(1 * time.Minute).Format("15:04")
			scheduleStartTime := fmt.Sprintf("daily=%s", formattedTime)
			log.Infof("Scheduled fstrom start time: %v", scheduleStartTime)
			err := EnableScheduledFSTrim(storageNodes, 10, scheduleStartTime)
			log.FailOnError(err, "failed to enable scheduled fs trim")

			Inst().AppList = []string{"fio-fstrim"}
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("fspxrestart-%d", i))...)
			}
		})
		ValidateApplications(contexts)

		stepLog = "select a node where fs trim running and get the fs trim status"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, ctx := range contexts {
				appVolumes, err := Inst().S.GetVolumes(ctx)
				log.FailOnError(err, "Failed to get volumes list for the application: %v", ctx.App.Key)

				for _, v := range appVolumes {
					selectedVol, err = Inst().V.InspectVolume(v.ID)
					log.FailOnError(err, "Failed to get volumes details using volume ID: %v", v.ID)

					attachedNode := selectedVol.AttachedOn
					log.Infof("Volume attached on node: %v", attachedNode)
					selectedNodeForRestart, err = node.GetNodeByIP(attachedNode)
					log.FailOnError(err, "Failed to get the node details by nodeIP: %v", attachedNode)

					checkFSTrimRunning := func() (interface{}, bool, error) {
						fsTrimStatusesBfrRestart, err = Inst().V.GetAutoFsTrimStatus(selectedNodeForRestart.DataIp)
						if err != nil {
							return nil, true, fmt.Errorf("failed to get AutoFstrim status for node [%v]: %v", selectedNodeForRestart, err)
						}
						if status, exists := fsTrimStatusesBfrRestart[selectedVol.Id]; exists {
							if status != opsapi.FilesystemTrim_FS_TRIM_FAILED {
								return true, false, nil
							}
							return false, true, fmt.Errorf("FSTrim status failed for the volume %v", selectedVol.Id)
						}
						return false, true, fmt.Errorf("FSTrim status not available for the volume %v, Retrying...", selectedVol.Id)
					}
					_, err = task.DoRetryWithTimeout(checkFSTrimRunning, fsTrimTimeout, fsTrimRetryInterval)
					log.FailOnError(err, "Failed to get the fstrim status for the node: %v", selectedNodeForRestart.DataIp)
					break
				}
			}
		})

		stepLog = "Update replication factor on volumes"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, ctx := range contexts {
				appVolumes, err := Inst().S.GetVolumes(ctx)
				log.FailOnError(err, "Failed to get volumes list for the application: %v", ctx.App.Key)

				for _, v := range appVolumes {
					currRep, err := Inst().V.GetReplicationFactor(v)
					log.FailOnError(err, "Failed to get Repl factor for volume %s", v.Name)

					appVol, err := Inst().V.InspectVolume(v.ID)
					log.FailOnError(err, fmt.Sprintf("error inspecting volume [%s]", v.ID))

					if _, ok := fsTrimStatusesBfrRestart[appVol.Id]; ok {
						opts := volume.Options{
							ValidateReplicationUpdateTimeout: validateReplicationUpdateTimeout,
						}
						if currRep == 3 {
							log.Infof("Current replica is %v, reducing the replica value to %v", currRep, currRep-1)
							err = Inst().V.SetReplicationFactor(v, currRep-1, nil, nil, true, opts)
							log.FailOnError(err, "Failed to set replication factor for volume: %v", v.Name)

							newRep, err := Inst().V.GetReplicationFactor(v)
							log.FailOnError(err, "Failed to get replication factor for volume: %v", v.Name)
							dash.VerifyFatal(newRep < currRep, true, fmt.Sprintf("Validate set repl factor to %d", currRep-1))
						} else {
							log.Infof("Current replication is %v, increasing the replica value to %v", currRep, currRep+1)
							err = Inst().V.SetReplicationFactor(v, currRep+1, nil, nil, true, opts)
							log.FailOnError(err, "Failed to set replication factor for volume: %v", v.Name)

							newRep, err := Inst().V.GetReplicationFactor(v)
							log.FailOnError(err, "Failed to get replication factor for volume: %v", v.Name)
							dash.VerifyFatal(newRep > currRep, true, fmt.Sprintf("Validate set repl factor to %d", currRep+1))
						}
					}
				}
			}
		})

		stepLog = "Restart PX on the selected node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			err = Inst().V.RestartDriver(selectedNodeForRestart, nil)
			log.FailOnError(err, "Error occured while Restart PX on node: %v", selectedNodeForRestart.Name)
			log.Infof("PX restarted successfully on node %v", selectedNodeForRestart.Name)

			selectedVolafterRestart, err := Inst().V.InspectVolume(selectedVol.Id)
			log.FailOnError(err, "Failed to get the volume details for: %v", selectedVol.Id)

			volAttachedNodeAfterRestart = selectedVolafterRestart.AttachedOn
			log.Infof("Volume attcahed node after node restart: %v", volAttachedNodeAfterRestart)
			log.Infof("Waiting for five minutes to scheduled fs trim to start after node restart")
			time.Sleep(5 * time.Minute)
		})

		stepLog = "Verify FSTrim schedule continues after PX restart"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			newFsTrimStatusesAftrRestart, err := Inst().V.GetAutoFsTrimStatus(volAttachedNodeAfterRestart)
			log.FailOnError(err, "error getting autofs status")
			for k := range fsTrimStatusesBfrRestart {
				val, ok := newFsTrimStatusesAftrRestart[k]
				dash.VerifySafely(ok, true, fmt.Sprintf("verify autofstrim started for volume %s", k))
				dash.VerifySafely(val != opsapi.FilesystemTrim_FS_TRIM_FAILED, true, fmt.Sprintf("verify fstrim status for volume %s, current status %v", k, val))
			}
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{VerifyFstrimWithNodeRestart}", Label("staging", "p0", "positive", "px_vol_ops"), func() {
	/*
	   https://purestorage.atlassian.net/browse/HAZEL-1070
	   1. Enabel scheduled FSTrim on the cluster
	   2. Reboot the node
	   3. Verify scheduled FSTrim continue to happen
	*/

	JustBeforeEach(func() {
		StartTorpedoTest("VerifyFstrimWithNodeRestart", "Verify Fstrim schedule after Node Reboot", nil, 0)
	})
	var (
		contexts = make([]*scheduler.Context, 0)
	)

	stepLog := "Verify Fstrim schedule after Node Reboot"
	It(stepLog, func() {
		log.InfoD(stepLog)

		var (
			storageNodes                []node.Node
			selectedStorageNode         node.Node
			selectedNodeForRestart      node.Node
			selectedVol                 *opsapi.Volume
			volAttachedNodeAfterRestart string
			volAttachedNode             node.Node
			fsTrimStatusesBfrRestart    map[string]opsapi.FilesystemTrim_FilesystemTrimStatus
			appList                     = Inst().AppList
			fsTrimTimeout               = 120 * time.Minute
			fsTrimRetryInterval         = 5 * time.Minute
		)

		cleanup := func() {
			log.Info("Executing cleanup tasks")
			DestroyApps(contexts, nil)
			_ = Inst().V.SetClusterOpts(selectedStorageNode, map[string]string{
				"--fstrim-schedule-start": ""})
			Inst().AppList = appList
		}
		defer cleanup()

		stepLog = "Enable Fstrim schedule on the cluster and schedule application"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			storageNodes = node.GetStorageNodes()
			selectedStorageNode = storageNodes[0]

			// Daily: daily=HH:MM, Weekly: weekly=day@hh:mm
			log.Infof("Enable scheduled fs trim on the cluster")
			formattedTime := time.Now().UTC().Add(1 * time.Minute).Format("15:04")
			scheduleStartTime := fmt.Sprintf("daily=%s", formattedTime)
			err := EnableScheduledFSTrim(storageNodes, 10, scheduleStartTime)
			log.FailOnError(err, "failed to enable scheduled fs trim")

			Inst().AppList = []string{"fio-fstrim"}
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("fspxrestart-%d", i))...)
			}
		})
		ValidateApplications(contexts)

		stepLog = "Select a node where fs trim running and get the fs trim status"
		Step(stepLog, func() {
			for _, ctx := range contexts {
				appVolumes, err := Inst().S.GetVolumes(ctx)
				log.FailOnError(err, "Failed to get volumes list for the application: %v", ctx.App.Key)

				for _, v := range appVolumes {
					vol, err := Inst().V.InspectVolume(v.ID)
					log.FailOnError(err, "Failed to get volumes details using: %v", v.ID)

					selectedVol = vol
					attachedNode := vol.AttachedOn
					log.Infof("Volume attached on node: %v", attachedNode)
					selectedNodeForRestart, err = node.GetNodeByIP(attachedNode)
					log.FailOnError(err, "Failed to get the node details by node IP: %v", attachedNode)

					_, err = CheckFSTrimRunningOnNode(selectedNodeForRestart, vol, fsTrimTimeout, fsTrimRetryInterval)
					log.FailOnError(err, "Failed to get the fstrim status for the node: %v", selectedNodeForRestart.DataIp)
					break
				}
			}
			fsTrimStatusesBfrRestart, err = Inst().V.GetAutoFsTrimStatus(selectedNodeForRestart.DataIp)
			log.FailOnError(err, "Failed to get the fstrim status for the node: %v", selectedNodeForRestart.Name)
		})

		stepLog = "Reboot the selected node and get volume attached on node after restart"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			err = Inst().N.RebootNodeAndWait(selectedNodeForRestart)
			log.FailOnError(err, "Failed to reboot node and wait till it is up")

			selectedVolafterRestart, err := Inst().V.InspectVolume(selectedVol.Id)
			log.FailOnError(err, "Failed to get the volume details for: %v after node reboot", selectedVol.Id)

			volAttachedNodeAfterRestart = selectedVolafterRestart.AttachedOn
			volAttachedNode, err = node.GetNodeByIP(volAttachedNodeAfterRestart)
			log.FailOnError(err, "Failed to get the node details by IP: %v", volAttachedNodeAfterRestart)
			log.Infof("Volume attcahed node after node reboot: %v", volAttachedNodeAfterRestart)
		})

		stepLog = "verifiy FSTrim schedule running after node restart"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			newFsTrimStatusesAftrRestart, err := CheckFSTrimRunningOnNode(volAttachedNode, selectedVol, fsTrimTimeout, fsTrimRetryInterval)
			log.FailOnError(err, "Failed to get the fstrim status for the node: %v", volAttachedNodeAfterRestart)

			for k := range fsTrimStatusesBfrRestart {
				val, ok := newFsTrimStatusesAftrRestart[k]
				dash.VerifySafely(ok, true, fmt.Sprintf("verify fstrim started for volume %s", k))
				dash.VerifySafely(val != opsapi.FilesystemTrim_FS_TRIM_FAILED, true, fmt.Sprintf("verify fstrim status for the volume %s, current status %v", k, val))
			}
		})

	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{StopPXOnKVDBNodeAndNewKVDBNodeWillBeUp}", Label("staging", "p0", "positive", "kvdb_ops"), func() {
	/*
		https://purestorage.atlassian.net/browse/HAZEL-1080
		1. Stop PX on kvdb Node
		2. Wait for 5 mins and restart PX on stopped node
		3. New kvdb node would comeup
		3. Make sure old kvdb node is no longer kvdb memeber and should be devoid of kvdb driver
	*/
	JustBeforeEach(func() {
		StartTorpedoTest("StopPXOnKVDBNodeAndNewKVDBNodeWillBeUp", "Stop PX on KVDB node and make sure the stopped node is no longer a KVDB member.Stopped kvdb node should be devoid of kvdb driver", nil, 0)
		//check if DMThin enabled, if so skip the KVDBNodePXStopAndStart trigger
		isDMthin, err := IsDMthin()
		log.FailOnError(err, "Failed to check if DMthin is enabled or not")
		if isDMthin {
			Skip("This test is not applicable when DMThin is enabled")
		}
		//Check for metadata on selected node, if true then skip the test
		isMetadata, err := IsMetadataEnabled()
		log.FailOnError(err, "Failed to check if Metadata is present or not")
		if isMetadata {
			Skip("This test is not applicable when there is metadata")
		}
	})
	var (
		kvdbNodesIDsBeforePXStop, kvdbNodesIDsAfterPXStop []string
		kvdbDriverBeforePXStop, kvdbDriverAfterPXStop     string
		nodeForPXStop                                     node.Node
		contexts                                          = make([]*scheduler.Context, 0)
	)

	stepLog := "Stop PX on KVDB node and make sure the stopped node is no longer a KVDB member.Stopped kvdb node should be devoid of kvdb driver"
	It(stepLog, func() {

		stepLog = "Schedule application"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				uniqueChar := string('a' + rune(i%26))
				appName := fmt.Sprintf("kvdb-%d-%s", i, uniqueChar)
				contexts = append(contexts, ScheduleApplications(appName)...)
			}
		})

		stepLog := "Check kvdb node health before stopping PX"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			isHealthybeforePXStop, err := CheckKVDBNodesHealth()
			log.FailOnError(err, "Failed to check kvdb nodes health before PX stop")
			dash.VerifyFatal(isHealthybeforePXStop, true, "Is kvdb nodes are healthy before PX stop?")
		})

		stepLog = "Getting all kvdb node IDs before PX stop"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			kvdbNodesIDsBeforePXStop, err = GetAllkvdbNodeIDs()
			log.FailOnError(err, "Failed to get kvdb node IDs")
			log.InfoD("KVDB node IDs before stopping PX : [%v]", kvdbNodesIDsBeforePXStop)
		})

		//selecting node to stop and start PX
		index := rand.Intn(len(kvdbNodesIDsBeforePXStop))
		selectedNode := kvdbNodesIDsBeforePXStop[index]
		log.Infof("Node ID selected for PX to stop : [%v]", selectedNode)
		nodeForPXStop, err = node.GetNodeDetailsByNodeID(selectedNode)
		log.FailOnError(err, "Failed to get node details for node : [%v]", selectedNode)
		log.Infof("Node selected for PX to stop : [%v]", nodeForPXStop.Name)

		stepLog = "Getting kvdb driver on selected node before PX stop"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			kvdbDriverBeforePXStop, err = GetKvdbDriveOnNode(nodeForPXStop)
			log.FailOnError(err, "Failed to get kvdb driver on node : [%v]", nodeForPXStop.Name)
			log.InfoD("kvdb driver before PX stop : [%v]", kvdbDriverBeforePXStop)
		})

		stepLog = "Stopping PX on kvdb node, wait for 5 mins and start PX again on same node"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			log.Infof("Stop volume driver [%s] on node: [%s]", Inst().V.String(), nodeForPXStop)
			StopVolDriverAndWait([]node.Node{nodeForPXStop})

			//Wait till new kvdb memeber comes up
			err := WaitForKVDBMembers()
			log.FailOnError(err, "Failed waiting for KVDB members to be active")

			//Start the PX on the sane node where it was stopped
			log.Infof("Starting volume driver [%s] on node [%s]", Inst().V.String(), nodeForPXStop)
			StartVolDriverAndWait([]node.Node{nodeForPXStop})
		})

		stepLog = "Getting kvdb nodes IDs after PX stop"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			kvdbNodesIDsAfterPXStop, err = GetAllkvdbNodeIDs()
			log.FailOnError(err, "Failed to kvdb nodes after PX stop")
			log.InfoD("KVDB node IDs after stopping PX on KVDB node : [%v]", kvdbNodesIDsAfterPXStop)
		})

		stepLog = "Checking Health of kvdb nodes after PX stop"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			isHealthyAfterPXStop, err := CheckKVDBNodesHealth()
			log.FailOnError(err, "Failed to check of kvdb nodes after PX stop, err : [%v]", err)
			dash.VerifyFatal(isHealthyAfterPXStop, true, "Is kvdb nodes are healthy after PX stop ?")
		})

		stepLog = "Getting kvdb driver on selected node after PX stop"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			kvdbDriverAfterPXStop, err = GetKvdbDriveOnNode(nodeForPXStop)
			log.FailOnError(err, "Failed to get kvdb driver on node after PX stop: [%v]", nodeForPXStop.Name)
			log.InfoD("kvdb driver after PX stop : [%v]", kvdbDriverAfterPXStop)
		})

		stepLog = "Validation kvdb nodes after PX restart"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			kvdbNodeBeforePXStop, kvdbNodeAfterPXStop, err := FindReplacedKvdbNode(kvdbNodesIDsBeforePXStop, kvdbNodesIDsAfterPXStop)
			log.FailOnError(err, "Failed to find the new kvdb node after PX resatrt")
			//validating creation of new kvdb node after PX stop
			dash.VerifyFatal(kvdbNodeBeforePXStop != kvdbNodeAfterPXStop, true, fmt.Sprintf("Successfully created new kvdb node [%v] in place of [%v]", kvdbNodeAfterPXStop, kvdbNodeBeforePXStop))
			//Validating absence of kvdb driver after PX start
			dash.VerifyFatal(kvdbDriverBeforePXStop != "" && kvdbDriverAfterPXStop == "", true, fmt.Sprintf("Validating kvdb driver [%v] before PX stop and kvdb driver [%v] after PX stop ?", kvdbDriverBeforePXStop, kvdbDriverAfterPXStop))
		})

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})

})
