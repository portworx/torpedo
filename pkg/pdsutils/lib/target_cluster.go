package lib

import (
	"github.com/sirupsen/logrus"
)

type TargetCluster struct {
	*cluster
}

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

func NewTargetCluster(context string) *TargetCluster {
	return &TargetCluster{
		cluster: &cluster{
			kubeconfig: context,
		},
	}
}
