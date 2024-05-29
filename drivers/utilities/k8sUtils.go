package utilities

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/osutils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	timeInterval = 10 * time.Second
	timeOut      = 30 * time.Minute
)

var (
	k8sCore = core.Instance()
)

var (
	Instance *Torpedo
	k8sCores = core.Instance()
)

type Torpedo struct {
	InstanceID string
	S          scheduler.Driver
	V          volume.Driver
	N          node.Driver
}

// Inst returns the Torpedo instances
func Inst() *Torpedo {
	return Instance
}

//// Check the DS related PV usage and resize in case of 90% full
//func CheckStorageFullCondition(namespace string, deploymentName string, thresholdPercentage float64) error {
//	log.Infof("Check PVC Usage")
//	f := func() (interface{}, bool, error) {
//		initialCapacity, _ := GetVolumeCapacityInGB(namespace, deploymentName)
//		floatCapacity := float64(initialCapacity)
//		consumedCapacity, err := GetVolumeUsage(namespace, deploymentName)
//		log.FailOnError(err, "unable to calculate vol usage")
//		threshold := thresholdPercentage * (floatCapacity / 100)
//		log.InfoD("threshold value calculated is- [%v]", threshold)
//		if consumedCapacity >= threshold {
//			log.Infof("The PVC capacity was %v , the consumed PVC in floating point value is- %v", initialCapacity, consumedCapacity)
//			return nil, false, nil
//		}
//		return nil, true, fmt.Errorf("threshold not achieved for the PVC, ")
//	}
//	_, err := task.DoRetryWithTimeout(f, 35*time.Minute, 20*time.Second)
//	return err
//}

func GetDbMasterNode(namespace string, dsName string, deploymentName string, kubeConfig string) (string, bool) {
	var command, dbMaster string
	log.Debugf("DS Name - [%s]", dsName)
	log.Debugf("KubeConfig - [%s]", kubeConfig)
	log.Debugf("Deployment Name - [%s]", deploymentName)
	log.Debugf("Namespace - [%s]", namespace)
	switch strings.ToLower(dsName) {
	case "postgresql":
		log.Infof("DSName - [%s]", dsName)
		command = fmt.Sprintf("patronictl list | grep -i leader | awk '{print $2}'")
		dbMaster, _ = ExecuteCommandInStatefulSetPod(deploymentName, namespace, command, kubeConfig)
	// TODO : Add this support for MySql

	default:
		return "", false
	}
	return dbMaster, true
}

// ExecuteCommandInStatefulSetPod executes the provided command inside a pod within the specified StatefulSet.
func ExecuteCommandInStatefulSetPod(statefulsetName, namespace, command string, kubeConfig string) (string, error) {
	podName, err := GetAnyPodName(statefulsetName, namespace)
	if err != nil {
		return "", err
	}

	return ExecCommandInPod(podName, namespace, command, kubeConfig)
}

func GetAnyPodName(statefulName, namespace string) (string, error) {
	rand.Seed(time.Now().UnixNano())
	inst := apps.Instance()
	sts, err := inst.GetStatefulSet(statefulName, namespace)
	if err != nil {
		return "", err
	}
	podList, err := inst.GetStatefulSetPods(sts)

	randomIndex := rand.Intn(len(podList))
	randomElement := podList[randomIndex]
	return randomElement.GetName(), nil
}

func ExecCommandInPod(podName, namespace, command string, kubeConfig string) (string, error) {
	cmd := fmt.Sprintf("kubectl --kubeconfig %s -n %s exec -it %s -- %s", kubeConfig, namespace, podName, command)
	log.Infof("Command: ", cmd)
	output, _, err := osutils.ExecShell(cmd)
	if err != nil {
		return "", err
	}
	log.Infof("Terminal output: %v", output)

	return string(output), nil
}

// DeleteK8sPods deletes the pods in given namespace
func DeleteK8sPods(pod string, namespace string, kubeConfigPath string) error {
	cmd := fmt.Sprintf("kubectl --kubeconfig %v -n %v delete pod %v", kubeConfigPath, namespace, pod)
	log.Infof("Command: ", cmd)
	output, _, err := osutils.ExecShell(cmd)
	if err != nil {
		return err
	}
	log.Infof("Terminal output: %v", output)
	return nil
}

func DeleteK8sPod(podName, namespace string, force bool) error {
	log.Debug("DeleteK8sMethod is called ")
	err := k8sCores.DeletePod(podName, namespace, force)
	if err != nil {
		return err
	}
	return nil
}

