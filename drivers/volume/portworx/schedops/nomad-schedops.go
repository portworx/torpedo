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
	// TODO: Implement this method
	return nil
}

func (nm *nomadSchedOps) ValidateRemoveLabels(vol *volume.Volume) error {
	// TODO: Implement this method
	return nil
}

func (nm *nomadSchedOps) GetVolumeName(vol *volume.Volume) string {
	// TODO: Implement this method
	return ""
}

func (nm *nomadSchedOps) ValidateVolumeCleanup(n node.Driver) error {
	// TODO: Implement this method
	return nil
}

func (nm *nomadSchedOps) ValidateVolumeSetup(vol *volume.Volume, d node.Driver) error {
	// TODO: Implement this method
	return nil
}

func (nm *nomadSchedOps) ValidateSnapshot(volParams map[string]string, parent *api.Volume) error {
	// TODO: Implement this method
	return nil
}

func (nm *nomadSchedOps) GetServiceEndpoint() (string, error) {
	// TODO: Implement this method
	return "", nil
}

func (nm *nomadSchedOps) UpgradePortworx(string, string, string, string) error {
	// TODO: Implement this method
	return nil
}

func (nm *nomadSchedOps) IsPXReadyOnNode(n node.Node) bool {
	// TODO: Implement this method
	return true
}

func (nm *nomadSchedOps) IsPXEnabled(n node.Node) (bool, error) {
	// TODO: Implement this method
	return true, nil
}

func (nm *nomadSchedOps) StartPxOnNode(n node.Node) error {
	// TODO: Implement this method
	return nil
}

func (nm *nomadSchedOps) StopPxOnNode(n node.Node) error {
	// TODO: Implement this method
	return nil
}

func (nm *nomadSchedOps) GetRemotePXNodes(destKubeConfig string) ([]node.Node, error) {
	// TODO: Implement this methid
	return nil, nil
}

func init() {
	nm := &nomadSchedOps{}
	Register("nomad", nm)
}
