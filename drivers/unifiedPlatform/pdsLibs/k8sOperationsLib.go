package pdslibs

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/units"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"time"
)

var instance *Torpedo

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

const (
	IncreasePvcSizeBy1gb = "increase-pvc-size-by-1gb"
)

func ExecuteK8sOperations(functionName string, namespace string, deployment map[string]string) error {
	switch functionName {
	case IncreasePvcSizeBy1gb:
		_, err := IncreasePVCby1Gig(namespace, deployment, 1)
		if err != nil {
			return err
		}

	}
	return nil
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
		k8sOps := k8sCore
		storageSize := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
		extraAmount, _ := resource.ParseQuantity(fmt.Sprintf("%dGi", sizeInGb))
		storageSize.Add(extraAmount)
		pvc.Spec.Resources.Requests[corev1.ResourceStorage] = storageSize
		_, err := k8sOps.UpdatePersistentVolumeClaim(&pvc)
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
	k8sOps := k8sCore
	var vols []*volume.Volume
	labelSelector := make(map[string]string)
	labelSelector["name"] = "deployment.GetClusterResourceName()"
	pvcList, _ := k8sOps.GetPersistentVolumeClaims(namespace, labelSelector)
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
	return pvcCapacity, err
}
