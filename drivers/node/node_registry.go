package node

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/pborman/uuid"
)

var (
	nodeRegistry = make(map[string]Node)
	lock         sync.RWMutex
)

// AddNode adds a node to the node collection
func AddNode(n Node) error {
	if n.uuid != "" {
		return fmt.Errorf("UUID should not be set to add new node")
	}
	lock.Lock()
	defer lock.Unlock()
	n.uuid = uuid.New()
	nodeRegistry[n.uuid] = n
	return nil
}

// UpdateNode updates a given node if it exists in the node collection
func UpdateNode(n Node) error {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := nodeRegistry[n.uuid]; !ok {
		return fmt.Errorf("node to be updated does not exist")
	}
	nodeRegistry[n.uuid] = n
	return nil
}

// DeleteNode method delete a given node if exist in the node collection
func DeleteNode(n Node) error {
	if n.uuid == "" {
		return fmt.Errorf("UUID should be set to delete existing node")
	}
	lock.Lock()
	defer lock.Unlock()
	delete(nodeRegistry, n.uuid)
	return nil
}

// GetNodes returns all the nodes from the node collection
func GetNodes() []Node {
	var nodeList []Node
	for _, n := range nodeRegistry {
		nodeList = append(nodeList, n)
	}
	return nodeList
}

// GetWorkerNodes returns only the worker nodes/agent nodes
func GetWorkerNodes() []Node {
	var nodeList []Node
	for _, n := range nodeRegistry {
		if n.Type == TypeWorker {
			nodeList = append(nodeList, n)
		}
	}
	return nodeList
}

// StructToString returns the string representation of the specified struct
func StructToString(s interface{}) string {
	if stringer, ok := s.(fmt.Stringer); ok {
		return stringer.String()
	}
	v := reflect.ValueOf(s)
	t := v.Type()
	var fields []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			fieldVal := v.Field(i)
			var fieldString string
			if stringer, ok := fieldVal.Interface().(fmt.Stringer); ok {
				fieldString = fmt.Sprintf("%s: %s", field.Name, stringer.String())
			} else {
				switch fieldVal.Kind() {
				case reflect.Ptr:
					if fieldVal.IsNil() {
						fieldString = fmt.Sprintf("%s: nil", field.Name)
					} else if fieldVal.Type().Elem().Kind() == reflect.Struct {
						fieldString = fmt.Sprintf("%s: %s", field.Name, StructToString(fieldVal.Elem().Interface()))
					} else {
						fieldString = fmt.Sprintf("%s: %v", field.Name, fieldVal.Elem())
					}
				case reflect.Slice:
					if fieldVal.IsNil() {
						fieldString = fmt.Sprintf("%s: nil", field.Name)
					} else {
						fieldString = fmt.Sprintf("%s: %v", field.Name, fieldVal.Interface())
					}
				case reflect.Map:
					if fieldVal.IsNil() {
						fieldString = fmt.Sprintf("%s: nil", field.Name)
					} else {
						fieldString = fmt.Sprintf("%s: %v", field.Name, fieldVal.Interface())
					}
				case reflect.Struct:
					fieldString = fmt.Sprintf("%s: %s", field.Name, StructToString(fieldVal.Interface()))
				case reflect.String:
					if fieldVal.Len() == 0 {
						fieldString = fmt.Sprintf("%s: \"\"", field.Name)
					} else {
						fieldString = fmt.Sprintf("%s: %v", field.Name, fieldVal.Interface())
					}
				default:
					fieldString = fmt.Sprintf("%s: %v", field.Name, fieldVal.Interface())
				}
			}
			fields = append(fields, fieldString)
		}
	}
	return fmt.Sprintf("%s: {%s}", t.Name(), strings.Join(fields, ", "))
}

// GetMasterNodes returns only the master nodes/agent nodes
func GetMasterNodes() []Node {
	var nodeList []Node
	fmt.Printf("Node Registry - %v\n", nodeRegistry)
	fmt.Printf("Node Registry from the function - %s\n", StructToString(nodeRegistry))
	for _, n := range nodeRegistry {
		if n.Type == TypeMaster {
			nodeList = append(nodeList, n)
		}
	}
	return nodeList
}

// IsMasterNode returns true if node is a Masternode
func IsMasterNode(n Node) bool {
	for _, each := range GetMasterNodes() {
		if each.uuid == n.uuid {
			return true
		}
	}
	return false
}

