package aws

import (
	"fmt"

	"github.com/portworx/torpedo/drivers/volume"
	torpedovolume "github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/log"
)

const (
	// DriverName is the name of the aws driver implementation
	DriverName = "aws"
	// AwsStorage AWS storage driver name
	AwsStorage torpedovolume.StorageProvisionerType = "aws"
)

// Provisioners types of supported provisioners
var provisioners = map[torpedovolume.StorageProvisionerType]torpedovolume.StorageProvisioner{
	AwsStorage: "kubernetes.io/aws-ebs",
}

type aws struct {
	torpedovolume.DefaultDriver
}

func (d *aws) String() string {
	return string(AwsStorage)
}

func (d *aws) ValidateVolumeCleanup() error {
	return nil
}

func (d *aws) RefreshDriverEndpoints() error {
	return nil
}

func (d *aws) Init(volOpts volume.InitOptions) error {
	log.Infof("Using the AWS EBS volume driver with provisioner %s under scheduler: %v", volOpts.StorageProvisionerType, volOpts.SchedulerDriverName)
	d.StorageDriver = DriverName
	// Set provisioner for torpedo
	if volOpts.StorageProvisionerType != "" {
		if p, ok := provisioners[volOpts.StorageProvisionerType]; ok {
			d.StorageProvisioner = p
		} else {
			d.StorageProvisioner = provisioners[torpedovolume.DefaultStorageProvisionerType]
		}
	} else {
		return fmt.Errorf("Provisioner is empty for volume driver: %s", DriverName)
	}
	return nil
}

// DeepCopy deep copies the driver instance
func (d *aws) DeepCopy() volume.Driver {
	out := *d
	return &out
}

func init() {
	torpedovolume.Register(DriverName, provisioners, &aws{})
}
