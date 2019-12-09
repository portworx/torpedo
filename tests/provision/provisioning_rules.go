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
	// Node label and whether it needs to be set on node remove
	GetLabels() ([]labelDict,int)
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
	vpsRulesReplica = make(map[string]vpsTemplate)
	vpsRulesVolume = make(map[string]vpsTemplate)
	vpsRulesMix = make(map[string]vpsTemplate)
	vpsRulesMixScale = make(map[string]vpsTemplate)
	vpsRulesPending = make(map[string]vpsTemplate)
	vpsRulesDefaultLabel = make(map[string]vpsTemplate)
)

// Register registers the given vps rule
func Register(name string, d vpsTemplate, cat int) error {

	if cat == 1 { 
		if _, ok := vpsRulesReplica[name]; !ok {
			vpsRulesReplica[name] = d
		} else {
			return fmt.Errorf("vps rule: %s is already registered", name)
		}
	} else if cat == 2 {
		if _, ok := vpsRulesVolume[name]; !ok {
			vpsRulesVolume[name] = d
		} else {
			return fmt.Errorf("vps rule: %s is already registered", name)
		}
	} else if cat == 3 {
		if _, ok := vpsRulesMix[name]; !ok {
			vpsRulesMix[name] = d
		} else {
			return fmt.Errorf("vps rule: %s is already registered", name)
		}
	} else if cat == 4 {
		if _, ok := vpsRulesMixScale[name]; !ok {
			vpsRulesMixScale[name] = d
		} else {
			return fmt.Errorf("vps rule: %s is already registered", name)
		}
	} else if cat == 5 {
		if _, ok := vpsRulesPending[name]; !ok {
			vpsRulesPending[name] = d
		} else {
			return fmt.Errorf("vps rule: %s is already registered", name)
		}
	} else if cat == 6 {
		if _, ok := vpsRulesDefaultLabel[name]; !ok {
			vpsRulesDefaultLabel[name] = d
		} else {
			return fmt.Errorf("vps rule: %s is already registered", name)
		}
	} else {
			return fmt.Errorf("vps rule category: %d, is not valid", cat)
	}

	return nil
}

// GetVpsRules return the list of vps rules
func GetVpsRules(cat int) map[string]vpsTemplate {
	if cat == 1 {
		return vpsRulesReplica
	} else if cat ==2 {
		return vpsRulesVolume
	} else if cat ==3 {
		return vpsRulesMix
	} else if cat ==4 {
		return vpsRulesMixScale
	} else if cat ==5 {
		return vpsRulesPending
	} else if cat ==6 {
		return vpsRulesDefaultLabel
	} else {
		return nil
	}

}





/*
 *  
 *     Replica  Affinity and Anti-Affinity related test cases
 *
 */



type vpscase1 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

//# Case-1--enforcemnt: Required
func (v *vpscase1) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"media_type": mediaSsd, "vps_test": "test"}
	node2lbl := labelDict{"media_type": mediaSata, "vps_test": "test"}
	node3lbl := labelDict{"media_type": mediaSsd, "vps_test": "test"}
	node4lbl := labelDict{"media_type": mediaSsd, "vps_test": "test"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl)
	return lbldata, 1
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

func (v *vpscase2) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"media_type": mediaSsd, "vps_test": "test"}
	node2lbl := labelDict{"media_type": mediaSata, "vps_test": "test"}
	node3lbl := labelDict{"media_type": mediaSsd, "vps_test": "test"}
	node4lbl := labelDict{"media_type": mediaSsd, "vps_test": "test"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl)
	return lbldata,1
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

func (v *vpscase3) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"iops": "90", "latency": "50"}
	node2lbl := labelDict{"iops": "80", "latency": "40"}
	node3lbl := labelDict{"iops": "70", "latency": "30"}
	node4lbl := labelDict{"iops": "60", "latency": "20"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl)
	return lbldata,1
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

func (v *vpscase4) GetLabels() ([]labelDict,int) {

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
	return lbldata,1
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



//#---- Case 5 ----T1052921  Verify Replica Anti-Affinity with topology keys
type vpscase5 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase5) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "asia"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "asia"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "south", "failure-domain.beta.kubernetes.io/px_region": "eu"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "south", "failure-domain.beta.kubernetes.io/px_region": "eu"}
	node7lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node8lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl,node5lbl, node6lbl,node7lbl,node8lbl)
	return lbldata,1
}

func (v *vpscase5) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

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
		volnodelist["mysql-data-aggr"]["rnodes1"] = append(volnodelist["mysql-data-aggr"]["rnodes1"], lnode)
	}

	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonewest"] {
		volnodelist["mysql-data"]["rnodes2"] = append(volnodelist["mysql-data"]["rnodes2"], lnode)
		volnodelist["mysql-data-seq"]["rnodes2"] = append(volnodelist["mysql-data-seq"]["rnodes2"], lnode)
		// Aggr replicas are spread across all nodes
		volnodelist["mysql-data-aggr"]["rnodes2"] = append(volnodelist["mysql-data-aggr"]["rnodes2"], lnode)
	}

	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonesouth"] {
		volnodelist["mysql-data"]["rnodes3"] = append(volnodelist["mysql-data"]["rnodes3"], lnode)
		volnodelist["mysql-data-seq"]["rnodes3"] = append(volnodelist["mysql-data-seq"]["rnodes3"], lnode)
		// Aggr replicas are spread across all nodes
		volnodelist["mysql-data-aggr"]["rnodes3"] = append(volnodelist["mysql-data-aggr"]["rnodes3"], lnode)
	}

	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonenorth"] {
		volnodelist["mysql-data"]["rnodes4"] = append(volnodelist["mysql-data"]["rnodes4"], lnode)
		volnodelist["mysql-data-seq"]["rnodes4"] = append(volnodelist["mysql-data-seq"]["rnodes4"], lnode)
		// Aggr replicas are spread across all nodes
		volnodelist["mysql-data-aggr"]["rnodes4"] = append(volnodelist["mysql-data-aggr"]["rnodes4"], lnode)
	}
	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase5) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	for _, appvol := range appVolumes {

		
		for vol, vnodes := range volscheck {

			if appvol.Name == vol {
				replicaset, err := Inst().V.GetReplicatNodeSets(appvol)
				logrus.Debugf("==Replicas for vol: %s, Volume should have replicas on nodes:%v , Volume replicas are present on nodes :%v ", vol, vnodes, replicaset)
				Expect(err).NotTo(HaveOccurred())
				Expect(replicaset).NotTo(BeEmpty())

				for _,replicas := range replicaset {
					// Must have (required)
					// There are  3 replicas and 4 sets to check in.
					// In the 4 set, the replica should be place in the 3 of the sets.
					// A set cannot containe more than 1 replica
					
					Expect(replicas).NotTo(BeEmpty())

				    totalrepfound :=0
					// Check in set 1
					foundinset := 0
					for _, mnode := range vnodes["rnodes1"] {
						found := ""
						repOnNodeCnt:=0
						// Check whether replica is on the expected set of nodes
						for _, rnode := range  replicas {
							logrus.Debugf("Expected replica to be on Node:%v Replica is on Node:%v", mnode, rnode)
							if mnode == rnode {
								found = rnode
								repOnNodeCnt++
							}
						}

						if  found != "" {	   
							Expect(repOnNodeCnt).Should(BeNumerically("<=", 1))
							foundinset++
						}
					}

					Expect(foundinset).Should(BeNumerically("<=", 1))
					if foundinset ==1 {
						totalrepfound++
					}

					// Check in set 2
					foundinset = 0
					for _, mnode := range vnodes["rnodes2"] {
						found := ""
						repOnNodeCnt:=0
						// Check whether replica is on the expected set of nodes
						for _, rnode := range  replicas {
							logrus.Debugf("Expected replica to be on Node:%v Replica is on Node:%v", mnode, rnode)
							if mnode == rnode {
								found = rnode
								repOnNodeCnt++
							}
						}

						if  found != "" {	   
							Expect(repOnNodeCnt).Should(BeNumerically("<=", 1))
							foundinset++
						}
					}

					Expect(foundinset).Should(BeNumerically("<=", 1))
					if foundinset ==1 {
						totalrepfound++
					}

					// Check in set 3
					foundinset = 0
					for _, mnode := range vnodes["rnodes3"] {
						found := ""
						repOnNodeCnt:=0
						// Check whether replica is on the expected set of nodes
						for _, rnode := range  replicas {
							logrus.Debugf("Expected replica to be on Node:%v Replica is on Node:%v", mnode, rnode)
							if mnode == rnode {
								found = rnode
								repOnNodeCnt++
							}
						}

						if  found != "" {	   
							Expect(repOnNodeCnt).Should(BeNumerically("<=", 1))
							foundinset++
						}
					}

					Expect(foundinset).Should(BeNumerically("<=", 1))
					if foundinset ==1 {
						totalrepfound++
					}

					// Check in set 4
					foundinset = 0
					for _, mnode := range vnodes["rnodes4"] {
						found := ""
						repOnNodeCnt:=0
						// Check whether replica is on the expected set of nodes
						for _, rnode := range  replicas {
							logrus.Debugf("Expected replica to be on Node:%v Replica is on Node:%v", mnode, rnode)
							if mnode == rnode {
								found = rnode
								repOnNodeCnt++
							}
						}

						if  found != "" {	   
							Expect(repOnNodeCnt).Should(BeNumerically("<=", 1))
							foundinset++
						}
					}

					Expect(foundinset).Should(BeNumerically("<=", 1))
					if foundinset ==1 {
						totalrepfound++
					}

					if vol == "mysql-data-seq" || vol =="mysql-data" {
						// These are repl:3 vol
						Expect(totalrepfound).Should(Equal(3))
					} else {
						// These are repl:2 aggr:2 volume
						Expect(totalrepfound).Should(Equal(2))
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
}


//StorageClass placement_strategy mapping
func (v *vpscase5) GetScStrategyMap() map[string] string{
	return map[string] string {"placement-1":"placement-1", "placement-2":"placement-1", "placement-3":"placement-3"}
}

func (v *vpscase5) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-1
spec:
  replicaAntiAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
---
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-3
spec:
  replicaAntiAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_region`
	return vpsSpec
}

func (v *vpscase5) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}


//

//#---- Case 6 ---- T809554  Verify Replica Affinity with nodes not having the required labels
type vpscase6 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase6) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	return lbldata,0
}

func (v *vpscase6) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

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
	volnodelist["mysql-data"]["rnodes"] = []string{}
	volnodelist["mysql-data-seq"]["pnodes"] = []string{}
	volnodelist["mysql-data-seq"]["nnodes"] = []string{}
	volnodelist["mysql-data-seq"]["rnodes"] = []string{}
	volnodelist["mysql-data-aggr"]["pnodes"] = []string{}
	volnodelist["mysql-data-aggr"]["nnodes"] = []string{}
	volnodelist["mysql-data-aggr"]["rnodes"] = []string{}

	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase6) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	for _, appvol := range appVolumes {

		
		for vol, vnodes := range volscheck {

			if appvol.Name == vol {
				//Replicas should be in pending state and hence they should not
				// land on any node
				replicas, err := Inst().V.GetReplicaSetNodes(appvol)
				logrus.Debugf("==Replicas for vol: %s, Volume should not have replicas on any nodes:%v , Volume replicas are present on nodes :%v ", vol, vnodes, replicas)
				Expect(err).To(HaveOccurred())
				Expect(replicas).To(BeEmpty())

				
			}
		}
	}
}


//StorageClass placement_strategy mapping
func (v *vpscase6) GetScStrategyMap() map[string] string{
	return map[string] string {"placement-1":"placement-1", "placement-2":"placement-1", "placement-3":"placement-1"}
}

func (v *vpscase6) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-1
spec:
  replicaAffinity:
  - enforcement: required
    matchExpressions:
    - key: "region"
      operator: In
      values:
      - "infra"`
	return vpsSpec
}

func (v *vpscase6) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}




/*
 *  
 *     Volume  Affinity and Anti-Affinity related test cases
 *
 */

//#---- Case 7 ---- T809548  Verify volume affinity  --operator: Exists
type vpscase7 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase7) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	return lbldata,0
}

