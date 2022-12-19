package azure

import (
	"fmt"

	v1 "github.com/libopenstorage/operator/pkg/apis/core/v1"
	torpedovolume "github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/drivers/volume/portworx/schedops"
	"github.com/portworx/torpedo/pkg/errors"
	"github.com/portworx/torpedo/pkg/log"
)

const (
	// DriverName is the name of the azure driver implementation
	DriverName = "azure"
	// AzureStorage Azure storage driver name
	AzureStorage torpedovolume.StorageProvisionerType = "azure"
)

// Provisioners types of supported provisioners
var provisioners = map[torpedovolume.StorageProvisionerType]torpedovolume.StorageProvisionerType{
	AzureStorage: "kubernetes.io/azure-disk",
}

type azure struct {
	schedOps schedops.Driver
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

func (d *azure) ValidateStorageCluster(endpointURL, endpointVersion string, autoUpdateComponents bool) error {
	// TODO: Add implementation
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "ValidateStorageCluster()",
	}
}

func (d *azure) UpdateAndValidateStorageCluster(cluster *v1.StorageCluster, f func(*v1.StorageCluster) *v1.StorageCluster, specGenUrl string, autoUpdateComponents bool) (*v1.StorageCluster, error) {
	// TODO: Add implementation
	return &v1.StorageCluster{}, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "UpdateAndValidateStorageCluster()",
	}
}

func (d *azure) Init(sched, nodeDriver, token, storageProvisioner, csiGenericDriverConfigMap string) error {
	log.Infof("Using the Azure volume driver with provisioner %s under scheduler: %v", storageProvisioner, sched)
	torpedovolume.StorageDriver = DriverName
	// Set provisioner for torpedo
	if storageProvisioner != "" {
		if p, ok := provisioners[torpedovolume.StorageProvisionerType(storageProvisioner)]; ok {
			torpedovolume.StorageProvisioner = p
		} else {
			return fmt.Errorf("driver %s, does not support provisioner %s", DriverName, storageProvisioner)
		}
	} else {
		torpedovolume.StorageProvisioner = provisioners[torpedovolume.DefaultStorageProvisioner]
	}
	return nil
}

func init() {
	torpedovolume.Register(DriverName, provisioners, &azure{})
}
