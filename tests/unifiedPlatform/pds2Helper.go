package tests

import (
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/pkg/units"
	"github.com/portworx/torpedo/tests"
	corev1 "k8s.io/api/core/v1"
)

var k8sCores = core.Instance()

func GetPvsAndPVCsForDep(namespace string, deploymentName string) (*corev1.PersistentVolumeClaimList, []*volume.Volume) {
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
	_, vols := GetPvsAndPVCsForDep(namespace, deploymentName)
	for _, vol := range vols {
		appVol, err := tests.Inst().V.InspectVolume(vol.ID)
		if err != nil {
			return 0, err
		}
		pvcCapacity = appVol.Spec.Size / units.GiB
	}
	return pvcCapacity, nil
}