func (v *vpscase7) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

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
	volnodelist["mysql-data"]["rnodes"] = []string{}
	volnodelist["mysql-data-seq"]["pnodes"] = []string{}
	volnodelist["mysql-data-seq"]["nnodes"] = []string{}
	volnodelist["mysql-data-seq"]["rnodes"] = []string{}

	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase7) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	logrus.Infof("Case 7 T809548  Verify volume affinity 'exists', mysql-data-seq volume's replica should on nodes where mysql-data volume's replicas are present ")	

	var mysqlDataReplNodes []string
	var mysqlDataSeqReplNodes []string

	for _, appvol := range appVolumes {

		replicas, err := Inst().V.GetReplicaSetNodes(appvol)
		Expect(err).NotTo(HaveOccurred())
		Expect(replicas).NotTo(BeEmpty())

		if appvol.Name == "mysql-data" {
			mysqlDataReplNodes = replicas
		} else if appvol.Name == "mysql-data-seq" {
			mysqlDataSeqReplNodes = replicas
		}
				
	}

	// mysql-data-seq replicas should land on  the nodes where mysql-data volume replicas are present
	for _,rnode := range mysqlDataSeqReplNodes {
		found :=""
		for _,mnode := range mysqlDataReplNodes {
			logrus.Debugf("Volume (mysql-data-seq) replica node :%v should be same as mysql-data  Replica Node:%v", mnode, rnode)
			if rnode == mnode{
				logrus.Debugf("Volume should have replica on :%v Replica Node:%v", mnode, rnode)
				found = rnode
				break
			}
		}
		Expect(found).NotTo(BeEmpty(), fmt.Sprintf("Replica (%v) of Volume '%v' is not in the list of expected nodes(%v)", rnode, "mysql-data-seq", mysqlDataReplNodes))

	}

}


//StorageClass placement_strategy mapping
func (v *vpscase7) GetScStrategyMap() map[string] string{
	return map[string] string {"placement-1":"placement-2", "placement-2":"placement-2", "placement-3":""}
}

func (v *vpscase7) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-2
spec:
  volumeAffinity:
  - enforcement: required
    matchExpressions:
    - key: app
      operator: Exists`
	return vpsSpec
}

func (v *vpscase7) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}


/*
 *  
 *     Volume  Affinity and Anti-Affinity related test cases
 *
 */

//#---- Case 8 ---- T809548  Verify volume affinity - operator: In
type vpscase8 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase8) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	return lbldata,0
}

func (v *vpscase8) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

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
	volnodelist["mysql-data"]["rnodes"] = []string{}
	volnodelist["mysql-data-seq"]["pnodes"] = []string{}
	volnodelist["mysql-data-seq"]["nnodes"] = []string{}
	volnodelist["mysql-data-seq"]["rnodes"] = []string{}

	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase8) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	logrus.Infof("Case 8 T809548  Verify volume affinity 'In', mysql-data-seq volume's replica should be on nodes where mysql-data volume's replicas are present ")	

	var mysqlDataReplNodes []string
	var mysqlDataSeqReplNodes []string

	for _, appvol := range appVolumes {

		replicas, err := Inst().V.GetReplicaSetNodes(appvol)
		Expect(err).NotTo(HaveOccurred())
		Expect(replicas).NotTo(BeEmpty())

		if appvol.Name == "mysql-data" {
			mysqlDataReplNodes = replicas
		} else if appvol.Name == "mysql-data-seq" {
			mysqlDataSeqReplNodes = replicas
		}
				
	}

	// mysql-data-seq replicas should land on  the nodes where mysql-data volume replicas are present
	for _,rnode := range mysqlDataSeqReplNodes {
		found :=""
		for _,mnode := range mysqlDataReplNodes {
			logrus.Debugf("Volume (mysql-data-seq) replica node :%v should be same as mysql-data  Replica Node:%v", mnode, rnode)
			if rnode == mnode{
				logrus.Debugf("Volume should have replica on :%v Replica Node:%v", mnode, rnode)
				found = rnode
				break
			}
		}
		Expect(found).NotTo(BeEmpty(), fmt.Sprintf("Replica (%v) of Volume '%v' is not in the list of expected nodes(%v)", rnode, "mysql-data-seq", mysqlDataReplNodes))

	}

}


//StorageClass placement_strategy mapping
func (v *vpscase8) GetScStrategyMap() map[string] string{
	return map[string] string {"placement-1":"placement-2", "placement-2":"placement-2", "placement-3":""}
}

func (v *vpscase8) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-2
spec:
  volumeAffinity:
  - enforcement: required
    matchExpressions:
    - key: app
      operator: In
      values:
      - "mysql"`
	return vpsSpec
}

func (v *vpscase8) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}



//#---- Case 9 ---- T809548  Verify volume affinity - operator: DoesNotExist
type vpscase9 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase9) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	return lbldata,0
}

func (v *vpscase9) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

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
	volnodelist["mysql-data"]["rnodes"] = []string{}
	volnodelist["mysql-data-seq"]["pnodes"] = []string{}
	volnodelist["mysql-data-seq"]["nnodes"] = []string{}
	volnodelist["mysql-data-seq"]["rnodes"] = []string{}

	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase9) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	logrus.Infof("Case 9 T809548  Verify volume affinity 'DoesNotExist', mysql-data-seq volume's replica should not be on nodes where mysql-data volume's replicas are present ")	

	var mysqlDataReplNodes []string
	var mysqlDataSeqReplNodes []string

	for _, appvol := range appVolumes {

		replicas, err := Inst().V.GetReplicaSetNodes(appvol)
		Expect(err).NotTo(HaveOccurred())
		Expect(replicas).NotTo(BeEmpty())

		if appvol.Name == "mysql-data" {
			mysqlDataReplNodes = replicas
		} else if appvol.Name == "mysql-data-seq" {
			mysqlDataSeqReplNodes = replicas
		}
				
	}

	// mysql-data-seq replicas should land on  the nodes where mysql-data volume replicas are present
	for _,rnode := range mysqlDataSeqReplNodes {
		found :=""
		for _,mnode := range mysqlDataReplNodes {
			logrus.Debugf("Volume (mysql-data-seq) replica node :%v should not be same as mysql-data  Replica Node:%v", mnode, rnode)
			if rnode == mnode{
				logrus.Debugf("Volume should not have replica on :%v Replica Node:%v", mnode, rnode)
				found = rnode
				break
			}
		}
		Expect(found).To(BeEmpty(), fmt.Sprintf("Replica (%v) of Volume '%v' is  in the list of not expected nodes(%v)", rnode, "mysql-data-seq", mysqlDataReplNodes))

	}

}


//StorageClass placement_strategy mapping
func (v *vpscase9) GetScStrategyMap() map[string] string{
	return map[string] string {"placement-1":"placement-2", "placement-2":"placement-2", "placement-3":""}
}

