package vsphere

import (
	"context"
	"fmt"
	"github.com/portworx/sched-ops/k8s/core/configmap"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/libopenstorage/cloudops"
	"github.com/libopenstorage/cloudops/backoff"
	"github.com/libopenstorage/cloudops/unsupported"
	"github.com/libopenstorage/cloudops/vsphere/lib/vsphere/vclib"
	"github.com/libopenstorage/cloudops/vsphere/lib/vsphere/vclib/diskmanagers"
	"github.com/sirupsen/logrus"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/units"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/govmomi/vslm"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	vSphereDataStoreLock = "vsphere-ds-lock"
)

type vsphereOps struct {
	cloudops.Compute
	vm     *vclib.VirtualMachine
	conn   *vclib.VSphereConnection
	cfg    *VSphereConfig
	dsLock configmap.ConfigMap
}

var (
	exponentialBackoff = wait.Backoff{
		Duration: 2 * time.Second, // the base duration
		Factor:   1.2,             // Duration is multiplied by factor each iteration
		Jitter:   1.0,             // The amount of jitter applied each iteration
		Steps:    12,              // Exit with error after this many steps
	}

	vmdkMatcherRegex = regexp.MustCompile("\\[(.+)\\](.+)") // e.g [px-datastore-02] osd-provisioned-disks/PX-DO-NOT-DELETE-122be943-485f-4a66-b665-b592f685a3de.vmdk"

)

// VirtualDisk encapsulates the existing virtual disk object to add a managed object
// reference to the datastore of the disk
type VirtualDisk struct {
	diskmanagers.VirtualDisk
	// DatastoreRef is the managed object reference of the datastore on which the disk belongs
	DatastoreRef types.ManagedObjectReference
}

