package vsphere

import (
	"fmt"
	"testing"

	"github.com/libopenstorage/cloudops"
	"github.com/libopenstorage/cloudops/test"
	"github.com/libopenstorage/cloudops/vsphere/lib/vsphere/vclib"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
	"github.com/vmware/govmomi/object"
)

const (
	newDiskSizeInKB = 2097152 // 2GB
	newDiskPrefix   = "openstorage-test"
)

var (
	datastoreForTest string
	driver           cloudops.Ops
)

var diskName = fmt.Sprintf("%s-%s", newDiskPrefix, uuid.New())

func initVsphere(t *testing.T) (cloudops.Ops, map[string]interface{}) {
	cfg, err := ReadVSphereConfigFromEnv()
	require.NoError(t, err, "failed to get vsphere config from env")

	cfg.VMUUID, err = cloudops.GetEnvValueStrict("VSPHERE_VM_UUID")
	require.NoError(t, err, "failed to get vsphere config from env variable VSPHERE_VM_UUID")

	datastoreForTest, err = cloudops.GetEnvValueStrict("VSPHERE_TEST_DATASTORE")
	require.NoError(t, err, "failed to get datastore from env variable VSPHERE_TEST_DATASTORE")

	driver, err = NewClient(cfg)
	require.NoError(t, err, "failed to instantiate storage ops driver")

	diskOptions := &vclib.VolumeOptions{
		Name:       diskName,
		CapacityKB: newDiskSizeInKB,
		Datastore:  datastoreForTest,
		DiskFormat: vclib.EagerZeroedThickDiskType,
	}

	return driver, map[string]interface{}{
		diskName: diskOptions,
	}
}

func sizeCheck(template interface{}, targetSize uint64) bool {
	_, ok := template.(*object.VirtualDiskInfo)
	if !ok {
		return false
	}
	// inspect of vmdks does not return size
	return true
}

func TestAll(t *testing.T) {
	if IsDevMode() {
		drivers := make(map[string]cloudops.Ops)
		diskTemplates := make(map[string]map[string]interface{})

		d, disks := initVsphere(t)
		drivers[d.Name()] = d
		diskTemplates[d.Name()] = disks

		test.RunTest(drivers, diskTemplates, sizeCheck, t)
	} else {
		fmt.Printf("skipping vSphere tests as environment is not set...\n")
		t.Skip("skipping vSphere tests as environment is not set...")
	}
}

func TestDevicePath(t *testing.T) {
	if IsDevMode() {
		d, _ := initVsphere(t)
		require.NotNil(t, d)

		// Based on your VM and environment, set
		// VSPHERE_VM_TEST_DEVICE_PATH="[Phy-vsanDatastore] 260f0d5d-207e-2372-3d57-ac1f6b204d08/PX-DO-NOT-DELETE-6004befe-b554-4283-bc6a-efacc4a72010.vmdk"
		testDevPath, err := cloudops.GetEnvValueStrict("VSPHERE_VM_TEST_DEVICE_PATH")
		if err != nil {
			t.Skip("skipping vSphere device path test as test device path is not set...")
		}

		attachedPath, err := d.DevicePath(testDevPath)
		require.NoError(t, err, "failed to get attached device path")
		require.NotEmpty(t, attachedPath)
	} else {
		fmt.Printf("skipping vSphere device path test as environment is not set...\n")
		t.Skip("skipping vSphere device test as environment is not set...")
	}

}