func (v *vpscase9) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-2
spec:
  volumeAffinity:
  - enforcement: required
    matchExpressions:
    - key: app
      operator: DoesNotExist`
	return vpsSpec
}

func (v *vpscase9) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}




//#---- Case 10 ---- T809548  Verify volume affinity - operator: NotIn
type vpscase10 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase10) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	return lbldata,0
}

func (v *vpscase10) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

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
	volnodelist["mysql-data"]["rnodes"] = []string{}
	volnodelist["mysql-data-seq"]["pnodes"] = []string{}
	volnodelist["mysql-data-seq"]["nnodes"] = []string{}
	volnodelist["mysql-data-seq"]["rnodes"] = []string{}

	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase10) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	logrus.Infof("Case 10 T809548  Verify volume affinity 'NotIn', mysql-data-seq volume's replica should not be on nodes where mysql-data volume's replicas are present ")	

	var mysqlDataReplNodes []string
	var mysqlDataSeqReplNodes []string

	for _, appvol := range appVolumes {

		replicas, err := Inst().V.GetReplicaSetNodes(appvol)
		Expect(err).NotTo(HaveOccurred())
		Expect(replicas).NotTo(BeEmpty())

		if appvol.Name == "mysql-data" {
			mysqlDataReplNodes = replicas
		} else if appvol.Name == "mysql-data-seq" {
			mysqlDataSeqReplNodes = replicas
		}
				
	}

	// mysql-data-seq replicas should land on  the nodes where mysql-data volume replicas are present
	for _,rnode := range mysqlDataSeqReplNodes {
		found :=""
		for _,mnode := range mysqlDataReplNodes {
				logrus.Debugf("Volume (mysql-data-seq) replica node :%v should not be same as mysql-data  Replica Node:%v", mnode, rnode)
			if rnode == mnode{
				logrus.Debugf("Volume should not have replica on :%v Replica Node:%v", mnode, rnode)
				found = rnode
				break
			}
		}
		Expect(found).To(BeEmpty(), fmt.Sprintf("Replica (%v) of Volume '%v' is  in the list of not expected nodes(%v)", rnode, "mysql-data-seq", mysqlDataReplNodes))

	}

}


//StorageClass placement_strategy mapping
func (v *vpscase10) GetScStrategyMap() map[string] string{
	return map[string] string {"placement-1":"placement-2", "placement-2":"placement-2", "placement-3":""}
}

func (v *vpscase10) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-2
spec:
  volumeAffinity:
  - enforcement: required
    matchExpressions:
    - key: app
      operator: NotIn
      values:
      - "mysql"`
	return vpsSpec
}

func (v *vpscase10) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}

//T809549 Verify Volume Anti-Affinity  


//#---- Case 11 ---- T809549  Verify volume Anit-Affinity  --operator: Exists
type vpscase11 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase11) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	return lbldata,0
}

func (v *vpscase11) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

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
	volnodelist["mysql-data"]["rnodes"] = []string{}
	volnodelist["mysql-data-seq"]["pnodes"] = []string{}
	volnodelist["mysql-data-seq"]["nnodes"] = []string{}
	volnodelist["mysql-data-seq"]["rnodes"] = []string{}

	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase11) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	logrus.Infof("Case 11 T809549  Verify volume anti-affinity 'exists', mysql-data-seq volume's replica should not on come nodes where mysql-data volume's replicas are present ")	

	var mysqlDataReplNodes []string
	var mysqlDataSeqReplNodes []string

	for _, appvol := range appVolumes {

		replicas, err := Inst().V.GetReplicaSetNodes(appvol)
		Expect(err).NotTo(HaveOccurred())
		Expect(replicas).NotTo(BeEmpty())

		if appvol.Name == "mysql-data" {
			mysqlDataReplNodes = replicas
		} else if appvol.Name == "mysql-data-seq" {
			mysqlDataSeqReplNodes = replicas
		}
				
	}

	// mysql-data-seq replicas should land on  the nodes where mysql-data volume replicas are present
	for _,rnode := range mysqlDataSeqReplNodes {
		found :=""
		for _,mnode := range mysqlDataReplNodes {
			logrus.Debugf("Volume (mysql-data-seq) replica node :%v should be come on mysql-data  Replica Node:%v", mnode, rnode)
			if rnode == mnode{
				logrus.Debugf("Volume should not have replica on :%v Replica Node:%v", mnode, rnode)
				found = rnode
				break
			}
		}
		Expect(found).To(BeEmpty(), fmt.Sprintf("Replica (%v) of Volume '%v' should not come on the nodes(%v)", rnode, "mysql-data-seq", mysqlDataReplNodes))

	}

}


//StorageClass placement_strategy mapping
func (v *vpscase11) GetScStrategyMap() map[string] string{
	return map[string] string {"placement-1":"placement-2", "placement-2":"placement-2", "placement-3":""}
}

func (v *vpscase11) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-2
spec:
  volumeAntiAffinity:
  - enforcement: required
    matchExpressions:
    - key: app
      operator: Exists`
	return vpsSpec
}

func (v *vpscase11) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}


//#---- Case 12 ---- T809549  Verify volume anti-affinity - operator: In
type vpscase12 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase12) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	return lbldata,0
}

func (v *vpscase12) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

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
	volnodelist["mysql-data"]["rnodes"] = []string{}
	volnodelist["mysql-data-seq"]["pnodes"] = []string{}
	volnodelist["mysql-data-seq"]["nnodes"] = []string{}
	volnodelist["mysql-data-seq"]["rnodes"] = []string{}

	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase12) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	logrus.Infof("Case 12 T809549  Verify volume anti-affinity 'In', mysql-data-seq volume's replica should not be on nodes where mysql-data volume's replicas are present ")	

	var mysqlDataReplNodes []string
	var mysqlDataSeqReplNodes []string

	for _, appvol := range appVolumes {

		replicas, err := Inst().V.GetReplicaSetNodes(appvol)
		Expect(err).NotTo(HaveOccurred())
		Expect(replicas).NotTo(BeEmpty())

		if appvol.Name == "mysql-data" {
			mysqlDataReplNodes = replicas
		} else if appvol.Name == "mysql-data-seq" {
			mysqlDataSeqReplNodes = replicas
		}
				
	}

	// mysql-data-seq replicas should land on  the nodes where mysql-data volume replicas are present
	for _,rnode := range mysqlDataSeqReplNodes {
		found :=""
		for _,mnode := range mysqlDataReplNodes {
			logrus.Debugf("Volume (mysql-data-seq) replica node :%v should not be on mysql-data  Replica Node:%v", mnode, rnode)
			if rnode == mnode{
				logrus.Debugf("Volume should not have replica on :%v Replica Node:%v", mnode, rnode)
				found = rnode
				break
			}
		}
		Expect(found).To(BeEmpty(), fmt.Sprintf("Replica (%v) of Volume '%v' should not come on the the nodes( %v)", rnode, "mysql-data-seq", mysqlDataReplNodes))

	}

}


//StorageClass placement_strategy mapping
func (v *vpscase12) GetScStrategyMap() map[string] string{
	return map[string] string {"placement-1":"placement-2", "placement-2":"placement-2", "placement-3":""}
}

func (v *vpscase12) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-2
spec:
  volumeAntiAffinity:
  - enforcement: required
    matchExpressions:
    - key: app
      operator: In
      values:
      - "mysql"`
	return vpsSpec
}

func (v *vpscase12) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}



//#---- Case 13 ---- T809549  Verify volume anti-affinity - operator: DoesNotExist
type vpscase13 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase13) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	return lbldata,0
}

func (v *vpscase13) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

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
	volnodelist["mysql-data"]["rnodes"] = []string{}
	volnodelist["mysql-data-seq"]["pnodes"] = []string{}
	volnodelist["mysql-data-seq"]["nnodes"] = []string{}
	volnodelist["mysql-data-seq"]["rnodes"] = []string{}

	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase13) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	logrus.Infof("Case 13 T809549  Verify volume anti-affinity 'DoesNotExist', mysql-data-seq volume's replica should be on nodes where mysql-data volume's replicas are present ")	

	var mysqlDataReplNodes []string
	var mysqlDataSeqReplNodes []string

	for _, appvol := range appVolumes {

		replicas, err := Inst().V.GetReplicaSetNodes(appvol)
		Expect(err).NotTo(HaveOccurred())
		Expect(replicas).NotTo(BeEmpty())

		if appvol.Name == "mysql-data" {
			mysqlDataReplNodes = replicas
		} else if appvol.Name == "mysql-data-seq" {
			mysqlDataSeqReplNodes = replicas
		}
				
	}

	// mysql-data-seq replicas should land on  the nodes where mysql-data volume replicas are present
	for _,rnode := range mysqlDataSeqReplNodes {
		found :=""
		for _,mnode := range mysqlDataReplNodes {
			logrus.Debugf("Volume (mysql-data-seq) replica node :%v should be on mysql-data  Replica Node:%v", mnode, rnode)
			if rnode == mnode{
				logrus.Debugf("Volume should have replica on :%v Replica Node:%v", mnode, rnode)
				found = rnode
				break
			}
		}
		Expect(found).NotTo(BeEmpty(), fmt.Sprintf("Replica (%v) of Volume '%v' is  not the list of  expected nodes(%v)", rnode, "mysql-data-seq", mysqlDataReplNodes))

	}

}


//StorageClass placement_strategy mapping
func (v *vpscase13) GetScStrategyMap() map[string] string{
	return map[string] string {"placement-1":"placement-2", "placement-2":"placement-2", "placement-3":""}
}

func (v *vpscase13) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-2
spec:
  volumeAntiAffinity:
  - enforcement: required
    matchExpressions:
    - key: app
      operator: DoesNotExist`
	return vpsSpec
}

func (v *vpscase13) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}




//#---- Case 14 ---- T809549  Verify volume anti-affinity - operator: NotIn
type vpscase14 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase14) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	return lbldata,0
}

func (v *vpscase14) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

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
	volnodelist["mysql-data"]["rnodes"] = []string{}
	volnodelist["mysql-data-seq"]["pnodes"] = []string{}
	volnodelist["mysql-data-seq"]["nnodes"] = []string{}
	volnodelist["mysql-data-seq"]["rnodes"] = []string{}

	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase14) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	logrus.Infof("Case 14 T809549  Verify volume anti-affinity 'NotIn', mysql-data-seq volume's replica should be on nodes where mysql-data volume's replicas are present ")	

	var mysqlDataReplNodes []string
	var mysqlDataSeqReplNodes []string

	for _, appvol := range appVolumes {

		replicas, err := Inst().V.GetReplicaSetNodes(appvol)
		Expect(err).NotTo(HaveOccurred())
		Expect(replicas).NotTo(BeEmpty())

		if appvol.Name == "mysql-data" {
			mysqlDataReplNodes = replicas
		} else if appvol.Name == "mysql-data-seq" {
			mysqlDataSeqReplNodes = replicas
		}
				
	}

	for _,rnode := range mysqlDataSeqReplNodes {
		found :=""
		for _,mnode := range mysqlDataReplNodes {
				logrus.Debugf("Volume (mysql-data-seq) replica node :%v should be on mysql-data  Replica Node:%v", mnode, rnode)
			if rnode == mnode{
				logrus.Debugf("Volume should have replica on :%v Replica Node:%v", mnode, rnode)
				found = rnode
				break
			}
		}
		Expect(found).NotTo(BeEmpty(), fmt.Sprintf("Replica (%v) of Volume '%v' is not in the list of expected nodes(%v)", rnode, "mysql-data-seq", mysqlDataReplNodes))

	}

}


//StorageClass placement_strategy mapping
func (v *vpscase14) GetScStrategyMap() map[string] string{
	return map[string] string {"placement-1":"placement-2", "placement-2":"placement-2", "placement-3":""}
}

func (v *vpscase14) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-2
spec:
  volumeAntiAffinity:
  - enforcement: required
    matchExpressions:
    - key: app
      operator: NotIn
      values:
      - "mysql"`
	return vpsSpec
}