// NewClient creates a new vsphere cloudops instance
func NewClient(cfg *VSphereConfig) (cloudops.Ops, error) {
	vSphereConn := &vclib.VSphereConnection{
		Username:          cfg.User,
		Password:          cfg.Password,
		Hostname:          cfg.VCenterIP,
		Insecure:          cfg.InsecureFlag,
		RoundTripperCount: cfg.RoundTripperCount,
		Port:              cfg.VCenterPort,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	vmObj, err := GetVMObject(ctx, vSphereConn, cfg.VMUUID)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("Using following configuration for vsphere:")
	logrus.Debugf("  vCenter: %s:%s", cfg.VCenterIP, cfg.VCenterPort)
	logrus.Debugf("  Datacenter: %s", vmObj.Datacenter.Name())
	logrus.Debugf("  VMUUID: %s", cfg.VMUUID)

	configMap, err := configmap.New(vSphereDataStoreLock, nil, 0, 420, 5*time.Second, 10*time.Second)
	if err != nil {
		logrus.Errorf(err.Error())
		return nil, err
	}

	return backoff.NewExponentialBackoffOps(
		&vsphereOps{
			Compute: unsupported.NewUnsupportedCompute(),
			cfg:     cfg,
			vm:      vmObj,
			conn:    vSphereConn,
			dsLock:  configMap,
		},
		isExponentialError,
		exponentialBackoff,
	), nil
}

func (ops *vsphereOps) Name() string { return string(cloudops.Vsphere) }

func (ops *vsphereOps) InstanceID() string { return ops.cfg.VMUUID }

func (ops *vsphereOps) Create(opts interface{}, labels map[string]string) (interface{}, error) {
	volumeOptions, ok := opts.(*vclib.VolumeOptions)
	if !ok {
		return nil, fmt.Errorf("invalid volume options specified to create: %v", opts)
	}

	if len(volumeOptions.Tags) == 0 {
		volumeOptions.Tags = labels
	} else {
		for k, v := range labels {
			volumeOptions.Tags[k] = v
		}
	}

	if len(volumeOptions.Datastore) == 0 {
		return nil, fmt.Errorf("datastore is required for the create call")
	}

	datastore := strings.TrimSpace(volumeOptions.Datastore)
	logrus.Infof("Given datastore/datastore cluster: %s for new disk", datastore)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if ops.vm == nil {
		return nil, fmt.Errorf("vm is not set")
	}

	vmObj, err := ops.renewVM(ctx, ops.vm)
	if err != nil {
		return nil, err
	}

	isPod, storagePod, err := IsStoragePod(ctx, vmObj, volumeOptions.Datastore)
	if err != nil {
		return nil, err
	}

	if isPod {
		datastore, err = ops.getDatastoreToUseInStoragePod(ctx, vmObj, volumeOptions, storagePod)
		if err != nil {
			return nil, err
		}
	}

	logrus.Infof("Using datastore: %s for new disk", datastore)
	ds, err := vmObj.Datacenter.GetDatastoreByName(ctx, datastore)
	if err != nil {
		logrus.Errorf("Failed to get datastore: %s due to: %v", datastore, err)
		return nil, err
	}

	volumeOptions.Datastore = datastore

	var (
		keepAfterDeleteVm = true
		disk              diskmanagers.VirtualDisk
	)

	about := vmObj.Client().ServiceContent.About
	apiVersion, err := version.NewVersion(about.ApiVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to detect vSphere API version due to: %v", err)
	}

	keepDiskVersion, err := version.NewVersion(keepAfterDeleteVMApiVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse vSphere API version that supports keepAfterDeleteVM due to: %v", err)
	}

	if apiVersion.GreaterThan(keepDiskVersion) || apiVersion.Equal(keepDiskVersion) {
		// create disk using the new API so it doesn't get deleted after VM deletion
		m := vslm.NewObjectManager(vmObj.Client())
		var provisioningType string
		switch volumeOptions.DiskFormat {
		case vclib.LazyZeroedThickDiskType:
			provisioningType = string(types.BaseConfigInfoDiskFileBackingInfoProvisioningTypeLazyZeroedThick)
		case vclib.ThinDiskType:
			provisioningType = string(types.BaseConfigInfoDiskFileBackingInfoProvisioningTypeThin)
		case vclib.ZeroedThickDiskType:
			fallthrough
		case vclib.EagerZeroedThickDiskType:
			fallthrough
		default:
			provisioningType = string(types.BaseConfigInfoDiskFileBackingInfoProvisioningTypeEagerZeroedThick)
		}

		spec := types.VslmCreateSpec{
			Name:              volumeOptions.Name,
			CapacityInMB:      int64(volumeOptions.CapacityKB*1024) / units.MB,
			KeepAfterDeleteVm: &keepAfterDeleteVm,
			BackingSpec: &types.VslmCreateSpecDiskFileBackingSpec{
				VslmCreateSpecBackingSpec: types.VslmCreateSpecBackingSpec{
					Datastore: ds.Reference(),
				},
				ProvisioningType: provisioningType,
			},
		}

		task, err := m.CreateDisk(ctx, spec)
		if err != nil {
			return nil, err
		}

		res, err := task.WaitForResult(ctx)
		if err != nil {
			return nil, err
		}

		if res == nil || res.Result == nil {
			return nil, fmt.Errorf("got empty result while creating vSphere disk")
		}

		myDisk := res.Result.(types.VStorageObject)
		logrus.Infof("Created vSphere disk (VMDK) with ID: %s", myDisk.Config.Id.Id)

		fileInfo, ok := myDisk.Config.Backing.(*types.BaseConfigInfoDiskFileBackingInfo)
		if !ok {
			return nil, fmt.Errorf("failed to get disk file path: %v", myDisk)
		}
		disk = diskmanagers.VirtualDisk{
			DiskPath:      fileInfo.FilePath,
			VolumeOptions: volumeOptions,
		}
	} else {
		diskBasePath := filepath.Clean(ds.Path(diskDirectory)) + "/"
		err = ds.CreateDirectory(ctx, diskBasePath, false)
		if err != nil && err != vclib.ErrFileAlreadyExist {
			logrus.Errorf("Cannot create dir %#v. err %s", diskBasePath, err)
			return nil, err
		}

		diskPath := diskBasePath + volumeOptions.Name + ".vmdk"
		disk = diskmanagers.VirtualDisk{
			DiskPath:      diskPath,
			VolumeOptions: volumeOptions,
		}

		diskPath, err = disk.Create(ctx, ds)
		if err != nil {
			logrus.Errorf("Failed to create a vsphere volume with volumeOptions: %+v on "+
				"datastore: %s. err: %+v", volumeOptions, datastore, err)
			return nil, err
		}

		// Get the canonical path for the volume path.
		canonicalVolumePath, err := getCanonicalVolumePath(ctx, vmObj.Datacenter, diskPath)
		if err != nil {
			logrus.Errorf("Failed to get canonical vsphere disk path for: %s with "+
				"volumeOptions: %+v on datastore: %s. err: %+v", diskPath, volumeOptions, datastore, err)
			return nil, err
		}

		disk.DiskPath = canonicalVolumePath
	}

	return &VirtualDisk{
		VirtualDisk:  disk,
		DatastoreRef: ds.Reference(),
	}, nil
}

func (ops *vsphereOps) GetDeviceID(vDisk interface{}) (string, error) {
	disk, ok := vDisk.(*VirtualDisk)
	if !ok {
		return "", fmt.Errorf("invalid input: %v to GetDeviceID", vDisk)
	}

	return disk.DiskPath, nil
}

// Attach takes in the path of the vmdk file and returns where it is attached inside the vm instance
func (ops *vsphereOps) Attach(diskPath string, options map[string]string) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vmObj, err := ops.renewVM(ctx, ops.vm)
	if err != nil {
		return "", err
	}

	volOpts := &vclib.VolumeOptions{SCSIControllerType: vclib.PVSCSIControllerType}
	attachMode, ok := options[DiskAttachMode]
	if ok {
		if strings.TrimSpace(attachMode) == string(types.VirtualDiskModePersistent) {
			about := vmObj.Client().ServiceContent.About
			apiVersion, err := version.NewVersion(about.ApiVersion)
			if err != nil {
				return "", fmt.Errorf("failed to detect vSphere API version due to: %v", err)
			}

			keepDiskVersion, err := version.NewVersion(keepAfterDeleteVMApiVersion)
			if err != nil {
				return "", fmt.Errorf("failed to parse vSphere API version that supports keepAfterDeleteVM due to: %v", err)
			}
			if apiVersion.LessThan(keepDiskVersion) {
				return "", fmt.Errorf("attaching disk as persistent is not supported for version less than %s", keepDiskVersion)
			}
		}
		volOpts.DiskMode = attachMode
	}
	diskUUID, err := vmObj.AttachDisk(ctx, diskPath, volOpts)
	if err != nil {
		logrus.Errorf("Failed to attach vsphere disk: %s for VM: %s. err: +%v", diskPath, vmObj.Name(), err)
		return "", err
	}

	return path.Join(diskByIDPath, diskSCSIPrefix+diskUUID), nil
}

