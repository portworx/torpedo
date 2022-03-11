package lib

import (
	"github.com/sirupsen/logrus"
)

// TargetCluster khash
type TargetCluster struct {
	*cluster
}

// LogStatus khaskh
func (cp *TargetCluster) LogStatus() {
	logrus.Info("Target Cluster:")

	cp.describePods(pdsSystemNamespace)

	logrus.Info("API server logs:")
	cp.logComponent(pdsSystemNamespace, "api-server")

	logrus.Info("API worker logs:\n")
	cp.logComponent(pdsSystemNamespace, "api-worker")

	logrus.Info("faktory logs:\n")
	cp.logComponent(pdsSystemNamespace, "faktory")
}

// NewTargetCluster lsajajsklj
func NewTargetCluster(context string) *TargetCluster {
	return &TargetCluster{
		cluster: &cluster{
			kubeconfig: context,
		},
	}
}
