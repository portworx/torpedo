package gce_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/libopenstorage/cloudops"
	"github.com/libopenstorage/cloudops/gce"
	"github.com/libopenstorage/cloudops/test"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
	compute "google.golang.org/api/compute/v1"
)

const (
	newDiskSizeInGB    = 10
	newDiskPrefix      = "gce-test"
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
		test.RunTest(drivers, diskTemplates, sizeCheck, t)
	} else {
		fmt.Printf("skipping GCE tests as environment is not set...\n")
		t.Skip("skipping GCE tests as environment is not set...")
	}
}

func TestInspectInstance(t *testing.T) {
	if gce.IsDevMode() {
		d, _ := initGCE(t)
		info, err := d.InspectInstance(os.Getenv("GCE_INSTANCE_NAME"))
		require.NoError(t, err)
		require.NotNil(t, info)
		logrus.Infof("[debug] got instance info: %v", info)
	} else {
		fmt.Printf("skipping GCE tests as environment is not set...\n")
		t.Skip("skipping GCE tests as environment is not set...")
	}

}

func sizeCheck(template interface{}, targetSize uint64) bool {
	disk, ok := template.(*compute.Disk)
	if !ok {
		return false
	}
	if disk.SizeGb == 0 {
		return false
	}
	return targetSize == uint64(disk.SizeGb)
}