func (ops *vsphereOps) Detach(diskPath string) error {
	return ops.detachInternal(diskPath, ops.cfg.VMUUID)
}

func (ops *vsphereOps) DetachFrom(diskPath, instanceID string) error {
	return ops.detachInternal(diskPath, instanceID)
}

func (ops *vsphereOps) detachInternal(diskPath, instanceID string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var vmObj *vclib.VirtualMachine
	var err error
	if instanceID == ops.cfg.VMUUID {
		vmObj, err = ops.renewVM(ctx, ops.vm)
		if err != nil {
			return err
		}
	} else {
		vmObj, err = GetVMObject(ctx, ops.conn, instanceID)
		if err != nil {
			return err
		}
	}

	if err := vmObj.DetachDisk(ctx, diskPath); err != nil {
		logrus.Errorf("Failed to detach vsphere disk: %s for VM: %s. err: +%v", diskPath, vmObj.Name(), err)
		return err
	}

	return nil
}

// Delete virtual disk at given path
func (ops *vsphereOps) Delete(diskPath string) error {
	return ops.deleteInternal(diskPath, ops.cfg.VMUUID)
}

func (ops *vsphereOps) DeleteFrom(diskPath, instanceID string) error {
	return ops.deleteInternal(diskPath, instanceID)
}

func (ops *vsphereOps) deleteInternal(diskPath, instanceID string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var vmObj *vclib.VirtualMachine
	var err error
	if instanceID == ops.cfg.VMUUID {
		vmObj, err = ops.renewVM(ctx, ops.vm)
		if err != nil {
			return err
		}
	} else {
		vmObj, err = GetVMObject(ctx, ops.conn, instanceID)
		if err != nil {
			return err
		}
	}

	disk := diskmanagers.VirtualDisk{
		DiskPath:      diskPath,
		VolumeOptions: &vclib.VolumeOptions{},
		VMOptions:     &vclib.VMOptions{},
	}

	err = disk.Delete(ctx, vmObj.Datacenter)
	if err != nil {
		logrus.Errorf("Failed to delete vsphere disk: %s. err: %+v", diskPath, err)
	}

	return err
}

