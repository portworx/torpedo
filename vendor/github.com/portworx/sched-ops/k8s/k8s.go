package k8s

import (
	"fmt"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/portworx/sched-ops/task"
	apps_api "k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	storage_api "k8s.io/api/storage/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/apps/v1beta2"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	masterLabelKey        = "node-role.kubernetes.io/master"
	hostnameKey           = "kubernetes.io/hostname"
	pvcStorageClassKey    = "volume.beta.kubernetes.io/storage-class"
	labelUpdateMaxRetries = 5
)

// Ops is an interface to perform any kubernetes related operations
type Ops interface {
	NamespaceOps
	NodeOps
	ServiceOps
	StatefulSetOps
	DeploymentOps
	DaemonSetOps
	PodOps
	StorageClassOps
	PersistentVolumeClaimOps
}

// NamespaceOps is an interface to perform namespace operations
type NamespaceOps interface {
	// CreateNamespace creates a namespace with given name and metadata
	CreateNamespace(string, map[string]string) (*v1.Namespace, error)
	// DeleteNamespace deletes a namespace with given name
	DeleteNamespace(string) error
}

// NodeOps is an interface to perform k8s node operations
type NodeOps interface {
	// GetNodes talks to the k8s api server and gets the nodes in the cluster
	GetNodes() (*v1.NodeList, error)
	// GetNodeByName returns the k8s node given it's name
	GetNodeByName(string) (*v1.Node, error)
	// SearchNodeByAddresses searches corresponding k8s node match any of the given address
	SearchNodeByAddresses(addresses []string) (*v1.Node, error)
	// IsNodeReady checks if node with given name is ready. Returns nil is ready.
	IsNodeReady(string) error
	// IsNodeMaster returns true if given node is a kubernetes master node
	IsNodeMaster(v1.Node) bool
	// GetLabelsOnNode gets all the labels on the given node
	GetLabelsOnNode(string) (map[string]string, error)
	// AddLabelOnNode adds a label key=value on the given node
	AddLabelOnNode(string, string, string) error
	// RemoveLabelOnNode removes the label with key on given node
	RemoveLabelOnNode(string, string) error
	// WatchNode sets up a watcher that listens for the changes on Node.
	WatchNode(node *v1.Node, fn NodeWatchFunc) error
}

// ServiceOps is an interface to perform k8s service operations
type ServiceOps interface {
	// GetService gets the service by the name
	GetService(string, string) (*v1.Service, error)
	// CreateService creates the given service
	CreateService(*v1.Service) (*v1.Service, error)
	// DeleteService deletes the given service
	DeleteService(*v1.Service) error
	// ValidateDeletedService validates if given service is deleted
	ValidateDeletedService(string, string) error
	// DescribeService gets the service status
	DescribeService(string, string) (*v1.ServiceStatus, error)
}

// StatefulSetOps is an interface to perform k8s stateful set operations
type StatefulSetOps interface {
	// CreateStatefulSet creates the given statefulset
	CreateStatefulSet(*apps_api.StatefulSet) (*apps_api.StatefulSet, error)
	// DeleteStatefulSet deletes the given statefulset
	DeleteStatefulSet(*apps_api.StatefulSet) error
	// ValidateStatefulSet validates the given statefulset if it's running and healthy
	ValidateStatefulSet(*apps_api.StatefulSet) error
	// ValidateTerminatedStatefulSet validates if given deployment is terminated
	ValidateTerminatedStatefulSet(*apps_api.StatefulSet) error
	// GetStatefulSetPods returns pods for the given statefulset
	GetStatefulSetPods(*apps_api.StatefulSet) ([]v1.Pod, error)
	// DescribeStatefulSet gets status of the statefulset
	DescribeStatefulSet(string, string) (*apps_api.StatefulSetStatus, error)
}

// DeploymentOps is an interface to perform k8s deployment operations
type DeploymentOps interface {
	// CreateDeployment creates the given deployment
	CreateDeployment(*apps_api.Deployment) (*apps_api.Deployment, error)
	// DeleteDeployment deletes the given deployment
	DeleteDeployment(*apps_api.Deployment) error
	// ValidateDeployment validates the given deployment if it's running and healthy
	ValidateDeployment(*apps_api.Deployment) error
	// ValidateTerminatedDeployment validates if given deployment is terminated
	ValidateTerminatedDeployment(*apps_api.Deployment) error
	// GetDeploymentPods returns pods for the given deployment
	GetDeploymentPods(*apps_api.Deployment) ([]v1.Pod, error)
	// DescribeDeployment gets the deployment status
	DescribeDeployment(string, string) (*apps_api.DeploymentStatus, error)
}

