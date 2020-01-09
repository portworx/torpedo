package tests

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

const (
	mediaSsd  = "SSD"
	mediaSata = "SATA"
)

type labelDict map[string]interface{}

type vpsTemplate interface {
	// Node label and whether it needs to be set on node remove
	GetLabels() ([]labelDict, int)
	// Get StorageClass placement_strategy
	GetScStrategyMap() map[string]string
	// Vps Spec
	GetSpec() string
	// Clean up
	CleanVps()
}

var (
	vpsRulesReplica      = make(map[string]vpsTemplate)
	vpsRulesVolume       = make(map[string]vpsTemplate)
	vpsRulesMix          = make(map[string]vpsTemplate)
	vpsRulesMixScale     = make(map[string]vpsTemplate)
	vpsRulesPending      = make(map[string]vpsTemplate)
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
	} else if cat == 2 {
		return vpsRulesVolume
	} else if cat == 3 {
		return vpsRulesMix
	} else if cat == 4 {
		return vpsRulesMixScale
	} else if cat == 5 {
		return vpsRulesPending
	} else if cat == 6 {
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
func (v *vpscase1) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"media_type": mediaSsd, "vps_test": "test"}
	node2lbl := labelDict{"media_type": mediaSata, "vps_test": "test"}
	node3lbl := labelDict{"media_type": mediaSsd, "vps_test": "test"}
	node4lbl := labelDict{"media_type": mediaSsd, "vps_test": "test"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl)
	return lbldata, 1
}

func (v *vpscase1) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-1", "placement-2": "placement-2", "placement-3": ""}
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

func (v *vpscase2) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"media_type": mediaSsd, "vps_test": "test"}
	node2lbl := labelDict{"media_type": mediaSata, "vps_test": "test"}
	node3lbl := labelDict{"media_type": mediaSsd, "vps_test": "test"}
	node4lbl := labelDict{"media_type": mediaSsd, "vps_test": "test"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl)
	return lbldata, 1
}

//StorageClass placement_strategy mapping
func (v *vpscase2) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-1", "placement-2": "placement-2", "placement-3": ""}
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

func (v *vpscase3) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"iops": "90", "latency": "50"}
	node2lbl := labelDict{"iops": "80", "latency": "40"}
	node3lbl := labelDict{"iops": "70", "latency": "30"}
	node4lbl := labelDict{"iops": "60", "latency": "20"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl)
	return lbldata, 1
}

//StorageClass placement_strategy mapping
func (v *vpscase3) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-1", "placement-2": "placement-2", "placement-3": ""}
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

func (v *vpscase4) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node7lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node8lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl, node7lbl, node8lbl)
	return lbldata, 1
}

//StorageClass placement_strategy mapping
func (v *vpscase4) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-1", "placement-2": "placement-1", "placement-3": "placement-3"}
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

func (v *vpscase5) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "asia"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "asia"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "south", "failure-domain.beta.kubernetes.io/px_region": "eu"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "south", "failure-domain.beta.kubernetes.io/px_region": "eu"}
	node7lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node8lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl, node7lbl, node8lbl)
	return lbldata, 1
}

//StorageClass placement_strategy mapping
func (v *vpscase5) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-1", "placement-2": "placement-1", "placement-3": "placement-3"}
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

func (v *vpscase6) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	return lbldata, 0
}

//StorageClass placement_strategy mapping
func (v *vpscase6) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-1", "placement-2": "placement-1", "placement-3": "placement-1"}
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

func (v *vpscase7) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	return lbldata, 0
}

//StorageClass placement_strategy mapping
func (v *vpscase7) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-2", "placement-2": "placement-2", "placement-3": ""}
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
      operator: Exists
      values:
      - ""`
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

func (v *vpscase8) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	return lbldata, 0
}

//StorageClass placement_strategy mapping
func (v *vpscase8) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-2", "placement-2": "placement-2", "placement-3": ""}
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

func (v *vpscase9) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	return lbldata, 0
}

//StorageClass placement_strategy mapping
func (v *vpscase9) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-2", "placement-2": "placement-2", "placement-3": ""}
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
      operator: DoesNotExist
      values:
      - ""`
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

func (v *vpscase10) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	return lbldata, 0
}

//StorageClass placement_strategy mapping
func (v *vpscase10) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-2", "placement-2": "placement-2", "placement-3": ""}
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

func (v *vpscase11) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	return lbldata, 0
}

