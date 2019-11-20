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
    // Get StorageClass placement_strategy
	GetScStrategyMap() map[string] string

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

//StorageClass placement_strategy mapping
func (v *vpscase1) GetScStrategyMap() map[string]string {
	return map[string]string {"placement-1":"placement-1", "placement-2":"placement-2", "placement-3":"",}
}

func (v *vpscase1) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-1
spec:
  replicaAffinity:
  - enforcement: required
    matchExpressions:
    - key: media_type
      operator: In
      values:
      - "SSD"
---
apiVersion: portworx.io/v1beta2 
kind: VolumePlacementStrategy
metadata:
  name: placement-2
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


//StorageClass placement_strategy mapping
func (v *vpscase2) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1":"placement-1", "placement-2":"placement-2", "placement-3":""}
}

func (v *vpscase2) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-1
spec:
  replicaAffinity:
  - enforcement: preferred
    matchExpressions:
    - key: media_type
      operator: In
      values:
      - "SSD"
---
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-2
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



//#---- Case 3 ----T809561: Verify Lt, Gt operators using latency and iops 
type vpscase3 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase3) GetLabels() []labelDict {

	lbldata := []labelDict{}
	node1lbl := labelDict{"iops": "90", "latency": "50"}
	node2lbl := labelDict{"iops": "80", "latency": "40"}
	node3lbl := labelDict{"iops": "70", "latency": "30"}
	node4lbl := labelDict{"iops": "60", "latency": "20"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl)
	return lbldata
}

func (v *vpscase3) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

	for key, val := range lblnodes {
		logrus.Debugf("label node: key:%v Val:%v", key, val)
	}

	//Create 3 node lists (requiredNodes, prefNodes, notOnNodes)
	volnodelist := map[string]map[string][]string{}
	volnodelist["mysql-data"] = map[string][]string{}
	volnodelist["mysql-data-seq"] = map[string][]string{}
	volnodelist["mysql-data"]["rnodes"] = []string{}
	volnodelist["mysql-data"]["nnodes"] = []string{}
	volnodelist["mysql-data-seq"]["pnodes"] = []string{}
	volnodelist["mysql-data-seq"]["nnodes"] = []string{}

	volnodelist["mysql-data"]["rnodes"] = append(volnodelist["mysql-data"]["rnodes"], lblnodes["iops90"][0])
	volnodelist["mysql-data"]["rnodes"] = append(volnodelist["mysql-data"]["rnodes"], lblnodes["iops80"][0])
	volnodelist["mysql-data"]["rnodes"] = append(volnodelist["mysql-data"]["rnodes"], lblnodes["iops70"][0])

	volnodelist["mysql-data-seq"]["rnodes"] = append(volnodelist["mysql-data-seq"]["rnodes"], lblnodes["latency40"][0])
	volnodelist["mysql-data-seq"]["rnodes"] = append(volnodelist["mysql-data-seq"]["rnodes"], lblnodes["latency30"][0])
	volnodelist["mysql-data-seq"]["rnodes"] = append(volnodelist["mysql-data-seq"]["rnodes"], lblnodes["latency20"][0])

	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase3) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

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


//StorageClass placement_strategy mapping
func (v *vpscase3) GetScStrategyMap() map[string] string{
	return map[string] string {"placement-1":"placement-1", "placement-2":"placement-2", "placement-3":""}
}

func (v *vpscase3) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-1
spec:
  replicaAffinity:
  - enforcement: required
    matchExpressions:
    - key: iops
      operator: Gt
      values:
      - "60"
