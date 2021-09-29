package tests

import (
	"fmt"
	"time"

	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/portworx/torpedo/tests"
)

const (
	defaultCommandRetry   = 5 * time.Second
	defaultCommandTimeout = 1 * time.Minute
)

var _ = Describe("{NFSServerFailover}", func() {
	var contexts []*scheduler.Context
	logrus.Infof("testing 2")
	It("has to setup, validate, failover, make sure pods on old server got restarted, and teardown apps", func() {
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("nfsserverfailover-%d", i))...)
		}

		ValidateApplications(contexts)

		for _, ctx := range contexts {
			oneMinute := time.Second * time.Duration(60)
			var nodesNonReplica []node.Node
			var volume *volume.Volume
			Step("disable scheduling on non replica nodes", func() {
				vols, err := Inst().S.GetVolumes(ctx)
				if err != nil {
					logrus.Infof("get volumes error %v", err)
				}
				volume = vols[0]
				replicaSets, err := Inst().V.GetReplicaSets(volume)
				if err != nil {
					logrus.Infof("get replicaSets error %v", err)
				}
				var replicaNodes []string
				for _, replicaSet := range replicaSets {
					nodes := replicaSet.Nodes
					replicaNodes = append(replicaNodes, nodes...)
				}
				allNodes := node.GetWorkerNodes()
				for _, node := range allNodes {
					if !contains(replicaNodes, node.VolDriverNodeID) {
						nodesNonReplica = append(nodesNonReplica, node)
					}
				}
				for _, node := range nodesNonReplica {
					Inst().S.DisableSchedulingOnNode(node)
				}

			})

			// scale down and then scale up the app, so that pods are only scheduled on replica nodes
			Step(fmt.Sprintf("scale down app: %s to 0 ", ctx.App.Key), func() {
				applicationScaleUpMap, err := Inst().S.GetScaleFactorMap(ctx)
				Expect(err).NotTo(HaveOccurred())
				for name := range applicationScaleUpMap {
					applicationScaleUpMap[name] = int32(0)
				}
				err = Inst().S.ScaleApplication(ctx, applicationScaleUpMap)
				Expect(err).NotTo(HaveOccurred())
			})

			Step(fmt.Sprintf("scale up app: %s to 2, and re-enable scheduling on all nodes", ctx.App.Key), func() {
				applicationScaleUpMap, err := Inst().S.GetScaleFactorMap(ctx)
				Expect(err).NotTo(HaveOccurred())
				for name := range applicationScaleUpMap {
					applicationScaleUpMap[name] = int32(2)
				}
				err = Inst().S.ScaleApplication(ctx, applicationScaleUpMap)
				Expect(err).NotTo(HaveOccurred())
				ValidateApplications(contexts)
				for _, node := range nodesNonReplica {
					Inst().S.EnableSchedulingOnNode(node)
				}
			})

			Step("fail over nfs server, and make sure the pod on server gets restarted", func() {
				oldServer, err := Inst().V.GetNodeForVolume(volume, defaultCommandTimeout, defaultCommandRetry)
				Expect(err).NotTo(HaveOccurred())
				logrus.Infof("old nfs server %v [%v]", oldServer.SchedulerNodeName, oldServer.Addresses[0])
				pods, err := core.Instance().GetPodsUsingPV(volume.ID)
				Expect(err).NotTo(HaveOccurred())
				var oldPodOnOldServer corev1.Pod
				for _, pod := range pods {
					if pod.Spec.NodeName == oldServer.Name {
						oldPodOnOldServer = pod
					}
				}
				// make sure there is a pod running on the old nfs server
				Expect(&oldPodOnOldServer).NotTo(BeNil())
				logrus.Infof("pod on old server %v, creation time %v", oldPodOnOldServer.Name, oldPodOnOldServer.CreationTimestamp)

				timestampBeforeFailOver := time.Now()
				err = Inst().V.StopDriver([]node.Node{*oldServer}, false, nil)
				Expect(err).NotTo(HaveOccurred())
				err = Inst().V.WaitDriverDownOnNode(*oldServer)
				Expect(err).NotTo(HaveOccurred())
				logrus.Infof("stopped px on nfs server node %v [%v]", oldServer.SchedulerNodeName, oldServer.Addresses[0])

				var newServer *node.Node

				for i := 0; i < 10; i++ {
					server, err := Inst().V.GetNodeForVolume(volume, defaultCommandTimeout, defaultCommandRetry)
					Expect(err).NotTo(HaveOccurred())
					if server.Id != oldServer.Id {
						logrus.Infof("nfs server failed over, new nfs server is %s [%s]", server.SchedulerNodeName, server.Addresses[0])
						newServer = server
						break
					}
					time.Sleep(oneMinute)
				}
				// make sure nfs server failed over
				Expect(newServer).NotTo(BeNil())
				logrus.Infof("new nfs server is %v", newServer)

				logrus.Infof("start px on old nfs server Id %v, Name %v", oldServer.Id, oldServer.Name)
				Inst().V.StartDriver(*oldServer)
				err = Inst().V.WaitDriverUpOnNode(*oldServer, Inst().DriverStartTimeout)
				Expect(err).NotTo(HaveOccurred())
				logrus.Infof("px is up on old nfs server Id %v, Name %v", oldServer.Id, oldServer.Name)

				// make sure the pods on both old and new server are restarted
				pods, err = core.Instance().GetPodsUsingPV(volume.ID)
				Expect(err).NotTo(HaveOccurred())
				for _, pod := range pods {
					if pod.Spec.NodeName == oldServer.Name {
						logrus.Infof("pod on old server %v, creation time %v", oldPodOnOldServer.Name, oldPodOnOldServer.CreationTimestamp)
						logrus.Infof("After failover, pod on old server %v, creation time %v", pod.Name, pod.CreationTimestamp)
						Expect(pod.CreationTimestamp.After(timestampBeforeFailOver)).To(BeTrue())
					}
					if pod.Spec.NodeName == newServer.Name {
						logrus.Infof("After failover, pod on new server %v, creation time %v", pod.Name, pod.CreationTimestamp)
						Expect(pod.CreationTimestamp.After(timestampBeforeFailOver)).To(BeTrue())
					}
				}
			})
		}

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

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
