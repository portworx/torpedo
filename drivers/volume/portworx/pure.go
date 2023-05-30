package portworx

import (
	"context"
	"fmt"
	"strconv"

	"github.com/libopenstorage/openstorage/api"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/log"
)

const (
	// PureDriverName is the name of the portworx-pure driver implementation
	PureDriverName = "pure"
)

// pure is essentially the same as the portworx volume driver, just different in name. This way,
// we can have separate specs for pure volumes vs. normal portworx ones
type pure struct {
	portworx
}

func (p *pure) Init(volOpts volume.InitOptions) error {
	return p.portworx.Init(volOpts)
}

func (p *pure) String() string {
	return PureDriverName
}

func (p *pure) ValidateCreateSnapshot(volumeName string, params map[string]string) error {
	var token string
	token = p.getTokenForVolume(volumeName, params)
	if val, hasKey := params[refreshEndpointParam]; hasKey {
		refreshEndpoint, _ := strconv.ParseBool(val)
		p.refreshEndpoint = refreshEndpoint
	}

	volDriver := p.getVolDriver()
	// This is the only difference: we have to name snapshots with hyphens, not underscores
	_, err := volDriver.SnapshotCreate(p.getContextWithToken(context.Background(), token), &api.SdkVolumeSnapshotCreateRequest{VolumeId: volumeName, Name: volumeName + "-snapshot"})
	if err != nil {
		log.Errorf(fmt.Sprintf("error when creating local snapshot, Err: %v", err))
		return err
	}
	return nil
}
