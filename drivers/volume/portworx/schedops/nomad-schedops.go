package schedops

import (
	"github.com/libopenstorage/openstorage/api"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/errors"
)

type nomadSchedOps struct{}

func (nm *nomadSchedOps) ValidateOnNode(n node.Node) error {
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "ValidateOnNode",
	}
}

func (nm *nomadSchedOps) ValidateAddLabels(replicaNodes []api.Node, vol *api.Volume) error {
	// We do not have labels in DC/OS currently
	return nil
}

func (nm *nomadSchedOps) ValidateRemoveLabels(vol *volume.Volume) error {
	// We do not have labels in DC/OS currently
	return nil
}

func (nm *nomadSchedOps) GetVolumeName(vol *volume.Volume) string {
	return vol.Name
}

func (nm *nomadSchedOps) ValidateVolumeCleanup(n node.Driver) error {
	// TODO: Implement this
	return nil
}

func (nm *nomadSchedOps) ValidateVolumeSetup(vol *volume.Volume) error {
	// TODO: Implement this
	return nil
}

func (nm *nomadSchedOps) ValidateSnapshot(volParams map[string]string, parent *api.Volume) error {
	// TODO: Implement this
	return nil
}

func (nm *nomadSchedOps) GetServiceEndpoint() (string, error) {
	// PX driver is accessed directly on agent nodes. There is no DC/OS level
	// service endpoint which can be used to redirect the calls to PX driver
	return "", nil
}

func (nm *nomadSchedOps) UpgradePortworx(image, tag string) error {
	// TOOD: Implement this method
	return nil
}

func (nm *nomadSchedOps) IsPXReadyOnNode(n node.Node) bool {
	// TODO: Implement this method
	return true
}

// IsPXEnabled should return whether given node has px installed or not
func (nm *nomadSchedOps) IsPXEnabled(n node.Node) (bool, error) {
	// TODO: Implement this method
	return true, nil
}

// GetStorageInfo returns cluster pair info from destination clusterrefereced by kubeconfig
func (nm *nomadSchedOps) GetRemotePXNodes(destKubeConfig string) ([]node.Node, error) {
	// TODO: Implement this methid
	return nil, nil
}

func init() {
	nm := &nomadSchedOps{}
	Register("nomad", nm)
}
