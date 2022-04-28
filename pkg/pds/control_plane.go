package pds

import "time"

type ControlPlane struct {
	*cluster
}

func NewControlPlane(kubeconfig string) *ControlPlane {
	return &ControlPlane{
		cluster: &cluster{
			kubeconfig: kubeconfig,
		},
	}
}

func (cp *ControlPlane) ComponentLogsSince(since time.Time) []componentLog {
	return []componentLog{
		{"API Server", cp.logComponent(pdsSystemNamespace, "api-server", since)},
		{"API Worker", cp.logComponent(pdsSystemNamespace, "api-worker", since)},
		{"Faktory", cp.logComponent(pdsSystemNamespace, "faktory", since)},
	}
}