//StorageClass placement_strategy mapping
func (v *vpscase11) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-2", "placement-2": "placement-2", "placement-3": ""}
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
      operator: Exists
      values:
      - ""`
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

func (v *vpscase12) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	return lbldata, 0
}

//StorageClass placement_strategy mapping
func (v *vpscase12) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-2", "placement-2": "placement-2", "placement-3": ""}
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

func (v *vpscase13) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	return lbldata, 0
}


//StorageClass placement_strategy mapping
func (v *vpscase13) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-2", "placement-2": "placement-2", "placement-3": ""}
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
      operator: DoesNotExist
      values:
      - ""`
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

func (v *vpscase14) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	return lbldata, 0
}

//StorageClass placement_strategy mapping
func (v *vpscase14) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-2", "placement-2": "placement-2", "placement-3": ""}
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

func (v *vpscase15) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "south", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "south", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "south", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node7lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node8lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl, node7lbl, node8lbl)
	return lbldata, 1
}

//StorageClass placement_strategy mapping
func (v *vpscase15) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-2", "placement-2": "placement-2", "placement-3": "placement-3"}
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

func (v *vpscase16) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "central", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "middle", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "south", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node7lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node8lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl, node7lbl, node8lbl)
	return lbldata, 1
}

//StorageClass placement_strategy mapping
func (v *vpscase16) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-1", "placement-2": "placement-2", "placement-3": ""}
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

func (v *vpscase17) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "central", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "middle", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "south", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node7lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node8lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl, node7lbl, node8lbl)
	return lbldata, 1
}

//StorageClass placement_strategy mapping
func (v *vpscase17) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-1", "placement-2": "placement-2", "placement-3": ""}
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

func (v *vpscase18) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node7lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node8lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl, node7lbl, node8lbl)
	return lbldata, 1
}

//StorageClass placement_strategy mapping
func (v *vpscase18) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-1", "placement-2": "placement-2", "placement-3": ""}
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

func (v *vpscase19) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node7lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node8lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl, node7lbl, node8lbl)
	return lbldata, 1
}

//StorageClass placement_strategy mapping
func (v *vpscase19) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-1", "placement-2": "placement-2", "placement-3": ""}
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

func (v *vpscase20) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "south", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "south", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node7lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node8lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl, node7lbl, node8lbl)
	return lbldata, 1
}

//StorageClass placement_strategy mapping
func (v *vpscase20) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-1", "placement-2": "placement-2", "placement-3": ""}
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

func (v *vpscase21) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "middleast", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "central", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "south", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node7lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node8lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl, node7lbl, node8lbl)
	return lbldata, 1
}

//StorageClass placement_strategy mapping
func (v *vpscase21) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-1", "placement-2": "placement-1", "placement-3": ""}
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

func (v *vpscase22) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "central", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node7lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node8lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl, node7lbl, node8lbl)
	return lbldata, 1
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

	for _, appvol := range appVolumes {

		replicas, err := Inst().V.GetReplicaSetNodes(appvol)
		Expect(err).NotTo(HaveOccurred())
		Expect(replicas).NotTo(BeEmpty())

		esDataNodes[appvol.Name] = replicas
	}

	//Replicas of each volume should be in same set of zone
	volrepinzone := 0

	// Nodes with same label and value
	for _, repset := range volscheck["es-data-esnode-0"] {
		// for each node in the zone, check replica count should be one
		repcount := map[string]int{}
		volrepnotinzone := 0
		for _, mnode := range repset {
			//For each volume
			for key, appvol := range esDataNodes {
				//For each replica nodes of a volume, check with zone nodes
				for _, rnode := range appvol {
					if rnode == mnode {
						repcount[key] += 1
					}
				}

			}
		}

		for _, val := range repcount {
			if repcount["es-data-esnode-0"] == 3 && val != 3 {
				volrepnotinzone = 1
				break
			}
		}
		if repcount["es-data-esnode-0"] == 3 && volrepnotinzone == 0 {
			volrepinzone = 1
		}
	}

	Expect(volrepinzone).To(Equal(1), fmt.Sprintf("Due to volume replica affinity replicas of volumes es-data-esnodes: %v , should appear in same zone", esDataNodes))
}

//StorageClass placement_strategy mapping
func (v *vpscase22) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-1", "placement-2": "placement-1", "placement-3": ""}
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

func (v *vpscase23) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "east", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "middleast", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "west", "failure-domain.beta.kubernetes.io/px_region": "usa"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "central", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "south", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node7lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	node8lbl := labelDict{"failure-domain.beta.kubernetes.io/px_zone": "north", "failure-domain.beta.kubernetes.io/px_region": "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl, node7lbl, node8lbl)
	return lbldata, 1
}

//StorageClass placement_strategy mapping
func (v *vpscase23) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-1", "placement-2": "placement-1", "placement-3": ""}
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

