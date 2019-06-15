package portworx

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/client"
	"github.com/libopenstorage/openstorage/api/spec"
	"github.com/pborman/uuid"
	"github.com/portworx/sched-ops/k8s"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/node"
	torpedovolume "github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/drivers/volume/portworx/schedops"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// DriverName is the name of the portworx driver implementation
	DriverName              = "pxd"
	pxdClientSchedUserAgent = "pxd-sched"
	pxdRestPort             = 9001
	pxVersionLabel          = "PX Version"
	maintenanceOpRetries    = 3
	enterMaintenancePath    = "/entermaintenance"
	exitMaintenancePath     = "/exitmaintenance"
	pxSystemdServiceName    = "portworx.service"
	storageStatusUp         = "Up"
	tokenKey                = "token"
	clusterIP               = "ip"
	clusterPort             = "port"
	remoteKubeConfigPath    = "/tmp/kubeconfig"
)

const (
	defaultTimeout                   = 2 * time.Minute
	defaultRetryInterval             = 10 * time.Second
	defaultPxServicePort             = 9020
	maintenanceOpTimeout             = 1 * time.Minute
	maintenanceWaitTimeout           = 2 * time.Minute
	inspectVolumeTimeout             = 10 * time.Second
	inspectVolumeRetryInterval       = 2 * time.Second
	validateDeleteVolumeTimeout      = 3 * time.Minute
	validateReplicationUpdateTimeout = 10 * time.Minute
	validateClusterStartTimeout      = 2 * time.Minute
	validateNodeStartTimeout         = 3 * time.Minute
	validatePXStartTimeout           = 5 * time.Minute
	validateVolumeAttachedTimeout    = 30 * time.Second
	validateVolumeAttachedInterval   = 5 * time.Second
	validateNodeStopTimeout          = 5 * time.Minute
	stopDriverTimeout                = 5 * time.Minute
	crashDriverTimeout               = 2 * time.Minute
	startDriverTimeout               = 2 * time.Minute
	upgradeTimeout                   = 10 * time.Minute
	upgradeRetryInterval             = 30 * time.Second
	waitVolDriverToCrash             = 1 * time.Minute
)

type portworx struct {
	clusterManager     api.OpenStorageClusterClient
	nodeManager        api.OpenStorageNodeClient
	mountAttachManager api.OpenStorageMountAttachClient
	volDriver          api.OpenStorageVolumeClient
	clusterPairManager api.OpenStorageClusterPairClient
	schedOps           schedops.Driver
	nodeDriver         node.Driver
	refreshEndpoint    bool
}

func (d *portworx) String() string {
	return DriverName
}

func (d *portworx) Init(sched string, nodeDriver string) error {
	logrus.Infof("Using the Portworx volume driver under scheduler: %v", sched)
	var err error
	if d.nodeDriver, err = node.Get(nodeDriver); err != nil {
		return err
	}

	if d.schedOps, err = schedops.Get(sched); err != nil {
		return fmt.Errorf("failed to get scheduler operator for portworx. Err: %v", err)
	}

	if err = d.setDriver(); err != nil {
		return err
	}

	nodes, err := d.getClusterOnStart()
	if err != nil {
		return err
	}

	if len(nodes) == 0 {
		return fmt.Errorf("cluster inspect returned empty nodes")
	}

	err = d.updateNodes(nodes)
	if err != nil {
		return err
	}

	for _, n := range node.GetStorageDriverNodes() {
		if err := d.WaitDriverUpOnNode(n); err != nil {
			return err
		}
	}

	logrus.Infof("The following Portworx nodes are in the cluster:")
	for _, n := range nodes {
		logrus.Infof(
			"Node UID: %v Node IP: %v Node Status: %v",
			n.Id,
			n.DataIp,
			n.Status,
		)
	}

	return nil
}

func (d *portworx) RefreshDriverEndpoints() error {
	cluster, err := d.getClusterOnStart()
	if err != nil {
		return err
	}

	if len(cluster.Nodes) == 0 {
		return fmt.Errorf("cluster inspect returned empty nodes")
	}

	err = d.updateNodes(cluster.Nodes)
	if err != nil {
		return err
	}
	return nil
}

func (d *portworx) updateNodes(pxNodes []api.StorageNode) error {
	for _, n := range node.GetWorkerNodes() {
		if err := d.updateNode(n, pxNodes); err != nil {
			return err
		}
	}

	return nil
}