// GetStorageDriverNodes returns only the worker node where storage
// driver is installed
func GetStorageDriverNodes() []Node {
	var nodeList []Node
	for _, n := range nodeRegistry {
		if n.Type == TypeWorker && n.IsStorageDriverInstalled {
			nodeList = append(nodeList, n)
		}
	}
	return nodeList
}

// IsStorageNode returns true if the node is a storage node, false otherwise
func IsStorageNode(n Node) bool {
	return len(n.Pools) > 0
}

// GetStorageNodes gets all the nodes with non-empty StoragePools
func GetStorageNodes() []Node {
	var nodeList []Node
	storageDriverNodes := GetStorageDriverNodes()
	for _, n := range storageDriverNodes {
		if IsStorageNode(n) {
			nodeList = append(nodeList, n)
		}
	}
	return nodeList
}

// GetStorageLessNodes gets all the nodes with empty StoragePools
func GetStorageLessNodes() []Node {
	var nodeList []Node
	storageDriverNodes := GetStorageDriverNodes()
	for _, n := range storageDriverNodes {
		if !IsStorageNode(n) {
			nodeList = append(nodeList, n)
		}
	}
	return nodeList
}

// GetNodesByTopologyZoneLabel gets all the nodes with Topology Zone Value matching
func GetNodesByTopologyZoneLabel(zone string) []Node {
	var nodeList []Node
	for _, n := range nodeRegistry {
		if n.TopologyZone == zone {
			nodeList = append(nodeList, n)
		}
	}
	return nodeList
}

// GetNodesByTopologyRegionLabel gets all the nodes with Topology Region Value matching
func GetNodesByTopologyRegionLabel(region string) []Node {
	var nodeList []Node
	for _, n := range nodeRegistry {
		if n.TopologyRegion == region {
			nodeList = append(nodeList, n)
		}
	}
	return nodeList
}

// GetMetadataNodes gets all the nodes which serves as internal kvdb metadata node
func GetMetadataNodes() []Node {
	var nodeList []Node
	for _, n := range nodeRegistry {
		if n.IsMetadataNode {
			nodeList = append(nodeList, n)
		}
	}
	return nodeList
}

// GetNodesByName returns map of nodes where the node name is the key
func GetNodesByName() map[string]Node {
	nodeMap := make(map[string]Node)
	for _, n := range nodeRegistry {
		nodeMap[n.Name] = n
	}
	return nodeMap
}

// GetNodesByVoDriverNodeID returns map of nodes where volume driver node id is the key
func GetNodesByVoDriverNodeID() map[string]Node {
	nodeMap := make(map[string]Node)
	for _, n := range nodeRegistry {
		nodeMap[n.VolDriverNodeID] = n
	}
	return nodeMap
}

// Contains checks if the node is present in the given list of nodes
func Contains(nodes []Node, n Node) bool {
	for _, value := range nodes {
		if value.Name == n.Name {
			return true
		}
	}
	return false
}

// GetNodeByName returns a node which matches with given name
func GetNodeByName(nodeName string) (Node, error) {
	for _, n := range nodeRegistry {
		if n.Name == nodeName {
			return n, nil
		}
	}
	return Node{}, fmt.Errorf("failed: Node [%s] not found in node registry", nodeName)
}

// GetNodeByIP return a node which matches with given IP
func GetNodeByIP(nodeIP string) (Node, error) {
	for _, n := range nodeRegistry {
		for _, addr := range n.Addresses {
			if addr == nodeIP {
				return n, nil
			}
		}
	}
	return Node{}, fmt.Errorf("failed: Node with [%s] not found in node registry", nodeIP)
}

// CleanupRegistry removes entry of all nodes from registry
func CleanupRegistry() {
	nodeRegistry = make(map[string]Node)
}

// GetNodeDetailsByNodeName get node details for a given node name
func GetNodeDetailsByNodeName(nodeName string) (Node, error) {
	storageNodes := GetStorageNodes()

	for _, each := range storageNodes {
		if each.Name == nodeName {
			return each, nil
		}
	}
	return Node{}, fmt.Errorf("failed to get Node Details by Node Name [%s] ", nodeName)
}

// GetNodeDetailsByNodeID get node details for a given node name
func GetNodeDetailsByNodeID(nodeID string) (Node, error) {
	storageNodes := GetStorageNodes()

	for _, each := range storageNodes {
		if each.Id == nodeID {
			return each, nil
		}
	}
	return Node{}, fmt.Errorf("failed to get Node Details by Node ID [%s] ", nodeID)
}
