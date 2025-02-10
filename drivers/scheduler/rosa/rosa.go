package rosa

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	optest "github.com/pure-px/px-operator/pkg/util/test"
	opnshift "github.com/pure-px/sched-ops/k8s/openshift"
	"github.com/pure-px/sched-ops/task"
	"github.com/pure-px/torpedo/drivers/scheduler"
	"github.com/pure-px/torpedo/drivers/scheduler/k8s"
	"github.com/pure-px/torpedo/drivers/scheduler/openshift"
	"github.com/pure-px/torpedo/pkg/log"
	"github.com/pure-px/torpedo/pkg/osutils"
	"golang.org/x/sync/errgroup"
)

const (
	SchedName                   = "rosa"
	IsHCP                       = false
	OpenshiftMirror             = "https://mirror.openshift.com/pub/openshift-v4/clients/rosa/latest/rosa-linux.tar.gz"
	rosaCli                     = "rosa"
	defaultCmdTimeout           = 15 * time.Minute
	defaultCmdRetry             = 45 * time.Second
	defaultSleepInterval        = 300 * time.Second
	defaultUpgradeTimeout       = 4 * time.Hour
	defaultUpgradeRetryInterval = 5 * time.Minute
)

var (
	k8sOpenshift = opnshift.Instance()
	versionReg   = regexp.MustCompile(`^(stable|candidate|fast)-(\d\.\d+)?$`)
)

type rosa struct {
	k8s.K8s
	rosaVersion     string
	rosaClusterName string
}

type ClusterDetails struct {
	OpenshiftVersion string `json:"openshift_version"`
	Version          struct {
		ID                 string   `json:"id"`
		AvailableUpgrades  []string `json:"available_upgrades"`
		ChannelGroup       string   `json:"channel_group"`
		EndOfLifeTimestamp string   `json:"end_of_life_timestamp"`
	} `json:"version"`
}

type NodePool struct {
	ID               string `json:"id"`
	AvailabilityZone string `json:"availability_zone"`
	Replicas         int    `json:"replicas"`
	AWSNodePool      struct {
		InstanceType string `json:"instance_type"`
		RootVolume   struct {
			Size int `json:"size"`
		} `json:"root_volume"`
	} `json:"aws_node_pool"`
	AutoRepair bool   `json:"auto_repair"`
	Subnet     string `json:"subnet"`
	Version    struct {
		ID                string   `json:"id"`
		AvailableUpgrades []string `json:"available_upgrades"`
	} `json:"version"`
}

type NodePoolVersion struct {
	ID      string `json:"id"`
	Version struct {
		ID                string   `json:"id"`
		AvailableUpgrades []string `json:"available_upgrades"`
		RawID             string   `json:"raw_id"`
	} `json:"version"`

	ManagementUpgrade struct {
		Kind           string `json:"kind"`
		MaxSurge       string `json:"max_surge"`
		MaxUnavailable string `json:"max_unavailable"`
		Type           string `json:"type"`
	} `json:"management_upgrade"`

	NodeDrainGracePeriod struct {
		Unit  string `json:"unit"`
		Value int    `json:"value"`
	} `json:"node_drain_grace_period"`

	ScheduledUpgrade struct {
		State   string `json:"state"`
		Version string `json:"version"`
	} `json:"scheduledUpgrade"`
}

// String returns the string name of this driver.
func (r *rosa) String() string {
	return SchedName
}

func init() {
	r := &rosa{}
	err := scheduler.Register(SchedName, r)
	if err != nil {
		return
	}
}

// UpgradeScheduler upgrades the ROSA scheduler to the given version
func (r *rosa) UpgradeScheduler(version string) error {
	// Download ROSA client
	if err := r.downloadROSAClient(); err != nil {
		return err
	}

	// Get Openshift version
	ocpVersion, err := optest.GetOpenshiftVersion()
	if err != nil {
		return fmt.Errorf("failed to get Openshift version, Err: %v", err)
	}
	r.rosaVersion = ocpVersion

	clientVersion, err := openshift.GetClientVersion()
	if err != nil {
		return err
	}

	upgradeVersion := version
	if versionReg.MatchString(version) {
		upgradeVersion = clientVersion
	}

	r.rosaClusterName = os.Getenv("CLUSTER_ID")
	isHCP, err := strconv.ParseBool(os.Getenv("IS_HCP"))
	if err != nil {
		log.Warnf("failed to parse IS_HCP: %v", err)
		isHCP = IsHCP
	}
	log.Debugf("Cluster ID: %s, ROSA Version: %s, Upgrade Version: %s, IsHCP: %v", r.rosaClusterName, r.rosaVersion, upgradeVersion, isHCP)

	if upgradeVersion, err = r.getAvailableUpgradeVersions(); err != nil {
		return err
	}

	nodePools, err := r.listNodePool()
	if err != nil {
		return fmt.Errorf("failed to list ROSA machine pools: %v", err)
	}
	for _, nodePool := range nodePools {
		if err = r.updateMaxUnavailableWorkerMCP(nodePool.ID); err != nil {
			return err
		}
	}

	if isHCP {
		if err = r.startHCPUpgrade(upgradeVersion); err != nil {
			return err
		}
	} else {
		if err = r.startUpgrade(upgradeVersion); err != nil {
			return err
		}
	}

	if err := openshift.WaitUpgradeCompletion(clientVersion); err != nil {
		return err
	}

	log.Info("Waiting for all the nodes to become ready...")
	if err := openshift.WaitNodesToBeReady(); err != nil {
		return err
	}

	log.Infof("Successfully upgraded ROSA scheduler to version %s", upgradeVersion)
	return nil
}

