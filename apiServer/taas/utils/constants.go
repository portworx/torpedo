package utils

import (
	"time"

	"github.com/portworx/sched-ops/k8s/core"
)

type HelmPayload struct {
	Command string `json:"command"`
}

const (
	mysql                        = "mysql"
	cassandra                    = "cassandra"
	fio                          = "fio"
	mongodb                      = "mongodb"
	pgbench                      = "pgbench"
	testName                     = "taas-automation"
	pxNameSpace                  = "kube-system"
	defaultWaitRebootRetry       = 10 * time.Second
	defaultCommandRetry          = 5 * time.Second
	defaultCommandTimeout        = 1 * time.Minute
	defaultTestConnectionTimeout = 15 * time.Minute
)

var (
	k8sCore = core.Instance()
)

type PxctlStatus struct {
	Status        string `json:"status,omitempty"`
	NodeID        string `json:"nodeId,omitempty"`
	IP            string `json:"ip,omitempty"`
	ClusterID     string `json:"clusterId,omitempty"`
	ClusterUUID   string `json:"clusterUUID,omitempty"`
	TotalUsed     string `json:"totalUsed,omitempty"`
	TotalCapacity string `json:"totalCapacity,omitempty"`
}