// Desribe an instance of the virtual machine object to which ops is connected to
func (ops *vsphereOps) Describe() (interface{}, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return ops.renewVM(ctx, ops.vm)
}

// FreeDevices is not supported by this provider
func (ops *vsphereOps) FreeDevices(blockDeviceMappings []interface{}, rootDeviceName string) ([]string, error) {
	return nil, &cloudops.ErrNotSupported{
		Operation: "FreeDevices",
	}
}

func (ops *vsphereOps) Inspect(vmdksWithDS []*string) ([]interface{}, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vmObj, err := ops.renewVM(ctx, ops.vm)
	if err != nil {
		return nil, err
	}
	f := find.NewFinder(vmObj.Client(), true)
	f.SetDatacenter(vmObj.Datacenter.Datacenter)

	dsMap := make(map[string]*object.Datastore)
	disks := []interface{}{}
	for _, vmdkPathWithDS := range vmdksWithDS {
		matches := vmdkMatcherRegex.FindStringSubmatch(*vmdkPathWithDS)
		if len(matches) < 2 {
			return nil, fmt.Errorf("failed to inspect drive: "+
				"failed to parse datastore from vmdk path: %s", *vmdkPathWithDS)
		}
		dsName := matches[1]
		vmdkPath := strings.TrimSpace(matches[2])

		ds, exists := dsMap[dsName]
		var err error
		if !exists {
			ds, err = f.Datastore(context.TODO(), dsName)
			if err != nil {
				return nil, fmt.Errorf("failed to inspect drive: failed to "+
					"get datastore %s for vmdk %s: %s", dsName, vmdkPath, err)
			}
			dsMap[dsName] = ds
		}
		m := object.NewVirtualDiskManager(ds.Client())
		diskInfos, err := m.QueryVirtualDiskInfo(ctx, *vmdkPathWithDS, vmObj.Datacenter.Datacenter, false)
		if err != nil {
			return nil, fmt.Errorf("failed to inspect drive %v: %v", *vmdkPathWithDS, err)
		}
		if len(diskInfos) != 1 {
			return nil, fmt.Errorf("failed to inspect drive %v: found more than %d disks with the same name", *vmdkPathWithDS, len(diskInfos))
		}
		disks = append(disks, &diskInfos[0])
	}
	return disks, nil
}

// DeviceMappings returns map[local_attached_volume_path]->volume ID/NAME
func (ops *vsphereOps) DeviceMappings() (map[string]string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vmObj, err := ops.renewVM(ctx, ops.vm)
	if err != nil {
		return nil, err
	}

	vmDevices, err := vmObj.Device(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices for vm: %s", vmObj.Name())
	}

	// Go over all the devices attached on this vm and create a map of just the virtual disks and where
	// they are attached on the vm
	m := make(map[string]string)
	for _, device := range vmDevices {
		if vmDevices.TypeName(device) == "VirtualDisk" {
			virtualDevice := device.GetVirtualDevice()
			backing, ok := virtualDevice.Backing.(*types.VirtualDiskFlatVer2BackingInfo)
			if ok {
				devicePath, err := ops.DevicePath(backing.FileName)
				if err == nil && len(devicePath) != 0 { // TODO can ignore errors?
					m[devicePath] = backing.FileName
				}
			}
		}
	}

	return m, nil
}

// DevicePath for the given volume i.e path where it's attached
func (ops *vsphereOps) DevicePath(diskPath string) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vmObj, err := ops.renewVM(ctx, ops.vm)
	if err != nil {
		return "", err
	}

	vmName, err := vmObj.ObjectName(ctx)
	if err != nil {
		return "", err
	}

	attached, err := vmObj.IsDiskAttached(ctx, diskPath)
	if err != nil {
		return "", fmt.Errorf("failed to check if disk: %s is attached on vm: %s. err: %v",
			diskPath, vmObj.Name(), err)
	}

	if !attached {
		return "", fmt.Errorf("disk: %s is not attached on vm: %s", diskPath, vmName)
	}

	diskUUID, err := vmObj.Datacenter.GetVirtualDiskPage83Data(ctx, diskPath)
	if err != nil {
		return "", fmt.Errorf("failed to get device path for disk: %s on vm: %s err: %s", diskPath, vmName, err)
	}

	return path.Join(diskByIDPath, diskSCSIPrefix+diskUUID), nil
}

