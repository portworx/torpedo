package lib

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// ControlPlane struct comprise of cluster object
type ControlPlane struct {
	*cluster
}

// NewControlPlane function create ControlPlane object.
func NewControlPlane(context string) *ControlPlane {
	return &ControlPlane{
		cluster: &cluster{
			kubeconfig: context,
		},
	}
}

// LogStatus return logs for PDS components.
func (cp *ControlPlane) LogStatus() {
	logrus.Info("Control plane:")

	cp.describePods(pdsSystemNamespace)

	logrus.Info("API server logs:")
	cp.logComponent(pdsSystemNamespace, "api-server")

	logrus.Info("API worker logs:\n")
	cp.logComponent(pdsSystemNamespace, "api-worker")

	logrus.Info("faktory logs:\n")
	cp.logComponent(pdsSystemNamespace, "faktory")
}

// IsReachbale check for the control plane reachability.
func (cp *ControlPlane) IsReachbale(url string) bool {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	_, err := client.Get(url)
	if err != nil {
		logrus.Error(err.Error())
		return false
	}
	return true

}
