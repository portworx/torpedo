package azure

import (
	"fmt"

	"github.com/portworx/torpedo/drivers/volume"
	torpedovolume "github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/log"
)

const (
	// DriverName is the name of the azure driver implementation
	DriverName = "azure"
	// AzureStorage Azure storage driver name
	AzureStorage torpedovolume.StorageProvisionerType = "azure"
)

// Provisioners types of supported provisioners
var provisioners = map[torpedovolume.StorageProvisionerType]torpedovolume.StorageProvisioner{
	AzureStorage: "kubernetes.io/azure-disk",
}

type azure struct {
	torpedovolume.DefaultDriver
}

func (d *azure) String() string {
	return string(AzureStorage)
}

func (d *azure) ValidateVolumeCleanup() error {
	return nil
}

func (d *azure) RefreshDriverEndpoints() error {
	return nil
}

func (d *azure) Init(volOpts volume.InitOptions) error {
	log.Infof("Using the Azure volume driver with provisioner %s under scheduler: %v", volOpts.StorageProvisionerType, volOpts.SchedulerDriverName)
	d.StorageDriver = DriverName
	// Set provisioner for torpedo
	if volOpts.StorageProvisionerType != "" {
		if p, ok := provisioners[volOpts.StorageProvisionerType]; ok {
			d.StorageProvisioner = p
		} else {
			return fmt.Errorf("volume driver %s, does not support provisioner corresponding to type [%s]", DriverName, volOpts.StorageProvisionerType)
		}
	} else {
		d.StorageProvisioner = provisioners[torpedovolume.DefaultStorageProvisionerType]
	}
	return nil
}

// DeepCopy deep copies the driver instance
func (d *azure) DeepCopy() volume.Driver {
	out := *d
	return &out
}

func init() {
	torpedovolume.Register(DriverName, provisioners, &azure{})
}
