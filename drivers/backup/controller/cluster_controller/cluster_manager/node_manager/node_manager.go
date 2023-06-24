package node_manager

import "github.com/portworx/torpedo/drivers/backup/controller/cluster_controller"

type NodeMetaData struct {
	NodeName string
}

func (m *NodeMetaData) GetNodeName() string {
	return m.NodeName
}

func (m *NodeMetaData) SetNodeName(name string) {
	m.NodeName = name
}

func (m *NodeMetaData) GetNodeUid() string {
	return m.GetNodeName()
}

func NewNodeMetaData() *NodeMetaData {
	newNodeMetaData := &NodeMetaData{}
	newNodeMetaData.SetNodeName("")
	return newNodeMetaData
}

type NodeConfig struct {
	NodeMetaData      *NodeMetaData
	ClusterController *cluster_controller.ClusterController
}

func (c *NodeConfig) GetNodeMetaData() *NodeMetaData {
	return c.NodeMetaData
}

func (c *NodeConfig) SetNodeMetaData(metaData *NodeMetaData) {
	c.NodeMetaData = metaData
}

func (c *NodeConfig) GetClusterController() *cluster_controller.ClusterController {
	return c.ClusterController
}

func (c *NodeConfig) SetClusterController(controller *cluster_controller.ClusterController) {
	c.ClusterController = controller
}

type Node struct{}

func NewNode() *Node {
	newNode := &Node{}
	return newNode
}

type NodeManager struct {
	NodeMap         map[string]*Node
	RemovedNodesMap map[string][]*Node
}

func (m *NodeManager) GetNodeMap() map[string]*Node {
	return m.NodeMap
}

func (m *NodeManager) SetNodeMap(nodeMap map[string]*Node) {
	m.NodeMap = nodeMap
}

func (m *NodeManager) GetRemovedNodesMap() map[string][]*Node {
	return m.RemovedNodesMap
}

func (m *NodeManager) SetRemovedNodesMap(removedNodesMap map[string][]*Node) {
	m.RemovedNodesMap = removedNodesMap
}

func (m *NodeManager) GetNode(nodeUid string) *Node {
	return m.NodeMap[nodeUid]
}

func (m *NodeManager) IsNodePresent(nodeUid string) bool {
	_, isPresent := m.NodeMap[nodeUid]
	return isPresent
}

func (m *NodeManager) SetNode(nodeUid string, node *Node) {
	m.NodeMap[nodeUid] = node
}

func (m *NodeManager) DeleteNode(nodeUid string) {
	delete(m.NodeMap, nodeUid)
}

func (m *NodeManager) RemoveNode(nodeUid string) {
	m.RemovedNodesMap[nodeUid] = append(m.RemovedNodesMap[nodeUid], m.NodeMap[nodeUid])
	m.DeleteNode(nodeUid)
}

func NewNodeManager() *NodeManager {
	newNodeManger := &NodeManager{}
	newNodeManger.SetNodeMap(make(map[string]*Node, 0))
	newNodeManger.SetRemovedNodesMap(make(map[string][]*Node, 0))
	return newNodeManger
}
