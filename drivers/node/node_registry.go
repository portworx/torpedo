package node

import (
	"fmt"
	"sync"

	"github.com/pborman/uuid"
)

type NodeRegistry struct {
	sync.RWMutex
	Nodes map[string]Node
}

// AddNode adds a node to the node collection
func (nr *NodeRegistry) AddNode(n Node) error {
	if n.Uuid != "" {
		return fmt.Errorf("UUID should not be set to add new node")
	}
	nr.Lock()
	defer nr.Unlock()
	n.Uuid = uuid.New()
	nr.Nodes[n.Uuid] = n
	return nil
}

// UpdateNode updates a given node if it exists in the node collection
func (nr *NodeRegistry) UpdateNode(n Node) error {
	nr.Lock()
	defer nr.Unlock()
	if _, ok := nr.Nodes[n.Uuid]; !ok {
		return fmt.Errorf("node to be updated does not exist")
	}
	nr.Nodes[n.Uuid] = n
	return nil
}

// DeleteNode method delete a given node if exist in the node collection
func (nr *NodeRegistry) DeleteNode(n Node) error {
	if n.Uuid == "" {
		return fmt.Errorf("UUID should be set to delete existing node")
	}
	nr.Lock()
	defer nr.Unlock()
	delete(nr.Nodes, n.Uuid)
	return nil
}

// GetNodes returns all the nodes from the node collection
func (nr *NodeRegistry) GetNodes() []Node {
	var nodeList []Node
	for _, n := range nr.Nodes {
		nodeList = append(nodeList, n)
	}
	return nodeList
}

// GetWorkerNodes returns only the worker nodes/agent nodes
func (nr *NodeRegistry) GetWorkerNodes() []Node {
	var nodeList []Node
	for _, n := range nr.Nodes {
		if n.Type == TypeWorker {
			nodeList = append(nodeList, n)
		}
	}
	return nodeList
}

// GetMasterNodes returns only the master nodes/agent nodes
func (nr *NodeRegistry) GetMasterNodes() []Node {
	var nodeList []Node
	for _, n := range nr.Nodes {
		if n.Type == TypeMaster {
			nodeList = append(nodeList, n)
		}
	}
	return nodeList
}

// IsMasterNode returns true if node is a Masternode
func (nr *NodeRegistry) IsMasterNode(n Node) bool {
	for _, each := range nr.GetMasterNodes() {
		if each.Uuid == n.Uuid {
			return true
		}
	}
	return false
}

// GetStorageDriverNodes returns only the worker node where storage
// driver is installed
func (nr *NodeRegistry) GetStorageDriverNodes() []Node {
	var nodeList []Node
	for _, n := range nr.Nodes {
		if n.Type == TypeWorker && n.IsStorageDriverInstalled {
			nodeList = append(nodeList, n)
		}
	}
	return nodeList
}

// IsStorageNode returns true if the node is a storage node, false otherwise
func (nr *NodeRegistry) IsStorageNode(n Node) bool {
	return len(n.Pools) > 0
}

// GetStorageNodes gets all the nodes with non-empty StoragePools
func (nr *NodeRegistry) GetStorageNodes() []Node {
	var nodeList []Node
	storageDriverNodes := nr.GetStorageDriverNodes()
	for _, n := range storageDriverNodes {
		if nr.IsStorageNode(n) {
			nodeList = append(nodeList, n)
		}
	}
	return nodeList
}

// GetStorageLessNodes gets all the nodes with empty StoragePools
func (nr *NodeRegistry) GetStorageLessNodes() []Node {
	var nodeList []Node
	storageDriverNodes := nr.GetStorageDriverNodes()
	for _, n := range storageDriverNodes {
		if !nr.IsStorageNode(n) {
			nodeList = append(nodeList, n)
		}
	}
	return nodeList
}

// GetNodesByTopologyZoneLabel gets all the nodes with Topology Zone Value matching
func (nr *NodeRegistry) GetNodesByTopologyZoneLabel(zone string) []Node {
	var nodeList []Node
	for _, n := range nr.Nodes {
		if n.TopologyZone == zone {
			nodeList = append(nodeList, n)
		}
	}
	return nodeList
}

// GetNodesByTopologyRegionLabel gets all the nodes with Topology Region Value matching
func (nr *NodeRegistry) GetNodesByTopologyRegionLabel(region string) []Node {
	var nodeList []Node
	for _, n := range nr.Nodes {
		if n.TopologyRegion == region {
			nodeList = append(nodeList, n)
		}
	}
	return nodeList
}

// GetMetadataNodes gets all the nodes which serves as internal kvdb metadata node
func (nr *NodeRegistry) GetMetadataNodes() []Node {
	var nodeList []Node
	for _, n := range nr.Nodes {
		if n.IsMetadataNode {
			nodeList = append(nodeList, n)
		}
	}
	return nodeList
}

// GetNodesByName returns map of nodes where the node name is the key
func (nr *NodeRegistry) GetNodesByName() map[string]Node {
	nodeMap := make(map[string]Node)
	for _, n := range nr.Nodes {
		nodeMap[n.Name] = n
	}
	return nodeMap
}

// GetNodesByVoDriverNodeID returns map of nodes where volume driver node id is the key
func (nr *NodeRegistry) GetNodesByVoDriverNodeID() map[string]Node {
	nodeMap := make(map[string]Node)
	for _, n := range nr.Nodes {
		nodeMap[n.VolDriverNodeID] = n
	}
	return nodeMap
}

// Contains checks if the node is present in the given list of nodes
func (nr *NodeRegistry) Contains(nodes []Node, n Node) bool {
	for _, value := range nodes {
		if value.Name == n.Name {
			return true
		}
	}
	return false
}

// GetNodeByName returns a node which matches with given name
func (nr *NodeRegistry) GetNodeByName(nodeName string) (Node, error) {
	for _, n := range nr.Nodes {
		if n.Name == nodeName {
			return n, nil
		}
	}
	return Node{}, fmt.Errorf("failed: Node [%s] not found in node registry", nodeName)
}

// GetNodeByIP return a node which matches with given IP
func (nr *NodeRegistry) GetNodeByIP(nodeIP string) (Node, error) {
	for _, n := range nr.Nodes {
		for _, addr := range n.Addresses {
			if addr == nodeIP {
				return n, nil
			}
		}
	}
	return Node{}, fmt.Errorf("failed: Node with [%s] not found in node registry", nodeIP)
}

// CleanupRegistry removes entry of all nodes from registry
func (nr *NodeRegistry) CleanupRegistry() {
	nr.Nodes = make(map[string]Node)
}

// GetNodeDetailsByNodeName get node details for a given node name
func (nr *NodeRegistry) GetNodeDetailsByNodeName(nodeName string) (Node, error) {
	storageNodes := nr.GetStorageNodes()

	for _, each := range storageNodes {
		if each.Name == nodeName {
			return each, nil
		}
	}
	return Node{}, fmt.Errorf("failed to get Node Details by Node Name [%s] ", nodeName)
}

// GetNodeDetailsByNodeID get node details for a given node name
func (nr *NodeRegistry) GetNodeDetailsByNodeID(nodeID string) (Node, error) {
	storageNodes := nr.GetStorageNodes()

	for _, each := range storageNodes {
		if each.Id == nodeID {
			return each, nil
		}
	}
	return Node{}, fmt.Errorf("failed to get Node Details by Node ID [%s] ", nodeID)
}