// updateMaxUnavailableWorkerMCP updates the maxUnavailable and maxSurge for the given machine pool
func (r *rosa) updateMaxUnavailableWorkerMCP(poolID string) error {
	maxUnavailable := os.Getenv("NODEPOOL_MAX_UNAVAILABLE")
	maxSurge := os.Getenv("NODEPOOL_MAX_SURGE")
	nodeDrainGracePeriod := os.Getenv("NODEPOOL_NODE_DRAIN_GRACE_PERIOD")
	if maxUnavailable == "" || maxSurge == "" || nodeDrainGracePeriod == "" {
		log.Warnf("NODEPOOL_MAX_UNAVAILABLE, NODEPOOL_MAX_SURGE and NODEPOOL_NODE_DRAIN_GRACE_PERIOD not set! Using default values")
		maxUnavailable = "1"
		maxSurge = "1"
		nodeDrainGracePeriod = "30"
	}
	args := []string{"edit machinepool", poolID, "--cluster", r.rosaClusterName,
		"--max-unavailable", maxUnavailable, "--max-surge", maxSurge, "--node-drain-grace-period", nodeDrainGracePeriod}
	output, _, err := r.runRosa(args)
	if err != nil {
		return fmt.Errorf("failed to patch machine pool %s: %v", poolID, err)
	}
	log.Infof("Output: %s", string(output))
	log.Infof("Successfully updated machine pool %s with maxUnavailable=%s and maxSurge=%s\n", poolID, maxUnavailable, maxSurge)
	return nil
}

// startUpgrade upgrades the ROSA classic scheduler
func (r *rosa) startUpgrade(upgradeVersion string) error {
	args := []string{"upgrade cluster -y --cluster", r.rosaClusterName, "--version", upgradeVersion}
	output, _, err := r.runRosa(args)
	if err != nil {
		return fmt.Errorf("failed to upgrade ROSA scheduler: %v, output: %s", err, output)
	}
	t := func() (interface{}, bool, error) {
		clusterVersion, err := k8sOpenshift.GetClusterVersion("version")
		if err != nil {
			return nil, true, fmt.Errorf("failed to get cluster version. cause: %v", err)
		}

		desiredVersion := clusterVersion.Status.Desired.Version
		if desiredVersion != upgradeVersion {
			return nil, true, fmt.Errorf("version mismatch. expected: %s but got %s", upgradeVersion, desiredVersion)
		}
		log.Infof("Upgrade done!")
		return nil, false, nil
	}
	if _, err = task.DoRetryWithTimeout(t, defaultUpgradeTimeout, defaultUpgradeRetryInterval); err != nil {
		return err
	}
	return nil
}

