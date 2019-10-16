package tests

import (
	"fmt"
	. "github.com/onsi/gomega"
	"github.com/portworx/torpedo/drivers/volume"
	. "github.com/portworx/torpedo/tests"
	"github.com/sirupsen/logrus"
)


type labelDict map[string]interface{}

type VpsTemplate interface {
	// Node label
	GetLabels()  [] labelDict
	// Pvc label
	GetPvcNodeLabels( lblnodes map[string][]string) map[string]map[string][]string
	// Vps Spec
	GetSpec ( ) string
	// Validate 
	Validate (appVolumes []*volume.Volume,volscheck map[string]map[string][]string)
	// Clean up
	CleanVps()
	// 
	

}

var (
		vpsrules = make(map[string] VpsTemplate)
	)

// Register registers the given vps rule
func Register(name string, d VpsTemplate) error {
	if _, ok := vpsrules[name]; !ok {
		vpsrules[name] = d
	} else {
		return fmt.Errorf("vps rule: %s is already registered", name)
	}

	return nil
}

// 
func GetVpsRules() map[string] VpsTemplate {
	return vpsrules
}

//-------
type Vpscase1 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

//# Case-1--enforcemnt: Required

func (v * Vpscase1) GetLabels() [] labelDict {

	lbldata := []labelDict{}
	node1lbl := labelDict{"media_type": "SSD","vps_test":"test"}  
	node2lbl := labelDict{"media_type": "SATA","vps_test":"test"}
	node3lbl := labelDict{"media_type": "SSD","vps_test":"test"}
	node4lbl := labelDict{"media_type": "SSD","vps_test":"test"}
	lbldata = append(lbldata, node1lbl, node2lbl,node3lbl,node4lbl)
		return lbldata
}

func (v * Vpscase1) GetPvcNodeLabels( lblnodes map[string][]string) map[string]map[string][]string {

	for key,val := range lblnodes {
		logrus.Infof("label node: key:%v Val:%v", key,val)
	}

	//Create 3 node lists (requiredNodes, prefNodes, notOnNodes)
	volnodelist := map[string]map[string][]string{}
	volnodelist["mysql-data"] = map[string][]string{}
	volnodelist["mysql-data-seq"] = map[string][]string{}
	volnodelist ["mysql-data"]["pnodes"] = [] string{}
	volnodelist ["mysql-data"]["nnodes"] = [] string{}
	volnodelist ["mysql-data-seq"]["pnodes"] = [] string{}
	volnodelist ["mysql-data-seq"]["nnodes"] = [] string{}


	for _,lnode := range lblnodes["media_typeSSD"] {
		volnodelist ["mysql-data"]["rnodes"] = append(volnodelist ["mysql-data"]["rnodes"], lnode)
		volnodelist ["mysql-data-seq"]["rnodes"] = append(volnodelist ["mysql-data-seq"]["rnodes"], lnode)
	}
/*
	for _,lnode := range lblnodes["media_typeSATA"] {
		volnodelist ["mysql-data"]["rnodes"] = append(volnodelist ["mysql-data"]["rnodes"], lnode)
		volnodelist ["mysql-data-seq"]["rnodes"] = append(volnodelist ["mysql-data-seq"]["rnodes"], lnode)
	}
*/
	return volnodelist
}


/*
 * 1. Each rule template, will provide the expected output
 */

func (v * Vpscase1) Validate(appVolumes []*volume.Volume,volscheck map[string]map[string][]string) {

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

func (v * Vpscase1)  GetSpec( ) string {

	var vpsspec string 
//	logrus.Infof(" rules:%v ", rules)	
	vpsspec =`apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: ssd-sata-pool-placement-spread
spec:
  replicaAffinity:
#  - affectedReplicas: 1
  - enforcement: required
    matchExpressions:
    - key: media_type
      operator: In
      values:
      - "SSD"`
/*  replicaAffinity:
  - affectedReplicas: 1
    enforcement: required
    matchExpressions:
    - key: media_type
      operator: In
      values:
      - "SATA"`
*/
	return vpsspec
}

func (v * Vpscase1) CleanVps() {
	logrus.Infof("Cleanup test case contexti for: %v", v.name)
}
func init() {
	v := &Vpscase1{"case1",true}
	Register(v.name, v)
}


//#---- Case 2 ---- enforcement: preferred

type Vpscase2 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v * Vpscase2) GetLabels() [] labelDict {

	lbldata := []labelDict{}
	node1lbl := labelDict{"media_type": "SSD","vps_test":"test"}  
	node2lbl := labelDict{"media_type": "SATA","vps_test":"test"}
	node3lbl := labelDict{"media_type": "SSD","vps_test":"test"}
	node4lbl := labelDict{"media_type": "SSD","vps_test":"test"}
	lbldata = append(lbldata, node1lbl, node2lbl,node3lbl,node4lbl)
		return lbldata
}

func (v * Vpscase2) GetPvcNodeLabels( lblnodes map[string][]string) map[string]map[string][]string {

	for key,val := range lblnodes {
		logrus.Infof("label node: key:%v Val:%v", key,val)
	}

	//Create 3 node lists (requiredNodes, prefNodes, notOnNodes)
	volnodelist := map[string]map[string][]string{}
	volnodelist["mysql-data"] = map[string][]string{}
	volnodelist["mysql-data-seq"] = map[string][]string{}
	volnodelist ["mysql-data"]["rnodes"] = [] string{}
	volnodelist ["mysql-data"]["nnodes"] = [] string{}
	volnodelist ["mysql-data-seq"]["rnodes"] = [] string{}
	volnodelist ["mysql-data-seq"]["nnodes"] = [] string{}


	for _,lnode := range lblnodes["media_typeSSD"] {
		volnodelist ["mysql-data"]["pnodes"] = append(volnodelist ["mysql-data"]["pnodes"], lnode)
		volnodelist ["mysql-data-seq"]["pnodes"] = append(volnodelist ["mysql-data-seq"]["pnodes"], lnode)
	}
/*
	for _,lnode := range lblnodes["media_typeSATA"] {
		volnodelist ["mysql-data"]["rnodes"] = append(volnodelist ["mysql-data"]["rnodes"], lnode)
		volnodelist ["mysql-data-seq"]["rnodes"] = append(volnodelist ["mysql-data-seq"]["rnodes"], lnode)
	}
*/
	return volnodelist
}


/*
 * 1. Each rule template, will provide the expected output
 */

func (v * Vpscase2) Validate(appVolumes []*volume.Volume,volscheck map[string]map[string][]string) {

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




func (v * Vpscase2) GetSpec( ) string {

	var vpsspec string 
	//logrus.Infof(" rules:%v ", rules)	
	vpsspec =`apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: ssd-sata-pool-placement-spread
spec:
  replicaAffinity:
#  - affectedReplicas: 1
  - enforcement: preferred
    matchExpressions:
    - key: media_type
      operator: In
      values:
      - "SSD"`
/*  replicaAffinity:
  - affectedReplicas: 1
    enforcement: preferred
    matchExpressions:
    - key: media_type
      operator: In
      values:
      - "SATA"`*/
	return vpsspec
}

func (v * Vpscase2) CleanVps() {
	logrus.Infof("Cleanup test case contexti for: %v", v.name)
}
func init() {
	v := &Vpscase2{"case2",true}
	Register(v.name, v)
}


