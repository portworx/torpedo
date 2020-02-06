package tests

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

const (
	mediaSsd       = "SSD"
	mediaSata      = "SATA"
	mediaType      = "media_type"
	liops          = "iops"
	llatency       = "latency"
	pxZone         = "failure-domain.beta.kubernetes.io/px_zone"
	pxRegion       = "failure-domain.beta.kubernetes.io/px_region"
	kubeZone       = "failure-domain.beta.kubernetes.io/zone"
	ruleCategories = 6
)

type labelDict map[string]interface{}

//VpsTemplate provides interface for running individual testcases
type VpsTemplate interface {
	// Node label and whether it needs to be set on node remove
	GetLabels() ([]labelDict, int)
	// Vps Spec
	GetSpec() string
	// Clean up
	CleanVps()
}

//VpsRules contains list of all rules for VolumePlacementStrategy
var VpsRules = []map[string]VpsTemplate{}

// Register registers the given vps rule
func Register(name string, d VpsTemplate, cat int) error {

	if _, ok := VpsRules[cat][name]; !ok {
		VpsRules[cat][name] = d
	} else {
		return fmt.Errorf("vps rule: %s is already registered", name)
	}
	return nil
}

// GetVpsRules return the list of vps rules
func GetVpsRules(cat int) map[string]VpsTemplate {
	if cat >= 0 && cat <= ruleCategories {
		return VpsRules[cat]
	}
	return nil

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
	node1lbl := labelDict{mediaType: mediaSsd}
	node2lbl := labelDict{mediaType: mediaSata}
	node3lbl := labelDict{mediaType: mediaSsd}
	node4lbl := labelDict{mediaType: mediaSsd}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl)
	return lbldata, 1
}

func (v *vpscase1) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
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
	node1lbl := labelDict{mediaType: mediaSsd}
	node2lbl := labelDict{mediaType: mediaSata}
	node3lbl := labelDict{mediaType: mediaSsd}
	node4lbl := labelDict{mediaType: mediaSsd}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl)
	return lbldata, 1
}

func (v *vpscase2) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
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
	node1lbl := labelDict{liops: "90", llatency: "50"}
	node2lbl := labelDict{liops: "80", llatency: "40"}
	node3lbl := labelDict{liops: "70", llatency: "30"}
	node4lbl := labelDict{liops: "60", llatency: "20"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl)
	return lbldata, 1
}

func (v *vpscase3) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
spec:
  replicaAffinity:
  - enforcement: required
    matchExpressions:
    - key: iops
      operator: Gt
      values:
      - "60"`
	return vpsSpec
}

func (v *vpscase3) CleanVps() {
	logrus.Infof("Cleanup test case context for: %v", v.name)
}

//#---- Case 3.1 ----T809561: Verify Lt, Gt operators using latency and iops
type vpscase3_1 struct {
	//Case description
	name string
	// Enabled
	enabled bool
}

func (v *vpscase3_1) GetLabels() ([]labelDict, int) {

	lbldata := []labelDict{}
	node1lbl := labelDict{liops: "90", llatency: "50"}
	node2lbl := labelDict{liops: "80", llatency: "40"}
	node3lbl := labelDict{liops: "70", llatency: "30"}
	node4lbl := labelDict{liops: "60", llatency: "20"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl)
	return lbldata, 1
}

func (v *vpscase3_1) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
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

func (v *vpscase3_1) CleanVps() {
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
	node1lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node2lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node3lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node4lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node5lbl := labelDict{pxZone: "west", pxRegion: "usa"}
	node6lbl := labelDict{pxZone: "west", pxRegion: "usa"}
	node7lbl := labelDict{pxZone: "west", pxRegion: "usa"}
	node8lbl := labelDict{pxZone: "west", pxRegion: "usa"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl, node7lbl, node8lbl)
	return lbldata, 1
}

func (v *vpscase4) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
spec:
  replicaAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone`
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
	node1lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node2lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node3lbl := labelDict{pxZone: "west", pxRegion: "asia"}
	node4lbl := labelDict{pxZone: "west", pxRegion: "asia"}
	node5lbl := labelDict{pxZone: "south", pxRegion: "eu"}
	node6lbl := labelDict{pxZone: "south", pxRegion: "eu"}
	node7lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	node8lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl, node7lbl, node8lbl)
	return lbldata, 1
}

