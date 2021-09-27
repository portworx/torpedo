package tests

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/sirupsen/logrus"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/portworx/torpedo/tests"
)

const (
	defaultCommandRetry   = 5 * time.Second
	defaultCommandTimeout = 1 * time.Minute
)

var _ = Describe("{PodDeletedOnOldNFSServer}", func() {
	var contexts []*scheduler.Context

	It("has to setup, validate, failover, make sure pods on old server got deleted, and teardown apps", func() {
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("poddeletedonoldnfsserver-%d", i))...)
		}

		ValidateApplications(contexts)

		for _, ctx := range contexts {
			sleepTime := time.Second * time.Duration(60)
			var nodesNonReplica []node.Node
			Step(fmt.Sprintf("validate: %s's pvcs", ctx.App.Key), func() {
				vols, err := Inst().S.GetVolumes(ctx)
				logrus.Infof("testing vols %v", vols)
				if err != nil {
					logrus.Infof("get volumes error %v", err)
				}
				vol := vols[0]

				pods, err := core.Instance().GetPodsUsingPV(vol.ID)
				logrus.Infof("testing pod name %v %v", pods[0].Name, pods[1].Name)

				replicaSets, err := Inst().V.GetReplicaSets(vol)
				if err != nil {
					logrus.Infof("get replicaSets error %v", err)
				}
				var replicaNodes []string
				for _, replicaSet := range replicaSets {
					nodes := replicaSet.Nodes
					logrus.Infof("testing nodes %v, replicatSet %v", nodes, replicaSet)
					replicaNodes = append(replicaNodes, nodes...)
				}
				logrus.Infof("testing replicaNodes %v", replicaNodes)
				allNodes := node.GetWorkerNodes()
				for _, node := range allNodes {
					logrus.Infof("testing node %v", node)
					logrus.Infof("testing node name %v, VolDriverNodeID %v", node.Name, node.VolDriverNodeID)
					if !contains(replicaNodes, node.VolDriverNodeID) {
						nodesNonReplica = append(nodesNonReplica, node)
					}
				}
				logrus.Infof("testing nodesNonReplica %v", nodesNonReplica)
				for _, node := range nodesNonReplica {
					Inst().S.DisableSchedulingOnNode(node)
				}

			})

			Step(fmt.Sprintf("scale down app: %s to 0 ", ctx.App.Key), func() {
				applicationScaleUpMap, err := Inst().S.GetScaleFactorMap(ctx)
				Expect(err).NotTo(HaveOccurred())
				for name, _ := range applicationScaleUpMap {
					applicationScaleUpMap[name] = int32(0)
				}
				err = Inst().S.ScaleApplication(ctx, applicationScaleUpMap)
				Expect(err).NotTo(HaveOccurred())
				logrus.Infof("testing scale app to 0")
			})

			Step(fmt.Sprintf("scale up app: %s to 2, and re-enable scheduling on all nodes", ctx.App.Key), func() {
				applicationScaleUpMap, err := Inst().S.GetScaleFactorMap(ctx)
				Expect(err).NotTo(HaveOccurred())
				for name, _ := range applicationScaleUpMap {
					applicationScaleUpMap[name] = int32(2)
				}
				err = Inst().S.ScaleApplication(ctx, applicationScaleUpMap)
				Expect(err).NotTo(HaveOccurred())
				logrus.Infof("testing scale app to 2")

				vols, _ := Inst().S.GetVolumes(ctx)
				vol := vols[0]
				pods, _ := core.Instance().GetPodsUsingPV(vol.ID)

				logrus.Infof("testing pod name %v %v, nodeName %v %v", pods[0].Name, pods[1].Name, pods[0].Spec.NodeName, pods[1].Spec.NodeName)
				b, err := json.Marshal(pods[0])
				logrus.Infof("testing pod json %v, err %v", string(b), err)
				ValidateApplications(contexts)
				for _, node := range nodesNonReplica {
					Inst().S.EnableSchedulingOnNode(node)
				}
			})

			Step("stop px on nfs server", func() {
				vols, err := Inst().S.GetVolumes(ctx)
				Expect(err).NotTo(HaveOccurred())
				for _, vol := range vols {
					logrus.Infof("testing going to get oldServer")
					oldServer, err := Inst().V.GetNodeForVolume(vol, defaultCommandTimeout, defaultCommandRetry)
					logrus.Infof("testing got oldServer")
					Expect(err).NotTo(HaveOccurred())
					logrus.Infof("testing volume %s is attached on node %s [%s]", vol.ID, oldServer.SchedulerNodeName, oldServer.Addresses[0])
					time.Sleep(sleepTime * 3)
					err = Inst().V.StopDriver([]node.Node{*oldServer}, false, nil)
					Expect(err).NotTo(HaveOccurred())
					err = Inst().V.WaitDriverDownOnNode(*oldServer)
					Expect(err).NotTo(HaveOccurred())
					logrus.Infof("testing stop px on nfs server Id %v", oldServer.Id)
					time.Sleep(5 * sleepTime)
					logrus.Infof("testing waited 5 mins after px stop on old server")
					var newServer *node.Node
					for i := 0; i < 3; i++ {
						newServer, err = Inst().V.GetNodeForVolume(vol, defaultCommandTimeout, defaultCommandRetry)
						Expect(err).NotTo(HaveOccurred())
						logrus.Infof("testing new nfs server Id %v", newServer.Id)
						logrus.Infof("testing volume %s is attached on new nfs server %s [%s]", vol.ID, newServer.SchedulerNodeName, newServer.Addresses[0])
						if newServer.Id != oldServer.Id {
							logrus.Infof("testing nfs server failover, new nfs server%s [%s]", newServer.SchedulerNodeName, newServer.Addresses[0])
							break
						}
						time.Sleep(sleepTime)
					}
					logrus.Infof("testing start px on old nfs server Id %v, Name %v", oldServer.Id, oldServer.Name)
					Inst().V.StartDriver(*oldServer)
					err = Inst().V.WaitDriverUpOnNode(*oldServer, Inst().DriverStartTimeout)
					Expect(err).NotTo(HaveOccurred())
					logrus.Infof("testing px is up on old nfs server Id %v, Name %v", oldServer.Id, oldServer.Name)
					time.Sleep(sleepTime * 2)
					logrus.Infof("testing vol %v", vol)
					pods, err := core.Instance().GetPodsUsingPV(vol.ID)
					logrus.Infof("testing pod name %v %v", pods[0].Name, pods[1].Name)
					logrus.Infof("testing pods %v, creation time %v, err %v", pods[0], pods[0].CreationTimestamp, err)
				}
			})
			logrus.Infof("testing finish all steps")

			time.Sleep(sleepTime * 5)

			// find out the nfs server, and stop driver on it (how to find out nfs server node)
			// Inst().V.StopDriver([]node.Node{*n}, false, nil)
			// wait for nfs server to failover and start driver on it (how do we know it has been fail over, just sleep?)
			// Inst().V.StartDriver()
			// make sure pod on nfs server got deleted (how do we get pod info to find out whether the pod is deleted)
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
