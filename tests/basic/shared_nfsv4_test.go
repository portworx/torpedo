package tests

import (
	"fmt"
	"time"

	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/sirupsen/logrus"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/portworx/torpedo/tests"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("{NFSv4NotSupported}", func() {
	var contexts []*scheduler.Context

	It("has to create sv4 svc volume with nfsv4, validate, and teardown apps", func() {
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("nfsv4notsupported-%d", i))...)
		}

		// validatePVCs will make sure they are in pending state.
		// sleep 30 seconds here, to make sure the pending state is not just transient
		time.Sleep(30 * time.Second)

		for _, ctx := range contexts {
			Step(fmt.Sprintf("validate: %s's pvcs", ctx.App.Key), func() {
				validatePVCs(ctx)
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

func validatePVCs(ctx *scheduler.Context) {
	pvcs, err := Inst().S.GetPVCs(ctx)
	if err != nil {
		logrus.Infof("get pvc error %v", err)
	}
	Expect(len(pvcs)).To(Equal(3), "There should be 3 PVCs")
	for _, pvc := range pvcs {
		Expect(pvc.Phase).To(Equal(string(corev1.ClaimPending)), fmt.Sprintf("pvc %v should be in pending phase", pvc.Name))
	}
}

var _ = Describe("{PodDeletedOnOldNFSServer}", func() {
	var contexts []*scheduler.Context

	It("has to setup, validate, failover, make sure pods on old server got deleted, and teardown apps", func() {
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("poddeletedonoldnfsserver-%d", i))...)
		}

		ValidateApplications(contexts)

		for _, ctx := range contexts {
			Step(fmt.Sprintf("validate: %s's pvcs", ctx.App.Key), func() {
				vols, err := Inst().S.GetVolumes(ctx)
				if err != nil {
					logrus.Infof("get volumes error %v", err)
				}
				vol := vols[0]
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
				allNodes := node.GetWorkerNodes()
				var nodesNotWithReplica []node.Node
				for _, node := range allNodes {
					logrus.Infof("testing node %v", node)
					if !contains(replicaNodes, node.Name) {
						nodesNotWithReplica = append(nodesNotWithReplica, node)
					}
				}
				logrus.Infof("testing nodesNotWithReplica %v", nodesNotWithReplica)
				for _, node := range nodesNotWithReplica {
					Inst().S.DisableSchedulingOnNode(node)
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
