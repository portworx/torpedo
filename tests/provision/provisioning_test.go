package tests

import (
	"fmt"
	"testing"
	//	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/portworx/sched-ops/k8s"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/volume"
	. "github.com/portworx/torpedo/tests"
	"github.com/sirupsen/logrus"
)

type labelDict map[string]interface{}

func TestVps(t *testing.T) {
	RegisterFailHandler(Fail)

	var specReporters []Reporter
	junitReporter := reporters.NewJUnitReporter("/testresults/junit_basic.xml")
	specReporters = append(specReporters, junitReporter)
	RunSpecsWithDefaultAndCustomReporters(t, "Torpedo : Provisioning", specReporters)
}

var _ = BeforeSuite(func() {
	InitInstance()
})

// This test performs VolumePlacementStrategy's replica affinity  of application
// volume
var _ = Describe("{ReplicaAffinity}", func() {
	It("has to schedule app and verify the replication affinity", func() {

		var contexts []*scheduler.Context
		lbldata := []labelDict{}
		Step("get nodes and set labels", func() {
			// Label list
			node1lbl := labelDict{"media_type": "SSD"} //,
			node2lbl := labelDict{"media_type": "SATA"}
			lbldata = append(lbldata, node1lbl, node2lbl)
			status := SetNodeLabels(lbldata)
			Expect(status).To(BeEmpty())
		})
		volrules := []labelDict{}
		Step("Rules pf volume Placement", func() {

			rule1 := labelDict{"vpsname": "ssd-sata-pool-placement-spread",
				"label": "media_type:SSD", "affectedrepl": "1", "enforcement": "required"}
			rule2 := labelDict{"vpsname": "ssd-sata-pool-placement-spread",
				"label": "media_type:SSD", "affectedrepl": "1", "enforcement": "required"}
			volrules := append(volrules, rule1, rule2)

			logrus.Info("Rules to check per volume", volrules)
		})

		Step("Launch Application ", func() {
			for i := 0; i < Inst().ScaleFactor; i++ {
				contexts = append(contexts, ScheduleAndValidate(fmt.Sprintf("replicaaffinity-%d", i))...)
			}
		})

		Step("get volumes and replica affinity", func() {
			// Get volumes  Inst().S.GetVolumes(ctx)
			// Get Inst().V.GetReplicaSetNodes(vol)
			// Verify & Confirm replica placement

			for _, ctx := range contexts {
				ValidateVpsRules(ctx, lbldata)
			}

		})

		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

		for _, ctx := range contexts {
			TearDownContext(ctx, opts)
		}
	})
})

/*
 * To ways to Validate
 * 1. Each rule template, will provide the expected output
 * 2. Parse each rule, interpret and generate the expected output
 *
 */

//Support functions
//ValidateVpsRules check applied vps rules on the app
func ValidateVpsRules(ctx *scheduler.Context, labels []labelDict) {
	// Get Volumes
	// Get Replicas
	// Get Rules applied on the app
	// Get node labels
	// Verify rules
	//
	var rules []string
	var err error
	var appVolumes []*volume.Volume
	appVolumes, err = Inst().S.GetVolumes(ctx)
	Expect(err).NotTo(HaveOccurred())
	Expect(appVolumes).NotTo(BeEmpty())
	for _, vol := range appVolumes {
		replicas, err := Inst().V.GetReplicaSetNodes(vol)
		logrus.Infof("==Replicas for vol: %s, Replicas:%v ", vol, replicas)
		Expect(err).NotTo(HaveOccurred())
		Expect(replicas).NotTo(BeEmpty())

		// create list of vol-replica-node mapping
	}

	//TODO: refer to the spawn code for template: instance/instancelist.go
	//		instance/ansible.go
	//For each rule, verify the volume,replica placement
	for _, rule := range rules {
		logrus.Info("Rule:", rule)
		//ValidateVpsReplicaRule()
		//ValidateVpsVolumeRule()

	}
}

//SetNodeLabels set the provided labels on the portworx worker nodes
func SetNodeLabels(labels []labelDict) int {

	workerNodes := node.GetWorkerNodes()
	workerCnt := len(workerNodes)
	nodes2lbl := len(labels)

	if workerCnt < nodes2lbl {
		fmt.Printf("Required(%v) number of worker nodes(%v) not available", nodes2lbl, workerCnt)
		return 0
	}

	// Get nodes
	for key, nlbl := range labels {
		//TODO: Randomize node selection
		n := workerNodes[key]
		for lkey, lval := range nlbl {
			if err := k8s.Instance().AddLabelOnNode(n.Name, lkey, lval.(string)); err != nil {
				logrus.Errorf("Failed to set node label %v: %v Err: %v", lkey, nlbl, err)
				return 0
			}

		}

	}
	//TODO: Return node list with the labels attached
	return 1
}

var _ = AfterSuite(func() {
	PerformSystemCheck()
	ValidateCleanup()
})

func init() {
	ParseFlags()
}
