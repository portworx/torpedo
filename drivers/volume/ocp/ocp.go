package ocp

import (
	"fmt"
	"github.com/libopenstorage/openstorage/api"
	torpedovolume "github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/drivers/volume/portworx"
	"github.com/portworx/torpedo/drivers/volume/portworx/schedops"
	"github.com/portworx/torpedo/pkg/errors"
	"github.com/portworx/torpedo/pkg/log"
)

const (
	OcpDriverName        = "ocp"
	OcpRbdCephDriverName = "rbd-csi"
	OcpCephfsDriverName  = "cephfs-csi"
	OcpRgwDriverName     = "rgw-csi"
	// OcpServiceName is the name of the ocp storage driver implementation
	OcpServiceName = ""
)

// Provisioners types of supported provisioners
var provisionersForOcp = map[torpedovolume.StorageProvisionerType]torpedovolume.StorageProvisionerType{
	OcpRbdCephDriverName: "openshift-storage.rbd.csi.ceph.com",
	OcpCephfsDriverName:  "openshift-storage.cephfs.csi.ceph.com",
	OcpRgwDriverName:     "openshift-storage.ceph.rook.io/bucket",
}

type Ocp struct {
	schedOps schedops.Driver
	torpedovolume.DefaultDriver
}

func (o *Ocp) Init(sched, nodeDriver, token, storageProvisioner, csiGenericDriverConfigMap string) error {
	log.Infof("Using the OCP volume driver with provisioner %s under scheduler: %v", storageProvisioner, sched)
	torpedovolume.StorageDriver = OcpDriverName
	// Set provisioner for torpedo
	if storageProvisioner != "" {
		if p, ok := provisionersForOcp[torpedovolume.StorageProvisionerType(storageProvisioner)]; ok {
			torpedovolume.StorageProvisioner = p
		} else {
			return fmt.Errorf("driver %s, does not support provisioner %s", portworx.DriverName, storageProvisioner)
		}
	} else {
		return fmt.Errorf("Provisioner is empty for volume driver: %s", portworx.DriverName)
	}
	return nil
}

func (o *Ocp) UpdateProvisioner(provisioner string) error {
	torpedovolume.StorageProvisioner = torpedovolume.StorageProvisionerType(provisioner)
	return nil
}

func (o *Ocp) String() string {
	return OcpDriverName
}

func (o *Ocp) ValidateCreateVolume(name string, params map[string]string) error {
	// TODO: Implementation of ValidateCreateVolume will be provided in the coming PRs
	log.Warnf("ValidateCreateVolume function has not been implemented for volume driver - %s", o.String())
	return nil
}

func (o *Ocp) ValidateVolumeSetup(vol *torpedovolume.Volume) error {
	// TODO: Implementation of ValidateVolumeSetup will be provided in the coming PRs
	log.Warnf("ValidateVolumeSetup function has not been implemented for volume driver - %s", o.String())
	return nil
}

func (o *Ocp) ValidateDeleteVolume(vol *torpedovolume.Volume) error {
	// TODO: Implementation of ValidateDeleteVolume will be provided in the coming PRs
	log.Warnf("ValidateDeleteVolume function has not been implemented for volume driver - %s", o.String())
	return nil
}

func (o *Ocp) GetDriverVersion() (string, error) {
	// TODO: Implementation of ValidateDeleteVolume will be provided in the coming PRs
	log.Warnf("GetDriverVersion function has not been implemented for volume driver - %s", o.String())
	return "", nil
}

// RefreshDriverEndpoints get the updated driver endpoints for the cluster
func (o *Ocp) RefreshDriverEndpoints() error {

	log.Warnf("RefreshDriverEndpoints function has not been implemented for volume driver - %s", o.String())
	return nil
}

func (o *Ocp) GetProxySpecForAVolume(volume *torpedovolume.Volume) (*api.ProxySpec, error) {
	log.Warnf("GetProxySpecForAVolume function has not been implemented for volume driver - %s", o.String())
	return nil, nil
}

func (o *Ocp) InspectCurrentCluster() (*api.SdkClusterInspectCurrentResponse, error) {
	log.Warnf("InspectCurrentCluster function has not been implemented for volume driver - %s", o.String())
	return nil, nil
}

// InspectVolume inspects the volume with the given name
func (o *Ocp) InspectVolume(name string) (*api.Volume, error) {
	log.Warnf("InspectVolume function has not been implemented for volume driver - %s", o.String())
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "InspectVolume()",
	}
}

func init() {
	log.Infof("Registering ocp driver")
	torpedovolume.Register(OcpDriverName, provisionersForOcp, &Ocp{})
}