func (d *portworx) updateNode(n node.Node, pxNodes []api.StorageNode) error {
	isPX, err := d.schedOps.IsPXEnabled(n)
	if err != nil {
		return err
	}

	// No need to check in pxNodes if px is not installed
	if !isPX {
		return nil
	}

	for _, address := range n.Addresses {
		for _, pxNode := range pxNodes {
			if address == pxNode.DataIp || address == pxNode.MgmtIp || n.Name == pxNode.Hostname {
				n.VolDriverNodeID = pxNode.Id
				n.IsStorageDriverInstalled = isPX
				node.UpdateNode(n)
				return nil
			}
		}
	}

	// Return error where PX is not explicitly disabled but was not found installed
	return fmt.Errorf("failed to find px node for node: %v PX nodes: %v", n, pxNodes)
}

func (d *portworx) CleanupVolume(name string) error {
	volumes, err := d.getVolDriver().Enumerate(context.Background(), &api.SdkVolumeEnumerateRequest{}, nil)

	if err != nil {
		return err
	}

	for _, volID := range volumes.GetVolumeIds() {
		volumeInspectResponse, err := d.getVolDriver().Inspect(context.Background(), &api.SdkVolumeInspectRequest{VolumeId: volID})
		if err != nil {
			return err
		}
		v := volumeInspectResponse.Volume
		if v.Locator.Name == name {
			// First unmount this volume at all mount paths...
			for _, path := range v.AttachPath {
				if _, err = d.mountAttachManager.Unmount(context.Background(), &api.SdkVolumeUnmountRequest{VolumeId: v.Id, MountPath: path}); err != nil {
					err = fmt.Errorf(
						"error while unmounting %v at %v because of: %v",
						v.Id,
						path,
						err,
					)
					logrus.Infof("%v", err)
					return err
				}
			}

			if _, err = d.mountAttachManager.Detach(context.Background(), &api.SdkVolumeDetachRequest{VolumeId: v.Id}); err != nil {
				err = fmt.Errorf(
					"error while detaching %v because of: %v",
					v.Id,
					err,
				)
				logrus.Infof("%v", err)
				return err
			}

			if _, err := d.getVolDriver().Delete(context.Background(), &api.SdkVolumeDeleteRequest{VolumeId: v.Id}); err != nil {
				err = fmt.Errorf(
					"error while deleting %v because of: %v",
					v.Id,
					err,
				)
				logrus.Infof("%v", err)
				return err
			}

			logrus.Infof("successfully removed Portworx volume %v", name)

			return nil
		}
	}

	return nil
}

func (d *portworx) getPxNode(n node.Node /*, cManager cluster.Cluster*/) (api.StorageNode, error) {
	//if cManager == nil {
	//	cManager = d.getClusterManager()
	//}
	nodeInspectResponse, err := d.nodeManager.Inspect(context.Background(), &api.SdkNodeInspectRequest{NodeId: n.VolDriverNodeID})
	if (err == nil && nodeInspectResponse.Node.Status == api.Status_STATUS_OFFLINE) || (err != nil && nodeInspectResponse.Node.Status == api.Status_STATUS_NONE) {
		n, err = d.updateNodeID(n)
		if err != nil {
			return api.StorageNode{}, err
		}
		return d.getPxNode(n) //, cManager)
	} else if err != nil {
		return api.StorageNode{}, err
	}
	return *nodeInspectResponse.Node, nil
}