---
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-2
spec:
  replicaAffinity:
  - enforcement: required
    matchExpressions:
    - key: latency
      operator: Lt
      values:
      - "50"`
	return vpsSpec
}

func (v *vpscase3) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}




//#---- Case 4 ----T863792  Verify Replica Affinity with topology keys
type vpscase4 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase4) GetLabels() []labelDict {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node7lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node8lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl,node5lbl, node6lbl,node7lbl,node8lbl)
	return lbldata
}

func (v *vpscase4) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

	for key, val := range lblnodes {
		logrus.Debugf("label node: key:%v Val:%v", key, val)
	}

	//Create 3 node lists (requiredNodes, prefNodes, notOnNodes)
	volnodelist := map[string]map[string][]string{}
	volnodelist["mysql-data"] = map[string][]string{}
	volnodelist["mysql-data-seq"] = map[string][]string{}
	volnodelist["mysql-data-aggr"] = map[string][]string{}
	volnodelist["mysql-data"]["pnodes"] = []string{}
	volnodelist["mysql-data"]["nnodes"] = []string{}
	volnodelist["mysql-data-seq"]["pnodes"] = []string{}
	volnodelist["mysql-data-seq"]["nnodes"] = []string{}
	volnodelist["mysql-data-aggr"]["pnodes"] = []string{}
	volnodelist["mysql-data-aggr"]["nnodes"] = []string{}
	volnodelist["mysql-data-aggr"]["rnodes1"] = []string{}

	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zoneeast"] {
		volnodelist["mysql-data"]["rnodes1"] = append(volnodelist["mysql-data"]["rnodes1"], lnode)
		volnodelist["mysql-data-seq"]["rnodes1"] = append(volnodelist["mysql-data-seq"]["rnodes1"], lnode)
		// Add nodes for aggr in set-2 for validation simplification
		volnodelist["mysql-data-aggr"]["rnodes2"] = append(volnodelist["mysql-data-aggr"]["rnodes2"], lnode)
	}

	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonewest"] {
		volnodelist["mysql-data"]["rnodes2"] = append(volnodelist["mysql-data"]["rnodes2"], lnode)
		volnodelist["mysql-data-seq"]["rnodes2"] = append(volnodelist["mysql-data-seq"]["rnodes2"], lnode)
		// Aggr replicas are spread across all nodes
		volnodelist["mysql-data-aggr"]["rnodes2"] = append(volnodelist["mysql-data-aggr"]["rnodes2"], lnode)
	}
	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase4) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	for _, appvol := range appVolumes {

		
		for vol, vnodes := range volscheck {

			if appvol.Name == vol {
				replicas, err := Inst().V.GetReplicaSetNodes(appvol)
				logrus.Debugf("==Replicas for vol: %s, Volume should have replicas on nodes:%v , Volume replicas are present on nodes :%v ", vol, vnodes, replicas)
				Expect(err).NotTo(HaveOccurred())
				Expect(replicas).NotTo(BeEmpty())

				foundinset := false
				// Must have (required)
				for _, rnode := range replicas {
					found := ""
					// Check whether replica is on the expected set of nodes
					for _, mnode := range vnodes["rnodes1"] {
						logrus.Debugf("Expected replica to be on Node:%v Replica is on Node:%v", mnode, rnode)
						if mnode == rnode {
							found = rnode
							break
						}
					}
				    if found == "" {
						foundinset=false
						break
					} else {
						foundinset=true
					}
				}

				//If replicas are not present in first set of labeled nodes, check other set
				if foundinset==false {
					for _, rnode := range replicas  {
						found := ""
					    // Check whether replica is on the expected set of nodes
						for _, mnode := range vnodes["rnodes2"] {
						    logrus.Debugf("Expected replica to be on Node:%v Replica is on Node:%v", mnode, rnode)
							if mnode == rnode {
								found = rnode
								break
							}
						}
						Expect(found).NotTo(BeEmpty(), fmt.Sprintf("Replica (%v) of Volume '%v' is not in the list of expected nodes(%v)", rnode, appvol, vnodes["rnodes2"]))
					}
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


//StorageClass placement_strategy mapping
func (v *vpscase4) GetScStrategyMap() map[string] string{
	return map[string] string {"placement-1":"placement-1", "placement-2":"placement-1", "placement-3":"placement-3"}
}

func (v *vpscase4) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-1
spec:
  replicaAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
---
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-3
spec:
  replicaAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_region`
	return vpsSpec
}

func (v *vpscase4) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}






// Test case inits
func init() {
	v := &vpscase1{"case1", true}
	Register(v.name, v)
}

func init() {
	v := &vpscase2{"case2-T863374", true}
	Register(v.name, v)
}


func init() {
	v := &vpscase3{"case3-T809561", true}
	Register(v.name, v)
}



func init() {
	v := &vpscase4{"case4-T863792", true}
	Register(v.name, v)
}





