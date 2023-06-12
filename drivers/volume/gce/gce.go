package gce

import (
	"fmt"

	"github.com/portworx/torpedo/pkg/log"

	"github.com/portworx/torpedo/drivers/volume"
	torpedovolume "github.com/portworx/torpedo/drivers/volume"
)

const (
	// DriverName is the name of the gce driver implementation
	DriverName = "gce"
	// GceStorage GCE storage driver name
	GceStorage torpedovolume.StorageProvisionerType = "gce"
)

// Provisioners types of supported provisioners
var provisioners = map[torpedovolume.StorageProvisionerType]torpedovolume.StorageProvisioner{
	GceStorage: "kubernetes.io/gce-pd",
}

type gce struct {
	torpedovolume.DefaultDriver
}

func (d *gce) String() string {
	return string(GceStorage)
}

func (d *gce) Init(volOpts volume.InitOptions) error {
	log.Infof("Using the GCE volume driver with provisioner %s under scheduler: %v", volOpts.StorageProvisionerType, volOpts.SchedulerDriverName)
	d.StorageDriver = DriverName
	// Set provisioner for torpedo
	if volOpts.StorageProvisionerType != "" {
		if p, ok := provisioners[volOpts.StorageProvisionerType]; ok {
			d.StorageProvisioner = p
		} else {
			return fmt.Errorf("volume driver [%s], does not support provisioner corresponding to type [%s]", DriverName, volOpts.StorageProvisionerType)
		}
	} else {
		return fmt.Errorf("Provisioner is empty for volume driver: %s", DriverName)
	}
	return nil
}

// DeepCopy deep copies the driver instance
func (d *gce) DeepCopy() volume.Driver {
	out := *d
	return &out
}

func init() {
	torpedovolume.Register(DriverName, provisioners, &gce{})
}