func (d *portworx) GetStorageDevices(n node.Node) ([]string, error) {
	//const (
	//	storageInfoKey = "STORAGE-INFO"
	//	resourcesKey   = "Resources"
	//	pathKey        = "path"
	//)
	//
	pxNode, err := d.getPxNode(n) //, nil)
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
		apiNode, err := d.getPxNode(n) //, nil)
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
		apiNode, err := d.getPxNode(n) //, nil)
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

func (d *portworx) ValidateCreateVolume(name string, params map[string]string) error {
	t := func() (interface{}, bool, error) {
		volumeInspectResponse, err := d.getVolDriver().Inspect(context.Background(), &api.SdkVolumeInspectRequest{VolumeId: name})
		if err != nil {
			return nil, true, err
		}
		return volumeInspectResponse.Volume, false, nil
	}

	out, err := task.DoRetryWithTimeout(t, inspectVolumeTimeout, inspectVolumeRetryInterval)
	if err != nil {
		return &ErrFailedToInspectVolume{
			ID:    name,
			Cause: fmt.Sprintf("Volume inspect returned err: %v", err),
		}
	}

	vol := out.(*api.Volume)

	// Status
	if vol.Status != api.VolumeStatus_VOLUME_STATUS_UP {
		return &ErrFailedToInspectVolume{
			ID: name,
			Cause: fmt.Sprintf("Volume has invalid status. Expected:%v Actual:%v",
				api.VolumeStatus_VOLUME_STATUS_UP, vol.Status),
		}
	}

	// State
	if vol.State == api.VolumeState_VOLUME_STATE_ERROR || vol.State == api.VolumeState_VOLUME_STATE_DELETED {
		return &ErrFailedToInspectVolume{
			ID:    name,
			Cause: fmt.Sprintf("Volume has invalid state. Actual:%v", vol.State),
		}
	}

	// if the volume is a clone or a snap, validate it's parent
	if vol.Readonly || vol.Source.Parent != "" {
		parentResp, err := d.getVolDriver().Inspect(context.Background(), &api.SdkVolumeInspectRequest{VolumeId: vol.Source.Parent})
		if err != nil {
			return &ErrFailedToInspectVolume{
				ID:    name,
				Cause: fmt.Sprintf("Could not get parent with ID [%s]", vol.Source.Parent),
			}
		}
		if err := d.schedOps.ValidateSnapshot(params, parentResp.Volume); err != nil {
			return &ErrFailedToInspectVolume{
				ID:    name,
				Cause: fmt.Sprintf("Snapshot/Clone validation failed. %v", err),
			}
		}
		return nil
	}

	// Labels
	var pxNodes []api.StorageNode
	for _, rs := range vol.ReplicaSets {
		for _, n := range rs.Nodes {
			nodeResponse, err := d.nodeManager.Inspect(context.Background(), &api.SdkNodeInspectRequest{NodeId: n})
			if err != nil {
				return &ErrFailedToInspectVolume{
					ID:    name,
					Cause: fmt.Sprintf("Failed to inspect replica set node: %s err: %v", n, err),
				}
			}

			pxNodes = append(pxNodes, *nodeResponse.Node)
		}
	}

	// Spec
	requestedSpec, requestedLocator, _, err := spec.NewSpecHandler().SpecFromOpts(params)
	if err != nil {
		return &ErrFailedToInspectVolume{
			ID:    name,
			Cause: fmt.Sprintf("failed to parse requested spec of volume. Err: %v", err),
		}
	}

	delete(vol.Locator.VolumeLabels, "pvc") // special handling for the new pvc label added in k8s

	// Params/Options
	for k, v := range params {
		switch k {
		case api.SpecNodes:
			if !reflect.DeepEqual(v, vol.Spec.ReplicaSet.Nodes) {
				return errFailedToInspectVolume(name, k, v, vol.Spec.ReplicaSet.Nodes)
			}
		case api.SpecParent:
			if v != vol.Source.Parent {
				return errFailedToInspectVolume(name, k, v, vol.Source.Parent)
			}
		case api.SpecEphemeral:
			if requestedSpec.Ephemeral != vol.Spec.Ephemeral {
				return errFailedToInspectVolume(name, k, requestedSpec.Ephemeral, vol.Spec.Ephemeral)
			}
		case api.SpecFilesystem:
			if requestedSpec.Format != vol.Spec.Format {
				return errFailedToInspectVolume(name, k, requestedSpec.Format, vol.Spec.Format)
			}
		case api.SpecBlockSize:
			if requestedSpec.BlockSize != vol.Spec.BlockSize {
				return errFailedToInspectVolume(name, k, requestedSpec.BlockSize, vol.Spec.BlockSize)
			}
		case api.SpecHaLevel:
			if requestedSpec.HaLevel != vol.Spec.HaLevel {
				return errFailedToInspectVolume(name, k, requestedSpec.HaLevel, vol.Spec.HaLevel)
			}
		case api.SpecPriorityAlias:
			// Since IO priority isn't guaranteed, we aren't validating it here.
		case api.SpecSnapshotInterval:
			if requestedSpec.SnapshotInterval != vol.Spec.SnapshotInterval {
				return errFailedToInspectVolume(name, k, requestedSpec.SnapshotInterval, vol.Spec.SnapshotInterval)
			}
		case api.SpecSnapshotSchedule:
			// TODO currently volume spec has a different format than request
			// i.e request "daily=12:00,7" turns into "- freq: daily\n  hour: 12\n  retain: 7\n" in volume spec
			//if requestedSpec.SnapshotSchedule != vol.Spec.SnapshotSchedule {
			//	return errFailedToInspectVolume(name, k, requestedSpec.SnapshotSchedule, vol.Spec.SnapshotSchedule)
			//}
		case api.SpecAggregationLevel:
			if requestedSpec.AggregationLevel != vol.Spec.AggregationLevel {
				return errFailedToInspectVolume(name, k, requestedSpec.AggregationLevel, vol.Spec.AggregationLevel)
			}
		case api.SpecShared:
			if requestedSpec.Shared != vol.Spec.Shared {
				return errFailedToInspectVolume(name, k, requestedSpec.Shared, vol.Spec.Shared)
			}
		case api.SpecSticky:
			if requestedSpec.Sticky != vol.Spec.Sticky {
				return errFailedToInspectVolume(name, k, requestedSpec.Sticky, vol.Spec.Sticky)
			}
		case api.SpecGroup:
			if !reflect.DeepEqual(requestedSpec.Group, vol.Spec.Group) {
				return errFailedToInspectVolume(name, k, requestedSpec.Group, vol.Spec.Group)
			}
		case api.SpecGroupEnforce:
			if requestedSpec.GroupEnforced != vol.Spec.GroupEnforced {
				return errFailedToInspectVolume(name, k, requestedSpec.GroupEnforced, vol.Spec.GroupEnforced)
			}
		// portworx injects pvc name and namespace labels so response object won't be equal to request
		case api.SpecLabels:
			for requestedLabelKey, requestedLabelValue := range requestedLocator.VolumeLabels {
				if labelValue, exists := vol.Locator.VolumeLabels[requestedLabelKey]; !exists || requestedLabelValue != labelValue {
					return errFailedToInspectVolume(name, k, requestedLocator.VolumeLabels, vol.Locator.VolumeLabels)
				}
			}
		case api.SpecIoProfile:
			if requestedSpec.IoProfile != vol.Spec.IoProfile {
				return errFailedToInspectVolume(name, k, requestedSpec.IoProfile, vol.Spec.IoProfile)
			}
		case api.SpecSize:
			if requestedSpec.Size != vol.Spec.Size {
				return errFailedToInspectVolume(name, k, requestedSpec.Size, vol.Spec.Size)
			}
		default:
		}
	}

	logrus.Infof("Successfully inspected volume: %v (%v)", vol.Locator.Name, vol.Id)
	return nil
}

func (d *portworx) ValidateUpdateVolume(vol *torpedovolume.Volume) error {
	name := d.schedOps.GetVolumeName(vol)
	t := func() (interface{}, bool, error) {
		volumeInspectResponse, err := d.getVolDriver().Inspect(context.Background(), &api.SdkVolumeInspectRequest{VolumeId: name})
		if err != nil {
			return nil, true, err
		}
		return volumeInspectResponse.Volume, false, nil
	}

	out, err := task.DoRetryWithTimeout(t, inspectVolumeTimeout, inspectVolumeRetryInterval)
	if err != nil {
		return &ErrFailedToInspectVolume{
			ID:    name,
			Cause: fmt.Sprintf("Volume inspect returned err: %v", err),
		}
	}

	respVol := out.(*api.Volume)

	// Size Update
	if respVol.Spec.Size != vol.Size {
		return &ErrFailedToInspectVolume{
			ID: name,
			Cause: fmt.Sprintf("Volume size differs. Expected:%v Actual:%v",
				vol.Size, respVol.Spec.Size),
		}
	}
	return nil
}

func errIsNotFound(err error) bool {
	statusErr, _ := status.FromError(err)
	return statusErr.Code() == codes.NotFound || strings.Contains(err.Error(), "code = NotFound")
}

func (d *portworx) ValidateDeleteVolume(vol *torpedovolume.Volume) error {
	name := d.schedOps.GetVolumeName(vol)
	t := func() (interface{}, bool, error) {
		volumeInspectResponse, err := d.getVolDriver().Inspect(context.Background(), &api.SdkVolumeInspectRequest{VolumeId: name})
		if err != nil && errIsNotFound(err) {
			return nil, false, nil
		} else if err != nil {
			return nil, true, err
		}
		// TODO remove shared validation when PWX-6894 and PWX-8790 are fixed
		if volumeInspectResponse.Volume != nil && !vol.Shared {
			return nil, true, fmt.Errorf("Volume %v is not yet removed from the system", name)
		}
		return nil, false, nil
	}

	_, err := task.DoRetryWithTimeout(t, validateDeleteVolumeTimeout, defaultRetryInterval)
	if err != nil {
		return &ErrFailedToDeleteVolume{
			ID:    name,
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
		}

	}
	logrus.Infof("Sleeping for %v for volume driver to go down.", waitVolDriverToCrash)
	time.Sleep(waitVolDriverToCrash)
	return nil
}

func (d *portworx) GetNodeForVolume(vol *torpedovolume.Volume) (*node.Node, error) {
	name := d.schedOps.GetVolumeName(vol)
	t := func() (interface{}, bool, error) {
		volumeInspectResponse, err := d.getVolDriver().Inspect(context.Background(), &api.SdkVolumeInspectRequest{VolumeId: name})
		if err != nil {
			logrus.Warnf("Failed to inspect volume: %s due to: %v", name, err)
			return nil, true, err
		}
		return volumeInspectResponse.Volume, false, nil
	}

	v, err := task.DoRetryWithTimeout(t, inspectVolumeTimeout, inspectVolumeRetryInterval)
	if err != nil {
		return nil, &ErrFailedToInspectVolume{
			ID:    name,
			Cause: err.Error(),
		}
	}

	r := func() (interface{}, bool, error) {
		pxVol := v.(*api.Volume)
		for _, n := range node.GetStorageDriverNodes() {
			if n.VolDriverNodeID == pxVol.AttachedOn {
				return &n, false, nil
			}
		}

		// Snapshots may not be attached to a node
		if pxVol.Source.Parent != "" {
			return nil, false, nil
		}

		return nil, true, fmt.Errorf("Volume: %s is not attached on any node", name)
	}

	n, err := task.DoRetryWithTimeout(r, validateVolumeAttachedTimeout, validateVolumeAttachedInterval)
	if err != nil {
		return nil, &ErrFailedToValidateAttachment{
			ID:    name,
			Cause: err.Error(),
		}
	}

	if n != nil {
		node := n.(*node.Node)
		return node, nil
	}

	return nil, nil
}

func (d *portworx) ExtractVolumeInfo(params string) (string, map[string]string, error) {
	ok, volParams, volName := spec.NewSpecHandler().SpecOptsFromString(params)
	if !ok {
		return params, nil, fmt.Errorf("Unable to parse the volume options")
	}
	return volName, volParams, nil
}

func (d *portworx) RandomizeVolumeName(params string) string {
	re := regexp.MustCompile("(name=)([0-9A-Za-z_-]+)(,)?")
	return re.ReplaceAllString(params, "${1}${2}_"+uuid.New()+"${3}")
}

func (d *portworx) getClusterOnStart() ([]api.StorageNode, error) {
	t := func() (interface{}, bool, error) {
		cluster, err := d.getClusterManager().InspectCurrent(context.Background(), &api.SdkClusterInspectCurrentRequest{})
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

func (d *portworx) getPxNodes() ([]api.StorageNode, error) {
	nodes := make([]api.StorageNode, 0)
	nodeEnumerateResp, err := d.nodeManager.Enumerate(context.Background(), &api.SdkNodeEnumerateRequest{})
	if err != nil {
		return nodes, err
	}
	for _, n := range nodeEnumerateResp.GetNodeIds() {
		nodeResp, err := d.nodeManager.Inspect(context.Background(), &api.SdkNodeInspectRequest{NodeId: n})
		if err != nil {
			return nodes, err
		}
		nodes = append(nodes, *nodeResp.Node)
	}
	return nodes, nil
}

func (d *portworx) WaitDriverUpOnNode(n node.Node) error {
	t := func() (interface{}, bool, error) {
		pxNode, err := d.getPxNode(n) //, nil)
		if err != nil {
			return "", true, &ErrFailedToWaitForPx{
				Node:  n,
				Cause: err.Error(),
			}
		}

		if pxNode.Status != api.Status_STATUS_OK {
			return "", true, &ErrFailedToWaitForPx{
				Node: n,
				Cause: fmt.Sprintf("px cluster is usable but node status is not ok. Expected: %v Actual: %v",
					api.Status_STATUS_OK, pxNode.Status),
			}
		}

		//storageStatus := d.getStorageStatus(n)
		//if storageStatus != storageStatusUp {
		//	return "", true, &ErrFailedToWaitForPx{
		//		Node: n,
		//		Cause: fmt.Sprintf("px cluster is usable but storage status is not ok. Expected: %v Actual: %v",
		//			storageStatusUp, storageStatus),
		//	}
		//}

		logrus.Infof("px on node %s is now up. status: %v", pxNode.Id, pxNode.Status)

		return "", false, nil
	}

	if _, err := task.DoRetryWithTimeout(t, validatePXStartTimeout, defaultRetryInterval); err != nil {
		return err
	}

	// Check if PX pod is up
	t = func() (interface{}, bool, error) {
		if !d.schedOps.IsPXReadyOnNode(n) {
			return "", true, &ErrFailedToWaitForPx{
				Node:  n,
				Cause: fmt.Sprintf("PX is not ready on %s after %v", n.Name, validatePXStartTimeout),
			}
		}
		return "", false, nil
	}

	if _, err := task.DoRetryWithTimeout(t, validatePXStartTimeout, defaultRetryInterval); err != nil {
		return err
	}

	return nil
}

func (d *portworx) WaitDriverDownOnNode(n node.Node) error {
	t := func() (interface{}, bool, error) {
		// Check if px is down on all node addresses. We don't want to keep track
		// which was the actual interface px was listening on before it went down
		nodeInspectResponse, err := d.nodeManager.Inspect(context.Background(), &api.SdkNodeInspectRequest{NodeId: n.VolDriverNodeID}, nil)
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

func (d *portworx) WaitForUpgrade(n node.Node, image, tag string) error {
	t := func() (interface{}, bool, error) {
		pxNode, err := d.getPxNode(n) //, nil)
		if err != nil {
			return nil, true, &ErrFailedToWaitForPx{
				Node:  n,
				Cause: err.Error(),
			}
		}

		if pxNode.Status != api.Status_STATUS_OK {
			return nil, true, &ErrFailedToWaitForPx{
				Node: n,
				Cause: fmt.Sprintf("px cluster is usable but node status is not ok. Expected: %v Actual: %v",
					api.Status_STATUS_OK, pxNode.Status),
			}
		}

		// filter out first 3 octets from the tag
		matches := regexp.MustCompile(`^(\d+\.\d+\.\d+).*`).FindStringSubmatch(tag)
		if len(matches) != 2 {
			return nil, false, &ErrFailedToUpgradeVolumeDriver{
				Version: fmt.Sprintf("%s:%s", image, tag),
				Cause:   fmt.Sprintf("failed to parse first 3 octets of version from new version tag: %s", tag),
			}
		}

		pxVersion := pxNode.NodeLabels[pxVersionLabel]
		if !strings.HasPrefix(pxVersion, matches[1]) {
			return nil, true, &ErrFailedToUpgradeVolumeDriver{
				Version: fmt.Sprintf("%s:%s", image, tag),
				Cause: fmt.Sprintf("version on node %s is still %s. It was expected to begin with: %s",
					n.VolDriverNodeID, pxVersion, matches[1]),
			}
		}
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
		volumeInspectResponse, err := d.volDriver.Inspect(context.Background(), &api.SdkVolumeInspectRequest{VolumeId: name})
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
	name := d.schedOps.GetVolumeName(vol)
	t := func() (interface{}, bool, error) {
		volumeInspectResponse, err := d.volDriver.Inspect(context.Background(), &api.SdkVolumeInspectRequest{VolumeId: name})
		if err != nil && errIsNotFound(err) {
			return nil, false, err
		} else if err != nil {
			return nil, true, err
		}

		su := api.VolumeSpecUpdate{}
		su.GetHaLevel()
		spec := &api.VolumeSpecUpdate{
			HaLevelOpt:          &api.VolumeSpecUpdate_HaLevel{HaLevel: int64(replFactor)},
			SnapshotIntervalOpt: &api.VolumeSpecUpdate_SnapshotInterval{SnapshotInterval: math.MaxUint32},
			ReplicaSet:          &api.ReplicaSet{},
		}
		_, err = d.volDriver.Update(context.Background(), &api.SdkVolumeUpdateRequest{
			VolumeId: volumeInspectResponse.Volume.Id,
			Spec:     spec,
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
				volumeInspectResponse, err = d.volDriver.Inspect(context.Background(), &api.SdkVolumeInspectRequest{VolumeId: name})
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
			ID:    name,
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
	name := d.schedOps.GetVolumeName(vol)
	t := func() (interface{}, bool, error) {
		volResp, err := d.volDriver.Inspect(context.Background(), &api.SdkVolumeInspectRequest{VolumeId: name})
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
			ID:    name,
			Cause: err.Error(),
		}
	}
	aggrLevel, ok := iAggrLevel.(uint32)
	if !ok {
		return 0, &ErrFailedToGetAggregationLevel{
			ID:    name,
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
		if err = d.testAndSetEndpoint(endpoint); err == nil {
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
	for _, n := range node.GetWorkerNodes() {
		for _, addr := range n.Addresses {
			if err = d.testAndSetEndpoint(addr); err == nil {
				return nil
			}
			logrus.Infof("testAndSetEndpoint failed for %v: %v", endpoint, err)
		}
	}

	return fmt.Errorf("failed to get endpoint for portworx volume driver")
}

func (d *portworx) testAndSetEndpoint(endpoint string) error {
	pxEndpoint := d.constructURL(endpoint)

	conn, err := grpc.Dial(pxEndpoint, grpc.WithInsecure())
	if err != nil {
		return err
	}

	d.volDriver = api.NewOpenStorageVolumeClient(conn)
	d.clusterManager = api.NewOpenStorageClusterClient(conn)
	d.nodeManager = api.NewOpenStorageNodeClient(conn)
	d.mountAttachManager = api.NewOpenStorageMountAttachClient(conn)
	d.clusterPairManager = api.NewOpenStorageClusterPairClient(conn)
	logrus.Infof("Using %v as endpoint for portworx volume driver", pxEndpoint)

	return nil
}

func (d *portworx) StartDriver(n node.Node) error {
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

func (d *portworx) UpgradeDriver(images []torpedovolume.Image) error {
	if len(images) == 0 {
		return fmt.Errorf("no version supplied for upgrading portworx")
	}

	partsOci := make([]string, 2)
	partsPx := make([]string, 2)

	for _, image := range images {
		switch image.Type {
		case "", "oci":
			partsOci = strings.Split(image.Version, ":")
		case "px":
			partsPx = strings.Split(image.Version, ":")
		}
	}
	version := partsOci[1]
	if len(partsPx) > 0 {
		version = partsPx[1]
	}
	logrus.Infof("upgrading portworx to %s", version)

	ociImage := partsOci[0]
	ociTag := partsOci[1]
	pxImage := partsPx[0]
	pxTag := partsPx[1]
	if err := d.schedOps.UpgradePortworx(ociImage, ociTag, pxImage, pxTag); err != nil {
		return &ErrFailedToUpgradeVolumeDriver{
			Version: version,
			Cause:   err.Error(),
		}
	}

	for _, n := range node.GetStorageDriverNodes() {
		image := ociImage
		tag := ociTag
		if len(pxImage) > 0 && len(pxTag) > 0 {
			image = pxImage
			tag = pxTag
		}
		if err := d.WaitForUpgrade(n, image, tag); err != nil {
			return err
		}
	}

	return nil
}

// GetClusterPairingInfo returns cluster pair information
func (d *portworx) GetClusterPairingInfo() (map[string]string, error) {
	pairInfo := make(map[string]string)

	resp, err := d.clusterPairManager.GetToken(context.Background(), &api.SdkClusterPairGetTokenRequest{})
	if err != nil {
		return nil, err
	}
	logrus.Infof("Response for token: %v", resp.Result.Token)

	// file up cluster pair info
	//pairInfo[clusterIP] = pxNodes[0].Addresses[0]
	pairInfo[tokenKey] = resp.Result.Token
	//pairInfo[clusterPort] = strconv.Itoa(pxdRestPort)

	return pairInfo, nil
}

func (d *portworx) DecommissionNode(n node.Node) error {

	if err := k8s.Instance().AddLabelOnNode(n.Name, "px/enabled", "remove"); err != nil {
		return &ErrFailedToDecommissionNode{
			Node:  n.Name,
			Cause: fmt.Sprintf("Failed to set label on node: %v. Err: %v", n.Name, err),
		}
	}

	if err := d.StopDriver([]node.Node{n}, false); err != nil {
		return &ErrFailedToDecommissionNode{
			Node:  n.Name,
			Cause: fmt.Sprintf("Failed to stop driver on node: %v. Err: %v", n.Name, err),
		}
	}
	//clusterManager := d.nodeManager
	//nodeResp, err := clusterManager.Inspect(context.Background(), &api.SdkNodeInspectRequest{NodeId:n.VolDriverNodeID})
	//if err != nil {
	//	return &ErrFailedToDecommissionNode{
	//		Node:  n.Name,
	//		Cause: fmt.Sprintf("Failed to inspect node: %v. Err: %v", nodeResp.Node, err),
	//	}
	//}
	//if err = clusterManager.Remove([]api.Node{pxNode}, false); err != nil {
	//	return &ErrFailedToDecommissionNode{
	//		Node:  n.Name,
	//		Cause: err.Error(),
	//	}
	//}
	return nil
}

func (d *portworx) RejoinNode(n node.Node) error {

	opts := node.ConnectionOpts{
		IgnoreError:     false,
		TimeBeforeRetry: defaultRetryInterval,
		Timeout:         defaultTimeout,
	}
	_, err := d.nodeDriver.RunCommand(n, "/opt/pwx/bin/pxctl sv node-wipe --all", opts)
	if err != nil {
		return &ErrFailedToRejoinNode{
			Node:  n.Name,
			Cause: err.Error(),
		}
	}
	if err := k8s.Instance().RemoveLabelOnNode(n.Name, "px/service"); err != nil {
		return &ErrFailedToRejoinNode{
			Node:  n.Name,
			Cause: fmt.Sprintf("Failed to set label on node: %v. Err: %v", n.Name, err),
		}
	}
	if err := k8s.Instance().RemoveLabelOnNode(n.Name, "px/enabled"); err != nil {
		return &ErrFailedToRejoinNode{
			Node:  n.Name,
			Cause: fmt.Sprintf("Failed to set label on node: %v. Err: %v", n.Name, err),
		}
	}
	return nil
}

func (d *portworx) GetNodeStatus(n node.Node) (*api.Status, error) {
	nodeResponse, err := d.nodeManager.Inspect(context.Background(), &api.SdkNodeInspectRequest{NodeId: n.VolDriverNodeID})
	if err != nil {
		return &nodeResponse.Node.Status, &ErrFailedToGetNodeStatus{
			Node:  n.Name,
			Cause: fmt.Sprintf("Failed to check node status: %v. Err: %v", nodeResponse.Node, err),
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

//func (d *portworx) getClusterManagerByAddress(addr string) (cluster.Cluster, error) {
//	pxEndpoint := d.constructURL(addr)
//	cClient, err := clusterclient.NewClusterClient(pxEndpoint, "v1")
//	if err != nil {
//		return nil, err
//	}
//
//	return clusterclient.ClusterManager(cClient), nil
//}

//func (d *portworx) getVolumeDriverByAddress(addr string) (volume.VolumeDriver, error) {
//	pxEndpoint := d.constructURL(addr)
//
//	dClient, err := volumeclient.NewDriverClient(pxEndpoint, DriverName, "", pxdClientSchedUserAgent)
//	if err != nil {
//		return nil, err
//	}
//
//	return volumeclient.VolumeDriver(dClient), nil
//}
//
func (d *portworx) maintenanceOp(n node.Node, op string) error {
	url := fmt.Sprintf("http://%s:%d", n.Addresses[0], pxdRestPort)
	c, err := client.NewClient(url, "", "")
	if err != nil {
		return err
	}
	req := c.Get().Resource(op)
	resp := req.Do()
	return resp.Error()
}

func (d *portworx) constructURL(ip string) string {
	return fmt.Sprintf("%s:%d", ip, defaultPxServicePort)
}

//func (d *portworx) getStorageStatus(n node.Node) string {
//	const (
//		storageInfoKey = "STORAGE-INFO"
//		statusKey      = "Status"
//	)
//	pxNode, err := d.getPxNode(n) //, nil)
//	if err != nil {
//		return err.Error()
//	}
//
//	storageInfo, ok := pxNode.Disks[""].NodeData[storageInfoKey]
//	if !ok {
//		return fmt.Sprintf("Unable to find storage info for node: %v", n.Name)
//	}
//	storageInfoMap := storageInfo.(map[string]interface{})
//
//	statusInfo, ok := storageInfoMap[statusKey]
//	if !ok || storageInfoMap == nil {
//		return fmt.Sprintf("Unable to find status info for node: %v", n.Name)
//	}
//	status := statusInfo.(string)
//	return status
//}

func (d *portworx) GetReplicaSetNodes(torpedovol *torpedovolume.Volume) ([]string, error) {
	var pxNodes []string
	volName := d.schedOps.GetVolumeName(torpedovol)
	volumeInspectResponse, err := d.getVolDriver().Inspect(context.Background(), &api.SdkVolumeInspectRequest{VolumeId: volName})
	if err != nil {
		return nil, &ErrFailedToInspectVolume{
			ID:    torpedovol.Name,
			Cause: err.Error(),
		}
	}

	for _, rs := range volumeInspectResponse.Volume.ReplicaSets {
		for _, n := range rs.Nodes {
			nodeInpectResponse, err := d.nodeManager.Inspect(context.Background(), &api.SdkNodeInspectRequest{NodeId: n})
			if err != nil {
				return nil, &ErrFailedToInspectVolume{
					ID:    torpedovol.Name,
					Cause: fmt.Sprintf("Failed to inspect replica set node: %s err: %v", n, err),
				}
			}
			nodeName := nodeInpectResponse.Node.SchedulerNodeName
			if nodeName == "" {
				nodeName = nodeInpectResponse.Node.Hostname
			}
			pxNodes = append(pxNodes, nodeName)
		}
	}
	return pxNodes, nil
}

func (d *portworx) updateNodeID(n node.Node) (node.Node, error) {
	nodes, err := d.getPxNodes()
	if err != nil {
		return n, err
	}
	for _, nd := range nodes {
		for _, addr := range n.Addresses {
			if nd.GetMgmtIp() == addr || nd.DataIp == addr {
				n.VolDriverNodeID = nd.Id
				node.UpdateNode(n)
				return n, nil
			}
		}
	}
	return n, fmt.Errorf("node %v not found in cluster", n)
}

func getGroupMatches(groupRegex *regexp.Regexp, str string) map[string]string {
	match := groupRegex.FindStringSubmatch(str)
	result := make(map[string]string)
	for i, name := range groupRegex.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return result
}

func init() {
	torpedovolume.Register(DriverName, &portworx{})
}