func (v *vpscase14) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}





//#---- Case 15 ---- T864665 Verify volume affinity with topology keys
type vpscase15 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase15) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "south", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "south", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "south", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node7lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node8lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl,node5lbl, node6lbl,node7lbl,node8lbl)
	return lbldata,1
}

func (v *vpscase15) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

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
	volnodelist["mysql-data"]["rnodes"] = []string{}
	volnodelist["mysql-data-seq"]["pnodes"] = []string{}
	volnodelist["mysql-data-seq"]["nnodes"] = []string{}
	volnodelist["mysql-data-seq"]["rnodes"] = []string{}
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zoneeast"] {
		volnodelist["mysql-data"]["rnodes"] = append(volnodelist["mysql-data"]["rnodes"], lnode)
		volnodelist["mysql-data-seq"]["rnodes"] = append(volnodelist["mysql-data-seq"]["rnodes"], lnode)
	}

	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_regionusa"] {
		volnodelist["mysql-data-aggr"]["rnodes"] = append(volnodelist["mysql-data-aggr"]["rnodes"], lnode)
	}

	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase15) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	logrus.Infof("Case 15 T864665  Verify volume affinity Topologykey, mysql-data and mysql-data-seq  should come together on nodes with px_zone label east and replicas of mysql-data-aggr come on nodes within same px_region of usa")	


	for _, appvol := range appVolumes {

		for vol, vnodes := range volscheck {
			if appvol.Name == vol {
				replicas, err := Inst().V.GetReplicaSetNodes(appvol)
				logrus.Debugf("==Replicas for vol: %s, Volume should have replicas on nodes:%v , Volume replicas are present on nodes :%v ", vol, vnodes, replicas)
				Expect(err).NotTo(HaveOccurred())
				Expect(replicas).NotTo(BeEmpty())

				// Must have (required)
				for _, rnode := range replicas {
					found := ""
					// Check whether replica is on the expected set of nodes
					for _, mnode := range vnodes["rnodes"] {
						logrus.Debugf("Expected replica to be on Node:%v Replica is on Node:%v", mnode, rnode)
						if mnode == rnode {
							found = rnode
							break
						}
					}
					Expect(found).NotTo(BeEmpty(), fmt.Sprintf("Volume '%v' replica %v , is not in the expected list of node:'%v'", appvol, rnode,vnodes["rnodes"] ))
				}

				// Go for next volume
				break
			}	
		}
	}


}


//StorageClass placement_strategy mapping
func (v *vpscase15) GetScStrategyMap() map[string] string{
	return map[string] string {"placement-1":"placement-2", "placement-2":"placement-2", "placement-3":"placement-3"}
}

func (v *vpscase15) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-2
spec:
  volumeAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
    matchExpressions:
      - key: app
        operator: In
        values:
          - "mysql"
  - enforcement: required
    matchExpressions:
      - key: "failure-domain.beta.kubernetes.io/px_zone"
        operator: In
        values:
          - "east"
---
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-3
spec:
  volumeAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_region
    matchExpressions:
      - key: app
        operator: In
        values:
          - "mysql"
  - enforcement: required
    matchExpressions:
      - key: "failure-domain.beta.kubernetes.io/px_region"
        operator: In
        values:
          - "usa"
---
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-1
spec:
  volumeAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
    matchExpressions:
      - key: "failure-domain.beta.kubernetes.io/px_zone"
        operator: In
        values:
          - "east"`
	return vpsSpec
}

func (v *vpscase15) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}


//#---- Case 16 ---- T1053359 Verify volume anti-affinity with topology keys
type vpscase16 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase16) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "central", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "middle", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "south", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node7lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node8lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl,node5lbl, node6lbl,node7lbl,node8lbl)
	return lbldata,1
}

func (v *vpscase16) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

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
	volnodelist["mysql-data"]["rnodes"] = []string{}
	volnodelist["mysql-data-seq"]["pnodes"] = []string{}
	volnodelist["mysql-data-seq"]["nnodes"] = []string{}
	volnodelist["mysql-data-seq"]["rnodes"] = []string{}
	//Create a list of nodes in px_zone east and north,
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zoneeast"] {
		volnodelist["mysql-data"]["rnodes"] = append(volnodelist["mysql-data"]["rnodes"], lnode)
		volnodelist["mysql-data-seq"]["rnodes"] = append(volnodelist["mysql-data-seq"]["rnodes"], lnode)
	}

	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonenorth"] {
		volnodelist["mysql-data"]["rnodes1"] = append(volnodelist["mysql-data"]["rnodes1"], lnode)
		volnodelist["mysql-data-seq"]["rnodes1"] = append(volnodelist["mysql-data-seq"]["rnodes1"], lnode)
	}
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonewest"] {
		volnodelist["mysql-data"]["rnodes2"] = append(volnodelist["mysql-data"]["rnodes2"], lnode)
		volnodelist["mysql-data-seq"]["rnodes2"] = append(volnodelist["mysql-data-seq"]["rnodes2"], lnode)
	}
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonecentral"] {
		volnodelist["mysql-data"]["rnodes3"] = append(volnodelist["mysql-data"]["rnodes3"], lnode)
		volnodelist["mysql-data-seq"]["rnodes3"] = append(volnodelist["mysql-data-seq"]["rnodes3"], lnode)
	}
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonemiddle"] {
		volnodelist["mysql-data"]["rnodes4"] = append(volnodelist["mysql-data"]["rnodes4"], lnode)
		volnodelist["mysql-data-seq"]["rnodes4"] = append(volnodelist["mysql-data-seq"]["rnodes4"], lnode)
	}

	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase16) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	logrus.Infof("Case 16 T1053359  Verify volume anti-affinity Topologykey, mysql-data and mysql-data-seq should not come together on same px_zone label")	

	var mysqlDataReplNodes []string
	var mysqlDataSeqReplNodes []string

	for _, appvol := range appVolumes {

		replicas, err := Inst().V.GetReplicaSetNodes(appvol)
		Expect(err).NotTo(HaveOccurred())
		Expect(replicas).NotTo(BeEmpty())

		if appvol.Name == "mysql-data" {
			mysqlDataReplNodes = replicas
		} else if appvol.Name == "mysql-data-seq" {
			mysqlDataSeqReplNodes = replicas
		}
				
	}

	/*
	// Replicas should not be on same node
	for _,rnode := range mysqlDataSeqReplNodes {
		found :=""
		for _,mnode := range mysqlDataReplNodes {
				logrus.Debugf("Volume (mysql-data-seq) replica node :%v should inot be on mysql-data  Replica Node:%v", mnode, rnode)
			if rnode == mnode{
				logrus.Debugf("Volume is having replica :%v  on Replica Node:%v", mnode, rnode)
				found = rnode
				break
			}
		}
		Expect(found).To(BeEmpty(), fmt.Sprintf("Replica (%v) of Volume '%v'  should not have replica in nodes(%v)", rnode, "mysql-data-seq", mysqlDataReplNodes))

	}*/

	//Replicas should not be in same zone
	for _, repset:= range volscheck["mysql-data"] {
		// for each node in the zone, check replica count should be one
		repfoundseq :=0
		repfound :=0
		for _, mnode := range repset {
			for _,rnode := range mysqlDataSeqReplNodes { 			 
				if rnode == mnode {
					repfoundseq++
				}
			}

			for _,rnode := range mysqlDataReplNodes { 			 
				if rnode == mnode {
					repfound++
				}
			}
		}
		if repfoundseq >=1 && repfound >= 1 {
			logrus.Debugf("Both volumes are having replicas in the same zone, nodes :%v  replica count of mysql-data:%v  replica count of mysql-data-seq:%v", repset, repfound, repfoundseq)
			Expect(1).NotTo(Equal(2), fmt.Sprintf("px_zone nodes(%v) has more than one( mysql-data: %v & mysql-data-seq:%v) replica of the volumes in  Volume Anti-affinity test case",repset,repfound, repfoundseq ) )

		}
	}

}


//StorageClass placement_strategy mapping
func (v *vpscase16) GetScStrategyMap() map[string] string{
	return map[string] string {"placement-1":"placement-1", "placement-2":"placement-2", "placement-3":""}
}

func (v *vpscase16) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-2
spec:
  volumeAntiAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
    matchExpressions:
      - key: app
        operator: In
        values:
          - "mysql"
---
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-3
spec:
  volumeAntiAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_region
    matchExpressions:
      - key: app
        operator: In
        values:
          - "mysql"
---
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-1
spec:
  volumeAntiAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
    matchExpressions:
      - key: app
        operator: In
        values:
          - "mysql"`
	return vpsSpec
}

func (v *vpscase16) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}



//#---- Case 17 ---- T870615 Verify volume anti-affinity multiple rules 
type vpscase17 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase17) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "central", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "middle", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "south", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node7lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node8lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl,node5lbl, node6lbl,node7lbl,node8lbl)
	return lbldata,1
}