// DaemonSetOps is an interface to perform k8s daemon set operations
type DaemonSetOps interface {
	// GetDaemonSet gets the the daemon set with given name
	GetDaemonSet(string, string) (*apps_api.DaemonSet, error)
	// UpdateDaemonSet updates the given daemon set
	UpdateDaemonSet(*apps_api.DaemonSet) error
}

// PodOps is an interface to perform k8s pod operations
type PodOps interface {
	// GetPods returns pods for the given namespace
	GetPods(string) (*v1.PodList, error)
	// GetPodsByOwner returns pods for the given owner and namespace
	GetPodsByOwner(string, string) ([]v1.Pod, error)
	// DeletePods deletes the given pods
	DeletePods([]v1.Pod) error
	// IsPodRunning checks if all containers in a pod are in running state
	IsPodRunning(v1.Pod) bool
}

// StorageClassOps is an interface to perform k8s storage class operations
type StorageClassOps interface {
	// CreateStorageClass creates the given storage class
	CreateStorageClass(*storage_api.StorageClass) (*storage_api.StorageClass, error)
	// DeleteStorageClass deletes the given storage class
	DeleteStorageClass(string) error
	// GetStorageClassParams returns the parameters of the given sc in the native map format
	GetStorageClassParams(*storage_api.StorageClass) (map[string]string, error)
	// ValidateStorageClass validates the given storage class
	ValidateStorageClass(string) (*storage_api.StorageClass, error)
}

// PersistentVolumeClaimOps is an interface to perform k8s PVC operations
type PersistentVolumeClaimOps interface {
	// CreatePersistentVolumeClaim creates the given persistent volume claim
	CreatePersistentVolumeClaim(*v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error)
	// DeletePersistentVolumeClaim deletes the given persistent volume claim
	DeletePersistentVolumeClaim(*v1.PersistentVolumeClaim) error
	// ValidatePersistentVolumeClaim validates the given pvc
	ValidatePersistentVolumeClaim(*v1.PersistentVolumeClaim) error
	// GetVolumeForPersistentVolumeClaim returns the back volume for the given PVC
	GetVolumeForPersistentVolumeClaim(*v1.PersistentVolumeClaim) (string, error)
	// GetPersistentVolumeClaimParams fetches custom parameters for the given PVC
	GetPersistentVolumeClaimParams(*v1.PersistentVolumeClaim) (map[string]string, error)
	// GetPersistentVolumeClaimStatus returns the status of the given pvc
	GetPersistentVolumeClaimStatus(*v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaimStatus, error)
}

var (
	instance Ops
	once     sync.Once
)

type k8sOps struct {
	client *kubernetes.Clientset
}

// Instance returns a singleton instance of k8sOps type
func Instance() Ops {
	once.Do(func() {
		instance = &k8sOps{}
	})
	return instance
}

// Initialize the k8s client if uninitialized
func (k *k8sOps) initK8sClient() error {
	if k.client == nil {
		k8sClient, err := getK8sClient()
		if err != nil {
			return err
		}

		// Quick validation if client connection works
		_, err = k8sClient.ServerVersion()
		if err != nil {
			return fmt.Errorf("failed to connect to k8s server: %s", err)
		}

		k.client = k8sClient
	}
	return nil
}

func (k *k8sOps) CreateNamespace(name string, metadata map[string]string) (*v1.Namespace, error) {
	if err := k.initK8sClient(); err != nil {
		return nil, err
	}

	return k.client.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:   name,
			Labels: metadata,
		},
	})
}

func (k *k8sOps) DeleteNamespace(name string) error {
	if err := k.initK8sClient(); err != nil {
		return err
	}

	return k.client.CoreV1().Namespaces().Delete(name, &meta_v1.DeleteOptions{})
}

