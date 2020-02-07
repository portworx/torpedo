package portworx

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	snapv1 "github.com/kubernetes-incubator/external-storage/snapshot/pkg/apis/crd/v1"
	apapi "github.com/libopenstorage/autopilot-api/pkg/apis/autopilot/v1alpha1"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/client"
	clusterclient "github.com/libopenstorage/openstorage/api/client/cluster"
	"github.com/libopenstorage/openstorage/api/spec"
	"github.com/libopenstorage/openstorage/cluster"
	"github.com/pborman/uuid"
	"github.com/portworx/sched-ops/k8s/core"
	talisman "github.com/portworx/sched-ops/k8s/talisman"
	"github.com/portworx/sched-ops/task"
	talisman_v1beta1 "github.com/portworx/talisman/pkg/apis/portworx/v1beta1"
	talisman_v1beta2 "github.com/portworx/talisman/pkg/apis/portworx/v1beta2"
	"github.com/portworx/torpedo/drivers/node"
	torpedovolume "github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/drivers/volume/portworx/schedops"
	"github.com/portworx/torpedo/pkg/aututils"
	tp_errors "github.com/portworx/torpedo/pkg/errors"
	"github.com/portworx/torpedo/pkg/units"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	//PortworxStorage portworx storage name
	PortworxStorage torpedovolume.StorageProvisionerType = "portworx"
	//PortworxCsi csi storage name
	PortworxCsi torpedovolume.StorageProvisionerType = "csi"
	// PxPoolAvailableCapacityMetric is metric for pool available capacity
	PxPoolAvailableCapacityMetric = "100 * ( px_pool_stats_available_bytes/ px_pool_stats_total_bytes)"
	// PxPoolTotalCapacityMetric is metric for pool total capacity
	PxPoolTotalCapacityMetric = "px_pool_stats_total_bytes/(1024*1024*1024)"
	// PxVolumeUsagePercentMetric is metric for volume usage percentage
	PxVolumeUsagePercentMetric = "100 * (px_volume_usage_bytes / px_volume_capacity_bytes)"
	// PxVolumeCapacityPercentMetric is metric for volume capacity percentage
	PxVolumeCapacityPercentMetric = "px_volume_capacity_bytes / 1000000000"
	// VolumeSpecAction is name for volume spec action
	VolumeSpecAction = "openstorage.io.action.volume/resize"
	// StorageSpecAction is name for storage spec action
	StorageSpecAction = "openstorage.io.action.storagepool/expand"
	// RuleActionsScalePercentage is name for scale percentage rule action
	RuleActionsScalePercentage = "scalepercentage"
	// RuleScaleType is name for scale type
	RuleScaleType = "scaletype"
	// RuleScaleTypeAddDisk is name for add disk scale type
	RuleScaleTypeAddDisk = "add-disk"
	// RuleScaleTypeResizeDisk is name for resize disk scale type
	RuleScaleTypeResizeDisk = "resize-disk"
	// RuleMaxSize is name for rule max size
	RuleMaxSize = "maxsize"
)

const (
	// DriverName is the name of the portworx driver implementation
	DriverName           = "pxd"
	pxDiagPath           = "/remotediags"
	pxVersionLabel       = "PX Version"
	enterMaintenancePath = "/entermaintenance"
	exitMaintenancePath  = "/exitmaintenance"
	pxSystemdServiceName = "portworx.service"
	tokenKey             = "token"
	clusterIP            = "ip"
	clusterPort          = "port"
)

const (
	defaultTimeout                   = 2 * time.Minute
	defaultRetryInterval             = 10 * time.Second
	maintenanceOpTimeout             = 1 * time.Minute
	maintenanceWaitTimeout           = 2 * time.Minute
	inspectVolumeTimeout             = 30 * time.Second
	inspectVolumeRetryInterval       = 2 * time.Second
	validateDeleteVolumeTimeout      = 3 * time.Minute
	validateReplicationUpdateTimeout = 10 * time.Minute
	validateClusterStartTimeout      = 2 * time.Minute
	validatePXStartTimeout           = 5 * time.Minute
	validateNodeStopTimeout          = 5 * time.Minute
	validateStoragePoolSizeTimeout   = 3 * time.Hour
	validateStoragePoolSizeInterval  = 30 * time.Second
	getNodeTimeout                   = 3 * time.Minute
	getNodeRetryInterval             = 5 * time.Second
	stopDriverTimeout                = 5 * time.Minute
	crashDriverTimeout               = 2 * time.Minute
	startDriverTimeout               = 2 * time.Minute
	upgradeTimeout                   = 10 * time.Minute
	upgradeRetryInterval             = 30 * time.Second
	upgradePerNodeTimeout            = 15 * time.Minute
	waitVolDriverToCrash             = 1 * time.Minute
)

const (
	secretName      = "openstorage.io/auth-secret-name"
	secretNamespace = "openstorage.io/auth-secret-namespace"
)

// Provisioners types of supported provisioners
var provisioners = map[torpedovolume.StorageProvisionerType]torpedovolume.StorageProvisionerType{
	PortworxStorage: "kubernetes.io/portworx-volume",
	PortworxCsi:     "pxd.portworx.com",
}

var deleteVolumeLabelList = []string{"auth-token", "pv.kubernetes.io", "volume.beta.kubernetes.io", "kubectl.kubernetes.io", "volume.kubernetes.io"}
var k8sCore = core.Instance()
var k8sTalisman = talisman.Instance()

type portworx struct {
	legacyClusterManager cluster.Cluster
	clusterManager       api.OpenStorageClusterClient
	nodeManager          api.OpenStorageNodeClient
	mountAttachManager   api.OpenStorageMountAttachClient
	volDriver            api.OpenStorageVolumeClient
	clusterPairManager   api.OpenStorageClusterPairClient
	alertsManager        api.OpenStorageAlertsClient
	schedOps             schedops.Driver
	nodeDriver           node.Driver
	refreshEndpoint      bool
	token                string
}

// TODO temporary solution until sdk supports metadataNode response
type metadataNode struct {
	PeerUrls   []string `json:"PeerUrls"`
	ClientUrls []string `json:"ClientUrls"`
	Leader     bool     `json:"Leader"`
	DbSize     int      `json:"DbSize"`
	IsHealthy  bool     `json:"IsHealthy"`
	ID         string   `json:"ID"`
}

// DiagRequestConfig is a request object which provides all the configuration details
// to PX for running diagnostics on a node. This object can also be passed over
// the wire through an API server for remote diag requests.
type DiagRequestConfig struct {
	// OutputFile for the diags.tgz
	OutputFile string
	// DockerHost config
	DockerHost string
	// ContainerName for PX
	ContainerName string
	// ExecPath of the program making this request (pxctl)
	ExecPath string
	// Profile when set diags command only dumps the go profile
	Profile bool
	// Live gets live diagnostics
	Live bool
	// Upload uploads the diags.tgz to s3
	Upload bool
	// All gets all possible diagnostics from PX
	All bool
	// Force overwrite of existing diags file.
	Force bool
	// OnHost indicates whether diags is being run on the host
	// or inside the container
	OnHost bool
	// Token for security authentication (if enabled)of the program making this request (pxctl)
	Token string
	// Extra indicates whether diags should attempt to collect extra information
	Extra bool
}

func (d *portworx) String() string {
	return DriverName
}

func (d *portworx) Init(sched string, nodeDriver string, token string, storageProvisioner string) error {
	logrus.Infof("Using the Portworx volume driver with provisioner %s under scheduler: %v", storageProvisioner, sched)
	var err error

	d.token = token

	if d.nodeDriver, err = node.Get(nodeDriver); err != nil {
		return err
	}

	if d.schedOps, err = schedops.Get(sched); err != nil {
		return fmt.Errorf("failed to get scheduler operator for portworx. Err: %v", err)
	}

	if err = d.setDriver(); err != nil {
		return err
	}

	storageNodes, err := d.getStorageNodesOnStart()
	if err != nil {
		return err
	}

	if len(storageNodes) == 0 {
		return fmt.Errorf("cluster inspect returned empty nodes")
	}

	err = d.updateNodes(storageNodes)
	if err != nil {
		return err
	}
	for _, n := range node.GetStorageDriverNodes() {
		if err = d.WaitDriverUpOnNode(n, validatePXStartTimeout); err != nil {
			return err
		}
	}

	logrus.Infof("The following Portworx nodes are in the cluster:")
	for _, n := range storageNodes {
		logrus.Infof(
			"Node UID: %v Node IP: %v Node Status: %v",
			n.Id,
			n.DataIp,
			n.Status,
		)
	}
	// Set provisioner for torpedo
	if storageProvisioner != "" {
		if p, ok := provisioners[torpedovolume.StorageProvisionerType(storageProvisioner)]; ok {
			torpedovolume.StorageProvisioner = p
		} else {
			return fmt.Errorf("driver %s, does not support provisioner %s", DriverName, storageProvisioner)
		}
	} else {
		torpedovolume.StorageProvisioner = provisioners[torpedovolume.DefaultStorageProvisioner]
	}
	return nil
}

func (d *portworx) RefreshDriverEndpoints() error {
	storageNodes, err := d.getStorageNodesOnStart()
	if err != nil {
		return err
	}

	if len(storageNodes) == 0 {
		return fmt.Errorf("cluster inspect returned empty nodes")
	}

	err = d.updateNodes(storageNodes)
	if err != nil {
		return err
	}
	return nil
}

func (d *portworx) updateNodes(pxNodes []api.StorageNode) error {
	for _, n := range node.GetWorkerNodes() {
		if err := d.updateNode(&n, pxNodes); err != nil {
			return err
		}
	}

	return nil
}

func (d *portworx) updateNode(n *node.Node, pxNodes []api.StorageNode) error {
	isPX, err := d.schedOps.IsPXEnabled(*n)
	if err != nil {
		return err
	}

	// No need to check in pxNodes if px is not installed
	if !isPX {
		return nil
	}

	for _, address := range n.Addresses {
		for _, pxNode := range pxNodes {
			if address == pxNode.DataIp || address == pxNode.MgmtIp || n.Name == pxNode.SchedulerNodeName {
				if len(pxNode.Id) > 0 {
					n.StorageNode = pxNode
					n.VolDriverNodeID = pxNode.Id
					n.IsStorageDriverInstalled = isPX
					isMetadataNode, err := d.isMetadataNode(*n, address)
					if err != nil {
						return err
					}
					n.IsMetadataNode = isMetadataNode

					if n.StoragePools == nil {
						for _, pxNodePool := range pxNode.Pools {
							storagePool := node.StoragePool{
								StoragePool:       pxNodePool,
								StoragePoolAtInit: pxNodePool,
							}
							n.StoragePools = append(n.StoragePools, storagePool)
						}
					} else {
						for idx, nodeStoragePool := range n.StoragePools {
							for _, pxNodePool := range pxNode.Pools {
								if nodeStoragePool.Uuid == pxNodePool.Uuid {
									n.StoragePools[idx].StoragePool = pxNodePool
								}
							}
						}
					}
					if err = node.UpdateNode(*n); err != nil {
						return fmt.Errorf("failed to update node %s. Cause: %v", n.Name, err)
					}
				} else {
					return fmt.Errorf("StorageNodeId is empty for node %v", pxNode)
				}
				return nil
			}
		}
	}

	// Return error where PX is not explicitly disabled but was not found installed
	return fmt.Errorf("failed to find px node for node: %v PX nodes: %v", n, pxNodes)
}

func (d *portworx) isMetadataNode(node node.Node, address string) (bool, error) {
	members, err := d.getKvdbMembers(node)
	if err != nil {
		return false, fmt.Errorf("failed to get metadata nodes. Cause: %v", err)
	}

	ipRegex := regexp.MustCompile(`http://(?P<address>.*):d+`)
	for _, value := range members {
		for _, url := range value.ClientUrls {
			result := getGroupMatches(ipRegex, url)
			if val, ok := result["address"]; ok && address == val {
				logrus.Debugf("Node %s is a metadata node", node.Name)
				return true, nil
			}
		}
	}
	return false, nil
}

func (d *portworx) CleanupVolume(volumeName string) error {
	volDriver := d.getVolDriver()
	volumes, err := volDriver.Enumerate(d.getContext(), &api.SdkVolumeEnumerateRequest{}, nil)
	if err != nil {
		return err
	}

	for _, volumeID := range volumes.GetVolumeIds() {
		volumeInspectResponse, err := volDriver.Inspect(d.getContext(), &api.SdkVolumeInspectRequest{VolumeId: volumeID})
		if err != nil {
			return err
		}
		pxVolume := volumeInspectResponse.Volume
		if pxVolume.Locator.Name == volumeName {
			// First unmount this volume at all mount paths...
			for _, path := range pxVolume.AttachPath {
				if _, err = d.getMountAttachManager().Unmount(d.getContext(), &api.SdkVolumeUnmountRequest{VolumeId: pxVolume.Id, MountPath: path}); err != nil {
					err = fmt.Errorf(
						"error while unmounting %v at %v because of: %v",
						pxVolume.Id,
						path,
						err,
					)
					logrus.Infof("%v", err)
					return err
				}
			}

			if _, err = d.mountAttachManager.Detach(d.getContext(), &api.SdkVolumeDetachRequest{VolumeId: pxVolume.Id}); err != nil {
				err = fmt.Errorf(
					"error while detaching %v because of: %v",
					pxVolume.Id,
					err,
				)
				logrus.Infof("%v", err)
				return err
			}

			if _, err := volDriver.Delete(d.getContext(), &api.SdkVolumeDeleteRequest{VolumeId: pxVolume.Id}); err != nil {
				err = fmt.Errorf(
					"error while deleting %v because of: %v",
					pxVolume.Id,
					err,
				)
				logrus.Infof("%v", err)
				return err
			}

			logrus.Infof("successfully removed Portworx volume %v", volumeName)

			return nil
		}
	}

	return nil
}

func (d *portworx) getPxNode(n *node.Node, nManager ...api.OpenStorageNodeClient) (api.StorageNode, error) {
	if len(nManager) == 0 {
		nManager = []api.OpenStorageNodeClient{d.getNodeManager()}
	}
	logrus.Debugf("Inspecting node [%s] with volume driver node id [%s]", n.Name, n.VolDriverNodeID)
	nodeInspectResponse, err := nManager[0].Inspect(d.getContext(), &api.SdkNodeInspectRequest{NodeId: n.VolDriverNodeID})
	if isNodeNotFound(err) {
		logrus.Warnf("node %s with ID %s not found, trying to update node ID...", n.Name, n.VolDriverNodeID)
		n, err = d.updateNodeID(n, nManager...)
		if err != nil {
			return api.StorageNode{Status: api.Status_STATUS_NONE}, err
		}
		return d.getPxNode(n, nManager...)
	}
	return *nodeInspectResponse.Node, nil
}

func isNodeNotFound(err error) bool {
	st, _ := status.FromError(err)
	// TODO when a node is not found sometimes we get an error code internal, as workaround we check for internal error and substring
	return err != nil && (st.Code() == codes.NotFound || (st.Code() == codes.Internal && strings.Contains(err.Error(), "Unable to locate node")))
}

func (d *portworx) getPxVersionOnNode(n node.Node, nodeManager ...api.OpenStorageNodeClient) (string, error) {

	t := func() (interface{}, bool, error) {
		logrus.Debugf("Getting PX Version on node [%s]", n.Name)
		pxNode, err := d.getPxNode(&n, nodeManager...)
		if err != nil {
			return "", false, err
		}
		if pxNode.Status != api.Status_STATUS_OK {
			return "", true, fmt.Errorf("px cluster is usable but node status is not ok. Expected: %v Actual: %v",
				api.Status_STATUS_OK, pxNode.Status)
		}
		pxVersion := pxNode.NodeLabels[pxVersionLabel]
		return pxVersion, false, nil
	}
	pxVersion, err := task.DoRetryWithTimeout(t, getNodeTimeout, getNodeRetryInterval)
	if err != nil {
		return "", fmt.Errorf("Timeout after %v waiting to get PX Version", getNodeTimeout)
	}
	return fmt.Sprintf("%v", pxVersion), nil
}

func (d *portworx) GetStorageDevices(n node.Node) ([]string, error) {
	pxNode, err := d.getPxNode(&n)
	if err != nil {
		return nil, err
	}

	devPaths := make([]string, 0)
	for _, value := range pxNode.Disks {
		devPaths = append(devPaths, value.Path)
	}
	return devPaths, nil
}

func (d *portworx) RecoverDriver(n node.Node) error {

	t := func() (interface{}, bool, error) {
		if err := d.maintenanceOp(n, enterMaintenancePath); err != nil {
			return nil, true, err
		}
		return nil, false, nil
	}

	if _, err := task.DoRetryWithTimeout(t, maintenanceOpTimeout, defaultRetryInterval); err != nil {
		return err
	}
	t = func() (interface{}, bool, error) {
		apiNode, err := d.getPxNode(&n)
		if err != nil {
			return nil, true, err
		}
		if apiNode.Status == api.Status_STATUS_MAINTENANCE {
			return nil, false, nil
		}
		return nil, true, fmt.Errorf("Node %v is not in Maintenance mode", n.Name)
	}

	if _, err := task.DoRetryWithTimeout(t, maintenanceWaitTimeout, defaultRetryInterval); err != nil {
		return &ErrFailedToRecoverDriver{
			Node:  n,
			Cause: err.Error(),
		}
	}
	t = func() (interface{}, bool, error) {
		if err := d.maintenanceOp(n, exitMaintenancePath); err != nil {
			return nil, true, err
		}
		return nil, false, nil
	}

	if _, err := task.DoRetryWithTimeout(t, maintenanceOpTimeout, defaultRetryInterval); err != nil {
		return err
	}

	t = func() (interface{}, bool, error) {
		apiNode, err := d.getPxNode(&n)
		if err != nil {
			return nil, true, err
		}
		if apiNode.Status == api.Status_STATUS_OK {
			return nil, false, nil
		}
		return nil, true, fmt.Errorf("Node %v is not up after exiting  Maintenance mode", n.Name)
	}

	if _, err := task.DoRetryWithTimeout(t, maintenanceWaitTimeout, defaultRetryInterval); err != nil {
		return err
	}

	return nil
}