func (v *vpscase17) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

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
	volnodelist["mysql-data"]["rnodes"] = []string{}
	volnodelist["mysql-data-seq"]["pnodes"] = []string{}
	volnodelist["mysql-data-seq"]["nnodes"] = []string{}
	volnodelist["mysql-data-seq"]["rnodes"] = []string{}

	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase17) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	logrus.Infof("Case 17 T870615  Verify volume anti-affinity with multiple rules , mysql-data and mysql-data-seq should not come together on a node")	

	var mysqlDataReplNodes []string
	var mysqlDataSeqReplNodes []string

	for _, appvol := range appVolumes {

		replicas, err := Inst().V.GetReplicaSetNodes(appvol)
		Expect(err).NotTo(HaveOccurred())
		Expect(replicas).NotTo(BeEmpty())

		if appvol.Name == "mysql-data" {
			mysqlDataReplNodes = replicas
		} else if appvol.Name == "mysql-data-seq" {
			mysqlDataSeqReplNodes = replicas
		}
				
	}

	// Replicas should not be on same node
	for _,rnode := range mysqlDataSeqReplNodes {
		found :=""
		for _,mnode := range mysqlDataReplNodes {
				logrus.Debugf("Volume (mysql-data-seq) replica node :%v should inot be on mysql-data  Replica Node:%v", mnode, rnode)
			if rnode == mnode{
				logrus.Debugf("Volume is having replica :%v  on Replica Node:%v", mnode, rnode)
				found = rnode
				break
			}
		}
		Expect(found).To(BeEmpty(), fmt.Sprintf("Replica (%v) of Volume '%v'  should not have replica in nodes(%v)", rnode, "mysql-data-seq", mysqlDataReplNodes))

	}

}


//StorageClass placement_strategy mapping
func (v *vpscase17) GetScStrategyMap() map[string] string{
	return map[string] string {"placement-1":"placement-1", "placement-2":"placement-2", "placement-3":""}
}

func (v *vpscase17) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-1
spec:
  volumeAntiAffinity:
  - enforcement: required
    matchExpressions:
      - key: app
        operator: In
        values:
          - "mysql"
  - enforcement: required
    matchExpressions:
      - key: voltype
        operator: In
        values:
         - "seq"
---
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-2
spec:
  volumeAntiAffinity:
  - enforcement: required
    matchExpressions:
      - key: app
        operator: In
        values:
          - "mysql"
  - enforcement: required
    matchExpressions:
      - key: voltype
        operator: In
        values:
         - "data"`
	return vpsSpec
}

func (v *vpscase17) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}


/*
 *  
 *     Replicas & Volume  Affinity and Anti-Affinity related test cases
 *
 */



//#---- Case 18 ---- T866365 Verify replica and volume affinity topology 
// keys	 with volume labels 
type vpscase18 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase18) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node7lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node8lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl,node5lbl, node6lbl,node7lbl,node8lbl)
	return lbldata,1
}

func (v *vpscase18) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

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
	volnodelist["mysql-data"]["rnodes"] = []string{}
	volnodelist["mysql-data-seq"]["pnodes"] = []string{}
	volnodelist["mysql-data-seq"]["nnodes"] = []string{}
	volnodelist["mysql-data-seq"]["rnodes"] = []string{}

	//Create a list of nodes in px_zone east and north,
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zoneeast"] {
		volnodelist["mysql-data"]["rnodes"] = append(volnodelist["mysql-data"]["rnodes"], lnode)
		volnodelist["mysql-data-seq"]["rnodes"] = append(volnodelist["mysql-data-seq"]["rnodes"], lnode)
	}

	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonenorth"] {
		volnodelist["mysql-data"]["rnodes1"] = append(volnodelist["mysql-data"]["rnodes1"], lnode)
		volnodelist["mysql-data-seq"]["rnodes1"] = append(volnodelist["mysql-data-seq"]["rnodes1"], lnode)
	}
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonewest"] {
		volnodelist["mysql-data"]["rnodes2"] = append(volnodelist["mysql-data"]["rnodes2"], lnode)
		volnodelist["mysql-data-seq"]["rnodes2"] = append(volnodelist["mysql-data-seq"]["rnodes2"], lnode)
	}
	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase18) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	logrus.Infof("Case 18 T866365 Verify replica and volume affinity topology keys with volume labels ")	

	var mysqlDataReplNodes []string
	var mysqlDataSeqReplNodes []string

	for _, appvol := range appVolumes {

		replicas, err := Inst().V.GetReplicaSetNodes(appvol)
		Expect(err).NotTo(HaveOccurred())
		Expect(replicas).NotTo(BeEmpty())

		if appvol.Name == "mysql-data" {
			mysqlDataReplNodes = replicas
		} else if appvol.Name == "mysql-data-seq" {
			mysqlDataSeqReplNodes = replicas
		}
				
	}

	//Replicas should be in same zone
	nodeinzone :=0
	for _, repset:= range volscheck["mysql-data"] {
		// for each node in the zone, check replica count should be one
		repfoundseq :=0
		repfound :=0
		for _, mnode := range repset {
			for _,rnode := range mysqlDataSeqReplNodes { 			 
				if rnode == mnode {
					repfoundseq++
				}
			}

			for _,rnode := range mysqlDataReplNodes { 			 
				if rnode == mnode {
					repfound++
				}
			}
		}
		if repfoundseq == 3 && repfound == 3 {
			nodeinzone = 1
		}
	}

	Expect(nodeinzone).To(Equal(1), fmt.Sprintf("The replicas of volume mysql-data: %v & mysql-data-seq:%v are not in same zone",mysqlDataReplNodes ,mysqlDataSeqReplNodes ) )

}


//StorageClass placement_strategy mapping
func (v *vpscase18) GetScStrategyMap() map[string] string{
	return map[string] string {"placement-1":"placement-1", "placement-2":"placement-2", "placement-3":""}
}

func (v *vpscase18) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-2
spec:
  volumeAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
    matchExpressions:
      - key: app
        operator: In
        values:
          - "mysql"
  replicaAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
---
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-3
spec:
  volumeAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_region
    matchExpressions:
      - key: app
        operator: In
        values:
          - "mysql"
---
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-1
spec:
  volumeAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
    matchExpressions:
      - key: app
        operator: In
        values:
          - "mysql"
  replicaAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone`
	return vpsSpec
}

func (v *vpscase18) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}




//#---- Case 19 ---- T866790
//Verify replica affinity and volume anti-affinity topology keys with volume labels 
// keys	 with volume labels 
type vpscase19 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase19) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node7lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node8lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl,node5lbl, node6lbl,node7lbl,node8lbl)
	return lbldata,1
}

func (v *vpscase19) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

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
	volnodelist["mysql-data"]["rnodes"] = []string{}
	volnodelist["mysql-data-seq"]["pnodes"] = []string{}
	volnodelist["mysql-data-seq"]["nnodes"] = []string{}
	volnodelist["mysql-data-seq"]["rnodes"] = []string{}

	//Create a list of nodes in px_zone east and north,
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zoneeast"] {
		volnodelist["mysql-data"]["rnodes"] = append(volnodelist["mysql-data"]["rnodes"], lnode)
		volnodelist["mysql-data-seq"]["rnodes"] = append(volnodelist["mysql-data-seq"]["rnodes"], lnode)
	}

	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonenorth"] {
		volnodelist["mysql-data"]["rnodes1"] = append(volnodelist["mysql-data"]["rnodes1"], lnode)
		volnodelist["mysql-data-seq"]["rnodes1"] = append(volnodelist["mysql-data-seq"]["rnodes1"], lnode)
	}
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonewest"] {
		volnodelist["mysql-data"]["rnodes2"] = append(volnodelist["mysql-data"]["rnodes2"], lnode)
		volnodelist["mysql-data-seq"]["rnodes2"] = append(volnodelist["mysql-data-seq"]["rnodes2"], lnode)
	}
	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase19) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	logrus.Infof("Case 19 T866790 Verify replica affinity and volume anti-affinity topology keys with volume labels. The 2 volumes should come in different zone, with replicas staying within the zone ")	

	var mysqlDataReplNodes []string
	var mysqlDataSeqReplNodes []string

	for _, appvol := range appVolumes {

		replicas, err := Inst().V.GetReplicaSetNodes(appvol)
		Expect(err).NotTo(HaveOccurred())
		Expect(replicas).NotTo(BeEmpty())

		if appvol.Name == "mysql-data" {
			mysqlDataReplNodes = replicas
		} else if appvol.Name == "mysql-data-seq" {
			mysqlDataSeqReplNodes = replicas
		}
				
	}

	//Replicas should be in same zone
	nodeinseqzone :=0
	nodeindatazone :=0
	for _, repset:= range volscheck["mysql-data"] {
		// for each node in the zone, check replica count should be one
		repfoundseq :=0
		repfound :=0
		for _, mnode := range repset {
			for _,rnode := range mysqlDataSeqReplNodes { 			 
				if rnode == mnode {
					repfoundseq++
				}
			}

			for _,rnode := range mysqlDataReplNodes { 			 
				if rnode == mnode {
					repfound++
				}
			}
		}
		if repfoundseq == 3 && repfound == 0 {
			nodeinseqzone = 1
		} else if repfoundseq == 0 && repfound == 3 {
			nodeindatazone = 1
		}
	}

	Expect(nodeinseqzone + nodeindatazone).To(Equal(2), fmt.Sprintf("Some or all of the replicas of volume mysql-data: %v & mysql-data-seq:%v are in same zone",mysqlDataReplNodes ,mysqlDataSeqReplNodes ) )

}


//StorageClass placement_strategy mapping
func (v *vpscase19) GetScStrategyMap() map[string] string{
	return map[string] string {"placement-1":"placement-1", "placement-2":"placement-2", "placement-3":""}
}

func (v *vpscase19) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-2
spec:
  volumeAntiAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
    matchExpressions:
      - key: app
        operator: In
        values:
          - "mysql"
  replicaAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
---
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-3
spec:
  volumeAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_region
    matchExpressions:
      - key: app
        operator: In
        values:
          - "mysql"
