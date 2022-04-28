package pds

type TargetCluster struct {
	*cluster
}

func NewTargetCluster(context string) *TargetCluster {
	return &TargetCluster{
		cluster: &cluster{
			kubeconfig: context,
		},
	}
}
