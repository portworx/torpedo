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
		var nodelist []node.Node
		var volcheck [] string
		var vpsspec string

// loop	for multiple replica affinity test cases	
		Step("get nodes and set labels", func() {
			lbldata := getTestLabels()
			lblnode := SetNodeLabels(lbldata)
			logrus.Info("Nodes containing label", lblnode)
			Expect(lblnode).NotTo(BeEmpty())
			volcheck,nodelist = pvcNodeMap(lblnode)	
		})


		Step("Rules of volume Placement", func () {
			volrules := getTestRules()
			logrus.Info("Rules to check per volume", volrules)
			vpsspec = getVpsSpec(volrules)
			applyVpsSpec(vpsspec)
		})

		Step("Launch Application ", func() {
			for i := 0; i < Inst().ScaleFactor; i++ {
				contexts = append(contexts, ScheduleAndValidate(fmt.Sprintf("replicaaffinity-%d", i))...)
			}
		})

		Step("Validate volumes and replica affinity", func() {

			for _, ctx := range contexts {
				ValidateVpsRules(ctx, volcheck, nodelist)
			}

		})

		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

		//TODO:
		//Clean labels set on node
		//Clean  Vps  kubectl delete vps ssd-sata-pool-placement-spread
		cleanVps()
		for _, ctx := range contexts {
			TearDownContext(ctx, opts)
		}
	})
})






//Support functions

func getTestLabels () [] labelDict {
		lbldata := []labelDict{}
		node1lbl := labelDict{"media_type": "SSD"} //, 
		node2lbl := labelDict{"media_type": "SATA"}
		lbldata = append(lbldata, node1lbl, node2lbl)
		return lbldata
}

func getTestRules () [] labelDict{

		volrules := []labelDict{}
		rule1 := labelDict {"vpsname":"ssd-sata-pool-placement-spread",
		"label":"media_type:SSD", "affectedrepl":"1", "enforcement":"required"  }
		rule2 := labelDict { "vpsname":"ssd-sata-pool-placement-spread", 
		"label":"media_type:SSD", "affectedrepl":"1", "enforcement":"required"  }
		volrules = append(volrules, rule1,rule2)
		return volrules 
}


//pvcNodeMap  The nodes on which this pvc is  expected to have replica
func pvcNodeMap(labelnodes []node.Node ) ([]string, []node.Node) {

	var volcheck [] string


	for key,val := range labelnodes {
		logrus.Infof("label node: key:%v Val:%v", key,val)
		logrus.Infof(" node details: node name Val:%v", val.Name)
	}
	volcheck=append(volcheck, "mysql-data")
	volcheck=append(volcheck, "mysql-data-seq")

	//Create 3 node lists (mustNodes, preferedNodes, notOnNodes)

	return volcheck, labelnodes
}
 	


/*
 * To ways to Validate
 * 1. Each rule template, will provide the expected output
 * 2. Parse each rule, interpret and generate the expected output
 *
 */

//ValidateVpsRules check applied vps rules on the app
func ValidateVpsRules(ctx *scheduler.Context,volscheck [] string, mustnodes []node.Node) {
	// Get Volumes
	// Get Replicas
	// Get Rules applied on the app
	// Get node labels
	// Verify rules
	//
	var err error
	var appVolumes []*volume.Volume
	appVolumes, err = Inst().S.GetVolumes(ctx)
	Expect(err).NotTo(HaveOccurred())
	Expect(appVolumes).NotTo(BeEmpty())
	logrus.Infof("Deployed volumes:%v,  volumes to check for %v and nodes on which the volumes should be %v", appVolumes, volscheck, mustnodes)
	for _, appvol := range appVolumes {
		for _,vol := range volscheck {
			if appvol.Name == vol {
				replicas, err := Inst().V.GetReplicaSetNodes(appvol)
				logrus.Infof("==Replicas for vol: %s, appvol:%v Replicas:%v ", vol, appvol,replicas)
				Expect(err).NotTo(HaveOccurred())
				Expect(replicas).NotTo(BeEmpty())

				// Must have (required)
				for _,mnode := range mustnodes {
					found := 0
					for _,rnode := range replicas {
						logrus.Infof("mnode:%v rnode:%v", mnode.Name, rnode)
						if mnode.Name == rnode {
							found=1
							break	
						}
					}
					 if found == 0 {
						logrus.Errorf("Volume '%v' does not have replica on node:'%v'", appvol,mnode)
					}
					  // Expect(found).NotTo(BeEmpty())
				}


				// Preferred
				//

				// NotonNode
			}
		}
	}
}

//SetNodeLabels set the provided labels on the portworx worker nodes
func SetNodeLabels(labels []labelDict) []node.Node {

appNodes  := make([]node.Node,0)
	workerNodes := node.GetWorkerNodes()
	workerCnt := len(workerNodes)
	nodes2lbl := len(labels)

	if workerCnt < nodes2lbl {
		fmt.Printf("Required(%v) number of worker nodes(%v) not available", nodes2lbl, workerCnt)
		return appNodes
	}

	// Get nodes
	for key, nlbl := range labels {
		//TODO: Randomize node selection
		n := workerNodes[key]
		for lkey, lval := range nlbl {
			if err := k8s.Instance().AddLabelOnNode(n.Name, lkey, lval.(string)); err != nil {
				logrus.Errorf("Failed to set node label %v: %v Err: %v", lkey, nlbl, err)
				return appNodes
			}
			appNodes = append(appNodes, n)
		}

	}
	//TODO: Return node list with the labels attached
	return appNodes
}


func getVpsSpec(rules [] labelDict ) string {

	var vpsspec string 
	logrus.Infof(" rules:%v ", rules)	
	vpsspec =`
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: ssd-sata-pool-placement-spread
spec:
	`
	return vpsspec
}

func applyVpsSpec (vpsspec string) error {
	var err error
	logrus.Infof("vpsspec:%v", vpsspec)

	return err
}

func cleanVps() {
	logrus.Infof("Cleanup test case context")
}

var _ = AfterSuite(func() {
	PerformSystemCheck()
	ValidateCleanup()
})

func init() {
	ParseFlags()
}
