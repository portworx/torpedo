package pds

// TargetCluster wraps a PDS target cluster.
type TargetCluster struct {
	*cluster
}

// NewTargetCluster creates a TargetCluster instance with the specified kubeconfig.
// Fails if a kubernetes go-client cannot be configured based on the kubeconfig.
func NewTargetCluster(kubeconfig string) (*TargetCluster, error) {
	cluster, err := newCluster(kubeconfig)
	if err != nil {
		return nil, err
	}
	return &TargetCluster{cluster}, nil
}
