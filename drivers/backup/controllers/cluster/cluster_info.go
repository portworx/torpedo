package cluster

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/backup/utils"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/tests"
)

// ClusterMetaData holds the metadata of a cluster
type ClusterMetaData struct {
	Id         string
	Name       string
	ConfigPath string
}

// ClusterInfo holds the necessary information to create a ClusterController
type ClusterInfo struct {
	*ClusterMetaData
	inCluster             bool
	hyperConverged        bool
	storageLessNodeLabels map[string]string
	storageLessNodes      []node.Node
}

// setStorageLessLabels sets the storage-less labels of the ClusterInfo
func setStorageLessLabels(c *ClusterInfo) {
	c.storageLessNodeLabels = make(map[string]string, 0)
	c.storageLessNodeLabels["storage"] = "NO"
}

// setStorageLessNodes sets the storage-less nodes of the ClusterInfo
func setStorageLessNodes(c *ClusterInfo) error {
	c.storageLessNodes = make([]node.Node, 0)
	storageLessNodes := node.GetStorageLessNodes()
	if len(storageLessNodes) == 0 {
		err := fmt.Errorf("no storage less nodes available in the cluster at path [%s]", c.ConfigPath)
		return utils.ProcessError(err)
	}
	c.storageLessNodes = storageLessNodes
	for _, storageLessNode := range c.storageLessNodes {
		for labelKey, labelValue := range c.storageLessNodeLabels {
			err := tests.Inst().S.AddLabelOnNode(storageLessNode, labelKey, labelValue)
			if err != nil {
				debugMessage := fmt.Sprintf("storage-less node: [%v]; label: key [%s], value [%s]", storageLessNode, labelKey, labelValue)
				return utils.ProcessError(err, debugMessage)
			}
		}
	}
	return nil
}

// IsInCluster configures the ClusterInfo to represent an in-cluster
func (c *ClusterInfo) IsInCluster() *ClusterInfo {
	c.ConfigPath = GlobalInClusterConfigPath
	c.inCluster = true
	return c
}

// IsHyperConverged configures the ClusterInfo to represent a hyper-converged-cluster
func (c *ClusterInfo) IsHyperConverged() *ClusterInfo {
	c.hyperConverged = true
	return c
}

// GetController returns a new ClusterController instance based on the ClusterInfo
func (c *ClusterInfo) GetController() (*ClusterController, error) {
	clController := &ClusterController{
		ClusterInfo:    c,
		namespaces:     make(map[string]*NamespaceInfo),
		appKeyCountMap: make(map[string]int),
	}
	if !clController.ClusterInfo.hyperConverged {
		setStorageLessLabels(c)
		err := setStorageLessNodes(c)
		if err != nil {
			debugMessage := fmt.Sprintf("cluster-info: [%v]", c)
			return nil, utils.ProcessError(err, debugMessage)
		}
	}
	return clController, nil
}