func (v *vpscase5) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
spec:
  replicaAntiAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone`
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

func (v *vpscase6) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
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

func (v *vpscase7) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
spec:
  volumeAffinity:
  - enforcement: required
    matchExpressions:
    - key: appvps-{VOL_KEY}
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

func (v *vpscase8) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
spec:
  volumeAffinity:
  - enforcement: required
    matchExpressions:
    - key: appvps-{VOL_KEY}
      operator: In
      values:
      - "{VOL_LABEL}"`
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

func (v *vpscase9) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
spec:
  volumeAffinity:
  - enforcement: required
    matchExpressions:
    - key: appvps-{VOL_KEY}
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

func (v *vpscase10) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
spec:
  volumeAffinity:
  - enforcement: required
    matchExpressions:
    - key: appvps-{VOL_KEY}
      operator: NotIn
      values:
      - "{VOL_LABEL}"`
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

func (v *vpscase11) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
spec:
  volumeAntiAffinity:
  - enforcement: required
    matchExpressions:
    - key: appvps-{VOL_KEY}
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

func (v *vpscase12) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
spec:
  volumeAntiAffinity:
  - enforcement: required
    matchExpressions:
    - key: appvps-{VOL_KEY}
      operator: In
      values:
      - "{VOL_LABEL}"`
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

func (v *vpscase13) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
spec:
  volumeAntiAffinity:
  - enforcement: required
    matchExpressions:
    - key: appvps-{VOL_KEY}
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

func (v *vpscase14) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
spec:
  volumeAntiAffinity:
  - enforcement: required
    matchExpressions:
    - key: appvps-{VOL_KEY}
      operator: NotIn
      values:
      - "{VOL_LABEL}"`
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
	node1lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node2lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node3lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node4lbl := labelDict{pxZone: "south", pxRegion: "usa"}
	node5lbl := labelDict{pxZone: "south", pxRegion: "jp"}
	node6lbl := labelDict{pxZone: "south", pxRegion: "jp"}
	node7lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	node8lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl, node7lbl, node8lbl)
	return lbldata, 1
}

func (v *vpscase15) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
spec:
  volumeAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
    matchExpressions:
      - key: appvps-{VOL_KEY}
        operator: In
        values:
          - "{VOL_LABEL}"
  - enforcement: required
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
	node1lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node2lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node3lbl := labelDict{pxZone: "west", pxRegion: "usa"}
	node4lbl := labelDict{pxZone: "central", pxRegion: "usa"}
	node5lbl := labelDict{pxZone: "middle", pxRegion: "jp"}
	node6lbl := labelDict{pxZone: "south", pxRegion: "jp"}
	node7lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	node8lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl, node7lbl, node8lbl)
	return lbldata, 1
}

func (v *vpscase16) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
spec:
  volumeAntiAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
    matchExpressions:
      - key: appvps-{VOL_KEY}
        operator: In
        values:
          - "{VOL_LABEL}"`
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
	node1lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node2lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node3lbl := labelDict{pxZone: "west", pxRegion: "usa"}
	node4lbl := labelDict{pxZone: "central", pxRegion: "usa"}
	node5lbl := labelDict{pxZone: "middle", pxRegion: "jp"}
	node6lbl := labelDict{pxZone: "south", pxRegion: "jp"}
	node7lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	node8lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl, node7lbl, node8lbl)
	return lbldata, 1
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
      - key: appvps-{VOL_KEY}
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
	node1lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node2lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node3lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node4lbl := labelDict{pxZone: "west", pxRegion: "usa"}
	node5lbl := labelDict{pxZone: "west", pxRegion: "jp"}
	node6lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	node7lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	node8lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl, node7lbl, node8lbl)
	return lbldata, 1
}

func (v *vpscase18) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
spec:
  volumeAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
    matchExpressions:
      - key: appvps-{VOL_KEY}
        operator: In
        values:
          - "{VOL_LABEL}"
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
	node1lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node2lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node3lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node4lbl := labelDict{pxZone: "west", pxRegion: "usa"}
	node5lbl := labelDict{pxZone: "west", pxRegion: "jp"}
	node6lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	node7lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	node8lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl, node7lbl, node8lbl)
	return lbldata, 1
}

