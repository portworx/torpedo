package gke

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/libopenstorage/cloudops"
	"github.com/libopenstorage/cloudops/gce"
	"github.com/pure-px/torpedo/drivers/node"
	"github.com/pure-px/torpedo/drivers/node/ssh"
	"github.com/pure-px/torpedo/drivers/scheduler"
	kube "github.com/pure-px/torpedo/drivers/scheduler/k8s"
	"github.com/pure-px/torpedo/pkg/log"
)

const (
	// SchedName is the name of the gke driver
	SchedName = "gke"

	defaultGkeUpgradeTimeout = 90 * time.Minute
)

type Gke struct {
	ssh.SSH
	kube.K8s
	ops           cloudops.Ops
	instanceGroup string
	nodePoolList  []string
}

func (g *Gke) String() string {
	return SchedName
}

func init() {
	g := &Gke{}
	scheduler.Register(SchedName, g)
}

func (g *Gke) Init(schedOpts scheduler.InitOptions) error {
	ops, err := gce.NewClient()
	if err != nil {
		return err
	}
	g.ops = ops

	err = g.K8s.Init(schedOpts)
	if err != nil {
		return err
	}

	return nil
}

// UpgradeScheduler performs GKE cluster upgrade to a specified version
func (g *Gke) UpgradeScheduler(version string) error {
	log.Infof("Starting GKE cluster upgrade to [%s]", version)

	instanceGroup := os.Getenv("INSTANCE_GROUP")
	if len(instanceGroup) != 0 {
		g.nodePoolList = append(g.nodePoolList, instanceGroup)
	} else {
		g.nodePoolList = append(g.nodePoolList, "default-pool")
	}

	// If NODE_POOL_LIST is passed, will use this list instead of INSTANCE_GROUP or default-pool to perform upgrades
	nodePoolList := os.Getenv("NODE_POOL_LIST")
	if len(nodePoolList) != 0 {
		g.nodePoolList = strings.Split(nodePoolList, ",")
		log.Infof("Got list of node pools to upgrade %v from NODE_POOL_LIST env var passed to torpedo", g.nodePoolList)
	}

	// Upgrade GKE Control Plane version
	if err := g.upgradeGkeControlPlaneVersion(version, defaultGkeUpgradeTimeout); err != nil {
		return err
	}

	// Update GKE node pool(s) upgrade strategy
	upgradeStrategy := os.Getenv("GKE_UPGRADE_STRATEGY")
	surgeSetting := os.Getenv("GKE_SURGE_VALUE")
	if upgradeStrategy == "" {
		upgradeStrategy = "surge"
	}
	if surgeSetting == "" {
		surgeSetting = "default"
	}

	if err := g.updateGkeNodeGroupUpgradeStrategy(g.nodePoolList,
		upgradeStrategy, defaultGkeUpgradeTimeout, surgeSetting); err != nil {
		return err
	}

	// Upgrade GKE node pool(s) version
	if err := g.upgradeGkeNodeGroupVersion(g.nodePoolList, version, defaultGkeUpgradeTimeout); err != nil {
		return err
	}

	log.Infof("Successfully upgraded GKE cluster to [%s]", version)
	return nil
}

func (g *Gke) SetASGClusterSize(perZoneCount int64, timeout time.Duration) error {
	instanceGroup := os.Getenv("INSTANCE_GROUP")
	if len(instanceGroup) != 0 {
		g.instanceGroup = instanceGroup
	} else {
		g.instanceGroup = "default-pool"
	}
	// GCP SDK requires per zone cluster size
	if err := g.ops.SetInstanceGroupSize(g.instanceGroup, perZoneCount, timeout); err != nil {
		return fmt.Errorf("failed to set size of node pool %s. Error: %v", g.instanceGroup, err)
	}

	return nil
}

func (g *Gke) GetASGClusterSize() (int64, error) {
	instanceGroup := os.Getenv("INSTANCE_GROUP")
	if len(instanceGroup) != 0 {
		g.instanceGroup = instanceGroup
	} else {
		g.instanceGroup = "default-pool"
	}
	nodeCount, err := g.ops.GetInstanceGroupSize(g.instanceGroup)
	if err != nil {
		return 0, fmt.Errorf("failed to get size of GKE Node Pool [%s] Err: %v", g.instanceGroup, err)
	}

	return nodeCount, nil
}

// upgradeGkeControlPlaneVersion upgrades GKE Control Plane to a specified version
func (g *Gke) upgradeGkeControlPlaneVersion(version string, timeout time.Duration) error {
	log.Infof("Upgrade GKE Control Plane version to [%s]..", version)
	if err := g.ops.SetClusterVersion(version, timeout); err != nil {
		return fmt.Errorf("failed to set version for GKE Control Plane to [%s], Err: %v", version, err)
	}
	log.Infof("GKE Control Plane version was successfully set to [%s]", version)
	return nil
}

// upgradeGkeNodeGroupVersion upgrades GKE node pool(s) to a specified version
func (g *Gke) upgradeGkeNodeGroupVersion(nodePoolList []string, version string, timeout time.Duration) error {
	log.Infof("Preparing to upgrade [%d] GKE node pool(s) %v..", len(nodePoolList), nodePoolList)
	for _, nodePool := range nodePoolList {
		log.Infof("Upgrade GKE node pool [%s] version to [%s]", nodePool, version)
		if err := g.ops.SetInstanceGroupVersion(nodePool, version, timeout); err != nil {
			return fmt.Errorf("failed to set version for GKE node pool [%s] to [%s], Err: %v", nodePool, version, err)
		}
		log.Infof("GKE node pool [%s] version was successfully set to [%s]", nodePool, version)
	}
	log.Infof("Successfully upgraded [%d] GKE node pool(s) %v", len(nodePoolList), nodePoolList)
	return nil
}

// updateGkeNodeGroupUpgradeStrategy updates GKE node pool(s) upgrade strategy
func (g *Gke) updateGkeNodeGroupUpgradeStrategy(nodePoolList []string, upgradeStrategy string, timeout time.Duration, surgeSetting string) error {
	log.Infof("Preparing to update [%d] GKE node pool(s) %v..", len(nodePoolList), nodePoolList)
	for _, nodePool := range nodePoolList {
		log.Infof("Updating GKE node pool [%s] upgrade strategy to [%s]", nodePool, upgradeStrategy)
		if err := g.ops.SetInstanceUpgradeStrategy(nodePool, upgradeStrategy, timeout, surgeSetting); err != nil {
			return fmt.Errorf("failed to set upgrade strategy for GKE Node Group [%s] to [%s], Err: %v", nodePool, upgradeStrategy, err)
		}
		log.Infof("GKE node pool [%s] upgrade strategy was successfully set to [%s]", nodePool, upgradeStrategy)
	}
	log.Infof("Successfully updated [%d] GKE node pool(s) %v", len(nodePoolList), nodePoolList)
	return nil
}

func (g *Gke) DeleteNode(node node.Node) error {
	if err := g.ops.DeleteInstance(node.Name, node.Zone, 10*time.Minute); err != nil {
		return err
	}
	return nil
}

func (g *Gke) GetZones() ([]string, error) {
	storageDriverNodes := node.GetStorageDriverNodes()
	nZones := make(map[string]bool)
	for _, sNode := range storageDriverNodes {
		if _, ok := nZones[sNode.Zone]; !ok {
			nZones[sNode.Zone] = true
		}

	}
	asgZones := make([]string, 0)
	for k := range nZones {
		asgZones = append(asgZones, k)
	}
	return asgZones, nil
}
