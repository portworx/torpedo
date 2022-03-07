package lib

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type ControlPlane struct {
	*cluster
}

func NewControlPlane(context string) *ControlPlane {
	return &ControlPlane{
		cluster: &cluster{
			kubeconfig: context,
		},
	}
}

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

func (cp *ControlPlane) IsReachbale(url string) bool {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(url)
	if err != nil {
		logrus.Error(err.Error())
		return false
	} else {
		logrus.Info(string(resp.StatusCode) + resp.Status)
	}
	return true
}