---
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-1
spec:
  volumeAntiAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
    matchExpressions:
      - key: app
        operator: In
        values:
          - "mysql"
  replicaAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone`
	return vpsSpec
}

func (v *vpscase19) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}


//#---- Case 20 ---- T866790
//Verify replica anti-affinity and volume affinity topology keys with volume lables
//
type vpscase20 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase20) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "south", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "south", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node7lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node8lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl,node5lbl, node6lbl,node7lbl,node8lbl)
	return lbldata,1
}

func (v *vpscase20) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

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
	volnodelist["mysql-data"]["rnodes"] = []string{}
	volnodelist["mysql-data-seq"]["pnodes"] = []string{}
	volnodelist["mysql-data-seq"]["nnodes"] = []string{}
	volnodelist["mysql-data-seq"]["rnodes"] = []string{}

	//Create a list of nodes in px_zone east and north,
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zoneeast"] {
		volnodelist["mysql-data"]["rnodes"] = append(volnodelist["mysql-data"]["rnodes"], lnode)
		volnodelist["mysql-data-seq"]["rnodes"] = append(volnodelist["mysql-data-seq"]["rnodes"], lnode)
	}

	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonenorth"] {
		volnodelist["mysql-data"]["rnodes1"] = append(volnodelist["mysql-data"]["rnodes1"], lnode)
		volnodelist["mysql-data-seq"]["rnodes1"] = append(volnodelist["mysql-data-seq"]["rnodes1"], lnode)
	}
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonewest"] {
		volnodelist["mysql-data"]["rnodes2"] = append(volnodelist["mysql-data"]["rnodes2"], lnode)
		volnodelist["mysql-data-seq"]["rnodes2"] = append(volnodelist["mysql-data-seq"]["rnodes2"], lnode)
	}
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonesouth"] {
		volnodelist["mysql-data"]["rnodes3"] = append(volnodelist["mysql-data"]["rnodes3"], lnode)
		volnodelist["mysql-data-seq"]["rnodes3"] = append(volnodelist["mysql-data-seq"]["rnodes3"], lnode)
	}
	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase20) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	logrus.Infof("Case 20 T866790 Verify replica anti-affinity and volume affinity topology keys with volume lables. Replicas of a volume should not come together in px_zone, replicas of different volume can come together")	

	var mysqlDataReplNodes []string
	var mysqlDataSeqReplNodes []string

	for _, appvol := range appVolumes {

		replicas, err := Inst().V.GetReplicaSetNodes(appvol)
		Expect(err).NotTo(HaveOccurred())
		Expect(replicas).NotTo(BeEmpty())

		if appvol.Name == "mysql-data" {
			mysqlDataReplNodes = replicas
		} else if appvol.Name == "mysql-data-seq" {
			mysqlDataSeqReplNodes = replicas
		}
				
	}

	//Replicas should not be in same zone, replicas of different volume should 
	//be in same zone
	volrepinzone :=0
	for _, repset:= range volscheck["mysql-data"] {
		// for each node in the zone, check replica count should be one
		repfoundseq :=0
		repfound :=0
		for _, mnode := range repset {
			for _,rnode := range mysqlDataSeqReplNodes { 			 
				if rnode == mnode {
					repfoundseq++
				}
			}

			for _,rnode := range mysqlDataReplNodes { 			 
				if rnode == mnode {
					repfound++
				}
			}
		}
		Expect(repfoundseq).Should(BeNumerically("<=", 1), fmt.Sprintf("More than one replicas is found in same zone for  mysql-data-seq:%v ",mysqlDataSeqReplNodes ) )
		Expect(repfound).Should(BeNumerically("<=", 1), fmt.Sprintf("More than one replica is found in same zone for volume mysql-data: %v", mysqlDataReplNodes ) )
		if(repfound == 1 && repfoundseq ==0 ) || (repfound == 0 && repfoundseq ==1 ) {
			volrepinzone =1
		}
	}

	Expect(volrepinzone).To(Equal(0), fmt.Sprintf("Due to volume affinity replicas of volumes mysql-data: %v &  mysql-data-seq:%v should have been in same zone",mysqlDataReplNodes,mysqlDataSeqReplNodes ) )
}


//StorageClass placement_strategy mapping
func (v *vpscase20) GetScStrategyMap() map[string] string{
	return map[string] string {"placement-1":"placement-1", "placement-2":"placement-2", "placement-3":""}
}

func (v *vpscase20) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-2
spec:
  volumeAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
    matchExpressions:
      - key: app
        operator: In
        values:
          - "mysql"
  replicaAntiAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
---
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-3
spec:
  volumeAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_region
    matchExpressions:
      - key: app
        operator: In
        values:
          - "mysql"
---
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-1
spec:
  volumeAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
    matchExpressions:
      - key: app
        operator: In
        values:
          - "mysql"
  replicaAntiAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone`
	return vpsSpec
}

func (v *vpscase20) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}




//#---- Case 21 ---- T867640
//Verify replica anti-affinity and volume anti-affinity topology keys with volume labels 
type vpscase21 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase21) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "middleast", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "central", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "south", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node7lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node8lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl,node5lbl, node6lbl,node7lbl,node8lbl)
	return lbldata,1
}

func (v *vpscase21) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

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
	volnodelist["mysql-data"]["rnodes"] = []string{}
	volnodelist["mysql-data-seq"]["pnodes"] = []string{}
	volnodelist["mysql-data-seq"]["nnodes"] = []string{}
	volnodelist["mysql-data-seq"]["rnodes"] = []string{}

	//Create a list of nodes in px_zone east and north,
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zoneeast"] {
		volnodelist["mysql-data"]["rnodes"] = append(volnodelist["mysql-data"]["rnodes"], lnode)
		volnodelist["mysql-data-seq"]["rnodes"] = append(volnodelist["mysql-data-seq"]["rnodes"], lnode)
	}

	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonenorth"] {
		volnodelist["mysql-data"]["rnodes1"] = append(volnodelist["mysql-data"]["rnodes1"], lnode)
		volnodelist["mysql-data-seq"]["rnodes1"] = append(volnodelist["mysql-data-seq"]["rnodes1"], lnode)
	}
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonewest"] {
		volnodelist["mysql-data"]["rnodes2"] = append(volnodelist["mysql-data"]["rnodes2"], lnode)
		volnodelist["mysql-data-seq"]["rnodes2"] = append(volnodelist["mysql-data-seq"]["rnodes2"], lnode)
	}
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonesouth"] {
		volnodelist["mysql-data"]["rnodes3"] = append(volnodelist["mysql-data"]["rnodes3"], lnode)
		volnodelist["mysql-data-seq"]["rnodes3"] = append(volnodelist["mysql-data-seq"]["rnodes3"], lnode)
	}
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonemiddleast"] {
		volnodelist["mysql-data"]["rnodes3"] = append(volnodelist["mysql-data"]["rnodes3"], lnode)
		volnodelist["mysql-data-seq"]["rnodes3"] = append(volnodelist["mysql-data-seq"]["rnodes3"], lnode)
	}
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonecentral"] {
		volnodelist["mysql-data"]["rnodes3"] = append(volnodelist["mysql-data"]["rnodes3"], lnode)
		volnodelist["mysql-data-seq"]["rnodes3"] = append(volnodelist["mysql-data-seq"]["rnodes3"], lnode)
	}
	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase21) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	logrus.Infof("Case 21 T867640 Verify replica anti-affinity and volume anti-affinity topology keys with volume labels ")	

	var mysqlDataReplNodes []string
	var mysqlDataSeqReplNodes []string

	for _, appvol := range appVolumes {

		replicas, err := Inst().V.GetReplicaSetNodes(appvol)
		Expect(err).NotTo(HaveOccurred())
		Expect(replicas).NotTo(BeEmpty())

		if appvol.Name == "mysql-data" {
			mysqlDataReplNodes = replicas
		} else if appvol.Name == "mysql-data-seq" {
			mysqlDataSeqReplNodes = replicas
		}
				
	}

	//Replicas should not be in same zone, replicas of different volume should 
	//be in same zone
	volrepinzone :=0
	for _, repset:= range volscheck["mysql-data"] {
		// for each node in the zone, check replica count should be one
		repfoundseq :=0
		repfound :=0
		for _, mnode := range repset {
			for _,rnode := range mysqlDataSeqReplNodes { 			 
				if rnode == mnode {
					repfoundseq++
				}
			}

			for _,rnode := range mysqlDataReplNodes { 			 
				if rnode == mnode {
					repfound++
				}
			}
		}
		if (repfound >=1  && repfoundseq >=1) || repfound > 1 || repfoundseq > 1   {
			volrepinzone =1
		}
	}

	Expect(volrepinzone).To(Equal(0), fmt.Sprintf("Due to volume & replicas anti-affinity replicas of volumes mysql-data: %v &  mysql-data-seq:%v should not appear in same zone",mysqlDataReplNodes,mysqlDataSeqReplNodes ) )
}


//StorageClass placement_strategy mapping
func (v *vpscase21) GetScStrategyMap() map[string] string{
	return map[string] string {"placement-1":"placement-1", "placement-2":"placement-1", "placement-3":""}
}