// GetPods returns the list of pods in namespace
func GetPods(namespace string) (*corev1.PodList, error) {
	podList, err := k8sCores.GetPods(namespace, nil)
	if err != nil {
		return nil, err
	}
	return podList, err
}

// KillPodsInNamespace Kill All pods matching podName string in a given namespace
func KillPodsInNamespace(ns string, podName string) error {
	var Pods []corev1.Pod

	podList, err := GetPods(ns)
	if err != nil {
		return err
	}

	for _, pod := range podList.Items {
		if strings.Contains(pod.Name, podName) {
			log.Infof("Pod Name is : %v", pod.Name)
			Pods = append(Pods, pod)
		}
	}

	for _, pod := range Pods {
		log.InfoD("Deleting Pod: %s", pod.Name)
		err = DeleteK8sPods(pod.Name, ns, "")
		if err != nil {
			return err
		}
		log.InfoD("Successfully Killed Pod: %v", pod.Name)
	}
	return err
}

func GetVolumeNodesOnWhichPxIsRunning() []node.Node {
	var (
		nodesToStopPx []node.Node
		stopPxNode    []node.Node
	)
	// Initialise the slices
	nodesToStopPx = make([]node.Node, 0)
	stopPxNode = make([]node.Node, 0)

	stopPxNode = node.GetStorageNodes()
	log.InfoD("PX the node with vol running found is-  %v ", stopPxNode)
	if len(stopPxNode) > 0 {
		nodesToStopPx = append(nodesToStopPx, stopPxNode[0])
	}
	return nodesToStopPx
}

// StopPxOnReplicaVolumeNode is used to STOP PX on the given list of nodes
func StopPxOnReplicaVolumeNode() error {
	nodesToStopPx := GetVolumeNodesOnWhichPxIsRunning()
	err := Inst().V.StopDriver(nodesToStopPx, true, nil)
	if err != nil {
		log.FailOnError(err, "Error while trying to STOP PX on the volNode- [%v]", nodesToStopPx)
	}
	log.InfoD("PX stopped successfully on node %v", nodesToStopPx)
	return nil
}

// DeleteNamespace will delete the namespace from the cluster
func DeleteNamespace(namespace string) error {
	k8sCore := core.Instance()
	err := k8sCore.DeleteNamespace(namespace)
	if err != nil {
		return fmt.Errorf("Error while deleting namespace [%s]", err.Error())
	}
	return nil
}

// DeleteDeploymentPods deletes the given pods
func DeletePods(podList []corev1.Pod) error {
	err := k8sCores.DeletePods(podList, true)
	if err != nil {
		return err
	}
	return nil
}

// ValidatePods returns err if pods are not up
func ValidatePods(namespace string, podName string) error {
	var newPods []corev1.Pod
	newPodList, err := GetPods(namespace)
	if err != nil {
		return err
	}

	if podName != "" {
		for _, pod := range newPodList.Items {
			if strings.Contains(pod.Name, podName) {
				log.Infof("%v", pod.Name)
				newPods = append(newPods, pod)
			}
		}
	} else {
		//reinitializing the pods
		newPods = append(newPods, newPodList.Items...)
	}

	//validate deployment pods are up and running
	for _, pod := range newPods {
		log.Infof("pds system pod name %v", pod.Name)
		err = k8sCores.ValidatePod(&pod, timeOut, timeInterval)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetDeploymentPods returns the pods list for a given deployment and given namespace
func GetDeploymentPods(deploymentName string, namespace string) ([]corev1.Pod, error) {

	depPods := make([]corev1.Pod, 0)
	pods, err := GetPods(namespace)
	if err != nil {
		log.FailOnError(err, "failed to get deployment pods")
	}
	for _, pod := range pods.Items {
		log.Infof("%v", pod.Name)
		if strings.Contains(pod.Name, deploymentName) {
			depPods = append(depPods, pod)
		}
	}
	return depPods, nil
}

// GetPodAge gets the pod age of pods of a given deployment and namespace
func GetPodAge(deploymentName string, namespace string) (float64, error) {
	var podAge time.Duration
	pods, err := GetDeploymentPods(deploymentName, namespace)
	log.FailOnError(err, "Unable to fetch deployment pods")
	for _, pod := range pods {
		currentTime := metav1.Now()
		podCreationTime := pod.GetCreationTimestamp().Time
		podAge = currentTime.Time.Sub(podCreationTime)
	}
	podAgeInt := podAge.Minutes()
	return podAgeInt, nil
}

// Check if a deployment specific PV and associated PVC is still present. If yes then delete both of them
func DeletePvandPVCs(resourceName string, delPod bool) error {
	log.Debugf("Starting to delete the PV and PVCs for resource %v\n", resourceName)
	var claimName string
	pv_list, err := k8sCore.GetPersistentVolumes()
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("persistant volumes Not Found due to : %v", err)
		}
		return err
	}
	for _, vol := range pv_list.Items {
		if vol.Spec.ClaimRef != nil {
			claimName = vol.Spec.ClaimRef.Name
		} else {
			log.Infof("No PVC bounded to the PV - %v", vol.Name)
			continue
		}
		flag := strings.Contains(claimName, resourceName)
		if flag {
			err := CheckAndDeleteIndependentPV(vol.Name, delPod)
			if err != nil {
				return fmt.Errorf("unable to delete the associated PV and PVCS due to : %v .Please check manually", err)
			}
			log.Debugf("The PV : %v and its associated PVC : %v is deleted !", vol.Name, claimName)
		}
	}
	return nil
}