func (d *portworx) ValidateCreateVolume(appVols map[string]map[string]string, appName string, appToken string) error {

	/*Volume affinity/Anti Affinity  related data creation : START*/
	var appVolList = map[string]*api.Volume{}

	// Loop over all volumes of the app
	for appVol, params := range appVols {

		logrus.Infof("get %s app's volume: %s inspected by the volume driver", appName, appVol)
		var token string
		params["auth-token"] = appToken
		token = d.getTokenForVolume(appVol, params)
		volDriver := d.getVolDriver()
		t := func() (interface{}, bool, error) {
			volumeInspectResponse, err := volDriver.Inspect(d.getContextWithToken(context.Background(), token), &api.SdkVolumeInspectRequest{VolumeId: appVol})
			if err != nil {
				return nil, true, err
			}
			return volumeInspectResponse.Volume, false, nil
		}

		out, err := task.DoRetryWithTimeout(t, inspectVolumeTimeout, inspectVolumeRetryInterval)
		if err != nil {
			return &ErrFailedToInspectVolume{
				ID:    appVol,
				Cause: fmt.Sprintf("Volume inspect returned err: %v", err),
			}
		}

		appVolList[appVol] = out.(*api.Volume)

	}

	//Get list of volumes and its replica per each node of the App
	nodeReplMap := d.getReplicaNodeMap(appVolList)
	//Get all nodes having app's volume
	appVolNodes, err := d.getAppVolNodes(appVols, appToken)
	/*Volume affinity/Anti Affinity  related data : END*/
	if err != nil {
		return &ErrFailedToInspectVolume{
			Cause: fmt.Sprintf("failed to generate node list having all volumes of the App. Err: %v", err),
		}
	}

	for appVol, params := range appVols {
		var token string
		params["auth-token"] = appToken
		token = d.getTokenForVolume(appVol, params)
		volDriver := d.getVolDriver()

		vol := appVolList[appVol]
		// Status
		if vol.Status != api.VolumeStatus_VOLUME_STATUS_UP {
			return &ErrFailedToInspectVolume{
				ID: appVol,
				Cause: fmt.Sprintf("Volume has invalid status. Expected:%v Actual:%v",
					api.VolumeStatus_VOLUME_STATUS_UP, vol.Status),
			}
		}

		// State
		if vol.State == api.VolumeState_VOLUME_STATE_ERROR || vol.State == api.VolumeState_VOLUME_STATE_DELETED {
			return &ErrFailedToInspectVolume{
				ID:    appVol,
				Cause: fmt.Sprintf("Volume has invalid state. Actual:%v", vol.State),
			}
		}

		// if the volume is a clone or a snap, validate it's parent
		if vol.IsSnapshot() || vol.IsClone() {
			parentResp, err := volDriver.Inspect(d.getContextWithToken(context.Background(), token), &api.SdkVolumeInspectRequest{VolumeId: vol.Source.Parent})
			if err != nil {
				return &ErrFailedToInspectVolume{
					ID:    appVol,
					Cause: fmt.Sprintf("Could not get parent with ID [%s]", vol.Source.Parent),
				}

			}

			if err := d.schedOps.ValidateSnapshot(params, parentResp.Volume); err != nil {
				return &ErrFailedToInspectVolume{
					ID:    appVol,
					Cause: fmt.Sprintf("Snapshot/Clone validation failed. %v", err),
				}
			}
			//TODO: IS validation for snapshot or clone needs to be done of VPS
			// Continue to investigate next volume
			continue
		}
		// Labels
		var pxNodes []api.StorageNode

		for _, rs := range vol.ReplicaSets {
			for _, n := range rs.Nodes {
				nodeResponse, err := d.getNodeManager().Inspect(d.getContextWithToken(context.Background(), token), &api.SdkNodeInspectRequest{NodeId: n})
				if err != nil {
					return &ErrFailedToInspectVolume{
						ID:    appVol,
						Cause: fmt.Sprintf("Failed to inspect replica set node: %s err: %v", n, err),
					}
				}

				pxNodes = append(pxNodes, *nodeResponse.Node)
			}
		}

		//logrus.Infof("Volume Replicas: %v (%v) VolumeNodes %v", vol.Id, vol.ReplicaSets, pxNodes)

		// Spec
		requestedSpec, requestedLocator, _, err := spec.NewSpecHandler().SpecFromOpts(params)
		if err != nil {
			return &ErrFailedToInspectVolume{
				ID:    appVol,
				Cause: fmt.Sprintf("failed to parse requested spec of volume. Err: %v", err),
			}
		}

		delete(vol.Locator.VolumeLabels, "pvc") // special handling for the new pvc label added in k8s
		deleteLabelsFromRequestedSpec(requestedLocator)

		// Params/Options
		for k, v := range params {
			switch k {
			case api.SpecNodes:
				if !reflect.DeepEqual(v, vol.Spec.ReplicaSet.Nodes) {
					return errFailedToInspectVolume(appVol, k, v, vol.Spec.ReplicaSet.Nodes)
				}
			case api.SpecParent:
				if v != vol.Source.Parent {
					return errFailedToInspectVolume(appVol, k, v, vol.Source.Parent)
				}
			case api.SpecEphemeral:
				if requestedSpec.Ephemeral != vol.Spec.Ephemeral {
					return errFailedToInspectVolume(appVol, k, requestedSpec.Ephemeral, vol.Spec.Ephemeral)
				}
			case api.SpecFilesystem:
				if requestedSpec.Format != vol.Spec.Format {
					return errFailedToInspectVolume(appVol, k, requestedSpec.Format, vol.Spec.Format)
				}
			case api.SpecBlockSize:
				if requestedSpec.BlockSize != vol.Spec.BlockSize {
					return errFailedToInspectVolume(appVol, k, requestedSpec.BlockSize, vol.Spec.BlockSize)
				}
			case api.SpecHaLevel:
				if requestedSpec.HaLevel != vol.Spec.HaLevel {
					return errFailedToInspectVolume(appVol, k, requestedSpec.HaLevel, vol.Spec.HaLevel)
				}
			case api.SpecPriorityAlias:
				// Since IO priority isn't guaranteed, we aren't validating it here.
			case api.SpecSnapshotInterval:
				if requestedSpec.SnapshotInterval != vol.Spec.SnapshotInterval {
					return errFailedToInspectVolume(appVol, k, requestedSpec.SnapshotInterval, vol.Spec.SnapshotInterval)
				}
			case api.SpecSnapshotSchedule:
				// TODO currently volume spec has a different format than request
				// i.e request "daily=12:00,7" turns into "- freq: daily\n  hour: 12\n  retain: 7\n" in volume spec
				//if requestedSpec.SnapshotSchedule != vol.Spec.SnapshotSchedule {
				//	return errFailedToInspectVolume(name, k, requestedSpec.SnapshotSchedule, vol.Spec.SnapshotSchedule)
				//}
			case api.SpecAggregationLevel:
				if requestedSpec.AggregationLevel != vol.Spec.AggregationLevel {
					return errFailedToInspectVolume(appVol, k, requestedSpec.AggregationLevel, vol.Spec.AggregationLevel)
				}
			case api.SpecShared:
				if requestedSpec.Shared != vol.Spec.Shared {
					return errFailedToInspectVolume(appVol, k, requestedSpec.Shared, vol.Spec.Shared)
				}
			case api.SpecSticky:
				if requestedSpec.Sticky != vol.Spec.Sticky {
					return errFailedToInspectVolume(appVol, k, requestedSpec.Sticky, vol.Spec.Sticky)
				}
			case api.SpecGroup:
				if !reflect.DeepEqual(requestedSpec.Group, vol.Spec.Group) {
					return errFailedToInspectVolume(appVol, k, requestedSpec.Group, vol.Spec.Group)
				}
			case api.SpecGroupEnforce:
				if requestedSpec.GroupEnforced != vol.Spec.GroupEnforced {
					return errFailedToInspectVolume(appVol, k, requestedSpec.GroupEnforced, vol.Spec.GroupEnforced)
				}
			// portworx injects pvc name and namespace labels so response object won't be equal to request
			case api.SpecLabels:
				for requestedLabelKey, requestedLabelValue := range requestedLocator.VolumeLabels {
					// check requested label is not in 'ignore' list
					if labelValue, exists := vol.Locator.VolumeLabels[requestedLabelKey]; !exists || requestedLabelValue != labelValue {
						return errFailedToInspectVolume(appVol, k, requestedLocator.VolumeLabels, vol.Locator.VolumeLabels)
					}
				}
			case api.SpecIoProfile:
				if requestedSpec.IoProfile != vol.Spec.IoProfile {
					return errFailedToInspectVolume(appVol, k, requestedSpec.IoProfile, vol.Spec.IoProfile)
				}
			case api.SpecSize:
				if requestedSpec.Size != vol.Spec.Size {
					return errFailedToInspectVolume(appVol, k, requestedSpec.Size, vol.Spec.Size)
				}
			default:
			}
		}
		err = d.ValidateVps(vol, appVols, pxNodes, nodeReplMap, appVolNodes)
		if err != nil {
			return &ErrFailedToInspectVolume{
				ID:    appVol,
				Cause: fmt.Sprintf("failed to validate VolumePlacementStratergy. Err: %v", err),
			}
		}
		logrus.Infof("Successfully inspected volume: %v (%v)", vol.Locator.Name, vol.Id)

	}

	return nil
}

// Validate the volume replicas as per VolumePlacementStrategy rule applied to the volume
func (d *portworx) ValidateVps(vol *api.Volume, appVols map[string]map[string]string, volNodes []api.StorageNode, nodeReplMap map[string][]*api.Volume, appVolNodes map[string]api.StorageNode) error {

	logrus.Debugf("Volume details: %v ===\n (%v) Volumes per Node List:%v, all nodes of the App's Volumes: %v ", vol, appVols, nodeReplMap, appVolNodes)
	logrus.Infof("Validate VPS  for Volume:%v ,  VPS Rule :(%v)", vol.Id, vol.Spec.GetPlacementStrategy())

	if vpsrule := vol.Spec.GetPlacementStrategy(); vpsrule != nil {

		// Get Volume Labels
		volLabels := vol.Spec.GetVolumeLabels()

		//Get vps spec from k8s
		volVpsRule, err := k8sTalisman.GetVolumePlacementStrategy(volLabels["placement_strategy"])
		logrus.Debugf("====Volume details-5: %v ===\n Placement strategy (%v)", vol.Id, volVpsRule)

		if err != nil {
			return err
		}

		// For each VPS  group rules
		// Group1 -Replica Affinity Group
		if volVpsRule.Spec.ReplicaAffinity != nil {
			for _, rRule := range volVpsRule.Spec.ReplicaAffinity {

				logrus.Infof("Validate Replica Affinity Rule Vol:%v === RA: Spec %v", vol.Id, rRule.CommonPlacementSpec)
				err = d.ValidateReplicaAffinity(vol, rRule, volNodes)
				if err != nil {
					return &ErrFailedToInspectVolume{
						ID:    vol.Id,
						Cause: fmt.Sprintf("failed to validate ReplicaAffinity. Err: %v", err),
					}
				}
			}
		}
		// Group2 -Replica Anti Affinity Group
		if volVpsRule.Spec.ReplicaAntiAffinity != nil {
			for _, raRule := range volVpsRule.Spec.ReplicaAntiAffinity {

				logrus.Infof("Validate Replica Anti Affinity Rule Vol:%v === RAA: Spec %v", vol.Id, raRule.CommonPlacementSpec)
				err = d.ValidateReplicaAntiAffinity(vol, raRule, volNodes)
				if err != nil {
					return &ErrFailedToInspectVolume{
						ID:    vol.Id,
						Cause: fmt.Sprintf("failed to validate ReplicaAntiAffinity. Err: %v", err),
					}
				}
			}
		}
		// Group3 -Volume Affinity Group
		if volVpsRule.Spec.VolumeAffinity != nil {
			for _, vRule := range volVpsRule.Spec.VolumeAffinity {

				logrus.Infof("Validate Volume Affinity Rule Vol:%v === VA: Spec %v", vol.Id, vRule)
				err = d.ValidateVolumeAffinity(vol, vRule, volNodes, nodeReplMap, appVolNodes)
				if err != nil {
					return &ErrFailedToInspectVolume{
						ID:    vol.Id,
						Cause: fmt.Sprintf("failed to validate VolumeAffinity. Err: %v", err),
					}
				}
			}
		}
		// Group4 -Volume Anti Affinity Group
		if volVpsRule.Spec.VolumeAntiAffinity != nil {
			for _, vaRule := range volVpsRule.Spec.VolumeAntiAffinity {

				logrus.Infof("Validate Volume Anti Affinity Rule Vol:%v === VAA: Spec %v", vol.Id, vaRule)
				err = d.ValidateVolumeAntiAffinity(vol, vaRule, volNodes, nodeReplMap, appVolNodes)
				if err != nil {
					return &ErrFailedToInspectVolume{
						ID:    vol.Id,
						Cause: fmt.Sprintf("failed to validate VolumeAntiAffinity. Err: %v", err),
					}
				}
			}
		}

	} else {
		logrus.Infof("Validate VolumePlacementStrategy,  Volume (%v) doesnot have any VPS rule applied to it", vol.Id)
	}

	return nil
}

/* VolumePlacementStrategy validate rule function */

// Get poolid on which the volume replica is placed
func (d *portworx) getReplicaPoolMap(vol *api.Volume) []map[string]string {

	var replPoolMapList = []map[string]string{}
	if vol != nil {
		for rinx, replicaset := range vol.ReplicaSets {
			var replPoolMap = map[string]string{}
			logrus.Infof("getReplicaPoolMap  vol:%v replicaset:%v", vol.Id, replicaset)
			for inx, node := range replicaset.Nodes {
				replPoolMap[node] = vol.ReplicaSets[rinx].PoolUuids[inx]
			}
			replPoolMapList = append(replPoolMapList, replPoolMap)
		}

	}
	return replPoolMapList
}

// Get node with the list of all volume replicas
func (d *portworx) getReplicaNodeMap(appVolList map[string]*api.Volume) map[string][]*api.Volume {

	var replNodeMap = map[string][]*api.Volume{}
	if appVolList != nil {

		for _, vol := range appVolList {
			if vol != nil {
				for _, replicaset := range vol.ReplicaSets {
					logrus.Infof("getReplicaPoolMap  vol:%v replicaset:%v", vol.Id, replicaset)
					for _, node := range replicaset.Nodes {
						replNodeMap[node] = append(replNodeMap[node], vol)
					}
				}

			}
		}
	}
	return replNodeMap
}

//Get all nodes having app's volume
func (d *portworx) getAppVolNodes(appVols map[string]map[string]string, appToken string) (map[string]api.StorageNode, error) {
	logrus.Infof("getAppVolNodes  get all nodes having app's volume: %v", appVols)

	var pxNodes = map[string]api.StorageNode{}
	for appVol, params := range appVols {
		var token string
		params["auth-token"] = appToken
		token = d.getTokenForVolume(appVol, params)
		volDriver := d.getVolDriver()
		t := func() (interface{}, bool, error) {
			volumeInspectResponse, err := volDriver.Inspect(d.getContextWithToken(context.Background(), token), &api.SdkVolumeInspectRequest{VolumeId: appVol})
			if err != nil {
				return nil, true, err
			}
			return volumeInspectResponse.Volume, false, nil
		}

		out, err := task.DoRetryWithTimeout(t, inspectVolumeTimeout, inspectVolumeRetryInterval)
		if err != nil {
			return nil, &ErrFailedToInspectVolume{
				ID:    appVol,
				Cause: fmt.Sprintf("Volume inspect returned err: %v", err),
			}
		}

		vol := out.(*api.Volume)
		for _, rs := range vol.ReplicaSets {
			for _, n := range rs.Nodes {
				if _, ok := pxNodes[n]; ok {
					continue
				} else {
					nodeResponse, err := d.getNodeManager().Inspect(d.getContextWithToken(context.Background(), token), &api.SdkNodeInspectRequest{NodeId: n})
					if err != nil {
						return nil, &ErrFailedToInspectVolume{
							ID:    appVol,
							Cause: fmt.Sprintf("Failed to inspect replica set node: %s err: %v", n, err),
						}
					}

					pxNodes[n] = *nodeResponse.Node
				}
			}
		}
	}
	return pxNodes, nil
}

// Create nodes list group based on the topology key
// { "value":["node1", node2] , "value1" : ["node3", "node4"] }
func (d *portworx) GroupTopologyNodes(topologykey string, volNodes []api.StorageNode) map[string]map[string]api.StorageNode {
	logrus.Infof("GroupTopologyNodes  Group nodes on values of the topology key: %v", topologykey)
	var nodelist = map[string]map[string]api.StorageNode{}

	if topologykey != "" {
		//Group Volume nodes based on the topology key value
		for _, vnode := range volNodes {
			//Get topology key value of the volume node
			for _, nodePool := range vnode.Pools {
				tkval := nodePool.Labels[topologykey]
				logrus.Debugf("GroupTopologyNodes grouping nodes together on value :%v for key:%v for node:%v  NodeLabels:%v", tkval, topologykey, vnode.Id, vnode.NodeLabels)
				if nodelist[tkval] == nil {
					nodelist[tkval] = map[string]api.StorageNode{}
				}

				nodelist[tkval][vnode.Id] = vnode
			}
		}
	} else {
		//Group all volume nodes into one set
		nodelist["all"] = map[string]api.StorageNode{}
		for _, vnode := range volNodes {

			nodelist["all"][vnode.Id] = vnode
		}
	}

	return nodelist
}