func (v *vpscase19) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
spec:
  volumeAntiAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
    matchExpressions:
      - key: appvps-{VOL_KEY}
        operator: In
        values:
          - "{VOL_LABEL}"
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
	node1lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node2lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node3lbl := labelDict{pxZone: "west", pxRegion: "usa"}
	node4lbl := labelDict{pxZone: "west", pxRegion: "usa"}
	node5lbl := labelDict{pxZone: "south", pxRegion: "jp"}
	node6lbl := labelDict{pxZone: "south", pxRegion: "jp"}
	node7lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	node8lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl, node7lbl, node8lbl)
	return lbldata, 1
}

func (v *vpscase20) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
spec:
  volumeAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
    matchExpressions:
      - key: appvps-{VOL_KEY}
        operator: In
        values:
          - "{VOL_LABEL}"
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
	node1lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node2lbl := labelDict{pxZone: "middleast", pxRegion: "usa"}
	node3lbl := labelDict{pxZone: "west", pxRegion: "usa"}
	node4lbl := labelDict{pxZone: "west", pxRegion: "usa"}
	node5lbl := labelDict{pxZone: "central", pxRegion: "jp"}
	node6lbl := labelDict{pxZone: "south", pxRegion: "jp"}
	node7lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	node8lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl, node7lbl, node8lbl)
	return lbldata, 1
}

func (v *vpscase21) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
spec:
  volumeAntiAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
    matchExpressions:
      - key: appvps-{VOL_KEY}
        operator: In
        values:
          - "{VOL_LABEL}"
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
	node1lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node2lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node3lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node4lbl := labelDict{pxZone: "west", pxRegion: "usa"}
	node5lbl := labelDict{pxZone: "central", pxRegion: "jp"}
	node6lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	node7lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	node8lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl, node7lbl, node8lbl)
	return lbldata, 1
}

func (v *vpscase22) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
spec:
  volumeAffinity:
  - enforcement: required
    topologyKey: failure-domain.beta.kubernetes.io/px_zone
    matchExpressions:
      - key: appvps-{VOL_KEY}
        operator: In
        values:
          - "{VOL_LABEL}"
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
	node1lbl := labelDict{pxZone: "east", pxRegion: "usa"}
	node2lbl := labelDict{pxZone: "middleast", pxRegion: "usa"}
	node3lbl := labelDict{pxZone: "west", pxRegion: "usa"}
	node4lbl := labelDict{pxZone: "west", pxRegion: "usa"}
	node5lbl := labelDict{pxZone: "central", pxRegion: "jp"}
	node6lbl := labelDict{pxZone: "south", pxRegion: "jp"}
	node7lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	node8lbl := labelDict{pxZone: "north", pxRegion: "jp"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl, node7lbl, node8lbl)
	return lbldata, 1
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
	node1lbl := labelDict{kubeZone: "east"}  //, "failure-domain.beta.kubernetes.io/region": "usa"}
	node2lbl := labelDict{kubeZone: "east"}  //, "failure-domain.beta.kubernetes.io/region": "usa"}
	node3lbl := labelDict{kubeZone: "west"}  //, "failure-domain.beta.kubernetes.io/region": "asia"}
	node4lbl := labelDict{kubeZone: "west"}  //, "failure-domain.beta.kubernetes.io/region": "asia"}
	node5lbl := labelDict{kubeZone: "south"} //, "failure-domain.beta.kubernetes.io/region": "eu"}
	node6lbl := labelDict{kubeZone: "south"} //, "failure-domain.beta.kubernetes.io/region": "eu"}
	lbldata = append(lbldata, node1lbl, node2lbl, node3lbl, node4lbl, node5lbl, node6lbl)
	return lbldata, 1
}

