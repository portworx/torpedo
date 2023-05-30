package pso

import (
	"fmt"
	"strings"

	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/volume"
	torpedovolume "github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/drivers/volume/portworx/schedops"
	"github.com/portworx/torpedo/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// PureDriverName is the name of the portworx-pure driver implementation
	PureDriverName = "pso"
	PsoServiceName = "pso-csi-controller"
)

// Provisioners types of supported provisioners
var provisionersForPure = map[torpedovolume.StorageProvisionerType]torpedovolume.StorageProvisioner{
	PureDriverName: "pure-csi",
}

// pure is essentially the same as the portworx volume driver, just different in name. This way,
// we can have separate specs for pure volumes vs. normal portworx ones
type Pso struct {
	schedOps schedops.Driver
	torpedovolume.DefaultDriver
}

func (d *Pso) Init(volOpts volume.InitOptions) error {
	log.Infof("Using the Pure volume driver with provisioner %s under scheduler: %v", volOpts.StorageProvisionerType, volOpts.SchedulerDriverName)

	d.StorageDriver = PureDriverName
	// Set provisioner for torpedo
	if volOpts.StorageProvisionerType != "" {
		if p, ok := provisionersForPure[volOpts.StorageProvisionerType]; ok {
			d.StorageProvisioner = p
		} else {
			return fmt.Errorf("driver %s, does not support provisioner corresponding to type [%s]", PureDriverName, volOpts.StorageProvisionerType)
		}
	} else {
		return fmt.Errorf("Provisioner is empty for volume driver: %s", PureDriverName)
	}
	return nil
}

// DeepCopy deep copies the driver instance
func (d *Pso) DeepCopy() volume.Driver {
	out := *d
	//FIX: I'm unsure if this is a truly deep or shallow copy
	return &out
}

func (d *Pso) String() string {
	return PureDriverName
}

func (d *Pso) ValidateCreateVolume(name string, params map[string]string) error {
	// TODO: Implementation of ValidateCreateVolume will be provided in the coming PRs
	log.Warnf("ValidateCreateVolume function has not been implemented for volume driver - %s", d.String())
	return nil
}

func (d *Pso) ValidateVolumeSetup(vol *torpedovolume.Volume) error {
	// TODO: Implementation of ValidateVolumeSetup will be provided in the coming PRs
	log.Warnf("ValidateVolumeSetup function has not been implemented for volume driver - %s", d.String())
	return nil
}

func (d *Pso) ValidateDeleteVolume(vol *torpedovolume.Volume) error {
	// TODO: Implementation of ValidateDeleteVolume will be provided in the coming PRs
	log.Warnf("ValidateDeleteVolume function has not been implemented for volume driver - %s", d.String())
	return nil
}

func (d *Pso) GetDriverVersion() (string, error) {
	labelSelectors := map[string]string{
		"app": "pso-csi-node",
	}
	namespace, err := GetPsoNamespace()
	if err != nil {
		return "", err
	}
	pods, err := core.Instance().GetPods(namespace, labelSelectors)
	if err != nil {
		return "", err
	}
	podImage := pods.Items[0].Spec.Containers[1].Image
	psoVersion := strings.Split(podImage, ":")[1]
	log.Infof("PSO Version - %s", psoVersion)
	return psoVersion, nil
}

// GetPsoNamespace returns namespace where PSO is running
func GetPsoNamespace() (string, error) {
	allServices, err := core.Instance().ListServices("", metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get list of services. Err: %v", err)
	}
	for _, svc := range allServices.Items {
		if svc.Name == PsoServiceName {
			return svc.Namespace, nil
		}
	}
	return "", fmt.Errorf("can't find PSO service [%s] from list of services", PsoServiceName)
}

func init() {
	log.Infof("Registering pso driver")
	torpedovolume.Register(PureDriverName, provisionersForPure, &Pso{})
}
