package pure

import (
	"fmt"
	torpedovolume "github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/drivers/volume/portworx"
	"github.com/portworx/torpedo/drivers/volume/portworx/schedops"
	"github.com/portworx/torpedo/pkg/log"
)

const (
	// PureDriverName is the name of the portworx-pure driver implementation
	PureDriverName = "pure"
)

// Provisioners types of supported provisioners
var provisionersForPure = map[torpedovolume.StorageProvisionerType]torpedovolume.StorageProvisionerType{
	PureDriverName: "pure-csi",
}

// pure is essentially the same as the portworx volume driver, just different in name. This way,
// we can have separate specs for pure volumes vs. normal portworx ones
type pure struct {
	schedOps schedops.Driver
	torpedovolume.DefaultDriver
}

func (d *pure) Init(sched, nodeDriver, token, storageProvisioner, csiGenericDriverConfigMap string) error {
	log.Infof("Using the Pure volume driver with provisioner %s under scheduler: %v", storageProvisioner, sched)
	torpedovolume.StorageDriver = portworx.DriverName
	// Set provisioner for torpedo
	if storageProvisioner != "" {
		if p, ok := provisionersForPure[torpedovolume.StorageProvisionerType(storageProvisioner)]; ok {
			torpedovolume.StorageProvisioner = p
		} else {
			return fmt.Errorf("driver %s, does not support provisioner %s", portworx.DriverName, storageProvisioner)
		}
	} else {
		return fmt.Errorf("Provisioner is empty for volume driver: %s", portworx.DriverName)
	}
	return nil
}

func (d *pure) String() string {
	return PureDriverName
}

func (d *pure) ValidateCreateVolume(name string, params map[string]string) error {
	return nil
}

func (d *pure) ValidateVolumeSetup(vol *torpedovolume.Volume) error {
	return nil
}

func (d *pure) ValidateDeleteVolume(vol *torpedovolume.Volume) error {
	return nil
}

func (d *pure) GetDriverVersion() (string, error) {
	return "1.0.0-1.0.0", nil
}

func init() {
	log.Infof("Registering pure")
	torpedovolume.Register(PureDriverName, provisionersForPure, &pure{})
}
