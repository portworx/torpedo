package cluster_metadata

import (
	. "github.com/portworx/torpedo/drivers/backup/cluster_controller/cluster_utils"
)

// ClusterMetaData represents the metadata for cluster_manager.Cluster
type ClusterMetaData struct {
	ConfigPath string
}

// GetConfigPath returns the ConfigPath associated with the ClusterMetaData
func (m *ClusterMetaData) GetConfigPath() string {
	return m.ConfigPath
}

// SetConfigPath sets the ConfigPath for the ClusterMetaData
func (m *ClusterMetaData) SetConfigPath(configPath string) *ClusterMetaData {
	m.ConfigPath = configPath
	return m
}

// GetClusterUID returns the cluster_manager.Cluster UID
func (m *ClusterMetaData) GetClusterUID() string {
	return m.GetConfigPath()
}

// NewClusterMetaData creates a new instance of the ClusterMetaData
func NewClusterMetaData(configPath string) *ClusterMetaData {
	newClusterMetaData := &ClusterMetaData{}
	newClusterMetaData.SetConfigPath(configPath)
	return newClusterMetaData
}

// NewDefaultClusterMetaData creates a new instance of the ClusterMetaData with default values
func NewDefaultClusterMetaData() *ClusterMetaData {
	return NewClusterMetaData(DefaultConfigPath)
}
