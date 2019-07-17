package gce_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/libopenstorage/cloudops"
	"github.com/libopenstorage/cloudops/gce"
	"github.com/libopenstorage/cloudops/test"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
	compute "google.golang.org/api/compute/v1"
)

const (
	newDiskSizeInGB    = 10
	newDiskPrefix      = "openstorage-test"
	newDiskDescription = "Disk created by Openstorage tests"
)

var diskName = fmt.Sprintf("%s-%s", newDiskPrefix, uuid.New())

func initGCE(t *testing.T) (cloudops.Ops, map[string]interface{}) {
	driver, err := gce.NewClient()
	require.NoError(t, err, "failed to instantiate storage ops driver")

	template := &compute.Disk{
		Description: newDiskDescription,
		Name:        diskName,
		SizeGb:      newDiskSizeInGB,
		Zone:        os.Getenv("GCE_INSTANCE_ZONE"),
	}

	return driver, map[string]interface{}{
		diskName: template,
	}
}

func TestAll(t *testing.T) {
	if gce.IsDevMode() {
		drivers := make(map[string]cloudops.Ops)
		diskTemplates := make(map[string]map[string]interface{})

		d, disks := initGCE(t)
		drivers[d.Name()] = d
		diskTemplates[d.Name()] = disks
		test.RunTest(drivers, diskTemplates, t)
	} else {
		fmt.Printf("skipping GCE tests as environment is not set...\n")
		t.Skip("skipping GCE tests as environment is not set...")
	}

}
