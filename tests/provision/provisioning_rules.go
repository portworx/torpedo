package tests

import (
	"fmt"
	. "github.com/onsi/gomega"
	"github.com/portworx/torpedo/drivers/volume"
	. "github.com/portworx/torpedo/tests"
	"github.com/sirupsen/logrus"
)

const (
	mediaSsd  = "SSD"
	mediaSata = "SATA"
)

type labelDict map[string]interface{}

type vpsTemplate interface {
	// Node label
	GetLabels() []labelDict
	// Pvc label
	GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string
	// Vps Spec
	GetSpec() string
	// Validate
	Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string)
	// Clean up
	CleanVps()
}

var (
	vpsRules = make(map[string]vpsTemplate)
)

// Register registers the given vps rule
func Register(name string, d vpsTemplate) error {
	if _, ok := vpsRules[name]; !ok {
		vpsRules[name] = d
	} else {
		return fmt.Errorf("vps rule: %s is already registered", name)
	}

	return nil
}

// GetVpsRules return the list of vps rules
func GetVpsRules() map[string]vpsTemplate {
	return vpsRules
}

type vpscase1 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

//# Case-1--enforcemnt: Required
func (v *vpscase1) GetLabels() []labelDict {

	lbldata := []labelDict{}
	node1lbl := labelDict{"media_type": mediaSsd, "vps_test": "test"}
	node2lbl := labelDict{"media_type": mediaSata, "vps_test": "test"}
	node3lbl := labelDict{"media_type": mediaSsd, "vps_test": "test"}
	node4lbl := labelDict{"media_type": mediaSsd, "vps_test": "test"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl)
	return lbldata
}

func (v *vpscase1) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

	for key, val := range lblnodes {
		logrus.Debugf("label node: key:%v Val:%v", key, val)
	}

	//Create 3 node lists (requiredNodes, prefNodes, notOnNodes)
	volnodelist := map[string]map[string][]string{}
	volnodelist["mysql-data"] = map[string][]string{}
	volnodelist["mysql-data-seq"] = map[string][]string{}
	volnodelist["mysql-data"]["pnodes"] = []string{}
	volnodelist["mysql-data"]["nnodes"] = []string{}
	volnodelist["mysql-data-seq"]["pnodes"] = []string{}
	volnodelist["mysql-data-seq"]["nnodes"] = []string{}

	for _, lnode := range lblnodes["media_typeSSD"] {
		volnodelist["mysql-data"]["rnodes"] = append(volnodelist["mysql-data"]["rnodes"], lnode)
		volnodelist["mysql-data-seq"]["rnodes"] = append(volnodelist["mysql-data-seq"]["rnodes"], lnode)
	}

	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */
func (v *vpscase1) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	for _, appvol := range appVolumes {

		for vol, vnodes := range volscheck {

			if appvol.Name == vol {
				replicas, err := Inst().V.GetReplicaSetNodes(appvol)
				logrus.Debugf("==Replicas for vol: %s, appvol:%v Replicas:%v ", vol, appvol, replicas)
				Expect(err).NotTo(HaveOccurred())
				Expect(replicas).NotTo(BeEmpty())

				// Must have (required)
				for _, mnode := range vnodes["rnodes"] {
					found := ""
					for _, rnode := range replicas {
						logrus.Debugf("Expected Volume Node:%v Replica Node:%v", mnode, rnode)
						if mnode == rnode {
							found = rnode
							break
						}
					}
					Expect(found).NotTo(BeEmpty(), fmt.Sprintf("Volume '%v' does not have replica on node:'%v'", appvol, mnode))
				}

				// Preferred
				for _, mnode := range vnodes["pnodes"] {
					found := ""
					for _, rnode := range replicas {
						logrus.Debugf("Preferred Volume Node:%v Replica Node:%v", mnode, rnode)
						if mnode == rnode {
							found = rnode
							break
						}
					}
					if found != "" {
						logrus.Infof("Volume '%v' has replica on node:'%v'", appvol, mnode)
					}
				}

				// NotonNode
				for _, mnode := range vnodes["nnodes"] {
					var found string
					for _, rnode := range replicas {
						logrus.Debugf("Volume should not have replica on :%v Replica Node:%v", mnode, rnode)
						if mnode == rnode {
							found = rnode
							break
						}
					}
					Expect(found).To(BeEmpty(), fmt.Sprintf("Volume '%v' has replica on node:'%v'", appvol, mnode))
				}
			}
		}
	}
}