func (d *portworx) GroupTopologyAppNodes(topologykey string, volNodes map[string]api.StorageNode) map[string]map[string]api.StorageNode {
	logrus.Infof("GroupTopologyAppNodes  Group nodes on values of the topology key: %v", topologykey)
	var nodelist = map[string]map[string]api.StorageNode{}

	if topologykey != "" {
		//Group Volume nodes based on the topology key value
		for _, vnode := range volNodes {
			//Get topology key value of the volume node
			for _, nodePool := range vnode.Pools {
				tkval := nodePool.Labels[topologykey]
				logrus.Debugf("GroupTopologyNodes grouping nodes together on value :%v for key:%v for node:%v  NodeLabels:%v", tkval, topologykey, vnode.Id, vnode.NodeLabels)
				if nodelist[tkval] == nil {
					nodelist[tkval] = map[string]api.StorageNode{}
				}

				nodelist[tkval][vnode.Id] = vnode
			}
		}
	} else {
		//Group all volume nodes into one set
		for inx, vnode := range volNodes {
			nodelist[inx] = map[string]api.StorageNode{}

			nodelist[inx][vnode.Id] = vnode
		}
	}

	return nodelist
}

// ReplicaMatchExpression
func (d *portworx) ReplicaMatchExpression(node api.StorageNode, matchExp *talisman_v1beta1.LabelSelectorRequirement, poolUUID string) (bool, error) {
	logrus.Infof("ReplicaMatchExpression started nodeId: %v matchExpression:%v ", node.Id, matchExp)

	logrus.Debugf("ReplicaMatchExpression rule:%v  mkey:%v mOperator:%v mValues:%v", node, matchExp.Key, matchExp.Operator, matchExp.Values)
	if node.Pools != nil {
		// each storage  pool on the node  check  for pool lables
		for _, spool := range node.Pools {

			//Check the nodes storagepool on which this replica is residing
			if spool.Uuid == poolUUID {

				//Check whether the pool's label contain the key
				if tval, ok := spool.Labels[matchExp.Key]; ok {

					// Check for Operators: In, Exists, NotIn, DoesNotExist
					logrus.Infof("ReplicaMatchExpression key found, check for  rule:%v  mkey:%v mOperator:%v mValues:%v tval:%v", node, matchExp.Key, matchExp.Operator, matchExp.Values, tval)
					switch matchExp.Operator {
					case "NotIn":
						for _, mValue := range matchExp.Values {
							if mValue == tval {
								//The node should not have the label value, hence return error
								return false, &ErrFailedToInspectVolume{
									Cause: fmt.Sprintf("ReplicaAffinity with matchExpression 'NotIn' with values %v is placed on node:%v having label value:(%v:%v)", matchExp.Values, node.Id, matchExp.Key, tval),
								}
							}
						}
					case "In":
						for _, mValue := range matchExp.Values {
							if mValue == tval {
								//The node should have the label value, hence return true
								logrus.Infof("ReplicaMatchExpression  mkey:%v mOperator:%v is placed on the node: %v having label key:(%v:%v)", matchExp.Key, matchExp.Operator, node.Id, matchExp.Key, tval)
								return true, nil
							}
						}
						return false, &ErrFailedToInspectVolume{
							Cause: fmt.Sprintf("ReplicaAffinity with matchExpression Operator:'%v' is placed on node:%v not having label key value:(%v:%v)", matchExp.Operator, node.Id, matchExp.Key, tval),
						}
					case "Exists":
						logrus.Infof("ReplicaMatchExpression  mkey:%v mOperator:%v exist on the node: %v", matchExp.Key, matchExp.Operator, node.Id)
						return true, nil
					case "DoesNotExist":
						return false, &ErrFailedToInspectVolume{
							Cause: fmt.Sprintf("ReplicaAffinity with matchExpression Operator:'%v' is placed on node:%v having label key:(%v:%v)", matchExp.Operator, node.Id, matchExp.Key, tval),
						}
					case "Gt":
						itval, err := strconv.ParseInt(tval, 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("ReplicaAffinity with matchExpression Operator:'%v', unable to convert node value to integer, placed on node:%v having label key:(%v:%v)", matchExp.Operator, node.Id, matchExp.Key, tval),
							}
						}
						mValue, err := strconv.ParseInt(matchExp.Values[0], 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("ReplicaAffinity with matchExpression Operator:'%v', unable to convert key value(%v) to integer, placed on node:%v having label key:(%v:%v)", matchExp.Operator, matchExp.Values[0], node.Id, matchExp.Key, tval),
							}
						}

						// Check node value is greater than spec value
						if itval <= mValue {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("ReplicaAffinity with matchExpression Operator:'%v', is placed on node:%v having node value(%v) not greater than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, itval, mValue, node.Id, matchExp.Key, tval),
							}
						}
						logrus.Infof("ReplicaAffinity with matchExpression Operator:'%v', is placed on node:%v having node value(%v) greater than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, itval, mValue, node.Id, matchExp.Key, tval)
						return true, nil

					case "Lt":
						itval, err := strconv.ParseInt(tval, 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("ReplicaAffinity with matchExpression Operator:'%v', unable to convert node value to integer, placed on node:%v having label key:(%v:%v)", matchExp.Operator, node.Id, matchExp.Key, tval),
							}
						}
						mValue, err := strconv.ParseInt(matchExp.Values[0], 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("ReplicaAffinity with matchExpression Operator:'%v', unable to convert key value(%v) to integer, placed on node:%v having label key:(%v:%v)", matchExp.Operator, matchExp.Values[0], node.Id, matchExp.Key, tval),
							}
						}

						// Check node value is less than spec value
						if itval >= mValue {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("ReplicaAffinity with matchExpression Operator:'%v', is placed on node:%v having node value(%v) not less than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, itval, mValue, node.Id, matchExp.Key, tval),
							}
						}
						logrus.Infof("ReplicaAffinity with matchExpression Operator:'%v', is placed on node:%v having node value(%v) less than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, itval, mValue, node.Id, matchExp.Key, tval)
						return true, nil

					default:
						return false, &ErrFailedToInspectVolume{
							Cause: fmt.Sprintf("ReplicaAffinity with UNKNOWN matchExpression Operator:'%v' is placed on node:%v having label key:(%v:%v)", matchExp.Operator, node.Id, matchExp.Key, tval),
						}
					}

				} else {
					logrus.Infof("ReplicaMatchExpression key not found, check for  rule:%v  mkey:%v mOperator:%v mValues:%v tval:%v", node, matchExp.Key, matchExp.Operator, matchExp.Values, tval)
					switch matchExp.Operator {
					case "DoesNotExist":
						logrus.Infof("ReplicaMatchExpression  mkey:%v mOperator:%v does not exist on the node: %v", matchExp.Key, matchExp.Operator, node.Id)
						return true, nil
					case "Exists":
						return false, &ErrFailedToInspectVolume{
							Cause: fmt.Sprintf("ReplicaAffinity with matchExpression Operator:'Exists' is placed on node:%v not having label key:(%v:%v)", node.Id, matchExp.Key, tval),
						}
					default:
						return false, &ErrFailedToInspectVolume{
							Cause: fmt.Sprintf("ReplicaAffinity with matchExpression Operator:'%v' is placed on node:%v not having label key:(%v:%v)", matchExp.Operator, node.Id, matchExp.Key, tval),
						}
					}

				}
			}
		}
	}

	return false, &ErrFailedToInspectVolume{
		Cause: fmt.Sprintf("ReplicaAffinity with matchExpression Operator:'%v' could not be validated for node:%v  having label key:(%v)", matchExp.Key, node.Id, matchExp.Key),
	}
}

// ReplicaAntiAffinityMatchExpression
func (d *portworx) ReplicaAntiMatchExpression(node api.StorageNode, matchExp *talisman_v1beta1.LabelSelectorRequirement, poolUUID string) (bool, error) {
	logrus.Infof("ReplicaAntiMatchExpression started nodeId: %v matchExpression:%v ", node.Id, matchExp)

	logrus.Debugf("ReplicaAntiMatchExpression rule:%v  mkey:%v mOperator:%v mValues:%v", node, matchExp.Key, matchExp.Operator, matchExp.Values)
	if node.Pools != nil {
		// each storage  pool on the node  check  for pool lables
		for _, spool := range node.Pools {

			//Check the nodes storagepool on which this replica is residing
			if spool.Uuid == poolUUID {

				//Check whether the pool's label contain the key
				if tval, ok := spool.Labels[matchExp.Key]; ok {

					// Check for Operators: In, Exists, NotIn, DoesNotExist
					logrus.Infof("ReplicaAntiMatchExpression key found, check for  rule:%v  mkey:%v mOperator:%v mValues:%v tval:%v", node, matchExp.Key, matchExp.Operator, matchExp.Values, tval)
					switch matchExp.Operator {
					case "NotIn":
						for _, mValue := range matchExp.Values {
							if mValue == tval {
								//The node should not have the label value, hence return error
								return true, nil
							}
						}
						return false, &ErrFailedToInspectVolume{
							Cause: fmt.Sprintf("ReplicaAntiAffinity with matchExpression 'NotIn' with values %v is placed on node:%v not having label value:(%v:%v)", matchExp.Values, node.Id, matchExp.Key, tval),
						}
					case "In":
						for _, mValue := range matchExp.Values {
							if mValue == tval {
								//The node should not have the label value, hence return false
								return false, &ErrFailedToInspectVolume{
									Cause: fmt.Sprintf("ReplicaAntiMatchExpression  mkey:%v mOperator:%v is placed on the node: %v having label key:(%v:%v)", matchExp.Key, matchExp.Operator, node.Id, matchExp.Key, tval)}
							}
						}
						return true, nil
					case "Exists":
						return false, &ErrFailedToInspectVolume{
							Cause: fmt.Sprintf("ReplicaAntiAffinity with matchExpression Operator:'%v' is placed on node:%v having label key:(%v:%v)", matchExp.Operator, node.Id, matchExp.Key, tval),
						}
					case "DoesNotExist":
						logrus.Infof("ReplicaAntiMatchExpression  mkey:%v mOperator:%v exist on the node: %v", matchExp.Key, matchExp.Operator, node.Id)
						return true, nil
					case "Gt":
						itval, err := strconv.ParseInt(tval, 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("ReplicaAntiAffinity with matchExpression Operator:'%v', unable to convert node value to integer, placed on node:%v having label key:(%v:%v)", matchExp.Operator, node.Id, matchExp.Key, tval),
							}
						}
						mValue, err := strconv.ParseInt(matchExp.Values[0], 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("ReplicaAntiAffinity with matchExpression Operator:'%v', unable to convert key value(%v) to integer, placed on node:%v having label key:(%v:%v)", matchExp.Operator, matchExp.Values[0], node.Id, matchExp.Key, tval),
							}
						}

						// Check node value is greater than spec value
						if itval <= mValue {
							logrus.Infof("ReplicaAntiAffinity with matchExpression Operator:'%v', is placed on node:%v having node value(%v) less than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, itval, mValue, node.Id, matchExp.Key, tval)
							return true, nil
						}
						return false, &ErrFailedToInspectVolume{
							Cause: fmt.Sprintf("ReplicaAntiAffinity with matchExpression Operator:'%v', is placed on node:%v having node value(%v)  greater than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, itval, mValue, node.Id, matchExp.Key, tval),
						}

					case "Lt":
						itval, err := strconv.ParseInt(tval, 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("ReplicaAntiAffinity with matchExpression Operator:'%v', unable to convert node value to integer, placed on node:%v having label key:(%v:%v)", matchExp.Operator, node.Id, matchExp.Key, tval),
							}
						}
						mValue, err := strconv.ParseInt(matchExp.Values[0], 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("ReplicaAntiAffinity with matchExpression Operator:'%v', unable to convert key value(%v) to integer, placed on node:%v having label key:(%v:%v)", matchExp.Operator, matchExp.Values[0], node.Id, matchExp.Key, tval),
							}
						}

						// Check node value is less than spec value
						if itval >= mValue {
							logrus.Infof("ReplicaAntiAffinity with matchExpression Operator:'%v', is placed on node:%v having node value(%v) greater than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, itval, mValue, node.Id, matchExp.Key, tval)
							return true, nil
						}
						return false, &ErrFailedToInspectVolume{
							Cause: fmt.Sprintf("ReplicaAntiAffinity with matchExpression Operator:'%v', is placed on node:%v having node value(%v)  less than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, itval, mValue, node.Id, matchExp.Key, tval),
						}

					default:
						return false, &ErrFailedToInspectVolume{
							Cause: fmt.Sprintf("ReplicaAntiAffinity with UNKNOWN matchExpression Operator:'%v' is placed on node:%v having label key:(%v:%v)", matchExp.Operator, node.Id, matchExp.Key, tval),
						}
					}

				} else {
					logrus.Infof("ReplicaAntiMatchExpression key not found, check for  rule:%v  mkey:%v mOperator:%v mValues:%v tval:%v", node, matchExp.Key, matchExp.Operator, matchExp.Values, tval)
					switch matchExp.Operator {
					case "DoesNotExist":
						return false, &ErrFailedToInspectVolume{
							Cause: fmt.Sprintf("ReplicaAntiAffinity with matchExpression Operator:'%v' is placed on node:%v not having label key:(%v:%v)", matchExp.Operator, node.Id, matchExp.Key, tval),
						}
					case "Exists":
						logrus.Infof("ReplicaAntiMatchExpression  mkey:%v mOperator:%v does not exist on the node: %v", matchExp.Key, matchExp.Operator, node.Id)
						return true, nil
					default:
						return false, &ErrFailedToInspectVolume{
							Cause: fmt.Sprintf("ReplicaAntiAffinity with matchExpression Operator:'%v' is placed on node:%v not having label key:(%v:%v)", matchExp.Operator, node.Id, matchExp.Key, tval),
						}
					}

				}
			}
		}
	}

	return false, &ErrFailedToInspectVolume{
		Cause: fmt.Sprintf("ReplicaAntiAffinity with matchExpression Operator:'%v' could not be validated for node:%v  having label key:(%v)", matchExp.Key, node.Id, matchExp.Key),
	}
}

