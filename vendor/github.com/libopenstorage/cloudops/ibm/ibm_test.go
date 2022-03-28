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
	fmt.Printf("Inside init.. getting new IBM cloud client")
	driver, err := ibm.NewClient()
	require.NoError(t, err, "failed to instantiate storage ops driver")
	fmt.Printf("successfully got ibm client in initIBM")

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

/*func TestInspectInstance(t *testing.T) {
	if gce.IsDevMode() {
		d, _ := initIBM(t)
		info, err := d.InspectInstance(os.Getenv("GCE_INSTANCE_NAME"))
		require.NoError(t, err)
		require.NotNil(t, info)
		logrus.Infof("[debug] got instance info: %v", info)
	} else {
		fmt.Printf("skipping GCE tests as environment is not set...\n")
		t.Skip("skipping GCE tests as environment is not set...")
	}

}*/

func sizeCheck(template interface{}, targetSize uint64) bool {
	return true
}
