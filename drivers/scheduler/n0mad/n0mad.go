package n0mad

import (
	"fmt"
	"strings"

	nomad "github.com/hashicorp/nomad/api"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/scheduler/spec"
	"github.com/sirupsen/logrus"
)

const (
	// SchedName is the name of the kubernetes scheduler driver implementation
	SchedName = "n0mad"
	// N0madMaster is a nodemad api concact (subject to change)
	N0madMaster = "http://70.0.57.114:4646"
)

type n0mad struct {
	specFactory    *spec.Factory
	nodeDriverName string
}

// This function connects to API and returns client
func getClient(apiEndpoint string) (*nomad.Client, error) {
	// Nomad config
	config := &nomad.Config{
		Address: apiEndpoint, // IP of nomad master
	}

	// Use config to connect to client API
	c, err := nomad.NewClient(config)
	if err != nil {
		return c, err
	}
	return c, nil
}

func (nm *n0mad) Init(specDir, volDriverName, nodeDriverName string) error {
	// TODO: Implement this method
	fmt.Printf("KOKADBG: INIT!\n")

	// Connect to nomad API
	c, err := getClient(N0madMaster)
	if err != nil {
		return err
	}

	// Get list of nodes
	nodeList, _, err := c.Nodes().List(&nomad.QueryOptions{})
	if err != nil {
		return err
	}

	// Get info about each node from the list
	for _, nomadNode := range nodeList {
		nodeInfo, _, err := c.Nodes().Info(nomadNode.ID, &nomad.QueryOptions{})
		if err != nil {
			return err
		}

		newNode := parseN0madNode(nodeInfo)
		if err := nm.IsNodeReady(newNode); err != nil {
			return err
		}
		if err := node.AddNode(newNode); err != nil {
			return err
		}
	}

	nm.specFactory, err = spec.NewFactory(specDir, nm)
	if err != nil {
		return err
	}

	nm.nodeDriverName = nodeDriverName
	return nil
}

func parseN0madNode(n *nomad.Node) node.Node {
	var nodeType node.Type
	logrus.Infof("NODE: %s, %s", isLeader(), n.HTTPAddr)
	if strings.Contains(n.HTTPAddr, isLeader()) {
		nodeType = node.TypeMaster
	} else {
		nodeType = node.TypeWorker
	}

	return node.Node{
		Name:      n.Name,
		Addresses: []string{strings.TrimSuffix(n.HTTPAddr, ":4646")},
		Type:      nodeType,
	}
}

// Get cluster leader
func isLeader() string {
	// Connect to client API
	c, err := getClient(N0madMaster)
	if err != nil {
		return ""
	}

	// Get leader
	leader, err := c.Status().Leader()
	if err != nil {
		return ""
	}

	leader = strings.TrimSuffix(leader, ":4647")
	return leader
}

// String returns the string name of this driver.
func (nm *n0mad) String() string {
	return SchedName
}

func (nm *n0mad) IsNodeReady(n node.Node) error {
	// TODO: Implement this method
	return nil
}

func (nm *n0mad) ParseSpecs(specDir string) ([]interface{}, error) {
	// TODO: Implement this method
	var specs []interface{}
	return specs, nil
}

func (nm *n0mad) Schedule(instanceID string, options scheduler.ScheduleOptions) ([]*scheduler.Context, error) {
	// TODO: Implement this method
	err := nm.Init("asd", "sdfsd", "dsfsd")
	return []*scheduler.Context{}, err
}

func (nm *n0mad) GetNodesForApp(ctx *scheduler.Context) ([]node.Node, error) {
	// TODO: Implement this method
	var result []node.Node
	return result, nil
}

func (nm *n0mad) WaitForRunning() error {
	// TODO: Implement this method
	return nil
}

func (nm *n0mad) AddTasks() error {
	// TODO: Implement this method
	return nil
}

func (nm *n0mad) Destroy() error {
	// TODO: Implement this method
	return nil
}

func (nm *n0mad) WaitForDestroy() error {
	// TODO: Implement this method
	return nil
}

func (nm *n0mad) DeleteTasks() error {
	// TODO: Implement this method
	return nil
}

func (nm *n0mad) GetVolumeParameters() error {
	// TODO: Implement this method
	return nil
}

func (nm *n0mad) InspectVolumes() error {
	// TODO: Implement this method
	return nil
}

func (nm *n0mad) DeleteVolumes() error {
	// TODO: Implement this method
	return nil
}

func (nm *n0mad) GetVolumes() error {
	// TODO: Implement this method
	return nil
}

func (nm *n0mad) ResizeVolume() error {
	// TODO: Implement this method
	return nil
}

func (nm *n0mad) GetSnapshots() error {
	// TODO: Implement this method
	return nil
}

func (nm *n0mad) Describe() error {
	// TODO: Implement this method
	return nil
}

func (nm *n0mad) ScaleApplication() (map[string]int32, error) {
	// TODO: Implement this method
	var test map[string]int32
	return test, nil
}

func (nm *n0mad) GetScaledFactorMap() (map[string]int32, error) {
	// TODO: Implement this method
	var test map[string]int32
	return test, nil
}

func (nm *n0mad) StopSchedOnNode() error {
	// TODO: Implement this method
	return nil
}

func (nm *n0mad) StartSchedOnNode() error {
	// TODO: Implement this method
	return nil
}

func (nm *n0mad) RescanSpecs() error {
	// TODO: Implement this method
	return nil
}
