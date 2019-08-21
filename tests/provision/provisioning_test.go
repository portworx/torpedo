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
//type vnode map[string]map[string][]string interface{}

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
		var volnodes map[string]map[string][]string 
		var vpsspec string

// loop	for multiple replica affinity test cases	
		Step("get nodes and set labels", func() {
			lbldata := getTestLabels(getCaseReqLabel) //TODO: function argument for getting testcase labels
			lblnode := SetNodeLabels(lbldata)
			logrus.Info("Nodes containing label", lblnode)
			Expect(lblnode).NotTo(BeEmpty())
			volnodes  = pvcNodeMap(getCaseReqPvcNode,lblnode)	//TODO: function argument for vol's expected nodes
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
				ValidateVpsRules(getCaseReqValidate,ctx, volnodes) //TODO function arg for validating mapping
			}

		})

		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

		cleanVps() //TODO: function arg for cleaning up the testcase environment
		for _, ctx := range contexts {
			TearDownContext(ctx, opts)
		}
	})
})






//Support functions

//#--enforcemnt: Required

func getCaseReqLabel() [] labelDict {

	lbldata := []labelDict{}
	node1lbl := labelDict{"media_type": "SSD","vps_test":"test"}  
	node2lbl := labelDict{"media_type": "SATA","vps_test":"test"}
	lbldata = append(lbldata, node1lbl, node2lbl)
		return lbldata
}

func getCaseReqPvcNode( lblnodes map[string][]string) map[string]map[string][]string {

	for key,val := range lblnodes {
		logrus.Infof("label node: key:%v Val:%v", key,val)
	}

	//Create 3 node lists (requiredNodes, prefNodes, notOnNodes)
	volnodelist := map[string]map[string][]string{}
	volnodelist["mysql-data"] = map[string][]string{}
	volnodelist["mysql-data-aggr"] = map[string][]string{}
	volnodelist ["mysql-data"]["pnodes"] = [] string{}
	volnodelist ["mysql-data"]["nnodes"] = [] string{}
	volnodelist ["mysql-data-aggr"]["pnodes"] = [] string{}
	volnodelist ["mysql-data-aggr"]["nnodes"] = [] string{}


	for _,lnode := range lblnodes["media_typeSSD"] {
		volnodelist ["mysql-data"]["rnodes"] = append(volnodelist ["mysql-data"]["rnodes"], lnode)
		volnodelist ["mysql-data-aggr"]["rnodes"] = append(volnodelist ["mysql-data-aggr"]["rnodes"], lnode)
	}

	for _,lnode := range lblnodes["media_typeSATA"] {
		volnodelist ["mysql-data"]["rnodes"] = append(volnodelist ["mysql-data"]["rnodes"], lnode)
		volnodelist ["mysql-data-aggr"]["rnodes"] = append(volnodelist ["mysql-data-aggr"]["rnodes"], lnode)
	}

	return volnodelist
}


/*
 * 1. Each rule template, will provide the expected output
 */

func getCaseReqValidate(appVolumes []*volume.Volume,volscheck map[string]map[string][]string) {

	logrus.Infof("Deployed volumes:%v,  volumes to check for nodes placement %v ",
			appVolumes, volscheck)

	for _, appvol := range appVolumes {

		for vol, vnodes := range volscheck {

			if appvol.Name == vol {
				replicas, err := Inst().V.GetReplicaSetNodes(appvol)
				logrus.Infof("==Replicas for vol: %s, appvol:%v Replicas:%v ", vol, appvol,replicas)
				Expect(err).NotTo(HaveOccurred())
				Expect(replicas).NotTo(BeEmpty())

				// Must have (required)
				for _,mnode := range vnodes["rnodes"]  {
					found := ""
					for _,rnode := range replicas {
						logrus.Infof("Expected Volume Node:%v Replica Node:%v", mnode, rnode)
						if mnode == rnode {
							found=rnode
							break	
						}
					}
					 if found == "" {
						logrus.Errorf("Volume '%v' does not have replica on node:'%v'", appvol,mnode)
					  	Expect(found).NotTo(BeEmpty())
					}
				}


				// Preferred
				for _,mnode := range vnodes["pnodes"]  {
					found := ""
					for _,rnode := range replicas {
						logrus.Infof("Preferred Volume Node:%v Replica Node:%v", mnode, rnode)
						if mnode == rnode {
							found=rnode
							break	
						}
					}
					 if found != ""  {
						logrus.Infof("Volume '%v' has replica on node:'%v'", appvol,mnode)
					}
				}

				// NotonNode
				for _,mnode := range vnodes["nnodes"]  {
					var found  string
					for _,rnode := range replicas {
						logrus.Infof("Volume should not have replica on :%v Replica Node:%v", mnode, rnode)
						if mnode == rnode {
							found = rnode
							break	
						}
					}
					 if found != ""  {
						logrus.Errorf("Volume '%v' has replica on node:'%v'", appvol,mnode)
					    Expect(found).To(BeEmpty())
					}
				}
			}
		}
	}
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

func getVpsSpec(rules [] labelDict ) string {

	var vpsspec string 
	logrus.Infof(" rules:%v ", rules)	
	vpsspec =`
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: ssd-sata-pool-placement-spread
spec:
  replicaAffinity:
  - affectedReplicas: 1
#    enforcement: required
    enforcement: preferred
#    type: affinity
    matchExpressions:
    - key: media_type
      operator: In
      values:
      - "SSD"
  replicaAffinity:
  - affectedReplicas: 1
    enforcement: preferred
#    enforcement: required
#    type: affinity
    matchExpressions:
    - key: media_type
      operator: In
      values:
      - "SATA"
	`
	return vpsspec
}


//-- Common function

//ValidateVpsRules check applied vps rules on the app
func ValidateVpsRules(f func([]*volume.Volume,map[string]map[string][]string), ctx *scheduler.Context,volscheck map[string]map[string][]string) {
	var err error
	var appVolumes []*volume.Volume
	appVolumes, err = Inst().S.GetVolumes(ctx)
	Expect(err).NotTo(HaveOccurred())
	Expect(appVolumes).NotTo(BeEmpty())

	f (appVolumes,volscheck)


}


func getTestLabels (f func ()  [] labelDict ) [] labelDict {
	return f()
}

//pvcNodeMap  The nodes on which this pvc is expected to have replica
func pvcNodeMap(f func(map[string][]string) map[string]map[string][]string , val map[string][]string ) map[string]map[string][]string  {

	return f(val)
}
 	
//SetNodeLabels set the provided labels on the portworx worker nodes
func SetNodeLabels(labels []labelDict) map[string] [] string  {

	lblnodes := map[string] [] string {}
	workerNodes := node.GetWorkerNodes()
	workerCnt := len(workerNodes)
	nodes2lbl := len(labels)

	if workerCnt < nodes2lbl {
		fmt.Printf("Required(%v) number of worker nodes(%v) not available", nodes2lbl, workerCnt)
		return lblnodes
	}

	// Get nodes
	for key, nlbl := range labels {
		//TODO: Randomize node selection
		n := workerNodes[key]
		for lkey, lval := range nlbl {
			if err := k8s.Instance().AddLabelOnNode(n.Name, lkey, lval.(string)); err != nil {
				logrus.Errorf("Failed to set node label %v: %v Err: %v", lkey, nlbl, err)
				return lblnodes
			}
			lblnodes[lkey+lval.(string)]=append(lblnodes[lkey+lval.(string)],n.Name)
		}

	}
	//TODO: Return node list with the labels attached
	return lblnodes
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
