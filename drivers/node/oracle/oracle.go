package oracle

import (
	"errors"
	"os"
	"time"

	"github.com/portworx/torpedo/drivers/node/ssh"

	"github.com/libopenstorage/cloudops"
	oracleOps "github.com/libopenstorage/cloudops/oracle"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/sirupsen/logrus"
)

const (
	// DriverName is the name of the aws driver
	DriverName = "oracle"
)

type oracle struct {
	ssh.SSH
	ops               cloudops.Ops
	instanceID        string
	instanceGroupName string
}

func (o *oracle) String() string {
	return DriverName
}

// Init initializes the node driver for oracle under the given scheduler
func (o *oracle) Init(nodeOpts node.InitOptions) error {
	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		return errors.New("INSTANCE_ID not found")
	}
	o.instanceID = instanceID

	instanceGroupName := os.Getenv("INSTANCE_GROUP_NAME")
	if instanceGroupName == "" {
		return errors.New("INSTANCE_GROUP_NAME not found")
	}
	o.instanceGroupName = instanceGroupName

	ops, err := oracleOps.NewClient()
	if err != nil {
		return err
	}
	o.ops = ops

	return nil
}

// SetASGClusterSize sets node count per zone
func (o *oracle) SetASGClusterSize(perZoneCount int64, timeout time.Duration) error {
	err := o.ops.SetInstanceGroupSize(o.instanceGroupName, perZoneCount, timeout)
	if err != nil {
		logrus.Errorf("failed to set size of node pool %s. Error: %v", o.instanceGroupName, err)
		return err
	}

	return nil
}

// GetASGClusterSize gets node count for cluster
func (o *oracle) GetASGClusterSize() (int64, error) {
	size, err := o.ops.GetInstanceGroupSize(o.instanceGroupName)
	if err != nil {
		logrus.Errorf("failed to get size of node pool %s. Error: %v", o.instanceGroupName, err)
		return 0, err
	}
	return size, nil
}

// GetZones returns list of zones in which cluster is running
func (o *oracle) GetZones() ([]string, error) {
	asgInfo, err := o.ops.InspectInstanceGroupForInstance(o.instanceID)
	if err != nil {
		return []string{}, err
	}

	return asgInfo.Zones, nil
}

// SetClusterVersion sets desired version for cluster and its node pools
func (o *oracle) SetClusterVersion(version string, timeout time.Duration) error {
	logrus.Info("[Torpedo] Setting cluster version to :", version)
	err := o.ops.SetClusterVersion(version, timeout)
	if err != nil {
		logrus.Errorf("failed to set version for cluster. Error: %v", err)
		return err
	}
	logrus.Info("[Torpedo] Cluster version set successfully. Setting up node group version now ...")

	err = o.ops.SetInstanceGroupVersion(o.instanceGroupName, version, timeout)
	if err != nil {
		logrus.Errorf("failed to set version for instance group %s. Error: %v", o.instanceGroupName, err)
		return err
	}
	logrus.Info("[Torpedo] Node group version set successfully for group ", o.instanceGroupName)

	return nil
}

// DeleteNode deletes the given node
func (o *oracle) DeleteNode(node node.Node, timeout time.Duration) error {
	logrus.Info("[Torpedo] Deleting node with ID :", node.Id)
	err := o.ops.DeleteInstance(node.Id, "", timeout)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	i := &oracle{
		SSH: *ssh.New(),
	}

	node.Register(DriverName, i)
}