func (v *vpscase21) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-1
spec:
  volumeAntiAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
    matchExpressions:
      - key: app
        operator: In
        values:
          - "mysql"
  replicaAntiAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone`
	return vpsSpec
}

func (v *vpscase21) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}



//#---- Case 22 ---- T871040
// Verify statefulset/deployment scale up/down w.r.t replica and volume affinity rules
type vpscase22 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase22) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "central", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node7lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node8lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl,node5lbl, node6lbl,node7lbl,node8lbl)
	return lbldata,1
}

func (v *vpscase22) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

	for key, val := range lblnodes {
		logrus.Debugf("label node: key:%v Val:%v", key, val)
	}

	//Create 3 node lists (requiredNodes, prefNodes, notOnNodes)
	volnodelist := map[string]map[string][]string{}
	volnodelist["es-data-esnode-0"] = map[string][]string{}
	volnodelist["es-data-esnode-1"] = map[string][]string{}
	volnodelist["es-data-esnode-2"] = map[string][]string{}
	volnodelist["es-data-esnode-0"]["pnodes"] = []string{}
	volnodelist["es-data-esnode-0"]["nnodes"] = []string{}
	volnodelist["es-data-esnode-0"]["rnodes"] = []string{}
	volnodelist["es-data-esnode-1"]["pnodes"] = []string{}
	volnodelist["es-data-esnode-1"]["nnodes"] = []string{}
	volnodelist["es-data-esnode-1"]["rnodes"] = []string{}

	//Create a list of nodes in px_zone east and north,
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zoneeast"] {
		volnodelist["es-data-esnode-0"]["rnodes"] = append(volnodelist["es-data-esnode-0"]["rnodes"], lnode)
		volnodelist["es-data-esnode-1"]["rnodes"] = append(volnodelist["es-data-esnode-1"]["rnodes"], lnode)
		volnodelist["es-data-esnode-2"]["rnodes"] = append(volnodelist["es-data-esnode-2"]["rnodes"], lnode)
	}

	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonenorth"] {
		volnodelist["es-data-esnode-0"]["rnodes1"] = append(volnodelist["es-data-esnode-0"]["rnodes1"], lnode)
		volnodelist["es-data-esnode-1"]["rnodes1"] = append(volnodelist["es-data-esnode-1"]["rnodes1"], lnode)
		volnodelist["es-data-esnode-2"]["rnodes1"] = append(volnodelist["es-data-esnode-2"]["rnodes1"], lnode)
	}
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonewest"] {
		volnodelist["es-data-esnode-0"]["rnodes2"] = append(volnodelist["es-data-esnode-0"]["rnodes2"], lnode)
		volnodelist["es-data-esnode-1"]["rnodes2"] = append(volnodelist["es-data-esnode-1"]["rnodes2"], lnode)
		volnodelist["es-data-esnode-2"]["rnodes2"] = append(volnodelist["es-data-esnode-2"]["rnodes2"], lnode)
	}
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonecentral"] {
		volnodelist["es-data-esnode-0"]["rnodes3"] = append(volnodelist["es-data-esnode-0"]["rnodes3"], lnode)
		volnodelist["es-data-esnode-1"]["rnodes3"] = append(volnodelist["es-data-esnode-1"]["rnodes3"], lnode)
		volnodelist["es-data-esnode-2"]["rnodes3"] = append(volnodelist["es-data-esnode-2"]["rnodes3"], lnode)
	}
	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase22) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	logrus.Infof("Case 22 T871040 Verify statefulset/deployment scale up/down w.r.t replica and volume affinity rules ")	

	var esData0Nodes []string
	var esData1Nodes []string
	var esData2Nodes []string

	for _, appvol := range appVolumes {

		replicas, err := Inst().V.GetReplicaSetNodes(appvol)
		Expect(err).NotTo(HaveOccurred())
		Expect(replicas).NotTo(BeEmpty())

		if appvol.Name == "es-data-esnode-0" {
			esData0Nodes = replicas
		} else if appvol.Name == "es-data-esnode-1" {
			esData1Nodes = replicas
		} else if appvol.Name == "es-data-esnode-2" {
			esData2Nodes = replicas
		}
				
	}

	//Replicas of each volume should be in same set of zone
	volrepinzone :=0
	for _, repset:= range volscheck["es-data-esnode-0"] {
		// for each node in the zone, check replica count should be one
		repfound0 :=0
		repfound1 :=0
		repfound2 :=0
		for _, mnode := range repset {
			for _,rnode := range esData0Nodes { 			 
				if rnode == mnode {
					repfound0++
				}
			}

			for _,rnode := range esData1Nodes { 			 
				if rnode == mnode {
					repfound1++
				}
			}
			for _,rnode := range esData2Nodes { 			 
				if rnode == mnode {
					repfound2++
				}
			}
		}
		if (repfound0 ==3  && repfound1 ==3  && repfound2==3) {
			volrepinzone =1
		}
	}

	Expect(volrepinzone).To(Equal(1), fmt.Sprintf("Due to volume replica affinity replicas of volumes es-data-esnode-0: %v ,  es-data-esnode-1:%v  & es-data-esnode2: %v should appear in same zone",esData0Nodes,esData1Nodes, esData2Nodes ) )
}


//StorageClass placement_strategy mapping
func (v *vpscase22) GetScStrategyMap() map[string] string{
	return map[string] string {"placement-1":"placement-1", "placement-2":"placement-1", "placement-3":""}
}

func (v *vpscase22) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-1
spec:
  volumeAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
    matchExpressions:
      - key: app
        operator: In
        values:
          - "elastic"
  replicaAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone`
	return vpsSpec
}

func (v *vpscase22) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}




//#---- Case 23 ---- T955476
//Replica & Volume Affinity & Anti Affinity 
type vpscase23 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase23) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "middleast", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "central", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "south", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node7lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node8lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl,node5lbl, node6lbl,node7lbl,node8lbl)
	return lbldata,1
}

func (v *vpscase23) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

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
	volnodelist["mysql-data"]["rnodes"] = []string{}
	volnodelist["mysql-data-seq"]["pnodes"] = []string{}
	volnodelist["mysql-data-seq"]["nnodes"] = []string{}
	volnodelist["mysql-data-seq"]["rnodes"] = []string{}

	//Create a list of nodes in px_zone east and north,
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zoneeast"] {
		volnodelist["mysql-data"]["rnodes"] = append(volnodelist["mysql-data"]["rnodes"], lnode)
		volnodelist["mysql-data-seq"]["rnodes"] = append(volnodelist["mysql-data-seq"]["rnodes"], lnode)
	}

	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonenorth"] {
		volnodelist["mysql-data"]["rnodes1"] = append(volnodelist["mysql-data"]["rnodes1"], lnode)
		volnodelist["mysql-data-seq"]["rnodes1"] = append(volnodelist["mysql-data-seq"]["rnodes1"], lnode)
	}
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonewest"] {
		volnodelist["mysql-data"]["rnodes2"] = append(volnodelist["mysql-data"]["rnodes2"], lnode)
		volnodelist["mysql-data-seq"]["rnodes2"] = append(volnodelist["mysql-data-seq"]["rnodes2"], lnode)
	}
	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/px_zonesouth"] {
		volnodelist["mysql-data"]["rnodes3"] = append(volnodelist["mysql-data"]["rnodes3"], lnode)
		volnodelist["mysql-data-seq"]["rnodes3"] = append(volnodelist["mysql-data-seq"]["rnodes3"], lnode)
	}
	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase23) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	logrus.Infof("Case 21 T867640 Verify replica anti-affinity and volume anti-affinity topology keys with volume labels ")	

	var mysqlDataReplNodes []string
	var mysqlDataSeqReplNodes []string

	for _, appvol := range appVolumes {

		replicas, err := Inst().V.GetReplicaSetNodes(appvol)
		Expect(err).NotTo(HaveOccurred())
		Expect(replicas).NotTo(BeEmpty())

		if appvol.Name == "mysql-data" {
			mysqlDataReplNodes = replicas
		} else if appvol.Name == "mysql-data-seq" {
			mysqlDataSeqReplNodes = replicas
		}
				
	}

	//Replicas should not be in same zone, replicas of different volume should 
	//be in same zone
	volrepinzone :=0
	for _, repset:= range volscheck["mysql-data"] {
		// for each node in the zone, check replica count should be one
		repfoundseq :=0
		repfound :=0
		for _, mnode := range repset {
			for _,rnode := range mysqlDataSeqReplNodes { 			 
				if rnode == mnode {
					repfoundseq++
				}
			}

			for _,rnode := range mysqlDataReplNodes { 			 
				if rnode == mnode {
					repfound++
				}
			}
		}
		if (repfound >=1  && repfoundseq >=1) || repfound > 1 || repfoundseq > 1   {
			volrepinzone =1
		}
	}

	Expect(volrepinzone).To(Equal(0), fmt.Sprintf("Due to volume & replicas anti-affinity replicas of volumes mysql-data: %v &  mysql-data-seq:%v should not appear in same zone",mysqlDataReplNodes,mysqlDataSeqReplNodes ) )
}


//StorageClass placement_strategy mapping
func (v *vpscase23) GetScStrategyMap() map[string] string{
	return map[string] string {"placement-1":"placement-1", "placement-2":"placement-1", "placement-3":""}
}

func (v *vpscase23) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: vps-data-rule
spec:
  replicaAffinity:
  - enforcement: required
    matchExpressions:
      - key: media_type
        operator: In
        values:
          - "SSD"
  - enforcement: required
    matchExpressions:
      - key: ioprofile
        operator: NotIn
        values:
          - "REGULAR"

---
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: vps-seq-rule
spec:
  replicaAntiAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
  replicaAffinity:
  - enforcement: required
    matchExpressions:
      - key: ioprofile
        operator: In
        values:
          - "REGULAR"