func (k *k8sOps) GetNodes() (*v1.NodeList, error) {
	if err := k.initK8sClient(); err != nil {
		return nil, err
	}

	nodes, err := k.client.CoreV1().Nodes().List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

func (k *k8sOps) GetNodeByName(name string) (*v1.Node, error) {
	if err := k.initK8sClient(); err != nil {
		return nil, err
	}

	node, err := k.client.CoreV1().Nodes().Get(name, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return node, nil
}

func (k *k8sOps) IsNodeReady(name string) error {
	node, err := k.GetNodeByName(name)
	if err != nil {
		return err
	}

	for _, condition := range node.Status.Conditions {
		switch condition.Type {
		case v1.NodeConditionType(v1.NodeReady):
			if condition.Status != v1.ConditionStatus(v1.ConditionTrue) {
				return fmt.Errorf("node: %v is not ready as condition: %v (%v) is %v. Reason: %v",
					name, condition.Type, condition.Message, condition.Status, condition.Reason)
			}
		case v1.NodeConditionType(v1.NodeOutOfDisk),
			v1.NodeConditionType(v1.NodeMemoryPressure),
			v1.NodeConditionType(v1.NodeDiskPressure),
			v1.NodeConditionType(v1.NodeNetworkUnavailable):
			if condition.Status != v1.ConditionStatus(v1.ConditionFalse) {
				return fmt.Errorf("node: %v is not ready as condition: %v (%v) is %v. Reason: %v",
					name, condition.Type, condition.Message, condition.Status, condition.Reason)
			}
		}
	}

	return nil
}

func (k *k8sOps) IsNodeMaster(node v1.Node) bool {
	_, ok := node.Labels[masterLabelKey]
	return ok
}

func (k *k8sOps) GetLabelsOnNode(name string) (map[string]string, error) {
	node, err := k.GetNodeByName(name)
	if err != nil {
		return nil, err
	}

	return node.Labels, nil
}

// SearchNodeByAddresses searches the node based on the IP addresses, then it falls back to a
// search by hostname, and finally by the labels
func (k *k8sOps) SearchNodeByAddresses(addresses []string) (*v1.Node, error) {
	nodes, err := k.GetNodes()
	if err != nil {
		return nil, err
	}

	// sweep #1 - locating based on IP address
	for _, node := range nodes.Items {
		for _, addr := range node.Status.Addresses {
			switch addr.Type {
			case v1.NodeExternalIP:
				fallthrough
			case v1.NodeInternalIP:
				for _, ip := range addresses {
					if addr.Address == ip {
						return &node, nil
					}
				}
			}
		}
	}

	// sweep #2 - locating based on Hostname
	for _, node := range nodes.Items {
		for _, addr := range node.Status.Addresses {
			switch addr.Type {
			case v1.NodeHostName:
				for _, ip := range addresses {
					if addr.Address == ip {
						return &node, nil
					}
				}
			}
		}
	}

	// sweep #3 - locating based on labels
	for _, node := range nodes.Items {
		if hn, has := node.GetLabels()[hostnameKey]; has {
			for _, ip := range addresses {
				if hn == ip {
					return &node, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("failed to find k8s node for given addresses: %v", addresses)
}

func (k *k8sOps) AddLabelOnNode(name, key, value string) error {
	var err error
	if err := k.initK8sClient(); err != nil {
		return err
	}

	retryCnt := 0
	for retryCnt < labelUpdateMaxRetries {
		retryCnt++

		node, err := k.client.CoreV1().Nodes().Get(name, meta_v1.GetOptions{})
		if err != nil {
			return err
		}

		if val, present := node.Labels[key]; present && val == value {
			return nil
		}

		node.Labels[key] = value
		if _, err = k.client.CoreV1().Nodes().Update(node); err == nil {
			return nil
		}
	}

	return err
}

func (k *k8sOps) RemoveLabelOnNode(name, key string) error {
	var err error
	if err := k.initK8sClient(); err != nil {
		return err
	}

	retryCnt := 0
	for retryCnt < labelUpdateMaxRetries {
		retryCnt++

		node, err := k.client.CoreV1().Nodes().Get(name, meta_v1.GetOptions{})
		if err != nil {
			return err
		}

		if _, present := node.Labels[key]; present {
			delete(node.Labels, key)
			if _, err = k.client.CoreV1().Nodes().Update(node); err == nil {
				return nil
			}
		}
	}

	return err
}

// NodeWatchFunc is a callback provided to the WatchNode function
// which is invoked when the v1.Node object is changed.
type NodeWatchFunc func(node *v1.Node) error

func (k *k8sOps) WatchNode(node *v1.Node, watchNodeFn NodeWatchFunc) error {
	if node == nil {
		return fmt.Errorf("no node given to watch")
	}

	if err := k.initK8sClient(); err != nil {
		return err
	}

	// let's use internal FieldsSelector, instead of LabelsSelector (labels are volatile)
	listOptions := meta_v1.SingleObject(node.ObjectMeta)
	watchInterface, err := k.client.Core().Nodes().Watch(listOptions)
	if err != nil {
		return err
	}

	// fire off watch function
	go func() {
		for {
			select {
			case event, more := <-watchInterface.ResultChan():
				if !more {
					// log.Warn("Kubernetes node watch channel closed")
					return
				}
				if k8sNode, ok := event.Object.(*v1.Node); ok {
					// CHECKME: handle errors?
					watchNodeFn(k8sNode)
				}
			}
		}
	}()
	return nil
}

// Service APIs - BEGIN

func (k *k8sOps) CreateService(service *v1.Service) (*v1.Service, error) {
	if err := k.initK8sClient(); err != nil {
		return nil, err
	}

	ns := service.Namespace
	if len(ns) == 0 {
		ns = v1.NamespaceDefault
	}

	return k.client.CoreV1().Services(ns).Create(service)
}

func (k *k8sOps) DeleteService(service *v1.Service) error {
	if err := k.initK8sClient(); err != nil {
		return err
	}

	policy := meta_v1.DeletePropagationForeground
	return k.client.CoreV1().Services(service.Namespace).Delete(service.Name, &meta_v1.DeleteOptions{
		PropagationPolicy: &policy,
	})
}

func (k *k8sOps) GetService(svcName string, svcNS string) (*v1.Service, error) {
	if err := k.initK8sClient(); err != nil {
		return nil, err
	}

	if svcName == "" {
		return nil, fmt.Errorf("cannot return service obj without service name")
	}
	svc, err := k.client.CoreV1().Services(svcNS).Get(svcName, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return svc, nil

}

func (k *k8sOps) DescribeService(svcName string, svcNamespace string) (*v1.ServiceStatus, error) {
	svc, err := k.GetService(svcName, svcNamespace)
	if err != nil {
		return nil, err
	}
	return &svc.Status, err
}

func (k *k8sOps) ValidateDeletedService(svcName string, svcNS string) error {
	if err := k.initK8sClient(); err != nil {
		return err
	}

	if svcName == "" {
		return fmt.Errorf("cannot validate service without service name")
	}

	_, err := k.client.CoreV1().Services(svcNS).Get(svcName, meta_v1.GetOptions{})
	if err != nil {
		if matched, _ := regexp.MatchString(".+ not found", err.Error()); matched {
			return nil
		}
		return err
	}

	return nil
}

// Service APIs - END

// Deployment APIs - BEGIN

func (k *k8sOps) CreateDeployment(deployment *apps_api.Deployment) (*apps_api.Deployment, error) {
	if err := k.initK8sClient(); err != nil {
		return nil, err
	}

	ns := deployment.Namespace
	if len(ns) == 0 {
		ns = v1.NamespaceDefault
	}

	return k.appsClient().Deployments(ns).Create(deployment)
}

func (k *k8sOps) DeleteDeployment(deployment *apps_api.Deployment) error {
	if err := k.initK8sClient(); err != nil {
		return err
	}

	policy := meta_v1.DeletePropagationForeground
	return k.appsClient().Deployments(deployment.Namespace).Delete(deployment.Name, &meta_v1.DeleteOptions{
		PropagationPolicy: &policy,
	})
}

func (k *k8sOps) DescribeDeployment(depName string, depNamespace string) (*apps_api.DeploymentStatus, error) {
	if err := k.initK8sClient(); err != nil {
		return nil, err
	}
	dep, err := k.appsClient().Deployments(depNamespace).Get(depName, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return &dep.Status, err
}

func (k *k8sOps) ValidateDeployment(deployment *apps_api.Deployment) error {
	t := func() (interface{}, error) {
		if err := k.initK8sClient(); err != nil {
			return "", err
		}

		dep, err := k.appsClient().Deployments(deployment.Namespace).Get(deployment.Name, meta_v1.GetOptions{})
		if err != nil {
			return "", err
		}

		requiredReplicas := *dep.Spec.Replicas

		if requiredReplicas != 1 {
			shared := false
			foundPVC := false
			for _, vol := range dep.Spec.Template.Spec.Volumes {
				if vol.PersistentVolumeClaim != nil {
					foundPVC = true

					claim, err := k.client.CoreV1().
						PersistentVolumeClaims(dep.Namespace).
						Get(vol.PersistentVolumeClaim.ClaimName, meta_v1.GetOptions{})
					if err != nil {
						return "", err
					}

					if k.isPVCShared(claim) {
						shared = true
						break
					}
				}
			}

			if foundPVC && !shared {
				requiredReplicas = 1
			}
		}

		if requiredReplicas > dep.Status.AvailableReplicas {
			return "", &ErrAppNotReady{
				ID: dep.Name,
				Cause: fmt.Sprintf("Expected replicas: %v Available replicas: %v",
					requiredReplicas, dep.Status.AvailableReplicas),
			}
		}

		if requiredReplicas > dep.Status.ReadyReplicas {
			return "", &ErrAppNotReady{
				ID: dep.Name,
				Cause: fmt.Sprintf("Expected replicas: %v Ready replicas: %v",
					requiredReplicas, dep.Status.ReadyReplicas),
			}
		}

		pods, err := k.GetDeploymentPods(deployment)
		if err != nil || pods == nil {
			return "", &ErrAppNotReady{
				ID:    dep.Name,
				Cause: fmt.Sprintf("Failed to get pods for deployment. Err: %v", err),
			}
		}

		if len(pods) == 0 {
			return "", &ErrAppNotReady{
				ID:    dep.Name,
				Cause: "Application has 0 pods",
			}
		}

		// look for "requiredReplicas" number of pods in running state
		var notRunningPods []string
		var runningCount int32
		for _, pod := range pods {
			if !k.IsPodRunning(pod) {
				notRunningPods = append(notRunningPods, pod.Name)
			} else {
				runningCount++
			}
		}

		if runningCount >= requiredReplicas {
			return "", nil
		}

		return "", &ErrAppNotReady{
			ID:    dep.Name,
			Cause: fmt.Sprintf("pod(s): %#v not yet ready", notRunningPods),
		}
	}

	if _, err := task.DoRetryWithTimeout(t, 10*time.Minute, 10*time.Second); err != nil {
		return err
	}
	return nil
}

func (k *k8sOps) ValidateTerminatedDeployment(deployment *apps_api.Deployment) error {
	t := func() (interface{}, error) {
		if err := k.initK8sClient(); err != nil {
			return "", err
		}

		dep, err := k.appsClient().Deployments(deployment.Namespace).Get(deployment.Name, meta_v1.GetOptions{})
		if err != nil {
			if matched, _ := regexp.MatchString(".+ not found", err.Error()); matched {
				return "", nil
			}
			return "", err
		}

		pods, err := k.GetDeploymentPods(deployment)
		if err != nil {
			return "", &ErrAppNotTerminated{
				ID:    dep.Name,
				Cause: fmt.Sprintf("Failed to get pods for deployment. Err: %v", err),
			}
		}

		if pods != nil && len(pods) > 0 {
			return "", &ErrAppNotTerminated{
				ID:    dep.Name,
				Cause: fmt.Sprintf("pods: %#v is still present", pods),
			}
		}

		return "", nil
	}

	if _, err := task.DoRetryWithTimeout(t, 10*time.Minute, 10*time.Second); err != nil {
		return err
	}
	return nil
}

func (k *k8sOps) GetDeploymentPods(deployment *apps_api.Deployment) ([]v1.Pod, error) {
	if err := k.initK8sClient(); err != nil {
		return nil, err
	}

	rSets, err := k.appsClient().ReplicaSets(deployment.Namespace).List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, rSet := range rSets.Items {
		for _, owner := range rSet.OwnerReferences {
			if owner.Name == deployment.Name {
				return k.GetPodsByOwner(rSet.Name, rSet.Namespace)
			}
		}
	}

	return nil, nil
}

// Deployment APIs - END

// DaemonSet APIs - BEGIN

func (k *k8sOps) GetDaemonSet(name, namespace string) (*apps_api.DaemonSet, error) {
	if err := k.initK8sClient(); err != nil {
		return nil, err
	}

	if len(namespace) == 0 {
		namespace = v1.NamespaceDefault
	}

	ds, err := k.appsClient().DaemonSets(namespace).Get(name, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return ds, nil
}

func (k *k8sOps) UpdateDaemonSet(ds *apps_api.DaemonSet) error {
	if err := k.initK8sClient(); err != nil {
		return err
	}

	if _, err := k.appsClient().DaemonSets(ds.Namespace).Update(ds); err != nil {
		return err
	}
	return nil
}

// DaemonSet APIs - END

// StatefulSet APIs - BEGIN

func (k *k8sOps) CreateStatefulSet(statefulset *apps_api.StatefulSet) (*apps_api.StatefulSet, error) {
	if err := k.initK8sClient(); err != nil {
		return nil, err
	}

	ns := statefulset.Namespace
	if len(ns) == 0 {
		ns = v1.NamespaceDefault
	}

	return k.appsClient().StatefulSets(ns).Create(statefulset)
}

func (k *k8sOps) DeleteStatefulSet(statefulset *apps_api.StatefulSet) error {
	if err := k.initK8sClient(); err != nil {
		return err
	}

	policy := meta_v1.DeletePropagationForeground
	return k.appsClient().StatefulSets(statefulset.Namespace).Delete(statefulset.Name, &meta_v1.DeleteOptions{
		PropagationPolicy: &policy,
	})
}

func (k *k8sOps) DescribeStatefulSet(ssetName string, ssetNamespace string) (*apps_api.StatefulSetStatus, error) {
	if err := k.initK8sClient(); err != nil {
		return nil, err
	}
	sset, err := k.appsClient().StatefulSets(ssetNamespace).Get(ssetName, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return &sset.Status, err
}

func (k *k8sOps) ValidateStatefulSet(statefulset *apps_api.StatefulSet) error {
	t := func() (interface{}, error) {
		if err := k.initK8sClient(); err != nil {
			return "", err
		}
		sset, err := k.appsClient().StatefulSets(statefulset.Namespace).Get(statefulset.Name, meta_v1.GetOptions{})
		if err != nil {
			return "", err
		}

		if *sset.Spec.Replicas != sset.Status.Replicas { // Not sure if this is even needed but for now let's have one check before
			//readiness check
			return "", &ErrAppNotReady{
				ID:    sset.Name,
				Cause: fmt.Sprintf("Expected replicas: %v Observed replicas: %v", *sset.Spec.Replicas, sset.Status.Replicas),
			}
		}

		if *sset.Spec.Replicas != sset.Status.ReadyReplicas {
			return "", &ErrAppNotReady{
				ID:    sset.Name,
				Cause: fmt.Sprintf("Expected replicas: %v Ready replicas: %v", *sset.Spec.Replicas, sset.Status.ReadyReplicas),
			}
		}

		pods, err := k.GetStatefulSetPods(statefulset)
		if err != nil || pods == nil {
			return "", &ErrAppNotReady{
				ID:    sset.Name,
				Cause: fmt.Sprintf("Failed to get pods for statefulset. Err: %v", err),
			}
		}

		for _, pod := range pods {
			if !k.IsPodRunning(pod) {
				return "", &ErrAppNotReady{
					ID:    sset.Name,
					Cause: fmt.Sprintf("pod: %v is not yet ready", pod.Name),
				}
			}
		}

		return "", nil
	}

	if _, err := task.DoRetryWithTimeout(t, 10*time.Minute, 10*time.Second); err != nil {
		return err
	}
	return nil
}

func (k *k8sOps) GetStatefulSetPods(statefulset *apps_api.StatefulSet) ([]v1.Pod, error) {
	return k.GetPodsByOwner(statefulset.Name, statefulset.Namespace)
}

func (k *k8sOps) ValidateTerminatedStatefulSet(statefulset *apps_api.StatefulSet) error {
	t := func() (interface{}, error) {
		if err := k.initK8sClient(); err != nil {
			return "", err
		}

		sset, err := k.appsClient().StatefulSets(statefulset.Namespace).Get(statefulset.Name, meta_v1.GetOptions{})
		if err != nil {
			if matched, _ := regexp.MatchString(".+ not found", err.Error()); matched {
				return "", nil
			}

			return "", err
		}

		pods, err := k.GetStatefulSetPods(statefulset)
		if err != nil {
			return "", &ErrAppNotTerminated{
				ID:    sset.Name,
				Cause: fmt.Sprintf("Failed to get pods for statefulset. Err: %v", err),
			}
		}

		if pods != nil && len(pods) > 0 {
			return "", &ErrAppNotTerminated{
				ID:    sset.Name,
				Cause: fmt.Sprintf("pods: %#v is still present", pods),
			}
		}

		return "", nil
	}

	if _, err := task.DoRetryWithTimeout(t, 10*time.Minute, 10*time.Second); err != nil {
		return err
	}
	return nil
}

// StatefulSet APIs - END

func (k *k8sOps) DeletePods(pods []v1.Pod) error {
	if err := k.initK8sClient(); err != nil {
		return err
	}

	var gracePeriod int64
	gracePeriod = 0

	for _, pod := range pods {
		if err := k.client.CoreV1().Pods(pod.Namespace).Delete(pod.Name, &meta_v1.DeleteOptions{
			GracePeriodSeconds: &gracePeriod,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (k *k8sOps) GetPods(namespace string) (*v1.PodList, error) {
	if err := k.initK8sClient(); err != nil {
		return nil, err
	}

	return k.client.CoreV1().Pods(namespace).List(meta_v1.ListOptions{})
}

func (k *k8sOps) GetPodsByOwner(ownerName string, namespace string) ([]v1.Pod, error) {
	pods, err := k.GetPods(namespace)
	if err != nil {
		return nil, err
	}

	var result []v1.Pod
	for _, pod := range pods.Items {
		for _, owner := range pod.OwnerReferences {
			if owner.Name == ownerName {
				result = append(result, pod)
			}
		}
	}

	return result, nil
}

func (k *k8sOps) IsPodRunning(pod v1.Pod) bool {
	// If init containers are running, return false since the actual container would not have started yet
	for _, c := range pod.Status.InitContainerStatuses {
		if c.State.Running != nil {
			return false
		}
	}

	for _, c := range pod.Status.ContainerStatuses {
		if c.State.Running == nil {
			return false
		}
	}

	return true
}

// StorageClass APIs - BEGIN

func (k *k8sOps) CreateStorageClass(sc *storage_api.StorageClass) (*storage_api.StorageClass, error) {
	if err := k.initK8sClient(); err != nil {
		return nil, err
	}

	return k.client.StorageV1().StorageClasses().Create(sc)
}

func (k *k8sOps) DeleteStorageClass(name string) error {
	if err := k.initK8sClient(); err != nil {
		return err
	}

	return k.client.StorageV1().StorageClasses().Delete(name, &meta_v1.DeleteOptions{})
}

func (k *k8sOps) GetStorageClassParams(sc *storage_api.StorageClass) (map[string]string, error) {
	if err := k.initK8sClient(); err != nil {
		return nil, err
	}

	sc, err := k.client.StorageV1().StorageClasses().Get(sc.Name, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return sc.Parameters, nil
}

func (k *k8sOps) ValidateStorageClass(name string) (*storage_api.StorageClass, error) {
	if err := k.initK8sClient(); err != nil {
		return nil, err
	}

	sc, err := k.client.StorageV1().StorageClasses().Get(name, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return sc, nil
}

// StorageClass APIs - END

// PVC APIs - BEGIN

func (k *k8sOps) CreatePersistentVolumeClaim(pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	if err := k.initK8sClient(); err != nil {
		return nil, err
	}

	ns := pvc.Namespace
	if len(ns) == 0 {
		ns = v1.NamespaceDefault
	}

	return k.client.CoreV1().PersistentVolumeClaims(ns).Create(pvc)
}

func (k *k8sOps) DeletePersistentVolumeClaim(pvc *v1.PersistentVolumeClaim) error {
	if err := k.initK8sClient(); err != nil {
		return err
	}

	return k.client.CoreV1().PersistentVolumeClaims(pvc.Namespace).Delete(pvc.Name, &meta_v1.DeleteOptions{})
}

func (k *k8sOps) ValidatePersistentVolumeClaim(pvc *v1.PersistentVolumeClaim) error {
	t := func() (interface{}, error) {
		if err := k.initK8sClient(); err != nil {
			return "", err
		}

		result, err := k.client.CoreV1().
			PersistentVolumeClaims(pvc.Namespace).
			Get(pvc.Name, meta_v1.GetOptions{})
		if err != nil {
			return "", err
		}

		if result.Status.Phase == v1.ClaimBound {
			return "", nil
		}

		return "", &ErrPVCNotReady{
			ID:    result.Name,
			Cause: fmt.Sprintf("PVC expected status: %v PVC actual status: %v", v1.ClaimBound, result.Status.Phase),
		}
	}

	if _, err := task.DoRetryWithTimeout(t, 5*time.Minute, 10*time.Second); err != nil {
		return err
	}
	return nil
}

func (k *k8sOps) GetVolumeForPersistentVolumeClaim(pvc *v1.PersistentVolumeClaim) (string, error) {
	if err := k.initK8sClient(); err != nil {
		return "", err
	}

	result, err := k.client.CoreV1().
		PersistentVolumeClaims(pvc.Namespace).
		Get(pvc.Name, meta_v1.GetOptions{})
	if err != nil {
		return "", err
	}

	return result.Spec.VolumeName, nil
}

func (k *k8sOps) GetPersistentVolumeClaimStatus(pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaimStatus, error) {
	if err := k.initK8sClient(); err != nil {
		return nil, err
	}

	result, err := k.client.CoreV1().PersistentVolumeClaims(pvc.Namespace).Get(pvc.Name, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &result.Status, nil
}

func (k *k8sOps) GetPersistentVolumeClaimParams(pvc *v1.PersistentVolumeClaim) (map[string]string, error) {
	if err := k.initK8sClient(); err != nil {
		return nil, err
	}

	params := make(map[string]string)

	result, err := k.client.CoreV1().PersistentVolumeClaims(pvc.Namespace).Get(pvc.Name, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	capacity, ok := result.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	if !ok {
		return nil, fmt.Errorf("failed to get storage resource for pvc: %v", result.Name)
	}

	// We explicitly send the unit with so the client can compare it with correct units
	requestGB := uint64(roundUpSize(capacity.Value(), 1024*1024*1024))
	params["size"] = fmt.Sprintf("%dG", requestGB)

	scName, ok := result.Annotations[pvcStorageClassKey]
	if !ok {
		return nil, fmt.Errorf("failed to get storage class for pvc: %v", result.Name)
	}

	sc, err := k.client.StorageV1().StorageClasses().Get(scName, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	for key, value := range sc.Parameters {
		params[key] = value
	}

	return params, nil
}

// isPVCShared returns true if the PersistentVolumeClaim has been configured for use by multiple clients
func (k *k8sOps) isPVCShared(pvc *v1.PersistentVolumeClaim) bool {
	for _, mode := range pvc.Spec.AccessModes {
		if mode == v1.PersistentVolumeAccessMode(v1.ReadOnlyMany) ||
			mode == v1.PersistentVolumeAccessMode(v1.ReadWriteMany) {
			return true
		}
	}

	return false
}

// PVCs APIs - END

func (k *k8sOps) appsClient() v1beta2.AppsV1beta2Interface {
	return k.client.AppsV1beta2()
}

// getK8sClient instantiates a k8s client
func getK8sClient() (*kubernetes.Clientset, error) {
	var k8sClient *kubernetes.Clientset
	var err error

	kubeconfig := os.Getenv("KUBECONFIG")
	if len(kubeconfig) > 0 {
		k8sClient, err = loadClientFromKubeconfig(kubeconfig)
	} else {
		k8sClient, err = loadClientFromServiceAccount()
	}

	if err != nil {
		return nil, err
	}

	if k8sClient == nil {
		return nil, ErrK8SApiAccountNotSet
	}

	return k8sClient, nil
}

// loadClientFromServiceAccount loads a k8s client from a ServiceAccount specified in the pod running px
func loadClientFromServiceAccount() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func loadClientFromKubeconfig(kubeconfig string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func roundUpSize(volumeSizeBytes int64, allocationUnitBytes int64) int64 {
	return (volumeSizeBytes + allocationUnitBytes - 1) / allocationUnitBytes
}
