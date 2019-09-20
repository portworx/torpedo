package schedops

import (
	ap_api "github.com/libopenstorage/autopilot/pkg/apis/autopilot/v1alpha1"
	"github.com/libopenstorage/openstorage/api"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/errors"
)

type dcosSchedOps struct{}

func (d *dcosSchedOps) StartPxOnNode(n node.Node) error {
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "StartPxOnNode()",
	}
}

func (d *dcosSchedOps) StopPxOnNode(n node.Node) error {
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "StopPxOnNode()",
	}
}

func (d *dcosSchedOps) ValidateOnNode(n node.Node) error {
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "ValidateOnNode",
	}
}

func (d *dcosSchedOps) ValidateAddLabels(replicaNodes []api.Node, vol *api.Volume) error {
	// We do not have labels in DC/OS currently
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "ValidateAddLabels()",
	}
}

func (d *dcosSchedOps) ValidateRemoveLabels(vol *volume.Volume) error {
	// We do not have labels in DC/OS currently
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "ValidateRemoveLabels()",
	}
}

func (d *dcosSchedOps) GetVolumeName(vol *volume.Volume) string {
	return vol.Name
}

func (d *dcosSchedOps) ValidateVolumeCleanup(n node.Driver) error {
	// TODO: Implement this
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "ValidateVolumeCleanup()",
	}
}

func (d *dcosSchedOps) ValidateVolumeSetup(vol *volume.Volume, driver node.Driver) error {
	// TODO: Implement this
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "ValidateVolumeSetup()",
	}
}

func (d *dcosSchedOps) ValidateSnapshot(volParams map[string]string, parent *api.Volume) error {
	// TODO: Implement this
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "ValidateSnapshot()",
	}
}

func (d *dcosSchedOps) GetServiceEndpoint() (string, error) {
	// PX driver is accessed directly on agent nodes. There is no DC/OS level
	// service endpoint which can be used to redirect the calls to PX driver
	return "", &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "GetServiceEndpoint()",
	}
}

func (d *dcosSchedOps) UpgradePortworx(ociImage, ociTag, pxImage, pxTag string) error {
	// TOOD: Implement this method
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "UpgradePortworx()",
	}
}

func (d *dcosSchedOps) IsPXReadyOnNode(n node.Node) bool {
	// TODO: Implement this method
	return false
}

// IsPXEnabled should return whether given node has px installed or not
func (d *dcosSchedOps) IsPXEnabled(n node.Node) (bool, error) {
	// TODO: Implement this method
	return true, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "IsPXEnabled()",
	}
}

// GetStorageInfo returns cluster pair info from destination clusterrefereced by kubeconfig
func (d *dcosSchedOps) GetRemotePXNodes(destKubeConfig string) ([]node.Node, error) {
	// TODO: Implement this methid
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "GetRemotePXNodes()",
	}
}

// CreateAutopilotRule creates an autopilot rule
func (d *dcosSchedOps) CreateAutopilotRule(rule *ap_api.AutopilotRule) error {
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "CreateAutopilotRule()",
	}
}

func init() {
	d := &dcosSchedOps{}
	Register("dcos", d)
}