func (ops *vsphereOps) Enumerate(volumeIds []*string,
	labels map[string]string,
	setIdentifier string,
) (map[string][]interface{}, error) {
	return nil, &cloudops.ErrNotSupported{
		Operation: "Enumerate",
	}
}

func (ops *vsphereOps) Expand(
	vmdkPath string,
	newSizeInGiB uint64,
) (uint64, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vm, err := ops.renewVM(ctx, ops.vm)
	if err != nil {
		return 0, err
	}

	vmName, err := vm.ObjectName(ctx)
	if err != nil {
		return 0, err
	}

	vmDevices, err := vm.Device(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get devices for vm: %s", vmName)
	}

	var disks []*types.VirtualDisk
	for _, device := range vmDevices {
		switch md := device.(type) {
		case *types.VirtualDisk:
			b, ok := md.Backing.(types.BaseVirtualDeviceFileBackingInfo)
			if ok {
				if b.GetVirtualDeviceFileBackingInfo().FileName == vmdkPath {
					disks = append(disks, md)
				}
			}

		}
	}

	if len(disks) == 0 {
		return 0, cloudops.NewStorageError(
			cloudops.ErrVolNotFound,
			fmt.Sprintf("vmdk: %s was not found on vm: %s", vmdkPath, vmName),
			vmName)
	} else if len(disks) > 1 {
		return 0, cloudops.NewStorageError(
			cloudops.ErrVolNotFound,
			fmt.Sprintf("multiple vmdks (%d) were found for path: %s on vm: %s",
				len(disks), vmdkPath, vmName),
			vmName)
	}

	editDisk := disks[0]
	newSizeInKiB := int64(newSizeInGiB) * 1024 * 1024
	if editDisk.CapacityInKB >= newSizeInKiB {
		return uint64(editDisk.CapacityInKB / (1024 * 1024)), cloudops.NewStorageError(cloudops.ErrDiskGreaterOrEqualToExpandSize,
			fmt.Sprintf("disk is already has a size: %d KiB greater than or equal "+
				"requested size: %d KiB", editDisk.CapacityInKB, newSizeInKiB), "")
	}

	editDisk.CapacityInKB = newSizeInKiB
	spec := types.VirtualMachineConfigSpec{}
	config := &types.VirtualDeviceConfigSpec{
		Device:    editDisk,
		Operation: types.VirtualDeviceConfigSpecOperationEdit,
	}

	config.FileOperation = ""
	spec.DeviceChange = append(spec.DeviceChange, config)

	task, err := vm.Reconfigure(ctx, spec)
	if err != nil {
		return 0, err
	}

	err = task.Wait(ctx)
	if err != nil {
		return 0, fmt.Errorf("error resizing vmdk: %s due to:  %s", vmdkPath, err)
	}

	return newSizeInGiB, nil
}

// Snapshot the volume with given volumeID
func (ops *vsphereOps) Snapshot(volumeID string, readonly bool) (interface{}, error) {
	return nil, &cloudops.ErrNotSupported{
		Operation: "Snapshot",
	}
}

// SnapshotDelete deletes the snapshot with given ID
func (ops *vsphereOps) SnapshotDelete(snapID string) error {
	return &cloudops.ErrNotSupported{
		Operation: "SnapshotDelete",
	}
}

// ApplyTags will apply given labels/tags on the given volume
func (ops *vsphereOps) ApplyTags(volumeID string, labels map[string]string) error {
	return &cloudops.ErrNotSupported{
		Operation: "ApplyTags",
	}
}

// RemoveTags removes labels/tags from the given volume
func (ops *vsphereOps) RemoveTags(volumeID string, labels map[string]string) error {
	return &cloudops.ErrNotSupported{
		Operation: "RemoveTags",
	}
}

// Tags will list the existing labels/tags on the given volume
func (ops *vsphereOps) Tags(volumeID string) (map[string]string, error) {
	return nil, &cloudops.ErrNotSupported{
		Operation: "Tags",
	}
}

