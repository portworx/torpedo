package cluster_spec

import "github.com/portworx/torpedo/tests"

const (
	// DefaultScheduler is the default Scheduler for ClusterSpec
	DefaultScheduler = "k8s"
	// DefaultHyperconverged is the default Hyperconverged for ClusterSpec
	DefaultHyperconverged = true
)

var (
	// DefaultStorageProvisioner is the default StorageProvisioner for ClusterSpec
	DefaultStorageProvisioner = tests.Inst().Provisioner
)

// ClusterSpec represents the specification for a Cluster
type ClusterSpec struct {
	ConfigPath         string
	Scheduler          string
	Hyperconverged     bool // TODO: Handle non-hyperconverged clusters
	StorageProvisioner string
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

// GetScheduler returns the Scheduler associated with the ClusterSpec
func (s *ClusterSpec) GetScheduler() string {
	return s.Scheduler
}

// SetScheduler sets the Scheduler for the ClusterSpec
func (s *ClusterSpec) SetScheduler(scheduler string) *ClusterSpec {
	s.Scheduler = scheduler
	return s
}

// GetHyperconverged returns the Hyperconverged associated with the ClusterSpec
func (s *ClusterSpec) GetHyperconverged() bool {
	return s.Hyperconverged
}

// SetHyperconverged sets the Hyperconverged for the ClusterSpec
func (s *ClusterSpec) SetHyperconverged(hyperconverged bool) *ClusterSpec {
	s.Hyperconverged = hyperconverged
	return s
}

// GetStorageProvisioner returns the StorageProvisioner associated with the ClusterSpec
func (s *ClusterSpec) GetStorageProvisioner() string {
	return s.StorageProvisioner
}

// SetStorageProvisioner sets the StorageProvisioner for the ClusterSpec
func (s *ClusterSpec) SetStorageProvisioner(provisioner string) *ClusterSpec {
	s.StorageProvisioner = provisioner
	return s
}

// NewClusterSpec creates a new instance of the ClusterSpec
func NewClusterSpec(configPath string, scheduler string, hyperconverged bool, storageProvisioner string) *ClusterSpec {
	clusterSpec := &ClusterSpec{}
	clusterSpec.SetConfigPath(configPath)
	clusterSpec.SetScheduler(scheduler)
	clusterSpec.SetHyperconverged(hyperconverged)
	clusterSpec.SetStorageProvisioner(storageProvisioner)
	return clusterSpec
}

// NewDefaultClusterSpec creates a new instance of the ClusterSpec with default values
func NewDefaultClusterSpec(configPath string) *ClusterSpec {
	return NewClusterSpec(configPath, DefaultScheduler, DefaultHyperconverged, DefaultStorageProvisioner)
}
