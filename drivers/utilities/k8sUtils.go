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

func IncreasePVCby1Gig(namespace string, deployment map[string]string, sizeInGb uint64) (*volume.Volume, error) {
	log.Info("Resizing of the PVC begins")
	var vol *volume.Volume
	pvcList, _ := GetPvsAndPVCsfromDeployment(namespace, deployment)
	initialCapacity, err := GetVolumeCapacityInGB(namespace, deployment)
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
	newcapacity, err := GetVolumeCapacityInGB(namespace, deployment)
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

func GetPvsAndPVCsfromDeployment(namespace string, deployment map[string]string) (*corev1.PersistentVolumeClaimList, []*volume.Volume) {
	log.Infof("Get PVC List based on namespace and deployment")
	var vols []*volume.Volume
	labelSelector := make(map[string]string)
	labelSelector["name"] = "deployment.GetClusterResourceName()"
	pvcList, _ := k8sCores.GetPersistentVolumeClaims(namespace, labelSelector)
	for _, pvc := range pvcList.Items {
		vols = append(vols, &volume.Volume{
			ID: pvc.Spec.VolumeName,
		})
	}
	return pvcList, vols
}

func GetVolumeCapacityInGB(namespace string, deployment map[string]string) (uint64, error) {
	var pvcCapacity uint64
	_, vols := GetPvsAndPVCsfromDeployment(namespace, deployment)
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