// startHCPUpgrade upgrades the Hosted Control Plane and MachinePools
func (r *rosa) startHCPUpgrade(upgradeVersion string) error {
	// Upgrade control plane follow by machinepools upgrade
	args := []string{"upgrade cluster --control-plane -m auto -y --cluster", r.rosaClusterName, "--version", upgradeVersion}
	output, _, err := r.runRosa(args)
	if err != nil {
		return fmt.Errorf("failed to upgrade ROSA scheduler: %v, output: %s", err, output)
	}

	// Check if Upgrade is scheduled/started
	t := func() (interface{}, bool, error) {
		args = []string{"list upgrades --cluster", r.rosaClusterName}
		output, _, err = r.runRosa(args)
		if err != nil {
			return nil, true, fmt.Errorf("failed to list ROSA upgrades output: %s", output)
		}
		re := regexp.MustCompile(`(?i)\b(started)\b`)
		if re.MatchString(output) {
			log.Infof("Upgrade control plane started: %s", output)
			return nil, false, nil
		} else {
			return nil, true, fmt.Errorf("failed to schedule/start upgrade: %s", output)
		}
	}
	if _, err = task.DoRetryWithTimeout(t, defaultCmdTimeout, defaultCmdRetry); err != nil {
		return err
	}

	// Verify Hosted Control plane is upgraded
	t = func() (interface{}, bool, error) {
		args = []string{"describe cluster -o json --cluster", r.rosaClusterName}
		output, _, err = r.runRosa(args)
		if err != nil {
			return nil, true, fmt.Errorf("failed to describe ROSA cluster: %v", err)
		}
		var clusterDetails ClusterDetails
		if err = json.Unmarshal([]byte(output), &clusterDetails); err != nil {
			return nil, true, fmt.Errorf("failed to parse clusterDetails JSON: %v", err)
		}
		if clusterDetails.OpenshiftVersion != upgradeVersion {
			return nil, true, fmt.Errorf("version mismatch. expected: %s but got %s", upgradeVersion, clusterDetails.OpenshiftVersion)
		}
		log.Infof("Hosted Control Plane Upgrade Done!")
		return nil, false, nil
	}
	if _, err = task.DoRetryWithTimeout(t, defaultUpgradeTimeout, defaultUpgradeRetryInterval); err != nil {
		return err
	}

	time.Sleep(defaultSleepInterval) // Sleep to stabilise the control plane

	// Upgrade Machine pools
	nodePools, err := r.listNodePool()
	if err != nil {
		return fmt.Errorf("failed to list ROSA machine pools: %v", err)
	}
	for _, nodePool := range nodePools {
		if nodePool.ID != "torpedo-pool" {
			if err = r.upgradeNodePool(nodePool.ID, upgradeVersion); err != nil {
				return fmt.Errorf("failed to upgrade machine pool: %v", err)
			}
		}
	}

	// Validate for machinepool upgrade
	t = func() (interface{}, bool, error) {
		g, _ := errgroup.WithContext(context.Background())

		for _, nodePool := range nodePools {
			if nodePool.ID == "torpedo-pool" {
				continue
			}
			np := nodePool
			g.Go(func() error {
				versionDetail, err := r.describeNodePool(np.ID)
				if err != nil {
					return fmt.Errorf("failed to describe machine pool %s: %v", np.ID, err)
				}
				if versionDetail.Version.RawID != upgradeVersion {
					return fmt.Errorf("version mismatch. expected: %s but got %s", upgradeVersion, versionDetail.Version.RawID)
				}
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			return nil, true, err
		}
		log.Infof("Upgrade of MachinePools done!")
		return nil, false, nil
	}
	if _, err = task.DoRetryWithTimeout(t, defaultUpgradeTimeout, defaultUpgradeRetryInterval); err != nil {
		return err
	}

	// Verify Openshift Version update
	t = func() (interface{}, bool, error) {
		clusterVersion, err := k8sOpenshift.GetClusterVersion("version")
		if err != nil {
			return nil, true, fmt.Errorf("failed to get cluster version. cause: %v", err)
		}

		desiredVersion := clusterVersion.Status.Desired.Version
		if desiredVersion != upgradeVersion {
			return nil, true, fmt.Errorf("version mismatch. expected: %s but got %s", upgradeVersion, desiredVersion)
		}
		log.Infof("Opensnift Upgrade Done!")
		return nil, false, nil
	}
	if _, err = task.DoRetryWithTimeout(t, defaultUpgradeTimeout, defaultUpgradeRetryInterval); err != nil {
		return err
	}
	return nil
}

// describeNodePool describes the machine pool with version details
func (r *rosa) describeNodePool(poolID string) (*NodePoolVersion, error) {
	log.Infof("Describing machine pool %s of cluster %s", poolID, r.rosaClusterName)
	args := []string{"describe machinepool", poolID, "--cluster", r.rosaClusterName, "-o json"}
	output, _, err := r.runRosa(args)
	if err != nil {
		return nil, fmt.Errorf("error describing machine pool %s: %v", poolID, err)
	}
	var versionDetail NodePoolVersion
	if err := json.Unmarshal([]byte(output), &versionDetail); err != nil {
		return nil, fmt.Errorf("error parsing JSON for machine pool %s: %v", poolID, err)
	}
	log.Infof("NodePoolVersion: %v", versionDetail)
	return &versionDetail, nil
}

// listNodePool lists the machine pools for the given cluster
func (r *rosa) listNodePool() ([]NodePool, error) {
	log.Infof("Listing machine pools for cluster %s", r.rosaClusterName)
	args := []string{"list machinepool -o json --cluster", r.rosaClusterName}
	output, _, err := r.runRosa(args)
	if err != nil {
		return nil, fmt.Errorf("failed to list ROSA machine pools: %v, output: %s", err, output)
	}
	var nodePools []NodePool
	if err := json.Unmarshal([]byte(output), &nodePools); err != nil {
		return nil, fmt.Errorf("Error parsing NodePool JSON: %v\n", err)
	}
	log.Infof("NodePools: %v", nodePools)
	return nodePools, nil
}

// upgradeNodePool upgrades the machine pool to the given version
func (r *rosa) upgradeNodePool(poolID string, version string) error {
	log.Infof("Upgrading machine pool %s to version %s", poolID, version)
	args := []string{"upgrade", "machinepool", poolID, "--cluster", r.rosaClusterName, "--version", version, "-y"}
	output, _, err := r.runRosa(args)
	if err != nil {
		return fmt.Errorf("failed to upgrade machine pool: %v, output: %s", err, output)
	}
	log.Debugf("Output: %s", output)

	t := func() (interface{}, bool, error) {
		versionDetail, err := r.describeNodePool(poolID)
		if err != nil {
			return nil, true, fmt.Errorf("failed to describe machine pool %s: %v", poolID, err)
		}
		if versionDetail.ScheduledUpgrade.State != "started" && versionDetail.ScheduledUpgrade.Version != version {
			return nil, true, fmt.Errorf("waiting to schedule/start upgrade: %s", output)
		}
		return nil, false, nil
	}
	if _, err := task.DoRetryWithTimeout(t, defaultCmdTimeout, defaultCmdRetry); err != nil {
		return err
	}
	return nil
}

// getAvailableUpgradeVersions returns the latest available upgrade version for the given cluster
func (r *rosa) getAvailableUpgradeVersions() (string, error) {
	var upgradeVersion []string
	log.Infof("Fetching ROSA scheduler upgrade versions for %s", r.rosaClusterName)
	// Get available upgrades
	args := []string{"list", "upgrade", "-o", "json", "--cluster", r.rosaClusterName}
	output, _, err := r.runRosa(args)
	if err != nil {
		return "", fmt.Errorf("failed to list ROSA upgrades: %v, output: %s", err, string(output))
	}
	if err = json.Unmarshal([]byte(output), &upgradeVersion); err != nil {
		return "", fmt.Errorf("failed to parse ROSA upgrade versions: %v", err)
	}
	log.Infof("Available ROSA scheduler upgrade versions: %v", upgradeVersion)
	if len(upgradeVersion) == 0 {
		return "", fmt.Errorf("no available upgrades found")
	}
	return upgradeVersion[0], err // return the latest available version for upgrade
}

// downloadROSAClient Constructs URL, downloads and prepares OC CLI to be used based on OCP version given
func (r *rosa) downloadROSAClient() error {
	clientName := "rosa-linux.tar.gz"
	rosaBinaryDir := "/usr/local/bin"

	log.Infof("Downloading the latest ROSA client. May take some time...")

	log.Infof("Downloading ROSA client from URL [%s] to [%s]...", OpenshiftMirror, clientName)
	stdout, err := exec.Command("curl", "-o", clientName, "-L", OpenshiftMirror).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to download OpenShift client from [%s], Err %v %v", OpenshiftMirror, stdout, err)
	}
	log.Infof("Openshift client successfully downloaded from [%s] and saved as [%s]", OpenshiftMirror, clientName)

	log.Infof("Executing command [tar -xvf %s]...", clientName)
	stdout, err = exec.Command("tar", "-xvf", clientName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to extract [%s], Err: %v %v", clientName, err, string(stdout))
	}
	log.Infof("Successfully extracted [%s]", clientName)

	log.Infof("Executing command [cp ./rosa %s]...", rosaBinaryDir)
	stdout, err = exec.Command("cp", "./rosa", rosaBinaryDir).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to copy rosa binary to [%s], Err %v %v", rosaBinaryDir, err, string(stdout))
	}
	log.Infof("Successfully copied rosa binary to [%s]", rosaBinaryDir)

	log.Info("Executing command [rosa version]...")
	stdout, err = exec.Command("rosa", "version").CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get rosa version, Err: %v %v", err, string(stdout))
	}
	log.Infof("Successfully got ROSA version:\n%v\n", string(stdout))
	log.Infof("Logging into ROSA...")
	token := os.Getenv("ROSA_TOKEN")
	if token == "" {
		return fmt.Errorf("ROSA_TOKEN not passed")
	}
	args := []string{"login", fmt.Sprintf("--token=\"%s\"", token)}
	_, _, err = r.runRosa(args)
	if err != nil {
		return fmt.Errorf("failed to login to ROSA, Err: %v %v", err, string(stdout))
	}
	log.Infof("Successfully logged into ROSA")
	return nil
}

func (r *rosa) runRosa(rosaArgs []string) (string, string, error) {
	cmdArgs := []string{rosaCli}
	cmdArgs = append(cmdArgs, strings.Join(rosaArgs, " "))
	return osutils.ExecShell(strings.Join(cmdArgs, " "))
}
