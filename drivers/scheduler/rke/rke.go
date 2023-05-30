package rke

import (
	"fmt"

	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	kube "github.com/portworx/torpedo/drivers/scheduler/k8s"
)

const (
	// SchedName is the name of the kubernetes scheduler driver implementation
	SchedName = "rke"
	// SystemdSchedServiceName is the name of the system service resposible for scheduling
	SystemdSchedServiceName = "kubelet"
)

type Rke struct {
	kube.K8s
}

// DeepCopy create a deepcopy of rke and sets all sched-ops instances to default
func (k *Rke) DeepCopy() scheduler.Driver {
	if k == nil {
		return nil
	}
	out := k.K8s.DeepCopy()
	return out
}

func (k *Rke) SaveSchedulerLogsToFile(n node.Node, location string) error {
	// requires 2>&1 since docker logs command send the logs to stdrr instead of sdout
	cmd := fmt.Sprintf("docker logs %s > %s/kubelet.log 2>&1", SystemdSchedServiceName, location)
	_, err := k.NodeDriver.RunCommand(n, cmd, node.ConnectionOpts{
		Timeout:         kube.DefaultTimeout,
		TimeBeforeRetry: kube.DefaultRetryInterval,
		Sudo:            true,
	})
	return err
}

func init() {
	k := &Rke{}
	scheduler.Register(SchedName, k)
}
