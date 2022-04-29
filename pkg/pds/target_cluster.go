package pds

type TargetCluster struct {
	*cluster
}

func NewTargetCluster(kubeconfig string) (*TargetCluster, error) {
	cluster, err := newCluster(kubeconfig)
	if err != nil {
		return nil, err
	}
	return &TargetCluster{cluster}, nil
}
