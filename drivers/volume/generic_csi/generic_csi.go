package csi

import (
	"fmt"

	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/volume"
	torpedovolume "github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/log"
)

const (
	// DriverName is the name of the aws driver implementation
	DriverName = "csi"
	// CsiStorage CSI storage driver name
	CsiStorage torpedovolume.StorageProvisionerType = "generic_csi"
	// CsiStorageClassKey CSI Generic driver config map key name
	CsiStorageClassKey = "csi_storageclass_key"
)

// Provisioners types of supported provisioners
var provisioners = map[torpedovolume.StorageProvisionerType]torpedovolume.StorageProvisioner{
	CsiStorage: "csi",
}

type genericCsi struct {
	torpedovolume.DefaultDriver

	// below are pointers (manipulate carefully)

	k8sCore core.Ops
}

func (d *genericCsi) String() string {
	return string(CsiStorage)
}

func (d *genericCsi) ValidateVolumeCleanup() error {
	return nil
}

func (d *genericCsi) RefreshDriverEndpoints() error {
	return nil
}

func (d *genericCsi) Init(volOpts volume.InitOptions) error {
	log.Infof("Using the generic CSI volume driver with provisioner %s under scheduler: %v", volOpts.StorageProvisionerType, volOpts.SchedulerDriverName)
	d.k8sCore = volOpts.K8sCore

	d.StorageDriver = DriverName
	// Set provisioner for torpedo, from
	if volOpts.StorageProvisionerType == CsiStorage {
		// Get the provisioner from the config map
		configMap, err := d.k8sCore.GetConfigMap(volOpts.CsiGenericDriverConfigMap, "default")
		if err != nil {
			return fmt.Errorf("Failed to get config map for volume driver: %s, provisioner: %s", DriverName, volOpts.StorageProvisionerType)
		}
		if p, ok := configMap.Data[CsiStorageClassKey]; ok {
			d.StorageProvisioner = torpedovolume.StorageProvisioner(p)
		}
	} else {
		return fmt.Errorf("Invalid provisioner [%s] for volume driver [%s]", volOpts.StorageProvisionerType, DriverName)
	}
	return nil
}

// DeepCopy deep copies the driver instance
func (d *genericCsi) DeepCopy() volume.Driver {
	out := *d
	return &out
}

func init() {
	torpedovolume.Register(DriverName, provisioners, &genericCsi{})
}