// VolumeMatchExpression
func (d *portworx) VolumeMatchExpression(vol *api.Volume, matchExp *talisman_v1beta1.LabelSelectorRequirement, nodeReplList map[string][]*api.Volume, nodeGrp map[string]api.StorageNode) (bool, error) {
	logrus.Infof("VolumeMatchExpression started nodeId: %v matchExpression:%v ", vol.Id, matchExp)

	logrus.Debugf("VolumeMatchExpression rule:%v  mkey:%v mOperator:%v mValues:%v nodeReplList:%v, nodeGrp:%v", vol, matchExp.Key, matchExp.Operator, matchExp.Values, nodeReplList, nodeGrp)
	if vol != nil {
		// Check for Operators: In, Exists, NotIn, DoesNotExist
		switch matchExp.Operator {
		case "NotIn":
			//for each node in the node group
			for _, node := range nodeGrp {

				logrus.Debugf("VolumeMatchExpression: volume replicas on node %v: %v", node.Id, nodeReplList[node.Id])
				// for each volume on the node
				for _, nodeVol := range nodeReplList[node.Id] {
					// Check whether node is having the same volume or clone
					if nodeVol.Id == vol.Id || nodeVol.Source.Parent != "" {
						continue
					}
					if tval, ok := nodeVol.Spec.VolumeLabels[matchExp.Key]; ok {
						for _, mValue := range matchExp.Values {
							// For local snapshots, the snap and clone will be placed on the same node as the volume itself,
							// hence skip these volumes even if value is matched
							// TODO: Find a better way to find whether the volumes are related
							if mValue == tval {
								//The node should not have the label value, hence return error
								return false, &ErrFailedToInspectVolume{
									Cause: fmt.Sprintf("VolumeAffinity with matchExpression 'NotIn' with values %v is placed on node:%v having volume(%v) with label value:(%v:%v)", matchExp.Values, node.Id, nodeVol.Id, matchExp.Key, tval),
								}
							}
						}
					}
				}

				//Check for node labels being used in matchExpression
				for _, spool := range node.Pools {
					if tval, ok := spool.Labels[matchExp.Key]; ok {
						for _, mValue := range matchExp.Values {
							if mValue == tval {
								//The node should not have the label value, hence return error
								return false, &ErrFailedToInspectVolume{
									Cause: fmt.Sprintf("VolumeAffinity with matchExpression 'NotIn' with values %v is placed on node:%v having label value:(%v:%v)", matchExp.Values, node.Id, matchExp.Key, tval),
								}
							}
						}
					}
				}

			}
			return true, nil
		case "In":
			//for each node in the node group
			for _, node := range nodeGrp {

				logrus.Debugf("VolumeMatchExpression: volume replicas on node %v: %v", node.Id, nodeReplList[node.Id])
				// for each volume on the node
				for _, nodeVol := range nodeReplList[node.Id] {
					// Check whether node is having the same volume or clone
					if nodeVol.Id == vol.Id || nodeVol.Source.Parent != "" {
						logrus.Infof("VolumeMatchExpression  mkey:%v mOperator:%v  for volume %v with volume %v having source %v", matchExp.Key, matchExp.Operator, vol.Id, nodeVol.Id, nodeVol.Source)
						continue
					}
					logrus.Debugf("VolumeMatchExpression: mkey:%v, matchExp.Values:%v , VolumeLabels:%v", matchExp.Key, matchExp.Values, nodeVol.Spec.VolumeLabels)
					if tval, ok := nodeVol.Spec.VolumeLabels[matchExp.Key]; ok {
						logrus.Debugf("VolumeMatchExpression: Found key mkey:%v:%v, matchExp.Values:%v , VolumeLabels:%v", matchExp.Key, tval, matchExp.Values, nodeVol.Spec.VolumeLabels)

						for _, mValue := range matchExp.Values {
							logrus.Debugf("VolumeMatchExpression:  mkey:%v:%v, mValue:%v, matchExp.Values:%v , VolumeLabels:%v", matchExp.Key, tval, mValue, matchExp.Values, nodeVol.Spec.VolumeLabels)
							if mValue == tval {
								//The node should have the label value, hence return true
								logrus.Infof("VolumeMatchExpression  mkey:%v mOperator:%v is placed on the node: %v having volume (%v) volume label key:(%v:%v)", matchExp.Key, matchExp.Operator, node.Id, nodeVol.Id, matchExp.Key, tval)
								return true, nil
							}
						}
					}
				}

				//Check for node labels being used in matchExpression
				for _, spool := range node.Pools {
					if tval, ok := spool.Labels[matchExp.Key]; ok {
						for _, mValue := range matchExp.Values {
							if mValue == tval {
								//The node should have the label value, hence return error
								logrus.Infof("VolumeMatchExpression  mkey:%v mOperator:%v is placed on the node: %v having label key:(%v:%v)", matchExp.Key, matchExp.Operator, node.Id, matchExp.Key, tval)
								return true, nil
							}
						}
					}
				}
			}
			return false, &ErrFailedToInspectVolume{
				Cause: fmt.Sprintf("VolumeAffinity with matchExpression Operator:'%v' could not find another volume having volume label key:(%v)", matchExp.Operator, matchExp.Key),
			}
		case "Exists":
			//for each node in the node group
			for _, node := range nodeGrp {

				logrus.Debugf("VolumeMatchExpression: volume replicas on node %v: %v", node.Id, nodeReplList[node.Id])
				// for each volume on the node
				for _, nodeVol := range nodeReplList[node.Id] {
					// Check whether node is having the same volume or clone
					if nodeVol.Id == vol.Id || nodeVol.Source.Parent != "" {
						logrus.Infof("VolumeMatchExpression  mkey:%v mOperator:%v  for volume %v with volume %v having source %v", matchExp.Key, matchExp.Operator, vol.Id, nodeVol.Id, nodeVol.Source)
						continue
					}
					if _, ok := nodeVol.Spec.VolumeLabels[matchExp.Key]; ok {
						logrus.Infof("VolumeMatchExpression  mkey:%v mOperator:%v exist on the node: %v for volume %v", matchExp.Key, matchExp.Operator, node.Id, nodeVol.Id)
						return true, nil
					}
				}

				//Check for node labels being used in matchExpression
				for _, spool := range node.Pools {
					if _, ok := spool.Labels[matchExp.Key]; ok {
						//The node should have the key, hence return true
						logrus.Infof("VolumeMatchExpression  mkey:%v mOperator:%v exist on the node: %v ", matchExp.Key, matchExp.Operator, node.Id)
						return true, nil
					}
				}

			}
			return false, &ErrFailedToInspectVolume{
				Cause: fmt.Sprintf("VolumeAffinity with matchExpression Operator:'%v' could not find other another volume with key:%v", matchExp.Operator, matchExp.Key),
			}
		case "DoesNotExist":
			//for each node in the node group
			for _, node := range nodeGrp {

				// for each volume on the node
				for _, nodeVol := range nodeReplList[node.Id] {
					// Check whether node is having the same volume or clone
					if nodeVol.Id == vol.Id || nodeVol.Source.Parent != "" {
						continue
					}
					if tval, ok := nodeVol.Spec.VolumeLabels[matchExp.Key]; ok {
						return false, &ErrFailedToInspectVolume{
							Cause: fmt.Sprintf("VolumeAffinity with matchExpression Operator:'%v' is placed on node:%v having volume(%v) with volume label key:(%v:%v)", matchExp.Operator, node.Id, nodeVol.Id, matchExp.Key, tval),
						}
					}
				}

				//Check for node labels being used in matchExpression
				for _, spool := range node.Pools {
					if _, ok := spool.Labels[matchExp.Key]; ok {
						//The node should not have the key, hence return error
						return false, &ErrFailedToInspectVolume{
							Cause: fmt.Sprintf("VolumeAffinity with matchExpression Operator:'%v' is placed on node:%v having key:(%v)", matchExp.Operator, node.Id, matchExp.Key),
						}
					}
				}
			}
			return true, nil
		case "Gt":
			//for each node in the node group
			for _, node := range nodeGrp {

				logrus.Debugf("VolumeMatchExpression: volume replicas on node %v: %v", node.Id, nodeReplList[node.Id])
				// for each volume on the node
				for _, nodeVol := range nodeReplList[node.Id] {
					// Check whether node is having the same volume or clone
					if nodeVol.Id == vol.Id || nodeVol.Source.Parent != "" {
						continue
					}
					if tval, ok := nodeVol.Spec.VolumeLabels[matchExp.Key]; ok {
						itval, err := strconv.ParseInt(tval, 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAffinity with matchExpression Operator:'%v', unable to convert node value to integer, placed on node:%v for volume(%v), having label key:(%v:%v)", matchExp.Operator, node.Id, nodeVol.Id, matchExp.Key, tval),
							}
						}
						mValue, err := strconv.ParseInt(matchExp.Values[0], 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAffinity with matchExpression Operator:'%v', unable to convert key value(%v) to integer, placed on node:%v for volume(%v), having label key:(%v:%v)", matchExp.Operator, matchExp.Values[0], node.Id, nodeVol.Id, matchExp.Key, tval),
							}
						}

						// Check node value is greater than spec value
						if itval <= mValue {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAffinity with matchExpression Operator:'%v', is placed on node:%v for volume %v, having node value(%v) not greater than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, node.Id, nodeVol.Id, itval, mValue, matchExp.Key, tval),
							}
						}
						logrus.Infof("VolumeAffinity with matchExpression Operator:'%v', is placed on node:%v having volume(%v) label value(%v) greater than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, node.Id, nodeVol.Id, itval, mValue, matchExp.Key, tval)
						return true, nil

					}
				}

				//Check for node labels being used in matchExpression
				for _, spool := range node.Pools {
					if tval, ok := spool.Labels[matchExp.Key]; ok {
						itval, err := strconv.ParseInt(tval, 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAffinity with matchExpression Operator:'%v', unable to convert node value to integer, placed on node:%v , having label key:(%v:%v)", matchExp.Operator, node.Id, matchExp.Key, tval),
							}
						}
						mValue, err := strconv.ParseInt(matchExp.Values[0], 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAffinity with matchExpression Operator:'%v', unable to convert key value(%v) to integer, placed on node:%v, having label key:(%v:%v)", matchExp.Operator, matchExp.Values[0], node.Id, matchExp.Key, tval),
							}
						}

						// Check node value is greater than spec value
						if itval <= mValue {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAffinity with matchExpression Operator:'%v', is placed on node:%v having node value(%v) not greater than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, node.Id, itval, mValue, matchExp.Key, tval),
							}
						}
						logrus.Infof("VolumeAffinity with matchExpression Operator:'%v', is placed on node:%v having label value(%v) greater than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, node.Id, itval, mValue, matchExp.Key, tval)
						return true, nil

					}
				}
			}
			return false, &ErrFailedToInspectVolume{
				Cause: fmt.Sprintf("VolumeAffinity with matchExpression Operator:'%v' could not find another volume with key:%v", matchExp.Operator, matchExp.Key),
			}

		case "Lt":
			//for each node in the node group
			for _, node := range nodeGrp {

				logrus.Debugf("VolumeMatchExpression: volume replicas on node %v: %v", node.Id, nodeReplList[node.Id])
				// for each volume on the node
				for _, nodeVol := range nodeReplList[node.Id] {
					// Check whether node is having the same volume or clone
					if nodeVol.Id == vol.Id || nodeVol.Source.Parent != "" {
						continue
					}
					if tval, ok := nodeVol.Spec.VolumeLabels[matchExp.Key]; ok {

						itval, err := strconv.ParseInt(tval, 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAffinity with matchExpression Operator:'%v', unable to convert node value to integer, placed on node:%v for volume(%v), having label key:(%v:%v)", matchExp.Operator, node.Id, nodeVol.Id, matchExp.Key, tval),
							}
						}
						mValue, err := strconv.ParseInt(matchExp.Values[0], 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAffinity with matchExpression Operator:'%v', unable to convert key value(%v) to integer, placed on node:%v for volume(%v),having label key:(%v:%v)", matchExp.Operator, matchExp.Values[0], node.Id, nodeVol.Id, matchExp.Key, tval),
							}
						}

						// Check node value is less than spec value
						if itval >= mValue {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAffinity with matchExpression Operator:'%v', is placed on node:%v for volume(%v), having node value(%v) not less than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, node.Id, nodeVol.Id, itval, mValue, matchExp.Key, tval),
							}
						}
						logrus.Infof("VolumeAffinity with matchExpression Operator:'%v', is placed on node:%v for volume(%v), having node value(%v) less than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, node.Id, nodeVol.Id, itval, mValue, matchExp.Key, tval)
						return true, nil

					}
				}

				for _, spool := range node.Pools {
					if tval, ok := spool.Labels[matchExp.Key]; ok {
						itval, err := strconv.ParseInt(tval, 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAffinity with matchExpression Operator:'%v', unable to convert node value to integer, placed on node:%v, having label key:(%v:%v)", matchExp.Operator, node.Id, matchExp.Key, tval),
							}
						}
						mValue, err := strconv.ParseInt(matchExp.Values[0], 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAffinity with matchExpression Operator:'%v', unable to convert key value(%v) to integer, placed on node:%v,having label key:(%v:%v)", matchExp.Operator, matchExp.Values[0], node.Id, matchExp.Key, tval),
							}
						}

						// Check node value is less than spec value
						if itval >= mValue {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAffinity with matchExpression Operator:'%v', is placed on node:%v , having node value(%v) not less than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, node.Id, itval, mValue, matchExp.Key, tval),
							}
						}
						logrus.Infof("VolumeAffinity with matchExpression Operator:'%v', is placed on node:%v, having node value(%v) less than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, node.Id, itval, mValue, matchExp.Key, tval)
						return true, nil

					}
				}

			}
			return false, &ErrFailedToInspectVolume{
				Cause: fmt.Sprintf("VolumeAffinity with matchExpression Operator:'%v' could not find another volume with key:%v", matchExp.Operator, matchExp.Key),
			}

		default:
			return false, &ErrFailedToInspectVolume{
				Cause: fmt.Sprintf("VolumeAffinity with UNKNOWN matchExpression Operator:'%v'  having  key:%v", matchExp.Operator, matchExp.Key),
			}
		}

	}
	return false, &ErrFailedToInspectVolume{
		Cause: fmt.Sprintf("VolumeAffinity could not be validated for volume:%v", vol.Id),
	}
}

// VolumeAntiMatchExpression
// matching label volumes should not co-exists on same node or same topology group
func (d *portworx) VolumeAntiMatchExpression(vol *api.Volume, matchExp *talisman_v1beta1.LabelSelectorRequirement, nodeReplList map[string][]*api.Volume, nodeGrp map[string]api.StorageNode) (bool, error) {
	logrus.Infof("VolumeAntiMatchExpression started nodeId: %v matchExpression:%v ", vol.Id, matchExp)

	logrus.Debugf("VolumeAntiMatchExpression rule:%v  mkey:%v mOperator:%v mValues:%v", vol, matchExp.Key, matchExp.Operator, matchExp.Values)
	if vol != nil {
		// Check for Operators: In, Exists, NotIn, DoesNotExist
		switch matchExp.Operator {
		case "NotIn":
			//for each node in the node group
			for _, node := range nodeGrp {
				logrus.Debugf("VolumeMatchExpression: volume replicas on node %v: %v", node.Id, nodeReplList[node.Id])

				// for each volume on the node
				for _, nodeVol := range nodeReplList[node.Id] {
					// Check whether node is having the same volume or clone
					if nodeVol.Id == vol.Id || nodeVol.Source.Parent != "" {
						continue
					}
					if tval, ok := nodeVol.Spec.VolumeLabels[matchExp.Key]; ok {
						for _, mValue := range matchExp.Values {
							if mValue != tval {
								//The node should  have the label value, hence return error
								return false, &ErrFailedToInspectVolume{
									Cause: fmt.Sprintf("VolumeAntiAffinity with matchExpression 'NotIn' with values %v is placed on node:%v having volume(%v) with label key value:(%v:%v)", matchExp.Values, node.Id, nodeVol.Id, matchExp.Key, tval),
								}
							}
						}
					}
				}

				//Check for node labels being used in matchExpression
				for _, spool := range node.Pools {
					if tval, ok := spool.Labels[matchExp.Key]; ok {
						for _, mValue := range matchExp.Values {
							if mValue != tval {
								//The node should not have the label value, hence return error
								return false, &ErrFailedToInspectVolume{
									Cause: fmt.Sprintf("VolumeAntiAffinity with matchExpression 'NotIn' with values %v is placed on node:%v having label value:(%v:%v)", matchExp.Values, node.Id, matchExp.Key, tval),
								}
							}
						}
					}
				}

			}
			return true, nil
		case "In":
			//for each node in the node group
			for _, node := range nodeGrp {

				logrus.Debugf("VolumeMatchExpression: volume replicas on node %v: %v", node.Id, nodeReplList[node.Id])
				// for each volume on the node
				for _, nodeVol := range nodeReplList[node.Id] {
					// Check whether node is having the same volume or clone
					if nodeVol.Id == vol.Id || nodeVol.Source.Parent != "" {
						continue
					}
					if tval, ok := nodeVol.Spec.VolumeLabels[matchExp.Key]; ok {
						for _, mValue := range matchExp.Values {
							if mValue == tval {
								//The node should have the label value, hence return true
								return false, &ErrFailedToInspectVolume{
									Cause: fmt.Sprintf("VolumeMatchExpression  mkey:%v mOperator:%v is placed on the node: %v having volume (%v) volume label key:(%v:%v)", matchExp.Key, matchExp.Operator, node.Id, nodeVol.Id, matchExp.Key, tval),
								}
							}
						}
					}
				}

				//Check for node labels being used in matchExpression
				for _, spool := range node.Pools {
					if tval, ok := spool.Labels[matchExp.Key]; ok {
						for _, mValue := range matchExp.Values {
							if mValue == tval {
								//The node should not have the label value, hence return error
								return false, &ErrFailedToInspectVolume{
									Cause: fmt.Sprintf("VolumeAntiMatchExpression  mkey:%v mOperator:%v is placed on the node: %v having label key:(%v:%v)", matchExp.Key, matchExp.Operator, node.Id, matchExp.Key, tval),
								}
							}
						}
					}
				}
			}
			return true, nil
		case "Exists":
			//for each node in the node group
			for _, node := range nodeGrp {
				logrus.Debugf("VolumeMatchExpression: volume replicas on node %v: %v", node.Id, nodeReplList[node.Id])

				// for each volume on the node
				for _, nodeVol := range nodeReplList[node.Id] {
					// Check whether node is having the same volume or clone
					if nodeVol.Id == vol.Id || nodeVol.Source.Parent != "" {
						continue
					}
					if _, ok := nodeVol.Spec.VolumeLabels[matchExp.Key]; ok {

						//TODO: check whether the volume is parent or child of the current volume
						// Assuming if the volume source is not equal to nil, then they are related to each other

						return false, &ErrFailedToInspectVolume{
							Cause: fmt.Sprintf("VolumeAntiMatchExpression  mkey:%v mOperator:%v exist on the node: %v for volume %v", matchExp.Key, matchExp.Operator, node.Id, nodeVol.Id),
						}
					}
				}

				//Check for node labels being used in matchExpression
				for _, spool := range node.Pools {
					if _, ok := spool.Labels[matchExp.Key]; ok {
						//The node should have the key, hence return true
						return false, &ErrFailedToInspectVolume{
							Cause: fmt.Sprintf("VolumeAntiMatchExpression  mkey:%v mOperator:%v exist on the node: %v ", matchExp.Key, matchExp.Operator, node.Id),
						}
					}
				}

			}
			return true, nil
		case "DoesNotExist":
			//for each node in the node group
			for _, node := range nodeGrp {
				logrus.Debugf("VolumeMatchExpression: volume replicas on node %v: %v", node.Id, nodeReplList[node.Id])

				// for each volume on the node
				for _, nodeVol := range nodeReplList[node.Id] {
					// Check whether node is having the same volume or clone
					if nodeVol.Id == vol.Id || nodeVol.Source.Parent != "" {
						continue
					}
					if tval, ok := nodeVol.Spec.VolumeLabels[matchExp.Key]; ok {
						logrus.Infof("VolumeAntiAffinity with matchExpression Operator:'%v' is placed on node:%v having volume(%v) with volume label key:(%v:%v)", matchExp.Operator, node.Id, nodeVol.Id, matchExp.Key, tval)
						return true, nil
					}
				}

				//Check for node labels being used in matchExpression
				for _, spool := range node.Pools {
					if _, ok := spool.Labels[matchExp.Key]; ok {
						//The node should not have the key, hence return error
						logrus.Infof("VolumeAntiAffinity with matchExpression Operator:'%v' is placed on node:%v having key:(%v)", matchExp.Operator, node.Id, matchExp.Key)
						return true, nil
					}
				}
			}
			return false, &ErrFailedToInspectVolume{
				Cause: fmt.Sprintf("VolumeAntiAffinity with matchExpression Operator:'%v' for volume :%v with volume key:(%v) could not find another node/volume have the key", matchExp.Operator, vol.Id, matchExp.Key),
			}
		case "Gt":
			//for each node in the node group
			for _, node := range nodeGrp {
				logrus.Debugf("VolumeMatchExpression: volume replicas on node %v: %v", node.Id, nodeReplList[node.Id])

				// for each volume on the node
				for _, nodeVol := range nodeReplList[node.Id] {
					// Check whether node is having the same volume or clone
					if nodeVol.Id == vol.Id || nodeVol.Source.Parent != "" {
						continue
					}
					if tval, ok := nodeVol.Spec.VolumeLabels[matchExp.Key]; ok {
						itval, err := strconv.ParseInt(tval, 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAntiAffinity with matchExpression Operator:'%v', unable to convert node value to integer, placed on node:%v for volume(%v), having label key:(%v:%v)", matchExp.Operator, node.Id, nodeVol.Id, matchExp.Key, tval),
							}
						}
						mValue, err := strconv.ParseInt(matchExp.Values[0], 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAntiAffinity with matchExpression Operator:'%v', unable to convert key value(%v) to integer, placed on node:%v for volume(%v), having label key:(%v:%v)", matchExp.Operator, matchExp.Values[0], node.Id, nodeVol.Id, matchExp.Key, tval),
							}
						}

						// Check node value is greater than spec value
						if itval >= mValue {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAntiAffinity with matchExpression Operator:'%v', is placed on node:%v for volume %v, having node value(%v)  greater than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, node.Id, nodeVol.Id, itval, mValue, matchExp.Key, tval),
							}
						}
						logrus.Infof("VolumeAntiAffinity with matchExpression Operator:'%v', is placed on node:%v having volume(%v) label value(%v) less than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, node.Id, nodeVol.Id, itval, mValue, matchExp.Key, tval)
						return true, nil

					}
				}

				//Check for node labels being used in matchExpression
				for _, spool := range node.Pools {
					if tval, ok := spool.Labels[matchExp.Key]; ok {
						itval, err := strconv.ParseInt(tval, 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAntiAffinity with matchExpression Operator:'%v', unable to convert node value to integer, placed on node:%v , having label key:(%v:%v)", matchExp.Operator, node.Id, matchExp.Key, tval),
							}
						}
						mValue, err := strconv.ParseInt(matchExp.Values[0], 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAntiAffinity with matchExpression Operator:'%v', unable to convert key value(%v) to integer, placed on node:%v, having label key:(%v:%v)", matchExp.Operator, matchExp.Values[0], node.Id, matchExp.Key, tval),
							}
						}

						// Check node value is greater than spec value
						if itval >= mValue {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAntiAffinity with matchExpression Operator:'%v', is placed on node:%v having node value(%v) greater than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, node.Id, itval, mValue, matchExp.Key, tval),
							}
						}
						logrus.Infof("VolumeAntiAffinity with matchExpression Operator:'%v', is placed on node:%v having label value(%v) not greater than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, node.Id, itval, mValue, matchExp.Key, tval)
						return true, nil

					}
				}
			}
			return false, &ErrFailedToInspectVolume{
				Cause: fmt.Sprintf("VolumeAntiAffinity with matchExpression Operator:'%v' could not find another volume with key:%v", matchExp.Operator, matchExp.Key),
			}

		case "Lt":
			//for each node in the node group
			for _, node := range nodeGrp {
				logrus.Debugf("VolumeMatchExpression: volume replicas on node %v: %v", node.Id, nodeReplList[node.Id])

				// for each volume on the node
				for _, nodeVol := range nodeReplList[node.Id] {
					// Check whether node is having the same volume or clone
					if nodeVol.Id == vol.Id || nodeVol.Source.Parent != "" {
						continue
					}
					if tval, ok := nodeVol.Spec.VolumeLabels[matchExp.Key]; ok {

						itval, err := strconv.ParseInt(tval, 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAntiAffinity with matchExpression Operator:'%v', unable to convert node value to integer, placed on node:%v for volume(%v), having label key:(%v:%v)", matchExp.Operator, node.Id, nodeVol.Id, matchExp.Key, tval),
							}
						}
						mValue, err := strconv.ParseInt(matchExp.Values[0], 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAntiAffinity with matchExpression Operator:'%v', unable to convert key value(%v) to integer, placed on node:%v for volume(%v),having label key:(%v:%v)", matchExp.Operator, matchExp.Values[0], node.Id, nodeVol.Id, matchExp.Key, tval),
							}
						}

						// Check node value is less than spec value
						if itval <= mValue {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAntiAffinity with matchExpression Operator:'%v', is placed on node:%v for volume(%v), having node value(%v) less than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, node.Id, nodeVol.Id, itval, mValue, matchExp.Key, tval),
							}
						}
						logrus.Infof("VolumeAntiAffinity with matchExpression Operator:'%v', is placed on node:%v for volume(%v), having node value(%v) notless than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, node.Id, nodeVol.Id, itval, mValue, matchExp.Key, tval)
						return true, nil

					}
				}

				for _, spool := range node.Pools {
					if tval, ok := spool.Labels[matchExp.Key]; ok {
						itval, err := strconv.ParseInt(tval, 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAntiAffinity with matchExpression Operator:'%v', unable to convert node value to integer, placed on node:%v, having label key:(%v:%v)", matchExp.Operator, node.Id, matchExp.Key, tval),
							}
						}
						mValue, err := strconv.ParseInt(matchExp.Values[0], 10, 64)
						if err != nil {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAntiAffinity with matchExpression Operator:'%v', unable to convert key value(%v) to integer, placed on node:%v,having label key:(%v:%v)", matchExp.Operator, matchExp.Values[0], node.Id, matchExp.Key, tval),
							}
						}

						// Check node value is less than spec value
						if itval <= mValue {
							return false, &ErrFailedToInspectVolume{
								Cause: fmt.Sprintf("VolumeAntiAffinity with matchExpression Operator:'%v', is placed on node:%v , having node value(%v) less than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, node.Id, itval, mValue, matchExp.Key, tval),
							}
						}
						logrus.Infof("VolumeAntiAffinity with matchExpression Operator:'%v', is placed on node:%v, having node value(%v) not less than Spec value (%v)  for key:(%v:%v)", matchExp.Operator, node.Id, itval, mValue, matchExp.Key, tval)
						return true, nil

					}
				}

			}
			return false, &ErrFailedToInspectVolume{
				Cause: fmt.Sprintf("VolumeAntiAffinity with matchExpression Operator:'%v' could not find another volume with key:%v", matchExp.Operator, matchExp.Key),
			}

		default:
			return false, &ErrFailedToInspectVolume{
				Cause: fmt.Sprintf("VolumeAntiAffinity with UNKNOWN matchExpression Operator:'%v'  having  key:%v", matchExp.Operator, matchExp.Key),
			}
		}

	}
	return false, &ErrFailedToInspectVolume{
		Cause: fmt.Sprintf("VolumeAntiAffinity could not be validated for volume:%v", vol.Id),
	}
}