// Check if PV and associated PVC is still present. If yes then delete both of them
func CheckAndDeleteIndependentPV(name string, delPod bool) error {
	pv_check, err := k8sCore.GetPersistentVolume(name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return err
	}
	log.InfoD("Stranded PV Found by the name - %s. Going ahead to delete this PV and associated entities", name)
	if pv_check.Status.Phase == corev1.VolumeBound {
		if pv_check.Spec.ClaimRef != nil && pv_check.Spec.ClaimRef.Kind == "PersistentVolumeClaim" {
			namespace := pv_check.Spec.ClaimRef.Namespace
			pvc_name := pv_check.Spec.ClaimRef.Name
			// Delete all Pods in this namespace
			var newPods []corev1.Pod
			podList, err := GetPods(namespace)
			if err != nil {
				return err
			}
			for _, pod := range podList.Items {
				newPods = append(newPods, pod)
			}
			if delPod {
				err = DeletePods(newPods)
				if err != nil {
					return err
				}
			}

			// Delete PVC from figured out namespace
			err = k8sCore.DeletePersistentVolumeClaim(pvc_name, namespace)
			if err != nil {
				return err
			}
		}
	}
	// Delete PV as it is still available from previous run
	err = k8sCore.DeletePersistentVolume(name)
	if err != nil {
		return err
	}
	return nil
}

// RemoveFinalizersFromAllResources removes finalizers from all PDS resource for cleanup
func RemoveFinalizersFromAllResources(namespace string) error {

	ns, err := k8sCore.GetNamespace(namespace)
	if err != nil {
		return fmt.Errorf("unable to get namespace. Error - [%s]", err.Error())
	}

	allEndpoints, err := k8sCore.ListEndpoints(namespace, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("unable to list endpoints. Error - [%s]", err.Error())
	}
	log.Infof("Total number of endpoints found - [%s]", len(allEndpoints.Items))

	allServices, err := k8sCore.ListServices(namespace, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("unable to list services. Error - [%s]", err.Error())
	}
	log.Infof("Total number of services found - [%s]", len(allServices.Items))

	var allErrors []string

	log.Infof("Removing finalizers from endpoints")
	for _, endpoint := range allEndpoints.Items {

		log.Infof("Setting finalizer for [%s]", endpoint.Name)
		endpoint.SetFinalizers([]string{})

		_, err := k8sCore.UpdateEndpoints(&endpoint)
		if err != nil {
			allErrors = append(allErrors, err.Error())
		}
	}

	log.Infof("Removing finalizers from services")
	for _, service := range allServices.Items {

		log.Infof("Setting finalizer for [%s]", service.Name)
		service.SetFinalizers([]string{})

		_, err := k8sCore.UpdateService(&service)
		if err != nil {
			allErrors = append(allErrors, err.Error())
		}
	}

	log.Infof("Removing finalizers from namespace")
	ns.SetFinalizers([]string{})
	_, err = k8sCore.UpdateNamespace(ns)
	if err != nil {
		allErrors = append(allErrors, err.Error())
	}

	if len(allErrors) > 0 {
		return fmt.Errorf("error while removing finalizers. Error - [%s]", strings.Join(allErrors, "\n"))
	}

	return nil
}

// UpdatePDSNamespce updates the namespace
func UpdatePDSNamespce(name string, nsLables map[string]string) (*corev1.Namespace, error) {

	currentNSDetails, err := k8sCore.GetNamespace(name)
	if err != nil {
		return nil, err
	}

	nsSpec := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      nsLables,
			Annotations: currentNSDetails.Annotations,
		},
	}
	ns, err := k8sCore.UpdateNamespace(nsSpec)
	if err != nil {
		return nil, err
	}
	return ns, nil
}
