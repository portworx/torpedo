package ibm_test

import (
	"fmt"
	"os"
	"testing"

	"google.golang.org/api/compute/v1"

	"github.com/libopenstorage/cloudops"
	"github.com/libopenstorage/cloudops/ibm"
	"github.com/libopenstorage/cloudops/test"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
)

const (
	newDiskSizeInGB    = 10
	newDiskPrefix      = "ibm-test"
	newDiskDescription = "Disk created by Openstorage tests"
)

var diskName = fmt.Sprintf("%s-%s", newDiskPrefix, uuid.New())

func initIBM(t *testing.T) (cloudops.Ops, map[string]interface{}) {
	driver, err := ibm.NewClient()
	require.NoError(t, err, "failed to instantiate storage ops driver")

	template := &compute.Disk{
		Description: newDiskDescription,
		Name:        diskName,
		SizeGb:      newDiskSizeInGB,
		Zone:        os.Getenv("IBM_INSTANCE_ZONE"),
	}

	return driver, map[string]interface{}{
		diskName: template,
	}
}

func TestAll(t *testing.T) {
	fmt.Println("Inside TestALL")
	drivers := make(map[string]cloudops.Ops)
	diskTemplates := make(map[string]map[string]interface{})

	d, disks := initIBM(t)
	drivers[d.Name()] = d
	diskTemplates[d.Name()] = disks
	test.RunTest(drivers, diskTemplates, sizeCheck, t)
}

func sizeCheck(template interface{}, targetSize uint64) bool {
	return true
}