// ReplicaAffinity Validate  module
func (d *portworx) ValidateReplicaAffinity(vol *api.Volume, vpsRule *talisman_v1beta2.ReplicaPlacementSpec, volNodes []api.StorageNode) error {
	logrus.Infof("ValidateReplicaAffinity for vol: %v:%v and rule:%v ", vol.Id, vol, vpsRule)

	if vpsRule != nil {
		//Create node group list base on topology Key set
		nodeGrpList := d.GroupTopologyNodes(vpsRule.CommonPlacementSpec.TopologyKey, volNodes)

		//logrus.Infof("GroupTopologyNodes returned nodes for topology key: %v  nodelist: %v", vpsRule.CommonPlacementSpec.TopologyKey, nodeGrpList)
		replPoolList := d.getReplicaPoolMap(vol)

		// For each node group check volume replica exists
		for _, nodes := range nodeGrpList {

			// Does the rule have MatchExpression ,
			if vpsRule.CommonPlacementSpec.MatchExpressions != nil {
				for _, node := range nodes {

					for _, matchExp := range vpsRule.CommonPlacementSpec.MatchExpressions {
						for _, replPool := range replPoolList {
							if _, ok := replPool[node.Id]; ok {
								logrus.Infof("ValidateReplicaAffinity for vol:%v and rule:%v MatchExpression %v", vol.Id, vpsRule, matchExp)
								status, err := d.ReplicaMatchExpression(node, matchExp, replPool[node.Id])
								logrus.Infof("ValidateReplicaAffinity for vol:%v and rule:%v MatchExpression %v Status:%v err: %v", vol.Id, vpsRule, matchExp, status, err)
								//EnforcementType_preferred EnforcementType = 1
								if (err != nil) && (vpsRule.CommonPlacementSpec.Enforcement != talisman_v1beta1.EnforcementPreferred) {
									return err
								}
							}
						}

					}
				}
			}

			//Does  the rule have topology key set
			if vpsRule.CommonPlacementSpec.TopologyKey != "" {

				for _, replPool := range replPoolList {
					//All replicas should be either present or none should present in the node group
					found := 0
					for nodeid := range replPool {
						if _, ok := nodes[nodeid]; ok {

							found++
						}
					}
					if found != 0 && found != len(replPool) && vpsRule.CommonPlacementSpec.Enforcement != talisman_v1beta1.EnforcementPreferred {
						return &ErrFailedToInspectVolume{
							Cause: fmt.Sprintf("ValidateReplicaAffinity for volume (%v) does not have all replicas (%v/%v) in the same topologykey(%v) group", vol.Id, found, len(replPool), vpsRule.CommonPlacementSpec.TopologyKey),
						}
					}
				}
			}

		}

	} else {
		logrus.Infof("ValidateReplicaAffinity for vol:%v and rule:%v is empty", vol.Id, vpsRule)
	}

	return nil
}

// ReplicaAntiAffinity Validate  module
func (d *portworx) ValidateReplicaAntiAffinity(vol *api.Volume, vpsRule *talisman_v1beta2.ReplicaPlacementSpec, volNodes []api.StorageNode) error {
	logrus.Infof("ValidateReplicaAntiAffinity for vol:%v and rule:%v ", vol, vpsRule)

	if vpsRule != nil {
		//Create node group list base on topology Key set
		nodeGrpList := d.GroupTopologyNodes(vpsRule.CommonPlacementSpec.TopologyKey, volNodes)

		//logrus.Infof("GroupTopologyNodes returned nodes for topology key: %v  nodelist: %v", vpsRule.CommonPlacementSpec.TopologyKey, nodeGrpList)
		replPoolList := d.getReplicaPoolMap(vol)

		//Total replicas found across all node group list
		volReplFound := 0

		// For each node group check volume replica exists
		for tpname, nodes := range nodeGrpList {

			// Does the rule have MatchExpression ,
			if vpsRule.CommonPlacementSpec.MatchExpressions != nil {
				for _, node := range nodes {

					for _, matchExp := range vpsRule.CommonPlacementSpec.MatchExpressions {
						for _, replPool := range replPoolList {
							if _, ok := replPool[node.Id]; ok {
								logrus.Infof("ValidateReplicaAntiAffinity for vol:%v and rule:%v MatchExpression %v", vol.Id, vpsRule, matchExp)
								status, err := d.ReplicaAntiMatchExpression(node, matchExp, replPool[node.Id])
								logrus.Infof("ValidateReplicaAntiAffinity for vol:%v and rule:%v MatchExpression %v Status:%v err: %v", vol.Id, vpsRule, matchExp, status, err)
								//EnforcementType_preferred EnforcementType = 1
								if (err != nil) && (vpsRule.CommonPlacementSpec.Enforcement != talisman_v1beta1.EnforcementPreferred) {
									return err
								}
							}
						}

					}
				}
			}

			//Does  the rule have topology key set
			if vpsRule.CommonPlacementSpec.TopologyKey != "" {
				//All replicas should be either present or none should present in the node group
				for _, replPool := range replPoolList {
					found := 0
					logrus.Debugf("ValidateReplicaAntiAffinity topologykey check: replPool:%v, nodes:%v group(%v), found:%v,volReplFound:%v", replPool, nodes, tpname, found, volReplFound)
					for nodeid := range replPool {
						if _, ok := nodes[nodeid]; ok {

							volReplFound++
							found++
						}
					}
					//There should be only one replica in a zone
					if found != 0 && found != 1 && vpsRule.CommonPlacementSpec.Enforcement != talisman_v1beta1.EnforcementPreferred {
						return &ErrFailedToInspectVolume{
							Cause: fmt.Sprintf("ValidateReplicaAntiAffinity for volume (%v) has more than one replica (%v/%v) (Total replicas found:%v) in the same topologykey(%v:%v) group", vol.Id, found, len(replPool), volReplFound, vpsRule.CommonPlacementSpec.TopologyKey, tpname),
						}
					}
				}
			}

		}

	} else {
		logrus.Infof("ValidateReplicaAntiAffinity for vol:%v and rule:%v is empty", vol.Id, vpsRule)
	}

	return nil
}

// VolumeAffinity Validate  module
func (d *portworx) ValidateVolumeAffinity(vol *api.Volume, vpsRule *talisman_v1beta2.CommonPlacementSpec, volNodes []api.StorageNode, nodeReplList map[string][]*api.Volume, appVolNodes map[string]api.StorageNode) error {
	logrus.Infof("ValidateVolumeAffinity for vol:%v and rule:%v ", vol, vpsRule)

	if vpsRule != nil {
		//Create node group list based on topology Key set
		nodeGrpList := d.GroupTopologyAppNodes(vpsRule.TopologyKey, appVolNodes)
		replPoolList := d.getReplicaPoolMap(vol)

		if vpsRule.MatchExpressions != nil {

			// for each replicaset
			for _, replPool := range replPoolList {
				// for each node of the replicaset
				for nodeid := range replPool {
					//Check to which nodegroup does this node belong to
					for _, nodes := range nodeGrpList {
						if _, ok := nodes[nodeid]; ok {
							//Volume Affinity will always have atleast one MatchExpression
							for _, matchExp := range vpsRule.MatchExpressions {
								//MatchExpression check
								logrus.Infof("ValidateVolumeAffinity for vol:%v and rule:%v MatchExpression %v", vol.Id, vpsRule, matchExp)
								status, err := d.VolumeMatchExpression(vol, matchExp, nodeReplList, nodes)
								logrus.Infof("ValidateVolumeAffinity for vol:%v and rule:%v MatchExpression %v Status:%v err:%v", vol.Id, vpsRule, matchExp, status, err)
								//EnforcementType_preferred EnforcementType = 1
								if (err != nil) && (vpsRule.Enforcement != talisman_v1beta1.EnforcementPreferred) {
									return err
								}
							}
						}
					}
				}
			}
		}

	} else {
		logrus.Infof("ValidateVolumeAffinity for vol:%v and rule:%v is empty", vol.Id, vpsRule)
	}

	return nil
}

// VolumeAntiAffinity Validate  module
func (d *portworx) ValidateVolumeAntiAffinity(vol *api.Volume, vpsRule *talisman_v1beta2.CommonPlacementSpec, volNodes []api.StorageNode, nodeReplList map[string][]*api.Volume, appVolNodes map[string]api.StorageNode) error {
	logrus.Infof("ValidateVolumeAntiAffinity for vol:%v and rule:%v ", vol, vpsRule)

	if vpsRule != nil {
		//Create node group list base on topology Key set
		nodeGrpList := d.GroupTopologyAppNodes(vpsRule.TopologyKey, appVolNodes)
		replPoolList := d.getReplicaPoolMap(vol)

		if vpsRule.MatchExpressions != nil {

			// for each replicaset
			for _, replPool := range replPoolList {
				// for each node of the replicaset
				for nodeid := range replPool {
					//Check whether the node belongs to the nodegroup
					for _, nodes := range nodeGrpList {
						if _, ok := nodes[nodeid]; ok {
							//Volume AntiAffinity will have atleast one MatchExpression
							for _, matchExp := range vpsRule.MatchExpressions {
								//MatchExpression check
								logrus.Infof("ValidateVolumeAntiAffinity for vol:%v and rule:%v MatchExpression %v", vol.Id, vpsRule, matchExp)
								status, err := d.VolumeAntiMatchExpression(vol, matchExp, nodeReplList, nodes)
								logrus.Infof("ValidateVolumeAntiAffinity for vol:%v and rule:%v MatchExpression %v Status:%v err:%v", vol.Id, vpsRule, matchExp, status, err)
								//EnforcementType_preferred EnforcementType = 1
								if (err != nil) && (vpsRule.Enforcement != talisman_v1beta1.EnforcementPreferred) {
									return err
								}
							}
						}
					}
				}
			}
		}

	} else {
		logrus.Infof("ValidateVolumeAntiAffinity for vol:%v and rule:%v is empty", vol.Id, vpsRule)
	}

	return nil
}

/*VolumePlacementStrategy  END*/

func (d *portworx) ValidateUpdateVolume(vol *torpedovolume.Volume, params map[string]string) error {
	var token string
	volumeName := d.schedOps.GetVolumeName(vol)
	token = d.getTokenForVolume(volumeName, params)
	t := func() (interface{}, bool, error) {
		volumeInspectResponse, err := d.getVolDriver().Inspect(d.getContextWithToken(context.Background(), token), &api.SdkVolumeInspectRequest{VolumeId: volumeName})
		if err != nil {
			return nil, true, err
		}
		return volumeInspectResponse.Volume, false, nil
	}

	out, err := task.DoRetryWithTimeout(t, inspectVolumeTimeout, inspectVolumeRetryInterval)
	if err != nil {
		return &ErrFailedToInspectVolume{
			ID:    volumeName,
			Cause: fmt.Sprintf("Volume inspect returned err: %v", err),
		}
	}

	respVol := out.(*api.Volume)

	// Size Update
	if respVol.Spec.Size != vol.RequestedSize {
		return &ErrFailedToInspectVolume{
			ID: volumeName,
			Cause: fmt.Sprintf("Volume size differs. Expected:%v Actual:%v",
				vol.RequestedSize, respVol.Spec.Size),
		}
	}
	return nil
}

func errIsNotFound(err error) bool {
	statusErr, _ := status.FromError(err)
	return statusErr.Code() == codes.NotFound || strings.Contains(err.Error(), "code = NotFound")
}