// GetVMObject fetches the VirtualMachine object corresponding to the given virtual machine uuid
func GetVMObject(ctx context.Context, conn *vclib.VSphereConnection, vmUUID string) (*vclib.VirtualMachine, error) {
	// TODO change impl below using multiple goroutines and sync.WaitGroup to make it faster
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := conn.Connect(ctx); err != nil {
		return nil, err
	}

	if len(vmUUID) == 0 {
		return nil, fmt.Errorf("virtual machine uuid is required")
	}

	datacenterObjs, err := vclib.GetAllDatacenter(ctx, conn)
	if err != nil {
		return nil, err
	}

	// Lookup in each vsphere datacenter for this virtual machine
	for _, dc := range datacenterObjs {
		vm, err := dc.GetVMByUUID(ctx, vmUUID)
		if err != nil {
			if err != vclib.ErrNoVMFound {
				logrus.Warnf("failed to find vm with uuid: %s in datacenter: %s due to err: %v", vmUUID, dc.Name(), err)
				// don't let one bad egg fail entire search. keep looking.
			} else {
				logrus.Debugf("did not find vm with uuid: %s in datacenter: %s", vmUUID, dc.Name())
			}
			continue
		}

		if vm != nil {
			return vm, nil
		}
	}

	return nil, fmt.Errorf("failed to find vm with uuid: %s in any datacenter for vc: %s", vmUUID, conn.Hostname)
}