func (v *vpscase24) GetSpec() string {

	var vpsSpec string
	vpsSpec = `apiVersion: portworx.io/v1beta2
kind: VolumePlacementStrategy
metadata:
  name: {VPS_NAME}
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

// Initialize the VpsRule
func init() {

	// Initialize all category to empty list
	for i := 0; i <= ruleCategories; i++ {
		vpsRule := map[string]VpsTemplate{}
		VpsRules = append(VpsRules, vpsRule)
	}

	v1 := &vpscase1{"case1 Replica affinity to node labels", true}
	Register(v1.name, v1, 1)

	v2 := &vpscase2{"case2-T863374 Replica Affinity with enforcement=preferred", true}
	Register(v2.name, v2, 1)

	v3 := &vpscase3{"case3-T809561 Replica Affinity with  Lt, Gt operators using latency and iops as node labels", true}
	Register(v3.name, v3, 1)

	v3_1 := &vpscase3_1{"case3_1-T809561 Replica Affinity with  Lt, Gt operators using latency and iops as node labels", true}
	Register(v3_1.name, v3_1, 1)
	v4 := &vpscase4{"case4-T863792 Replica Affinity with topology keys", true}
	Register(v4.name, v4, 1)

	v5 := &vpscase5{"case5-T1052921 Replica Anti-Affinity with topology keys (with all nodes labeled)", true}
	Register(v5.name, v5, 1)

	//*/

	v6 := &vpscase6{"case6-T809554 Replica Affinity ,Volume creation should fail when VolumePlacementStrategy fails to find enough pools", true}
	Register(v6.name, v6, 5)

	/*
	 *
	 *     Volume  Affinity and Anti-Affinity related test cases init
	 *
	 */
	///*
	v7 := &vpscase7{"case7-T809548 Volume Affinity 'Exists'", true}
	Register(v7.name, v7, 2)

	v8 := &vpscase8{"case8-T809548 Volume Affinity 'In'", true}
	Register(v8.name, v8, 2)

	v9 := &vpscase9{"case9-T809548 Volume Affinity 'DoesNotExists'", true}
	Register(v9.name, v9, 2)
	v10 := &vpscase10{"case10-T809548 Volume Affinity 'NotIn'", true}
	Register(v10.name, v10, 2)

	// Volume Anti-affinity
	v11 := &vpscase11{"case11-T809549 Volume Anti-Affinity 'Exists'", true}
	Register(v11.name, v11, 2)

	v12 := &vpscase12{"case12-T809549 Volume Anti-Affinity 'In'", true}
	Register(v12.name, v12, 2)

	//*/
	/*
		v13 := &vpscase13{"case13-T809549 Volume Anti-Affinity 'DoesNotExists'", true}
			Register(v13.name, v13,2)

			v14 := &vpscase14{"case14-T809549 Volume Anti-Affinity 'NotIn'", true}
			Register(v14.name, v14,2)
	*/

	///*
	v15 := &vpscase15{"case15-T864665  Volume Affinity with topology key", true}
	Register(v15.name, v15, 2)
	v16 := &vpscase16{"case16-T1053359 Volume anti-affinity with topology keys", true}
	Register(v16.name, v16, 2)

	/*  This case is valid for apps having multiple volumes with different label
	v17 := &vpscase17{"case17-T870615  volume anti-affinity multiple rules", true}
	Register(v17.name, v17, 2)
	*/

	/*
	 *
	 *     Replicas & Volume  Affinity and Anti-Affinity related test cases init
	 *
	 */

	v18 := &vpscase18{"case18-T866365 Verify replica and volume affinity topology keys with volume labels", true}
	Register(v18.name, v18, 3)

	v19 := &vpscase19{"case19-T866790 replica affinity and volume anti-affinity topology keys with volume labels ", true}
	Register(v19.name, v19, 3)

	//*/

	v20 := &vpscase20{"case20-T867215 Verify replica anti-affinity and volume affinity topology keys with volume lables ", true}
	Register(v20.name, v20, 3)

	v21 := &vpscase21{"case21-T867640 Verify replica anti-affinity and volume anti-affinity topology keys with volume labels", true}
	Register(v21.name, v21, 3)

	// Volume replica scaling
	v22 := &vpscase22{"case22-T871040 Verify statefulset/deployment scale up/down w.r.t replica and volume affinity rules ", true}
	Register(v22.name, v22, 4)

	/*
	   // Pool labeling is pending
	   v23 := &vpscase23{"case23-T871040 T955476 Replica & Volume Affinity & Anti Affinity ", true}
	   	Register(v23.name, v23)


	*/

	/*Default node labels*/

	v24 := &vpscase24{"case24-T864240  Verify Replica Anti-Affinity with topology keys (with few nodes not set)", true}
	Register(v24.name, v24, 6)
}