func (v *vpscase24) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{"failure-domain.beta.kubernetes.io/zone": "east"}  //, "failure-domain.beta.kubernetes.io/region": "usa"}
	node2lbl := labelDict{"failure-domain.beta.kubernetes.io/zone": "east"}  //, "failure-domain.beta.kubernetes.io/region": "usa"}
	node3lbl := labelDict{"failure-domain.beta.kubernetes.io/zone": "west"}  //, "failure-domain.beta.kubernetes.io/region": "asia"}
	node4lbl := labelDict{"failure-domain.beta.kubernetes.io/zone": "west"}  //, "failure-domain.beta.kubernetes.io/region": "asia"}
	node5lbl := labelDict{"failure-domain.beta.kubernetes.io/zone": "south"} //, "failure-domain.beta.kubernetes.io/region": "eu"}
	node6lbl := labelDict{"failure-domain.beta.kubernetes.io/zone": "south"} //, "failure-domain.beta.kubernetes.io/region": "eu"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl)
	return lbldata, 1
}

//StorageClass placement_strategy mapping
func (v *vpscase24) GetScStrategyMap() map[string]string {
	return map[string]string{"placement-1": "placement-1", "placement-2": "placement-1", "placement-3": ""}
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
    topologyKey: failure-domain.beta.kubernetes.io/zone`
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
	Register(v.name, v, 1)
}

func init() {
	v := &vpscase2{"case2-T863374 Replica Affinity with enforcement=preferred", true}
	Register(v.name, v, 1)
}

func init() {
	v := &vpscase3{"case3-T809561 Replica Affinity with  Lt, Gt operators using latency and iops as node labels", true}
	Register(v.name, v, 1)
}

func init() {
	v := &vpscase4{"case4-T863792 Replica Affinity with topology keys", true}
	Register(v.name, v, 1)
}

func init() {
	v := &vpscase5{"case5-T1052921 Replica Anti-Affinity with topology keys (with all nodes labeled)", true}
	Register(v.name, v, 1)
}

//*/

func init() {
	v := &vpscase6{"case6-T809554 Replica Affinity ,Volume creation should fail when VolumePlacementStrategy fails to find enough pools", true}
	Register(v.name, v, 5)
}

/*
 *
 *     Volume  Affinity and Anti-Affinity related test cases init
 *
 */
///*
func init() {
	v := &vpscase7{"case7-T809548 Volume Affinity 'Exists'", true}
	Register(v.name, v, 2)
}

func init() {
	v := &vpscase8{"case8-T809548 Volume Affinity 'In'", true}
	Register(v.name, v, 2)
}

func init() {
	v := &vpscase9{"case9-T809548 Volume Affinity 'DoesNotExists'", true}
	Register(v.name, v, 2)
}

func init() {
	v := &vpscase10{"case10-T809548 Volume Affinity 'NotIn'", true}
	Register(v.name, v, 2)
}

// Volume Anti-affinity
func init() {
	v := &vpscase11{"case11-T809549 Volume Anti-Affinity 'Exists'", true}
	Register(v.name, v, 2)
}

func init() {
	v := &vpscase12{"case12-T809549 Volume Anti-Affinity 'In'", true}
	Register(v.name, v, 2)
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
	Register(v.name, v, 2)
}

func init() {
	v := &vpscase16{"case16-T1053359 Volume anti-affinity with topology keys", true}
	Register(v.name, v, 2)
}

func init() {
	v := &vpscase17{"case17-T870615  volume anti-affinity multiple rules", true}
	Register(v.name, v, 2)
}

/*
 *
 *     Replicas & Volume  Affinity and Anti-Affinity related test cases init
 *
 */

func init() {
	v := &vpscase18{"case18-T866365 Verify replica and volume affinity topology keys with volume labels", true}
	Register(v.name, v, 3)
}

func init() {
	v := &vpscase19{"case19-T866790 replica affinity and volume anti-affinity topology keys with volume labels ", true}
	Register(v.name, v, 3)
}

//*/

func init() {
	v := &vpscase20{"case20-T867215 Verify replica anti-affinity and volume affinity topology keys with volume lables ", true}
	Register(v.name, v, 3)
}

func init() {
	v := &vpscase21{"case21-T867640 Verify replica anti-affinity and volume anti-affinity topology keys with volume labels", true}
	Register(v.name, v, 3)
}

// Volume replica scaling
func init() {
	v := &vpscase22{"case22-T871040 Verify statefulset/deployment scale up/down w.r.t replica and volume affinity rules ", true}
	Register(v.name, v, 4)
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
	Register(v.name, v, 6)
}
