package tests

import (
	"fmt"
	"testing"
	"os"
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

		var vpsspec string
		vpsrules := GetVpsRules()

		for vkey,vrule := range vpsrules {
			var contexts []*scheduler.Context
			var volnodes map[string]map[string][]string 
	
			var lbldata  [] labelDict 
			Step("get nodes and set labels: "+vkey, func() {
				lbldata = getTestLabels(vrule.GetLabels) //TODO: function argument for getting testcase labels
				RemoveNodeLabels(lbldata)
				lblnode := SetNodeLabels(lbldata)
				logrus.Info("Nodes containing label", lblnode)
				Expect(lblnode).NotTo(BeEmpty())
				volnodes  = pvcNodeMap(vrule.GetPvcNodeLabels,lblnode)	//TODO: function argument for vol's expected nodes
			})

			Step("Rules of volume Placement: "+vkey, func () {
				//volrules := getTestRules()
				//logrus.Info("Rules to check per volume", volrules)
				vpsspec = getVpsSpec(vrule.GetSpec)
			})

			Step("==============Launch Application============== :"+vkey, func() {
				applyVpsSpec(vpsspec)
				logrus.Infof("=========== Spec Dir %v", Inst().SpecDir)
				Inst().S.RescanSpecs(Inst().SpecDir)
				for i := 0; i < Inst().ScaleFactor; i++ {
					contexts = append(contexts, ScheduleAndValidate(fmt.Sprintf("replicaaffinity-%d", i))...)
				}
			})

			Step("Validate volumes and replica affinity: "+vkey, func() {
				for _, ctx := range contexts {
					ValidateVpsRules(vrule.Validate,ctx, volnodes) //TODO function arg for validating mapping
				}

			})

			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

			vrule.CleanVps() //TODO: function arg for cleaning up the testcase environment
			//Remove labes from all nodes
			RemoveNodeLabels(lbldata)

			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		}
	})
})






//Support functions






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

// RemoveNodeLabels  remove the case specific lables from all nodes
func RemoveNodeLabels(labels []labelDict)  {

	workerNodes := node.GetWorkerNodes()

	 // Get nodes
	 for _,n := range workerNodes {
		 for _, nlbl := range labels {
			 for lkey, lval := range nlbl {
				 if err := k8s.Instance().RemoveLabelOnNode(n.Name, lkey); err != nil {
					 logrus.Errorf("Failed to remove node label %v=%v: %v Err: %v", lkey,lval , nlbl, err)
						 //return lblnodes
				 }
			 }

		 }
	 }
}



func getVpsSpec(f func () string  ) string  {
 	return f ()
}

func applyVpsSpec (vpsspec string) error {
	logrus.Infof("vpsspec:%v, ---SpecDir:%v--- App: %v ", vpsspec, Inst().SpecDir,Inst().AppList[0])

	appDir:= Inst().AppList[0]	
	f,err := os.Create(Inst().SpecDir +"/" + appDir +"/vps.yaml")
	if err != nil {
		logrus.Infof("Failed to create VPS spec: %v ",Inst().SpecDir +"/" + appDir +"/vps.yaml")
		return err
	}
	defer f.Close()

	nsize, err := f.WriteString(vpsspec)
	if err != nil {
		logrus.Infof("Failed to write VPS spec: %v ",Inst().SpecDir +"/" + appDir +"/vps.yaml")
		return err
	}
	f.Sync()
	logrus.Infof("==============Created VPS spec: %v size: %v",Inst().SpecDir +"/" + appDir +"/vps.yaml",nsize)
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