---
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: vps-aggr-rule
spec:
  volumeAntiAffinity:
  - enforcement: required
    matchExpressions:
      - key: voltype
        operator: In
        values:
          - "mysql-data"
  replicaAffinity:
  - enforcement: required
    matchExpressions:
      - key: media_type
        operator: In
        values:
          - "SSD"
  volumeAffinity:
  - enforcement: required
    matchExpressions:
      - key: voltype
        operator: In
        values:
          - "mysql-seq"`
	return vpsSpec
}

func (v *vpscase23) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}


//#---- Case 24 ----T864240  Verify Replica Anti-Affinity with topology keys
// With few nodes labeled 
type vpscase24 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase24) GetLabels() ([]labelDict,int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/zone": "east"}//, "failure-domain.beta.kubernetes.io/region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/zone": "east"}//, "failure-domain.beta.kubernetes.io/region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/zone": "west"}//, "failure-domain.beta.kubernetes.io/region": "asia"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/zone": "west"}//, "failure-domain.beta.kubernetes.io/region": "asia"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/zone": "south"}//, "failure-domain.beta.kubernetes.io/region": "eu"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/zone": "south"}//, "failure-domain.beta.kubernetes.io/region": "eu"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl,node5lbl, node6lbl)
	return lbldata,1
}

func (v *vpscase24) GetPvcNodeLabels(lblnodes map[string][]string) map[string]map[string][]string {

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

	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/zoneeast"] {
		volnodelist["mysql-data"]["rnodes1"] = append(volnodelist["mysql-data"]["rnodes1"], lnode)
		volnodelist["mysql-data-seq"]["rnodes1"] = append(volnodelist["mysql-data-seq"]["rnodes1"], lnode)
		// Add nodes for aggr in set-2 for validation simplification
		volnodelist["mysql-data-aggr"]["rnodes1"] = append(volnodelist["mysql-data-aggr"]["rnodes1"], lnode)
	}

	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/zonewest"] {
		volnodelist["mysql-data"]["rnodes2"] = append(volnodelist["mysql-data"]["rnodes2"], lnode)
		volnodelist["mysql-data-seq"]["rnodes2"] = append(volnodelist["mysql-data-seq"]["rnodes2"], lnode)
		// Aggr replicas are spread across all nodes
		volnodelist["mysql-data-aggr"]["rnodes2"] = append(volnodelist["mysql-data-aggr"]["rnodes2"], lnode)
	}

	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/zonesouth"] {
		volnodelist["mysql-data"]["rnodes3"] = append(volnodelist["mysql-data"]["rnodes3"], lnode)
		volnodelist["mysql-data-seq"]["rnodes3"] = append(volnodelist["mysql-data-seq"]["rnodes3"], lnode)
		// Aggr replicas are spread across all nodes
		volnodelist["mysql-data-aggr"]["rnodes3"] = append(volnodelist["mysql-data-aggr"]["rnodes3"], lnode)
	}

	for _, lnode := range lblnodes["failure-domain.beta.kubernetes.io/zonedefault"] {
		volnodelist["mysql-data"]["rnodes4"] = append(volnodelist["mysql-data"]["rnodes4"], lnode)
		volnodelist["mysql-data-seq"]["rnodes4"] = append(volnodelist["mysql-data-seq"]["rnodes4"], lnode)
		// Aggr replicas are spread across all nodes
		volnodelist["mysql-data-aggr"]["rnodes4"] = append(volnodelist["mysql-data-aggr"]["rnodes4"], lnode)
	}
	return volnodelist
}

/*
 * 1. Each rule template, will provide the expected output
 */

func (v *vpscase24) Validate(appVolumes []*volume.Volume, volscheck map[string]map[string][]string) {

	logrus.Debugf("Deployed volumes:%v,  volumes to check for nodes placement %v ",
		appVolumes, volscheck)

	for _, appvol := range appVolumes {

		
		for vol, vnodes := range volscheck {

			if appvol.Name == vol {
				replicaset, err := Inst().V.GetReplicatNodeSets(appvol)
				logrus.Debugf("==Replicas for vol: %s, Volume should have replicas on nodes:%v , Volume replicas are present on nodes :%v ", vol, vnodes, replicaset)
				Expect(err).NotTo(HaveOccurred())
				Expect(replicaset).NotTo(BeEmpty())

				for _,replicas := range replicaset {
					// Must have (required)
					// There are  3 replicas and 4 sets to check in.
					// In the 4 set, the replica should be place in the 3 of the sets.
					// A set cannot containe more than 1 replica
					
					Expect(replicas).NotTo(BeEmpty())

				    totalrepfound :=0
					// Check in set 1
					foundinset := 0
					for _, mnode := range vnodes["rnodes1"] {
						found := ""
						repOnNodeCnt:=0
						// Check whether replica is on the expected set of nodes
						for _, rnode := range  replicas {
							logrus.Debugf("Expected replica to be on Node:%v Replica is on Node:%v", mnode, rnode)
							if mnode == rnode {
								found = rnode
								repOnNodeCnt++
							}
						}

						if  found != "" {	   
							Expect(repOnNodeCnt).Should(BeNumerically("<=", 1))
							foundinset++
						}
					}

					Expect(foundinset).Should(BeNumerically("<=", 1))
					if foundinset ==1 {
						totalrepfound++
					}

					// Check in set 2
					foundinset = 0
					for _, mnode := range vnodes["rnodes2"] {
						found := ""
						repOnNodeCnt:=0
						// Check whether replica is on the expected set of nodes
						for _, rnode := range  replicas {
							logrus.Debugf("Expected replica to be on Node:%v Replica is on Node:%v", mnode, rnode)
							if mnode == rnode {
								found = rnode
								repOnNodeCnt++
							}
						}

						if  found != "" {	   
							Expect(repOnNodeCnt).Should(BeNumerically("<=", 1))
							foundinset++
						}
					}

					Expect(foundinset).Should(BeNumerically("<=", 1))
					if foundinset ==1 {
						totalrepfound++
					}

					// Check in set 3
					foundinset = 0
					for _, mnode := range vnodes["rnodes3"] {
						found := ""
						repOnNodeCnt:=0
						// Check whether replica is on the expected set of nodes
						for _, rnode := range  replicas {
							logrus.Debugf("Expected replica to be on Node:%v Replica is on Node:%v", mnode, rnode)
							if mnode == rnode {
								found = rnode
								repOnNodeCnt++
							}
						}

						if  found != "" {	   
							Expect(repOnNodeCnt).Should(BeNumerically("<=", 1))
							foundinset++
						}
					}

					Expect(foundinset).Should(BeNumerically("<=", 1))
					if foundinset ==1 {
						totalrepfound++
					}

					// Check in set 4
					foundinset = 0
					for _, mnode := range vnodes["rnodes4"] {
						found := ""
						repOnNodeCnt:=0
						// Check whether replica is on the expected set of nodes
						for _, rnode := range  replicas {
							logrus.Debugf("Expected replica to be on Node:%v Replica is on Node:%v", mnode, rnode)
							if mnode == rnode {
								found = rnode
								repOnNodeCnt++
							}
						}

						if  found != "" {	   
							Expect(repOnNodeCnt).Should(BeNumerically("<=", 1))
							foundinset++
						}
					}

					Expect(foundinset).Should(BeNumerically("<=", 1))
					if foundinset ==1 {
						totalrepfound++
					}

					if vol == "mysql-data-seq" || vol =="mysql-data" {
						// These are repl:3 vol
						Expect(totalrepfound).Should(Equal(3))
					} else {
						// These are repl:2 aggr:2 volume
						Expect(totalrepfound).Should(Equal(2))
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
}


//StorageClass placement_strategy mapping
func (v *vpscase24) GetScStrategyMap() map[string] string{
	return map[string] string {"placement-1":"placement-1", "placement-2":"placement-1", "placement-3":"placement-3"}
}

func (v *vpscase24) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-1
spec:
  replicaAntiAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/zone
---
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: placement-3
spec:
  replicaAntiAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/region`
	return vpsSpec
}

func (v *vpscase24) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}




// Test case inits
//

/*
 *  
 *     Replica  Affinity and Anti-Affinity related test cases init
 *
 */
///*
func init() {
	v := &vpscase1{"case1 Replica affinity to node labels", true}
	Register(v.name, v,1)
}

func init() {
	v := &vpscase2{"case2-T863374 Replica Affinity with enforcement=preferred", true}
	Register(v.name, v,1)
}


func init() {
	v := &vpscase3{"case3-T809561 Replica Affinity with  Lt, Gt operators using latency and iops as node labels", true}
	Register(v.name, v,1)
}


func init() {
	v := &vpscase4{"case4-T863792 Replica Affinity with topology keys", true}
	Register(v.name, v,1)
}


func init() {
	v := &vpscase5{"case5-T1052921 Replica Anti-Affinity with topology keys (with all nodes labeled)", true}
	Register(v.name, v,1)
}
//*/

func init() {
	v := &vpscase6{"case6-T809554 Replica Affinity ,Volume creation should fail when VolumePlacementStrategy fails to find enough pools", true}
	Register(v.name, v,5)
}


/*
 *  
 *     Volume  Affinity and Anti-Affinity related test cases init
 *
 */
///*
func init() {
	v := &vpscase7{"case7-T809548 Volume Affinity 'Exists'", true}
	Register(v.name, v,2)
}


func init() {
	v := &vpscase8{"case8-T809548 Volume Affinity 'In'", true}
	Register(v.name, v,2)
}


func init() {
	v := &vpscase9{"case9-T809548 Volume Affinity 'DoesNotExists'", true}
	Register(v.name, v,2)
}


func init() {
	v := &vpscase10{"case10-T809548 Volume Affinity 'NotIn'", true}
	Register(v.name, v,2)
}

// Volume Anti-affinity
func init() {
	v := &vpscase11{"case11-T809549 Volume Anti-Affinity 'Exists'", true}
	Register(v.name, v,2)
}



func init() {
	v := &vpscase12{"case12-T809549 Volume Anti-Affinity 'In'", true}
	Register(v.name, v,2)
}
//*/
/*
func init() {
	v := &vpscase13{"case13-T809549 Volume Anti-Affinity 'DoesNotExists'", true}
	Register(v.name, v,2)
}

func init() {
	v := &vpscase14{"case14-T809549 Volume Anti-Affinity 'NotIn'", true}
	Register(v.name, v,2)
}
*/

///*
func init() {
	v := &vpscase15{"case15-T864665  Volume Affinity with topology key", true}
	Register(v.name, v,2)
}


func init() {
	v := &vpscase16{"case16-T1053359 Volume anti-affinity with topology keys", true}
	Register(v.name, v,2)
}

func init() {
	v := &vpscase17{"case17-T870615  volume anti-affinity multiple rules", true}
	Register(v.name, v,2)
}



/*
 *  
 *     Replicas & Volume  Affinity and Anti-Affinity related test cases init
 *
 */


func init() {
	v := &vpscase18{"case18-T866365 Verify replica and volume affinity topology keys with volume labels", true}
	Register(v.name, v,3)
}

func init() {
	v := &vpscase19{"case19-T866790 replica affinity and volume anti-affinity topology keys with volume labels ", true}
	Register(v.name, v,3)
}


//*/

func init() {
	v := &vpscase20{"case20-T867215 Verify replica anti-affinity and volume affinity topology keys with volume lables ", true}
	Register(v.name, v,3)
}


func init() {
	v := &vpscase21{"case21-T867640 Verify replica anti-affinity and volume anti-affinity topology keys with volume labels", true}
	Register(v.name, v,3)
}



// Volume replica scaling
func init() {
	v := &vpscase22{"case22-T871040 Verify statefulset/deployment scale up/down w.r.t replica and volume affinity rules ", true}
	Register(v.name, v,4)
}

/*
// Pool labeling is pending
func init() {
	v := &vpscase23{"case23-T871040 T955476 Replica & Volume Affinity & Anti Affinity ", true}
	Register(v.name, v)
}


 */


/*Default node labels*/


func init() {
	v := &vpscase24{"case24-T864240  Verify Replica Anti-Affinity with topology keys (with few nodes not set)", true}
	Register(v.name, v,6)
}