func (d *portworx) ValidateDeleteVolume(vol *torpedovolume.Volume) error {
	volumeName := d.schedOps.GetVolumeName(vol)
	t := func() (interface{}, bool, error) {
		volumeInspectResponse, err := d.getVolDriver().Inspect(d.getContext(), &api.SdkVolumeInspectRequest{VolumeId: volumeName})
		if err != nil && errIsNotFound(err) {
			return nil, false, nil
		} else if err != nil {
			return nil, true, err
		}
		// TODO remove shared validation when PWX-6894 and PWX-8790 are fixed
		if volumeInspectResponse.Volume != nil && !vol.Shared {
			return nil, true, fmt.Errorf("Volume %v is not yet removed from the system", volumeName)
		}
		return nil, false, nil
	}

	_, err := task.DoRetryWithTimeout(t, validateDeleteVolumeTimeout, defaultRetryInterval)
	if err != nil {
		return &ErrFailedToDeleteVolume{
			ID:    volumeName,
			Cause: err.Error(),
		}
	}

	return nil
}

func (d *portworx) ValidateVolumeCleanup() error {
	return d.schedOps.ValidateVolumeCleanup(d.nodeDriver)
}

func (d *portworx) ValidateVolumeSetup(vol *torpedovolume.Volume) error {
	return d.schedOps.ValidateVolumeSetup(vol, d.nodeDriver)
}

func (d *portworx) StopDriver(nodes []node.Node, force bool) error {
	var err error
	for _, n := range nodes {
		logrus.Infof("Stopping volume driver on %s.", n.Name)
		if force {
			pxCrashCmd := "sudo pkill -9 px-storage"
			_, err = d.nodeDriver.RunCommand(n, pxCrashCmd, node.ConnectionOpts{
				Timeout:         crashDriverTimeout,
				TimeBeforeRetry: defaultRetryInterval,
			})
			if err != nil {
				logrus.Warnf("failed to run cmd : %s. on node %s err: %v", pxCrashCmd, n.Name, err)
				return err
			}
			logrus.Infof("Sleeping for %v for volume driver to go down.", waitVolDriverToCrash)
			time.Sleep(waitVolDriverToCrash)
		} else {
			err = d.schedOps.StopPxOnNode(n)
			if err != nil {
				return err
			}
			err = d.nodeDriver.Systemctl(n, pxSystemdServiceName, node.SystemctlOpts{
				Action: "stop",
				ConnectionOpts: node.ConnectionOpts{
					Timeout:         stopDriverTimeout,
					TimeBeforeRetry: defaultRetryInterval,
				}})
			if err != nil {
				logrus.Warnf("failed to run systemctl stopcmd  on node %s err: %v", n.Name, err)
				return err
			}
			logrus.Infof("Sleeping for %v for volume driver to gracefully go down.", waitVolDriverToCrash/6)
			time.Sleep(waitVolDriverToCrash / 6)
		}

	}
	return nil
}

func (d *portworx) GetNodeForVolume(vol *torpedovolume.Volume, timeout time.Duration, retryInterval time.Duration) (*node.Node, error) {
	volumeName := d.schedOps.GetVolumeName(vol)
	r := func() (interface{}, bool, error) {
		t := func() (interface{}, bool, error) {
			volumeInspectResponse, err := d.getVolDriver().Inspect(d.getContext(), &api.SdkVolumeInspectRequest{VolumeId: volumeName})
			if err != nil {
				logrus.Warnf("Failed to inspect volume: %s due to: %v", volumeName, err)
				return nil, true, err
			}
			return volumeInspectResponse.Volume, false, nil
		}

		v, err := task.DoRetryWithTimeout(t, inspectVolumeTimeout, inspectVolumeRetryInterval)
		if err != nil {
			return nil, false, &ErrFailedToInspectVolume{
				ID:    volumeName,
				Cause: err.Error(),
			}
		}
		pxVol := v.(*api.Volume)
		for _, n := range node.GetStorageDriverNodes() {
			if isVolumeAttachedOnNode(pxVol, n) {
				return &n, false, nil
			}
		}

		// Snapshots may not be attached to a node
		if pxVol.Source.Parent != "" {
			return nil, false, nil
		}

		return nil, true, fmt.Errorf("Volume: %s is not attached on any node", volumeName)
	}

	n, err := task.DoRetryWithTimeout(r, timeout, retryInterval)
	if err != nil {
		return nil, &ErrFailedToValidateAttachment{
			ID:    volumeName,
			Cause: err.Error(),
		}
	}

	if n != nil {
		node := n.(*node.Node)
		return node, nil
	}

	return nil, nil
}

func isVolumeAttachedOnNode(volume *api.Volume, node node.Node) bool {
	if node.VolDriverNodeID == volume.AttachedOn {
		return true
	}
	for _, ip := range node.Addresses {
		if ip == volume.AttachedOn {
			return true
		}
	}
	return false
}

func (d *portworx) ExtractVolumeInfo(params string) (string, map[string]string, error) {
	ok, volParams, volumeName := spec.NewSpecHandler().SpecOptsFromString(params)
	if !ok {
		return params, nil, fmt.Errorf("Unable to parse the volume options")
	}
	return volumeName, volParams, nil
}

func (d *portworx) RandomizeVolumeName(params string) string {
	re := regexp.MustCompile("(name=)([0-9A-Za-z_-]+)(,)?")
	return re.ReplaceAllString(params, "${1}${2}_"+uuid.New()+"${3}")
}

func (d *portworx) getStorageNodesOnStart() ([]api.StorageNode, error) {
	t := func() (interface{}, bool, error) {
		cluster, err := d.getClusterManager().InspectCurrent(d.getContext(), &api.SdkClusterInspectCurrentRequest{})
		if err != nil {
			return nil, true, err
		}
		if cluster.Cluster.Status != api.Status_STATUS_OK {
			return nil, true, &ErrFailedToWaitForPx{
				Cause: fmt.Sprintf("px cluster is still not up. Status: %v", cluster.Cluster.Status),
			}
		}
		return &cluster.Cluster, false, nil
	}

	_, err := task.DoRetryWithTimeout(t, validateClusterStartTimeout, defaultRetryInterval)
	if err != nil {
		return nil, err
	}

	return d.getPxNodes()
}

func (d *portworx) getPxNodes(nManagers ...api.OpenStorageNodeClient) ([]api.StorageNode, error) {
	var nodeManager api.OpenStorageNodeClient
	if nManagers == nil {
		nodeManager = d.getNodeManager()
	} else {
		nodeManager = nManagers[0]
	}
	nodes := make([]api.StorageNode, 0)
	nodeEnumerateResp, err := nodeManager.Enumerate(d.getContext(), &api.SdkNodeEnumerateRequest{})
	if err != nil {
		return nodes, err
	}
	for _, n := range nodeEnumerateResp.GetNodeIds() {
		nodeResp, err := nodeManager.Inspect(d.getContext(), &api.SdkNodeInspectRequest{NodeId: n})
		if err != nil {
			return nodes, err
		}
		nodes = append(nodes, *nodeResp.Node)
	}
	return nodes, nil
}

func (d *portworx) WaitDriverUpOnNode(n node.Node, timeout time.Duration) error {
	logrus.Debugf("waiting for PX node to be up: %s", n.Name)
	t := func() (interface{}, bool, error) {
		logrus.Debugf("Getting node info: %s", n.Name)
		pxNode, err := d.getPxNode(&n)
		if err != nil {
			return "", true, &ErrFailedToWaitForPx{
				Node:  n,
				Cause: fmt.Sprintf("failed to get node info [%s]. Err: %v", n.Name, err),
			}
		}

		logrus.Debugf("checking PX status on node: %s", n.Name)
		switch pxNode.Status {
		case api.Status_STATUS_DECOMMISSION, api.Status_STATUS_OK: // do nothing
		case api.Status_STATUS_OFFLINE:
			// in case node is offline and it is a storageless node, the id might have changed so update it
			if len(pxNode.Pools) == 0 {
				d.updateNodeID(&n, d.getNodeManager())
			}
		default:
			return "", true, &ErrFailedToWaitForPx{
				Node: n,
				Cause: fmt.Sprintf("px cluster is usable but node %s status is not ok. Expected: %v Actual: %v",
					n.Name, api.Status_STATUS_OK, pxNode.Status),
			}
		}

		logrus.Infof("px on node: %s is now up. status: %v", n.Name, pxNode.Status)

		return "", false, nil
	}

	if _, err := task.DoRetryWithTimeout(t, timeout, defaultRetryInterval); err != nil {
		return err
	}

	// Check if PX pod is up
	logrus.Debugf("checking if PX pod is up on node: %s", n.Name)
	t = func() (interface{}, bool, error) {
		if !d.schedOps.IsPXReadyOnNode(n) {
			return "", true, &ErrFailedToWaitForPx{
				Node:  n,
				Cause: fmt.Sprintf("px pod is not ready on node: %s after %v", n.Name, timeout),
			}
		}
		return "", false, nil
	}

	if _, err := task.DoRetryWithTimeout(t, timeout, defaultRetryInterval); err != nil {
		return err
	}

	logrus.Debugf("px is fully operational on node: %s", n.Name)
	return nil
}

func (d *portworx) WaitDriverDownOnNode(n node.Node) error {
	t := func() (interface{}, bool, error) {
		// to avoid getting the same node which driver has brought down
		nManager, err := d.pickAlternateClusterManager(n)
		if err != nil {
			return "", true, &ErrFailedToWaitForPx{
				Node:  n,
				Cause: err.Error(),
			}
		}
		nodeInspectResponse, err := nManager.Inspect(d.getContext(), &api.SdkNodeInspectRequest{NodeId: n.VolDriverNodeID})
		if err != nil {
			return "", true, &ErrFailedToWaitForPx{
				Node:  n,
				Cause: err.Error(),
			}
		}
		if nodeInspectResponse.Node.Status != api.Status_STATUS_OFFLINE {
			return "", true, &ErrFailedToWaitForPx{
				Node: n,
				Cause: fmt.Sprintf("px is not yet down on node. Expected: %v Actual: %v",
					api.Status_STATUS_OFFLINE, nodeInspectResponse.Node.Status),
			}
		}

		logrus.Infof("px on node %s is now down.", n.Name)
		return "", false, nil
	}

	if _, err := task.DoRetryWithTimeout(t, validateNodeStopTimeout, defaultRetryInterval); err != nil {
		return err
	}

	return nil
}

func (d *portworx) ValidateStoragePools() error {
	listApRules, err := d.schedOps.ListAutopilotRules()
	if err != nil {
		return err
	}

	if len(listApRules.Items) != 0 {
		expectedPoolSizes, err := d.getExpectedPoolSizes(listApRules)
		if err != nil {
			return err
		}

		// start a task to check if the pools are at their expected sizes
		t := func() (interface{}, bool, error) {
			allDone := true
			if err := d.RefreshDriverEndpoints(); err != nil {
				return nil, true, err
			}

			for _, n := range node.GetWorkerNodes() {
				for _, pool := range n.StoragePools {
					expectedSize := expectedPoolSizes[pool.Uuid]
					if expectedSize != pool.TotalSize {
						if pool.TotalSize > expectedSize {
							// no need to retry with this state as pool is already at larger size than expected
							err := fmt.Errorf("node: %s pool: %s was expanded to size: %d larger than expected: %d",
								n.Name, pool.Uuid, pool.TotalSize, expectedSize)
							logrus.Errorf(err.Error())
							return "", false, err
						}

						logrus.Infof("node: %s, pool: %s, size is not as expected. Expected: %v, Actual: %v",
							n.Name, pool.Uuid, expectedSize, pool.TotalSize)
						allDone = false
					} else {
						logrus.Infof("node: %s, pool: %s, size is as expected. Expected: %v",
							n.Name, pool.Uuid, expectedSize)
					}
				}
			}
			if allDone {
				return "", false, nil
			}
			return "", true, fmt.Errorf("some sizes of pools are not as expected")
		}

		if _, err := task.DoRetryWithTimeout(t, validateStoragePoolSizeTimeout, validateStoragePoolSizeInterval); err != nil {
			return err
		}
	}
	return nil
}

func (d *portworx) getExpectedPoolSizes(listApRules *apapi.AutopilotRuleList) (map[string]uint64, error) {
	fn := "getExpectedPoolSizes"
	var (
		expectedPoolSizes = map[string]uint64{}
		err               error
	)
	d.RefreshDriverEndpoints()
	for _, apRule := range listApRules.Items {
		for _, n := range node.GetWorkerNodes() {
			for _, pool := range n.StoragePools {
				apRuleLabels := apRule.Spec.Selector.LabelSelector.MatchLabels
				labelsMatch := false
				for k, v := range apRuleLabels {
					if apRuleLabels[k] == pool.Labels[k] && apRuleLabels[v] == pool.Labels[v] {
						labelsMatch = true
					}
				}

				if labelsMatch {
					expectedPoolSizes[pool.Uuid], err = d.EstimatePoolExpandSize(apRule, pool, n)
					if err != nil {
						return nil, err
					}
				} else {
					expectedPoolSizes[pool.Uuid] = pool.StoragePoolAtInit.TotalSize
				}
			}
		}
	}
	logrus.Debugf("%s: expected sizes of storage pools: %+v", fn, expectedPoolSizes)
	return expectedPoolSizes, nil
}

// pickAlternateClusterManager returns a different node than given one, useful in case you want to skip nodes which are down
func (d *portworx) pickAlternateClusterManager(n node.Node) (api.OpenStorageNodeClient, error) {
	// Check if px is down on all node addresses. We don't want to keep track
	// which was the actual interface px was listening on before it went down
	for _, alternateNode := range node.GetWorkerNodes() {
		if alternateNode.Name == n.Name {
			continue
		}

		for _, addr := range alternateNode.Addresses {
			nodeManager, err := d.getNodeManagerByAddress(addr)
			if err != nil {
				return nil, err
			}
			ns, err := nodeManager.Enumerate(d.getContext(), &api.SdkNodeEnumerateRequest{})
			if err != nil {
				// if not responding in this addr, continue and pick another one, log the error
				logrus.Warnf("failed to check node %s on addr %s. Cause: %v", n.Name, addr, err)
				continue
			}
			if len(ns.NodeIds) != 0 {
				return nodeManager, nil
			}
		}
	}
	return nil, fmt.Errorf("failed to get an alternate cluster manager for %s", n.Name)
}

// CreateAutopilotRules creates autopilot rules
func (d *portworx) CreateAutopilotRules(apRules []apapi.AutopilotRule) error {
	for _, apRule := range apRules {
		autopilotRule, err := d.schedOps.CreateAutopilotRule(apRule)
		if err != nil {
			return err
		}
		logrus.Infof("Created Autopilot rule: %+v", autopilotRule)
	}
	return nil
}

