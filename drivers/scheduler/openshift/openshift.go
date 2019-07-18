package openshift

import (
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	kube "github.com/portworx/torpedo/drivers/scheduler/k8s"
	"regexp"
)

const (
	// SchedName is the name of the kubernetes scheduler driver implementation
	SchedName = "openshift"
	// SnapshotParent is the parameter key for the parent of a snapshot
	SnapshotParent = "snapshot_parent"
	k8sPodsRootDir = "/var/lib/kubelet/pods"
	// DeploymentSuffix is the suffix for deployment names stored as keys in maps
	DeploymentSuffix = "-dep"
	// StatefulSetSuffix is the suffix for statefulset names stored as keys in maps
	StatefulSetSuffix = "-ss"
	// SystemdSchedServiceName is the name of the system service resposible for scheduling
	// TODO Change this when running on openshift for the proper service name
	SystemdSchedServiceName = "atomic-openshift-node"
)



var provisioners = map[string]string{
	kube.PortworxStorage: "kubernetes.io/portworx-volume",
	kube.CsiStorage:      "com.openstorage.pxd",
}

var (
	namespaceRegex = regexp.MustCompile("{{NAMESPACE}}")
)

type openshift struct {
	kube.K8s
}

func (k *openshift) StopSchedOnNode(n node.Node) error {
	driver, _ := node.Get(k.K8s.NodeDriverName)
	systemOpts := node.SystemctlOpts{
		ConnectionOpts: node.ConnectionOpts{
			Timeout:         kube.FindFilesOnWorkerTimeout,
			TimeBeforeRetry: kube.DefaultRetryInterval,
		},
		Action: "stop",
	}
	err := driver.Systemctl(n, SystemdSchedServiceName, systemOpts)
	if err != nil {
		return &scheduler.ErrFailedToStopSchedOnNode{
			Node:          n,
			SystemService: SystemdSchedServiceName,
			Cause:         err.Error(),
		}
	}
	return nil
}

func (k *openshift) StartSchedOnNode(n node.Node) error {
	driver, _ := node.Get(k.K8s.NodeDriverName)
	systemOpts := node.SystemctlOpts{
		ConnectionOpts: node.ConnectionOpts{
			Timeout:         kube.DefaultTimeout,
			TimeBeforeRetry: kube.DefaultRetryInterval,
		},
		Action: "start",
	}
	err := driver.Systemctl(n, SystemdSchedServiceName, systemOpts)
	if err != nil {
		return &scheduler.ErrFailedToStartSchedOnNode{
			Node:          n,
			SystemService: SystemdSchedServiceName,
			Cause:         err.Error(),
		}
	}
	return nil
}

func init() {
	k := &openshift{}
	scheduler.Register(SchedName, k)
}