func (v *vpscase1) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: ssd-sata-pool-placement
spec:
  replicaAffinity:
  - enforcement: required
    matchExpressions:
    - key: media_type
      operator: In
      values:
      - "SSD"`
	return vpsSpec
}

func (v *vpscase1) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}

//#---- Case 2 ---- enforcement: preferred
type vpscase2 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase2) GetLabels() []labelDict {

	lbldata := []labelDict{}
	node1lbl := labelDict{"media_type": mediaSsd, "vps_test": "test"}
	node2lbl := labelDict{"media_type": mediaSata, "vps_test": "test"}
	node3lbl := labelDict{"media_type": mediaSsd, "vps_test": "test"}
	node4lbl := labelDict{"media_type": mediaSsd, "vps_test": "test"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl)
	return lbldata
}

func (v *vpscase2) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

	for key, val := range lblnodes {
		logrus.Debugf("label node: key:%v Val:%v", key, val)
	}

	//Create 3 node lists (requiredNodes, prefNodes, notOnNodes)
	volnodelist := map[string]map[string][]string{}
	volnodelist["mysql-data"] = map[string][]string{}
	volnodelist["mysql-data-seq"] = map[string][]string{}
	volnodelist["mysql-data"]["rnodes"] = []string{}
	volnodelist["mysql-data"]["nnodes"] = []string{}
	volnodelist["mysql-data-seq"]["rnodes"] = []string{}
	volnodelist["mysql-data-seq"]["nnodes"] = []string{}

	for _, lnode := range lblnodes["media_typeSSD"] {
		volnodelist["mysql-data"]["pnodes"] = append(volnodelist["mysql-data"]["pnodes"], lnode)
		volnodelist["mysql-data-seq"]["pnodes"] = append(volnodelist["mysql-data-seq"]["pnodes"], lnode)
	}

	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase2) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	for _, appvol := range appVolumes {

		for vol, vnodes := range volscheck {

			if appvol.Name == vol {
				replicas, err := Inst().V.GetReplicaSetNodes(appvol)
				logrus.Debugf("==Replicas for vol: %s, appvol:%v Replicas:%v ", vol, appvol, replicas)
				Expect(err).NotTo(HaveOccurred())
				Expect(replicas).NotTo(BeEmpty())

				// Must have (required)
				for _, mnode := range vnodes["rnodes"] {
					found := ""
					for _, rnode := range replicas {
						logrus.Debugf("Expected Volume Node:%v Replica Node:%v", mnode, rnode)
						if mnode == rnode {
							found = rnode
							break
						}
					}
					Expect(found).NotTo(BeEmpty(), fmt.Sprintf("Volume '%v' does not have replica on node:'%v'", appvol, mnode))
				}

				// Preferred
				for _, mnode := range vnodes["pnodes"] {
					found := ""
					for _, rnode := range replicas {
						logrus.Debugf("Preferred Volume Node:%v Replica Node:%v", mnode, rnode)
						if mnode == rnode {
							found = rnode
							break
						}
					}
					if found != "" {
						logrus.Infof("Volume '%v' has replica on node:'%v'", appvol, mnode)
					}
				}

				// NotonNode
				for _, mnode := range vnodes["nnodes"] {
					var found string
					for _, rnode := range replicas {
						logrus.Debugf("Volume should not have replica on :%v Replica Node:%v", mnode, rnode)
						if mnode == rnode {
							found = rnode
							break
						}
					}
					Expect(found).To(BeEmpty(), fmt.Sprintf("Volume '%v' has replica on node:'%v'", appvol, mnode))
				}
			}
		}
	}
}

func (v *vpscase2) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: ssd-sata-pool-placement
spec:
  replicaAffinity:
  - enforcement: preferred
    matchExpressions:
    - key: media_type
      operator: In
      values:
      - "SSD"`
	return vpsSpec
}

func (v *vpscase2) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}

// Test case inits
func init() {
	v := &vpscase1{"case1", true}
	Register(v.name, v)
}

func init() {
	v := &vpscase2{"case2", true}
	Register(v.name, v)
}
