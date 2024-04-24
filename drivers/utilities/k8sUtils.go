package utilities

import (
	"fmt"
	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/osutils"
	"github.com/portworx/torpedo/pkg/units"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"math/rand"
	"strings"
	"time"
)

const (
	timeInterval = 10 * time.Second
	timeOut      = 30 * time.Minute
)

var (
	instance *Torpedo
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
	return instance
}

func IncreasePVCby1Gig(namespace string, deploymentName string, sizeInGb uint64) (*volume.Volume, error) {
	log.Info("Resizing of the PVC begins")
	var vol *volume.Volume
	pvcList, _ := GetPvsAndPVCsfromDeployment(namespace, deploymentName)
	initialCapacity, err := GetVolumeCapacityInGB(namespace, deploymentName)
	log.Debugf("Initial volume storage size is : %v", initialCapacity)
	if err != nil {
		return nil, err
	}
	for _, pvc := range pvcList.Items {
		storageSize := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
		extraAmount, _ := resource.ParseQuantity(fmt.Sprintf("%dGi", sizeInGb))
		storageSize.Add(extraAmount)
		pvc.Spec.Resources.Requests[corev1.ResourceStorage] = storageSize
		_, err := k8sCores.UpdatePersistentVolumeClaim(&pvc)
		if err != nil {
			return nil, err
		}
		sizeInt64, _ := storageSize.AsInt64()
		vol = &volume.Volume{
			Name:          pvc.Name,
			RequestedSize: uint64(sizeInt64),
		}
	}
	// wait for the resize to take effect
	time.Sleep(30 * time.Second)
	newcapacity, err := GetVolumeCapacityInGB(namespace, deploymentName)
	log.Infof("Resized volume storage size is : %v", newcapacity)
	if err != nil {
		return nil, err
	}
	if newcapacity > initialCapacity {
		log.InfoD("Successfully resized the pvc by 1gb")
		return vol, nil
	} else {
		return vol, err
	}
}

func GetPvsAndPVCsfromDeployment(namespace string, deploymentName string) (*corev1.PersistentVolumeClaimList, []*volume.Volume) {
	log.Infof("Get PVC List based on namespace and deployment")
	var vols []*volume.Volume
	labelSelector := make(map[string]string)
	labelSelector["name"] = deploymentName
	pvcList, _ := k8sCores.GetPersistentVolumeClaims(namespace, labelSelector)
	for _, pvc := range pvcList.Items {
		vols = append(vols, &volume.Volume{
			ID: pvc.Spec.VolumeName,
		})
	}
	return pvcList, vols
}

func GetVolumeCapacityInGB(namespace string, deploymentName string) (uint64, error) {
	var pvcCapacity uint64
	_, vols := GetPvsAndPVCsfromDeployment(namespace, deploymentName)
	for _, vol := range vols {
		appVol, err := Inst().V.InspectVolume(vol.ID)
		if err != nil {
			return 0, err
		}
		pvcCapacity = appVol.Spec.Size / units.GiB
	}
	return pvcCapacity, nil
}

func GetDbMasterNode(namespace string, dsName string, deployment string, kubeconfigPath string) (string, bool) {
	var command, dbMaster string
	switch dsName {
	case "deployment.Postgresql":
		command = fmt.Sprintf("patronictl list | grep -i leader | awk '{print $2}'")
		dbMaster, _ = ExecuteCommandInStatefulSetPod("deployment.GetClusterResourceName()", namespace, command)
		//log.FailOnError(err, "Failed while fetching db master pods=.")
		//log.Infof("Deployment %v of type %v have the master "+
		//"running at %v pod.", deployment.GetClusterResourceName(), dsName, dbMaster)
	case "deployment.Mysql":
		//_, connectionDetails, err := pdslib.ApiComponents.DataServiceDeployment.GetConnectionDetails("deployment.GetId()")
		//log.FailOnError(err, "Failed while fetching connection details.")
		//cred, err := pdslib.ApiComponents.DataServiceDeployment.GetDeploymentCredentials("deployment.GetId()")
		//log.FailOnError(err, "Failed while fetching credentials.")
		//command = fmt.Sprintf("mysqlsh --host=%v --port %v --user=innodb-config "+
		//" --password=%v -- cluster status", connectionDetails["host"], connectionDetails["port"], cred.GetPassword())
		dbMaster, _ = ExecuteCommandInStatefulSetPod("deployment.GetClusterResourceName()", namespace, command)
		//log.Infof("Deployment %v of type %v have the master "+
		//"running at %v pod.", deployment.GetClusterResourceName(), dsName, dbMaster)
	default:
		return "", false
	}
	return dbMaster, true
}

// ExecuteCommandInStatefulSetPod executes the provided command inside a pod within the specified StatefulSet.
func ExecuteCommandInStatefulSetPod(statefulsetName, namespace, command string) (string, error) {
	podName, err := GetAnyPodName(statefulsetName, namespace)
	if err != nil {
		return "", err
	}

	return ExecCommandInPod(podName, namespace, command)
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

func ExecCommandInPod(podName, namespace, command string) (string, error) {
	cmd := fmt.Sprintf("kubectl --kubeconfig %v -n %v exec -it %v -- %v", "targetCluster.kubeconfig", namespace, podName, command)
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
