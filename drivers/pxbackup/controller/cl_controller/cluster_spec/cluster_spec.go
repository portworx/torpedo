package cluster_spec

// ClusterSpec represents the specification for a Cluster
type ClusterSpec struct {
	ConfigPath string
}

// GetConfigPath returns the ConfigPath associated with the ClusterSpec
func (s *ClusterSpec) GetConfigPath() string {
	return s.ConfigPath
}

// SetConfigPath sets the ConfigPath for the ClusterSpec
func (s *ClusterSpec) SetConfigPath(configPath string) *ClusterSpec {
	s.ConfigPath = configPath
	return s
}

// NewClusterSpec creates a new instance of the ClusterSpec
func NewClusterSpec(configPath string) *ClusterSpec {
	clusterSpec := &ClusterSpec{}
	clusterSpec.SetConfigPath(configPath)
	return clusterSpec
}

// NewDefaultClusterSpec creates a new instance of the ClusterSpec with default values
func NewDefaultClusterSpec(configPath string) *ClusterSpec {
	return NewClusterSpec(configPath)
}
