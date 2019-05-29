package openshift

import (
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	kube "github.com/portworx/torpedo/drivers/scheduler/k8s"
	"regexp"
	"time"
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

const (
	statefulSetValidateTimeout   = 20 * time.Minute
	k8sNodeReadyTimeout          = 5 * time.Minute
	volDirCleanupTimeout         = 5 * time.Minute
	k8sObjectCreateTimeout       = 2 * time.Minute
	k8sDestroyTimeout            = 2 * time.Minute
	findFilesOnWorkerTimeout     = 1 * time.Minute
	deleteTasksWaitTimeout       = 3 * time.Minute
	defaultRetryInterval         = 10 * time.Second
	defaultTimeout               = 2 * time.Minute
	resizeSupportedAnnotationKey = "torpedo/resize-supported"
)

const (
	portworxStorage = "portworx"
	csiStorage      = "csi"
)

var provisioners = map[string]string{
	portworxStorage: "kubernetes.io/portworx-volume",
	csiStorage:      "com.openstorage.pxd",
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
			Timeout:         findFilesOnWorkerTimeout,
			TimeBeforeRetry: defaultRetryInterval,
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
			Timeout:         defaultTimeout,
			TimeBeforeRetry: defaultRetryInterval,
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
