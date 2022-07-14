package tests

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/testrailuttils"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/portworx/torpedo/tests"
)

var _ = Describe("{VirtLauncherStatfs}", func() {
	var testrailID, runID int
	var contexts, prunedContexts []*scheduler.Context
	var namespacePrefix string

	JustBeforeEach(func() {
		runID = testrailuttils.AddRunsToMilestone(testrailID) // TODO testrailID

		// Set up all apps
		contexts = nil
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("%s-%d", namespacePrefix, i))...)
		}

		// Skip the test if there are no virt-launcher-sim apps
		var foundBothTypes bool
		prunedContexts, foundBothTypes = pruneContexts(contexts)
		if !foundBothTypes {
			Skip("Did not find both virt launcher and test sharedv4 apps")
		}
		Step("scale the apps so that one pod runs on each worker node", func() {
			scaleApps(prunedContexts, len(node.GetWorkerNodes()))
		})
		ValidateApplications(contexts)
	})

	When("{statfs is invoked inside the virt-launcher pod}", func() {
		BeforeEach(func() {
			testrailID = 0 // TODO
			namespacePrefix = "virtlauncherstatfs"
		})

		It("should return nfs only for PX volumes", func() {
			logrus.Infof("validating %d pruned contexts", len(prunedContexts))
			for _, ctx := range prunedContexts {
				logrus.Infof("validating context %v", ctx.App.Key)
				validateStatfs(ctx)
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
	// 		for _, ctx := range virtLauncherContexts {
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

// Returns the contexts that are running either virt-launcher-sim apps or test-sharedv4 apps.
// The second return value is true if at least one context of each type was found.
func pruneContexts(contexts []*scheduler.Context) ([]*scheduler.Context, bool) {
	var foundVirtLauncher, foundTestSharedv4 bool
	var ret []*scheduler.Context
	for _, ctx := range contexts {
		if isVirtLauncherContext(ctx) {
			foundVirtLauncher = true
			ret = append(ret, ctx)
		} else if isTestSharedv4Context(ctx) {
			foundTestSharedv4 = true
			ret = append(ret, ctx)
		}
	}
	return ret, foundVirtLauncher && foundTestSharedv4
}

func isVirtLauncherContext(ctx *scheduler.Context) bool {
	return strings.HasPrefix(ctx.App.Key, "virt-launcher-sim")
}

func isTestSharedv4Context(ctx *scheduler.Context) bool {
	return strings.HasPrefix(ctx.App.Key, "test-sv4-svc") || strings.HasPrefix(ctx.App.Key, "test-sharedv4")
}

func validateStatfs(ctx *scheduler.Context) {
	// mount paths inside the app container
	sharedMountPath := "/shared-vol"
	localMountPath := "/local-vol"

	vols, err := Inst().S.GetVolumes(ctx)
	Expect(err).NotTo(HaveOccurred(), "failed to get volumes for context %s: %v", ctx.App.Key, err)
	Expect(len(vols)).To(Equal(1))
	vol := vols[0]

	pods, err := core.Instance().GetPodsUsingPV(vol.ID)
	Expect(err).NotTo(HaveOccurred(), "failed to get pods for volume %v of context %s: %v", vol.ID, ctx.App.Key, err)

	foundBindMount := false
	foundNFSMount := false
	for _, pod := range pods {
		logrus.Infof("validating statfs in pod %s in namespace %v", pod.Name, pod.Namespace)

		// Sharedv4 PX volume is mounted on path /sv4test. Check if it is nfs-mounted or bind-mounted.
		//
		// Bind mount sample output:
		//
		// $ kubectl exec -it virt-launcher-sim-dep-8476fffd8b-ds4l5 -c sv4test -- df -T /sv4test
		// Filesystem                     Type 1K-blocks  Used Available Use% Mounted on
		// /dev/pxd/pxd585819943023766088 ext4  51343840 54556  48651460   1% /sv4test
		//
		//
		// NFS mount sample output:
		//
		// $ kubetl exec -it virt-launcher-sim-dep-8476fffd8b-f7vbk -c sv4test -- df -T /sv4test
		// Filesystem                                          Type 1K-blocks  Used Available Use% Mounted on
		// 192.168.121.49:/var/lib/osd/pxns/585819943023766088 nfs   51344384 54272  48652288   1% /sv4test

		output := runCommandInContainer(&pod, []string{"df", "-T", sharedMountPath})
		isBindMount := regexp.MustCompile(`pxd.*ext4`).MatchString(output)
		if !isBindMount {
			// must be an NFS mount
			Expect(regexp.MustCompile(`pxns.*nfs`).MatchString(output)).To(BeTrue())
		}
		logrus.Infof("isBindMount=%v", isBindMount)

		// check statfs() on the PX volume
		output = runCommandInContainer(&pod, []string{"stat", "--format=%T", "-f", sharedMountPath})
		output = strings.TrimSpace(output)
		logrus.Infof("statfs(%v)=%v", sharedMountPath, output)

		if isVirtLauncherContext(ctx) {
			// should always be "nfs"
			Expect(output).To(Equal("nfs"),
				"statfs() did not return 'nfs' for virt launcher pod %s/%s", pod.Namespace, pod.Name)
		} else if isBindMount {
			Expect(output).NotTo(Equal("nfs"),
				"statfs() returned 'nfs' for bind-mounted pod %s/%s", pod.Namespace, pod.Name)
		} else {
			Expect(output).To(Equal("nfs"),
				"statfs() did not return 'nfs' for nfs-mounted pod %s/%s", pod.Namespace, pod.Name)
		}

		// Path /local-vol is an {emptydir} volume so it should never return nfs for the file system type
		output = runCommandInContainer(&pod, []string{"stat", "--format=%T", "-f", localMountPath})
		output = strings.TrimSpace(output)
		logrus.Infof("statfs(%v)=%v", localMountPath, output)
		Expect(output).NotTo(Equal("nfs"),
			"statfs() returned 'nfs' for local volume in pod %s/%s", pod.Namespace, pod.Name)

		if isBindMount {
			foundBindMount = true
		} else {
			foundNFSMount = true
		}
	}

	// Sanity check: since we run 1 pod on each of the worker nodes, we expect to find
	// 1 bind-mounted pod and at least 1 nfs-mounted pod.
	Expect(foundBindMount).To(BeTrue(), "bind-mounted pod not found for context %s", ctx.App.Key)
	Expect(foundNFSMount).To(BeTrue(), "nfs-mounted pod not found for context %s", ctx.App.Key)
	logrus.Infof("validated statfs for context %v", ctx.App.Key)
}

func runCommandInContainer(pod *corev1.Pod, cmd []string) string {
	container := "sv4test"
	output, err := core.Instance().RunCommandInPod(cmd, pod.Name, container, pod.Namespace)
	Expect(err).NotTo(HaveOccurred(),
		"failed to run command %v inside the pod %v/%v: %v", cmd, pod.Namespace, pod.Name, err)
	return output
}