func (d *portworx) IsStorageExpansionEnabled() (bool, error) {
	var listApRules *apapi.AutopilotRuleList
	var err error
	d.RefreshDriverEndpoints()
	if listApRules, err = d.schedOps.ListAutopilotRules(); err != nil {
		return false, err
	}

	if len(listApRules.Items) != 0 {
		for _, apRule := range listApRules.Items {
			for _, n := range node.GetWorkerNodes() {
				if isAutopilotMatchStoragePoolLabels(apRule, n.StoragePools) {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func isAutopilotMatchStoragePoolLabels(apRule apapi.AutopilotRule, sPools []node.StoragePool) bool {
	apRuleLabels := apRule.Spec.Selector.LabelSelector.MatchLabels
	for k, v := range apRuleLabels {
		for _, pool := range sPools {
			if poolLabelValue, ok := pool.Labels[k]; ok {
				if poolLabelValue == v {
					return true
				}
			}
		}
	}
	return false
}

func (d *portworx) WaitForUpgrade(n node.Node, tag string) error {
	t := func() (interface{}, bool, error) {

		// filter out first 3 octets from the tag
		matches := regexp.MustCompile(`^(\d+\.\d+\.\d+).*`).FindStringSubmatch(tag)
		if len(matches) != 2 {
			return nil, false, &ErrFailedToUpgradeVolumeDriver{
				Version: fmt.Sprintf("%s", tag),
				Cause:   fmt.Sprintf("failed to parse first 3 octets of version from new version tag: %s", tag),
			}
		}

		pxVersion, err := d.getPxVersionOnNode(n)
		if err != nil {
			return nil, true, &ErrFailedToWaitForPx{
				Node:  n,
				Cause: fmt.Sprintf("failed to get PX Version with error: %s", err),
			}
		}
		if !strings.HasPrefix(pxVersion, matches[1]) {
			return nil, true, &ErrFailedToUpgradeVolumeDriver{
				Version: fmt.Sprintf("%s", tag),
				Cause: fmt.Sprintf("version on node %s is still %s. It was expected to begin with: %s",
					n.VolDriverNodeID, pxVersion, matches[1]),
			}
		}

		logrus.Infof("version on node %s is %s. Expected version is %s", n.VolDriverNodeID, pxVersion, matches[1])

		return nil, false, nil
	}

	if _, err := task.DoRetryWithTimeout(t, upgradeTimeout, upgradeRetryInterval); err != nil {
		return err
	}
	return nil
}

func (d *portworx) GetReplicationFactor(vol *torpedovolume.Volume) (int64, error) {
	name := d.schedOps.GetVolumeName(vol)
	t := func() (interface{}, bool, error) {
		volumeInspectResponse, err := d.getVolDriver().Inspect(d.getContext(), &api.SdkVolumeInspectRequest{VolumeId: name})
		if err != nil && errIsNotFound(err) {
			return 0, false, err
		} else if err != nil {
			return 0, true, err
		}
		return volumeInspectResponse.Volume.Spec.HaLevel, false, nil
	}

	iReplFactor, err := task.DoRetryWithTimeout(t, validateReplicationUpdateTimeout, defaultRetryInterval)
	if err != nil {
		return 0, &ErrFailedToGetReplicationFactor{
			ID:    name,
			Cause: err.Error(),
		}
	}
	replFactor, ok := iReplFactor.(int64)
	if !ok {
		return 0, &ErrFailedToGetReplicationFactor{
			ID:    name,
			Cause: fmt.Sprintf("Replication factor is not of type int64"),
		}
	}

	return replFactor, nil
}

func (d *portworx) SetReplicationFactor(vol *torpedovolume.Volume, replFactor int64) error {
	volumeName := d.schedOps.GetVolumeName(vol)
	t := func() (interface{}, bool, error) {
		volDriver := d.getVolDriver()
		volumeInspectResponse, err := volDriver.Inspect(d.getContext(), &api.SdkVolumeInspectRequest{VolumeId: volumeName})
		if err != nil && errIsNotFound(err) {
			return nil, false, err
		} else if err != nil {
			return nil, true, err
		}

		volumeSpecUpdate := &api.VolumeSpecUpdate{
			HaLevelOpt:          &api.VolumeSpecUpdate_HaLevel{HaLevel: int64(replFactor)},
			SnapshotIntervalOpt: &api.VolumeSpecUpdate_SnapshotInterval{SnapshotInterval: math.MaxUint32},
			ReplicaSet:          &api.ReplicaSet{},
		}
		_, err = volDriver.Update(d.getContext(), &api.SdkVolumeUpdateRequest{
			VolumeId: volumeInspectResponse.Volume.Id,
			Spec:     volumeSpecUpdate,
		})
		if err != nil {
			return nil, false, err
		}
		quitFlag := false
		wdt := time.After(validateReplicationUpdateTimeout)
		for !quitFlag && !(areRepSetsFinal(volumeInspectResponse.Volume, replFactor) && isClean(volumeInspectResponse.Volume)) {
			select {
			case <-wdt:
				quitFlag = true
			default:
				volumeInspectResponse, err = volDriver.Inspect(d.getContext(), &api.SdkVolumeInspectRequest{VolumeId: volumeName})
				if err != nil && errIsNotFound(err) {
					return nil, false, err
				} else if err != nil {
					return nil, true, err
				}
				time.Sleep(defaultRetryInterval)
			}
		}
		if !(areRepSetsFinal(volumeInspectResponse.Volume, replFactor) && isClean(volumeInspectResponse.Volume)) {
			return 0, false, fmt.Errorf("Volume didn't successfully change to replication factor of %d", replFactor)
		}
		return 0, false, nil
	}

	if _, err := task.DoRetryWithTimeout(t, validateReplicationUpdateTimeout, defaultRetryInterval); err != nil {
		return &ErrFailedToSetReplicationFactor{
			ID:    volumeName,
			Cause: err.Error(),
		}
	}

	return nil
}

func (d *portworx) GetMaxReplicationFactor() int64 {
	return 3
}

func (d *portworx) GetMinReplicationFactor() int64 {
	return 1
}

func (d *portworx) GetAggregationLevel(vol *torpedovolume.Volume) (int64, error) {
	volumeName := d.schedOps.GetVolumeName(vol)
	t := func() (interface{}, bool, error) {
		volResp, err := d.getVolDriver().Inspect(d.getContext(), &api.SdkVolumeInspectRequest{VolumeId: volumeName})
		if err != nil && errIsNotFound(err) {
			return 0, false, err
		} else if err != nil {
			return 0, true, err
		}
		return volResp.Volume.Spec.AggregationLevel, false, nil
	}

	iAggrLevel, err := task.DoRetryWithTimeout(t, inspectVolumeTimeout, inspectVolumeRetryInterval)
	if err != nil {
		return 0, &ErrFailedToGetAggregationLevel{
			ID:    volumeName,
			Cause: err.Error(),
		}
	}
	aggrLevel, ok := iAggrLevel.(uint32)
	if !ok {
		return 0, &ErrFailedToGetAggregationLevel{
			ID:    volumeName,
			Cause: fmt.Sprintf("Aggregation level is not of type uint32"),
		}
	}

	return int64(aggrLevel), nil
}

func isClean(vol *api.Volume) bool {
	for _, v := range vol.RuntimeState {
		if v.GetRuntimeState()["RuntimeState"] != "clean" {
			return false
		}
	}
	return true
}

func areRepSetsFinal(vol *api.Volume, replFactor int64) bool {
	for _, rs := range vol.ReplicaSets {
		if int64(len(rs.GetNodes())) != replFactor {
			return false
		}
	}
	return true
}

func (d *portworx) setDriver() error {
	var err error
	var endpoint string

	// Try portworx-service first
	endpoint, err = d.schedOps.GetServiceEndpoint()
	if err == nil && endpoint != "" {
		if err = d.testAndSetEndpointUsingService(endpoint); err == nil {
			d.refreshEndpoint = false
			return nil
		}
		logrus.Infof("testAndSetEndpoint failed for %v: %v", endpoint, err)
	} else if err != nil && len(node.GetWorkerNodes()) == 0 {
		return err
	}

	// Try direct address of cluster nodes
	// Set refresh endpoint to true so that we try and get the new
	// and working driver if the endpoint we are hooked onto goes
	// down
	d.refreshEndpoint = true
	logrus.Infof("Getting new driver.")
	for _, n := range node.GetWorkerNodes() {
		for _, addr := range n.Addresses {
			if err = d.testAndSetEndpointUsingNodeIP(addr); err == nil {
				return nil
			}
			logrus.Infof("testAndSetEndpoint failed for %v: %v", endpoint, err)
		}
	}

	return fmt.Errorf("failed to get endpoint for portworx volume driver")
}

func (d *portworx) testAndSetEndpointUsingService(endpoint string) error {
	sdkPort, err := getSDKPort()
	if err != nil {
		return err
	}

	restPort, err := getRestPort()
	if err != nil {
		return err
	}

	return d.testAndSetEndpoint(endpoint, sdkPort, restPort)
}

func (d *portworx) testAndSetEndpointUsingNodeIP(ip string) error {
	sdkPort, err := getSDKContainerPort()
	if err != nil {
		return err
	}

	restPort, err := getRestContainerPort()
	if err != nil {
		return err
	}

	return d.testAndSetEndpoint(ip, sdkPort, restPort)
}

func (d *portworx) testAndSetEndpoint(endpoint string, sdkport, apiport int32) error {
	pxEndpoint := fmt.Sprintf("%s:%d", endpoint, sdkport)
	conn, err := grpc.Dial(pxEndpoint, grpc.WithInsecure())
	if err != nil {
		return err
	}

	d.clusterManager = api.NewOpenStorageClusterClient(conn)
	_, err = d.clusterManager.InspectCurrent(d.getContext(), &api.SdkClusterInspectCurrentRequest{})
	if st, ok := status.FromError(err); ok && st.Code() == codes.Unavailable {
		return err
	}

	d.volDriver = api.NewOpenStorageVolumeClient(conn)
	d.nodeManager = api.NewOpenStorageNodeClient(conn)
	d.mountAttachManager = api.NewOpenStorageMountAttachClient(conn)
	d.clusterPairManager = api.NewOpenStorageClusterPairClient(conn)
	d.alertsManager = api.NewOpenStorageAlertsClient(conn)
	if legacyClusterManager, err := d.getLegacyClusterManager(endpoint, apiport); err == nil {
		d.legacyClusterManager = legacyClusterManager
	} else {
		return err
	}
	logrus.Infof("Using %v as endpoint for portworx volume driver", pxEndpoint)

	return nil
}

func (d *portworx) getLegacyClusterManager(endpoint string, pxdRestPort int32) (cluster.Cluster, error) {
	pxEndpoint := fmt.Sprintf("http://%s:%d", endpoint, pxdRestPort)
	var cClient *client.Client
	var err error
	if d.token != "" {
		cClient, err = clusterclient.NewAuthClusterClient(pxEndpoint, "v1", d.token, "")
		if err != nil {
			return nil, err
		}
	} else {
		cClient, err = clusterclient.NewClusterClient(pxEndpoint, "v1")
		if err != nil {
			return nil, err
		}
	}

	clusterManager := clusterclient.ClusterManager(cClient)
	_, err = clusterManager.Enumerate()
	if err != nil {
		return nil, err
	}
	return clusterManager, nil
}

func (d *portworx) getContextWithToken(ctx context.Context, token string) context.Context {
	md, _ := metadata.FromOutgoingContext(ctx)
	md = metadata.Join(md, metadata.New(map[string]string{
		"authorization": "bearer " + token,
	}))
	return metadata.NewOutgoingContext(ctx, md)
}

func (d *portworx) getContext() context.Context {
	ctx := context.Background()
	if len(d.token) > 0 {
		return d.getContextWithToken(ctx, d.token)
	}
	return ctx
}

func (d *portworx) StartDriver(n node.Node) error {
	logrus.Infof("Starting volume driver on %s.", n.Name)
	err := d.schedOps.StartPxOnNode(n)
	if err != nil {
		return err
	}
	return d.nodeDriver.Systemctl(n, pxSystemdServiceName, node.SystemctlOpts{
		Action: "start",
		ConnectionOpts: node.ConnectionOpts{
			Timeout:         startDriverTimeout,
			TimeBeforeRetry: defaultRetryInterval,
		}})
}

func (d *portworx) UpgradeDriver(endpointURL string, endpointVersion string) error {
	upgradeFileName := "/upgrade.sh"

	if endpointURL == "" {
		return fmt.Errorf("no link supplied for upgrading portworx")
	}
	logrus.Infof("upgrading portworx from %s URL and %s endpoint version", endpointURL, endpointVersion)

	// Getting upgrade script
	fullEndpointURL := fmt.Sprintf("%s/%s/upgrade", endpointURL, endpointVersion)
	cmd := exec.Command("wget", "-O", upgradeFileName, fullEndpointURL)
	output, err := cmd.Output()
	logrus.Infof("%s", output)
	if err != nil {
		return fmt.Errorf("error on downloading endpoint: %+v", err)
	}
	// Check if downloaded file exists
	file, err := os.Stat(upgradeFileName)
	if err != nil {
		return fmt.Errorf("file %s doesn't exist", upgradeFileName)
	}
	logrus.Infof("file %s exists", upgradeFileName)

	// Check if downloaded file is not empty
	fileSize := file.Size()
	if fileSize == 0 {
		return fmt.Errorf("file %s is empty", upgradeFileName)
	}
	logrus.Infof("file %s is not empty", upgradeFileName)

	// Change permission on file to be able to execute
	cmd = exec.Command("chmod", "+x", upgradeFileName)
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("error on changing permission for %s file", upgradeFileName)
	}
	logrus.Infof("permission changed on file %s", upgradeFileName)

	nodeList := node.GetStorageDriverNodes()
	pxNode := nodeList[0]
	pxVersion, err := d.getPxVersionOnNode(pxNode)
	if err != nil {
		return fmt.Errorf("error on getting PX Version on node %s with err: %v", pxNode.Name, err)
	}
	// If PX Version less than 2.x.x.x, then we have to add timeout parameter to avoid test failure
	// more details in https://portworx.atlassian.net/browse/PWX-10108
	cmdArgs := []string{upgradeFileName, "-f"}
	majorPxVersion := pxVersion[:1]
	if majorPxVersion < "2" {
		cmdArgs = append(cmdArgs, "-u", strconv.Itoa(int(upgradePerNodeTimeout/time.Second)))
	}

	// Run upgrade script
	logrus.Infof("executing /bin/sh with params: %s\n", cmdArgs)
	cmd = exec.Command("/bin/sh", cmdArgs...)
	output, err = cmd.Output()

	// Print and replace all '\n' with new lines
	logrus.Infof("%s", strings.Replace(string(output[:]), `\n`, "\n", -1))
	if err != nil {
		return fmt.Errorf("error: %+v", err)
	}
	logrus.Infof("Portworx cluster upgraded successfully")

	for _, n := range node.GetStorageDriverNodes() {
		if err := d.WaitForUpgrade(n, endpointVersion); err != nil {
			return err
		}
	}

	return nil
}

// GetClusterPairingInfo returns cluster pair information
func (d *portworx) GetClusterPairingInfo() (map[string]string, error) {
	pairInfo := make(map[string]string)

	resp, err := d.clusterPairManager.GetToken(d.getContext(), &api.SdkClusterPairGetTokenRequest{})
	if err != nil {
		return nil, err
	}
	logrus.Infof("Response for token: %v", resp.Result.Token)

	// file up cluster pair info
	clusterIPAddress, err := d.schedOps.GetServiceEndpoint()
	if err != nil {
		return pairInfo, err
	}
	pairInfo[clusterIP] = clusterIPAddress
	pairInfo[tokenKey] = resp.Result.Token
	pwxServicePort, err := getSDKPort()
	if err != nil {
		return nil, err
	}
	pairInfo[clusterPort] = fmt.Sprintf("%d", pwxServicePort)

	return pairInfo, nil
}

func (d *portworx) DecommissionNode(n *node.Node) error {

	if err := k8sCore.AddLabelOnNode(n.Name, schedops.PXEnabledLabelKey, "remove"); err != nil {
		return &ErrFailedToDecommissionNode{
			Node:  n.Name,
			Cause: fmt.Sprintf("Failed to set label on node: %v. Err: %v", n.Name, err),
		}
	}

	if err := d.StopDriver([]node.Node{*n}, false); err != nil {
		return &ErrFailedToDecommissionNode{
			Node:  n.Name,
			Cause: fmt.Sprintf("Failed to stop driver on node: %v. Err: %v", n.Name, err),
		}
	}

	nodeResp, err := d.getNodeManager().Inspect(d.getContext(), &api.SdkNodeInspectRequest{NodeId: n.VolDriverNodeID})
	if err != nil {
		return &ErrFailedToDecommissionNode{
			Node:  n.Name,
			Cause: fmt.Sprintf("Failed to inspect node: %v. Err: %v", nodeResp.Node, err),
		}
	}

	// TODO replace when sdk supports node removal
	if err = d.legacyClusterManager.Remove([]api.Node{{Id: nodeResp.Node.Id}}, false); err != nil {
		return &ErrFailedToDecommissionNode{
			Node:  n.Name,
			Cause: err.Error(),
		}
	}

	// update node in registry
	n.IsStorageDriverInstalled = false
	if err = node.UpdateNode(*n); err != nil {
		return fmt.Errorf("failed to update node %s. Cause: %v", n.Name, err)
	}

	// force refresh endpoint
	d.refreshEndpoint = true

	return nil
}

func (d *portworx) RejoinNode(n *node.Node) error {

	opts := node.ConnectionOpts{
		IgnoreError:     false,
		TimeBeforeRetry: defaultRetryInterval,
		Timeout:         defaultTimeout,
	}
	if _, err := d.nodeDriver.RunCommand(*n, "/opt/pwx/bin/pxctl sv node-wipe --all", opts); err != nil {
		return &ErrFailedToRejoinNode{
			Node:  n.Name,
			Cause: err.Error(),
		}
	}
	if err := k8sCore.RemoveLabelOnNode(n.Name, schedops.PXServiceLabelKey); err != nil {
		return &ErrFailedToRejoinNode{
			Node:  n.Name,
			Cause: fmt.Sprintf("Failed to set label on node: %v. Err: %v", n.Name, err),
		}
	}
	if err := k8sCore.RemoveLabelOnNode(n.Name, schedops.PXEnabledLabelKey); err != nil {
		return &ErrFailedToRejoinNode{
			Node:  n.Name,
			Cause: fmt.Sprintf("Failed to set label on node: %v. Err: %v", n.Name, err),
		}
	}
	if err := k8sCore.UnCordonNode(n.Name, defaultTimeout, defaultRetryInterval); err != nil {
		return &ErrFailedToRejoinNode{
			Node:  n.Name,
			Cause: fmt.Sprintf("Failed to uncordon node: %v. Err: %v", n.Name, err),
		}
	}
	return nil
}

func (d *portworx) GetNodeStatus(n node.Node) (*api.Status, error) {
	nodeResponse, err := d.getNodeManager().Inspect(d.getContext(), &api.SdkNodeInspectRequest{NodeId: n.VolDriverNodeID})
	if err != nil {
		if isNodeNotFound(err) {
			apiSt := api.Status_STATUS_NONE
			return &apiSt, nil
		}
		return nil, &ErrFailedToGetNodeStatus{
			Node:  n.Name,
			Cause: fmt.Sprintf("Failed to check node status: %v. Err: %v", n.Name, err),
		}
	}
	return &nodeResponse.Node.Status, nil
}

func (d *portworx) getVolDriver() api.OpenStorageVolumeClient {
	if d.refreshEndpoint {
		d.setDriver()
	}
	return d.volDriver
}

func (d *portworx) getClusterManager() api.OpenStorageClusterClient {
	if d.refreshEndpoint {
		d.setDriver()
	}
	return d.clusterManager

}

func (d *portworx) getNodeManager() api.OpenStorageNodeClient {
	if d.refreshEndpoint {
		d.setDriver()
	}
	return d.nodeManager

}

func (d *portworx) getMountAttachManager() api.OpenStorageMountAttachClient {
	if d.refreshEndpoint {
		d.setDriver()
	}
	return d.mountAttachManager

}

func (d *portworx) getClusterPairManager() api.OpenStorageClusterPairClient {
	if d.refreshEndpoint {
		d.setDriver()
	}
	return d.clusterPairManager

}

func (d *portworx) getAlertsManager() api.OpenStorageAlertsClient {
	if d.refreshEndpoint {
		d.setDriver()
	}
	return d.alertsManager

}

func (d *portworx) getNodeManagerByAddress(addr string) (api.OpenStorageNodeClient, error) {
	pxPort, err := getSDKContainerPort()
	if err != nil {
		return nil, err
	}
	pxEndpoint := fmt.Sprintf("%s:%d", addr, pxPort)
	conn, err := grpc.Dial(pxEndpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	dClient := api.NewOpenStorageNodeClient(conn)
	_, err = dClient.Enumerate(d.getContext(), &api.SdkNodeEnumerateRequest{})
	if err != nil {
		return nil, err
	}

	return dClient, nil
}

func (d *portworx) maintenanceOp(n node.Node, op string) error {
	// TODO replace by sdk call whenever it is available
	pxdRestPort, err := getRestPort()
	if err != nil {
		return err
	}
	endpoint, err := d.schedOps.GetServiceEndpoint()
	var url string
	if err != nil {
		logrus.Warnf("unable to get service endpoint falling back to node addr %v", err)
		pxdRestPort, err = getRestContainerPort()
		if err != nil {
			return err
		}
		url = fmt.Sprintf("http://%s:%d", n.Addresses[0], pxdRestPort)
	} else {
		url = fmt.Sprintf("http://%s:%d", endpoint, pxdRestPort)
	}
	c, err := client.NewClient(url, "", "")
	if err != nil {
		return err
	}
	req := c.Get().Resource(op)
	resp := req.Do()
	return resp.Error()
}

func (d *portworx) GetReplicaSets(torpedovol *torpedovolume.Volume) ([]*api.ReplicaSet, error) {
	volumeName := d.schedOps.GetVolumeName(torpedovol)
	volumeInspectResponse, err := d.getVolDriver().Inspect(d.getContext(), &api.SdkVolumeInspectRequest{VolumeId: volumeName})
	if err != nil {
		return nil, &ErrFailedToInspectVolume{
			ID:    torpedovol.Name,
			Cause: err.Error(),
		}
	}

	return volumeInspectResponse.Volume.ReplicaSets, nil
}

func (d *portworx) updateNodeID(n *node.Node, nManager ...api.OpenStorageNodeClient) (*node.Node, error) {
	nodes, err := d.getPxNodes(nManager...)
	if err != nil {
		return n, err
	}
	if err = d.updateNode(n, nodes); err != nil {
		return &node.Node{}, fmt.Errorf("failed to update node ID for node %s. Cause: %v", n.Name, err)
	}
	return n, fmt.Errorf("node %v not found in cluster", n)
}

func getGroupMatches(groupRegex *regexp.Regexp, str string) map[string]string {
	match := groupRegex.FindStringSubmatch(str)
	result := make(map[string]string)
	if len(match) > 0 {
		for i, name := range groupRegex.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}
	}
	return result
}

// ValidateVolumeSnapshotRestore return nil if snapshot is restored successuflly to
// given volumes
// TODO: additionally check for restore objects in case of cloudsnap
func (d *portworx) ValidateVolumeSnapshotRestore(vol string, snapshotData *snapv1.VolumeSnapshotData, timeStart time.Time) error {
	snap := snapshotData.Spec.PortworxSnapshot.SnapshotID
	if snapshotData.Spec.PortworxSnapshot.SnapshotType == snapv1.PortworxSnapshotTypeCloud {
		snap = "in-place-restore-" + vol
	}

	tsStart := timestamp.Timestamp{
		Nanos:   int32(timeStart.UnixNano()),
		Seconds: timeStart.Unix(),
	}
	currentTime := time.Now()
	tsEnd := timestamp.Timestamp{
		Nanos:   int32(currentTime.UnixNano()),
		Seconds: currentTime.Unix(),
	}
	alerts, err := d.alertsManager.EnumerateWithFilters(d.getContext(), &api.SdkAlertsEnumerateWithFiltersRequest{
		Queries: []*api.SdkAlertsQuery{
			{
				Query: &api.SdkAlertsQuery_ResourceTypeQuery{
					ResourceTypeQuery: &api.SdkAlertsResourceTypeQuery{
						ResourceType: api.ResourceType_RESOURCE_TYPE_VOLUME,
					},
				},
				Opts: []*api.SdkAlertsOption{
					{Opt: &api.SdkAlertsOption_TimeSpan{
						TimeSpan: &api.SdkAlertsTimeSpan{
							StartTime: &tsStart,
							EndTime:   &tsEnd,
						},
					}},
				},
			},
		},
	})

	if err != nil {
		return err
	}
	// get volume and snap info
	volDriver := d.getVolDriver()
	pvcVol, err := volDriver.Inspect(d.getContext(), &api.SdkVolumeInspectRequest{VolumeId: vol})
	if err != nil {
		return fmt.Errorf("inspect failed for %v: %v", vol, err)
	}
	// form alert msg for snapshot restore
	grepMsg := "Volume " + pvcVol.Volume.GetLocator().GetName() +
		" (" + pvcVol.Volume.GetId() + ") restored from snapshot "
	snapVol, err := volDriver.Inspect(d.getContext(), &api.SdkVolumeInspectRequest{VolumeId: snap})
	if err != nil {
		// Restore object get deleted in case of cloudsnap
		logrus.Warnf("Snapshot volume %v not found: %v", snap, err)
		grepMsg = grepMsg + snap
	} else {
		grepMsg = grepMsg + snapVol.Volume.GetLocator().GetName() +
			" (" + snap + ")"
	}

	isSuccess := false
	alertsResp, err := alerts.Recv()
	if err != nil {
		return err
	}
	for _, alert := range alertsResp.Alerts {
		if strings.Contains(alert.GetMessage(), grepMsg) {
			isSuccess = true
			break
		}
	}
	if isSuccess {
		return nil
	}
	return fmt.Errorf("restore failed, expected alert to be present : %v", grepMsg)
}

func (d *portworx) getTokenForVolume(name string, params map[string]string) string {
	token := d.token
	var volSecret string
	var volSecretNamespace string
	if secret, ok := params[secretName]; ok {
		volSecret = secret
	}
	if namespace, ok := params[secretNamespace]; ok {
		volSecretNamespace = namespace
	}
	if volSecret != "" && volSecretNamespace != "" {
		if tk, ok := params["auth-token"]; ok {
			token = tk
		}
	}
	return token
}

func deleteLabelsFromRequestedSpec(expectedLocator *api.VolumeLocator) {
	for labelKey := range expectedLocator.VolumeLabels {
		if hasIgnorePrefix(labelKey) {
			delete(expectedLocator.VolumeLabels, labelKey)
		}
	}
}

func hasIgnorePrefix(str string) bool {
	for _, label := range deleteVolumeLabelList {
		if strings.HasPrefix(str, label) {
			return true
		}
	}
	return false
}

func (d *portworx) getKvdbMembers(n node.Node) (map[string]metadataNode, error) {
	kvdbMembers := make(map[string]metadataNode)
	pxdRestPort, err := getRestPort()
	if err != nil {
		return kvdbMembers, err
	}
	endpoint, err := d.schedOps.GetServiceEndpoint()
	var url string
	if err != nil {
		logrus.Warnf("unable to get service endpoint falling back to node addr %v", err)
		pxdRestPort, err = getRestContainerPort()
		if err != nil {
			return kvdbMembers, err
		}
		url = fmt.Sprintf("http://%s:%d", n.Addresses[0], pxdRestPort)
	} else {
		url = fmt.Sprintf("http://%s:%d", endpoint, pxdRestPort)
	}
	// TODO replace by sdk call whenever it is available
	logrus.Infof("Url to call %v", url)
	c, err := client.NewClient(url, "", "")
	if err != nil {
		return nil, err
	}
	req := c.Get().Resource("kvmembers")
	resp := req.Do()
	if resp.Error() != nil {
		if strings.Contains(resp.Error().Error(), "command not supported") {
			return kvdbMembers, nil
		}
		return kvdbMembers, resp.Error()
	}
	err = resp.Unmarshal(&kvdbMembers)
	return kvdbMembers, err
}

func (d *portworx) CollectDiags(n node.Node) error {
	var err error

	pxNode, err := d.getPxNode(&n)
	if err != nil {
		return err
	}

	opts := node.ConnectionOpts{
		IgnoreError:     false,
		TimeBeforeRetry: defaultRetryInterval,
		Timeout:         defaultTimeout,
		Sudo:            true,
	}

	logrus.Debugf("Collecting diags on node %v, because there was an error", pxNode.Hostname)

	if pxNode.Status == api.Status_STATUS_OFFLINE {
		logrus.Debugf("Node %v is offline, collecting diags using pxctl", pxNode.Hostname)

		// Only way to collect diags when PX is offline is using pxctl
		out, err := d.nodeDriver.RunCommand(n, "pxctl sv diags -a -f", opts)
		if err != nil {
			return fmt.Errorf("failed to collect diags on node %v, Err: %v %v", pxNode.Hostname, err, out)
		}
		logrus.Debugf("Successfully collected diags on node %v", pxNode.Hostname)
		return nil
	}

	url := fmt.Sprintf("http://%s:9014", n.Addresses[0])

	r := &DiagRequestConfig{
		DockerHost:    "unix:///var/run/docker.sock",
		OutputFile:    "/var/cores/diags.tar.gz",
		ContainerName: "",
		Profile:       false,
		Live:          true,
		Upload:        false,
		All:           true,
		Force:         true,
		OnHost:        true,
		Extra:         false,
	}

	c, err := client.NewClient(url, "", "")
	if err != nil {
		return err
	}
	req := c.Post().Resource(pxDiagPath).Body(r)
	resp := req.Do()
	if resp.Error() != nil {
		return fmt.Errorf("failed to collect diags on node %v, Err: %v", pxNode.Hostname, resp.Error())
	}
	logrus.Debugf("Successfully collected diags on node %v", pxNode.Hostname)
	return nil
}

// EstimatePoolExpandSize calculates the expected size based on autopilot rule, initial and workload sizes
func (d *portworx) EstimatePoolExpandSize(apRule apapi.AutopilotRule, pool node.StoragePool, node node.Node) (uint64, error) {
	// this method calculates expected pool size for given initial and workload sizes.
	// for ex: autopilot rule says scale storage pool by 50% with scale type adding disks when
	// available storage pool capacity is less that 70%. Initial storage pool size is 32Gb and
	// workload size on this pool is 10Gb
	// First, we get PX metric from the rule and calculate it's own value based on initial storage
	// pool size. In our example metric value will be (32Gb-10Gb*100) / 32Gb = 68.75
	// Second, we check if above metric matches condition in the rule conditions. Metric value is
	// less than 70% and we have to apply condition action, which will add another disk with 32Gb.
	// It will continue until metric value won't match condition in the rule

	// first check if the apRule is supported by torpedo
	var actionScaleType string
	for _, ruleAction := range apRule.Spec.Actions {
		if ruleAction.Name != aututils.StorageSpecAction {
			return 0, &tp_errors.ErrNotSupported{
				Type:      ruleAction.Name,
				Operation: "EstimatePoolExpandSize for action",
			}
		}

		if len(ruleAction.Params) == 0 {
			return 0, &tp_errors.ErrNotSupported{
				Type:      "without params",
				Operation: "Pool expand action",
			}
		}

		actionScaleType = ruleAction.Params[aututils.RuleScaleType]
		if len(actionScaleType) == 0 {
			return 0, &tp_errors.ErrNotSupported{
				Type:      "without param for scale type",
				Operation: "Pool expand action",
			}
		}
	}

	var (
		initialSize         = pool.StoragePoolAtInit.TotalSize
		workloadSize        = pool.WorkloadSize
		calculatedTotalSize = initialSize
		baseDiskSize        uint64
	)

	// adjust workloadSize by the initial usage that PX pools start with
	// TODO get this from porx: func (bm *btrfsMount) MkReserve(volname string, available uint64) error {
	poolBaseUsage := uint64(float64(initialSize) / 10)
	if initialSize < (32 * units.GiB) {
		poolBaseUsage = 3 * units.GiB
	}
	workloadSize += poolBaseUsage

	// get base disk size for the pool from the node spec
	for _, disk := range node.Disks {
		// NOTE: below medium check if a weak assumption and will fail if the installation has multiple pools on the node
		// with the same medium (pools with disks of different sizes but same medium). The SDK does not provide a direct
		// mapping of disks to pools so this the best we can do from SDK right now.
		if disk.Medium == pool.StoragePoolAtInit.Medium {
			baseDiskSize = disk.Size
		}
	}

	if baseDiskSize == 0 {
		return 0, fmt.Errorf("failed to detect base disk size for pool: %s", pool.Uuid)
	}

	//	The goal of the below for loop is to keep increasing calculatedTotalSize until the rule conditions match
	for {
		for _, conditionExpression := range apRule.Spec.Conditions.Expressions {
			var metricValue float64
			switch conditionExpression.Key {
			case aututils.PxPoolAvailableCapacityMetric:
				availableSize := int64(calculatedTotalSize) - int64(workloadSize)
				metricValue = float64(availableSize*100) / float64(calculatedTotalSize)
			case aututils.PxPoolTotalCapacityMetric:
				metricValue = float64(calculatedTotalSize) / units.GiB
			default:
				return 0, &tp_errors.ErrNotSupported{
					Type:      conditionExpression.Key,
					Operation: "Pool Condition Expression Key",
				}
			}

			if doesConditionMatch(metricValue, conditionExpression) {
				for _, ruleAction := range apRule.Spec.Actions {
					actionScalePercentage, err := strconv.ParseUint(ruleAction.Params[aututils.RuleActionsScalePercentage], 10, 64)
					if err != nil {
						return 0, err
					}

					requiredScaleSize := float64(calculatedTotalSize * actionScalePercentage / 100)
					if actionScaleType == aututils.RuleScaleTypeAddDisk {
						requiredNewDisks := uint64(math.Ceil(requiredScaleSize / float64(baseDiskSize)))
						calculatedTotalSize += requiredNewDisks * baseDiskSize
					} else {
						calculatedTotalSize += uint64(requiredScaleSize)
					}
				}
			} else {
				return calculatedTotalSize, nil
			}
		}
	}
}

// EstimatePoolExpandSize calculates the expected size of a volume based on autopilot rule, initial and workload sizes
func (d *portworx) EstimateVolumeExpandSize(apRule apapi.AutopilotRule, initialSize, workloadSize uint64) (uint64, error) {
	// this method calculates expected autopilot object size for given initial and workload sizes.
	for _, ruleAction := range apRule.Spec.Actions {
		if ruleAction.Name != aututils.VolumeSpecAction {
			return 0, &tp_errors.ErrNotSupported{
				Type:      ruleAction.Name,
				Operation: "EstimateVolumeExpandSize for action",
			}
		}
	}

	calculatedTotalSize := initialSize
	//	The goal of the below for loop is to keep increasing calculatedTotalSize until the rule conditions match
	for {
		for _, conditionExpression := range apRule.Spec.Conditions.Expressions {
			var metricValue float64
			switch conditionExpression.Key {
			case aututils.PxVolumeUsagePercentMetric:
				metricValue = float64(int64(workloadSize)) * 100 / float64(calculatedTotalSize)
			case aututils.PxVolumeTotalCapacityMetric:
				metricValue = float64(calculatedTotalSize) / units.GB
			default:
				return 0, &tp_errors.ErrNotSupported{
					Type:      conditionExpression.Key,
					Operation: "Volume Condition Expression Key",
				}
			}

			if doesConditionMatch(metricValue, conditionExpression) {
				for _, ruleAction := range apRule.Spec.Actions {
					actionScalePercentage, err := strconv.ParseUint(ruleAction.Params[aututils.RuleActionsScalePercentage], 10, 64)
					if err != nil {
						return 0, err
					}

					requiredScaleSize := float64(calculatedTotalSize * actionScalePercentage / 100)
					calculatedTotalSize += uint64(requiredScaleSize)

					// check if calculated size is more than maxsize
					if actionMaxSize, ok := ruleAction.Params[aututils.RuleMaxSize]; ok {
						maxSize, _ := strconv.ParseUint(actionMaxSize, 10, 64)
						if maxSize != 0 && calculatedTotalSize > maxSize {
							return maxSize, nil
						}
					}

				}
			} else {
				return calculatedTotalSize, nil
			}
		}
	}
}

func doesConditionMatch(metricValue float64, conditionExpression *apapi.LabelSelectorRequirement) bool {
	condExprValue, _ := strconv.ParseFloat(conditionExpression.Values[0], 64)
	return metricValue < condExprValue && conditionExpression.Operator == apapi.LabelSelectorOpLt ||
		metricValue > condExprValue && conditionExpression.Operator == apapi.LabelSelectorOpGt

}

// getRestPort gets the service port for rest api, required when using service endpoint
func getRestPort() (int32, error) {
	svc, err := k8sCore.GetService(schedops.PXServiceName, schedops.PXNamespace)
	if err != nil {
		return 0, err
	}
	for _, port := range svc.Spec.Ports {
		if port.Name == "px-api" {
			return port.Port, nil
		}
	}
	return 0, fmt.Errorf("px-api port not found in service")
}

// getRestContainerPort gets the rest api container port exposed in the node, required when using node ip
func getRestContainerPort() (int32, error) {
	svc, err := k8sCore.GetService(schedops.PXServiceName, schedops.PXNamespace)
	if err != nil {
		return 0, err
	}
	for _, port := range svc.Spec.Ports {
		if port.Name == "px-api" {
			return port.TargetPort.IntVal, nil
		}
	}
	return 0, fmt.Errorf("px-api target port not found in service")
}

// getSDKPort gets sdk service port, required when using service endpoint
func getSDKPort() (int32, error) {
	svc, err := k8sCore.GetService(schedops.PXServiceName, schedops.PXNamespace)
	if err != nil {
		return 0, err
	}
	for _, port := range svc.Spec.Ports {
		if port.Name == "px-sdk" {
			return port.Port, nil
		}
	}
	return 0, fmt.Errorf("px-sdk port not found in service")
}

// getSDKContainerPort gets the sdk container port in the node, required when using node ip
func getSDKContainerPort() (int32, error) {
	svc, err := k8sCore.GetService(schedops.PXServiceName, schedops.PXNamespace)
	if err != nil {
		return 0, err
	}
	for _, port := range svc.Spec.Ports {
		if port.Name == "px-sdk" {
			return port.TargetPort.IntVal, nil
		}
	}
	return 0, fmt.Errorf("px-sdk target port not found in service")
}

func init() {
	torpedovolume.Register(DriverName, &portworx{})
}
