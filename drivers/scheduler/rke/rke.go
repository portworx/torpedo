package rke

import (
	"fmt"
	rancherclientbase "github.com/rancher/norman/clientbase"
	rancherclient "github.com/rancher/rancher/pkg/client/generated/management/v3"

	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	kube "github.com/portworx/torpedo/drivers/scheduler/k8s"
	_ "github.com/rancher/norman/clientbase"
	_ "github.com/rancher/norman/types"
)

const (
	// SchedName is the name of the kubernetes scheduler driver implementation
	SchedName = "rke"
	// SystemdSchedServiceName is the name of the system service resposible for scheduling
	SystemdSchedServiceName = "kubelet"
)

type Rancher struct {
	kube.K8s
	client *rancherclient.Client
}

func (k *Rancher) Init(schedOpts scheduler.InitOptions) error {
	var err error
	rancherClientOpts := rancherclientbase.ClientOpts{
		URL:      schedOpts.Endpoint,
		TokenKey: schedOpts.Token,
		Insecure: true,
	}
	k.client, err = rancherclient.NewClient(&rancherClientOpts)
	if err != nil {
		return fmt.Errorf("error getting rancher client, %v", err)
	}
	return nil
}

func (k *Rancher) SaveSchedulerLogsToFile(n node.Node, location string) error {
	driver, _ := node.Get(k.K8s.NodeDriverName)
	// requires 2>&1 since docker logs command send the logs to stdrr instead of sdout
	cmd := fmt.Sprintf("docker logs %s > %s/kubelet.log 2>&1", SystemdSchedServiceName, location)
	_, err := driver.RunCommand(n, cmd, node.ConnectionOpts{
		Timeout:         kube.DefaultTimeout,
		TimeBeforeRetry: kube.DefaultRetryInterval,
		Sudo:            true,
	})
	return err
}

func init() {
	k := &Rancher{}
	scheduler.Register(SchedName, k)
}