func (ops *vsphereOps) renewVM(ctx context.Context, vm *vclib.VirtualMachine) (*vclib.VirtualMachine, error) {
	var client *vim25.Client
	err := ops.conn.Connect(ctx)
	if err != nil {
		client, err = ops.conn.NewClient(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		client = ops.conn.Client
	}

	vmObj := vm.RenewVM(client)
	return &vmObj, nil
}

// getDatastoreToUseInStoragePod asks the storage resource manager to recommend a datastore
// in the given storage pod (datastore cluster) for the required disk spec
func (ops *vsphereOps) getDatastoreToUseInStoragePod(
	ctx context.Context, vmObj *vclib.VirtualMachine,
	volumeOptions *vclib.VolumeOptions, storagePod *object.StoragePod) (string, error) {
	logrus.Infof("Using storage pod: %s", storagePod.Name())

	// devices is a list of devices in the virtual machine (disks and disk controllers) that
	// will be part of the request spec to storage resource manager
	var devices object.VirtualDeviceList
	scsi, err := devices.CreateSCSIController("scsi")
	if err != nil {
		return "", err
	}

	devices = append(devices, scsi)

	controller, err := devices.FindDiskController("scsi")
	if err != nil {
		return "", err
	}

	disk := &types.VirtualDisk{
		VirtualDevice: types.VirtualDevice{
			Key: devices.NewKey(),
			Backing: &types.VirtualDiskFlatVer2BackingInfo{
				DiskMode:        string(types.VirtualDiskModePersistent),
				ThinProvisioned: types.NewBool(true),
			},
		},
		CapacityInKB: int64(volumeOptions.CapacityKB),
	}

	devices = append(devices, disk)
	devices.AssignController(disk, controller)
	deviceChange, err := devices.ConfigSpec(types.VirtualDeviceConfigSpecOperationAdd)
	if err != nil {
		return "", err
	}

	spec := &types.VirtualMachineConfigSpec{
		Name: vmObj.Name(),
	}

	spec.DeviceChange = deviceChange
	ops.dsLock.LockWithKey(ops.cfg.VMUUID, storagePod.Name())
	recommendedDatastore, err := recommendDatastore(ctx, vmObj, storagePod, spec)
	ops.dsLock.UnlockWithKey(storagePod.Name())
	if err != nil {
		return "", err
	}

	return recommendedDatastore.Name(), nil
}

// recommendedDatastore recommends a datastore to use for the given storage pod by
// quering the storage resource manager
// logic borrowwed from recommendDatastore() at https://github.com/vmware/govmomi/blob/master/govc/vm/create.go#L455
func recommendDatastore(
	ctx context.Context,
	vmObj *vclib.VirtualMachine,
	storagePod *object.StoragePod,
	spec *types.VirtualMachineConfigSpec) (*object.Datastore, error) {
	sp := storagePod.Reference()

	// Build pod selection spec from config spec
	podSelectionSpec := types.StorageDrsPodSelectionSpec{
		StoragePod: &sp,
	}

	for _, deviceConfigSpec := range spec.DeviceChange {
		s := deviceConfigSpec.GetVirtualDeviceConfigSpec()
		if s.Operation != types.VirtualDeviceConfigSpecOperationAdd {
			continue
		}

		if s.FileOperation != types.VirtualDeviceConfigSpecFileOperationCreate {
			continue
		}

		d, ok := s.Device.(*types.VirtualDisk)
		if !ok {
			continue
		}

		podConfigForPlacement := types.VmPodConfigForPlacement{
			StoragePod: sp,
			Disk: []types.PodDiskLocator{
				{
					DiskId:          d.Key,
					DiskBackingInfo: d.Backing,
				},
			},
		}

		podSelectionSpec.InitialVmConfig = append(podSelectionSpec.InitialVmConfig, podConfigForPlacement)
	}

	resourcePool, err := vmObj.ResourcePool(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get vm resource pool due to: %v", err)
	}

	if resourcePool == nil {
		return nil, fmt.Errorf("failed to get vm resource pool")
	}

	resourcePoolRef := resourcePool.Reference()

	sps := types.StoragePlacementSpec{
		Type:             string(types.StoragePlacementSpecPlacementTypeCreate),
		PodSelectionSpec: podSelectionSpec,
		ConfigSpec:       spec,
		ResourcePool:     &resourcePoolRef,
		// lock the recommended resources for us for the next 90 seconds
		ResourceLeaseDurationSec: 90,
	}

	srm := object.NewStorageResourceManager(vmObj.Client())
	result, err := srm.RecommendDatastores(ctx, sps)
	if err != nil {
		logrus.Errorf("failed to get datastore recommendations due to: %v", err)
		return nil, err
	}

	// Use result to pin disks to recommended datastores
	recs := result.Recommendations
	if len(recs) == 0 {
		return nil, fmt.Errorf("no datastores recommendations")
	}

	var ds types.ManagedObjectReference
	/*
	recs are the recommendations made by vSphere. There will be multiple datastore recommendations made by vSphere
	and it's expected for us to select one manually based on any of the parameters given to us. The below code picks
	the best datastore based on two factors:
	1. Rating: Valid values range from 1 (lowest confidence) to 5 (highest confidence).
	2. Space Utilization after virtual disk placement: It's percentage value
	If there is a higher rating recommendation with the best space utilization, we'll select that.
	Many recommendations will be made with the same rating. In those cases we'll pick the one with lease space util.
	What is interesting will be when a lower recommendation has very less space util than a higher rating one. In order
	to consider those cases, let's assign a 9% value for each rating and decide which one to pick. Let's consider some examples:
	1. Rating 5, Util 41.00 vs Rating 4, Util 30.0000 -> We'll pick the second one since it's atleast 10% less utilized
	2. Rating 5, Util 50.000 vs Rating 3, Util 20.000 -> We'll pick the second one again for the same reason
	*/
	minSpaceUtilization := float32(200)
	const (
		maxRating = int32(5)
		normalizationPercent = float32(9)
	)
	for i, r := range recs {
		logrus.Infof("vsphere datastore recommendation #%v: %v", i, r)
		for _, a := range r.Action {
			action := a.(*types.StoragePlacementAction)
			logrus.Infof("vsphere recommendation action %v", action)
			normalizedUtilization := (float32(maxRating - r.Rating) * normalizationPercent) + action.SpaceUtilAfter
			if normalizedUtilization < minSpaceUtilization {
				 minSpaceUtilization = normalizedUtilization
				 ds = action.Destination
			}
		}
	}

	var mds mo.Datastore
	err = property.DefaultCollector(vmObj.Client()).RetrieveOne(ctx, ds, []string{"name"}, &mds)
	if err != nil {
		return nil, err
	}

	datastore := object.NewDatastore(vmObj.Client(), ds)
	datastore.InventoryPath = mds.Name

	return datastore, nil
}

// IsStoragePod checks if the object with given name is a StoragePod (Datastore cluster)
func IsStoragePod(ctx context.Context, vmObj *vclib.VirtualMachine, name string) (bool, *object.StoragePod, error) {
	f := find.NewFinder(vmObj.Client(), true)
	f.SetDatacenter(vmObj.Datacenter.Datacenter)
	sp, err := f.DatastoreCluster(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil, nil
		}

		logrus.Errorf("got error: %v fetching datastore cluster: %s", err, name)
		return false, nil, err
	}

	if sp == nil {
		return false, nil, nil
	}

	return true, sp, nil
}

func isExponentialError(err error) bool {
	retryErrors := map[string]struct{}{
		// ServerFaultCode is received from the vCenter API when we encounter intermittent errors on the vCenter
		// server side and typically they always get resolved on retries
		"ServerFaultCode": {},
	}
	if err != nil {
		for retryErr := range retryErrors {
			if strings.Contains(err.Error(), retryErr) {
				return true
			}
		}
	}
	return false
}
