package tests

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/libopenstorage/openstorage/pkg/mount"
	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/testrailuttils"
	"github.com/sirupsen/logrus"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/portworx/torpedo/tests"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	devicePathPrefix = "/dev/pxd/pxd"
)

var _ = Describe("{Sharedv4Functional}", func() {
	var testrailID, runID int
	var contexts, testSv4Contexts, testSharedV4Contexts []*scheduler.Context
	var workers []node.Node
	var numPods int
	var namespacePrefix string

	JustBeforeEach(func() {
		runID = testrailuttils.AddRunsToMilestone(testrailID)

		// Set up all apps
		contexts = nil
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("%s-%d", namespacePrefix, i))...)
		}

		// Skip the test if there are no test-sharedv4 apps
		testSharedV4Contexts = getTestSharedV4Contexts(contexts)
		if len(testSv4Contexts) == 0 && len(testSharedV4Contexts) == 0 {
			Skip("No test-sharedv4 apps were found")
		}
		workers = node.GetWorkerNodes()
		numPods = len(workers)

		Step("scale the test-sharedv4 apps so that one pod runs on each worker node", func() {
			scaleApps(testSharedV4Contexts, numPods)
		})
		ValidateApplications(contexts)
	})

	Context("{SharedV4MountRecovery}", func() {
		BeforeEach(func() {
			namespacePrefix = "sharedv4mountrecovery"
		})

		JustBeforeEach(func() {
			for _, ctx := range testSharedV4Contexts {
				vol, _, attachedNode := getSv4TestAppVol(ctx)
				k8sApps := apps.Instance()
				Step(
					fmt.Sprintf("setup app %s to deploy on single node", ctx.App.Key),
					func() {
						Step(fmt.Sprintf("add label to attached node for %s", ctx.App.Key), func() {
							err := core.Instance().AddLabelOnNode(attachedNode.Name, "attachedNode", "true")
							Expect(err).NotTo(HaveOccurred())
						})

						Step(fmt.Sprintf("update %s deployment with label", ctx.App.Key), func() {
							deployments, err := k8sApps.ListDeployments(ctx.GetID(), metav1.ListOptions{})
							Expect(err).NotTo(HaveOccurred())
							deployment := deployments.Items[0]
							deployment.Spec.Template.Spec.NodeSelector = map[string]string{"attachedNode": "true"}
							deployment.Spec.Template.Spec.Affinity = nil

							_, err = k8sApps.UpdateDeployment(&deployment)
							Expect(err).NotTo(HaveOccurred())
						})

						Step(fmt.Sprintf("scale down app: %s to 0 ", ctx.App.Key), func() {
							scaleApp(ctx, 0)
						})

						// wait until all pods are gone
						Step(fmt.Sprintf("wait for app %s to have 0 pods", ctx.App.Key), func() {
							waitForNumPodsToEqual(ctx, 0)
						})

						Step(fmt.Sprintf("scale the app back to numPods for %s", ctx.App.Key), func() {
							scaleApp(ctx, numPods)

							ValidateContext(ctx)
						})

						// validate the pods are all on one node
						Step(fmt.Sprintf("validate that all pods are running on the attached node for %s", ctx.App.Key), func() {
							pods, err := core.Instance().GetPodsUsingPV(vol.ID)
							Expect(err).NotTo(HaveOccurred())
							for _, pod := range pods {
								Expect(pod.Spec.NodeName).To(Equal(attachedNode.Name))
							}

						})
					})
			}

		})

		It("should set device path to RO and validate recovery", func() {
			for _, ctx := range testSharedV4Contexts {
				_, apiVol, attachedNode := getSv4TestAppVol(ctx)
				counterCollectionInterval := 3 * time.Duration(numPods) * time.Second
				devicePath := fmt.Sprintf("%s%s", devicePathPrefix, apiVol.Id)

				// In porx, we check the export path is read-only when it is previously RO and
				// changed to RW by fs due to occuring error. Specifically, when mnt.Opts (default)
				// is RW and mnt.VfsOpts (fs state) is RO.
				// By setting device path to RO, we mock the same behavior. Check state in
				// `/proc/self/mountinfo`
				Step(fmt.Sprintf("mark device path as RO %s", ctx.App.Key), func() {
					setPathToROMode(devicePath, attachedNode)
				})

				Step(fmt.Sprintf("validate the counters are inactive %s", ctx.App.Key), func() {
					counters := getAppCounters(apiVol, attachedNode, counterCollectionInterval)
					activePods := getActivePods(counters)
					Expect(len(activePods)).To(Equal(0))
				})

				Step(fmt.Sprintf("restart vol driver %s", ctx.App.Key), func() {
					restartVolumeDriverOnNode(attachedNode)
				})

				Step(fmt.Sprintf("validate counter are active for %s", ctx.App.Key), func() {
					counters := getAppCounters(apiVol, attachedNode, counterCollectionInterval)
					activePods := getActivePods(counters)
					Expect(len(activePods)).To(Equal(numPods))
				})

				Step(fmt.Sprintf("validate device path is set as RW for %s", ctx.App.Key), func() {
					mntList, err := mount.GetMounts()
					Expect(err).NotTo(HaveOccurred())
					var volumeMountRW = regexp.MustCompile(`,rw,|,rw|rw,|rw`)

					for _, mnt := range mntList {
						if mnt.Mountpoint != devicePath {
							continue
						}
						Expect(volumeMountRW.MatchString(mnt.Opts)).To(BeTrue())
						Expect(volumeMountRW.MatchString(mnt.VfsOpts)).To(BeTrue())
					}
				})
			}
		})

		JustAfterEach(func() {
			for _, ctx := range testSharedV4Contexts {
				_, _, attachedNode := getSv4TestAppVol(ctx)
				Step(fmt.Sprintf("remove label on node for %s", ctx.App.Key), func() {
					err := core.Instance().RemoveLabelOnNode(attachedNode.Name, "attachedNode")
					Expect(err).NotTo(HaveOccurred())
				})
			}
		})
	})

	// Template for additional tests
	// Context("{}", func() {
	// 	BeforeEach(func() {
	// 		testrailID = 0
	//      namespacePrefix = ""
	// 	})
	//
	// 	JustBeforeEach(func() {
	//		// since the apps are deployed by JustBeforeEach in the outer block,
	// 		// any test-specific changes to the deployed apps should go here.
	// 	})
	//
	// 	It("", func() {
	// 		for _, ctx := range testSharedV4Contexts {
	// 		}
	// 	})

	// 	AfterEach(func() {
	// 	})
	// })

	AfterEach(func() {
		Step("destroy apps", func() {
			if CurrentGinkgoTestDescription().Failed {
				logrus.Info("not destroying apps because the test failed\n")
				return
			}
			for _, ctx := range contexts {
				TearDownContext(ctx, map[string]bool{scheduler.OptionsWaitForResourceLeakCleanup: true})
			}
		})
	})

	JustAfterEach(func() {
		AfterEachTest(contexts, testrailID, runID)
	})
})

// returns the contexts that are running test-sharedv4* apps
func getTestSharedV4Contexts(contexts []*scheduler.Context) []*scheduler.Context {
	var testSharedV4Contexts []*scheduler.Context
	for _, ctx := range contexts {
		if !strings.HasPrefix(ctx.App.Key, "test-sharedv4") {
			continue
		}
		testSharedV4Contexts = append(testSharedV4Contexts, ctx)
	}
	return testSharedV4Contexts
}
